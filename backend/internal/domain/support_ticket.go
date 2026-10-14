package domain

import (
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	TicketCategoryBilling    = "billing"
	TicketCategoryAPIModels  = "api_models"
	TicketCategoryAccount    = "account"
	TicketCategoryOther      = "other"
	TicketStatusOpen         = "open"
	TicketStatusPendingAdmin = "pending_admin"
	TicketStatusPendingUser  = "pending_user"
	TicketStatusClosed       = "closed"
	TicketOpenMax            = 10
	TicketMessageMax         = 200
)

var (
	ErrTicketNotFound        = infraerrors.NotFound("TICKET_NOT_FOUND", "ticket not found")
	ErrTicketSubjectInvalid  = infraerrors.BadRequest("TICKET_SUBJECT_INVALID", "ticket subject exceeds 80 characters")
	ErrTicketBodyInvalid     = infraerrors.BadRequest("TICKET_BODY_INVALID", "ticket body exceeds 4000 characters")
	ErrTicketCategoryInvalid = infraerrors.BadRequest("TICKET_CATEGORY_INVALID", "invalid ticket category")
	ErrTicketStatusInvalid   = infraerrors.BadRequest("TICKET_STATUS_INVALID", "invalid ticket status")
	ErrTicketInputRequired   = infraerrors.BadRequest("TICKET_INPUT_REQUIRED", "ticket subject and body are required")
	ErrTicketInputInvalid    = infraerrors.BadRequest("TICKET_INPUT_INVALID", "ticket text contains invalid characters")
	ErrTicketQueryInvalid    = infraerrors.BadRequest("TICKET_QUERY_INVALID", "invalid ticket pagination or search")
	ErrTicketMessageLimit    = infraerrors.Conflict("TICKET_MESSAGE_LIMIT", "ticket conversation is full; please create a new ticket")
	ErrTicketClosed          = infraerrors.Conflict("TICKET_CLOSED", "ticket is closed")
	ErrTicketOpenLimit       = infraerrors.Conflict("TICKET_OPEN_LIMIT", "too many open tickets").WithMetadata(map[string]string{"max": "10"})
)

type SupportTicket struct {
	ID            int64
	UserID        int64
	UserEmail     string
	Subject       string
	Category      string
	Status        string
	Preview       string
	LastMessageAt time.Time
	ClosedAt      *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Messages      []SupportTicketMessage
}

type SupportTicketMessage struct {
	ID         int64
	TicketID   int64
	AuthorID   int64
	AuthorRole string
	Body       string
	CreatedAt  time.Time
}
