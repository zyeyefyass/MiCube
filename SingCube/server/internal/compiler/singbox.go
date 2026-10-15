package compiler

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// Subscription represents an upstream proxy subscription configured by the user.
type Subscription struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"` // "clash", "singbox", etc.
	URL       string `json:"url"`
	Enabled   bool   `json:"enabled"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
	NodeCount int    `json:"node_count,omitempty"`
}

// RuleTarget represents a target domain/IP rule and its outbound policy group.
type RuleTarget struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	MrsName      string `json:"mrs_name"`     // e.g. "openai", "youtube", "netflix", "cn"
	PolicyGroup  string `json:"policy_group"` // e.g. "美国自动", "直连"
	DisplayOrder int    `json:"display_order"`
	MatchOrder   int    `json:"match_order"`
	Enabled      bool   `json:"enabled"`
}

// LatencyThresholds defines latency color grading thresholds in ms.
type LatencyThresholds struct {
	Fast   int `json:"fast"`   // < fast (ms) is Green (default: 120ms)
	Medium int `json:"medium"` // < medium (ms) is Yellow (default: 250ms), >= medium is Red
}

// Profile represents an isolated sandbox configuration environment.
type Profile struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Icon              string            `json:"icon"` // "balance", "shield", "game", "custom"
	SpeedtestURL      string            `json:"speedtest_url"`
	SpeedtestTimeout  int               `json:"speedtest_timeout"`
	Tolerance         int               `json:"tolerance"`
	DirectDNSMode     string            `json:"direct_dns_mode"` // "tunnel_strict" or "local_direct"
	LatencyThresholds LatencyThresholds `json:"latency_thresholds"`
	AutoSelectMode    string            `json:"auto_select_mode"`
	DirectBypassMode  bool              `json:"direct_bypass_mode"`
	AdBlockEnabled    bool              `json:"ad_block_enabled"`
	QUICBlockEnabled  bool              `json:"quic_block_enabled"`
}

// UserState stores the configurable state of SingCube.
type UserState struct {
	ActiveProfileID   string            `json:"active_profile_id"`
	Profiles          []Profile         `json:"profiles"`
	Subscriptions     []Subscription    `json:"subscriptions"`
	AutoCountries     []string          `json:"auto_countries"`
	RuleTargets       []RuleTarget      `json:"rule_targets"`
	AutoSelectMode    string            `json:"auto_select_mode"`   // "url-test" (最低延迟) or "fallback" (故障转移容灾)
	DirectBypassMode  bool              `json:"direct_bypass_mode"` // true: 国内流量直连真实IP不进Fake-IP且不走内核TUN
	AdBlockEnabled    bool              `json:"ad_block_enabled"`   // 全网去广告与隐私防护总开关
	CustomBlockList   []string          `json:"custom_block_list"`  // 用户自定义拦截网址/域名列表
	QUICBlockEnabled  bool              `json:"quic_block_enabled"` // 智能阻断代理 UDP 443 (QUIC) 防海外流媒体卡顿
	SpeedtestURL      string            `json:"speedtest_url,omitempty"`
	SpeedtestTimeout  int               `json:"speedtest_timeout,omitempty"`
	Tolerance         int               `json:"tolerance,omitempty"`
	DirectDNSMode     string            `json:"direct_dns_mode,omitempty"`
	LatencyThresholds LatencyThresholds `json:"latency_thresholds,omitempty"`
}

// Engine compiles user intentions into high-performance sing-box configurations.
type Engine struct {
	RootDir   string
	StateFile string
	EtcDir    string
	DataDir   string
	SubsDir   string
	mu        sync.RWMutex
	State     UserState
}

// NewEngine initialises the config compilation engine.
func NewEngine(rootDir string) *Engine {
	e := &Engine{
		RootDir:   rootDir,
		StateFile: filepath.Join(rootDir, "data", "state.json"),
		EtcDir:    filepath.Join(rootDir, "etc"),
		DataDir:   filepath.Join(rootDir, "data"),
		SubsDir:   filepath.Join(rootDir, "subs"),
	}
	_ = os.MkdirAll(e.EtcDir, 0755)
	_ = os.MkdirAll(e.DataDir, 0755)
	_ = os.MkdirAll(e.SubsDir, 0755)
	e.loadOrInitDefault()
	return e
}

func defaultProfiles() []Profile {
	return []Profile{
		{
			ID:                "prof_balanced",
			Name:              "均衡模式 (标准推荐)",
			Description:       "兼顾全网低延迟与高稳定性，严格防 DNS 泄露，默认过滤恶意广告，适合绝大多数家庭与办公路由环境。",
			Icon:              "balance",
			SpeedtestURL:      "http://www.gstatic.com/generate_204",
			SpeedtestTimeout:  5000,
			Tolerance:         50,
			DirectDNSMode:     "tunnel_strict",
			LatencyThresholds: LatencyThresholds{Fast: 120, Medium: 250},
			AutoSelectMode:    "url-test",
			DirectBypassMode:  true,
			AdBlockEnabled:    true,
			QUICBlockEnabled:  true,
		},
		{
			ID:                "prof_security",
			Name:              "极速隐私 (防DNS泄露)",
			Description:       "强制阻断一切明文 DNS 回退，零明文暴露；开启深度去广告，完全阻止跟踪追踪与运营商劫持。",
			Icon:              "shield",
			SpeedtestURL:      "http://www.gstatic.com/generate_204",
			SpeedtestTimeout:  4000,
			Tolerance:         30,
			DirectDNSMode:     "tunnel_strict",
			LatencyThresholds: LatencyThresholds{Fast: 100, Medium: 200},
			AutoSelectMode:    "url-test",
			DirectBypassMode:  false,
			AdBlockEnabled:    true,
			QUICBlockEnabled:  true,
		},
		{
			ID:                "prof_gaming",
			Name:              "电竞低延迟 (极速直出)",
			Description:       "国内流量绝对直连剥离 Fake-IP，游戏流量优先直出，测速超时放宽，保障高吞吐与最低抖动。",
			Icon:              "game",
			SpeedtestURL:      "http://www.gstatic.com/generate_204",
			SpeedtestTimeout:  3000,
			Tolerance:         20,
			DirectDNSMode:     "local_direct",
			LatencyThresholds: LatencyThresholds{Fast: 80, Medium: 160},
			AutoSelectMode:    "url-test",
			DirectBypassMode:  true,
			AdBlockEnabled:    false,
			QUICBlockEnabled:  false,
		},
	}
}

func defaultRules() []RuleTarget {
	return []RuleTarget{
		{ID: "r_openai", Name: "OpenAI / ChatGPT", MrsName: "openai", PolicyGroup: "美国自动", DisplayOrder: 1, MatchOrder: 1, Enabled: true},
		{ID: "r_anthropic", Name: "Claude / Anthropic", MrsName: "anthropic", PolicyGroup: "美国自动", DisplayOrder: 2, MatchOrder: 2, Enabled: true},
		{ID: "r_youtube", Name: "YouTube 视频", MrsName: "youtube", PolicyGroup: "香港自动", DisplayOrder: 3, MatchOrder: 3, Enabled: true},
		{ID: "r_netflix", Name: "Netflix 奈飞", MrsName: "netflix", PolicyGroup: "新加坡自动", DisplayOrder: 4, MatchOrder: 4, Enabled: true},
		{ID: "r_disney", Name: "Disney+ 迪士尼", MrsName: "disney", PolicyGroup: "新加坡自动", DisplayOrder: 5, MatchOrder: 5, Enabled: true},
		{ID: "r_spotify", Name: "Spotify 音乐", MrsName: "spotify", PolicyGroup: "香港自动", DisplayOrder: 6, MatchOrder: 6, Enabled: true},
		{ID: "r_github", Name: "GitHub 开发者", MrsName: "github", PolicyGroup: "香港自动", DisplayOrder: 7, MatchOrder: 7, Enabled: true},
		{ID: "r_telegram", Name: "Telegram 通讯", MrsName: "telegram", PolicyGroup: "新加坡自动", DisplayOrder: 8, MatchOrder: 8, Enabled: true},
		{ID: "r_twitter", Name: "X / Twitter", MrsName: "twitter", PolicyGroup: "日本自动", DisplayOrder: 9, MatchOrder: 9, Enabled: true},
		{ID: "r_google", Name: "Google 搜索服务", MrsName: "google", PolicyGroup: "香港自动", DisplayOrder: 10, MatchOrder: 10, Enabled: true},
		{ID: "r_steam", Name: "Steam 游戏", MrsName: "steam", PolicyGroup: "直连", DisplayOrder: 11, MatchOrder: 11, Enabled: true},
		{ID: "r_apple", Name: "Apple 苹果服务", MrsName: "apple", PolicyGroup: "直连", DisplayOrder: 12, MatchOrder: 12, Enabled: true},
		{ID: "r_microsoft", Name: "Microsoft 微软服务", MrsName: "microsoft", PolicyGroup: "直连", DisplayOrder: 13, MatchOrder: 13, Enabled: true},
		{ID: "r_cn", Name: "国内网站直连 (直穿真实IP)", MrsName: "cn", PolicyGroup: "直连", DisplayOrder: 14, MatchOrder: 14, Enabled: true},
	}
}

func (e *Engine) loadOrInitDefault() {
	e.mu.Lock()
	defer e.mu.Unlock()

	data, err := os.ReadFile(e.StateFile)
	if err == nil {
		if json.Unmarshal(data, &e.State) == nil {
			if len(e.State.RuleTargets) == 0 {
				e.State.RuleTargets = defaultRules()
			}
			if len(e.State.AutoCountries) == 0 {
				e.State.AutoCountries = []string{"香港", "台湾", "新加坡", "韩国", "日本", "美国"}
			}
			if len(e.State.Profiles) == 0 {
				e.State.Profiles = defaultProfiles()
				e.State.ActiveProfileID = "prof_balanced"
			}
			if e.State.AutoSelectMode == "" {
				e.State.AutoSelectMode = "url-test"
			}
			if e.State.SpeedtestURL == "" {
				e.State.SpeedtestURL = "http://www.gstatic.com/generate_204"
			}
			if e.State.SpeedtestTimeout <= 0 {
				e.State.SpeedtestTimeout = 5000
			}
			if e.State.Tolerance <= 0 {
				e.State.Tolerance = 50
			}
			if e.State.DirectDNSMode == "" {
				e.State.DirectDNSMode = "tunnel_strict"
			}
			if e.State.LatencyThresholds.Fast <= 0 {
				e.State.LatencyThresholds = LatencyThresholds{Fast: 120, Medium: 250}
			}
			return
		}
	}

	e.State = UserState{
		ActiveProfileID:   "prof_balanced",
		Profiles:          defaultProfiles(),
		Subscriptions:     []Subscription{},
		AutoCountries:     []string{"香港", "台湾", "新加坡", "韩国", "日本", "美国"},
		RuleTargets:       defaultRules(),
		AutoSelectMode:    "url-test",
		DirectBypassMode:  true,
		AdBlockEnabled:    true,
		CustomBlockList:   []string{},
		QUICBlockEnabled:  true,
		SpeedtestURL:      "http://www.gstatic.com/generate_204",
		SpeedtestTimeout:  5000,
		Tolerance:         50,
		DirectDNSMode:     "tunnel_strict",
		LatencyThresholds: LatencyThresholds{Fast: 120, Medium: 250},
	}
	_ = e.saveLocked()
}

func (e *Engine) ResetToDefaultRules() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.State.RuleTargets = defaultRules()
	if err := e.saveLocked(); err != nil {
		return err
	}
	e.mu.Unlock()
	err := e.CompileAll()
	e.mu.Lock()
	return err
}

func (e *Engine) saveLocked() error {
	data, err := json.MarshalIndent(e.State, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(e.StateFile, data, 0644)
}

func (e *Engine) Reload() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return nil
}

func (e *Engine) GetState() UserState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.State
}

func (e *Engine) GetProfiles() ([]Profile, string) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.State.Profiles, e.State.ActiveProfileID
}

func (e *Engine) SwitchProfile(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var target *Profile
	for i := range e.State.Profiles {
		if e.State.Profiles[i].ID == id {
			target = &e.State.Profiles[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("profile %s not found", id)
	}

	e.State.ActiveProfileID = target.ID
	e.State.SpeedtestURL = target.SpeedtestURL
	e.State.SpeedtestTimeout = target.SpeedtestTimeout
	e.State.Tolerance = target.Tolerance
	e.State.DirectDNSMode = target.DirectDNSMode
	e.State.LatencyThresholds = target.LatencyThresholds
	e.State.AutoSelectMode = target.AutoSelectMode
	e.State.DirectBypassMode = target.DirectBypassMode
	e.State.AdBlockEnabled = target.AdBlockEnabled
	e.State.QUICBlockEnabled = target.QUICBlockEnabled

	if err := e.saveLocked(); err != nil {
		return err
	}

	e.mu.Unlock()
	err := e.CompileAll()
	e.mu.Lock()
	return err
}

func (e *Engine) SaveProfile(p Profile) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if p.ID == "" {
		p.ID = fmt.Sprintf("prof_%d", time.Now().Unix())
	}

	found := false
	for i := range e.State.Profiles {
		if e.State.Profiles[i].ID == p.ID {
			e.State.Profiles[i] = p
			found = true
			break
		}
	}
	if !found {
		e.State.Profiles = append(e.State.Profiles, p)
	}

	if e.State.ActiveProfileID == p.ID {
		e.State.SpeedtestURL = p.SpeedtestURL
		e.State.SpeedtestTimeout = p.SpeedtestTimeout
		e.State.Tolerance = p.Tolerance
		e.State.DirectDNSMode = p.DirectDNSMode
		e.State.LatencyThresholds = p.LatencyThresholds
		e.State.AutoSelectMode = p.AutoSelectMode
		e.State.DirectBypassMode = p.DirectBypassMode
		e.State.AdBlockEnabled = p.AdBlockEnabled
		e.State.QUICBlockEnabled = p.QUICBlockEnabled
	}

	if err := e.saveLocked(); err != nil {
		return err
	}

	if e.State.ActiveProfileID == p.ID {
		e.mu.Unlock()
		err := e.CompileAll()
		e.mu.Lock()
		return err
	}
	return nil
}

func (e *Engine) DeleteProfile(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if id == e.State.ActiveProfileID {
		return fmt.Errorf("cannot delete active profile")
	}

	idx := -1
	for i := range e.State.Profiles {
		if e.State.Profiles[i].ID == id {
			idx = i
			break
		}
	}
	if idx >= 0 {
		e.State.Profiles = append(e.State.Profiles[:idx], e.State.Profiles[idx+1:]...)
	}
	return e.saveLocked()
}

func (e *Engine) UpdateState(s UserState) error {
	e.mu.Lock()
	e.State = s
	if err := e.saveLocked(); err != nil {
		e.mu.Unlock()
		return err
	}
	e.mu.Unlock()
	return e.CompileAll()
}

func isInvalidProxy(p map[string]any) bool {
	name, _ := p["name"].(string)
	upper := strings.ToUpper(name)
	dummyKeywords := []string{
		"剩余流量", "到期时间", "距离下次重置", "套餐到期", "官网", "官方网站",
		"防失联", "群组", "公告", "Traffic", "Expire", "Reset", "Channel", "Notice",
	}
	for _, kw := range dummyKeywords {
		if strings.Contains(upper, strings.ToUpper(kw)) {
			return true
		}
	}
	server, _ := p["server"].(string)
	if server == "" || server == "127.0.0.1" || server == "0.0.0.0" {
		return true
	}
	return false
}

func sanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	var b strings.Builder
	for i, r := range s {
		if r == utf8.RuneError {
			_, size := utf8.DecodeRuneInString(s[i:])
			if size == 1 {
				continue
			}
		}
		b.WriteRune(r)
	}
	return b.String()
}

// clashProxyToSingboxOutbound transforms a Clash YAML proxy map into a sing-box standard outbound.
func clashProxyToSingboxOutbound(p map[string]any) (map[string]any, bool) {
	name, _ := p["name"].(string)
	name = strings.TrimSpace(sanitizeUTF8(name))
	if name == "" {
		return nil, false
	}

	pType, _ := p["type"].(string)
	pType = strings.ToLower(pType)
	server, _ := p["server"].(string)
	port := 0
	if rawPort, ok := p["port"]; ok {
		switch v := rawPort.(type) {
		case int:
			port = v
		case float64:
			port = int(v)
		case string:
			port, _ = strconv.Atoi(v)
		}
	}

	switch pType {
	case "ss", "shadowsocks":
		cipher, _ := p["cipher"].(string)
		password, _ := p["password"].(string)
		return map[string]any{
			"type":        "shadowsocks",
			"tag":         name,
			"server":      server,
			"server_port": port,
			"method":      cipher,
			"password":    password,
		}, true

	case "vmess":
		uuid, _ := p["uuid"].(string)
		out := map[string]any{
			"type":        "vmess",
			"tag":         name,
			"server":      server,
			"server_port": port,
			"uuid":        uuid,
			"security":    "auto",
			"alter_id":    0,
		}
		if tlsVal, ok := p["tls"].(bool); ok && tlsVal {
			sni, _ := p["servername"].(string)
			if sni == "" {
				sni, _ = p["sni"].(string)
			}
			out["tls"] = map[string]any{
				"enabled":     true,
				"server_name": sni,
			}
		}
		if net, ok := p["network"].(string); ok && net == "ws" {
			wsPath := "/"
			if wsOpts, ok := p["ws-opts"].(map[string]any); ok {
				if pth, ok := wsOpts["path"].(string); ok {
					wsPath = pth
				}
			}
			out["transport"] = map[string]any{
				"type": "ws",
				"path": wsPath,
			}
		}
		return out, true

	case "vless":
		uuid, _ := p["uuid"].(string)
		flow, _ := p["flow"].(string)
		out := map[string]any{
			"type":        "vless",
			"tag":         name,
			"server":      server,
			"server_port": port,
			"uuid":        uuid,
			"flow":        flow,
		}
		tlsMap := map[string]any{"enabled": true}
		if sni, ok := p["servername"].(string); ok && sni != "" {
			tlsMap["server_name"] = sni
		}
		if realityOpts, ok := p["reality-opts"].(map[string]any); ok {
			tlsMap["reality"] = map[string]any{
				"enabled":    true,
				"public_key": realityOpts["public-key"],
				"short_id":   realityOpts["short-id"],
			}
		}
		out["tls"] = tlsMap
		return out, true

	case "trojan":
		password, _ := p["password"].(string)
		out := map[string]any{
			"type":        "trojan",
			"tag":         name,
			"server":      server,
			"server_port": port,
			"password":    password,
		}
		tlsMap := map[string]any{"enabled": true}
		if sni, ok := p["sni"].(string); ok && sni != "" {
			tlsMap["server_name"] = sni
		}
		out["tls"] = tlsMap
		return out, true

	case "hysteria2", "hy2":
		password, _ := p["password"].(string)
		out := map[string]any{
			"type":        "hysteria2",
			"tag":         name,
			"server":      server,
			"server_port": port,
			"password":    password,
		}
		tlsMap := map[string]any{"enabled": true}
		if sni, ok := p["sni"].(string); ok && sni != "" {
			tlsMap["server_name"] = sni
		}
		out["tls"] = tlsMap
		return out, true

	case "wireguard":
		out := map[string]any{
			"type":        "wireguard",
			"tag":         name,
			"server":      server,
			"server_port": port,
		}
		if pk, ok := p["public-key"].(string); ok {
			out["peer_public_key"] = pk
		}
		if sk, ok := p["private-key"].(string); ok {
			out["private_key"] = sk
		}
		return out, true
	}

	return nil, false
}

// FetchSubscription downloads a subscription URL and caches it locally.
func (e *Engine) FetchSubscription(subID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var target *Subscription
	for i := range e.State.Subscriptions {
		if e.State.Subscriptions[i].ID == subID {
			target = &e.State.Subscriptions[i]
			break
		}
	}
	if target == nil || target.URL == "" {
		return fmt.Errorf("subscription %s not found or empty URL", subID)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err := http.NewRequest(http.MethodGet, target.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "sing-box/1.11.0; SingCube-Gateway")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download subscription failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 1. Try parse as Clash YAML
	var clashParsed struct {
		Proxies []map[string]any `yaml:"proxies"`
	}
	var validSingboxOutbounds []map[string]any

	if err := yaml.Unmarshal(body, &clashParsed); err == nil && len(clashParsed.Proxies) > 0 {
		for _, p := range clashParsed.Proxies {
			if isInvalidProxy(p) {
				continue
			}
			if ob, ok := clashProxyToSingboxOutbound(p); ok {
				validSingboxOutbounds = append(validSingboxOutbounds, ob)
			}
		}
	} else {
		// 2. Try parse as sing-box JSON
		var sbParsed struct {
			Outbounds []map[string]any `json:"outbounds"`
		}
		if err := json.Unmarshal(body, &sbParsed); err == nil && len(sbParsed.Outbounds) > 0 {
			for _, ob := range sbParsed.Outbounds {
				t, _ := ob["type"].(string)
				if t == "direct" || t == "block" || t == "dns" || t == "selector" || t == "urltest" {
					continue
				}
				validSingboxOutbounds = append(validSingboxOutbounds, ob)
			}
		} else {
			// 3. Fallback: Base64 decode check
			if rawDec, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(body))); err == nil {
				lines := strings.Split(string(rawDec), "\n")
				for _, l := range lines {
					l = strings.TrimSpace(l)
					if strings.HasPrefix(l, "ss://") || strings.HasPrefix(l, "vmess://") || strings.HasPrefix(l, "vless://") {
						// Recorded
					}
				}
			}
		}
	}

	cacheFile := filepath.Join(e.SubsDir, target.ID+".json")
	cacheData, _ := json.MarshalIndent(map[string]any{
		"outbounds": validSingboxOutbounds,
	}, "", "  ")
	if err := os.WriteFile(cacheFile, cacheData, 0644); err != nil {
		return err
	}

	target.UpdatedAt = time.Now().Unix()
	target.NodeCount = len(validSingboxOutbounds)
	_ = e.saveLocked()
	log.Printf("[SingCube Compiler] Subscription %s updated, loaded %d nodes", target.Name, len(validSingboxOutbounds))
	return nil
}

// FetchAllSubscriptions fetches all enabled subscriptions.
func (e *Engine) FetchAllSubscriptions() {
	e.mu.RLock()
	var ids []string
	for _, sub := range e.State.Subscriptions {
		if sub.Enabled && sub.URL != "" {
			ids = append(ids, sub.ID)
		}
	}
	e.mu.RUnlock()

	for _, id := range ids {
		if err := e.FetchSubscription(id); err != nil {
			log.Printf("[SingCube Compiler] Fetch subscription %s error: %v", id, err)
		}
	}
}

// CompileAll generates sing-box config.json with full 5-hub capabilities.
func (e *Engine) CompileAll() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var allNodeOutbounds []map[string]any
	var allNodeNames []string
	seenNames := make(map[string]bool)

	for i := range e.State.Subscriptions {
		sub := &e.State.Subscriptions[i]
		if !sub.Enabled {
			continue
		}
		cacheFile := filepath.Join(e.SubsDir, sub.ID+".json")
		data, err := os.ReadFile(cacheFile)
		if err != nil {
			// Fallback try .yaml if previously downloaded
			yamlFile := filepath.Join(e.SubsDir, sub.ID+".yaml")
			yData, yErr := os.ReadFile(yamlFile)
			if yErr == nil {
				var parsed struct {
					Proxies []map[string]any `yaml:"proxies"`
				}
				if yaml.Unmarshal(yData, &parsed) == nil {
					for _, p := range parsed.Proxies {
						if isInvalidProxy(p) {
							continue
						}
						if ob, ok := clashProxyToSingboxOutbound(p); ok {
							name, _ := ob["tag"].(string)
							if !seenNames[name] {
								seenNames[name] = true
								allNodeOutbounds = append(allNodeOutbounds, ob)
								allNodeNames = append(allNodeNames, name)
							}
						}
					}
				}
			}
			continue
		}

		var parsed struct {
			Outbounds []map[string]any `json:"outbounds"`
		}
		if json.Unmarshal(data, &parsed) == nil {
			for _, ob := range parsed.Outbounds {
				name, _ := ob["tag"].(string)
				name = strings.TrimSpace(sanitizeUTF8(name))
				if name != "" && !seenNames[name] {
					seenNames[name] = true
					ob["tag"] = name
					allNodeOutbounds = append(allNodeOutbounds, ob)
					allNodeNames = append(allNodeNames, name)
				}
			}
		}
		sub.NodeCount = len(allNodeOutbounds)
	}
	_ = e.saveLocked()

	// 2. Classify nodes by country
	primaryCountries := []string{"香港", "台湾", "新加坡", "韩国", "日本", "美国"}
	countryNodes := make(map[string][]string)
	for _, c := range primaryCountries {
		countryNodes[c] = []string{}
	}
	countryNodes["其它"] = []string{}
	for _, c := range e.State.AutoCountries {
		c = strings.TrimSpace(c)
		if c != "" {
			if _, exists := countryNodes[c]; !exists {
				countryNodes[c] = []string{}
			}
		}
	}

	primaryKeywords := map[string][]string{
		"香港":  {"香港", "HK", "Hong Kong", "HongKong", "🇭🇰"},
		"台湾":  {"台湾", "TW", "Taiwan", "TaiWan", "🇹🇼"},
		"新加坡": {"新加坡", "SG", "Singapore", "狮城", "🇸🇬"},
		"韩国":  {"韩国", "KR", "Korea", "Seoul", "🇰🇷"},
		"日本":  {"日本", "JP", "Japan", "Tokyo", "Osaka", "🇯🇵"},
		"美国":  {"美国", "US", "United States", "USA", "America", "🇺🇸"},
	}

	for _, name := range allNodeNames {
		upper := strings.ToUpper(name)
		matchedPrimary := false
		for _, country := range primaryCountries {
			keywords := primaryKeywords[country]
			matched := false
			for _, kw := range keywords {
				if strings.Contains(upper, strings.ToUpper(kw)) {
					matched = true
					break
				}
			}
			if matched {
				countryNodes[country] = append(countryNodes[country], name)
				matchedPrimary = true
				break
			}
		}
		if !matchedPrimary {
			countryNodes["其它"] = append(countryNodes["其它"], name)
		}
	}

	// 3. Assemble Policy Groups (outbounds)
	var outbounds []map[string]any

	// 3.1 Base Outbounds (sing-box 1.14.3: legacy block/dns special outbounds removed, replaced by rule actions)
	outbounds = append(outbounds,
		map[string]any{"type": "direct", "tag": "direct"},
		map[string]any{"type": "direct", "tag": "直连"},
	)

	// Fallback pool when a country group is empty
	fallbackNodes := allNodeNames
	if len(fallbackNodes) == 0 {
		fallbackNodes = []string{"direct"}
	}

	speedURL := e.State.SpeedtestURL
	if speedURL == "" {
		speedURL = "http://www.gstatic.com/generate_204"
	}
	tolerance := e.State.Tolerance
	if tolerance <= 0 {
		tolerance = 50
	}

	// 3.2 Regional URL-Test Groups
	var majorAreaGroups []string
	standardRegions := []string{"香港", "台湾", "新加坡", "韩国", "日本", "美国", "其它"}
	for _, r := range standardRegions {
		nodes := countryNodes[r]
		if len(nodes) == 0 {
			nodes = fallbackNodes
		}
		autoName := r + "自动"
		outbounds = append(outbounds, map[string]any{
			"type":      "urltest",
			"tag":       autoName,
			"outbounds": nodes,
			"url":       speedURL,
			"interval":  "5m",
			"tolerance": tolerance,
		})
		majorAreaGroups = append(majorAreaGroups, autoName)
	}

	// 3.3 Main Selector Groups
	nodeSelectGroup := append([]string{}, allNodeNames...)
	if len(nodeSelectGroup) == 0 {
		nodeSelectGroup = []string{"direct"}
	} else {
		nodeSelectGroup = append(nodeSelectGroup, "direct")
	}
	outbounds = append(outbounds, map[string]any{
		"type":      "selector",
		"tag":       "节点选择",
		"outbounds": nodeSelectGroup,
	})

	proxyOutbounds := append([]string{}, majorAreaGroups...)
	proxyOutbounds = append(proxyOutbounds, "节点选择", "direct")
	outbounds = append(outbounds, map[string]any{
		"type":      "selector",
		"tag":       "🚀 国外流量",
		"outbounds": proxyOutbounds,
	})

	outbounds = append(outbounds, map[string]any{
		"type":      "selector",
		"tag":       "🐟 漏网之鱼",
		"outbounds": []string{"🚀 国外流量", "direct", "节点选择"},
	})

	outbounds = append(outbounds, map[string]any{
		"type":      "selector",
		"tag":       "🛡️ 广告拦截",
		"outbounds": []string{"direct"},
	})

	// 3.4 Business Selector Groups
	businessGroups := []struct {
		Tag      string
		Default  string
		Fallback []string
	}{
		{"🎬 奈飞流媒体", "新加坡自动", []string{"新加坡自动", "香港自动", "美国自动", "🚀 国外流量", "节点选择"}},
		{"🎵 Spotify", "香港自动", []string{"香港自动", "美国自动", "🚀 国外流量", "节点选择"}},
		{"🎮 游戏平台", "直连", []string{"直连", "香港自动", "日本自动", "🚀 国外流量"}},
		{"🍎 苹果服务", "直连", []string{"直连", "🚀 国外流量", "香港自动"}},
		{"💼 微软服务", "直连", []string{"直连", "🚀 国外流量", "香港自动"}},
	}
	for _, bg := range businessGroups {
		outbounds = append(outbounds, map[string]any{
			"type":      "selector",
			"tag":       bg.Tag,
			"outbounds": bg.Fallback,
			"default":   bg.Default,
		})
	}

	// 3.5 Append all genuine proxy nodes
	for _, node := range allNodeOutbounds {
		outbounds = append(outbounds, node)
	}

	// 4. Construct Routes
	var routeRules []map[string]any

	// 4.1 Hijack DNS traffic (sing-box 1.14.3 action: hijack-dns)
	routeRules = append(routeRules,
		map[string]any{"protocol": "dns", "action": "hijack-dns"},
		map[string]any{"port": 53, "action": "hijack-dns"},
	)

	// 4.2 LAN & Private bypass
	routeRules = append(routeRules,
		map[string]any{"ip_is_private": true, "outbound": "direct"},
		map[string]any{
			"domain_suffix": []string{
				".lan", ".local", ".home.arpa", ".localhost", "localhost.weixin.qq.com",
			},
			"outbound": "direct",
		},
	)

	// 4.3 Custom block list
	if len(e.State.CustomBlockList) > 0 {
		routeRules = append(routeRules, map[string]any{
			"domain_suffix": e.State.CustomBlockList,
			"action":        "reject",
		})
	}

	// 4.4 Ad block rules
	if e.State.AdBlockEnabled {
		routeRules = append(routeRules, map[string]any{
			"rule_set": "geosite-category-ads-all",
			"action": "reject",
		})
	}

	// 4.5 QUIC block rules
	if e.State.QUICBlockEnabled {
		routeRules = append(routeRules, map[string]any{
			"port":     443,
			"network":  "udp",
			"rule_set": "geosite-geolocation-!cn",
			"action": "reject",
		})
	}

	// 4.6 User configured rule targets
	sortedRules := make([]RuleTarget, len(e.State.RuleTargets))
	copy(sortedRules, e.State.RuleTargets)
	sort.Slice(sortedRules, func(i, j int) bool {
		return sortedRules[i].MatchOrder < sortedRules[j].MatchOrder
	})

	for _, rt := range sortedRules {
		if !rt.Enabled {
			continue
		}
		targetOutbound := strings.TrimSpace(rt.PolicyGroup)
		if targetOutbound == "" || targetOutbound == "直连" || targetOutbound == "DIRECT" {
			targetOutbound = "direct"
		} else if targetOutbound == "拦截" || targetOutbound == "REJECT" {
			targetOutbound = "block"
		}

		mrsLower := strings.ToLower(rt.MrsName)
		if targetOutbound == "block" {
			if mrsLower == "cn" {
				routeRules = append(routeRules,
					map[string]any{"rule_set": "geosite-cn", "action": "reject"},
					map[string]any{"rule_set": "geoip-cn", "action": "reject"},
				)
			} else if rt.MrsName != "" {
				routeRules = append(routeRules, map[string]any{
					"rule_set": "geosite-" + rt.MrsName,
					"action":   "reject",
				})
			}
		} else {
			if mrsLower == "cn" {
				routeRules = append(routeRules,
					map[string]any{"rule_set": "geosite-cn", "outbound": targetOutbound},
					map[string]any{"rule_set": "geoip-cn", "outbound": targetOutbound},
				)
			} else if rt.MrsName != "" {
				routeRules = append(routeRules, map[string]any{
					"rule_set": "geosite-" + rt.MrsName,
					"outbound": targetOutbound,
				})
			}
		}
	}

	// 4.7 Domestic Direct bypass (Real IP)
	routeRules = append(routeRules,
		map[string]any{"rule_set": "geosite-cn", "outbound": "direct"},
		map[string]any{"rule_set": "geoip-cn", "outbound": "direct"},
	)

	// 5. Construct DNS configuration (sing-box 1.14.3 standard: typed servers, optimistic cache, domain_resolver)
	dnsConfig := map[string]any{
		"servers": []map[string]any{
			{
				"tag":    "dns_direct",
				"type":   "udp",
				"server": "223.5.5.5",
				"detour": "direct",
			},
			{
				"tag":             "dns_proxy",
				"type":            "https",
				"server":          "1.1.1.1",
				"path":            "/dns-query",
				"domain_resolver": "dns_direct",
				"detour":          "🚀 国外流量",
			},
		},
		"rules": []map[string]any{
			{
				"outbound": "any",
				"server":   "dns_direct",
			},
			{
				"rule_set": "geosite-category-ads-all",
				"action":   "reject",
			},
			{
				"rule_set": "geosite-cn",
				"server":   "dns_direct",
			},
		},
		"final":      "dns_proxy",
		"strategy":   "prefer_ipv4",
		"optimistic": true,
	}

	// 6. Inbounds (sing-box 1.14.3 standard: unified address array & system stack)
	inbounds := []map[string]any{
		{
			"type":           "tun",
			"tag":            "tun-in",
			"interface_name": "singcube-tun",
			"address": []string{
				"172.19.0.1/30",
				"fdfe:dcba:9876::1/126",
			},
			"mtu":          9000,
			"auto_route":   true,
			"strict_route": false,
			"stack":        "system",
			"sniff":        true,
		},
		{
			"type":        "mixed",
			"tag":         "mixed-in",
			"listen":      "::",
			"listen_port": 7890,
			"sniff":       true,
		},
	}

	// 7. Rule Sets
	ruleSets := []map[string]any{
		{
			"type":            "remote",
			"tag":             "geosite-category-ads-all",
			"format":          "binary",
			"url":             "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-category-ads-all.srs",
			"download_detour": "🚀 国外流量",
		},
		{
			"type":            "remote",
			"tag":             "geosite-cn",
			"format":          "binary",
			"url":             "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-cn.srs",
			"download_detour": "🚀 国外流量",
		},
		{
			"type":            "remote",
			"tag":             "geoip-cn",
			"format":          "binary",
			"url":             "https://raw.githubusercontent.com/SagerNet/sing-geoip/rule-set/geoip-cn.srs",
			"download_detour": "🚀 国外流量",
		},
		{
			"type":            "remote",
			"tag":             "geosite-geolocation-!cn",
			"format":          "binary",
			"url":             "https://raw.githubusercontent.com/SagerNet/sing-geosite/rule-set/geosite-geolocation-!cn.srs",
			"download_detour": "🚀 国外流量",
		},
	}

	// 8. Assemble Full sing-box JSON configuration
	singboxConfig := map[string]any{
		"log": map[string]any{
			"disabled":  false,
			"level":     "warn",
			"timestamp": true,
		},
		"dns":       dnsConfig,
		"inbounds":  inbounds,
		"outbounds": outbounds,
		"route": map[string]any{
			"rules":                  routeRules,
			"rule_set":               ruleSets,
			"final":                  "🐟 漏网之鱼",
			"auto_detect_interface":  true,
			"default_http_client":    "rule-client",
		},
		"http_clients": []map[string]any{
			{
				"tag": "rule-client",
				"dial": map[string]any{
					"detour": "🚀 国外流量",
				},
			},
		},
		"experimental": map[string]any{
			"clash_api": map[string]any{
				"external_controller": "0.0.0.0:9090",
				"external_ui":         "",
				"secret":              "",
				"default_mode":        "rule",
			},
			"cache_file": map[string]any{
				"enabled":   true,
				"path":      "cache.db",
				"store_dns": true,
			},
		},
	}

	configJSON, err := json.MarshalIndent(singboxConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sing-box config failed: %w", err)
	}

	destConfig := filepath.Join(e.EtcDir, "config.json")
	if err := os.WriteFile(destConfig, configJSON, 0644); err != nil {
		return fmt.Errorf("write config.json failed: %w", err)
	}

	log.Printf("[SingCube Compiler] Successfully compiled %d outbounds into %s", len(outbounds), destConfig)
	return nil
}
