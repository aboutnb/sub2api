package service_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/service/tickettest"
	"github.com/stretchr/testify/require"
)

func supportTicketSetup() (*tickettest.Repository, *service.SupportTicketService) {
	r := tickettest.New()
	r.AddUser(1, "owner@example.test")
	r.AddUser(2, "other@example.test")
	r.AddUser(9, "admin@example.test")
	return r, service.NewSupportTicketService(r)
}

func createSupportTicket(t *testing.T, s *service.SupportTicketService) *service.SupportTicket {
	t.Helper()
	ticket, err := s.Create(context.Background(), 1, service.CreateTicketInput{Subject: "Question", Body: "First message", Category: "billing"})
	require.NoError(t, err)
	return ticket
}

func TestSupportTicketCreatePreviewAndCopies(t *testing.T) {
	r, s := supportTicketSetup()
	body := "  第一行\n\t第二行  " + strings.Repeat("文", 100) + "  "
	ticket, err := s.Create(context.Background(), 1, service.CreateTicketInput{Subject: "  Help  ", Body: body, Category: " api_models "})
	require.NoError(t, err)
	require.Equal(t, "Help", ticket.Subject)
	require.Equal(t, "api_models", ticket.Category)
	require.Equal(t, "open", ticket.Status)
	require.Nil(t, ticket.ClosedAt)
	require.Len(t, []rune(ticket.Preview), 80)
	require.True(t, strings.HasPrefix(ticket.Preview, "第一行 第二行 "))
	require.Equal(t, strings.TrimSpace(body), ticket.Messages[0].Body)
	require.Equal(t, "user", ticket.Messages[0].AuthorRole)
	ticket.Subject = "mutated"
	ticket.Messages[0].Body = "mutated"
	stored, err := s.GetForUser(context.Background(), 1, ticket.ID)
	require.NoError(t, err)
	require.Equal(t, "Help", stored.Subject)
	require.Equal(t, strings.TrimSpace(body), stored.Messages[0].Body)
	stored.Subject = "replacement"
	require.NoError(t, r.UpdateTicket(context.Background(), stored))
	stored.Subject = "mutated after save"
	stored, err = r.GetTicket(context.Background(), stored.ID)
	require.NoError(t, err)
	require.Equal(t, "replacement", stored.Subject)
	require.Nil(t, stored.Messages)
}

func TestSupportTicketLimitsCloseAndAdminReopen(t *testing.T) {
	r, s := supportTicketSetup()
	ctx := context.Background()
	for i := 0; i < 9; i++ {
		createSupportTicket(t, s)
	}
	closed := createSupportTicket(t, s)
	_, err := s.CloseForUser(ctx, 1, closed.ID)
	require.NoError(t, err)
	createSupportTicket(t, s)
	_, err = s.Create(ctx, 1, service.CreateTicketInput{Subject: "extra", Body: "extra", Category: "other"})
	require.ErrorIs(t, err, service.ErrTicketOpenLimit)
	require.Equal(t, "10", infraerrors.FromError(err).Metadata["max"])
	_, err = s.UpdateStatus(ctx, closed.ID, "open")
	require.NoError(t, err)
	count, err := r.CountOpenByUser(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, 11, count)
	_, err = s.UpdateStatus(ctx, closed.ID, "closed")
	require.NoError(t, err)
	ticket, err := s.ReplyAsAdmin(ctx, 9, closed.ID, "We can help")
	require.NoError(t, err)
	require.Equal(t, "pending_user", ticket.Status)
	require.Nil(t, ticket.ClosedAt)
}

func TestSupportTicketIsolationAndTransitions(t *testing.T) {
	_, s := supportTicketSetup()
	ctx := context.Background()
	ticket := createSupportTicket(t, s)
	_, err := s.GetForUser(ctx, 2, ticket.ID)
	require.ErrorIs(t, err, service.ErrTicketNotFound)
	_, err = s.ReplyAsUser(ctx, 2, ticket.ID, "reply")
	require.ErrorIs(t, err, service.ErrTicketNotFound)
	_, err = s.CloseForUser(ctx, 2, ticket.ID)
	require.ErrorIs(t, err, service.ErrTicketNotFound)
	replied, err := s.ReplyAsUser(ctx, 1, ticket.ID, "  user\n reply  ")
	require.NoError(t, err)
	require.Equal(t, "pending_admin", replied.Status)
	require.Equal(t, "user reply", replied.Preview)
	require.True(t, replied.LastMessageAt.After(ticket.LastMessageAt))
	closed, err := s.CloseForUser(ctx, 1, ticket.ID)
	require.NoError(t, err)
	require.NotNil(t, closed.ClosedAt)
	require.Equal(t, replied.LastMessageAt, closed.LastMessageAt)
	require.Equal(t, replied.Preview, closed.Preview)
	_, err = s.ReplyAsUser(ctx, 2, ticket.ID, "reply")
	require.ErrorIs(t, err, service.ErrTicketNotFound)
	_, err = s.ReplyAsUser(ctx, 1, ticket.ID, "reply")
	require.ErrorIs(t, err, service.ErrTicketClosed)
	for i := 0; i < 2; i++ {
		again, err := s.CloseForUser(ctx, 1, ticket.ID)
		require.NoError(t, err)
		require.Equal(t, closed.ClosedAt, again.ClosedAt)
		require.Equal(t, closed.UpdatedAt, again.UpdatedAt)
		again, err = s.UpdateStatus(ctx, ticket.ID, "closed")
		require.NoError(t, err)
		require.Equal(t, closed.ClosedAt, again.ClosedAt)
		require.Equal(t, closed.UpdatedAt, again.UpdatedAt)
	}
	reopened, err := s.ReplyAsAdmin(ctx, 9, ticket.ID, "Admin reply")
	require.NoError(t, err)
	require.Equal(t, "pending_user", reopened.Status)
	require.Nil(t, reopened.ClosedAt)
	require.Len(t, reopened.Messages, 3)
	require.Equal(t, "admin", reopened.Messages[2].AuthorRole)
	open, err := s.UpdateStatus(ctx, ticket.ID, " open ")
	require.NoError(t, err)
	require.Nil(t, open.ClosedAt)
	require.Equal(t, reopened.LastMessageAt, open.LastMessageAt)
	unchanged, err := s.UpdateStatus(ctx, ticket.ID, "open")
	require.NoError(t, err)
	require.Equal(t, open.UpdatedAt, unchanged.UpdatedAt)
}

func TestSupportTicketValidationOrder(t *testing.T) {
	_, s := supportTicketSetup()
	for _, tc := range []struct {
		subject, body, category string
		err                     error
	}{
		{" ", " ", "wrong", service.ErrTicketInputRequired},
		{strings.Repeat("文", 81), " ", "wrong", service.ErrTicketSubjectInvalid},
		{"Subject", " ", "wrong", service.ErrTicketInputRequired},
		{"Subject", strings.Repeat("文", 4001), "wrong", service.ErrTicketBodyInvalid},
		{"Subject", "Body", "", service.ErrTicketCategoryInvalid},
		{"Subject", "Body", "all", service.ErrTicketCategoryInvalid},
		{"Subject", "Body", "Billing", service.ErrTicketCategoryInvalid},
		{"Subject", "Body", "missing", service.ErrTicketCategoryInvalid},
	} {
		_, err := s.Create(context.Background(), 1, service.CreateTicketInput{Subject: tc.subject, Body: tc.body, Category: tc.category})
		require.ErrorIs(t, err, tc.err)
	}
	_, err := s.Create(context.Background(), 1, service.CreateTicketInput{Subject: strings.Repeat("文", 80), Body: strings.Repeat("文", 4000), Category: "other"})
	require.NoError(t, err)
	for _, status := range []string{"all", "OPEN", "bad"} {
		_, _, err := s.ListForUser(context.Background(), 1, pagination.DefaultPagination(), "", status, "")
		require.ErrorIs(t, err, service.ErrTicketStatusInvalid)
	}
	for _, queue := range []string{"all", "open", "pending_admin", "UNANSWERED"} {
		_, _, err := s.ListForAdmin(context.Background(), pagination.DefaultPagination(), "", queue, "")
		require.ErrorIs(t, err, service.ErrTicketStatusInvalid)
	}
	_, err = s.ReplyAsUser(context.Background(), 1, 1, " ")
	require.ErrorIs(t, err, service.ErrTicketInputRequired)
	_, err = s.ReplyAsAdmin(context.Background(), 9, 1, strings.Repeat("文", 4001))
	require.ErrorIs(t, err, service.ErrTicketBodyInvalid)
}

func TestSupportTicketQueuesSearchAndPagination(t *testing.T) {
	r, s := supportTicketSetup()
	ctx := context.Background()
	open := createSupportTicket(t, s)
	pending := createSupportTicket(t, s)
	_, err := s.ReplyAsUser(ctx, 1, pending.ID, "More info")
	require.NoError(t, err)
	awaiting := createSupportTicket(t, s)
	_, err = s.ReplyAsAdmin(ctx, 9, awaiting.ID, "Please confirm")
	require.NoError(t, err)
	closed := createSupportTicket(t, s)
	_, err = s.CloseForUser(ctx, 1, closed.ID)
	require.NoError(t, err)
	for _, queue := range []string{"", "unanswered"} {
		items, result, err := s.ListForAdmin(ctx, pagination.DefaultPagination(), "all", queue, "")
		require.NoError(t, err)
		require.EqualValues(t, 2, result.Total)
		require.Equal(t, []int64{pending.ID, open.ID}, []int64{items[0].ID, items[1].ID})
		require.Equal(t, "owner@example.test", items[0].UserEmail)
		require.Nil(t, items[0].Messages)
	}
	for _, queue := range []string{"awaiting_user", "closed"} {
		items, result, err := s.ListForAdmin(ctx, pagination.DefaultPagination(), "billing", queue, "")
		require.NoError(t, err)
		require.EqualValues(t, 1, result.Total)
		require.Len(t, items, 1)
	}
	items, result, err := s.ListForAdmin(ctx, pagination.PaginationParams{Page: 2, PageSize: 1}, "", "", "OWNER@EXAMPLE")
	require.NoError(t, err)
	require.EqualValues(t, 2, result.Total)
	require.Equal(t, open.ID, items[0].ID)
	r.AddUser(1, `literal_%\@example.test`)
	items, _, err = s.ListForAdmin(ctx, pagination.DefaultPagination(), "", "", `_%\`)
	require.NoError(t, err)
	require.Len(t, items, 2)
	items, _, err = s.ListForUser(ctx, 2, pagination.DefaultPagination(), "", "", "")
	require.NoError(t, err)
	require.Empty(t, items)
	items, result, err = s.ListForUser(ctx, 1, pagination.PaginationParams{Page: 2, PageSize: 2}, "", "unclosed", "")
	require.NoError(t, err)
	require.EqualValues(t, 3, result.Total)
	require.Len(t, items, 1)
	items, result, err = s.ListForUser(ctx, 1, pagination.DefaultPagination(), "", "closed", "")
	require.NoError(t, err)
	require.EqualValues(t, 1, result.Total)
	require.Equal(t, closed.ID, items[0].ID)
	items, _, err = s.ListForUser(ctx, 1, pagination.DefaultPagination(), "", "", fmt.Sprintf("#%d", open.ID))
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, open.ID, items[0].ID)
	items, _, err = s.ListForUser(ctx, 2, pagination.DefaultPagination(), "", "", fmt.Sprintf("#%d", open.ID))
	require.NoError(t, err)
	require.Empty(t, items)
	items, _, err = s.ListForUser(ctx, 1, pagination.DefaultPagination(), "", "", "literal_%")
	require.NoError(t, err)
	require.Empty(t, items, "user search must not match the owner's email")
	_, _, err = s.ListForAdmin(ctx, pagination.DefaultPagination(), "bad", "", "")
	require.ErrorIs(t, err, service.ErrTicketCategoryInvalid)
}

type failingTicketUpdate struct {
	service.SupportTicketRepository
}

func (r failingTicketUpdate) UpdateTicket(context.Context, *service.SupportTicket) error {
	return errors.New("update failed")
}

func TestSupportTicketRollbackRestoresMessagesAndCounters(t *testing.T) {
	r, s := supportTicketSetup()
	ctx := context.Background()
	broken := service.NewSupportTicketService(failingTicketUpdate{r})
	_, err := broken.Create(ctx, 1, service.CreateTicketInput{Subject: "Question", Body: "Body", Category: "other"})
	require.Error(t, err)
	ticket := createSupportTicket(t, s)
	require.EqualValues(t, 1, ticket.ID)
	require.EqualValues(t, 1, ticket.Messages[0].ID)
	closed, err := s.CloseForUser(ctx, 1, ticket.ID)
	require.NoError(t, err)
	_, err = broken.ReplyAsAdmin(ctx, 9, ticket.ID, "Uncommitted")
	require.Error(t, err)
	stored, err := s.GetForAdmin(ctx, ticket.ID)
	require.NoError(t, err)
	require.Equal(t, "closed", stored.Status)
	require.Equal(t, closed.ClosedAt, stored.ClosedAt)
	require.Equal(t, closed.UpdatedAt, stored.UpdatedAt)
	require.Len(t, stored.Messages, 1)
	replied, err := s.ReplyAsAdmin(ctx, 9, ticket.ID, "Committed")
	require.NoError(t, err)
	require.EqualValues(t, 2, replied.Messages[1].ID)
}

func TestSupportTicketTransactionNoDeadlockAndNestedRollback(t *testing.T) {
	r, s := supportTicketSetup()
	ticket := createSupportTicket(t, s)
	done := make(chan error, 1)
	go func() {
		done <- r.WithTx(context.Background(), func(ctx context.Context) error { _, err := r.LockTicket(ctx, ticket.ID); return err })
	}()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("transaction callback deadlocked")
	}
	outerErr := errors.New("outer failed")
	err := r.WithTx(context.Background(), func(ctx context.Context) error {
		_, err := s.Create(ctx, 1, service.CreateTicketInput{Subject: "Nested", Body: "Body", Category: "other"})
		require.NoError(t, err)
		return outerErr
	})
	require.ErrorIs(t, err, outerErr)
	count, err := r.CountOpenByUser(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	err = r.WithTx(context.Background(), func(ctx context.Context) error {
		_, err := s.Create(ctx, 1, service.CreateTicketInput{Subject: "Keep", Body: "Body", Category: "other"})
		require.NoError(t, err)
		_, err = s.Create(ctx, 123, service.CreateTicketInput{Subject: "Missing user", Body: "Body", Category: "other"})
		require.ErrorIs(t, err, service.ErrUserNotFound)
		return nil
	})
	require.NoError(t, err)
	count, err = r.CountOpenByUser(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

func TestSupportTicketConcurrentCreationLimit(t *testing.T) {
	r, s := supportTicketSetup()
	var wg sync.WaitGroup
	errorsCh := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Create(context.Background(), 1, service.CreateTicketInput{Subject: "Concurrent", Body: "Body", Category: "other"})
			errorsCh <- err
		}()
	}
	wg.Wait()
	close(errorsCh)
	succeeded := 0
	for err := range errorsCh {
		if err == nil {
			succeeded++
		} else {
			require.ErrorIs(t, err, service.ErrTicketOpenLimit)
		}
	}
	require.Equal(t, 10, succeeded)
	count, err := r.CountOpenByUser(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 10, count)
}
