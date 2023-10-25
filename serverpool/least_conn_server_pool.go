package serverpool

import (
	"sync"

	"github.com/IRSHIT033/Atlas/backend"
)

type leastConnServerPool struct {
	backends []backend.Backend
	mux      sync.RWMutex
}

func (s *leastConnServerPool) GetNextValidPeer() backend.Backend {
	var leastConnectedPeer backend.Backend

	// first
	for _, b := range s.backends {
		if b.IsAlive() {
			leastConnectedPeer = b
			break
		}
	}

	for _, b := range s.backends {
		if !b.IsAlive() {
			continue
		}
		if leastConnectedPeer.GetActiveConnections() > b.GetActiveConnections() {
			leastConnectedPeer = b
		}
	}

	return leastConnectedPeer
}

func (s *leastConnServerPool) AddBackend(b backend.Backend) {
	s.backends = append(s.backends, b)
}

func (s *leastConnServerPool) GetServerPoolSize() int {
	return len(s.backends)
}

func (s *leastConnServerPool) GetBackends() []backend.Backend {
	return s.backends
}
