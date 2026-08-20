package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestSignupRiskIdentityFailsClosedWhenGrantStoreIsMissing(t *testing.T) {
	svc := &AuthService{cfg: &config.Config{Default: config.DefaultConfig{UserBalance: 5, UserConcurrency: 2}}}
	ctx := WithSignupRiskIdentity(context.Background(), "risk-fingerprint")

	grant := svc.prepareSignupGrant(ctx, "email")
	require.True(t, grant.deferred)
	require.Zero(t, grant.initialBalance())
	require.Zero(t, grant.initialConcurrency())

	allowed, err := svc.applySignupGrant(ctx, &User{ID: 42}, grant)
	require.ErrorIs(t, err, ErrServiceUnavailable)
	require.False(t, allowed)
}
