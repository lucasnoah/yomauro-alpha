package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

// middleware is the implementation of the rate-limiting HTTP middleware.
func (rl *RateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !rl.allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "too many login attempts, try again later",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// allow checks whether a request from the given key is within the rate limit.
// It records the current timestamp and evicts expired entries.
func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	cutoff := now.Add(-rl.window)

	e, ok := rl.entries[key]
	if !ok {
		e = &entry{}
		rl.entries[key] = e
	}

	// Evict timestamps outside the window.
	valid := e.timestamps[:0]
	for _, ts := range e.timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	e.timestamps = valid

	if len(e.timestamps) >= rl.limit {
		return false
	}

	e.timestamps = append(e.timestamps, now)
	return true
}

// Cleanup removes entries that have no timestamps within the current window.
// Call this periodically to prevent unbounded memory growth.
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := rl.now().Add(-rl.window)
	for key, e := range rl.entries {
		valid := e.timestamps[:0]
		for _, ts := range e.timestamps {
			if ts.After(cutoff) {
				valid = append(valid, ts)
			}
		}
		if len(valid) == 0 {
			delete(rl.entries, key)
		} else {
			e.timestamps = valid
		}
	}
}

// clientIP extracts the client IP from the request, checking
// X-Forwarded-For first, then falling back to RemoteAddr.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For may contain multiple IPs; use the first (client).
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return strings.TrimSpace(xff[:i])
			}
		}
		return strings.TrimSpace(xff)
	}
	// RemoteAddr is "IP:port"; strip the port.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
