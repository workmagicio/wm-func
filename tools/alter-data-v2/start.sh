#!/bin/bash

echo "🚀 启动 Alter Data V2 服务..."

# 检查是否存在构建的前端文件
if [ ! -d "dist" ] || [ ! -f "dist/index.html" ]; then
    echo "⚠️  未找到前端构建文件，开始构建..."
    ./build.sh
    if [ $? -ne 0 ]; then
        echo "❌ 构建失败"
        exit 1
    fi
fi

# 获取端口参数，默认8080
PORT=${1:-8080}

echo "🌐 启动服务器在端口 $PORT..."
echo "📱 前端应用: http://localhost:$PORT"
echo "🔌 API接口: http://localhost:$PORT/api/alter-data"
echo "💊 健康检查: http://localhost:$PORT/health"
echo ""
echo "按 Ctrl+C 停止服务器"
echo ""

# 启动Go服务器
go run main.go -port=$PORT
