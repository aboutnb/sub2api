package service

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	TicketCategoryBilling    = domain.TicketCategoryBilling
	TicketCategoryAPIModels  = domain.TicketCategoryAPIModels
	TicketCategoryAccount    = domain.TicketCategoryAccount
	TicketCategoryOther      = domain.TicketCategoryOther
	TicketStatusOpen         = domain.TicketStatusOpen
	TicketStatusPendingAdmin = domain.TicketStatusPendingAdmin
	TicketStatusPendingUser  = domain.TicketStatusPendingUser
	TicketStatusClosed       = domain.TicketStatusClosed
	TicketOpenMax            = domain.TicketOpenMax
	TicketMessageMax         = domain.TicketMessageMax
)

var (
	ErrTicketNotFound        = domain.ErrTicketNotFound
	ErrTicketSubjectInvalid  = domain.ErrTicketSubjectInvalid
	ErrTicketBodyInvalid     = domain.ErrTicketBodyInvalid
	ErrTicketCategoryInvalid = domain.ErrTicketCategoryInvalid
	ErrTicketStatusInvalid   = domain.ErrTicketStatusInvalid
	ErrTicketInputRequired   = domain.ErrTicketInputRequired
	ErrTicketInputInvalid    = domain.ErrTicketInputInvalid
	ErrTicketQueryInvalid    = domain.ErrTicketQueryInvalid
	ErrTicketMessageLimit    = domain.ErrTicketMessageLimit
	ErrTicketClosed          = domain.ErrTicketClosed
	ErrTicketOpenLimit       = domain.ErrTicketOpenLimit
)

type SupportTicket = domain.SupportTicket
type SupportTicketMessage = domain.SupportTicketMessage

type CreateTicketInput struct{ Subject, Body, Category string }

type SupportTicketListFilters struct {
	UserID   int64
	Category string
	Statuses []string
	Search   string
}

type SupportTicketRepository interface {
	WithTx(context.Context, func(context.Context) error) error
	LockUser(context.Context, int64) error
	CountOpenByUser(context.Context, int64) (int, error)
	CountMessages(context.Context, int64) (int, error)
	GetTicket(context.Context, int64) (*SupportTicket, error)
	TicketOwnerID(context.Context, int64) (int64, error)
	LockTicket(context.Context, int64) (*SupportTicket, error)
	InsertTicket(context.Context, *SupportTicket) error
	AddMessage(context.Context, *SupportTicketMessage) error
	UpdateTicket(context.Context, *SupportTicket) error
	List(context.Context, pagination.PaginationParams, SupportTicketListFilters) ([]SupportTicket, *pagination.PaginationResult, error)
	ListMessages(context.Context, int64) ([]SupportTicketMessage, error)
	EmailsByUserIDs(context.Context, []int64) (map[int64]string, error)
}

type SupportTicketService struct{ repo SupportTicketRepository }

func NewSupportTicketService(repo SupportTicketRepository) *SupportTicketService {
	return &SupportTicketService{repo: repo}
}

func ticketCategory(value string, includeAll bool) (string, error) {
	value = strings.TrimSpace(value)
	switch value {
	case "", "all":
		if includeAll {
			return "", nil
		}
	case TicketCategoryBilling, TicketCategoryAPIModels, TicketCategoryAccount, TicketCategoryOther:
		return value, nil
	}
	return "", ErrTicketCategoryInvalid
}

func validTicketStatus(value string) bool {
	switch value {
	case TicketStatusOpen, TicketStatusPendingAdmin, TicketStatusPendingUser, TicketStatusClosed:
		return true
	}
	return false
}

func validateTicketBody(body string) (string, error) {
	if !validTicketText(body) {
		return "", ErrTicketInputInvalid
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return "", ErrTicketInputRequired
	}
	if utf8.RuneCountInString(body) > 4000 {
		return "", ErrTicketBodyInvalid
	}
	return body, nil
}

func validTicketText(text string) bool {
	return utf8.ValidString(text) && !strings.ContainsRune(text, '\x00')
}

func validateTicketQuery(params pagination.PaginationParams, search string) error {
	if params.Page > 1000000 || params.PageSize > 100 || params.Page < 0 || params.PageSize < 0 ||
		!validTicketText(search) || utf8.RuneCountInString(search) > 200 {
		return ErrTicketQueryInvalid
	}
	return nil
}

func ticketPreview(body string) string {
	runes := []rune(strings.Join(strings.Fields(body), " "))
	if len(runes) > 80 {
		runes = runes[:80]
	}
	return string(runes)
}

func (s *SupportTicketService) Create(ctx context.Context, userID int64, input CreateTicketInput) (*SupportTicket, error) {
	if !validTicketText(input.Subject) {
		return nil, ErrTicketInputInvalid
	}
	input.Subject = strings.TrimSpace(input.Subject)
	if input.Subject == "" {
		return nil, ErrTicketInputRequired
	}
	if utf8.RuneCountInString(input.Subject) > 80 {
		return nil, ErrTicketSubjectInvalid
	}
	body, err := validateTicketBody(input.Body)
	if err != nil {
		return nil, err
	}
	category, err := ticketCategory(input.Category, false)
	if err != nil {
		return nil, err
	}
	var ticket *SupportTicket
	err = s.repo.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.LockUser(txCtx, userID); err != nil {
			return err
		}
		count, err := s.repo.CountOpenByUser(txCtx, userID)
		if err != nil {
			return err
		}
		if count >= TicketOpenMax {
			return ErrTicketOpenLimit
		}
		now := time.Now()
		ticket = &SupportTicket{UserID: userID, Subject: input.Subject, Category: category, Status: TicketStatusOpen, LastMessageAt: now, CreatedAt: now, UpdatedAt: now}
		if err := s.repo.InsertTicket(txCtx, ticket); err != nil {
			return err
		}
		message := &SupportTicketMessage{TicketID: ticket.ID, AuthorID: userID, AuthorRole: RoleUser, Body: body, CreatedAt: now}
		if err := s.repo.AddMessage(txCtx, message); err != nil {
			return err
		}
		ticket.Preview = ticketPreview(body)
		if err := s.repo.UpdateTicket(txCtx, ticket); err != nil {
			return err
		}
		ticket.Messages = []SupportTicketMessage{*message}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ticket, nil
}

func (s *SupportTicketService) ListForUser(ctx context.Context, userID int64, params pagination.PaginationParams, category, status, search string) ([]SupportTicket, *pagination.PaginationResult, error) {
	if err := validateTicketQuery(params, search); err != nil {
		return nil, nil, err
	}
	category, err := ticketCategory(category, true)
	if err != nil {
		return nil, nil, err
	}
	status = strings.TrimSpace(status)
	filters := SupportTicketListFilters{UserID: userID, Category: category, Search: strings.TrimSpace(search)}
	if status == "unclosed" {
		filters.Statuses = []string{TicketStatusOpen, TicketStatusPendingAdmin, TicketStatusPendingUser}
	} else if status != "" {
		if !validTicketStatus(status) {
			return nil, nil, ErrTicketStatusInvalid
		}
		filters.Statuses = []string{status}
	}
	return s.repo.List(ctx, params, filters)
}

func (s *SupportTicketService) ListForAdmin(ctx context.Context, params pagination.PaginationParams, category, queue, search string) ([]SupportTicket, *pagination.PaginationResult, error) {
	if err := validateTicketQuery(params, search); err != nil {
		return nil, nil, err
	}
	category, err := ticketCategory(category, true)
	if err != nil {
		return nil, nil, err
	}
	filters := SupportTicketListFilters{Category: category, Search: strings.TrimSpace(search)}
	switch strings.TrimSpace(queue) {
	case "", "unanswered":
		filters.Statuses = []string{TicketStatusOpen, TicketStatusPendingAdmin}
	case "awaiting_user":
		filters.Statuses = []string{TicketStatusPendingUser}
	case "closed":
		filters.Statuses = []string{TicketStatusClosed}
	default:
		return nil, nil, ErrTicketStatusInvalid
	}
	items, result, err := s.repo.List(ctx, params, filters)
	if err != nil {
		return nil, nil, err
	}
	if err := s.attachEmails(ctx, items); err != nil {
		return nil, nil, err
	}
	return items, result, nil
}

func (s *SupportTicketService) attachEmails(ctx context.Context, tickets []SupportTicket) error {
	ids := make([]int64, 0, len(tickets))
	seen := make(map[int64]bool)
	for _, ticket := range tickets {
		if !seen[ticket.UserID] {
			ids = append(ids, ticket.UserID)
			seen[ticket.UserID] = true
		}
	}
	if len(ids) == 0 {
		return nil
	}
	emails, err := s.repo.EmailsByUserIDs(ctx, ids)
	if err != nil {
		return err
	}
	for i := range tickets {
		tickets[i].UserEmail = emails[tickets[i].UserID]
	}
	return nil
}

func (s *SupportTicketService) detail(ctx context.Context, ticket *SupportTicket, admin bool) (*SupportTicket, error) {
	messages, err := s.repo.ListMessages(ctx, ticket.ID)
	if err != nil {
		return nil, err
	}
	ticket.Messages = messages
	if admin {
		tickets := []SupportTicket{*ticket}
		if err := s.attachEmails(ctx, tickets); err != nil {
			return nil, err
		}
		ticket.UserEmail = tickets[0].UserEmail
	}
	return ticket, nil
}

func (s *SupportTicketService) GetForUser(ctx context.Context, userID, ticketID int64) (*SupportTicket, error) {
	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket.UserID != userID {
		return nil, ErrTicketNotFound
	}
	return s.detail(ctx, ticket, false)
}

func (s *SupportTicketService) GetForAdmin(ctx context.Context, ticketID int64) (*SupportTicket, error) {
	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	return s.detail(ctx, ticket, true)
}

func (s *SupportTicketService) lockOwned(ctx context.Context, userID, ticketID int64) (*SupportTicket, error) {
	ownerID, err := s.repo.TicketOwnerID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ownerID != userID {
		return nil, ErrTicketNotFound
	}
	ticket, err := s.repo.LockTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket.UserID != userID {
		return nil, ErrTicketNotFound
	}
	return ticket, nil
}

func (s *SupportTicketService) reply(ctx context.Context, authorID, ticketID int64, body string, admin bool) (*SupportTicket, error) {
	body, err := validateTicketBody(body)
	if err != nil {
		return nil, err
	}
	err = s.repo.WithTx(ctx, func(txCtx context.Context) error {
		var ticket *SupportTicket
		var err error
		if admin {
			ticket, err = s.repo.LockTicket(txCtx, ticketID)
		} else {
			ticket, err = s.lockOwned(txCtx, authorID, ticketID)
		}
		if err != nil {
			return err
		}
		if !admin && ticket.Status == TicketStatusClosed {
			return ErrTicketClosed
		}
		// The ticket row is locked, so concurrent replies cannot exceed the cap.
		count, err := s.repo.CountMessages(txCtx, ticketID)
		if err != nil {
			return err
		}
		if count >= TicketMessageMax {
			return ErrTicketMessageLimit
		}
		now := time.Now()
		role, status := RoleUser, TicketStatusPendingAdmin
		if admin {
			role, status = RoleAdmin, TicketStatusPendingUser
		}
		message := &SupportTicketMessage{TicketID: ticketID, AuthorID: authorID, AuthorRole: role, Body: body, CreatedAt: now}
		if err := s.repo.AddMessage(txCtx, message); err != nil {
			return err
		}
		ticket.Status, ticket.ClosedAt = status, nil
		ticket.Preview, ticket.LastMessageAt, ticket.UpdatedAt = ticketPreview(body), now, now
		return s.repo.UpdateTicket(txCtx, ticket)
	})
	if err != nil {
		return nil, err
	}
	if admin {
		return s.GetForAdmin(ctx, ticketID)
	}
	return s.GetForUser(ctx, authorID, ticketID)
}

func (s *SupportTicketService) ReplyAsUser(ctx context.Context, userID, ticketID int64, body string) (*SupportTicket, error) {
	return s.reply(ctx, userID, ticketID, body, false)
}

func (s *SupportTicketService) ReplyAsAdmin(ctx context.Context, adminID, ticketID int64, body string) (*SupportTicket, error) {
	return s.reply(ctx, adminID, ticketID, body, true)
}

func setTicketStatus(ticket *SupportTicket, status string) bool {
	if ticket.Status == status {
		return false
	}
	now := time.Now()
	ticket.Status, ticket.UpdatedAt, ticket.ClosedAt = status, now, nil
	if status == TicketStatusClosed {
		ticket.ClosedAt = &now
	}
	return true
}

func (s *SupportTicketService) CloseForUser(ctx context.Context, userID, ticketID int64) (*SupportTicket, error) {
	err := s.repo.WithTx(ctx, func(txCtx context.Context) error {
		ticket, err := s.lockOwned(txCtx, userID, ticketID)
		if err != nil {
			return err
		}
		if !setTicketStatus(ticket, TicketStatusClosed) {
			return nil
		}
		return s.repo.UpdateTicket(txCtx, ticket)
	})
	if err != nil {
		return nil, err
	}
	return s.GetForUser(ctx, userID, ticketID)
}

func (s *SupportTicketService) UpdateStatus(ctx context.Context, ticketID int64, status string) (*SupportTicket, error) {
	status = strings.TrimSpace(status)
	if !validTicketStatus(status) {
		return nil, ErrTicketStatusInvalid
	}
	err := s.repo.WithTx(ctx, func(txCtx context.Context) error {
		ticket, err := s.repo.LockTicket(txCtx, ticketID)
		if err != nil {
			return err
		}
		if !setTicketStatus(ticket, status) {
			return nil
		}
		return s.repo.UpdateTicket(txCtx, ticket)
	})
	if err != nil {
		return nil, err
	}
	return s.GetForAdmin(ctx, ticketID)
}
