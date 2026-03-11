package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)

	handler := rl.Middleware(okHandler())

	for i := range 5 {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: got status %d, want %d", i+1, rr.Code, http.StatusOK)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)

	handler := rl.Middleware(okHandler())

	// Exhaust the limit.
	for range 5 {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "10.0.0.1:9999"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// 6th request should be blocked.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "10.0.0.1:9999"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("got status %d, want %d", rr.Code, http.StatusTooManyRequests)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := "too many login attempts, try again later"
	if body["error"] != want {
		t.Fatalf("got error %q, want %q", body["error"], want)
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	now := time.Now()
	rl := NewRateLimiter(2, time.Minute)
	rl.now = func() time.Time { return now }

	handler := rl.Middleware(okHandler())

	// Exhaust the limit.
	for range 2 {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.RemoteAddr = "10.0.0.2:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Advance time past the window.
	now = now.Add(61 * time.Second)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("after window reset: got status %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRateLimiter_IndependentPerIP(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	handler := rl.Middleware(okHandler())

	// First IP uses its single allowed request.
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "1.1.1.1:1000"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("first IP first request: got %d, want %d", rr.Code, http.StatusOK)
	}

	// Second IP should still be allowed.
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "2.2.2.2:2000"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("second IP first request: got %d, want %d", rr.Code, http.StatusOK)
	}

	// First IP is now blocked.
	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "1.1.1.1:1000"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("first IP second request: got %d, want %d", rr.Code, http.StatusTooManyRequests)
	}
}

func TestClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18, 150.172.238.178")

	got := clientIP(req)
	if got != "203.0.113.50" {
		t.Fatalf("got %q, want %q", got, "203.0.113.50")
	}
}

func TestClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.100:54321"

	got := clientIP(req)
	if got != "192.168.1.100" {
		t.Fatalf("got %q, want %q", got, "192.168.1.100")
	}
}

func TestCleanup_RemovesExpiredEntries(t *testing.T) {
	now := time.Now()
	rl := NewRateLimiter(5, time.Minute)
	rl.now = func() time.Time { return now }

	// Record a request.
	rl.allow("old-ip")

	// Advance past window.
	now = now.Add(2 * time.Minute)

	rl.Cleanup()

	rl.mu.Lock()
	_, exists := rl.entries["old-ip"]
	rl.mu.Unlock()

	if exists {
		t.Fatal("expected expired entry to be cleaned up")
	}
}

func TestCleanup_KeepsActiveEntries(t *testing.T) {
	now := time.Now()
	rl := NewRateLimiter(5, time.Minute)
	rl.now = func() time.Time { return now }

	rl.allow("active-ip")

	// Advance but stay within window.
	now = now.Add(30 * time.Second)
	rl.Cleanup()

	rl.mu.Lock()
	_, exists := rl.entries["active-ip"]
	rl.mu.Unlock()

	if !exists {
		t.Fatal("expected active entry to be kept")
	}
}

func TestClientIP_XForwardedFor_Whitespace(t *testing.T) {
	cases := []struct {
		xff  string
		want string
	}{
		{"203.0.113.50 , 70.41.3.18", "203.0.113.50"},   // trailing space before comma
		{"  203.0.113.50, 70.41.3.18", "203.0.113.50"},  // leading space
		{" 203.0.113.50 , 70.41.3.18", "203.0.113.50"},  // both
		{"  203.0.113.50  ", "203.0.113.50"},             // single IP with surrounding spaces
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", c.xff)
		got := clientIP(req)
		if got != c.want {
			t.Errorf("XFF %q: got %q, want %q", c.xff, got, c.want)
		}
	}
}

// TestRateLimiter_XFFSpoofingBypass documents that X-Forwarded-For is
// trusted without validation. A client behind no trusted proxy can set
// arbitrary XFF values and bypass the rate limit. This is expected behavior
// when deployed behind a trusted reverse proxy (which overwrites XFF), but
// is a security gap when the service is directly internet-facing.
func TestRateLimiter_XFFSpoofingBypass(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	handler := rl.Middleware(okHandler())

	// Exhaust the limit using one fake XFF IP.
	for range 3 {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Same RemoteAddr but different XFF value — bypasses the rate limit.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "5.6.7.8")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// This passes (200 OK) demonstrating the bypass is possible.
	if rr.Code != http.StatusOK {
		t.Fatalf("expected bypass to succeed, got %d", rr.Code)
	}
}
