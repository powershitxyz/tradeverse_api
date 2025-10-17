#!/bin/sh
set -e

# 配置参数
PROJECT_DIR="/data/chaos/source/api"         # 代码所在目录
OUTPUT_DIR="/data/chaos/api-server"           # 编译输出目录
BINARY_NAME="chaos-api"               # 生成的二进制文件名称

echo "switch direction：$PROJECT_DIR"
cd "$PROJECT_DIR"

echo "pull code for update..."
git fetch origin
git reset --hard origin/main
git clean -fd

echo "start compiling..."
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

go build -o "$OUTPUT_DIR/$BINARY_NAME" main.go
echo "compiled successfully and moved app to $OUTPUT_DIR/$BINARY_NAME"