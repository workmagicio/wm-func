#!/bin/bash

# 停止脚本 - 停止alter-data-v2服务

set -e

echo "🛑 停止 alter-data-v2 服务..."

cd "$(dirname "$0")"

# 停止并删除容器
echo "📦 停止容器..."
docker compose down

echo "🧹 清理未使用的资源..."
docker system prune -f

echo ""
echo "✅ 服务已停止！"
echo ""
echo "💡 如需重新启动，运行: ./start-docker.sh"
