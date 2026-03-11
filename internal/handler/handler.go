// Package handler provides HTTP handlers and routing for the API.
package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lucasnoah/yomauro/internal/middleware"
)

// Pinger can check database connectivity.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	db Pinger
}

// NewRouter creates a chi router with all routes registered.
func NewRouter(db Pinger) http.Handler {
	h := &Handler{db: db}

	r := chi.NewRouter()
	r.Use(middleware.RequestIDMiddleware)
	r.Get("/api/v1/health", h.Health)

	return r
}
