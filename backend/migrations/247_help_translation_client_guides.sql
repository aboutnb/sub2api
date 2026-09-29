-- Align untouched translation and WorkBuddy articles with the current client screens.
WITH candidates AS (
  SELECT id, CASE
    WHEN slug = 'clients/read-frog' THEN replace(replace(content_markdown,
      '3. 先在 API Providers 运行 Test Connection，再检查 Feature Providers 中翻译功能绑定的模型，选择新建服务商下的当前模型。',
      '3. 按图填写 API Key、Base URL 和 Model，勾选 Enter the name of the custom model。点击 Test Connection，提示成功即表示连接配置完成。Feature Providers 与 Advanced Options 保持折叠。'),
      '核对日期：2026-09-24。', '核对日期：2026-09-25。')
    WHEN slug = 'clients/kiss-translator' THEN replace(replace(replace(content_markdown,
      '2. 将下方 API URL 填入接口地址，Claude 服务包含 /v1/messages，OpenAI 服务包含 /v1/chat/completions；填写本站密钥与上方模型。',
      '2. 按图填写 URL、Key 和 Model。URL 必须是完整地址：Claude 服务包含 /v1/messages，OpenAI 服务包含 /v1/chat/completions。Sort Order 填 0，翻译风格 formal，Temperature 0，Max Tokens 20480。'),
      '4. 保存并设为当前翻译服务，刷新测试网页后选一小段公开文字翻译。',
      '4. 保存后刷新页面，选择刚添加的服务，再翻译一小段公开文字。不要改用通用自定义 JSON 接口。'),
      '核对日期：2026-09-24。', '核对日期：2026-09-25。')
    WHEN slug = 'clients/immersive-translate' THEN replace(replace(replace(replace(content_markdown,
      '打开设置，按本页 Provider 进入 Claude 或 OpenAI 翻译服务设置。',
      '打开设置，在「其他/自定义」中选择当前渠道对应的 Claude 1 或 OpenAI 1。'),
      '3. 在自定义模型设置填入上方模型。版本采用列表语法时，按官方说明用 +模型ID 添加模型，再选中它。',
      '3. 勾选「输入自定义模型名称」，填入上方模型。AI 智能上下文保持关闭，翻译策略选择「通用」。'),
      '4. 保存后打开浏览器扩展弹窗，选择刚配置的 Claude 或 OpenAI 作为当前服务，先测试一段短文本。',
      '4. 保存后打开任意英文网页查看翻译。不要把扩展弹窗里的旧默认服务当成完成标准。'),
      '核对日期：2026-09-24。', '核对日期：2026-09-25。')
    WHEN slug = 'clients/workbuddy' THEN replace(content_markdown,
      '2. 提供商选择「自定义 / Custom」。接口地址填写完整 /v1/chat/completions，自定义协议保持关闭。',
      '2. 提供商选择「自定义 / Custom」。接口地址填写完整 /v1/chat/completions，自定义协议保持关闭。高级配置里工具调用、图片输入和思考模式按客户端默认勾选；模型不支持图片或工具时再关闭。默认思考强度选自动，输入和输出长度不要猜测。')
    ELSE content_markdown END AS revised_content
  FROM help_documents
  WHERE slug IN ('clients/read-frog', 'clients/kiss-translator', 'clients/immersive-translate', 'clients/workbuddy')
    AND version = 2 AND updated_by IS NULL AND status = 'published'
    AND content_markdown = published_snapshot->>'content_markdown'
), revised AS (
  SELECT d.id, c.revised_content,
    CASE
      WHEN d.slug = 'clients/read-frog' THEN '在 API Providers 按渠道选择服务商，Base URL 使用 /v1，Test Connection 成功后再翻译。'
      WHEN d.slug = 'clients/kiss-translator' THEN '使用内置 Claude 或 OpenAI 服务，填写完整 URL、Sort Order 0 和模型。'
      WHEN d.slug = 'clients/immersive-translate' THEN '在其他/自定义填写完整接口地址；保存后打开英文网页查看翻译。'
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
