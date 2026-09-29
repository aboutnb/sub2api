-- New client guides only; preserve operator edits and existing publication history.
WITH seeds(slug, title, category, summary, content_markdown, selector_schema) AS (
 VALUES
 ('clients/dsh', 'dsh 接入', 'dsh', '通过官方模型设置界面添加本站 Responses 供应商。', '## 前置条件
准备本站 API Key，并确认绑定分组与本页一致。模型参数跟随上方当前分组列表；默认按版本号优先选择，没有发布日期时不保证为最新发布。API 地址仅使用密钥页默认端点。文件或字段中的 YOUR_API_KEY 必须替换为自己的密钥。

## 安装
从 DeepSeek Harness 官方仓库的安装说明安装 dsh。参考命令 `npm install -g @deepseek-ai/dsh`；先核对当前发行版的 Node.js 要求，再运行 `dsh --version` 和 `dsh web`。

## 认证与项目配置
1. 打开 dsh Web 设置，依次选择 Settings → Models → Add model provider → Custom model API。
2. Provider ID 填 site，Base URL 使用下方本站地址，API 选择 openai-responses。不要在 Base URL 末尾加 /responses。
3. 将 API Key 填到凭据字段，模型 ID 使用上方当前分组模型。可使用 Fetch available models 获取模型列表。
4. 保存后在模型选择器中选中 site 下的模型。模型配置更改在下一次请求生效。
5. 官方新版高级配置位于 $DSH_HOME/profiles/<profile>/cordis.patch.yml。不要把旧版 ~/.dsh/settings.yaml 当作所有版本通用格式。

## 验证与成功判据
发送“请只回复 OK”或翻译一段简短公开文本，确认客户端返回正常结果，再到本站用量记录核对模型与请求。模型列表成功、安装版本号、连接测试都不等同于实际推理成功。需要工具或图片时另行验证，不在测试中自动修改项目文件。

## 常见错误
模型列表失败时核对 /v1 地址和密钥权限。Responses 与 Messages 不是可互换的协议；不要只换地址不改 API。模型能力以本站实际支持为准，不盲填上下文窗口或图片能力。

## 官方来源与适用版本
[官方来源](https://github.com/deepseek-ai/deepseek-harness/blob/master/docs/user/guide/providers.md)

核对日期：2026-09-22。适用具备上述自定义提供商或端点功能的版本；尚未记录本站真实客户端端到端验证。保留现有配置并合并，不覆盖其他供应商或凭据。
', '{"clients":["dsh"]}'::jsonb),
 ('clients/pi', 'Pi 接入', 'pi', '使用用户级 models.json 添加自定义 Responses Provider。', '## 前置条件
准备本站 API Key，并确认绑定分组与本页一致。模型参数跟随上方当前分组列表；默认按版本号优先选择，没有发布日期时不保证为最新发布。API 地址仅使用密钥页默认端点。文件或字段中的 YOUR_API_KEY 必须替换为自己的密钥。

## 安装
从 [Pi 官网](https://pi.dev/) 查看当前安装方式和支持的系统，完成安装后运行 `pi --version`。此教程仅在本站 OpenAI 渠道展示。

## 认证与项目配置
1. 创建 ~/.pi/agent 目录，Windows 使用用户主目录下对应路径。
2. 将下方 JSON 合并到 ~/.pi/agent/models.json，保留已有 providers；site 是本教程的自定义供应商 ID。
3. baseUrl 使用本站 /v1 地址，api 采用 openai-responses，models[].id 与上方选择一致。
4. apiKey 示例仅含占位符。若改为环境变量引用，官方当前格式为 "$SITE_API_KEY"，裸字符串 SITE_API_KEY 不会读取环境变量。
5. 启动 Pi，打开 /model，选择 site 下的模型。重新打开 /model 会重载文件。

## 验证与成功判据
发送“请只回复 OK”或翻译一段简短公开文本，确认客户端返回正常结果，再到本站用量记录核对模型与请求。模型列表成功、安装版本号、连接测试都不等同于实际推理成功。需要工具或图片时另行验证，不在测试中自动修改项目文件。

## 常见错误
无法选模型时检查是否配置了认证；JSON 中不能带注释或尾逗号。仅列出模型不代表 Responses 推理正常。

## 官方来源与适用版本
[官方来源](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/models.md)

核对日期：2026-09-22。适用具备上述自定义提供商或端点功能的版本；尚未记录本站真实客户端端到端验证。保留现有配置并合并，不覆盖其他供应商或凭据。
', '{"clients":["pi"],"platforms":["openai"]}'::jsonb),
 ('clients/openclaw', 'OpenClaw 接入', 'openclaw', '通过自定义 Provider 接入本站 Chat Completions 模型。', '## 前置条件
准备本站 API Key，并确认绑定分组与本页一致。模型参数跟随上方当前分组列表；默认按版本号优先选择，没有发布日期时不保证为最新发布。API 地址仅使用密钥页默认端点。文件或字段中的 YOUR_API_KEY 必须替换为自己的密钥。

## 安装
从 [OpenClaw 官方文档](https://docs.openclaw.ai/) 完成安装及初始化。先记录当前版本，备份 ~/.openclaw/openclaw.json，不要用示例覆盖已有机器人或渠道配置。

## 认证与项目配置
1. 下方内容合并到 ~/.openclaw/openclaw.json；models.mode 采用 merge，保留现有供应商。
2. models.providers.site.baseUrl 使用本站默认端点的 /v1 地址，api 为 openai-completions。
3. models 中的 id 是实际模型，agents.defaults.model.primary 使用 site/模型ID。
4. 把占位密钥替换为本站密钥；长期使用优先按官方 SecretRef 或环境变量替换方式管理凭据。
5. 按当前版本配置重载方式让设置生效，在聊天界面选择该模型。不要未经验证添加 compat 或模型能力声明。

## 验证与成功判据
发送“请只回复 OK”或翻译一段简短公开文本，确认客户端返回正常结果，再到本站用量记录核对模型与请求。模型列表成功、安装版本号、连接测试都不等同于实际推理成功。需要工具或图片时另行验证，不在测试中自动修改项目文件。

## 常见错误
如果模型名称包含斜杠，保留完整模型 ID；primary 仍需添加 site/ 前缀。普通对话和工具调用要分别验证。不要把 openai-completions 地址配置成 openai-responses。

## 官方来源与适用版本
[官方来源](https://docs.openclaw.ai/gateway/config-tools/custom-providers)

核对日期：2026-09-22。适用具备上述自定义提供商或端点功能的版本；尚未记录本站真实客户端端到端验证。保留现有配置并合并，不覆盖其他供应商或凭据。
', '{"clients":["openclaw"]}'::jsonb),
 ('clients/hermes', 'Hermes 接入', 'hermes', '使用 custom Provider、配置文件和独立环境文件连接本站。', '## 前置条件
准备本站 API Key，并确认绑定分组与本页一致。模型参数跟随上方当前分组列表；默认按版本号优先选择，没有发布日期时不保证为最新发布。API 地址仅使用密钥页默认端点。文件或字段中的 YOUR_API_KEY 必须替换为自己的密钥。

## 安装
从 [Hermes 官方文档](https://hermes-agent.nousresearch.com/docs/) 安装 Hermes Agent，并完成初始化。Windows 的支持环境以该版本官方说明为准。配置前备份 ~/.hermes。

## 认证与项目配置
1. 可先运行 hermes model 进入交互式模型设置，选择自定义 OpenAI 兼容提供商。
2. 手动配置时，将下面 model 节点合并到 ~/.hermes/config.yaml，provider 设为 custom。
3. 将 SITE_API_KEY 写入 ~/.hermes/.env；不要将凭据放进项目仓库，限制文件读取权限。
4. model.default 是当前模型 ID，model.base_url 是本站 /v1 地址，model.key_env 引用 SITE_API_KEY。
5. 重启相关终端或 Hermes 进程，确认选择的是 custom Provider。

## 验证与成功判据
发送“请只回复 OK”或翻译一段简短公开文本，确认客户端返回正常结果，再到本站用量记录核对模型与请求。模型列表成功、安装版本号、连接测试都不等同于实际推理成功。需要工具或图片时另行验证，不在测试中自动修改项目文件。

## 常见错误
OPENAI_BASE_URL 属于 openai-api Provider，不代替 custom 的 model.base_url。旧 LLM_MODEL 环境变量不应继续使用。YAML 缩进不正确会导致配置不生效。

## 官方来源与适用版本
[官方来源](https://hermes-agent.nousresearch.com/docs/integrations/providers)

核对日期：2026-09-22。适用具备上述自定义提供商或端点功能的版本；尚未记录本站真实客户端端到端验证。保留现有配置并合并，不覆盖其他供应商或凭据。
', '{"clients":["hermes"]}'::jsonb),
 ('clients/workbuddy', 'WorkBuddy 接入', 'workbuddy', '通过 Custom 模型与完整请求地址连接本站。', '## 前置条件
准备本站 API Key，并确认绑定分组与本页一致。模型参数跟随上方当前分组列表；默认按版本号优先选择，没有发布日期时不保证为最新发布。API 地址仅使用密钥页默认端点。文件或字段中的 YOUR_API_KEY 必须替换为自己的密钥。

## 安装
从 [WorkBuddy 官网](https://www.workbuddy.ai/) 下载并安装客户端，进入 Settings → Model → 添加模型 → Custom。

## 认证与项目配置
1. 使用下方 URL、API Key 和 Model 字段。
2. 本教程明确开启 Custom Protocol，URL 因此使用完整 /v1/chat/completions 地址，客户端不再补全路径。
3. 若关闭 Custom Protocol，客户端会校验和补全路径，此时不能直接沿用本教程完整 URL 设定。
4. 保存并选择新模型，先尝试不需要工具的简单对话。
5. 旧 ~/.codebuddy/models.json 可能仍兼容，但优先使用官方推荐的图形界面。

## 验证与成功判据
发送“请只回复 OK”或翻译一段简短公开文本，确认客户端返回正常结果，再到本站用量记录核对模型与请求。模型列表成功、安装版本号、连接测试都不等同于实际推理成功。需要工具或图片时另行验证，不在测试中自动修改项目文件。

## 常见错误
重复 /chat/completions 通常是 Custom Protocol 没有开启。核对当前使用的模型配置，不要误改了未启用的条目。

## 官方来源与适用版本
[官方来源](https://www.workbuddy.ai/docs/workbuddy/From-Beginner-to-Expert-Guide/Function-Description/Model)

核对日期：2026-09-22。适用具备上述自定义提供商或端点功能的版本；尚未记录本站真实客户端端到端验证。保留现有配置并合并，不覆盖其他供应商或凭据。
', '{"clients":["workbuddy"]}'::jsonb),
 ('clients/zcode', 'ZCode 接入', 'zcode', '添加 OpenAI 兼容供应商与当前分组模型。', '## 前置条件
准备本站 API Key，并确认绑定分组与本页一致。模型参数跟随上方当前分组列表；默认按版本号优先选择，没有发布日期时不保证为最新发布。API 地址仅使用密钥页默认端点。文件或字段中的 YOUR_API_KEY 必须替换为自己的密钥。

## 安装
从 [ZCode 官网](https://zcode.z.ai/) 安装客户端。进入模型选择器 → 管理模型 → 模型设置 → 添加供应商。

## 认证与项目配置
1. 供应商名称填写“本站”，协议选择 OpenAI 兼容。
2. Base URL 和 API Key 按下方填写；地址为默认端点的 /v1 形式，不是完整 /chat/completions 请求路径。
3. 点击添加模型，填写上方已选模型的准确 ID。显示名称可以自定义，ID 不能随意缩写。
4. 开启供应商并在模型选择器中选择该模型。
5. 不要假设随意添加的 options 会被发送给服务器；自定义能力参数应以官方支持字段为准。

## 验证与成功判据
发送“请只回复 OK”或翻译一段简短公开文本，确认客户端返回正常结果，再到本站用量记录核对模型与请求。模型列表成功、安装版本号、连接测试都不等同于实际推理成功。需要工具或图片时另行验证，不在测试中自动修改项目文件。

## 常见错误
模型 ID 不存在时先核对分组权限。认证字段和协议必须匹配；本教程的 OpenAI 模式不能套用 Anthropic 专用端点。

## 官方来源与适用版本
[官方来源](https://zcode.z.ai/cn/docs/configuration)

核对日期：2026-09-22。适用具备上述自定义提供商或端点功能的版本；尚未记录本站真实客户端端到端验证。保留现有配置并合并，不覆盖其他供应商或凭据。
', '{"clients":["zcode"]}'::jsonb),
 ('clients/trae', 'TRAE SOLO 接入', 'trae', '按协议填写自定义模型，使用明确的完整 URL。', '## 前置条件
准备本站 API Key，并确认绑定分组与本页一致。模型参数跟随上方当前分组列表；默认按版本号优先选择，没有发布日期时不保证为最新发布。API 地址仅使用密钥页默认端点。文件或字段中的 YOUR_API_KEY 必须替换为自己的密钥。

## 安装
从 [TRAE 官网](https://www.trae.ai/) 获取 TRAE SOLO；确认所用版本和账号提供添加自定义模型功能。不要把 TraeCode CLI 的 trae_cli.yaml 用作 SOLO 桌面版配置。

## 认证与项目配置
1. 打开模型设置 → 添加模型 → 自定义配置。
2. API 格式按下方字段选择：Claude/Antigravity 使用 Anthropic Messages，其余渠道采用 OpenAI Chat Completions。
3. 本教程开启“完整 URL”，填写含协议路径的完整请求地址。不要让客户端再补一次 /v1/messages 或 /chat/completions。
4. 填写模型 ID 与 API 密钥。只有确认所选模型支持图片时才开启多模态。
5. 提交后回到对话界面，选择刚添加的模型。当前版本没有此功能时，不要将本站密钥填入官方服务商登录栏。

## 验证与成功判据
发送“请只回复 OK”或翻译一段简短公开文本，确认客户端返回正常结果，再到本站用量记录核对模型与请求。模型列表成功、安装版本号、连接测试都不等同于实际推理成功。需要工具或图片时另行验证，不在测试中自动修改项目文件。

## 常见错误
404 常见于完整 URL 开关与地址形式不匹配。SOLO、IDE 和 CLI 不是同一种配置文件；核对安装版本。

## 官方来源与适用版本
[官方来源](https://www.trae.ai/)

核对日期：2026-09-22。适用具备上述自定义提供商或端点功能的版本；尚未记录本站真实客户端端到端验证。界面字段同时参考 Krill 的 TRAE SOLO 教程；版本差异仍需按实际客户端核对。保留现有配置并合并，不覆盖其他供应商或凭据。
', '{"clients":["trae"]}'::jsonb),
 ('clients/read-frog', 'Read Frog 接入', 'read-frog', '浏览器扩展使用自定义 OpenAI 兼容服务翻译文本。', '## 前置条件
准备本站 API Key，并确认绑定分组与本页一致。模型参数跟随上方当前分组列表；默认按版本号优先选择，没有发布日期时不保证为最新发布。API 地址仅使用密钥页默认端点。文件或字段中的 YOUR_API_KEY 必须替换为自己的密钥。

## 安装
从 [Read Frog 官网](https://www.readfrog.app/) 的官方安装入口安装扩展。浏览器权限由用户确认，不要在非官方页面输入密钥。

## 认证与项目配置
1. 打开扩展选项 → API 服务商 → 添加 OpenAI 兼容服务商。
2. 填写下方 Base URL、API Key、模型 ID，启用服务商。
3. 将翻译所用模型切换为新建服务商下的当前模型，运行连接测试。
4. 在普通测试网页选择一小段非敏感文字进行翻译，核对本站用量。
5. 页面内容可能被扩展发送到模型接口；不要用私密文档进行首次测试。

## 验证与成功判据
发送“请只回复 OK”或翻译一段简短公开文本，确认客户端返回正常结果，再到本站用量记录核对模型与请求。模型列表成功、安装版本号、连接测试都不等同于实际推理成功。需要工具或图片时另行验证，不在测试中自动修改项目文件。

## 常见错误
浏览器网络、CORS、代理或扩展权限可能影响请求。翻译成功不意味着需要结构化输出的 AI 操作也可用，应分别验证。

## 官方来源与适用版本
[官方来源](https://www.readfrog.app/zh/docs/providers/openai-compatible-providers)

核对日期：2026-09-22。适用具备上述自定义提供商或端点功能的版本；尚未记录本站真实客户端端到端验证。保留现有配置并合并，不覆盖其他供应商或凭据。
', '{"clients":["read-frog"]}'::jsonb),
 ('clients/kiss-translator', 'Kiss Translator 接入', 'kiss-translator', '使用内置 OpenAI 翻译服务与完整 API 请求地址。', '## 前置条件
准备本站 API Key，并确认绑定分组与本页一致。模型参数跟随上方当前分组列表；默认按版本号优先选择，没有发布日期时不保证为最新发布。API 地址仅使用密钥页默认端点。文件或字段中的 YOUR_API_KEY 必须替换为自己的密钥。

## 安装
从 [Kiss Translator 官方仓库](https://github.com/fishjar/kiss-translator) 的发行入口安装扩展或用户脚本，核对来源和版本。

## 认证与项目配置
1. 进入扩展设置的翻译服务，添加或选择内置 OpenAI 服务，而不是通用自定义 JSON 接口。
2. 将下方 API URL 填入接口地址，包含 /v1/chat/completions；填写本站密钥与上方模型。
3. 如果有自定义模型输入，填准确模型 ID，保留内置 OpenAI 请求和响应处理逻辑。
4. 保存并设为当前翻译服务，选一小段公开文字翻译。
5. 若改用通用自定义接口，必须按官方 custom-api_v2.md 配置 OpenAI 请求体和响应 Hook；默认 {text, from, to} 不是 Chat Completions 格式。

## 验证与成功判据
发送“请只回复 OK”或翻译一段简短公开文本，确认客户端返回正常结果，再到本站用量记录核对模型与请求。模型列表成功、安装版本号、连接测试都不等同于实际推理成功。需要工具或图片时另行验证，不在测试中自动修改项目文件。

## 常见错误
请求体没有 messages 通常是误选通用自定义接口。不要把完整请求 URL 填入只收 Base URL 的其他客户端。

## 官方来源与适用版本
[官方来源](https://github.com/fishjar/kiss-translator/blob/dev/custom-api_v2.md)

核对日期：2026-09-22。适用具备上述自定义提供商或端点功能的版本；尚未记录本站真实客户端端到端验证。保留现有配置并合并，不覆盖其他供应商或凭据。
', '{"clients":["kiss-translator"]}'::jsonb),
 ('clients/immersive-translate', '沉浸式翻译 接入', 'immersive-translate', '使用 OpenAI 自定义地址和模型完成网页翻译。', '## 前置条件
准备本站 API Key，并确认绑定分组与本页一致。模型参数跟随上方当前分组列表；默认按版本号优先选择，没有发布日期时不保证为最新发布。API 地址仅使用密钥页默认端点。文件或字段中的 YOUR_API_KEY 必须替换为自己的密钥。

## 安装
从 [沉浸式翻译官网](https://immersivetranslate.com/) 安装官方扩展。打开设置，进入翻译服务的 OpenAI 设置。

## 认证与项目配置
1. 在 OpenAI 服务中填写本站 API Key。
2. 展开更多设置，启用自定义 API 地址，填写下方完整 /v1/chat/completions URL。
3. 在自定义模型设置填入上方模型。版本采用列表语法时，按官方说明用 +模型ID 添加模型，再选中它。
4. 保存并选择 OpenAI 作为当前翻译服务，先测试一段短文本。
5. 翻译大篇幅网页会消耗额度；不要把“测试服务”成功等同于整页翻译和所有 AI 功能都可用。

## 验证与成功判据
发送“请只回复 OK”或翻译一段简短公开文本，确认客户端返回正常结果，再到本站用量记录核对模型与请求。模型列表成功、安装版本号、连接测试都不等同于实际推理成功。需要工具或图片时另行验证，不在测试中自动修改项目文件。

## 常见错误
模型不出现在下拉列表时检查是否已添加并保存自定义模型。429 可降低并发并检查分组额度；不要反复重试整页请求。

## 官方来源与适用版本
[官方来源](https://immersivetranslate.com/docs/services/openai/)

核对日期：2026-09-22。适用具备上述自定义提供商或端点功能的版本；尚未记录本站真实客户端端到端验证。保留现有配置并合并，不覆盖其他供应商或凭据。
', '{"clients":["immersive-translate"]}'::jsonb)
)
INSERT INTO help_documents (slug, title, category, summary, content_markdown, selector_schema, status, version, published_snapshot, published_at)
SELECT slug, title, category, summary, content_markdown, selector_schema, 'published', 1,
 jsonb_build_object('title', title, 'category', category, 'summary', summary, 'content_markdown', content_markdown, 'selector_schema', selector_schema, 'version', 1, 'published_at', NOW()), NOW()
FROM seeds
ON CONFLICT (slug) DO NOTHING;

