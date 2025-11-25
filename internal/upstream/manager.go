package upstream

import (
	"sync"
	"time"

	"vibeway/internal/config"
)

type Upstream struct {
	Name           string
	URLs           []string
	LoadBalancer   LoadBalancer
	HealthChecker  *HealthChecker
	CircuitBreaker *CircuitBreaker

	activeRequests map[string]int64
	activeReqMu    sync.RWMutex
}

type Manager struct {
	upstreams map[string]*Upstream
	mu        sync.RWMutex
}

func NewManager(cfg map[string]config.UpstreamConfig) *Manager {
	m := &Manager{
		upstreams: make(map[string]*Upstream),
	}

	for name, uCfg := range cfg {
		lb := NewRoundRobin() // Default to RR

		u := &Upstream{
			Name:           name,
			URLs:           uCfg.URLs,
			activeRequests: make(map[string]int64),
			HealthChecker:  NewHealthChecker(uCfg.URLs, 10*time.Second), // Default interval
			CircuitBreaker: NewCircuitBreaker(
				uCfg.CircuitBreaker.FailureThreshold,
				time.Duration(uCfg.CircuitBreaker.ResetTimeoutMs)*time.Millisecond,
			),
		}

		if uCfg.LoadBalancer == "least_connections" {
			u.LoadBalancer = NewLeastConnections(u.GetActiveRequestCount)
		} else {
			u.LoadBalancer = lb
		}

		// Start health checks
		u.HealthChecker.Start()

		m.upstreams[name] = u
	}

	return m
}

func (m *Manager) GetUpstream(name string) (*Upstream, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.upstreams[name]
	return u, ok
}

func (u *Upstream) GetNextURL() (string, bool) {
	// Filter healthy URLs
	healthyURLs := u.HealthChecker.GetHealthyURLs()
	if len(healthyURLs) == 0 {
		return "", false
	}

	// Check circuit breaker
	if !u.CircuitBreaker.Allow() {
		return "", false
	}

	return u.LoadBalancer.Next(healthyURLs), true
}

func (u *Upstream) IncConnection(url string) {
	u.activeReqMu.Lock()
	defer u.activeReqMu.Unlock()
	u.activeRequests[url]++
}

func (u *Upstream) DecConnection(url string) {
	u.activeReqMu.Lock()
	defer u.activeReqMu.Unlock()
	if u.activeRequests[url] > 0 {
		u.activeRequests[url]--
	}
}

func (u *Upstream) GetActiveRequestCount(url string) int64 {
	u.activeReqMu.RLock()
	defer u.activeReqMu.RUnlock()
	return u.activeRequests[url]
}
