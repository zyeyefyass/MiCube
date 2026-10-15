#!/bin/sh
# ==============================================================================
# MiCube 官方极速一键部署脚本 (国内免翻加速 / 飞牛 fnOS / PVE / Linux)
# 架构支持: AMD64 (x86_64) / ARM64 (AArch64)
# ==============================================================================
set -e

echo "🧊 正在检测系统环境与网络连通性..."

# 0. 智能检测国内网络与加速镜像源
GH_PREFIX=""
if ! curl -fsSL -m 3 https://api.github.com >/dev/null 2>&1; then
  echo "⚡ 检测到当前处于中国大陆网络或无法直接直连 GitHub，自动启用国内极速加速镜像通道..."
  # 探测当前最快的高速 GitHub 反代加速节点
  if curl -fsSL -m 3 https://ghfast.top >/dev/null 2>&1; then
    GH_PREFIX="https://ghfast.top/"
  elif curl -fsSL -m 3 https://mirror.ghproxy.com >/dev/null 2>&1; then
    GH_PREFIX="https://mirror.ghproxy.com/"
  elif curl -fsSL -m 3 https://ghproxy.cn >/dev/null 2>&1; then
    GH_PREFIX="https://ghproxy.cn/"
  else
    GH_PREFIX="https://ghfast.top/"
  fi
  echo "✅ 已锁定国内加速通道: ${GH_PREFIX}"
fi

# 1. 检测架构
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) MICUBE_ARCH="amd64" ;;
  aarch64|arm64) MICUBE_ARCH="arm64" ;;
  *) echo "❌ 暂不支持的架构: $ARCH"; exit 1 ;;
esac

echo "✅ 检测到系统架构: $MICUBE_ARCH"

# 2. 检测 Docker 环境 (国内网络智能回退至阿里云源)
if ! command -v docker >/dev/null 2>&1; then
  echo "⚠️ 未检测到 Docker，正在自动安装 Docker..."
  if [ -n "$GH_PREFIX" ]; then
    echo "⚡ 正在通过国内加速源极速安装 Docker..."
    curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun
  else
    curl -fsSL https://get.docker.com | sh
  fi
fi

# 3. 准备宿主持久化目录与 TUN 虚拟网卡
# 智能适配飞牛 (fnOS) 存储池 (/vol1/1000/micube) 与通用 Linux (/opt/MiCube)
if [ -d "/vol1/1000" ] || [ -d "/vol1" ]; then
  if [ -d "/vol1/1000" ]; then
    BASE_DIR="/vol1/1000/micube"
  else
    BASE_DIR="/vol1/micube"
  fi
  echo "🐳 检测到飞牛 fnOS 存储池，数据将持久化保存在用户空间: $BASE_DIR"
else
  BASE_DIR="/opt/MiCube"
fi
DATA_DIR="$BASE_DIR/data"
ETC_DIR="$BASE_DIR/etc"
mkdir -p "$DATA_DIR" "$ETC_DIR"

if [ ! -c /dev/net/tun ]; then
  mkdir -p /dev/net
  mknod /dev/net/tun c 10 200 2>/dev/null || true
  chmod 666 /dev/net/tun 2>/dev/null || true
fi

# 4. 开启 IPv4 数据包转发
sysctl -w net.ipv4.ip_forward=1 >/dev/null 2>&1 || true

# 5. 准备并启动 MiCube 容器
echo "🚀 正在准备 MiCube 高性能网关镜像..."
IMAGE_NAME="micube:latest"

# 优先匹配本地离线包
if [ -f "MiCube-${MICUBE_ARCH}.tar" ]; then
  echo "📦 检测到本地离线镜像包 MiCube-${MICUBE_ARCH}.tar，正在秒级导入 (0流量)..."
  docker load -i "MiCube-${MICUBE_ARCH}.tar"
  IMAGE_NAME="micube:${MICUBE_ARCH}"
elif [ -f "micube-${MICUBE_ARCH}.tar" ]; then
  echo "📦 检测到本地离线镜像包 micube-${MICUBE_ARCH}.tar，正在秒级导入 (0流量)..."
  docker load -i "micube-${MICUBE_ARCH}.tar"
  IMAGE_NAME="micube:${MICUBE_ARCH}"
elif [ -f "micube.tar" ]; then
  echo "📦 检测到本地离线镜像包 micube.tar，正在秒级导入 (0流量)..."
  docker load -i "micube.tar"
else
  # 本地无离线包，若有加速镜像源，先尝试从国内加速下载离线包，彻底免去 Docker Hub 封锁
  TAR_URL="${GH_PREFIX}https://github.com/zyeyefyass/MiCube/releases/download/Test-V0.91/MiCube-${MICUBE_ARCH}.tar"
  echo "🌐 正在从高速源获取 MiCube-${MICUBE_ARCH}.tar 离线镜像..."
  if curl -fL --progress-bar "$TAR_URL" -o "MiCube-${MICUBE_ARCH}.tar"; then
    echo "✅ 高速下载完成，正在秒级载入镜像..."
    docker load -i "MiCube-${MICUBE_ARCH}.tar"
    IMAGE_NAME="micube:${MICUBE_ARCH}"
  else
    echo "⚠️ 离线包下载失败，尝试常规 Docker 拉取..."
    docker pull ghcr.io/zyeyefyass/micube:latest 2>/dev/null && IMAGE_NAME="ghcr.io/zyeyefyass/micube:latest" || IMAGE_NAME="fyasshole/micube:latest"
  fi
fi

echo "🚀 正在启动 MiCube 高性能网关容器 (${IMAGE_NAME})..."
docker rm -f micube 2>/dev/null || true

docker run -d \
  --name micube \
  --restart always \
  --net host \
  --privileged \
  -v /dev/net/tun:/dev/net/tun \
  -v "$DATA_DIR":/opt/MiCube/data \
  -v "$ETC_DIR":/opt/MiCube/etc \
  -e TZ=Asia/Shanghai \
  -e PORT=2026 \
  "$IMAGE_NAME"

# 6. 获取当前主机 IP 并打印欢迎信息
HOST_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || echo "你的设备IP")

echo "=========================================================="
echo " 🧊 MiCube 容器已成功启动！"
echo " 控制台管理地址: http://${HOST_IP}:2026"
echo " DNS 防泄露网关: 端口 53"
echo " 混合代理端口:   端口 7890"
echo "=========================================================="
