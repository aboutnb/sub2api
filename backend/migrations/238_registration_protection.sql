-- Successful registrations are committed in the same transaction as users.
CREATE TABLE IF NOT EXISTS registration_source_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE REFERENCES users(id) ON DELETE SET NULL,
    ip_hash VARCHAR(64) NOT NULL,
    identity_hash VARCHAR(64) NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    trigger_path VARCHAR(512) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS registration_source_accounts_ip_time
    ON registration_source_accounts(ip_hash, created_at);
CREATE INDEX IF NOT EXISTS registration_source_accounts_identity_time
    ON registration_source_accounts(identity_hash, created_at);

CREATE TABLE IF NOT EXISTS registration_risk_events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    ip_hash VARCHAR(64) NOT NULL,
    identity_hash VARCHAR(64) NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    trigger_path VARCHAR(512) NOT NULL DEFAULT '',
    reason VARCHAR(128) NOT NULL,
    action VARCHAR(32) NOT NULL,
    observed_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS registration_risk_events_time ON registration_risk_events(created_at DESC);
CREATE INDEX IF NOT EXISTS registration_risk_events_user ON registration_risk_events(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS registration_source_blocks (
    id BIGSERIAL PRIMARY KEY,
    scope VARCHAR(16) NOT NULL CHECK (scope IN ('ip', 'ip_ua')),
    source_hash VARCHAR(64) NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    reason VARCHAR(128) NOT NULL,
    failure_count INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    released_at TIMESTAMPTZ,
    released_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    release_note VARCHAR(512) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS registration_source_blocks_lookup
    ON registration_source_blocks(scope, source_hash, expires_at) WHERE released_at IS NULL;

-- A negative concurrency alone never establishes a risk classification.
CREATE TABLE IF NOT EXISTS registration_risk_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    reason VARCHAR(128) NOT NULL,
    status VARCHAR(16) NOT NULL CHECK (status IN ('observed', 'restricted', 'released')),
    previous_concurrency INTEGER NOT NULL CHECK (previous_concurrency >= -1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    reviewed_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    review_note VARCHAR(512) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS registration_risk_accounts_status ON registration_risk_accounts(status, created_at DESC);
