package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	emailBroadcastRateInterval  = 200 * time.Millisecond
	emailBroadcastUpdateTimeout = 10 * time.Second
)

type EmailBroadcastService struct {
	repo         EmailBroadcastRepository
	notification *NotificationEmailService

	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
	stop   sync.Once
}

func NewEmailBroadcastService(repo EmailBroadcastRepository, notification *NotificationEmailService) *EmailBroadcastService {
	ctx, cancel := context.WithCancel(context.Background())
	return &EmailBroadcastService{repo: repo, notification: notification, ctx: ctx, cancel: cancel}
}

func ProvideEmailBroadcastService(repo EmailBroadcastRepository, notification *NotificationEmailService) *EmailBroadcastService {
	service := NewEmailBroadcastService(repo, notification)
	service.Start()
	return service
}

func (s *EmailBroadcastService) Start() {
	if s == nil || s.repo == nil || s.notification == nil {
		return
	}
	s.once.Do(func() {
		go s.worker()
	})
}

func (s *EmailBroadcastService) Stop() {
	if s == nil {
		return
	}
	s.stop.Do(func() {
		if s.cancel != nil {
			s.cancel()
		}
	})
}

func (s *EmailBroadcastService) worker() {
	ticker := time.NewTicker(emailBroadcastRateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.runOnce()
		}
	}
}

func (s *EmailBroadcastService) runOnce() {
	ctx, cancel := context.WithTimeout(s.ctx, 45*time.Second)
	defer cancel()
	delivery, err := s.repo.ClaimNextRecipient(ctx)
	if err != nil || delivery == nil {
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			slog.Error("email broadcast claim failed", "error", err)
		}
		return
	}

	sendCtx, sendCancel := context.WithTimeout(s.ctx, 30*time.Second)
	err = s.sendSnapshot(sendCtx, delivery)
	sendCancel()
	if err == nil {
		updateCtx, updateCancel := context.WithTimeout(context.Background(), emailBroadcastUpdateTimeout)
		updateErr := s.repo.CompleteRecipient(updateCtx, delivery.Recipient.ID)
		updateCancel()
		if updateErr != nil {
			slog.Error("email broadcast completion failed", "recipient_id", delivery.Recipient.ID, "error", updateErr)
		}
		return
	}

	message := strings.TrimSpace(err.Error())
	if len(message) > 500 {
		message = message[:500]
	}
	// SMTP may accept a message before the client observes an error; never retry
	// an ambiguous broadcast delivery automatically.
	updateCtx, updateCancel := context.WithTimeout(context.Background(), emailBroadcastUpdateTimeout)
	updateErr := s.repo.FailRecipient(updateCtx, delivery.Recipient.ID, message)
	updateCancel()
	if updateErr != nil {
		slog.Error("email broadcast failure update failed", "recipient_id", delivery.Recipient.ID, "error", updateErr)
	}
}

func (s *EmailBroadcastService) sendSnapshot(ctx context.Context, delivery *EmailBroadcastDelivery) error {
	if delivery == nil || s.notification == nil || s.notification.emailService == nil {
		return fmt.Errorf("email broadcast service is not configured")
	}
	task := delivery.Task
	locale := normalizeNotificationLocale(delivery.Recipient.Locale)
	if delivery.Recipient.Locale == "" {
		locale = s.notification.ResolveRecipientLocale(ctx, dereferenceInt64(delivery.Recipient.UserID), delivery.Recipient.Email)
	}
	template, ok := task.TemplateSnapshots[locale]
	if !ok {
		template = task.TemplateSnapshots[notificationEmailDefaultLocale]
	}
	if strings.TrimSpace(template.Subject) == "" || strings.TrimSpace(template.HTML) == "" {
		return fmt.Errorf("broadcast template snapshot is incomplete")
	}
	input := NotificationEmailSendInput{
		Event:          task.Event,
		Locale:         locale,
		RecipientEmail: delivery.Recipient.Email,
		RecipientName:  delivery.Recipient.RecipientName,
		UserID:         dereferenceInt64(delivery.Recipient.UserID),
		Variables:      task.Variables,
	}
	variables := s.notification.runtimeVariables(ctx, task.Event, locale, input)
	rendered, err := renderNotificationEmail(task.Event, template.Subject, template.HTML, variables, nil)
	if err != nil {
		return err
	}
	return s.notification.emailService.SendEmail(ctx, delivery.Recipient.Email, rendered.Subject, rendered.HTML)
}

func dereferenceInt64(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func (s *EmailBroadcastService) EstimateRecipients(ctx context.Context, audience EmailBroadcastAudience) (int64, error) {
	if s == nil || s.repo == nil {
		return 0, infraerrors.New(http.StatusServiceUnavailable, "EMAIL_BROADCAST_UNAVAILABLE", "email broadcast service is not configured")
	}
	normalized, err := normalizeEmailBroadcastAudience(audience)
	if err != nil {
		return 0, err
	}
	return s.repo.CountEligibleRecipients(ctx, normalized)
}

func (s *EmailBroadcastService) ListTasks(ctx context.Context, params pagination.PaginationParams) ([]EmailBroadcastTask, *pagination.PaginationResult, error) {
	if s == nil || s.repo == nil {
		return nil, nil, infraerrors.New(http.StatusServiceUnavailable, "EMAIL_BROADCAST_UNAVAILABLE", "email broadcast service is not configured")
	}
	tasks, result, err := s.repo.ListTasks(ctx, params)
	for index := range tasks {
		tasks[index].TemplateSnapshots = nil
	}
	return tasks, result, err
}

func (s *EmailBroadcastService) GetTask(ctx context.Context, taskID int64) (*EmailBroadcastTask, error) {
	if s == nil || s.repo == nil {
		return nil, infraerrors.New(http.StatusServiceUnavailable, "EMAIL_BROADCAST_UNAVAILABLE", "email broadcast service is not configured")
	}
	task, err := s.repo.GetTask(ctx, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, infraerrors.NotFound("EMAIL_BROADCAST_NOT_FOUND", "email broadcast task not found")
	}
	if task != nil {
		task.TemplateSnapshots = nil
	}
	return task, err
}

func (s *EmailBroadcastService) ListRecipients(ctx context.Context, taskID int64, status string, params pagination.PaginationParams) ([]EmailBroadcastRecipient, *pagination.PaginationResult, error) {
	if _, err := s.GetTask(ctx, taskID); err != nil {
		return nil, nil, err
	}
	return s.repo.ListRecipients(ctx, taskID, status, params)
}

func (s *EmailBroadcastService) CreateTask(ctx context.Context, input EmailBroadcastCreateInput, createdBy int64) (*EmailBroadcastTask, error) {
	if s == nil || s.repo == nil || s.notification == nil {
		return nil, infraerrors.New(http.StatusServiceUnavailable, "EMAIL_BROADCAST_UNAVAILABLE", "email broadcast service is not configured")
	}
	if createdBy <= 0 {
		return nil, infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_CREATOR", "invalid creator")
	}
	title := strings.TrimSpace(input.Title)
	if title == "" || len([]rune(title)) > 200 {
		return nil, infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_TITLE", "title is required and must be at most 200 characters")
	}
	if input.ScheduledAt.IsZero() {
		input.ScheduledAt = time.Now().UTC()
	}
	if input.ScheduledAt.Before(time.Now().UTC().Add(-time.Minute)) {
		return nil, infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_SCHEDULE", "scheduled_at cannot be in the past")
	}
	variables := make(map[string]string, len(input.Variables))
	for key, value := range input.Variables {
		key = strings.TrimSpace(key)
		if key == "" || len(key) > 80 {
			continue
		}
		variables[key] = strings.TrimSpace(value)
	}
	variables, err := normalizeEmailBroadcastVariables(variables)
	if err != nil {
		return nil, err
	}
	audience, err := normalizeEmailBroadcastAudience(input.Audience)
	if err != nil {
		return nil, err
	}
	event, err := normalizeEmailBroadcastEvent(input.Event)
	if err != nil {
		return nil, err
	}
	snapshots := make(map[string]EmailBroadcastTemplateSnapshot, 2)
	for _, locale := range []string{notificationEmailDefaultLocale, notificationEmailLocaleChinese} {
		template, err := s.notification.GetTemplate(ctx, event, locale)
		if err != nil {
			return nil, infraerrors.BadRequest("EMAIL_BROADCAST_TEMPLATE_INVALID", err.Error())
		}
		snapshots[locale] = EmailBroadcastTemplateSnapshot{Subject: template.Subject, HTML: template.HTML}
	}
	status := EmailBroadcastStatusPending
	if input.ScheduledAt.After(time.Now().UTC().Add(30 * time.Second)) {
		status = EmailBroadcastStatusScheduled
	}
	task := &EmailBroadcastTask{Title: title, Event: event, Status: status, Variables: variables, Audience: audience, TemplateSnapshots: snapshots, CreatedBy: createdBy, ScheduledAt: input.ScheduledAt.UTC()}
	if err := s.repo.CreateTask(ctx, task); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no active recipients") {
			return nil, infraerrors.BadRequest("EMAIL_BROADCAST_NO_RECIPIENTS", "no active users with an email address")
		}
		return nil, fmt.Errorf("create email broadcast task: %w", err)
	}
	return task, nil
}

type EmailBroadcastCreateInput struct {
	Title       string
	Event       string
	ScheduledAt time.Time
	Variables   map[string]string
	Audience    EmailBroadcastAudience
}

func (s *EmailBroadcastService) SendTest(ctx context.Context, email, locale, event string, variables map[string]string) error {
	if s == nil || s.notification == nil {
		return infraerrors.New(http.StatusServiceUnavailable, "EMAIL_BROADCAST_UNAVAILABLE", "email broadcast service is not configured")
	}
	address, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil || !strings.Contains(address.Address, "@") {
		return infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_EMAIL", "a valid recipient email is required")
	}
	variables, err = normalizeEmailBroadcastVariables(variables)
	if err != nil {
		return err
	}
	event, err = normalizeEmailBroadcastEvent(event)
	if err != nil {
		return err
	}
	return s.notification.Send(ctx, NotificationEmailSendInput{
		Event:          event,
		Locale:         normalizeNotificationLocale(locale),
		RecipientEmail: address.Address,
		RecipientName:  emailRecipientName(address.Address),
		Variables:      variables,
	})
}

func normalizeEmailBroadcastEvent(event string) (string, error) {
	event = strings.ToLower(strings.TrimSpace(event))
	if event == "" {
		return NotificationEmailEventBroadcast, nil
	}
	if event != NotificationEmailEventBroadcast && event != NotificationEmailEventReactivation {
		return "", infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_EVENT", "unsupported email broadcast event")
	}
	return event, nil
}

func normalizeEmailBroadcastVariables(input map[string]string) (map[string]string, error) {
	variables := make(map[string]string, len(input)+8)
	for key, value := range input {
		key = strings.TrimSpace(key)
		if key == "" || len(key) > 80 {
			continue
		}
		variables[key] = strings.TrimSpace(value)
	}
	if variables["broadcast_subject_zh"] == "" && variables["maintenance_title"] != "" {
		legacyBody := strings.TrimSpace(strings.Join([]string{
			variables["maintenance_start"] + " - " + variables["maintenance_end"] + " (" + variables["maintenance_timezone"] + ")",
			variables["maintenance_impact"],
		}, "\n\n"))
		variables["broadcast_subject_zh"] = variables["maintenance_title"]
		variables["broadcast_heading_zh"] = variables["maintenance_title"]
		variables["broadcast_body_zh"] = legacyBody
		variables["broadcast_action_zh"] = variables["maintenance_action"]
	}
	for _, field := range []string{"subject", "heading", "body", "action"} {
		zhKey := "broadcast_" + field + "_zh"
		enKey := "broadcast_" + field + "_en"
		if variables[zhKey] == "" {
			variables[zhKey] = variables[enKey]
		}
		if variables[enKey] == "" {
			variables[enKey] = variables[zhKey]
		}
	}
	for _, key := range []string{"broadcast_subject_zh", "broadcast_heading_zh", "broadcast_body_zh", "broadcast_subject_en", "broadcast_heading_en", "broadcast_body_en"} {
		if variables[key] == "" {
			return nil, infraerrors.BadRequest("EMAIL_BROADCAST_MISSING_CONTENT", key+" is required")
		}
	}
	limits := map[string]int{
		"broadcast_subject_zh": 200,
		"broadcast_subject_en": 200,
		"broadcast_heading_zh": 300,
		"broadcast_heading_en": 300,
		"broadcast_body_zh":    20000,
		"broadcast_body_en":    20000,
		"broadcast_action_zh":  2000,
		"broadcast_action_en":  2000,
	}
	for key, limit := range limits {
		if len([]rune(variables[key])) > limit {
			return nil, infraerrors.BadRequest("EMAIL_BROADCAST_CONTENT_TOO_LONG", key+" is too long")
		}
	}
	return variables, nil
}

func normalizeEmailBroadcastAudience(input EmailBroadcastAudience) (EmailBroadcastAudience, error) {
	mode := strings.ToLower(strings.TrimSpace(input.Mode))
	if mode == "" {
		mode = EmailBroadcastAudienceAll
	}
	audience := EmailBroadcastAudience{Mode: mode}
	switch mode {
	case EmailBroadcastAudienceAll:
		return audience, nil
	case EmailBroadcastAudienceRole:
		seen := make(map[string]struct{}, len(input.Roles))
		for _, role := range input.Roles {
			role = strings.ToLower(strings.TrimSpace(role))
			if role != "admin" && role != "user" {
				return EmailBroadcastAudience{}, infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_AUDIENCE", "roles must contain only admin or user")
			}
			if _, ok := seen[role]; !ok {
				seen[role] = struct{}{}
				audience.Roles = append(audience.Roles, role)
			}
		}
		if len(audience.Roles) == 0 {
			return EmailBroadcastAudience{}, infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_AUDIENCE", "at least one role is required")
		}
	case EmailBroadcastAudienceGroups:
		seen := make(map[int64]struct{}, len(input.GroupIDs))
		for _, groupID := range input.GroupIDs {
			if groupID <= 0 {
				continue
			}
			if _, ok := seen[groupID]; !ok {
				seen[groupID] = struct{}{}
				audience.GroupIDs = append(audience.GroupIDs, groupID)
			}
		}
		if len(audience.GroupIDs) == 0 {
			return EmailBroadcastAudience{}, infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_AUDIENCE", "at least one group is required")
		}
	case EmailBroadcastAudienceSelected:
		if len(input.Emails) > 5000 {
			return EmailBroadcastAudience{}, infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_AUDIENCE", "at most 5000 email addresses can be selected")
		}
		seen := make(map[string]struct{}, len(input.Emails))
		for _, raw := range input.Emails {
			address, err := mail.ParseAddress(strings.TrimSpace(raw))
			if err != nil || !strings.Contains(address.Address, "@") {
				return EmailBroadcastAudience{}, infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_AUDIENCE", "selected recipients contain an invalid email address")
			}
			email := strings.ToLower(strings.TrimSpace(address.Address))
			if _, ok := seen[email]; !ok {
				seen[email] = struct{}{}
				audience.Emails = append(audience.Emails, email)
			}
		}
		if len(audience.Emails) == 0 {
			return EmailBroadcastAudience{}, infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_AUDIENCE", "at least one email address is required")
		}
	case EmailBroadcastAudienceInactive:
		if input.InactiveDays < 1 || input.InactiveDays > 3650 {
			return EmailBroadcastAudience{}, infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_AUDIENCE", "inactive_days must be between 1 and 3650")
		}
		audience.InactiveDays = input.InactiveDays
	default:
		return EmailBroadcastAudience{}, infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_AUDIENCE", "unsupported audience mode")
	}
	return audience, nil
}

func (s *EmailBroadcastService) CancelTask(ctx context.Context, taskID, canceledBy int64) error {
	if canceledBy <= 0 {
		return infraerrors.BadRequest("EMAIL_BROADCAST_INVALID_CANCELLER", "invalid canceller")
	}
	changed, err := s.repo.CancelTask(ctx, taskID, canceledBy)
	if err != nil {
		return err
	}
	if !changed {
		return infraerrors.New(http.StatusConflict, "EMAIL_BROADCAST_NOT_CANCELABLE", "task is already finished or not found")
	}
	return nil
}

func (s *EmailBroadcastService) RetryFailed(ctx context.Context, taskID int64) (int64, error) {
	count, err := s.repo.RetryFailed(ctx, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, infraerrors.New(http.StatusConflict, "EMAIL_BROADCAST_NO_FAILED_RECIPIENTS", "no failed recipients to retry")
	}
	return count, err
}
