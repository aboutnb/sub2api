//go:build integration

package repository

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestChannelMonitorPlatformCatalogMigration(t *testing.T) {
	ctx := context.Background()
	tx, err := integrationDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	// A temporary table shadows the real configuration without altering it.
	_, err = tx.ExecContext(ctx, `CREATE TEMP TABLE channel_monitor_v2_config (
        id integer PRIMARY KEY, version integer, enabled boolean DEFAULT true,
        platforms jsonb DEFAULT '[]', group_ids bigint[] DEFAULT '{42}',
        updated_at timestamptz DEFAULT NOW()
    ) ON COMMIT DROP;
    INSERT INTO channel_monitor_v2_config (id, version, platforms) VALUES
    (1, 7, '[{"platform":"openai","enabled":true,"models":["gpt-custom"]},
             {"platform":"deepseek","enabled":false,"models":["deepseek-custom"]},
             {"platform":"custom","enabled":false,"models":["private"]}]'),
    (2, 9, '[]');`)
	require.NoError(t, err)
	sql, err := migrations.FS.ReadFile("255_channel_monitor_platform_catalog.sql")
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, string(sql))
	require.NoError(t, err)
	var raw []byte
	var version int
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT platforms, version FROM channel_monitor_v2_config WHERE id=1 AND enabled AND group_ids='{42}'`).Scan(&raw, &version))
	require.Equal(t, 8, version)
	var platforms []service.ChannelMonitorV2PlatformConfig
	require.NoError(t, json.Unmarshal(raw, &platforms))
	byPlatform := make(map[string]service.ChannelMonitorV2PlatformConfig)
	for _, p := range platforms {
		byPlatform[p.Platform] = p
	}
	require.Equal(t, []string{"gpt-custom"}, byPlatform["openai"].Models)
	require.False(t, byPlatform["deepseek"].Enabled)
	require.Equal(t, []string{"deepseek-custom"}, byPlatform["deepseek"].Models)
	require.Equal(t, []string{"private"}, byPlatform["custom"].Models)
	for _, p := range []string{"kimi", "zhipu", "minimax", "opencode_go", "typesafe"} {
		require.True(t, byPlatform[p].Enabled, p)
		require.Empty(t, byPlatform[p].Models, p)
	}
	_, err = tx.ExecContext(ctx, string(sql))
	require.NoError(t, err)
	var repeated []byte
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT platforms, version FROM channel_monitor_v2_config WHERE id=1`).Scan(&repeated, &version))
	require.JSONEq(t, string(raw), string(repeated))
	require.Equal(t, 8, version)
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT platforms, version FROM channel_monitor_v2_config WHERE id=2`).Scan(&raw, &version))
	require.JSONEq(t, `[]`, string(raw))
	require.Equal(t, 9, version)
	_, err = tx.ExecContext(ctx, `INSERT INTO channel_monitor_v2_config (id, version) VALUES (3, 1)`)
	require.NoError(t, err)
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT platforms FROM channel_monitor_v2_config WHERE id=3`).Scan(&raw))
	require.NoError(t, json.Unmarshal(raw, &platforms))
	require.Len(t, platforms, 12)

	// The v0.2.15 extension follows the same preservation and idempotency rules.
	_, err = tx.ExecContext(ctx, `UPDATE channel_monitor_v2_config SET platforms = platforms || '[{"platform":"cline","enabled":false,"models":["private-cline"]}]'::jsonb WHERE id=1`)
	require.NoError(t, err)
	sql, err = migrations.FS.ReadFile("256_channel_monitor_v0215_platforms.sql")
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, string(sql))
	require.NoError(t, err)
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT platforms, version FROM channel_monitor_v2_config WHERE id=1`).Scan(&raw, &version))
	require.Equal(t, 9, version)
	require.NoError(t, json.Unmarshal(raw, &platforms))
	for _, p := range platforms {
		if p.Platform == "cline" {
			require.False(t, p.Enabled)
			require.Equal(t, []string{"private-cline"}, p.Models)
		}
		if p.Platform == "command_code" {
			require.True(t, p.Enabled)
		}
	}
	_, err = tx.ExecContext(ctx, string(sql))
	require.NoError(t, err)
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT platforms, version FROM channel_monitor_v2_config WHERE id=1`).Scan(&repeated, &version))
	require.JSONEq(t, string(raw), string(repeated))
	require.Equal(t, 9, version)
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT platforms, version FROM channel_monitor_v2_config WHERE id=2`).Scan(&raw, &version))
	require.JSONEq(t, `[]`, string(raw))
	require.Equal(t, 9, version)
	_, err = tx.ExecContext(ctx, `INSERT INTO channel_monitor_v2_config (id, version) VALUES (4, 1)`)
	require.NoError(t, err)
	require.NoError(t, tx.QueryRowContext(ctx, `SELECT platforms FROM channel_monitor_v2_config WHERE id=4`).Scan(&raw))
	require.NoError(t, json.Unmarshal(raw, &platforms))
	require.Len(t, platforms, 14)
}
