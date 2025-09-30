#!/bin/bash
set -e  # 遇到错误立即退出

# 配置变量
INSTANCE_NAME="lan-env"
ZONE="us-west2-b"
REMOTE_DIR="/home/$(whoami)/alter-v3"
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"

echo "🛑 停止远端服务..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="
    mkdir -p ${REMOTE_DIR}
    cd ${REMOTE_DIR}
    if [ -f docker-compose.yml ]; then
        docker compose down 2>/dev/null || true
    fi
" || echo "⚠️ 服务可能未运行，继续部署..."

echo ""
echo "🧹 清理远端旧文件..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="
    mkdir -p ${REMOTE_DIR}
    cd ${REMOTE_DIR}
    rm -rf dist bin 2>/dev/null || true
    echo '✅ 已清理远端目录'
"

echo ""
echo "🧹 清理本地构建文件..."
rm -rf ${PROJECT_ROOT}/dist ${PROJECT_ROOT}/bin 2>/dev/null || true
echo "✅ 已清理本地构建目录"

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
echo "📤 上传文件到服务器..."
# 确保远端目录存在
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="mkdir -p ${REMOTE_DIR}"

# 上传二进制文件
if [ -d ${PROJECT_ROOT}/bin ]; then
    gcloud compute scp --recurse ${PROJECT_ROOT}/bin ${INSTANCE_NAME}:${REMOTE_DIR}/ --zone=${ZONE}
    echo "✅ 二进制文件上传完成"
else
    echo "❌ bin 目录不存在"
    exit 1
fi

# 上传前端文件
if [ -d ${PROJECT_ROOT}/dist ]; then
    gcloud compute scp --recurse ${PROJECT_ROOT}/dist ${INSTANCE_NAME}:${REMOTE_DIR}/ --zone=${ZONE}
    echo "✅ 前端文件上传完成"
else
    echo "❌ dist 目录不存在"
    exit 1
fi

## 上传配置文件
#if [ -f ${PROJECT_ROOT}/config.json ]; then
#    gcloud compute scp ${PROJECT_ROOT}/config.json ${INSTANCE_NAME}:${REMOTE_DIR}/ --zone=${ZONE}
#    echo "✅ 配置文件上传完成"
#else
#    echo "❌ config.json 不存在"
#    exit 1
#fi

# 上传 docker-compose 文件
if [ -f ${PROJECT_ROOT}/docker-compose.yml ]; then
    gcloud compute scp ${PROJECT_ROOT}/docker-compose.yml ${INSTANCE_NAME}:${REMOTE_DIR}/ --zone=${ZONE}
    echo "✅ docker-compose.yml 上传完成"
else
    echo "❌ docker-compose.yml 不存在"
    exit 1
fi

# 上传 nginx 配置文件
if [ -f ${PROJECT_ROOT}/nginx.conf ]; then
    gcloud compute scp ${PROJECT_ROOT}/nginx.conf ${INSTANCE_NAME}:${REMOTE_DIR}/ --zone=${ZONE}
    echo "✅ nginx.conf 上传完成"
else
    echo "⚠️ nginx.conf 不存在，将使用默认配置"
fi

echo ""
echo "🚀 启动服务..."
gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command="cd ${REMOTE_DIR} && docker compose up -d"
echo "✅ 服务启动完成"

echo ""
echo "🎉 部署完成！"
echo "📋 查看服务状态: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose ps'"
echo "📋 查看日志: gcloud compute ssh ${INSTANCE_NAME} --zone=${ZONE} --command='cd ${REMOTE_DIR} && docker compose logs -f'"
