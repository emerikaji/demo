package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleOrganizer UserRole = "organizer"
	RoleAttendee  UserRole = "attendee"
)

// ValidateRoleAssignment checks if a target role can be assigned by the current user.
func ValidateRoleAssignment(actorRole, targetRole UserRole) error {
	// Only an existing admin can create or promote a user to 'admin'
	if targetRole == RoleAdmin && actorRole != RoleAdmin {
		return ErrUnauthorizedRole
	}
	return nil
}

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose in JSON responses
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusPublished EventStatus = "published"
	EventStatusCancelled EventStatus = "cancelled"
)

type Event struct {
	ID               uuid.UUID   `json:"id"`
	OrganizerID      uuid.UUID   `json:"organizer_id"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	Capacity         int         `json:"capacity"`
	RemainingTickets int         `json:"remaining_tickets"`
	Status           EventStatus `json:"status"`
	StartsAt         time.Time   `json:"starts_at"`
	CreatedAt        time.Time   `json:"created_at"`
}

type TicketStatus string

const (
	TicketStatusReserved  TicketStatus = "reserved"  // Held for 15 minutes awaiting payment
	TicketStatusConfirmed TicketStatus = "confirmed" // Paid & active
	TicketStatusCancelled TicketStatus = "cancelled" // User or system released ticket
)

type Ticket struct {
	ID        uuid.UUID    `json:"id"`
	EventID   uuid.UUID    `json:"event_id"`
	UserID    uuid.UUID    `json:"user_id"`
	Status    TicketStatus `json:"status"`
	ExpiresAt time.Time    `json:"expires_at"` // Reservation TTL
	CreatedAt time.Time    `json:"created_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
