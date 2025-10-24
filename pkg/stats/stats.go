package stats

import (
	"fmt"
	"sync"
	"time"
)

type Statistics struct {
	mu              sync.RWMutex
	totalSent       uint64
	totalReceived   uint64
	startTime       time.Time
	lastReportTime  time.Time
	receivedPackets map[uint64]bool
}

func NewStatistics() *Statistics {
	now := time.Now()
	return &Statistics{
		startTime:       now,
		lastReportTime:  now,
		receivedPackets: make(map[uint64]bool),
	}
}

func (s *Statistics) IncrementSent() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalSent++
}

func (s *Statistics) IncrementReceived(seq uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalReceived++
	s.receivedPackets[seq] = true
}

func (s *Statistics) GetStats() (sent, received uint64, lossRate float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	sent = s.totalSent
	received = s.totalReceived
	
	if sent > 0 {
		lossRate = float64(sent-received) / float64(sent) * 100
	}
	
	return
}

func (s *Statistics) Report() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	duration := time.Since(s.startTime)
	sent := s.totalSent
	received := s.totalReceived
	lost := sent - received
	lossRate := 0.0
	
	if sent > 0 {
		lossRate = float64(lost) / float64(sent) * 100
	}
	
	return fmt.Sprintf(
		"Duration: %v | Sent: %d | Received: %d | Lost: %d | Loss Rate: %.2f%%",
		duration.Round(time.Second), sent, received, lost, lossRate,
	)
}

func (s *Statistics) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.totalSent = 0
	s.totalReceived = 0
	s.receivedPackets = make(map[uint64]bool)
	s.lastReportTime = time.Now()
}
