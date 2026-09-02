package ratelimit

import (
	"net/url"
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

func TestURLLimiterPerHost(t *testing.T) {
	urlLimiter := NewURLLimiter(map[string]time.Duration{
		"slow.example": 50 * time.Millisecond,
	})

	// A configured host is rate limited, whatever is requested from it.
	start := time.Now()
	wait(t, urlLimiter, "https://slow.example")
	wait(t, urlLimiter, "https://slow.example/sparql")
	if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
		t.Errorf("expected configured host to be rate limited, got %v", elapsed)
	}

	// An unconfigured host is not rate limited.
	start = time.Now()
	for i := 0; i < 100; i++ {
		wait(t, urlLimiter, "https://fast.example/sparql")
	}
	if elapsed := time.Since(start); elapsed > 20*time.Millisecond {
		t.Errorf("expected unconfigured host to be unlimited, got %v", elapsed)
	}
}

func TestURLLimiterMatchesPatterns(t *testing.T) {
	urlLimiter := NewURLLimiter(map[string]time.Duration{
		"*.slow.example": 50 * time.Millisecond,
	})

	// Hosts matching the pattern share the rate limit, whatever their path.
	start := time.Now()
	wait(t, urlLimiter, "https://api.slow.example/one")
	wait(t, urlLimiter, "https://images.slow.example/two/three")
	if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
		t.Errorf("expected matching hosts to share the rate limit, got %v", elapsed)
	}

	// The pattern doesn't match the domain itself or unrelated hosts.
	start = time.Now()
	wait(t, urlLimiter, "https://slow.example/one")
	wait(t, urlLimiter, "https://fast.example/two")
	if elapsed := time.Since(start); elapsed > 20*time.Millisecond {
		t.Errorf("expected unmatched hosts to be unlimited, got %v", elapsed)
	}
}

func TestURLLimiterPrefersTheMostSpecificPattern(t *testing.T) {
	urlLimiter := NewURLLimiter(map[string]time.Duration{
		"*":                             0,
		"slow.example/*":                50 * time.Millisecond,
		"slow.example/fast/*":           0,
		"slow.example/fast/slowest.ttl": 50 * time.Millisecond,
	})

	// The more specific path wins over the pattern covering the whole host.
	start := time.Now()
	wait(t, urlLimiter, "https://slow.example/fast/one.ttl")
	wait(t, urlLimiter, "https://slow.example/fast/two.ttl")
	if elapsed := time.Since(start); elapsed > 20*time.Millisecond {
		t.Errorf("expected the more specific path to win, got %v", elapsed)
	}

	// A pattern without wildcards wins over the paths covering it.
	start = time.Now()
	wait(t, urlLimiter, "https://slow.example/fast/slowest.ttl")
	wait(t, urlLimiter, "https://slow.example/fast/slowest.ttl")
	if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
		t.Errorf("expected the pattern without wildcards to win, got %v", elapsed)
	}

	// The host pattern wins over the catch-all for the rest of the host.
	start = time.Now()
	wait(t, urlLimiter, "https://slow.example/one.ttl")
	wait(t, urlLimiter, "https://slow.example/two.ttl")
	if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
		t.Errorf("expected the host pattern to win, got %v", elapsed)
	}
}

func TestNilURLLimiterIsNoOp(t *testing.T) {
	var urlLimiter *URLLimiter
	wait(t, urlLimiter, "https://anything.example")
}

// wait applies the limiter of the given URL.
func wait(t *testing.T, urlLimiter *URLLimiter, location string) {
	t.Helper()

	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatal(err)
	}
	urlLimiter.Wait(parsed)
}
