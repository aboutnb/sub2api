package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDecodeCheckinRequestAcceptsOnlyOneStrictObject(t *testing.T) {
	testCases := []struct {
		name       string
		body       string
		wantMode   string
		wantReason string
	}{
		{name: "normal", body: `{"mode":"normal"}`, wantMode: "normal"},
		{name: "lucky", body: `{"mode":"lucky"}`, wantMode: "lucky"},
		{name: "unknown field", body: `{"mode":"normal","amount":100}`, wantReason: "CHECKIN_INVALID_REQUEST"},
		{name: "concatenated objects", body: `{"mode":"normal"}{"mode":"lucky"}`, wantReason: "CHECKIN_INVALID_REQUEST"},
		{name: "invalid mode", body: `{"mode":"admin"}`, wantReason: "CHECKIN_INVALID_MODE"},
		{name: "empty", body: ``, wantReason: "CHECKIN_INVALID_REQUEST"},
		{name: "too large", body: `{"mode":"normal","padding":"` + strings.Repeat("x", 2048) + `"}`, wantReason: "CHECKIN_REQUEST_TOO_LARGE"},
	}

	gin.SetMode(gin.TestMode)
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest("POST", "/api/v1/user/checkin", strings.NewReader(testCase.body))

			request, err := decodeCheckinRequest(ctx)

			if testCase.wantReason == "" {
				require.NoError(t, err)
				require.Equal(t, testCase.wantMode, request.Mode)
				return
			}
			require.Equal(t, testCase.wantReason, infraerrors.Reason(err))
		})
	}
}

func TestValidateCheckinIdempotencyKey(t *testing.T) {
	require.NoError(t, validateCheckinIdempotencyKey("checkin-2026-07-26-request-1"))
	require.Equal(t, "IDEMPOTENCY_KEY_REQUIRED", infraerrors.Reason(validateCheckinIdempotencyKey("")))
	require.Equal(t, "IDEMPOTENCY_KEY_INVALID", infraerrors.Reason(validateCheckinIdempotencyKey(strings.Repeat("x", 129))))
}

func TestScopedCheckinIdempotencyKeySeparatesUsers(t *testing.T) {
	first := scopedCheckinIdempotencyKey(7, "same-client-key")
	require.Equal(t, first, scopedCheckinIdempotencyKey(7, "same-client-key"))
	require.NotEqual(t, first, scopedCheckinIdempotencyKey(8, "same-client-key"))
	require.Len(t, first, 64)
}

func TestCheckinTurnstileTokenReadsDedicatedHeaderWithCompatibilityFallback(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("POST", "/api/v1/user/checkin", nil)
	ctx.Request.Header.Set("CF-Turnstile-Response", "fallback-token")
	require.Equal(t, "fallback-token", checkinTurnstileToken(ctx))

	ctx.Request.Header.Set("X-Turnstile-Token", "dedicated-token")
	require.Equal(t, "dedicated-token", checkinTurnstileToken(ctx))
}
