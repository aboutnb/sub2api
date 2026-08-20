package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadUSDTSettingsFromEnvironment(t *testing.T) {
	resetViperWithJWTSecret(t)
	t.Setenv("USDT_PAYMENT_ENABLED_NETWORKS", "tron, bsc, polygon")
	t.Setenv("USDT_PAYMENT_REQUEST_TIMEOUT_SECONDS", "9")
	t.Setenv("USDT_PAYMENT_RECONCILE_INTERVAL_SECONDS", "4")
	t.Setenv("USDT_PAYMENT_RECONCILE_BATCH_SIZE", "37")
	t.Setenv("USDT_PAYMENT_WEBHOOK_CLOCK_SKEW_SECONDS", "420")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, []string{"tron", "bsc", "polygon"}, cfg.USDTPayment.EnabledNetworks)
	require.Equal(t, 9, cfg.USDTPayment.RequestTimeoutSeconds)
	require.Equal(t, 4, cfg.USDTPayment.ReconcileIntervalSeconds)
	require.Equal(t, 37, cfg.USDTPayment.ReconcileBatchSize)
	require.Equal(t, 420, cfg.USDTPayment.WebhookClockSkewSeconds)
}
