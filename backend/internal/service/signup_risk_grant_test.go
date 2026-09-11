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
	require.Equal(t, -1, grant.initialConcurrency())

	allowed, err := svc.applySignupGrant(ctx, &User{ID: 42}, grant)
	require.ErrorIs(t, err, ErrServiceUnavailable)
	require.False(t, allowed)
}

type signupRiskGrantStoreStub struct {
	allowed            bool
	restoreConcurrency int
}

func (s *signupRiskGrantStoreStub) ClaimSignupGrant(ctx context.Context, _ int64, _ string) (bool, error) {
	s.restoreConcurrency = SignupRiskRestoreConcurrency(ctx)
	return s.allowed, nil
}

func TestSignupRiskRestrictionPreservesConfiguredConcurrency(t *testing.T) {
	for _, expected := range []int{-1, 0, 4} {
		store := &signupRiskGrantStoreStub{}
		svc := &AuthService{signupRiskGrantStore: store}
		user := &User{ID: 42, Concurrency: -1}
		allowed, err := svc.applySignupGrant(WithSignupRiskIdentity(context.Background(), "risk"), user,
			signupGrantApplication{plan: signupGrantPlan{Concurrency: expected}, deferred: true})
		require.NoError(t, err)
		require.False(t, allowed)
		require.Equal(t, expected, store.restoreConcurrency)
		require.Equal(t, -1, user.Concurrency)
	}
}

func (s *signupRiskGrantStoreStub) SignupGrantAllowed(context.Context, int64) (bool, error) {
	return s.allowed, nil
}

func (s *signupRiskGrantStoreStub) ReleaseSignupGrant(context.Context, int64) error {
	return nil
}

type signupGrantUserRepoStub struct {
	UserRepository
	updateCalls            int
	updatedConcurrency     int
	updateConcurrencyCalls int
}

func (s *signupGrantUserRepoStub) UpdateConcurrency(_ context.Context, _ int64, amount int) error {
	s.updateConcurrencyCalls++
	s.updatedConcurrency = amount
	return nil
}

func (s *signupGrantUserRepoStub) Update(_ context.Context, user *User, fields UserUpdateFields) error {
	s.updateCalls++
	if fields.Concurrency {
		s.updatedConcurrency = user.Concurrency
	}
	return nil
}

func TestSignupRiskGrantWritesUnlimitedConcurrencyAfterApproval(t *testing.T) {
	repo := &signupGrantUserRepoStub{}
	svc := &AuthService{
		userRepo:             repo,
		signupRiskGrantStore: &signupRiskGrantStoreStub{allowed: true},
	}
	ctx := WithSignupRiskIdentity(context.Background(), "risk-fingerprint")
	grant := signupGrantApplication{
		plan:     signupGrantPlan{Concurrency: 0},
		deferred: true,
	}
	user := &User{ID: 42, Concurrency: -1}

	allowed, err := svc.applySignupGrant(ctx, user, grant)

	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, 1, repo.updateCalls)
	require.Zero(t, repo.updateConcurrencyCalls)
	require.Equal(t, 0, repo.updatedConcurrency)
	require.Equal(t, 0, user.Concurrency)
}

func TestSignupGrantNormalizesInvalidConfiguredConcurrency(t *testing.T) {
	svc := &AuthService{cfg: &config.Config{Default: config.DefaultConfig{UserConcurrency: -2}}}

	grant := svc.resolveSignupGrantPlan(context.Background(), "email")

	require.Equal(t, -1, grant.Concurrency)
	require.Equal(t, -1, (signupGrantApplication{plan: grant}).initialConcurrency())
}
