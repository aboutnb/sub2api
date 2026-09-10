-- Promote the untouched legacy lucky-check-in defaults to the agreed
-- 65% positive distribution and -0.08..0.20 multiplier range. Treat the
-- four related values as one fingerprint so customized sites are preserved.
DO $$
DECLARE
    positive_probability TEXT;
    multiplier_min TEXT;
    multiplier_max TEXT;
    multiplier_tiers TEXT;
    legacy_tiers_match BOOLEAN := FALSE;
BEGIN
    SELECT value INTO positive_probability
    FROM settings
    WHERE key = 'checkin_lucky_positive_probability';

    SELECT value INTO multiplier_min
    FROM settings
    WHERE key = 'checkin_lucky_min_multiplier';

    SELECT value INTO multiplier_max
    FROM settings
    WHERE key = 'checkin_lucky_max_multiplier';

    SELECT value INTO multiplier_tiers
    FROM settings
    WHERE key = 'checkin_lucky_multiplier_positive_tiers';

    BEGIN
        legacy_tiers_match := multiplier_tiers::JSONB =
            '[{"min":"0.01","max":"0.10","weight":"100"}]'::JSONB;
    EXCEPTION WHEN OTHERS THEN
        legacy_tiers_match := FALSE;
    END;

    IF positive_probability ~ '^[+]?[0-9]+(\.[0-9]+)?$'
        AND positive_probability::NUMERIC = 70
        AND multiplier_min ~ '^-[0-9]+(\.[0-9]+)?$'
        AND multiplier_min::NUMERIC = -0.05
        AND multiplier_max ~ '^[+]?[0-9]+(\.[0-9]+)?$'
        AND multiplier_max::NUMERIC = 0.10
        AND legacy_tiers_match THEN
        UPDATE settings
        SET value = CASE key
                WHEN 'checkin_lucky_positive_probability' THEN '65'
                WHEN 'checkin_lucky_min_multiplier' THEN '-0.08'
                WHEN 'checkin_lucky_max_multiplier' THEN '0.20'
                WHEN 'checkin_lucky_multiplier_positive_tiers' THEN
                    '[{"min":"0.01","max":"0.10","weight":"70"},{"min":"0.11","max":"0.15","weight":"20"},{"min":"0.16","max":"0.20","weight":"10"}]'
            END,
            updated_at = NOW()
        WHERE key IN (
            'checkin_lucky_positive_probability',
            'checkin_lucky_min_multiplier',
            'checkin_lucky_max_multiplier',
            'checkin_lucky_multiplier_positive_tiers'
        );

        UPDATE settings
        SET value = CASE
                WHEN value ~ '^[0-9]+$' THEN (value::BIGINT + 1)::TEXT
                ELSE '1'
            END,
            updated_at = NOW()
        WHERE key = 'checkin_config_version';
    END IF;
END $$;
