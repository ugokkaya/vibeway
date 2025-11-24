package upstream

import (
	"net/http"
	"sync"
	"time"

	"vibeway/pkg/logger"
)

type HealthChecker struct {
	urls        []string
	interval    time.Duration
	healthyURLs []string
	mu          sync.RWMutex
	stop        chan struct{}
}

func NewHealthChecker(urls []string, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		urls:        urls,
		interval:    interval,
		healthyURLs: urls, // Assume all healthy initially
		stop:        make(chan struct{}),
	}
}

func (hc *HealthChecker) Start() {
	go func() {
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				hc.check()
			case <-hc.stop:
				return
			}
		}
	}()
}

func (hc *HealthChecker) Stop() {
	close(hc.stop)
}

func (hc *HealthChecker) check() {
	var healthy []string
	for _, url := range hc.urls {
		if hc.isHealthy(url) {
			healthy = append(healthy, url)
		} else {
			logger.Warn("Upstream unhealthy", map[string]interface{}{"url": url})
		}
	}

	hc.mu.Lock()
	hc.healthyURLs = healthy
	hc.mu.Unlock()
}

func (hc *HealthChecker) isHealthy(url string) bool {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	// Assuming a simple TCP connect or root GET for now.
	// Ideally this should be configurable (e.g., /health endpoint)
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func (hc *HealthChecker) GetHealthyURLs() []string {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	// Return a copy to be safe
	urls := make([]string, len(hc.healthyURLs))
	copy(urls, hc.healthyURLs)
	return urls
}
