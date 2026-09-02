package ratelimit

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// How Snowman recovers when a server rate limits it anyway. Projects control
// rate limiting through the per-URL intervals, these are not configurable.
var (
	maxRetries  = 3
	initialWait = time.Second
	maxWait     = 60 * time.Second
	sleep       = time.Sleep
)

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

// parseRetryAfter interprets a Retry-After header, which either holds a number
// of seconds or a HTTP date. The boolean return value reports whether the
// header could be interpreted.
func parseRetryAfter(value string, reference time.Time) (time.Duration, bool) {
	value = strings.TrimSpace(value)

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

// Do issues the given request while respecting the rate limit configured for
// its host and retrying for as long as the server rate limits us. The last
// response is returned once the server stops rate limiting us or once we run
// out of retries, callers still have to check its status code.
func Do(client *http.Client, req *http.Request) (*http.Response, error) {
	wait := initialWait

	for attempt := 0; ; attempt++ {
		CurrentURLLimiter.Wait(req.URL)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if !IsRateLimited(resp) || attempt >= maxRetries {
			return resp, nil
		}

		// Respect the delay the server asks for, otherwise back off with a
		// wait that grows for each retry.
		delay := wait
		if requested, ok := parseRetryAfter(resp.Header.Get("Retry-After"), time.Now()); ok {
			if requested > maxWait {
				fmt.Println("Rate limited by " + req.URL.Host + " which asked us to wait " + requested.String() +
					", longer than Snowman is willing to wait for a single retry. Giving up.")
				return resp, nil
			}
			delay = requested
		} else if wait *= 2; wait > maxWait {
			wait = maxWait
		}

		// The client closes the request body after each call, restore it so
		// that the request can be issued again.
		if req.Body != nil {
			if req.GetBody == nil {
				return resp, nil
			}
			body, err := req.GetBody()
			if err != nil {
				return resp, nil
			}
			req.Body = body
		}

		// Discard the body we won't read so that the connection can be reused.
		io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		resp.Body.Close()

		fmt.Println("Rate limited (HTTP " + strconv.Itoa(resp.StatusCode) + ") by " + req.URL.Host +
			", waiting " + delay.String() + " before retry " + strconv.Itoa(attempt+1) + " of " + strconv.Itoa(maxRetries) + ".")
		sleep(delay)
	}
}
