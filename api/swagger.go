package api

import (
	_ "embed"
	"net/http"
)

// ─── Swagger Route ───────────────────────────────────────────────────────────

//go:embed public/swagger.yaml
var openapiYAML []byte

// handleSpec : GET /v1/swagger.yaml
func (h *Handler) handleSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openapiYAML)
}

// ─────────────────────────────────────────────────────────────────────────────
