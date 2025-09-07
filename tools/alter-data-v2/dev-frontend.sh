#!/bin/bash

# 前端开发模式启动脚本
# 用于快速启动 alter-data-v2 前端开发服务器

set -e  # 遇到错误时退出

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRONTEND_DIR="$SCRIPT_DIR/frontend"

echo "🚀 启动前端开发服务器..."

# 检查frontend目录是否存在
if [ ! -d "$FRONTEND_DIR" ]; then
    echo "❌ 错误: frontend 目录不存在 ($FRONTEND_DIR)"
    exit 1
fi

# 进入frontend目录
cd "$FRONTEND_DIR"
echo "📁 当前目录: $(pwd)"

# 检查是否有package.json
if [ ! -f "package.json" ]; then
    echo "❌ 错误: package.json 文件不存在"
    exit 1
fi

# 设置代理（根据用户的记忆配置）
echo "🔧 配置网络代理..."
if command -v xp >/dev/null 2>&1; then
    echo "使用 xp 命令设置代理..."
    xp
else
    echo "⚠️  警告: xp 命令不可用，跳过代理设置"
fi

# 检查并安装依赖
if [ ! -d "node_modules" ] || [ "package.json" -nt "node_modules" ]; then
    echo "📦 安装/更新依赖包..."
    if command -v yarn >/dev/null 2>&1; then
        echo "使用 yarn 安装依赖..."
        yarn install
    else
        echo "使用 npm 安装依赖..."
        npm install
    fi
else
    echo "✅ 依赖包已是最新"
fi

echo ""
echo "🎯 启动开发服务器..."
echo "💡 开发服务器通常会运行在: http://localhost:5173"
echo "🔄 修改代码后会自动热更新"
echo "❌ 使用 Ctrl+C 停止服务器"
echo ""

# 启动开发服务器
if command -v yarn >/dev/null 2>&1; then
    yarn dev
else
    npm run dev
fi
