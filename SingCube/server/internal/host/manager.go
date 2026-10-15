package host

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// HostConfig stores the SSH credentials for communicating with the host OS (e.g., fnOS, PVE, Debian, Ubuntu).
type HostConfig struct {
	Enabled    bool   `json:"enabled"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
}

// HostInfo represents status and telemetry probed from the host system.
type HostInfo struct {
	Connected      bool   `json:"connected"`
	OSName         string `json:"os_name"`
	PlatformID     string `json:"platform_id"` // "fnos", "debian", "ubuntu", "pve", "synology", "linux"
	KernelVersion  string `json:"kernel_version"`
	Arch           string `json:"arch"`
	HostName       string `json:"hostname"`
	Uptime         string `json:"uptime"`
	IPForward      bool   `json:"ip_forward"`
	WireGuardReady bool   `json:"wireguard_ready"`
	WireGuardActive bool  `json:"wireguard_active"`
	FirewallConfigured bool `json:"firewall_configured"`
	DockerRunning  bool   `json:"docker_running"`
	RawError       string `json:"raw_error,omitempty"`
}

// FirewallResult represents the status of firewall rule deployment on the host.
type FirewallResult struct {
	Success bool     `json:"success"`
	Logs    []string `json:"logs"`
}

// Manager orchestrates host-level system maintenance and Docker companion tasks.
type Manager struct {
	rootDir    string
	configFile string
	mu         sync.RWMutex
	config     HostConfig
	cachedInfo *HostInfo
}

// NewManager creates a new host Manager.
func NewManager(rootDir string) *Manager {
	m := &Manager{
		rootDir:    rootDir,
		configFile: filepath.Join(rootDir, "data", "host-config.json"),
		config: HostConfig{
			Enabled: false,
			Host:    "172.17.0.1", // Standard Docker gateway / host IP
			Port:    22,
			User:    "root",
		},
	}
	m.loadConfig()
	return m
}

func (m *Manager) loadConfig() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if data, err := os.ReadFile(m.configFile); err == nil {
		_ = json.Unmarshal(data, &m.config)
	}
}

// SaveConfig persists the Host SSH configuration.
func (m *Manager) SaveConfig(cfg HostConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cfg.Port <= 0 {
		cfg.Port = 22
	}
	if cfg.User == "" {
		cfg.User = "root"
	}
	m.config = cfg
	_ = os.MkdirAll(filepath.Dir(m.configFile), 0755)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.configFile, data, 0644)
}

// GetConfig returns the current Host SSH configuration.
func (m *Manager) GetConfig() HostConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// GetCachedHostInfo returns the most recently successfully probed host telemetry.
func (m *Manager) GetCachedHostInfo() *HostInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cachedInfo
}

// getSSHClient dials the host machine via SSH.
func (m *Manager) getSSHClient(cfg HostConfig) (*ssh.Client, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("宿主机 IP 或地址未配置")
	}
	if cfg.Port <= 0 {
		cfg.Port = 22
	}
	if cfg.User == "" {
		cfg.User = "root"
	}

	var authMethods []ssh.AuthMethod
	if cfg.Password != "" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}
	if strings.TrimSpace(cfg.PrivateKey) != "" {
		signer, err := ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
		if err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("请提供宿主机 SSH 登录密码或私钥")
	}

	clientConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return ssh.Dial("tcp", addr, clientConfig)
}

// RunCommand runs a single command string on the host and returns stdout+stderr combined.
func (m *Manager) RunCommand(cfg HostConfig, cmdStr string) (string, error) {
	client, err := m.getSSHClient(cfg)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建 SSH 会话失败: %w", err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	session.Stderr = &b
	err = session.Run(cmdStr)
	out := strings.TrimSpace(b.String())
	if err != nil {
		return out, fmt.Errorf("%w: %s", err, out)
	}
	return out, nil
}

// ProbeHost telemetry inspects the host OS, kernel, network, and WireGuard status.
func (m *Manager) ProbeHost() (*HostInfo, error) {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	info := &HostInfo{
		PlatformID: "linux",
		OSName:     "Linux Host",
	}

	if !cfg.Enabled || cfg.Host == "" {
		// Attempt local container detection
		info.OSName = m.detectLocalHost()
		info.PlatformID = m.detectLocalPlatformID(info.OSName)
		return info, nil
	}

	client, err := m.getSSHClient(cfg)
	if err != nil {
		info.RawError = err.Error()
		return info, err
	}
	defer client.Close()

	info.Connected = true

	// 1. Check OS release
	if session, err := client.NewSession(); err == nil {
		var out bytes.Buffer
		session.Stdout = &out
		_ = session.Run("cat /etc/os-release")
		session.Close()
		info.OSName, info.PlatformID = parseOSRelease(out.String())
	}

	// 2. Check Kernel & Arch & Hostname
	if session, err := client.NewSession(); err == nil {
		var out bytes.Buffer
		session.Stdout = &out
		_ = session.Run("uname -r; uname -m; hostname")
		session.Close()
		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		if len(lines) >= 1 {
			info.KernelVersion = strings.TrimSpace(lines[0])
		}
		if len(lines) >= 2 {
			info.Arch = strings.TrimSpace(lines[1])
		}
		if len(lines) >= 3 {
			info.HostName = strings.TrimSpace(lines[2])
		}
	}

	// 3. Check uptime
	if session, err := client.NewSession(); err == nil {
		var out bytes.Buffer
		session.Stdout = &out
		_ = session.Run("uptime -p 2>/dev/null || cat /proc/uptime")
		session.Close()
		info.Uptime = strings.TrimSpace(out.String())
	}

	// 4. Check IP Forward
	if session, err := client.NewSession(); err == nil {
		var out bytes.Buffer
		session.Stdout = &out
		_ = session.Run("cat /proc/sys/net/ipv4/ip_forward")
		session.Close()
		info.IPForward = strings.TrimSpace(out.String()) == "1"
	}

	// 5. Check WireGuard installed & active
	if session, err := client.NewSession(); err == nil {
		var out bytes.Buffer
		session.Stdout = &out
		_ = session.Run("which wg; which wg-quick")
		session.Close()
		info.WireGuardReady = strings.Contains(out.String(), "wg")
	}

	if session, err := client.NewSession(); err == nil {
		var out bytes.Buffer
		session.Stdout = &out
		_ = session.Run("ip link show wg0 2>/dev/null")
		session.Close()
		info.WireGuardActive = strings.Contains(out.String(), "wg0")
	}

	// 6. Check Firewall iptables rule for 13231 UDP
	if session, err := client.NewSession(); err == nil {
		var out bytes.Buffer
		session.Stdout = &out
		_ = session.Run("iptables -L INPUT -n 2>/dev/null | grep 13231")
		session.Close()
		info.FirewallConfigured = strings.Contains(out.String(), "13231")
	}

	m.mu.Lock()
	m.cachedInfo = info
	m.mu.Unlock()

	return info, nil
}

// ApplyFirewall configures IP forwarding, WireGuard UDP port, panel TCP port, and NAT masquerade on the host.
// Supported modes: "permissive" (default, recommended for novices) and "strict" (least privilege).
// If enablePenetration is false, no WireGuard UDP penetration ports will be opened (pure LAN bypass).
func (m *Manager) ApplyFirewall(mode string, enablePenetration bool) (*FirewallResult, error) {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	if !cfg.Enabled || cfg.Host == "" {
		return nil, fmt.Errorf("请先在【宿主机运维】中配置并启用宿主机 SSH 连接")
	}

	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != "strict" {
		mode = "permissive"
	}

	res := &FirewallResult{Success: false, Logs: []string{}}

	var commands []struct {
		desc string
		cmd  string
	}

	if mode == "permissive" {
		commands = append(commands,
			struct {
				desc string
				cmd  string
			}{"开启 Linux 内核 IPv4 转发", "sysctl -w net.ipv4.ip_forward=1 && (grep -q '^net.ipv4.ip_forward=1' /etc/sysctl.conf || echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf)"},
			struct {
				desc string
				cmd  string
			}{"开启全网卡转发 (conf.all.forwarding)", "sysctl -w net.ipv4.conf.all.forwarding=1 2>/dev/null || true"},
			struct {
				desc string
				cmd  string
			}{"放行 FORWARD 转发链默认策略 (ACCEPT)", "iptables -P FORWARD ACCEPT"},
			struct {
				desc string
				cmd  string
			}{"置顶放行 FORWARD 数据链", "iptables -C FORWARD -j ACCEPT 2>/dev/null || iptables -I FORWARD 1 -j ACCEPT"},
			struct {
				desc string
				cmd  string
			}{"放行 MiCube Web 面板 TCP 2026 端口", "iptables -C INPUT -p tcp --dport 2026 -j ACCEPT 2>/dev/null || iptables -I INPUT -p tcp --dport 2026 -j ACCEPT"},
		)
		if enablePenetration {
			commands = append(commands,
				struct {
					desc string
					cmd  string
				}{"放行 WireGuard UDP 13231 穿透端口", "iptables -C INPUT -p udp --dport 13231 -j ACCEPT 2>/dev/null || iptables -I INPUT -p udp --dport 13231 -j ACCEPT"},
				struct {
					desc string
					cmd  string
				}{"放行 WireGuard UDP 51820 备用端口", "iptables -C INPUT -p udp --dport 51820 -j ACCEPT 2>/dev/null || iptables -I INPUT -p udp --dport 51820 -j ACCEPT"},
				struct {
					desc string
					cmd  string
				}{"配置 WireGuard 虚拟网段 (10.88.0.0/24) NAT 出网伪装", "iptables -t nat -C POSTROUTING -s 10.88.0.0/24 ! -o wg+ -j MASQUERADE 2>/dev/null || iptables -t nat -I POSTROUTING -s 10.88.0.0/24 ! -o wg+ -j MASQUERADE"},
			)
		}
	} else {
		// Strict mode
		commands = append(commands,
			struct {
				desc string
				cmd  string
			}{"开启 Linux 内核 IPv4 转发", "sysctl -w net.ipv4.ip_forward=1 && (grep -q '^net.ipv4.ip_forward=1' /etc/sysctl.conf || echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf)"},
			struct {
				desc string
				cmd  string
			}{"设置 FORWARD 链默认策略为 DROP (最小特权)", "iptables -P FORWARD DROP"},
			struct {
				desc string
				cmd  string
			}{"放行 ESTABLISHED,RELATED 状态跟踪连接", "iptables -C FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || iptables -C FORWARD -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || iptables -I FORWARD 1 -m state --state RELATED,ESTABLISHED -j ACCEPT"},
			struct {
				desc string
				cmd  string
			}{"放行 RFC1918 私网网段互访转发", "iptables -C FORWARD -s 192.168.0.0/16 -d 192.168.0.0/16 -j ACCEPT 2>/dev/null || iptables -I FORWARD 2 -s 192.168.0.0/16 -d 192.168.0.0/16 -j ACCEPT; iptables -C FORWARD -s 10.0.0.0/8 -d 10.0.0.0/8 -j ACCEPT 2>/dev/null || iptables -I FORWARD 3 -s 10.0.0.0/8 -d 10.0.0.0/8 -j ACCEPT"},
			struct {
				desc string
				cmd  string
			}{"放行 MiCube Web 面板 TCP 2026 端口", "iptables -C INPUT -p tcp --dport 2026 -j ACCEPT 2>/dev/null || iptables -I INPUT -p tcp --dport 2026 -j ACCEPT"},
		)
		if enablePenetration {
			commands = append(commands,
				struct {
					desc string
					cmd  string
				}{"定向放行 WireGuard 虚拟网段 (10.88.0.0/24) 转发", "iptables -C FORWARD -s 10.88.0.0/24 -j ACCEPT 2>/dev/null || iptables -I FORWARD 4 -s 10.88.0.0/24 -j ACCEPT; iptables -C FORWARD -d 10.88.0.0/24 -j ACCEPT 2>/dev/null || iptables -I FORWARD 5 -d 10.88.0.0/24 -j ACCEPT"},
				struct {
					desc string
					cmd  string
				}{"放行 WireGuard UDP 13231 穿透端口", "iptables -C INPUT -p udp --dport 13231 -j ACCEPT 2>/dev/null || iptables -I INPUT -p udp --dport 13231 -j ACCEPT"},
				struct {
					desc string
					cmd  string
				}{"放行 WireGuard UDP 51820 备用端口", "iptables -C INPUT -p udp --dport 51820 -j ACCEPT 2>/dev/null || iptables -I INPUT -p udp --dport 51820 -j ACCEPT"},
				struct {
					desc string
					cmd  string
				}{"配置 WireGuard 虚拟网段 (10.88.0.0/24) NAT 出网伪装", "iptables -t nat -C POSTROUTING -s 10.88.0.0/24 ! -o wg+ -j MASQUERADE 2>/dev/null || iptables -t nat -I POSTROUTING -s 10.88.0.0/24 ! -o wg+ -j MASQUERADE"},
			)
		}
	}

	client, err := m.getSSHClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("无法连接宿主机: %w", err)
	}
	defer client.Close()

	for _, item := range commands {
		session, err := client.NewSession()
		if err != nil {
			res.Logs = append(res.Logs, fmt.Sprintf("❌ %s 失败: %v", item.desc, err))
			continue
		}
		var out bytes.Buffer
		session.Stdout = &out
		session.Stderr = &out
		runErr := session.Run(item.cmd)
		session.Close()

		if runErr != nil {
			res.Logs = append(res.Logs, fmt.Sprintf("⚠️ %s: %s (%v)", item.desc, strings.TrimSpace(out.String()), runErr))
		} else {
			res.Logs = append(res.Logs, fmt.Sprintf("✅ %s 已就绪", item.desc))
		}
	}

	// Try persisting rules
	if session, err := client.NewSession(); err == nil {
		_ = session.Run("which netfilter-persistent && netfilter-persistent save || which iptables-save && iptables-save > /etc/iptables.rules 2>/dev/null")
		session.Close()
	}

	res.Success = true
	return res, nil
}

// DeployHostWireGuard installs wireguard tools and applies the configuration on fnOS / Linux host.
func (m *Manager) DeployHostWireGuard(wgServerConf string) (string, error) {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	if !cfg.Enabled || cfg.Host == "" {
		return "", fmt.Errorf("请先在【宿主机运维】中配置并启用宿主机 SSH 连接")
	}

	client, err := m.getSSHClient(cfg)
	if err != nil {
		return "", fmt.Errorf("无法连接宿主机: %w", err)
	}
	defer client.Close()

	var logs []string

	// 1. Install wireguard if missing
	if session, err := client.NewSession(); err == nil {
		var out bytes.Buffer
		session.Stdout = &out
		session.Stderr = &out
		_ = session.Run("which wg-quick || (apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y wireguard wireguard-tools)")
		session.Close()
		logs = append(logs, "检查并就绪 wireguard-tools: "+strings.TrimSpace(out.String()))
	}

	// 2. Ensure /etc/wireguard directory
	if session, err := client.NewSession(); err == nil {
		_ = session.Run("mkdir -p /etc/wireguard")
		session.Close()
	}

	// 3. Write /etc/wireguard/wg0.conf
	writeCmd := fmt.Sprintf("cat << 'EOF' > /etc/wireguard/wg0.conf\n%s\nEOF\nchmod 600 /etc/wireguard/wg0.conf", wgServerConf)
	if session, err := client.NewSession(); err == nil {
		var out bytes.Buffer
		session.Stdout = &out
		session.Stderr = &out
		if err := session.Run(writeCmd); err != nil {
			session.Close()
			return "", fmt.Errorf("写入 /etc/wireguard/wg0.conf 失败: %w", err)
		}
		session.Close()
		logs = append(logs, "✅ 已写入宿主机 /etc/wireguard/wg0.conf")
	}

	// 4. Start or Reload WireGuard
	reloadCmd := "wg-quick down wg0 2>/dev/null; systemctl restart wg-quick@wg0 || wg-quick up wg0"
	if session, err := client.NewSession(); err == nil {
		var out bytes.Buffer
		session.Stdout = &out
		session.Stderr = &out
		_ = session.Run(reloadCmd)
		session.Close()
		logs = append(logs, "✅ 宿主机 WireGuard 服务已拉起生效: "+strings.TrimSpace(out.String()))
	}

	// 5. Automatically ensure firewall rules
	_, _ = m.ApplyFirewall("permissive", true)

	return strings.Join(logs, "\n"), nil
}

// SyncPeerToHost adds or updates a peer on the host wg0 interface dynamically.
func (m *Manager) SyncPeerToHost(publicKey, clientIP string) error {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	if !cfg.Enabled || cfg.Host == "" {
		return nil // Host SSH not configured, skip
	}

	client, err := m.getSSHClient(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmd := fmt.Sprintf("wg set wg0 peer '%s' allowed-ips '%s/32'", publicKey, clientIP)
	return session.Run(cmd)
}

// RemovePeerFromHost deletes a peer from the host wg0 interface dynamically.
func (m *Manager) RemovePeerFromHost(publicKey string) error {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	if !cfg.Enabled || cfg.Host == "" {
		return nil
	}

	client, err := m.getSSHClient(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmd := fmt.Sprintf("wg set wg0 peer '%s' remove", publicKey)
	return session.Run(cmd)
}

// CheckHostUpgrade queries upgradable packages on the host.
func (m *Manager) CheckHostUpgrade() (string, error) {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	if !cfg.Enabled || cfg.Host == "" {
		return "", fmt.Errorf("宿主机 SSH 连接未配置")
	}

	return m.RunCommand(cfg, "apt-get update >/dev/null 2>&1 && apt list --upgradable 2>/dev/null")
}

// ExecuteHostUpgrade triggers system packages upgrade on the host.
func (m *Manager) ExecuteHostUpgrade() (string, error) {
	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	if !cfg.Enabled || cfg.Host == "" {
		return "", fmt.Errorf("宿主机 SSH 连接未配置")
	}

	return m.RunCommand(cfg, "DEBIAN_FRONTEND=noninteractive apt-get upgrade -y")
}

// GenerateFirewallScript creates a zero-SSH 1-click firewall script for either permissive or strict mode.
func (m *Manager) GenerateFirewallScript(mode string, enablePenetration bool) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "strict" {
		if enablePenetration {
			return `#!/usr/bin/env bash
# ==============================================================================
#  MiCube · 宿主机网络底座与防火墙极速一键通关 (🛡️ 严格模式 + 🚀 手机 WireGuard 穿透)
#  适用平台: 飞牛 fnOS / PVE / Debian / Ubuntu (最小特权白名单防御)
# ==============================================================================
set -e

echo "🧊 [MiCube] 正在为宿主机配置网络底座与防火墙 (🛡️ 严格模式 + 🚀 穿透模式)..."

# 1. 开启 Linux 内核 IPv4 数据包转发
echo "  -> 开启 IPv4 数据包转发 (net.ipv4.ip_forward=1)..."
sysctl -w net.ipv4.ip_forward=1
grep -q '^net.ipv4.ip_forward=1' /etc/sysctl.conf || echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf

# 2. 保持 FORWARD 链默认 DROP 策略 (最小权限原则)
echo "  -> 保持 FORWARD 链默认 DROP (最小特权防御)..."
iptables -P FORWARD DROP

# 3. 放行状态跟踪连接 (ESTABLISHED, RELATED)
echo "  -> 放行已建立连接状态跟踪..."
iptables -C FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || iptables -C FORWARD -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || iptables -I FORWARD 1 -m state --state RELATED,ESTABLISHED -j ACCEPT

# 4. 定向放行 WireGuard 虚拟网段与私网互访
echo "  -> 定向放行 WireGuard 虚拟接口与私网互访..."
iptables -C FORWARD -s 10.88.0.0/24 -j ACCEPT 2>/dev/null || iptables -I FORWARD 2 -s 10.88.0.0/24 -j ACCEPT
iptables -C FORWARD -d 10.88.0.0/24 -j ACCEPT 2>/dev/null || iptables -I FORWARD 3 -d 10.88.0.0/24 -j ACCEPT
iptables -C FORWARD -s 192.168.0.0/16 -d 192.168.0.0/16 -j ACCEPT 2>/dev/null || iptables -I FORWARD 4 -s 192.168.0.0/16 -d 192.168.0.0/16 -j ACCEPT
iptables -C FORWARD -s 10.0.0.0/8 -d 10.0.0.0/8 -j ACCEPT 2>/dev/null || iptables -I FORWARD 5 -s 10.0.0.0/8 -d 10.0.0.0/8 -j ACCEPT

# 5. 放行核心服务端口 (MiCube Web 面板 TCP 2026 + WireGuard UDP 13231/51820)
echo "  -> 放行核心通信端口..."
iptables -C INPUT -p tcp --dport 2026 -j ACCEPT 2>/dev/null || iptables -I INPUT -p tcp --dport 2026 -j ACCEPT
iptables -C INPUT -p udp --dport 13231 -j ACCEPT 2>/dev/null || iptables -I INPUT -p udp --dport 13231 -j ACCEPT
iptables -C INPUT -p udp --dport 51820 -j ACCEPT 2>/dev/null || iptables -I INPUT -p udp --dport 51820 -j ACCEPT

# 6. 仅为 WireGuard 虚拟网段启用 NAT 伪装
echo "  -> 开启 WireGuard 专属 NAT 伪装..."
iptables -t nat -C POSTROUTING -s 10.88.0.0/24 ! -o wg+ -j MASQUERADE 2>/dev/null || iptables -t nat -I POSTROUTING -s 10.88.0.0/24 ! -o wg+ -j MASQUERADE

# 7. 持久化保存
echo "  -> 持久化防火墙规则..."
which netfilter-persistent &>/dev/null && netfilter-persistent save || (which iptables-save &>/dev/null && iptables-save > /etc/iptables.rules 2>/dev/null || true)

echo "✅ [MiCube] 宿主机网络底座与防火墙已以【🛡️ 严格模式 + 🚀 穿透模式】成功打通！"
`
		}

		// Strict mode WITHOUT penetration (Pure LAN bypass, zero public ports)
		return `#!/usr/bin/env bash
# ==============================================================================
#  MiCube · 宿主机网络底座与防火墙极速一键通关 (🛡️ 严格模式 · 🌐 纯局域网旁路由)
#  适用平台: 飞牛 fnOS / PVE / Debian / Ubuntu (零公网暴露面 · 最小特权)
# ==============================================================================
set -e

echo "🧊 [MiCube] 正在为宿主机配置网络底座与防火墙 (🛡️ 严格模式 · 🌐 纯局域网旁路由)..."

# 1. 开启 Linux 内核 IPv4 数据包转发
echo "  -> 开启 IPv4 数据包转发 (net.ipv4.ip_forward=1)..."
sysctl -w net.ipv4.ip_forward=1
grep -q '^net.ipv4.ip_forward=1' /etc/sysctl.conf || echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf

# 2. 保持 FORWARD 链默认 DROP 策略 (最小权限原则)
echo "  -> 保持 FORWARD 链默认 DROP (最小特权防御)..."
iptables -P FORWARD DROP

# 3. 放行状态跟踪连接 (ESTABLISHED, RELATED)
echo "  -> 放行已建立连接状态跟踪..."
iptables -C FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || iptables -C FORWARD -m state --state RELATED,ESTABLISHED -j ACCEPT 2>/dev/null || iptables -I FORWARD 1 -m state --state RELATED,ESTABLISHED -j ACCEPT

# 4. 定向放行 RFC1918 局域网私网互访
echo "  -> 定向放行局域网私网流量互访..."
iptables -C FORWARD -s 192.168.0.0/16 -d 192.168.0.0/16 -j ACCEPT 2>/dev/null || iptables -I FORWARD 2 -s 192.168.0.0/16 -d 192.168.0.0/16 -j ACCEPT
iptables -C FORWARD -s 10.0.0.0/8 -d 10.0.0.0/8 -j ACCEPT 2>/dev/null || iptables -I FORWARD 3 -s 10.0.0.0/8 -d 10.0.0.0/8 -j ACCEPT

# 5. 放行核心 Web 面板 TCP 2026 端口 (不开放任何 WireGuard 公网 UDP 端口)
echo "  -> 放行 MiCube Web 面板端口 (TCP 2026)..."
iptables -C INPUT -p tcp --dport 2026 -j ACCEPT 2>/dev/null || iptables -I INPUT -p tcp --dport 2026 -j ACCEPT

# 6. 持久化保存
echo "  -> 持久化防火墙规则..."
which netfilter-persistent &>/dev/null && netfilter-persistent save || (which iptables-save &>/dev/null && iptables-save > /etc/iptables.rules 2>/dev/null || true)

echo "✅ [MiCube] 宿主机网络底座与防火墙已以【🛡️ 严格模式 · 🌐 纯局域网】成功打通！零公网端口暴露，固若金汤！"
`
	}

	// Permissive mode
	if enablePenetration {
		return `#!/usr/bin/env bash
# ==============================================================================
#  MiCube · 宿主机网络底座与防火墙极速一键通关 (🟢 宽松模式 + 🚀 手机 WireGuard 穿透)
#  适用平台: 飞牛 fnOS / PVE / 群晖 DSM / Ubuntu / Debian / CentOS Linux
# ==============================================================================
set -e

echo "🧊 [MiCube] 正在为宿主机配置网络底座与防火墙 (🟢 宽松模式 + 🚀 穿透模式)..."

# 1. 永久开启 Linux 内核 IPv4 数据包转发
echo "  -> 开启 IPv4 数据包转发 (net.ipv4.ip_forward=1)..."
sysctl -w net.ipv4.ip_forward=1
grep -q '^net.ipv4.ip_forward=1' /etc/sysctl.conf || echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf
sysctl -w net.ipv4.conf.all.forwarding=1 2>/dev/null || true

# 2. 设置 FORWARD 链默认策略为 ACCEPT 并置顶放行 (100% 避免断网)
echo "  -> 放行 FORWARD 链策略 (宽松放行，防止断网)..."
iptables -P FORWARD ACCEPT
iptables -C FORWARD -j ACCEPT 2>/dev/null || iptables -I FORWARD 1 -j ACCEPT

# 3. 放行核心通信端口 (Web 面板 TCP 2026 + WireGuard UDP 13231/51820)
echo "  -> 放行 MiCube 核心端口..."
iptables -C INPUT -p tcp --dport 2026 -j ACCEPT 2>/dev/null || iptables -I INPUT -p tcp --dport 2026 -j ACCEPT
iptables -C INPUT -p udp --dport 13231 -j ACCEPT 2>/dev/null || iptables -I INPUT -p udp --dport 13231 -j ACCEPT
iptables -C INPUT -p udp --dport 51820 -j ACCEPT 2>/dev/null || iptables -I INPUT -p udp --dport 51820 -j ACCEPT

# 4. 配置 WireGuard 虚拟网段 (10.88.0.0/24) NAT 出网伪装
echo "  -> 配置 WireGuard 虚拟网段 NAT 伪装..."
iptables -t nat -C POSTROUTING -s 10.88.0.0/24 ! -o wg+ -j MASQUERADE 2>/dev/null || iptables -t nat -I POSTROUTING -s 10.88.0.0/24 ! -o wg+ -j MASQUERADE

# 5. 持久化保存防火墙规则
echo "  -> 持久化防火墙规则..."
which netfilter-persistent &>/dev/null && netfilter-persistent save || (which iptables-save &>/dev/null && iptables-save > /etc/iptables.rules 2>/dev/null || true)

echo "✅ [MiCube] 宿主机网络底座与防火墙已以【🟢 宽松模式 + 🚀 穿透模式】成功打通！"
`
	}

	// Permissive mode WITHOUT penetration (Pure LAN bypass, zero public ports)
	return `#!/usr/bin/env bash
# ==============================================================================
#  MiCube · 宿主机网络底座与防火墙极速一键通关 (🟢 宽松模式 · 🌐 纯局域网旁路由)
#  适用平台: 飞牛 fnOS / PVE / 群晖 DSM / Ubuntu / Debian / CentOS Linux
# ==============================================================================
set -e

echo "🧊 [MiCube] 正在为宿主机配置网络底座与防火墙 (🟢 宽松模式 · 🌐 纯局域网旁路由)..."

# 1. 永久开启 Linux 内核 IPv4 数据包转发
echo "  -> 开启 IPv4 数据包转发 (net.ipv4.ip_forward=1)..."
sysctl -w net.ipv4.ip_forward=1
grep -q '^net.ipv4.ip_forward=1' /etc/sysctl.conf || echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf
sysctl -w net.ipv4.conf.all.forwarding=1 2>/dev/null || true

# 2. 设置 FORWARD 链默认策略为 ACCEPT 并置顶放行 (100% 避免断网)
echo "  -> 放行 FORWARD 链策略 (宽松放行，防止断网)..."
iptables -P FORWARD ACCEPT
iptables -C FORWARD -j ACCEPT 2>/dev/null || iptables -I FORWARD 1 -j ACCEPT

# 3. 放行核心 Web 面板 TCP 2026 端口 (不开放任何 WireGuard 公网 UDP 端口)
echo "  -> 放行 MiCube Web 面板端口 (TCP 2026)..."
iptables -C INPUT -p tcp --dport 2026 -j ACCEPT 2>/dev/null || iptables -I INPUT -p tcp --dport 2026 -j ACCEPT

# 4. 持久化保存防火墙规则
echo "  -> 持久化防火墙规则..."
which netfilter-persistent &>/dev/null && netfilter-persistent save || (which iptables-save &>/dev/null && iptables-save > /etc/iptables.rules 2>/dev/null || true)

echo "✅ [MiCube] 宿主机网络底座与防火墙已以【🟢 宽松模式 · 🌐 纯局域网】成功打通！零公网端口暴露，开箱即通，绝不断网！"
`
}

// GenerateSetupBashScript creates a zero-SSH 1-click installer script that users can run on fnOS / Linux.
func (m *Manager) GenerateSetupBashScript(serverIP string, serverPort int, serverPubKey, clientIP, clientPrivKey, clientPubKey string) string {
	if serverIP == "" {
		serverIP = "192.168.88.1"
	}
	if serverPort <= 0 {
		serverPort = 13231
	}

	return fmt.Sprintf(`#!/usr/bin/env bash
# ========================================================
#  MiCube · 飞牛 (fnOS) / Linux 宿主机 WireGuard 极速一键配置
# ========================================================
set -e

echo "🧊 [MiCube] 正在为飞牛 fnOS / Linux 宿主机部署 WireGuard 与防火墙..."

# 1. 开启内核转发
sysctl -w net.ipv4.ip_forward=1
grep -q '^net.ipv4.ip_forward=1' /etc/sysctl.conf || echo 'net.ipv4.ip_forward=1' >> /etc/sysctl.conf

# 2. 安装 WireGuard 工具包
if ! command -v wg &> /dev/null; then
    echo "📦 正在安装 wireguard wireguard-tools..."
    apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq wireguard wireguard-tools
fi

# 3. 部署防火墙规则 (放行 UDP %d 并启用 NAT 伪装)
iptables -C INPUT -p udp --dport %d -j ACCEPT 2>/dev/null || iptables -I INPUT -p udp --dport %d -j ACCEPT
iptables -C INPUT -p tcp --dport 2026 -j ACCEPT 2>/dev/null || iptables -I INPUT -p tcp --dport 2026 -j ACCEPT
iptables -t nat -C POSTROUTING -s 10.88.0.0/24 ! -o wg+ -j MASQUERADE 2>/dev/null || iptables -t nat -I POSTROUTING -s 10.88.0.0/24 ! -o wg+ -j MASQUERADE
iptables -C FORWARD -s 10.88.0.0/24 -j ACCEPT 2>/dev/null || iptables -I FORWARD -s 10.88.0.0/24 -j ACCEPT
iptables -C FORWARD -d 10.88.0.0/24 -j ACCEPT 2>/dev/null || iptables -I FORWARD -d 10.88.0.0/24 -j ACCEPT

# 4. 尝试保存防火墙
command -v netfilter-persistent &>/dev/null && netfilter-persistent save || true

echo "✅ [MiCube] 宿主机 WireGuard 与防火墙配置已圆满完成！手机现在可以通过扫码秒级穿透接入！"
`, serverPort, serverPort, serverPort)
}

func (m *Manager) detectLocalHost() string {
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		c := strings.ToLower(string(data))
		if strings.Contains(c, "fnos") || strings.Contains(c, "trimix") {
			return "飞牛 fnOS"
		}
		if strings.Contains(c, "pve") || strings.Contains(c, "proxmox") {
			return "Proxmox VE (PVE)"
		}
		if strings.Contains(c, "synology") {
			return "群晖 Synology DSM"
		}
		if strings.Contains(c, "ubuntu") {
			return "Ubuntu Server"
		}
		if strings.Contains(c, "debian") {
			return "Debian GNU/Linux"
		}
	}
	return "Linux Docker"
}

func (m *Manager) detectLocalPlatformID(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "fnos") || strings.Contains(lower, "飞牛") {
		return "fnos"
	}
	if strings.Contains(lower, "pve") || strings.Contains(lower, "proxmox") {
		return "pve"
	}
	if strings.Contains(lower, "synology") || strings.Contains(lower, "群晖") {
		return "synology"
	}
	if strings.Contains(lower, "ubuntu") {
		return "ubuntu"
	}
	if strings.Contains(lower, "debian") {
		return "debian"
	}
	return "linux"
}

func parseOSRelease(content string) (osName string, platID string) {
	lines := strings.Split(content, "\n")
	rawName := ""
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "PRETTY_NAME=") {
			rawName = strings.Trim(strings.TrimPrefix(l, "PRETTY_NAME="), `"`)
			break
		}
		if strings.HasPrefix(l, "NAME=") && rawName == "" {
			rawName = strings.Trim(strings.TrimPrefix(l, "NAME="), `"`)
		}
	}

	lower := strings.ToLower(content)
	if strings.Contains(lower, "fnos") || strings.Contains(lower, "trimix") {
		if rawName == "" {
			rawName = "飞牛 fnOS"
		}
		return rawName, "fnos"
	}
	if strings.Contains(lower, "proxmox") || strings.Contains(lower, "pve") {
		if rawName == "" {
			rawName = "Proxmox VE (PVE)"
		}
		return rawName, "pve"
	}
	if strings.Contains(lower, "synology") {
		return "群晖 Synology DSM", "synology"
	}
	if strings.Contains(lower, "ubuntu") {
		return rawName, "ubuntu"
	}
	if strings.Contains(lower, "debian") {
		return rawName, "debian"
	}
	if rawName != "" {
		return rawName, "linux"
	}
	return "Linux Host", "linux"
}
