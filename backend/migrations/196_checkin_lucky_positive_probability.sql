-- Separate the chance of a positive lucky result from its multiplier or amount
-- range so changing reward magnitude does not silently change the odds.
INSERT INTO settings (key, value, updated_at)
VALUES ('checkin_lucky_positive_probability', '70', NOW())
ON CONFLICT (key) DO NOTHING;
