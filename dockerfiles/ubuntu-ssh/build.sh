#!/bin/bash
set -e

# 镜像名称和标签
IMAGE_NAME="dws/ubuntu-ssh"
IMAGE_TAG="${1:-24.04}"
FULL_IMAGE="${IMAGE_NAME}:${IMAGE_TAG}"

echo "Building ${FULL_IMAGE}..."

# 构建镜像
docker build -t "${FULL_IMAGE}" .

echo "=========================================="
echo "Build completed: ${FULL_IMAGE}"
echo ""
echo "Usage:"
echo "  docker run -d -p 2222:22 ${FULL_IMAGE}"
echo "  docker run -d -p 2222:22 -e SSH_PASSWORD=your_password ${FULL_IMAGE}"
echo ""
echo "Connect:"
echo "  ssh root@localhost -p 2222"
echo "  Default password: dws"
echo "=========================================="

# 可选：推送到本地 registry（如果需要）
if [ "$2" = "push" ]; then
    echo "Pushing to registry..."
    docker push "${FULL_IMAGE}"
fi

