package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type registrationProtectionRepository struct{ db *sql.DB }

func (r *registrationProtectionRepository) CleanupBefore(ctx context.Context, cutoff time.Time, batch int) (int64, error) {
	if batch < 1 || batch > 5000 {
		return 0, fmt.Errorf("invalid registration cleanup batch")
	}
	var total int64
	for _, item := range []struct{ table, condition string }{
		{"registration_risk_events", "created_at < $1"},
		{"registration_source_blocks", "expires_at < $1"},
		{"registration_source_accounts", "created_at < $1 AND NOT EXISTS (SELECT 1 FROM registration_risk_accounts r WHERE r.user_id=registration_source_accounts.user_id AND r.status <> 'released')"},
	} {
		result, err := r.db.ExecContext(ctx, `DELETE FROM `+item.table+` WHERE id IN (SELECT id FROM `+item.table+` WHERE `+item.condition+` ORDER BY id LIMIT $2 FOR UPDATE SKIP LOCKED)`, cutoff, batch)
		if err != nil {
			return total, err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return total, err
		}
		total += count
	}
	return total, nil
}

func (r *registrationProtectionRepository) ReleaseSourceBlock(ctx context.Context, id, actorID int64, note string, clear func(context.Context, string) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var scope, hash string
	err = tx.QueryRowContext(ctx, `SELECT scope,source_hash FROM registration_source_blocks
	 WHERE id=$1 AND released_at IS NULL AND expires_at>NOW() FOR UPDATE`, id).Scan(&scope, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		return service.ErrRegistrationRiskNotFound
	}
	if err != nil {
		return err
	}
	dimension := "identity"
	if scope == "ip" {
		dimension = "ip"
	}
	// A failed counter reset must leave the persistent restriction active.
	if err = clear(ctx, "registration_failure:"+dimension+":"+hash); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE registration_source_blocks SET released_at=NOW(),released_by=$2,release_note=$3 WHERE id=$1`, id, actorID, note)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *registrationProtectionRepository) CheckSourceBlock(ctx context.Context, source service.RegistrationSource) (time.Duration, error) {
	var seconds float64
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(EXTRACT(EPOCH FROM (MAX(expires_at)-NOW())),0)
	 FROM registration_source_blocks WHERE released_at IS NULL AND expires_at > NOW()
	 AND ((scope='ip' AND source_hash=$1) OR (scope='ip_ua' AND source_hash=$2))`, source.IPHash, source.IdentityHash).Scan(&seconds)
	return time.Duration(seconds * float64(time.Second)), err
}

func (r *registrationProtectionRepository) ActivateSourceBlock(ctx context.Context, source service.RegistrationSource, scope string, count int64) error {
	if scope != "ip" && scope != "ip_ua" {
		return fmt.Errorf("invalid registration block scope")
	}
	hash := source.IPHash
	if scope == "ip_ua" {
		hash = source.IdentityHash
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,238))`, scope+":"+hash); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO registration_source_blocks
	 (scope,source_hash,ip_address,user_agent,reason,failure_count,expires_at)
	 SELECT $1::varchar,$2::varchar,$3::text,$4::varchar,'registration_failure_limit',$5::integer,NOW()+($6 * INTERVAL '1 minute')
	 WHERE NOT EXISTS (SELECT 1 FROM registration_source_blocks WHERE scope=$1 AND source_hash=$2 AND released_at IS NULL AND expires_at > NOW())`,
		scope, hash, source.IPAddress, source.UserAgent, count, source.Policy.BlockMinutes)
	if err != nil {
		return err
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if inserted > 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO registration_risk_events
		 (ip_hash,identity_hash,ip_address,user_agent,trigger_path,reason,action,observed_count)
		 VALUES ($1,$2,$3,$4,$5,'registration_failure_limit','blocked',$6)`, source.IPHash, source.IdentityHash, source.IPAddress, source.UserAgent, source.Path, count)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func NewRegistrationProtectionRepository(db *sql.DB) service.RegistrationProtectionRepository {
	return &registrationProtectionRepository{db: db}
}

func (r *registrationProtectionRepository) List(ctx context.Context, f service.RegistrationRiskFilter) (*service.RegistrationRiskPage, error) {
	var query string
	switch f.Kind {
	case "accounts":
		query = `SELECT r.id, r.user_id, u.email, u.concurrency, r.reason, r.status, r.previous_concurrency,
		 r.created_at, r.updated_at, r.reviewed_at, r.reviewed_by, r.review_note,
		 COALESCE(s.ip_address,'') AS ip_address, COALESCE(s.user_agent,'') AS user_agent
		 FROM registration_risk_accounts r JOIN users u ON u.id=r.user_id
		 LEFT JOIN registration_source_accounts s ON s.user_id=r.user_id WHERE u.deleted_at IS NULL`
	case "sources":
		query = `SELECT s.id, s.user_id, COALESCE(u.email,'') AS email, s.ip_address, s.user_agent, s.trigger_path, s.created_at, '' AS status
		 FROM registration_source_accounts s LEFT JOIN users u ON u.id=s.user_id`
	case "events":
		query = `SELECT e.id, e.user_id, COALESCE(u.email,'') AS email, e.ip_address, e.user_agent, e.trigger_path,
		 e.reason, e.action, e.observed_count, e.created_at, '' AS status
		 FROM registration_risk_events e LEFT JOIN users u ON u.id=e.user_id`
	case "blocks":
		query = `SELECT b.id, b.scope, b.ip_address, b.user_agent, b.reason, b.failure_count, b.created_at, b.expires_at,
		 b.released_at, b.released_by, b.release_note, '' AS email,
		 CASE WHEN b.released_at IS NOT NULL THEN 'released' WHEN b.expires_at > NOW() THEN 'active' ELSE 'expired' END AS status
		 FROM registration_source_blocks b`
	default:
		return nil, fmt.Errorf("invalid registration record kind")
	}
	// All query fragments are selected from internal constants; filters remain bound parameters.
	filtered := ` FROM (` + query + `) records WHERE ($1 = '' OR $1 = 'all' OR status = $1)
	 AND ($2 = '' OR ip_address ILIKE '%' || $2 || '%' OR user_agent ILIKE '%' || $2 || '%' OR email ILIKE '%' || $2 || '%')`
	result := &service.RegistrationRiskPage{Page: f.Page, PageSize: f.PageSize}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+filtered, f.Status, f.Query).Scan(&result.Total); err != nil {
		return nil, err
	}
	var raw []byte
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(page)), '[]'::jsonb) FROM (SELECT *`+filtered+` ORDER BY id DESC LIMIT $3 OFFSET $4) page`, f.Status, f.Query, f.PageSize, (f.Page-1)*f.PageSize).Scan(&raw)
	if err != nil {
		return nil, err
	}
	result.Items = raw
	return result, nil
}

func (r *registrationProtectionRepository) ReviewAccount(ctx context.Context, id, actorID int64, action, note string) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	var userID int64
	var current, previous int
	var status, role string
	err = tx.QueryRowContext(ctx, `SELECT r.user_id, u.concurrency, r.previous_concurrency, r.status, u.role
	 FROM registration_risk_accounts r JOIN users u ON u.id=r.user_id
	 WHERE r.id=$1 AND u.deleted_at IS NULL FOR UPDATE OF u,r`, id).Scan(&userID, &current, &previous, &status, &role)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, service.ErrRegistrationRiskNotFound
	}
	if err != nil {
		return 0, err
	}
	if role == service.RoleAdmin {
		return 0, service.ErrRegistrationRiskConflict
	}
	target, next := -1, "restricted"
	switch action {
	case "restrict":
		if status == "restricted" || current < 0 {
			return 0, service.ErrRegistrationRiskConflict
		}
		previous = current
	case "release":
		if status != "restricted" || current != -1 {
			return 0, service.ErrRegistrationRiskConflict
		}
		target, next = previous, "released"
	default:
		return 0, service.ErrRegistrationRiskConflict
	}
	if _, err = tx.ExecContext(ctx, `UPDATE users SET concurrency=$2, updated_at=NOW() WHERE id=$1`, userID, target); err != nil {
		return 0, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE registration_risk_accounts SET status=$2, previous_concurrency=$3, updated_at=NOW(), reviewed_at=NOW(), reviewed_by=$4, review_note=$5 WHERE id=$1`, id, next, previous, actorID, note); err != nil {
		return 0, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO registration_risk_events
	 (user_id,ip_hash,identity_hash,ip_address,user_agent,trigger_path,reason,action)
	 SELECT $1,COALESCE(s.ip_hash,''),COALESCE(s.identity_hash,''),COALESCE(s.ip_address,''),COALESCE(s.user_agent,''),
	 '/admin/registration-protection/accounts','admin_review',$2
	 FROM users u LEFT JOIN registration_source_accounts s ON s.user_id=u.id WHERE u.id=$1`, userID, next); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return userID, nil
}
