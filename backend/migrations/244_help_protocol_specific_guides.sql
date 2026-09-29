-- Update only untouched seed articles; preserve operator drafts and publication history.
WITH candidates AS (
  SELECT id, CASE
    WHEN slug = 'clients/cherry-studio' THEN replace(replace(replace(replace(replace(replace(content_markdown, '这里使用 OpenAI 兼容接口，不需要登录 OpenAI 官方账号。', '协议随本站平台切换：Claude、Antigravity、DeepSeek、MiniMax 使用 Anthropic；其余平台使用 OpenAI 兼容接口。无需登录上游官方账号。'), '供应商类型选择 OpenAI。', '供应商类型按上方当前渠道生成的 Provider 字段选择，不能固定为 OpenAI。'), 'API Host 使用本站默认端点的 `/v1` 地址，不要填写完整的 `/chat/completions` 路径。', 'Anthropic 的 API Host 使用本站默认端点根地址（Antigravity 带 /antigravity 前缀）；OpenAI 使用 `/v1` 基础地址。不要填写完整消息路径。'), '应只有一层 `/v1`，并以 `/chat/completions` 结束。', 'Anthropic 应以 `/v1/messages` 结束，OpenAI 应以 `/v1/chat/completions` 结束。'), '提供 OpenAI 供应商、自定义 API Host 和模型设置的桌面版本。', '提供 Anthropic / OpenAI 供应商、自定义 API Host 和模型设置的桌面版本。'), '2026-09-22', '2026-09-24')
    WHEN slug = 'clients/read-frog' THEN replace(replace(replace(replace(content_markdown, '添加 OpenAI 兼容服务商。', '按本页 Provider 选择 Anthropic 或 OpenAI 兼容服务商。Claude、Antigravity、DeepSeek、MiniMax 平台采用 Anthropic，其余平台保留 OpenAI 兼容模式。'), '填写下方 Base URL、API Key、模型 ID，启用服务商。', '填写下方 Base URL、API Key、模型 ID，勾选自定义模型名称并启用服务商。Base URL 使用生成的 /v1 基础地址，不是完整消息请求路径。'), '将翻译所用模型切换为新建服务商下的当前模型，运行连接测试。', '先在 API Providers 运行 Test Connection，再检查 Feature Providers 中翻译功能绑定的模型，选择新建服务商下的当前模型。'), '2026-09-22', '2026-09-24')
    WHEN slug = 'clients/kiss-translator' THEN replace(replace(replace(replace(replace(replace(content_markdown, '添加或选择内置 OpenAI 服务，而不是通用自定义 JSON 接口。', '点击 Add，按本页 Provider 选择内置 Claude 或 OpenAI 服务，而不是通用自定义 JSON 接口。Claude、Antigravity、DeepSeek、MiniMax 平台使用 Claude。'), '包含 /v1/chat/completions；', 'Claude 服务包含 /v1/messages，OpenAI 服务包含 /v1/chat/completions；'), '保留内置 OpenAI 请求和响应处理逻辑。', '保留所选内置服务的请求和响应处理逻辑。'), '保存并设为当前翻译服务，选一小段公开文字翻译。', '保存并设为当前翻译服务，刷新测试网页后选一小段公开文字翻译。'), '配置 OpenAI 请求体和响应 Hook', '配置对应协议的请求体和响应 Hook'), '2026-09-22', '2026-09-24')
    WHEN slug = 'clients/immersive-translate' THEN replace(replace(replace(replace(replace(content_markdown, '进入翻译服务的 OpenAI 设置。', '按本页 Provider 进入 Claude 或 OpenAI 翻译服务设置。'), '在 OpenAI 服务中填写本站 API Key。', '在当前服务中填写本站 API Key。Claude、Antigravity、DeepSeek、MiniMax 平台使用 Claude，其余平台使用 OpenAI。'), '填写下方完整 /v1/chat/completions URL。', '填写下方生成的完整 URL：Claude 使用 /v1/messages，OpenAI 使用 /v1/chat/completions。'), '保存并选择 OpenAI 作为当前翻译服务，先测试一段短文本。', '保存后打开浏览器扩展弹窗，选择刚配置的 Claude 或 OpenAI 作为当前服务，先测试一段短文本。'), '2026-09-22', '2026-09-24')
    ELSE content_markdown END AS revised_content
  FROM help_documents
  WHERE slug IN ('clients/cherry-studio', 'clients/read-frog', 'clients/kiss-translator', 'clients/immersive-translate')
    AND version = 1 AND updated_by IS NULL AND status = 'published'
    AND content_markdown = published_snapshot->>'content_markdown'
    AND publication_history = '[]'::jsonb
), revised AS (
  SELECT d.id, c.revised_content,
    CASE WHEN d.category = 'cherry-studio'
      THEN jsonb_set(d.selector_schema, '{protocols}', '["messages","chat-completions"]'::jsonb)
      ELSE d.selector_schema END AS revised_schema,
    CASE WHEN d.category IN ('read-frog','kiss-translator','immersive-translate')
      THEN '按当前渠道选择对应翻译服务、地址格式与模型，分别验证连接和实际翻译。'
      ELSE d.summary END AS revised_summary
  FROM help_documents d JOIN candidates c ON d.id = c.id
  WHERE c.revised_content <> d.content_markdown
)
UPDATE help_documents d
SET content_markdown = r.revised_content,
    summary = r.revised_summary,
    selector_schema = r.revised_schema,
    publication_history = d.publication_history || jsonb_build_array(d.published_snapshot),
    version = d.version + 1,
    published_snapshot = d.published_snapshot || jsonb_build_object(
      'content_markdown', r.revised_content, 'summary', r.revised_summary,
      'selector_schema', r.revised_schema, 'version', d.version + 1, 'published_at', NOW()),
    published_at = NOW(), updated_at = NOW()
FROM revised r WHERE d.id = r.id;
