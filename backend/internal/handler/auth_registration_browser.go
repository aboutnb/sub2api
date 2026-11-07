package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const registrationBrowserCookie = "registration_browser"
const registrationFeatureHeader = "X-Registration-Features"

func (h *AuthHandler) registrationBrowserSignature(value string) string {
	mac := hmac.New(sha256.New, h.registrationChallengeSecret())
	_, _ = mac.Write([]byte("registration-browser-v1\x00" + value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (h *AuthHandler) registrationBrowserHash(c *gin.Context) string {
	value, err := c.Cookie(registrationBrowserCookie)
	if err != nil || len(value) > 256 {
		return ""
	}
	parts := strings.Split(value, ".")
	if len(parts) != 3 || len(parts[0]) != 32 {
		return ""
	}
	expires, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || expires <= time.Now().Unix() || expires > time.Now().Add(31*24*time.Hour).Unix() ||
		!constantTimeStringEqual(parts[2], h.registrationBrowserSignature(parts[0]+"."+parts[1])) {
		return ""
	}
	return h.registrationSensitiveHash("registration-browser-id-v1", parts[0])
}

func (h *AuthHandler) ensureRegistrationBrowser(c *gin.Context) {
	if h.registrationBrowserHash(c) != "" {
		return
	}
	value := randomRegistrationChallengeID(24) + "." + strconv.FormatInt(time.Now().Add(30*24*time.Hour).Unix(), 10)
	value += "." + h.registrationBrowserSignature(value)
	// Request TLS can terminate at a trusted proxy; Origin is not used to trust an IP.
	secure := c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") || strings.HasPrefix(strings.ToLower(c.GetHeader("Origin")), "https://")
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(registrationBrowserCookie, value, 30*24*60*60, "/", "", secure, true)
}

func (h *AuthHandler) registrationFeatureHash(c *gin.Context) string {
	value := c.GetHeader(registrationFeatureHeader)
	if len(value) != 64 {
		return ""
	}
	if _, err := hex.DecodeString(value); err != nil {
		return ""
	}
	// Client features can be spoofed; admission combines them with server-side history.
	return h.registrationSensitiveHash("registration-features-v1", strings.ToLower(value)+"\x00"+normalizeRegistrationUserAgent(c.Request.UserAgent()))
}
