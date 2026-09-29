-- Correct untouched GUI guides that contradicted the live client tutorials.
WITH candidates AS (
  SELECT id, CASE
    WHEN slug = 'clients/cursor' THEN replace(replace(replace(content_markdown,
      '1. 在 Models 中找到 OpenAI API Key，填写本站密钥。
2. 如果当前版本提供 Override OpenAI Base URL，启用并填写下方地址，再保存或验证。此地址来自本站密钥页默认端点。
3. 添加本站密钥分组支持的准确模型 ID，然后在聊天模型选择器中选中它。
4. 如果当前版本没有自定义 Base URL 选项，停止使用此接入方式；仅将本站密钥填入官方 OpenAI 地址不会生效。可改用本站教程支持的其他客户端。',
      '1. 打开 Cursor Settings → Network，将 HTTP Compatibility Mode 改为 HTTP/1.1。这是接入步骤，不是可选的故障排查。
2. 进入 Models，展开 API Keys。打开 OpenAI API Key，填入本站密钥。
3. 打开 Override OpenAI Base URL，填入当前线路 URL。地址以 /v1 结尾，不要写成 /chat/completions。设置页没有 Provider，也不要按旧示意去 Add Model。
4. 保存后回到对话，选择当前渠道可用模型并发送一条消息。如果当前版本没有自定义 Base URL，停止使用此接入方式；只把密钥填进官方 OpenAI 地址不会生效。'),
      '核对日期：2026-09-22。', '核对日期：2026-09-25。'),
      '配置自定义 OpenAI 地址，了解聊天模型与 Tab 补全的区别。',
      '先把 HTTP Compatibility Mode 改为 HTTP/1.1，再填写 OpenAI API Key 和线路 /v1 地址。')
    WHEN slug = 'clients/trae' THEN replace(replace(replace(replace(content_markdown,
      '按协议填写自定义模型，使用明确的完整 URL。',
      '按渠道选择 API 格式，线路地址不要带协议路径，完整 URL 保持关闭。'),
      '2. API 格式按下方字段选择：Claude/Antigravity 使用 Anthropic Messages，其余渠道采用 OpenAI Chat Completions。
3. 本教程开启“完整 URL”，填写含协议路径的完整请求地址。不要让客户端再补一次 /v1/messages 或 /chat/completions。',
      '2. API 格式按下方字段选择：Claude、Antigravity、DeepSeek、MiniMax 使用 Anthropic Messages，其余渠道采用 OpenAI Chat Completions。
3. 保持“完整 URL”关闭。自定义请求地址只填线路 URL，不要以斜杠结尾，也不要自带 /v1/messages 或 /chat/completions；客户端会把对应路径补到末尾。'),
      '404 常见于完整 URL 开关与地址形式不匹配。',
      '404 常见于完整 URL 被打开，或地址里已经写了客户端还会追加的路径。'),
      '核对日期：2026-09-22。', '核对日期：2026-09-25。')
    WHEN slug = 'clients/workbuddy' THEN replace(replace(replace(replace(content_markdown,
      '通过 Custom 模型与完整请求地址连接本站。',
      '通过自定义模型连接本站：接口地址包含 /chat/completions，自定义协议保持关闭。'),
      '2. 本教程明确开启 Custom Protocol，URL 因此使用完整 /v1/chat/completions 地址，客户端不再补全路径。
3. 若关闭 Custom Protocol，客户端会校验和补全路径，此时不能直接沿用本教程完整 URL 设定。',
      '2. 提供商选择「自定义 / Custom」。接口地址填写完整 /v1/chat/completions，自定义协议保持关闭。
3. 不要打开自定义协议后再把 /chat/completions 写进地址，否则可能重复拼接。'),
      '重复 /chat/completions 通常是 Custom Protocol 没有开启。核对当前使用的模型配置，不要误改了未启用的条目。',
      '重复 /chat/completions 通常是自定义协议被打开，同时又填写了已经包含该路径的地址。核对当前使用的模型配置，不要误改了未启用的条目。'),
      '核对日期：2026-09-22。', '核对日期：2026-09-25。')
    WHEN slug = 'clients/zcode' THEN replace(replace(replace(content_markdown,
      '1. 供应商名称填写“本站”，协议选择 OpenAI 兼容。
2. Base URL 和 API Key 按下方填写；地址为默认端点的 /v1 形式，不是完整 /chat/completions 请求路径。',
      '1. 供应商名称可自定义。当前界面没有协议下拉，不要去找 OpenAI Compatible 选项。
2. API Base URL 和 API Key 按下方填写；地址为默认端点的 /v1 形式，不是完整 /chat/completions 请求路径。'),
      '模型 ID 不存在时先核对分组权限。认证字段和协议必须匹配；本教程的 OpenAI 模式不能套用 Anthropic 专用端点。',
      '模型 ID 不存在时先核对分组权限。ZCode 使用 OpenAI 兼容的 /v1 地址，不要改成 /v1/messages。'),
      '核对日期：2026-09-22。', '核对日期：2026-09-25。')
    ELSE content_markdown END AS revised_content
  FROM help_documents
  WHERE slug IN ('clients/cursor', 'clients/trae', 'clients/workbuddy', 'clients/zcode')
    AND version = 1 AND updated_by IS NULL AND status = 'published'
    AND content_markdown = published_snapshot->>'content_markdown'
    AND publication_history = '[]'::jsonb
), revised AS (
  SELECT d.id, c.revised_content,
    CASE
      WHEN d.slug = 'clients/cursor' THEN '先把 HTTP Compatibility Mode 改为 HTTP/1.1，再填写 OpenAI API Key 和线路 /v1 地址。'
      WHEN d.slug = 'clients/trae' THEN '按渠道选择 API 格式，线路地址不要带协议路径，完整 URL 保持关闭。'
      WHEN d.slug = 'clients/workbuddy' THEN '通过自定义模型连接本站：接口地址包含 /chat/completions，自定义协议保持关闭。'
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
