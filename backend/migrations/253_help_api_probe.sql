-- Replace the static API examples with a short companion to the in-page probe.
-- Skip rows an operator has edited.
WITH seeds(slug, summary, content_markdown) AS (
VALUES
('api', $help253$按当前分组的协议发送一次最小探测。$help253$, $help253$## 怎么用
在上方选择分组、模型和已启用的密钥。协议按分组自动选择；该分组有多种协议时可以切换。页面不会自动发送。

探测是一次非流式请求，内容固定为 hi。它会消耗额度，并记入用量。

## 状态
| 状态 | 优先检查 |
| --- | --- |
| `401` | 密钥是否启用，以及鉴权头是否发出 |
| `403` | 分组、额度、账号状态 |
| `404` | 协议路径和模型 ID |
| `429` | 密钥限额、订阅窗口、上游限流 |

模型列表成功不等于这次探测成功。
$help253$)
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
