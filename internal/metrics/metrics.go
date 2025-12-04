package metrics

import (
	"sync"
	"time"
)

type Metrics struct {
	// Connection metrics
	TotalConnections     int64
	ActiveConnections    int64
	ConnectionsCreated   int64
	ConnectionsClosed    int64

	// Peer metrics
	TotalPeers           int64
	ActivePeers          int64
	PeersAnnounced       int64
	PeersDiscovered      int64

	// Hub metrics
	TotalHubs            int64
	BootstrapConnected   int64
	CrossHubMessages     int64

	// Message metrics
	MessagesProcessed    int64
	MessageErrors        int64
	MessagesBroadcast    int64

	// Timing
	StartTime            time.Time
	LastCleanup          time.Time

	mu sync.RWMutex
}

var globalMetrics = &Metrics{
	StartTime: time.Now(),
}

func GetMetrics() *Metrics {
	return globalMetrics
}

func (m *Metrics) ConnectionOpened() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalConnections++
	m.ActiveConnections++
	m.ConnectionsCreated++
}

func (m *Metrics) ConnectionClosed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ActiveConnections > 0 {
		m.ActiveConnections--
	}
	m.ConnectionsClosed++
}

func (m *Metrics) PeerAnnounced() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalPeers++
	m.ActivePeers++
	m.PeersAnnounced++
}

func (m *Metrics) PeerDiscovered() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PeersDiscovered++
}

func (m *Metrics) PeerRemoved() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ActivePeers > 0 {
		m.ActivePeers--
	}
}

func (m *Metrics) HubConnected() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.BootstrapConnected++
}

func (m *Metrics) CrossHubMessageSent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CrossHubMessages++
}

func (m *Metrics) MessageProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesProcessed++
}

func (m *Metrics) MessageFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessageErrors++
}

func (m *Metrics) MessageBroadcast(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesBroadcast += count
}

func (m *Metrics) CleanupPerformed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LastCleanup = time.Now()
}

func (m *Metrics) Snapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.StartTime)

	return map[string]interface{}{
		"timestamp":           time.Now().Format(time.RFC3339),
		"uptime_ms":           uptime.Milliseconds(),
		"connections": map[string]interface{}{
			"total":   m.TotalConnections,
			"active":  m.ActiveConnections,
			"created": m.ConnectionsCreated,
			"closed":  m.ConnectionsClosed,
		},
		"peers": map[string]interface{}{
			"total":      m.TotalPeers,
			"active":     m.ActivePeers,
			"announced":  m.PeersAnnounced,
			"discovered": m.PeersDiscovered,
		},
		"hubs": map[string]interface{}{
			"bootstrap_connected": m.BootstrapConnected,
			"cross_hub_messages":  m.CrossHubMessages,
		},
		"messages": map[string]interface{}{
			"processed":  m.MessagesProcessed,
			"errors":     m.MessageErrors,
			"broadcast":  m.MessagesBroadcast,
		},
	}
}
