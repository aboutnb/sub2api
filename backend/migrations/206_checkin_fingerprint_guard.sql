-- Add an independent rolling-window limit for the HMAC(IP + User-Agent)
-- fingerprint. The existing IP guard remains separate to cover multiple
-- devices behind one address.
INSERT INTO settings (key, value, updated_at)
VALUES
    ('checkin_fingerprint_window_minutes', '1440', NOW()),
    ('checkin_fingerprint_max_users', '1', NOW())
ON CONFLICT (key) DO NOTHING;
