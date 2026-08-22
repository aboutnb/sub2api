package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// RegisterUSDTPaymentRoutes registers the isolated BEpusdt module. These routes
// do not share the RMB payment provider registry or webhook handlers.
func RegisterUSDTPaymentRoutes(
	v1 *gin.RouterGroup,
	usdtHandler *handler.USDTPaymentHandler,
	jwtAuth middleware.JWTAuthMiddleware,
	settingService *service.SettingService,
	panelRateLimiter *middleware.PanelRateLimiter,
) {
	user := v1.Group("/usdt")
	user.Use(gin.HandlerFunc(jwtAuth))
	user.Use(middleware.BackendModeUserGuard(settingService))
	user.Use(panelRateLimiter.Global())
	{
		user.GET("/config", usdtHandler.GetConfig)
		user.POST("/orders", usdtHandler.CreateOrder)
		user.GET("/orders/:id", usdtHandler.GetOrder)
	}

	v1.POST("/usdt/webhook/bepusdt", usdtHandler.Webhook)
	v1.POST("/usdt/webhook/bepusdt/native", usdtHandler.NativeWebhook)
}
