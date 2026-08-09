-- Tighten the original check-in IP guard default. Deployments still using the
-- former default are updated; administrators can explicitly choose another
-- value from the check-in center afterwards.
UPDATE settings
SET value = '5', updated_at = NOW()
WHERE key = 'checkin_ip_max_users'
  AND value = '20';

INSERT INTO settings (key, value, updated_at)
VALUES ('checkin_ip_max_users', '5', NOW())
ON CONFLICT (key) DO NOTHING;
