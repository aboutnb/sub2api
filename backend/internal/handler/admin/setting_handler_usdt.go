package admin

import (
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type testUSDTPaymentRequest struct {
	APIBase                  string   `json:"api_base"`
	PublicBaseURL            string   `json:"public_base_url"`
	PublicCallbackBaseURL    string   `json:"public_callback_base_url"`
	KeyID                    string   `json:"key_id"`
	APISecret                string   `json:"api_secret"`
	Fiat                     string   `json:"fiat"`
	EnabledNetworks          []string `json:"enabled_networks"`
	OrderTimeoutSeconds      int      `json:"order_timeout_seconds"`
	LatePaymentWindowMinutes int      `json:"late_payment_window_minutes"`
	RequestTimeoutSeconds    int      `json:"request_timeout_seconds"`
	ReconcileIntervalSeconds int      `json:"reconcile_interval_seconds"`
	ReconcileBatchSize       int      `json:"reconcile_batch_size"`
	WebhookClockSkewSeconds  int      `json:"webhook_clock_skew_seconds"`
}

func (h *SettingHandler) TestUSDTPaymentConnection(c *gin.Context) {
	if h.usdtPaymentSettingsService == nil {
		response.ErrorFrom(c, infraerrors.ServiceUnavailable("USDT_SETTINGS_UNAVAILABLE", "USDT payment settings service is unavailable"))
		return
	}
	var req testUSDTPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid USDT test request: "+err.Error())
		return
	}
	result, err := h.usdtPaymentSettingsService.TestConnection(c.Request.Context(), service.USDTPaymentAdminSettings{
		Enabled: true, APIBase: req.APIBase, PublicBaseURL: req.PublicBaseURL,
		PublicCallbackBaseURL: req.PublicCallbackBaseURL, KeyID: req.KeyID, APISecret: req.APISecret,
		Fiat: req.Fiat, EnabledNetworks: req.EnabledNetworks,
		OrderTimeoutSeconds: req.OrderTimeoutSeconds, LatePaymentWindowMinutes: req.LatePaymentWindowMinutes,
		RequestTimeoutSeconds: req.RequestTimeoutSeconds, ReconcileIntervalSeconds: req.ReconcileIntervalSeconds,
		ReconcileBatchSize: req.ReconcileBatchSize, WebhookClockSkewSeconds: req.WebhookClockSkewSeconds,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}
