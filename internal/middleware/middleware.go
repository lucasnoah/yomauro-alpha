// Package middleware provides HTTP middleware for the yomauro API.
package middleware

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// RateLimiter tracks request counts per key within a sliding window
// and rejects requests that exceed the configured limit.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*entry
	limit   int
	window  time.Duration
	now     func() time.Time // for testing
}

type entry struct {
	timestamps []time.Time
}

// Drainer tracks in-flight HTTP requests and supports graceful connection
// draining. During shutdown, it rejects new requests with HTTP 503 and
// waits for in-flight requests to complete.
type Drainer struct {
	wg       sync.WaitGroup
	draining atomic.Bool
}

// NewDrainer creates a Drainer ready to track in-flight requests.
func NewDrainer() *Drainer {
	return &Drainer{}
}

// NewRateLimiter creates a RateLimiter that allows at most limit requests
// per key within the given window duration.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		entries: make(map[string]*entry),
		limit:   limit,
		window:  window,
		now:     time.Now,
	}
}

// Middleware returns an http.Handler middleware that rate-limits requests
// by client IP address. Requests exceeding the limit receive HTTP 429
// with a JSON error body.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return rl.middleware(next)
}
