// Package api is the implementation for each route in the API.
package api

import (
	"net/http"
)

// Handler stores API dependencies and exposes the routes.
type Handler struct{}

// New is a Handler generator.
func New() *Handler {
	return &Handler{}
}

// Routes provides the pattern-to-handler map for the API.
func (h *Handler) Routes() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"GET /demo": h.Demo,
	}
}

// ─── GET Routes ──────────────────────────────────────────────────────────────

func (h *Handler) Demo(w http.ResponseWriter, r *http.Request) {

}

// ─── POST Routes ─────────────────────────────────────────────────────────────

// ─────────────────────────────────────────────────────────────────────────────
