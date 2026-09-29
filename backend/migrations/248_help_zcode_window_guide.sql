-- Replace the untouched ZCode article with the current settings-window flow.
WITH candidates AS (
  SELECT id, replace(content_markdown,
    '1. 供应商名称可自定义。当前界面没有协议下拉，不要去找 OpenAI Compatible 选项。
2. API Base URL 和 API Key 按下方填写；地址为默认端点的 /v1 形式，不是完整 /chat/completions 请求路径。
3. 点击添加模型，填写上方已选模型的准确 ID。显示名称可以自定义，ID 不能随意缩写。
4. 开启供应商并在模型选择器中选择该模型。
5. 不要假设随意添加的 options 会被发送给服务器；自定义能力参数应以官方支持字段为准。',
    '1. 打开 ZCode 设置并进入「模型设置」，点击「+ 添加供应商」。供应商名称可自定义，当前界面没有协议下拉。
2. API Base URL 填当前线路的 /v1，不要写成 /chat/completions。API Key 填自己的密钥，然后保存供应商。
3. 在该供应商的模型列表中点击「添加模型」。Model ID 填当前选择的模型。上下文窗口由客户端按模型自动带出，没有带出时不要手填猜测值。
4. 返回工作区，点击输入框右下角的「选择模型」，展开本站供应商并选择刚添加的模型。
5. 发送一条消息。看到 ZCode 正常回复即表示配置完成。不要假设随意添加的 options 会被发送给服务器。') AS revised_content
  FROM help_documents
  WHERE slug = 'clients/zcode'
    AND version = 2 AND updated_by IS NULL AND status = 'published'
    AND content_markdown = published_snapshot->>'content_markdown'
), revised AS (
  SELECT d.id, c.revised_content,
    '按 ZCode 模型设置窗口添加供应商和模型，地址使用线路 /v1。' AS revised_summary
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
