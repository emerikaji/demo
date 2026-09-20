package store_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"demo/domain"
	"demo/store"

	"github.com/google/uuid"
)

// setupStore creates a fresh instance of MemoryStore for each test run.
func setupStore() *store.MemoryStore {
	return store.NewMemoryStore()
}

// ─── UserRepository Tests ────────────────────────────────────────────────────

func TestCreateUser(t *testing.T) {
	t.Parallel()

	t.Run("successfully creates user with default attendee role", func(t *testing.T) {
		t.Parallel()
		s := setupStore()
		ctx := context.Background()

		user := &domain.User{
			ID:        uuid.New(),
			Email:     "  Jane.Doe@Example.com  ",
			CreatedAt: time.Now(),
		}

		err := s.CreateUser(ctx, user)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Fetch user to verify normalization and default role assignment
		fetched, err := s.GetUserByID(ctx, user.ID)
		if err != nil {
			t.Fatalf("expected user to exist, got %v", err)
		}

		expectedEmail := "jane.doe@example.com"

		if fetched.Email != expectedEmail {
			t.Errorf("expected email %s, got '%s'", expectedEmail, fetched.Email)
		}
		if fetched.Role != domain.RoleAttendee {
			t.Errorf("expected default role 'attendee', got '%s'", fetched.Role)
		}
	})

	t.Run("returns ErrConflict when ID already exists", func(t *testing.T) {
		t.Parallel()
		s := setupStore()
		ctx := context.Background()

		userID := uuid.New()
		u1 := &domain.User{ID: userID, Email: "drusilla@example.com"}
		u2 := &domain.User{ID: userID, Email: "anastella@example.com"}

		_ = s.CreateUser(ctx, u1)
		err := s.CreateUser(ctx, u2)

		if !errors.Is(err, domain.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("returns ErrEmailAlreadyExists when email is registered", func(t *testing.T) {
		t.Parallel()
		s := setupStore()
		ctx := context.Background()

		u1 := &domain.User{ID: uuid.New(), Email: "jane.doe@example.com"}
		u2 := &domain.User{ID: uuid.New(), Email: "JANE.DOE@example.com"}

		_ = s.CreateUser(ctx, u1)
		err := s.CreateUser(ctx, u2)

		if !errors.Is(err, domain.ErrEmailAlreadyExists) {
			t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
		}
	})
}

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	t.Run("successfully updates email and updates lookup index", func(t *testing.T) {
		t.Parallel()
		s := setupStore()
		ctx := context.Background()

		user := &domain.User{ID: uuid.New(), Email: "old@example.com", Role: domain.RoleAttendee}
		_ = s.CreateUser(ctx, user)

		// Update email
		user.Email = "new@example.com"
		err := s.UpdateUser(ctx, user)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Old email should no longer be found
		_, err = s.GetUserByEmail(ctx, "old@example.com")
		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("expected old email lookup to return ErrUserNotFound, got %v", err)
		}

		// New email should successfully resolve
		fetched, err := s.GetUserByEmail(ctx, "new@example.com")
		if err != nil || fetched.ID != user.ID {
			t.Errorf("failed to fetch user via updated email index")
		}
	})

	t.Run("fails to update email if target email belongs to another user", func(t *testing.T) {
		t.Parallel()
		s := setupStore()
		ctx := context.Background()

		u1 := &domain.User{ID: uuid.New(), Email: "drusilla@example.com"}
		u2 := &domain.User{ID: uuid.New(), Email: "anastella@example.com"}
		_ = s.CreateUser(ctx, u1)
		_ = s.CreateUser(ctx, u2)

		// Try updating u2's email to u1's email
		u2.Email = "drusilla@example.com"
		err := s.UpdateUser(ctx, u2)
		if !errors.Is(err, domain.ErrEmailAlreadyExists) {
			t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
		}
	})
}

func TestGetUserByID(t *testing.T) {
	t.Parallel()

	t.Run("returns ErrUserNotFound for missing ID", func(t *testing.T) {
		t.Parallel()
		s := setupStore()

		_, err := s.GetUserByID(context.Background(), uuid.New())
		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestGetUserByEmail(t *testing.T) {
	t.Parallel()

	t.Run("fetches user case-insensitively", func(t *testing.T) {
		t.Parallel()
		s := setupStore()
		ctx := context.Background()

		user := &domain.User{ID: uuid.New(), Email: "jane.doe@example.com"}
		_ = s.CreateUser(ctx, user)

		fetched, err := s.GetUserByEmail(ctx, "JANE.DOE@EXAMPLE.COM")
		if err != nil {
			t.Fatalf("expected to find user, got %v", err)
		}
		if fetched.ID != user.ID {
			t.Errorf("expected user ID %s, got %s", user.ID, fetched.ID)
		}
	})
}

// ─── EventRepository Tests ───────────────────────────────────────────────────

func TestCreateAndGetEvent(t *testing.T) {
	t.Parallel()

	s := setupStore()
	ctx := context.Background()

	event := &domain.Event{
		ID:               uuid.New(),
		OrganizerID:      uuid.New(),
		Title:            "GolangConf",
		Capacity:         100,
		RemainingTickets: 100,
		Status:           domain.EventStatusPublished,
	}

	if err := s.CreateEvent(ctx, event); err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	fetched, err := s.GetEvent(ctx, event.ID)
	if err != nil {
		t.Fatalf("failed to get event: %v", err)
	}

	if fetched.Title != event.Title {
		t.Errorf("expected title '%s', got '%s'", event.Title, fetched.Title)
	}
}

func TestUpdateEvent(t *testing.T) {
	t.Parallel()

	t.Run("cannot update cancelled event", func(t *testing.T) {
		t.Parallel()
		s := setupStore()
		ctx := context.Background()

		event := &domain.Event{
			ID:               uuid.New(),
			Capacity:         50,
			RemainingTickets: 50,
			Status:           domain.EventStatusCancelled,
		}
		_ = s.CreateEvent(ctx, event)

		event.Title = "New Title"
		err := s.UpdateEvent(ctx, event)
		if !errors.Is(err, domain.ErrEventCancelled) {
			t.Errorf("expected ErrEventCancelled, got %v", err)
		}
	})

	t.Run("prevents lowering capacity below issued tickets", func(t *testing.T) {
		t.Parallel()
		s := setupStore()
		ctx := context.Background()

		// Initial capacity 10, remaining 2 -> 8 tickets issued/sold
		event := &domain.Event{
			ID:               uuid.New(),
			Capacity:         10,
			RemainingTickets: 2,
			Status:           domain.EventStatusPublished,
		}
		_ = s.CreateEvent(ctx, event)

		// Attempt to reduce total capacity to 5 (less than 8 issued)
		event.Capacity = 5
		err := s.UpdateEvent(ctx, event)
		if !errors.Is(err, domain.ErrInvalidCapacity) {
			t.Errorf("expected ErrInvalidCapacity, got %v", err)
		}
	})

	t.Run("adjusts remaining tickets when capacity is increased", func(t *testing.T) {
		t.Parallel()
		s := setupStore()
		ctx := context.Background()

		// 10 capacity, 2 remaining = 8 tickets issued
		event := &domain.Event{
			ID:               uuid.New(),
			Capacity:         10,
			RemainingTickets: 2,
			Status:           domain.EventStatusPublished,
		}
		_ = s.CreateEvent(ctx, event)

		// Increase capacity to 20. Remaining should expand to 12 (20 total - 8 issued)
		event.Capacity = 20
		if err := s.UpdateEvent(ctx, event); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		updated, _ := s.GetEvent(ctx, event.ID)
		if updated.RemainingTickets != 12 {
			t.Errorf("expected remaining tickets to be 12, got %d", updated.RemainingTickets)
		}
	})
}

func TestListEvents(t *testing.T) {
	t.Parallel()

	s := setupStore()
	ctx := context.Background()

	// Create test organizers
	organizerA := uuid.New()
	organizerB := uuid.New()

	publishedStatus := domain.EventStatusPublished
	draftStatus := domain.EventStatusDraft
	cancelledStatus := domain.EventStatusCancelled

	// Organizer A: 2 Published, 1 Draft
	for i := 0; i < 2; i++ {
		_ = s.CreateEvent(ctx, &domain.Event{
			ID:          uuid.New(),
			OrganizerID: organizerA,
			Title:       fmt.Sprintf("Organizer A Published %d", i),
			Status:      publishedStatus,
		})
	}
	_ = s.CreateEvent(ctx, &domain.Event{
		ID:          uuid.New(),
		OrganizerID: organizerA,
		Title:       "Organizer A Draft",
		Status:      draftStatus,
	})

	// Organizer B: 1 Published, 1 Draft, 1 Cancelled
	_ = s.CreateEvent(ctx, &domain.Event{
		ID:          uuid.New(),
		OrganizerID: organizerB,
		Title:       "Organizer B Published",
		Status:      publishedStatus,
	})
	_ = s.CreateEvent(ctx, &domain.Event{
		ID:          uuid.New(),
		OrganizerID: organizerB,
		Title:       "Organizer B Draft",
		Status:      draftStatus,
	})
	_ = s.CreateEvent(ctx, &domain.Event{
		ID:          uuid.New(),
		OrganizerID: organizerB,
		Title:       "Organizer B Cancelled",
		Status:      cancelledStatus,
	})

	t.Run("lists all events when no filter is specified", func(t *testing.T) {
		t.Parallel()

		events, err := s.ListEvents(ctx, domain.EventFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 6 {
			t.Errorf("expected 6 events, got %d", len(events))
		}
	})

	t.Run("filters events by UserID (organizer)", func(t *testing.T) {
		t.Parallel()

		events, err := s.ListEvents(ctx, domain.EventFilter{
			UserID: &organizerA,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 3 {
			t.Errorf("expected 3 events for Organizer A, got %d", len(events))
		}
		for _, e := range events {
			if e.OrganizerID != organizerA {
				t.Errorf("expected OrganizerID %s, got %s", organizerA, e.OrganizerID)
			}
		}
	})

	t.Run("filters events by both UserID and Status", func(t *testing.T) {
		t.Parallel()

		events, err := s.ListEvents(ctx, domain.EventFilter{
			UserID: &organizerA,
			Status: &publishedStatus,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("expected 2 published events for Organizer A, got %d", len(events))
		}
		for _, e := range events {
			if e.OrganizerID != organizerA || e.Status != domain.EventStatusPublished {
				t.Errorf("event did not match combined filter criteria: %+v", e)
			}
		}
	})

	t.Run("applies pagination limit and offset with filters", func(t *testing.T) {
		t.Parallel()

		events, err := s.ListEvents(ctx, domain.EventFilter{
			Status: &publishedStatus,
			Limit:  2,
			Offset: 1,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("expected 2 published events, got %d", len(events))
		}
	})
}

// ─── TicketRepository Tests ──────────────────────────────────────────────────

func TestTicketLifecycle(t *testing.T) {
	t.Parallel()

	s := setupStore()
	ctx := context.Background()

	event := &domain.Event{
		ID:               uuid.New(),
		Capacity:         1,
		RemainingTickets: 1,
		Status:           domain.EventStatusPublished,
	}
	_ = s.CreateEvent(ctx, event)

	ticket := &domain.Ticket{
		ID:        uuid.New(),
		EventID:   event.ID,
		UserID:    uuid.New(),
		Status:    domain.TicketStatusReserved,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	// 1. Reserve Ticket
	if err := s.ReserveTicket(ctx, ticket); err != nil {
		t.Fatalf("failed to reserve ticket: %v", err)
	}

	// Verify event capacity decreased
	evt, _ := s.GetEvent(ctx, event.ID)
	if evt.RemainingTickets != 0 {
		t.Errorf("expected remaining tickets to be 0, got %d", evt.RemainingTickets)
	}

	// 2. Reject reservation when sold out
	secondTicket := &domain.Ticket{ID: uuid.New(), EventID: event.ID, UserID: uuid.New()}
	if err := s.ReserveTicket(ctx, secondTicket); !errors.Is(err, domain.ErrSoldOut) {
		t.Errorf("expected ErrSoldOut, got %v", err)
	}

	// 3. Confirm Reservation
	if err := s.ConfirmReservation(ctx, ticket.ID); err != nil {
		t.Fatalf("failed to confirm ticket: %v", err)
	}

	confirmed, _ := s.GetTicket(ctx, ticket.ID)
	if confirmed.Status != domain.TicketStatusConfirmed {
		t.Errorf("expected status 'confirmed', got '%s'", confirmed.Status)
	}

	// 4. Cancel Reservation & Restore Capacity
	if err := s.CancelReservation(ctx, ticket.ID); err != nil {
		t.Fatalf("failed to cancel reservation: %v", err)
	}

	evt, _ = s.GetEvent(ctx, event.ID)
	if evt.RemainingTickets != 1 {
		t.Errorf("expected restored capacity to be 1, got %d", evt.RemainingTickets)
	}
}

func TestConfirmExpiredTicket(t *testing.T) {
	t.Parallel()

	s := setupStore()
	ctx := context.Background()

	event := &domain.Event{ID: uuid.New(), Capacity: 10, RemainingTickets: 10}
	_ = s.CreateEvent(ctx, event)

	ticket := &domain.Ticket{
		ID:        uuid.New(),
		EventID:   event.ID,
		UserID:    uuid.New(),
		Status:    domain.TicketStatusReserved,
		ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired in past
	}
	_ = s.ReserveTicket(ctx, ticket)

	err := s.ConfirmReservation(ctx, ticket.ID)
	if !errors.Is(err, domain.ErrTicketExpired) {
		t.Errorf("expected ErrTicketExpired, got %v", err)
	}
}

// ─── Concurrency Safety Test ─────────────────────────────────────────────────

func TestConcurrentTicketReservations(t *testing.T) {
	t.Parallel()

	s := setupStore()
	ctx := context.Background()

	totalCapacity := 50
	event := &domain.Event{
		ID:               uuid.New(),
		Capacity:         totalCapacity,
		RemainingTickets: totalCapacity,
	}
	_ = s.CreateEvent(ctx, event)

	var wg sync.WaitGroup
	workers := 100 // 100 concurrent workers competing for 50 tickets

	successCount := 0
	soldOutCount := 0
	var countMu sync.Mutex

	for range workers {
		wg.Go(func() {
			ticket := &domain.Ticket{
				ID:        uuid.New(),
				EventID:   event.ID,
				UserID:    uuid.New(),
				Status:    domain.TicketStatusReserved,
				ExpiresAt: time.Now().Add(15 * time.Minute),
			}

			err := s.ReserveTicket(ctx, ticket)

			countMu.Lock()
			defer countMu.Unlock()
			if err == nil {
				successCount++
			} else if errors.Is(err, domain.ErrSoldOut) {
				soldOutCount++
			}
		})
	}

	wg.Wait()

	if successCount != totalCapacity {
		t.Errorf("expected exactly %d successful reservations, got %d", totalCapacity, successCount)
	}

	if soldOutCount != workers-totalCapacity {
		t.Errorf("expected exactly %d sold out errors, got %d", workers-totalCapacity, soldOutCount)
	}

	evt, _ := s.GetEvent(ctx, event.ID)
	if evt.RemainingTickets != 0 {
		t.Errorf("expected remaining tickets to be 0, got %d", evt.RemainingTickets)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
