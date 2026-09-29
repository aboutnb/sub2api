-- Refresh only untouched starter publications. Operator drafts/publications stay intact.
WITH guides(slug, summary, content) AS (
 VALUES
 ('clients/codex', '配置用户级 provider、认证与模型目录，完成第一次 Responses 请求。', $codex$
## 前置条件
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
1. 在 **API Keys → 使用密钥 → Codex** 获取模型目录，下载 `codex-models.json`。
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

## 官方来源与验证记录
- [Codex CLI](https://developers.openai.com/codex/cli)
- [Advanced Configuration](https://developers.openai.com/codex/config-advanced)
- [Authentication](https://developers.openai.com/codex/auth)

资料核对：2026-09-22。适用支持自定义 provider、Responses 与 `model_catalog_json` 的 Codex CLI；具体版本以安装后的 `codex --version` 为准。文档和配置结构已核对，尚未标记本站各分组真实推理通过。
$codex$),
 ('clients/opencode', '安装 OpenCode，配置 provider 与协议，选择模型并验证调用。', $opencode$
## 前置条件
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

## 官方来源与验证记录
- [OpenCode 安装](https://opencode.ai/docs/)
- [Providers](https://opencode.ai/docs/providers/)
- [Configuration](https://opencode.ai/docs/config/)

资料核对：2026-09-22。本文匹配 `provider / npm / options` 配置结构；安装版本若使用不同 schema，请先按对应版本文档核对。尚未记录本站各分组真实推理验证。
$opencode$),
 ('clients/gemini-cli', '使用 Gemini API Key 模式接入本站，清理旧认证并验证原生协议调用。', $gemini$
## 前置条件
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

## 官方来源与验证记录
- [Installation](https://geminicli.com/docs/get-started/installation/)
- [Authentication](https://geminicli.com/docs/get-started/authentication/)
- [Configuration](https://geminicli.com/docs/reference/configuration/)

资料核对：2026-09-22。适用官方配置文档中支持 `GOOGLE_GEMINI_BASE_URL` 的版本；记录本机 `gemini --version` 以便排查兼容性。尚未记录本站各分组真实推理验证。
$gemini$),
 ('clients/grok-cli', '使用 xAI 官方 Grok Build，配置本站 Responses 网关与自定义模型。', $grok$
## 前置条件
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

## 官方来源与验证记录
- [Grok Build 官方仓库与安装](https://github.com/xai-org/grok-build)
- [Custom Models](https://github.com/xai-org/grok-build/blob/main/crates/codegen/xai-grok-pager/docs/user-guide/11-custom-models.md)
- [Configuration Reference](https://github.com/xai-org/grok-build/blob/main/crates/codegen/xai-grok-pager/docs/user-guide/26-config-reference.md)

资料核对：2026-09-22。适用官方 Grok Build 的自定义模型 TOML 格式；具体本机版本以 `grok --version` 为准。安装来源与字段已核对，本站真实推理尚未实测。
$grok$)
)
UPDATE help_documents AS d
SET summary = guides.summary,
    content_markdown = btrim(guides.content),
    publication_history = d.publication_history || jsonb_build_array(d.published_snapshot),
    published_snapshot = d.published_snapshot || jsonb_build_object(
      'summary', guides.summary, 'content_markdown', btrim(guides.content),
      'version', d.version + 1, 'published_at', NOW()),
    version = d.version + 1,
    published_at = NOW(),
    updated_at = NOW()
FROM guides
WHERE d.slug = guides.slug
  AND d.version = 1
  AND d.updated_by IS NULL
  AND d.status = 'published'
  AND d.published_snapshot IS NOT NULL
  AND d.publication_history = '[]'::jsonb
  AND d.content_markdown = d.published_snapshot->>'content_markdown';
