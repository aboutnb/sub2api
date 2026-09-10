package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"

	"github.com/gin-gonic/gin"
)

const defaultPublicAccessGuardHeader = "x-sub2api-publish-key"

// PublicAccessPublishKey returns the configured public access publish key.
// This value is deliberately public: it is only a noise filter, not an auth secret.
func PublicAccessPublishKey(cfg *config.Config) string {
	if cfg == nil || !cfg.Security.PublicAccessGuard.Enabled {
		return ""
	}
	return strings.TrimSpace(cfg.Security.PublicAccessGuard.PublishKey)
}

func PublicAccessHeaderName(cfg *config.Config) string {
	if cfg == nil {
		return defaultPublicAccessGuardHeader
	}
	header := strings.TrimSpace(cfg.Security.PublicAccessGuard.HeaderName)
	if header == "" {
		return defaultPublicAccessGuardHeader
	}
	return header
}

func RequirePublicAccessPublishKey(cfg *config.Config) gin.HandlerFunc {
	expected := PublicAccessPublishKey(cfg)
	headerName := PublicAccessHeaderName(cfg)

	return func(c *gin.Context) {
		if expected == "" {
			c.Next()
			return
		}
		got := strings.TrimSpace(c.GetHeader(headerName))
		if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "PUBLIC_ACCESS_KEY_REQUIRED",
				"message": "Public access key is required",
			})
			return
		}
		c.Next()
	}
}

func RejectMalformedGatewayAPIKey(cfg *config.Config, writeError GatewayErrorWriter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg == nil ||
			!cfg.Security.PublicAccessGuard.Enabled ||
			!cfg.Security.PublicAccessGuard.RejectMalformedGatewayKeys {
			c.Next()
			return
		}

		apiKey := extractGatewayAPIKeyCandidate(c)
		if apiKey == "" || looksLikeGatewayAPIKey(cfg, apiKey) {
			c.Next()
			return
		}

		writeError(c, http.StatusUnauthorized, "Invalid API key")
		c.Abort()
	}
}

func extractGatewayAPIKeyCandidate(c *gin.Context) string {
	if c == nil {
		return ""
	}

	if auth := strings.TrimSpace(c.GetHeader("Authorization")); auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			if key := strings.TrimSpace(parts[1]); key != "" {
				return key
			}
		}
	}

	for _, header := range []string{"x-api-key", "x-goog-api-key"} {
		if key := strings.TrimSpace(c.GetHeader(header)); key != "" {
			return key
		}
	}

	if key := strings.TrimSpace(c.Query("key")); key != "" {
		return key
	}
	if key := strings.TrimSpace(c.Query("api_key")); key != "" {
		return key
	}
	return ""
}

func looksLikeGatewayAPIKey(cfg *config.Config, key string) bool {
	key = strings.TrimSpace(key)
	if len(key) < 16 {
		return false
	}
	if !hasAllowedAPIKeyPrefix(cfg, key) {
		return false
	}
	for _, r := range key {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}

func hasAllowedAPIKeyPrefix(cfg *config.Config, key string) bool {
	prefixes := cfg.Security.PublicAccessGuard.GatewayKeyAllowedPrefixes
	if p := strings.TrimSpace(cfg.Default.APIKeyPrefix); p != "" {
		prefixes = append(prefixes, p)
	}
	if len(prefixes) == 0 {
		prefixes = []string{"sk-"}
	}
	for _, prefix := range prefixes {
		if prefix != "" && strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}
