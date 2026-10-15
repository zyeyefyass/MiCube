package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SystemStatus struct {
	Version       string  `json:"version"`
	Running       bool    `json:"running"`
	PID           int     `json:"pid"`
	UptimeSeconds int64   `json:"uptime_seconds"`
	CoreMemoryMB  float64 `json:"core_memory_mb"`
	PanelMemoryMB float64 `json:"panel_memory_mb"`
	TotalMemoryMB float64 `json:"total_memory_mb"`
	Platform      string  `json:"platform"`
	PlatformID    string  `json:"platform_id"`
	Arch          string  `json:"arch"`
	Autostart     bool    `json:"autostart"`
	TunMode       string  `json:"tun_mode"`
}

type Supervisor struct {
	RootDir       string
	CtlScript     string
	ConfigPath    string
	CoreAPI       string
	startTime     time.Time
	mu            sync.Mutex
	onStateChange func(bool)

	// PlatformProvider optionally provides live detected OS and version info
	PlatformProvider func() (string, string)
}

func NewSupervisor(rootDir string, cb func(bool)) *Supervisor {
	return &Supervisor{
		RootDir:       rootDir,
		CtlScript:     rootDir + "/container/bin/singcube-ctl",
		ConfigPath:    rootDir + "/etc/config.json",
		CoreAPI:       "http://127.0.0.1:9090",
		startTime:     time.Now(),
		onStateChange: cb,
	}
}

func (s *Supervisor) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := exec.Command(s.CtlScript, "start").Run()
	if err == nil && s.onStateChange != nil {
		s.startTime = time.Now()
		s.onStateChange(true)
	}
	return err
}

func (s *Supervisor) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.onStateChange != nil {
		s.onStateChange(false)
	}
	return exec.Command(s.CtlScript, "stop").Run()
}

func (s *Supervisor) Restart() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := exec.Command(s.CtlScript, "restart").Run()
	if err == nil && s.onStateChange != nil {
		s.startTime = time.Now()
		s.onStateChange(err == nil)
	}
	return err
}

func (s *Supervisor) Reload() error {
	payload, _ := json.Marshal(map[string]string{"path": s.ConfigPath})
	req, err := http.NewRequest(http.MethodPut, s.CoreAPI+"/configs?force=true", bytes.NewReader(payload))
	if err == nil {
		req.Header.Set("Content-Type", "application/json")
		if resp, err := (&http.Client{Timeout: 3 * time.Second}).Do(req); err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
				return nil
			}
		}
	}
	return s.Restart()
}

func (s *Supervisor) SetAutostart(enable bool) error {
	action := "disable"
	if enable {
		action = "enable"
	}
	return exec.Command(s.CtlScript, action).Run()
}

func (s *Supervisor) GetStatus() SystemStatus {
	pid := s.getSingboxPID()
	running := pid > 0
	if s.onStateChange != nil {
		s.onStateChange(running)
	}
	coreMem := getProcessRSSMB(pid)
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	panelMem := float64(ms.Sys) / 1024 / 1024
	_, errAuto := os.Stat(s.RootDir + "/data/autostart")
	var uptime int64
	if running {
		uptime = int64(time.Since(s.startTime).Seconds())
	}
	platName, platID := detectHostPlatform()
	if s.PlatformProvider != nil {
		if pName, pID := s.PlatformProvider(); pName != "" {
			platName = pName
			platID = pID
		}
	}
	ver := "Test V0.91"
	if data, err := os.ReadFile(filepath.Join(s.RootDir, "meta.json")); err == nil {
		var meta struct {
			Version string `json:"version"`
		}
		if json.Unmarshal(data, &meta) == nil && meta.Version != "" {
			ver = meta.Version
		}
	}
	return SystemStatus{
		Version:       ver,
		Running:       running,
		PID:           pid,
		UptimeSeconds: uptime,
		CoreMemoryMB:  coreMem,
		PanelMemoryMB: panelMem,
		TotalMemoryMB: coreMem + panelMem,
		Platform:      platName,
		PlatformID:    platID,
		Arch:          runtime.GOARCH,
		Autostart:     errAuto == nil,
		TunMode:       "TUN SystemStack (sing-box) · 极低延迟 · 40MB+超轻量",
	}
}

func detectHostPlatform() (string, string) {
	if runtime.GOOS == "windows" {
		return "Windows 11 / Server", "windows"
	}
	if runtime.GOOS == "darwin" {
		return "macOS", "apple"
	}

	platEnv := strings.ToLower(os.Getenv("SINGCUBE_PLATFORM"))
	if platEnv == "" {
		platEnv = strings.ToLower(os.Getenv("MICUBE_PLATFORM"))
	}
	if platEnv == "routeros" {
		return "RouterOS v7.x", "routeros"
	}
	hostName := strings.ToLower(os.Getenv("HOSTNAME"))
	if hostName == "chr" || strings.Contains(hostName, "routeros") || strings.Contains(hostName, "mikrotik") {
		return "RouterOS v7.x", "routeros"
	}

	if _, err := os.Stat("/disk1"); err == nil {
		return "RouterOS v7.x", "routeros"
	}
	if data, err := os.ReadFile("/proc/version"); err == nil {
		if strings.Contains(strings.ToLower(string(data)), "mikrotik") {
			return "RouterOS v7.x", "routeros"
		}
	}

	if _, err := os.Stat("/etc/pve"); err == nil {
		return "Proxmox VE (PVE)", "pve"
	}
	if out, err := exec.Command("which", "pveversion").Output(); err == nil && len(out) > 0 {
		return "Proxmox VE (PVE)", "pve"
	}

	if data, err := os.ReadFile("/sys/class/dmi/id/sys_vendor"); err == nil {
		if strings.Contains(strings.ToLower(string(data)), "vmware") {
			return "VMware ESXi", "esxi"
		}
	}
	if data, err := os.ReadFile("/sys/class/dmi/id/product_name"); err == nil {
		if strings.Contains(strings.ToLower(string(data)), "vmware") {
			return "VMware ESXi", "esxi"
		}
	}

	if _, err := os.Stat("/etc/synoinfo.conf"); err == nil {
		return "群晖 Synology DSM", "synology"
	}

	if _, err := os.Stat("/etc/unraid-version"); err == nil {
		return "Unraid OS", "unraid"
	}

	if _, err := os.Stat("/etc/openwrt_release"); err == nil {
		return "OpenWrt Linux", "openwrt"
	}

	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		c := strings.ToLower(string(data))
		if strings.Contains(c, "fnos") || strings.Contains(c, "trimix") {
			return "飞牛 fnOS", "fnos"
		}
		if strings.Contains(c, "pve") || strings.Contains(c, "proxmox") {
			return "Proxmox VE (PVE)", "pve"
		}
		if strings.Contains(c, "ubuntu") {
			return "Ubuntu Server", "ubuntu"
		}
		if strings.Contains(c, "debian") {
			return "Debian GNU/Linux", "debian"
		}
	}

	return "RouterOS v7.x", "routeros"
}

func (s *Supervisor) getSingboxPID() int {
	data, err := os.ReadFile(s.RootDir + "/data/singbox.pid")
	if err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			return pid
		}
	}
	data, err = os.ReadFile(s.RootDir + "/data/sing-box.pid")
	if err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			return pid
		}
	}
	out, err := exec.Command("pidof", "sing-box").Output()
	if err == nil {
		fields := strings.Fields(string(out))
		if len(fields) > 0 {
			if pid, err := strconv.Atoi(fields[0]); err == nil {
				return pid
			}
		}
	}
	return 0
}

func getProcessRSSMB(pid int) float64 {
	if pid <= 0 {
		return 0
	}
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, _ := strconv.ParseFloat(fields[1], 64)
				return kb / 1024.0
			}
		}
	}
	return 0
}
