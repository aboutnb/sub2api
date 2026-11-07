package handler

import (
	"context"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9" //nolint:depguard // Registration failure counter test.
	"github.com/stretchr/testify/require"
)

func TestRegistrationFailurePairCountsBothDimensionsAndCaps(t *testing.T) {
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		values, err := registrationFailurePairScript.Run(ctx, rdb, []string{"ip", "identity"}, 60000, 5, 2).Int64Slice()
		require.NoError(t, err)
		require.EqualValues(t, min(i+1, 5), values[0])
		require.EqualValues(t, min(i+1, 2), values[1])
	}
	server.FastForward(time.Minute)
	values, err := registrationFailurePairScript.Run(ctx, rdb, []string{"ip", "identity"}, 60000, 5, 2).Int64Slice()
	require.NoError(t, err)
	require.Equal(t, []int64{1, 1}, values)
}

func TestRegistrationFailureClassificationExcludesSignedExpiryAndFastSubmit(t *testing.T) {
	h := newRegistrationChallengeTestHandler()
	c, _ := newRegistrationChallengeTestContext()
	submission := buildRegistrationChallengeSubmissionForTest(t, h, c, "test@example.com", "register", "")
	payload, err := h.parseRegistrationChallengeToken(submission.Token)
	require.NoError(t, err)
	payload.ExpiresAt = time.Now().Add(-time.Minute).UnixMilli()
	submission.Token, err = h.signRegistrationChallengePayload(payload)
	require.NoError(t, err)
	require.False(t, h.registrationChallengeFailureIsAbuse(submission))
	submission.TrapValue = "bot"
	require.True(t, h.registrationChallengeFailureIsAbuse(submission))
	submission.TrapValue = ""
	submission.Token += "tampered"
	require.True(t, h.registrationChallengeFailureIsAbuse(submission))
	require.True(t, h.registrationChallengeFailureIsAbuse(nil))
	payload.IssuedAt = time.Now().UnixMilli()
	payload.ExpiresAt = time.Now().Add(time.Minute).UnixMilli()
	payload.MinElapsedMS = 900
	submission.Token, err = h.signRegistrationChallengePayload(payload)
	require.NoError(t, err)
	require.False(t, h.registrationChallengeFailureIsAbuse(submission))
}

type registrationAdmissionDenyStub struct {
	service.RegistrationProtectionRepository
	emails []string
}

func (s *registrationAdmissionDenyStub) CheckEmailAdmission(_ context.Context, _ service.RegistrationSource, email string) error {
	s.emails = append(s.emails, email)
	return service.ErrRegistrationDenied
}
func TestRegistrationAdmissionRejectsBeforeCaptchaAndMail(t *testing.T) {
	for _, method := range []string{"send", "register", "oauth-send"} {
		t.Run(method, func(t *testing.T) {
			h := newRegistrationChallengeTestHandler()
			h.settingSvc = service.NewSettingService(&oauthPendingFlowSettingRepoStub{values: map[string]string{}}, h.cfg)
			repo := &registrationAdmissionDenyStub{}
			h.registrationProtection = service.NewRegistrationProtectionService(repo, h.settingSvc, nil, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/auth", strings.NewReader(`{"email":"blocked@example.com","password":"Password123!"}`))
			c.Request.Header.Set("Content-Type", "application/json")
			// authService is deliberately nil: calling CAPTCHA, mail, or registration
			// after a denial panics, so this also asserts the ordering of side effects.
			switch method {
			case "send":
				h.SendVerifyCode(c)
			case "register":
				h.Register(c)
			default:
				h.SendPendingOAuthVerifyCode(c)
			}
			require.Equal(t, 403, w.Code)
			require.Equal(t, []string{"blocked@example.com"}, repo.emails)
			require.Contains(t, w.Body.String(), "REGISTRATION_UNAVAILABLE")
			require.NotContains(t, w.Body.String(), "Retry")
			require.Empty(t, w.Header().Get("Retry-After"))
		})
	}
}
