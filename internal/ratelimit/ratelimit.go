package ratelimit

import (
	"sync"
	"time"
)

// CurrentDomainLimiter holds the per-domain rate limits for non-SPARQL HTTP
// requests. It is set up during build and is nil when no limits apply.
var CurrentDomainLimiter *DomainLimiter

// Limiter enforces a minimum interval between successive calls to Wait.
type Limiter struct {
	mu       sync.Mutex
	interval time.Duration
	last     time.Time
}

// NewLimiter returns a Limiter enforcing the given minimum interval between
// requests. A non-positive interval disables rate limiting.
func NewLimiter(interval time.Duration) *Limiter {
	return &Limiter{interval: interval}
}

// Wait blocks until enough time has passed since the previous call to satisfy
// the configured interval. A nil Limiter or non-positive interval is a no-op.
func (l *Limiter) Wait() {
	if l == nil || l.interval <= 0 {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.last.IsZero() {
		if remaining := l.interval - time.Since(l.last); remaining > 0 {
			time.Sleep(remaining)
		}
	}
	l.last = time.Now()
}

// DomainLimiter holds a separate Limiter per host.
type DomainLimiter struct {
	limiters map[string]*Limiter
}

// NewDomainLimiter returns a DomainLimiter built from a map of host to minimum
// request interval.
func NewDomainLimiter(intervals map[string]time.Duration) *DomainLimiter {
	limiters := make(map[string]*Limiter, len(intervals))
	for host, interval := range intervals {
		limiters[host] = NewLimiter(interval)
	}
	return &DomainLimiter{limiters: limiters}
}

// Wait applies the rate limit configured for the given host, if any. A nil
// DomainLimiter or an unconfigured host is a no-op.
func (d *DomainLimiter) Wait(host string) {
	if d == nil {
		return
	}
	d.limiters[host].Wait()
}
