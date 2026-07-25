-- Daily check-in records. The business key makes retries and concurrent tabs idempotent.
CREATE TABLE IF NOT EXISTS checkin_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    checkin_date DATE NOT NULL,
    mode VARCHAR(16) NOT NULL CHECK (mode IN ('normal', 'lucky')),
    random_value DECIMAL(20,8) NOT NULL,
    reward_amount DECIMAL(20,8) NOT NULL,
    balance_before DECIMAL(20,8) NOT NULL,
    balance_after DECIMAL(20,8) NOT NULL,
    checked_in_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT checkin_records_user_date_unique UNIQUE (user_id, checkin_date)
);

CREATE INDEX IF NOT EXISTS idx_checkin_records_user_date
    ON checkin_records (user_id, checkin_date DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_checkin_records_date
    ON checkin_records (checkin_date DESC, id DESC);

-- Keep the feature disabled by default while providing valid safe parameters for
-- an administrator who enables it later.
INSERT INTO settings (key, value, updated_at)
VALUES
    ('checkin_enabled', 'false', NOW()),
    ('checkin_normal_min', '0.01', NOW()),
    ('checkin_normal_max', '0.05', NOW()),
    ('checkin_lucky_min_multiplier', '-0.05', NOW()),
    ('checkin_lucky_max_multiplier', '0.10', NOW()),
    ('checkin_config_version', '1', NOW())
ON CONFLICT (key) DO NOTHING;
