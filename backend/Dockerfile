# 使用官方 Go 镜像作为构建阶段
FROM golang:1.24.4-alpine AS builder

# 安装必要的系统依赖
RUN apk add --no-cache git gcc musl-dev

# 设置工作目录
WORKDIR /app

# 克隆代码
RUN git clone https://github.com/ccj241/binance.git .

# 进入后端目录
WORKDIR /app/backend

# 下载依赖
RUN go mod download

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# 构建 create-admin 工具
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o create-admin ./cmd/create-admin/main.go || echo "create-admin build skipped"

# 使用轻量级的 Alpine 镜像作为运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 创建非 root 用户
RUN addgroup -g 1000 -S appuser && \
    adduser -u 1000 -S appuser -G appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/backend/main .
COPY --from=builder /app/backend/create-admin ./create-admin || echo "create-admin not found"

# 更改文件所有权
RUN chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 23337

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:23337/health || exit 1

# 运行应用
CMD ["./main"]