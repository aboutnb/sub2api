-- Tighten published help articles. Skip rows an operator has edited.
WITH seeds(slug, summary, content_markdown) AS (
VALUES
('api', $help250$区分认证测试与真实推理。$help250$, $help250$## 前置条件
从密钥页复制接口根地址、密钥和可用模型。下面命令为 macOS/Linux 示例。先在本机设置 API_BASE_URL、API_KEY 和 MODEL，不要把真实值写进帮助文档。

## 模型列表
```sh
curl "$API_BASE_URL/v1/models" -H "Authorization: Bearer $API_KEY"
```

## Responses 推理
仅适用于支持 Responses 的分组。实际请求可能产生费用。
```sh
curl "$API_BASE_URL/v1/responses" -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" -d "{\"model\":\"$MODEL\",\"input\":\"Reply OK\",\"max_output_tokens\":32}"
```

## 配置与成功判据
API_BASE_URL 为不带 /v1 的根地址；含额外网关前缀时保留该前缀。模型列表返回 200 后，仍需确认推理响应有实际模型输出。Chat Completions、Messages 和 Gemini 原生接口使用不同路径及请求体。

## 常见错误
401 检查密钥；403 检查权限；404 检查协议和模型；429 检查额度与并发。排查时不要发送含密钥的完整请求头。

## 官方来源
[OpenAI Responses](https://platform.openai.com/docs/api-reference/responses)
$help250$),
('clients/cherry-studio', $help250$添加本站供应商、填写默认端点、选择模型并开始对话。$help250$, $help250$## 前置条件
准备本站 API Key，确认绑定的分组与页面上方一致，并在密钥页查看该分组的可用模型。协议随本站平台切换：Claude、Antigravity、DeepSeek、MiniMax 使用 Anthropic；其余平台使用 OpenAI 兼容接口。无需登录上游官方账号。

## 安装
从 [Cherry Studio 官网](https://www.cherry-ai.com/) 下载适合 macOS、Windows 或 Linux 的客户端。安装并启动后，进入“设置 → 模型服务”。已安装的用户先记录当前版本并保留原有供应商配置。

## 认证与项目配置
1. 添加一个新的供应商，名称可填“本站”，供应商类型按上方当前渠道生成的 Provider 字段选择，不能固定为 OpenAI。
2. 将下面示意中的 API Host 和 API Key 分别填入对应字段。Anthropic 的 API Host 使用本站默认端点根地址（Antigravity 带 /antigravity 前缀）；OpenAI 使用 /v1 基础地址。不要填写完整消息路径。
3. 查看客户端显示的最终请求地址，Anthropic 应以 /v1/messages 结束，OpenAI 应以 /v1/chat/completions 结束。若出现重复 /v1/v1，检查该版本的自动补全规则，不要自行切换到其他服务商地址。
4. 获取模型列表，或手动添加密钥页显示的准确模型 ID；将示意中的 YOUR_MODEL_ID 替换成它。模型显示名称可自定义，模型 ID 不可随意修改。
5. 启用该供应商，在新对话的模型选择器中选择刚添加的模型。

## 验证与成功判据
在新对话中发送一条短消息。收到模型回复后，到本站用量记录核对请求。供应商连接检查或模型列表成功不能代替实际对话。

## 常见错误
- 401：检查密钥是否有效，以及输入时是否带入空格。
- 404：检查 API Host 拼接结果与模型 ID。不要混用 Chat Completions、Responses 和 Gemini 原生地址。
- 能列出模型但不能对话：核对分组额度、模型权限和请求日志。
- 工具或图片失败：先验证纯文本，再使用该分组明确支持相应能力的模型。

## 官方来源
[Cherry Studio 官网](https://www.cherry-ai.com/) · [官方文档](https://docs.cherry-ai.com/) · [官方源代码与发行版本](https://github.com/CherryHQ/cherry-studio)

适用提供 Anthropic / OpenAI 供应商、自定义 API Host 和模型设置的桌面版本。地址自动补全规则以当前安装版本为准。
$help250$),
('clients/claude-code', $help250$安装、认证与 Anthropic 协议配置。$help250$, $help250$## 前置条件
选择支持 Claude Code 的分组。先备份现有配置；下方配置与密钥页使用同一生成器，合并配置时不要覆盖其他设置。

## 安装
官方推荐 Native Install。macOS、Linux 或 WSL：
```sh
curl -fsSL https://claude.ai/install.sh | bash
```
Windows PowerShell：
```powershell
irm https://claude.ai/install.ps1 | iex
```
npm 是备选方式；Node.js 版本要求随 Claude Code 版本变化，请以官方 Setup 页面当前要求为准：
```sh
npm install -g @anthropic-ai/claude-code
```

## 认证与项目配置
在下方选择平台、分组、系统和 Shell，复制 ANTHROPIC_BASE_URL 与 ANTHROPIC_AUTH_TOKEN 配置，或合并到 ~/.claude/settings.json。示例占位密钥必须替换为自己在密钥页生成的密钥。

## 验证与成功判据
```sh
claude --version
claude doctor
claude
```
在项目目录中发送一次简短请求。正常回复并在用量页查到请求才说明实际链路成功。

## 常见错误
检查旧环境变量是否覆盖 settings.json；确认分组支持 Messages 协议。不要把参考站的专属兼容变量一并复制。

## 官方来源
[Claude Code Setup](https://code.claude.com/docs/en/setup) · [Settings](https://code.claude.com/docs/en/settings)

Native Install 不依赖本地 Node.js；npm 的运行时要求以官方 Setup 页面为准。
$help250$),
('clients/cline', $help250$在 VS Code 中通过 OpenAI Compatible 供应商连接本站。$help250$, $help250$## 前置条件
准备 VS Code、本站 API Key 和当前分组的模型 ID。用于编辑文件或执行工具时，模型必须支持对应工具调用能力；基础聊天通过不代表所有 Agent 功能可用。

## 安装
在 VS Code 扩展市场搜索 Cline，核对发布者与 [Cline 官方安装文档](https://docs.cline.bot/getting-started/installing-cline) 后安装。打开侧边栏 Cline 面板，进入设置。

## 认证与项目配置
1. 将 API Provider 设为 OpenAI Compatible，不要选择需要官方账号登录的模式。
2. 填写下面的 Base URL、API Key 与 Model ID。地址已使用本站默认端点并补齐 /v1，不要再追加 /chat/completions。
3. 上下文窗口、最大输出和图片能力按所选模型实际能力设置，不沿用其他模型的能力或价格。
4. 若设置区分 Plan 和 Act 的模型配置，分别检查两种模式的供应商、地址与模型。保存后返回对话。
5. 不需要向项目写入这些凭据，也不要提交包含密钥的设置文件。

## 验证与成功判据
先在 Plan 模式发送一条不修改项目的短消息，收到正常回复并核对本站用量记录。需要 Agent 功能时，再在测试项目中验证一次工具调用，并按客户端提示审核操作。

## 常见错误
- 404：Base URL 应只含一层 /v1；核对模型 ID 的大小写和前缀。
- 普通回答正常但工具调用失败：检查模型能力和工具协议，不能仅修改模型名称冒充支持。
- 切换模式后失败：核对 Plan 与 Act 是否使用不同供应商配置。

## 官方来源
[Cline OpenAI Compatible](https://docs.cline.bot/provider-config/openai-compatible) · [官方安装说明](https://docs.cline.bot/getting-started/installing-cline)

适用提供 OpenAI Compatible 设置的版本，以实际安装界面及官方说明为准。
$help250$),
('clients/codex', $help250$配置用户级 provider、认证与模型目录，完成第一次 Responses 请求。$help250$, $help250$## 前置条件
选择支持 Codex 的分组，并使用绑定到该分组的本站 API Key。Codex 在这里走 Responses 协议，不能把 Chat Completions 地址直接当作 Responses 地址。

先备份已有的用户配置。macOS / Linux 的目录是 `~/.codex/`；Windows 是 `%USERPROFILE%\.codex\`。使用 WSL 时，在 WSL 内安装和配置，不要混用 Windows 用户目录。下面的系统选择应与运行 Codex 的环境一致。

## 安装
在终端执行：
```sh
npm install -g @openai/codex
codex --version
```
macOS 也可使用官方支持的 Homebrew 安装方式：`brew install --cask codex`。选择一种安装方式即可，不要让多个安装位置争用 PATH。先记录 `codex --version` 输出，排错时使用该版本对应的官方文档。

## 认证与项目配置
1. 在 API Keys → 使用密钥 → Codex 获取模型目录，下载 `codex-models.json`。
2. 将目录文件放到 `.codex` 用户目录。下面生成的 `config.toml` 中 `model_catalog_json` 必须指向这个真实存在的文件。
3. 合并生成的 `config.toml`。`model_provider` 必须与 `[model_providers.<名称>]` 中的名称一致；`base_url` 是本站网关地址，`wire_api` 为 `responses`。
4. 用当前密钥模型目录中的实际模型 ID 替换示例 `model`，不要仅凭示例名称判断可用性。
5. 按生成结果选择一种认证方式：有 `env_key` 时，在同一终端设置相应环境变量；生成了 `auth.json` 时，将它放到同一用户目录。不要把密钥字符串填进 `env_key`，它需要的是变量名。

**认证方式不可混用。** `requires_openai_auth = true` 使用 OpenAI 认证路径并忽略 `env_key`。本站的兼容配置可能通过 `auth.json` 提供本站密钥；这不意味着需要购买或登录 ChatGPT。第三方 provider 使用 `env_key` 时，应设置 `requires_openai_auth = false`。界面中的直接 Bearer 模式是本站现有的实验配置，不能把它与环境变量模式同时启用。

Provider 定义保存在用户级配置中，不要只写入项目 `.codex/config.toml`。保留原来的 MCP、权限和其他 provider 设置，避免整文件覆盖。

## 验证与成功判据
进入一个自己的测试项目目录，启动客户端：
```sh
codex
```
在客户端中用 `/status` 检查当前模型和 provider。然后输入：
```text
只回复 OK，不读取或修改任何文件，不调用工具。
```
收到正常回复后，在本站“使用记录”核对时间、模型和分组。这个请求可能产生少量费用。模型目录加载成功与真实推理成功应分别记录。

**Codex WebSocket：** 只有该分组界面提供 WebSocket 选项时才使用对应配置。普通 Codex 成功不代表 WebSocket 已成功；握手失败时先切回普通 Codex 检查认证、协议和网关支持。

## 常见错误
- **找不到 codex 命令：** 重开终端，检查 npm 全局目录是否加入 PATH；避免在 Windows 安装后到 WSL 直接执行。
- **找不到 model_catalog_json：** 检查目录文件是否下载、文件扩展名是否被自动添加为 `.txt`、Windows 路径是否与生成配置一致。
- **401 / 登录冲突：** 检查当前 provider、`env_key` 与 `requires_openai_auth` 的组合，以及旧 `auth.json` 或凭据存储。先备份再调整，不要盲目删除整个 `.codex` 目录。
- **404 / 模型不存在：** 检查模型 ID 与密钥分组；不要在 `base_url` 后手动再追加 `/responses`。
- **429：** 查看本站额度、分组限制和并发；重复重装客户端不能解决额度问题。

## 官方来源
- [Codex CLI](https://developers.openai.com/codex/cli)
- [Advanced Configuration](https://developers.openai.com/codex/config-advanced)
- [Authentication](https://developers.openai.com/codex/auth)

适用支持自定义 provider、Responses 与 `model_catalog_json` 的 Codex CLI；具体版本以安装后的 `codex --version` 为准。
$help250$),
('clients/cursor', $help250$先把 HTTP Compatibility Mode 改为 HTTP/1.1，再填写 OpenAI API Key 和线路 /v1 地址。$help250$, $help250$## 前置条件
准备 OpenAI 渠道分组的本站 API Key 和实际可用模型 ID。Cursor 只在 OpenAI 渠道展示：不同版本与模型可能使用 Responses 或 Chat Completions，不能把其他渠道视为同等兼容。

Cursor 的请求会经过其服务器。本站默认端点必须能被 Cursor 服务器访问；localhost、127.0.0.1 和局域网地址不能用于这种接入。不要为此随意公开本地服务，应由站点管理员配置正式默认端点。

## 安装
从 [Cursor 官网](https://cursor.com/downloads) 下载并安装客户端，打开项目后进入 Cursor Settings → Models。界面位置可能随版本调整。

## 认证与项目配置
1. 打开 Cursor Settings → Network，将 HTTP Compatibility Mode 改为 HTTP/1.1。这是接入步骤，不是只在代理或 VPN 下才做的排查。
2. 进入 Models，展开 API Keys。打开 OpenAI API Key，填入本站密钥。
3. 打开 Override OpenAI Base URL，填入当前线路 URL。地址以 /v1 结尾，不要写成 /chat/completions。设置页没有 Provider，也不要按旧示意去 Add Model。
4. 保存后回到对话，选择当前渠道可用模型并发送一条消息。如果当前版本没有自定义 Base URL，停止使用此接入方式；只把密钥填进官方 OpenAI 地址不会生效。
5. 自定义地址可能影响其他模型请求。保留原设置，切换回官方供应商时检查覆盖地址是否仍启用。

## 验证与成功判据
在 Chat 中发送一条简短纯文本消息，核对模型回复和本站用量记录。再单独验证所需 Agent 功能。自有 API Key 不覆盖 Cursor Tab 补全，不能用补全是否正常判断本站接入结果。

## 常见错误
- 地址验证失败：检查正式端点是否可从公网访问、证书是否有效，以及密钥和模型权限。
- 缺少 messages 或 input：可能是客户端采用的协议与网关处理不匹配。记录模型、Cursor 版本和请求 ID，分别核对 Responses 与 Chat Completions。
- 官方模型也失败：检查全局 Base URL 覆盖是否把其他供应商请求改到了本站。

## 官方来源
[Cursor API Keys](https://cursor.com/help/models-and-usage/api-keys) · [官方论坛：自定义地址兼容问题](https://forum.cursor.com/t/the-custom-override-of-the-openai-base-url-is-unusable/152675)

官方帮助确认自有密钥的聊天模型范围与服务器转发。Base URL 限制参考官方论坛的问题记录，不视为当前所有版本的保证。适用具有 Override OpenAI Base URL 的版本。
$help250$),
('clients/dsh', $help250$通过官方模型设置界面添加本站 Responses 供应商。$help250$, $help250$## 前置条件
准备绑定当前分组的本站 API Key。下方 API 固定为 openai-responses，Base URL 使用 /v1。

## 安装
从 DeepSeek Harness 官方仓库的安装说明安装 dsh。参考命令 `npm install -g @deepseek-ai/dsh`；先核对当前发行版的 Node.js 要求，再运行 `dsh --version` 和 `dsh web`。

## 认证与项目配置
1. 打开 dsh Web 设置，依次选择 Settings → Models → Add model provider → Custom model API。
2. Provider ID 填 site，Base URL 使用下方本站地址，API 选择 openai-responses。不要在 Base URL 末尾加 /responses，也不要改成 Messages。
3. 将 API Key 填到凭据字段，模型 ID 使用上方当前分组模型。可使用 Fetch available models 获取模型列表。
4. 保存后在模型选择器中选中 site 下的模型。模型配置更改在下一次请求生效。
5. 官方新版高级配置位于 $DSH_HOME/profiles/<profile>/cordis.patch.yml。不要把旧版 ~/.dsh/settings.yaml 当作所有版本通用格式。

## 验证与成功判据
选中 site 下的模型后发送一条短消息。确认回复，再到本站用量记录核对模型与请求。模型列表成功不等于推理成功。

## 常见错误
模型列表失败时核对 /v1 地址和密钥权限。Responses 与 Messages 不是可互换的协议。模型能力以本站实际支持为准，不盲填上下文窗口或图片能力。

## 官方来源
[官方来源](https://github.com/deepseek-ai/deepseek-harness/blob/master/docs/user/guide/providers.md)
$help250$),
('clients/gemini-cli', $help250$使用 Gemini API Key 模式接入本站，清理旧认证并验证原生协议调用。$help250$, $help250$## 前置条件
选择支持 Gemini CLI 的分组。本站密钥不是 Google 登录凭据，也不是 Vertex AI 的服务账号。需要客户端支持 Gemini 原生接口和 `GOOGLE_GEMINI_BASE_URL`，不能使用任意 OpenAI 兼容地址替代。

官方安装文档当前要求 Node.js 20.0.0 或更新版本。先在终端检查：
```sh
node --version
npm --version
```

## 安装
执行稳定版安装：
```sh
npm install -g @google/gemini-cli
gemini --version
```
macOS / Linux 也可使用 `brew install gemini-cli`。选择一种安装方式即可。Windows 优先使用 PowerShell；如果用 WSL，则在教程上方选择 Linux，并在 WSL 内安装。

## 认证与项目配置
1. 在准备启动 Gemini 的同一个终端里清理旧认证变量，再执行下面生成的环境变量命令。
2. `GEMINI_API_KEY` 填本站密钥；`GOOGLE_GEMINI_BASE_URL` 保持生成器给出的网关根地址，不要手动拼接 `/models`、`/generateContent` 等路径。
3. `GEMINI_MODEL` 只是示例默认模型，必须改为当前密钥分组实际支持的 Gemini 模型 ID。老版本示例模型不保证仍可调用。
4. 启动 `gemini` 后选择 **Use Gemini API key**。如果之前登录过 Google，可在交互界面用 `/auth` 重新选择认证方式。

**自定义地址只在 API Key 模式生效。** 官方配置页说明 `GOOGLE_GEMINI_BASE_URL` 对应 `gemini-api-key` 模式；Google 登录和 Vertex AI 是另外的路径。远程地址应使用 HTTPS，本机回环地址除外。

上面的清理命令只影响当前终端，不会删除云项目或全局凭据。还要检查项目 `.env` 和用户 `~/.gemini/.env`：CLI 会加载找到的首个环境文件，而不是合并所有文件。旧 `GOOGLE_API_KEY` 或 `GOOGLE_GENAI_USE_VERTEXAI` 设置可能把请求切换到 Vertex 路径。

如需持久保存，可将相同变量写入用户 `.gemini/.env`，使用 `变量名=值` 格式，不带 `export`、`set` 或 `$env:` 前缀。先备份已有文件，不要把含密钥的 `.env` 提交到仓库。

## 验证与成功判据
在设置了变量的终端启动：
```sh
gemini
```
确认选择了 API Key 认证和当前分组支持的模型，然后发送：
```text
只回复 OK，不读取或修改任何文件，不调用工具。
```
在本站用量记录确认该请求确实经过本站。能打开 Google 登录页、看到启动画面或收到 Google 官方错误，均不能证明本站网关已接通。

## 常见错误
- **仍然要求 Google 浏览器登录：** 使用 `/auth` 切换到 Gemini API Key；检查是否仍有旧登录方式被选中。
- **提示 Vertex 项目或区域缺失：** 当前进入了 Vertex 路径，检查 `GOOGLE_GENAI_USE_VERTEXAI` 与旧环境文件。不要为本站密钥临时创建 Google 项目。
- **401：** 检查本站密钥，不要将 Google AI Studio 密钥和本站网关地址混搭。
- **404 / 模型不可用：** 替换示例 `GEMINI_MODEL`，确认分组支持该模型与 Gemini 原生协议。
- **变量设置后仍无效：** 保证设置与启动发生在同一终端；CMD、PowerShell 和 Bash 的赋值语法不能混用。
- **配置地址被拒绝：** 查看是否为远程 HTTP 地址，并按官方要求使用 HTTPS。不要关闭 TLS 证书校验绕过错误。

## 官方来源
- [Installation](https://geminicli.com/docs/get-started/installation/)
- [Authentication](https://geminicli.com/docs/get-started/authentication/)
- [Configuration](https://geminicli.com/docs/reference/configuration/)

适用官方配置文档中支持 `GOOGLE_GEMINI_BASE_URL` 的版本；记录本机 `gemini --version` 以便排查兼容性。
$help250$),
('clients/grok-cli', $help250$使用 xAI 官方 Grok Build，配置本站 Responses 网关与自定义模型。$help250$, $help250$## 前置条件
本文对应 **xAI 官方 `xai-org/grok-build`**，可执行命令为 `grok`。同名的第三方 npm 工具未必支持这里的 TOML 配置，不要混用。

准备绑定 Grok 分组的本站密钥，确认一个可用的文本模型 ID。当前共享生成器的 Grok 文本模型使用 Responses；图片、视频功能属于其他端点，不在本教程的文本验证范围内。

## 安装
macOS / Linux / Git Bash 的官方安装命令：
```sh
curl -fsSL https://x.ai/cli/install.sh | bash
```
Windows PowerShell：
```powershell
irm https://x.ai/cli/install.ps1 | iex
```
安装后重开终端，执行 `grok --version`。上方系统选择会展示适合当前环境的命令。不要把 `npm install -g grok` 当作官方 Grok Build 的安装命令。

## 认证与项目配置
1. 创建用户 `.grok` 目录。macOS / Linux 配置文件为 `~/.grok/config.toml`；Windows 为 `%USERPROFILE%\.grok\config.toml`。
2. 在当前终端设置 `GROK_MODELS_BASE_URL` 与 `XAI_API_KEY`，把占位密钥换成本站密钥。
3. 合并下面生成的 TOML，保留 `[endpoints].models_base_url` 与 `models_list_url` 的配套地址。
4. 为需要的文本模型保留 `[model.<名称>]` 条目。`model` 字段必须是密钥实际可用的模型 ID；`env_key = "XAI_API_KEY"` 引用当前终端变量。
5. 每个本站 Grok 文本模型条目保持 `api_backend = "responses"`。官方默认可能是 Chat Completions，不写该字段会选择不同协议。

设置自定义 `models_base_url` 后，官方文档支持使用 Bearer API Key，不需要运行 `grok login`。不要拿本站密钥尝试官方账号浏览器登录。

**不要覆盖 `cli_chat_proxy_base_url`。** 它是会话服务地址，不是本站推理地址。如果复制过旧模板，请从 `[endpoints]` 中删除这一覆盖行，并检查 `GROK_CLI_CHAT_PROXY_BASE_URL` 是否还指向本站。只删除这一个旧字段，不要删除整个配置文件。

## 验证与成功判据
先确认安装和有效配置：
```sh
grok --version
grok inspect
grok models
```
`grok inspect` 用于本机排查；分享输出前先移除密钥、用户目录等敏感信息。`grok models` 能显示列表不等于推理成功。

然后启动：
```sh
grok
```
在 `/model` 中选择刚配置的本站文本模型，发送：
```text
只回复 OK，不读取或修改任何文件，不调用工具。
```
确认正常回复，再在本站“使用记录”核对模型、分组和时间。不要为了验证文本接入而启动图片、视频或其他额外付费操作。

## 常见错误
- **配置字段无法识别：** 确认安装来源是 `xai-org/grok-build`，记录版本后对照对应版本配置文档。
- **401 / 被引导去登录：** 检查 `models_base_url` 是否生效、`XAI_API_KEY` 是否在当前终端设置，以及 `env_key` 是否拼写一致。
- **请求到了 Chat Completions：** 检查所选模型条目是否明确设置 `api_backend = "responses"`。
- **模型菜单存在但请求 404：** 本地 `[model.*]` 声明不能证明上游支持；核对实际模型 ID 和分组权限。
- **会话代理报错：** 检查旧 `cli_chat_proxy_base_url` 覆盖，不要把会话服务与推理服务地址混用。

## 官方来源
- [Grok Build 官方仓库与安装](https://github.com/xai-org/grok-build)
- [Custom Models](https://github.com/xai-org/grok-build/blob/main/crates/codegen/xai-grok-pager/docs/user-guide/11-custom-models.md)
- [Configuration Reference](https://github.com/xai-org/grok-build/blob/main/crates/codegen/xai-grok-pager/docs/user-guide/26-config-reference.md)

适用官方 Grok Build 的自定义模型 TOML 格式；具体本机版本以 `grok --version` 为准。
$help250$),
('clients/hermes', $help250$按生成配置设置 custom provider；只有出现 api_mode 时才填写。$help250$, $help250$## 前置条件
准备绑定当前分组的本站 API Key。下方配置与密钥页使用同一生成器，YOUR_API_KEY 必须换成自己的密钥。配置前备份 ~/.hermes。

## 安装
macOS / Linux 可参考 `curl -fsSL https://raw.githubusercontent.com/NousResearch/hermes-agent/main/scripts/install.sh | bash`，再执行 `source ~/.zshrc` 和 `hermes --version`。Windows 以 [Hermes 官方文档](https://hermes-agent.nousresearch.com/docs/) 为准。

## 认证与项目配置
1. 将下面的 model 节点合并到 ~/.hermes/config.yaml。provider 为 custom，model.default 为当前模型 ID，model.api_key 填自己的密钥。
2. api_mode 只在生成配置里出现时才写入。OpenAI 为 codex_responses，base_url 用 /v1。Claude、Antigravity、DeepSeek、MiniMax 为 anthropic_messages，base_url 用根地址（Antigravity 加 /antigravity）。其余渠道不要写 api_mode，base_url 用 /v1。
3. 不要把密钥另存到 ~/.hermes/.env，也不要把凭据放进项目仓库。
4. 也可以用生成的终端命令设置同样字段：`hermes config set model.provider`、`model.base_url`、`model.default`；生成结果含 api_mode 时再设置 `model.api_mode`。然后执行 `hermes chat -m` 当前模型 ID。
5. 重启 Hermes 进程，确认当前 provider 是 custom。

## 验证与成功判据
发送一条短消息。收到回复后，到本站用量记录核对模型与请求。

## 常见错误
OPENAI_BASE_URL 属于 openai-api Provider，不代替 custom 的 model.base_url。旧 LLM_MODEL 环境变量不应继续使用。YAML 缩进不正确会导致配置不生效。codex_responses 与 anthropic_messages 不能互换，也不要给没有 api_mode 的渠道补上这个字段。

## 官方来源
[官方来源](https://hermes-agent.nousresearch.com/docs/integrations/providers)
$help250$),
('clients/immersive-translate', $help250$在其他/自定义填写完整接口地址；保存后打开英文网页查看翻译。$help250$, $help250$## 前置条件
准备绑定当前分组的本站 API Key。Claude、Antigravity、DeepSeek、MiniMax 使用 Claude，其余平台使用 OpenAI。

## 安装
从 [沉浸式翻译官网](https://immersivetranslate.com/) 安装官方扩展。打开设置，在「其他/自定义」中选择当前渠道对应的 Claude 1 或 OpenAI 1。

## 认证与项目配置
1. 在当前服务中填写本站 API Key。
2. 展开更多设置，启用自定义 API 地址，填写下方生成的完整 URL：Claude 使用 /v1/messages，OpenAI 使用 /v1/chat/completions。
3. 勾选「输入自定义模型名称」，填入上方模型。AI 智能上下文保持关闭，翻译策略选择「通用」。
4. 保存后打开英文网页查看翻译。不要把扩展弹窗里的旧默认服务当成完成标准。

## 验证与成功判据
网页出现翻译，并且本站用量记录中有对应请求，才算完成。大篇幅网页会消耗额度。“测试服务”成功不能代替整页翻译，也不代表所有 AI 功能可用。

## 常见错误
模型不出现在下拉列表时检查是否已添加并保存自定义模型。429 可降低并发并检查分组额度；不要反复重试整页请求。

## 官方来源
[官方来源](https://immersivetranslate.com/docs/services/openai/)
$help250$),
('clients/kiss-translator', $help250$使用内置 Claude 或 OpenAI 服务，填写完整 URL、Sort Order 0 和模型。$help250$, $help250$## 前置条件
准备绑定当前分组的本站 API Key。Claude、Antigravity、DeepSeek、MiniMax 使用内置 Claude 服务，其余平台使用内置 OpenAI 服务。

## 安装
从 [Kiss Translator 官方仓库](https://github.com/fishjar/kiss-translator) 的发行入口安装扩展或用户脚本，核对来源和版本。

## 认证与项目配置
1. 进入扩展设置的翻译服务，点击 Add，按本页 Provider 选择内置 Claude 或 OpenAI 服务，而不是通用自定义 JSON 接口。
2. 按图填写 URL、Key 和 Model。URL 必须是完整地址：Claude 服务包含 /v1/messages，OpenAI 服务包含 /v1/chat/completions。Sort Order 填 0，翻译风格 formal，Temperature 0，Max Tokens 20480。
3. 如果有自定义模型输入，填准确模型 ID，保留所选内置服务的请求和响应处理逻辑。
4. 保存后刷新页面，选择刚添加的服务，再翻译一小段公开文字。

## 验证与成功判据
翻译结果出现，并且本站用量记录中有对应请求，才算完成。

## 常见错误
请求体没有 messages 通常是误选通用自定义接口。不要把完整请求 URL 填入只收 Base URL 的其他客户端。若改用通用自定义接口，必须按官方 custom-api_v2.md 配置对应协议的请求体和响应 Hook；默认 {text, from, to} 不是 Chat Completions 格式。

## 官方来源
[官方来源](https://github.com/fishjar/kiss-translator/blob/dev/custom-api_v2.md)
$help250$),
('clients/openclaw', $help250$按当前渠道写入 openai-responses、anthropic-messages 或 openai-completions。$help250$, $help250$## 前置条件
准备绑定当前分组的本站 API Key。先记录当前版本，备份 ~/.openclaw/openclaw.json。YOUR_API_KEY 必须换成自己的密钥。

## 安装
从 [OpenClaw 官方文档](https://docs.openclaw.ai/) 完成安装及初始化。不要用示例覆盖已有机器人或渠道配置。

## 认证与项目配置
1. 下方内容合并到 ~/.openclaw/openclaw.json；models.mode 采用 merge，保留现有供应商。
2. models.providers.site.api 随渠道变化：OpenAI 为 openai-responses，baseUrl 用 /v1；Claude、Antigravity、DeepSeek、MiniMax 为 anthropic-messages，baseUrl 用根地址（Antigravity 加 /antigravity）；其余平台为 openai-completions，baseUrl 用 /v1。
3. models 中的 id 是实际模型，agents.defaults.model.primary 使用 site/模型ID。
4. 把占位密钥替换为本站密钥；长期使用优先按官方 SecretRef 或环境变量替换方式管理凭据。
5. 按当前版本配置重载方式让设置生效，在聊天界面选择该模型。不要未经验证添加 compat 或模型能力声明。

## 验证与成功判据
选择该模型后发送一条短消息，再到本站用量记录核对模型与请求。普通对话和工具调用要分别验证。

## 常见错误
如果模型名称包含斜杠，保留完整模型 ID；primary 仍需添加 site/ 前缀。不要混用 openai-responses、anthropic-messages 和 openai-completions；协议必须和当前渠道一致。

## 官方来源
[官方来源](https://docs.openclaw.ai/gateway/config-tools/custom-providers)
$help250$),
('clients/opencode', $help250$安装 OpenCode，配置 provider 与协议，选择模型并验证调用。$help250$, $help250$## 前置条件
准备本站 API Key，并确认绑定分组。进入自己的项目目录，备份已有 `opencode.json` / `opencode.jsonc`。本站生成器使用官方文档中的 `provider`、`npm`、`options.baseURL` 配置结构；不要混入其他版本的 `providers`、`package`、`settings` 字段。

## 安装
安装 Node.js 后，在终端执行：
```sh
npm install -g opencode-ai
opencode --version
```
Windows、macOS 和 Linux 都可以使用该 npm 安装方式。运行时要求以当前 OpenCode 安装文档为准。若出现权限错误，优先修复 Node.js / npm 的用户安装目录，不要直接以管理员身份运行整个编码客户端。

## 认证与项目配置
1. 在项目根目录创建 `opencode.json`，将下面生成的 JSON 合并进去；已有 `opencode.jsonc` 时不要同时维护两份互相冲突的配置。
2. `provider` 下的键就是 provider ID。使用 `/connect → Other` 时，输入与这个键完全相同的 ID，再输入本站密钥。
3. `/connect` 只保存凭据，不会自动填写接口地址或添加自定义模型，仍然需要 JSON 配置。
4. 当前生成器也可能在 `options.apiKey` 中提供占位值。选择直接配置方式时，把 `YOUR_API_KEY` 替换为本站密钥；选择 `/connect` 保存凭据时，删除 JSON 中的占位 `apiKey` 属性，避免它覆盖保存的凭据。注意 JSON 逗号。
5. 用当前分组支持的真实模型 ID 核对 `models` 中的键。顶层 `model` 如存在，格式为 `provider-ID/model-ID`，两部分都必须与配置一致。

**协议由 provider SDK 决定。** Chat Completions 使用 `@ai-sdk/openai-compatible`；Responses 使用 `@ai-sdk/openai`；Anthropic 和 Gemini 原生接口分别使用对应 SDK。保持下方生成器选定的 SDK 与地址配套，不要只替换地址或凭经验追加 `/v1`。

Antigravity 分组可能同时给出 Claude 和 Gemini 两个示例，它们是不同 provider 的方案。首次接入先选一个；需要同时使用时，应合并 `provider` 对象，而不是复制两段独立 JSON 到同一文件。

含密钥的项目配置不要提交到版本库。使用 `/connect` 时，官方说明凭据存放在用户目录的 `~/.local/share/opencode/auth.json`。

## 验证与成功判据
在放置配置文件的项目目录启动：
```sh
opencode
```
需要保存凭据时执行 `/connect`。随后执行 `/models`，选择刚配置的 provider 和模型，再发送：
```text
只回复 OK，不读取或修改任何文件，不调用工具。
```
模型出现在菜单中只说明配置被加载。收到回复并在本站“使用记录”找到对应请求，才表示推理链路成功。

## 常见错误
- **provider 没出现：** 检查运行目录、配置文件名、JSON 语法和 provider ID；重启 OpenCode 重新加载配置。
- **保存密钥后仍然 401：** 检查 `/connect` 的 ID 是否一致，以及 JSON 中是否仍有 `apiKey: "YOUR_API_KEY"` 占位值。
- **列表有模型但调用 404：** `models` 可由本地声明产生，不证明网关提供该模型。核对密钥权限、模型 ID、SDK 协议和 `baseURL`。
- **流式解析失败：** 检查是否把 Chat Completions SDK 接到 Responses 地址，或把 Anthropic / Gemini 原生地址当作 OpenAI 协议使用。
- **设置不生效：** 检查项目、用户与环境变量提供的多份配置，避免修改的文件被更高优先级来源覆盖。

## 官方来源
- [OpenCode 安装](https://opencode.ai/docs/)
- [Providers](https://opencode.ai/docs/providers/)
- [Configuration](https://opencode.ai/docs/config/)

本文匹配 `provider / npm / options` 配置结构。安装版本若使用不同 schema，先按对应版本文档核对。
$help250$),
('clients/pi', $help250$使用用户级 models.json 添加自定义 Responses Provider。$help250$, $help250$## 前置条件
准备绑定 OpenAI 分组的本站 API Key。此客户端只在 OpenAI 渠道展示，API 使用 openai-responses。YOUR_API_KEY 必须换成自己的密钥。

## 安装
参考命令 `npm install -g --ignore-scripts @earendil-works/pi-coding-agent`。先按 [Pi 官网](https://pi.dev/) 核对 Node.js 要求，再运行 `pi --version`。

## 认证与项目配置
1. 创建 ~/.pi/agent 目录，Windows 使用用户主目录下对应路径。
2. 将下方 JSON 合并到 ~/.pi/agent/models.json，保留已有 providers；site 是本教程的自定义供应商 ID。
3. baseUrl 使用本站 /v1 地址，api 采用 openai-responses，models[].id 与上方选择一致。
4. apiKey 示例仅含占位符。若改为环境变量引用，官方当前格式为 "$SITE_API_KEY"，裸字符串 SITE_API_KEY 不会读取环境变量。
5. 启动 Pi，打开 /model，选择 site 下的模型。重新打开 /model 会重载文件。

## 验证与成功判据
选择 site 下的模型后发送一条短消息，再到本站用量记录核对模型与请求。仅列出模型不代表 Responses 推理正常。

## 常见错误
无法选模型时检查是否配置了认证；JSON 中不能带注释或尾逗号。

## 官方来源
[官方来源](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/models.md)
$help250$),
('clients/read-frog', $help250$在 API Providers 按渠道选择服务商，Base URL 使用 /v1，Test Connection 成功后再翻译。$help250$, $help250$## 前置条件
准备绑定当前分组的本站 API Key。Claude、Antigravity、DeepSeek、MiniMax 使用 Anthropic，其余平台使用 OpenAI 兼容服务商。

## 安装
从 [Read Frog 官网](https://www.readfrog.app/) 的官方安装入口安装扩展。不要在非官方页面输入密钥。

## 认证与项目配置
1. 打开扩展选项 → API 服务商，按本页 Provider 选择 Anthropic 或 OpenAI 兼容服务商。
2. 填写下方 Base URL、API Key 和模型 ID，勾选 Enter the name of the custom model，并启用服务商。Base URL 使用生成的 /v1 基础地址，不是完整消息路径。Feature Providers 与 Advanced Options 保持折叠。
3. 点击 Test Connection。提示成功只说明连接检查通过。
4. 在普通网页选择一小段非敏感文字翻译，核对本站用量。不要用私密文档做首次测试。

## 验证与成功判据
翻译结果出现，并且本站用量记录中有对应请求，才算完成。Test Connection 成功不等于翻译可用；翻译成功也不代表需要结构化输出的 AI 操作可用。

## 常见错误
浏览器网络、CORS、代理或扩展权限可能影响请求。

## 官方来源
[官方来源](https://www.readfrog.app/zh/docs/providers/openai-compatible-providers)
$help250$),
('clients/roo-code', $help250$配置 OpenAI Compatible，分别验证对话与原生工具调用。$help250$, $help250$## 前置条件
准备 VS Code、本站 API Key 与支持原生工具调用的模型。Roo Code 的 OpenAI Compatible 供应商要求模型支持相应工具格式；仅支持普通文本的模型不满足 Agent 使用条件。

## 安装
从 [Roo Code 官方仓库](https://github.com/RooCodeInc/Roo-Code) 查看发行及安装说明，核对扩展来源、版本和维护状态后安装。打开 Roo Code 面板，进入设置。不要使用来路不明的同名扩展。

## 认证与项目配置
1. 创建或选择一个 API 配置，供应商选择 OpenAI Compatible。
2. 填写下方 Base URL、API Key 与 Model ID。Base URL 来自本站默认端点，不使用 Roo 官方服务地址，也不使用参考站地址。
3. 根据模型能力设置上下文、最大输出和图片支持；无法确认的能力不要启用。
4. 保存配置，确认当前模式选择的是刚设置的 API 配置。多个配置之间切换时检查地址和模型。

## 验证与成功判据
先发送一个不需要修改文件的短请求，在本站用量记录核对实际推理。再在测试项目中验证一次工具调用，按扩展提示审批。连接测试不等于全部功能可用。

## 常见错误
- 工具调用格式错误：选择支持 OpenAI 原生工具调用的模型，并核对上游兼容性。
- 模型不存在：填写当前密钥所属分组的模型 ID，显示名称不能替代 ID。
- 切换模式后仍用旧地址：核对当前选择的 API 配置，避免修改了未启用的配置。

## 官方来源
[Roo Code OpenAI Compatible](https://docs.roocode.com/providers/openai-compatible) · [官方仓库](https://github.com/RooCodeInc/Roo-Code)

适用提供 OpenAI Compatible 的版本。
$help250$),
('clients/sillytavern', $help250$在酒馆渠道用 Chat Completion 的自定义 OpenAI 兼容端点接入。$help250$, $help250$## 前置条件
准备本站 API Key，并选择上方分组。酒馆不是新的网关平台，客户端只发送 OpenAI Chat Completions。地址使用当前线路的 /v1，不要自行追加 /chat/completions，密钥不要加 Bearer。

可以对接：
- 分组下拉中的分组。这里列出未开启「仅 Claude Code」的分组，包含 OpenAI、Grok、Claude、Gemini、Antigravity，以及 Kimi、智谱、DeepSeek、MiniMax 等隐藏渠道下的分组。
- OpenAI、Grok、Kimi、智谱、DeepSeek、MiniMax 由 OpenAI Chat Completions 网关转发。Claude、Gemini、Antigravity 由本站转换后转发。合成分组只接受该分组能够解析的模型。

不能对接：
- 开启「仅 Claude Code」的分组不会出现在本渠道。这类分组只接受 Claude Code 的 /v1/messages。
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
发送一句短对话。SillyTavern 收到回复，并且本站用量记录中的模型与请求一致，才算完成。模型列表、连接状态和测试消息按钮本身不等于角色对话已经可用。

## 常见错误
地址多写了 /chat/completions 时，实际请求会变成 /v1/chat/completions/chat/completions。401 时检查密钥是否误加 Bearer，以及分组是否仅限 Claude Code。模型列表为空时手填模型 ID，不要改用文本补全。

## 官方来源
[官方来源](https://docs.sillytavern.app/usage/api-connections/openai/)

适用当前提供自定义 Chat Completion 端点的版本。
$help250$),
('clients/tavernai', $help250$在酒馆渠道用 TavernAI 2 的 Custom Chat Completions 接入。$help250$, $help250$## 前置条件
准备本站 API Key，并选择上方分组。这里的 TavernAI 指 TavernAI 2，不是旧版 TavernAI 的反向代理。客户端使用 Chat Completions，地址填当前线路的 /v1。

可以对接的分组与 SillyTavern 相同：未开启「仅 Claude Code」的文字对话分组。OpenAI、Grok、Kimi、智谱、DeepSeek、MiniMax 走 OpenAI Chat Completions 网关；Claude、Gemini、Antigravity 由本站转换；合成分组只接受它能解析的模型。

不能对接：仅 Claude Code 分组、图片与 Embedding 等非对话模型、Text Completions，以及 Media Tools 的生图提供商。不要把内置 OpenAI 提供商的密钥页当成完成配置。

## 安装
从 [TavernAI 官方仓库](https://github.com/TavernAI/TavernAI) 下载当前 TavernAI 2 发行包。旧版说明见 TavernAI-v1，不要按旧版反向代理填写。[快速开始](https://tavernai.net/docs/quick-start/) 只演示打开提供商设置，字段以当前版本的 Custom Chat Completions 为准。

## 认证与项目配置
1. 点击顶部插头图标，打开「AI 提供商连接」。不要停在 Provider 为 OpenAI 的内置页。
2. 基础提供商选择 Custom。模型端点选择 Chat Completions，不要选 Text Completions。当前版本没有模型端点下拉时，保持 Custom 即可。
3. API 地址填下方 /v1，不要写成 /chat/completions。若界面出现「使用直接 API 地址（忽略所选端点）」，打开它，避免地址被所选端点模板改写。
4. API 密钥填本站密钥，不要加 Bearer。模型 ID 填上方模型；也可以先点提供商模型的 Load，失败后再手填。
5. 点击连接。连接检查失败但实际消息成功时，才打开「跳过连接检查」。官方快速开始截图在演示内置 OpenAI 时勾选了它，本站先不要勾。
6. Media Tools、ComfyUI 和本地 GGUF 不在本教程范围内。

## 验证与成功判据
连接成功后发送一条短文字消息。TavernAI 收到回复，并且本站用量记录中的模型与请求一致，才算完成。Load 成功或状态显示 Connected 都不等于对话可用。

## 常见错误
选成内置 OpenAI 后无法改写本站地址。API 地址多写 /chat/completions，或没打开直接地址，都会打到错误路径。Text Completions 会请求 /v1/completions，本教程不使用该接口。

## 官方来源
[官方来源](https://github.com/TavernAI/TavernAI)

适用当前提供 Custom provider 与 Chat Completions 的 TavernAI 2。
$help250$),
('clients/trae', $help250$按渠道选择 API 格式，线路地址不要带协议路径，完整 URL 保持关闭。$help250$, $help250$## 前置条件
准备绑定当前分组的本站 API Key。API 格式随渠道变化，地址使用密钥页默认端点。YOUR_API_KEY 必须换成自己的密钥。

## 安装
从 [TRAE 官网](https://www.trae.ai/) 获取 TRAE SOLO；确认所用版本和账号提供添加自定义模型功能。不要把 TraeCode CLI 的 trae_cli.yaml 用作 SOLO 桌面版配置。

## 认证与项目配置
1. 打开模型设置 → 添加模型 → 自定义配置。
2. API 格式按下方字段选择：Claude、Antigravity、DeepSeek、MiniMax 使用 Anthropic Messages，其余渠道采用 OpenAI Chat Completions。
3. 保持“完整 URL”关闭。自定义请求地址只填线路 URL，不要以斜杠结尾，也不要自带 /v1/messages 或 /chat/completions；客户端会把对应路径补到末尾。
4. 填写模型 ID 与 API 密钥。只有确认所选模型支持图片时才开启多模态。
5. 提交后回到对话界面，选择刚添加的模型。当前版本没有此功能时，不要将本站密钥填入官方服务商登录栏。

## 验证与成功判据
选择刚添加的模型，发送一条短消息。确认回复后，到本站用量记录核对模型与请求。

## 常见错误
404 常见于完整 URL 被打开，或地址里已经写了客户端还会追加的路径。SOLO、IDE 和 CLI 不是同一种配置文件；核对安装版本。

## 官方来源
[官方来源](https://www.trae.ai/)

适用提供上述自定义模型配置的 TRAE SOLO 版本。
$help250$),
('clients/workbuddy', $help250$通过自定义模型连接本站：接口地址包含 /chat/completions，自定义协议保持关闭。$help250$, $help250$## 前置条件
准备绑定当前分组的本站 API Key。接口地址使用下方生成的完整 /v1/chat/completions，自定义协议保持关闭。

## 安装
从 [WorkBuddy 官网](https://www.workbuddy.ai/) 下载并安装客户端，进入 Settings → Model → 添加模型 → Custom。

## 认证与项目配置
1. 使用下方 URL、API Key 和 Model 字段。
2. 提供商选择「自定义 / Custom」。接口地址填写完整 /v1/chat/completions，自定义协议保持关闭。高级配置里工具调用、图片输入和思考模式按客户端默认勾选；模型不支持图片或工具时再关闭。默认思考强度选自动，输入和输出长度不要猜测。
3. 不要打开自定义协议后再把 /chat/completions 写进地址，否则可能重复拼接。
4. 保存并选择新模型。旧 ~/.codebuddy/models.json 可能仍兼容，但优先使用官方推荐的图形界面。

## 验证与成功判据
发送一条不需要工具的短消息。收到回复后，到本站用量记录核对模型与请求。连接测试不等于工具或图片可用。

## 常见错误
重复 /chat/completions 通常是自定义协议被打开，同时又填写了已经包含该路径的地址。核对当前使用的模型配置，不要误改了未启用的条目。

## 官方来源
[官方来源](https://www.workbuddy.ai/docs/workbuddy/From-Beginner-to-Expert-Guide/Function-Description/Model)
$help250$),
('clients/zcode', $help250$按 ZCode 模型设置窗口添加供应商和模型，地址使用线路 /v1。$help250$, $help250$## 前置条件
准备绑定当前分组的本站 API Key。API Base URL 使用当前线路的 /v1。

## 安装
从 [ZCode 官网](https://zcode.z.ai/) 安装客户端。进入模型选择器 → 管理模型 → 模型设置 → 添加供应商。

## 认证与项目配置
1. 打开 ZCode 设置并进入「模型设置」，点击「+ 添加供应商」。供应商名称可自定义，当前界面没有协议下拉。
2. API Base URL 填当前线路的 /v1，不要写成 /chat/completions。API Key 填自己的密钥，然后保存供应商。
3. 在该供应商的模型列表中点击「添加模型」。Model ID 填当前选择的模型。上下文窗口由客户端按模型自动带出，没有带出时不要手填猜测值。
4. 返回工作区，点击输入框右下角的「选择模型」，展开本站供应商并选择刚添加的模型。
5. 发送一条消息。看到 ZCode 正常回复即表示配置完成。不要假设随意添加的 options 会被发送给服务器。

## 验证与成功判据
工作区里选中该模型并收到正常回复，且本站用量记录能对上，才算完成。模型出现在列表中不等于请求已经发出。

## 常见错误
模型 ID 不存在时先核对分组权限。ZCode 使用 OpenAI 兼容的 /v1 地址，不要改成 /v1/messages。

## 官方来源
[官方来源](https://zcode.z.ai/cn/docs/configuration)
$help250$),
('faq', $help250$认证、计费、模型与连接问题。$help250$, $help250$## 配置未生效
确认配置文件路径及 Shell 类型，检查旧环境变量，然后重新启动客户端。

## 模型不可用
核对密钥绑定分组、可用模型名称和接口协议。不要用全站模型列表推断单个密钥权限。

## 认证失败
检查密钥是否有效、禁用或过期。区分官方账号登录与第三方 API Key。

## 超时与限流
先查看用量记录和渠道状态，核对并发与额度；模型列表正常也可能出现推理超时。

## 余额与订阅
以当前分组、订阅和订单页面为准，不套用其他站点的折扣、倍率、退款或套餐说明。

## 成功判据
真实调用收到正常回复，并在本项目用量记录中找到对应请求。请求可能产生费用，页面不会自动发送推理请求。客户端特定错误见对应教程中的官方链接。
$help250$),
('quick-start', $help250$完成第一次客户端调用。$help250$, $help250$## 前置条件
登录并确认有可用额度与分组，在 API Keys 页面创建密钥。密钥所属分组决定可用模型与协议。

## 操作步骤
1. 打开密钥页的“使用教程”，自动带入平台与分组。
2. 选择客户端、系统和 Shell，复制配置示例。
3. 在密钥页的“使用密钥”中生成含真实凭据的配置，只在自己的设备使用。
4. 关闭并重新打开终端或客户端。

## 验证与成功判据
先查询模型列表，再发送一次最小推理请求。列表成功只证明认证与列表路由；客户端收到正常模型回复且用量页出现对应请求才算完成。实际推理可能产生费用。

## 常见错误
401：核对密钥状态。404：核对协议路径与模型名。429：核对分组额度和并发限制。超时：查看请求日志和渠道状态。
$help250$)
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
