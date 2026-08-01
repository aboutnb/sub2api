package admin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type checkinDecimalString string

func (value *checkinDecimalString) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		return fmt.Errorf("check-in decimal value is required")
	}
	if strings.HasPrefix(raw, `"`) {
		var decoded string
		if err := json.Unmarshal(data, &decoded); err != nil {
			return err
		}
		*value = checkinDecimalString(decoded)
		return nil
	}
	if _, err := strconv.ParseFloat(raw, 64); err != nil {
		return fmt.Errorf("invalid check-in decimal value: %w", err)
	}
	*value = checkinDecimalString(raw)
	return nil
}

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
	Enabled                      bool                               `json:"enabled"`
	NormalEnabled                *bool                              `json:"normal_enabled"`
	LuckyEnabled                 *bool                              `json:"lucky_enabled"`
	NormalMin                    checkinDecimalString               `json:"normal_min"`
	NormalMax                    checkinDecimalString               `json:"normal_max"`
	LuckyRewardType              string                             `json:"lucky_reward_type"`
	LuckyPositiveProbability     checkinDecimalString               `json:"lucky_positive_probability"`
	LuckyMultiplierPositiveTiers []service.AdminCheckinPositiveTier `json:"lucky_multiplier_positive_tiers"`
	LuckyAmountPositiveTiers     []service.AdminCheckinPositiveTier `json:"lucky_amount_positive_tiers"`
	LuckyMinMultiply             checkinDecimalString               `json:"lucky_min_multiplier"`
	LuckyMaxMultiply             checkinDecimalString               `json:"lucky_max_multiplier"`
	LuckyAmountMin               checkinDecimalString               `json:"lucky_amount_min"`
	LuckyAmountMax               checkinDecimalString               `json:"lucky_amount_max"`
	RiskEnabled                  bool                               `json:"risk_control_enabled"`
	MinAccountAgeHours           int                                `json:"min_account_age_hours"`
	IPWindowMinutes              int                                `json:"ip_window_minutes"`
	IPMaxUsers                   int                                `json:"ip_max_users"`
	UnrechargedEnabled           bool                               `json:"unrecharged_reduction_enabled"`
	UnrechargedCheckinThreshold  int                                `json:"unrecharged_checkin_threshold"`
	UnrechargedNormalPercent     checkinDecimalString               `json:"unrecharged_normal_reward_percent"`
	ExpectedVersion              int64                              `json:"expected_config_version"`
	ChangeReason                 string                             `json:"change_reason"`
}

func (h *CheckinHandler) UpdateConfig(c *gin.Context) {
	var req updateCheckinConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorFrom(c, service.ErrCheckinConfigInput)
		return
	}
	if req.NormalEnabled == nil || req.LuckyEnabled == nil {
		response.ErrorFrom(c, service.ErrCheckinConfigInput)
		return
	}
	config, err := h.service.UpdateConfig(c.Request.Context(), service.AdminCheckinConfigUpdate{
		Enabled:                      req.Enabled,
		NormalEnabled:                *req.NormalEnabled,
		LuckyEnabled:                 *req.LuckyEnabled,
		NormalMin:                    string(req.NormalMin),
		NormalMax:                    string(req.NormalMax),
		LuckyRewardType:              req.LuckyRewardType,
		LuckyPositiveProbability:     string(req.LuckyPositiveProbability),
		LuckyMultiplierPositiveTiers: req.LuckyMultiplierPositiveTiers,
		LuckyAmountPositiveTiers:     req.LuckyAmountPositiveTiers,
		LuckyMinMultiply:             string(req.LuckyMinMultiply),
		LuckyMaxMultiply:             string(req.LuckyMaxMultiply),
		LuckyAmountMin:               string(req.LuckyAmountMin),
		LuckyAmountMax:               string(req.LuckyAmountMax),
		RiskEnabled:                  req.RiskEnabled,
		MinAccountAgeHours:           req.MinAccountAgeHours,
		IPWindowMinutes:              req.IPWindowMinutes,
		IPMaxUsers:                   req.IPMaxUsers,
		UnrechargedEnabled:           req.UnrechargedEnabled,
		UnrechargedCheckinThreshold:  req.UnrechargedCheckinThreshold,
		UnrechargedNormalPercent:     string(req.UnrechargedNormalPercent),
		ExpectedVersion:              req.ExpectedVersion,
		ChangeReason:                 req.ChangeReason,
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
		"result":                            "accepted",
		"enabled":                           config.Enabled,
		"normal_enabled":                    config.NormalEnabled,
		"lucky_enabled":                     config.LuckyEnabled,
		"unrecharged_reduction_enabled":     config.UnrechargedEnabled,
		"unrecharged_checkin_threshold":     config.UnrechargedCheckinThreshold,
		"unrecharged_normal_reward_percent": config.UnrechargedNormalPercent,
		"config_version":                    config.ConfigVersion,
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
