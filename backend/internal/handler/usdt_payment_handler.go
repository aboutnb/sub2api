package handler

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/usdtpayment"
	"github.com/gin-gonic/gin"
)

const maxUSDTWebhookBodySize = 1 << 20

type USDTPaymentHandler struct {
	service *usdtpayment.Service
}

func NewUSDTPaymentHandler(service *usdtpayment.Service) *USDTPaymentHandler {
	return &USDTPaymentHandler{service: service}
}

func (h *USDTPaymentHandler) GetConfig(c *gin.Context) {
	capabilities, err := h.service.Capabilities(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	result := gin.H{"enabled": h.service.Enabled(), "networks": capabilities}
	if h.service.Enabled() {
		if rate, rateErr := h.service.ExchangeRate(c.Request.Context()); rateErr == nil {
			result["rate"] = rate.Rate
			result["rate_crypto"] = rate.Crypto
			result["rate_fiat"] = rate.Fiat
			result["rate_updated_at"] = rate.UpdatedAt
		} else {
			// Keep the network health response available, but let the UI fail
			// closed until a fresh rate can be fetched for order creation.
			result["rate_error"] = "USDT/CNY rate is temporarily unavailable"
		}
	}
	response.Success(c, result)
}

func (h *USDTPaymentHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	var req usdtpayment.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.CreateOrder(
		c.Request.Context(), subject.UserID, req, c.ClientIP(), c.Request.Host,
		c.Request.Referer(), c.GetHeader("Accept-Language"),
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *USDTPaymentHandler) GetOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || orderID <= 0 {
		response.BadRequest(c, "Invalid USDT order id")
		return
	}
	result, err := h.service.GetOrder(c.Request.Context(), subject.UserID, orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *USDTPaymentHandler) Webhook(c *gin.Context) {
	raw, err := io.ReadAll(io.LimitReader(c.Request.Body, maxUSDTWebhookBodySize+1))
	if err != nil || len(raw) > maxUSDTWebhookBodySize {
		c.String(http.StatusBadRequest, "invalid body")
		return
	}
	headers := make(map[string]string, len(c.Request.Header))
	for key := range c.Request.Header {
		headers[strings.ToLower(key)] = c.GetHeader(key)
	}
	payload, err := h.service.VerifyWebhook(raw, headers, c.Request.URL.EscapedPath(), time.Now())
	if err != nil {
		c.String(http.StatusUnauthorized, "verify failed")
		return
	}
	if err := h.service.HandleWebhook(c.Request.Context(), *payload, raw); err != nil {
		c.String(http.StatusInternalServerError, "handle failed")
		return
	}
	c.String(http.StatusOK, "success")
}
