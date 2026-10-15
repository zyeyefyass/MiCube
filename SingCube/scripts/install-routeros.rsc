# ==============================================================================
# SingCube (sing-box 内核版) RouterOS v7 官方一键容器初始化脚本
# 适用硬件: CCR / RB 系列 / hEX / CHR / x86 RouterOS v7.x
# ==============================================================================

# 1. 创建 VETH 专用虚拟网卡并加入 LAN 网桥 (默认分配 IP: 192.168.88.3/24，网关指向 ROS: 192.168.88.1)
/interface/veth/add name=singcube address=192.168.88.3/24 gateway=192.168.88.1
/interface/bridge/port add bridge=bridge interface=singcube

# 2. 指定临时提取路径 (请确认磁盘名为 disk1 或 usb1)
/container/config/set tmpdir=disk1/pull

# 3. 创建容器实例并配置自启 (自动挂载当前磁盘下的 SingCube-amd64.tar 离线镜像)
/container/add file=disk1/SingCube-amd64.tar interface=singcube root-dir=disk1/singcube \
    envlist="" hostname=SingCube logging=yes start-on-boot=yes

# 4. 立即启动 SingCube 容器
/container/start [find where interface=singcube]
