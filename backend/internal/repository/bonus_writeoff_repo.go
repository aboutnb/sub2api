package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *userRepository) GetBonusWriteoff(ctx context.Context, id int64, page int) (json.RawMessage, error) {
	rows, err := r.sql.QueryContext(ctx, `SELECT jsonb_build_object(
 'policy',(SELECT to_jsonb(s) FROM recharge_bonus_writeoff_settings s WHERE user_id=$1),
 'batches',COALESCE((SELECT jsonb_agg(b ORDER BY b.id DESC) FROM
 (SELECT * FROM recharge_bonus_writeoff_batches WHERE user_id=$1 ORDER BY id DESC LIMIT 50) b),'[]'::jsonb),
 'total',(SELECT count(*) FROM recharge_bonus_writeoff_audit WHERE user_id=$1),
 'items',COALESCE((SELECT jsonb_agg(a ORDER BY a.id DESC) FROM
 (SELECT * FROM recharge_bonus_writeoff_audit WHERE user_id=$1 ORDER BY id DESC LIMIT 50 OFFSET $2) a),'[]'::jsonb))
 FROM users WHERE id=$1 AND deleted_at IS NULL`, id, (page-1)*50)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrUserNotFound
	}
	var data []byte
	err = rows.Scan(&data)
	return json.RawMessage(data), err
}

func (r *userRepository) SetBonusWriteoff(ctx context.Context, id, actor int64, p service.BonusWriteoffPolicy) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	c := tx.Client()
	rows, err := c.QueryContext(ctx, `SELECT id FROM users WHERE id=$1 AND deleted_at IS NULL FOR UPDATE`, id)
	if err != nil {
		return err
	}
	found := rows.Next()
	rowErr := rows.Err()
	rows.Close()
	if rowErr != nil {
		return rowErr
	}
	if !found {
		return service.ErrUserNotFound
	}
	// Retiring old batches on every policy edit avoids retroactive catch-up or changing existing agreements.
	_, err = c.ExecContext(ctx, `UPDATE recharge_bonus_writeoff_batches SET active=false WHERE user_id=$1 AND active`, id)
	if err != nil {
		return err
	}
	_, err = c.ExecContext(ctx, `INSERT INTO recharge_bonus_writeoff_settings(user_id,enabled,effective_at,amount_limit,consent_note,updated_by)
 VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(user_id) DO UPDATE SET
 enabled=EXCLUDED.enabled,effective_at=EXCLUDED.effective_at,enabled_at=NOW(),
 amount_limit=EXCLUDED.amount_limit,written_off=0,consent_note=EXCLUDED.consent_note,
 updated_by=EXCLUDED.updated_by,revision=recharge_bonus_writeoff_settings.revision+1,updated_at=NOW()`, id, p.Enabled, p.EffectiveAt, p.AmountLimit, p.ConsentNote, actor)
	if err != nil {
		return err
	}
	detail, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = c.ExecContext(ctx, `INSERT INTO recharge_bonus_writeoff_audit(user_id,action,actor_id,detail) VALUES($1,'policy',$2,$3::jsonb)`, id, actor, string(detail))
	if err != nil {
		return err
	}
	return tx.Commit()
}

func applyBonusWriteoff(ctx context.Context, tx *sql.Tx, id int64, cost float64, key string) (float64, error) {
	var extra float64
	err := tx.QueryRowContext(ctx, `SELECT apply_recharge_bonus_writeoff($1,$2,$3)`, id, cost, key).Scan(&extra)
	return extra, err
}

func markBonusUsage(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `SELECT set_config('aivoza.bonus_usage','on',true)`)
	return err
}

// The legacy usage service already owns a transaction and its usage-log dedup key.
func (r *userRepository) ApplyBonusWriteoff(ctx context.Context, id int64, cost float64, key string) error {
	c := clientFromContext(ctx, r.client)
	if c == r.client {
		return errors.New("bonus writeoff requires an existing transaction")
	}
	_, err := c.ExecContext(ctx, `SELECT apply_recharge_bonus_writeoff($1,$2,$3)`, id, cost, key)
	return err
}
