package usdtpayment

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const quoteColumns = `id, payment_order_id, merchant_order_id, provider_trade_id,
fiat_currency, fiat_amount, crypto_currency, network, trade_type, crypto_amount,
exchange_rate, receiving_address, payment_url, upstream_created_at, upstream_expires_at,
provider_status, transaction_hash, chain_transfer_at, block_number, last_reconciled_at,
next_reconcile_at, reconcile_attempts, reconcile_lease_until, last_error, created_at, updated_at`

type rowScanner interface {
	Scan(...any) error
}

func scanQuote(scanner rowScanner) (*Quote, error) {
	var q Quote
	if err := scanner.Scan(
		&q.ID, &q.PaymentOrderID, &q.MerchantOrderID, &q.ProviderTradeID,
		&q.FiatCurrency, &q.FiatAmount, &q.CryptoCurrency, &q.Network, &q.TradeType, &q.CryptoAmount,
		&q.ExchangeRate, &q.ReceivingAddress, &q.PaymentURL, &q.UpstreamCreatedAt, &q.UpstreamExpiresAt,
		&q.ProviderStatus, &q.TransactionHash, &q.ChainTransferAt, &q.BlockNumber, &q.LastReconciledAt,
		&q.NextReconcileAt, &q.ReconcileAttempts, &q.ReconcileLeaseUntil, &q.LastError, &q.CreatedAt, &q.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *Repository) SaveQuote(ctx context.Context, q *Quote) error {
	if r == nil || r.db == nil || q == nil {
		return errors.New("USDT quote repository is unavailable")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `
INSERT INTO usdt_payment_quotes (
    payment_order_id, merchant_order_id, provider_trade_id, fiat_currency, fiat_amount,
    crypto_currency, network, trade_type, crypto_amount, exchange_rate, receiving_address,
    payment_url, upstream_created_at, upstream_expires_at, provider_status, next_reconcile_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,NOW())
ON CONFLICT (payment_order_id) DO NOTHING`,
		q.PaymentOrderID, q.MerchantOrderID, q.ProviderTradeID, q.FiatCurrency, q.FiatAmount,
		q.CryptoCurrency, q.Network, q.TradeType, q.CryptoAmount, q.ExchangeRate, q.ReceivingAddress,
		q.PaymentURL, q.UpstreamCreatedAt, q.UpstreamExpiresAt, q.ProviderStatus)
	if err != nil {
		return fmt.Errorf("insert USDT quote: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
UPDATE payment_orders
SET pay_url=$1, expires_at=$2, payment_type=$3, provider_key=$4,
    provider_snapshot=$5::jsonb, updated_at=NOW()
WHERE id=$6 AND out_trade_no=$7 AND status='PENDING'`,
		q.PaymentURL, q.UpstreamExpiresAt, PaymentType, ProviderKey,
		`{"schema_version":1,"provider_key":"bepusdt","payment_mode":"redirect","currency":"CNY"}`,
		q.PaymentOrderID, q.MerchantOrderID)
	if err != nil {
		return fmt.Errorf("attach USDT quote to payment order: %w", err)
	}
	return tx.Commit()
}

func (r *Repository) GetQuoteForUser(ctx context.Context, paymentOrderID, userID int64) (*Quote, string, float64, float64, float64, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT `+prefixColumns("q", quoteColumns)+`, o.status, o.amount, o.pay_amount, o.fee_rate
FROM usdt_payment_quotes q
JOIN payment_orders o ON o.id=q.payment_order_id
WHERE q.payment_order_id=$1 AND o.user_id=$2`, paymentOrderID, userID)
	var status string
	var amount, payAmount, feeRate float64
	q, err := scanQuoteWithOrder(row, &status, &amount, &payAmount, &feeRate)
	return q, status, amount, payAmount, feeRate, err
}

func scanQuoteWithOrder(scanner rowScanner, status *string, amount, payAmount, feeRate *float64) (*Quote, error) {
	var q Quote
	err := scanner.Scan(
		&q.ID, &q.PaymentOrderID, &q.MerchantOrderID, &q.ProviderTradeID,
		&q.FiatCurrency, &q.FiatAmount, &q.CryptoCurrency, &q.Network, &q.TradeType, &q.CryptoAmount,
		&q.ExchangeRate, &q.ReceivingAddress, &q.PaymentURL, &q.UpstreamCreatedAt, &q.UpstreamExpiresAt,
		&q.ProviderStatus, &q.TransactionHash, &q.ChainTransferAt, &q.BlockNumber, &q.LastReconciledAt,
		&q.NextReconcileAt, &q.ReconcileAttempts, &q.ReconcileLeaseUntil, &q.LastError, &q.CreatedAt, &q.UpdatedAt,
		status, amount, payAmount, feeRate,
	)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func prefixColumns(prefix, columns string) string {
	result := ""
	for i, column := range splitColumns(columns) {
		if i > 0 {
			result += ","
		}
		result += prefix + "." + column
	}
	return result
}

func splitColumns(columns string) []string {
	var result []string
	current := ""
	for _, ch := range columns {
		if ch == ',' {
			result = append(result, trimSpace(current))
			current = ""
			continue
		}
		current += string(ch)
	}
	return append(result, trimSpace(current))
}

func trimSpace(value string) string {
	start, end := 0, len(value)
	for start < end && (value[start] == ' ' || value[start] == '\n' || value[start] == '\t') {
		start++
	}
	for end > start && (value[end-1] == ' ' || value[end-1] == '\n' || value[end-1] == '\t') {
		end--
	}
	return value[start:end]
}

func (r *Repository) GetQuoteByMerchantOrderID(ctx context.Context, orderID string) (*Quote, error) {
	return scanQuote(r.db.QueryRowContext(ctx, `SELECT `+quoteColumns+` FROM usdt_payment_quotes WHERE merchant_order_id=$1`, orderID))
}

func (r *Repository) ClaimDueQuotes(ctx context.Context, limit int, lease time.Duration) ([]*Quote, error) {
	rows, err := r.db.QueryContext(ctx, `
WITH due AS (
    SELECT id FROM usdt_payment_quotes
    WHERE provider_status IN ('waiting','confirming')
      AND next_reconcile_at <= NOW()
      AND (reconcile_lease_until IS NULL OR reconcile_lease_until < NOW())
    ORDER BY next_reconcile_at
    FOR UPDATE SKIP LOCKED
    LIMIT $1
)
UPDATE usdt_payment_quotes q
SET reconcile_lease_until=NOW()+($2 * INTERVAL '1 second'),
    reconcile_attempts=q.reconcile_attempts+1, updated_at=NOW()
FROM due WHERE q.id=due.id
RETURNING `+prefixColumns("q", quoteColumns), limit, int(lease.Seconds()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var quotes []*Quote
	for rows.Next() {
		q, err := scanQuote(rows)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, q)
	}
	return quotes, rows.Err()
}

func (r *Repository) RecordReconcile(ctx context.Context, id int64, status string, next time.Time, reconcileErr error) error {
	var errorText any
	if reconcileErr != nil {
		errorText = reconcileErr.Error()
	}
	_, err := r.db.ExecContext(ctx, `
UPDATE usdt_payment_quotes SET provider_status=$1, last_reconciled_at=NOW(),
next_reconcile_at=$2, reconcile_lease_until=NULL, last_error=$3, updated_at=NOW()
WHERE id=$4`, status, next, errorText, id)
	return err
}

func (r *Repository) RecordConfirmed(ctx context.Context, q *Quote, txHash string, transferAt time.Time, blockNumber int64) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE usdt_payment_quotes SET provider_status='succeeded', transaction_hash=$1,
chain_transfer_at=$2, block_number=$3, last_reconciled_at=NOW(), reconcile_lease_until=NULL,
last_error=NULL, updated_at=NOW() WHERE id=$4`, txHash, transferAt, blockNumber, q.ID)
	if err != nil {
		return fmt.Errorf("record confirmed USDT quote: %w", err)
	}
	return nil
}

// ReserveConfirmation binds a chain transaction to the frozen quote before
// local fulfillment. The unique transaction-hash index then rejects a proof
// that was already used by another order without crediting that order first.
func (r *Repository) ReserveConfirmation(ctx context.Context, q *Quote, txHash string, transferAt time.Time, blockNumber int64) error {
	if r == nil || r.db == nil || q == nil {
		return errors.New("USDT quote repository is unavailable")
	}
	result, err := r.db.ExecContext(ctx, `
UPDATE usdt_payment_quotes
SET provider_status=CASE WHEN provider_status='succeeded' THEN provider_status ELSE 'confirming' END,
    transaction_hash=$1, chain_transfer_at=$2, block_number=$3,
    last_reconciled_at=NOW(), reconcile_lease_until=NULL, last_error=NULL, updated_at=NOW()
WHERE id=$4 AND (transaction_hash IS NULL OR transaction_hash=$1)`,
		txHash, transferAt, blockNumber, q.ID)
	if err != nil {
		return fmt.Errorf("reserve USDT confirmation: %w", err)
	}
	if rows, rowsErr := result.RowsAffected(); rowsErr == nil && rows == 0 {
		return errors.New("USDT quote transaction hash is already bound to another proof")
	}
	return nil
}

func (r *Repository) InsertWebhookEvent(ctx context.Context, payload WebhookPayload, raw []byte) (bool, error) {
	payloadHash := sha256Hex(raw)
	result, err := r.db.ExecContext(ctx, `
INSERT INTO usdt_webhook_events (event_id,payload_hash,merchant_order_id,provider_trade_id,payload)
VALUES ($1,$2,$3,$4,$5::jsonb) ON CONFLICT (event_id) DO NOTHING`,
		payload.EventID, payloadHash, payload.OrderID, payload.TradeID, string(raw))
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	if rows > 0 {
		return true, nil
	}
	var existingHash string
	if err := r.db.QueryRowContext(ctx, `SELECT payload_hash FROM usdt_webhook_events WHERE event_id=$1`, payload.EventID).Scan(&existingHash); err != nil {
		return false, err
	}
	if existingHash != payloadHash {
		return false, errors.New("webhook event_id was reused with a different payload")
	}
	return false, nil
}

type webhookEvent struct {
	ID      int64
	Payload WebhookPayload
}

func (r *Repository) ClaimWebhookEvents(ctx context.Context, limit int) ([]webhookEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
WITH due AS (
    SELECT id FROM usdt_webhook_events
    WHERE ((status IN ('pending','failed') AND next_attempt_at <= NOW())
       OR (status='processing' AND updated_at < NOW()-INTERVAL '30 seconds'))
    ORDER BY next_attempt_at FOR UPDATE SKIP LOCKED LIMIT $1
)
UPDATE usdt_webhook_events e SET status='processing', attempts=e.attempts+1, updated_at=NOW()
FROM due WHERE e.id=due.id RETURNING e.id,e.payload`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []webhookEvent
	for rows.Next() {
		var item webhookEvent
		var raw []byte
		if err := rows.Scan(&item.ID, &raw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &item.Payload); err != nil {
			_ = r.CompleteWebhookEvent(ctx, item.ID, err)
			continue
		}
		events = append(events, item)
	}
	return events, rows.Err()
}

func (r *Repository) CompleteWebhookEvent(ctx context.Context, id int64, processErr error) error {
	if processErr == nil {
		_, err := r.db.ExecContext(ctx, `UPDATE usdt_webhook_events SET status='processed',processed_at=NOW(),last_error=NULL,updated_at=NOW() WHERE id=$1`, id)
		return err
	}
	_, err := r.db.ExecContext(ctx, `UPDATE usdt_webhook_events SET status='failed',last_error=$1,next_attempt_at=NOW()+INTERVAL '2 seconds',updated_at=NOW() WHERE id=$2`, processErr.Error(), id)
	return err
}

func (r *Repository) MarkWebhookEventByEventID(ctx context.Context, eventID string, processErr error) error {
	if processErr == nil {
		_, err := r.db.ExecContext(ctx, `UPDATE usdt_webhook_events SET status='processed',processed_at=NOW(),last_error=NULL,updated_at=NOW() WHERE event_id=$1`, eventID)
		return err
	}
	_, err := r.db.ExecContext(ctx, `UPDATE usdt_webhook_events SET status='failed',attempts=attempts+1,last_error=$1,next_attempt_at=NOW()+INTERVAL '2 seconds',updated_at=NOW() WHERE event_id=$2`, processErr.Error(), eventID)
	return err
}

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
