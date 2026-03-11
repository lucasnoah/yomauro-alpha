package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucasnoah/yomauro/internal/handler"
)

type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(_ context.Context) error {
	return m.err
}

func TestHealth_OK(t *testing.T) {
	router := handler.NewRouter(&mockPinger{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status \"ok\", got %q", body["status"])
	}
}

func TestHealth_Unhealthy(t *testing.T) {
	router := handler.NewRouter(&mockPinger{err: errors.New("connection refused")})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "unhealthy" {
		t.Fatalf("expected status \"unhealthy\", got %q", body["status"])
	}
	if body["error"] != "database ping failed" {
		t.Fatalf("expected error \"database ping failed\", got %q", body["error"])
	}
}
