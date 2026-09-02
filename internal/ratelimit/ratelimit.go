// Package ratelimit issues HTTP requests while respecting the rate limiting
// responses of the server on the other end. Instead of failing a build the
// moment an endpoint asks us to slow down, requests are retried after the
// delay requested by the server or, if none is given, after an exponentially
// increasing backoff.
package ratelimit

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Config holds the rate limiting options used when issuing requests.
type Config struct {
	// MaxRetries is the number of times a rate limited request is retried,
	// zero disables retrying.
	MaxRetries int
	// InitialBackoff is the duration waited before the first retry when the
	// server doesn't tell us how long to wait, it's doubled for each retry.
	InitialBackoff time.Duration
	// MaxWait is the longest duration we wait before a single retry. Backoffs
	// are capped by it while a longer delay requested by the server makes us
	// give up as we can't respect it.
	MaxWait time.Duration
}

// DefaultConfig returns the rate limiting options used unless the project
// configures something else.
func DefaultConfig() Config {
	return Config{
		MaxRetries:     3,
		InitialBackoff: time.Second,
		MaxWait:        60 * time.Second,
	}
}

// sleep is replaced in tests.
var sleep = time.Sleep

// IsRateLimited reports whether the given response is the server asking us to
// slow down. HTTP 503 is only treated as rate limiting when it carries a
// Retry-After header as it's otherwise used for all kinds of outages.
func IsRateLimited(resp *http.Response) bool {
	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		return true
	case http.StatusServiceUnavailable:
		return resp.Header.Get("Retry-After") != ""
	}
	return false
}

// ParseRetryAfter interprets a Retry-After header, which either holds a number
// of seconds or a HTTP date. The boolean return value reports whether the
// header could be interpreted.
func ParseRetryAfter(value string, reference time.Time) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds < 0 {
			return 0, false
		}
		return time.Duration(seconds) * time.Second, true
	}

	if date, err := http.ParseTime(value); err == nil {
		wait := date.Sub(reference)
		if wait < 0 {
			wait = 0
		}
		return wait, true
	}

	return 0, false
}

// backoff returns the duration to wait before the given retry, the first retry
// being attempt zero.
func backoff(config Config, attempt int) time.Duration {
	wait := config.InitialBackoff
	for i := 0; i < attempt; i++ {
		if wait >= config.MaxWait/2 {
			return config.MaxWait
		}
		wait *= 2
	}
	if wait > config.MaxWait {
		return config.MaxWait
	}
	return wait
}

// rewind restores the request body so that the request can be issued again,
// the http client closes it after each call.
func rewind(req *http.Request) bool {
	if req.Body == nil {
		return true
	}
	if req.GetBody == nil {
		return false
	}

	body, err := req.GetBody()
	if err != nil {
		return false
	}
	req.Body = body
	return true
}

// drain reads and closes the body of a response we are about to discard so
// that the underlying connection can be reused.
func drain(resp *http.Response) {
	io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	resp.Body.Close()
}

// Do issues the given request, waiting and retrying for as long as the server
// responds with a rate limiting status. The last response is returned once the
// server stops rate limiting us or once we run out of retries, meaning that
// callers still have to check the status code of the returned response.
func Do(client *http.Client, req *http.Request, config Config) (*http.Response, error) {
	if client == nil {
		client = http.DefaultClient
	}

	for attempt := 0; ; attempt++ {
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if !IsRateLimited(resp) || attempt >= config.MaxRetries {
			return resp, nil
		}

		wait, requested := ParseRetryAfter(resp.Header.Get("Retry-After"), time.Now())
		if requested {
			if wait > config.MaxWait {
				fmt.Println("Rate limited by " + req.URL.Host + " which asked us to wait " + wait.String() +
					", longer than the configured rate_limit.max_wait of " + config.MaxWait.String() + ". Giving up.")
				return resp, nil
			}
		} else {
			wait = backoff(config, attempt)
		}

		if !rewind(req) {
			fmt.Println("Rate limited by " + req.URL.Host + " but the request body can't be sent again. Giving up.")
			return resp, nil
		}

		drain(resp)

		fmt.Println("Rate limited (HTTP " + strconv.Itoa(resp.StatusCode) + ") by " + req.URL.Host +
			", waiting " + wait.String() + " before retry " + strconv.Itoa(attempt+1) + " of " + strconv.Itoa(config.MaxRetries) + ".")
		sleep(wait)
	}
}
