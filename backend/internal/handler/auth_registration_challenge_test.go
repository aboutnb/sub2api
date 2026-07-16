package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9" //nolint:depguard // tests wire the same Redis client accepted by auth routes.
	"github.com/stretchr/testify/require"
)

func newRegistrationChallengeTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 registration-challenge-test")
	req.Header.Set("CF-Connecting-IP", "203.0.113.42")
	ginCtx.Request = req
	return ginCtx, recorder
}

func newRegistrationChallengeTestHandler() *AuthHandler {
	return &AuthHandler{
		cfg: &config.Config{
			JWT: config.JWTConfig{
				Secret: "registration-challenge-test-secret",
			},
		},
	}
}

func attachRegistrationChallengeRedis(t *testing.T, handler *AuthHandler) {
	t.Helper()

	mr := miniredis.RunT(t)
	handler.SetRegistrationRiskRedis(redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	}))
	t.Cleanup(func() {
		if handler.redisClient != nil {
			_ = handler.redisClient.Close()
		}
	})
}

func buildRegistrationChallengeSubmissionForTest(
	t *testing.T,
	h *AuthHandler,
	c *gin.Context,
	email string,
	action string,
	trapValue string,
) *RegistrationChallengeSubmission {
	t.Helper()

	challenge, err := h.issueRegistrationChallenge(c)
	require.NoError(t, err)

	payload, err := h.parseRegistrationChallengeToken(challenge.Token)
	require.NoError(t, err)
	now := time.Now()
	payload.IssuedAt = now.Add(-registrationChallengeMinElapsed - 50*time.Millisecond).UnixMilli()
	payload.ExpiresAt = now.Add(registrationChallengeTTL).UnixMilli()
	challenge.Token, err = h.signRegistrationChallengePayload(payload)
	require.NoError(t, err)
	challenge.IssuedAt = payload.IssuedAt
	challenge.ExpiresAt = payload.ExpiresAt

	completedAt := time.Now().UnixMilli()
	return &RegistrationChallengeSubmission{
		Token:       challenge.Token,
		CompletedAt: completedAt,
		Proof: registrationChallengeProof(
			challenge.Token,
			normalizeRegistrationChallengeEmail(email),
			action,
			completedAt,
			challenge.TrapField,
			challenge.Salt,
		),
		TrapField: challenge.TrapField,
		TrapValue: trapValue,
	}
}

func newRegistrationChallengeJSONRequestForTest(
	t *testing.T,
	h *AuthHandler,
	c *gin.Context,
	path string,
	email string,
	action string,
	payload map[string]any,
) *http.Request {
	t.Helper()

	probeReq := httptest.NewRequest(http.MethodPost, path, nil)
	probeReq.Header.Set("Content-Type", "application/json")
	probeReq.Header.Set("User-Agent", "Mozilla/5.0 registration-challenge-test")
	probeReq.Header.Set("CF-Connecting-IP", "203.0.113.42")
	c.Request = probeReq

	if payload == nil {
		payload = make(map[string]any)
	}
	payload["registration_challenge"] = buildRegistrationChallengeSubmissionForTest(t, h, c, email, action, "")
	bodyBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 registration-challenge-test")
	req.Header.Set("CF-Connecting-IP", "203.0.113.42")
	return req
}

func TestRegistrationChallengeEndpointReturnsToken(t *testing.T) {
	handler := newRegistrationChallengeTestHandler()
	ginCtx, recorder := newRegistrationChallengeTestContext()
	ginCtx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/registration-challenge", nil)
	ginCtx.Request.Header.Set("User-Agent", "Mozilla/5.0 registration-challenge-test")
	ginCtx.Request.Header.Set("CF-Connecting-IP", "203.0.113.42")

	handler.GetRegistrationChallenge(ginCtx)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"token"`)
	require.Contains(t, recorder.Body.String(), `"trap_field"`)
}

func TestRequireRegistrationChallengeAcceptsValidSubmission(t *testing.T) {
	handler := newRegistrationChallengeTestHandler()
	ginCtx, _ := newRegistrationChallengeTestContext()
	submission := buildRegistrationChallengeSubmissionForTest(t, handler, ginCtx, "User@Example.com", "register", "")

	err := handler.requireRegistrationChallenge(ginCtx, "register", " user@example.com ", submission)

	require.NoError(t, err)
}

func TestRequireRegistrationChallengeAcceptsSkewedClientClock(t *testing.T) {
	handler := newRegistrationChallengeTestHandler()
	ginCtx, _ := newRegistrationChallengeTestContext()
	submission := buildRegistrationChallengeSubmissionForTest(t, handler, ginCtx, "user@example.com", "register", "")
	payload, err := handler.parseRegistrationChallengeToken(submission.Token)
	require.NoError(t, err)

	submission.CompletedAt = time.Now().Add(24 * time.Hour).UnixMilli()
	submission.Proof = registrationChallengeProof(
		submission.Token,
		"user@example.com",
		"register",
		submission.CompletedAt,
		payload.TrapField,
		payload.Salt,
	)

	require.NoError(t, handler.requireRegistrationChallenge(ginCtx, "register", "user@example.com", submission))
}

func TestRequireRegistrationChallengeRejectsReplay(t *testing.T) {
	handler := newRegistrationChallengeTestHandler()
	attachRegistrationChallengeRedis(t, handler)
	ginCtx, _ := newRegistrationChallengeTestContext()
	submission := buildRegistrationChallengeSubmissionForTest(t, handler, ginCtx, "user@example.com", "register", "")

	require.NoError(t, handler.requireRegistrationChallenge(ginCtx, "register", "user@example.com", submission))
	err := handler.requireRegistrationChallenge(ginCtx, "register", "user@example.com", submission)

	require.Error(t, err)
	require.Equal(t, "REGISTRATION_CHALLENGE_INVALID", infraerrors.Reason(err))
}

func TestRequireRegistrationChallengeRateLimitsEmail(t *testing.T) {
	handler := newRegistrationChallengeTestHandler()
	attachRegistrationChallengeRedis(t, handler)
	ginCtx, _ := newRegistrationChallengeTestContext()

	for i := 0; i < 3; i++ {
		submission := buildRegistrationChallengeSubmissionForTest(t, handler, ginCtx, "user@example.com", "send_verify_code", "")
		require.NoError(t, handler.requireRegistrationChallenge(ginCtx, "send_verify_code", "user@example.com", submission))
	}
	submission := buildRegistrationChallengeSubmissionForTest(t, handler, ginCtx, "user@example.com", "send_verify_code", "")
	err := handler.requireRegistrationChallenge(ginCtx, "send_verify_code", "user@example.com", submission)

	require.Error(t, err)
	require.Equal(t, "REGISTRATION_TOO_MANY_ATTEMPTS", infraerrors.Reason(err))
}

func TestRequireRegistrationChallengeAllowsSharedNetworkAndUserAgent(t *testing.T) {
	handler := newRegistrationChallengeTestHandler()
	attachRegistrationChallengeRedis(t, handler)
	ginCtx, _ := newRegistrationChallengeTestContext()

	for i := 0; i < 50; i++ {
		email := fmt.Sprintf("shared-network-%d@example.com", i)
		submission := buildRegistrationChallengeSubmissionForTest(t, handler, ginCtx, email, "send_verify_code", "")
		require.NoError(t, handler.requireRegistrationChallenge(ginCtx, "send_verify_code", email, submission))
	}
}

func TestRequireRegistrationChallengeRejectsMissingSubmission(t *testing.T) {
	handler := newRegistrationChallengeTestHandler()
	ginCtx, _ := newRegistrationChallengeTestContext()

	err := handler.requireRegistrationChallenge(ginCtx, "register", "user@example.com", nil)

	require.Error(t, err)
	require.Equal(t, "REGISTRATION_CHALLENGE_REQUIRED", infraerrors.Reason(err))
	require.Equal(t, "注册验证缺失，请重试", infraerrors.Message(err))
}

func TestRequireRegistrationChallengeRejectsTrapValue(t *testing.T) {
	handler := newRegistrationChallengeTestHandler()
	ginCtx, _ := newRegistrationChallengeTestContext()
	submission := buildRegistrationChallengeSubmissionForTest(t, handler, ginCtx, "user@example.com", "register", "bot-filled-value")

	err := handler.requireRegistrationChallenge(ginCtx, "register", "user@example.com", submission)

	require.Error(t, err)
	require.Equal(t, "REGISTRATION_CHALLENGE_INVALID", infraerrors.Reason(err))
}

func TestRequireRegistrationChallengeRejectsFingerprintMismatch(t *testing.T) {
	handler := newRegistrationChallengeTestHandler()
	ginCtx, _ := newRegistrationChallengeTestContext()
	submission := buildRegistrationChallengeSubmissionForTest(t, handler, ginCtx, "user@example.com", "register", "")
	ginCtx.Request.Header.Set("User-Agent", "curl/8.0")

	err := handler.requireRegistrationChallenge(ginCtx, "register", "user@example.com", submission)

	require.Error(t, err)
	require.Equal(t, "REGISTRATION_CHALLENGE_INVALID", infraerrors.Reason(err))
}
