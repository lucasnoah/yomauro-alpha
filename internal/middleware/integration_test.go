package middleware_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lucasnoah/yomauro/internal/middleware"
)

// TestDrainerIntegration tests graceful shutdown with a real HTTP server
// and an actual in-flight request, simulating what happens during SIGTERM.
func TestDrainerIntegration(t *testing.T) {
	d := middleware.NewDrainer()
	slowRelease := make(chan struct{})
	slowStarted := make(chan struct{})

	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		close(slowStarted)
		<-slowRelease
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("done"))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{Handler: d.Middleware(mux)}

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("could not listen: %v", err)
	}
	addr := ln.Addr().String()
	go srv.Serve(ln)
	t.Cleanup(func() {
		srv.Shutdown(context.Background())
	})

	// Verify normal request passes through.
	resp, err := http.Get("http://" + addr + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("health check failed: err=%v code=%v", err, resp)
	}

	// Start an in-flight slow request.
	slowDone := make(chan *http.Response, 1)
	go func() {
		r, err := http.Get("http://" + addr + "/slow")
		if err != nil {
			slowDone <- nil
			return
		}
		slowDone <- r
	}()

	select {
	case <-slowStarted:
	case <-time.After(time.Second):
		t.Fatal("slow handler did not start within 1s")
	}

	// Trigger drain — simulates SIGTERM handler calling drainer.Drain().
	drainCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	drainDone := make(chan error, 1)
	go func() {
		drainDone <- d.Drain(drainCtx)
	}()

	// Allow the drain goroutine to set the flag.
	time.Sleep(10 * time.Millisecond)

	// New requests during drain must get 503 + Retry-After.
	// Note: Go's HTTP client strips hop-by-hop headers like Connection from
	// responses; Connection:close is verified in unit tests via ResponseRecorder.
	client := &http.Client{Timeout: time.Second}
	resp503, err503 := client.Get("http://" + addr + "/health")
	if err503 != nil {
		t.Fatalf("expected 503 response during drain, got error: %v", err503)
	}
	if resp503.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("during drain: expected 503, got %d", resp503.StatusCode)
	}
	if v := resp503.Header.Get("Retry-After"); v != "5" {
		t.Errorf("during drain: expected Retry-After:5, got %q", v)
	}

	// Drain must not complete while in-flight request is outstanding.
	select {
	case <-drainDone:
		t.Fatal("drain returned while in-flight request is still running")
	case <-time.After(50 * time.Millisecond):
	}

	// Release the slow request.
	close(slowRelease)

	select {
	case r := <-slowDone:
		if r == nil || r.StatusCode != http.StatusOK {
			t.Errorf("in-flight request did not complete with 200: %v", r)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("in-flight request did not complete within 2s after release")
	}

	// Drain must complete now.
	select {
	case err := <-drainDone:
		if err != nil {
			t.Errorf("drain returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("drain did not complete within 2s after in-flight request finished")
	}
}

// TestDrainerIntegration_NoInflight tests that drain completes immediately
// when there are no in-flight requests, and the server then stops accepting.
func TestDrainerIntegration_NoInflight(t *testing.T) {
	d := middleware.NewDrainer()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{Handler: d.Middleware(mux)}
	ts := httptest.NewServer(d.Middleware(mux))
	t.Cleanup(ts.Close)

	// Normal request works.
	resp, err := http.Get(ts.URL + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("pre-drain health check failed: %v", err)
	}

	// Drain with no in-flight requests.
	if err := d.Drain(context.Background()); err != nil {
		t.Fatalf("drain failed: %v", err)
	}

	// Subsequent requests are rejected.
	resp2, err2 := http.Get(ts.URL + "/health")
	if err2 != nil {
		t.Fatalf("expected response after drain, got error: %v", err2)
	}
	if resp2.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("after drain: expected 503, got %d", resp2.StatusCode)
	}

	_ = srv.Shutdown(context.Background())
}
