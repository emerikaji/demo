package api

import (
	"demo/util"
	"net/http"
)

func (h *Handler) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	_ = util.EncodeJSON(w, http.StatusOK, util.Map{
		"status": "available",
	})
}
