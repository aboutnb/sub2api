package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
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
			AddRow(service.RoleAdmin, service.StatusActive, 10.0))
	mock.ExpectQuery("SELECT id, user_id, checkin_date, mode, reward_type, random_value, reward_amount").
		WithArgs(int64(7), "2026-07-26").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "checkin_date", "mode", "reward_type", "random_value", "reward_amount",
			"balance_before", "balance_after", "checked_in_at",
		}))
	mock.ExpectQuery("UPDATE users").
		WithArgs(0.02, int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(10.02))
	mock.ExpectQuery("INSERT INTO checkin_records").
		WithArgs(int64(7), "2026-07-26", "normal", service.CheckinRewardTypeAmount, 0.02, 0.02, 10.0, 10.02).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "checkin_date", "mode", "reward_type", "random_value", "reward_amount",
			"balance_before", "balance_after", "checked_in_at",
		}).AddRow(9, checkedInAt, "normal", service.CheckinRewardTypeAmount, 0.02, 0.02, 10.0, 10.02, checkedInAt))
	mock.ExpectCommit()

	repo := &checkinRepository{db: db}
	record, newlyCheckedIn, err := repo.Apply(context.Background(), 7, "2026-07-26", "normal", func(float64) (float64, float64, string, error) {
		return 0.02, 0.02, service.CheckinRewardTypeAmount, nil
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
			AddRow(service.RoleUser, service.StatusActive, -50.0))
	mock.ExpectRollback()

	repo := &checkinRepository{db: db}
	calculated := false
	_, newlyCheckedIn, err := repo.Apply(context.Background(), 7, "2026-07-26", "normal", func(float64) (float64, float64, string, error) {
		calculated = true
		return 0.01, 0.01, service.CheckinRewardTypeAmount, nil
	})

	require.ErrorIs(t, err, service.ErrCheckinNegativeBalance)
	require.False(t, newlyCheckedIn)
	require.False(t, calculated)
	require.NoError(t, mock.ExpectationsWereMet())
}
