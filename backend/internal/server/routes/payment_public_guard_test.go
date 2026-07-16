package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPaymentPublicRoutesRequirePublicAccessKeyWhenEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	cfg.Security.PublicAccessGuard.Enabled = true
	cfg.Security.PublicAccessGuard.ProtectSitePublicPOST = true
	cfg.Security.PublicAccessGuard.PublishKey = "pub-test-key"

	router := gin.New()
	v1 := router.Group("/api/v1")
	RegisterPaymentRoutes(
		v1,
		nil,
		nil,
		nil,
		func(c *gin.Context) { c.Next() },
		func(c *gin.Context) { c.Next() },
		nil,
		nil,
		cfg,
		servermiddleware.RequirePublicAccessPublishKey(cfg),
	)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/payment/public/orders/verify", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}
