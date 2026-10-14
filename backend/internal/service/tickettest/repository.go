// Package tickettest provides an in-memory repository for support ticket tests.
package tickettest

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type state struct {
	tickets             map[int64]service.SupportTicket
	messages            map[int64][]service.SupportTicketMessage
	users               map[int64]string
	ticketID, messageID int64
}

type Repository struct {
	mu, txMu sync.Mutex
	state    state
}
type txKey struct{}

func New() *Repository {
	return &Repository{state: state{tickets: make(map[int64]service.SupportTicket), messages: make(map[int64][]service.SupportTicketMessage), users: make(map[int64]string)}}
}

func cloneTicket(t service.SupportTicket) service.SupportTicket {
	if t.ClosedAt != nil {
		closedAt := *t.ClosedAt
		t.ClosedAt = &closedAt
	}
	t.Messages = append([]service.SupportTicketMessage(nil), t.Messages...)
	return t
}

func cloneState(s state) state {
	out := state{tickets: make(map[int64]service.SupportTicket), messages: make(map[int64][]service.SupportTicketMessage), users: make(map[int64]string), ticketID: s.ticketID, messageID: s.messageID}
	for id, ticket := range s.tickets {
		out.tickets[id] = cloneTicket(ticket)
	}
	for id, messages := range s.messages {
		out.messages[id] = append([]service.SupportTicketMessage(nil), messages...)
	}
	for id, email := range s.users {
		out.users[id] = email
	}
	return out
}

func (r *Repository) AddUser(id int64, email string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.users[id] = email
}

func (r *Repository) WithTx(ctx context.Context, fn func(context.Context) error) (err error) {
	if ctx.Value(txKey{}) == r {
		return fn(ctx)
	}
	r.txMu.Lock()
	defer r.txMu.Unlock()
	r.mu.Lock()
	snapshot := cloneState(r.state)
	r.mu.Unlock()
	committed := false
	defer func() {
		if !committed {
			r.mu.Lock()
			r.state = snapshot
			r.mu.Unlock()
		}
	}()
	// Callbacks use repository methods which take their own locks.
	err = fn(context.WithValue(ctx, txKey{}, r))
	committed = err == nil
	return err
}

func (r *Repository) LockUser(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.state.users[id]; !ok {
		return service.ErrUserNotFound
	}
	return nil
}

func (r *Repository) CountOpenByUser(_ context.Context, userID int64) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, t := range r.state.tickets {
		if t.UserID == userID && t.Status != service.TicketStatusClosed {
			count++
		}
	}
	return count, nil
}

func (r *Repository) GetTicket(_ context.Context, id int64) (*service.SupportTicket, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ticket, ok := r.state.tickets[id]
	if !ok {
		return nil, service.ErrTicketNotFound
	}
	ticket = cloneTicket(ticket)
	return &ticket, nil
}

func (r *Repository) LockTicket(ctx context.Context, id int64) (*service.SupportTicket, error) {
	return r.GetTicket(ctx, id)
}
func (r *Repository) TicketOwnerID(ctx context.Context, id int64) (int64, error) {
	t, err := r.GetTicket(ctx, id)
	if err != nil {
		return 0, err
	}
	return t.UserID, nil
}

func (r *Repository) InsertTicket(_ context.Context, t *service.SupportTicket) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.ticketID++
	t.ID = r.state.ticketID
	stored := cloneTicket(*t)
	stored.Messages = nil
	r.state.tickets[t.ID] = stored
	return nil
}

func (r *Repository) AddMessage(_ context.Context, m *service.SupportTicketMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.state.tickets[m.TicketID]; !ok {
		return service.ErrTicketNotFound
	}
	r.state.messageID++
	m.ID = r.state.messageID
	r.state.messages[m.TicketID] = append(r.state.messages[m.TicketID], *m)
	return nil
}

func (r *Repository) UpdateTicket(_ context.Context, t *service.SupportTicket) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.state.tickets[t.ID]; !ok {
		return service.ErrTicketNotFound
	}
	stored := cloneTicket(*t)
	stored.Messages = nil
	r.state.tickets[t.ID] = stored
	return nil
}

func (r *Repository) List(_ context.Context, params pagination.PaginationParams, filters service.SupportTicketListFilters) ([]service.SupportTicket, *pagination.PaginationResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]service.SupportTicket, 0)
	for _, ticket := range r.state.tickets {
		if filters.UserID != 0 && ticket.UserID != filters.UserID {
			continue
		}
		if filters.Category != "" && ticket.Category != filters.Category {
			continue
		}
		if len(filters.Statuses) > 0 {
			match := false
			for _, status := range filters.Statuses {
				if ticket.Status == status {
					match = true
				}
			}
			if !match {
				continue
			}
		}
		search := strings.ToLower(filters.Search)
		id, _ := strconv.ParseInt(strings.TrimPrefix(search, "#"), 10, 64)
		emailMatch := filters.UserID == 0 && strings.Contains(strings.ToLower(r.state.users[ticket.UserID]), search)
		if search != "" && !strings.Contains(strings.ToLower(ticket.Subject), search) && !emailMatch && ticket.ID != id {
			continue
		}
		items = append(items, cloneTicket(ticket))
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].LastMessageAt.Equal(items[j].LastMessageAt) {
			return items[i].ID > items[j].ID
		}
		return items[i].LastMessageAt.After(items[j].LastMessageAt)
	})
	total := len(items)
	start := params.Offset()
	if start > total {
		start = total
	}
	end := start + params.Limit()
	if end > total {
		end = total
	}
	pages := (total + params.Limit() - 1) / params.Limit()
	return items[start:end], &pagination.PaginationResult{Total: int64(total), Page: params.Page, PageSize: params.Limit(), Pages: pages}, nil
}

func (r *Repository) CountMessages(_ context.Context, id int64) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.state.messages[id]), nil
}

func (r *Repository) ListMessages(_ context.Context, id int64) ([]service.SupportTicketMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]service.SupportTicketMessage{}, r.state.messages[id]...), nil
}

func (r *Repository) EmailsByUserIDs(_ context.Context, ids []int64) (map[int64]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[int64]string)
	for _, id := range ids {
		out[id] = r.state.users[id]
	}
	return out, nil
}

var _ service.SupportTicketRepository = (*Repository)(nil)
