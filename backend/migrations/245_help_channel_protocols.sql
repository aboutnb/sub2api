-- Align untouched OpenClaw, Hermes, and Pi articles with channel-specific protocols.
WITH candidates AS (
  SELECT id, CASE
    WHEN slug = 'clients/openclaw' THEN replace(replace(replace(replace(content_markdown,
      '通过自定义 Provider 接入本站 Chat Completions 模型。',
      '按当前渠道选择 OpenClaw API：OpenAI 用 openai-responses，Claude 类平台用 anthropic-messages，其余用 openai-completions。'),
      '2. models.providers.site.baseUrl 使用本站默认端点的 /v1 地址，api 为 openai-completions。',
      '2. models.providers.site.api 随渠道变化：OpenAI 为 openai-responses，baseUrl 用 /v1；Claude、Antigravity、DeepSeek、MiniMax 为 anthropic-messages，baseUrl 用根地址（Antigravity 加 /antigravity）；其余平台为 openai-completions，baseUrl 用 /v1。'),
      '不要把 openai-completions 地址配置成 openai-responses。',
      '不要混用 openai-responses、anthropic-messages 和 openai-completions；协议必须和当前渠道一致。'),
      '2026-09-22', '2026-09-25')
    WHEN slug = 'clients/hermes' THEN replace(replace(replace(replace(content_markdown,
      '使用 custom Provider、配置文件和独立环境文件连接本站。',
      '使用 custom Provider，并按渠道设置 api_mode。'),
      '从 [Hermes 官方文档](https://hermes-agent.nousresearch.com/docs/) 安装 Hermes Agent，并完成初始化。Windows 的支持环境以该版本官方说明为准。配置前备份 ~/.hermes。',
      'macOS / Linux 可参考 `curl -fsSL https://raw.githubusercontent.com/NousResearch/hermes-agent/main/scripts/install.sh | bash`，再执行 `source ~/.zshrc` 和 `hermes --version`。Windows 以 [Hermes 官方文档](https://hermes-agent.nousresearch.com/docs/) 为准。配置前备份 ~/.hermes。'),
      '4. model.default 是当前模型 ID，model.base_url 是本站 /v1 地址，model.key_env 引用 SITE_API_KEY。',
      '4. model.default 是当前模型 ID，model.api_key 填写 YOUR_API_KEY。OpenAI 渠道的 api_mode 为 codex_responses，base_url 用 /v1；Claude、Antigravity、DeepSeek、MiniMax 的 api_mode 为 anthropic_messages，base_url 用根地址；Grok 不设置 api_mode。'),
      '2026-09-22', '2026-09-25')
    WHEN slug = 'clients/pi' THEN replace(replace(content_markdown,
      '从 [Pi 官网](https://pi.dev/) 查看当前安装方式和支持的系统，完成安装后运行 `pi --version`。此教程仅在本站 OpenAI 渠道展示。',
      '参考命令 `npm install -g --ignore-scripts @earendil-works/pi-coding-agent`。先按 [Pi 官网](https://pi.dev/) 核对 Node.js 要求，再运行 `pi --version`。此教程仅在本站 OpenAI 渠道展示。'),
      '2026-09-22', '2026-09-25')
    ELSE content_markdown END AS revised_content
  FROM help_documents
  WHERE slug IN ('clients/openclaw', 'clients/hermes', 'clients/pi')
    AND version = 1 AND updated_by IS NULL AND status = 'published'
    AND content_markdown = published_snapshot->>'content_markdown'
    AND publication_history = '[]'::jsonb
), revised AS (
  SELECT d.id, c.revised_content,
    CASE
      WHEN d.slug = 'clients/openclaw' THEN '按当前渠道写入 openai-responses、anthropic-messages 或 openai-completions。'
      WHEN d.slug = 'clients/hermes' THEN '按当前渠道设置 custom provider 和对应 api_mode。'
      ELSE d.summary
    END AS revised_summary
  FROM help_documents d JOIN candidates c ON d.id = c.id
  WHERE c.revised_content <> d.content_markdown
)
UPDATE help_documents d
SET content_markdown = r.revised_content,
    summary = r.revised_summary,
    publication_history = d.publication_history || jsonb_build_array(d.published_snapshot),
    version = d.version + 1,
    published_snapshot = d.published_snapshot || jsonb_build_object(
      'content_markdown', r.revised_content, 'summary', r.revised_summary,
      'selector_schema', d.selector_schema, 'version', d.version + 1, 'published_at', NOW()),
    published_at = NOW(), updated_at = NOW()
FROM revised r WHERE d.id = r.id;
