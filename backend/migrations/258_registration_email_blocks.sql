-- Persistent admission decisions survive source rotation and risk-log cleanup.
CREATE TABLE registration_email_blocks (
 email_identity TEXT PRIMARY KEY,
 reason VARCHAR(128) NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 released_at TIMESTAMPTZ,
 release_note TEXT NOT NULL DEFAULT ''
);

-- Preserve the reviewed disabled accounts when this guard is first deployed.
WITH reviewed AS (
 SELECT DISTINCT LOWER(TRIM(u.email)) AS email
 FROM users u JOIN registration_risk_events e ON e.user_id=u.id
 WHERE u.status='disabled' AND e.reason='manual_batch_registration_abuse' AND e.action='disabled'
), parts AS (
 SELECT SPLIT_PART(SPLIT_PART(email,'@',1),'+',1) AS local_part,
 RTRIM(SPLIT_PART(email,'@',2),'.') AS domain FROM reviewed
)
INSERT INTO registration_email_blocks(email_identity,reason)
SELECT CASE WHEN domain IN ('gmail.com','googlemail.com')
 THEN REPLACE(local_part,'.','')||'@gmail.com' ELSE local_part||'@'||domain END,
 'manual_batch_registration_abuse' FROM parts
ON CONFLICT(email_identity) DO NOTHING;
