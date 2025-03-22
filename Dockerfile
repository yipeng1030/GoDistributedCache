# Stage 1: 构建二进制文件，目标平台为 linux/arm64
FROM --platform=linux/arm64 golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o mycache main/main.go

# Stage 2: 构建运行镜像，确保使用 linux/arm64 平台
FROM --platform=linux/arm64 alpine:latest
WORKDIR /app
COPY --from=builder /app/mycache .
RUN chmod +x mycache
EXPOSE 8001
CMD ["./mycache"]