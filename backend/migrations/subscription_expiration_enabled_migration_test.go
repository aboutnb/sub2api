package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubscriptionExpirationEnabledMigrationDefaultsToHistoricalBehavior(t *testing.T) {
	content, err := FS.ReadFile("235_subscription_expiration_enabled.sql")
	require.NoError(t, err)

	sql := strings.ToLower(strings.Join(strings.Fields(string(content)), " "))
	require.Contains(t, sql, "values ('subscription_expiration_enabled', 'true'")
	require.Contains(t, sql, "on conflict (key) do nothing")
}
