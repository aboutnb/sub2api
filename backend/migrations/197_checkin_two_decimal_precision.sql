-- New check-ins settle rewards and random values to two decimal places. Keep
-- the original scale on historical rows so their immutable audit values remain
-- valid, while the application explicitly marks all new rows with scale 2.
UPDATE settings
SET value = CASE
        WHEN key IN ('checkin_lucky_min_multiplier', 'checkin_lucky_amount_min')
            THEN LEAST(ROUND(value::numeric, 2), -0.01)::text
        WHEN key IN ('checkin_lucky_max_multiplier', 'checkin_lucky_amount_max')
            THEN GREATEST(ROUND(value::numeric, 2), 0.01)::text
        ELSE ROUND(value::numeric, 2)::text
    END,
    updated_at = NOW()
WHERE key IN (
        'checkin_normal_min',
        'checkin_normal_max',
        'checkin_lucky_positive_probability',
        'checkin_lucky_min_multiplier',
        'checkin_lucky_max_multiplier',
        'checkin_lucky_amount_min',
        'checkin_lucky_amount_max'
    )
  AND value ~ '^[+-]?([0-9]+(\.[0-9]*)?|\.[0-9]+)$';

ALTER TABLE checkin_records
    ADD COLUMN IF NOT EXISTS calculation_scale SMALLINT NOT NULL DEFAULT 8;

ALTER TABLE checkin_records
    DROP CONSTRAINT IF EXISTS checkin_records_calculation_scale_check,
    DROP CONSTRAINT IF EXISTS checkin_records_reward_integrity;

ALTER TABLE checkin_records
    ADD CONSTRAINT checkin_records_calculation_scale_check
        CHECK (calculation_scale IN (2, 8)),
    ADD CONSTRAINT checkin_records_reward_integrity CHECK (
        (
            mode = 'normal'
            AND reward_type = 'amount'
            AND random_value BETWEEN 0 AND 100
            AND random_value = ROUND(random_value, calculation_scale)
            AND reward_amount = random_value
        )
        OR
        (
            mode = 'lucky'
            AND reward_type = 'amount'
            AND random_value BETWEEN -100 AND 100
            AND random_value = ROUND(random_value, calculation_scale)
            AND reward_amount = CASE
                WHEN calculation_scale = 2
                    THEN GREATEST(random_value, -TRUNC(balance_before, 2))
                ELSE GREATEST(random_value, -balance_before)
            END
        )
        OR
        (
            mode = 'lucky'
            AND reward_type = 'multiplier'
            AND random_value BETWEEN -1 AND 10
            AND random_value = ROUND(random_value, calculation_scale)
            AND reward_amount = CASE
                WHEN calculation_scale = 2
                    THEN GREATEST(ROUND(balance_before * random_value, 2), -TRUNC(balance_before, 2))
                ELSE GREATEST(ROUND(balance_before * random_value, 8), -balance_before)
            END
        )
    );
