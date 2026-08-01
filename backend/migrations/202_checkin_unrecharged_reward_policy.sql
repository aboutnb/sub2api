-- Reduce normal check-in rewards for users who have accumulated free check-ins
-- without any positive balance recharge. The policy is editable by admins.
INSERT INTO settings (key, value, updated_at)
VALUES
    ('checkin_unrecharged_reduction_enabled', 'true', NOW()),
    ('checkin_unrecharged_checkin_threshold', '3', NOW()),
    ('checkin_unrecharged_normal_reward_percent', '50', NOW())
ON CONFLICT (key) DO NOTHING;
