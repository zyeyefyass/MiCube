#!/bin/sh
# ==============================================================================
# MiCube 官方极速一键部署脚本 (Docker / 飞牛 fnOS / PVE / Linux)
# 架构支持: AMD64 (x86_64) / ARM64 (AArch64)
# ==============================================================================
set -e

echo "🧊 正在检测系统环境与依赖..."

# 1. 检测架构
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) MICUBE_ARCH="amd64" ;;
  aarch64|arm64) MICUBE_ARCH="arm64" ;;
  *) echo "❌ 暂不支持的架构: $ARCH"; exit 1 ;;
esac

echo "✅ 检测到系统架构: $MICUBE_ARCH"

# 2. 检测 Docker 环境
if ! command -v docker >/dev/null 2>&1; then
  echo "⚠️ 未检测到 Docker，正在自动安装 Docker..."
  curl -fsSL https://get.docker.com | sh
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

# 5. 拉取并启动 MiCube 容器
echo "🚀 正在启动 MiCube 高性能网关容器..."
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
  fyasshole/micube:latest

# 6. 获取当前主机 IP 并打印欢迎信息
HOST_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || echo "你的设备IP")

echo "=========================================================="
echo " 🧊 MiCube 容器已成功启动！"
echo " 控制台管理地址: http://${HOST_IP}:2026"
echo " DNS 防泄露网关: 端口 53"
echo " 混合代理端口:   端口 7890"
echo "=========================================================="
