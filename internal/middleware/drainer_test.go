package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/lucasnoah/yomauro/internal/middleware"
)

func TestDrainerMiddleware_PassesThrough(t *testing.T) {
	d := middleware.NewDrainer()
	h := d.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestDrainerMiddleware_RejectsDuringDrain(t *testing.T) {
	d := middleware.NewDrainer()
	h := d.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called during drain")
	}))

	// Drain with no in-flight requests completes immediately.
	if err := d.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if v := rec.Header().Get("Connection"); v != "close" {
		t.Fatalf("expected Connection: close, got %q", v)
	}
	if v := rec.Header().Get("Retry-After"); v != "5" {
		t.Fatalf("expected Retry-After: 5, got %q", v)
	}
}

func TestDrainWaitsForInflight(t *testing.T) {
	d := middleware.NewDrainer()
	release := make(chan struct{})
	started := make(chan struct{})

	h := d.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusOK)
	}))

	// Start an in-flight request.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}()

	// Wait until the handler is actually running before draining.
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start within 1s")
	}

	// Start draining.
	drained := make(chan error, 1)
	go func() {
		drained <- d.Drain(context.Background())
	}()

	// Drain should not complete while request is in-flight.
	select {
	case <-drained:
		t.Fatal("Drain returned while request is still in-flight")
	case <-time.After(50 * time.Millisecond):
	}

	// Release the in-flight request.
	close(release)
	wg.Wait()

	select {
	case err := <-drained:
		if err != nil {
			t.Fatalf("Drain returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Drain did not return after in-flight request completed")
	}
}

func TestDrainRespectsContextTimeout(t *testing.T) {
	d := middleware.NewDrainer()
	release := make(chan struct{})
	defer close(release)
	started := make(chan struct{})

	h := d.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
	}))

	// Start a request that blocks until release is closed.
	go func() {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}()

	// Wait until the handler is actually running before draining.
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start within 1s")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := d.Drain(ctx)
	if err == nil {
		t.Fatal("expected context deadline error, got nil")
	}
}

func TestDraining(t *testing.T) {
	d := middleware.NewDrainer()
	if d.Draining() {
		t.Fatal("expected Draining() to be false initially")
	}

	if err := d.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !d.Draining() {
		t.Fatal("expected Draining() to be true after Drain()")
	}
}
