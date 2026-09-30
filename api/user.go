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
	// 1. Parse and validate path parameter
	userIDStr := r.PathValue("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		util.RespondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	// 2. JWT Verification: Extract claims attached by JWTMiddleware
	claims, err := h.tokens.Parse(r.Header.Get("Authorization"))
	if err != nil || claims == nil {
		util.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 3. Identity Verification: Ensure authenticated user matches requested path ID
	if claims.UserID != userID {
		util.RespondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx := r.Context()

	// 4. Fetch user to verify existence and check role
	user, err := h.userStore.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			util.RespondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// 5. Role validation: only organizers can access user events route
	if user.Role != domain.RoleOrganizer {
		util.RespondError(w, http.StatusForbidden, "forbidden: organizer role required")
		return
	}

	// 6. Fetch all events (draft + published) created by this organizer
	events, err := h.eventStore.ListEvents(ctx, domain.EventFilter{UserID: &userID})
	if err != nil {
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Ensure empty slice serializes to `[]` instead of `null`
	if events == nil {
		events = []*domain.Event{}
	}

	// 7. Respond with 200 OK
	if err := util.EncodeJSON(w, http.StatusOK, map[string]any{"events": events}); err != nil {
		util.RespondError(w, http.StatusInternalServerError, "internal server error")
		return
	}
}

func (h *Handler) handleListUserTickets(w http.ResponseWriter, r *http.Request) {

}

// ─────────────────────────────────────────────────────────────────────────────
