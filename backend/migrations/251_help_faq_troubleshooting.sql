-- Replace the untouched FAQ with the ordered troubleshooting guide.
WITH seeds(slug, summary, content_markdown) AS (
VALUES
('faq', $help251$按固定顺序排查 Key、分组、模型、地址和协议。$help251$, $help251$## 按固定顺序排查

1. **Key**：重新复制，检查启用状态、有效期、额度和 IP 规则。
2. **分组**：确认 Key 已分配目标分组。
3. **模型**：确认目标 ID 仍在该 Key 的 `/models` 中。
4. **Base URL**：检查最终请求 URL，排除重复 `/v1` 或端点。
5. **协议**：用非流式、无工具的最小请求验证 Responses、Messages 或 Chat Completions。
6. **高级参数**：移除图片、工具、推理档位、长上下文和客户端插件。
7. **上游**：最后再判断容量、限流、账号状态或临时故障。

## 常见状态码

| 状态 | 优先检查 |
| --- | --- |
| `400` | 请求结构、工具 schema、推理参数、图片格式 |
| `401` | Key、鉴权头、环境变量是否进入客户端进程 |
| `403` | 分组、余额、账号状态、IP 或风控策略 |
| `404` | Base URL、重复端点、模型 ID、协议未开放 |
| `429` | Key 限额、订阅窗口、上游限流和 `Retry-After` |
| `5xx` | 请求 ID、上游容量、超时和故障转移结果 |

## 模型表现或速度异常

不要先假设模型被替换。使用相同 Key、模型、协议和最小输入重复两到三次，并核对：

- 客户端实际发送的模型 ID 和推理档位；
- 是否发生上下文压缩、工具调用或自动子代理；
- 用量页中的请求模型、上游模型、首 Token、总耗时和状态；
- 同一时间是否有 `429`、容量或故障转移记录。

容量波动可能只影响某个上游账号或模型。不要在没有时间、模型和请求 ID 的情况下归因到整个服务。

## 参数不兼容

若纯文本成功、开启高级能力后失败，每次只恢复一项：流式、system、图片、工具、推理档位、长上下文。第三方客户端可能发送本站或上游不接受的字段；`400` 时保存脱敏请求结构，不要连续重试相同错误。

## 提交诊断信息

提供客户端与完整版本、操作系统、发生时间和时区、Key ID（不是 Key 内容）、分组、请求模型、最终 URL、协议、stream、状态码、耗时、`X-Request-ID` 和脱敏错误体。截图必须遮住 Key、Cookie、账号、源码和对话内容。

## 最小化与重试策略

从 curl 非流式单消息开始，成功后按顺序恢复客户端和高级能力。不要对 `400`、`401`、`403` 自动重试；对 `429` 和临时 `5xx` 尊重 `Retry-After` 或使用带抖动的指数退避，并设置最大重试次数。

排查模型质量或容量时，必须关联准确的用户代理、请求模型和窄时间窗口。仅凭模型显示名称、客户端弹窗或一条主观回答不能确定根因。
$help251$)
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
