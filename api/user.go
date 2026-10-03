package api

import (
	"demo/domain"
	"demo/util"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// ─── User Routes ─────────────────────────────────────────────────────────────

// handleListUserEvents : GET /v1/user/{id}/events
func (h *Handler) handleListUserEvents(w http.ResponseWriter, r *http.Request) {

	// ─── Parse Param And Header ──────────────────────────────────────────

	uidString := r.PathValue("id")
	uid, err := uuid.Parse(uidString)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	claims, err := h.tokens.Parse(r.Header.Get("Authorization"))
	if err != nil || claims == nil {
		util.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// ─── Check User Authentication ───────────────────────────────────────

	if claims.UserID != uid {
		util.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx := r.Context()

	user, err := h.userStore.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			util.RespondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if user.Role != domain.RoleOrganizer {
		util.RespondError(w, http.StatusForbidden, "forbidden: organizer role required")
		return
	}

	// ─── Fetch Corresponding Events ──────────────────────────────────────

	events, err := h.eventStore.ListEvents(ctx, domain.EventFilter{UserID: &uid})
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if events == nil {
		events = []*domain.Event{}
	}

	// ─── Respond With Status OK ──────────────────────────────────────────

	if err := util.EncodeJSON(w, http.StatusOK, map[string]any{"events": events}); err != nil {
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
}

// handleListUserTickets : GET /v1/user/{id}/tickets
func (h *Handler) handleListUserTickets(w http.ResponseWriter, r *http.Request) {

	// ─── Parse Param And Header ──────────────────────────────────────────

	uidString := r.PathValue("id")
	uid, err := uuid.Parse(uidString)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	claims, err := h.tokens.Parse(r.Header.Get("Authorization"))
	if err != nil || claims == nil {
		util.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// ─── Check User Authentication ───────────────────────────────────────

	if claims.UserID != uid {
		util.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx := r.Context()

	user, err := h.userStore.GetUserByID(ctx, uid)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			util.RespondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if user.Role != domain.RoleAttendee {
		util.RespondError(w, http.StatusForbidden, "forbidden: attendee role required")
		return
	}

	// ─── Fetch Corresponding Tickets ─────────────────────────────────────

	tickets, err := h.ticketStore.ListTickets(ctx, domain.TicketFilter{UserID: &uid, Status: domain.TicketStatusConfirmed})
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if tickets == nil {
		tickets = []*domain.Ticket{}
	}

	// ─── Respond With Status OK ──────────────────────────────────────────

	if err := util.EncodeJSON(w, http.StatusOK, map[string]any{"tickets": tickets}); err != nil {
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
}

// ─────────────────────────────────────────────────────────────────────────────
