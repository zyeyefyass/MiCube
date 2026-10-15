package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zyeyefyass/singcube/server/internal/accelerator"
	"github.com/zyeyefyass/singcube/server/internal/adblock"
	"github.com/zyeyefyass/singcube/server/internal/ai"
	"github.com/zyeyefyass/singcube/server/internal/backup"
	"github.com/zyeyefyass/singcube/server/internal/compiler"
	"github.com/zyeyefyass/singcube/server/internal/core"
	"github.com/zyeyefyass/singcube/server/internal/device"
	"github.com/zyeyefyass/singcube/server/internal/dns"
	"github.com/zyeyefyass/singcube/server/internal/host"
	"github.com/zyeyefyass/singcube/server/internal/inspector"
	"github.com/zyeyefyass/singcube/server/internal/ros"
	"github.com/zyeyefyass/singcube/server/internal/telemetry"
	"github.com/zyeyefyass/singcube/server/internal/tgbot"
	"github.com/zyeyefyass/singcube/server/internal/webui"
	"github.com/zyeyefyass/singcube/server/internal/wireguard"
)

// jsonOK writes a JSON success response.
func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(data)
}

// jsonErr writes a JSON error response with the given HTTP status code.
func jsonErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// corsMiddleware adds CORS headers for the Vue dev server and SPA.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	defaultRoot := os.Getenv("MICUBE_ROOT")
	if defaultRoot == "" {
		if _, err := os.Stat("etc"); err == nil {
			defaultRoot = "."
		} else {
			defaultRoot = "/opt/SingCube"
		}
	}
	defaultPort := os.Getenv("PORT")
	if defaultPort == "" {
		defaultPort = "2026"
	}

	flagRoot := flag.String("root", defaultRoot, "Path to SingCube root directory")
	flagPort := flag.String("port", defaultPort, "Port to listen on")
	flagListen := flag.String("listen", "", "Address/port to listen on (e.g. :2026 or 0.0.0.0:2026)")
	flag.Parse()

	rootDir := *flagRoot
	port := *flagPort
	if *flagListen != "" {
		port = strings.TrimPrefix(*flagListen, ":")
	}

	// ── DNS Smart Relay ──
	dnsRelay := dns.NewSmartRelay("0.0.0.0:53", "127.0.0.1:7853")
	if err := dnsRelay.Start(); err != nil {
		log.Printf("[SingCube] DNS relay failed to start: %v (non-fatal)", err)
	}

	// ── Core Supervisor ──
	sup := core.NewSupervisor(rootDir, func(running bool) {
		dnsRelay.SetCoreHealthy(running)
	})
	if _, err := os.Stat(filepath.Join(rootDir, "data", "autostart")); err == nil {
		go func() {
			time.Sleep(1 * time.Second)
			if err := sup.Start(); err != nil {
				log.Printf("[SingCube] Auto-start core error: %v", err)
			} else {
				log.Printf("[SingCube] Core auto-started successfully")
			}
		}()
	}

	// ── Config Compiler ──
	comp := compiler.NewEngine(rootDir)
	_ = comp.CompileAll()

	// ── Route Inspector ──
	routeInspector := inspector.NewInspector(comp)

	// ── Telemetry Manager ──
	dbPath := filepath.Join(rootDir, "data", "singcube.sqlite")
	teleMgr, err := telemetry.NewManager(dbPath)
	if err != nil {
		log.Printf("[SingCube] Telemetry init failed: %v (non-fatal)", err)
	}

	// ── KSpeeder Download & Mirror Accelerator ──
	kspeeder := accelerator.NewKSpeeder(rootDir)

	// ── ROS Manager ──
	rosMgr := ros.NewManager(rootDir)
	rosMgr.URLOptimizer = func(rawURL string) string {
		return kspeeder.OptimizeURL(rawURL)
	}

	// ── AI Copilot ──
	copilot := ai.NewCopilot(rootDir, comp, routeInspector, teleMgr, func() error {
		return sup.Reload()
	})

	// ── WireGuard Manager ──
	wgMgr := wireguard.NewManager(rootDir, rosMgr)

	// ── Host & Docker Companion Manager (fnOS, PVE, Linux) ──
	hostMgr := host.NewManager(rootDir)
	wgMgr.SetHostManager(hostMgr)

	// Platform Provider for live RouterOS and fnOS version reporting
	sup.PlatformProvider = func() (string, string) {
		if conn, plat := rosMgr.GetFullPlatformInfo(); conn && plat != "" {
			return plat, "routeros"
		}
		if hInfo := hostMgr.GetCachedHostInfo(); hInfo != nil && hInfo.OSName != "" && hInfo.Connected {
			return hInfo.OSName, hInfo.PlatformID
		}
		return "", ""
	}

	// ── AdBlock Metrics & Real-time Interceptor ──
	adblockMgr := adblock.NewManager("http://127.0.0.1:9090")

	// ── Backup Manager ──
	backupMgr := backup.NewManager(rootDir)

	// ── LAN Device Discovery & Alias Manager ──
	devMgr := device.NewManager(rootDir, rosMgr, "http://127.0.0.1:9090")

	// Inject live router topology and status into Qwen AI system context
	ai.ContextBuilderFunc = func() string {
		st := sup.GetStatus()
		cState := comp.GetState()
		wgInfo := wgMgr.GetInfo()
		wgCfg, _ := wgInfo["config"].(wireguard.ServerConfig)
		return fmt.Sprintf("- 宿主系统平台: %s (%s)\n- 核心运行状态: %v (PID: %d, Uptime: %ds, 核心内存: %.1fMB)\n- 当前激活配置沙盒: %s (选路算法: %s, 直连真实IP: %v, 广告拦截: %v)\n- 移动穿透 DNS 场景模式: %s\n- WireGuard 接口: %s (公网端点: %s, 授权设备数: %d)\n",
			st.Platform, st.Arch, st.Running, st.PID, st.UptimeSeconds, st.CoreMemoryMB,
			cState.ActiveProfileID, cState.AutoSelectMode, cState.DirectBypassMode, cState.AdBlockEnabled,
			wgCfg.DirectDNSMode,
			wgCfg.InterfaceName, wgInfo["effective_endpoint"], wgInfo["total_peers"],
		)
	}

	// ── Telegram Bot ──
	botSvc := tgbot.NewService(rootDir, sup, comp, routeInspector, teleMgr)
	botSvc.AIChat = func(prompt string) (string, error) {
		return copilot.Chat(prompt)
	}

	// ────────────────────────────────────────────
	// HTTP Router
	// ────────────────────────────────────────────
	mux := http.NewServeMux()

	// ── System Status ──
	mux.HandleFunc("/api/v1/system/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		jsonOK(w, sup.GetStatus())
	})

	// ── Core Control ──
	mux.HandleFunc("/api/v1/core/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		if err := sup.Start(); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"status": "started"})
	})

	mux.HandleFunc("/api/v1/core/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		if err := sup.Stop(); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"status": "stopped"})
	})

	mux.HandleFunc("/api/v1/core/restart", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		if err := sup.Restart(); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"status": "restarted"})
	})

	mux.HandleFunc("/api/v1/core/reload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		if err := sup.Reload(); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"status": "reloaded"})
	})

	// ── Autostart Toggle ──
	mux.HandleFunc("/api/v1/autostart", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			st := sup.GetStatus()
			jsonOK(w, map[string]bool{"autostart": st.Autostart})
		case http.MethodPost:
			var body struct {
				Enable bool `json:"enable"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				jsonErr(w, http.StatusBadRequest, "invalid JSON body")
				return
			}
			if err := sup.SetAutostart(body.Enable); err != nil {
				jsonErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			jsonOK(w, map[string]bool{"autostart": body.Enable})
		default:
			jsonErr(w, http.StatusMethodNotAllowed, "GET or POST only")
		}
	})

	// ── Config State (GET = read, PUT = update & recompile) ──
	mux.HandleFunc("/api/v1/config/state", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jsonOK(w, comp.GetState())
		case http.MethodPut, http.MethodPost:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				jsonErr(w, http.StatusBadRequest, "cannot read body")
				return
			}
			var state compiler.UserState
			if err := json.Unmarshal(body, &state); err != nil {
				jsonErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
				return
			}
			if err := comp.UpdateState(state); err != nil {
				jsonErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			jsonOK(w, map[string]string{"status": "compiled"})
		default:
			jsonErr(w, http.StatusMethodNotAllowed, "GET, PUT or POST only")
		}
	})

	// ── Multi-Profile Sandbox Management ──
	mux.HandleFunc("/api/v1/profiles", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			profs, activeID := comp.GetProfiles()
			jsonOK(w, map[string]any{
				"profiles":  profs,
				"active_id": activeID,
			})
		case http.MethodPost, http.MethodPut:
			var p compiler.Profile
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				jsonErr(w, http.StatusBadRequest, "invalid profile JSON")
				return
			}
			if err := comp.SaveProfile(p); err != nil {
				jsonErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			_ = sup.Reload()
			jsonOK(w, map[string]string{"status": "saved"})
		case http.MethodDelete:
			id := r.URL.Query().Get("id")
			if id == "" {
				jsonErr(w, http.StatusBadRequest, "id parameter required")
				return
			}
			if err := comp.DeleteProfile(id); err != nil {
				jsonErr(w, http.StatusBadRequest, err.Error())
				return
			}
			jsonOK(w, map[string]string{"status": "deleted"})
		default:
			jsonErr(w, http.StatusMethodNotAllowed, "GET, POST, PUT or DELETE only")
		}
	})

	mux.HandleFunc("/api/v1/profiles/switch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var req struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
			jsonErr(w, http.StatusBadRequest, "need {\"id\": \"profile_id\"}")
			return
		}
		if err := comp.SwitchProfile(req.ID); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		_ = sup.Reload()
		jsonOK(w, map[string]string{"status": "switched", "active_id": req.ID})
	})

	// ── Force Recompile ──
	mux.HandleFunc("/api/v1/config/compile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		if err := comp.CompileAll(); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"status": "compiled"})
	})

	// ── Reset Rules to Factory Presets ──
	mux.HandleFunc("/api/v1/config/rules/reset-defaults", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		if err := comp.ResetToDefaultRules(); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		_ = sup.Reload()
		jsonOK(w, map[string]string{"status": "reset", "message": "已成功恢复官方推荐规则预设！"})
	})

	// ── Update Subscriptions ──
	mux.HandleFunc("/api/v1/config/subscriptions/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var body struct {
			ID string `json:"id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.ID != "" {
			if err := comp.FetchSubscription(body.ID); err != nil {
				jsonErr(w, http.StatusInternalServerError, err.Error())
				return
			}
		} else {
			comp.FetchAllSubscriptions()
		}
		_ = comp.CompileAll()
		_ = sup.Reload()
		jsonOK(w, map[string]string{"status": "updated"})
	})

	// ── Backup Export & Import ──
	mux.HandleFunc("/api/v1/backup/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		st := sup.GetStatus()
		data, err := backupMgr.ExportBackup(st.Version)
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		filename := fmt.Sprintf("micube-backup-%s.json", time.Now().Format("20060102-150405"))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		_ = json.NewEncoder(w).Encode(data)
	})

	mux.HandleFunc("/api/v1/backup/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var b backup.BackupData
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			jsonErr(w, http.StatusBadRequest, "无效的 JSON 格式备份文件: "+err.Error())
			return
		}
		if err := backupMgr.ImportBackup(&b); err != nil {
			jsonErr(w, http.StatusBadRequest, err.Error())
			return
		}
		_ = comp.Reload()
		_ = sup.Reload()
		wgMgr.Reload()
		jsonOK(w, map[string]any{
			"status":  "restored",
			"message": "配置恢复成功，核心已自动热重载生效！",
		})
	})

	// ── LAN Devices Discovery & Aliases ──
	mux.HandleFunc("/api/v1/devices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		devices := devMgr.GetDevices()
		jsonOK(w, map[string]any{
			"devices": devices,
			"total":   len(devices),
		})
	})

	mux.HandleFunc("/api/v1/devices/alias", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var req struct {
			IP   string `json:"ip"`
			MAC  string `json:"mac"`
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonErr(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.IP == "" && req.MAC == "" {
			jsonErr(w, http.StatusBadRequest, "至少提供 IP 或 MAC 地址")
			return
		}
		if err := devMgr.SetAlias(req.IP, req.MAC, req.Name); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]any{
			"status": "saved",
			"ip":     req.IP,
			"mac":    req.MAC,
			"name":   req.Name,
		})
	})

	// ── Route Inspector ──
	mux.HandleFunc("/api/v1/inspector/probe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var body struct {
			Host string `json:"host"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Host == "" {
			jsonErr(w, http.StatusBadRequest, "need {\"host\":\"example.com\"}")
			return
		}
		result := routeInspector.Probe(body.Host)
		jsonOK(w, result)
	})

	// ── Mihomo Live Connections & Session Inspector ──
	mux.HandleFunc("/api/v1/connections", func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{Timeout: 5 * time.Second}
		switch r.Method {
		case http.MethodGet:
			resp, err := client.Get("http://127.0.0.1:9090/connections")
			if err != nil {
				jsonErr(w, http.StatusBadGateway, "core not responding: "+err.Error())
				return
			}
			defer resp.Body.Close()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			_, _ = io.Copy(w, resp.Body)
		case http.MethodDelete:
			req, err := http.NewRequest(http.MethodDelete, "http://127.0.0.1:9090/connections", nil)
			if err != nil {
				jsonErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			resp, err := client.Do(req)
			if err != nil {
				jsonErr(w, http.StatusBadGateway, "core not responding: "+err.Error())
				return
			}
			defer resp.Body.Close()
			jsonOK(w, map[string]string{"status": "closed_all"})
		default:
			jsonErr(w, http.StatusMethodNotAllowed, "GET or DELETE only")
		}
	})

	mux.HandleFunc("/api/v1/connections/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			jsonErr(w, http.StatusMethodNotAllowed, "DELETE only")
			return
		}
		connID := strings.TrimPrefix(r.URL.Path, "/api/v1/connections/")
		if connID == "" {
			jsonErr(w, http.StatusBadRequest, "connection ID required")
			return
		}
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequest(http.MethodDelete, "http://127.0.0.1:9090/connections/"+connID, nil)
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			jsonErr(w, http.StatusBadGateway, "core not responding: "+err.Error())
			return
		}
		defer resp.Body.Close()
		jsonOK(w, map[string]string{"status": "closed", "id": connID})
	})

	// ── DNS & Fake-IP Cache Flush ──
	mux.HandleFunc("/api/v1/dns/flush", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		client := &http.Client{Timeout: 3 * time.Second}
		req, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:9090/cache/fakeip/flush", nil)
		if req != nil {
			_, _ = client.Do(req)
		}
		jsonOK(w, map[string]string{"status": "flushed"})
	})

	// ── Telemetry ──
	mux.HandleFunc("/api/v1/telemetry/drilldown", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		q := r.URL.Query()
		direction := q.Get("direction")
		if direction == "" {
			direction = "upload"
		}
		hours, _ := strconv.Atoi(q.Get("hours"))
		if hours <= 0 {
			hours = 24
		}
		dim := q.Get("dimension")
		if dim == "" {
			dim = "rule"
		}
		filterDim := q.Get("filter_dim")
		filterVal := q.Get("filter_val")
		result := teleMgr.QueryDrilldown(direction, hours, dim, filterDim, filterVal)
		jsonOK(w, result)
	})

	mux.HandleFunc("/api/v1/telemetry/nodes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		jsonOK(w, teleMgr.GetNodeTelemetry())
	})

	mux.HandleFunc("/api/v1/telemetry/delay-test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var body struct {
			URL     string `json:"url"`
			Timeout int    `json:"timeout"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.URL == "" {
			body.URL = comp.GetState().SpeedtestURL
		}
		if body.Timeout <= 0 {
			body.Timeout = comp.GetState().SpeedtestTimeout
		}
		teleMgr.TriggerBatchDelayTest(body.URL, body.Timeout)
		jsonOK(w, map[string]string{"status": "triggered", "url": body.URL})
	})

	mux.HandleFunc("/api/v1/telemetry/realtime", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		jsonOK(w, teleMgr.GetRealtimeSpeed())
	})

	// ── Proxy Groups & Node Selection ──
	mux.HandleFunc("/api/v1/proxies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		resp, err := http.Get("http://127.0.0.1:9090/proxies")
		if err != nil {
			jsonErr(w, http.StatusBadGateway, "mihomo core api unavailable: "+err.Error())
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, "read mihomo response error: "+err.Error())
			return
		}

		var rawData struct {
			Proxies map[string]struct {
				Name    string `json:"name"`
				Type    string `json:"type"`
				Now     string `json:"now"`
				All     []string `json:"all"`
				History []struct {
					Time  string `json:"time"`
					Delay int    `json:"delay"`
				} `json:"history"`
				Alive bool `json:"alive"`
				UDP   bool `json:"udp"`
			} `json:"proxies"`
		}
		if err := json.Unmarshal(body, &rawData); err != nil {
			jsonErr(w, http.StatusInternalServerError, "decode mihomo proxies error: "+err.Error())
			return
		}

		type NodeItem struct {
			Name  string `json:"name"`
			Type  string `json:"type"`
			Delay int    `json:"delay"`
			Alive bool   `json:"alive"`
		}

		type GroupItem struct {
			Name     string     `json:"name"`
			Type     string     `json:"type"`
			Now      string     `json:"now"`
			All      []NodeItem `json:"all"`
			IsSelect bool       `json:"is_select"`
		}

		nodeMap := make(map[string]NodeItem)
		totalProxies := 0
		for name, p := range rawData.Proxies {
			ptype := strings.ToLower(p.Type)
			delay := 0
			if len(p.History) > 0 {
				delay = p.History[len(p.History)-1].Delay
			}
			isGroup := ptype == "selector" || ptype == "urltest" || ptype == "fallback" || ptype == "loadbalance" || ptype == "relay"
			if !isGroup && ptype != "direct" && ptype != "reject" {
				totalProxies++
			}
			nodeMap[name] = NodeItem{
				Name:  name,
				Type:  p.Type,
				Delay: delay,
				Alive: p.Alive,
			}
		}

		var groups []GroupItem
		for name, p := range rawData.Proxies {
			ptype := strings.ToLower(p.Type)
			if ptype == "selector" || ptype == "urltest" || ptype == "fallback" || ptype == "loadbalance" || ptype == "relay" {
				var children []NodeItem
				for _, childName := range p.All {
					if n, ok := nodeMap[childName]; ok {
						children = append(children, n)
					} else {
						children = append(children, NodeItem{Name: childName, Type: "Unknown"})
					}
				}
				groups = append(groups, GroupItem{
					Name:     name,
					Type:     p.Type,
					Now:      p.Now,
					All:      children,
					IsSelect: ptype == "selector",
				})
			}
		}

		sortPriority := func(name string) int {
			if name == "节点总览" {
				return 1
			}
			if name == "自动优选" {
				return 2
			}
			if name == "故障转移" {
				return 3
			}
			if strings.HasSuffix(name, "节点") {
				return 10
			}
			if strings.HasSuffix(name, "自动") {
				return 20
			}
			if strings.HasSuffix(name, "故转") || strings.HasSuffix(name, "故障转移") {
				return 30
			}
			if strings.Contains(name, "漏网之鱼") {
				return 50
			}
			return 100
		}

		sort.SliceStable(groups, func(i, j int) bool {
			pi, pj := sortPriority(groups[i].Name), sortPriority(groups[j].Name)
			if pi != pj {
				return pi < pj
			}
			return groups[i].Name < groups[j].Name
		})

		jsonOK(w, map[string]any{
			"groups":        groups,
			"total_proxies": totalProxies,
		})
	})

	mux.HandleFunc("/api/v1/proxies/select", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			jsonErr(w, http.StatusMethodNotAllowed, "POST or PUT only")
			return
		}
		var req struct {
			Group string `json:"group"`
			Name  string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Group == "" || req.Name == "" {
			jsonErr(w, http.StatusBadRequest, "need group and name")
			return
		}

		payload, _ := json.Marshal(map[string]string{"name": req.Name})
		targetURL := fmt.Sprintf("http://127.0.0.1:9090/proxies/%s", url.PathEscape(req.Group))
		httpReq, err := http.NewRequest(http.MethodPut, targetURL, bytes.NewReader(payload))
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			jsonErr(w, http.StatusBadGateway, "failed to contact mihomo core: "+err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			respBody, _ := io.ReadAll(resp.Body)
			jsonErr(w, resp.StatusCode, fmt.Sprintf("mihomo returned %d: %s", resp.StatusCode, string(respBody)))
			return
		}

		jsonOK(w, map[string]any{
			"status": "selected",
			"group":  req.Group,
			"name":   req.Name,
		})
	})

	mux.HandleFunc("/api/v1/proxies/delay", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var req struct {
			Name    string `json:"name"`
			URL     string `json:"url"`
			Timeout int    `json:"timeout"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Name == "" {
			jsonErr(w, http.StatusBadRequest, "need name")
			return
		}
		if req.URL == "" {
			req.URL = comp.GetState().SpeedtestURL
			if req.URL == "" {
				req.URL = "http://www.gstatic.com/generate_204"
			}
		}
		if req.Timeout <= 0 {
			req.Timeout = comp.GetState().SpeedtestTimeout
			if req.Timeout <= 0 {
				req.Timeout = 5000
			}
		}

		targetURL := fmt.Sprintf("http://127.0.0.1:9090/proxies/%s/delay?url=%s&timeout=%d",
			url.PathEscape(req.Name), url.QueryEscape(req.URL), req.Timeout)
		client := &http.Client{Timeout: time.Duration(req.Timeout+2000) * time.Millisecond}
		resp, err := client.Get(targetURL)
		if err != nil {
			jsonOK(w, map[string]any{"name": req.Name, "delay": 0, "error": err.Error()})
			return
		}
		defer resp.Body.Close()

		var delayResp struct {
			Delay int `json:"delay"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&delayResp)
		jsonOK(w, map[string]any{
			"name":  req.Name,
			"delay": delayResp.Delay,
		})
	})

	// ── ROS / Update Manager ──
	mux.HandleFunc("/api/v1/ros/overview", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		jsonOK(w, rosMgr.GetOverview())
	})

	mux.HandleFunc("/api/v1/ros/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jsonOK(w, rosMgr.GetConfig())
		case http.MethodPut:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				jsonErr(w, http.StatusBadRequest, "cannot read body")
				return
			}
			var cfg map[string]string
			if err := json.Unmarshal(body, &cfg); err != nil {
				jsonErr(w, http.StatusBadRequest, "invalid JSON")
				return
			}
			if err := rosMgr.SaveConfig(cfg); err != nil {
				jsonErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			jsonOK(w, map[string]string{"status": "saved"})
		default:
			jsonErr(w, http.StatusMethodNotAllowed, "GET or PUT only")
		}
	})

	// ── ROS Policy Routing & Firewall Matrix ──
	mux.HandleFunc("/api/v1/ros/routing/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		report, err := rosMgr.InspectRouting()
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, report)
	})

	mux.HandleFunc("/api/v1/ros/routing/apply", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var req struct {
			TargetIP string `json:"target_ip"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		res, err := rosMgr.ApplyRouting(req.TargetIP)
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, res)
	})

	mux.HandleFunc("/api/v1/ros/routing/clean", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		res, err := rosMgr.CleanRoutingRules()
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, res)
	})

	mux.HandleFunc("/api/v1/ros/routing/script", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		targetIP := r.URL.Query().Get("target_ip")
		script := rosMgr.GenerateWinBoxScript(targetIP)
		jsonOK(w, map[string]string{"script": script})
	})

	// ── Host & Docker Companion Endpoints (fnOS / Linux) ──
	mux.HandleFunc("/api/v1/host/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jsonOK(w, hostMgr.GetConfig())
		case http.MethodPost:
			var cfg host.HostConfig
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				jsonErr(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
				return
			}
			if err := hostMgr.SaveConfig(cfg); err != nil {
				jsonErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			jsonOK(w, map[string]any{"ok": true, "config": hostMgr.GetConfig()})
		default:
			jsonErr(w, http.StatusMethodNotAllowed, "GET or POST only")
		}
	})

	mux.HandleFunc("/api/v1/host/info", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		info, err := hostMgr.ProbeHost()
		if err != nil && info.RawError == "" {
			info.RawError = err.Error()
		}
		jsonOK(w, info)
	})

	handleHostFirewallApply := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var req struct {
			Mode              string `json:"mode"`
			EnablePenetration *bool  `json:"enable_penetration"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Mode == "" {
			req.Mode = r.URL.Query().Get("mode")
		}
		if req.Mode == "" {
			req.Mode = "permissive"
		}
		enablePenetration := true
		if req.EnablePenetration != nil {
			enablePenetration = *req.EnablePenetration
		} else if r.URL.Query().Get("penetration") == "false" || r.URL.Query().Get("enable_penetration") == "false" {
			enablePenetration = false
		}
		res, err := hostMgr.ApplyFirewall(req.Mode, enablePenetration)
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, res)
	}
	mux.HandleFunc("/api/v1/host/firewall/apply", handleHostFirewallApply)
	mux.HandleFunc("/api/v1/host/firewall", handleHostFirewallApply)

	handleHostFirewallScript := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		mode := r.URL.Query().Get("mode")
		if mode == "" {
			mode = "permissive"
		}
		enablePenetration := true
		if r.URL.Query().Get("penetration") == "false" || r.URL.Query().Get("enable_penetration") == "false" {
			enablePenetration = false
		}
		script := hostMgr.GenerateFirewallScript(mode, enablePenetration)
		jsonOK(w, map[string]any{
			"mode":               mode,
			"enable_penetration": enablePenetration,
			"script":             script,
		})
	}
	mux.HandleFunc("/api/v1/host/firewall/script", handleHostFirewallScript)
	mux.HandleFunc("/api/v1/host/script", handleHostFirewallScript)

	mux.HandleFunc("/api/v1/host/wireguard/deploy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		serverConf := wgMgr.GenerateHostWireGuardConfig()
		out, err := hostMgr.DeployHostWireGuard(serverConf)
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]any{"ok": true, "output": out})
	})

	mux.HandleFunc("/api/v1/host/wireguard/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		conf := wgMgr.GenerateHostWireGuardConfig()
		jsonOK(w, map[string]string{"config": conf})
	})

	mux.HandleFunc("/api/v1/host/wireguard/script", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		info := wgMgr.GetInfo()
		cfg, _ := info["config"].(wireguard.ServerConfig)
		script := hostMgr.GenerateSetupBashScript(rosMgr.GetOutboundIP(), cfg.ListenPort, cfg.PublicKey, "", "", "")
		jsonOK(w, map[string]string{"script": script})
	})

	mux.HandleFunc("/api/v1/host/upgrade/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		out, err := hostMgr.CheckHostUpgrade()
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"output": out})
	})

	mux.HandleFunc("/api/v1/host/upgrade/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		out, err := hostMgr.ExecuteHostUpgrade()
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"output": out})
	})

	// ── AdBlock Live Stats & Blocked Domain Stream ──
	mux.HandleFunc("/api/v1/adblock/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		cState := comp.GetState()
		jsonOK(w, adblockMgr.GetStats(cState.AdBlockEnabled, len(cState.CustomBlockList)))
	})

	mux.HandleFunc("/api/v1/adblock/blocked", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			cState := comp.GetState()
			jsonOK(w, adblockMgr.GetStats(cState.AdBlockEnabled, len(cState.CustomBlockList)))
		case http.MethodDelete:
			adblockMgr.ClearRecords()
			jsonOK(w, map[string]string{"status": "cleared"})
		default:
			jsonErr(w, http.StatusMethodNotAllowed, "GET or DELETE only")
		}
	})

	mux.HandleFunc("/api/v1/ros/upgrade-mihomo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		ver, err := rosMgr.UpgradeMihomoCore(func() error { return sup.Restart() })
		if err != nil {
			if strings.Contains(err.Error(), "already at latest version") {
				jsonOK(w, map[string]string{"upgraded_to": ver, "message": "内核已是最新版本"})
				return
			}
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"upgraded_to": ver})
	})

	mux.HandleFunc("/api/v1/ros/upgrade-firmware", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		if err := rosMgr.TriggerROSFirmwareUpgrade(); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"status": "upgrade_triggered"})
	})

	mux.HandleFunc("/api/v1/system/upgrade-panel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		ver, err := rosMgr.UpgradePanel(func() error {
			log.Printf("[SingCube] Restarting process for upgrade...")
			os.Exit(0) // In container, docker/init will restart container with new binary
			return nil
		})
		if err != nil {
			if strings.Contains(err.Error(), "already at the latest version") {
				jsonOK(w, map[string]string{"status": "already_latest", "version": ver, "message": "MiCube 系统已是最新版本"})
				return
			}
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"status": "upgraded", "version": ver})
	})

	// ── GeoIP / GeoSite & MRS Rule Providers Update ──
	mux.HandleFunc("/api/v1/system/geodata/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		jsonOK(w, rosMgr.GetGeoDataStatus())
	})

	mux.HandleFunc("/api/v1/system/geodata/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		res, err := rosMgr.UpdateGeoData(func() error {
			return sup.Restart()
		})
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, res)
	})

	// ── AI Copilot ──
	mux.HandleFunc("/api/v1/ai/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var body struct {
			Prompt  string `json:"prompt"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			jsonErr(w, http.StatusBadRequest, "invalid json body")
			return
		}
		prompt := body.Prompt
		if prompt == "" {
			prompt = body.Message
		}
		if strings.TrimSpace(prompt) == "" {
			jsonErr(w, http.StatusBadRequest, "need prompt or message")
			return
		}
		reply, err := copilot.Chat(prompt)
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"reply": reply})
	})

	mux.HandleFunc("/api/v1/ai/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jsonOK(w, copilot.GetConfig())
		case http.MethodPost, http.MethodPut:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				jsonErr(w, http.StatusBadRequest, "cannot read body")
				return
			}
			var cfg map[string]any
			if err := json.Unmarshal(body, &cfg); err != nil {
				jsonErr(w, http.StatusBadRequest, "invalid JSON")
				return
			}
			if err := copilot.SaveConfig(cfg); err != nil {
				jsonErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			jsonOK(w, map[string]string{"status": "ok"})
		default:
			jsonErr(w, http.StatusMethodNotAllowed, "GET, POST or PUT only")
		}
	})

	mux.HandleFunc("/api/v1/ai/diagnose", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		reply, err := copilot.Diagnose()
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"diagnosis": reply, "report": reply})
	})

	// ── KSpeeder Download & Mirror Accelerator (iStoreOS 生态强力复刻) ──
	mux.HandleFunc("/api/v1/accelerator/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		jsonOK(w, kspeeder.GetStatus())
	})

	mux.HandleFunc("/api/v1/accelerator/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			jsonErr(w, http.StatusMethodNotAllowed, "POST or PUT only")
			return
		}
		var cfg accelerator.Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			jsonErr(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if err := kspeeder.UpdateConfig(cfg); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"status": "saved"})
	})

	mux.HandleFunc("/api/v1/accelerator/test-github", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		results := kspeeder.TestGitHubProxies()
		active := kspeeder.GetStatus().ActiveGitHubProxy
		jsonOK(w, map[string]any{"proxies": results, "active": active, "best_proxy": active})
	})

	mux.HandleFunc("/api/v1/accelerator/test-docker", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		results := kspeeder.TestDockerMirrors()
		active := kspeeder.GetStatus().ActiveDockerMirror
		jsonOK(w, map[string]any{"mirrors": results, "active": active, "best_mirror": active})
	})

	mux.HandleFunc("/api/v1/accelerator/convert-url", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
			jsonErr(w, http.StatusBadRequest, "need {\"url\": \"...\"}")
			return
		}
		optimized := kspeeder.OptimizeURL(req.URL)
		jsonOK(w, map[string]string{"original": req.URL, "optimized": optimized, "accelerated_url": optimized})
	})

	mux.HandleFunc("/api/v1/accelerator/ros-script", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		jsonOK(w, map[string]string{"script": kspeeder.GenerateROSRegistryScript()})
	})

	mux.HandleFunc("/api/v1/accelerator/docker-daemon", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		daemonJSON := kspeeder.GenerateDockerDaemonJSON()
		jsonOK(w, map[string]string{"json": daemonJSON, "daemon_json": daemonJSON})
	})

	// ── Telegram Bot Config ──
	mux.HandleFunc("/api/v1/tgbot/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			jsonOK(w, botSvc.GetConfig())
		case http.MethodPut:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				jsonErr(w, http.StatusBadRequest, "cannot read body")
				return
			}
			var cfg any
			if err := json.Unmarshal(body, &cfg); err != nil {
				jsonErr(w, http.StatusBadRequest, "invalid JSON")
				return
			}
			if err := botSvc.SaveConfig(cfg); err != nil {
				jsonErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			jsonOK(w, map[string]string{"status": "saved"})
		default:
			jsonErr(w, http.StatusMethodNotAllowed, "GET or PUT only")
		}
	})

	// ── WireGuard Intranet Penetration ──
	mux.HandleFunc("/api/v1/wireguard/info", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		jsonOK(w, wgMgr.GetInfo())
	})

	mux.HandleFunc("/api/v1/wireguard/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			jsonErr(w, http.StatusMethodNotAllowed, "POST or PUT only")
			return
		}
		var cfg wireguard.ServerConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			jsonErr(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if err := wgMgr.UpdateConfig(cfg); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"status": "saved"})
	})

	mux.HandleFunc("/api/v1/wireguard/init-ros", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		if err := wgMgr.InitRouterOSServer(); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"status": "initialized"})
	})

	mux.HandleFunc("/api/v1/wireguard/peers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req struct {
				Name string `json:"name"`
				Mode string `json:"mode"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			peer, conf, err := wgMgr.CreatePeer(req.Name, req.Mode)
			if err != nil {
				jsonErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			jsonOK(w, map[string]any{"peer": peer, "config": conf})
		case http.MethodDelete:
			id := r.URL.Query().Get("id")
			if id == "" {
				var req struct {
					ID string `json:"id"`
				}
				_ = json.NewDecoder(r.Body).Decode(&req)
				id = req.ID
			}
			if id == "" {
				jsonErr(w, http.StatusBadRequest, "id parameter required")
				return
			}
			if err := wgMgr.DeletePeer(id); err != nil {
				jsonErr(w, http.StatusInternalServerError, err.Error())
				return
			}
			jsonOK(w, map[string]string{"status": "deleted"})
		default:
			jsonErr(w, http.StatusMethodNotAllowed, "POST or DELETE only")
		}
	})

	mux.HandleFunc("/api/v1/wireguard/peers/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			jsonErr(w, http.StatusBadRequest, "id parameter required")
			return
		}
		conf, err := wgMgr.GetClientConfig(id)
		if err != nil {
			jsonErr(w, http.StatusNotFound, err.Error())
			return
		}
		jsonOK(w, map[string]string{"id": id, "config": conf})
	})

	mux.HandleFunc("/api/v1/wireguard/ros-script", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		script := wgMgr.GenerateROSScript()
		jsonOK(w, map[string]string{"script": script})
	})

	mux.HandleFunc("/api/v1/wireguard/align/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		report, err := wgMgr.InspectWireGuardAlignment()
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, report)
	})

	mux.HandleFunc("/api/v1/wireguard/align/apply", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		res, err := wgMgr.AlignWireGuardToROS()
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, res)
	})

	mux.HandleFunc("/api/v1/wireguard/peers/rename", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var req struct {
			PublicKey string `json:"public_key"`
			Name      string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonErr(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.PublicKey == "" || req.Name == "" {
			jsonErr(w, http.StatusBadRequest, "public_key and name required")
			return
		}
		if err := wgMgr.RenamePeer(req.PublicKey, req.Name); err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]string{"status": "renamed", "name": req.Name})
	})

	mux.HandleFunc("/api/v1/wireguard/peers/live-status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonErr(w, http.StatusMethodNotAllowed, "GET only")
			return
		}
		id := r.URL.Query().Get("id")
		status, err := wgMgr.GetPeerLiveStatus(id)
		if err != nil {
			jsonErr(w, http.StatusNotFound, err.Error())
			return
		}
		jsonOK(w, status)
	})

	mux.HandleFunc("/api/v1/wireguard/wizard/auto-setup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonErr(w, http.StatusMethodNotAllowed, "POST only")
			return
		}
		var req struct {
			DeviceName string `json:"device_name"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		peer, conf, err := wgMgr.WizardAutoSetup(req.DeviceName)
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, map[string]any{
			"peer":   peer,
			"config": conf,
		})
	})

	// ── Static WebUI (SPA fallback) ──
	mux.Handle("/", webui.Handler())

	// ── Start Server ──
	log.Printf("[SingCube] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("[SingCube]   🧊 MiCube Server v1.0.0")
	log.Printf("[SingCube]   Panel:  http://0.0.0.0:%s", port)
	log.Printf("[SingCube]   API:    http://0.0.0.0:%s/api/v1/", port)
	log.Printf("[SingCube] ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if err := http.ListenAndServe(":"+port, corsMiddleware(mux)); err != nil {
		log.Fatalf("[SingCube] Server failed: %v", err)
	}
}
