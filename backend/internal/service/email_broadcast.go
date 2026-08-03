package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	EmailBroadcastStatusScheduled       = "scheduled"
	EmailBroadcastStatusPending         = "pending"
	EmailBroadcastStatusRunning         = "running"
	EmailBroadcastStatusSucceeded       = "succeeded"
	EmailBroadcastStatusPartiallyFailed = "partially_failed"
	EmailBroadcastStatusCanceled        = "canceled"

	EmailBroadcastRecipientPending = "pending"
	EmailBroadcastRecipientSending = "sending"
	EmailBroadcastRecipientSent    = "sent"
	EmailBroadcastRecipientFailed  = "failed"

	EmailBroadcastAudienceAll      = "all"
	EmailBroadcastAudienceRole     = "role"
	EmailBroadcastAudienceGroups   = "groups"
	EmailBroadcastAudienceSelected = "selected"
)

type EmailBroadcastTemplateSnapshot struct {
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

type EmailBroadcastTask struct {
	ID                int64                                     `json:"id"`
	Title             string                                    `json:"title"`
	Event             string                                    `json:"event"`
	Status            string                                    `json:"status"`
	Variables         map[string]string                         `json:"variables"`
	Audience          EmailBroadcastAudience                    `json:"audience"`
	TemplateSnapshots map[string]EmailBroadcastTemplateSnapshot `json:"template_snapshots,omitempty"`
	CreatedBy         int64                                     `json:"created_by"`
	ScheduledAt       time.Time                                 `json:"scheduled_at"`
	TotalRecipients   int64                                     `json:"total_recipients"`
	SentCount         int64                                     `json:"sent_count"`
	FailedCount       int64                                     `json:"failed_count"`
	CanceledBy        *int64                                    `json:"canceled_by,omitempty"`
	CanceledAt        *time.Time                                `json:"canceled_at,omitempty"`
	StartedAt         *time.Time                                `json:"started_at,omitempty"`
	FinishedAt        *time.Time                                `json:"finished_at,omitempty"`
	CreatedAt         time.Time                                 `json:"created_at"`
	UpdatedAt         time.Time                                 `json:"updated_at"`
}

type EmailBroadcastAudience struct {
	Mode     string   `json:"mode"`
	Roles    []string `json:"roles,omitempty"`
	GroupIDs []int64  `json:"group_ids,omitempty"`
	Emails   []string `json:"emails,omitempty"`
}

type EmailBroadcastRecipient struct {
	ID            int64      `json:"id"`
	TaskID        int64      `json:"task_id"`
	UserID        *int64     `json:"user_id,omitempty"`
	Email         string     `json:"email"`
	RecipientName string     `json:"recipient_name"`
	Locale        string     `json:"locale"`
	Status        string     `json:"status"`
	Attempts      int        `json:"attempts"`
	LastError     *string    `json:"last_error,omitempty"`
	SentAt        *time.Time `json:"sent_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type EmailBroadcastDelivery struct {
	Task      EmailBroadcastTask
	Recipient EmailBroadcastRecipient
}

type EmailBroadcastRepository interface {
	CountEligibleRecipients(ctx context.Context, audience EmailBroadcastAudience) (int64, error)
	CreateTask(ctx context.Context, task *EmailBroadcastTask) error
	ListTasks(ctx context.Context, params pagination.PaginationParams) ([]EmailBroadcastTask, *pagination.PaginationResult, error)
	GetTask(ctx context.Context, taskID int64) (*EmailBroadcastTask, error)
	ListRecipients(ctx context.Context, taskID int64, status string, params pagination.PaginationParams) ([]EmailBroadcastRecipient, *pagination.PaginationResult, error)
	ClaimNextRecipient(ctx context.Context, staleAfter time.Duration) (*EmailBroadcastDelivery, error)
	CompleteRecipient(ctx context.Context, recipientID int64) error
	RetryRecipient(ctx context.Context, recipientID int64, errorMessage string, nextAttemptAt time.Time) error
	FailRecipient(ctx context.Context, recipientID int64, errorMessage string) error
	CancelTask(ctx context.Context, taskID, canceledBy int64) (bool, error)
	RetryFailed(ctx context.Context, taskID int64) (int64, error)
}
