package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newAuthRoutesTestRouter(redisClient *redis.Client) *gin.Engine {
	return newAuthRoutesTestRouterWithConfig(redisClient, &config.Config{}, nil)
}

func newAuthRoutesTestRouterWithConfig(redisClient *redis.Client, cfg *config.Config, publicAccessGuard gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")

	RegisterAuthRoutes(
		v1,
		&handler.Handlers{
			Auth:    &handler.AuthHandler{},
			Setting: &handler.SettingHandler{},
		},
		servermiddleware.JWTAuthMiddleware(func(c *gin.Context) {
			c.Next()
		}),
		redisClient,
		nil,
		cfg,
		publicAccessGuard,
	)

	return router
}

func TestAuthRoutesRateLimitFailCloseWhenRedisUnavailable(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  50 * time.Millisecond,
		ReadTimeout:  50 * time.Millisecond,
		WriteTimeout: 50 * time.Millisecond,
	})
	t.Cleanup(func() {
		_ = rdb.Close()
	})

	router := newAuthRoutesTestRouter(rdb)
	paths := []string{
		"/api/v1/auth/register",
		"/api/v1/auth/login",
		"/api/v1/auth/login/2fa",
		"/api/v1/auth/send-verify-code",
		"/api/v1/auth/oauth/pending/send-verify-code",
	}

	for _, path := range paths {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "203.0.113.10:12345"

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusTooManyRequests, w.Code, "path=%s", path)
		require.Contains(t, w.Body.String(), "rate limit exceeded", "path=%s", path)
	}
}

func TestAuthRoutesPublicAccessGuardProtectsOnlyPublicPOST(t *testing.T) {
	cfg := &config.Config{}
	cfg.Security.PublicAccessGuard.Enabled = true
	cfg.Security.PublicAccessGuard.ProtectSitePublicPOST = true
	cfg.Security.PublicAccessGuard.PublishKey = "pub-test-key"
	cfg.Security.PublicAccessGuard.HeaderName = "x-sub2api-publish-key"

	router := newAuthRoutesTestRouterWithConfig(nil, cfg, servermiddleware.RequirePublicAccessPublishKey(cfg))

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/oauth/linuxdo/start", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)
	require.NotEqual(t, http.StatusUnauthorized, getW.Code)

	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(`{}`))
	postReq.Header.Set("Content-Type", "application/json")
	postW := httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)
	require.Equal(t, http.StatusUnauthorized, postW.Code)

	postReq = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", strings.NewReader(`{}`))
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("x-sub2api-publish-key", "pub-test-key")
	postW = httptest.NewRecorder()
	router.ServeHTTP(postW, postReq)
	require.NotEqual(t, http.StatusUnauthorized, postW.Code)
}
