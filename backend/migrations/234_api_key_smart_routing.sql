CREATE TABLE IF NOT EXISTS api_key_smart_routes (
    api_key_id BIGINT PRIMARY KEY REFERENCES api_keys(id) ON DELETE CASCADE,
    platform VARCHAR(32) NOT NULL,
    subscription_type VARCHAR(32) NOT NULL,
    strategy VARCHAR(16) NOT NULL CHECK (strategy IN ('auto', 'price', 'speed', 'success', 'custom')),
    price_weight SMALLINT NOT NULL CHECK (price_weight BETWEEN 0 AND 100),
    speed_weight SMALLINT NOT NULL CHECK (speed_weight BETWEEN 0 AND 100),
    success_weight SMALLINT NOT NULL CHECK (success_weight BETWEEN 0 AND 100),
    rate_guard_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    max_rate_multiplier DOUBLE PRECISION,
    max_image_rate_multiplier DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT api_key_smart_routes_weights_sum CHECK (
        price_weight + speed_weight + success_weight = 100
    ),
    CONSTRAINT api_key_smart_routes_max_rate_non_negative CHECK (
        max_rate_multiplier IS NULL OR (
            max_rate_multiplier >= 0 AND max_rate_multiplier < 'Infinity'::DOUBLE PRECISION
        )
    ),
    CONSTRAINT api_key_smart_routes_max_image_rate_non_negative CHECK (
        max_image_rate_multiplier IS NULL OR (
            max_image_rate_multiplier >= 0 AND max_image_rate_multiplier < 'Infinity'::DOUBLE PRECISION
        )
    )
);

CREATE TABLE IF NOT EXISTS api_key_smart_route_groups (
    api_key_id BIGINT NOT NULL REFERENCES api_key_smart_routes(api_key_id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    position SMALLINT NOT NULL CHECK (position BETWEEN 0 AND 19),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (api_key_id, group_id),
    UNIQUE (api_key_id, position)
);

CREATE INDEX IF NOT EXISTS idx_api_key_smart_route_groups_group
    ON api_key_smart_route_groups(group_id, api_key_id);

INSERT INTO settings (key, value, updated_at)
VALUES ('smart_routing_enabled', 'false', NOW())
ON CONFLICT (key) DO NOTHING;

COMMENT ON TABLE api_key_smart_routes IS
    'Optional smart routing configuration. Row absence keeps the API key in legacy single-group mode.';
COMMENT ON TABLE api_key_smart_route_groups IS
    'Ordered candidate groups for an API key smart routing configuration.';
