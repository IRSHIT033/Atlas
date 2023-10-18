package frontend

import (
	"net/http"

	"github.com/IRSHIT033/Atlas/serverpool"
)

const (
	RETRY_ATTEMPTED int = 0
)

func AllowRetry(r *http.Request) bool {
	if _, ok := r.Context().Value(RETRY_ATTEMPTED).(bool); ok {
		return false
	}
	return true
}

type ILoadBalancer interface {
	Serve(http.ResponseWriter, *http.Request)
}

type LoadBalancer struct {
	serverPool serverpool.ServerPool
}

func (lb *LoadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	peer := lb.serverPool.GetNextValidPeer()
	if peer != nil {
		peer.Serve(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func NewLoadBalancer(serverPool serverpool.ServerPool) ILoadBalancer {
	return &LoadBalancer{
		serverPool: serverPool,
	}
}
