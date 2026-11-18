package repository

import (
	"context"
	"encoding/json"
	"fmt"

	dbent "github.com/Wei-Shaw/sub2api/ent"
)

// Only a server-issued browser identity is strong enough for this rule.
// Shared IPs, user agents and client-supplied features alone are not evidence.
func blockAffiliateSelfReferral(ctx context.Context, client *dbent.Client, inviterID, inviteeID int64, orderID *int64) (bool, error) {
	rows, err := client.QueryContext(ctx, `
SELECT CASE
 WHEN $1::bigint = $2::bigint THEN 'affiliate_self_referral'
 WHEN EXISTS (SELECT 1 FROM affiliate_risk_pairs WHERE inviter_id=$1 AND invitee_id=$2)
 THEN 'affiliate_same_browser'
 WHEN EXISTS (
   SELECT 1 FROM registration_source_accounts a
   JOIN registration_source_accounts b ON a.browser_hash=b.browser_hash
   WHERE a.user_id=$1 AND b.user_id=$2 AND a.browser_hash<>''
 ) THEN 'affiliate_same_browser'
 ELSE '' END`, inviterID, inviteeID)
	if err != nil {
		return false, fmt.Errorf("check affiliate risk: %w", err)
	}
	var reason string
	if rows.Next() {
		err = rows.Scan(&reason)
	}
	readErr := rows.Err()
	closeErr := rows.Close()
	if err != nil {
		return false, err
	}
	if readErr != nil {
		return false, readErr
	}
	if closeErr != nil {
		return false, closeErr
	}
	if reason == "" {
		return false, nil
	}

	// Keep the pair decision after registration-source retention expires.
	_, err = client.ExecContext(ctx, `INSERT INTO affiliate_risk_pairs (inviter_id,invitee_id,reason)
 VALUES ($1,$2,$3) ON CONFLICT (inviter_id,invitee_id) DO NOTHING`, inviterID, inviteeID, reason)
	if err != nil {
		return false, fmt.Errorf("record affiliate risk pair: %w", err)
	}
	if orderID != nil {
		message := "邀请双方注册使用同一浏览器标识，本订单不发放邀请返利；充值到账不受影响"
		if reason == "affiliate_self_referral" {
			message = "邀请人与被邀请人为同一账号，本订单不发放邀请返利；充值到账不受影响"
		}
		detail, err := json.Marshal(map[string]any{"reason": reason, "message": message, "inviter_id": inviterID, "invitee_id": inviteeID})
		if err != nil {
			return false, err
		}
		_, err = client.ExecContext(ctx, `INSERT INTO payment_audit_logs (order_id,action,detail,operator,created_at)
 VALUES ($1,'AFFILIATE_REBATE_RISK_BLOCKED',$2,'system',NOW()) ON CONFLICT (order_id,action) DO NOTHING`, fmt.Sprint(*orderID), string(detail))
		if err != nil {
			return false, fmt.Errorf("audit affiliate risk: %w", err)
		}
	}
	return true, nil
}
