package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"

	"code.gitea.io/sdk/gitea"
	"gopkg.in/yaml.v3"
)

// 配置变量
var (
	// 基础配置
	Debug      = getEnvBool("PLUGIN_DEBUG", false)
	ServerType = getEnv("SERVERTYPE", "gitea")
	Token      = getEnv("TOKEN", "")
	ServerURL  = getEnv("SERVER_URL", "https://git.local.lan")

	// 模板配置 - Woodpecker 风格（优先）+ Drone 兼容
	NamespaceTemplate = getEnvWithFallback("WOODPECKER_CONFIG_NAMESPACE_TEMP", "DRONE_CONFIG_NAMESPACE_TEMP", "{{ .Repo.Owner }}")
	RepoNameTemplate  = getEnvWithFallback("WOODPECKER_CONFIG_REPONAME_TEMP", "DRONE_CONFIG_REPONAME_TEMP", "woodpeckerfiles")
	BranchTemplate    = getEnvWithFallback("WOODPECKER_CONFIG_BRANCH_TEMP", "DRONE_CONFIG_BRANCH_TEMP", "{{ .Pipeline.Branch }}")
	PathTemplate      = getEnvWithFallback("WOODPECKER_CONFIG_YAMLPATH_TEMP", "DRONE_CONFIG_YAMLPATH_TEMP", "{{ .Repo.Name }}/{{ .Pipeline.Branch }}")

	// 兼容旧版配置
	GiteaURL   = getEnv("GITEA_URL", ServerURL)
	GiteaToken = getEnv("GITEA_TOKEN", Token)
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvWithFallback(primary, fallback, defaultValue string) string {
	// 优先使用 primary 环境变量
	if value := os.Getenv(primary); value != "" {
		return value
	}
	// 如果 primary 不存在，使用 fallback
	if value := os.Getenv(fallback); value != "" {
		return value
	}
	// 都不存在，使用默认值
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

func debugLog(format string, args ...interface{}) {
	if Debug {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// Woodpecker 请求结构
type ConfigRequest struct {
	Repo     RepoInfo     `json:"repo"`
	Pipeline PipelineInfo `json:"pipeline"`
	Config   ConfigInfo   `json:"config"`
}

type RepoInfo struct {
	Name      string `json:"name"`
	Owner     string `json:"owner"`
	Namespace string `json:"-"` // 别名，用于模板，从 Owner 复制值
	FullName  string `json:"full_name"`
	CloneURL  string `json:"clone_url"`
	Branch    string `json:"default_branch"`
}

type PipelineInfo struct {
	Branch string `json:"branch"`
	Commit string `json:"commit"`
	Ref    string `json:"ref"`
}

type ConfigInfo struct {
	Data string `json:"data"`
}

// 模板数据
type TemplateData struct {
	Repo     RepoInfo
	Pipeline PipelineInfo
}

// Woodpecker 响应结构（多文件）
type ConfigResponse struct {
	Configs []ConfigFile `json:"configs"`
}

type ConfigFile struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// Gitea API 响应
type GiteaFile struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

// 渲染模板
func renderTemplate(tmplStr string, data TemplateData) (string, error) {
	tmpl, err := template.New("config").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse template error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template error: %w", err)
	}

	result := buf.String()
	debugLog("Template: %s => %s", tmplStr, result)
	return result, nil
}

// 从 Git 服务器获取文件
func fetchFilesFromGitServer(req ConfigRequest) ([]GiteaFile, error) {
	// 准备模板数据
	data := TemplateData{
		Repo:     req.Repo,
		Pipeline: req.Pipeline,
	}

	// 渲染模板
	namespace, err := renderTemplate(NamespaceTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("render namespace template: %w", err)
	}

	repoName, err := renderTemplate(RepoNameTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("render reponame template: %w", err)
	}

	branch, err := renderTemplate(BranchTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("render branch template: %w", err)
	}

	path, err := renderTemplate(PathTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("render path template: %w", err)
	}

	debugLog("Resolved values - Namespace: %s, Repo: %s, Branch: %s, Path: %s",
		namespace, repoName, branch, path)

	// 根据服务器类型调用相应的函数
	switch strings.ToLower(ServerType) {
	case "gitea":
		return fetchFilesFromGitea(namespace, repoName, branch, path)
	case "github":
		return fetchFilesFromGitHub(namespace, repoName, branch, path)
	case "gitlab":
		return fetchFilesFromGitLab(namespace, repoName, branch, path)
	default:
		return nil, fmt.Errorf("unsupported server type: %s", ServerType)
	}
}

// 从 Gitea 获取目录下所有文件
func fetchFilesFromGitea(namespace, repo, branch, path string) ([]GiteaFile, error) {
	debugLog("fetchFilesFromGitea - namespace: %s, repo: %s, branch: %s, path: %s",
		namespace, repo, branch, path)

	// 创建 Gitea 客户端
	client, err := gitea.NewClient(GiteaURL,
		gitea.SetToken(GiteaToken),
		gitea.SetHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}),
	)
	if err != nil {
		debugLog("ERROR: Failed to create Gitea client: %v", err)
		return nil, err
	}

	// 获取目录内容列表
	contentsList, _, err := client.ListContents(namespace, repo, branch, path)
	if err != nil {
		debugLog("ERROR: Failed to get directory contents: %v", err)
		return nil, err
	}

	debugLog("Found %d items in directory", len(contentsList))

	// 处理每个文件
	var result []GiteaFile
	for _, content := range contentsList {
		// 只处理 .yml 和 .yaml 文件
		if content.Type == "file" && (strings.HasSuffix(content.Name, ".yml") || strings.HasSuffix(content.Name, ".yaml")) {
			debugLog("  Processing file: %s", content.Name)

			// 获取文件内容
			fileContent, _, err := client.GetFile(namespace, repo, branch, content.Path)
			if err != nil {
				debugLog("    ERROR: Failed to fetch file: %v", err)
				continue
			}

			// Gitea SDK 的 GetFile() 返回的是原始字节（已解码），直接使用
			giteaFile := GiteaFile{
				Name:    content.Name,
				Path:    content.Path,
				Type:    "file",
				Content: string(fileContent),
			}

			debugLog("    ✓ Loaded %s (%d bytes)", content.Name, len(giteaFile.Content))
			result = append(result, giteaFile)
		} else {
			debugLog("  Skipping: %s (type: %s)", content.Name, content.Type)
		}
	}

	debugLog("Total files loaded: %d", len(result))
	return result, nil
}

// 处理配置请求
func handleConfigRequest(w http.ResponseWriter, r *http.Request) {
	if Debug {
		fmt.Println("=== Config Request Start ===")
	}

	// 1. 解析请求
	var req ConfigRequest
	body, _ := io.ReadAll(r.Body)
	debugLog("Request body: %s", string(body))

	if err := json.Unmarshal(body, &req); err != nil {
		debugLog("ERROR: Failed to parse request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 复制 Owner 到 Namespace（用于模板兼容性）
	req.Repo.Namespace = req.Repo.Owner

	debugLog("Parsed request - Repo: %s, Branch: %s, Owner: %s",
		req.Repo.Name, req.Pipeline.Branch, req.Repo.Owner)

	// 2. 从 Git 服务器获取所有配置文件
	files, err := fetchFilesFromGitServer(req)
	if err != nil {
		debugLog("ERROR: Failed to fetch files: %v", err)
		// 如果目录不存在，返回 204（使用仓库自己的配置）
		w.WriteHeader(http.StatusNoContent)
		return
	}

	debugLog("Found %d config files", len(files))

	// 3. 构建响应
	var configs []ConfigFile

	for _, file := range files {
		// 去掉 .yml 后缀作为 pipeline 名称
		name := strings.TrimSuffix(file.Name, ".yml")
		name = strings.TrimSuffix(name, ".yaml")

		debugLog("  - %s (%d bytes)", file.Name, len(file.Content))

		// SDK 已经返回原始 YAML 内容，直接使用
		// 验证是否是有效的 YAML
		var testData interface{}
		if err := yaml.Unmarshal([]byte(file.Content), &testData); err != nil {
			debugLog("    WARNING: YAML validation failed: %v", err)
		} else {
			debugLog("    ✓ YAML validation passed")
		}

		configs = append(configs, ConfigFile{
			Name: name,
			Data: file.Content,
		})
	}

	// 4. 返回多个配置文件
	response := ConfigResponse{
		Configs: configs,
	}

	if Debug {
		responseJSON, _ := json.MarshalIndent(response, "", "  ")
		fmt.Printf("Response (formatted):\n%s\n", string(responseJSON))
	}

	// 设置 HTTP headers
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// 返回 JSON
	json.NewEncoder(w).Encode(response)

	if Debug {
		fmt.Println("=== Config Request End ===")
	}
}

func main() {
	fmt.Println("Woodpecker Config Provider (Enhanced Multi-file) starting on :8000")
	fmt.Println("Server Type:", ServerType)
	fmt.Println("Server URL:", ServerURL)
	fmt.Println("Template Repo:", RepoNameTemplate)
	fmt.Println("Debug Mode:", Debug)

	if Token == "" && GiteaToken == "" {
		fmt.Println("WARNING: TOKEN is not set!")
	} else {
		tokenToShow := Token
		if tokenToShow == "" {
			tokenToShow = GiteaToken
		}
		if len(tokenToShow) > 16 {
			fmt.Printf("Token configured: %s...%s\n", tokenToShow[:8], tokenToShow[len(tokenToShow)-8:])
		}
	}

	fmt.Println("\nTemplate Configuration:")
	fmt.Println("  Namespace:", NamespaceTemplate)
	fmt.Println("  RepoName:", RepoNameTemplate)
	fmt.Println("  Branch:", BranchTemplate)
	fmt.Println("  Path:", PathTemplate)

	// 配置路由
	http.HandleFunc("/ciconfig", func(w http.ResponseWriter, r *http.Request) {
		debugLog("Received request: %s %s", r.Method, r.URL.Path)

		if r.Method == "POST" {
			handleConfigRequest(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Health check endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		debugLog("Health check: %s %s", r.Method, r.URL.Path)

		if r.Method == "GET" && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "ok",
				"service": "Woodpecker Config Provider (Enhanced Multi-file)",
				"version": "2.0.0",
				"config": map[string]string{
					"server_type":    ServerType,
					"namespace_tmpl": NamespaceTemplate,
					"reponame_tmpl":  RepoNameTemplate,
					"branch_tmpl":    BranchTemplate,
					"path_tmpl":      PathTemplate,
					"debug":          fmt.Sprintf("%v", Debug),
				},
			})
			return
		}

		http.NotFound(w, r)
	})

	fmt.Println("\nStarting HTTP server on :8000")
	http.ListenAndServe(":8000", nil)
}
