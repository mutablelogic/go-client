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

// RoundTrip implements http.RoundTripper. When a rate limit is configured each
// call reserves its send-slot under the lock (advancing t.ts immediately) and
// then sleeps outside the lock until that slot arrives. Reserving the slot
// before releasing the lock ensures that concurrent callers are assigned
// distinct, strictly ordered slots rather than all sleeping for the same
// duration and proceeding together.
func (t *RateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.rate > 0 {
		t.mu.Lock()
		interval := time.Duration(float32(time.Second) / t.rate)
		var slot time.Time
		if t.ts.IsZero() {
			slot = time.Now() // first request: no wait
		} else {
			slot = t.ts.Add(interval) // queue behind last reservation
		}
		t.ts = slot // reserve this slot before releasing the lock
		t.mu.Unlock()

		if delay := time.Until(slot); delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-req.Context().Done():
				timer.Stop()
				return nil, req.Context().Err()
			case <-timer.C:
			}
		}
	}
	rt := t.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}
