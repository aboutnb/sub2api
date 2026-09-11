package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type signupRiskGrantRepository struct {
	db *sql.DB
}

func NewSignupRiskGrantRepository(db *sql.DB) service.SignupRiskGrantStore {
	return &signupRiskGrantRepository{db: db}
}

// ClaimSignupGrant serializes the check-and-insert with a transaction-scoped
// advisory lock. This avoids two concurrent registrations from both receiving
// the initial balance/subscription package.
func (r *signupRiskGrantRepository) ClaimSignupGrant(ctx context.Context, userID int64, fingerprint string) (bool, error) {
	if r == nil || r.db == nil || userID <= 0 || fingerprint == "" {
		return false, errors.New("signup risk grant repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 208))`, fingerprint); err != nil {
		return false, err
	}
	var existingAllowed bool
	userErr := tx.QueryRowContext(ctx, `SELECT grant_allowed FROM signup_risk_accounts WHERE user_id = $1`, userID).Scan(&existingAllowed)
	if userErr == nil {
		if err := tx.Commit(); err != nil {
			return false, err
		}
		return existingAllowed, nil
	}
	if !errors.Is(userErr, sql.ErrNoRows) {
		return false, userErr
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM signup_risk_accounts WHERE fingerprint_hash = $1)`, fingerprint).Scan(&exists); err != nil {
		return false, err
	}
	allowed := !exists
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO signup_risk_accounts (user_id, fingerprint_hash, grant_allowed)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO NOTHING`, userID, fingerprint, allowed); err != nil {
		return false, err
	}
	if !allowed {
		if _, err := tx.ExecContext(ctx, `INSERT INTO registration_risk_accounts
			(user_id, reason, status, previous_concurrency)
			VALUES ($1, 'duplicate_signup_identity', 'restricted', $2)
			ON CONFLICT (user_id) DO UPDATE SET reason=EXCLUDED.reason, status=EXCLUDED.status,
			previous_concurrency=EXCLUDED.previous_concurrency, updated_at=NOW()
			WHERE registration_risk_accounts.status='observed' AND registration_risk_accounts.reviewed_at IS NULL`, userID, service.SignupRiskRestoreConcurrency(ctx)); err != nil {
			return false, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO registration_risk_events
			(user_id, ip_hash, identity_hash, ip_address, user_agent, trigger_path, reason, action)
			SELECT user_id, ip_hash, identity_hash, ip_address, user_agent, trigger_path,
			'duplicate_signup_identity', 'restricted'
			FROM registration_source_accounts WHERE user_id = $1`, userID); err != nil {
			return false, err
		}
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return allowed, nil
}

func (r *signupRiskGrantRepository) ReleaseSignupGrant(ctx context.Context, userID int64) error {
	if r == nil || r.db == nil || userID <= 0 {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM signup_risk_accounts WHERE user_id = $1`, userID)
	return err
}

func (r *signupRiskGrantRepository) SignupGrantAllowed(ctx context.Context, userID int64) (bool, error) {
	if r == nil || r.db == nil || userID <= 0 {
		return false, errors.New("signup risk grant repository is not configured")
	}
	var allowed bool
	err := r.db.QueryRowContext(ctx, `SELECT grant_allowed FROM signup_risk_accounts WHERE user_id = $1`, userID).Scan(&allowed)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return allowed, err
}
