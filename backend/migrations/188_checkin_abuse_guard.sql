-- Check-in abuse controls are enabled by default and apply only to a user's
-- first settlement for the current business date.
INSERT INTO settings (key, value, updated_at)
VALUES
    ('checkin_risk_control_enabled', 'true', NOW()),
    ('checkin_min_account_age_hours', '24', NOW()),
    ('checkin_ip_window_minutes', '10', NOW()),
    ('checkin_ip_max_users', '20', NOW())
ON CONFLICT (key) DO NOTHING;
