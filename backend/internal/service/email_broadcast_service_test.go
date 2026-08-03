package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type emailBroadcastRepositoryStub struct {
	created *EmailBroadcastTask
	count   int64
}

func (r *emailBroadcastRepositoryStub) CountEligibleRecipients(context.Context, EmailBroadcastAudience) (int64, error) {
	return r.count, nil
}

func (r *emailBroadcastRepositoryStub) CreateTask(_ context.Context, task *EmailBroadcastTask) error {
	copy := *task
	copy.ID = 42
	copy.TotalRecipients = r.count
	copy.CreatedAt = time.Now().UTC()
	copy.UpdatedAt = copy.CreatedAt
	*task = copy
	r.created = &copy
	return nil
}

func (r *emailBroadcastRepositoryStub) ListTasks(context.Context, pagination.PaginationParams) ([]EmailBroadcastTask, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{}, nil
}

func (r *emailBroadcastRepositoryStub) GetTask(context.Context, int64) (*EmailBroadcastTask, error) {
	return r.created, nil
}

func (r *emailBroadcastRepositoryStub) ListRecipients(context.Context, int64, string, pagination.PaginationParams) ([]EmailBroadcastRecipient, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{}, nil
}

func (r *emailBroadcastRepositoryStub) ClaimNextRecipient(context.Context) (*EmailBroadcastDelivery, error) {
	return nil, nil
}

func (r *emailBroadcastRepositoryStub) CompleteRecipient(context.Context, int64) error { return nil }
func (r *emailBroadcastRepositoryStub) FailRecipient(context.Context, int64, string) error {
	return nil
}
func (r *emailBroadcastRepositoryStub) CancelTask(context.Context, int64, int64) (bool, error) {
	return true, nil
}
func (r *emailBroadcastRepositoryStub) RetryFailed(context.Context, int64) (int64, error) {
	return 0, nil
}

func validEmailBroadcastVariables() map[string]string {
	return map[string]string{
		"broadcast_subject_zh": "服务器升级通知",
		"broadcast_heading_zh": "服务器升级",
		"broadcast_body_zh":    "服务可能短暂不可用。",
		"broadcast_action_zh":  "请提前保存工作。",
		"broadcast_subject_en": "Server upgrade notice",
		"broadcast_heading_en": "Server upgrade",
		"broadcast_body_en":    "The service may be briefly unavailable.",
		"broadcast_action_en":  "Please save ongoing work.",
	}
}

func TestEmailBroadcastCreateTaskSnapshotsBothLocales(t *testing.T) {
	ctx := context.Background()
	settings := newNotificationEmailMemorySettingRepo()
	notification := NewNotificationEmailService(settings, nil)
	_, err := notification.UpdateTemplate(ctx, NotificationEmailEventBroadcast, "zh", "自定义：{{broadcast_subject_zh}}", "<p>{{broadcast_body_zh}}</p>")
	require.NoError(t, err)
	repo := &emailBroadcastRepositoryStub{count: 307}
	service := NewEmailBroadcastService(repo, notification)

	task, err := service.CreateTask(ctx, EmailBroadcastCreateInput{
		Title:       "August maintenance",
		ScheduledAt: time.Now().UTC().Add(2 * time.Hour),
		Variables:   validEmailBroadcastVariables(),
		Audience:    EmailBroadcastAudience{Mode: EmailBroadcastAudienceRole, Roles: []string{"user"}},
	}, 9)
	require.NoError(t, err)
	require.Equal(t, int64(42), task.ID)
	require.Equal(t, int64(307), task.TotalRecipients)
	require.Equal(t, EmailBroadcastStatusScheduled, task.Status)
	require.Equal(t, EmailBroadcastAudienceRole, task.Audience.Mode)
	require.Contains(t, task.TemplateSnapshots, "en")
	require.Equal(t, "自定义：{{broadcast_subject_zh}}", task.TemplateSnapshots["zh"].Subject)

	_, err = notification.UpdateTemplate(ctx, NotificationEmailEventBroadcast, "zh", "后来修改：{{broadcast_subject_zh}}", "<p>{{broadcast_action_zh}}</p>")
	require.NoError(t, err)
	require.Equal(t, "自定义：{{broadcast_subject_zh}}", repo.created.TemplateSnapshots["zh"].Subject)
}

func TestEmailBroadcastCreateTaskRequiresBroadcastContent(t *testing.T) {
	repo := &emailBroadcastRepositoryStub{count: 1}
	service := NewEmailBroadcastService(repo, NewNotificationEmailService(newNotificationEmailMemorySettingRepo(), nil))
	variables := validEmailBroadcastVariables()
	delete(variables, "broadcast_body_zh")
	delete(variables, "broadcast_body_en")

	_, err := service.CreateTask(context.Background(), EmailBroadcastCreateInput{
		Title:       "Maintenance",
		ScheduledAt: time.Now().UTC().Add(time.Hour),
		Variables:   variables,
	}, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "broadcast_body_zh")
	require.Nil(t, repo.created)
}

func TestNormalizeEmailBroadcastAudience(t *testing.T) {
	audience, err := normalizeEmailBroadcastAudience(EmailBroadcastAudience{
		Mode:   EmailBroadcastAudienceSelected,
		Emails: []string{" User@Example.com ", "user@example.com"},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"user@example.com"}, audience.Emails)

	_, err = normalizeEmailBroadcastAudience(EmailBroadcastAudience{Mode: EmailBroadcastAudienceGroups})
	require.Error(t, err)
}

func TestMaintenanceEmailTemplateIsTransactionalAndPreviewable(t *testing.T) {
	service := NewNotificationEmailService(newNotificationEmailMemorySettingRepo(), nil)
	var maintenance *NotificationEmailEventInfo
	for _, info := range service.ListEventInfos() {
		if info.Event == NotificationEmailEventMaintenance {
			copy := info
			maintenance = &copy
			break
		}
	}
	require.NotNil(t, maintenance)
	require.False(t, maintenance.Optional)
	require.Contains(t, maintenance.Placeholders, "maintenance_start")

	preview, err := service.PreviewTemplate(context.Background(), NotificationEmailPreviewInput{
		Event:  NotificationEmailEventMaintenance,
		Locale: "zh-CN",
		Variables: map[string]string{
			"maintenance_title":    "服务器升级",
			"maintenance_start":    "23:00",
			"maintenance_end":      "00:30",
			"maintenance_timezone": "Asia/Shanghai",
			"maintenance_impact":   "服务短暂不可用",
			"maintenance_action":   "请提前保存工作",
		},
	})
	require.NoError(t, err)
	require.Contains(t, preview.Subject, "服务器升级")
	require.Contains(t, preview.HTML, "23:00")
	require.Contains(t, preview.HTML, "服务短暂不可用")
}

func TestGeneralBroadcastTemplateRendersLocalizedContent(t *testing.T) {
	service := NewNotificationEmailService(newNotificationEmailMemorySettingRepo(), nil)
	preview, err := service.PreviewTemplate(context.Background(), NotificationEmailPreviewInput{
		Event:  NotificationEmailEventBroadcast,
		Locale: "zh-CN",
		Variables: map[string]string{
			"broadcast_subject_zh": "域名迁移通知",
			"broadcast_heading_zh": "新域名已启用",
			"broadcast_body_zh":    "请使用 https://aivoza.com\n原域名继续保留。",
			"broadcast_action_zh":  "无需修改账号。",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "域名迁移通知", preview.Subject)
	require.Contains(t, preview.HTML, "https://aivoza.com")
	require.Contains(t, preview.HTML, "原域名继续保留")
}
