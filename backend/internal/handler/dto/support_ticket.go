package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type SupportTicket struct {
	ID            int64                  `json:"id"`
	UserID        int64                  `json:"user_id"`
	UserEmail     string                 `json:"user_email,omitempty"`
	Subject       string                 `json:"subject"`
	Category      string                 `json:"category"`
	Status        string                 `json:"status"`
	Preview       string                 `json:"preview"`
	LastMessageAt time.Time              `json:"last_message_at"`
	ClosedAt      *time.Time             `json:"closed_at"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Messages      []SupportTicketMessage `json:"messages,omitempty"`
}

type SupportTicketMessage struct {
	ID         int64     `json:"id"`
	TicketID   int64     `json:"ticket_id"`
	AuthorID   int64     `json:"author_id"`
	AuthorRole string    `json:"author_role"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}

func SupportTicketFromService(t *service.SupportTicket) *SupportTicket {
	if t == nil {
		return nil
	}
	out := &SupportTicket{ID: t.ID, UserID: t.UserID, UserEmail: t.UserEmail, Subject: t.Subject, Category: t.Category, Status: t.Status, Preview: t.Preview, LastMessageAt: t.LastMessageAt, ClosedAt: t.ClosedAt, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt}
	for _, m := range t.Messages {
		out.Messages = append(out.Messages, SupportTicketMessage{ID: m.ID, TicketID: m.TicketID, AuthorID: m.AuthorID, AuthorRole: m.AuthorRole, Body: m.Body, CreatedAt: m.CreatedAt})
	}
	return out
}

func SupportTicketsFromService(tickets []service.SupportTicket) []SupportTicket {
	out := make([]SupportTicket, 0, len(tickets))
	for i := range tickets {
		out = append(out, *SupportTicketFromService(&tickets[i]))
	}
	return out
}
