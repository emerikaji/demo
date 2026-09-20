package store

import (
	"context"
	"strings"
	"sync"
	"time"

	"demo/domain"

	"github.com/google/uuid"
)

type MemoryStore struct {
	mu           sync.RWMutex
	users        map[uuid.UUID]*domain.User
	usersByEmail map[string]*domain.User
	events       map[uuid.UUID]*domain.Event
	tickets      map[uuid.UUID]*domain.Ticket
}

// Check interface implementation at runtime
var (
	_ domain.UserRepository   = (*MemoryStore)(nil)
	_ domain.EventRepository  = (*MemoryStore)(nil)
	_ domain.TicketRepository = (*MemoryStore)(nil)
)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:        make(map[uuid.UUID]*domain.User),
		usersByEmail: make(map[string]*domain.User),
		events:       make(map[uuid.UUID]*domain.Event),
		tickets:      make(map[uuid.UUID]*domain.Ticket),
	}
}

// ─── UserRepository ──────────────────────────────────────────────────────────

func (s *MemoryStore) CreateUser(ctx context.Context, user *domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	normalizedEmail := strings.ToLower(strings.TrimSpace(user.Email))

	if _, exists := s.users[user.ID]; exists {
		return domain.ErrConflict
	}

	if _, exists := s.usersByEmail[normalizedEmail]; exists {
		return domain.ErrEmailAlreadyExists
	}

	copied := *user
	copied.Email = normalizedEmail

	// Default role to attendee if unspecified
	if copied.Role == "" {
		copied.Role = domain.RoleAttendee
	}

	s.users[user.ID] = &copied
	s.usersByEmail[normalizedEmail] = &copied
	return nil
}

func (s *MemoryStore) UpdateUser(ctx context.Context, updated *domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.users[updated.ID]
	if !exists {
		return domain.ErrUserNotFound
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(updated.Email))

	// If email is changing, ensure new email isn't already taken by someone else
	if normalizedEmail != existing.Email {
		if _, taken := s.usersByEmail[normalizedEmail]; taken {
			return domain.ErrEmailAlreadyExists
		}
		// Remove old email from lookup index
		delete(s.usersByEmail, existing.Email)
	}

	copied := *updated
	copied.Email = normalizedEmail

	s.users[updated.ID] = &copied
	s.usersByEmail[normalizedEmail] = &copied
	return nil
}

func (s *MemoryStore) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, domain.ErrUserNotFound
	}

	copied := *user
	return &copied, nil
}

func (s *MemoryStore) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	user, exists := s.usersByEmail[normalizedEmail]
	if !exists {
		return nil, domain.ErrUserNotFound
	}

	copied := *user
	return &copied, nil
}

// ─── EventRepository ─────────────────────────────────────────────────────────

func (s *MemoryStore) CreateEvent(ctx context.Context, event *domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[event.ID]; exists {
		return domain.ErrConflict
	}

	copied := *event
	s.events[event.ID] = &copied
	return nil
}

func (s *MemoryStore) UpdateEvent(ctx context.Context, updated *domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, exists := s.events[updated.ID]
	if !exists {
		return domain.ErrNotFound
	}

	// Rule 1: Cannot modify canceled events
	if current.Status == domain.EventStatusCancelled {
		return domain.ErrEventCancelled
	}

	// Rule 2: Enforce capacity constraints on published events
	ticketsSoldOrReserved := current.Capacity - current.RemainingTickets
	if updated.Capacity < ticketsSoldOrReserved {
		return domain.ErrInvalidCapacity // Cannot drop capacity below claimed tickets
	}

	// Re-calculate remaining tickets based on new capacity
	updated.RemainingTickets = updated.Capacity - ticketsSoldOrReserved

	copied := *updated
	s.events[updated.ID] = &copied
	return nil
}

func (s *MemoryStore) GetEvent(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, exists := s.events[id]
	if !exists {
		return nil, domain.ErrNotFound
	}

	copied := *event
	return &copied, nil
}

func (s *MemoryStore) ListEvents(ctx context.Context, filter domain.EventFilter) ([]*domain.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.Event
	skipped := 0

	for _, event := range s.events {
		// Filter by status if requested
		if filter.Status != nil && event.Status != *filter.Status {
			continue
		}

		if filter.UserID != nil && event.OrganizerID != *filter.UserID {
			continue
		}

		// Apply pagination offset
		if skipped < filter.Offset {
			skipped++
			continue
		}

		// Apply pagination limit
		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}

		copied := *event
		result = append(result, &copied)
	}

	return result, nil
}

// ─── TicketRepository ────────────────────────────────────────────────────────

func (s *MemoryStore) ReserveTicket(ctx context.Context, ticket *domain.Ticket) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, exists := s.events[ticket.EventID]
	if !exists {
		return domain.ErrNotFound
	}

	if event.RemainingTickets <= 0 {
		return domain.ErrSoldOut
	}

	// Deduct capacity atomically within memory lock
	event.RemainingTickets--

	copied := *ticket
	s.tickets[ticket.ID] = &copied
	return nil
}

func (s *MemoryStore) ConfirmReservation(ctx context.Context, ticketID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ticket, exists := s.tickets[ticketID]
	if !exists {
		return domain.ErrNotFound
	}

	if time.Now().After(ticket.ExpiresAt) && ticket.Status == domain.TicketStatusReserved {
		return domain.ErrTicketExpired
	}

	ticket.Status = domain.TicketStatusConfirmed
	return nil
}

func (s *MemoryStore) CancelReservation(ctx context.Context, ticketID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ticket, exists := s.tickets[ticketID]
	if !exists {
		return domain.ErrNotFound
	}

	if ticket.Status == domain.TicketStatusCancelled {
		return nil // Idempotent
	}

	ticket.Status = domain.TicketStatusCancelled

	// Restore capacity
	if event, exists := s.events[ticket.EventID]; exists {
		event.RemainingTickets++
	}

	return nil
}

func (s *MemoryStore) GetTicket(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ticket, exists := s.tickets[id]
	if !exists {
		return nil, domain.ErrNotFound
	}

	copied := *ticket
	return &copied, nil
}

func (s *MemoryStore) ListUserTickets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*domain.Ticket
	i := 0
	for _, ticket := range s.tickets {
		if ticket.UserID != userID {
			continue
		}
		if i < offset {
			i++
			continue
		}
		if len(result) >= limit {
			break
		}
		copied := *ticket
		result = append(result, &copied)
	}

	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
