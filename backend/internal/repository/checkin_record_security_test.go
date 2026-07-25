package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

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
	_, newlyCheckedIn, err := repo.Apply(context.Background(), 7, "2026-07-26", "normal", func(float64) (float64, float64, error) {
		calculated = true
		return 0.01, 0.01, nil
	})

	require.ErrorIs(t, err, service.ErrCheckinNegativeBalance)
	require.False(t, newlyCheckedIn)
	require.False(t, calculated)
	require.NoError(t, mock.ExpectationsWereMet())
}
