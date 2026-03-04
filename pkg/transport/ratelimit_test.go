package transport_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	transport "github.com/mutablelogic/go-client/pkg/transport"
	assert "github.com/stretchr/testify/assert"
)

func TestNewRateLimit_NilParentUsesDefault(t *testing.T) {
	assert := assert.New(t)
	r := transport.NewRateLimit(nil, 1)
	assert.NotNil(r)
	var _ http.RoundTripper = r
}
func TestNewRateLimit_WithParent(t *testing.T) {
	assert := assert.New(t)
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "ok"), nil
	})
	r := transport.NewRateLimit(inner, 10)
	assert.NotNil(r)
}
func TestRateLimit_ZeroRateNoThrottle(t *testing.T) {
	assert := assert.New(t)
	var calls int32
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return stubResp(200, "text/plain", "ok"), nil
	})
	rl := transport.NewRateLimit(inner, 0)
	const n = 5
	for i := 0; i < n; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		resp, err := rl.RoundTrip(req)
		assert.NoError(err)
		resp.Body.Close()
	}
	assert.Equal(int32(n), atomic.LoadInt32(&calls))
}
func TestRateLimit_ThrottlesRequests(t *testing.T) {
	const rate = float32(10)
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "ok"), nil
	})
	rl := transport.NewRateLimit(inner, rate)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := rl.RoundTrip(req)
	assert.NoError(t, err)
	resp.Body.Close()
	start := time.Now()
	req = httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err = rl.RoundTrip(req)
	elapsed := time.Since(start)
	assert.NoError(t, err)
	resp.Body.Close()
	minDelay := time.Duration(float64(time.Second) / float64(rate))
	assert.GreaterOrEqual(t, elapsed, minDelay/2)
}
func TestRateLimit_ContextCancellation(t *testing.T) {
	assert := assert.New(t)
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return stubResp(200, "text/plain", "ok"), nil
	})
	rl := transport.NewRateLimit(inner, 0.1)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	resp, err := rl.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, _ = http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com/", nil)
	start := time.Now()
	_, err = rl.RoundTrip(req)
	elapsed := time.Since(start)
	assert.ErrorIs(err, context.Canceled)
	assert.Less(elapsed, 2*time.Second)
}
func TestRateLimit_ForwardsRequest(t *testing.T) {
	assert := assert.New(t)
	var gotMethod string
	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotMethod = req.Method
		return stubResp(204, "text/plain", ""), nil
	})
	rl := transport.NewRateLimit(inner, 0)
	req := httptest.NewRequest(http.MethodDelete, "http://example.com/resource", nil)
	resp, err := rl.RoundTrip(req)
	assert.NoError(err)
	resp.Body.Close()
	assert.Equal(http.MethodDelete, gotMethod)
	assert.Equal(204, resp.StatusCode)
}

// TestRateLimit_ConcurrentRequestsAreThrottled verifies that concurrent
// callers are assigned distinct time slots and do not all proceed at once.
// Prior to the slot-reservation fix, N goroutines would all read the same
// t.ts, compute the same delay, sleep in parallel, and fire simultaneously —
// violating the configured rate limit.
func TestRateLimit_ConcurrentRequestsAreThrottled(t *testing.T) {
	const rate = float32(10) // 10 req/s → ~100ms interval
	const n = 3

	var mu sync.Mutex
	var timestamps []time.Time

	inner := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		mu.Lock()
		timestamps = append(timestamps, time.Now())
		mu.Unlock()
		return stubResp(200, "text/plain", "ok"), nil
	})
	rl := transport.NewRateLimit(inner, rate)

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
			resp, err := rl.RoundTrip(req)
			if err == nil {
				resp.Body.Close()
			}
		}()
	}
	close(start)
	wg.Wait()

	assert.Equal(t, n, len(timestamps), "all requests should complete")

	// Sort by arrival time and verify each adjacent pair is at least ~80% of
	// the configured interval apart (20% slack for scheduler jitter).
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i].Before(timestamps[j]) })
	minInterval := time.Duration(float64(time.Second)/float64(rate)) * 8 / 10
	for i := 1; i < len(timestamps); i++ {
		gap := timestamps[i].Sub(timestamps[i-1])
		assert.GreaterOrEqualf(t, gap, minInterval,
			"requests %d and %d fired too close together: %v (want >= %v)", i-1, i, gap, minInterval)
	}
}
