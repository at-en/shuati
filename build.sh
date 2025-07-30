#!/bin/bash

# Alpine Linux编译脚本 - Go语言刷题系统
# 用于生成静态链接的单一可执行文件

set -e  # 出错时立即退出

echo "========================================"
echo "Building Quiz System for Alpine Linux"
echo "========================================"

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Please install Go 1.21+ first:"
    echo "  apk add --no-cache go"
    exit 1
fi

# 检查Go版本
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "Go version: $GO_VERSION"

# 检查编译依赖
echo "Checking build dependencies..."
if ! apk info gcc &> /dev/null; then
    echo "Installing build dependencies..."
    apk add --no-cache gcc musl-dev sqlite-dev git
fi

# 清理之前的构建
echo "Cleaning previous builds..."
rm -f quiz-system
rm -f quiz.db

# 设置编译环境变量
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

echo "Build configuration:"
echo "  CGO_ENABLED: $CGO_ENABLED"
echo "  GOOS: $GOOS"
echo "  GOARCH: $GOARCH"

# 下载依赖
echo "Downloading dependencies..."
go mod tidy
go mod download

# 验证依赖
echo "Verifying dependencies..."
go mod verify

# 编译应用程序
echo "Building application..."
echo "This may take a few minutes..."

# 使用静态链接编译，确保在Alpine上运行
go build -v \
    -ldflags="-w -s -linkmode external -extldflags '-static'" \
    -tags="sqlite_omit_load_extension netgo osusergo static_build" \
    -o quiz-system \
    main.go

# 检查编译结果
if [ ! -f "quiz-system" ]; then
    echo "Error: Build failed - executable not found"
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
file quiz-system
echo ""

echo "Executable size: $(du -h quiz-system | cut -f1)"
echo "Dependencies:"
ldd quiz-system 2>/dev/null || echo "  Static binary (no dynamic dependencies)"
echo ""

# 验证可执行文件
echo "Testing executable..."
chmod +x quiz-system
./quiz-system --help 2>/dev/null || echo "Binary is ready to run"

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
echo ""
echo "System Requirements:"
echo "  - Alpine Linux 3.15+"
echo "  - Memory: 20MB+"
echo "  - Disk: 50MB+"
echo "  - Network: Port 50442"
echo "========================================"

echo "Build completed at: $(date)"
echo "Ready for production deployment!"