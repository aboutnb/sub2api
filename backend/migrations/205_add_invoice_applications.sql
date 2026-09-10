-- 205_add_invoice_applications.sql
-- Keep local ownership and workflow state for XZNOAuth invoice applications.

CREATE TABLE IF NOT EXISTS invoice_applications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    order_nos JSONB NOT NULL DEFAULT '[]'::jsonb,
    need_pay_tax BOOLEAN NOT NULL DEFAULT FALSE,
    tax_order_nos JSONB NOT NULL DEFAULT '[]'::jsonb,
    validation_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    external_id VARCHAR(128),
    status VARCHAR(32) NOT NULL DEFAULT 'draft',
    title VARCHAR(255),
    recipient_email VARCHAR(255),
    total_amount VARCHAR(32) NOT NULL DEFAULT '',
    currency VARCHAR(8) NOT NULL DEFAULT 'CNY',
    external_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    request_id VARCHAR(128),
    error_code VARCHAR(128),
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoice_applications_user_created
    ON invoice_applications (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_invoice_applications_user_status
    ON invoice_applications (user_id, status);

CREATE UNIQUE INDEX IF NOT EXISTS idx_invoice_applications_external_id
    ON invoice_applications (external_id)
    WHERE external_id IS NOT NULL AND external_id <> '';
