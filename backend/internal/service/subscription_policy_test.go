package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type subscriptionPolicyReaderStub struct {
	value string
	err   error
	calls int
}

func (s *subscriptionPolicyReaderStub) GetValue(context.Context, string) (string, error) {
	s.calls++
	return s.value, s.err
}

func TestSubscriptionPolicy_DefaultsToExpirationEnabled(t *testing.T) {
	require.True(t, NewSubscriptionPolicy(nil).ExpirationEnabled(context.Background()))

	reader := &subscriptionPolicyReaderStub{err: errors.New("settings unavailable")}
	require.True(t, NewSubscriptionPolicy(reader).ExpirationEnabled(context.Background()))
	require.Equal(t, 1, reader.calls)
}

func TestSubscriptionPolicy_CachesAndInvalidatesSetting(t *testing.T) {
	reader := &subscriptionPolicyReaderStub{value: "false"}
	policy := NewSubscriptionPolicy(reader)

	require.False(t, policy.ExpirationEnabled(context.Background()))
	reader.value = "true"
	require.False(t, policy.ExpirationEnabled(context.Background()))
	require.Equal(t, 1, reader.calls)

	policy.Invalidate()
	require.True(t, policy.ExpirationEnabled(context.Background()))
	require.Equal(t, 2, reader.calls)
}

func TestStaticSubscriptionPolicy_ReturnsConfiguredValue(t *testing.T) {
	require.True(t, NewStaticSubscriptionPolicy(true).ExpirationEnabled(context.Background()))
	require.False(t, NewStaticSubscriptionPolicy(false).ExpirationEnabled(context.Background()))
}
