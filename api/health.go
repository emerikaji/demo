package api

import "net/http"

func (h *Handler) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	_ = encodeJSON(w, http.StatusOK, Map{
		"status": "available",
	})
}
