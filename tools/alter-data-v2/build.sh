#!/bin/bash

# 交叉编译脚本 - 编译Linux二进制文件

set -e

echo "🔨 开始交叉编译..."

# 进入项目根目录
cd "$(dirname "$0")/../../"

# 设置编译环境
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

# 创建输出目录
mkdir -p tools/alter-data-v2/bin

# 编译应用
echo "📦 编译 alter-data-v2..."
go build -a -installsuffix cgo -ldflags '-w -s' -o tools/alter-data-v2/bin/app tools/alter-data-v2/main.go

# 检查编译结果
if [ -f "tools/alter-data-v2/bin/app" ]; then
    echo "✅ 编译成功!"
    
    # 显示文件信息
    echo "📊 二进制文件信息:"
    ls -lh tools/alter-data-v2/bin/app
    file tools/alter-data-v2/bin/app
else
    echo "❌ 编译失败!"
    exit 1
fi

echo "🎉 交叉编译完成!"