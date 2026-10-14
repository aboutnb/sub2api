package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type supportTicketRepository struct{ client *dbent.Client }

func NewSupportTicketRepository(client *dbent.Client) service.SupportTicketRepository {
	return &supportTicketRepository{client: client}
}

func (r *supportTicketRepository) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if dbent.TxFromContext(ctx) != nil {
		return fn(ctx)
	}
	tx, err := r.client.Tx(ctx)
	if errors.Is(err, dbent.ErrTxStarted) {
		return fn(ctx)
	}
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(dbent.NewTxContext(ctx, tx)); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *supportTicketRepository) LockUser(ctx context.Context, id int64) error {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, "SELECT id FROM users WHERE id=$1 AND deleted_at IS NULL FOR UPDATE", id)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return service.ErrUserNotFound
	}
	var userID int64
	if err := rows.Scan(&userID); err != nil {
		return err
	}
	return rows.Err()
}

func (r *supportTicketRepository) CountOpenByUser(ctx context.Context, userID int64) (int, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, "SELECT COUNT(*) FROM support_tickets WHERE user_id=$1 AND status<>$2", userID, service.TicketStatusClosed)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return 0, rows.Err()
	}
	var count int
	if err := rows.Scan(&count); err != nil {
		return 0, err
	}
	return count, rows.Err()
}

const ticketSelectColumns = "t.id, t.user_id, t.subject, t.category, t.status, t.preview, t.last_message_at, t.closed_at, t.created_at, t.updated_at"

func scanSupportTicket(rows *sql.Rows) (*service.SupportTicket, error) {
	var ticket service.SupportTicket
	if err := rows.Scan(&ticket.ID, &ticket.UserID, &ticket.Subject, &ticket.Category, &ticket.Status, &ticket.Preview, &ticket.LastMessageAt, &ticket.ClosedAt, &ticket.CreatedAt, &ticket.UpdatedAt); err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *supportTicketRepository) get(ctx context.Context, id int64, lock bool) (*service.SupportTicket, error) {
	query := "SELECT " + ticketSelectColumns + " FROM support_tickets t WHERE t.id=$1"
	if lock {
		query += " FOR UPDATE"
	}
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrTicketNotFound
	}
	ticket, err := scanSupportTicket(rows)
	if err != nil {
		return nil, err
	}
	return ticket, rows.Err()
}

func (r *supportTicketRepository) GetTicket(ctx context.Context, id int64) (*service.SupportTicket, error) {
	return r.get(ctx, id, false)
}
func (r *supportTicketRepository) LockTicket(ctx context.Context, id int64) (*service.SupportTicket, error) {
	return r.get(ctx, id, true)
}

func (r *supportTicketRepository) TicketOwnerID(ctx context.Context, id int64) (int64, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, "SELECT user_id FROM support_tickets WHERE id=$1", id)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, err
		}
		return 0, service.ErrTicketNotFound
	}
	var ownerID int64
	if err := rows.Scan(&ownerID); err != nil {
		return 0, err
	}
	return ownerID, rows.Err()
}

func (r *supportTicketRepository) InsertTicket(ctx context.Context, t *service.SupportTicket) error {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `INSERT INTO support_tickets
	(user_id, subject, category, status, preview, last_message_at, closed_at, created_at, updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, t.UserID, t.Subject, t.Category, t.Status, t.Preview, t.LastMessageAt, t.ClosedAt, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	if err := rows.Scan(&t.ID); err != nil {
		return err
	}
	return rows.Err()
}

func (r *supportTicketRepository) AddMessage(ctx context.Context, m *service.SupportTicketMessage) error {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, `INSERT INTO support_ticket_messages
	(ticket_id, author_id, author_role, body, created_at) VALUES ($1,$2,$3,$4,$5) RETURNING id`, m.TicketID, m.AuthorID, m.AuthorRole, m.Body, m.CreatedAt)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	if err := rows.Scan(&m.ID); err != nil {
		return err
	}
	return rows.Err()
}

func (r *supportTicketRepository) UpdateTicket(ctx context.Context, t *service.SupportTicket) error {
	result, err := clientFromContext(ctx, r.client).ExecContext(ctx, `UPDATE support_tickets
	SET status=$2, preview=$3, last_message_at=$4, closed_at=$5, updated_at=$6 WHERE id=$1`, t.ID, t.Status, t.Preview, t.LastMessageAt, t.ClosedAt, t.UpdatedAt)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return service.ErrTicketNotFound
	}
	return nil
}

func ticketListWhere(filters service.SupportTicketListFilters) (string, []any) {
	clauses := []string{"TRUE"}
	args := []any{}
	arg := func(value any) string { args = append(args, value); return fmt.Sprintf("$%d", len(args)) }
	if filters.UserID != 0 {
		clauses = append(clauses, "t.user_id="+arg(filters.UserID))
	}
	if filters.Category != "" {
		clauses = append(clauses, "t.category="+arg(filters.Category))
	}
	if len(filters.Statuses) > 0 {
		placeholders := make([]string, 0, len(filters.Statuses))
		for _, status := range filters.Statuses {
			placeholders = append(placeholders, arg(status))
		}
		clauses = append(clauses, "t.status IN ("+strings.Join(placeholders, ",")+")")
	}
	if filters.Search != "" {
		placeholder := arg("%" + escapeLikeWildcards(filters.Search) + "%")
		searchClauses := []string{"t.subject ILIKE " + placeholder + ` ESCAPE '\'`}
		if filters.UserID == 0 {
			searchClauses = append(searchClauses, "COALESCE(u.email, '') ILIKE "+placeholder+` ESCAPE '\'`)
		}
		if id, err := strconv.ParseInt(strings.TrimPrefix(filters.Search, "#"), 10, 64); err == nil && id > 0 {
			searchClauses = append(searchClauses, "t.id="+arg(id))
		}
		clauses = append(clauses, "("+strings.Join(searchClauses, " OR ")+")")
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *supportTicketRepository) List(ctx context.Context, params pagination.PaginationParams, filters service.SupportTicketListFilters) ([]service.SupportTicket, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	where, args := ticketListWhere(filters)
	from := " FROM support_tickets t LEFT JOIN users u ON u.id=t.user_id"
	rows, err := client.QueryContext(ctx, "SELECT COUNT(*)"+from+where, args...)
	if err != nil {
		return nil, nil, err
	}
	var total int64
	if rows.Next() {
		err = rows.Scan(&total)
	} else {
		err = rows.Err()
	}
	closeErr := rows.Close()
	if err != nil {
		return nil, nil, err
	}
	if closeErr != nil {
		return nil, nil, closeErr
	}
	query := "SELECT " + ticketSelectColumns + from + where + fmt.Sprintf(" ORDER BY t.last_message_at DESC, t.id DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, params.Limit(), params.Offset())
	rows, err = client.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.SupportTicket, 0)
	for rows.Next() {
		ticket, err := scanSupportTicket(rows)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, *ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *supportTicketRepository) CountMessages(ctx context.Context, id int64) (int, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, "SELECT COUNT(*) FROM support_ticket_messages WHERE ticket_id=$1", id)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	var count int
	if !rows.Next() {
		return 0, rows.Err()
	}
	if err := rows.Scan(&count); err != nil {
		return 0, err
	}
	return count, rows.Err()
}

func (r *supportTicketRepository) ListMessages(ctx context.Context, id int64) ([]service.SupportTicketMessage, error) {
	rows, err := clientFromContext(ctx, r.client).QueryContext(ctx, "SELECT id, ticket_id, author_id, author_role, body, created_at FROM support_ticket_messages WHERE ticket_id=$1 ORDER BY id ASC", id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.SupportTicketMessage, 0)
	for rows.Next() {
		var m service.SupportTicketMessage
		if err := rows.Scan(&m.ID, &m.TicketID, &m.AuthorID, &m.AuthorRole, &m.Body, &m.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (r *supportTicketRepository) EmailsByUserIDs(ctx context.Context, ids []int64) (map[int64]string, error) {
	emails := make(map[int64]string)
	if len(ids) == 0 {
		return emails, nil
	}
	users, err := clientFromContext(ctx, r.client).User.Query().Where(dbuser.IDIn(ids...)).All(mixins.SkipSoftDelete(ctx))
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		emails[user.ID] = user.Email
	}
	return emails, nil
}
