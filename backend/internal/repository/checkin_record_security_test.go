package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCheckinApplyAllowsActiveAdminAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	checkedInAt := time.Date(2026, 7, 26, 9, 30, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT role, status, balance").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"role", "status", "balance"}).
			AddRow(service.RoleAdmin, service.StatusActive, "10.00000000"))
	mock.ExpectQuery("SELECT id, user_id, checkin_date, mode, reward_type, random_value, reward_amount").
		WithArgs(int64(7), "2026-07-26").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "checkin_date", "mode", "reward_type", "random_value", "reward_amount",
			"balance_before", "balance_after", "checked_in_at",
		}))
	mock.ExpectQuery("UPDATE users").
		WithArgs("0.02", int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow("10.02000000"))
	mock.ExpectQuery("INSERT INTO checkin_records").
		WithArgs(int64(7), "2026-07-26", "normal", service.CheckinRewardTypeAmount, "0.02", "0.02", "10.00000000", "10.02000000", service.CheckinCalculationScale).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "checkin_date", "mode", "reward_type", "random_value", "reward_amount",
			"balance_before", "balance_after", "checked_in_at",
		}).AddRow(9, checkedInAt, "normal", service.CheckinRewardTypeAmount, 0.02, 0.02, 10.0, 10.02, checkedInAt))
	mock.ExpectExec("INSERT INTO redeem_codes").
		WithArgs(int64(9), service.RedeemTypeCheckin, "0.02", service.StatusUsed, int64(7), checkedInAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := &checkinRepository{db: db}
	record, newlyCheckedIn, err := repo.Apply(context.Background(), 7, "2026-07-26", "normal", func(decimal.Decimal) (decimal.Decimal, decimal.Decimal, string, error) {
		return decimal.RequireFromString("0.02"), decimal.RequireFromString("0.02"), service.CheckinRewardTypeAmount, nil
	})

	require.NoError(t, err)
	require.True(t, newlyCheckedIn)
	require.Equal(t, int64(7), record.UserID)
	require.InDelta(t, 10.02, record.BalanceAfter, 0.00000001)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckinApplyRejectsBalanceThatTurnedNegativeInsideTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT role, status, balance").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"role", "status", "balance"}).
			AddRow(service.RoleUser, service.StatusActive, "-50.00000000"))
	mock.ExpectRollback()

	repo := &checkinRepository{db: db}
	calculated := false
	_, newlyCheckedIn, err := repo.Apply(context.Background(), 7, "2026-07-26", "normal", func(decimal.Decimal) (decimal.Decimal, decimal.Decimal, string, error) {
		calculated = true
		return decimal.RequireFromString("0.01"), decimal.RequireFromString("0.01"), service.CheckinRewardTypeAmount, nil
	})

	require.ErrorIs(t, err, service.ErrCheckinNegativeBalance)
	require.False(t, newlyCheckedIn)
	require.False(t, calculated)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCheckinApplyPreservesExactHighBalanceDecimals(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	checkedInAt := time.Date(2026, 7, 27, 7, 39, 7, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT role, status, balance::text").
		WithArgs(int64(6)).
		WillReturnRows(sqlmock.NewRows([]string{"role", "status", "balance"}).
			AddRow(service.RoleUser, service.StatusActive, "199999839.31129506"))
	mock.ExpectQuery("SELECT id, user_id, checkin_date, mode, reward_type, random_value, reward_amount").
		WithArgs(int64(6), "2026-07-27").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "checkin_date", "mode", "reward_type", "random_value", "reward_amount",
			"balance_before", "balance_after", "checked_in_at",
		}))
	mock.ExpectQuery("UPDATE users").
		WithArgs("-15999987.14", int64(6)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow("183999852.17129506"))
	mock.ExpectQuery("INSERT INTO checkin_records").
		WithArgs(
			int64(6), "2026-07-27", "lucky", service.CheckinRewardTypeMultiplier,
			"-0.08", "-15999987.14", "199999839.31129506", "183999852.17129506", service.CheckinCalculationScale,
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "checkin_date", "mode", "reward_type", "random_value", "reward_amount",
			"balance_before", "balance_after", "checked_in_at",
		}).AddRow(
			19, checkedInAt, "lucky", service.CheckinRewardTypeMultiplier, -0.08, -15999987.14,
			199999839.31129506, 183999852.17129506, checkedInAt,
		))
	mock.ExpectExec("INSERT INTO redeem_codes").
		WithArgs(int64(19), service.RedeemTypeCheckin, "-15999987.14", service.StatusUsed, int64(6), checkedInAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := &checkinRepository{db: db}
	record, newlyCheckedIn, err := repo.Apply(context.Background(), 6, "2026-07-27", "lucky", func(balance decimal.Decimal) (decimal.Decimal, decimal.Decimal, string, error) {
		require.Equal(t, "199999839.31129506", balance.StringFixed(8))
		return decimal.RequireFromString("-15999987.14"), decimal.RequireFromString("-0.08"), service.CheckinRewardTypeMultiplier, nil
	})

	require.NoError(t, err)
	require.True(t, newlyCheckedIn)
	require.Equal(t, int64(6), record.UserID)
	require.InDelta(t, -15999987.14, record.RewardAmount, 0.001)
	require.NoError(t, mock.ExpectationsWereMet())
}
