# ==============================================================================
# MiCube RouterOS v7 CT 容器极速一键部署脚本
# 固件要求: RouterOS v7.4+ (开启 container.npk 组件)
# 镜像架构: AMD64 (x86 / x86_64 软路由)
# ==============================================================================

# 1. 宿主网络接口配置 (VETH 虚拟网卡与 Bridge 桥接)
/interface/veth/add name=veth-micube address=10.0.0.11/24 gateway=10.0.0.1
/interface/bridge/port/add bridge=bridge1 interface=veth-micube comment="MiCube-Gateway-VETH"

# 2. 宿主磁盘持久化挂载配置 (数据与配置持久化)
/container/mounts/add name=micube-etc src=disk1/micube/etc dst=/opt/MiCube/etc
/container/mounts/add name=micube-data src=disk1/micube/data dst=/opt/MiCube/data

# 3. 部署并导入 MiCube 容器镜像 (请先将 MiCube-amd64.tar 上传至 RouterOS Files)
/container/add \
  file=MiCube-amd64.tar \
  interface=veth-micube \
  root-dir=disk1/micube \
  mounts=micube-etc,micube-data \
  logging=yes \
  start-on-boot=yes \
  comment="MiCube-HighPerformance-Gateway"

# 4. 打印完成指引
:put "=========================================================="
:put " MiCube 容器已创建成功！"
:put " 请等待 RouterOS 提取完成 (Status: stopped 后点击 Start)"
:put " 控制台访问地址: http://10.0.0.11:2026"
:put " DNS 防泄露监听: 10.0.0.11:53"
:put " HTTP/SOCKS 代理: 10.0.0.11:7890"
:put "=========================================================="
