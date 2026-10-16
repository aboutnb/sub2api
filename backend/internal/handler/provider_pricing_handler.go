package handler

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ProviderPricingHandler struct {
	pricingService *service.ProviderPricingService
}

func NewProviderPricingHandler(pricingService *service.ProviderPricingService) *ProviderPricingHandler {
	return &ProviderPricingHandler{pricingService: pricingService}
}

type providerPricingResponse struct {
	SchemaVersion string              `json:"schema_version"`
	Success       bool                `json:"success"`
	Message       string              `json:"message"`
	Data          providerPricingData `json:"data"`
}

type providerPricingData struct {
	Currency   string                         `json:"currency"`
	PriceUnit  string                         `json:"price_unit"`
	SiteName   string                         `json:"site_name,omitempty"`
	SiteDomain string                         `json:"site_domain,omitempty"`
	UpdatedAt  string                         `json:"updated_at"`
	Models     []service.ProviderPricingModel `json:"models"`
}

// Get serves the unauthenticated Hvoy Provider Pricing API.
func (h *ProviderPricingHandler) Get(c *gin.Context) {
	if h == nil || h.pricingService == nil {
		c.JSON(http.StatusServiceUnavailable, providerPricingResponse{
			SchemaVersion: "1.1",
			Success:       false,
			Message:       "provider pricing is unavailable",
		})
		return
	}

	snapshot, err := h.pricingService.Snapshot(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, providerPricingResponse{
			SchemaVersion: "1.1",
			Success:       false,
			Message:       "provider pricing is temporarily unavailable",
		})
		return
	}

	remaining := time.Until(snapshot.UpdatedAt.Add(service.ProviderPricingCacheTTL))
	maxAge := max(0, int(math.Ceil(remaining.Seconds())))
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	c.JSON(http.StatusOK, providerPricingResponse{
		SchemaVersion: snapshot.SchemaVersion,
		Success:       snapshot.Success,
		Message:       snapshot.Message,
		Data: providerPricingData{
			Currency:   snapshot.Currency,
			PriceUnit:  snapshot.PriceUnit,
			SiteName:   snapshot.SiteName,
			SiteDomain: snapshot.SiteDomain,
			UpdatedAt:  snapshot.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			Models:     snapshot.Models,
		},
	})
}
