package device

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/zyeyefyass/singcube/server/internal/ros"
)

// Device represents an identified LAN endpoint.
type Device struct {
	IP        string `json:"ip"`
	MAC       string `json:"mac"`
	HostName  string `json:"hostname"`
	Name      string `json:"name"` // User custom alias or fallback
	IsCustom  bool   `json:"is_custom"`
	Interface string `json:"interface"`
	Source    string `json:"source"`
	Online    bool   `json:"online"`
	LastSeen  int64  `json:"last_seen"`
}

// DeviceStore stores the persistent aliases mapping.
type DeviceStore struct {
	Aliases map[string]string `json:"aliases"` // key can be IP or UPPER(MAC)
}

// Manager manages LAN device discovery, alias persistence, and RouterOS sync.
type Manager struct {
	rootDir    string
	storeFile  string
	rosMgr     *ros.Manager
	clashURL   string
	mu         sync.RWMutex
	store      DeviceStore
	httpClient *http.Client
}

// NewManager creates a new device Manager.
func NewManager(rootDir string, rosMgr *ros.Manager, clashURL string) *Manager {
	if clashURL == "" {
		clashURL = "http://127.0.0.1:9090"
	}
	m := &Manager{
		rootDir:    rootDir,
		storeFile:  filepath.Join(rootDir, "data", "devices.json"),
		rosMgr:     rosMgr,
		clashURL:   clashURL,
		httpClient: &http.Client{Timeout: 3 * time.Second},
		store:      DeviceStore{Aliases: make(map[string]string)},
	}
	m.loadStore()
	return m
}

func (m *Manager) loadStore() {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := os.ReadFile(m.storeFile)
	if err == nil {
		_ = json.Unmarshal(data, &m.store)
	}
	if m.store.Aliases == nil {
		m.store.Aliases = make(map[string]string)
	}
}

func (m *Manager) saveStoreLocked() error {
	_ = os.MkdirAll(filepath.Dir(m.storeFile), 0755)
	data, err := json.MarshalIndent(m.store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.storeFile, data, 0644)
}

// SetAlias updates or deletes a custom name for an IP or MAC address.
func (m *Manager) SetAlias(ip, mac, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name = strings.TrimSpace(name)
	ip = strings.TrimSpace(ip)
	mac = strings.ToUpper(strings.TrimSpace(mac))

	if name == "" {
		if ip != "" {
			delete(m.store.Aliases, ip)
		}
		if mac != "" {
			delete(m.store.Aliases, mac)
		}
	} else {
		if ip != "" {
			m.store.Aliases[ip] = name
		}
		if mac != "" {
			m.store.Aliases[mac] = name
		}
	}

	if err := m.saveStoreLocked(); err != nil {
		return err
	}

	// Try to sync alias back to RouterOS DHCP / ARP comment if connected
	if m.rosMgr != nil && mac != "" {
		go func(targetMAC, aliasName string) {
			// Find lease in RouterOS and patch comment
			if data, err := m.rosMgr.QueryROS("/ip/dhcp-server/lease"); err == nil {
				var leases []map[string]any
				if json.Unmarshal(data, &leases) == nil {
					for _, l := range leases {
						lMac, _ := l["mac-address"].(string)
						lId, _ := l[".id"].(string)
						if strings.EqualFold(lMac, targetMAC) && lId != "" {
							_ = m.rosMgr.PatchROS("/ip/dhcp-server/lease/"+lId, map[string]string{
								"comment": aliasName,
							})
						}
					}
				}
			}
		}(mac, name)
	}

	return nil
}

// ResolveName returns the user custom alias for an IP or MAC, or empty string.
func (m *Manager) ResolveName(ip, mac string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if mac != "" {
		if name, ok := m.store.Aliases[strings.ToUpper(mac)]; ok && name != "" {
			return name
		}
	}
	if ip != "" {
		if name, ok := m.store.Aliases[ip]; ok && name != "" {
			return name
		}
	}
	return ""
}

// GetDevices returns a comprehensive list of LAN devices discovered across all sources.
func (m *Manager) GetDevices() []Device {
	m.mu.RLock()
	defer m.mu.RUnlock()

	deviceMap := make(map[string]*Device) // key: IP
	macToIP := make(map[string]string)

	now := time.Now().Unix()

	// 1. Fetch from RouterOS DHCP server leases (richest info)
	if m.rosMgr != nil {
		if data, err := m.rosMgr.QueryROS("/ip/dhcp-server/lease"); err == nil {
			var leases []map[string]any
			if json.Unmarshal(data, &leases) == nil {
				for _, l := range leases {
					ip, _ := l["address"].(string)
					mac, _ := l["mac-address"].(string)
					host, _ := l["host-name"].(string)
					comment, _ := l["comment"].(string)
					status, _ := l["status"].(string)
					if ip == "" {
						continue
					}
					macUpper := strings.ToUpper(mac)
					d := &Device{
						IP:        ip,
						MAC:       macUpper,
						HostName:  host,
						Source:    "RouterOS DHCP",
						Online:    status == "bound",
						LastSeen:  now,
						Interface: "bridge",
					}
					if comment != "" {
						d.Name = comment
						d.IsCustom = true
					}
					deviceMap[ip] = d
					if macUpper != "" {
						macToIP[macUpper] = ip
					}
				}
			}
		}

		// 2. Fetch from RouterOS ARP table
		if data, err := m.rosMgr.QueryROS("/ip/arp"); err == nil {
			var arps []map[string]any
			if json.Unmarshal(data, &arps) == nil {
				for _, a := range arps {
					ip, _ := a["address"].(string)
					mac, _ := a["mac-address"].(string)
					comment, _ := a["comment"].(string)
					iface, _ := a["interface"].(string)
					if ip == "" {
						continue
					}
					macUpper := strings.ToUpper(mac)
					if existing, ok := deviceMap[ip]; ok {
						if existing.MAC == "" {
							existing.MAC = macUpper
						}
						if existing.Interface == "" {
							existing.Interface = iface
						}
						if existing.Name == "" && comment != "" {
							existing.Name = comment
							existing.IsCustom = true
						}
					} else {
						d := &Device{
							IP:        ip,
							MAC:       macUpper,
							Interface: iface,
							Source:    "RouterOS ARP",
							Online:    true,
							LastSeen:  now,
						}
						if comment != "" {
							d.Name = comment
							d.IsCustom = true
						}
						deviceMap[ip] = d
						if macUpper != "" {
							macToIP[macUpper] = ip
						}
					}
				}
			}
		}
	}

	// 3. Read local Linux /proc/net/arp
	if f, err := os.Open("/proc/net/arp"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		isHeader := true
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if isHeader {
				isHeader = false
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 6 {
				ip := parts[0]
				macUpper := strings.ToUpper(parts[3])
				iface := parts[5]
				if macUpper == "00:00:00:00:00:00" || ip == "" {
					continue
				}
				if existing, ok := deviceMap[ip]; ok {
					if existing.MAC == "" {
						existing.MAC = macUpper
					}
					if existing.Interface == "" {
						existing.Interface = iface
					}
				} else {
					deviceMap[ip] = &Device{
						IP:        ip,
						MAC:       macUpper,
						Interface: iface,
						Source:    "Linux ARP",
						Online:    true,
						LastSeen:  now,
					}
					macToIP[macUpper] = ip
				}
			}
		}
	}

	// 4. Inject active connections client IPs from Clash
	req, err := http.NewRequest(http.MethodGet, m.clashURL+"/connections", nil)
	if err == nil {
		if resp, err := m.httpClient.Do(req); err == nil {
			var connData struct {
				Connections []struct {
					Metadata struct {
						SourceIP string `json:"sourceIP"`
					} `json:"metadata"`
				} `json:"connections"`
			}
			if json.NewDecoder(resp.Body).Decode(&connData) == nil {
				for _, c := range connData.Connections {
					srcIP := c.Metadata.SourceIP
					if srcIP != "" && srcIP != "127.0.0.1" && srcIP != "::1" {
						if existing, ok := deviceMap[srcIP]; ok {
							existing.Online = true
							existing.LastSeen = now
						} else {
							deviceMap[srcIP] = &Device{
								IP:       srcIP,
								Source:   "活跃链路",
								Online:   true,
								LastSeen: now,
							}
						}
					}
				}
			}
			resp.Body.Close()
		}
	}

	// 5. Apply user custom aliases from local store
	var result []Device
	for ip, d := range deviceMap {
		// Priority 1: Check alias by MAC
		if d.MAC != "" {
			if alias, ok := m.store.Aliases[d.MAC]; ok && alias != "" {
				d.Name = alias
				d.IsCustom = true
			}
		}
		// Priority 2: Check alias by IP
		if !d.IsCustom {
			if alias, ok := m.store.Aliases[ip]; ok && alias != "" {
				d.Name = alias
				d.IsCustom = true
			}
		}
		// Fallback name
		if d.Name == "" {
			if d.HostName != "" {
				d.Name = d.HostName
			} else {
				d.Name = d.IP
			}
		}
		result = append(result, *d)
	}

	// Also add any stored aliases that might not be in ARP currently (offline devices)
	for key, alias := range m.store.Aliases {
		if strings.Contains(key, ":") { // MAC
			if _, exists := macToIP[key]; !exists {
				result = append(result, Device{
					IP:       "-",
					MAC:      key,
					Name:     alias,
					IsCustom: true,
					Source:   "已保存离线设备",
					Online:   false,
					LastSeen: 0,
				})
			}
		} else if strings.Contains(key, ".") { // IP
			if _, exists := deviceMap[key]; !exists {
				result = append(result, Device{
					IP:       key,
					MAC:      "-",
					Name:     alias,
					IsCustom: true,
					Source:   "已保存离线设备",
					Online:   false,
					LastSeen: 0,
				})
			}
		}
	}

	return result
}
