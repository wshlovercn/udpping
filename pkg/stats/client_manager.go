package stats

import (
	"fmt"
	"sync"
	"time"
)

type ClientInfo struct {
	Address          string
	Statistics       *Statistics
	LastPacketTime   time.Time
	FirstPacketTime  time.Time
	LastFeedbackTime time.Time
	mu               sync.RWMutex
}

func NewClientInfo(address string) *ClientInfo {
	now := time.Now()
	return &ClientInfo{
		Address:          address,
		Statistics:       NewStatistics(),
		FirstPacketTime:  now,
		LastPacketTime:   now,
		LastFeedbackTime: now,
	}
}

func (c *ClientInfo) UpdateLastPacketTime() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastPacketTime = time.Now()
}

func (c *ClientInfo) ShouldSendFeedback(interval time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.LastFeedbackTime) >= interval
}

func (c *ClientInfo) UpdateFeedbackTime() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastFeedbackTime = time.Now()
}

func (c *ClientInfo) IsInactive(timeout time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.LastPacketTime) > timeout
}

type ClientManager struct {
	clients          map[string]*ClientInfo
	mu               sync.RWMutex
	feedbackInterval time.Duration
	inactiveTimeout  time.Duration
}

func NewClientManager(feedbackInterval, inactiveTimeout time.Duration) *ClientManager {
	return &ClientManager{
		clients:          make(map[string]*ClientInfo),
		feedbackInterval: feedbackInterval,
		inactiveTimeout:  inactiveTimeout,
	}
}

func (cm *ClientManager) GetOrCreateClient(address string) *ClientInfo {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	client, exists := cm.clients[address]
	if !exists {
		client = NewClientInfo(address)
		cm.clients[address] = client
	}
	return client
}

func (cm *ClientManager) GetClient(address string) (*ClientInfo, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	client, exists := cm.clients[address]
	return client, exists
}

func (cm *ClientManager) GetAllClients() []*ClientInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	clients := make([]*ClientInfo, 0, len(cm.clients))
	for _, client := range cm.clients {
		clients = append(clients, client)
	}
	return clients
}

func (cm *ClientManager) CleanupInactiveClients() int {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	removed := 0
	for address, client := range cm.clients {
		if client.IsInactive(cm.inactiveTimeout) {
			delete(cm.clients, address)
			removed++
		}
	}
	return removed
}

func (cm *ClientManager) GetClientCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.clients)
}

func (cm *ClientManager) GetReport() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if len(cm.clients) == 0 {
		return "No active clients"
	}

	report := fmt.Sprintf("Active clients: %d\n", len(cm.clients))
	for address, client := range cm.clients {
		sent, received, lossRate := client.Statistics.GetStats()
		duration := time.Since(client.FirstPacketTime)
		report += fmt.Sprintf("  %s - Duration: %v | Sent: %d | Received: %d | Loss: %.2f%%\n",
			address, duration.Round(time.Second), sent, received, lossRate)
	}
	return report
}
