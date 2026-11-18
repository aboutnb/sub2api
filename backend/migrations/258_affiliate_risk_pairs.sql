-- Rebate-only restrictions; neither registration nor paid balance is blocked.
CREATE TABLE IF NOT EXISTS affiliate_risk_pairs (
    inviter_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (inviter_id, invitee_id)
);
