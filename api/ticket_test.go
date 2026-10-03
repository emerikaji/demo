package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"demo/domain"
	"demo/store"

	"github.com/google/uuid"
)

// ─── Test Battery ────────────────────────────────────────────────────────────

var ticketTests = []apiTest{

	// ─── Get Ticket ──────────────────────────────────────────────────────

	{
		name:           "get ticket - missing auth header should return 401",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketID1),
		method:         http.MethodGet,
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"unauthorized"}`,
	},
	{
		name:           "get ticket - invalid UUID path param with valid JWT should return 400",
		route:          "/v1/tickets/invalid-uuid",
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"invalid ticket id"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDJane
		},
	},
	{
		name:           "get ticket - non-existent ticket should return 404",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketID1),
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusNotFound,
		expectedBody:   `{"error":"ticket not found"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDJane
			err := s.CreateUser(context.Background(), &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}
		},
	},
	{
		name:           "get ticket - JWT subject mismatch should return 403",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketID1),
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token-for-alexandrina"},
		expectedStatus: http.StatusForbidden,
		expectedBody:   `{"error":"forbidden: ticket does not belong to user"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDAlexandrina

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			err = s.CreateUser(ctx, &domain.User{
				ID:    userIDAlexandrina,
				Email: "alexandrina@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed second user: %v", err)
			}

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      uuid.New(),
				Title:            "GolangConf 2026",
				Capacity:         100,
				RemainingTickets: 100,
				Status:           domain.EventStatusPublished,
			})
			if err != nil {
				t.Fatalf("failed to seed event: %v", err)
			}

			err = s.ReserveTicket(ctx, &domain.Ticket{
				ID:        ticketID1,
				EventID:   eventID1,
				UserID:    userIDJane,
				Status:    domain.TicketStatusConfirmed,
				CreatedAt: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed ticket: %v", err)
			}
		},
	},
	{
		name:           "get ticket - non-attendee role should return 403",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketID1),
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusForbidden,
		expectedBody:   `{"error":"forbidden: attendee role required"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDJane

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "organizer@example.com",
				Role:  domain.RoleOrganizer,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}
		},
	},
	{
		name:           "get ticket - successful retrieval of confirmed ticket returns 200 with ticket entity",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketID1),
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusOK,
		expectedBody:   fmt.Sprintf(`{"id":"%s","event_id":"%s","user_id":"%s","status":"confirmed","created_at":"2026-10-01T12:00:00Z"}`, ticketID1, eventID1, userIDJane),
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDJane

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      uuid.New(),
				Title:            "GolangConf 2026",
				Capacity:         100,
				RemainingTickets: 100,
				Status:           domain.EventStatusPublished,
			})
			if err != nil {
				t.Fatalf("failed to seed event: %v", err)
			}

			err = s.ReserveTicket(ctx, &domain.Ticket{
				ID:        ticketID1,
				EventID:   eventID1,
				UserID:    userIDJane,
				Status:    domain.TicketStatusConfirmed,
				CreatedAt: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed ticket: %v", err)
			}
		},
	},
	{
		name:           "get ticket - successful retrieval of reserved ticket returns 200 with ticket entity",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketIDReserved),
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusOK,
		expectedBody:   fmt.Sprintf(`{"id":"%s","event_id":"%s","user_id":"%s","status":"reserved","created_at":"2026-10-01T12:00:00Z"}`, ticketIDReserved, eventID1, userIDJane),
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDJane

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      uuid.New(),
				Title:            "GolangConf 2026",
				Capacity:         100,
				RemainingTickets: 100,
				Status:           domain.EventStatusPublished,
			})
			if err != nil {
				t.Fatalf("failed to seed event: %v", err)
			}

			err = s.ReserveTicket(ctx, &domain.Ticket{
				ID:        ticketIDReserved,
				EventID:   eventID1,
				UserID:    userIDJane,
				Status:    domain.TicketStatusReserved,
				ExpiresAt: time.Now().Add(15 * time.Minute),
				CreatedAt: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed ticket: %v", err)
			}
		},
	},

	// ─── Reserve Ticket ──────────────────────────────────────────────────

	{
		name:           "reserve ticket - missing auth header should return 401",
		route:          "/v1/tickets",
		method:         http.MethodPost,
		body:           []byte(fmt.Sprintf(`{"event_id":"%s"}`, eventID1)),
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"unauthorized"}`,
	},
	{
		name:           "reserve ticket - empty body should return 400",
		route:          "/v1/tickets",
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(``),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"body must not be empty"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDJane
		},
	},
	{
		name:           "reserve ticket - missing event_id should return 400",
		route:          "/v1/tickets",
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(`{}`),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"event_id is required"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDJane
		},
	},
	{
		name:           "reserve ticket - non-attendee role should return 403",
		route:          "/v1/tickets",
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(fmt.Sprintf(`{"event_id":"%s"}`, eventID1)),
		expectedStatus: http.StatusForbidden,
		expectedBody:   `{"error":"forbidden: attendee role required"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDJane

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "organizer@example.com",
				Role:  domain.RoleOrganizer,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}
		},
	},
	{
		name:           "reserve ticket - non-existent event should return 404",
		route:          "/v1/tickets",
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(fmt.Sprintf(`{"event_id":"%s"}`, eventID1)),
		expectedStatus: http.StatusNotFound,
		expectedBody:   `{"error":"event not found"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDJane

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}
		},
	},
	{
		name:           "reserve ticket - sold out event should return 409",
		route:          "/v1/tickets",
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(fmt.Sprintf(`{"event_id":"%s"}`, eventIDSoldOut)),
		expectedStatus: http.StatusConflict,
		expectedBody:   `{"error":"event is sold out"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDJane

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventIDSoldOut,
				OrganizerID:      uuid.New(),
				Title:            "Sold Out Concert",
				Capacity:         10,
				RemainingTickets: 0,
				Status:           domain.EventStatusPublished,
			})
			if err != nil {
				t.Fatalf("failed to seed sold out event: %v", err)
			}
		},
	},
	{
		name:           "reserve ticket - successful reservation should return 201 with reserved ticket",
		route:          "/v1/tickets",
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(fmt.Sprintf(`{"id":"%s","event_id":"%s","created_at":"2026-10-01T12:00:00Z"}`, ticketIDReserved, eventID1)),
		expectedStatus: http.StatusCreated,
		expectedBody:   fmt.Sprintf(`{"id":"%s","event_id":"%s","user_id":"%s","status":"reserved","created_at":"2026-10-01T12:00:00Z"}`, ticketIDReserved, eventID1, userIDJane),
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDJane

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      uuid.New(),
				Title:            "GolangConf 2026",
				Capacity:         100,
				RemainingTickets: 100,
				Status:           domain.EventStatusPublished,
			})
			if err != nil {
				t.Fatalf("failed to seed event: %v", err)
			}
		},
	},

	// ─── Confirm Reservation ─────────────────────────────────────────────

	{
		name:           "confirm reservation - missing auth header should return 401",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketIDReserved),
		method:         http.MethodPost,
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"unauthorized"}`,
	},
	{
		name:           "confirm reservation - invalid UUID path param with valid JWT should return 400",
		route:          "/v1/tickets/invalid-uuid",
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"invalid ticket id"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDJane
		},
	},
	{
		name:           "confirm reservation - non-existent ticket should return 404",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketIDReserved),
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusNotFound,
		expectedBody:   `{"error":"ticket not found"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDJane
			err := s.CreateUser(context.Background(), &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}
		},
	},
	{
		name:           "confirm reservation - JWT subject mismatch should return 403",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketIDReserved),
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token-for-alexandrina"},
		expectedStatus: http.StatusForbidden,
		expectedBody:   `{"error":"forbidden: ticket does not belong to user"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDAlexandrina

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      uuid.New(),
				Title:            "GolangConf 2026",
				Capacity:         100,
				RemainingTickets: 100,
				Status:           domain.EventStatusPublished,
			})
			if err != nil {
				t.Fatalf("failed to seed event: %v", err)
			}

			err = s.ReserveTicket(ctx, &domain.Ticket{
				ID:        ticketIDReserved,
				EventID:   eventID1,
				UserID:    userIDJane,
				Status:    domain.TicketStatusReserved,
				ExpiresAt: time.Now().Add(15 * time.Minute),
				CreatedAt: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed ticket: %v", err)
			}
		},
	},
	{
		name:           "confirm reservation - expired reservation should return 400",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketIDExpired),
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"ticket reservation has expired"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDJane

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      uuid.New(),
				Title:            "GolangConf 2026",
				Capacity:         100,
				RemainingTickets: 100,
				Status:           domain.EventStatusPublished,
			})
			if err != nil {
				t.Fatalf("failed to seed event: %v", err)
			}

			err = s.ReserveTicket(ctx, &domain.Ticket{
				ID:        ticketIDExpired,
				EventID:   eventID1,
				UserID:    userIDJane,
				Status:    domain.TicketStatusReserved,
				ExpiresAt: time.Now().Add(-5 * time.Minute),
				CreatedAt: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed expired ticket: %v", err)
			}
		},
	},
	{
		name:           "confirm reservation - successful confirmation should return 200 with confirmed ticket",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketIDReserved),
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusOK,
		expectedBody:   fmt.Sprintf(`{"id":"%s","event_id":"%s","user_id":"%s","status":"confirmed","created_at":"2026-10-01T12:00:00Z"}`, ticketIDReserved, eventID1, userIDJane),
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDJane

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      uuid.New(),
				Title:            "GolangConf 2026",
				Capacity:         100,
				RemainingTickets: 100,
				Status:           domain.EventStatusPublished,
			})
			if err != nil {
				t.Fatalf("failed to seed event: %v", err)
			}

			err = s.ReserveTicket(ctx, &domain.Ticket{
				ID:        ticketIDReserved,
				EventID:   eventID1,
				UserID:    userIDJane,
				Status:    domain.TicketStatusReserved,
				ExpiresAt: time.Now().Add(15 * time.Minute),
				CreatedAt: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed ticket: %v", err)
			}
		},
	},

	// ─── Cancel Reservation ──────────────────────────────────────────────

	{
		name:           "cancel reservation - missing auth header should return 401",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketIDReserved),
		method:         http.MethodDelete,
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"unauthorized"}`,
	},
	{
		name:           "cancel reservation - invalid UUID path param with valid JWT should return 400",
		route:          "/v1/tickets/invalid-uuid",
		method:         http.MethodDelete,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"invalid ticket id"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDJane
		},
	},
	{
		name:           "cancel reservation - non-existent ticket should return 404",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketIDReserved),
		method:         http.MethodDelete,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusNotFound,
		expectedBody:   `{"error":"ticket not found"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDJane
			err := s.CreateUser(context.Background(), &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}
		},
	},
	{
		name:           "cancel reservation - JWT subject mismatch should return 403",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketIDReserved),
		method:         http.MethodDelete,
		headers:        map[string]string{"Authorization": "Bearer valid-token-for-alexandrina"},
		expectedStatus: http.StatusForbidden,
		expectedBody:   `{"error":"forbidden: ticket does not belong to user"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDAlexandrina

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      uuid.New(),
				Title:            "GolangConf 2026",
				Capacity:         100,
				RemainingTickets: 100,
				Status:           domain.EventStatusPublished,
			})
			if err != nil {
				t.Fatalf("failed to seed event: %v", err)
			}

			err = s.ReserveTicket(ctx, &domain.Ticket{
				ID:        ticketIDReserved,
				EventID:   eventID1,
				UserID:    userIDJane,
				Status:    domain.TicketStatusReserved,
				ExpiresAt: time.Now().Add(15 * time.Minute),
				CreatedAt: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed ticket: %v", err)
			}
		},
	},
	{
		name:           "cancel reservation - successful cancellation should return 200 with cancelled ticket",
		route:          fmt.Sprintf("/v1/tickets/%s", ticketIDReserved),
		method:         http.MethodDelete,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusOK,
		expectedBody:   fmt.Sprintf(`{"id":"%s","event_id":"%s","user_id":"%s","status":"cancelled","created_at":"2026-10-01T12:00:00Z"}`, ticketIDReserved, eventID1, userIDJane),
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDJane

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      uuid.New(),
				Title:            "GolangConf 2026",
				Capacity:         100,
				RemainingTickets: 100,
				Status:           domain.EventStatusPublished,
			})
			if err != nil {
				t.Fatalf("failed to seed event: %v", err)
			}

			err = s.ReserveTicket(ctx, &domain.Ticket{
				ID:        ticketIDReserved,
				EventID:   eventID1,
				UserID:    userIDJane,
				Status:    domain.TicketStatusReserved,
				ExpiresAt: time.Now().Add(15 * time.Minute),
				CreatedAt: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed ticket: %v", err)
			}
		},
	},
}

// ─── Run Test ────────────────────────────────────────────────────────────────

func TestTickets(t *testing.T) {
	testRoutesInMemory(t, ticketTests)
	testRoutesHTTP(t, ticketTests)
}

// ─────────────────────────────────────────────────────────────────────────────
