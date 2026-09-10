-- Keep both existing check-in modes available after upgrading while allowing
-- administrators to control their user-facing availability independently.
INSERT INTO settings (key, value, updated_at)
VALUES
    ('checkin_normal_enabled', 'true', NOW()),
    ('checkin_lucky_enabled', 'true', NOW())
ON CONFLICT (key) DO NOTHING;
