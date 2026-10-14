-- In-site support tickets. Idempotent and non-cascading.

CREATE TABLE IF NOT EXISTS support_tickets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    subject VARCHAR(80) NOT NULL,
    category VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    preview TEXT NOT NULL DEFAULT '',
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS support_ticket_messages (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL,
    author_id BIGINT NOT NULL,
    author_role VARCHAR(16) NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_support_tickets_user_status
    ON support_tickets (user_id, status);

CREATE INDEX IF NOT EXISTS idx_support_tickets_recent
    ON support_tickets (last_message_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_support_ticket_messages_ticket
    ON support_ticket_messages (ticket_id, id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'support_tickets_user_id_fkey'
    ) THEN
        ALTER TABLE support_tickets
            ADD CONSTRAINT support_tickets_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES users (id);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'support_ticket_messages_ticket_id_fkey'
    ) THEN
        ALTER TABLE support_ticket_messages
            ADD CONSTRAINT support_ticket_messages_ticket_id_fkey
            FOREIGN KEY (ticket_id) REFERENCES support_tickets (id);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'support_ticket_messages_author_id_fkey'
    ) THEN
        ALTER TABLE support_ticket_messages
            ADD CONSTRAINT support_ticket_messages_author_id_fkey
            FOREIGN KEY (author_id) REFERENCES users (id);
    END IF;
END $$;
