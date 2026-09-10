-- Expose check-in balance changes through the existing user/admin balance
-- history without mixing system entries into redeemable code inventory.
INSERT INTO redeem_codes (code, type, value, status, used_by, used_at, created_at)
SELECT
    'SYS-CHECKIN-' || id,
    'checkin',
    reward_amount,
    'used',
    user_id,
    checked_in_at,
    created_at
FROM checkin_records
ON CONFLICT (code) DO NOTHING;
