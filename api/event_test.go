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

var eventTests = []apiTest{

	// ─── List Events ─────────────────────────────────────────────────────

	{
		name:           "list events - empty store returns empty list",
		route:          "/v1/events",
		method:         http.MethodGet,
		expectedStatus: http.StatusOK,
		expectedBody:   `{"events":[]}`,
	},
	{
		name:           "list events - returns only published events omitting draft and cancelled",
		route:          "/v1/events",
		method:         http.MethodGet,
		expectedStatus: http.StatusOK,
		expectedBody:   fmt.Sprintf(`{"events":[{"id":"%s","organizer_id":"%s","title":"GolangConf 2026","description":"Go-conference with a program built on real-world tasks","capacity":100,"remaining_tickets":100,"status":"published","starts_at":"2027-04-20T10:00:00Z","created_at":"2026-09-01T00:00:00Z"}]}`, eventID1, userIDJane),
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()

			// 1. Published event (must appear in list)
			err := s.CreateEvent(ctx, &domain.Event{
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

			// 2. Draft event (must NOT appear)
			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventIDDraft,
				OrganizerID:      userIDJane,
				Title:            "Internal Draft Workshop",
				Description:      "Draft event not yet public",
				Capacity:         20,
				RemainingTickets: 20,
				Status:           domain.EventStatusDraft,
				StartsAt:         time.Date(2027, 5, 10, 10, 0, 0, 0, time.UTC),
				CreatedAt:        time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed draft event: %v", err)
			}

			// 3. Cancelled event (must NOT appear)
			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventIDCancelled,
				OrganizerID:      userIDJane,
				Title:            "Cancelled Meetup",
				Description:      "Cancelled event",
				Capacity:         50,
				RemainingTickets: 50,
				Status:           domain.EventStatusCancelled,
				StartsAt:         time.Date(2027, 6, 1, 10, 0, 0, 0, time.UTC),
				CreatedAt:        time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed cancelled event: %v", err)
			}
		},
	},
	{
		name:           "list events - filter by organizer user_id",
		route:          fmt.Sprintf("/v1/events?user_id=%s", userIDJane),
		method:         http.MethodGet,
		expectedStatus: http.StatusOK,
		expectedBody:   fmt.Sprintf(`{"events":[{"id":"%s","organizer_id":"%s","title":"GolangConf 2026","description":"Go-conference","capacity":100,"remaining_tickets":100,"status":"published","starts_at":"2027-04-20T10:00:00Z","created_at":"2026-09-01T00:00:00Z"}]}`, eventID1, userIDJane),
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()

			// Event by Jane
			err := s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      userIDJane,
				Title:            "GolangConf 2026",
				Description:      "Go-conference",
				Capacity:         100,
				RemainingTickets: 100,
				Status:           domain.EventStatusPublished,
				StartsAt:         time.Date(2027, 4, 20, 10, 0, 0, 0, time.UTC),
				CreatedAt:        time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed Jane's event: %v", err)
			}

			// Event by John
			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID2,
				OrganizerID:      userIDJohn,
				Title:            "Rust Summit 2026",
				Description:      "Rust-conference",
				Capacity:         150,
				RemainingTickets: 150,
				Status:           domain.EventStatusPublished,
				StartsAt:         time.Date(2027, 5, 20, 10, 0, 0, 0, time.UTC),
				CreatedAt:        time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed John's event: %v", err)
			}
		},
	},
	{
		name:           "list events - pagination limit and offset filter",
		route:          "/v1/events?limit=1&offset=1",
		method:         http.MethodGet,
		expectedStatus: http.StatusOK,
		expectedBody:   fmt.Sprintf(`{"events":[{"id":"%s","organizer_id":"%s","title":"Rust Summit 2026","description":"Rust-conference","capacity":150,"remaining_tickets":150,"status":"published","starts_at":"2027-05-20T10:00:00Z","created_at":"2026-09-02T00:00:00Z"}]}`, eventID2, userIDJohn),
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()

			err := s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      userIDJane,
				Title:            "GolangConf 2026",
				Description:      "Go-conference",
				Capacity:         100,
				RemainingTickets: 100,
				Status:           domain.EventStatusPublished,
				StartsAt:         time.Date(2027, 4, 20, 10, 0, 0, 0, time.UTC),
				CreatedAt:        time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed event 1: %v", err)
			}

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID2,
				OrganizerID:      userIDJohn,
				Title:            "Rust Summit 2026",
				Description:      "Rust-conference",
				Capacity:         150,
				RemainingTickets: 150,
				Status:           domain.EventStatusPublished,
				StartsAt:         time.Date(2027, 5, 20, 10, 0, 0, 0, time.UTC),
				CreatedAt:        time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Fatalf("failed to seed event 2: %v", err)
			}
		},
	},

	// ─── Get Event ───────────────────────────────────────────────────────

	{
		name:           "get event - invalid UUID path param should return 400",
		route:          "/v1/events/invalid-uuid",
		method:         http.MethodGet,
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"invalid event id"}`,
	},
	{
		name:           "get event - non-existent event should return 404",
		route:          fmt.Sprintf("/v1/events/%s", uuid.MustParse("00000000-0000-0000-0000-000000000999")),
		method:         http.MethodGet,
		expectedStatus: http.StatusNotFound,
		expectedBody:   `{"error":"event not found"}`,
	},
	{
		name:           "get event - successful retrieval of published event should return 200",
		route:          fmt.Sprintf("/v1/events/%s", eventID1),
		method:         http.MethodGet,
		expectedStatus: http.StatusOK,
		expectedBody:   fmt.Sprintf(`{"id":"%s","organizer_id":"%s","title":"GolangConf 2026","description":"Go-conference with a program built on real-world tasks","capacity":100,"remaining_tickets":100,"status":"published","starts_at":"2027-04-20T10:00:00Z","created_at":"2026-09-01T00:00:00Z"}`, eventID1, userIDJane),
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			err := s.CreateEvent(context.Background(), &domain.Event{
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

	// ─── Create Event ────────────────────────────────────────────────────

	{
		name:           "create event - missing auth header should return 401",
		route:          "/v1/events",
		method:         http.MethodPost,
		body:           []byte(`{"title":"New Conf","description":"Desc","capacity":100,"starts_at":"2027-06-01T10:00:00Z"}`),
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"unauthorized"}`,
	},
	{
		name:           "create event - attendee role should return 403 forbidden",
		route:          "/v1/events",
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(`{"title":"New Conf","description":"Desc","capacity":100,"starts_at":"2027-06-01T10:00:00Z"}`),
		expectedStatus: http.StatusForbidden,
		expectedBody:   `{"error":"forbidden: organizer role required"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDAlexandrina

			err := s.CreateUser(context.Background(), &domain.User{
				ID:    userIDAlexandrina,
				Email: "alexandrina@example.com",
				Role:  domain.RoleAttendee,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}
		},
	},
	{
		name:           "create event - empty body should return 400",
		route:          "/v1/events",
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(``),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"body must not be empty"}`,
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
		name:           "create event - missing required fields should return 400",
		route:          "/v1/events",
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(`{"description":"Missing title, capacity and starts_at"}`),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"title, capacity, and starts_at are required"}`,
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
		name:           "create event - invalid capacity <= 0 should return 400",
		route:          "/v1/events",
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(`{"title":"New Conf","capacity":0,"starts_at":"2027-06-01T10:00:00Z"}`),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"capacity must be greater than 0"}`,
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
		name:           "create event - successful creation should return 201",
		route:          "/v1/events",
		method:         http.MethodPost,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(`{"title":"Go Global 2027","description":"Global Go Summit","capacity":250,"starts_at":"2027-09-15T09:00:00Z"}`),
		expectedStatus: http.StatusCreated,
		expectedBody:   `{"message":"event created successfully"}`,
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

	// ─── Update Event ────────────────────────────────────────────────────

	{
		name:           "update event - missing auth header should return 401",
		route:          fmt.Sprintf("/v1/events/%s", eventID1),
		method:         http.MethodPatch,
		body:           []byte(`{"title":"Updated Title"}`),
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"unauthorized"}`,
	},
	{
		name:           "update event - invalid UUID path param should return 400",
		route:          "/v1/events/invalid-uuid",
		method:         http.MethodPatch,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(`{"title":"Updated Title"}`),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"invalid event id"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDJane
		},
	},
	{
		name:           "update event - empty body should return 400",
		route:          fmt.Sprintf("/v1/events/%s", eventID1),
		method:         http.MethodPatch,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(``),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"body must not be empty"}`,
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

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      userIDJane,
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
	{
		name:           "update event - non-existent event should return 404",
		route:          fmt.Sprintf("/v1/events/%s", uuid.MustParse("00000000-0000-0000-0000-000000000999")),
		method:         http.MethodPatch,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(`{"title":"Updated Title"}`),
		expectedStatus: http.StatusNotFound,
		expectedBody:   `{"error":"event not found"}`,
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
		name:           "update event - non-owning organizer should return 403 forbidden",
		route:          fmt.Sprintf("/v1/events/%s", eventID1),
		method:         http.MethodPatch,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(`{"title":"Hacked Title"}`),
		expectedStatus: http.StatusForbidden,
		expectedBody:   `{"error":"forbidden: you do not own this event"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDJohn // Authenticated as John

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJohn,
				Email: "john@example.com",
				Role:  domain.RoleOrganizer,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			// Event owned by Jane
			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      userIDJane,
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
	{
		name:           "update event - cannot update cancelled event should return 400",
		route:          fmt.Sprintf("/v1/events/%s", eventIDCancelled),
		method:         http.MethodPatch,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(`{"title":"Reactivated Title"}`),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"cannot update a cancelled event"}`,
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

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventIDCancelled,
				OrganizerID:      userIDJane,
				Title:            "Cancelled Meetup",
				Capacity:         50,
				RemainingTickets: 50,
				Status:           domain.EventStatusCancelled,
			})
			if err != nil {
				t.Fatalf("failed to seed event: %v", err)
			}
		},
	},
	{
		name:           "update event - capacity below issued tickets should return 400",
		route:          fmt.Sprintf("/v1/events/%s", eventID1),
		method:         http.MethodPatch,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(`{"capacity":5}`),
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"new capacity cannot be less than tickets already issued"}`,
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

			// 10 capacity, 2 remaining -> 8 tickets issued
			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      userIDJane,
				Title:            "GolangConf 2026",
				Capacity:         10,
				RemainingTickets: 2,
				Status:           domain.EventStatusPublished,
			})
			if err != nil {
				t.Fatalf("failed to seed event: %v", err)
			}
		},
	},
	{
		name:           "update event - successful update should return 200",
		route:          fmt.Sprintf("/v1/events/%s", eventID1),
		method:         http.MethodPatch,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		body:           []byte(`{"title":"GolangConf 2026 - Extended","capacity":150}`),
		expectedStatus: http.StatusOK,
		expectedBody:   `{"message":"event updated successfully"}`,
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

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      userIDJane,
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

	// ─── Cancel Event ────────────────────────────────────────────────────

	{
		name:           "cancel event - missing auth header should return 401",
		route:          fmt.Sprintf("/v1/events/%s", eventID1),
		method:         http.MethodDelete,
		expectedStatus: http.StatusUnauthorized,
		expectedBody:   `{"error":"unauthorized"}`,
	},
	{
		name:           "cancel event - invalid UUID path param should return 400",
		route:          "/v1/events/invalid-uuid",
		method:         http.MethodDelete,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"invalid event id"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			j.returnClaims.UserID = userIDJane
		},
	},
	{
		name:           "cancel event - non-existent event should return 404",
		route:          fmt.Sprintf("/v1/events/%s", uuid.MustParse("00000000-0000-0000-0000-000000000999")),
		method:         http.MethodDelete,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusNotFound,
		expectedBody:   `{"error":"event not found"}`,
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
		name:           "cancel event - non-owning organizer should return 403 forbidden",
		route:          fmt.Sprintf("/v1/events/%s", eventID1),
		method:         http.MethodDelete,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusForbidden,
		expectedBody:   `{"error":"forbidden: you do not own this event"}`,
		setup: func(t *testing.T, s *store.MemoryStore, j *MockJWT) {
			ctx := context.Background()
			j.returnClaims.UserID = userIDJohn

			err := s.CreateUser(ctx, &domain.User{
				ID:    userIDJohn,
				Email: "john@example.com",
				Role:  domain.RoleOrganizer,
			})
			if err != nil {
				t.Fatalf("failed to seed user: %v", err)
			}

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      userIDJane,
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
	{
		name:           "cancel event - already cancelled event should return 400",
		route:          fmt.Sprintf("/v1/events/%s", eventIDCancelled),
		method:         http.MethodDelete,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusBadRequest,
		expectedBody:   `{"error":"cannot update a cancelled event"}`,
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

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventIDCancelled,
				OrganizerID:      userIDJane,
				Title:            "Cancelled Meetup",
				Capacity:         50,
				RemainingTickets: 50,
				Status:           domain.EventStatusCancelled,
			})
			if err != nil {
				t.Fatalf("failed to seed event: %v", err)
			}
		},
	},
	{
		name:           "cancel event - successful cancellation should return 200",
		route:          fmt.Sprintf("/v1/events/%s", eventID1),
		method:         http.MethodDelete,
		headers:        map[string]string{"Authorization": "Bearer valid-token"},
		expectedStatus: http.StatusOK,
		expectedBody:   `{"message":"event cancelled successfully"}`,
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

			err = s.CreateEvent(ctx, &domain.Event{
				ID:               eventID1,
				OrganizerID:      userIDJane,
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
}

// ─── Run Test ────────────────────────────────────────────────────────────────

func TestEvent(t *testing.T) {
	testRoutesInMemory(t, eventTests)
	testRoutesHTTP(t, eventTests)
}

// ─────────────────────────────────────────────────────────────────────────────
