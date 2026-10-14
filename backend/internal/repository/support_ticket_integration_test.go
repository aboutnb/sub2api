//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSupportTicketPostgresMessageCapAndCloseRace(t *testing.T) {
	ctx := context.Background()
	userID := ticketIntegrationUser(t, "ticket-message-cap")
	s := service.NewSupportTicketService(NewSupportTicketRepository(integrationEntClient))
	ticket, err := s.Create(ctx, userID, service.CreateTicketInput{Subject: "Cap", Body: "First", Category: "other"})
	require.NoError(t, err)
	_, err = integrationDB.ExecContext(ctx, `INSERT INTO support_ticket_messages(ticket_id,author_id,author_role,body,created_at)
		SELECT $1,$2,'user','seed',NOW() FROM generate_series(1,198)`, ticket.ID, userID)
	require.NoError(t, err)
	results := make(chan error, 16)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := s.ReplyAsUser(ctx, userID, ticket.ID, "last"); results <- err }()
	}
	wg.Wait()
	close(results)
	succeeded := 0
	for err := range results {
		if err == nil {
			succeeded++
		} else {
			require.ErrorIs(t, err, service.ErrTicketMessageLimit)
		}
	}
	require.Equal(t, 1, succeeded)
	var count int
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT count(*) FROM support_ticket_messages WHERE ticket_id=$1", ticket.ID).Scan(&count))
	require.Equal(t, service.TicketMessageMax, count)
	for i := 0; i < 5; i++ {
		ticket, err := s.Create(ctx, userID, service.CreateTicketInput{Subject: "Close race", Body: "First", Category: "other"})
		require.NoError(t, err)
		errs := make(chan error, 2)
		go func() { _, err := s.CloseForUser(ctx, userID, ticket.ID); errs <- err }()
		go func() { _, err := s.ReplyAsUser(ctx, userID, ticket.ID, "racing reply"); errs <- err }()
		for j := 0; j < 2; j++ {
			err := <-errs
			if err != nil {
				require.ErrorIs(t, err, service.ErrTicketClosed)
			}
		}
		final, err := s.GetForUser(ctx, userID, ticket.ID)
		require.NoError(t, err)
		require.Equal(t, service.TicketStatusClosed, final.Status)
		require.NotNil(t, final.ClosedAt)
	}
}

func ticketIntegrationUser(t *testing.T, emailPrefix string) int64 {
	t.Helper()
	user, err := integrationEntClient.User.Create().SetEmail(fmt.Sprintf("%s-%d@example.test", emailPrefix, time.Now().UnixNano())).SetPasswordHash("test").SetRole(service.RoleUser).SetStatus(service.StatusActive).Save(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx := context.Background()
		_, err := integrationDB.ExecContext(ctx, "DELETE FROM support_ticket_messages WHERE author_id=$1 OR ticket_id IN (SELECT id FROM support_tickets WHERE user_id=$1)", user.ID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(ctx, "DELETE FROM support_tickets WHERE user_id=$1", user.ID)
		require.NoError(t, err)
		_, err = integrationDB.ExecContext(ctx, "DELETE FROM users WHERE id=$1", user.ID)
		require.NoError(t, err)
	})
	return user.ID
}

func TestSupportTicketPostgresConcurrencySearchAndSoftDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	userID := ticketIntegrationUser(t, `ticket_%`)
	adminID := ticketIntegrationUser(t, "ticket-admin")
	r := NewSupportTicketRepository(integrationEntClient)
	s := service.NewSupportTicketService(r)
	results := make(chan error, 16)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Create(ctx, userID, service.CreateTicketInput{Subject: `literal_%\`, Body: "Initial message", Category: "api_models"})
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	succeeded := 0
	for err := range results {
		if err == nil {
			succeeded++
		} else {
			require.ErrorIs(t, err, service.ErrTicketOpenLimit)
		}
	}
	require.Equal(t, 10, succeeded)
	items, result, err := s.ListForAdmin(ctx, pagination.PaginationParams{Page: 2, PageSize: 3}, "api_models", "", `literal_%\`)
	require.NoError(t, err)
	require.EqualValues(t, 10, result.Total)
	require.Len(t, items, 3)
	_, err = integrationDB.ExecContext(ctx, "UPDATE users SET deleted_at=NOW() WHERE id=$1", userID)
	require.NoError(t, err)
	items, result, err = s.ListForAdmin(ctx, pagination.DefaultPagination(), "", "", "ticket_%")
	require.NoError(t, err)
	require.EqualValues(t, 10, result.Total)
	require.Contains(t, items[0].UserEmail, "ticket_%")
	_, err = s.Create(ctx, userID, service.CreateTicketInput{Subject: "No new ticket", Body: "Body", Category: "other"})
	require.ErrorIs(t, err, service.ErrUserNotFound)
	closed, err := s.CloseForUser(ctx, userID, items[0].ID)
	require.NoError(t, err)
	require.NotNil(t, closed.ClosedAt)
	reopened, err := s.ReplyAsAdmin(ctx, adminID, closed.ID, "Reply\nfrom admin")
	require.NoError(t, err)
	require.Nil(t, reopened.ClosedAt)
	require.Equal(t, "pending_user", reopened.Status)
	require.Len(t, reopened.Messages, 2)
	require.Equal(t, "Reply from admin", reopened.Preview)
}

func TestSupportTicketPostgresOuterRollback(t *testing.T) {
	ctx := context.Background()
	userID := ticketIntegrationUser(t, "ticket-rollback")
	r := NewSupportTicketRepository(integrationEntClient)
	s := service.NewSupportTicketService(r)
	tx, err := integrationEntClient.Tx(ctx)
	require.NoError(t, err)
	ticket, err := s.Create(dbent.NewTxContext(ctx, tx), userID, service.CreateTicketInput{Subject: "Nested", Body: "Body", Category: "other"})
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
	_, err = r.GetTicket(ctx, ticket.ID)
	require.ErrorIs(t, err, service.ErrTicketNotFound)
	messages, err := r.ListMessages(ctx, ticket.ID)
	require.NoError(t, err)
	require.Empty(t, messages)
	outerErr := errors.New("rollback")
	err = r.WithTx(ctx, func(txCtx context.Context) error {
		_, err := s.Create(txCtx, userID, service.CreateTicketInput{Subject: "Nested callback", Body: "Body", Category: "other"})
		if err != nil {
			return err
		}
		return outerErr
	})
	require.ErrorIs(t, err, outerErr)
	count, err := r.CountOpenByUser(ctx, userID)
	require.NoError(t, err)
	require.Zero(t, count)
}

func TestSupportTicketPostgresUserTabsAndSearch(t *testing.T) {
	ctx := context.Background()
	userID := ticketIntegrationUser(t, "ticket-user-search")
	otherID := ticketIntegrationUser(t, "ticket-foreign-search")
	s := service.NewSupportTicketService(NewSupportTicketRepository(integrationEntClient))
	first, err := s.Create(ctx, userID, service.CreateTicketInput{Subject: `literal_%\`, Body: "Details", Category: "other"})
	require.NoError(t, err)
	second, err := s.Create(ctx, userID, service.CreateTicketInput{Subject: "Model issue", Body: "Details", Category: "api_models"})
	require.NoError(t, err)
	_, err = s.ReplyAsAdmin(ctx, otherID, second.ID, "Please confirm")
	require.NoError(t, err)
	third, err := s.Create(ctx, userID, service.CreateTicketInput{Subject: "Resolved", Body: "Details", Category: "billing"})
	require.NoError(t, err)
	_, err = s.CloseForUser(ctx, userID, third.ID)
	require.NoError(t, err)
	items, result, err := s.ListForUser(ctx, userID, pagination.PaginationParams{Page: 2, PageSize: 1}, "", "unclosed", "")
	require.NoError(t, err)
	require.EqualValues(t, 2, result.Total)
	require.Len(t, items, 1)
	require.Equal(t, first.ID, items[0].ID)
	items, result, err = s.ListForUser(ctx, userID, pagination.DefaultPagination(), "", "closed", "")
	require.NoError(t, err)
	require.EqualValues(t, 1, result.Total)
	require.Equal(t, third.ID, items[0].ID)
	for _, search := range []string{fmt.Sprint(first.ID), fmt.Sprintf("#%d", first.ID), `_%\`} {
		items, result, err = s.ListForUser(ctx, userID, pagination.DefaultPagination(), "", "", search)
		require.NoError(t, err)
		require.EqualValues(t, 1, result.Total)
		require.Equal(t, first.ID, items[0].ID)
	}
	items, _, err = s.ListForUser(ctx, otherID, pagination.DefaultPagination(), "", "", fmt.Sprintf("#%d", first.ID))
	require.NoError(t, err)
	require.Empty(t, items)
	items, _, err = s.ListForUser(ctx, userID, pagination.DefaultPagination(), "", "", "ticket-user-search")
	require.NoError(t, err)
	require.Empty(t, items, "user search matches subjects and IDs, not email")
}
