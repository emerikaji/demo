package api

import (
	"demo/domain"
	"demo/util"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// ─── Ticket Routes ───────────────────────────────────────────────────────────

// handleGetTicket : GET /v1/tickets/{id}
func (h *Handler) handleGetTicket(w http.ResponseWriter, r *http.Request) {

	// ─── Parse Param ─────────────────────────────────────────────────────

	ticketidString := r.PathValue("id")

	ticketid, err := uuid.Parse(ticketidString)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, "invalid ticket id")
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

	if user.Role != domain.RoleAttendee {
		util.RespondError(w, http.StatusForbidden, "attendee role required")
		return
	}

	// ─── Fetch And Verify Ticket ─────────────────────────────────────────

	ticket, err := h.ticketStore.GetTicket(r.Context(), ticketid)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			util.RespondError(w, http.StatusNotFound, "ticket not found")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if ticket.UserID != user.ID {
		util.RespondError(w, http.StatusForbidden, "you do not own this ticket")
		return
	}

	// ─── Respond With Status Ok ──────────────────────────────────────────

	if err := util.EncodeJSON(w, http.StatusOK, map[string]any{"ticket": ticket}); err != nil {
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
}

// handleReserveTicket : POST /v1/tickets
func (h *Handler) handleReserveTicket(w http.ResponseWriter, r *http.Request) {

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

	if user.Role != domain.RoleAttendee {
		util.RespondError(w, http.StatusForbidden, "attendee role required")
		return
	}

	// ─── Parse And Validate JSON ─────────────────────────────────────────

	ticket := &domain.Ticket{
		ID:     uuid.New(),
		UserID: user.ID,
	}

	if err := util.DecodeJSON(w, r, ticket); err != nil {
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

	if ticket.EventID == uuid.Nil {
		util.RespondError(w, http.StatusBadRequest, "event id is required")
		return
	}

	// ─── Reserve Ticket ──────────────────────────────────────────────────

	if err := h.ticketStore.ReserveTicket(r.Context(), ticket); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			util.RespondError(w, http.StatusNotFound, "event not found")
			return
		}
		if errors.Is(err, domain.ErrSoldOut) {
			util.RespondError(w, http.StatusConflict, "event is sold out")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// ─── Respond With Status Created ─────────────────────────────────────

	_ = util.EncodeJSON(w, http.StatusCreated, util.Map{
		"message": "reservation created successfully",
	})
}

// handleConfirmReservation : POST /v1/tickets/{id}
func (h *Handler) handleConfirmReservation(w http.ResponseWriter, r *http.Request) {

	// ─── Parse Param ─────────────────────────────────────────────────────

	ticketidString := r.PathValue("id")

	ticketid, err := uuid.Parse(ticketidString)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, "invalid ticket id")
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

	if user.Role != domain.RoleAttendee {
		util.RespondError(w, http.StatusForbidden, "attendee role required")
		return
	}

	// ─── Fetch And Verify Ticket ─────────────────────────────────────────

	ticket, err := h.ticketStore.GetTicket(r.Context(), ticketid)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			util.RespondError(w, http.StatusNotFound, "ticket not found")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if ticket.UserID != user.ID {
		util.RespondError(w, http.StatusForbidden, "you do not own this ticket")
		return
	}

	// ─── Confirm Reservation ─────────────────────────────────────────────

	if err := h.ticketStore.ConfirmReservation(r.Context(), ticket); err != nil {
		if errors.Is(err, domain.ErrTicketExpired) {
			util.RespondError(w, http.StatusBadRequest, "ticket reservation has expired")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// ─── Respond With 200 OK ─────────────────────────────────────────────

	_ = util.EncodeJSON(w, http.StatusOK, util.Map{
		"message": "reservation confirmed successfully",
	})
}

// handleCancelReservation : DELETE /v1/tickets/{id}
func (h *Handler) handleCancelReservation(w http.ResponseWriter, r *http.Request) {

	// ─── Parse Param ─────────────────────────────────────────────────────

	ticketidString := r.PathValue("id")

	ticketid, err := uuid.Parse(ticketidString)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, "invalid ticket id")
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

	if user.Role != domain.RoleAttendee {
		util.RespondError(w, http.StatusForbidden, "attendee role required")
		return
	}

	// ─── Fetch And Verify Ticket ─────────────────────────────────────────

	ticket, err := h.ticketStore.GetTicket(r.Context(), ticketid)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			util.RespondError(w, http.StatusNotFound, "ticket not found")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if ticket.UserID != user.ID {
		util.RespondError(w, http.StatusForbidden, "you do not own this ticket")
		return
	}

	// ─── Cancel Reservation Or Delete Ticket ─────────────────────────────

	response := util.Map{}

	if ticket.Status == domain.TicketStatusReserved {
		response["message"] = "reservation cancelled successfully"
	}

	if ticket.Status == domain.TicketStatusConfirmed {
		response["message"] = "ticket deleted successfully"
	}

	if err := h.ticketStore.CancelReservation(r.Context(), ticket); err != nil {
		if errors.Is(err, domain.ErrTicketExpired) {
			util.RespondError(w, http.StatusBadRequest, "ticket reservation has expired")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	_ = util.EncodeJSON(w, http.StatusOK, response)
}

// ─────────────────────────────────────────────────────────────────────────────
