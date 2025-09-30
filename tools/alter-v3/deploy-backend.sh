#!/bin/bash
set -e  # 遇到错误立即退出

# 配置变量
INSTANCE_NAME="lan-env"
ZONE="us-west2-b"
REMOTE_DIR="/home/$(whoami)/alter-v3"
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"

echo "🔧 只部署后端服务..."
echo ""

echo "🧹 清理本地后端构建文件..."
rm -rf ${PROJECT_ROOT}/bin 2>/dev/null || true
echo "✅ 已清理本地 bin 目录"

echo ""
echo "🔧 交叉编译 Go 后端..."
cd ${PROJECT_ROOT}
if [ -f main.go ]; then
    mkdir -p bin
    GOOS=linux GOARCH=amd64 go build -o bin/alter-v3 main.go
    if [ -f bin/alter-v3 ]; then
        echo "✅ Go 后端编译完成"
    else
        echo "❌ Go 后端编译失败"
        exit 1
    fi
else
    echo "❌ main.go 不存在"
    exit 1
fi

echo ""
echo "🛑 停止远端后端容器..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="
    cd ${REMOTE_DIR}
    if [ -f docker-compose.yml ]; then
        docker compose stop alter-v3-backend 2>/dev/null || true
    fi
" || echo "⚠️ 后端容器可能未运行，继续部署..."

echo ""
echo "🧹 清理远端后端文件..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="
    cd ${REMOTE_DIR} 
    rm -rf bin 2>/dev/null || true
    echo '✅ 已清理远端后端目录'
"

echo ""
echo "📤 上传后端文件到服务器..."
# 确保远端目录存在
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="mkdir -p ${REMOTE_DIR}"

# 上传二进制文件
if [ -d ${PROJECT_ROOT}/bin ]; then
    gcloud compute scp --recurse ${PROJECT_ROOT}/bin ${INSTANCE_NAME}:${REMOTE_DIR}/ --zone=${ZONE}
    echo "✅ 后端文件上传完成"
else
    echo "❌ bin 目录不存在"
    exit 1
fi

## 上传配置文件（如果有更新）
#if [ -f ${PROJECT_ROOT}/config.json ]; then
#    gcloud compute scp ${PROJECT_ROOT}/config.json ${INSTANCE_NAME}:${REMOTE_DIR}/ --zone=${ZONE}
#    echo "✅ 配置文件上传完成"
#fi

echo ""
echo "🚀 启动后端容器..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="
    cd ${REMOTE_DIR}
    docker compose up -d alter-v3-backend
"
echo "✅ 后端容器启动完成"

echo ""
echo "🎉 后端部署完成！"
echo "🔌 API地址: http://10.168.0.10:8081/api/config"
echo "📋 查看后端容器状态: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose ps alter-v3-backend'"
echo "📋 查看后端日志: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose logs -f alter-v3-backend'"
