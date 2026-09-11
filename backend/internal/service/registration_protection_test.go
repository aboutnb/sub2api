package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRegistrationProtectionSettingsDefaultsAndPersistence(t *testing.T) {
	repo := &panelRateLimitSettingRepo{}
	svc := &SettingService{settingRepo: repo}
	ctx := context.Background()
	settings, err := svc.GetRegistrationProtectionSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, settings.IdentitySuccessLimit)
	require.Equal(t, 10, settings.IPSuccessLimit)
	require.Equal(t, 24, settings.SuccessWindowHours)
	settings.Enabled = false
	settings.SuccessWindowHours = 48
	require.NoError(t, svc.SetRegistrationProtectionSettings(ctx, settings))
	got, err := svc.GetRegistrationProtectionSettings(ctx)
	require.NoError(t, err)
	require.Equal(t, settings, got)
}

func TestRegistrationProtectionSettingsRejectInvalidPolicy(t *testing.T) {
	for _, mutate := range []func(*RegistrationProtectionSettings){
		func(s *RegistrationProtectionSettings) { s.IdentitySuccessLimit = 0 },
		func(s *RegistrationProtectionSettings) { s.SuccessWindowHours = 721 },
		func(s *RegistrationProtectionSettings) { s.BlockMinutes = -1 },
		func(s *RegistrationProtectionSettings) { s.IdentitySuccessLimit = s.IPSuccessLimit + 1 },
		func(s *RegistrationProtectionSettings) { s.IdentityFailureLimit = s.IPFailureLimit + 1 },
	} {
		settings := DefaultRegistrationProtectionSettings()
		mutate(&settings)
		require.Error(t, settings.Validate())
	}
}

func TestRegistrationSourceRequiresTrustedIdentity(t *testing.T) {
	_, ok := RegistrationSourceFromContext(context.Background())
	require.False(t, ok)
	_, ok = RegistrationSourceFromContext(WithRegistrationSource(context.Background(), RegistrationSource{IPHash: "ip"}))
	require.False(t, ok)
	source := RegistrationSource{IPHash: "ip", IdentityHash: "identity", Policy: DefaultRegistrationProtectionSettings()}
	got, ok := RegistrationSourceFromContext(WithRegistrationSource(context.Background(), source))
	require.True(t, ok)
	require.Equal(t, source, got)
}

func TestRegistrationProtectionCacheUpdatesAndFailsClosedAfterExpiry(t *testing.T) {
	repo := &panelRateLimitSettingRepo{}
	svc := &SettingService{settingRepo: repo}
	ctx := context.Background()
	settings, err := svc.GetRegistrationProtectionSettingsCached(ctx)
	require.NoError(t, err)
	_, err = svc.GetRegistrationProtectionSettingsCached(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, repo.getValueCalls)
	settings.SuccessWindowHours = 48
	require.NoError(t, svc.SetRegistrationProtectionSettings(ctx, settings))
	got, err := svc.GetRegistrationProtectionSettingsCached(ctx)
	require.NoError(t, err)
	require.Equal(t, 48, got.SuccessWindowHours)
	repo.getValueErr = errors.New("database offline")
	svc.registrationProtectionCache.Store(&cachedRegistrationProtection{settings: settings, expiresAt: time.Now().Add(-time.Second)})
	_, err = svc.GetRegistrationProtectionSettingsCached(ctx)
	require.Error(t, err)
}
