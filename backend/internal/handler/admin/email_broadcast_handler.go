package admin

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type EmailBroadcastHandler struct {
	service *service.EmailBroadcastService
}

func NewEmailBroadcastHandler(broadcastService *service.EmailBroadcastService) *EmailBroadcastHandler {
	return &EmailBroadcastHandler{service: broadcastService}
}

type emailBroadcastRequest struct {
	Title       string                         `json:"title"`
	ScheduledAt *string                        `json:"scheduled_at"`
	Email       string                         `json:"email"`
	Locale      string                         `json:"locale"`
	SubjectZH   string                         `json:"subject_zh"`
	HeadingZH   string                         `json:"heading_zh"`
	BodyZH      string                         `json:"body_zh"`
	ActionZH    string                         `json:"action_zh"`
	SubjectEN   string                         `json:"subject_en"`
	HeadingEN   string                         `json:"heading_en"`
	BodyEN      string                         `json:"body_en"`
	ActionEN    string                         `json:"action_en"`
	Audience    service.EmailBroadcastAudience `json:"audience"`

	// Legacy maintenance fields remain accepted for API compatibility.
	MaintenanceTitle    string `json:"maintenance_title"`
	MaintenanceStart    string `json:"maintenance_start"`
	MaintenanceEnd      string `json:"maintenance_end"`
	MaintenanceTimezone string `json:"maintenance_timezone"`
	MaintenanceImpact   string `json:"maintenance_impact"`
	MaintenanceAction   string `json:"maintenance_action"`
}

func (r emailBroadcastRequest) variables() map[string]string {
	return map[string]string{
		"broadcast_subject_zh": strings.TrimSpace(r.SubjectZH),
		"broadcast_heading_zh": strings.TrimSpace(r.HeadingZH),
		"broadcast_body_zh":    strings.TrimSpace(r.BodyZH),
		"broadcast_action_zh":  strings.TrimSpace(r.ActionZH),
		"broadcast_subject_en": strings.TrimSpace(r.SubjectEN),
		"broadcast_heading_en": strings.TrimSpace(r.HeadingEN),
		"broadcast_body_en":    strings.TrimSpace(r.BodyEN),
		"broadcast_action_en":  strings.TrimSpace(r.ActionEN),
		"maintenance_title":    strings.TrimSpace(r.MaintenanceTitle),
		"maintenance_start":    strings.TrimSpace(r.MaintenanceStart),
		"maintenance_end":      strings.TrimSpace(r.MaintenanceEnd),
		"maintenance_timezone": strings.TrimSpace(r.MaintenanceTimezone),
		"maintenance_impact":   strings.TrimSpace(r.MaintenanceImpact),
		"maintenance_action":   strings.TrimSpace(r.MaintenanceAction),
	}
}

func (h *EmailBroadcastHandler) Estimate(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "Email broadcast service unavailable")
		return
	}
	count, err := h.service.EstimateRecipients(c.Request.Context(), service.EmailBroadcastAudience{Mode: service.EmailBroadcastAudienceAll})
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, gin.H{"eligible_recipients": count})
}

func (h *EmailBroadcastHandler) EstimateFiltered(c *gin.Context) {
	if h == nil || h.service == nil {
		response.Error(c, http.StatusServiceUnavailable, "Email broadcast service unavailable")
		return
	}
	var request struct {
		Audience service.EmailBroadcastAudience `json:"audience"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	count, err := h.service.EstimateRecipients(c.Request.Context(), request.Audience)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, gin.H{"eligible_recipients": count})
}

func (h *EmailBroadcastHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	tasks, result, err := h.service.ListTasks(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: pageSize})
	if response.ErrorFrom(c, err) {
		return
	}
	response.Paginated(c, tasks, result.Total, page, pageSize)
}

func (h *EmailBroadcastHandler) Get(c *gin.Context) {
	taskID, ok := emailBroadcastTaskID(c)
	if !ok {
		return
	}
	task, err := h.service.GetTask(c.Request.Context(), taskID)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, task)
}

func (h *EmailBroadcastHandler) ListRecipients(c *gin.Context) {
	taskID, ok := emailBroadcastTaskID(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.service.ListRecipients(c.Request.Context(), taskID, c.Query("status"), pagination.PaginationParams{Page: page, PageSize: pageSize})
	if response.ErrorFrom(c, err) {
		return
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}

func (h *EmailBroadcastHandler) Create(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	var request emailBroadcastRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	scheduledAt := time.Now().UTC()
	if request.ScheduledAt != nil && strings.TrimSpace(*request.ScheduledAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*request.ScheduledAt))
		if err != nil {
			response.BadRequest(c, "scheduled_at must use RFC3339 format")
			return
		}
		scheduledAt = parsed.UTC()
	}
	input := service.EmailBroadcastCreateInput{
		Title: request.Title, ScheduledAt: scheduledAt, Variables: request.variables(), Audience: request.Audience,
	}
	executeAdminIdempotentJSON(c, "admin.email_broadcasts.create", request, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		task, err := h.service.CreateTask(ctx, input, subject.UserID)
		if task != nil {
			task.TemplateSnapshots = nil
		}
		return task, err
	})
}

func (h *EmailBroadcastHandler) Test(c *gin.Context) {
	var request emailBroadcastRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.service.SendTest(c.Request.Context(), request.Email, request.Locale, request.variables()); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Test email sent"})
}

func (h *EmailBroadcastHandler) Cancel(c *gin.Context) {
	taskID, ok := emailBroadcastTaskID(c)
	if !ok {
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Unauthorized")
		return
	}
	if err := h.service.CancelTask(c.Request.Context(), taskID, subject.UserID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": taskID, "status": service.EmailBroadcastStatusCanceled})
}

func (h *EmailBroadcastHandler) RetryFailed(c *gin.Context) {
	taskID, ok := emailBroadcastTaskID(c)
	if !ok {
		return
	}
	count, err := h.service.RetryFailed(c.Request.Context(), taskID)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, gin.H{"id": taskID, "retried_recipients": count})
}

func emailBroadcastTaskID(c *gin.Context) (int64, bool) {
	taskID, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || taskID <= 0 {
		response.BadRequest(c, "Invalid task id")
		return 0, false
	}
	return taskID, true
}
