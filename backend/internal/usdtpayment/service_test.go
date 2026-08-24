package usdtpayment

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/shopspring/decimal"
)

func quoteRows(q *Quote) *sqlmock.Rows {
	return sqlmock.NewRows(splitColumns(quoteColumns)).AddRow(
		q.ID, q.PaymentOrderID, q.MerchantOrderID, q.ProviderTradeID,
		q.FiatCurrency, q.FiatAmount, q.CryptoCurrency, q.Network, q.TradeType, q.CryptoAmount,
		q.ExchangeRate, q.ReceivingAddress, q.PaymentURL, q.UpstreamCreatedAt, q.UpstreamExpiresAt,
		q.ProviderStatus, nil, nil, nil, nil, q.NextReconcileAt, q.ReconcileAttempts, nil, nil,
		q.CreatedAt, q.UpdatedAt,
	)
}

type testPaymentBridge struct {
	confirmErr error
	confirmed  int
}

func (b *testPaymentBridge) PrepareUSDTOrder(context.Context, PrepareOrderRequest) (*PreparedOrder, error) {
	return nil, errors.New("not implemented")
}

func (b *testPaymentBridge) ConfirmUSDTOrder(context.Context, ConfirmOrderRequest) error {
	b.confirmed++
	return b.confirmErr
}

func (b *testPaymentBridge) FailUSDTOrderBeforeQuote(context.Context, int64, error) error {
	return nil
}

func testQuoteAndProof(now time.Time) (*Quote, *UpstreamOrder) {
	quote := &Quote{
		ID:                10,
		PaymentOrderID:    20,
		MerchantOrderID:   "sub2_order",
		ProviderTradeID:   "trade_1",
		FiatCurrency:      "CNY",
		FiatAmount:        "102",
		CryptoCurrency:    "USDT",
		Network:           "tron",
		TradeType:         "usdt.trc20",
		CryptoAmount:      "14.5714",
		ReceivingAddress:  "TAddress",
		UpstreamCreatedAt: now.Add(-time.Minute),
		UpstreamExpiresAt: now.Add(time.Minute),
	}
	proof := &UpstreamOrder{
		OrderID:            quote.MerchantOrderID,
		TradeID:            quote.ProviderTradeID,
		StatusName:         "succeeded",
		Fiat:               quote.FiatCurrency,
		Amount:             quote.FiatAmount,
		Crypto:             quote.CryptoCurrency,
		ActualAmount:       quote.CryptoAmount,
		TradeType:          quote.TradeType,
		Network:            quote.Network,
		Token:              quote.ReceivingAddress,
		BlockTransactionID: "0xhash",
		TransferAt:         now.Unix(),
		Confirmation:       map[string]any{"confirmed": true, "block_number": float64(123)},
	}
	return quote, proof
}

func TestConfirmProofDoesNotMarkQuoteSucceededWhenBridgeFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	}()

	bridge := &testPaymentBridge{confirmErr: errors.New("temporary fulfillment failure")}
	service := &Service{config: testUSDTConfig(), repository: NewRepository(db), bridge: bridge}
	quote, proof := testQuoteAndProof(time.Now().UTC().Truncate(time.Second))
	expectation := mock.ExpectExec(regexp.QuoteMeta("UPDATE usdt_payment_quotes"))
	expectation = expectation.WithArgs(proof.BlockTransactionID, time.Unix(proof.TransferAt, 0), int64(123), quote.ID)
	expectation.WillReturnResult(sqlmock.NewResult(0, 1))

	if err := service.confirmProof(context.Background(), quote, proof); err == nil {
		t.Fatal("expected bridge failure")
	}
	if bridge.confirmed != 1 {
		t.Fatalf("expected one bridge attempt, got %d", bridge.confirmed)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("quote was persisted despite fulfillment failure: %v", err)
	}
}

func TestConfirmProofDoesNotFulfillWhenTransactionHashReservationFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	}()

	bridge := &testPaymentBridge{}
	service := &Service{config: testUSDTConfig(), repository: NewRepository(db), bridge: bridge}
	quote, proof := testQuoteAndProof(time.Now().UTC().Truncate(time.Second))
	expectation := mock.ExpectExec(regexp.QuoteMeta("UPDATE usdt_payment_quotes"))
	expectation = expectation.WithArgs(proof.BlockTransactionID, time.Unix(proof.TransferAt, 0), int64(123), quote.ID)
	expectation.WillReturnError(errors.New("duplicate transaction hash"))

	if err := service.confirmProof(context.Background(), quote, proof); err == nil {
		t.Fatal("expected transaction hash reservation failure")
	}
	if bridge.confirmed != 0 {
		t.Fatalf("bridge ran before transaction hash reservation: %d", bridge.confirmed)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfirmProofRecordsQuoteAfterSuccessfulBridge(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	}()

	bridge := &testPaymentBridge{}
	service := &Service{config: testUSDTConfig(), repository: NewRepository(db), bridge: bridge}
	now := time.Now().UTC().Truncate(time.Second)
	quote, proof := testQuoteAndProof(now)
	for _, query := range []string{"UPDATE usdt_payment_quotes", "UPDATE usdt_payment_quotes SET provider_status='succeeded'"} {
		expectation := mock.ExpectExec(regexp.QuoteMeta(query))
		expectation = expectation.WithArgs(proof.BlockTransactionID, time.Unix(proof.TransferAt, 0), int64(123), quote.ID)
		expectation.WillReturnResult(sqlmock.NewResult(0, 1))
	}

	if err := service.confirmProof(context.Background(), quote, proof); err != nil {
		t.Fatalf("confirm proof: %v", err)
	}
	if bridge.confirmed != 1 {
		t.Fatalf("expected one bridge attempt, got %d", bridge.confirmed)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfirmProofRejectsFrozenAmountAndNetworkMismatches(t *testing.T) {
	service := &Service{config: testUSDTConfig(), bridge: &testPaymentBridge{}}
	quote, proof := testQuoteAndProof(time.Now().UTC().Truncate(time.Second))

	proof.Amount = "103"
	if err := service.confirmProof(context.Background(), quote, proof); err == nil {
		t.Fatal("accepted a fiat amount mismatch")
	}
	proof.Amount = quote.FiatAmount
	proof.Network = "bsc"
	if err := service.confirmProof(context.Background(), quote, proof); err == nil {
		t.Fatal("accepted a network mismatch")
	}
}

func TestSameFrozenQuoteUsesExactFinancialAndRoutingFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	left, _ := testQuoteAndProof(now)
	left.PaymentURL = "https://pay.example/checkout/trade_1"
	left.ExchangeRate = "7.2000"
	right := *left
	right.FiatAmount = "102.00"
	right.CryptoAmount = "14.571400"
	right.ExchangeRate = "7.2"
	if !sameFrozenQuote(left, &right) {
		t.Fatal("equivalent decimal encodings should match")
	}

	right.Network = "bsc"
	if sameFrozenQuote(left, &right) {
		t.Fatal("network mismatch was accepted")
	}
	right = *left
	right.ProviderTradeID = "trade_2"
	if sameFrozenQuote(left, &right) {
		t.Fatal("provider trade ID mismatch was accepted")
	}
	right = *left
	right.ReceivingAddress = "different-address"
	if sameFrozenQuote(left, &right) {
		t.Fatal("receiving address mismatch was accepted")
	}
}

func TestParseRequestedUSDTAmountIsDecimalExact(t *testing.T) {
	amount, err := parseRequestedUSDTAmount(" 10.12345678 ")
	if err != nil || amount.String() != "10.12345678" {
		t.Fatalf("amount=%s err=%v", amount, err)
	}
	for _, raw := range []string{"", "0", "-1", "1.123456789", "not-a-number"} {
		if _, err := parseRequestedUSDTAmount(raw); err == nil {
			t.Fatalf("accepted invalid amount %q", raw)
		}
	}
}

func TestValidateMinimumUSDTAmount(t *testing.T) {
	if err := validateMinimumUSDTAmount(decimal.RequireFromString("4.99999999"), config.USDTPaymentConfig{MinimumAmount: 5}); err == nil {
		t.Fatal("accepted amount below configured minimum")
	}
	if err := validateMinimumUSDTAmount(decimal.RequireFromString("5"), config.USDTPaymentConfig{MinimumAmount: 5}); err != nil {
		t.Fatalf("rejected exact minimum: %v", err)
	}
	if got := minimumUSDTAmount(config.USDTPaymentConfig{}).String(); got != "5" {
		t.Fatalf("default minimum = %s, want 5", got)
	}
}

func TestReconcileScheduleStaysFastForPendingAndSlowsAfterExpiry(t *testing.T) {
	service := &Service{config: config.USDTPaymentConfig{ReconcileIntervalSeconds: 2}}
	now := time.Now()
	quote := &Quote{UpstreamCreatedAt: now.Add(-time.Minute)}
	if got := service.reconcileStateDelay(quote, "waiting"); got != 2*time.Second {
		t.Fatalf("fresh pending delay = %v", got)
	}
	quote.UpstreamCreatedAt = now.Add(-5 * time.Minute)
	if got := service.reconcileStateDelay(quote, "waiting"); got != 5*time.Second {
		t.Fatalf("mature pending delay = %v", got)
	}
	if got := service.reconcileStateDelay(quote, "expired"); got != 15*time.Second {
		t.Fatalf("late-payment delay = %v", got)
	}
	if got := service.reconcileErrorDelay(100); got != 30*time.Second {
		t.Fatalf("error backoff = %v", got)
	}
}

func TestSaveQuoteRejectsConflictingExistingQuote(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	}()
	now := time.Now().UTC().Truncate(time.Second)
	existing, _ := testQuoteAndProof(now)
	existing.ID = 1
	existing.PaymentURL = "https://pay.example/trade_1"
	existing.ExchangeRate = "7"
	existing.ProviderStatus = "waiting"
	existing.NextReconcileAt = now
	existing.CreatedAt = now
	existing.UpdatedAt = now
	incoming := *existing
	incoming.Network = "bsc"

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO usdt_payment_quotes").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT[\\s\\S]*FROM usdt_payment_quotes WHERE payment_order_id=\\$1 FOR UPDATE").
		WithArgs(existing.PaymentOrderID).WillReturnRows(quoteRows(existing))
	mock.ExpectRollback()

	err = NewRepository(db).SaveQuote(context.Background(), &incoming)
	if err == nil || !regexp.MustCompile("conflicts").MatchString(err.Error()) {
		t.Fatalf("error = %v, want frozen quote conflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSaveQuoteAcceptsIdenticalExistingQuote(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
	}()
	now := time.Now().UTC().Truncate(time.Second)
	quote, _ := testQuoteAndProof(now)
	quote.ID = 1
	quote.PaymentURL = "https://pay.example/trade_1"
	quote.ExchangeRate = "7"
	quote.ProviderStatus = "waiting"
	quote.NextReconcileAt = now
	quote.CreatedAt = now
	quote.UpdatedAt = now

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO usdt_payment_quotes").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT[\\s\\S]*FROM usdt_payment_quotes WHERE payment_order_id=\\$1 FOR UPDATE").
		WithArgs(quote.PaymentOrderID).WillReturnRows(quoteRows(quote))
	mock.ExpectExec("UPDATE payment_orders").
		WithArgs("", quote.UpstreamExpiresAt, PaymentType, ProviderKey, sqlmock.AnyArg(), quote.PaymentOrderID, quote.MerchantOrderID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := NewRepository(db).SaveQuote(context.Background(), quote); err != nil {
		t.Fatalf("save identical quote: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestIsUpstreamOrderNotFound(t *testing.T) {
	for _, err := range []error{
		errors.New("BEpusdt cashier info returned 400: order not found"),
		errors.New("BEpusdt cashier info returned 400: 订单不存在"),
	} {
		if !isUpstreamOrderNotFound(err) {
			t.Fatalf("did not classify terminal upstream error: %v", err)
		}
	}
	if isUpstreamOrderNotFound(errors.New("BEpusdt cashier HTTP 502")) {
		t.Fatal("classified an upstream transport error as terminal")
	}
}
