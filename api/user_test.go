package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"demo/domain"
	"demo/store"
)

// ─── Test Battery ────────────────────────────────────────────────────────────

var userTests = []apiTest{

	// ─── Events ──────────────────────────────────────────────────────────

	{
		name:           "list user events - missing auth header should return 401",
		route:          fmt.Sprintf("/v1/user/%s/events", userIDJane),
		method:         http.MethodGet,
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"unauthorized"}`,
	},
	{
		name:           "list user events - invalid UUID path param with valid JWT should return 400",
		route:          "/v1/user/invalid-uuid/events",
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"invalid user id"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDJane
		},
	},
	{
		name:           "list user events - JWT subject mismatch should return 401",
		route:          fmt.Sprintf("/v1/user/%s/events", userIDJane),
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token-for-alexandrina"},
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"unauthorized"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDAlexandrina
		},
	},
	{
		name:           "list user events - authenticated organizer with no events returns empty list",
		route:          fmt.Sprintf("/v1/user/%s/events", userIDJane),
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusOK,
		expectedBody:   `{"events":[]}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDJane

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
		name:           "list user events - successful list returns both draft and published events",
		route:          fmt.Sprintf("/v1/user/%s/events", userIDJane),
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusOK,
		expectedBody:   fmt.Sprintf(`{"events":[{"id":"%s","organizer_id":"%s","title":"GolangConf 2026","description":"Go-conference with a program built on real-world tasks","capacity":100,"remaining_tickets":100,"status":"published","starts_at":"2027-04-20T10:00:00Z","created_at":"2026-09-01T00:00:00Z"},{"id":"%s","organizer_id":"%s","title":"Private Workshop","description":"Internal draft workshop","capacity":20,"remaining_tickets":20,"status":"draft","starts_at":"2027-05-10T10:00:00Z","created_at":"2026-09-02T00:00:00Z"}]}`, eventID1, userIDJane, eventIDDraft, userIDJane),
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDJane

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJane,
				Email: "janedoe@example.com",
				Role:  domain.RoleOrganizer,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			// 1. Published event
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
				t.Fatalf("failed to seed published event: %v", err)
			}

			// 2. Draft event (visible to authorized organizer)
			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventIDDraft,
				OrganizerID:      userIDJane,
				Title:            "Private Workshop",
				Description:      "Internal draft workshop",
				Capacity:         20,
				RemainingTickets: 20,
				Status:           domain.EventStatusDraft,
				StartsAt:         time.Date(2027, 5, 10, 10, 0, 0, 0, time.UTC),
				CreatedAt:        time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed draft event: %v", err)
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
		name:           "list user tickets - JWT subject mismatch should return 401",
		route:          fmt.Sprintf("/v1/user/%s/tickets", userIDJane),
		method:         http.MethodGet,
		headers:        map[string]string{"Authorization": "Bearer valid-token-for-alexandrina"},
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"unauthorized"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDAlexandrina

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
				t.Fatalf("failed to seed published event: %v", err)
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
