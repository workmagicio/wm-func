#!/bin/bash
set -e  # 遇到错误立即退出

# 配置变量
INSTANCE_NAME="lan-env"
ZONE="us-west2-b"
REMOTE_DIR="/home/$(whoami)/alter-data-v2"


echo ""
echo "🔨 构建后端二进制文件..."
GOOS=linux GOARCH=amd64 go build .
mv alter-data-v2 app

echo "🛑 停止远端服务..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="cd ${REMOTE_DIR} && docker compose down" || echo "⚠️ 服务可能未运行，继续部署..."



echo ""
echo "🚀 传输二进制文件..."
gcloud compute scp --zone=${ZONE} app ${INSTANCE_NAME}:${REMOTE_DIR}/bin/


echo ""
echo "🚀 启动新服务..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="cd ${REMOTE_DIR} && docker compose up -d"

