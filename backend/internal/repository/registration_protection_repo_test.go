package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestRegistrationSourceReleaseKeepsBlockWhenCounterResetFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT scope,source_hash").WithArgs(int64(8)).WillReturnRows(sqlmock.NewRows([]string{"scope", "source_hash"}).AddRow("ip_ua", "identity-hash"))
	mock.ExpectRollback()
	err = NewRegistrationProtectionRepository(db).ReleaseSourceBlock(context.Background(), 8, 1, "reviewed", func(_ context.Context, key string) error {
		require.Equal(t, "registration_failure:identity:identity-hash", key)
		return errors.New("redis unavailable")
	})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRegistrationSourceReleasePreservesAuditAndResetsOnlyMatchedCounter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT scope,source_hash").WithArgs(int64(8)).WillReturnRows(sqlmock.NewRows([]string{"scope", "source_hash"}).AddRow("ip", "ip-hash"))
	mock.ExpectExec("UPDATE registration_source_blocks").WithArgs(int64(8), int64(1), "reviewed").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	err = NewRegistrationProtectionRepository(db).ReleaseSourceBlock(context.Background(), 8, 1, "reviewed", func(_ context.Context, key string) error {
		require.Equal(t, "registration_failure:ip:ip-hash", key)
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRegistrationRiskReviewRestoresSavedConcurrency(t *testing.T) {
	for _, previous := range []int{-1, 0, 5} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT r.user_id").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"user_id", "concurrency", "previous_concurrency", "status", "role"}).AddRow(42, -1, previous, "restricted", "user"))
		mock.ExpectExec("UPDATE users SET concurrency").WithArgs(int64(42), previous).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("UPDATE registration_risk_accounts").WithArgs(int64(7), "released", previous, int64(1), "reviewed").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("INSERT INTO registration_risk_events").WithArgs(int64(42), "released").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		id, err := NewRegistrationProtectionRepository(db).ReviewAccount(context.Background(), 7, 1, "release", "reviewed")
		require.NoError(t, err)
		require.EqualValues(t, 42, id)
		require.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestRegistrationRiskReviewRejectsChangedOrAdminAccount(t *testing.T) {
	for _, tc := range []struct {
		current              int
		status, role, action string
	}{
		{3, "restricted", "user", "release"},
		{-1, "released", "user", "release"},
		{-1, "observed", "user", "restrict"},
		{-1, "restricted", "admin", "release"},
	} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT r.user_id").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"user_id", "concurrency", "previous_concurrency", "status", "role"}).AddRow(42, tc.current, 5, tc.status, tc.role))
		mock.ExpectRollback()
		_, err = NewRegistrationProtectionRepository(db).ReviewAccount(context.Background(), 7, 1, tc.action, "reviewed")
		require.ErrorIs(t, err, service.ErrRegistrationRiskConflict)
		require.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestRegistrationRiskReviewRollsBackOnAuditFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT r.user_id").WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"user_id", "concurrency", "previous_concurrency", "status", "role"}).AddRow(42, 3, 0, "observed", "user"))
	mock.ExpectExec("UPDATE users SET concurrency").WithArgs(int64(42), -1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE registration_risk_accounts").WithArgs(int64(7), "restricted", 3, int64(1), "reviewed").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO registration_risk_events").WithArgs(int64(42), "restricted").WillReturnError(errors.New("audit unavailable"))
	mock.ExpectRollback()
	_, err = NewRegistrationProtectionRepository(db).ReviewAccount(context.Background(), 7, 1, "restrict", "reviewed")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
