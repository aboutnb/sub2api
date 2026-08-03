//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestEmailBroadcastRepositoryFinishRecipientPersistsDeliveryState(t *testing.T) {
	tests := []struct {
		name              string
		finish            func(*emailBroadcastRepository, context.Context, int64) error
		recipientStatus   string
		taskStatus        string
		sentCount         int64
		failedCount       int64
		expectSentAt      bool
		expectedLastError string
	}{
		{
			name: "sent",
			finish: func(repo *emailBroadcastRepository, ctx context.Context, recipientID int64) error {
				return repo.CompleteRecipient(ctx, recipientID)
			},
			recipientStatus: service.EmailBroadcastRecipientSent,
			taskStatus:      service.EmailBroadcastStatusSucceeded,
			sentCount:       1,
			expectSentAt:    true,
		},
		{
			name: "failed",
			finish: func(repo *emailBroadcastRepository, ctx context.Context, recipientID int64) error {
				return repo.FailRecipient(ctx, recipientID, "smtp rejected")
			},
			recipientStatus:   service.EmailBroadcastRecipientFailed,
			taskStatus:        service.EmailBroadcastStatusPartiallyFailed,
			failedCount:       1,
			expectedLastError: "smtp rejected",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			user := mustCreateUser(t, integrationEntClient, &service.User{
				Email:        fmt.Sprintf("email-broadcast-finish-%s-%d@example.com", test.name, time.Now().UnixNano()),
				PasswordHash: "hash",
				Role:         service.RoleAdmin,
				Status:       service.StatusActive,
			})

			var taskID, recipientID int64
			err := integrationDB.QueryRowContext(ctx, `
				INSERT INTO email_broadcast_tasks
					(title, event, status, template_snapshots, created_by, scheduled_at, total_recipients)
				VALUES ($1, $2, $3, '{}'::jsonb, $4, NOW(), 1)
				RETURNING id`, "finish recipient integration", service.NotificationEmailEventBroadcast,
				service.EmailBroadcastStatusRunning, user.ID).Scan(&taskID)
			require.NoError(t, err)
			t.Cleanup(func() {
				_, _ = integrationDB.ExecContext(ctx, `DELETE FROM email_broadcast_tasks WHERE id = $1`, taskID)
				_, _ = integrationDB.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
			})

			err = integrationDB.QueryRowContext(ctx, `
				INSERT INTO email_broadcast_recipients
					(task_id, user_id, email, recipient_name, status, attempts, claimed_at)
				VALUES ($1, $2, $3, $4, $5, 1, NOW())
				RETURNING id`, taskID, user.ID, user.Email, "integration", service.EmailBroadcastRecipientSending).Scan(&recipientID)
			require.NoError(t, err)

			repo := &emailBroadcastRepository{db: integrationDB}
			require.NoError(t, test.finish(repo, ctx, recipientID))

			var recipientStatus, lastError string
			var sentAt *time.Time
			err = integrationDB.QueryRowContext(ctx, `
				SELECT status, COALESCE(last_error, ''), sent_at
				FROM email_broadcast_recipients WHERE id = $1`, recipientID).
				Scan(&recipientStatus, &lastError, &sentAt)
			require.NoError(t, err)
			require.Equal(t, test.recipientStatus, recipientStatus)
			require.Equal(t, test.expectedLastError, lastError)
			if test.expectSentAt {
				require.NotNil(t, sentAt)
			} else {
				require.Nil(t, sentAt)
			}

			var taskStatus string
			var sentCount, failedCount int64
			err = integrationDB.QueryRowContext(ctx, `
				SELECT status, sent_count, failed_count FROM email_broadcast_tasks WHERE id = $1`, taskID).
				Scan(&taskStatus, &sentCount, &failedCount)
			require.NoError(t, err)
			require.Equal(t, test.taskStatus, taskStatus)
			require.Equal(t, test.sentCount, sentCount)
			require.Equal(t, test.failedCount, failedCount)
		})
	}
}

func TestEmailBroadcastRepositoryClaimDoesNotReclaimSendingRecipient(t *testing.T) {
	ctx := context.Background()
	user := mustCreateUser(t, integrationEntClient, &service.User{
		Email:        fmt.Sprintf("email-broadcast-claim-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hash",
		Role:         service.RoleAdmin,
		Status:       service.StatusActive,
	})

	var taskID, sendingID, pendingID int64
	err := integrationDB.QueryRowContext(ctx, `
		INSERT INTO email_broadcast_tasks
			(title, event, status, template_snapshots, created_by, scheduled_at, total_recipients)
		VALUES ($1, $2, $3, '{}'::jsonb, $4, NOW(), 2)
		RETURNING id`, "claim recipient integration", service.NotificationEmailEventBroadcast,
		service.EmailBroadcastStatusPending, user.ID).Scan(&taskID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(ctx, `DELETE FROM email_broadcast_tasks WHERE id = $1`, taskID)
		_, _ = integrationDB.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
	})

	err = integrationDB.QueryRowContext(ctx, `
		INSERT INTO email_broadcast_recipients
			(task_id, user_id, email, recipient_name, status, attempts, claimed_at)
		VALUES ($1, $2, $3, $4, $5, 8, NOW() - INTERVAL '1 hour')
		RETURNING id`, taskID, user.ID, user.Email, "already processed", service.EmailBroadcastRecipientSending).Scan(&sendingID)
	require.NoError(t, err)

	// A nullable user_id keeps this recipient distinct from the protected row.
	err = integrationDB.QueryRowContext(ctx, `
		INSERT INTO email_broadcast_recipients
			(task_id, email, recipient_name, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id`, taskID, "pending-claim-"+user.Email, "new recipient", service.EmailBroadcastRecipientPending).Scan(&pendingID)
	require.NoError(t, err)

	repo := &emailBroadcastRepository{db: integrationDB}
	delivery, err := repo.ClaimNextRecipient(ctx)
	require.NoError(t, err)
	require.NotNil(t, delivery)
	require.Equal(t, pendingID, delivery.Recipient.ID)
	require.Equal(t, 1, delivery.Recipient.Attempts)

	var sendingStatus string
	var sendingAttempts int
	var sendingClaimedAt *time.Time
	err = integrationDB.QueryRowContext(ctx, `
		SELECT status, attempts, claimed_at
		FROM email_broadcast_recipients WHERE id = $1`, sendingID).
		Scan(&sendingStatus, &sendingAttempts, &sendingClaimedAt)
	require.NoError(t, err)
	require.Equal(t, service.EmailBroadcastRecipientSending, sendingStatus)
	require.Equal(t, 8, sendingAttempts)
	require.NotNil(t, sendingClaimedAt)
}
