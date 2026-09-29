-- Tavern channel guides. These clients are not new gateway platforms.
WITH seeds(slug, title, category, summary, content_markdown, selector_schema) AS (
 VALUES
 ('clients/sillytavern', 'SillyTavern 接入', 'sillytavern', '在酒馆渠道用 Chat Completion 的自定义 OpenAI 兼容端点接入。', '## 前置条件
准备本站 API Key，并选择上方分组。酒馆不是新的网关平台，客户端只发送 OpenAI Chat Completions。地址使用当前线路的 /v1，不要自行追加 /chat/completions，密钥不要加 Bearer。

可以对接：
- 分组下拉中的分组。这里列出未开启「仅 Claude Code」的分组，包含 OpenAI、Grok、Claude、Gemini、Antigravity，以及 Kimi、智谱、DeepSeek、MiniMax 等隐藏渠道下的分组。
- OpenAI、Grok、Kimi、智谱、DeepSeek、MiniMax 由 OpenAI Chat Completions 网关转发。Claude、Gemini、Antigravity 由本站转换后转发。合成分组只接受该分组能够解析的模型。

不能对接：
- 开启「仅 Claude Code」的分组会拒绝 Chat Completions，只接受 /v1/messages，因此不会出现在本渠道。
- 图片、Embedding、语音、视频和审核模型，以及 /v1/completions 文本补全。
- SillyTavern 的内置 OpenAI、Claude、Text Completion、Kobold、NovelAI。

## 安装
按 [SillyTavern 官方仓库](https://github.com/SillyTavern/SillyTavern) 的当前安装说明安装，并对照 [Chat Completions 文档](https://docs.sillytavern.app/usage/api-connections/openai/)。

## 认证与项目配置
1. API 选择「聊天补全 / Chat Completion」，不要选「文本补全」。
2. 聊天补全来源选择「自定义（兼容 OpenAI）」。不要选内置 OpenAI 或 Claude。
3. 自定义端点（基础 URL）填下方 /v1，不要以斜杠结尾，也不要写成 /chat/completions。客户端会自动补上该路径。
4. 自定义 API 密钥填本站密钥。界面标为选填，接入本站时必须填写，且不要加 Bearer。
5. 点击连接。若 /v1/models 有返回，从列表选择上方模型；否则在「输入模型名」手填完整模型 ID。
6. 点击「发送测试消息」。只有请求实际成功、SillyTavern 仍报警时，才勾选「绕过 API 状态检查」。
7. 若上游拒绝连续同角色消息，再把提示词后处理改为合并连续同角色。不要一开始就选「单条用户消息」。

## 验证与成功判据
发送“请只回复 OK”或一句短对话，确认 SillyTavern 收到回复，再到本站用量记录核对模型和请求。模型列表、连接状态和测试消息按钮本身不等于角色对话已经可用。

## 常见错误
地址多写了 /chat/completions 时，实际请求会变成 /v1/chat/completions/chat/completions。401 时检查密钥是否误加 Bearer，以及分组是否仅限 Claude Code。模型列表为空时手填模型 ID，不要改用文本补全。

## 官方来源与适用版本
[官方来源](https://docs.sillytavern.app/usage/api-connections/openai/)

核对日期：2026-09-25。适用当前提供自定义 Chat Completion 端点的版本；尚未记录本站真实客户端端到端验证。
', '{"clients":["sillytavern"],"platforms":["tavern"]}'::jsonb),
 ('clients/tavernai', 'TavernAI 接入', 'tavernai', '在酒馆渠道用 TavernAI 2 的 Custom Chat Completions 接入。', '## 前置条件
准备本站 API Key，并选择上方分组。这里的 TavernAI 指 TavernAI 2，不是旧版 TavernAI 的反向代理。客户端使用 Chat Completions，地址填当前线路的 /v1。

可以对接的分组与 SillyTavern 相同：未开启「仅 Claude Code」的文字对话分组。OpenAI、Grok、Kimi、智谱、DeepSeek、MiniMax 走 OpenAI Chat Completions 网关；Claude、Gemini、Antigravity 由本站转换；合成分组只接受它能解析的模型。

不能对接：仅 Claude Code 分组、图片与 Embedding 等非对话模型、Text Completions，以及 Media Tools 的生图提供商。不要把内置 OpenAI 提供商的密钥页当成完成配置。

## 安装
从 [TavernAI 官方仓库](https://github.com/TavernAI/TavernAI) 下载当前 TavernAI 2 发行包。旧版说明见 TavernAI-v1，不要按旧版反向代理填写。快速开始只说明打开 provider settings：https://tavernai.net/docs/quick-start/

## 认证与项目配置
1. 点击顶部插头图标，打开「AI 提供商连接」。不要停在 Provider 为 OpenAI 的内置页。
2. 基础提供商选择 Custom。模型端点选择 Chat Completions，不要选 Text Completions。当前版本没有模型端点下拉时，保持 Custom 即可。
3. API 地址填下方 /v1，不要写成 /chat/completions。若界面出现「使用直接 API 地址（忽略所选端点）」，打开它，避免地址被所选端点模板改写。
4. API 密钥填本站密钥，不要加 Bearer。模型 ID 填上方模型；也可以先点提供商模型的 Load，失败后再手填。
5. 点击连接。连接检查失败但实际消息成功时，才打开「跳过连接检查」。官方快速开始截图在演示内置 OpenAI 时勾选了它，本站先不要勾。
6. Media Tools、ComfyUI 和本地 GGUF 不在本教程范围内。

## 验证与成功判据
连接成功后发送一条短文字消息，确认 TavernAI 收到回复，再到本站用量记录核对模型与请求。Load 成功或状态显示 Connected 都不等于对话可用。

## 常见错误
选成内置 OpenAI 后无法改写本站地址。API 地址多写 /chat/completions，或没打开直接地址，都会打到错误路径。Text Completions 会请求 /v1/completions，本教程不使用该接口。

## 官方来源与适用版本
[官方来源](https://github.com/TavernAI/TavernAI)

核对日期：2026-09-25。字段名来自当前仓库 locales/app 的 en.json 与 zh-CN.json，以及 2.0–2.2 更新说明中的 Custom provider Chat Completions。官方快速开始没有逐项列出这些字段。尚未记录本站真实客户端端到端验证。
', '{"clients":["tavernai"],"platforms":["tavern"]}'::jsonb)
)
INSERT INTO help_documents (slug, title, category, summary, content_markdown, selector_schema, status, version, published_snapshot, published_at)
SELECT slug, title, category, summary, content_markdown, selector_schema, 'published', 1,
 jsonb_build_object('title', title, 'category', category, 'summary', summary, 'content_markdown', content_markdown, 'selector_schema', selector_schema, 'version', 1, 'published_at', NOW()), NOW()
FROM seeds
ON CONFLICT (slug) DO NOTHING;
