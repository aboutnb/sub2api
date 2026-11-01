//go:build unit

package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type turnstileSecurityVerifier struct {
	result *TurnstileVerifyResponse
	err    error
	calls  int
}

func (v *turnstileSecurityVerifier) VerifyToken(context.Context, string, string, string) (*TurnstileVerifyResponse, error) {
	v.calls++
	return v.result, v.err
}

func TestTurnstileRejectsUnverifiedTokens(t *testing.T) {
	for _, tc := range []struct {
		name, token       string
		result            *TurnstileVerifyResponse
		providerErr, want error
		calls             int
	}{
		{name: "missing", want: ErrTurnstileVerificationFailed},
		{name: "blank", token: "  ", want: ErrTurnstileVerificationFailed},
		{name: "oversized", token: strings.Repeat("a", 2049), want: ErrTurnstileVerificationFailed},
		{name: "forged", token: "forged", result: &TurnstileVerifyResponse{ErrorCodes: []string{"invalid-input-response"}}, want: ErrTurnstileVerificationFailed, calls: 1},
		{name: "expired or replayed", token: "used", result: &TurnstileVerifyResponse{ErrorCodes: []string{"timeout-or-duplicate"}}, want: ErrTurnstileVerificationFailed, calls: 1},
		{name: "provider unavailable", token: "token", providerErr: errors.New("private verification URL"), want: ErrServiceUnavailable, calls: 1},
		{name: "empty provider response", token: "token", want: ErrServiceUnavailable, calls: 1},
		{name: "verified", token: "token", result: &TurnstileVerifyResponse{Success: true}, calls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			verifier := &turnstileSecurityVerifier{result: tc.result, err: tc.providerErr}
			err := NewTurnstileService(nil, verifier).VerifyTokenWithSecret(context.Background(), "secret", tc.token, "203.0.113.10")
			if tc.want == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.want)
			}
			require.Equal(t, tc.calls, verifier.calls)
		})
	}
}

func TestTurnstileSettingsFailureDoesNotDisableVerification(t *testing.T) {
	settings := NewSettingService(&settingRepoStub{err: errors.New("unavailable")}, nil)
	verifier := &turnstileSecurityVerifier{result: &TurnstileVerifyResponse{Success: true}}
	err := NewTurnstileService(settings, verifier).VerifyToken(context.Background(), "token", "203.0.113.10")
	require.ErrorIs(t, err, ErrServiceUnavailable)
	require.Zero(t, verifier.calls)
}
