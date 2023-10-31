package serverpool

import (
	"math"
	"sync"

	"github.com/IRSHIT033/Atlas/backend"
)

type leastConnServerPool struct {
	backends []backend.Backend
	mux      sync.RWMutex
}

func (s *leastConnServerPool) GetNextValidPeer() backend.Backend {
	
	var leastConnectedPeer backend.Backend
    
	s.mux.Lock()
	defer s.mux.Unlock()

	minConnCount:= math.MaxInt64

	for _, b := range s.backends {
		if !b.IsAlive() {
			continue
		}
		effectiveConnCount := b.GetActiveConnections() / b.GetWeight()
		if  effectiveConnCount < minConnCount {
			leastConnectedPeer = b
			minConnCount = effectiveConnCount
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
