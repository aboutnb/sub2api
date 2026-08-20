-- Persist the one-time signup benefit decision per risk identity and account.
-- The identity is a server-side HMAC of the trusted client IP + normalized UA;
-- raw request metadata is never stored.
CREATE TABLE IF NOT EXISTS signup_risk_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    fingerprint_hash VARCHAR(64) NOT NULL,
    grant_allowed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT signup_risk_accounts_user_unique UNIQUE (user_id)
);

CREATE INDEX IF NOT EXISTS signup_risk_accounts_fingerprint_idx
    ON signup_risk_accounts (fingerprint_hash);

CREATE UNIQUE INDEX IF NOT EXISTS signup_risk_accounts_one_grant_per_fingerprint
    ON signup_risk_accounts (fingerprint_hash)
    WHERE grant_allowed = TRUE;
