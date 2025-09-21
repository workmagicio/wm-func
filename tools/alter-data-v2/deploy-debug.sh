#!/bin/bash
set -e  # 遇到错误立即退出

# 配置变量
INSTANCE_NAME="lan-env"
ZONE="us-west2-b"
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

echo "🐳 传输调试版Docker Compose配置..."
gcloud compute scp --zone=${ZONE} ${PROJECT_ROOT}/docker-compose.debug.yml ${INSTANCE_NAME}:${REMOTE_DIR}/docker-compose.yml

echo ""
echo "✅ 传输完成！"

echo ""
echo "🚀 启动新服务（调试模式）..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="cd ${REMOTE_DIR} && docker compose up -d"

echo ""
echo "📊 查看服务状态..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="cd ${REMOTE_DIR} && docker compose ps"

echo ""
echo "📋 查看API服务日志（最近50行）..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="cd ${REMOTE_DIR} && docker compose logs --tail=50 api"

echo ""
echo "✅ 调试模式部署完成！"

# 获取实例内部IP
INTERNAL_IP=$(gcloud compute instances describe ${INSTANCE_NAME} --zone=${ZONE} --format='get(networkInterfaces[0].networkIP)')

echo "🌐 访问地址 (内部IP):"
echo "   应用: http://${INTERNAL_IP} (前端+API)"
echo "   调试: http://${INTERNAL_IP}:8081 (备用端口)"
echo "   API:  http://${INTERNAL_IP}/api/"
echo ""
echo "💡 调试用命令："
echo "   查看API日志: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose logs -f api'"
echo "   查看所有日志: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose logs -f'"
echo "   连接实例: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE}"
echo "   停止服务: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose down'"
echo "   重启服务: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose restart'"
echo "   进入API容器: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose exec api sh'"
