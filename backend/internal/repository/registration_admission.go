package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type registrationAdmissionQuerier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

// Only server-side registration history decides a denial. A shared feature
// alone, a missing cookie alone, and transient infrastructure errors do not.
func registrationAdmissionReason(ctx context.Context, db registrationAdmissionQuerier, source service.RegistrationSource, email string) (string, error) {
	p := source.Policy
	rows, err := db.QueryContext(ctx, `SELECT
 EXISTS(SELECT 1 FROM registration_email_blocks WHERE email_identity=$1 AND released_at IS NULL),
 $2::boolean AND EXISTS(SELECT 1 FROM registration_source_blocks WHERE released_at IS NULL AND expires_at>NOW() AND ((scope='ip' AND source_hash=$3) OR (scope='ip_ua' AND source_hash=$4))),
 $2::boolean AND $5::text<>'' AND (
 (SELECT COUNT(*) + 1 >= $6 AND COUNT(DISTINCT ip_hash) + CASE WHEN COUNT(*) FILTER (WHERE ip_hash=$3)=0 THEN 1 ELSE 0 END >= $7 FROM registration_source_accounts WHERE feature_hash=$5 AND created_at>NOW()-($8 * INTERVAL '1 minute'))
 OR (SELECT COUNT(DISTINCT s.user_id)>=2 AND COUNT(DISTINCT s.ip_hash)>=2 FROM registration_source_accounts s JOIN users u ON u.id=s.user_id WHERE s.feature_hash=$5 AND u.status='disabled' AND EXISTS(SELECT 1 FROM registration_risk_events e WHERE e.user_id=u.id AND e.reason='manual_batch_registration_abuse' AND e.action='disabled'))
 )`, service.NormalizeEmailForAliasDedup(email), p.Enabled, source.IPHash, source.IdentityHash, source.FeatureHash, p.ObserveFingerprintLimit, p.FingerprintDistinctIPs, p.FingerprintWindowMinutes)
	if err != nil {
		return "", err
	}
	defer func() { _ = rows.Close() }()
	var emailBlocked, sourceBlocked, cluster bool
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("missing admission result")
	}
	if err = rows.Scan(&emailBlocked, &sourceBlocked, &cluster); err != nil {
		return "", err
	}
	if emailBlocked {
		return "email_blocked", nil
	}
	if sourceBlocked {
		return "source_blocked", nil
	}
	if cluster {
		return "registration_cluster", nil
	}
	return "", nil
}

func (r *registrationProtectionRepository) denyEmail(ctx context.Context, source service.RegistrationSource, email, reason string) error {
	// Persistence must complete even if the requester disconnects. It is outside
	// a rejected account transaction so rollback cannot erase the email marker.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer cancel()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `INSERT INTO registration_email_blocks(email_identity,reason) VALUES($1,$2) ON CONFLICT(email_identity) DO UPDATE SET reason=EXCLUDED.reason,created_at=NOW(),released_at=NULL,release_note='' WHERE registration_email_blocks.released_at IS NOT NULL`, service.NormalizeEmailForAliasDedup(email), reason)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n > 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO registration_risk_events(ip_hash,identity_hash,ip_address,user_agent,trigger_path,reason,action) VALUES($1,$2,$3,$4,$5,$6,'denied')`, source.IPHash, source.IdentityHash, source.IPAddress, source.UserAgent, source.Path, reason)
		if err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return service.ErrRegistrationDenied
}

func (r *registrationProtectionRepository) CheckEmailAdmission(ctx context.Context, source service.RegistrationSource, email string) error {
	reason, err := registrationAdmissionReason(ctx, r.db, source, email)
	if err != nil {
		return service.ErrServiceUnavailable
	}
	if reason != "" {
		return r.denyEmail(ctx, source, email, reason)
	}
	if !source.Policy.Enabled {
		return nil
	}
	// Reuse the final-registration quota checks before mail/CAPTCHA work.
	if err := checkRegistrationSourceQuotaQuery(ctx, r.db, source); err != nil {
		if _, ok := err.(*service.RegistrationQuotaExceeded); ok {
			return r.denyEmail(ctx, source, email, "source_quota")
		}
		return service.ErrServiceUnavailable
	}
	return nil
}
