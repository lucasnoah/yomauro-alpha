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
		// Add before checking the flag to avoid a race where Drain's
		// wg.Wait returns before this request is tracked.
		d.wg.Add(1)
		if d.draining.Load() {
			d.wg.Done()
			w.Header().Set("Connection", "close")
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		defer d.wg.Done()
		next.ServeHTTP(w, r)
	})
}

// Drain marks the server as draining and blocks until all in-flight
// requests complete or the context is canceled.
func (d *Drainer) Drain(ctx context.Context) error {
	d.draining.Store(true)
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
