package transport

import (
	"net/http"
	"sync"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// RateLimitTransport is an http.RoundTripper middleware that enforces a
// maximum request rate (requests per second). It sleeps before forwarding
// each request when necessary, and respects context cancellation during
// the sleep.
type RateLimitTransport struct {
	http.RoundTripper
	mu   sync.Mutex
	rate float32
	ts   time.Time
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewRateLimit wraps parent in a RateLimitTransport that allows at most
// rate requests per second. A rate of 0 disables throttling.
// If parent is nil, http.DefaultTransport is used.
func NewRateLimit(parent http.RoundTripper, rate float32) *RateLimitTransport {
	if parent == nil {
		parent = http.DefaultTransport
	}
	return &RateLimitTransport{RoundTripper: parent, rate: rate}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// RoundTrip implements http.RoundTripper. When a rate limit is configured it
// sleeps until the inter-request interval has elapsed, then records the send
// time and forwards the request. The sleep is cancelled early if ctx is done.
func (t *RateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.rate > 0 {
		t.mu.Lock()
		var delay time.Duration
		if !t.ts.IsZero() {
			next := t.ts.Add(time.Duration(float32(time.Second) / t.rate))
			delay = time.Until(next)
		}
		if delay > 0 {
			t.mu.Unlock()
			timer := time.NewTimer(delay)
			select {
			case <-req.Context().Done():
				timer.Stop()
				return nil, req.Context().Err()
			case <-timer.C:
			}
			t.mu.Lock()
		}
		t.ts = time.Now()
		t.mu.Unlock()
	}
	rt := t.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}
