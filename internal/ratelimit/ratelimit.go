package ratelimit

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// CurrentURLLimiter holds the per-URL-pattern rate limits applied to the HTTP
// requests Snowman makes. It is set up during build and is nil when no limits
// apply.
var CurrentURLLimiter *URLLimiter

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

// URLLimiter holds a separate Limiter per URL pattern. Requests matching
// the same pattern share its Limiter, so a pattern limits the requests made to
// the service behind it rather than to each of its URLs.
type URLLimiter struct {
	limiters []patternLimiter
}

type patternLimiter struct {
	pattern string
	matcher *regexp.Regexp
	limiter *Limiter
}

// isPattern reports whether the given URL pattern contains wildcards.
func isPattern(pattern string) bool {
	return strings.ContainsAny(pattern, "*?")
}

// normalizePattern prepares a URL pattern for matching. The scheme is dropped
// as it isn't matched against, and a pattern without a path is expanded so
// that it matches every URL on its host.
func normalizePattern(pattern string) string {
	pattern = strings.ToLower(pattern)
	pattern = strings.TrimPrefix(pattern, "https://")
	pattern = strings.TrimPrefix(pattern, "http://")

	if !strings.Contains(pattern, "/") {
		pattern += "/*"
	}
	return pattern
}

// compilePattern turns a normalized URL pattern into a matcher where "*"
// matches any sequence of characters and "?" a single one.
func compilePattern(pattern string) *regexp.Regexp {
	expression := strings.NewReplacer(`\*`, `.*`, `\?`, `.`).Replace(regexp.QuoteMeta(pattern))
	return regexp.MustCompile("^" + expression + "$")
}

// matchTarget returns the string a URL is matched against, its host name
// followed by its path. Neither the scheme nor the port are matched against.
func matchTarget(location *url.URL) string {
	path := location.EscapedPath()
	if path == "" {
		path = "/"
	}
	return strings.ToLower(location.Hostname()) + path
}

// NewURLLimiter returns a URLLimiter built from a map of URL pattern to
// minimum request interval. The patterns are ordered so that the most specific
// match wins: patterns without wildcards before those with, longer patterns
// before shorter ones.
func NewURLLimiter(intervals map[string]time.Duration) *URLLimiter {
	limiters := make([]patternLimiter, 0, len(intervals))
	for pattern, interval := range intervals {
		normalized := normalizePattern(pattern)
		limiters = append(limiters, patternLimiter{normalized, compilePattern(normalized), NewLimiter(interval)})
	}

	sort.Slice(limiters, func(i, j int) bool {
		a, b := limiters[i].pattern, limiters[j].pattern
		if isPattern(a) != isPattern(b) {
			return !isPattern(a)
		}
		if len(a) != len(b) {
			return len(a) > len(b)
		}
		return a < b
	})

	return &URLLimiter{limiters: limiters}
}

// Wait applies the rate limit of the first pattern matching the given URL, if
// any. A nil URLLimiter or an unmatched URL is a no-op.
func (u *URLLimiter) Wait(location *url.URL) {
	if u == nil {
		return
	}

	target := matchTarget(location)
	for _, entry := range u.limiters {
		if entry.matcher.MatchString(target) {
			entry.limiter.Wait()
			return
		}
	}
}
