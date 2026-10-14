package repository

import (
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSupportTicketSearchEscapesWildcards(t *testing.T) {
	query, args := ticketListWhere(service.SupportTicketListFilters{Search: `owner_%\`, Category: "billing", Statuses: []string{"open", "pending_admin"}})
	require.Equal(t, 2, strings.Count(query, `ESCAPE '\'`))
	require.Contains(t, query, "COALESCE(u.email, '') ILIKE $4")
	require.Equal(t, []any{"billing", "open", "pending_admin", `%owner\_\%\\%`}, args)
}

func TestSupportTicketUserSearchScopesIDAndSubject(t *testing.T) {
	query, args := ticketListWhere(service.SupportTicketListFilters{UserID: 7, Search: "#42"})
	require.Contains(t, query, "t.user_id=$1")
	require.Contains(t, query, "t.subject ILIKE $2")
	require.Contains(t, query, "OR t.id=$3")
	require.NotContains(t, query, "u.email")
	require.Equal(t, []any{int64(7), "%#42%", int64(42)}, args)
}
