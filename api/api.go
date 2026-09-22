// Package api is the implementation for each route in the API.
package api

import (
	"demo/domain"
	"demo/util"
	"net/http"
)

// Handler stores API dependencies and exposes the routes.
type Handler struct {
	userStore   domain.UserRepository
	eventStore  domain.EventRepository
	ticketStore domain.TicketRepository
	tokens      util.JWTProvider
}

// New is a Handler generator.
func New(userStore domain.UserRepository, eventStore domain.EventRepository, ticketStore domain.TicketRepository, tokens util.JWTProvider) *Handler {
	return &Handler{
		userStore:   userStore,
		eventStore:  eventStore,
		ticketStore: ticketStore,
		tokens:      tokens,
	}
}

// Routes provides the pattern-to-handler map for the API.
func (h *Handler) Routes() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		// Health
		"GET /v1/health": h.handleHealthCheck,

		// Auth
		"POST /v1/auth/register": h.handleRegisterUser,
		"POST /v1/auth/login":    h.handleLoginUser,

		/*
			// User
			"GET /v1/user/{id}/events":  h.handleListUserEvents,
			"GET /v1/user/{id}/tickets": h.handleListUserTickets,

			// Events
			"GET /v1/events":              h.handleListEvents,
			"GET /v1/events/{id}":         h.handleGetEvent,
			"POST /v1/events":             h.handleCreateEvent,
			"PATCH /v1/events/{id}":       h.handleUpdateEvent,
			"POST /v1/events/{id}/cancel": h.handleCancelEvent,

			// Tickets & Reservations
			"GET /v1/tickets/{id}":     h.handleGetTicket,
			"POST /v1/tickets/reserve": h.handleReserveTicket,
			"POST /v1/tickets/confirm": h.handleConfirmReservation,
			"POST /v1/tickets/cancel":  h.handleCancelReservation,

		*/
	}
}

// ─────────────────────────────────────────────────────────────────────────────
