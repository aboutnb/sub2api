package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRequirePublicAccessPublishKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("disabled_allows_request", func(t *testing.T) {
		router := gin.New()
		router.Use(RequirePublicAccessPublishKey(&config.Config{}))
		router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("enabled_requires_publish_key_header", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.Security.PublicAccessGuard.Enabled = true
		cfg.Security.PublicAccessGuard.PublishKey = "pub-test-key"
		cfg.Security.PublicAccessGuard.HeaderName = "x-site-publish-key"

		router := gin.New()
		router.Use(RequirePublicAccessPublishKey(cfg))
		router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		require.Equal(t, http.StatusUnauthorized, w.Code)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("x-site-publish-key", "pub-test-key")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestRejectMalformedGatewayAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	cfg.Security.PublicAccessGuard.Enabled = true
	cfg.Security.PublicAccessGuard.RejectMalformedGatewayKeys = true
	cfg.Security.PublicAccessGuard.GatewayKeyAllowedPrefixes = []string{"sk-"}

	t.Run("malformed_key_aborts_before_next_handler", func(t *testing.T) {
		called := false
		router := gin.New()
		router.Use(RejectMalformedGatewayAPIKey(cfg, AnthropicErrorWriter))
		router.GET("/", func(c *gin.Context) {
			called = true
			c.Status(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer junk")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusUnauthorized, w.Code)
		require.False(t, called)
	})

	t.Run("well_formed_key_allows_next_handler", func(t *testing.T) {
		router := gin.New()
		router.Use(RejectMalformedGatewayAPIKey(cfg, AnthropicErrorWriter))
		router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer sk-1234567890abcdef")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNoContent, w.Code)
	})
}
