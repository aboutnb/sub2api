package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRegistrationQuotaResponseDoesNotExposeSource(t *testing.T) {
	err := fmt.Errorf("create: %w", &RegistrationQuotaExceeded{Dimension: "identity_hash", RetryAfter: 1500 * time.Millisecond})
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	require.True(t, response.ErrorFrom(c, err))
	require.Equal(t, http.StatusTooManyRequests, recorder.Code)
	require.Equal(t, "2", recorder.Header().Get("Retry-After"))
	require.Contains(t, recorder.Body.String(), "REGISTRATION_SOURCE_QUOTA_EXCEEDED")
	require.NotContains(t, recorder.Body.String(), "identity_hash")
}
