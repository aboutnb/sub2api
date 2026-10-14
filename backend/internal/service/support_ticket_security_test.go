package service_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSupportTicketRejectsInvalidTextAndQuery(t *testing.T) {
	_, s := supportTicketSetup()
	ctx := context.Background()
	for _, input := range []service.CreateTicketInput{
		{Subject: "bad\x00title", Body: "ok", Category: "other"},
		{Subject: "ok", Body: "bad\x00body", Category: "other"},
		{Subject: "\xff", Body: "ok", Category: "other"},
	} {
		_, err := s.Create(ctx, 1, input)
		require.ErrorIs(t, err, service.ErrTicketInputInvalid)
	}
	for _, search := range []string{"\x00", strings.Repeat("界", 201), "\xff"} {
		_, _, err := s.ListForUser(ctx, 1, pagination.DefaultPagination(), "", "", search)
		require.ErrorIs(t, err, service.ErrTicketQueryInvalid)
		_, _, err = s.ListForAdmin(ctx, pagination.DefaultPagination(), "", "", search)
		require.ErrorIs(t, err, service.ErrTicketQueryInvalid)
	}
	_, _, err := s.ListForUser(ctx, 1, pagination.PaginationParams{Page: int(^uint(0) >> 1), PageSize: 100}, "", "", "")
	require.ErrorIs(t, err, service.ErrTicketQueryInvalid)
}

func TestSupportTicketMessageCapIsAtomic(t *testing.T) {
	r, s := supportTicketSetup()
	ticket := createSupportTicket(t, s)
	ctx := context.Background()
	for i := 1; i < service.TicketMessageMax-1; i++ {
		require.NoError(t, r.AddMessage(ctx, &service.SupportTicketMessage{TicketID: ticket.ID, AuthorID: 1, AuthorRole: "user", Body: "old"}))
	}
	results := make(chan error, 16)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := s.ReplyAsUser(ctx, 1, ticket.ID, "last"); results <- err }()
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else {
			require.ErrorIs(t, err, service.ErrTicketMessageLimit)
		}
	}
	require.Equal(t, 1, success)
	count, err := r.CountMessages(ctx, ticket.ID)
	require.NoError(t, err)
	require.Equal(t, service.TicketMessageMax, count)
	_, err = s.ReplyAsAdmin(ctx, 9, ticket.ID, "admin cannot bypass cap")
	require.ErrorIs(t, err, service.ErrTicketMessageLimit)
	_, err = s.CloseForUser(ctx, 1, ticket.ID)
	require.NoError(t, err)
}
