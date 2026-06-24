// Package health exposes standard liveness and readiness HTTP handlers.
package health

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// Checker reports whether a dependency is ready.
type Checker func(r *http.Request) error

// Handler serves GET /healthz (always OK) and GET /readyz (runs checkers).
type Handler struct {
	ready atomic.Bool
	check Checker
}

// New returns a Handler with an optional readiness checker.
func New(check Checker) *Handler {
	h := &Handler{check: check}
	h.ready.Store(true)
	return h
}

// SetReady toggles readiness (useful in tests or graceful startup).
func (h *Handler) SetReady(v bool) {
	h.ready.Store(v)
}

// Register mounts /healthz and /readyz on mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.Live)
	mux.HandleFunc("GET /readyz", h.Ready)
}

// Live responds 200 when the process is running.
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready responds 200 only when ready and all checkers pass.
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	if !h.ready.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
		return
	}
	if h.check != nil {
		if err := h.check(r); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "not_ready",
				"error":  err.Error(),
			})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
