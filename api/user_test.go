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

// Static User UUIDs for deterministic path parameters
var (
	userIDJane        = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userIDAlexandrina = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	eventID1          = uuid.MustParse("00000000-0000-0000-0000-000000000010")
	ticketID1         = uuid.MustParse("00000000-0000-0000-0000-000000000100")
	ticketIDReserved  = uuid.MustParse("00000000-0000-0000-0000-000000000200")
)

// ─── Test Battery ────────────────────────────────────────────────────────────

var userTests = []apiTest{

	// ─── Events ──────────────────────────────────────────────────────────

	{
		name:           "list user events - invalid UUID path param should return 400",
		route:          "/v1/user/invalid-uuid/events",
		method:         http.MethodGet,
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"invalid user id"}`,
	},
	{
		name:           "list user events - non-existent user should return 404",
		route:          fmt.Sprintf("/v1/user/%s/events", userIDAlexandrina),
		method:         http.MethodGet,
		expectedStatus: http.StatusNotFound,
		expectedBody:   `{"error":"user not found"}`,
	},
	{
		name:           "list user events - user with no organized events should return empty list",
		route:          fmt.Sprintf("/v1/user/%s/events", userIDJane),
		method:         http.MethodGet,
		expectedStatus: http.StatusOK,
		expectedBody:   `{"events":[]}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			err := s.CreateUser(context.Background(), &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleOrganizer,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}
		},
	},
	{
		name:           "list user events - successful list should return 200 with organized events",
		route:          fmt.Sprintf("/v1/user/%s/events", userIDJane),
		method:         http.MethodGet,
		expectedStatus: http.StatusOK,
		expectedBody:   fmt.Sprintf(`{"events":[{"id":"%s","organizer_id":"%s","title":"GolangConf 2026","description":"Go-conference with a program built on real-world tasks","capacity":100,"remaining_tickets":100,"status":"published","starts_at":"2027-04-20T10:00:00Z","created_at":"2026-09-01T00:00:00Z"}]}`, eventID1, userIDJane),
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleOrganizer,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      userIDJane,
				Title:            "GolangConf 2026",
				Description:      "Go-conference with a program built on real-world tasks",
				Capacity:         100,
				RemainingTickets: 100,
				Status:           domain.EventStatusPublished,
				StartsAt:         time.Date(2027, 4, 20, 10, 0, 0, 0, time.UTC),
				CreatedAt:        time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed event: %v", err)
			}
		},
	},

	// ─── Tickets ─────────────────────────────────────────────────────────

	{
		name:           "list user tickets - missing auth header should return 401",
		route:          fmt.Sprintf("/v1/user/%s/tickets", userIDAlexandrina),
		method:         http.MethodGet,
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"unauthorized"}`,
	},
	{
		name:           "list user tickets - JWT subject mismatch should return 401",
		route:          fmt.Sprintf("/v1/user/%s/tickets", userIDJane),
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token-for-alexandrina"},
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"unauthorized"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDAlexandrina
		},
	},
	{
		name:           "list user tickets - invalid UUID path param with valid JWT should return 400",
		route:          "/v1/user/invalid-uuid/tickets",
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"invalid user id"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDAlexandrina
		},
	},
	{
		name:           "list user tickets - authenticated user with no confirmed tickets returns empty list",
		route:          fmt.Sprintf("/v1/user/%s/tickets", userIDJane),
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusOK,
		expectedBody:   `{"tickets":[]}`,
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
		name:           "list user tickets - excludes unconfirmed reservations and omits expires_at on confirmed tickets",
		route:          fmt.Sprintf("/v1/user/%s/tickets", userIDJane),
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusOK,
		expectedBody:   fmt.Sprintf(`{"tickets":[{"id":"%s","event_id":"%s","user_id":"%s","status":"confirmed","created_at":"2026-10-01T12:00:00Z"}]}`, ticketID1, eventID1, userIDJane),
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

			// 1. Confirmed ticket (should appear without expires_at)
			err = s.ReserveTicket(ctx, &domain.Ticket{
				ID:        ticketID1,
				EventID:   eventID1,
				UserID:    userIDJane,
				Status:    domain.TicketStatusConfirmed,
				CreatedAt: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed confirmed ticket: %v", err)
			}

			// 2. Pending reservation (must NOT appear in confirmed tickets endpoint)
			err = s.ReserveTicket(ctx, &domain.Ticket{
				ID:        ticketIDReserved,
				EventID:   eventID1,
				UserID:    userIDJane,
				Status:    domain.TicketStatusReserved,
				ExpiresAt: time.Now().Add(15 * time.Minute),
				CreatedAt: time.Now(),
			})
			if err != nil {
				t.Fatalf("failed to seed reservation: %v", err)
			}
		},
	},
}

// ─── Run Test ────────────────────────────────────────────────────────────────

func TestUserRoutes(t *testing.T) {
	testRoutesInMemory(t, userTests)
	testRoutesHTTP(t, userTests)
}

// ─────────────────────────────────────────────────────────────────────────────
