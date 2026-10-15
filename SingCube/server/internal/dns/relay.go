package dns

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"
)

// SmartRelay provides strict anti-DNS-leak relaying on port 53 (UDP + TCP).
// In StrictAntiLeak mode, queries are NEVER forwarded to external plaintext
// public DNS. If Mihomo DNS (127.0.0.1:7853) is unreachable, it returns SERVFAIL
// instead of exposing client query records to the local ISP or public internet.
type SmartRelay struct {
	ListenAddr     string
	MihomoDNS      string
	StrictAntiLeak bool
	coreHealthy    atomic.Bool
	stopChan       chan struct{}
}

// NewSmartRelay creates a new relay with strict anti-leak enabled by default.
func NewSmartRelay(listenAddr, mihomoDNS string) *SmartRelay {
	return &SmartRelay{
		ListenAddr:     listenAddr,
		MihomoDNS:      mihomoDNS,
		StrictAntiLeak: true, // Strict Anti-DNS-Leak by default
		stopChan:       make(chan struct{}),
	}
}

// SetCoreHealthy updates the Mihomo core running state.
func (r *SmartRelay) SetCoreHealthy(h bool) {
	r.coreHealthy.Store(h)
}

// Start begins listening on both UDP and TCP port 53.
func (r *SmartRelay) Start() error {
	udpAddr, err := net.ResolveUDPAddr("udp", r.ListenAddr)
	if err != nil {
		return err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", r.ListenAddr)
	if err != nil {
		udpConn.Close()
		return err
	}
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		udpConn.Close()
		return err
	}

	// UDP listener
	go func() {
		defer udpConn.Close()
		buf := make([]byte, 4096)
		for {
			select {
			case <-r.stopChan:
				return
			default:
				n, clientAddr, err := udpConn.ReadFromUDP(buf)
				if err != nil {
					continue
				}
				req := make([]byte, n)
				copy(req, buf[:n])

				go func(c *net.UDPConn, a *net.UDPAddr, query []byte) {
					resp := r.resolve(query)
					if resp != nil {
						_, _ = c.WriteToUDP(resp, a)
					}
				}(udpConn, clientAddr, req)
			}
		}
	}()

	// TCP listener
	go func() {
		defer tcpListener.Close()
		for {
			select {
			case <-r.stopChan:
				return
			default:
				clientConn, err := tcpListener.AcceptTCP()
				if err != nil {
					continue
				}
				go r.handleTCP(clientConn)
			}
		}
	}()

	log.Printf("[DNS] Strict Anti-Leak DNS Relay listening on %s (UDP/TCP)", r.ListenAddr)
	return nil
}

// Stop terminates the relay.
func (r *SmartRelay) Stop() {
	close(r.stopChan)
}

// resolve executes the strict forwarding policy.
func (r *SmartRelay) resolve(req []byte) []byte {
	// 1. Try Mihomo Core DNS first (Fake-IP / Smart Routing)
	if r.coreHealthy.Load() {
		resp, err := exchangeUDP(r.MihomoDNS, req, 1200*time.Millisecond)
		if err == nil && len(resp) > 0 {
			return resp
		}
		// Quick second attempt
		resp, err = exchangeUDP(r.MihomoDNS, req, 1200*time.Millisecond)
		if err == nil && len(resp) > 0 {
			return resp
		}
	}

	// 2. Strict Anti-Leak Enforcement:
	// If StrictAntiLeak is ON, NEVER fallback to plaintext public DNS (223.5.5.5).
	// Doing so would instantly leak target domains to the ISP.
	// Instead, return standard SERVFAIL to prevent exposure while core restarts.
	if r.StrictAntiLeak {
		return buildServfailResponse(req)
	}

	// If strict mode is explicitly disabled, emergency fallback to encrypted or local DNS
	resp, _ := exchangeUDP("223.5.5.5:53", req, 1000*time.Millisecond)
	return resp
}

// handleTCP handles incoming DNS-over-TCP requests.
func (r *SmartRelay) handleTCP(conn *net.TCPConn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))

	// DNS over TCP prefixes each message with a 2-byte length
	var length uint16
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return
	}

	resp := r.resolve(buf)
	if resp != nil {
		respLen := uint16(len(resp))
		_ = binary.Write(conn, binary.BigEndian, respLen)
		_, _ = conn.Write(resp)
	}
}

// exchangeUDP performs a standard UDP query with timeout.
func exchangeUDP(upstream string, req []byte, timeout time.Duration) ([]byte, error) {
	conn, err := net.DialTimeout("udp", upstream, timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))
	if _, err := conn.Write(req); err != nil {
		return nil, err
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// buildServfailResponse constructs a valid DNS SERVFAIL response packet
// for the given query to safely signal temporary failure without leaking info.
func buildServfailResponse(query []byte) []byte {
	if len(query) < 12 {
		return nil
	}
	resp := make([]byte, len(query))
	copy(resp, query)

	// Set QR=1 (response), RCODE=2 (Server failure), RA=1
	resp[2] = (resp[2] | 0x80) & 0xFE // QR=1, Opcode preserved
	resp[3] = (resp[3] & 0xF0) | 0x02 // RCODE=2 (SERVFAIL), RA=0
	return resp
}
