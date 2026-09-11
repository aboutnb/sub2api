package repository

import (
	"context"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// The caller holds both source locks until the user transaction commits.
func checkRegistrationSourceQuota(ctx context.Context, client *dbent.Client, source service.RegistrationSource) error {
	if !source.Policy.Enabled {
		return nil
	}
	if err := source.Policy.Validate(); err != nil {
		return err
	}
	window := time.Duration(source.Policy.SuccessWindowHours) * time.Hour
	for _, dim := range []struct {
		column string
		hash   string
		limit  int
	}{
		{"ip_hash", source.IPHash, source.Policy.IPSuccessLimit},
		{"identity_hash", source.IdentityHash, source.Policy.IdentitySuccessLimit},
	} {
		// Column names are internal constants, never request values.
		// The newest limit records determine when a slot becomes available,
		// including when an administrator lowers a previously higher limit.
		rows, err := client.QueryContext(ctx, fmt.Sprintf(`SELECT COUNT(*), COALESCE(EXTRACT(EPOCH FROM (MIN(created_at) + ($2 * INTERVAL '1 second') - CURRENT_TIMESTAMP)), 0) FROM (SELECT created_at FROM registration_source_accounts WHERE %s = $1 AND created_at > CURRENT_TIMESTAMP - ($2 * INTERVAL '1 second') ORDER BY created_at DESC LIMIT $3) recent`, dim.column), dim.hash, int64(window/time.Second), dim.limit)
		if err != nil {
			return err
		}
		var count int
		var remaining float64
		if !rows.Next() {
			err = rows.Err()
			_ = rows.Close()
			if err != nil {
				return err
			}
			return fmt.Errorf("registration source quota query returned no row")
		}
		err = rows.Scan(&count, &remaining)
		closeErr := rows.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if count >= dim.limit {
			return &service.RegistrationQuotaExceeded{Dimension: dim.column, RetryAfter: time.Duration(remaining * float64(time.Second))}
		}
	}
	return nil
}

func recordRegistrationSourceAccount(ctx context.Context, client *dbent.Client, userID int64, source service.RegistrationSource) error {
	_, err := client.ExecContext(ctx, `INSERT INTO registration_source_accounts (user_id, ip_hash, identity_hash, ip_address, user_agent, trigger_path) VALUES ($1,$2,$3,$4,$5,$6)`, userID, source.IPHash, source.IdentityHash, source.IPAddress, source.UserAgent, source.Path)
	if err != nil || !source.Policy.Enabled {
		return err
	}
	// Observation never changes concurrency. The existing signup grant decision
	// may subsequently promote this fresh account to an explicit restriction.
	_, err = client.ExecContext(ctx, `INSERT INTO registration_risk_accounts (user_id,reason,status,previous_concurrency)
	 SELECT $1,'identity_registration_observed','observed',u.concurrency FROM users u WHERE u.id=$1 AND
	 (SELECT COUNT(*) FROM registration_source_accounts WHERE identity_hash=$2 AND created_at > CURRENT_TIMESTAMP - ($3 * INTERVAL '1 hour')) >= $4
	 ON CONFLICT (user_id) DO NOTHING`, userID, source.IdentityHash, source.Policy.SuccessWindowHours, source.Policy.ObserveIdentitySuccessLimit)
	if err != nil {
		return err
	}
	_, err = client.ExecContext(ctx, `INSERT INTO registration_risk_events (user_id,ip_hash,identity_hash,ip_address,user_agent,trigger_path,reason,action,observed_count)
	 SELECT $1::bigint,$2::varchar,$3::varchar,$4::text,$5::varchar,$6::varchar,'identity_registration_observed','observed',
	 (SELECT COUNT(*) FROM registration_source_accounts WHERE identity_hash=$3 AND created_at > CURRENT_TIMESTAMP - ($7 * INTERVAL '1 hour'))
	 WHERE EXISTS (SELECT 1 FROM registration_risk_accounts WHERE user_id=$1 AND status='observed')`,
		userID, source.IPHash, source.IdentityHash, source.IPAddress, source.UserAgent, source.Path, source.Policy.SuccessWindowHours)
	return err
}
