#!/bin/sh
set -e
SINGCUBE_ROOT="/opt/SingCube"
DATA="$SINGCUBE_ROOT/data"
ETC="$SINGCUBE_ROOT/etc"
mkdir -p "$DATA" "$ETC"

# 1. 虚拟 TUN 网卡设备就绪
if [ ! -c /dev/net/tun ]; then
  mkdir -p /dev/net
  mknod /dev/net/tun c 10 200 2>/dev/null || true
  chmod 666 /dev/net/tun 2>/dev/null || true
fi

# 2. 宿主平台环境感知
if [ -z "$SINGCUBE_PLATFORM" ]; then
  if grep -qa "container" /proc/1/environ 2>/dev/null && ! [ -f /.dockerenv ]; then
    export SINGCUBE_PLATFORM="routeros"
  else
    export SINGCUBE_PLATFORM="docker"
  fi
fi

# 3. 策略路由表自愈 (防止 RouterOS/Linux 出现环路死锁)
if command -v ip >/dev/null 2>&1; then
  while ip rule show | grep -Eq '^(1|2|3):\s+from all lookup (local|main|default)'; do
    pref=$(ip rule show | grep -E '^(1|2|3):\s+from all lookup (local|main|default)' | head -n1 | cut -d: -f1)
    ip rule del pref "$pref" 2>/dev/null || break
  done
  ip rule show | grep -Eq 'pref 200 .*lookup local' || ip rule add pref 200 from all lookup local 2>/dev/null || true
  ip rule show | grep -Eq 'pref 32766 .*lookup main' || ip rule add pref 32766 from all lookup main 2>/dev/null || true
  ip rule show | grep -Eq 'pref 32770 .*lookup default' || ip rule add pref 32770 from all lookup default 2>/dev/null || true
fi

# 4. 内核网络协议栈高吞吐优化与 BBR
echo 1 > /proc/sys/net/ipv4/ip_forward 2>/dev/null || true
echo 1 > /proc/sys/net/ipv6/conf/all/forwarding 2>/dev/null || true
for i in /proc/sys/net/ipv4/conf/*/rp_filter; do echo 0 > "$i" 2>/dev/null || true; done
for i in /proc/sys/net/ipv4/conf/*/send_redirects; do echo 0 > "$i" 2>/dev/null || true; done
[ -f /proc/sys/net/core/rmem_max ] && echo 16777216 > /proc/sys/net/core/rmem_max 2>/dev/null || true
[ -f /proc/sys/net/core/wmem_max ] && echo 16777216 > /proc/sys/net/core/wmem_max 2>/dev/null || true
echo "4096 87380 16777216" > /proc/sys/net/ipv4/tcp_rmem 2>/dev/null || true
echo "4096 65536 16777216" > /proc/sys/net/ipv4/tcp_wmem 2>/dev/null || true
echo 3 > /proc/sys/net/ipv4/tcp_fastopen 2>/dev/null || true
echo 1 > /proc/sys/net/ipv4/tcp_window_scaling 2>/dev/null || true
echo bbr > /proc/sys/net/ipv4/tcp_congestion_control 2>/dev/null || true
ulimit -n 65535 2>/dev/null || ulimit -n 4096 2>/dev/null || true

# 5. 读取或初始化面板端口
panel_port=$(cat "$DATA/panel-port" 2>/dev/null | tr -dc '0-9')
if [ -z "$panel_port" ]; then
  panel_port="${PORT:-2026}"
  echo "$panel_port" > "$DATA/panel-port"
fi

# 6. 开机自启判断
if [ -f "$DATA/autostart" ] && [ -f "$ETC/config.json" ]; then
  "$SINGCUBE_ROOT/container/bin/singcube-ctl" start || true
fi

export PORT="$panel_port"
export HOST="::"
export SINGCUBE_ROOT="$SINGCUBE_ROOT"
export GOMEMLIMIT="${GOMEMLIMIT:-32MiB}"
export GOGC="${GOGC:-30}"

trap '"$SINGCUBE_ROOT/container/bin/singcube-ctl" stop; exit 0' TERM INT

# 7. 嵌入式 Dropbear 运维 SSH 服务
(
  DROPBEAR_BIN=$(command -v dropbear || echo "")
  if [ -n "$DROPBEAR_BIN" ]; then
    mkdir -p /etc/dropbear
    [ -f /etc/dropbear/dropbear_rsa_host_key ] || dropbearkey -t rsa -f /etc/dropbear/dropbear_rsa_host_key 2>/dev/null || true
    [ -f /etc/dropbear/dropbear_ed25519_host_key ] || dropbearkey -t ed25519 -f /etc/dropbear/dropbear_ed25519_host_key 2>/dev/null || true
    echo "root:${SSH_PASSWORD:-1111}" | chpasswd 2>/dev/null || true
    $DROPBEAR_BIN -p "${SSH_PORT:-22}" -B -R 2>/dev/null || true
  fi
) &

exec "$SINGCUBE_ROOT/bin/singcube-server"
