#!/bin/sh
# ==============================================================================
# SingCube (sing-box 内核版) 官方极速一键部署脚本
# 适用平台: Docker / 飞牛 fnOS / PVE LXC / 群晖 NAS / Debian / Ubuntu / Linux
# 架构支持: AMD64 (x86_64) / ARM64 (AArch64)
# 特性: 内置国内无代理极速 CDN 加速，支持 0 流量本地离线包导入
# ==============================================================================
set -e

echo "🌌 [SingCube] 正在检测系统硬件架构与网络环境..."

# 1. 架构检测
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) SC_ARCH="amd64" ;;
  aarch64|arm64) SC_ARCH="arm64" ;;
  *) echo "❌ 暂不支持的架构: $ARCH"; exit 1 ;;
esac
echo "✅ 检测到系统架构: $SC_ARCH"

# 2. 网络探测与 CDN 加速选择 (解决国内无科学上网无法访问 GitHub / Docker Hub)
IS_CHINA=0
if ! curl -s --connect-timeout 3 https://api.github.com >/dev/null 2>&1; then
  echo "⚡ 检测到当前处于国内直连网络环境，已自动为您启用全球 CDN 镜像加速通道！"
  IS_CHINA=1
fi

CDN_PREFIX=""
if [ "$IS_CHINA" -eq 1 ]; then
  CANDIDATE_CDNS="https://ghfast.top/ https://mirror.ghproxy.com/ https://ghproxy.cn/"
  for cdn in $CANDIDATE_CDNS; do
    if curl -s --connect-timeout 2 "${cdn}https://raw.githubusercontent.com" >/dev/null 2>&1; then
      CDN_PREFIX="$cdn"
      break
    fi
  done
  [ -z "$CDN_PREFIX" ] && CDN_PREFIX="https://ghfast.top/"
  echo "🚀 已优选高速 CDN 加速节点: $CDN_PREFIX"
fi

# 3. Docker 环境就绪检测
if ! command -v docker >/dev/null 2>&1; then
  echo "⚠️ 未检测到 Docker 运行时，正在为您全自动安装..."
  if [ "$IS_CHINA" -eq 1 ]; then
    curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun
  else
    curl -fsSL https://get.docker.com | sh
  fi
fi

# 4. 准备宿主持久化目录与 TUN 虚拟网卡
# 智能适配飞牛 fnOS (/vol1/1000/singcube) 与标准 Linux (/opt/SingCube)
if [ -d "/vol1/1000" ] || [ -d "/vol1" ]; then
  if [ -d "/vol1/1000" ]; then
    BASE_DIR="/vol1/1000/singcube"
  else
    BASE_DIR="/vol1/singcube"
  fi
  echo "🐳 检测到飞牛 fnOS 存储池，数据将持久化保存在用户空间: $BASE_DIR"
else
  BASE_DIR="/opt/SingCube"
fi
DATA_DIR="$BASE_DIR/data"
ETC_DIR="$BASE_DIR/etc"
mkdir -p "$DATA_DIR" "$ETC_DIR"

if [ ! -c /dev/net/tun ]; then
  mkdir -p /dev/net
  mknod /dev/net/tun c 10 200 2>/dev/null || true
  chmod 666 /dev/net/tun 2>/dev/null || true
fi

# 5. 开启内核 IPv4 数据包转发
sysctl -w net.ipv4.ip_forward=1 >/dev/null 2>&1 || true

# 6. 容器镜像获取 (优先使用本地离线镜像包 -> CDN 直链加速下载 -> Docker Hub)
IMAGE_NAME="singcube:latest"

if [ -f "SingCube-${SC_ARCH}.tar" ]; then
  echo "📦 [0流量极速] 检测到当前目录下的离线镜像包 SingCube-${SC_ARCH}.tar，正在秒级导入..."
  docker load -i "SingCube-${SC_ARCH}.tar"
  IMAGE_NAME="singcube:${SC_ARCH}"
elif [ -f "singcube-${SC_ARCH}.tar" ]; then
  echo "📦 [0流量极速] 检测到当前目录下的离线镜像包 singcube-${SC_ARCH}.tar，正在秒级导入..."
  docker load -i "singcube-${SC_ARCH}.tar"
  IMAGE_NAME="singcube:${SC_ARCH}"
elif [ -f "singcube.tar" ]; then
  echo "📦 [0流量极速] 检测到当前目录下的离线镜像包 singcube.tar，正在秒级导入..."
  docker load -i "singcube.tar"
else
  if [ "$IS_CHINA" -eq 1 ]; then
    TAR_NAME="SingCube-${SC_ARCH}.tar"
    TAR_URL="${CDN_PREFIX}https://github.com/zyeyefyass/SingCube/releases/download/Test-V0.91/${TAR_NAME}"
    echo "⚡ 正在通过国内加速节点秒级下载镜像归档包 (${TAR_NAME})..."
    if curl -fSL --progress-bar "$TAR_URL" -o "/tmp/${TAR_NAME}"; then
      echo "📦 下载完成，正在导入 Docker 镜像..."
      docker load -i "/tmp/${TAR_NAME}"
      rm -f "/tmp/${TAR_NAME}"
      IMAGE_NAME="singcube:${SC_ARCH}"
    else
      echo "⚠️ 加速下载失败，尝试通过 Docker Registry 拉取..."
      docker pull ghcr.io/zyeyefyass/singcube:latest 2>/dev/null && IMAGE_NAME="ghcr.io/zyeyefyass/singcube:latest" || IMAGE_NAME="fyasshole/singcube:latest"
    fi
  else
    echo "🌐 正在从 GitHub Packages / Docker Hub 拉取最新 SingCube 镜像..."
    docker pull ghcr.io/zyeyefyass/singcube:latest 2>/dev/null && IMAGE_NAME="ghcr.io/zyeyefyass/singcube:latest" || IMAGE_NAME="fyasshole/singcube:latest"
  fi
fi

# 7. 部署并启动 SingCube 容器
echo "🚀 正在启动 SingCube (sing-box 内核) 旁路由网关容器..."
docker rm -f singcube 2>/dev/null || true

docker run -d \
  --name singcube \
  --restart always \
  --net host \
  --privileged \
  -v /dev/net/tun:/dev/net/tun \
  -v "$DATA_DIR":/opt/SingCube/data \
  -v "$ETC_DIR":/opt/SingCube/etc \
  -e TZ=Asia/Shanghai \
  -e PORT=2026 \
  "$IMAGE_NAME"

# 8. 获取主机 IP 并打印欢迎卡片
HOST_IP=$(hostname -I 2>/dev/null | awk '{print $1}' || echo "你的设备IP")

echo "=========================================================="
echo " 🌌 SingCube (sing-box 极速内核) 容器已成功启动！"
echo " 控制台管理地址: http://${HOST_IP}:2026"
echo " DNS 防泄露网关: 端口 53"
echo " 本地混合代理:   端口 7890"
echo " 内核控制接口:   http://127.0.0.1:9090 (Clash API 兼容)"
echo " 数据持久化目录: $BASE_DIR"
echo "=========================================================="
