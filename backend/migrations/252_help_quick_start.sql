-- Align the quick-start article with the selectors already shown above it.
WITH seeds(slug, summary, content_markdown) AS (
VALUES
('quick-start', $help252$按上方的渠道和客户端完成第一次调用。$help252$, $help252$## 从上方进入
先选渠道，再选客户端。后面的步骤、分组和模型都跟随这次选择。密钥页的“使用教程”会直接打开对应客户端，并带入该密钥的平台和分组，不必再从本页重选。

## 只用一份配置
客户端页面中的配置是示例。`YOUR_API_KEY` 和 `YOUR_MODEL_ID` 要换成当前密钥及其可用模型。需要已填入真实密钥的配置时，在密钥页打开“使用密钥”。不要把示例和“使用密钥”保存成两套配置。

## 确认调用成功
写入配置后，关闭并重新打开终端或客户端，再发送一次最短消息。客户端收到正常回复，并且用量页出现对应请求，才算完成。只看到模型列表不算完成。这次请求可能产生费用。

## 失败时先核对
401：核对进入客户端的密钥是否与“使用密钥”中的一致。404：核对地址、协议和模型 ID 是否属于当前分组。429：核对额度与并发。一次只改一项。
$help252$)
)
UPDATE help_documents AS d
SET summary = s.summary,
    content_markdown = s.content_markdown,
    publication_history = COALESCE(d.publication_history, '[]'::jsonb) || jsonb_build_array(d.published_snapshot),
    version = d.version + 1,
    published_snapshot = d.published_snapshot || jsonb_build_object(
      'summary', s.summary,
      'content_markdown', s.content_markdown,
      'version', d.version + 1,
      'published_at', NOW()
    ),
    published_at = NOW(),
    updated_at = NOW()
FROM seeds AS s
WHERE d.slug = s.slug
  AND d.updated_by IS NULL
  AND d.status = 'published'
  AND d.content_markdown = d.published_snapshot->>'content_markdown'
  AND d.content_markdown IS DISTINCT FROM s.content_markdown;
