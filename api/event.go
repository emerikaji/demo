package api

import (
	"demo/domain"
	"demo/util"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

// ─── Event Routes ────────────────────────────────────────────────────────────

// handleListEvents : GET /v1/events
func (h *Handler) handleListEvents(w http.ResponseWriter, r *http.Request) {

	// ─── Parse Params ────────────────────────────────────────────────────

	limitString := r.URL.Query().Get("limit")
	offsetString := r.URL.Query().Get("offset")
	uidString := r.URL.Query().Get("user_id")

	log.Println("limitString:", limitString, "offsetString:", offsetString, "uidString:", uidString)

	filter := domain.EventFilter{Status: domain.EventStatusPublished, Limit: 20, Offset: 0}
	var err error

	if limitString != "" {
		filter.Limit, err = strconv.Atoi(limitString)
		if err != nil {
			util.RespondError(w, http.StatusBadRequest, "invalid limit value")
			return
		}
	}

	if offsetString != "" {
		filter.Offset, err = strconv.Atoi(offsetString)
		if err != nil {
			util.RespondError(w, http.StatusBadRequest, "invalid offset value")
			return
		}
	}

	if uidString != "" {
		var uid uuid.UUID
		uid, err = uuid.Parse(uidString)
		if err != nil {
			util.RespondError(w, http.StatusBadRequest, "invalid user id")
			return
		}
		filter.UserID = &uid
	}

	// ─── Fetch Corresponding Events ──────────────────────────────────────

	events, err := h.eventStore.ListEvents(r.Context(), filter)
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if events == nil {
		events = []*domain.Event{}
	}

	// ─── Respond With Status Ok ──────────────────────────────────────────

	if err := util.EncodeJSON(w, http.StatusOK, map[string]any{"events": events}); err != nil {
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
}

// handleGetEvent : GET /v1/events/{ID}
func (h *Handler) handleGetEvent(w http.ResponseWriter, r *http.Request) {

}

// handleCreateEvent : POST /v1/events
func (h *Handler) handleCreateEvent(w http.ResponseWriter, r *http.Request) {

}

// handleUpdateEvent : PATCH /v1/events/{ID}
func (h *Handler) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {

}

// handleCancelEvent : DELETE /v1/events/{ID}
func (h *Handler) handleCancelEvent(w http.ResponseWriter, r *http.Request) {

}

// ─────────────────────────────────────────────────────────────────────────────
