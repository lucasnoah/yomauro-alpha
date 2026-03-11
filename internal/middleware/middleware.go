// Package middleware provides HTTP middleware for the yomauro API.
package middleware

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// requestIDKey is the context key for the request ID value.
type requestIDKey struct{}

// RequestID extracts the request ID from the context. Returns an empty
// string if no request ID is present.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// RequestIDMiddleware returns an http.Handler middleware that assigns a
// unique request ID to each request. If the incoming request carries an
// X-Request-ID header, that value is reused; otherwise a new ID is
// generated. The ID is stored in the request context and set on the
// response as the X-Request-ID header.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return requestIDMiddleware(next)
}

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
	mu       sync.Mutex
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
