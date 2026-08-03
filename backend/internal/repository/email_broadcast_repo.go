package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type emailBroadcastRepository struct {
	db *sql.DB
}

func NewEmailBroadcastRepository(db *sql.DB) service.EmailBroadcastRepository {
	return &emailBroadcastRepository{db: db}
}

func (r *emailBroadcastRepository) CountEligibleRecipients(ctx context.Context, audience service.EmailBroadcastAudience) (int64, error) {
	where, args := emailBroadcastAudienceWhere(audience, 1)
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users u WHERE `+where, args...).Scan(&count)
	return count, err
}

func (r *emailBroadcastRepository) CreateTask(ctx context.Context, task *service.EmailBroadcastTask) error {
	if task == nil {
		return fmt.Errorf("email broadcast task is required")
	}
	variablesJSON, err := json.Marshal(task.Variables)
	if err != nil {
		return fmt.Errorf("encode broadcast variables: %w", err)
	}
	templatesJSON, err := json.Marshal(task.TemplateSnapshots)
	if err != nil {
		return fmt.Errorf("encode broadcast templates: %w", err)
	}
	audienceJSON, err := json.Marshal(task.Audience)
	if err != nil {
		return fmt.Errorf("encode broadcast audience: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO email_broadcast_tasks
			(title, event, status, variables, audience, template_snapshots, created_by, scheduled_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`,
		task.Title, task.Event, task.Status, variablesJSON, audienceJSON, templatesJSON, task.CreatedBy, task.ScheduledAt,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return err
	}

	where, audienceArgs := emailBroadcastAudienceWhere(task.Audience, 2)
	recipientArgs := append([]any{task.ID}, audienceArgs...)
	result, err := tx.ExecContext(ctx, `
		INSERT INTO email_broadcast_recipients (task_id, user_id, email, recipient_name)
		SELECT $1, u.id, LOWER(BTRIM(u.email)), COALESCE(NULLIF(BTRIM(u.username), ''), SPLIT_PART(u.email, '@', 1))
		FROM users u
		WHERE `+where+`
		ORDER BY u.id`, recipientArgs...)
	if err != nil {
		return err
	}
	task.TotalRecipients, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if task.TotalRecipients == 0 {
		return fmt.Errorf("no active recipients")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE email_broadcast_tasks SET total_recipients = $1 WHERE id = $2`, task.TotalRecipients, task.ID); err != nil {
		return err
	}
	return tx.Commit()
}

const emailBroadcastTaskColumns = `id, title, event, status, variables, audience, template_snapshots, created_by,
	scheduled_at, total_recipients, sent_count, failed_count, canceled_by, canceled_at,
	started_at, finished_at, created_at, updated_at`

type emailBroadcastRowScanner interface {
	Scan(dest ...any) error
}

func scanEmailBroadcastTask(row emailBroadcastRowScanner) (*service.EmailBroadcastTask, error) {
	var task service.EmailBroadcastTask
	var variablesJSON, audienceJSON, templatesJSON []byte
	var canceledBy sql.NullInt64
	var canceledAt, startedAt, finishedAt sql.NullTime
	err := row.Scan(
		&task.ID, &task.Title, &task.Event, &task.Status, &variablesJSON, &audienceJSON, &templatesJSON, &task.CreatedBy,
		&task.ScheduledAt, &task.TotalRecipients, &task.SentCount, &task.FailedCount, &canceledBy, &canceledAt,
		&startedAt, &finishedAt, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(variablesJSON, &task.Variables); err != nil {
		return nil, fmt.Errorf("decode broadcast variables: %w", err)
	}
	if err := json.Unmarshal(audienceJSON, &task.Audience); err != nil {
		return nil, fmt.Errorf("decode broadcast audience: %w", err)
	}
	if err := json.Unmarshal(templatesJSON, &task.TemplateSnapshots); err != nil {
		return nil, fmt.Errorf("decode broadcast templates: %w", err)
	}
	if canceledBy.Valid {
		value := canceledBy.Int64
		task.CanceledBy = &value
	}
	if canceledAt.Valid {
		task.CanceledAt = &canceledAt.Time
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		task.FinishedAt = &finishedAt.Time
	}
	return &task, nil
}

func emailBroadcastAudienceWhere(audience service.EmailBroadcastAudience, firstArg int) (string, []any) {
	where := "u.status = 'active' AND u.deleted_at IS NULL AND BTRIM(u.email) <> ''"
	args := make([]any, 0, 1)
	placeholder := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", firstArg+len(args)-1)
	}
	switch audience.Mode {
	case service.EmailBroadcastAudienceRole:
		where += " AND u.role = ANY(" + placeholder(pq.Array(audience.Roles)) + ")"
	case service.EmailBroadcastAudienceGroups:
		where += " AND EXISTS (SELECT 1 FROM user_allowed_groups uag WHERE uag.user_id = u.id AND uag.group_id = ANY(" + placeholder(pq.Array(audience.GroupIDs)) + "))"
	case service.EmailBroadcastAudienceSelected:
		where += " AND LOWER(BTRIM(u.email)) = ANY(" + placeholder(pq.Array(audience.Emails)) + ")"
	}
	return where, args
}

func (r *emailBroadcastRepository) ListTasks(ctx context.Context, params pagination.PaginationParams) ([]service.EmailBroadcastTask, *pagination.PaginationResult, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM email_broadcast_tasks`).Scan(&total); err != nil {
		return nil, nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+emailBroadcastTaskColumns+`
		FROM email_broadcast_tasks ORDER BY created_at DESC, id DESC LIMIT $1 OFFSET $2`, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.EmailBroadcastTask, 0)
	for rows.Next() {
		item, err := scanEmailBroadcastTask(rows)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *emailBroadcastRepository) GetTask(ctx context.Context, taskID int64) (*service.EmailBroadcastTask, error) {
	return scanEmailBroadcastTask(r.db.QueryRowContext(ctx, `SELECT `+emailBroadcastTaskColumns+` FROM email_broadcast_tasks WHERE id = $1`, taskID))
}

const emailBroadcastRecipientColumns = `id, task_id, user_id, email, recipient_name, locale, status, attempts, last_error, sent_at, created_at, updated_at`
const emailBroadcastRecipientReturningColumns = `r.id, r.task_id, r.user_id, r.email, r.recipient_name, r.locale, r.status, r.attempts, r.last_error, r.sent_at, r.created_at, r.updated_at`

func scanEmailBroadcastRecipient(row emailBroadcastRowScanner) (*service.EmailBroadcastRecipient, error) {
	var item service.EmailBroadcastRecipient
	var userID sql.NullInt64
	var lastError sql.NullString
	var sentAt sql.NullTime
	if err := row.Scan(&item.ID, &item.TaskID, &userID, &item.Email, &item.RecipientName, &item.Locale, &item.Status, &item.Attempts, &lastError, &sentAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	if userID.Valid {
		value := userID.Int64
		item.UserID = &value
	}
	if lastError.Valid {
		item.LastError = &lastError.String
	}
	if sentAt.Valid {
		item.SentAt = &sentAt.Time
	}
	return &item, nil
}

func (r *emailBroadcastRepository) ListRecipients(ctx context.Context, taskID int64, status string, params pagination.PaginationParams) ([]service.EmailBroadcastRecipient, *pagination.PaginationResult, error) {
	where := "task_id = $1"
	args := []any{taskID}
	if status = strings.TrimSpace(status); status != "" {
		args = append(args, status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM email_broadcast_recipients WHERE `+where, args...).Scan(&total); err != nil {
		return nil, nil, err
	}
	args = append(args, params.Limit(), params.Offset())
	rows, err := r.db.QueryContext(ctx, `SELECT `+emailBroadcastRecipientColumns+` FROM email_broadcast_recipients WHERE `+where+
		fmt.Sprintf(" ORDER BY id ASC LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.EmailBroadcastRecipient, 0)
	for rows.Next() {
		item, err := scanEmailBroadcastRecipient(rows)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return items, paginationResultFromTotal(total, params), nil
}

func (r *emailBroadcastRepository) ClaimNextRecipient(ctx context.Context, staleAfter time.Duration) (*service.EmailBroadcastDelivery, error) {
	staleSeconds := int64(staleAfter.Seconds())
	if staleSeconds < 30 {
		staleSeconds = 120
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	row := tx.QueryRowContext(ctx, `
		WITH candidate AS (
			SELECT r.id
			FROM email_broadcast_recipients r
			JOIN email_broadcast_tasks t ON t.id = r.task_id
			WHERE t.status IN ($1, $2, $3)
			  AND t.scheduled_at <= NOW()
			  AND ((r.status = $4 AND r.next_attempt_at <= NOW())
			    OR (r.status = $5 AND r.claimed_at < NOW() - ($6 * interval '1 second')))
			ORDER BY t.scheduled_at ASC, t.id ASC, r.id ASC
			LIMIT 1
			FOR UPDATE OF r, t SKIP LOCKED
		)
		UPDATE email_broadcast_recipients r
		SET status = $5, attempts = attempts + 1, claimed_at = NOW(), last_error = NULL, updated_at = NOW()
		FROM candidate WHERE r.id = candidate.id
		RETURNING `+emailBroadcastRecipientReturningColumns,
		service.EmailBroadcastStatusScheduled, service.EmailBroadcastStatusPending, service.EmailBroadcastStatusRunning,
		service.EmailBroadcastRecipientPending, service.EmailBroadcastRecipientSending, staleSeconds)
	recipient, err := scanEmailBroadcastRecipient(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE email_broadcast_tasks SET status = $1, started_at = COALESCE(started_at, NOW()), updated_at = NOW() WHERE id = $2 AND status <> $3`,
		service.EmailBroadcastStatusRunning, recipient.TaskID, service.EmailBroadcastStatusCanceled); err != nil {
		return nil, err
	}
	task, err := scanEmailBroadcastTask(tx.QueryRowContext(ctx, `SELECT `+emailBroadcastTaskColumns+` FROM email_broadcast_tasks WHERE id = $1`, recipient.TaskID))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &service.EmailBroadcastDelivery{Task: *task, Recipient: *recipient}, nil
}

func (r *emailBroadcastRepository) CompleteRecipient(ctx context.Context, recipientID int64) error {
	return r.finishRecipient(ctx, recipientID, service.EmailBroadcastRecipientSent, "")
}

func (r *emailBroadcastRepository) FailRecipient(ctx context.Context, recipientID int64, errorMessage string) error {
	return r.finishRecipient(ctx, recipientID, service.EmailBroadcastRecipientFailed, errorMessage)
}

func (r *emailBroadcastRepository) finishRecipient(ctx context.Context, recipientID int64, status, errorMessage string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var taskID int64
	sent := status == service.EmailBroadcastRecipientSent
	err = tx.QueryRowContext(ctx, `UPDATE email_broadcast_recipients SET status = $1, last_error = NULLIF($2, ''), sent_at = CASE WHEN $3 THEN NOW() ELSE sent_at END, claimed_at = NULL, updated_at = NOW() WHERE id = $4 AND status = $5 RETURNING task_id`,
		status, errorMessage, sent, recipientID, service.EmailBroadcastRecipientSending).Scan(&taskID)
	if err != nil {
		return err
	}
	sentIncrement, failedIncrement := int64(0), int64(0)
	if status == service.EmailBroadcastRecipientSent {
		sentIncrement = 1
	} else {
		failedIncrement = 1
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE email_broadcast_tasks
		SET sent_count = sent_count + $1, failed_count = failed_count + $2,
			status = CASE WHEN sent_count + failed_count + $1 + $2 >= total_recipients
				THEN CASE WHEN failed_count + $2 > 0 THEN $3 ELSE $4 END ELSE status END,
			finished_at = CASE WHEN sent_count + failed_count + $1 + $2 >= total_recipients THEN NOW() ELSE finished_at END,
			updated_at = NOW()
		WHERE id = $5 AND status <> $6`, sentIncrement, failedIncrement,
		service.EmailBroadcastStatusPartiallyFailed, service.EmailBroadcastStatusSucceeded, taskID, service.EmailBroadcastStatusCanceled); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *emailBroadcastRepository) RetryRecipient(ctx context.Context, recipientID int64, errorMessage string, nextAttemptAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE email_broadcast_recipients SET status = $1, last_error = $2, claimed_at = NULL, next_attempt_at = $3, updated_at = NOW() WHERE id = $4 AND status = $5`,
		service.EmailBroadcastRecipientPending, errorMessage, nextAttemptAt, recipientID, service.EmailBroadcastRecipientSending)
	return err
}

func (r *emailBroadcastRepository) CancelTask(ctx context.Context, taskID, canceledBy int64) (bool, error) {
	result, err := r.db.ExecContext(ctx, `UPDATE email_broadcast_tasks SET status = $1, canceled_by = $2, canceled_at = NOW(), finished_at = NOW(), updated_at = NOW() WHERE id = $3 AND status IN ($4, $5, $6)`,
		service.EmailBroadcastStatusCanceled, canceledBy, taskID, service.EmailBroadcastStatusScheduled, service.EmailBroadcastStatusPending, service.EmailBroadcastStatusRunning)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count > 0, err
}

func (r *emailBroadcastRepository) RetryFailed(ctx context.Context, taskID int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `UPDATE email_broadcast_recipients SET status = $1, attempts = 0, last_error = NULL, claimed_at = NULL, next_attempt_at = NOW(), updated_at = NOW()
		WHERE task_id = $2 AND status = $3
		  AND EXISTS (SELECT 1 FROM email_broadcast_tasks WHERE id = $2 AND status = $4)`,
		service.EmailBroadcastRecipientPending, taskID, service.EmailBroadcastRecipientFailed, service.EmailBroadcastStatusPartiallyFailed)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, sql.ErrNoRows
	}
	if _, err := tx.ExecContext(ctx, `UPDATE email_broadcast_tasks SET status = $1, failed_count = GREATEST(failed_count - $2, 0), finished_at = NULL, canceled_by = NULL, canceled_at = NULL, updated_at = NOW() WHERE id = $3 AND status = $4`,
		service.EmailBroadcastStatusPending, count, taskID, service.EmailBroadcastStatusPartiallyFailed); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}
