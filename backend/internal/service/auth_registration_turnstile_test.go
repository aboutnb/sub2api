//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func registrationTurnstileSettings() map[string]string {
	return map[string]string{
		SettingKeyRegistrationEnabled: "true", SettingKeyTurnstileEnabled: "true",
		SettingKeyTurnstileSecretKey: "secret", SettingKeyEmailVerifyEnabled: "true",
	}
}

func TestRegistrationTurnstileGuardsAllAccountCreationServices(t *testing.T) {
	ctx := context.Background()
	for _, path := range []string{"password", "legacy_oauth", "oauth", "verified_email", "email_code", "auto_email", "creation_guard"} {
		t.Run(path, func(t *testing.T) {
			repo := &userRepoStub{getByEmailErr: ErrUserNotFound}
			svc := newAuthService(repo, registrationTurnstileSettings(), &emailCacheStub{}, nil)
			svc.refreshTokenCache = &refreshTokenCacheStub{}
			var err error
			switch path {
			case "password":
				_, _, err = svc.RegisterWithVerification(ctx, "new@example.com", "password", "123456", "", "", "")
			case "legacy_oauth":
				_, _, err = svc.LoginOrRegisterOAuth(ctx, "new@example.com", "new")
			case "oauth":
				_, _, err = svc.LoginOrRegisterOAuthWithTokenPair(ctx, "new@example.com", "new", "", "", "linuxdo")
			case "verified_email":
				_, _, err = svc.RegisterVerifiedOAuthEmailAccount(ctx, "new@example.com", "password", "", "github")
			case "email_code":
				_, _, err = svc.RegisterOAuthEmailAccount(ctx, "new@example.com", "password", "123456", "", "oidc")
			case "auto_email":
				_, err = svc.createEmailOAuthUser(ctx, "new@example.com", "new", "oidc", "", "")
			case "creation_guard":
				err = svc.createUserWithRegistrationEmailGuard(ctx, &User{Email: "new@example.com"})
			}
			require.ErrorIs(t, err, ErrTurnstileVerificationFailed)
			require.Empty(t, repo.created)
		})
	}
}

func TestRegistrationTurnstileEmailProofCannotBeSkipped(t *testing.T) {
	for _, scenario := range []string{"valid", "missing", "unverified", "wrong_secret", "expired", "wrong_ua", "wrong_code", "attempts"} {
		t.Run(scenario, func(t *testing.T) {
			repo := &userRepoStub{nextID: 1}
			cache := &emailCacheStub{}
			svc := newAuthService(repo, registrationTurnstileSettings(), cache, nil)
			verifier := &turnstileVerifierSpy{}
			svc.turnstileService = NewTurnstileService(svc.settingService, verifier)
			ctx := WithRegistrationVerificationContext(context.Background(), RegistrationVerificationContext{Action: "send_verify_code", UserAgentHash: "ua"})
			verified, err := svc.VerifyRegistrationCaptcha(ctx, CaptchaProof{TurnstileToken: "good-token"}, "203.0.113.1")
			require.NoError(t, err)
			source, ok := registrationVerificationContextFrom(verified)
			require.True(t, ok)
			require.NotEmpty(t, source.TurnstileBinding)
			cache.data = &VerificationCodeData{Code: "123456", ExpiresAt: time.Now().Add(time.Minute), RegistrationAction: "send_verify_code", RegistrationUserAgentHash: "ua", TurnstileBinding: source.TurnstileBinding}
			code := "123456"
			switch scenario {
			case "missing":
				cache.data = nil
			case "unverified":
				cache.data.TurnstileBinding = ""
			case "wrong_secret":
				cache.data.TurnstileBinding = "different-secret"
			case "expired":
				cache.data.ExpiresAt = time.Now().Add(-time.Second)
			case "wrong_ua":
				cache.data.RegistrationUserAgentHash = "other"
			case "wrong_code":
				code = "000000"
			case "attempts":
				cache.data.Attempts = maxVerifyCodeAttempts
			}
			ctx = WithRegistrationVerificationContext(context.Background(), RegistrationVerificationContext{Action: "register", UserAgentHash: "ua"})
			ctx, err = svc.VerifyRegistrationSubmission(ctx, CaptchaProof{}, "203.0.113.1", "new@example.com", code)
			if err == nil {
				_, _, err = svc.RegisterWithVerification(ctx, "new@example.com", "password", code, "", "", "")
			}
			if scenario == "valid" {
				require.NoError(t, err)
				require.Len(t, repo.created, 1)
			} else {
				require.Error(t, err)
				require.Empty(t, repo.created)
			}
			require.Equal(t, 1, verifier.called, "email flow must not reuse a consumed token")
		})
	}
}

func TestRegistrationTurnstileDoesNotBlockExistingOAuthLogin(t *testing.T) {
	user := &User{ID: 7, Email: "existing@example.com", Role: RoleUser, Status: StatusActive}
	repo := &userRepoStub{user: user}
	svc := newAuthService(repo, registrationTurnstileSettings(), nil, nil)
	_, got, err := svc.LoginOrRegisterOAuth(context.Background(), user.Email, "existing")
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)
	require.Empty(t, repo.created)
}

func TestRegistrationTurnstilePermitSurvivesEmailQueueContext(t *testing.T) {
	svc := newAuthService(&userRepoStub{}, registrationTurnstileSettings(), nil, nil)
	svc.turnstileService = NewTurnstileService(svc.settingService, &turnstileVerifierSpy{})
	ctx := WithRegistrationVerificationContext(context.Background(), RegistrationVerificationContext{Action: "send_verify_code", UserAgentHash: "ua"})
	ctx, err := svc.VerifyRegistrationCaptcha(ctx, CaptchaProof{TurnstileToken: "token"}, "203.0.113.1")
	require.NoError(t, err)
	queue := &EmailQueueService{taskChan: make(chan EmailTask, 1)}
	require.NoError(t, queue.EnqueueVerifyCodeWithContext(ctx, "new@example.com", "test"))
	task := <-queue.taskChan
	require.NotNil(t, task.RegistrationVerificationCtx)
	require.NotEmpty(t, task.RegistrationVerificationCtx.TurnstileBinding)
}

func TestRegistrationTurnstileAcceptsVerifiedDirectRegistration(t *testing.T) {
	settings := registrationTurnstileSettings()
	settings[SettingKeyEmailVerifyEnabled] = "false"
	repo := &userRepoStub{nextID: 1}
	svc := newAuthService(repo, settings, nil, nil)
	svc.turnstileService = NewTurnstileService(svc.settingService, &turnstileVerifierSpy{})
	ctx, err := svc.VerifyRegistrationSubmission(context.Background(), CaptchaProof{TurnstileToken: "token"}, "203.0.113.1", "new@example.com", "")
	require.NoError(t, err)
	_, user, err := svc.RegisterWithVerification(ctx, "new@example.com", "password", "", "", "", "")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Len(t, repo.created, 1)
}

func TestRegistrationTurnstileRejectsSettingsFailureAndStalePermit(t *testing.T) {
	repo := &settingRepoStub{values: registrationTurnstileSettings()}
	svc := newAuthServiceForCaptchaRepoTest(repo, false, &turnstileVerifierSpy{}, nil)
	ctx, err := svc.VerifyRegistrationCaptcha(context.Background(), CaptchaProof{TurnstileToken: "token"}, "203.0.113.1")
	require.NoError(t, err)
	require.NoError(t, svc.requireRegistrationTurnstile(ctx))
	repo.values[SettingKeyTurnstileSecretKey] = "rotated-secret"
	require.ErrorIs(t, svc.requireRegistrationTurnstile(ctx), ErrTurnstileVerificationFailed)
	repo.err = errors.New("settings offline")
	require.ErrorIs(t, svc.requireRegistrationTurnstile(ctx), ErrServiceUnavailable)
}
