package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestRequestID_GeneratesWhenAbsent(t *testing.T) {
	var ctxID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID = RequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestIDMiddleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rr.Code, http.StatusOK)
	}

	respID := rr.Header().Get("X-Request-ID")
	if respID == "" {
		t.Fatal("expected X-Request-ID response header to be set")
	}
	if !uuidPattern.MatchString(respID) {
		t.Fatalf("response X-Request-ID %q is not a valid UUID v4", respID)
	}
	if ctxID != respID {
		t.Fatalf("context ID %q does not match response header %q", ctxID, respID)
	}
}

func TestRequestID_PropagatesExisting(t *testing.T) {
	const incoming = "abc-123-existing-id"
	var ctxID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID = RequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestIDMiddleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", incoming)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("X-Request-ID") != incoming {
		t.Fatalf("response X-Request-ID %q, want %q", rr.Header().Get("X-Request-ID"), incoming)
	}
	if ctxID != incoming {
		t.Fatalf("context ID %q, want %q", ctxID, incoming)
	}
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	handler := RequestIDMiddleware(okHandler())

	ids := make(map[string]struct{})
	for range 100 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		id := rr.Header().Get("X-Request-ID")
		if _, exists := ids[id]; exists {
			t.Fatalf("duplicate request ID: %s", id)
		}
		ids[id] = struct{}{}
	}
}

func TestRequestID_EmptyContextWithoutMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	id := RequestID(req.Context())
	if id != "" {
		t.Fatalf("expected empty string without middleware, got %q", id)
	}
}

func TestGenerateID_Format(t *testing.T) {
	for range 50 {
		id := generateID()
		if !uuidPattern.MatchString(id) {
			t.Fatalf("generateID() = %q, does not match UUID v4 pattern", id)
		}
	}
}
