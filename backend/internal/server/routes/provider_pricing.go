package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterProviderPricingRoutes registers the public Hvoy pricing contract.
func RegisterProviderPricingRoutes(r *gin.Engine, h *handler.Handlers, panelRateLimiter *middleware.PanelRateLimiter) {
	provider := r.Group("/api/provider")
	if panelRateLimiter != nil {
		provider.Use(panelRateLimiter.PublicIP())
	}
	provider.GET("/pricing", h.ProviderPricing.Get)
}
