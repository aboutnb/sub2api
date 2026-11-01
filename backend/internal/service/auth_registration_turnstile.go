package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
	"time"
)

type registrationTurnstileKey struct{}

// The permit is only minted after Siteverify succeeds, or restored from an
// unexpired, source-bound email code issued after successful verification.
// It is never read from client JSON, headers, or an OAuth identity claim.
func (s *AuthService) registrationTurnstileBinding(ctx context.Context) (string, error) {
	if s == nil || s.settingService == nil {
		return "", ErrServiceUnavailable
	}
	cfg, err := s.settingService.GetCaptchaProviderConfig(ctx)
	if err != nil {
		return "", ErrServiceUnavailable
	}
	if !cfg.TurnstileEnabled {
		if s.cfg != nil && s.cfg.Server.Mode == "release" && s.cfg.Turnstile.Required && !cfg.Tencent.Enabled && !cfg.Aliyun.Enabled {
			return "", ErrTurnstileNotConfigured
		}
		return "", nil
	}
	if captchaProvidersConflict(cfg.TurnstileEnabled, cfg.Tencent.Enabled, cfg.Aliyun.Enabled) {
		return "", ErrCaptchaProviderConflict
	}
	if strings.TrimSpace(cfg.TurnstileSecretKey) == "" {
		return "", ErrTurnstileNotConfigured
	}
	hash := sha256.Sum256([]byte("registration-turnstile-v1\x00" + cfg.TurnstileSecretKey))
	return hex.EncodeToString(hash[:]), nil
}

// RegistrationNeedsTurnstile also routes configuration failures to the
// verification form rather than allowing automatic account creation.
func (s *AuthService) RegistrationNeedsTurnstile(ctx context.Context) bool {
	binding, err := s.registrationTurnstileBinding(ctx)
	return err != nil || binding != ""
}

func withRegistrationTurnstile(ctx context.Context, binding string) context.Context {
	ctx = context.WithValue(ctx, registrationTurnstileKey{}, binding)
	if source, ok := registrationVerificationContextFrom(ctx); ok {
		source.TurnstileBinding = binding
		ctx = WithRegistrationVerificationContext(ctx, source)
	}
	return ctx
}

func (s *AuthService) requireRegistrationTurnstile(ctx context.Context) error {
	binding, err := s.registrationTurnstileBinding(ctx)
	if err != nil {
		return err
	}
	if binding == "" {
		return nil
	}
	permit, _ := ctx.Value(registrationTurnstileKey{}).(string)
	if subtle.ConstantTimeCompare([]byte(permit), []byte(binding)) != 1 {
		return ErrTurnstileVerificationFailed
	}
	return nil
}

func (s *AuthService) VerifyRegistrationCaptcha(ctx context.Context, proof CaptchaProof, remoteIP string) (context.Context, error) {
	binding, err := s.registrationTurnstileBinding(ctx)
	if err != nil {
		return ctx, err
	}
	if binding != "" {
		// Re-read only to obtain the secret, and verify the exact policy that
		// minted the binding. A concurrent disable must never mint a permit.
		cfg, err := s.settingService.GetCaptchaProviderConfig(ctx)
		if err != nil {
			return ctx, ErrServiceUnavailable
		}
		hash := sha256.Sum256([]byte("registration-turnstile-v1\x00" + cfg.TurnstileSecretKey))
		if !cfg.TurnstileEnabled || hex.EncodeToString(hash[:]) != binding || s.turnstileService == nil {
			return ctx, ErrTurnstileNotConfigured
		}
		if err := s.turnstileService.VerifyTokenWithSecret(ctx, cfg.TurnstileSecretKey, proof.TurnstileToken, remoteIP); err != nil {
			return ctx, err
		}
	} else {
		if err := s.VerifyCaptcha(ctx, proof, remoteIP); err != nil {
			return ctx, err
		}
	}
	return withRegistrationTurnstile(ctx, binding), nil
}

// VerifyRegistrationSubmission preserves the one-use token email flow, but a
// nonempty email code alone is never sufficient proof of a completed CAPTCHA.
func (s *AuthService) VerifyRegistrationSubmission(ctx context.Context, proof CaptchaProof, remoteIP, email, code string) (context.Context, error) {
	if !s.IsEmailVerifyEnabled(ctx) || strings.TrimSpace(code) == "" {
		return s.VerifyRegistrationCaptcha(ctx, proof, remoteIP)
	}
	binding, err := s.registrationTurnstileBinding(ctx)
	if err != nil {
		return ctx, err
	}
	if binding == "" {
		return ctx, nil
	}
	if s.emailService == nil || s.emailService.cache == nil {
		return ctx, ErrServiceUnavailable
	}
	data, err := s.emailService.cache.GetVerificationCode(ctx, strings.TrimSpace(strings.ToLower(email)))
	if err != nil || data == nil {
		return ctx, ErrInvalidVerifyCode
	}
	if data.Attempts >= maxVerifyCodeAttempts {
		return ctx, ErrVerifyCodeMaxAttempts
	}
	if !data.ExpiresAt.After(time.Now()) {
		return ctx, ErrInvalidVerifyCode
	}
	if subtle.ConstantTimeCompare([]byte(data.TurnstileBinding), []byte(binding)) != 1 {
		return ctx, ErrTurnstileVerificationFailed
	}
	if err := verifyRegistrationCodeContext(ctx, data, email); err != nil {
		return ctx, err
	}
	if subtle.ConstantTimeCompare([]byte(data.Code), []byte(code)) != 1 {
		// Keep the existing failed-attempt counter. Even if the cached code
		// changes during this check, do not mint a permit from this attempt.
		if err := s.emailService.VerifyCode(ctx, email, code); err != nil {
			return ctx, err
		}
		return ctx, ErrInvalidVerifyCode
	}
	// RegisterWithVerification still consumes the code before creating a user.
	return withRegistrationTurnstile(ctx, binding), nil
}
