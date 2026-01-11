# Woodpecker Config Provider - Multi-file Support

支持从 Gitea `dronefiles` 仓库读取**多个配置文件**的 Woodpecker Config Provider。

## 功能特点

✅ **多文件支持** - 自动读取目录下所有 `.yml` 文件
✅ **ARM64 原生支持** - 针对 Apple Silicon 优化
✅ **Gitea 集成** - 从 dronefiles 仓库读取配置
✅ **动态路径** - 支持 `{{ .Repo.Name }}/{{ .Repo.Branch }}/` 路径结构
✅ **完全兼容** - 兼容 Woodpecker Config Extension API

## 工作原理

### 1. 目录结构

```
dronefiles/
└── myrepo/
    └── main/
        ├── build.yml      # 构建 pipeline
        ├── test.yml       # 测试 pipeline
        └── deploy.yml     # 部署 pipeline
```

### 2. 工作流程

```
1. 推送代码到 myrepo
2. Woodpecker Server 接收 Webhook
3. 调用 Config Provider: POST /ciconfig
4. Config Provider 从 dronefiles/myrepo/main/ 读取所有 .yml 文件
5. 返回多个配置文件给 Woodpecker
6. Woodpecker 执行所有 pipeline（可并行）
```

### 3. 配置文件示例

**dronefiles/myrepo/main/build.yml:**
```yaml
steps:
  - name: build
    image: golang:1.21
    commands:
      - go build -o app .
```

**dronefiles/myrepo/main/test.yml:**
```yaml
steps:
  - name: test
    image: golang:1.21
    commands:
      - go test -v ./...
```

**dronefiles/myrepo/main/deploy.yml:**
```yaml
when:
  branch: main
  event: push

steps:
  - name: deploy
    image: alpine
    commands:
      - echo "Deploying to production..."
```

## 构建和部署

### 1. 构建 ARM64 镜像

```bash
# 构建支持 ARM64 的镜像
docker buildx build --platform linux/arm64 \
  -t woodpecker-config-provider-multifile:arm64 \
  -f Dockerfile .

# 或者构建多平台镜像
docker buildx build --platform linux/amd64,linux/arm64 \
  -t woodpecker-config-provider-multifile:latest \
  -f Dockerfile .
```

### 2. 配置环境变量

编辑 `main.go` 中的配置：

```go
const (
    GiteaURL      = "https://git.local.lan"
    GiteaToken    = "your-token-here"
    TemplateRepo  = "dronefiles"
    TemplateOwner = "admin"
)
```

或者使用环境变量：

```bash
export GITEA_URL="https://git.local.lan"
export GITEA_TOKEN="your-token"
export TEMPLATE_REPO="dronefiles"
export TEMPLATE_OWNER="admin"
```

### 3. 添加到 docker-compose.yml

```yaml
woodpecker-config-provider:
  image: woodpecker-config-provider-multifile:arm64
  container_name: woodpecker-config-provider
  restart: unless-stopped
  volumes:
    - ./certs/ca.crt:/etc/ssl/certs/ca-certificates.crt:ro
  environment:
    - GITEA_URL=https://git.local.lan
    - GITEA_TOKEN=42ae0e5b0238ddb3ec80f7ebd208e829e0d37d5a
    - TEMPLATE_REPO=dronefiles
    - TEMPLATE_OWNER=admin
  networks:
    - traefik
```

### 4. 更新 Woodpecker Server 配置

```yaml
woodpecker-server:
  environment:
    - WOODPECKER_CONFIG_SERVICE_ENDPOINT=http://woodpecker-config-provider:8000
```

## 使用方式

### 1. 创建 dronefiles 仓库

在 Gitea 中创建 `dronefiles` 仓库（私有）。

### 2. 添加配置文件

按照以下结构组织配置：

```
dronefiles/
├── myrepo/
│   ├── main/
│   │   ├── build.yml
│   │   ├── test.yml
│   │   └── deploy.yml
│   └── develop/
│       ├── build.yml
│       └── test.yml
└── another-repo/
    └── main/
        └── build.yml
```

### 3. 推送代码触发构建

推送代码到 `myrepo` 的 `main` 分支，Woodpecker 会：
1. 自动从 `dronefiles/myrepo/main/` 读取所有 `.yml` 文件
2. 创建多个 pipeline（build、test、deploy）
3. 在 UI 中分别显示每个 pipeline 的执行结果

## API 响应格式

### 请求

```json
POST /ciconfig
{
  "repo": {
    "name": "myrepo",
    "namespace": "admin",
    "branch": "main"
  },
  "build": {
    "branch": "main"
  }
}
```

### 响应（多文件）

```json
{
  "configs": [
    {
      "name": "build",
      "data": "steps:\n  - name: build\n    image: golang:1.21\n    commands:\n      - go build -o app ."
    },
    {
      "name": "test",
      "data": "steps:\n  - name: test\n    image: golang:1.21\n    commands:\n      - go test -v ./..."
    },
    {
      "name": "deploy",
      "data": "when:\n  branch: main\nsteps:\n  - name: deploy\n    image: alpine\n    commands:\n      - echo 'Deploying...'"
    }
  ]
}
```

## 优势

### vs 单文件配置

| 特性 | 单文件 | 多文件 |
|------|--------|--------|
| **组织性** | ❌ 所有步骤在一个文件 | ✅ 按功能分离 |
| **可维护性** | ❌ 文件可能很长 | ✅ 每个文件独立 |
| **复用性** | ❌ 难以复用 | ✅ 可以共享通用配置 |
| **并行执行** | ⚠️ 需要手动配置 | ✅ 自动并行 |
| **UI 展示** | ❌ 单个 pipeline | ✅ 多个 pipeline 分别显示 |

### vs Drone config-plugin

| 特性 | Drone | Woodpecker (多文件) |
|------|-------|---------------------|
| **配置方式** | 单个 .drone.yml | 多个 .yml 文件 |
| **模板变量** | 有限支持 | ✅ 完整支持 |
| **并行执行** | ⚠️ 需要配置 | ✅ 自动并行 |
| **UI 展示** | 单个 pipeline | ✅ 多个 pipeline |

## 故障排查

### 问题 1：无法读取配置

**检查：**
```bash
# 测试 Gitea API
curl -k -H "Authorization: token YOUR_TOKEN" \
  "https://git.local.lan/api/v1/repos/admin/dronefiles/contents/myrepo/main"
```

### 问题 2：配置未生效

**检查日志：**
```bash
docker-compose logs woodpecker-config-provider
```

### 问题 3：ARM64 构建失败

**使用 buildx：**
```bash
docker buildx create --use
docker buildx build --platform linux/arm64 -t config-provider:arm64 .
```

## 开发和测试

### 本地运行

```bash
# 设置环境变量
export GITEA_URL="https://git.local.lan"
export GITEA_TOKEN="your-token"

# 运行
go run main.go
```

### 测试 API

```bash
curl -X POST http://localhost:8000/ciconfig \
  -H "Content-Type: application/json" \
  -d '{
    "repo": {"name": "myrepo", "namespace": "admin", "branch": "main"},
    "build": {"branch": "main"}
  }'
```

## 下一步

1. ✅ 构建 ARM64 镜像
2. ✅ 配置环境变量
3. ✅ 创建 dronefiles 仓库
4. ✅ 添加多个配置文件
5. ✅ 测试多 pipeline 执行

## 参考

- [Woodpecker Config Extensions](https://woodpecker-ci.org/docs/administration/advanced/config-extensions)
- [Gitea API 文档](https://docs.gitea.io/en-us/api-usage/)
