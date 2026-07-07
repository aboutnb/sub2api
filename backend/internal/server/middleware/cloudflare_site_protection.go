package middleware

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	iputil "github.com/Wei-Shaw/sub2api/internal/pkg/ip"

	"github.com/gin-gonic/gin"
)

var (
	defaultCloudflareProtectedPrefixes = []string{"/"}
	defaultCloudflareBypassPaths       = []string{"/health"}
	defaultCloudflareBypassPrefixes    = []string{
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
)

// CloudflareSiteProtection rejects direct-to-origin site requests that do not
// look like they came through Cloudflare or an explicitly trusted reverse proxy.
func CloudflareSiteProtection(cfg *config.Config) gin.HandlerFunc {
	if cfg == nil || !cfg.Security.CloudflareSiteProtection.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	siteCfg := cfg.Security.CloudflareSiteProtection
	requiredHeaders := normalizeMiddlewareStringSlice(siteCfg.RequiredHeaders, nil)
	secretHeader := strings.TrimSpace(siteCfg.RequiredSecretHeader)
	secretValue := strings.TrimSpace(siteCfg.RequiredSecretValue)
	trustedProxyRules := iputil.CompileIPRules(siteCfg.TrustedProxyCIDRs)
	protectedPrefixes := normalizeMiddlewareStringSlice(siteCfg.ProtectedPrefixes, defaultCloudflareProtectedPrefixes)
	bypassPaths := normalizeMiddlewareStringSlice(siteCfg.BypassPaths, defaultCloudflareBypassPaths)
	bypassPrefixes := normalizeMiddlewareStringSlice(siteCfg.BypassPrefixes, defaultCloudflareBypassPrefixes)

	return func(c *gin.Context) {
		if !shouldProtectCloudflareSitePath(c.Request.URL.Path, protectedPrefixes, bypassPaths, bypassPrefixes) {
			c.Next()
			return
		}

		if trustedProxyRules.PatternCount > 0 {
			allowed, _ := iputil.CheckIPRestrictionWithCompiledRules(requestRemoteIP(c), trustedProxyRules, nil)
			if !allowed {
				abortCloudflareSiteProtection(c)
				return
			}
		}

		if secretHeader != "" {
			got := strings.TrimSpace(c.GetHeader(secretHeader))
			if subtle.ConstantTimeCompare([]byte(got), []byte(secretValue)) != 1 {
				abortCloudflareSiteProtection(c)
				return
			}
		}

		if !hasRequiredCloudflareHeaders(c, requiredHeaders) {
			abortCloudflareSiteProtection(c)
			return
		}

		c.Next()
	}
}

func shouldProtectCloudflareSitePath(path string, protectedPrefixes, bypassPaths, bypassPrefixes []string) bool {
	if path == "" {
		path = "/"
	}
	for _, bypassPath := range bypassPaths {
		if path == bypassPath {
			return false
		}
	}
	for _, prefix := range bypassPrefixes {
		if pathHasPrefixBoundary(path, prefix) {
			return false
		}
	}
	for _, prefix := range protectedPrefixes {
		if pathHasPrefixBoundary(path, prefix) {
			return true
		}
	}
	return false
}

func pathHasPrefixBoundary(path, prefix string) bool {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return false
	}
	if prefix == "/" {
		return true
	}
	prefix = strings.TrimRight(prefix, "/")
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func hasRequiredCloudflareHeaders(c *gin.Context, requiredHeaders []string) bool {
	for _, header := range requiredHeaders {
		value := strings.TrimSpace(c.GetHeader(header))
		if value == "" {
			return false
		}
		if strings.EqualFold(header, "CF-Connecting-IP") && net.ParseIP(value) == nil {
			return false
		}
	}
	return true
}

func requestRemoteIP(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	remoteAddr := strings.TrimSpace(c.Request.RemoteAddr)
	if remoteAddr == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return strings.Trim(host, "[]")
	}
	return strings.Trim(remoteAddr, "[]")
}

func abortCloudflareSiteProtection(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"code":    "CLOUDFLARE_SITE_PROTECTION_REQUIRED",
		"message": "Cloudflare site protection is required",
	})
}

func normalizeMiddlewareStringSlice(values, defaults []string) []string {
	if len(values) == 0 {
		return defaults
	}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return defaults
	}
	return normalized
}
