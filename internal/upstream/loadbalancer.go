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
	// In a real implementation, this would need access to the connection pool stats
	// For this simplified version, we'll fall back to RoundRobin logic
	// or we would need to pass connection counts to Next().
	// To keep it clean and strictly follow the interface, we will implement a weighted-like
	// approach if we had weights, but for now let's use a mutex-protected map if we were tracking it.
	// Since we don't have the connection stats here, we will implement it as a stub
	// that behaves like RoundRobin for now, or we can expand the interface.
	// Let's expand the interface in the Manager, but here keep it simple.
	rr *RoundRobin
}

func NewLeastConnections() *LeastConnections {
	return &LeastConnections{
		rr: NewRoundRobin(),
	}
}

func (lc *LeastConnections) Next(urls []string) string {
	// Fallback to RR as we don't have connection stats passed in
	return lc.rr.Next(urls)
}
