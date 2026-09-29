CREATE TABLE IF NOT EXISTS help_documents (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(160) NOT NULL UNIQUE,
    title VARCHAR(240) NOT NULL,
    category VARCHAR(60) NOT NULL,
    summary VARCHAR(1000) NOT NULL DEFAULT '',
    content_markdown TEXT NOT NULL,
    selector_schema JSONB NOT NULL DEFAULT '{}'::jsonb,
    published_snapshot JSONB NULL,
    publication_history JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    version INTEGER NOT NULL DEFAULT 1,
    published_at TIMESTAMPTZ NULL,
    updated_by BIGINT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS help_documents_slug_idx ON help_documents (slug);
CREATE INDEX IF NOT EXISTS help_documents_category_status_idx ON help_documents (category, status);
CREATE INDEX IF NOT EXISTS help_documents_updated_at_idx ON help_documents (updated_at);

WITH seeds(slug, title, category, summary, content_markdown, selector_schema) AS (
 VALUES
  ('quick-start', '快速开始', 'quick-start', '完成第一次客户端调用。', '## 前置条件
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

## 维护信息
资料核对日期：2026-09-22。客户端版本与网关兼容性需按各篇教程验证。本初始教程未标记为实测通过。', '{}'::jsonb),
  ('clients/claude-code', 'Claude Code 接入', 'claude-code', '安装、认证与 Anthropic 协议配置。', '## 前置条件
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

## 官方来源与验证记录
[Claude Code Setup](https://code.claude.com/docs/en/setup) · [Settings](https://code.claude.com/docs/en/settings)
资料核对：2026-09-22。Native Install 不依赖本地 Node.js；npm 的运行时要求以官方 Setup 页面为准。本项目端到端验证尚未记录。', '{"clients":["claude"]}'::jsonb),
  ('clients/codex', 'Codex 接入', 'codex', '区分官方认证与第三方 provider 认证。', '## 前置条件
使用支持 Responses 的分组。备份 ~/.codex/config.toml，按当前客户端版本核对配置字段。

## 安装
```sh
npm install -g @openai/codex
codex --version
```

## 认证与项目配置
选择下方系统与认证模式，合并生成的 config.toml。模型目录文件需在密钥页获取与下载，不要引用不存在的目录文件。
requires_openai_auth = true 使用 OpenAI 认证路径（项目的兼容模式可能生成 auth.json）。使用 env_key 的第三方 provider 必须设 requires_openai_auth = false；env_key 是环境变量名称，不能填写密钥本身。
当前项目 OpenAI 分组另有实验性直接 Bearer 模式，与 env_key 模式不同；按密钥页生成结果选择，不要混合两种认证方式。

## 验证与成功判据
```sh
codex
```
启动后核对所选 provider 和模型，发送一次最小请求。模型列表可用不代表 Responses 推理成功。

## 常见错误
认证冲突时检查旧登录状态和凭据存储；更换 provider 后完整退出再启动。模型目录缺失时在密钥页下载相应目录。

## 官方来源与验证记录
[Advanced Configuration](https://developers.openai.com/codex/config-advanced) · [Authentication](https://developers.openai.com/codex/auth)
资料核对：2026-09-22。适用具有自定义 Responses provider 能力的版本；具体实测版本尚未记录。', '{"clients":["codex","codex-ws"],"protocols":["responses"]}'::jsonb),
  ('clients/opencode', 'OpenCode 接入', 'opencode', '配置提供商、协议与模型。', '## 前置条件
准备可用分组与 API Key，备份现有 opencode.json。

## 安装
按 [OpenCode 官方介绍](https://opencode.ai/docs/) 选择当前系统的安装方式，然后运行：
```sh
opencode --version
```

## 认证与项目配置
使用下方生成的 opencode.json。在客户端通过 /connect → Other 添加凭据；如果使用生成器的 apiKey 字段，不要提交该配置到公开仓库。
Chat Completions 使用 @ai-sdk/openai-compatible；Responses 使用 @ai-sdk/openai。原生 Anthropic 与 Gemini 分别使用对应 SDK，不能仅替换地址而保留错误协议。

## 验证与成功判据
启动 opencode，执行 /models 选择当前分组支持的模型，再发送一次最小请求并检查用量记录。

## 常见错误
模型列表为空时核对 provider ID 与配置中的 models 键。404 时核对 baseURL 和协议，不要重复拼接 /v1。

## 官方来源与验证记录
[OpenCode Providers](https://opencode.ai/docs/providers/)
资料核对：2026-09-22。适用支持自定义 provider 的版本；具体端到端实测版本尚未记录。', '{"clients":["opencode"]}'::jsonb),
  ('clients/gemini-cli', 'Gemini CLI 接入', 'gemini-cli', '区分 Gemini API Key 和 Vertex AI。', '## 前置条件
选择 Gemini CLI 兼容分组，并确认客户端支持自定义 Gemini 地址。

## 安装
按 [Gemini CLI 官方文档](https://geminicli.com/docs/get-started/) 安装，记录版本：
```sh
gemini --version
```

## 认证与项目配置
使用下方共享生成器的配置。官方 Gemini API Key 流程使用 GEMINI_API_KEY；Vertex AI 的 GOOGLE_API_KEY、ADC、项目和区域设置是另一套认证路径。旧 GOOGLE_API_KEY 或 Vertex 配置可能与当前模式冲突。
本项目生成的 GOOGLE_GEMINI_BASE_URL 属于网关接入配置；官方认证页并不能单独证明所有版本都支持这个覆盖变量。

## 验证与成功判据
```sh
gemini
```
选择 API key 模式，确认请求实际到达本项目并获得正常回复。只看到登录成功不算网关验证成功。

## 常见错误
仍连接官方服务时检查地址覆盖是否被当前版本识别；不要将第三方密钥用于 Google 登录或 Vertex ADC。

## 官方来源与验证记录
[Gemini CLI Authentication](https://geminicli.com/docs/get-started/authentication/)
资料核对：2026-09-22。自定义地址兼容版本及端到端实测尚未记录。', '{"clients":["gemini"]}'::jsonb),
  ('clients/grok-cli', 'Grok CLI 接入', 'grok-cli', '使用项目现有 Grok CLI 配置格式。', '## 前置条件
使用本项目支持的 Grok CLI 实现与 Grok 分组。名称相同的第三方 CLI 不一定支持相同配置文件。

## 安装
按所用 Grok CLI 实现的官方 README 安装，不要仅根据“grok”名称安装来源不明的软件。客户端版本应记录在运营维护说明中。

## 认证与项目配置
在下方选择 Grok 分组和系统，复制环境变量与 ~/.grok/config.toml 配置。使用项目生成的模型列表与推理地址；不要混用 xAI 官方登录与第三方密钥。

## 验证与成功判据
```sh
grok --version
grok
```
先验证模型列表，再发送一次最小推理请求，并核对本项目用量记录。

## 常见错误
字段不识别说明所安装的实现或版本与模板不匹配，不应继续把模板当作兼容证明。

## 官方来源与验证记录
[xAI API 文档](https://docs.x.ai/) 仅作为协议参考，不作为具体 Grok CLI 安装来源。
项目模板盘点日期：2026-09-22。CLI 官方发行来源、适用版本及端到端验证尚待记录。', '{"clients":["grok"],"platforms":["grok"]}'::jsonb),
  ('api', 'API 调用', 'api', '区分认证测试与真实推理。', '## 前置条件
从密钥页复制接口根地址、密钥和可用模型。下面命令为 macOS/Linux 示例，先在本机设置 API_BASE_URL、API_KEY 和 MODEL 环境变量，禁止将真实值发到帮助文档。

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

## 官方来源与验证记录
[OpenAI Responses](https://platform.openai.com/docs/api-reference/responses)
资料核对：2026-09-22。各分组协议能力以实际配置和验证为准。', '{}'::jsonb),
  ('faq', '常见问题', 'faq', '认证、计费、模型与连接问题。', '## 配置未生效
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
真实调用收到正常回复，并在本项目用量记录中找到对应请求。请求可能产生费用，页面不会自动发送推理请求。

## 来源与维护记录
来源：本项目密钥、用量和分组配置流程。核对日期：2026-09-22。遇到客户端特定错误请参照相应教程官方链接。', '{}'::jsonb)
)
INSERT INTO help_documents (slug, title, category, summary, content_markdown, selector_schema, status, version, published_snapshot, published_at)
SELECT slug, title, category, summary, content_markdown, selector_schema, 'published', 1,
 jsonb_build_object('title', title, 'category', category, 'summary', summary, 'content_markdown', content_markdown, 'selector_schema', selector_schema, 'version', 1, 'published_at', NOW()), NOW()
FROM seeds
ON CONFLICT (slug) DO NOTHING;
