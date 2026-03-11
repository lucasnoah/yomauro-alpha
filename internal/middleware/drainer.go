package middleware

import (
	"context"
	"log/slog"
	"net/http"
)

// Middleware returns an HTTP middleware that tracks in-flight requests.
// When draining, it rejects new requests with HTTP 503 and a Connection: close header.
func (d *Drainer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hold the mutex while checking the flag and incrementing the
		// counter. Drain holds the same mutex before calling wg.Wait, so
		// it is impossible for wg.Add and wg.Wait to execute concurrently.
		d.mu.Lock()
		if d.draining.Load() {
			d.mu.Unlock()
			w.Header().Set("Connection", "close")
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		d.wg.Add(1)
		d.mu.Unlock()
		defer d.wg.Done()
		next.ServeHTTP(w, r)
	})
}

// Drain marks the server as draining and blocks until all in-flight
// requests complete or the context is canceled.
func (d *Drainer) Drain(ctx context.Context) error {
	// Set the flag under the mutex so that any middleware goroutine that
	// acquires the lock after this point sees draining=true and skips
	// wg.Add. Once the lock is released, no new wg.Add calls can occur,
	// making the subsequent wg.Wait race-free.
	d.mu.Lock()
	d.draining.Store(true)
	d.mu.Unlock()
	slog.Info("draining connections")
	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		slog.Info("all connections drained")
		return nil
	case <-ctx.Done():
		slog.Warn("drain timeout exceeded")
		return ctx.Err()
	}
}

// Draining reports whether the server is in draining mode.
func (d *Drainer) Draining() bool {
	return d.draining.Load()
}
