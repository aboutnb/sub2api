package service

import (
	"context"
	"strings"
)

// SignupRiskGrantStore records the one-time signup benefit decision for a
// server-derived risk identity. Implementations must serialize concurrent
// claims for the same fingerprint and return true only for the first account.
type SignupRiskGrantStore interface {
	ClaimSignupGrant(context.Context, int64, string) (bool, error)
	SignupGrantAllowed(context.Context, int64) (bool, error)
	ReleaseSignupGrant(context.Context, int64) error
}

type signupRiskIdentityContextKey struct{}

type signupRiskConcurrencyContextKey struct{}

func withSignupRiskConcurrency(ctx context.Context, concurrency int) context.Context {
	return context.WithValue(ctx, signupRiskConcurrencyContextKey{}, normalizeUserConcurrency(concurrency))
}

// SignupRiskRestoreConcurrency is the configured pre-restriction value, not
// the provisional -1 used while signup benefit eligibility is checked.
func SignupRiskRestoreConcurrency(ctx context.Context) int {
	if ctx != nil {
		if value, ok := ctx.Value(signupRiskConcurrencyContextKey{}).(int); ok {
			return value
		}
	}
	return -1
}

// WithSignupRiskIdentity attaches a server-derived, non-reversible identity to
// a registration request. The raw IP and User-Agent are deliberately absent.
func WithSignupRiskIdentity(ctx context.Context, fingerprint string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, signupRiskIdentityContextKey{}, strings.TrimSpace(fingerprint))
}

func signupRiskIdentityFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(signupRiskIdentityContextKey{}).(string)
	return strings.TrimSpace(value)
}
