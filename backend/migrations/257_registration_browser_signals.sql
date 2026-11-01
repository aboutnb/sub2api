ALTER TABLE registration_source_accounts
    ADD COLUMN IF NOT EXISTS browser_hash VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS feature_hash VARCHAR(64) NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS registration_source_accounts_browser_time
    ON registration_source_accounts(browser_hash, created_at) WHERE browser_hash <> '';
CREATE INDEX IF NOT EXISTS registration_source_accounts_feature_time
    ON registration_source_accounts(feature_hash, created_at) WHERE feature_hash <> '';
