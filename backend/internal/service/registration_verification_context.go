package service

import (
	"context"
	"strings"
)

type registrationVerificationContextKey struct{}

// RegistrationVerificationContext is stored with registration verification codes.
// Values are pre-hashed by the handler so Redis never stores raw IP or UA strings.
type RegistrationVerificationContext struct {
	Action            string
	ClientIPHash      string
	UserAgentHash     string
	NetworkBucketHash string
}

func WithRegistrationVerificationContext(ctx context.Context, value RegistrationVerificationContext) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	value.Action = strings.TrimSpace(value.Action)
	value.ClientIPHash = strings.TrimSpace(value.ClientIPHash)
	value.UserAgentHash = strings.TrimSpace(value.UserAgentHash)
	value.NetworkBucketHash = strings.TrimSpace(value.NetworkBucketHash)
	return context.WithValue(ctx, registrationVerificationContextKey{}, value)
}

func registrationVerificationContextFrom(ctx context.Context) (RegistrationVerificationContext, bool) {
	if ctx == nil {
		return RegistrationVerificationContext{}, false
	}
	value, ok := ctx.Value(registrationVerificationContextKey{}).(RegistrationVerificationContext)
	if !ok {
		return RegistrationVerificationContext{}, false
	}
	value.Action = strings.TrimSpace(value.Action)
	value.ClientIPHash = strings.TrimSpace(value.ClientIPHash)
	value.UserAgentHash = strings.TrimSpace(value.UserAgentHash)
	value.NetworkBucketHash = strings.TrimSpace(value.NetworkBucketHash)
	return value, value.Action != "" || value.ClientIPHash != "" || value.UserAgentHash != "" || value.NetworkBucketHash != ""
}
