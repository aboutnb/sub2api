-- 203_add_email_broadcast_tasks.sql
-- Persistent email broadcast tasks and per-recipient delivery state.

CREATE TABLE IF NOT EXISTS email_broadcast_tasks (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    event VARCHAR(100) NOT NULL,
    status VARCHAR(24) NOT NULL,
    variables JSONB NOT NULL DEFAULT '{}'::jsonb,
    template_snapshots JSONB NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    total_recipients BIGINT NOT NULL DEFAULT 0,
    sent_count BIGINT NOT NULL DEFAULT 0,
    failed_count BIGINT NOT NULL DEFAULT 0,
    canceled_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    canceled_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_broadcast_tasks_status_schedule
    ON email_broadcast_tasks(status, scheduled_at, id);

CREATE INDEX IF NOT EXISTS idx_email_broadcast_tasks_created_at
    ON email_broadcast_tasks(created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS email_broadcast_recipients (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES email_broadcast_tasks(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    email VARCHAR(255) NOT NULL,
    recipient_name VARCHAR(100) NOT NULL DEFAULT '',
    locale VARCHAR(8) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error TEXT,
    claimed_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_email_broadcast_recipient UNIQUE (task_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_email_broadcast_recipients_claim
    ON email_broadcast_recipients(status, next_attempt_at, id);

CREATE INDEX IF NOT EXISTS idx_email_broadcast_recipients_task_status
    ON email_broadcast_recipients(task_id, status, id);
