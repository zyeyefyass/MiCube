package telemetry

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// ──────────────────────────────────────────────
// Schema & Types
// ──────────────────────────────────────────────

// TrafficRecord stores one 10-second traffic sample.
type TrafficRecord struct {
	Timestamp int64  `json:"ts"`
	Direction string `json:"direction"` // "upload" or "download"
	Rule      string `json:"rule"`
	Node      string `json:"node"`
	Host      string `json:"host"`
	Bytes     int64  `json:"bytes"`
}

// NodeLatency stores one delay measurement for a proxy node.
type NodeLatency struct {
	Timestamp int64  `json:"ts"`
	Node      string `json:"node"`
	LatencyMs int    `json:"latency_ms"`
	Alive     bool   `json:"alive"`
}

// SwitchLogEntry records when mihomo switched between proxy nodes.
type SwitchLogEntry struct {
	Timestamp int64  `json:"ts"`
	Group     string `json:"group"`
	FromNode  string `json:"from_node"`
	ToNode    string `json:"to_node"`
	Reason    string `json:"reason"`
}

// DrilldownItem represents a single aggregated row in drill-down queries.
type DrilldownItem struct {
	Key       string  `json:"key"`
	Bytes     int64   `json:"bytes"`
	Percent   float64 `json:"percent"`
	HumanSize string  `json:"human_size"`
}

// DrilldownResult is the response for QueryDrilldown.
type DrilldownResult struct {
	GrandTotal     int64           `json:"grand_total"`
	GrandTotalSize string          `json:"grand_total_size"`
	Items          []DrilldownItem `json:"items"`
}

// NodeHistory tracks latency history for a single node.
type NodeHistory struct {
	Node    string         `json:"node"`
	Samples []NodeLatency  `json:"samples"`
	AvgMs   float64        `json:"avg_ms"`
	Alive   bool           `json:"alive"`
}

// NodeTelemetryResult is the response for GetNodeTelemetry.
type NodeTelemetryResult struct {
	NodeHistory map[string]*NodeHistory `json:"node_history"`
	SwitchLogs  []SwitchLogEntry        `json:"switch_logs"`
}

// SpeedSample stores one point in time for the live waveform.
type SpeedSample struct {
	Timestamp int64 `json:"ts"`
	UpBps     int64 `json:"up_bps"`
	DownBps   int64 `json:"down_bps"`
}

// RealtimeSpeedResult is returned to the frontend for the live traffic chart.
type RealtimeSpeedResult struct {
	UpSpeedBps     int64         `json:"up_speed_bps"`
	DownSpeedBps   int64         `json:"down_speed_bps"`
	UpSpeedHuman   string        `json:"up_speed_human"`
	DownSpeedHuman string        `json:"down_speed_human"`
	ActiveConns    int           `json:"active_conns"`
	History        []SpeedSample `json:"history"`
}

// ──────────────────────────────────────────────
// Manager
// ──────────────────────────────────────────────

// Manager handles all telemetry storage and querying via SQLite.
type Manager struct {
	db              *sql.DB
	mihomoAPI       string
	mu              sync.RWMutex
	stopChan        chan struct{}
	speedMu         sync.Mutex
	lastUpTotal     int64
	lastDownTotal   int64
	lastSpeedTs     time.Time
	currUpSpeed     int64
	currDownSpeed   int64
	currActiveConns int
	speedHistory    []SpeedSample
}

// NewManager opens (or creates) the SQLite database and starts the background collector.
func NewManager(dbPath string) (*Manager, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal=WAL&_busy_timeout=3000")
	if err != nil {
		return nil, fmt.Errorf("telemetry: open db: %w", err)
	}

	// Enforce WAL mode and connection limits for low-memory environments.
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)

	m := &Manager{
		db:           db,
		mihomoAPI:    "http://127.0.0.1:9090",
		stopChan:     make(chan struct{}),
		speedHistory: make([]SpeedSample, 0, 32),
	}

	if err := m.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("telemetry: init schema: %w", err)
	}

	// Start background traffic & latency collection.
	go m.collectLoop()

	return m, nil
}

// Close stops the background collector and closes the database.
func (m *Manager) Close() error {
	close(m.stopChan)
	return m.db.Close()
}

// ──────────────────────────────────────────────
// Schema Initialization
// ──────────────────────────────────────────────

func (m *Manager) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS traffic (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		ts        INTEGER NOT NULL,
		direction TEXT    NOT NULL,
		rule      TEXT    NOT NULL DEFAULT '',
		node      TEXT    NOT NULL DEFAULT '',
		host      TEXT    NOT NULL DEFAULT '',
		bytes     INTEGER NOT NULL DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_traffic_ts ON traffic(ts);
	CREATE INDEX IF NOT EXISTS idx_traffic_dir_ts ON traffic(direction, ts);

	CREATE TABLE IF NOT EXISTS node_latency (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		ts         INTEGER NOT NULL,
		node       TEXT    NOT NULL,
		latency_ms INTEGER NOT NULL DEFAULT 0,
		alive      INTEGER NOT NULL DEFAULT 1
	);
	CREATE INDEX IF NOT EXISTS idx_latency_ts ON node_latency(ts);

	CREATE TABLE IF NOT EXISTS switch_log (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		ts        INTEGER NOT NULL,
		grp       TEXT    NOT NULL DEFAULT '',
		from_node TEXT    NOT NULL DEFAULT '',
		to_node   TEXT    NOT NULL DEFAULT '',
		reason    TEXT    NOT NULL DEFAULT ''
	);
	CREATE INDEX IF NOT EXISTS idx_switch_ts ON switch_log(ts);
	`
	_, err := m.db.Exec(schema)
	return err
}

// ──────────────────────────────────────────────
// Background Collection
// ──────────────────────────────────────────────

// collectLoop polls the mihomo RESTful API every 10 seconds for traffic data,
// and every 60 seconds for proxy latency data.
func (m *Manager) collectLoop() {
	trafficTicker := time.NewTicker(10 * time.Second)
	latencyTicker := time.NewTicker(60 * time.Second)
	// Purge data older than 7 days, check once per hour.
	purgeTicker := time.NewTicker(1 * time.Hour)
	defer trafficTicker.Stop()
	defer latencyTicker.Stop()
	defer purgeTicker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-trafficTicker.C:
			m.collectTraffic()
		case <-latencyTicker.C:
			m.collectLatency()
		case <-purgeTicker.C:
			m.purgeOldData(7 * 24 * time.Hour)
		}
	}
}

// collectTraffic reads the mihomo /connections endpoint and records byte counters.
func (m *Manager) collectTraffic() {
	resp, err := httpGet(m.mihomoAPI + "/connections")
	if err != nil {
		return
	}

	var data struct {
		DownloadTotal int64 `json:"downloadTotal"`
		UploadTotal   int64 `json:"uploadTotal"`
		Connections   []struct {
			Metadata struct {
				Host        string `json:"host"`
				ProcessPath string `json:"processPath"`
			} `json:"metadata"`
			Rule     string   `json:"rule"`
			RulePayload string `json:"rulePayload"`
			Chains   []string `json:"chains"`
			Upload   int64    `json:"upload"`
			Download int64    `json:"download"`
		} `json:"connections"`
	}
	if err := json.Unmarshal(resp, &data); err != nil {
		return
	}

	now := time.Now().Unix()
	tx, err := m.db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO traffic (ts, direction, rule, node, host, bytes) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	// Aggregate per-connection data into upload/download records.
	for _, conn := range data.Connections {
		node := ""
		if len(conn.Chains) > 0 {
			node = conn.Chains[0]
		}
		host := conn.Metadata.Host
		rule := conn.Rule
		if conn.RulePayload != "" {
			rule = rule + "(" + conn.RulePayload + ")"
		}

		if conn.Upload > 0 {
			_, _ = stmt.Exec(now, "upload", rule, node, host, conn.Upload)
		}
		if conn.Download > 0 {
			_, _ = stmt.Exec(now, "download", rule, node, host, conn.Download)
		}
	}

	_ = tx.Commit()
}

// collectLatency reads the mihomo /proxies endpoint and stores latency data.
func (m *Manager) collectLatency() {
	resp, err := httpGet(m.mihomoAPI + "/proxies")
	if err != nil {
		return
	}

	var data struct {
		Proxies map[string]struct {
			Name    string `json:"name"`
			Type    string `json:"type"`
			History []struct {
				Delay int `json:"delay"`
			} `json:"history"`
			Alive bool `json:"alive"`
		} `json:"proxies"`
	}
	if err := json.Unmarshal(resp, &data); err != nil {
		return
	}

	now := time.Now().Unix()
	tx, err := m.db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO node_latency (ts, node, latency_ms, alive) VALUES (?, ?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	for name, proxy := range data.Proxies {
		// Skip built-in group types.
		ptype := strings.ToLower(proxy.Type)
		if ptype == "selector" || ptype == "urltest" || ptype == "fallback" || ptype == "direct" || ptype == "reject" || ptype == "relay" || ptype == "loadbalance" {
			continue
		}

		latency := 0
		if len(proxy.History) > 0 {
			latency = proxy.History[len(proxy.History)-1].Delay
		}
		alive := 0
		if proxy.Alive {
			alive = 1
		}
		_, _ = stmt.Exec(now, name, latency, alive)
	}

	_ = tx.Commit()
}

// purgeOldData removes records older than the given duration.
func (m *Manager) purgeOldData(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge).Unix()
	_, _ = m.db.Exec("DELETE FROM traffic WHERE ts < ?", cutoff)
	_, _ = m.db.Exec("DELETE FROM node_latency WHERE ts < ?", cutoff)
	_, _ = m.db.Exec("DELETE FROM switch_log WHERE ts < ?", cutoff)
	_, _ = m.db.Exec("PRAGMA incremental_vacuum")
}

// ──────────────────────────────────────────────
// Query: Drilldown
// ──────────────────────────────────────────────

// QueryDrilldown aggregates traffic by a given dimension (rule, node, host)
// for the specified direction and time window.
// Optional filter: filterDim + filterVal narrows the query (e.g., filter by a specific rule).
func (m *Manager) QueryDrilldown(direction string, hours int, dimension, filterDim, filterVal string) map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if direction != "upload" && direction != "download" {
		direction = "upload"
	}
	if dimension == "" {
		dimension = "rule"
	}
	// Validate dimension to prevent SQL injection.
	switch dimension {
	case "rule", "node", "host":
		// OK
	default:
		dimension = "rule"
	}

	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour).Unix()

	var query string
	var args []any

	if filterDim != "" && filterVal != "" {
		// Validate filter dimension.
		switch filterDim {
		case "rule", "node", "host":
			// OK
		default:
			filterDim = "rule"
		}
		query = fmt.Sprintf(
			"SELECT %s, SUM(bytes) as total FROM traffic WHERE direction = ? AND ts >= ? AND %s = ? GROUP BY %s ORDER BY total DESC LIMIT 100",
			dimension, filterDim, dimension,
		)
		args = []any{direction, cutoff, filterVal}
	} else {
		query = fmt.Sprintf(
			"SELECT %s, SUM(bytes) as total FROM traffic WHERE direction = ? AND ts >= ? GROUP BY %s ORDER BY total DESC LIMIT 100",
			dimension, dimension,
		)
		args = []any{direction, cutoff}
	}

	rows, err := m.db.Query(query, args...)
	if err != nil {
		log.Printf("[Telemetry] drilldown query error: %v", err)
		return map[string]any{"grand_total": 0, "grand_total_size": "0 B", "items": []any{}}
	}
	defer rows.Close()

	var items []DrilldownItem
	var grandTotal int64

	for rows.Next() {
		var key string
		var total int64
		if err := rows.Scan(&key, &total); err != nil {
			continue
		}
		if key == "" {
			key = "(unknown)"
		}
		items = append(items, DrilldownItem{Key: key, Bytes: total, HumanSize: humanBytes(total)})
		grandTotal += total
	}

	// Calculate percentages.
	for i := range items {
		if grandTotal > 0 {
			items[i].Percent = float64(items[i].Bytes) / float64(grandTotal) * 100
		}
	}

	if items == nil {
		items = []DrilldownItem{}
	}

	return map[string]any{
		"grand_total":      grandTotal,
		"grand_total_size": humanBytes(grandTotal),
		"items":            items,
	}
}

// ──────────────────────────────────────────────
// Query: Node Telemetry
// ──────────────────────────────────────────────

// GetNodeTelemetry returns recent latency history for all tracked nodes
// plus the last 50 node switch log entries.
func (m *Manager) GetNodeTelemetry() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := NodeTelemetryResult{
		NodeHistory: make(map[string]*NodeHistory),
		SwitchLogs:  []SwitchLogEntry{},
	}

	// Last 2 hours of latency data.
	cutoff := time.Now().Add(-2 * time.Hour).Unix()
	rows, err := m.db.Query("SELECT ts, node, latency_ms, alive FROM node_latency WHERE ts >= ? ORDER BY ts ASC", cutoff)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var nl NodeLatency
			var alive int
			if err := rows.Scan(&nl.Timestamp, &nl.Node, &nl.LatencyMs, &alive); err != nil {
				continue
			}
			nl.Alive = alive == 1

			nh, ok := result.NodeHistory[nl.Node]
			if !ok {
				nh = &NodeHistory{Node: nl.Node}
				result.NodeHistory[nl.Node] = nh
			}
			nh.Samples = append(nh.Samples, nl)
			nh.Alive = nl.Alive
		}
	}

	// Calculate avg latency per node.
	for _, nh := range result.NodeHistory {
		totalMs := 0
		count := 0
		for _, s := range nh.Samples {
			if s.LatencyMs > 0 {
				totalMs += s.LatencyMs
				count++
			}
		}
		if count > 0 {
			nh.AvgMs = float64(totalMs) / float64(count)
		}
	}

	// Last 50 switch logs.
	slRows, err := m.db.Query("SELECT ts, grp, from_node, to_node, reason FROM switch_log ORDER BY ts DESC LIMIT 50")
	if err == nil {
		defer slRows.Close()
		for slRows.Next() {
			var sl SwitchLogEntry
			if err := slRows.Scan(&sl.Timestamp, &sl.Group, &sl.FromNode, &sl.ToNode, &sl.Reason); err != nil {
				continue
			}
			result.SwitchLogs = append(result.SwitchLogs, sl)
		}
	}
	// Reverse to chronological order.
	sort.Slice(result.SwitchLogs, func(i, j int) bool {
		return result.SwitchLogs[i].Timestamp < result.SwitchLogs[j].Timestamp
	})

	return map[string]any{
		"node_history": result.NodeHistory,
		"switch_logs":  result.SwitchLogs,
	}
}

// ──────────────────────────────────────────────
// Action: Batch Delay Test
// ──────────────────────────────────────────────

// TriggerBatchDelayTest asks mihomo to run a delay test on all proxy groups and individual proxies,
// using the specified target URL and timeout.
func (m *Manager) TriggerBatchDelayTest(testURL string, timeoutMs int) {
	if testURL == "" {
		testURL = "http://cp.cloudflare.com/generate_204"
	}
	if timeoutMs <= 0 {
		timeoutMs = 4000
	}

	// Get all proxies.
	resp, err := httpGet(m.mihomoAPI + "/proxies")
	if err != nil {
		log.Printf("[Telemetry] batch delay test: cannot get proxies: %v", err)
		return
	}

	var data struct {
		Proxies map[string]struct {
			Type string `json:"type"`
		} `json:"proxies"`
	}
	if err := json.Unmarshal(resp, &data); err != nil {
		return
	}

	client := &http.Client{Timeout: 12 * time.Second}
	// Throttle with worker pool (max 16 concurrent tests) to prevent CPU spike and bufferbloat
	sem := make(chan struct{}, 16)
	var wg sync.WaitGroup

	for name, proxy := range data.Proxies {
		ptype := strings.ToLower(proxy.Type)
		var delayURL string
		if ptype == "selector" || ptype == "urltest" || ptype == "fallback" {
			delayURL = fmt.Sprintf("%s/group/%s/delay?url=%s&timeout=%d", m.mihomoAPI, name, testURL, timeoutMs)
		} else if ptype != "direct" && ptype != "reject" && ptype != "relay" && ptype != "loadbalance" {
			delayURL = fmt.Sprintf("%s/proxies/%s/delay?url=%s&timeout=%d", m.mihomoAPI, name, testURL, timeoutMs)
		}
		if delayURL != "" {
			wg.Add(1)
			go func(target string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				if resp, err := client.Get(target); err == nil {
					resp.Body.Close()
				}
			}(delayURL)
		}
	}

	// Wait for all worker tasks to complete, then collect fresh latency data.
	go func() {
		wg.Wait()
		time.Sleep(500 * time.Millisecond)
		m.collectLatency()
	}()
}

// RecordSwitch stores a node switch event.
func (m *Manager) RecordSwitch(group, fromNode, toNode, reason string) {
	now := time.Now().Unix()
	_, _ = m.db.Exec(
		"INSERT INTO switch_log (ts, grp, from_node, to_node, reason) VALUES (?, ?, ?, ?, ?)",
		now, group, fromNode, toNode, reason,
	)
}

// GetRealtimeSpeed polls the instant traffic rates and returns historical waveform samples.
func (m *Manager) GetRealtimeSpeed() RealtimeSpeedResult {
	m.speedMu.Lock()
	defer m.speedMu.Unlock()

	resp, err := httpGet(m.mihomoAPI + "/connections")
	if err == nil {
		var data struct {
			DownloadTotal int64 `json:"downloadTotal"`
			UploadTotal   int64 `json:"uploadTotal"`
			Connections   []any `json:"connections"`
		}
		if json.Unmarshal(resp, &data) == nil {
			now := time.Now()
			if !m.lastSpeedTs.IsZero() {
				delta := now.Sub(m.lastSpeedTs).Seconds()
				if delta >= 0.4 {
					upDiff := data.UploadTotal - m.lastUpTotal
					downDiff := data.DownloadTotal - m.lastDownTotal
					if upDiff < 0 {
						upDiff = 0
					}
					if downDiff < 0 {
						downDiff = 0
					}
					m.currUpSpeed = int64(float64(upDiff) / delta)
					m.currDownSpeed = int64(float64(downDiff) / delta)
					m.lastUpTotal = data.UploadTotal
					m.lastDownTotal = data.DownloadTotal
					m.lastSpeedTs = now
					m.currActiveConns = len(data.Connections)

					m.speedHistory = append(m.speedHistory, SpeedSample{
						Timestamp: now.Unix(),
						UpBps:     m.currUpSpeed,
						DownBps:   m.currDownSpeed,
					})
					if len(m.speedHistory) > 30 {
						m.speedHistory = m.speedHistory[len(m.speedHistory)-30:]
					}
				}
			} else {
				m.lastUpTotal = data.UploadTotal
				m.lastDownTotal = data.DownloadTotal
				m.lastSpeedTs = now
				m.currActiveConns = len(data.Connections)
			}
		}
	}

	histCopy := make([]SpeedSample, len(m.speedHistory))
	copy(histCopy, m.speedHistory)

	return RealtimeSpeedResult{
		UpSpeedBps:     m.currUpSpeed,
		DownSpeedBps:   m.currDownSpeed,
		UpSpeedHuman:   humanBytes(m.currUpSpeed) + "/s",
		DownSpeedHuman: humanBytes(m.currDownSpeed) + "/s",
		ActiveConns:    m.currActiveConns,
		History:        histCopy,
	}
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

// httpGet performs a simple GET request with a short timeout.
func httpGet(url string) ([]byte, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// humanBytes formats byte counts into human-readable strings.
func humanBytes(b int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)
	switch {
	case b >= TB:
		return fmt.Sprintf("%.2f TB", float64(b)/float64(TB))
	case b >= GB:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.2f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.2f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
