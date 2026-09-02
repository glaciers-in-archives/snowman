package ratelimit

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// waits replaces the sleep function with one recording the durations it is
// asked to wait, the returned function restores the original.
func waits(recorded *[]time.Duration) func() {
	original := sleep
	sleep = func(duration time.Duration) { *recorded = append(*recorded, duration) }
	return func() { sleep = original }
}

// do issues a request against the given test server through Do.
func do(t *testing.T, server *httptest.Server, method string, body io.Reader) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, server.URL, body)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := Do(server.Client(), req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestParseRetryAfter(t *testing.T) {
	reference := time.Date(2023, 5, 17, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		value    string
		expected time.Duration
		ok       bool
	}{
		{"30", 30 * time.Second, true},
		{"-30", 0, false},
		{"", 0, false},
		{"soon", 0, false},
		{"Wed, 17 May 2023 12:01:00 GMT", time.Minute, true},
		{"Wed, 17 May 2023 11:00:00 GMT", 0, true},
	}

	for _, test := range tests {
		duration, ok := parseRetryAfter(test.value, reference)
		if ok != test.ok || (ok && duration != test.expected) {
			t.Errorf("parseRetryAfter(%q) = %v, %v, want %v, %v", test.value, duration, ok, test.expected, test.ok)
		}
	}
}

func TestDoRespectsRetryAfter(t *testing.T) {
	var waited []time.Duration
	defer waits(&waited)()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("results"))
	}))
	defer server.Close()

	resp := do(t, server, "GET", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK || calls != 2 {
		t.Fatalf("expected the retry to succeed, got HTTP %d after %d calls", resp.StatusCode, calls)
	}
	if len(waited) != 1 || waited[0] != 2*time.Second {
		t.Errorf("expected to wait the requested 2s, got %v", waited)
	}
	if body, _ := io.ReadAll(resp.Body); string(body) != "results" {
		t.Errorf("expected the body of the last response, got %q", string(body))
	}
}

func TestDoBacksOffWithoutRetryAfter(t *testing.T) {
	var waited []time.Duration
	defer waits(&waited)()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	// The rate limited response is returned once we run out of retries.
	resp := do(t, server, "GET", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests || calls != maxRetries+1 {
		t.Fatalf("expected %d calls ending in HTTP 429, got %d calls ending in HTTP %d", maxRetries+1, calls, resp.StatusCode)
	}
	if len(waited) != 3 || waited[0] != time.Second || waited[1] != 2*time.Second || waited[2] != 4*time.Second {
		t.Errorf("expected the wait to grow from 1s, got %v", waited)
	}
}

func TestDoResendsRequestBodies(t *testing.T) {
	defer waits(&[]time.Duration{})()

	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		bodies = append(bodies, string(body))

		if len(bodies) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	resp := do(t, server, "POST", strings.NewReader("query=SELECT%20*"))
	defer resp.Body.Close()

	if len(bodies) != 2 || bodies[0] != bodies[1] {
		t.Errorf("expected the retry to carry the same body, got %q", bodies)
	}
}

func TestDoGivesUpWhenAskedToWaitTooLong(t *testing.T) {
	var waited []time.Duration
	defer waits(&waited)()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	resp := do(t, server, "GET", nil)
	defer resp.Body.Close()

	if calls != 1 || len(waited) != 0 {
		t.Errorf("expected the request not to be retried, got %d calls and the waits %v", calls, waited)
	}
}

func TestDoOnlyRetriesRateLimitedResponses(t *testing.T) {
	defer waits(&[]time.Duration{})()

	// HTTP 503 is only rate limiting when it carries a Retry-After header.
	for _, status := range []int{http.StatusInternalServerError, http.StatusServiceUnavailable} {
		var calls int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.WriteHeader(status)
		}))

		do(t, server, "GET", nil).Body.Close()
		if calls != 1 {
			t.Errorf("expected HTTP %d not to be retried, got %d calls", status, calls)
		}
		server.Close()
	}
}

func TestDoRespectsTheURLRule(t *testing.T) {
	defer waits(&[]time.Duration{})()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	original := CurrentURLLimiter
	defer func() { CurrentURLLimiter = original }()
	CurrentURLLimiter = NewURLLimiter(map[string]time.Duration{"127.0.0.1": 50 * time.Millisecond})

	// The first attempt goes through right away, the retry waits for the limit.
	start := time.Now()
	do(t, server, "GET", nil).Body.Close()

	if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
		t.Errorf("expected the retry to respect the rate limit, took %v", elapsed)
	}
}
