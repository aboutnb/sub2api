package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCloudflareSiteProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("disabled_allows_site_request", func(t *testing.T) {
		router := newCloudflareSiteProtectionTestRouter(&config.Config{})

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("enabled_blocks_site_request_without_cloudflare_headers", func(t *testing.T) {
		router := newCloudflareSiteProtectionTestRouter(defaultEnabledCloudflareSiteProtectionConfig())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

		require.Equal(t, http.StatusForbidden, w.Code)
		require.Contains(t, w.Body.String(), "CLOUDFLARE_SITE_PROTECTION_REQUIRED")
	})

	t.Run("enabled_allows_site_request_with_cloudflare_headers", func(t *testing.T) {
		router := newCloudflareSiteProtectionTestRouter(defaultEnabledCloudflareSiteProtectionConfig())
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("CF-Connecting-IP", "203.0.113.10")
		req.Header.Set("CF-Ray", "abc123-SJC")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("invalid_cf_connecting_ip_is_blocked", func(t *testing.T) {
		router := newCloudflareSiteProtectionTestRouter(defaultEnabledCloudflareSiteProtectionConfig())
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("CF-Connecting-IP", "not-an-ip")
		req.Header.Set("CF-Ray", "abc123-SJC")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("bypasses_health_api_setup_and_gateway_prefixes", func(t *testing.T) {
		router := newCloudflareSiteProtectionTestRouter(defaultEnabledCloudflareSiteProtectionConfig())

		for _, path := range []string{"/health", "/api/v1/auth/me", "/api/event_logging/batch", "/setup/status", "/v1/messages", "/v1beta/models", "/responses", "/backend-api/codex/responses", "/antigravity/v1/messages"} {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
			require.Equal(t, http.StatusNoContent, w.Code, path)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v123", nil))
		require.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("required_secret_header_can_protect_without_cf_headers", func(t *testing.T) {
		cfg := defaultEnabledCloudflareSiteProtectionConfig()
		cfg.Security.CloudflareSiteProtection.RequiredHeaders = nil
		cfg.Security.CloudflareSiteProtection.RequiredSecretHeader = "X-Origin-Guard"
		cfg.Security.CloudflareSiteProtection.RequiredSecretValue = "origin-secret"
		router := newCloudflareSiteProtectionTestRouter(cfg)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		require.Equal(t, http.StatusForbidden, w.Code)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Origin-Guard", "origin-secret")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("trusted_proxy_cidrs_check_remote_addr", func(t *testing.T) {
		cfg := defaultEnabledCloudflareSiteProtectionConfig()
		cfg.Security.CloudflareSiteProtection.RequiredHeaders = nil
		cfg.Security.CloudflareSiteProtection.TrustedProxyCIDRs = []string{"203.0.113.0/24"}
		router := newCloudflareSiteProtectionTestRouter(cfg)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "198.51.100.10:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code)

		req = httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "203.0.113.10:12345"
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNoContent, w.Code)
	})
}

func newCloudflareSiteProtectionTestRouter(cfg *config.Config) *gin.Engine {
	router := gin.New()
	router.Use(CloudflareSiteProtection(cfg))
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/") {
			c.Status(http.StatusNoContent)
			return
		}
		c.Status(http.StatusNotFound)
	})
	return router
}

func defaultEnabledCloudflareSiteProtectionConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Security.CloudflareSiteProtection.Enabled = true
	cfg.Security.CloudflareSiteProtection.RequiredHeaders = []string{"CF-Connecting-IP", "CF-Ray"}
	cfg.Security.CloudflareSiteProtection.ProtectedPrefixes = []string{"/"}
	cfg.Security.CloudflareSiteProtection.BypassPaths = []string{"/health"}
	cfg.Security.CloudflareSiteProtection.BypassPrefixes = []string{
		"/api",
		"/setup",
		"/v1",
		"/v1beta",
		"/responses",
		"/chat/completions",
		"/embeddings",
		"/images",
		"/videos",
		"/backend-api/codex",
		"/antigravity",
	}
	return cfg
}
