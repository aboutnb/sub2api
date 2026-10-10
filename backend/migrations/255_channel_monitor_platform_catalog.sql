-- Extend passive monitoring to platforms added after the original six.
-- Preserve existing entries, including disabled platforms and model lists.
WITH missing AS (
    SELECT c.id,
           jsonb_agg(jsonb_build_object('platform', p.platform, 'enabled', true, 'models', '[]'::jsonb)
                     ORDER BY p.platform) AS platforms
    FROM channel_monitor_v2_config c
    CROSS JOIN (VALUES ('kimi'), ('zhipu'), ('deepseek'), ('minimax'), ('opencode_go'), ('typesafe')) AS p(platform)
    WHERE jsonb_array_length(c.platforms) > 0
      AND NOT EXISTS (
          SELECT 1 FROM jsonb_array_elements(c.platforms) AS existing
          WHERE lower(trim(existing->>'platform')) = p.platform
      )
    GROUP BY c.id
)
UPDATE channel_monitor_v2_config c
SET platforms = c.platforms || missing.platforms,
    version = c.version + 1,
    updated_at = NOW()
FROM missing
WHERE c.id = missing.id;

-- Keep fresh installations aligned; legacy Kiro is still a valid dimension.
ALTER TABLE channel_monitor_v2_config ALTER COLUMN platforms SET DEFAULT
'[{"platform":"anthropic","enabled":true,"models":[]},{"platform":"openai","enabled":true,"models":[]},{"platform":"gemini","enabled":true,"models":[]},{"platform":"antigravity","enabled":true,"models":[]},{"platform":"grok","enabled":true,"models":[]},{"platform":"kiro","enabled":true,"models":[]},{"platform":"kimi","enabled":true,"models":[]},{"platform":"zhipu","enabled":true,"models":[]},{"platform":"deepseek","enabled":true,"models":[]},{"platform":"minimax","enabled":true,"models":[]},{"platform":"opencode_go","enabled":true,"models":[]},{"platform":"typesafe","enabled":true,"models":[]}]'::jsonb;
