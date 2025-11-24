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
		if uCfg.LoadBalancer == "least_connections" {
			// lb = NewLeastConnections() // Placeholder for now
		}

		u := &Upstream{
			Name:          name,
			URLs:          uCfg.URLs,
			LoadBalancer:  lb,
			HealthChecker: NewHealthChecker(uCfg.URLs, 10*time.Second), // Default interval
			CircuitBreaker: NewCircuitBreaker(
				uCfg.CircuitBreaker.FailureThreshold,
				time.Duration(uCfg.CircuitBreaker.ResetTimeoutMs)*time.Millisecond,
			),
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
