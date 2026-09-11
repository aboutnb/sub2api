package service

import (
	"fmt"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestRegistrationVerificationMarkerPreservesPublicError(t *testing.T) {
	err := &VerificationFailure{Cause: ErrInvalidVerifyCode}
	require.ErrorIs(t, err, ErrInvalidVerifyCode)
	require.Equal(t, infraerrors.Code(ErrInvalidVerifyCode), infraerrors.Code(err))
	require.Equal(t, infraerrors.Message(ErrInvalidVerifyCode), infraerrors.Message(err))
	require.True(t, IsRegistrationVerificationFailure(fmt.Errorf("verify: %w", err)))
	for _, ordinary := range []error{ErrInvalidVerifyCode, ErrVerifyCodeMaxAttempts, ErrTurnstileInvalidSecretKey, ErrServiceUnavailable, ErrTurnstileVerificationFailed} {
		require.False(t, IsRegistrationVerificationFailure(ordinary))
	}
}
