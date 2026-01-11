# Woodpecker Config Provider - Multi-Platform SDK

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-ARM64%20%7C%20AMD64-2496ED?logo=docker)](https://hub.docker.com/)

支持 **Gitea、GitHub、GitLab** 三大平台的 Woodpecker CI 配置提供器，实现集中式多文件 Pipeline 配置管理。

## ✨ 核心特性

### 🌐 多平台支持
- ✅ **Gitea** - 使用官方 SDK (`code.gitea.io/sdk/gitea`)
- ✅ **GitHub** - 使用官方 SDK (`github.com/google/go-github/v57`)
- ✅ **GitLab** - 使用官方 SDK (`gitlab.com/gitlab-org/api/client-go`)

### 📁 多文件配置
- ✅ 自动读取目录下所有 `.yml` 和 `.yaml` 文件
- ✅ 支持 `build.yml`、`test.yml`、`deploy.yml` 分离
- ✅ 每个文件独立显示在 Woodpecker UI

### 🎯 模板引擎
- ✅ Go template 语法支持
- ✅ 动态路径：`{{ .Repo.Name }}/{{ .Pipeline.Branch }}`
- ✅ 灵活的命名空间配置

### 🚀 生产就绪
- ✅ DEBUG 模式调试
- ✅ 完整的测试覆盖

## 📋 目录结构示例

```
dronefiles/                    # 配置仓库
├── project-a/
│   ├── main/
│   │   ├── build.yml         # 构建 pipeline
│   │   ├── test.yml          # 测试 pipeline
│   │   └── deploy.yml        # 部署 pipeline
│   └── develop/
│       ├── build.yml
│       └── test.yml
├── project-b/
│   └── main/
│       └── build.yml
└── shared/
    └── common.yml
```

## 🚀 快速开始

### 方式 1: Docker 部署（推荐）

```yaml
# docker-compose.yml
services:
  woodpecker-config-provider:
    image: ghcr.io/yahuiwong/woodpecker-config-provider:latest
    ports:
      - "8000:8000"
    environment:
      # 基础配置
      - SERVERTYPE=gitea                    # gitea | github | gitlab
      - SERVER_URL=https://git.example.com
      - TOKEN=your_access_token

      # 模板配置（使用 Woodpecker 风格命名）
      - WOODPECKER_CONFIG_NAMESPACE_TEMP={{ .Repo.Owner }}
      - WOODPECKER_CONFIG_REPONAME_TEMP=dronefiles
      - WOODPECKER_CONFIG_BRANCH_TEMP={{ .Pipeline.Branch }}
      - WOODPECKER_CONFIG_YAMLPATH_TEMP={{ .Repo.Name }}/{{ .Pipeline.Branch }}

      # 可选：调试模式
      - PLUGIN_DEBUG=false
```

### 方式 2: 本地构建

```bash
# 克隆仓库
git clone https://github.com/YahuiWong/woodpecker-config-provider.git
cd woodpecker-config-provider

# 安装依赖
go mod download

# 构建
go build -o woodpecker-config-provider .

# 运行
export SERVERTYPE=gitea
export SERVER_URL=https://git.example.com
export TOKEN=your_token
./woodpecker-config-provider
```

### 方式 3: 从源码构建 Docker 镜像

```bash
# 构建镜像（支持 ARM64 和 AMD64）
docker build -t woodpecker-config-provider:latest .

# 运行
docker run -p 8000:8000 \
  -e SERVERTYPE=gitea \
  -e SERVER_URL=https://git.example.com \
  -e TOKEN=your_token \
  woodpecker-config-provider:latest
```

## ⚙️ 配置指南

### Gitea 配置

```yaml
environment:
  - SERVERTYPE=gitea
  - SERVER_URL=https://git.example.com
  - TOKEN=your_gitea_access_token

  # 模板配置
  - WOODPECKER_CONFIG_NAMESPACE_TEMP={{ .Repo.Owner }}
  - WOODPECKER_CONFIG_REPONAME_TEMP=dronefiles
  - WOODPECKER_CONFIG_BRANCH_TEMP={{ .Pipeline.Branch }}
  - WOODPECKER_CONFIG_YAMLPATH_TEMP={{ .Repo.Name }}/{{ .Pipeline.Branch }}
```

**生成 Token:**
1. Gitea → 用户设置 → 应用 → 访问令牌
2. 权限：`repo:read`

### GitHub 配置

```yaml
environment:
  - SERVERTYPE=github
  - SERVER_URL=https://api.github.com     # GitHub.com
  # - SERVER_URL=https://github.enterprise.com/api/v3  # GitHub Enterprise
  - TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx

  - WOODPECKER_CONFIG_NAMESPACE_TEMP={{ .Repo.Owner }}
  - WOODPECKER_CONFIG_REPONAME_TEMP=dronefiles
  - WOODPECKER_CONFIG_BRANCH_TEMP={{ .Pipeline.Branch }}
  - WOODPECKER_CONFIG_YAMLPATH_TEMP={{ .Repo.Name }}/{{ .Pipeline.Branch }}
```

**生成 Token:**
1. GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. 权限：`repo` (Full control of private repositories)

### GitLab 配置

```yaml
environment:
  - SERVERTYPE=gitlab
  - SERVER_URL=https://gitlab.com         # GitLab.com
  # - SERVER_URL=https://gitlab.company.com  # Self-hosted
  - TOKEN=glpat-xxxxxxxxxxxxxxxxxxxx

  - WOODPECKER_CONFIG_NAMESPACE_TEMP={{ .Repo.Owner }}
  - WOODPECKER_CONFIG_REPONAME_TEMP=dronefiles
  - WOODPECKER_CONFIG_BRANCH_TEMP={{ .Pipeline.Branch }}
  - WOODPECKER_CONFIG_YAMLPATH_TEMP={{ .Repo.Name }}/{{ .Pipeline.Branch }}
```

**生成 Token:**
1. GitLab → User Settings → Access Tokens
2. 权限：`read_api`, `read_repository`

## 📝 环境变量参考

### 基础配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SERVERTYPE` | `gitea` | Git 平台类型：`gitea`/`github`/`gitlab` |
| `SERVER_URL` | `https://git.local.lan` | Git 服务器 URL |
| `TOKEN` | - | 访问令牌（必需） |
| `PLUGIN_DEBUG` | `false` | 启用调试日志 |

### 模板配置（Woodpecker 风格）

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `WOODPECKER_CONFIG_NAMESPACE_TEMP` | `{{ .Repo.Owner }}` | 命名空间模板 |
| `WOODPECKER_CONFIG_REPONAME_TEMP` | `dronefiles` | 配置仓库名 |
| `WOODPECKER_CONFIG_BRANCH_TEMP` | `{{ .Pipeline.Branch }}` | 分支模板 |
| `WOODPECKER_CONFIG_YAMLPATH_TEMP` | `{{ .Repo.Name }}/{{ .Pipeline.Branch }}` | 配置路径模板 |

### 兼容配置（Drone 风格，自动 fallback）

| 变量 | 说明 |
|------|------|
| `DRONE_CONFIG_NAMESPACE_TEMP` | 如果 `WOODPECKER_*` 未设置则使用 |
| `DRONE_CONFIG_REPONAME_TEMP` | 同上 |
| `DRONE_CONFIG_BRANCH_TEMP` | 同上 |
| `DRONE_CONFIG_YAMLPATH_TEMP` | 同上 |
| `GITEA_URL` | fallback to `SERVER_URL` |
| `GITEA_TOKEN` | fallback to `TOKEN` |

## 🎨 模板语法

支持 Go template 语法，可用变量：

```go
// 仓库信息
.Repo.Name      // 仓库名称，如 "myproject"
.Repo.Owner     // 仓库所有者，如 "admin"
.Repo.FullName  // 完整名称，如 "admin/myproject"

// Pipeline 信息
.Pipeline.Branch // 分支名称，如 "main"
.Pipeline.Commit // 提交 SHA
.Pipeline.Ref    // Git ref
```

### 模板示例

```yaml
# 示例 1: 标准配置
WOODPECKER_CONFIG_YAMLPATH_TEMP={{ .Repo.Name }}/{{ .Pipeline.Branch }}
# 结果: myproject/main

# 示例 2: 多租户配置
WOODPECKER_CONFIG_NAMESPACE_TEMP={{ .Repo.Owner }}
WOODPECKER_CONFIG_YAMLPATH_TEMP={{ .Repo.Owner }}/{{ .Repo.Name }}/{{ .Pipeline.Branch }}
# 结果: admin/myproject/main

# 示例 3: 固定分支
WOODPECKER_CONFIG_BRANCH_TEMP=main
WOODPECKER_CONFIG_YAMLPATH_TEMP={{ .Repo.Name }}/common
# 结果: myproject/common（所有分支共用）
```

## 🔌 集成 Woodpecker Server

### 1. 更新 Woodpecker Server 配置

```yaml
woodpecker-server:
  environment:
    # 启用配置服务
    - WOODPECKER_CONFIG_SERVICE_ENDPOINT=http://woodpecker-config-provider:8000/ciconfig

    # 允许的主机（如果使用自签名证书）
    - WOODPECKER_EXTENSIONS_ALLOWED_HOSTS=woodpecker-config-provider,loopback,private
```

### 2. 重启 Woodpecker Server

```bash
docker-compose restart woodpecker-server
```

### 3. 验证集成

推送代码到任意仓库，检查 Woodpecker UI 是否显示多个 pipeline。

## 🧪 配置文件示例

### build.yml
```yaml
steps:
  - name: build
    image: golang:1.24
    commands:
      - go mod download
      - go build -o app .
      - echo "Build completed"
```

### test.yml
```yaml
steps:
  - name: test
    image: golang:1.24
    commands:
      - go test -v ./...
      - go test -race ./...
```

### deploy.yml
```yaml
when:
  - event: push
    branch: main

steps:
  - name: deploy
    image: alpine
    commands:
      - echo "Deploying to production..."
      - ./deploy.sh
```

## 📊 API 文档

### Endpoint: `POST /ciconfig`

**请求格式:**
```json
{
  "repo": {
    "name": "myproject",
    "owner": "admin",
    "full_name": "admin/myproject",
    "default_branch": "main"
  },
  "pipeline": {
    "branch": "main",
    "commit": "abc123...",
    "ref": "refs/heads/main"
  }
}
```

**响应格式:**
```json
{
  "configs": [
    {
      "name": "build",
      "data": "steps:\n  - name: build\n    image: golang:1.24\n    commands:\n      - go build ."
    },
    {
      "name": "test",
      "data": "steps:\n  - name: test\n    image: golang:1.24\n    commands:\n      - go test ./..."
    }
  ]
}
```

### Health Check: `GET /`

```bash
curl http://localhost:8000/
```

**响应:**
```json
{
  "status": "ok",
  "service": "Woodpecker Config Provider (Enhanced Multi-file)",
  "version": "2.0.0",
  "config": {
    "server_type": "gitea",
    "namespace_tmpl": "{{ .Repo.Owner }}",
    "reponame_tmpl": "dronefiles",
    "branch_tmpl": "{{ .Pipeline.Branch }}",
    "path_tmpl": "{{ .Repo.Name }}/{{ .Pipeline.Branch }}",
    "debug": "false"
  }
}
```

## 🐛 调试与故障排查

### 启用调试模式

```yaml
environment:
  - PLUGIN_DEBUG=true
```

### 查看日志

```bash
# Docker
docker-compose logs -f woodpecker-config-provider

# 本地运行
# 日志会输出到 stdout
```

### 调试日志示例

```
[DEBUG] Parsed request - Repo: myproject, Branch: main, Owner: admin
[DEBUG] Template: {{ .Repo.Owner }} => admin
[DEBUG] Template: dronefiles => dronefiles
[DEBUG] Template: {{ .Pipeline.Branch }} => main
[DEBUG] Template: {{ .Repo.Name }}/{{ .Pipeline.Branch }} => myproject/main
[DEBUG] Resolved values - Namespace: admin, Repo: dronefiles, Branch: main, Path: myproject/main
[DEBUG] fetchFilesFromGitea - namespace: admin, repo: dronefiles, branch: main, path: myproject/main
[DEBUG] Found 3 items in directory
[DEBUG]   Processing file: build.yml
[DEBUG]     ✓ Loaded build.yml (245 bytes)
[DEBUG]   Processing file: test.yml
[DEBUG]     ✓ Loaded test.yml (189 bytes)
[DEBUG]   Processing file: deploy.yml
[DEBUG]     ✓ Loaded deploy.yml (156 bytes)
[DEBUG] Total files loaded: 3
[DEBUG]     ✓ YAML validation passed
```

### 常见问题

#### 1. 404 错误：配置目录不存在

```
[DEBUG] Response status: 404
```

**原因：**
- dronefiles 仓库中不存在对应的目录
- 路径模板配置错误

**解决：**
```bash
# 检查路径
# 如果路径是 myproject/main，确保 dronefiles 仓库中存在该目录
```

#### 2. 401 错误：认证失败

```
[DEBUG] ERROR: HTTP request failed: 401 Unauthorized
```

**原因：**
- Token 无效或过期
- Token 权限不足

**解决：**
- 重新生成 Token
- 确保 Token 权限包含 `repo:read` (Gitea) 或 `repo` (GitHub) 或 `read_api` (GitLab)

#### 3. 无配置文件返回

```
[DEBUG] Total files loaded: 0
[DEBUG] Found 0 config files
```

**原因：**
- 目录中没有 `.yml` 或 `.yaml` 文件
- 文件被忽略（非 file 类型）

**解决：**
- 检查 dronefiles 仓库中的文件扩展名
- 确保文件类型为 `file`，不是 `dir` 或其他

#### 4. YAML 解析失败

```
[DEBUG] WARNING: YAML validation failed: yaml: line 5: did not find expected key
```

**原因：**
- YAML 语法错误
- 缩进问题

**解决：**
- 使用 YAML 验证工具检查语法
- 确保使用空格缩进（不要用 Tab）

## 🛠️ 开发指南

### 环境要求

- Go 1.24+
- Docker (可选)

### 本地开发

```bash
# 克隆仓库
git clone https://github.com/YahuiWong/woodpecker-config-provider.git
cd woodpecker-config-provider

# 安装依赖
go mod download

# 运行测试
go test -v ./...

# 运行（带 DEBUG）
export PLUGIN_DEBUG=true
export SERVERTYPE=gitea
export SERVER_URL=https://git.example.com
export TOKEN=your_token
go run .
```

### 运行测试

```bash
# 运行所有测试
go test -v ./...

# 运行特定测试
go test -v -run TestTemplateRendering

# 查看测试覆盖率
go test -cover ./...
```

### 构建二进制

```bash
# 本地架构
go build -o woodpecker-config-provider .

# 交叉编译 ARM64
GOOS=linux GOARCH=arm64 go build -o woodpecker-config-provider-arm64 .

# 交叉编译 AMD64
GOOS=linux GOARCH=amd64 go build -o woodpecker-config-provider-amd64 .

# 带优化标志
go build -ldflags="-w -s" -o woodpecker-config-provider .
```

### Docker 构建

```bash
# 构建单架构
docker build -t woodpecker-config-provider:latest .

# 构建多架构（需要 buildx）
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t woodpecker-config-provider:latest \
  --push .
```

## 📦 项目结构

```
.
├── main.go                    # 主程序（Gitea SDK + 核心逻辑）
├── github_gitlab.go           # GitHub 和 GitLab SDK 实现
├── main_test.go              # ConfigResponse 和 YAML 解析测试
├── yaml_test.go              # YAML 边界情况测试
├── template_test.go          # 模板渲染测试
├── go.mod                    # Go 依赖
├── go.sum                    # 依赖校验
├── Dockerfile                # Docker 镜像构建
├── .dockerignore            # Docker 构建忽略文件
├── .gitignore               # Git 忽略规则
└── README.md                # 本文件
```

## 📚 依赖库

| 库 | 版本 | 用途 |
|----|------|------|
| `code.gitea.io/sdk/gitea` | v0.22.1 | Gitea API 客户端 |
| `github.com/google/go-github/v57` | v57.0.0 | GitHub API 客户端 |
| `gitlab.com/gitlab-org/api/client-go` | v1.11.0 | GitLab API 客户端 |
| `gopkg.in/yaml.v3` | v3.0.1 | YAML 解析 |

完整依赖列表请查看 `go.mod`。

## 🔄 更新日志

### v2.0.0 (2026-01-11)

**重大更新：**
- ✨ 迁移到官方 SDK（Gitea、GitHub、GitLab）
- ✨ 支持 WOODPECKER_* 环境变量命名
- ✨ 完整的 GitHub 和 GitLab 支持
- ✨ Go 1.24 支持

**改进：**
- 🔧 移除 USE_BASE64 选项（SDK 自动处理）
- 🔧 简化配置（减少冗余环境变量）
- 🔧 清理未使用代码
- 📝 更新文档

**修复：**
- 🐛 修复 Gitea SDK GetFile() Base64 解码问题
- 🐛 修复 Go 版本兼容性

### v1.0.0 (2026-01-10)

- 🎉 初始版本
- ✅ 基础多文件配置支持
- ✅ Gitea 集成
- ✅ ARM64 支持

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 🔗 相关链接

- [Woodpecker CI](https://woodpecker-ci.org/)
- [Woodpecker Config Extensions](https://woodpecker-ci.org/docs/administration/external-configuration-api)
- [Gitea API](https://docs.gitea.io/en-us/api-usage/)
- [GitHub REST API](https://docs.github.com/en/rest)
- [GitLab API](https://docs.gitlab.com/ee/api/)

## 💬 支持

如有问题或建议：
- 📝 提交 [Issue](https://github.com/YahuiWong/woodpecker-config-provider/issues)
- 💡 参与 [Discussions](https://github.com/YahuiWong/woodpecker-config-provider/discussions)

---

**由 Claude Sonnet 4.5 协助开发** 🤖
