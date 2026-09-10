package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type CheckinHandler struct {
	service *service.CheckinService
}

func NewCheckinHandler(checkinService *service.CheckinService) *CheckinHandler {
	return &CheckinHandler{service: checkinService}
}

func (h *CheckinHandler) GetStatus(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	status, err := h.service.Status(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, status)
}

type checkinRequest struct {
	Mode string `json:"mode"`
}

const maxCheckinRequestBodyBytes int64 = 1024

var (
	errCheckinInvalidRequest  = infraerrors.BadRequest("CHECKIN_INVALID_REQUEST", "request must contain only a valid check-in mode")
	errCheckinRequestTooLarge = infraerrors.New(http.StatusRequestEntityTooLarge, "CHECKIN_REQUEST_TOO_LARGE", "daily check-in request body is too large")
)

func (h *CheckinHandler) CheckIn(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	req, err := decodeCheckinRequest(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if err := validateCheckinIdempotencyKey(c.GetHeader("Idempotency-Key")); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Request.Header.Set("Idempotency-Key", scopedCheckinIdempotencyKey(subject.UserID, c.GetHeader("Idempotency-Key")))
	clientIP := ip.GetSecurityClientIP(c, false)
	identity := service.CheckinIdentity{IP: clientIP}
	if c.Request != nil {
		identity.UserAgent = c.Request.UserAgent()
	}
	executeUserIdempotentJSON(c, "user.checkin.claim", req, 26*time.Hour, func(ctx context.Context) (any, error) {
		record, newlyCheckedIn, checkinErr := h.service.CheckInWithIdentityAndCaptcha(ctx, subject.UserID, req.Mode, identity, checkinTurnstileToken(c))
		if checkinErr != nil {
			middleware2.SetAuditExtra(c, map[string]any{
				"result":     "rejected",
				"error_code": infraerrors.Reason(checkinErr),
			})
			return nil, checkinErr
		}
		middleware2.SetAuditExtra(c, map[string]any{"result": "accepted"})
		return gin.H{
			"newly_checked_in": newlyCheckedIn,
			"record":           record,
		}, nil
	})
}

func checkinTurnstileToken(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if token := c.GetHeader("X-Turnstile-Token"); token != "" {
		return token
	}
	return c.GetHeader("CF-Turnstile-Response")
}

func scopedCheckinIdempotencyKey(userID int64, raw string) string {
	return service.HashIdempotencyKey(strconv.FormatInt(userID, 10) + "\n" + raw)
}

func validateCheckinIdempotencyKey(raw string) error {
	key, err := service.NormalizeIdempotencyKey(raw)
	if err != nil {
		return err
	}
	if key == "" {
		return service.ErrIdempotencyKeyRequired
	}
	return nil
}

func decodeCheckinRequest(c *gin.Context) (checkinRequest, error) {
	var req checkinRequest
	if c == nil || c.Request == nil || c.Request.Body == nil {
		return req, errCheckinInvalidRequest
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCheckinRequestBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return req, errCheckinRequestTooLarge
		}
		return req, errCheckinInvalidRequest
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return req, errCheckinInvalidRequest
	}
	if req.Mode != "normal" && req.Mode != "lucky" {
		return req, service.ErrCheckinInvalidMode
	}
	return req, nil
}

func (h *CheckinHandler) GetRecords(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.service.Records(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}
