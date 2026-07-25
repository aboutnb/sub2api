-- Defense in depth for check-in settlement data. Application validation and
-- the user-row transaction remain primary; these constraints reject future
-- regressions or unexpected writes that would violate the settlement model.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'checkin_records_balance_integrity'
    ) THEN
        ALTER TABLE checkin_records
            ADD CONSTRAINT checkin_records_balance_integrity CHECK (
                balance_before >= 0
                AND balance_after >= 0
                AND balance_after = GREATEST(balance_before + reward_amount, 0)
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'checkin_records_reward_integrity'
    ) THEN
        ALTER TABLE checkin_records
            ADD CONSTRAINT checkin_records_reward_integrity CHECK (
                (
                    mode = 'normal'
                    AND random_value BETWEEN 0 AND 100
                    AND reward_amount = random_value
                )
                OR
                (
                    mode = 'lucky'
                    AND random_value BETWEEN -1 AND 1
                    AND reward_amount BETWEEN -balance_before AND balance_before
                )
            );
    END IF;
END $$;
