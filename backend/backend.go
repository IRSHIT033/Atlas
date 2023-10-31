package backend

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend interface {
	SetAlive(bool)
	IsAlive() bool
	GetURL() *url.URL
	GetWeight() int
	GetActiveConnections() int
	Serve(http.ResponseWriter, *http.Request)
}

type backend struct {
	url          *url.URL
	alive        bool
	mux          sync.RWMutex
	connections  int
	weight       int
	reverseProxy *httputil.ReverseProxy
}

func (b *backend) GetActiveConnections() int {
	b.mux.RLock()
	connections := b.connections
	b.mux.RUnlock()
	return connections
}

func (b *backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.alive = alive
	b.mux.Unlock()
}

func (b *backend) IsAlive() bool {
	b.mux.RLock()
	alive := b.alive
	defer b.mux.RUnlock()
	return alive
}

func (b *backend) GetWeight() int {
	return b.weight
}

func (b *backend) GetURL() *url.URL {
	return b.url
}

func (b *backend) Serve(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		b.mux.Lock()
		b.connections--
		b.mux.Unlock()
	}()

	b.mux.Lock()
	b.connections++
	b.mux.Unlock()
	b.reverseProxy.ServeHTTP(rw, req)
}

func NewBackend(u *url.URL, w int, rp *httputil.ReverseProxy) Backend {
	return &backend{
		url:          u,
		weight:       w,
		alive:        true,
		reverseProxy: rp,
	}
}
