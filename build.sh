#!/bin/bash

# Alpine Linux编译脚本 - Go语言刷题系统
# 适配 arm64 和 amd64 架构，并正确处理 sqlite3 在 Alpine (musl) 上的编译
set -e  # 出错时立即退出

echo "========================================"
echo "Building Quiz System for Alpine Linux"
echo "========================================"

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "Go version: $GO_VERSION"

# 检查并安装编译依赖 (gcc 和 musl-dev 是 CGO 必需的)
echo "Checking build dependencies..."
apk add --no-cache build-base git

# 清理之前的构建
echo "Cleaning previous builds..."
rm -f quiz-system
rm -f quiz.db # 如果需要的话

# ================= 关键修改点 1: 架构设置 =================
# 移除硬编码的 GOARCH=amd64，让 Go 自动检测
unset GOARCH
# go env GOARCH 会显示自动检测到的架构
export CGO_ENABLED=1
export GOOS=linux

echo "Build configuration:"
echo "  CGO_ENABLED: $CGO_ENABLED"
echo "  GOOS: $GOOS"
echo "  GOARCH: $(go env GOARCH)" # 显示自动检测的架构

# 下载和验证依赖
echo "Downloading dependencies..."
go mod tidy
go mod download
echo "Verifying dependencies..."
go mod verify

# ================= 关键修改点 2: 编译命令 =================
# 编译应用程序，使用 libsqlite3 标签来适配 Alpine/musl
echo "Building application..."
echo "This may take a few minutes..."
go build -v \
    -tags="libsqlite3" \
    -ldflags="-w -s" \
    -o quiz-system \
    main.go

# 检查编译结果
if [ ! -f "quiz-system" ]; then
    echo "Error: Build failed - executable not found" >&2
    exit 1
fi

# 显示文件信息
echo "========================================"
echo "Build completed successfully!"
echo "========================================"

echo "Executable information:"
ls -lh quiz-system
echo ""

echo "File type:"
# 'file' 命令可能不存在，先检查
if command -v file &> /dev/null; then
    file quiz-system
fi
echo ""

echo "Executable size: $(du -h quiz-system | cut -f1)"
echo "Dependencies:"
# 在 Alpine 中 ldd 是 /lib/ld-musl-*.so.1，直接调用可执行文件并加上 ldd 选项
ldd ./quiz-system 2>/dev/null || echo "  Static binary (no dynamic dependencies)"
echo ""

# ... (后续的部署说明部分保持不变) ...

echo "========================================"
echo "Deployment Instructions:"
echo "========================================"
echo "1. Copy files to target server:"
echo "   - quiz-system (executable)"
echo "   - questions.json (question database)"
echo ""
echo "2. Run the application:"
echo "   ./quiz-system"
echo ""
echo "3. Access the system:"
echo "   http://localhost:50442"
echo "========================================"

echo "Build completed at: $(date)"
echo "Ready for production deployment!"

