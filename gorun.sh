#!/bin/bash

# 通用 Go 编译运行脚本
# 支持从子目录编译运行

set -e

# 获取项目根目录（脚本所在目录）
ROOT_DIR=$(dirname "$(readlink -f "$0")")

# 默认构建当前目录
BUILD_DIR="./cmd"
if [ $# -gt 0 ]; then
    BUILD_DIR="$1"
    shift
fi

# 切换到项目根目录
cd "$ROOT_DIR"

# 检查并初始化 go.mod
if [ ! -f "go.mod" ]; then
    echo "⚠️  go.mod not found, initializing new module..."
    MODULE_NAME=$(basename "$(pwd)")
    go mod init "$MODULE_NAME"
fi

# 同步依赖
echo "📦 Synchronizing dependencies..."
go mod tidy

# 获取模块名
MODULE_NAME=$(go list -m)
EXEC_NAME=$(basename "$MODULE_NAME")

# 创建输出目录
BIN_DIR="$ROOT_DIR/bin"
mkdir -p "$BIN_DIR"

# 构建项目
echo "🛠  Building $EXEC_NAME from $BUILD_DIR..."
# 构建项目后立即设置执行权限
go build -o "$BIN_DIR/$EXEC_NAME" "./$BUILD_DIR" && chmod +x "$BIN_DIR/$EXEC_NAME"

# 检查构建结果
if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

# 运行程序
echo "🚀 Running $EXEC_NAME..."
"$BIN_DIR/$EXEC_NAME" "$@"