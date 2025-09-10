#!/bin/bash
set -e # 任何命令失败则立即退出

# --- 配置区 ---
PROJECT_ID="glass-ranger-446609-p9"
REGION="us-east1"
SERVICE_NAME="knocommerce-public-service" # 服务名称，可自定义

IMAGE_TAG=$(date +%Y%m%d-%H%M%S)
FULL_IMAGE_NAME="us-east1-docker.pkg.dev/${PROJECT_ID}/cloud-run-source-deploy/${SERVICE_NAME}:${IMAGE_TAG}"

# --- 脚本开始 ---
echo "STEP 1: 本地交叉编译 Go 程序..."
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
PROJECT_ROOT=$(cd "${SCRIPT_DIR}/../.." && pwd)
(cd "${PROJECT_ROOT}" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -v -o "${SCRIPT_DIR}/server" ./cmd/knocommerce_first_sync)
echo "编译完成, 可执行文件 'server' 已生成。"

echo "STEP 2: 本地构建 Docker 镜像 (强制为 linux/amd64)..."
docker build --platform linux/amd64 -t "${FULL_IMAGE_NAME}" -f Dockerfile .
echo "镜像构建完成: ${FULL_IMAGE_NAME}"

echo "STEP 3: 推送镜像到 Google Artifact Registry..."
docker push "${FULL_IMAGE_NAME}"
echo "镜像推送完成。"

echo "STEP 4: 部署镜像为可公开访问的 Cloud Run Service..."
gcloud run deploy "${SERVICE_NAME}" \
    --image "${FULL_IMAGE_NAME}" \
    --region "${REGION}" \
    --project "${PROJECT_ID}" \
    --platform "managed" \
    --vpc-connector "projects/${PROJECT_ID}/locations/${REGION}/connectors/default" \
    --vpc-egress "private-ranges-only" \
#    --allow-unauthenticated

echo "STEP 5: 清理本地编译产物和镜像..."
rm server
docker rmi "${FULL_IMAGE_NAME}"
echo "本地清理完成。"

echo "✅ 公开服务部署成功！"
echo "服务 URL (可从互联网直接访问):"
# 部署成功后，gcloud 会输出服务的 URL