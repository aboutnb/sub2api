SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

-- Authentication abuse bans. One row represents the latest lifecycle for an
-- IP-wide or IP+UA scoped ban and remains available for administrator review.
CREATE TABLE IF NOT EXISTS auth_ip_bans (
    id BIGSERIAL PRIMARY KEY,
    ip_address INET NOT NULL,
    ban_scope VARCHAR(16) NOT NULL DEFAULT 'ip',
    ua_hash VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    ua_category VARCHAR(32) NOT NULL DEFAULT 'unknown',
    source VARCHAR(64) NOT NULL DEFAULT 'auth_auto_ban',
    reason VARCHAR(128) NOT NULL DEFAULT '',
    trigger_path VARCHAR(512) NOT NULL DEFAULT '',
    target_identifier VARCHAR(255) NOT NULL DEFAULT '',
    failure_count INT NOT NULL DEFAULT 0,
    first_seen_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL,
    banned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    released_at TIMESTAMPTZ,
    released_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    release_note VARCHAR(255) NOT NULL DEFAULT '',
    ban_count INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT auth_ip_bans_scope_check CHECK (ban_scope IN ('ip', 'ip_ua')),
    CONSTRAINT auth_ip_bans_failure_count_check CHECK (failure_count >= 0),
    CONSTRAINT auth_ip_bans_ban_count_check CHECK (ban_count > 0),
    CONSTRAINT auth_ip_bans_expiry_check CHECK (expires_at > banned_at),
    CONSTRAINT auth_ip_bans_identity_unique UNIQUE (ip_address, ban_scope, ua_hash)
);

CREATE INDEX IF NOT EXISTS idx_auth_ip_bans_status_expiry
    ON auth_ip_bans (released_at, expires_at DESC);
CREATE INDEX IF NOT EXISTS idx_auth_ip_bans_updated
    ON auth_ip_bans (updated_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_auth_ip_bans_ip
    ON auth_ip_bans (ip_address);
