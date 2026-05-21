package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterEnforcesInterval(t *testing.T) {
	limiter := NewLimiter(50 * time.Millisecond)

	start := time.Now()
	for i := 0; i < 3; i++ {
		limiter.Wait()
	}
	elapsed := time.Since(start)

	// The first call returns immediately, the next two each wait ~50ms.
	if elapsed < 90*time.Millisecond {
		t.Errorf("expected at least 90ms for 3 rate-limited calls, got %v", elapsed)
	}
}

func TestLimiterZeroIntervalIsNoOp(t *testing.T) {
	limiter := NewLimiter(0)

	start := time.Now()
	for i := 0; i < 100; i++ {
		limiter.Wait()
	}

	if elapsed := time.Since(start); elapsed > 20*time.Millisecond {
		t.Errorf("expected unlimited limiter to be near-instant, got %v", elapsed)
	}
}

func TestNilLimiterIsNoOp(t *testing.T) {
	var limiter *Limiter
	limiter.Wait()
}

func TestDomainLimiterPerHost(t *testing.T) {
	domainLimiter := NewDomainLimiter(map[string]time.Duration{
		"slow.example": 50 * time.Millisecond,
	})

	// A configured host is rate limited.
	start := time.Now()
	domainLimiter.Wait("slow.example")
	domainLimiter.Wait("slow.example")
	if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
		t.Errorf("expected configured host to be rate limited, got %v", elapsed)
	}

	// An unconfigured host is not rate limited.
	start = time.Now()
	for i := 0; i < 100; i++ {
		domainLimiter.Wait("fast.example")
	}
	if elapsed := time.Since(start); elapsed > 20*time.Millisecond {
		t.Errorf("expected unconfigured host to be unlimited, got %v", elapsed)
	}
}

func TestNilDomainLimiterIsNoOp(t *testing.T) {
	var domainLimiter *DomainLimiter
	domainLimiter.Wait("anything.example")
}
