package upstream

import (
	"sync/atomic"
)

type LoadBalancer interface {
	Next(urls []string) string
}

type RoundRobin struct {
	counter uint64
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

func (rb *RoundRobin) Next(urls []string) string {
	if len(urls) == 0 {
		return ""
	}
	count := atomic.AddUint64(&rb.counter, 1)
	return urls[(int(count)-1)%len(urls)]
}

// LeastConnections would require tracking active connections,
// which is more complex to implement without a central state manager.
// For now, we will implement a placeholder or a simple random if needed,
// but the requirement asked for RoundRobin and LeastConnections.
// We'll stick to RoundRobin as the default and implement a basic structure for LeastConn.

type LeastConnections struct {
	rr       *RoundRobin
	getStats func(string) int64
}

func NewLeastConnections(getStats func(string) int64) *LeastConnections {
	return &LeastConnections{
		rr:       NewRoundRobin(),
		getStats: getStats,
	}
}

func (lc *LeastConnections) Next(urls []string) string {
	if len(urls) == 0 {
		return ""
	}

	// If no stats function provided, fallback to RR
	if lc.getStats == nil {
		return lc.rr.Next(urls)
	}

	var best string
	var min int64 = -1

	for _, u := range urls {
		count := lc.getStats(u)
		if min == -1 || count < min {
			min = count
			best = u
		}
	}

	// If we found a best candidate
	if best != "" {
		return best
	}

	// Fallback
	return lc.rr.Next(urls)
}
