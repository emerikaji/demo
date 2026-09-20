package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotFound           = errors.New("resource not found")
	ErrConflict           = errors.New("resource already exists")
	ErrSoldOut            = errors.New("event is sold out")
	ErrTicketExpired      = errors.New("ticket reservation has expired")
	ErrEventCancelled     = errors.New("cannot update a cancelled event")
	ErrInvalidCapacity    = errors.New("new capacity cannot be less than tickets already issued")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("user with this email already exists")
	ErrUnauthorizedRole   = errors.New("insufficient privileges to assign this role")
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type EventFilter struct {
	Status *EventStatus // Optional status filter
	UserID *uuid.UUID   // Optional user filter
	Limit  int
	Offset int
}

type EventRepository interface {
	CreateEvent(ctx context.Context, event *Event) error
	UpdateEvent(ctx context.Context, event *Event) error
	GetEvent(ctx context.Context, id uuid.UUID) (*Event, error)
	ListEvents(ctx context.Context, filter EventFilter) ([]*Event, error)
}

type TicketRepository interface {
	ReserveTicket(ctx context.Context, ticket *Ticket) error
	ConfirmReservation(ctx context.Context, ticketID uuid.UUID) error
	CancelReservation(ctx context.Context, ticketID uuid.UUID) error
	GetTicket(ctx context.Context, id uuid.UUID) (*Ticket, error)
	ListUserTickets(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Ticket, error)
}

// ─────────────────────────────────────────────────────────────────────────────
