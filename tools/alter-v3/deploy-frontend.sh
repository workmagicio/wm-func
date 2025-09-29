#!/bin/bash
set -e  # 遇到错误立即退出

# 配置变量
INSTANCE_NAME="lan-env"
ZONE="us-west2-b"
REMOTE_DIR="/home/$(whoami)/alter-v3"
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"

echo "🎨 只部署前端服务..."
echo ""

echo "🧹 清理本地前端构建文件..."
rm -rf ${PROJECT_ROOT}/dist 2>/dev/null || true
echo "✅ 已清理本地 dist 目录"

echo ""
echo "🎨 编译前端..."
if [ -d ${PROJECT_ROOT}/frontend ]; then
    cd ${PROJECT_ROOT}/frontend
    if [ -f package.json ]; then
        npm run build
        echo "✅ 前端编译完成"
        
        echo ""
        echo "📦 移动前端构建文件..."
        cd ${PROJECT_ROOT}
        if [ -d frontend/build ]; then
            mv frontend/build dist
            echo "✅ 前端文件已移动到 dist 目录"
        else
            echo "❌ frontend/build 目录不存在，编译可能失败"
            exit 1
        fi
    else
        echo "❌ frontend/package.json 不存在"
        exit 1
    fi
else
    echo "❌ frontend 目录不存在"
    exit 1
fi

echo ""
echo "🛑 停止远端前端容器..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="
    cd ${REMOTE_DIR}
    if [ -f docker-compose.yml ]; then
        docker compose stop alter-v3 2>/dev/null || true
    fi
" || echo "⚠️ 前端容器可能未运行，继续部署..."

echo ""
echo "🧹 清理远端前端文件..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="
    cd ${REMOTE_DIR} 
    rm -rf dist 2>/dev/null || true
    echo '✅ 已清理远端前端目录'
"

echo ""
echo "📤 上传前端文件到服务器..."
# 确保远端目录存在
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="mkdir -p ${REMOTE_DIR}"

# 上传前端文件
if [ -d ${PROJECT_ROOT}/dist ]; then
    gcloud compute scp --recurse ${PROJECT_ROOT}/dist ${INSTANCE_NAME}:${REMOTE_DIR}/ --zone=${ZONE}
    echo "✅ 前端文件上传完成"
else
    echo "❌ dist 目录不存在"
    exit 1
fi

echo ""
echo "🚀 启动前端容器..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="
    cd ${REMOTE_DIR}
    docker compose up -d alter-v3
"
echo "✅ 前端容器启动完成"

echo ""
echo "🎉 前端部署完成！"
echo "🌐 访问地址: http://10.168.0.10:8090/"
echo "📋 查看前端容器状态: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose ps alter-v3'"
echo "📋 查看前端日志: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose logs -f alter-v3'"
