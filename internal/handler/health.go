package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Health checks database connectivity and reports service status.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := h.db.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  "database ping failed",
		}); err != nil {
			slog.Error("health: failed to write response", "error", err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	}); err != nil {
		slog.Error("health: failed to write response", "error", err)
	}
}
