//go:build integration

package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func bonusFixture(t *testing.T, old float64) (int64, int64) {
	t.Helper()
	ctx := context.Background()
	u := mustCreateUser(t, integrationEntClient, &service.User{Email: fmt.Sprintf("bonus-%d@example.test", time.Now().UnixNano()), PasswordHash: "hash", Balance: old + 110})
	t.Cleanup(func() {
		for _, table := range []string{"recharge_bonus_writeoff_audit", "recharge_bonus_writeoff_holds", "recharge_bonus_writeoff_batches", "recharge_bonus_writeoff_settings"} {
			_, cleanupErr := integrationDB.Exec("DELETE FROM "+table+" WHERE user_id=$1", u.ID)
			require.NoError(t, cleanupErr)
		}
	})
	o, err := integrationEntClient.PaymentOrder.Create().SetUserID(u.ID).SetUserEmail(u.Email).SetUserName("test").SetClientIP("127.0.0.1").SetSrcHost("test").SetAmount(110).SetPayAmount(100).SetBonusAmount(10).SetRechargeCode(fmt.Sprintf("bonus-%d", time.Now().UnixNano())).SetPaymentType("test").SetPaymentTradeNo("").SetExpiresAt(time.Now().Add(time.Hour)).SetStatus("COMPLETED").Save(ctx)
	require.NoError(t, err)
	_, err = integrationDB.Exec(`INSERT INTO recharge_bonus_writeoff_settings(user_id,enabled,consent_note,updated_by) VALUES($1,true,'consent',1)`, u.ID)
	require.NoError(t, err)
	_, err = integrationDB.Exec(`INSERT INTO recharge_bonus_writeoff_batches(user_id,order_id,principal,bonus,policy_revision) VALUES($1,$2,100,10,1)`, u.ID, o.ID)
	require.NoError(t, err)
	return u.ID, o.ID
}

func bonusDebit(t *testing.T, id int64, cost float64) float64 {
	t.Helper()
	ctx := context.Background()
	tx, err := integrationDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()
	require.NoError(t, markBonusUsage(ctx, tx))
	_, _, err = deductUsageBillingBalance(ctx, tx, id, cost)
	require.NoError(t, err)
	extra, err := applyBonusWriteoff(ctx, tx, id, cost, "test")
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	return extra
}

func TestBonusWriteoffHistoricalFundsAndExactTotal(t *testing.T) {
	id, _ := bonusFixture(t, 20)
	require.Zero(t, bonusDebit(t, id, 20))
	require.Equal(t, 1.0, bonusDebit(t, id, 10))
	require.Equal(t, 9.0, bonusDebit(t, id, 90))
	var balance float64
	require.NoError(t, integrationDB.QueryRow(`SELECT balance FROM users WHERE id=$1`, id).Scan(&balance))
	require.Zero(t, balance)
}

func TestBonusWriteoffLimitAndAdjustment(t *testing.T) {
	id, _ := bonusFixture(t, 0)
	_, err := integrationDB.Exec(`UPDATE recharge_bonus_writeoff_settings SET amount_limit=2 WHERE user_id=$1`, id)
	require.NoError(t, err)
	require.Equal(t, 2.0, bonusDebit(t, id, 30))
	require.Zero(t, bonusDebit(t, id, 10))
	_, err = integrationDB.Exec(`UPDATE users SET balance=balance-1 WHERE id=$1`, id)
	require.NoError(t, err)
	require.Zero(t, bonusDebit(t, id, 10))
	var active int
	require.NoError(t, integrationDB.QueryRow(`SELECT count(*) FROM recharge_bonus_writeoff_batches WHERE user_id=$1 AND active`, id).Scan(&active))
	require.Zero(t, active)
}

func TestBonusWriteoffRefundRestoresOnce(t *testing.T) {
	id, order := bonusFixture(t, 0)
	require.Equal(t, 1.0, bonusDebit(t, id, 10))
	for _, status := range []string{"REFUNDING", "REFUND_FAILED", "REFUNDING"} {
		_, err := integrationDB.Exec(`UPDATE payment_orders SET status=$1 WHERE id=$2`, status, order)
		require.NoError(t, err)
	}
	var balance float64
	require.NoError(t, integrationDB.QueryRow(`SELECT balance FROM users WHERE id=$1`, id).Scan(&balance))
	require.Equal(t, 100.0, balance)
	require.Zero(t, bonusDebit(t, id, 10))
}

func TestBonusWriteoffConcurrentMicroDebits(t *testing.T) {
	id, _ := bonusFixture(t, 0)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				bonusDebit(t, id, 1)
			}
		}()
	}
	wg.Wait()
	var balance, written float64
	require.NoError(t, integrationDB.QueryRow(`SELECT balance FROM users WHERE id=$1`, id).Scan(&balance))
	require.Zero(t, balance)
	require.NoError(t, integrationDB.QueryRow(`SELECT written_off FROM recharge_bonus_writeoff_settings WHERE user_id=$1`, id).Scan(&written))
	require.Equal(t, 10.0, written)
}

func TestBonusWriteoffFrozenFunds(t *testing.T) {
	ctx := context.Background()
	t.Run("historical hold never consumes new principal", func(t *testing.T) {
		id, _ := bonusFixture(t, 0)
		_, err := integrationDB.Exec(`UPDATE users SET frozen_balance=20 WHERE id=$1`, id)
		require.NoError(t, err)
		require.Equal(t, 1.0, bonusDebit(t, id, 10))
		tx, err := integrationDB.BeginTx(ctx, nil)
		require.NoError(t, err)
		defer tx.Rollback()
		result, err := captureUsageBillingBatchImageBalance(ctx, tx, &service.BatchImageBalanceHoldCommand{UserID: id, BatchID: "old", HoldAmount: 20, ActualAmount: 20})
		require.NoError(t, err)
		require.Equal(t, 99.0, *result.NewBalance)
		require.NoError(t, tx.Commit())
	})
	t.Run("new hold protects proportional bonus and settles actual cost", func(t *testing.T) {
		id, _ := bonusFixture(t, 0)
		tx, err := integrationDB.BeginTx(ctx, nil)
		require.NoError(t, err)
		defer tx.Rollback()
		result, err := reserveUsageBillingBatchImageBalance(ctx, tx, &service.BatchImageBalanceHoldCommand{UserID: id, BatchID: "new", HoldAmount: 100})
		require.NoError(t, err)
		require.Zero(t, *result.NewBalance)
		require.Equal(t, 110.0, *result.FrozenBalance)
		require.NoError(t, tx.Commit())
		tx, err = integrationDB.BeginTx(ctx, nil)
		require.NoError(t, err)
		defer tx.Rollback()
		result, err = captureUsageBillingBatchImageBalance(ctx, tx, &service.BatchImageBalanceHoldCommand{UserID: id, BatchID: "new", HoldAmount: 100, ActualAmount: 50, RequestID: "capture-new"})
		require.NoError(t, err)
		require.Equal(t, 55.0, *result.NewBalance)
		require.Zero(t, *result.FrozenBalance)
		require.NoError(t, tx.Commit())
		require.Equal(t, 5.0, bonusDebit(t, id, 50))
	})
}

func TestBonusWriteoffDisabledAndFractional(t *testing.T) {
	id, _ := bonusFixture(t, 0)
	_, err := integrationDB.Exec(`UPDATE recharge_bonus_writeoff_settings SET enabled=false WHERE user_id=$1`, id)
	require.NoError(t, err)
	require.Zero(t, bonusDebit(t, id, 10))
	_, err = integrationDB.Exec(`UPDATE recharge_bonus_writeoff_settings SET enabled=true WHERE user_id=$1`, id)
	require.NoError(t, err)
	require.InDelta(t, 0.00000001, bonusDebit(t, id, 0.0000001), 1e-12)
	require.InDelta(t, 8.99999999, bonusDebit(t, id, 89.9999999), 1e-8)
}

func TestBonusWriteoffEnrollmentBoundaryAndRetry(t *testing.T) {
	id, order := bonusFixture(t, 0)
	_, err := integrationDB.Exec(`DELETE FROM recharge_bonus_writeoff_batches WHERE order_id=$1`, order)
	require.NoError(t, err)
	_, err = integrationDB.Exec(`UPDATE payment_orders SET provider_snapshot='{"recharge_bonus_mode":"bonus"}',order_type='balance' WHERE id=$1`, order)
	require.NoError(t, err)
	enroll := func(want int) {
		t.Helper()
		_, err := integrationDB.Exec(`SELECT register_recharge_bonus_batch($1,(SELECT recharge_code FROM payment_orders WHERE id=$2))`, id, order)
		require.NoError(t, err)
		var n int
		require.NoError(t, integrationDB.QueryRow(`SELECT count(*) FROM recharge_bonus_writeoff_batches WHERE user_id=$1`, id).Scan(&n))
		require.Equal(t, want, n)
	}
	enroll(0) // Existing order predates activation.
	_, err = integrationDB.Exec(`UPDATE payment_orders SET provider_snapshot=provider_snapshot||'{"bonus_writeoff_revision":1}'::jsonb WHERE id=$1`, order)
	require.NoError(t, err)
	_, err = integrationDB.Exec(`UPDATE recharge_bonus_writeoff_settings SET enabled_at=NOW()-INTERVAL '1 day',effective_at=NOW()+INTERVAL '1 day' WHERE user_id=$1`, id)
	require.NoError(t, err)
	enroll(0)
	_, err = integrationDB.Exec(`UPDATE recharge_bonus_writeoff_settings SET effective_at=NOW()-INTERVAL '1 day',enabled=false WHERE user_id=$1`, id)
	require.NoError(t, err)
	enroll(0)
	_, err = integrationDB.Exec(`UPDATE recharge_bonus_writeoff_settings SET enabled=true WHERE user_id=$1`, id)
	require.NoError(t, err)
	enroll(1)
	enroll(1)
}

func TestBonusWriteoffBillingDedup(t *testing.T) {
	id, _ := bonusFixture(t, 0)
	ctx := context.Background()
	key := mustCreateApiKey(t, integrationEntClient, &service.APIKey{UserID: id, Key: fmt.Sprintf("sk-bonus-%d", id), Name: "bonus"})
	account := mustCreateAccount(t, integrationEntClient, &service.Account{Name: fmt.Sprintf("bonus-%d", id), Type: service.AccountTypeAPIKey})
	t.Cleanup(func() {
		_, err := integrationDB.Exec(`UPDATE accounts SET status='disabled' WHERE id=$1`, account.ID)
		require.NoError(t, err)
	})
	repo := NewUsageBillingRepository(integrationEntClient, integrationDB)
	command := &service.UsageBillingCommand{UserID: id, APIKeyID: key.ID, AccountID: account.ID, AccountType: service.AccountTypeAPIKey, RequestID: fmt.Sprintf("bonus-request-%d", id), BalanceCost: 10, APIKeyQuotaCost: 10}
	result, err := repo.Apply(ctx, command)
	require.NoError(t, err)
	require.Equal(t, 1.0, result.BonusWrittenOff)
	result, err = repo.Apply(ctx, command)
	require.NoError(t, err)
	require.False(t, result.Applied)
	command.RequestID += "-batch"
	_, err = repo.ApplyBatch(ctx, []*service.UsageBillingCommand{command, command})
	require.Error(t, err)
	results, err := repo.ApplyBatch(ctx, []*service.UsageBillingCommand{command})
	require.NoError(t, err)
	require.Len(t, results, 1)
	_, err = repo.ApplyBatch(ctx, []*service.UsageBillingCommand{command})
	require.NoError(t, err)
	var balance, quota float64
	require.NoError(t, integrationDB.QueryRow(`SELECT balance FROM users WHERE id=$1`, id).Scan(&balance))
	require.Equal(t, 88.0, balance)
	require.NoError(t, integrationDB.QueryRow(`SELECT quota_used FROM api_keys WHERE id=$1`, key.ID).Scan(&quota))
	require.Equal(t, 20.0, quota)
}

func TestBonusWriteoffNewOrderSnapshotAndMixedRatios(t *testing.T) {
	id, _ := bonusFixture(t, 0)
	ctx := context.Background()
	code := fmt.Sprintf("bonus-mixed-%d", id)
	o, err := integrationEntClient.PaymentOrder.Create().SetUserID(id).SetUserEmail("mixed@example.test").SetUserName("test").SetClientIP("127.0.0.1").SetSrcHost("test").SetAmount(120).SetPayAmount(100).SetBonusAmount(20).SetRechargeCode(code).SetPaymentType("test").SetPaymentTradeNo("").SetExpiresAt(time.Now().Add(time.Hour)).SetStatus("COMPLETED").SetOrderType("balance").SetProviderSnapshot(map[string]any{"recharge_bonus_mode": "bonus"}).Save(ctx)
	require.NoError(t, err)
	_, err = integrationDB.Exec(`UPDATE users SET balance=balance+120 WHERE id=$1`, id)
	require.NoError(t, err)
	_, err = integrationDB.Exec(`SELECT register_recharge_bonus_batch($1,$2)`, id, code)
	require.NoError(t, err)
	var count int
	require.NoError(t, integrationDB.QueryRow(`SELECT count(*) FROM recharge_bonus_writeoff_batches WHERE order_id=$1`, o.ID).Scan(&count))
	require.Equal(t, 1, count)
	require.Equal(t, 20.0, bonusDebit(t, id, 150))
	require.Equal(t, 10.0, bonusDebit(t, id, 50))
}

func TestBonusWriteoffPolicySaveRetiresOldBatches(t *testing.T) {
	id, _ := bonusFixture(t, 0)
	require.Equal(t, 1.0, bonusDebit(t, id, 10))
	repo := NewUserRepository(integrationEntClient, integrationDB).(service.BonusWriteoffRepository)
	require.NoError(t, repo.SetBonusWriteoff(context.Background(), id, 1, service.BonusWriteoffPolicy{Enabled: true, EffectiveAt: time.Now().Add(-time.Hour), ConsentNote: "renewed agreement"}))
	require.Zero(t, bonusDebit(t, id, 10))
	data, err := repo.GetBonusWriteoff(context.Background(), id, 1)
	require.NoError(t, err)
	require.Contains(t, string(data), "renewed agreement")
	var active int
	require.NoError(t, integrationDB.QueryRow(`SELECT count(*) FROM recharge_bonus_writeoff_batches WHERE user_id=$1 AND active`, id).Scan(&active))
	require.Zero(t, active)
}

func TestBonusWriteoffHoldCancellationAndInsufficientFunds(t *testing.T) {
	id, _ := bonusFixture(t, 0)
	ctx := context.Background()
	key := mustCreateApiKey(t, integrationEntClient, &service.APIKey{UserID: id, Key: fmt.Sprintf("sk-hold-bonus-%d", id), Name: "bonus"})
	repo := NewUsageBillingRepository(integrationEntClient, integrationDB)
	batch := fmt.Sprintf("hold-bonus-%d", id)
	cmd := &service.BatchImageBalanceHoldCommand{UserID: id, APIKeyID: key.ID, BatchID: batch, RequestID: service.BatchImageHoldRequestID(batch), HoldAmount: 110}
	_, err := repo.ReserveBatchImageBalance(ctx, cmd)
	require.ErrorIs(t, err, service.ErrBatchImageInsufficientBalance)
	cmd.HoldAmount = 100
	result, err := repo.ReserveBatchImageBalance(ctx, cmd)
	require.NoError(t, err)
	require.Equal(t, 110.0, *result.FrozenBalance)
	cmd.RequestID += "-release"
	result, err = repo.ReleaseBatchImageBalance(ctx, cmd)
	require.NoError(t, err)
	require.Equal(t, 110.0, *result.NewBalance)
	require.Zero(t, *result.FrozenBalance)
	result, err = repo.ReleaseBatchImageBalance(ctx, cmd)
	require.NoError(t, err)
	require.False(t, result.Applied)
	require.Equal(t, 10.0, bonusDebit(t, id, 100))
}
