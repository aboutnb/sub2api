-- Keep the check-in challenge independent from the global authentication challenge.
-- Existing installations remain unchanged until an administrator explicitly enables it.
INSERT INTO settings (key, value, updated_at)
VALUES ('checkin_turnstile_enabled', 'false', NOW())
ON CONFLICT (key) DO NOTHING;
