ALTER TABLE api_key_smart_routes
    ADD COLUMN IF NOT EXISTS cross_platform BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS prefer_platform VARCHAR(32) NOT NULL DEFAULT '';

COMMENT ON COLUMN api_key_smart_routes.cross_platform IS
    'When true, candidate groups may use different platforms. Manual routes default to false.';
COMMENT ON COLUMN api_key_smart_routes.prefer_platform IS
    'Optional platform ranked ahead of the score. Empty keeps score-only ordering.';

CREATE UNIQUE INDEX IF NOT EXISTS api_keys_image_studio_prompt_route
    ON api_keys (user_id)
    WHERE purpose = 'image_studio' AND group_id IS NULL AND deleted_at IS NULL;
