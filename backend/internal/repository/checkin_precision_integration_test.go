//go:build integration

package repository

import (
	"context"
	"database/sql"
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

	var historyType, historyValue, historyNotes string
	var historyUserID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		SELECT type, value::text, used_by, notes
		FROM redeem_codes
		WHERE code = $1`, fmt.Sprintf("SYS-CHECKIN-%d", record.ID)).Scan(&historyType, &historyValue, &historyUserID, &historyNotes))
	require.Equal(t, service.RedeemTypeCheckin, historyType)
	require.Equal(t, "-15999987.14000000", historyValue)
	require.Equal(t, userID, historyUserID)
	require.Equal(t, service.CheckinHistoryModeNote("lucky"), historyNotes)
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

func TestCheckinBalanceHistoryModeMigrationBackfillsExistingRecords(t *testing.T) {
	ctx := context.Background()
	tx := testTx(t)
	email := fmt.Sprintf("checkin-history-mode-%d@example.com", time.Now().UnixNano())
	var userID int64
	require.NoError(t, tx.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, role, status, balance, concurrency)
		VALUES ($1, 'hash', 'user', 'active', 10, 1)
		RETURNING id`, email).Scan(&userID))

	checkedInAt := time.Date(2026, 7, 31, 8, 30, 0, 0, time.UTC)
	var recordID int64
	require.NoError(t, tx.QueryRowContext(ctx, `
		INSERT INTO checkin_records
			(user_id, checkin_date, mode, reward_type, random_value, reward_amount,
			 balance_before, balance_after, calculation_scale, checked_in_at)
		VALUES ($1, '2026-07-31', 'lucky', 'multiplier', 0.02, 0.20, 10, 10.20, 2, $2)
		RETURNING id`, userID, checkedInAt).Scan(&recordID))
	_, err := tx.ExecContext(ctx, `
		INSERT INTO redeem_codes (code, type, value, status, used_by, used_at, created_at)
		VALUES ('SYS-CHECKIN-' || $1, 'checkin', 0.20, 'used', $2, $3, $3)`,
		recordID, userID, checkedInAt)
	require.NoError(t, err)

	migrationSQL, err := migrations.FS.ReadFile("201_checkin_balance_history_mode.sql")
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, string(migrationSQL))
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, string(migrationSQL))
	require.NoError(t, err)

	var notes string
	require.NoError(t, tx.QueryRowContext(ctx, `
		SELECT notes FROM redeem_codes WHERE code = 'SYS-CHECKIN-' || $1`, recordID).Scan(&notes))
	require.Equal(t, service.CheckinHistoryModeNote("lucky"), notes)
}

func TestCheckinDefaultDistributionMigrationUpgradesOnlyLegacyDefaults(t *testing.T) {
	ctx := context.Background()
	tx := testTx(t)
	_, err := tx.ExecContext(ctx, `
		UPDATE settings
		SET value = CASE key
			WHEN 'checkin_lucky_positive_probability' THEN '70'
			WHEN 'checkin_lucky_min_multiplier' THEN '-0.05'
			WHEN 'checkin_lucky_max_multiplier' THEN '0.10'
			WHEN 'checkin_lucky_multiplier_positive_tiers' THEN '[{"min":"0.01","max":"0.10","weight":"100"}]'
			WHEN 'checkin_config_version' THEN '7'
			ELSE value
		END
		WHERE key IN (
			'checkin_lucky_positive_probability',
			'checkin_lucky_min_multiplier',
			'checkin_lucky_max_multiplier',
			'checkin_lucky_multiplier_positive_tiers',
			'checkin_config_version'
		)`)
	require.NoError(t, err)

	migrationSQL, err := migrations.FS.ReadFile("199_checkin_default_lucky_distribution.sql")
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, string(migrationSQL))
	require.NoError(t, err)

	values := readCheckinDefaultSettings(t, ctx, tx)
	require.Equal(t, "65", values["checkin_lucky_positive_probability"])
	require.Equal(t, "-0.08", values["checkin_lucky_min_multiplier"])
	require.Equal(t, "0.20", values["checkin_lucky_max_multiplier"])
	require.JSONEq(t, `[{"min":"0.01","max":"0.10","weight":"70"},{"min":"0.11","max":"0.15","weight":"20"},{"min":"0.16","max":"0.20","weight":"10"}]`, values["checkin_lucky_multiplier_positive_tiers"])
	require.Equal(t, "8", values["checkin_config_version"])

	_, err = tx.ExecContext(ctx, `
		UPDATE settings
		SET value = CASE key
			WHEN 'checkin_lucky_positive_probability' THEN '63'
			WHEN 'checkin_lucky_min_multiplier' THEN '-0.07'
			WHEN 'checkin_lucky_max_multiplier' THEN '0.18'
			WHEN 'checkin_lucky_multiplier_positive_tiers' THEN '[{"min":"0.01","max":"0.18","weight":"100"}]'
			ELSE value
		END
		WHERE key IN (
			'checkin_lucky_positive_probability',
			'checkin_lucky_min_multiplier',
			'checkin_lucky_max_multiplier',
			'checkin_lucky_multiplier_positive_tiers'
		)`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, string(migrationSQL))
	require.NoError(t, err)

	values = readCheckinDefaultSettings(t, ctx, tx)
	require.Equal(t, "63", values["checkin_lucky_positive_probability"])
	require.Equal(t, "-0.07", values["checkin_lucky_min_multiplier"])
	require.Equal(t, "0.18", values["checkin_lucky_max_multiplier"])
	require.JSONEq(t, `[{"min":"0.01","max":"0.18","weight":"100"}]`, values["checkin_lucky_multiplier_positive_tiers"])
	require.Equal(t, "8", values["checkin_config_version"])
}

func readCheckinDefaultSettings(t *testing.T, ctx context.Context, tx interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}) map[string]string {
	t.Helper()
	rows, err := tx.QueryContext(ctx, `
		SELECT key, value
		FROM settings
		WHERE key IN (
			'checkin_lucky_positive_probability',
			'checkin_lucky_min_multiplier',
			'checkin_lucky_max_multiplier',
			'checkin_lucky_multiplier_positive_tiers',
			'checkin_config_version'
		)`)
	require.NoError(t, err)
	defer rows.Close()

	values := map[string]string{}
	for rows.Next() {
		var key, value string
		require.NoError(t, rows.Scan(&key, &value))
		values[key] = value
	}
	require.NoError(t, rows.Err())
	return values
}
