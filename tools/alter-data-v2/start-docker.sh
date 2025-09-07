#!/bin/bash

# 启动脚本 - Docker方式运行alter-data-v2服务

set -e

echo "🚀 启动 alter-data-v2 服务..."

# 创建必要目录
mkdir -p data/redis
mkdir -p logs

# 先进行交叉编译
echo "🔨 交叉编译应用..."
./build.sh

# 构建并启动服务
echo "📦 构建Docker镜像..."
docker compose build --no-cache

echo "🐳 启动服务..."
docker compose up -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 健康检查
echo "🔍 检查服务状态..."
if curl -s http://localhost:8081/health > /dev/null; then
    echo "✅ API服务启动成功"
else
    echo "❌ API服务启动失败"
    docker compose logs api
    exit 1
fi

if curl -s http://localhost/health > /dev/null; then
    echo "✅ Nginx服务启动成功"
else
    echo "⚠️  Nginx服务可能有问题，但API直接访问正常"
fi

# 显示服务信息
echo ""
echo "🎉 服务启动成功！"
echo "📊 API地址: http://localhost:8081"
echo "🌐 前端地址: http://localhost"
echo "📈 Redis地址: localhost:6379"
echo ""
echo "💡 常用命令:"
echo "  查看日志: docker compose logs -f [service_name]"
echo "  停止服务: docker compose down"
echo "  重启服务: docker compose restart"
echo "  进入容器: docker compose exec [service_name] sh"
