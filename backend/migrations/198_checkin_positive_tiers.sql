-- Positive lucky rewards use explicit weighted tiers. Existing configurations
-- keep a single uniform tier, except the agreed 0.20 multiplier range which
-- starts with the 70/20/10 distribution.
DO $$
DECLARE
    multiplier_max NUMERIC := 0.10;
    amount_max NUMERIC := 0.10;
    multiplier_tiers TEXT;
    amount_tiers TEXT;
BEGIN
    SELECT CASE WHEN value ~ '^[+]?[0-9]+(\.[0-9]+)?$' THEN value::NUMERIC ELSE 0.10 END
    INTO multiplier_max
    FROM settings
    WHERE key = 'checkin_lucky_max_multiplier';

    SELECT CASE WHEN value ~ '^[+]?[0-9]+(\.[0-9]+)?$' THEN value::NUMERIC ELSE 0.10 END
    INTO amount_max
    FROM settings
    WHERE key = 'checkin_lucky_amount_max';

    multiplier_max := COALESCE(multiplier_max, 0.10);
    amount_max := COALESCE(amount_max, 0.10);

    IF multiplier_max = 0.20 THEN
        multiplier_tiers := '[{"min":"0.01","max":"0.10","weight":"70"},{"min":"0.11","max":"0.15","weight":"20"},{"min":"0.16","max":"0.20","weight":"10"}]';
    ELSE
        multiplier_tiers := jsonb_build_array(jsonb_build_object(
            'min', '0.01',
            'max', to_char(multiplier_max, 'FM999990.00'),
            'weight', '100'
        ))::TEXT;
    END IF;

    amount_tiers := jsonb_build_array(jsonb_build_object(
        'min', '0.01',
        'max', to_char(amount_max, 'FM999990.00'),
        'weight', '100'
    ))::TEXT;

    INSERT INTO settings (key, value, updated_at)
    VALUES
        ('checkin_lucky_multiplier_positive_tiers', multiplier_tiers, NOW()),
        ('checkin_lucky_amount_positive_tiers', amount_tiers, NOW())
    ON CONFLICT (key) DO NOTHING;
END $$;
