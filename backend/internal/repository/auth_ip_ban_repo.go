package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type authIPBanRepository struct {
	db *sql.DB
}

func NewAuthIPBanRepository(db *sql.DB) service.AuthIPBanRepository {
	return &authIPBanRepository{db: db}
}

const authIPBanSelectColumns = `
  b.id,
  host(b.ip_address),
  b.ban_scope,
  b.ua_hash,
  b.user_agent,
  b.ua_category,
  b.source,
  b.reason,
  b.trigger_path,
  b.target_identifier,
  b.failure_count,
  b.first_seen_at,
  b.last_seen_at,
  b.banned_at,
  b.expires_at,
  b.released_at,
  b.released_by_user_id,
  COALESCE(u.email, ''),
  b.release_note,
  b.ban_count,
  b.created_at,
  b.updated_at`

func scanAuthIPBan(scan func(dest ...any) error) (*service.AuthIPBan, error) {
	record := &service.AuthIPBan{}
	var releasedAt sql.NullTime
	var releasedByUserID sql.NullInt64
	if err := scan(
		&record.ID,
		&record.IPAddress,
		&record.BanScope,
		&record.UAHash,
		&record.UserAgent,
		&record.UACategory,
		&record.Source,
		&record.Reason,
		&record.TriggerPath,
		&record.TargetIdentifier,
		&record.FailureCount,
		&record.FirstSeenAt,
		&record.LastSeenAt,
		&record.BannedAt,
		&record.ExpiresAt,
		&releasedAt,
		&releasedByUserID,
		&record.ReleasedByEmail,
		&record.ReleaseNote,
		&record.BanCount,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if releasedAt.Valid {
		value := releasedAt.Time
		record.ReleasedAt = &value
	}
	if releasedByUserID.Valid {
		value := releasedByUserID.Int64
		record.ReleasedByUserID = &value
	}
	record.Status = authIPBanStatus(record, time.Now().UTC())
	return record, nil
}

func (r *authIPBanRepository) FindActive(ctx context.Context, ipAddress, uaHash string, now time.Time) (*service.AuthIPBan, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil auth IP ban repository")
	}
	query := `SELECT` + authIPBanSelectColumns + `
FROM auth_ip_bans b
LEFT JOIN users u ON u.id = b.released_by_user_id
WHERE b.ip_address = $1::inet
  AND b.released_at IS NULL
  AND b.expires_at > $3
  AND (b.ban_scope = 'ip' OR (b.ban_scope = 'ip_ua' AND b.ua_hash = $2))
ORDER BY CASE WHEN b.ban_scope = 'ip' THEN 0 ELSE 1 END, b.expires_at DESC
LIMIT 1`
	record, err := scanAuthIPBan(r.db.QueryRowContext(ctx, query, ipAddress, uaHash, now.UTC()).Scan)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	record.Status = "active"
	return record, nil
}

func (r *authIPBanRepository) Activate(ctx context.Context, input *service.AuthIPBanActivation) (*service.AuthIPBan, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil auth IP ban repository")
	}
	if input == nil {
		return nil, fmt.Errorf("nil auth IP ban activation")
	}
	query := `
INSERT INTO auth_ip_bans (
  ip_address, ban_scope, ua_hash, user_agent, ua_category, source, reason,
  trigger_path, target_identifier, failure_count, first_seen_at, last_seen_at,
  banned_at, expires_at
) VALUES (
  $1::inet, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
ON CONFLICT (ip_address, ban_scope, ua_hash) DO UPDATE SET
  user_agent = EXCLUDED.user_agent,
  ua_category = EXCLUDED.ua_category,
  source = EXCLUDED.source,
  reason = EXCLUDED.reason,
  trigger_path = EXCLUDED.trigger_path,
  target_identifier = EXCLUDED.target_identifier,
  failure_count = EXCLUDED.failure_count,
  first_seen_at = EXCLUDED.first_seen_at,
  last_seen_at = EXCLUDED.last_seen_at,
  banned_at = EXCLUDED.banned_at,
  expires_at = EXCLUDED.expires_at,
  released_at = NULL,
  released_by_user_id = NULL,
  release_note = '',
  ban_count = auth_ip_bans.ban_count + 1,
  updated_at = NOW()
RETURNING
  id, host(ip_address), ban_scope, ua_hash, user_agent, ua_category, source,
  reason, trigger_path, target_identifier, failure_count, first_seen_at,
  last_seen_at, banned_at, expires_at, released_at, released_by_user_id,
  '', release_note, ban_count, created_at, updated_at`
	record, err := scanAuthIPBan(r.db.QueryRowContext(
		ctx,
		query,
		input.IPAddress,
		input.BanScope,
		input.UAHash,
		input.UserAgent,
		input.UACategory,
		input.Source,
		input.Reason,
		input.TriggerPath,
		input.TargetIdentifier,
		input.FailureCount,
		input.FirstSeenAt.UTC(),
		input.LastSeenAt.UTC(),
		input.BannedAt.UTC(),
		input.ExpiresAt.UTC(),
	).Scan)
	if err != nil {
		return nil, err
	}
	record.Status = "active"
	return record, nil
}

func (r *authIPBanRepository) List(ctx context.Context, filter *service.AuthIPBanFilter) (*service.AuthIPBanList, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil auth IP ban repository")
	}
	if filter == nil {
		filter = &service.AuthIPBanFilter{}
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	clauses := []string{"1=1"}
	args := make([]any, 0, 4)
	switch strings.ToLower(strings.TrimSpace(filter.Status)) {
	case "active":
		clauses = append(clauses, "b.released_at IS NULL AND b.expires_at > NOW()")
	case "expired":
		clauses = append(clauses, "b.released_at IS NULL AND b.expires_at <= NOW()")
	case "released":
		clauses = append(clauses, "b.released_at IS NOT NULL")
	}
	if query := strings.TrimSpace(filter.Query); query != "" {
		args = append(args, "%"+escapeLikePattern(query)+"%")
		placeholder := "$" + itoa(len(args))
		clauses = append(clauses, "(host(b.ip_address) ILIKE "+placeholder+
			" OR b.user_agent ILIKE "+placeholder+
			" OR b.target_identifier ILIKE "+placeholder+
			" OR b.reason ILIKE "+placeholder+")")
	}
	where := "WHERE " + strings.Join(clauses, " AND ")

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM auth_ip_bans b "+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	queryArgs := append(append([]any{}, args...), pageSize, offset)
	query := `SELECT` + authIPBanSelectColumns + `
FROM auth_ip_bans b
LEFT JOIN users u ON u.id = b.released_by_user_id
` + where + `
ORDER BY
  CASE WHEN b.released_at IS NULL AND b.expires_at > NOW() THEN 0
       WHEN b.released_at IS NULL THEN 1 ELSE 2 END,
  b.updated_at DESC, b.id DESC
LIMIT $` + itoa(len(args)+1) + ` OFFSET $` + itoa(len(args)+2)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]*service.AuthIPBan, 0, pageSize)
	for rows.Next() {
		record, err := scanAuthIPBan(rows.Scan)
		if err != nil {
			return nil, err
		}
		items = append(items, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &service.AuthIPBanList{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (r *authIPBanRepository) GetByID(ctx context.Context, id int64) (*service.AuthIPBan, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil auth IP ban repository")
	}
	query := `SELECT` + authIPBanSelectColumns + `
FROM auth_ip_bans b
LEFT JOIN users u ON u.id = b.released_by_user_id
WHERE b.id = $1`
	record, err := scanAuthIPBan(r.db.QueryRowContext(ctx, query, id).Scan)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, service.ErrAuthIPBanNotFound
		}
		return nil, err
	}
	return record, nil
}

func (r *authIPBanRepository) Release(ctx context.Context, id, releasedByUserID int64, note string, now time.Time) (*service.AuthIPBan, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil auth IP ban repository")
	}
	query := `
UPDATE auth_ip_bans SET
  released_at = $2,
  released_by_user_id = NULLIF($3, 0),
  release_note = $4,
  updated_at = $2
WHERE id = $1 AND released_at IS NULL AND expires_at > $2
RETURNING
  id, host(ip_address), ban_scope, ua_hash, user_agent, ua_category, source,
  reason, trigger_path, target_identifier, failure_count, first_seen_at,
  last_seen_at, banned_at, expires_at, released_at, released_by_user_id,
  '', release_note, ban_count, created_at, updated_at`
	record, err := scanAuthIPBan(r.db.QueryRowContext(ctx, query, id, now.UTC(), releasedByUserID, note).Scan)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, service.ErrAuthIPBanNotFound
		}
		return nil, err
	}
	record.Status = "released"
	return record, nil
}

func authIPBanStatus(record *service.AuthIPBan, now time.Time) string {
	if record == nil {
		return ""
	}
	if record.ReleasedAt != nil {
		return "released"
	}
	if !record.ExpiresAt.After(now) {
		return "expired"
	}
	return "active"
}
