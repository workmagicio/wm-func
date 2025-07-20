#!/bin/bash
set -e # 任何命令失败则立即退出

# --- 配置区 ---
PROJECT_ID="glass-ranger-446609-p9"
REGION="us-east1"
JOB_NAME="shopify-customer-first-order"

IMAGE_TAG=$(date +%Y%m%d-%H%M%S)
FULL_IMAGE_NAME="us-east1-docker.pkg.dev/${PROJECT_ID}/cloud-run-source-deploy/${JOB_NAME}:${IMAGE_TAG}"

# --- 脚本开始 ---
echo "STEP 1: 本地交叉编译 Go 程序..."
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)
PROJECT_ROOT=$(cd "${SCRIPT_DIR}/../.." && pwd)
(cd "${PROJECT_ROOT}" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -v -o "${SCRIPT_DIR}/server" ./cmd/shopify-customer-first-order)
echo "编译完成, 可执行文件 'server' 已生成。"

echo "STEP 2: 本地构建 Docker 镜像 (强制为 linux/amd64)..."
docker build --platform linux/amd64 -t "${FULL_IMAGE_NAME}" -f Dockerfile .
echo "镜像构建完成: ${FULL_IMAGE_NAME}"

echo "STEP 3: 推送镜像到 Google Artifact Registry..."
docker push "${FULL_IMAGE_NAME}"

echo "STEP 4: 部署镜像..."
gcloud run jobs deploy "${JOB_NAME}" \
    --image "${FULL_IMAGE_NAME}" \
    --tasks 1 \
    --region "${REGION}" \
    --project "${PROJECT_ID}" \
    --network "default" \
    --subnet "default" \
    --vpc-egress "private-ranges-only"


# 【新增步骤】
echo "STEP 5: 立即执行 Job..."
gcloud run jobs execute "${JOB_NAME}" \
    --region "${REGION}" \


echo "STEP 6: 清理本地编译产物..."
rm server
docker rmi "${FULL_IMAGE_NAME}"
echo "本地清理完成。"

echo "✅ 部署并执行成功！"