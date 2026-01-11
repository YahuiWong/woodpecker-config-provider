package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/google/go-github/v57/github"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// 从 GitHub 获取目录下所有文件
func fetchFilesFromGitHub(namespace, repo, branch, path string) ([]GiteaFile, error) {
	debugLog("fetchFilesFromGitHub - namespace: %s, repo: %s, branch: %s, path: %s",
		namespace, repo, branch, path)

	// 创建 HTTP 客户端（支持自签名证书）
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// 创建 GitHub 客户端
	client := github.NewClient(httpClient).WithAuthToken(Token)

	// 如果是自托管 GitHub Enterprise，设置 BaseURL
	if !strings.Contains(ServerURL, "api.github.com") {
		var err error
		client, err = client.WithEnterpriseURLs(ServerURL, ServerURL)
		if err != nil {
			debugLog("ERROR: Failed to set GitHub Enterprise URL: %v", err)
			return nil, err
		}
	}

	ctx := context.Background()

	// 获取目录内容
	_, directoryContent, _, err := client.Repositories.GetContents(ctx, namespace, repo, path, &github.RepositoryContentGetOptions{
		Ref: branch,
	})
	if err != nil {
		debugLog("ERROR: Failed to get directory contents: %v", err)
		return nil, err
	}

	debugLog("Found %d items in directory", len(directoryContent))

	// 处理每个文件
	var result []GiteaFile
	for _, content := range directoryContent {
		// 只处理 .yml 和 .yaml 文件
		if content.GetType() == "file" && (strings.HasSuffix(content.GetName(), ".yml") || strings.HasSuffix(content.GetName(), ".yaml")) {
			debugLog("  Processing file: %s", content.GetName())

			// 获取文件内容
			fileContent, _, _, err := client.Repositories.GetContents(ctx, namespace, repo, content.GetPath(), &github.RepositoryContentGetOptions{
				Ref: branch,
			})
			if err != nil {
				debugLog("    ERROR: Failed to fetch file: %v", err)
				continue
			}

			// GitHub SDK 已经解码了 Base64
			decodedContent, err := fileContent.GetContent()
			if err != nil {
				debugLog("    ERROR: Failed to get content: %v", err)
				continue
			}

			giteaFile := GiteaFile{
				Name:    content.GetName(),
				Path:    content.GetPath(),
				Type:    "file",
				Content: decodedContent,
			}

			debugLog("    ✓ Loaded %s (%d bytes)", content.GetName(), len(giteaFile.Content))
			result = append(result, giteaFile)
		} else {
			debugLog("  Skipping: %s (type: %s)", content.GetName(), content.GetType())
		}
	}

	debugLog("Total files loaded: %d", len(result))
	return result, nil
}

// 从 GitLab 获取目录下所有文件
func fetchFilesFromGitLab(namespace, repo, branch, path string) ([]GiteaFile, error) {
	debugLog("fetchFilesFromGitLab - namespace: %s, repo: %s, branch: %s, path: %s",
		namespace, repo, branch, path)

	// 创建 HTTP 客户端（支持自签名证书）
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// 创建 GitLab 客户端
	client, err := gitlab.NewClient(Token,
		gitlab.WithBaseURL(ServerURL),
		gitlab.WithHTTPClient(httpClient),
	)
	if err != nil {
		debugLog("ERROR: Failed to create GitLab client: %v", err)
		return nil, err
	}

	// GitLab 项目 ID（格式：namespace/repo）
	projectID := namespace + "/" + repo

	// 列出目录树
	treeOptions := &gitlab.ListTreeOptions{
		Path: &path,
		Ref:  &branch,
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	trees, _, err := client.Repositories.ListTree(projectID, treeOptions)
	if err != nil {
		debugLog("ERROR: Failed to list tree: %v", err)
		return nil, err
	}

	debugLog("Found %d items in directory", len(trees))

	// 处理每个文件
	var result []GiteaFile
	for _, tree := range trees {
		// 只处理 .yml 和 .yaml 文件
		if tree.Type == "blob" && (strings.HasSuffix(tree.Name, ".yml") || strings.HasSuffix(tree.Name, ".yaml")) {
			debugLog("  Processing file: %s", tree.Name)

			// 获取文件内容
			fileOptions := &gitlab.GetFileOptions{
				Ref: &branch,
			}

			file, _, err := client.RepositoryFiles.GetFile(projectID, tree.Path, fileOptions)
			if err != nil {
				debugLog("    ERROR: Failed to fetch file: %v", err)
				continue
			}

			// GitLab SDK 返回 Base64 编码的内容，需要解码
			decodedContent, err := base64.StdEncoding.DecodeString(file.Content)
			if err != nil {
				debugLog("    ERROR: Failed to decode base64: %v", err)
				continue
			}

			giteaFile := GiteaFile{
				Name:    tree.Name,
				Path:    tree.Path,
				Type:    "file",
				Content: string(decodedContent),
			}

			debugLog("    ✓ Loaded %s (%d bytes)", tree.Name, len(giteaFile.Content))
			result = append(result, giteaFile)
		} else {
			debugLog("  Skipping: %s (type: %s)", tree.Name, tree.Type)
		}
	}

	debugLog("Total files loaded: %d", len(result))
	return result, nil
}
