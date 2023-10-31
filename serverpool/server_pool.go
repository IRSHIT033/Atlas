package serverpool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/IRSHIT033/Atlas/backend"
	"github.com/IRSHIT033/Atlas/utils"
	"go.uber.org/zap"
)

type ServerPool interface {
	GetBackends() []backend.Backend
	GetNextValidPeer() backend.Backend
	AddBackend(backend.Backend)
	GetServerPoolSize() int
}

type roundRobinServerPool struct {
	backends  []backend.Backend
	mux       sync.RWMutex
	current   int
	maxWeight int
	gcd       int
}

// For weighted round robin

func findGCD(a, b int) int {
	if b == 0 {
		return a
	}
	return findGCD(b, a%b)
}

func calculateGCDofWeights(backends []backend.Backend) int {
	gcd := backends[0].GetWeight()
	for _, b := range backends {
		gcd = findGCD(gcd, b.GetWeight())
	}
	return gcd
}

func getMaxWeight(backends []backend.Backend) int {
	maxWeight := backends[0].GetWeight()
	for _, b := range backends {
		if b.GetWeight() > maxWeight {
			maxWeight = b.GetWeight()
		}
	}
	return maxWeight
}

func (s *roundRobinServerPool) Rotate() backend.Backend {
	s.mux.Lock()
	if s.maxWeight == 0 { // for round robin
		s.current = (s.current + 1) % s.GetServerPoolSize()
	} else { // for weighted round robin
		for {
			s.current = (s.current + 1) % s.GetServerPoolSize()
			if s.current == 0 {
				s.maxWeight = s.maxWeight - s.gcd
				if s.maxWeight <= 0 {
					s.maxWeight = getMaxWeight(s.backends)
				}
			}
			if s.backends[s.current].GetWeight() >= s.maxWeight {
				break
			}
		}
	}
	s.mux.Unlock()
	return s.backends[s.current]
}

func (s *roundRobinServerPool) GetNextValidPeer() backend.Backend {
	s.maxWeight = getMaxWeight(s.backends)
	s.gcd = calculateGCDofWeights(s.backends)

	for i := 0; i < s.GetServerPoolSize(); i++ {
		nextPeer := s.Rotate()
		if nextPeer.IsAlive() {
			return nextPeer
		}
	}
	return nil
}

func (s *roundRobinServerPool) GetBackends() []backend.Backend {
	return s.backends
}

func (s *roundRobinServerPool) AddBackend(b backend.Backend) {
	s.backends = append(s.backends, b)
}

func (s *roundRobinServerPool) GetServerPoolSize() int {
	return len(s.backends)
}

func HealthCheck(ctx context.Context, s ServerPool) {
	aliveChannel := make(chan bool, 1)

	for _, b := range s.GetBackends() {
		b := b
		requestCtx, stop := context.WithTimeout(ctx, 10*time.Second)
		defer stop()
		status := "up"
		go backend.IsBackendAlive(requestCtx, aliveChannel, b.GetURL())

		select {
		case <-ctx.Done():
			utils.Logger.Info("Gracefully shutting down health check")
			return
		case alive := <-aliveChannel:
			b.SetAlive(alive)
			if !alive {
				status = "down"
			}
		}
		utils.Logger.Debug(
			"URL Status",
			zap.String("URL", b.GetURL().String()),
			zap.String("status", status),
		)
	}
}

func NewServerPool(strategy utils.LoadBalanceStrategy) (ServerPool, error) {
	switch strategy {
	case utils.RoundRobin:
		return &roundRobinServerPool{
			backends: make([]backend.Backend, 0),
			current:  0,
			gcd:      0,
		}, nil

	case utils.LeastConnected:
		return &leastConnServerPool{
			backends: make([]backend.Backend, 0),
		}, nil
	default:
		return nil, fmt.Errorf("invalid strategy")
	}
}

func LaunchHealthCheck(ctx context.Context, sp ServerPool) {
	t := time.NewTicker(time.Second * 20)
	utils.Logger.Info("Starting health check...")
	for {
		select {
		case <-t.C:
			go HealthCheck(ctx, sp)
		case <-ctx.Done():
			utils.Logger.Info("Closing Health Check")
			return
		}
	}
}
