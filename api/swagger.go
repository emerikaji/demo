package api

import (
	_ "embed"
	"net/http"
)

// ─── Swagger Routes ──────────────────────────────────────────────────────────

//go:embed public/openapi.yaml
var openapiYAML []byte

// handleSpec : GET /v1/openapi.yaml
func (h *Handler) handleSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	_, _ = w.Write(openapiYAML)
}

//go:embed public/swagger.html
var swaggerHTML []byte

// handleSwagger : GET /v1/swagger
func (h *Handler) handleSwagger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(swaggerHTML)
}

// ─────────────────────────────────────────────────────────────────────────────
