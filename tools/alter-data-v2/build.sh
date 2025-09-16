#!/bin/bash

# 交叉编译脚本 - 编译Linux二进制文件

set -e

echo "🔨 开始交叉编译..."

# 进入项目根目录（当前目录就是项目根目录）
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_ROOT"

# 设置编译环境
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

# 创建输出目录
mkdir -p bin

# 编译应用
echo "📦 编译 alter-data-v2..."
go build -a -installsuffix cgo -ldflags '-w -s' -o bin/app main.go

# 检查编译结果
if [ -f "bin/app" ]; then
    echo "✅ 编译成功!"
    
    # 显示文件信息
    echo "📊 二进制文件信息:"
    ls -lh bin/app
    file bin/app
else
    echo "❌ 编译失败!"
    exit 1
fi

echo "🎉 交叉编译完成!"