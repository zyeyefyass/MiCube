package tgbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zyeyefyass/singcube/server/internal/compiler"
	"github.com/zyeyefyass/singcube/server/internal/core"
	"github.com/zyeyefyass/singcube/server/internal/inspector"
	"github.com/zyeyefyass/singcube/server/internal/telemetry"
)

// BotConfig stores Telegram Bot settings.
type BotConfig struct {
	Enabled       bool    `json:"enabled"`
	Token         string  `json:"token"`
	AllowedChatIDs []int64 `json:"allowed_chat_ids"`
	ProxyURL      string  `json:"proxy_url"` // e.g. "http://127.0.0.1:7890"
}

// Service manages the Telegram Bot long-polling and command handler.
type Service struct {
	rootDir    string
	configFile string
	config     BotConfig
	mu         sync.RWMutex
	stopChan   chan struct{}
	running    bool

	supervisor *core.Supervisor
	compiler   *compiler.Engine
	inspector  *inspector.Inspector
	telemetry  *telemetry.Manager

	AIChat func(prompt string) (string, error)
}

// NewService creates a new Telegram Bot service.
func NewService(rootDir string, sup *core.Supervisor, comp *compiler.Engine, ins *inspector.Inspector, tele *telemetry.Manager) *Service {
	s := &Service{
		rootDir:    rootDir,
		configFile: filepath.Join(rootDir, "data", "tgbot-config.json"),
		supervisor: sup,
		compiler:   comp,
		inspector:  ins,
		telemetry:  tele,
		config: BotConfig{
			Enabled:       false,
			Token:         "",
			AllowedChatIDs: []int64{},
			ProxyURL:      "http://127.0.0.1:7890",
		},
	}
	s.loadConfig()
	if s.config.Enabled && s.config.Token != "" {
		s.Start()
	}
	return s
}

// Start begins the Telegram update polling loop in the background.
func (s *Service) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.stopChan = make(chan struct{})
	s.running = true
	s.mu.Unlock()

	go s.pollLoop()
	log.Printf("[TgBot] Bot service started")
}

// Stop terminates the polling loop.
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return
	}
	close(s.stopChan)
	s.running = false
	log.Printf("[TgBot] Bot service stopped")
}

// GetConfig returns current config with token masked.
func (s *Service) GetConfig() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	maskedToken := ""
	if s.config.Token != "" {
		if len(s.config.Token) > 10 {
			maskedToken = s.config.Token[:6] + "..." + s.config.Token[len(s.config.Token)-4:]
		} else {
			maskedToken = "****"
		}
	}

	return map[string]any{
		"enabled":          s.config.Enabled,
		"token":            maskedToken,
		"has_token":        s.config.Token != "",
		"allowed_chat_ids": s.config.AllowedChatIDs,
		"proxy_url":        s.config.ProxyURL,
		"running":          s.running,
	}
}

// SaveConfig updates and saves config, restarting bot if necessary.
func (s *Service) SaveConfig(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var incoming BotConfig
	if err := json.Unmarshal(data, &incoming); err != nil {
		return err
	}

	s.mu.Lock()
	if incoming.Token != "" {
		s.config.Token = incoming.Token
	}
	s.config.Enabled = incoming.Enabled
	if len(incoming.AllowedChatIDs) > 0 {
		s.config.AllowedChatIDs = incoming.AllowedChatIDs
	}
	if incoming.ProxyURL != "" {
		s.config.ProxyURL = incoming.ProxyURL
	}
	_ = s.saveConfigLocked()
	shouldRun := s.config.Enabled && s.config.Token != ""
	s.mu.Unlock()

	s.Stop()
	if shouldRun {
		s.Start()
	}
	return nil
}

func (s *Service) loadConfig() {
	data, err := os.ReadFile(s.configFile)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &s.config)
}

func (s *Service) saveConfigLocked() error {
	_ = os.MkdirAll(filepath.Dir(s.configFile), 0755)
	data, _ := json.MarshalIndent(s.config, "", "  ")
	return os.WriteFile(s.configFile, data, 0644)
}

// ──────────────────────────────────────────────
// Polling & Telegram API
// ──────────────────────────────────────────────

func (s *Service) getClient() *http.Client {
	s.mu.RLock()
	proxyStr := s.config.ProxyURL
	s.mu.RUnlock()

	tr := &http.Transport{}
	if proxyStr != "" {
		if pURL, err := url.Parse(proxyStr); err == nil {
			tr.Proxy = http.ProxyURL(pURL)
		}
	}
	return &http.Client{
		Transport: tr,
		Timeout:   35 * time.Second,
	}
}

func (s *Service) pollLoop() {
	offset := 0
	for {
		select {
		case <-s.stopChan:
			return
		default:
		}

		s.mu.RLock()
		token := s.config.Token
		s.mu.RUnlock()
		if token == "" {
			time.Sleep(5 * time.Second)
			continue
		}

		apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=25", token, offset)
		client := s.getClient()
		resp, err := client.Get(apiURL)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var updResp struct {
			Ok     bool `json:"ok"`
			Result []struct {
				UpdateID int `json:"update_id"`
				Message  *struct {
					Chat struct {
						ID int64 `json:"id"`
					} `json:"chat"`
					Text string `json:"text"`
				} `json:"message"`
			} `json:"result"`
		}

		if err := json.Unmarshal(body, &updResp); err != nil || !updResp.Ok {
			time.Sleep(3 * time.Second)
			continue
		}

		for _, item := range updResp.Result {
			offset = item.UpdateID + 1
			if item.Message != nil && item.Message.Text != "" {
				s.handleMessage(item.Message.Chat.ID, item.Message.Text)
			}
		}
	}
}

func (s *Service) isAuthorized(chatID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.config.AllowedChatIDs) == 0 {
		return true // Allow all if not restricted
	}
	for _, id := range s.config.AllowedChatIDs {
		if id == chatID {
			return true
		}
	}
	return false
}

func (s *Service) sendMessage(chatID int64, text string) {
	s.mu.RLock()
	token := s.config.Token
	s.mu.RUnlock()
	if token == "" {
		return
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload, _ := json.Marshal(map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	})

	client := s.getClient()
	_, _ = client.Post(apiURL, "application/json", bytes.NewReader(payload))
}

// ──────────────────────────────────────────────
// Command Dispatching & Voice Matching
// ──────────────────────────────────────────────

func (s *Service) handleMessage(chatID int64, text string) {
	if !s.isAuthorized(chatID) {
		s.sendMessage(chatID, "⛔ 抱歉，您无权操作此 MiCube 网关。")
		return
	}

	text = strings.TrimSpace(text)
	parts := strings.Fields(text)
	cmd := ""
	if len(parts) > 0 {
		cmd = strings.ToLower(parts[0])
	}

	// 1. Check slash command
	switch cmd {
	case "/start", "/help":
		s.sendHelp(chatID)
		return
	case "/status":
		s.sendStatus(chatID)
		return
	case "/restart":
		s.sendMessage(chatID, "🔄 正在重启 Mihomo 核心...")
		if err := s.supervisor.Restart(); err != nil {
			s.sendMessage(chatID, fmt.Sprintf("❌ 重启失败: %v", err))
		} else {
			s.sendMessage(chatID, "✅ Mihomo 核心已成功重启！")
		}
		return
	case "/reload":
		s.sendMessage(chatID, "⟳ 正在热重载配置...")
		if err := s.supervisor.Reload(); err != nil {
			s.sendMessage(chatID, fmt.Sprintf("❌ 重载失败: %v", err))
		} else {
			s.sendMessage(chatID, "✅ 配置热重载成功！")
		}
		return
	case "/stop":
		_ = s.supervisor.Stop()
		s.sendMessage(chatID, "⏹ 核心已停止。")
		return
	case "/start_core":
		_ = s.supervisor.Start()
		s.sendMessage(chatID, "▶ 核心已启动。")
		return
	case "/update":
		s.sendMessage(chatID, "📥 正在拉取机场订阅并重新编译...")
		go func() {
			s.compiler.FetchAllSubscriptions()
			_ = s.compiler.CompileAll()
			_ = s.supervisor.Reload()
			s.sendMessage(chatID, "✅ 订阅更新并重载完成！")
		}()
		return
	case "/test":
		s.sendMessage(chatID, "⚡ 已触发节点批量测速，正在进行...")
		state := s.compiler.GetState()
		s.telemetry.TriggerBatchDelayTest(state.SpeedtestURL, state.SpeedtestTimeout)
		return
	case "/probe":
		if len(parts) < 2 {
			s.sendMessage(chatID, "用法: `/probe google.com`")
			return
		}
		target := parts[1]
		res := s.inspector.Probe(target)
		data, _ := json.MarshalIndent(res, "", "  ")
		s.sendMessage(chatID, fmt.Sprintf("🔍 *路由探测结果 (%s)*:\n```json\n%s\n```", target, string(data)))
		return
	}

	// 2. Natural language / voice intent fallback
	if matchedAction, ok := s.ParseLocalVoiceCommand(text); ok {
		s.sendMessage(chatID, fmt.Sprintf("🎙 识别到语音/自然语言意图: *%s*", matchedAction))
		switch matchedAction {
		case "status":
			s.sendStatus(chatID)
		case "restart":
			_ = s.supervisor.Restart()
			s.sendMessage(chatID, "✅ 核心已完成重启。")
		case "update":
			s.compiler.FetchAllSubscriptions()
			_ = s.compiler.CompileAll()
			_ = s.supervisor.Reload()
			s.sendMessage(chatID, "✅ 订阅更新成功。")
		case "test":
			state := s.compiler.GetState()
			s.telemetry.TriggerBatchDelayTest(state.SpeedtestURL, state.SpeedtestTimeout)
			s.sendMessage(chatID, "⚡ 已触发测速。")
		}
		return
	}

	if s.AIChat != nil {
		s.sendMessage(chatID, "🤖 MiCube 通义千问正在思考中...")
		reply, err := s.AIChat(text)
		if err == nil && reply != "" {
			s.sendMessage(chatID, reply)
			return
		}
	}

	s.sendMessage(chatID, "❓ 未知命令，输入 /help 查看支持的控制指令。")
}

// ParseLocalVoiceCommand checks if a transcription matches known intents.
func (s *Service) ParseLocalVoiceCommand(text string) (string, bool) {
	t := strings.ToLower(text)
	switch {
	case strings.Contains(t, "状态") || strings.Contains(t, "运行") || strings.Contains(t, "怎么样"):
		return "status", true
	case strings.Contains(t, "重启") || strings.Contains(t, "重开"):
		return "restart", true
	case strings.Contains(t, "更新订阅") || strings.Contains(t, "拉取节点") || strings.Contains(t, "更新节点"):
		return "update", true
	case strings.Contains(t, "测速") || strings.Contains(t, "测试延迟"):
		return "test", true
	}
	return "", false
}

func (s *Service) sendHelp(chatID int64) {
	msg := `🧊 *MiCube 智能网关机器人*

支持的操作指令：
/status - 📊 查看当前运行状态与内存
/restart - 🔄 重启 Mihomo 代理核心
/reload - ⟳ 热重载规则配置文件
/update - 📥 更新机场订阅节点
/test - ⚡ 触发全部节点延迟测速
/probe <域名> - 🔍 诊断目标路由与延迟
/stop - ⏹ 停止代理核心
/start_core - ▶ 启动代理核心

💡 也支持直接发送："查询状态"、"重启"、"测速"、"更新订阅"`
	s.sendMessage(chatID, msg)
}

func (s *Service) sendStatus(chatID int64) {
	st := s.supervisor.GetStatus()
	statusEmoji := "🟢 运行中"
	if !st.Running {
		statusEmoji = "🔴 已停止"
	}

	msg := fmt.Sprintf(`🧊 *MiCube 网关状态报告*
━━━━━━━━━━━━━━━━━
• 状态: %s
• 进程 PID: %d
• 架构: %s
• 运行时长: %d 秒
• 核心内存: %.1f MB
• 面板内存: %.1f MB
• 总计内存: %.1f MB
• TUN 模式: %s
• 开机自启: %s`,
		statusEmoji,
		st.PID,
		st.Arch,
		st.UptimeSeconds,
		st.CoreMemoryMB,
		st.PanelMemoryMB,
		st.TotalMemoryMB,
		st.TunMode,
		strconv.FormatBool(st.Autostart),
	)
	s.sendMessage(chatID, msg)
}
