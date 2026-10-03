package api

import (
	"demo/domain"
	"demo/util"
	"errors"
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

// handleGetEvent : GET /v1/events/{id}
func (h *Handler) handleGetEvent(w http.ResponseWriter, r *http.Request) {

	// ─── Parse Param ─────────────────────────────────────────────────────

	eventidString := r.PathValue("id")

	eventid, err := uuid.Parse(eventidString)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	// ─── Fetch Event ─────────────────────────────────────────────────────

	event, err := h.eventStore.GetEvent(r.Context(), eventid)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			util.RespondError(w, http.StatusNotFound, "event not found")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// ─── Respond With Status Ok ──────────────────────────────────────────

	if err := util.EncodeJSON(w, http.StatusOK, map[string]any{"event": event}); err != nil {
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
}

// handleCreateEvent : POST /v1/events
func (h *Handler) handleCreateEvent(w http.ResponseWriter, r *http.Request) {

	// ─── Parse And Verify Auth ───────────────────────────────────────────

	claims, err := h.tokens.Parse(r.Header.Get("Authorization"))
	if err != nil || claims == nil {
		util.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userStore.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			util.RespondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if user.Role != domain.RoleOrganizer {
		util.RespondError(w, http.StatusForbidden, "organizer role required")
		return
	}

	// ─── Parse And Validate JSON ─────────────────────────────────────────

	event := &domain.Event{
		ID:          uuid.New(),
		OrganizerID: user.ID,
	}

	if err := util.DecodeJSON(w, r, event); err != nil {
		if errors.Is(err, util.ErrEmptyBody) {
			util.RespondError(w, http.StatusBadRequest, "body must not be empty")
			return
		}
		if errors.Is(err, util.ErrJSONValues) {
			util.RespondError(w, http.StatusBadRequest, "body cannot contain more than one json value")
			return
		}
		if errors.Is(err, util.ErrInvalidJSON) {
			util.RespondError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if errors.Is(err, util.ErrInvalidJSONType) {
			util.RespondError(w, http.StatusBadRequest, "invalid json types")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if event.Title == "" || event.StartsAt.IsZero() || event.Capacity == 0 {
		util.RespondError(w, http.StatusBadRequest, "title, capacity, and starts_at are required")
		return
	}

	if event.Capacity < 1 {
		util.RespondError(w, http.StatusBadRequest, "capacity must be greater than 0")
		return
	}

	// ─── Create Event ────────────────────────────────────────────────────

	if err := h.eventStore.CreateEvent(r.Context(), event); err != nil {
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// ─── Respond With Status Created ─────────────────────────────────────

	_ = util.EncodeJSON(w, http.StatusCreated, util.Map{
		"message": "event created successfully",
	})
}

// handleUpdateEvent : PATCH /v1/events/{id}
func (h *Handler) handleUpdateEvent(w http.ResponseWriter, r *http.Request) {

	// ─── Parse Param ─────────────────────────────────────────────────────

	eventidString := r.PathValue("id")

	eventid, err := uuid.Parse(eventidString)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	// ─── Parse And Verify Auth ───────────────────────────────────────────

	claims, err := h.tokens.Parse(r.Header.Get("Authorization"))
	if err != nil || claims == nil {
		util.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userStore.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			util.RespondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if user.Role != domain.RoleOrganizer {
		util.RespondError(w, http.StatusForbidden, "organizer role required")
		return
	}

	// ─── Fetch And Verify Event ──────────────────────────────────────────

	event, err := h.eventStore.GetEvent(r.Context(), eventid)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			util.RespondError(w, http.StatusNotFound, "event not found")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if event.OrganizerID != user.ID {
		util.RespondError(w, http.StatusForbidden, "you do not own this event")
		return
	}

	// ─── Parse And Validate JSON ─────────────────────────────────────────

	if err := util.DecodeJSON(w, r, event); err != nil {
		if errors.Is(err, util.ErrEmptyBody) {
			util.RespondError(w, http.StatusBadRequest, "body must not be empty")
			return
		}
		if errors.Is(err, util.ErrJSONValues) {
			util.RespondError(w, http.StatusBadRequest, "body cannot contain more than one json value")
			return
		}
		if errors.Is(err, util.ErrInvalidJSON) {
			util.RespondError(w, http.StatusBadRequest, "invalid json")
			return
		}
		if errors.Is(err, util.ErrInvalidJSONType) {
			util.RespondError(w, http.StatusBadRequest, "invalid json types")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// ─── Update Event ────────────────────────────────────────────────────

	if err := h.eventStore.UpdateEvent(r.Context(), event); err != nil {
		if errors.Is(err, domain.ErrEventCancelled) {
			util.RespondError(w, http.StatusBadRequest, "cannot update a cancelled event")
			return
		}
		if errors.Is(err, domain.ErrInvalidCapacity) {
			util.RespondError(w, http.StatusBadRequest, "new capacity cannot be less than tickets already issued")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// ─── Respond With Status OK ──────────────────────────────────────────

	_ = util.EncodeJSON(w, http.StatusOK, util.Map{
		"message": "event updated successfully",
	})
}

// handleCancelEvent : DELETE /v1/events/{id}
func (h *Handler) handleCancelEvent(w http.ResponseWriter, r *http.Request) {

	// ─── Parse Param ─────────────────────────────────────────────────────

	eventidString := r.PathValue("id")

	eventid, err := uuid.Parse(eventidString)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	// ─── Parse And Verify Auth ───────────────────────────────────────────

	claims, err := h.tokens.Parse(r.Header.Get("Authorization"))
	if err != nil || claims == nil {
		util.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userStore.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			util.RespondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if user.Role != domain.RoleOrganizer {
		util.RespondError(w, http.StatusForbidden, "organizer role required")
		return
	}

	// ─── Fetch And Verify Event ──────────────────────────────────────────

	event, err := h.eventStore.GetEvent(r.Context(), eventid)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			util.RespondError(w, http.StatusNotFound, "event not found")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if event.OrganizerID != user.ID {
		util.RespondError(w, http.StatusForbidden, "you do not own this event")
		return
	}

	switch event.Status {
	case domain.EventStatusCancelled:
		util.RespondError(w, http.StatusBadRequest, "cannot cancel an already cancelled event")
		return
	case domain.EventStatusPublished:
		event.Status = domain.EventStatusCancelled

		if err := h.eventStore.UpdateEvent(r.Context(), event); err != nil {
			util.RespondError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		_ = util.EncodeJSON(w, http.StatusOK, util.Map{
			"message": "event cancelled successfully",
		})
	case domain.EventStatusDraft:
		if err := h.eventStore.DeleteEventDraft(r.Context(), eventid); err != nil {
			util.RespondError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		_ = util.EncodeJSON(w, http.StatusOK, util.Map{
			"message": "event deleted successfully",
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
