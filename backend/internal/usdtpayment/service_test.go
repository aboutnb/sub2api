package usdtpayment

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

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
	defer db.Close()

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
	defer db.Close()

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
	defer db.Close()

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
