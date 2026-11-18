//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAffiliateRiskRebateGate(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		browserA, browserB     string
		self, sources, blocked bool
	}{
		{name: "same_browser", browserA: "signed-browser", browserB: "signed-browser", sources: true, blocked: true},
		{name: "shared_ip_ua_features", browserA: "browser-a", browserB: "browser-b", sources: true},
		{name: "empty_browser", sources: true},
		{name: "legacy_without_source"},
		{name: "self", self: true, blocked: true},
	} {
		for _, freeze := range []int{0, 24} {
			t.Run(fmt.Sprintf("%s/freeze%d", tc.name, freeze), func(t *testing.T) {
				tx := testEntTx(t)
				ctx := dbent.NewTxContext(context.Background(), tx)
				client := tx.Client()
				repo := NewAffiliateRepository(client, integrationDB)
				create := func(suffix string) int64 {
					u := mustCreateUser(t, client, &service.User{Email: fmt.Sprintf("risk-%s-%d@example.com", suffix, time.Now().UnixNano()), PasswordHash: "hash", Role: service.RoleUser, Status: service.StatusActive, Concurrency: 5, Balance: 100})
					_, err := repo.EnsureUserAffiliate(ctx, u.ID)
					require.NoError(t, err)
					return u.ID
				}
				inviter, invitee := create("inviter"), create("invitee")
				if tc.self {
					invitee = inviter
				}
				if tc.sources {
					for id, browser := range map[int64]string{inviter: tc.browserA, invitee: tc.browserB} {
						_, err := client.ExecContext(ctx, `INSERT INTO registration_source_accounts
 (user_id,ip_hash,identity_hash,ip_address,user_agent,browser_hash,feature_hash)
 VALUES ($1,'shared-ip','shared-identity','203.0.113.1','shared-UA',$2,'shared-features')`, id, browser)
						require.NoError(t, err)
					}
				}
				order, err := client.PaymentOrder.Create().SetUserID(invitee).
					SetUserEmail("risk@example.com").SetUserName("risk-test").
					SetAmount(100).SetPayAmount(100).SetRechargeCode(fmt.Sprintf("risk-%d", invitee)).
					SetPaymentType("test").SetPaymentTradeNo(fmt.Sprintf("risk-%d", invitee)).
					SetClientIP("203.0.113.1").SetSrcHost("test.example.com").
					SetExpiresAt(time.Now().Add(time.Hour)).Save(ctx)
				require.NoError(t, err)
				orderID := order.ID
				applied, err := repo.AccrueQuota(ctx, inviter, invitee, 5, freeze, &orderID)
				require.NoError(t, err)
				require.Equal(t, !tc.blocked, applied)
				expected := 5.0
				if tc.blocked {
					expected = 0
				}
				require.Equal(t, expected, querySingleFloat(t, ctx, client, `SELECT aff_quota+aff_frozen_quota FROM user_affiliates WHERE user_id=$1`, inviter))
				require.Equal(t, expected, querySingleFloat(t, ctx, client, `SELECT aff_history_quota FROM user_affiliates WHERE user_id=$1`, inviter))
				require.Equal(t, 100.0, querySingleFloat(t, ctx, client, `SELECT balance FROM users WHERE id=$1`, invitee))
				if tc.blocked {
					// Retention and callback retries must not undo a recorded decision.
					_, err = client.ExecContext(ctx, `DELETE FROM registration_source_accounts WHERE user_id IN ($1,$2)`, inviter, invitee)
					require.NoError(t, err)
					applied, err = repo.AccrueQuota(ctx, inviter, invitee, 5, freeze, &orderID)
					require.NoError(t, err)
					require.False(t, applied)
					require.Equal(t, 1, querySingleInt(t, ctx, client, `SELECT COUNT(*) FROM payment_audit_logs WHERE order_id=$1 AND action='AFFILIATE_REBATE_RISK_BLOCKED'`, fmt.Sprint(orderID)))
					require.Equal(t, 0, querySingleInt(t, ctx, client, `SELECT COUNT(*) FROM user_affiliate_ledger WHERE user_id=$1`, inviter))
				}
			})
		}
	}
}
