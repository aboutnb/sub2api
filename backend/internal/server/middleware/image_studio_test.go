package middleware

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
)

func TestImageStudioFrameHeadersOnlyAllowStudio(t *testing.T) {
	router := gin.New()
	router.Use(SecurityHeaders(config.CSPConfig{Enabled: true}, nil))
	router.GET("/*path", func(c *gin.Context) { c.Status(200) })
	for _, path := range []string{"/image-studio-app/", "/image-studio-app/index.html", "/image-studio", "/other"} {
		result := httptest.NewRecorder()
		router.ServeHTTP(result, httptest.NewRequest("GET", path, nil))
		if path == "/image-studio-app/" || path == "/image-studio-app/index.html" {
			require.Equal(t, "SAMEORIGIN", result.Header().Get("X-Frame-Options"))
			require.Contains(t, result.Header().Get("Content-Security-Policy"), "frame-ancestors 'self'")
			require.NotContains(t, result.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'")
		} else {
			require.Equal(t, "DENY", result.Header().Get("X-Frame-Options"))
		}
	}
}
