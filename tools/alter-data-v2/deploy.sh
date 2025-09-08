#!/bin/bash
set -e  # 遇到错误立即退出

# 配置变量
INSTANCE_NAME="alter-data-xk"
ZONE="us-east1-d"
REMOTE_DIR="/home/$(whoami)/alter-data-v2"
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"

echo "🛑 停止远端服务..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="cd ${REMOTE_DIR} && docker compose down" || echo "⚠️ 服务可能未运行，继续部署..."

echo ""
echo "🔨 构建后端二进制文件..."
bash ${PROJECT_ROOT}/build.sh

echo ""
echo "🎨 构建前端资源..."
bash ${PROJECT_ROOT}/build-frontend.sh

echo ""
echo "📦 创建远程目录结构..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="mkdir -p ${REMOTE_DIR}/{bin,data/redis,cache,logs,dist}"

echo ""
echo "🚀 传输二进制文件..."
gcloud compute scp --zone=${ZONE} ${PROJECT_ROOT}/bin/app ${INSTANCE_NAME}:${REMOTE_DIR}/bin/

echo "🎯 传输前端静态文件..."
gcloud compute scp --zone=${ZONE} --recurse ${PROJECT_ROOT}/dist ${INSTANCE_NAME}:${REMOTE_DIR}/


# 只有第一次需要传下面内容
#echo "🐳 传输Docker Compose配置..."
#gcloud compute scp --zone=${ZONE} ${PROJECT_ROOT}/docker-compose.prod.yml ${INSTANCE_NAME}:${REMOTE_DIR}/docker-compose.yml
#
#echo "⚙️ 传输Nginx配置..."
#gcloud compute scp --zone=${ZONE} ${PROJECT_ROOT}/nginx.conf ${INSTANCE_NAME}:${REMOTE_DIR}/
#
#echo "📋 传输Dockerfile..."
#gcloud compute scp --zone=${ZONE} ${PROJECT_ROOT}/Dockerfile.prod ${INSTANCE_NAME}:${REMOTE_DIR}/Dockerfile
#
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

# 获取实例内部IP
INTERNAL_IP=$(gcloud compute instances describe ${INSTANCE_NAME} --zone=${ZONE} --format='get(networkInterfaces[0].networkIP)')

echo "🌐 访问地址 (内部IP):"
echo "   前端: http://${INTERNAL_IP}"
echo "   API:  http://${INTERNAL_IP}:8081"
echo ""
echo "💡 有用的命令："
echo "   查看日志: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose logs -f'"
echo "   连接实例: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE}"
echo "   停止服务: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose down'"
echo "   重启服务: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose restart'"
