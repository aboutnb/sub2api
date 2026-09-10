package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSmartRoutingMigrationKeepsAPIKeySchemaModular(t *testing.T) {
	content, err := FS.ReadFile("234_api_key_smart_routing.sql")
	require.NoError(t, err)
	sql := strings.ToLower(string(content))

	require.Contains(t, sql, "create table if not exists api_key_smart_routes")
	require.Contains(t, sql, "create table if not exists api_key_smart_route_groups")
	require.Contains(t, sql, "references api_keys(id) on delete cascade")
	require.Contains(t, sql, "primary key (api_key_id, group_id)")
	require.Contains(t, sql, "unique (api_key_id, position)")
	require.Contains(t, sql, "price_weight + speed_weight + success_weight = 100")
	require.Contains(t, sql, "max_rate_multiplier < 'infinity'::double precision")
	require.NotContains(t, sql, "alter table api_keys")
}
