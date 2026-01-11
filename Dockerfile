# 多阶段构建，支持 ARM64
FROM golang:1.24-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH=arm64

WORKDIR /app

# 复制依赖文件
COPY go.mod ./
# 如果 go.sum 存在则复制，不存在则跳过
COPY go.sum ./

# 下载依赖（如果有的话）
RUN go mod download || true

# 复制源代码
COPY . .

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-w -s" -o woodpecker-config-provider .

# 最终镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/woodpecker-config-provider .

EXPOSE 8000

CMD ["./woodpecker-config-provider"]
