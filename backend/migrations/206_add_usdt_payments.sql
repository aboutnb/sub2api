-- 206_add_usdt_payments.sql
-- Isolated durable state for BEpusdt USDT quotes, callbacks, and reconciliation.

CREATE TABLE IF NOT EXISTS usdt_payment_quotes (
    id BIGSERIAL PRIMARY KEY,
    payment_order_id BIGINT NOT NULL REFERENCES payment_orders(id) ON DELETE CASCADE,
    merchant_order_id VARCHAR(64) NOT NULL,
    provider_trade_id VARCHAR(128) NOT NULL,
    fiat_currency VARCHAR(8) NOT NULL,
    fiat_amount VARCHAR(32) NOT NULL,
    crypto_currency VARCHAR(16) NOT NULL DEFAULT 'USDT',
    network VARCHAR(32) NOT NULL,
    trade_type VARCHAR(32) NOT NULL,
    crypto_amount VARCHAR(64) NOT NULL,
    exchange_rate VARCHAR(64) NOT NULL,
    receiving_address VARCHAR(256) NOT NULL,
    payment_url TEXT NOT NULL,
    upstream_created_at TIMESTAMPTZ NOT NULL,
    upstream_expires_at TIMESTAMPTZ NOT NULL,
    provider_status VARCHAR(32) NOT NULL DEFAULT 'waiting',
    transaction_hash VARCHAR(256),
    chain_transfer_at TIMESTAMPTZ,
    block_number BIGINT,
    last_reconciled_at TIMESTAMPTZ,
    next_reconcile_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reconcile_attempts INTEGER NOT NULL DEFAULT 0,
    reconcile_lease_until TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_usdt_quotes_payment_order
    ON usdt_payment_quotes (payment_order_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_usdt_quotes_merchant_order
    ON usdt_payment_quotes (merchant_order_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_usdt_quotes_provider_trade
    ON usdt_payment_quotes (provider_trade_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_usdt_quotes_transaction_hash
    ON usdt_payment_quotes (transaction_hash)
    WHERE transaction_hash IS NOT NULL AND transaction_hash <> '';
CREATE INDEX IF NOT EXISTS idx_usdt_quotes_reconcile_due
    ON usdt_payment_quotes (next_reconcile_at)
    WHERE provider_status IN ('waiting', 'confirming');

CREATE TABLE IF NOT EXISTS usdt_webhook_events (
    id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(128) NOT NULL,
    payload_hash VARCHAR(64) NOT NULL,
    merchant_order_id VARCHAR(64) NOT NULL,
    provider_trade_id VARCHAR(128) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_usdt_webhook_event_id
    ON usdt_webhook_events (event_id);
CREATE INDEX IF NOT EXISTS idx_usdt_webhook_events_pending
    ON usdt_webhook_events (next_attempt_at)
    WHERE status IN ('pending', 'failed');
