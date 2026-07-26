-- Add an explicit lucky reward strategy while preserving all existing records
-- and multiplier settings. Normal rewards are always fixed amounts.
INSERT INTO settings (key, value, updated_at)
VALUES
    ('checkin_lucky_reward_type', 'multiplier', NOW()),
    ('checkin_lucky_amount_min', '-0.05', NOW()),
    ('checkin_lucky_amount_max', '0.10', NOW())
ON CONFLICT (key) DO NOTHING;

ALTER TABLE checkin_records
    ADD COLUMN IF NOT EXISTS reward_type VARCHAR(16) NOT NULL DEFAULT 'multiplier';

UPDATE checkin_records
SET reward_type = 'amount'
WHERE mode = 'normal' AND reward_type <> 'amount';

ALTER TABLE checkin_records
    DROP CONSTRAINT IF EXISTS checkin_records_reward_type_check,
    DROP CONSTRAINT IF EXISTS checkin_records_reward_integrity;

ALTER TABLE checkin_records
    ADD CONSTRAINT checkin_records_reward_type_check
        CHECK (reward_type IN ('amount', 'multiplier')),
    ADD CONSTRAINT checkin_records_reward_integrity CHECK (
        (
            mode = 'normal'
            AND reward_type = 'amount'
            AND random_value BETWEEN 0 AND 100
            AND reward_amount = random_value
        )
        OR
        (
            mode = 'lucky'
            AND reward_type = 'amount'
            AND random_value BETWEEN -100 AND 100
            AND reward_amount = GREATEST(random_value, -balance_before)
        )
        OR
        (
            mode = 'lucky'
            AND reward_type = 'multiplier'
            AND random_value BETWEEN -1 AND 10
            AND reward_amount = GREATEST(ROUND(balance_before * random_value, 8), -balance_before)
        )
    );
