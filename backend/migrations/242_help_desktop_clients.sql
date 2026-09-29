-- Seed new guides without replacing operator-maintained articles.
WITH seeds(slug, title, category, summary, content_markdown, selector_schema) AS (
 VALUES
 ('clients/cherry-studio', 'Cherry Studio 接入', 'cherry-studio', '添加本站供应商、填写默认端点、选择模型并开始对话。', $guide$
## 前置条件
准备本站 API Key，确认绑定的分组与页面上方一致，并在密钥页查看该分组的可用模型。这里使用 OpenAI 兼容接口，不需要登录 OpenAI 官方账号。

## 安装
从 [Cherry Studio 官网](https://www.cherry-ai.com/) 下载适合 macOS、Windows 或 Linux 的客户端。安装并启动后，进入“设置 → 模型服务”。已安装的用户先记录当前版本并保留原有供应商配置。

## 认证与项目配置
1. 添加一个新的供应商，名称可填“本站”，供应商类型选择 OpenAI。
2. 将下面示意中的 API Host 和 API Key 分别填入对应字段。API Host 使用本站默认端点的 `/v1` 地址，不要填写完整的 `/chat/completions` 路径。
3. 查看客户端显示的最终请求地址，应只有一层 `/v1`，并以 `/chat/completions` 结束。若出现重复 `/v1/v1`，检查该版本的自动补全规则，不要自行切换到其他服务商地址。
4. 获取模型列表，或手动添加密钥页显示的准确模型 ID；将示意中的 YOUR_MODEL_ID 替换成它。模型显示名称可自定义，模型 ID 不可随意修改。
5. 启用该供应商，在新对话的模型选择器中选择刚添加的模型。

## 验证与成功判据
在新对话中发送“请只回复 OK”。收到模型回复后，到本站用量记录核对请求。供应商连接检查或模型列表成功仅证明部分链路可用，不能代替实际对话验证。

## 常见错误
- 401：检查密钥是否有效，以及输入时是否带入空格。
- 404：检查 API Host 拼接结果与模型 ID。不要混用 Chat Completions、Responses 和 Gemini 原生地址。
- 能列出模型但不能对话：核对分组额度、模型权限和请求日志。
- 工具或图片失败：先验证纯文本，再使用该分组明确支持相应能力的模型。

## 官方来源与验证记录
[Cherry Studio 官网](https://www.cherry-ai.com/) · [官方文档](https://docs.cherry-ai.com/) · [官方源代码与发行版本](https://github.com/CherryHQ/cherry-studio)

适用范围：提供 OpenAI 供应商、自定义 API Host 和模型设置的桌面版本。整理日期：2026-09-22；官方文档正文抓取未完成，地址自动补全细节需按安装版本核对。本教程尚未记录真实客户端推理验证，不标记为实测通过。
$guide$, '{"clients":["cherry-studio"],"protocols":["chat-completions"]}'::jsonb),
 ('clients/cursor', 'Cursor 接入', 'cursor', '配置自定义 OpenAI 地址，了解聊天模型与 Tab 补全的区别。', $guide$
## 前置条件
准备 OpenAI 渠道分组的本站 API Key 和实际可用模型 ID。本教程仅展示 OpenAI 渠道：Cursor 不同版本与模型可能使用 Responses 或 Chat Completions，不能把所有渠道视为同等兼容。

Cursor 的请求会经过其服务器。本站默认端点必须能被 Cursor 服务器访问；localhost、127.0.0.1 和局域网地址不能用于这种接入。不要为此随意公开本地服务，应由站点管理员配置正式默认端点。

## 安装
从 [Cursor 官网](https://cursor.com/downloads) 下载并安装客户端，打开项目后进入 Cursor Settings → Models。界面位置可能随版本调整。

## 认证与项目配置
1. 在 Models 中找到 OpenAI API Key，填写本站密钥。
2. 如果当前版本提供 Override OpenAI Base URL，启用并填写下方地址，再保存或验证。此地址来自本站密钥页默认端点。
3. 添加本站密钥分组支持的准确模型 ID，然后在聊天模型选择器中选中它。
4. 如果当前版本没有自定义 Base URL 选项，停止使用此接入方式；仅将本站密钥填入官方 OpenAI 地址不会生效。可改用本站教程支持的其他客户端。
5. 自定义地址可能影响其他模型请求。保留原设置，切换回官方供应商时检查覆盖地址是否仍启用。

## 验证与成功判据
在 Chat 中发送一条简短纯文本消息，核对模型回复和本站用量记录。再单独验证所需 Agent 功能。自有 API Key 不覆盖 Cursor Tab 补全，不能用补全是否正常判断本站接入结果。

## 常见错误
- 地址验证失败：检查正式端点是否可从公网访问、证书是否有效，以及密钥和模型权限。
- 缺少 messages 或 input：可能是客户端采用的协议与网关处理不匹配。记录模型、Cursor 版本和请求 ID，分别核对 Responses 与 Chat Completions。
- 官方模型也失败：检查全局 Base URL 覆盖是否把其他供应商请求改到了本站。

## 官方来源与验证记录
[Cursor API Keys](https://cursor.com/help/models-and-usage/api-keys) · [官方论坛：自定义地址兼容问题](https://forum.cursor.com/t/the-custom-override-of-the-openai-base-url-is-unusable/152675)

核对日期：2026-09-22。官方帮助确认自有密钥的聊天模型范围与服务器转发；Base URL 限制参考官方论坛的问题记录，不视为当前所有版本保证。适用具有 Override OpenAI Base URL 的版本；尚未记录真实客户端推理验证。
$guide$, '{"clients":["cursor"],"platforms":["openai"],"protocols":["responses","chat-completions"]}'::jsonb),
 ('clients/cline', 'Cline 接入', 'cline', '在 VS Code 中通过 OpenAI Compatible 供应商连接本站。', $guide$
## 前置条件
准备 VS Code、本站 API Key 和当前分组的模型 ID。用于编辑文件或执行工具时，模型必须支持对应工具调用能力；基础聊天通过不代表所有 Agent 功能可用。

## 安装
在 VS Code 扩展市场搜索 Cline，核对发布者与 [Cline 官方安装文档](https://docs.cline.bot/getting-started/installing-cline) 后安装。打开侧边栏 Cline 面板，进入设置。

## 认证与项目配置
1. 将 API Provider 设为 OpenAI Compatible，不要选择需要官方账号登录的模式。
2. 填写下面的 Base URL、API Key 与 Model ID。地址已使用本站默认端点并补齐 `/v1`，不要再追加 `/chat/completions`。
3. 上下文窗口、最大输出和图片能力按所选模型实际能力设置，不沿用其他模型的能力或价格。
4. 若设置区分 Plan 和 Act 的模型配置，分别检查两种模式的供应商、地址与模型。保存后返回对话。
5. 不需要向项目写入这些凭据，也不要提交包含密钥的设置文件。

## 验证与成功判据
先在 Plan 模式发送简短问题，收到正常回复并核对本站用量记录。需要 Agent 功能时，再在测试项目中验证一次工具调用，并按客户端提示审核操作。模型列表与工具验证应分别记录。

## 常见错误
- 404：Base URL 应只含一层 `/v1`；核对模型 ID 的大小写和前缀。
- 普通回答正常但工具调用失败：检查模型能力和工具协议，不能仅修改模型名称冒充支持。
- 切换模式后失败：核对 Plan 与 Act 是否使用不同供应商配置。

## 官方来源与验证记录
[Cline OpenAI Compatible](https://docs.cline.bot/provider-config/openai-compatible) · [官方安装说明](https://docs.cline.bot/getting-started/installing-cline)

核对日期：2026-09-22。适用提供 OpenAI Compatible 设置的 Cline 版本。本站尚未记录具体扩展版本与端到端推理验证；以实际安装界面及官方说明为准。
$guide$, '{"clients":["cline"],"protocols":["chat-completions"]}'::jsonb),
 ('clients/roo-code', 'Roo Code 接入', 'roo-code', '配置 OpenAI Compatible，分别验证对话与原生工具调用。', $guide$
## 前置条件
准备 VS Code、本站 API Key 与支持原生工具调用的模型。Roo Code 的 OpenAI Compatible 供应商要求模型支持相应工具格式；仅支持普通文本的模型不满足 Agent 使用条件。

## 安装
从 [Roo Code 官方仓库](https://github.com/RooCodeInc/Roo-Code) 查看发行及安装说明，核对扩展来源、版本和维护状态后安装。打开 Roo Code 面板，进入设置。不要使用来路不明的同名扩展。

## 认证与项目配置
1. 创建或选择一个 API 配置，供应商选择 OpenAI Compatible。
2. 填写下方 Base URL、API Key 与 Model ID。Base URL 来自本站默认端点，不使用 Roo 官方服务地址，也不使用参考站地址。
3. 根据模型能力设置上下文、最大输出和图片支持；无法确认的能力不要启用。
4. 保存配置，确认当前模式选择的是刚设置的 API 配置。多个配置之间切换时检查地址和模型。

## 验证与成功判据
先发送一个不需要修改文件的简短请求，在本站用量记录核对实际推理。再在测试项目中验证一次工具调用，按扩展提示审批。记录扩展版本、模型 ID、对话结果和工具结果，不把连接测试等同于全部功能可用。

## 常见错误
- 工具调用格式错误：选择支持 OpenAI 原生工具调用的模型，并核对上游兼容性。
- 模型不存在：填写当前密钥所属分组的模型 ID，显示名称不能替代 ID。
- 切换模式后仍用旧地址：核对当前选择的 API 配置，避免修改了未启用的配置。

## 官方来源与验证记录
[Roo Code OpenAI Compatible](https://docs.roocode.com/providers/openai-compatible) · [官方仓库](https://github.com/RooCodeInc/Roo-Code)

核对日期：2026-09-22；供应商文档标注更新于 2026-05-15。适用提供 OpenAI Compatible 的版本。尚未记录本站真实客户端推理验证。
$guide$, '{"clients":["roo-code"],"protocols":["chat-completions"]}'::jsonb)
)
INSERT INTO help_documents (slug, title, category, summary, content_markdown, selector_schema, status, version, published_snapshot, published_at)
SELECT slug, title, category, summary, content_markdown, selector_schema, 'published', 1,
 jsonb_build_object('title', title, 'category', category, 'summary', summary, 'content_markdown', content_markdown, 'selector_schema', selector_schema, 'version', 1, 'published_at', NOW()), NOW()
FROM seeds
ON CONFLICT (slug) DO NOTHING;
