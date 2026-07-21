package middleware

import (
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	authIPBanTargetContextKey = "auth_ip_ban_target"
	authIPBanReasonContextKey = "auth_ip_ban_reason"
)

func SetAuthAttemptTarget(c *gin.Context, target string) {
	if c == nil {
		return
	}
	c.Set(authIPBanTargetContextKey, strings.TrimSpace(target))
}

func SetAuthAttemptFailureReason(c *gin.Context, reason string) {
	if c == nil {
		return
	}
	c.Set(authIPBanReasonContextKey, strings.TrimSpace(reason))
}

// AuthIPBan protects only the authentication routes on which it is mounted.
// API-key gateway requests never pass through this middleware.
func AuthIPBan(banService *service.AuthIPBanService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if banService == nil {
			c.Next()
			return
		}

		clientIP := SecurityClientIP(c)
		userAgent := c.Request.UserAgent()
		active, err := banService.CheckActive(c.Request.Context(), clientIP, userAgent)
		if err != nil {
			slog.Error("auth_ip_ban.check_failed", "ip", clientIP, "error", err)
			c.Next()
			return
		}
		if active != nil {
			retrySeconds := int(math.Ceil(time.Until(active.ExpiresAt).Seconds()))
			if retrySeconds < 1 {
				retrySeconds = 1
			}
			c.Header("Retry-After", strconv.Itoa(retrySeconds))
			AbortWithError(c, http.StatusTooManyRequests, "AUTH_IP_TEMPORARILY_BANNED", "登录失败次数过多，该网络地址已被临时限制，请稍后再试")
			return
		}

		c.Next()
		status := c.Writer.Status()
		target := c.GetString(authIPBanTargetContextKey)
		if status >= 200 && status < 400 {
			banService.ClearFailures(c.Request.Context(), clientIP, userAgent, target)
			return
		}
		// Client-side authentication failures indicate probing or credential errors.
		// Server failures are deliberately excluded to avoid banning users during outages.
		if status < 400 || status >= 500 {
			return
		}
		reason := c.GetString(authIPBanReasonContextKey)
		if reason == "" {
			reason = "auth_request_rejected"
		}
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		if _, err := banService.RecordFailure(
			c.Request.Context(),
			clientIP,
			userAgent,
			target,
			path,
			reason,
		); err != nil {
			slog.Error("auth_ip_ban.record_failure_failed", "ip", clientIP, "path", path, "error", err)
		}
	}
}
