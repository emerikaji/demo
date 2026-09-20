package api

import (
	"net/http"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Routes() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"GET /demo": h.Demo,
	}
}

func (h *Handler) Demo(w http.ResponseWriter, r *http.Request) {
}
