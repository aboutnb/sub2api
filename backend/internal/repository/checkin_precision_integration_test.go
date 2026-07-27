//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCheckinApplySettlesHighBalanceAtTwoDecimals(t *testing.T) {
	ctx := context.Background()
	email := fmt.Sprintf("checkin-precision-%d@example.com", time.Now().UnixNano())
	var userID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, role, status, balance, concurrency)
		VALUES ($1, 'hash', 'user', 'active', 199999839.31129506, 1)
		RETURNING id`, email).Scan(&userID))
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})

	repo := &checkinRepository{db: integrationDB}
	record, newlyCheckedIn, err := repo.Apply(ctx, userID, "2026-07-27", "lucky", func(balance decimal.Decimal) (decimal.Decimal, decimal.Decimal, string, error) {
		require.Equal(t, "199999839.31129506", balance.StringFixed(8))
		multiplier := decimal.RequireFromString("-0.08")
		return balance.Mul(multiplier).Round(service.CheckinCalculationScale), multiplier, service.CheckinRewardTypeMultiplier, nil
	})

	require.NoError(t, err)
	require.True(t, newlyCheckedIn)
	require.Equal(t, -15999987.14, record.RewardAmount)
	require.Equal(t, 183999852.17129506, record.BalanceAfter)

	var calculationScale int
	var storedReward string
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		SELECT calculation_scale, reward_amount::text
		FROM checkin_records
		WHERE id = $1`, record.ID).Scan(&calculationScale, &storedReward))
	require.Equal(t, service.CheckinCalculationScale, calculationScale)
	require.Equal(t, "-15999987.14000000", storedReward)
}

func TestCheckinPrecisionMigrationNormalizesLegacySettings(t *testing.T) {
	ctx := context.Background()
	tx := testTx(t)
	_, err := tx.ExecContext(ctx, `
		UPDATE settings
		SET value = CASE key
			WHEN 'checkin_normal_max' THEN '.12345678'
			WHEN 'checkin_lucky_min_multiplier' THEN '-0.004'
			ELSE value
		END
		WHERE key IN ('checkin_normal_max', 'checkin_lucky_min_multiplier')`)
	require.NoError(t, err)

	migrationSQL, err := migrations.FS.ReadFile("197_checkin_two_decimal_precision.sql")
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, string(migrationSQL))
	require.NoError(t, err)

	values := map[string]string{}
	rows, err := tx.QueryContext(ctx, `
		SELECT key, value
		FROM settings
		WHERE key IN ('checkin_normal_max', 'checkin_lucky_min_multiplier')`)
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var key, value string
		require.NoError(t, rows.Scan(&key, &value))
		values[key] = value
	}
	require.NoError(t, rows.Err())
	require.Equal(t, "0.12", values["checkin_normal_max"])
	require.Equal(t, "-0.01", values["checkin_lucky_min_multiplier"])
}
