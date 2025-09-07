#!/bin/bash
set -e  # 遇到错误立即退出

echo "🛑 停止远端服务..."
ssh -i ~/.ssh/ali-us-va-default-key.pem ecs-user@10.10.2.238 "cd /home/ecs-user/alter-data && docker compose down" || echo "⚠️ 服务可能未运行，继续部署..."

echo ""
echo "📦 构建Linux二进制文件..."
GOOS=linux GOARCH=amd64 go build .

echo ""
echo "🚀 传输二进制文件..."
scp -i ~/.ssh/ali-us-va-default-key.pem ./alter-data ecs-user@10.10.2.238:/home/ecs-user/alter-data/

echo "🎨 传输静态文件..."
scp -r -i ~/.ssh/ali-us-va-default-key.pem ./static ecs-user@10.10.2.238:/home/ecs-user/alter-data/

echo "🐳 传输Docker Compose配置..."
scp -i ~/.ssh/ali-us-va-default-key.pem ./docker compose.yaml ecs-user@10.10.2.238:/home/ecs-user/alter-data/

echo ""
echo "✅ 传输完成！"

echo "🚀 启动新服务..."
ssh -i ~/.ssh/ali-us-va-default-key.pem ecs-user@10.10.2.238 "cd /home/ecs-user/alter-data && docker compose up -d"

echo "📊 查看服务状态..."
ssh -i ~/.ssh/ali-us-va-default-key.pem ecs-user@10.10.2.238 "cd /home/ecs-user/alter-data && docker compose ps"

echo ""
echo "✅ 部署完成！"
echo "🌐 访问地址: http://10.10.2.238:8090"
echo ""
echo "💡 查看日志命令："
echo "   ssh -i ~/.ssh/ali-us-va-default-key.pem ecs-user@10.10.2.238 'cd /home/ecs-user/alter-data && docker compose logs -f'"
