#!/bin/bash

echo "🚀 开始构建 Alter Data V2 应用..."

# 检查是否在正确的目录
if [ ! -f "main.go" ]; then
    echo "❌ 错误: 请在 alter-data-v2 目录下运行此脚本"
    exit 1
fi

echo "📦 构建前端应用..."

# 进入前端目录
cd frontend

# 检查是否已安装依赖
if [ ! -d "node_modules" ]; then
    echo "📥 安装前端依赖..."
    npm install
else
    echo "✅ 前端依赖已安装"
fi

# 构建前端
echo "🔨 构建前端代码..."
npm run build

if [ $? -ne 0 ]; then
    echo "❌ 前端构建失败"
    exit 1
fi

# 返回根目录
cd ..

echo "✅ 前端构建完成"
echo "📁 前端文件已输出到 dist/ 目录"

echo "🎉 构建完成!"
echo ""
echo "启动服务器请运行:"
echo "  go run main.go"
echo ""
echo "或者使用:"
echo "  ./start.sh"
