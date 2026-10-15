package inspector

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/zyeyefyass/singcube/server/internal/compiler"
)

// Inspector performs real route probing: DNS resolution, rule matching,
// and HTTP connectivity tests to diagnose routing behaviour.
type Inspector struct {
	Compiler  *compiler.Engine
	MihomoAPI string
}

// NewInspector creates an Inspector wired to the config compiler.
func NewInspector(c *compiler.Engine) *Inspector {
	return &Inspector{Compiler: c, MihomoAPI: "http://127.0.0.1:9090"}
}

// ProbeResult is the structured output of a route inspection.
type ProbeResult struct {
	TargetHost  string            `json:"target_host"`
	Consistent  bool              `json:"consistent"`
	Theoretical TheoreticalResult `json:"theoretical"`
	Actual      ActualResult      `json:"actual"`
	DNS         DNSResult         `json:"dns"`
	Error       string            `json:"error,omitempty"`
}

// TheoreticalResult describes what rule *should* match for a host.
type TheoreticalResult struct {
	MatchedRuleName string `json:"matched_rule_name"`
	OutboundGroup   string `json:"outbound_group"`
}

// ActualResult describes what mihomo actually does with a host.
type ActualResult struct {
	ResolvedIP     string `json:"resolved_ip"`
	ActualNode     string `json:"actual_node"`
	HTTPStatusCode int    `json:"http_status_code"`
	LatencyMs      int64  `json:"latency_ms"`
}

// DNSResult stores DNS resolution details.
type DNSResult struct {
	RawIPs   []string `json:"raw_ips"`
	IsFakeIP bool     `json:"is_fake_ip"`
	Resolver string   `json:"resolver"`
	TimingMs int64    `json:"timing_ms"`
}

// Probe performs a complete route diagnosis for the given hostname.
func (ins *Inspector) Probe(host string) map[string]any {
	result := ProbeResult{
		TargetHost: host,
	}

	// Step 1: DNS Resolution via system (which goes through mihomo's DNS).
	dnsStart := time.Now()
	ips, dnsErr := net.LookupHost(host)
	dnsTiming := time.Since(dnsStart).Milliseconds()

	result.DNS.TimingMs = dnsTiming
	result.DNS.Resolver = "mihomo-dns (127.0.0.1:7853)"

	if dnsErr != nil {
		result.Error = fmt.Sprintf("DNS resolution failed: %v", dnsErr)
		result.DNS.RawIPs = []string{}
	} else {
		result.DNS.RawIPs = ips
		if len(ips) > 0 {
			result.Actual.ResolvedIP = ips[0]
			// Detect fake-ip range (198.18.0.0/15).
			ip := net.ParseIP(ips[0])
			if ip != nil {
				_, fakeNet, _ := net.ParseCIDR("198.18.0.0/15")
				if fakeNet != nil && fakeNet.Contains(ip) {
					result.DNS.IsFakeIP = true
				}
			}
		}
	}

	// Step 2: Query mihomo rules endpoint to find the theoretical match.
	theoretical := ins.matchRule(host)
	result.Theoretical = theoretical

	// Step 3: Find what node mihomo is actually using via /connections.
	actualNode, actualRule := ins.findActiveConnection(host)
	if actualNode != "" {
		result.Actual.ActualNode = actualNode
	} else {
		result.Actual.ActualNode = theoretical.OutboundGroup
	}

	// Step 4: HTTP connectivity test through the proxy.
	httpStart := time.Now()
	statusCode := ins.testHTTP(host)
	httpLatency := time.Since(httpStart).Milliseconds()
	result.Actual.HTTPStatusCode = statusCode
	result.Actual.LatencyMs = httpLatency

	// Step 5: Consistency check.
	if actualRule != "" {
		result.Consistent = strings.Contains(actualRule, theoretical.MatchedRuleName) ||
			actualNode == theoretical.OutboundGroup ||
			actualNode == "DIRECT"
	} else {
		result.Consistent = statusCode > 0
	}

	// Marshal to map[string]any for generic API response.
	data, _ := json.Marshal(result)
	var out map[string]any
	_ = json.Unmarshal(data, &out)
	return out
}

// matchRule queries mihomo's /rules endpoint to find the best rule match.
func (ins *Inspector) matchRule(host string) TheoreticalResult {
	resp, err := httpGet(ins.MihomoAPI + "/rules")
	if err != nil {
		return TheoreticalResult{MatchedRuleName: "MATCH", OutboundGroup: "DIRECT"}
	}

	var data struct {
		Rules []struct {
			Type    string `json:"type"`
			Payload string `json:"payload"`
			Proxy   string `json:"proxy"`
		} `json:"rules"`
	}
	if err := json.Unmarshal(resp, &data); err != nil {
		return TheoreticalResult{MatchedRuleName: "MATCH", OutboundGroup: "DIRECT"}
	}

	host = strings.ToLower(host)
	for _, rule := range data.Rules {
		payload := strings.ToLower(rule.Payload)
		matched := false
		switch strings.ToUpper(rule.Type) {
		case "DOMAIN":
			matched = host == payload
		case "DOMAIN-SUFFIX":
			matched = host == payload || strings.HasSuffix(host, "."+payload)
		case "DOMAIN-KEYWORD":
			matched = strings.Contains(host, payload)
		case "MATCH":
			matched = true
		}
		if matched {
			return TheoreticalResult{
				MatchedRuleName: fmt.Sprintf("%s,%s", rule.Type, rule.Payload),
				OutboundGroup:   rule.Proxy,
			}
		}
	}

	return TheoreticalResult{MatchedRuleName: "MATCH", OutboundGroup: "DIRECT"}
}

// findActiveConnection looks for an active connection to the host in mihomo.
func (ins *Inspector) findActiveConnection(host string) (node, rule string) {
	resp, err := httpGet(ins.MihomoAPI + "/connections")
	if err != nil {
		return "", ""
	}

	var data struct {
		Connections []struct {
			Metadata struct {
				Host string `json:"host"`
			} `json:"metadata"`
			Rule   string   `json:"rule"`
			Chains []string `json:"chains"`
		} `json:"connections"`
	}
	if err := json.Unmarshal(resp, &data); err != nil {
		return "", ""
	}

	host = strings.ToLower(host)
	for _, conn := range data.Connections {
		if strings.ToLower(conn.Metadata.Host) == host || strings.HasPrefix(strings.ToLower(conn.Metadata.Host), host) {
			nodeName := ""
			if len(conn.Chains) > 0 {
				nodeName = conn.Chains[0]
			}
			return nodeName, conn.Rule
		}
	}

	return "", ""
}

// testHTTP performs a lightweight HTTP GET through the system's proxy path.
func (ins *Inspector) testHTTP(host string) int {
	url := "http://" + host
	if !strings.Contains(host, ":") {
		url = "https://" + host
	}

	client := &http.Client{
		Timeout: 8 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return 0
	}
	req.Header.Set("User-Agent", "MiCube-Inspector/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

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
