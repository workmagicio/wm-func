#!/bin/bash

# 前端启动脚本 - 用于后端debug模式开发
# 只启动前端，后端通过IDE debug启动

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRONTEND_DIR="$SCRIPT_DIR/frontend"

echo "🚀 启动前端开发服务器..."

# 进入frontend目录
cd "$FRONTEND_DIR"

# 检查依赖
if [ ! -d "node_modules" ]; then
    echo "📦 安装依赖包..."
    npm install
fi

echo ""
echo "🎯 启动前端开发服务器..."
echo "💡 前端服务器: http://localhost:5173"
echo "🔧 后端请手动通过IDE debug启动"
echo "❌ 使用 Ctrl+C 停止前端服务器"
echo ""

# 启动前端开发服务器
npm run dev
