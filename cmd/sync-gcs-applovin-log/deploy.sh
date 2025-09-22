#!/bin/bash
set -e  # 遇到错误立即退出

# 配置变量
INSTANCE_NAME="lan-env"
ZONE="us-west2-b"
REMOTE_DIR="/home/$(whoami)/sync-gcs-applovin-log"
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"

echo ""
echo "🔨 构建后端二进制文件..."
GOOS=linux GOARCH=amd64 go build .


echo "🛑 停止远端服务..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="cd ${REMOTE_DIR} && docker compose down" || echo "⚠️ 服务可能未运行，继续部署..."

echo ""
echo "📦 创建远程目录结构..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="mkdir -p ${REMOTE_DIR}/data"

echo ""
echo "🚀 传输二进制文件..."
gcloud compute scp --zone=${ZONE} ${PROJECT_ROOT}/sync-gcs-applovin-log ${INSTANCE_NAME}:${REMOTE_DIR}

echo "🐳 传输Docker Compose配置..."
gcloud compute scp --zone=${ZONE} ${PROJECT_ROOT}/docker-compose.yml ${INSTANCE_NAME}:${REMOTE_DIR}/docker-compose.yml

echo ""
echo "✅ 传输完成！"

echo ""
echo "🚀 启动新服务..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="cd ${REMOTE_DIR} && docker compose up -d"

echo ""
echo "📊 查看服务状态..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="cd ${REMOTE_DIR} && docker compose ps"

echo ""
echo "✅ 部署完成！"