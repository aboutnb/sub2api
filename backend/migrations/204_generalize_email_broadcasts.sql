-- 204_generalize_email_broadcasts.sql
-- Persist the audience definition used to build each immutable recipient snapshot.

ALTER TABLE email_broadcast_tasks
    ADD COLUMN IF NOT EXISTS audience JSONB NOT NULL DEFAULT '{"mode":"all"}'::jsonb;
