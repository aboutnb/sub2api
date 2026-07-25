package admin

import (
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type CheckinHandler struct {
	service *service.AdminCheckinService
}

func NewCheckinHandler(checkinService *service.AdminCheckinService) *CheckinHandler {
	return &CheckinHandler{service: checkinService}
}

func (h *CheckinHandler) GetConfig(c *gin.Context) {
	config, err := h.service.Config(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, config)
}

type updateCheckinConfigRequest struct {
	Enabled            bool   `json:"enabled"`
	NormalMin          string `json:"normal_min" binding:"required"`
	NormalMax          string `json:"normal_max" binding:"required"`
	LuckyMinMultiply   string `json:"lucky_min_multiplier" binding:"required"`
	LuckyMaxMultiply   string `json:"lucky_max_multiplier" binding:"required"`
	RiskEnabled        bool   `json:"risk_control_enabled"`
	MinAccountAgeHours int    `json:"min_account_age_hours"`
	IPWindowMinutes    int    `json:"ip_window_minutes"`
	IPMaxUsers         int    `json:"ip_max_users"`
	ExpectedVersion    int64  `json:"expected_config_version"`
	ChangeReason       string `json:"change_reason" binding:"required"`
}

func (h *CheckinHandler) UpdateConfig(c *gin.Context) {
	var req updateCheckinConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "enabled, ranges and change_reason are required")
		return
	}
	config, err := h.service.UpdateConfig(c.Request.Context(), service.AdminCheckinConfigUpdate{
		Enabled:            req.Enabled,
		NormalMin:          req.NormalMin,
		NormalMax:          req.NormalMax,
		LuckyMinMultiply:   req.LuckyMinMultiply,
		LuckyMaxMultiply:   req.LuckyMaxMultiply,
		RiskEnabled:        req.RiskEnabled,
		MinAccountAgeHours: req.MinAccountAgeHours,
		IPWindowMinutes:    req.IPWindowMinutes,
		IPMaxUsers:         req.IPMaxUsers,
		ExpectedVersion:    req.ExpectedVersion,
		ChangeReason:       req.ChangeReason,
	})
	if err != nil {
		servermiddleware.SetAuditExtra(c, map[string]any{
			"result":     "rejected",
			"error_code": infraerrors.Reason(err),
		})
		response.ErrorFrom(c, err)
		return
	}
	servermiddleware.SetAuditExtra(c, map[string]any{
		"result":         "accepted",
		"enabled":        config.Enabled,
		"config_version": config.ConfigVersion,
	})
	response.Success(c, config)
}

func (h *CheckinHandler) GetOverview(c *gin.Context) {
	overview, err := h.service.Overview(c.Request.Context(), c.Query("date"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, overview)
}

func (h *CheckinHandler) GetRecords(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	var userID *int64
	if raw := strings.TrimSpace(c.Query("user_id")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			response.BadRequest(c, "user_id must be a positive integer")
			return
		}
		userID = &parsed
	}
	items, total, err := h.service.Records(c.Request.Context(), service.AdminCheckinRecordFilter{
		Page: page, PageSize: pageSize, Date: c.Query("date"), UserID: userID, Email: c.Query("email"),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}
