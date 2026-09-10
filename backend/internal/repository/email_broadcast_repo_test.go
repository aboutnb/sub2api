package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestEmailBroadcastCountEligibleRecipientsFiltersInactiveAndDeletedUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users u WHERE u.status = 'active' AND u.deleted_at IS NULL AND BTRIM\(u.email\) <> ''`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(307)))

	repo := NewEmailBroadcastRepository(db)
	count, err := repo.CountEligibleRecipients(context.Background(), service.EmailBroadcastAudience{Mode: service.EmailBroadcastAudienceAll})
	require.NoError(t, err)
	require.Equal(t, int64(307), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEmailBroadcastAudienceWhere(t *testing.T) {
	tests := []struct {
		name     string
		audience service.EmailBroadcastAudience
		contains string
		arg      any
	}{
		{name: "role", audience: service.EmailBroadcastAudience{Mode: service.EmailBroadcastAudienceRole, Roles: []string{"user"}}, contains: "u.role = ANY($3)"},
		{name: "groups", audience: service.EmailBroadcastAudience{Mode: service.EmailBroadcastAudienceGroups, GroupIDs: []int64{9}}, contains: "uag.group_id = ANY($3)"},
		{name: "selected", audience: service.EmailBroadcastAudience{Mode: service.EmailBroadcastAudienceSelected, Emails: []string{"user@example.com"}}, contains: "LOWER(BTRIM(u.email)) = ANY($3)"},
		{name: "inactive", audience: service.EmailBroadcastAudience{Mode: service.EmailBroadcastAudienceInactive, InactiveDays: 7}, contains: "COALESCE(u.last_active_at, u.last_login_at, u.created_at) <= NOW() - make_interval(days => $3::int)", arg: 7},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			where, args := emailBroadcastAudienceWhere(test.audience, 3)
			require.True(t, strings.Contains(where, test.contains))
			require.Len(t, args, 1)
			if test.arg != nil {
				require.Equal(t, test.arg, args[0])
			}
		})
	}
}
