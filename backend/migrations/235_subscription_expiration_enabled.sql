-- 235: add the unified subscription expiration policy switch.
-- Existing installations default to the historical behavior (expiry enforced).
INSERT INTO settings (key, value, updated_at)
VALUES ('subscription_expiration_enabled', 'true', NOW())
ON CONFLICT (key) DO NOTHING;
