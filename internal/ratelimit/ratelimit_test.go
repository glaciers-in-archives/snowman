package ratelimit

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// recordSleeps replaces the sleep function with one recording the durations it
// is asked to wait, the returned function restores the original.
func recordSleeps(waited *[]time.Duration) func() {
	original := sleep
	sleep = func(duration time.Duration) {
		*waited = append(*waited, duration)
	}
	return func() { sleep = original }
}

func testConfig() Config {
	return Config{
		MaxRetries:     3,
		InitialBackoff: time.Second,
		MaxWait:        60 * time.Second,
	}
}

func TestParseRetryAfter(t *testing.T) {
	reference := time.Date(2023, 5, 17, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		value    string
		expected time.Duration
		ok       bool
	}{
		{"seconds", "30", 30 * time.Second, true},
		{"zero seconds", "0", 0, true},
		{"padded seconds", " 30 ", 30 * time.Second, true},
		{"negative seconds", "-30", 0, false},
		{"empty", "", 0, false},
		{"garbage", "soon", 0, false},
		{"http date", "Wed, 17 May 2023 12:01:00 GMT", time.Minute, true},
		{"http date in the past", "Wed, 17 May 2023 11:00:00 GMT", 0, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			duration, ok := ParseRetryAfter(test.value, reference)
			if ok != test.ok {
				t.Fatalf("Expected ok to be %v, got %v", test.ok, ok)
			}
			if ok && duration != test.expected {
				t.Fatalf("Expected %v, got %v", test.expected, duration)
			}
		})
	}
}

func TestBackoffDoubles(t *testing.T) {
	config := Config{InitialBackoff: time.Second, MaxWait: 60 * time.Second}

	expected := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}
	for attempt, want := range expected {
		if got := backoff(config, attempt); got != want {
			t.Fatalf("Expected backoff %v for attempt %d, got %v", want, attempt, got)
		}
	}
}

func TestBackoffIsCappedByMaxWait(t *testing.T) {
	config := Config{InitialBackoff: time.Second, MaxWait: 5 * time.Second}

	if got := backoff(config, 10); got != 5*time.Second {
		t.Fatalf("Expected the backoff to be capped at 5s, got %v", got)
	}

	config = Config{InitialBackoff: 10 * time.Second, MaxWait: 5 * time.Second}
	if got := backoff(config, 0); got != 5*time.Second {
		t.Fatalf("Expected the backoff to be capped at 5s, got %v", got)
	}
}

func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		retryAfter string
		expected   bool
	}{
		{"too many requests", http.StatusTooManyRequests, "", true},
		{"service unavailable with retry after", http.StatusServiceUnavailable, "10", true},
		{"service unavailable without retry after", http.StatusServiceUnavailable, "", false},
		{"internal server error", http.StatusInternalServerError, "10", false},
		{"ok", http.StatusOK, "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: test.statusCode, Header: http.Header{}}
			if test.retryAfter != "" {
				resp.Header.Set("Retry-After", test.retryAfter)
			}
			if got := IsRateLimited(resp); got != test.expected {
				t.Fatalf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestDoRetriesRateLimitedRequests(t *testing.T) {
	var waited []time.Duration
	defer recordSleeps(&waited)()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls <= 2 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("results"))
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := Do(server.Client(), req, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected the request to eventually succeed, got HTTP %d", resp.StatusCode)
	}
	if calls != 3 {
		t.Fatalf("Expected 3 calls, got %d", calls)
	}
	if len(waited) != 2 || waited[0] != 2*time.Second || waited[1] != 2*time.Second {
		t.Fatalf("Expected to wait the requested 2s twice, got %v", waited)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "results" {
		t.Fatalf("Expected the body of the last response, got %q", string(body))
	}
}

func TestDoFallsBackToBackoffWithoutRetryAfter(t *testing.T) {
	var waited []time.Duration
	defer recordSleeps(&waited)()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := Do(server.Client(), req, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("Expected the last response to be returned, got HTTP %d", resp.StatusCode)
	}
	if calls != 4 {
		t.Fatalf("Expected the request to be retried 3 times, got %d calls", calls)
	}

	expected := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second}
	if len(waited) != len(expected) {
		t.Fatalf("Expected %d waits, got %v", len(expected), waited)
	}
	for i, want := range expected {
		if waited[i] != want {
			t.Fatalf("Expected wait %d to be %v, got %v", i, want, waited[i])
		}
	}
}

func TestDoResendsRequestBodies(t *testing.T) {
	defer recordSleeps(&[]time.Duration{})()

	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		bodies = append(bodies, string(body))

		if len(bodies) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	req, err := http.NewRequest("POST", server.URL, strings.NewReader("query=SELECT%20*"))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := Do(server.Client(), req, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected the retried request to succeed, got HTTP %d", resp.StatusCode)
	}
	if len(bodies) != 2 {
		t.Fatalf("Expected 2 calls, got %d", len(bodies))
	}
	if bodies[0] != bodies[1] {
		t.Fatalf("Expected the retry to carry the same body, got %q and %q", bodies[0], bodies[1])
	}
}

func TestDoGivesUpWhenAskedToWaitLongerThanMaxWait(t *testing.T) {
	var waited []time.Duration
	defer recordSleeps(&waited)()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := Do(server.Client(), req, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("Expected the rate limited response to be returned, got HTTP %d", resp.StatusCode)
	}
	if calls != 1 {
		t.Fatalf("Expected the request not to be retried, got %d calls", calls)
	}
	if len(waited) != 0 {
		t.Fatalf("Expected no waiting, got %v", waited)
	}
}

func TestDoWithoutRetries(t *testing.T) {
	defer recordSleeps(&[]time.Duration{})()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	config := testConfig()
	config.MaxRetries = 0

	resp, err := Do(server.Client(), req, config)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if calls != 1 {
		t.Fatalf("Expected a single call when retrying is disabled, got %d", calls)
	}
}

func TestDoIgnoresOtherStatuses(t *testing.T) {
	defer recordSleeps(&[]time.Duration{})()

	statuses := []int{http.StatusOK, http.StatusInternalServerError, http.StatusNotFound, http.StatusServiceUnavailable}

	for _, status := range statuses {
		var calls int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.WriteHeader(status)
		}))

		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			server.Close()
			t.Fatal(err)
		}

		resp, err := Do(server.Client(), req, testConfig())
		if err != nil {
			server.Close()
			t.Fatal(err)
		}
		resp.Body.Close()

		if calls != 1 {
			t.Fatalf("Expected HTTP %d not to be retried, got %d calls", status, calls)
		}
		server.Close()
	}
}
