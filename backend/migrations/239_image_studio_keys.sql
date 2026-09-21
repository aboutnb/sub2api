ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS purpose VARCHAR(32) NOT NULL DEFAULT 'standard';
CREATE UNIQUE INDEX IF NOT EXISTS api_keys_image_studio_owner_group
    ON api_keys (user_id, group_id)
    WHERE purpose = 'image_studio' AND deleted_at IS NULL;
