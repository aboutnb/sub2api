# FlowAI 分支变更台账

> 适用分支：`sub2api-flowai`
>
> 这不是上游项目的 release note，而是 FlowAI 分支的业务上下文和所有权记录。
> 上游合并、功能调整和正式发布都必须以本台账与
> [FLOWAI_BRANCH_CONTRACT.md](./FLOWAI_BRANCH_CONTRACT.md) 为准。

## 1. 台账目的

FlowAI 分支长期保留了一组与上游 `main` 不同的产品功能、调度语义、部署方式和
数据迁移。单纯执行 `git merge upstream/main` 或按文件选择一方，可能会出现以下
问题：

- 上游同名文件覆盖 FlowAI 的业务逻辑或 i18n key；
- 账号优先级方向被改回数值越大优先；
- 已执行迁移被重写，触发生产 checksum 错误；
- 邮件、支付或签到状态被当作普通重试而重复发送/重复结算；
- 只更新应用镜像，遗漏 Mihomo、Caddy、locale 或持久化配置；
- 服务器重新构建导致发布窗口不可控。

因此每次合并上游或发布前，按以下顺序执行：

1. 记录当前分支、HEAD、上游引用、应用版本和工作树状态。
2. 查看 `git diff upstream/main...HEAD`、上游新增迁移和本台账的受保护路径。
3. 对每个冲突按“上游新增行为 / FlowAI 保留行为 / 有意改变的契约”分类，禁止整树覆盖。
4. 如果语义发生变化，先更新契约、测试和本台账，再合并代码或生成发布镜像。
5. 执行 `make check-flowai-contract` 和发布清单中的目标测试；证据写入发布记录。

自动检查会核对本台账中的提交索引。提交信息带 `[flowai-governance]` 的文档维护提交
可以不重复登记，但该提交只能修改 FlowAI 治理文档、部署说明、检查脚本、Makefile、
`.gitignore` 和对应工作流；任何代码、迁移或合并提交都必须在后续发布前补入索引。提交
索引和迁移索引之间的 HTML 标记是机器检查边界，不要删除、改名或把其他说明放进标记区。

## 1.1 每次上游合并的固定流程

1. `git fetch upstream main` 后，记录 `upstream/main` 的完整 hash；不要更新或切换
   本地 `main`，然后执行 `make review-flowai-upstream`。若有待合入提交，必须先完成人工
   核对，再用 `FLOWAI_UPSTREAM_REVIEW_ACK=<完整 hash>` 重跑并取得通过结果。
2. 合并前查看 `git diff --name-status upstream/main...HEAD`，并单独查看上游新增迁移、
   调度代码、i18n 聚合入口、部署 compose 和锁文件。
3. 只使用 `git merge --no-ff upstream/main`。冲突时按第 6 节矩阵逐项解决，禁止用
   上游文件整文件覆盖 FlowAI 实现。
4. 合并完成后，把合并提交 hash、每个冲突区域的保留结论和验证方式写入本台账；新增
   功能提交同时登记行为、受保护路径、测试和迁移影响。
5. 先运行 `make check-flowai-contract`，再运行目标测试；发布候选最后运行
   `make check-flowai-contract-strict`。任一门禁失败都不能生成镜像或连接 23 服务器。

未来新增功能条目至少包含以下信息：

```text
提交 hash / 日期：
功能和用户可见行为：
必须保留的实现路径与 i18n key：
数据库迁移及是否可回滚：
正向测试、回归测试和未验证项：
与上游同名文件的冲突处理：
```

## 2. 当前快照

| 项目 | 值 |
| --- | --- |
| 发布分支 | `sub2api-flowai` |
| 本次核对日期 | 2026-08-30（Asia/Shanghai） |
| 业务代码基线 HEAD（本次合并前） | `592d28f277b9824617aa657cae755a103909c7bd` |
| 本次核对 HEAD（文档更新前） | `52d8138e27d8e64a377d414d5173e25d9ef9b5ff` |
| 最后已审并合入的上游基线 | `upstream/main` = `b5827cfd54d58c248a9480b800444d0b40f0c6ea`（0.1.183） |
| 当前抓取但尚未合入的上游 | 无（已合入） |
| 应用版本 | `0.1.183` |
| 相对上游的非合并提交 | 88 个功能/修复提交、8 个治理提交；治理提交按受限规则由脚本动态豁免 |
| 相对上游的文件差异 | 440 个文件，约 48481 行新增、1633 行删除 |
| 发布镜像 | `ghcr.io/aboutnb/sub2api:sub2api-flowai-<sha12>` |
| 生产发布目标 | 23 服务器，使用预构建镜像 |

快照值会变化，当前行为契约不会因上游版本号变化而自动变化。下一次发布必须重新
记录 HEAD、上游基线、版本、镜像 digest 和实际验收结果。

### 2.1 本次上游合并审阅（已完成）

2026-08-29 刷新并审阅 `upstream/main` 完整提交
`b5827cfd54d58c248a9480b800444d0b40f0c6ea`（0.1.183）。本次上游相对已审基线新增
40 个提交，随后以 `ea2096d33360f45ddd4ec4142e457c4d648ee493` 合入
`sub2api-flowai`。

- 上游的 40 个提交已完整进入合并提交；`main` 未修改，且本次没有新增或修改 SQL 迁移。
- 上游新增的 DeepSeek 定价、Anthropic Fable 阈值、OpenAI/Spark 流式故障转移、模型定价、
  上游倍率探测、智谱团队版配置及充值币种显示已合入；不改变 FlowAI 的账号调度、并发、
  GM/EasyPay、签到、邮件和独立部署边界。
- `PaymentView.vue` 是唯一内容冲突：保留 FlowAI 当前充值/订阅布局和 GM/EasyPay 支付选择，
  仅吸收上游按所选币种渲染充值倍率的参数；对应中英文 key 和测试均保留。
- 账号调度仍为 priority 升序，即 1 最高优先级；受保护的中英文 i18n 聚合入口和账号文案
  未被上游覆盖或删除。后续以契约检查和发布测试结果作为放行证据。
- 为对齐当前调度契约，提交 `52d8138e2` 修正 schedulable projection 集成测试中与
  priority 升序相反的断言；目标用例已在本地 PostgreSQL/Redis 测试容器中通过。

合并前预审确认命令：

```bash
FLOWAI_UPSTREAM_REVIEW_ACK=b5827cfd54d58c248a9480b800444d0b40f0c6ea \
  make review-flowai-upstream
```

确认值必须与当时的 `upstream/main` 完整 hash 一致；上游再次变化时必须重新审阅，不能
复用旧确认值。

## 3. 当前行为总表

| 功能域 | FlowAI 当前约定 | 主要实现入口 | 主要回归测试 |
| --- | --- | --- | --- |
| 账号调度 | 账号 `priority` 为 1 时最高，数值越小越优先；相同优先级再比较负载、LRU、OAuth 等 | `backend/internal/service/gateway_scheduling.go`、`openai_account_scheduler.go`、`openai_gateway_scheduling.go`、`gemini_messages_compat_service.go`、`backend/internal/repository/account_repo.go` | `gateway_account_selection_test.go`、`scheduler_layered_filter_test.go`、OpenAI scheduler 测试、`account_repo_sort_integration_test.go` |
| 用户/账号并发 | `-1` 立即拒绝，`0` 不限，正数为上限；小于 `-1` 不得写入 | `backend/internal/service/concurrency_service.go`、`backend/internal/handler/gateway_helper.go`、设置/用户管理路径 | `concurrency_service_test.go`、`gateway_helper_hotpath_test.go` |
| 风险注册赠送 | 授权判断前余额为 0、并发为 `-1`；一次性授权成功后才写入实际赠送值 | `backend/internal/service/auth_service.go`、`signup_risk_context.go`、`signup_risk_grant_repo.go` | `signup_risk_grant_test.go`、签到安全测试 |
| Project Mihomo | 支持 URL/静态源、多源、兼容请求头、节点测速、筛选、自动路由、多 listener 和账号池分配；状态持久化 | `backend/internal/service/project_mihomo_service.go`、admin handler、`deploy/docker-compose.preview.yml` | `project_mihomo_service_test.go`、admin handler 测试 |
| 账号导入/批量测试 | 导入可指定分组/代理池，返回新账号集合；批量测试需人工点击开始，不能自动误发请求 | `backend/internal/handler/admin/account_data.go`、`BatchAccountTestModal.vue`、`AccountsView.vue` | `data-import.spec.ts`、`BatchAccountTestModal.spec.ts`、账号测试 i18n 测试 |
| USDT/发票 | 生产使用独立 GM 服务并通过 EasyPay 自定义方式接入；BEpusdt 专用应用代码已移除，历史迁移保持不变；XZNOAuth 发票状态和 PDF 流程保持不变 | `backend/internal/payment/provider/easypay.go`、EasyPay provider 配置、`backend/migrations/206_add_usdt_payments.sql`、`invoice_service.go` | EasyPay custom method/provider 测试、invoice service 测试、支付 API 测试 |
| 签到/奖励 | 每用户每业务日最多结算一次；普通/幸运模式、概率、倍率/固定金额、阶梯、精度和未充值策略均受事务与风控约束 | `backend/internal/service/checkin_service.go`、`checkin_record_repo.go` | `checkin_service_test.go`、精度/安全/handler 测试 |
| 邮件广播 | 受众和模板快照持久化；`FOR UPDATE SKIP LOCKED` 领取；发送结果不确定时不自动重试 | `email_broadcast_repo.go`、`email_broadcast_service.go` | repository integration tests、service tests |
| 注册/访问安全 | 注册 challenge、邮箱策略、IP 封禁、公开 POST 发布密钥、上游错误脱敏和 Cloudflare 保护必须保留 | auth handlers/services、middleware、`upstream_error_sanitize.go` | auth/middleware/service 安全测试 |
| i18n | 中文和英文模块、聚合入口、key 级合并都必须保留；账号优先级文案必须与调度方向一致 | `frontend/src/i18n/index.ts`、`frontend/src/i18n/locales/{zh,en}/` | locale compile/collision/default/account/checkin 测试 |
| 站点配置 | 社区群、订阅页开关、充值/订阅费用策略、主题、品牌 logo、公告和法律文档是分支功能 | settings service、`App.vue`、`useThemeMode.ts`、相关 views | Settings、router、branding、announcement 测试 |
| 发布 | GitHub Actions 构建 GHCR 镜像；23 服务器只 pull 并重建应用服务，不在生产机打包 | `.github/workflows/preview-image.yml`、`deploy/deploy-preview-image.sh` | workflow/compose 渲染、health smoke test |

## 4. 功能与提交台账

### 4.1 账号调度优先级

- 最终规则是**账号优先级 1 最高，数值越小越优先**。当前实现同时覆盖普通 Gateway、分层
  调度、OpenAI 高级 scheduler、传统 OpenAI、Gemini、图片账号排序和数据库查询。
- `account_groups.priority ASC` 是分组成员顺序；`accounts.priority ASC` 是账号自身
  调度顺序。两者不能互换。
- 错误透传规则的 `priority` 仍是数值越小越先匹配，位于
  `backend/internal/service/error_passthrough_service.go`，不属于账号调度。
- 历史提交 `562193408` 曾把“1 最高”实现为升序；后续合并和修复恢复为 DESC，本次
  FlowAI 重新确认升序为当前契约。这个历史过程必须保留在上下文里，避免再次无审阅地
  接受上游方向。

受保护路径：

- `backend/internal/service/gateway_scheduling.go`
- `backend/internal/service/openai_account_scheduler.go`
- `backend/internal/service/openai_gateway_scheduling.go`
- `backend/internal/service/gemini_messages_compat_service.go`
- `backend/internal/service/batch_image_public.go`
- `backend/internal/repository/account_repo.go`
- `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`

### 4.2 并发与风险注册

- `-1` 是 deny-all：在服务层和 Gateway 等待队列入口立即返回，不调用用户 Redis
  槽位，也不增加等待计数。
- `0` 保持历史 unlimited 行为；正数继续走 Redis 槽位和原有等待策略。
- 默认设置、各注册来源默认值、管理员编辑和批量设置都必须拒绝或安全归一化小于
  `-1` 的值。
- 风险注册在服务端风险指纹授权完成前必须使用 `-1`。授权失败、存储不可用或重复
  领取时 fail closed；只有 `ClaimSignupGrant` 成功后才应用 `0` 或正数等实际配置。

### 4.3 Project Mihomo 与账号代理池

FlowAI 的 Mihomo 控制面不是单一订阅 URL：

- 支持多个订阅 URL、稳定 key/name、URL/后端/静态三种获取模式和兼容请求头；上游
  返回 403 时要记录来源拒绝，不能把问题误判成 YAML 解析错误。
- 支持节点名称筛选、区域、延迟测试、自动路由容差/间隔、多端口 listener，以及
  在更新配置时复用现有 listener/账号绑定，避免无意清空生产代理。
- 创建/导入账号可从 Project Mihomo 池分配端口；已有直接代理 OAuth 流程仍需明确
  区分，不能静默改成池代理。
- Linux compose 中应用通过 `mihomo-sub2api` 网络别名访问 controller，运行时状态
  位于 `/app/data/mihomo`。发布只替换应用镜像，不用空目录覆盖该状态。

### 4.4 账号导入与批量测试

- 导入流程支持分组和代理池选择，后端返回本次新建账号 ID；前端批量测试只预加载
  这些账号，不把历史账号混入。
- 批量测试 modal 的开始动作必须显式触发；进度、失败原因、SSE 状态码和按钮文案
  使用 i18n key，不能在英文界面回显中文硬编码。
- 上游改动 `AccountTestModal.vue`、导入 DTO 或 SSE 事件时，先核对这条链路的返回
  ID、选中集合和 locale，再解决冲突。

### 4.5 支付、发票和费用策略

- 人民币支付的充值手续费率、手续费是否计入余额、订阅购买是否收手续费是独立的
  管理配置；前端展示和后端结算必须使用同一快照。
- XZNOAuth 发票流程保存本地 `invoice_applications` 状态，覆盖草稿、校验、税费、
  申请、取消和 PDF 下载；client secret 只能留在服务端。
- 2026-08-29 起，23 服务器停止生产 BEpusdt；Sub2API 中专用 quote/order/reconcile/webhook
  代码已移除，`206_add_usdt_payments.sql` 保持原样用于历史数据兼容，不代表生产入口仍启用。
- GM 是独立项目：使用 `gm-epusdt` 分支的不可变 GHCR 镜像，独立 Compose、容器、配置
  和数据目录；域名由 Caddy 直接代理到 GM，不在 BEpusdt 目录上改造。
- Sub2API 使用普通 EasyPay provider 实例接入 GM。USDT 自定义方式需要把前台方式映射为
  GM 可用 selector，例如 `usdt_trc20 -> usdt.tron`，并使用 EasyPay 回调
  `/api/v1/payment/webhook/easypay`。
- GM 的 `supported_assets` 为空时不得启用 Sub2API 支付方式。启用前必须同时验证钱包、
  RPC/监听、创建订单、回调、两端订单状态和链上证明；容器健康、页面打开或 HTTP 200
  不能单独证明链上结算。

### 4.6 签到和奖励

- 业务日期唯一性、事务账务、重复请求幂等、负余额保护、风控指纹和审计不可被关闭
  某个模式绕过。
- 普通签到与幸运签到的奖励方向、概率、倍率/固定金额、正向阶梯、两位小数结算和
  未充值用户折减是独立配置；未充值策略不应误伤幸运签到。
- 配置版本和管理员确认/并发更新检查必须保留；损坏配置、风控依赖不可用时 fail
  closed。

### 4.7 邮件广播和通知

- 任务创建时固定受众、变量和中英文模板快照；发送状态由数据库驱动。
- 领取必须使用行锁和 `SKIP LOCKED`，状态路径为 `pending -> sending -> sent/failed`。
  已发送记录不能重置成 `pending`。
- SMTP 已接受但客户端得到错误时结果不确定，不能自动重发；`sending` 残留需先查
  provider/SMTP 证据再人工处理。
- 通知邮件模板、维护公告、登录/注册邮件和公告弹窗均属于用户可见功能，不能只合并
  后端发送逻辑而丢掉模板或 locale。

### 4.8 注册、访问和上游错误安全

- 注册 challenge 绑定请求指纹、最短耗时、陷阱字段和一次性消费；邮箱域名/别名/一次性
  邮箱策略在注册和 OAuth 补全流程都生效。
- 已认证邮箱换绑不属于新账号注册：仍执行注册后缀白名单、精确地址/别名占用查重和事务
  原子守卫，允许已验证的自身新别名参与换绑，但不会放宽其他用户已占用的收件箱。
- 登录失败 IP（或 IP+UA）封禁使用独立 Redis counter 和管理页面；清理或改 key 时
  不能影响并发槽位。
- 公开访问 guard 只保护配置指定的公开 POST/API 场景；发布密钥不得写入日志或客户端
  错误。Cloudflare Turnstile/site protection 配置和 CORS 规则一起核对。
- 上游 URL、token、query secret 和内部地址在同步/流式/多平台错误中必须脱敏，但
  管理端诊断仍要保留足够的 upstream status、request id 和分类信息。

### 4.9 i18n、站点配置和品牌

- 真正的 locale 来源是 `frontend/src/i18n/locales/zh/`、`frontend/src/i18n/locales/en/`
  及其 `index.ts` 聚合入口；不要因上游同名文件变化删除整个目录或 FlowAI key。
- 中文和英文必须同时更新。账号管理文案写“1 为最高优先级，数值越小越优先使用”；错误透传
  namespace 才能写“小数值优先”。并发文案必须解释 `-1/0/正数`。
- 社区群名称/图标/链接、用户订阅页开关、购买订阅入口、充值费用策略和公开设置注入
  是一条完整 API/store/router/UI 链路，缺少任意一层都会导致刷新后回显错误。
- 主题初始化、主题切换、FlowAI logo/favicon、法律文档构建输入和公告弹窗属于品牌
  约定；Docker ignore 变更要确保法律文档和 locale 仍进入镜像。

### 4.10 构建与发布

- `.github/workflows/preview-image.yml` 只接受 `sub2api-flowai`，先执行契约检查，再
  构建 `linux/amd64` 镜像，并同时推送可变分支 tag 和短 SHA 不可变 tag。
- 23 服务器通过 `deploy/deploy-preview-image.sh` pull 后只重建 `sub2api` 应用服务，
  等待 `/health`；PostgreSQL、Redis、Mihomo 和 `/app/data` 不得被删除或重置。
- 服务器不得作为正式发布构建机执行 `docker compose build`、`pnpm build` 或 `go build`。
  回滚只切换已验证的镜像 tag；数据库迁移不能用回滚镜像盲目逆向。

## 5. FlowAI 迁移台账

以下文件是当前分支相对上游新增、修改或必须保护的迁移。数字前缀可能与上游其他功能
重复，必须以完整文件名识别，不能因为编号相同而覆盖或删除。

<!-- FLOWAI_MIGRATION_LEDGER_BEGIN -->
| 文件 | 功能域 | 发布注意事项 |
| --- | --- | --- |
| `backend/migrations/001_init.sql` | 已执行迁移 checksum 兼容 | 只能保留已登记的兼容改动，不能继续直接编辑 |
| `backend/migrations/185_auth_ip_bans.sql` | 登录 IP 封禁 | 与独立 counter、管理接口一起验证 |
| `backend/migrations/191_daily_checkin.sql` | 签到基础表/索引 | 只追加，不改已执行内容 |
| `backend/migrations/192_checkin_abuse_guard.sql` | 签到滥用防护 | 核对业务日期和指纹约束 |
| `backend/migrations/193_checkin_security_integrity.sql` | 签到安全完整性 | 失败时保留数据库状态 |
| `backend/migrations/194_checkin_flexible_lucky_rewards.sql` | 幸运奖励配置 | 与配置版本测试配套 |
| `backend/migrations/195_checkin_mode_switches.sql` | 模式开关 | 关闭模式不能绕过唯一性 |
| `backend/migrations/196_checkin_lucky_positive_probability.sql` | 幸运正向概率 | 核对概率边界 |
| `backend/migrations/197_checkin_two_decimal_precision.sql` | 两位小数 | 与账务精度测试配套 |
| `backend/migrations/198_checkin_positive_tiers.sql` | 正向阶梯 | 检查连续覆盖和权重 |
| `backend/migrations/199_checkin_default_lucky_distribution.sql` | 默认分布 | 不覆盖管理员已有配置 |
| `backend/migrations/200_checkin_balance_history.sql` | 余额历史 | 核对 history mode 和展示 |
| `backend/migrations/201_checkin_balance_history_mode.sql` | 历史模式标记 | 与回放/审计一致 |
| `backend/migrations/202_checkin_unrecharged_reward_policy.sql` | 未充值奖励策略 | 普通/幸运模式影响范围不同 |
| `backend/migrations/203_add_email_broadcast_tasks.sql` | 邮件任务/收件人 | 与状态机和快照一起迁移 |
| `backend/migrations/204_generalize_email_broadcasts.sql` | 通用广播字段 | 不重置已有 sent 状态 |
| `backend/migrations/205_add_invoice_applications.sql` | 发票申请 | 与 Ent 生成代码和 XZNOAuth 流程一致 |
| `backend/migrations/206_add_usdt_payments.sql` | 历史 USDT quote/order/event 表 | 保持迁移内容不变，不再由新的 BEpusdt 应用链路写入 |
| `backend/migrations/206_checkin_fingerprint_guard.sql` | 签到指纹防护 | 与同编号 USDT 文件并存 |
| `backend/migrations/207_tighten_checkin_ip_guard_default.sql` | 签到 IP 默认策略 | 不要误认为上游同编号迁移 |
| `backend/migrations/208_signup_risk_grant_guard.sql` | 注册赠送一次性领取 | 与 `-1` 初始并发顺序一致 |
| `backend/migrations/231_add_usage_log_requested_reasoning_effort.sql` | 记录映射前请求推理强度 | 本次上游 0.1.183 新增；可空字段，不改历史数据 |
| `backend/migrations/231_user_restrict_public_groups.sql` | 用户公开分组访问限制 | 本次上游 0.1.183 新增；默认 false，保留现有用户行为 |
<!-- FLOWAI_MIGRATION_LEDGER_END -->

`backend/migrations/001_init.sql` 的内容曾为保留生产 checksum 做兼容性修复（提交
`71a8c43ea`、`aeac5e5d3`）。已经执行的迁移原则上不可编辑；必须通过新增迁移或
明确的 `migrationChecksumCompatibilityRules` 处理历史误改，不能删除
`schema_migrations` 或手工改 checksum。

## 6. 上游合并冲突矩阵

| 冲突区域 | 合并前必须回答的问题 | 放行证据 |
| --- | --- | --- |
| 调度代码/SQL | `a.priority` 是否仍为 ASC？`ag.priority` 是否仍独立为 ASC？OpenAI score 是否把较小 priority 映射为更高分？ | 目标单测 + `make check-flowai-contract` |
| 并发/注册 | `-1` 是否在 Redis/等待队列前拒绝？风险账号是否先创建为 `-1`？ | concurrency、hotpath、signup risk tests |
| i18n | zh/en key 是否都存在？聚合入口是否仍加载模块？是否把错误透传文案复制到账号文案？ | locale compile/collision/default tests |
| 迁移/Ent/wire | 是否保留所有完整文件名？是否修改了已执行 SQL？生成代码是否与 schema 一致？ | migration runner、schema、backend tests |
| 支付/发票 | 普通支付、XZNOAuth、USDT 是否仍为正确的独立链路？回调是否幂等？ | payment/invoice/USDT tests + 配置核对 |
| 签到/邮件 | 唯一结算和 at-most-once 发送是否仍成立？是否把 `sending`/已发送重置？ | integration tests + 数据库状态核对 |
| Mihomo/部署 | controller 别名、Caddy、locale、`/app/data/mihomo` 和环境变量是否同时存在？ | compose config + 服务日志/health |
| 依赖/构建 | Go/Node 版本、锁文件、Docker build context 是否兼容？ | lint/typecheck/build/Actions |

遇到无法确认的冲突，停止发布并保留冲突文件、数据库日志和上游 commit；不能以“测试
通过”替代业务语义审阅。当前 `7b693ae42` 待审快照在完成前不得发布。

## 7. 历史提交索引（非合并提交）

下面的索引覆盖当前快照中相对 `upstream/main` 的全部 85 个功能/修复非合并提交。治理
文档提交按上面的受限规则动态豁免，但仍会被路径检查；脚本会逐个检查功能提交 hash
是否存在于标记区，新增代码提交未登记时，CI/发布门禁失败。

<!-- FLOWAI_LEDGER_NON_MERGE_BEGIN -->
| 日期 | 提交 | 说明 | 责任域 |
| --- | --- | --- | --- |
| 2026-05-19 | `9c65bb8d6` | feat: add theme-aware logo and preview deploy config | 品牌/部署 |
| 2026-05-20 | `9a05e6d83` | feat: add project mihomo proxy management | Mihomo |
| 2026-05-21 | `625f2ff3a` | feat: refine multi-port proxy management | Mihomo |
| 2026-05-21 | `e3c1062b7` | chore: update default multi-port proxy ports | Mihomo/部署 |
| 2026-05-22 | `332921be4` | fix preview mihomo controller deployment | Mihomo/部署 |
| 2026-05-27 | `1a04f35aa` | feat: improve project mihomo proxy routing | Mihomo |
| 2026-05-27 | `822b4a87f` | feat: assign accounts from project mihomo pool | Mihomo/账号 |
| 2026-05-28 | `72e9ed14a` | fix: improve project mihomo proxy handling | Mihomo |
| 2026-06-11 | `455f3a2a3` | feat: add account import groups and batch testing | 账号管理 |
| 2026-06-11 | `f738f7a70` | fix: include legal docs in docker build | 构建/法律文档 |
| 2026-06-12 | `c1c65040f` | feat: add public access guard | 访问安全 |
| 2026-07-01 | `b310b6801` | fix: prefer higher account priority | 账号调度 |
| 2026-07-02 | `3373360fc` | chore: set preview domain to flowai.cyou | 域名/部署 |
| 2026-07-02 | `bf465a3c3` | fix: improve project mihomo subscription compatibility | Mihomo |
| 2026-07-03 | `7e6a21a45` | fix: qualify grouped account scheduling query | 账号调度/SQL |
| 2026-07-03 | `4bb0ec81e` | chore: bump version to 0.1.142 | 版本 |
| 2026-07-06 | `3cb924ad9` | feat: improve upstream tracing and announcement popup | 诊断/公告 |
| 2026-07-07 | `dd945ded6` | feat: harden registration and deploy contact updates | 注册/部署 |
| 2026-07-07 | `38a8f8dc3` | feat: support static project mihomo sources | Mihomo |
| 2026-07-08 | `5a05a2d96` | feat: improve project mihomo proxy management | Mihomo |
| 2026-07-08 | `a7f544b54` | feat: redesign notification email templates | 邮件 |
| 2026-07-08 | `551e2c533` | chore: use prebuilt preview image deployment | 部署 |
| 2026-07-08 | `491c656d7` | fix: use semver for preview image builds | 构建 |
| 2026-07-10 | `ad5816474` | fix: align flowai priority semantics after upstream sync | 账号调度 |
| 2026-07-10 | `dcccd801c` | fix: restore low-value account priority | 账号调度 |
| 2026-07-10 | `5f08828fa` | fix: satisfy ci lint after upstream sync | CI |
| 2026-07-10 | `fd0874431` | feat: add configurable community group link | 站点配置 |
| 2026-07-11 | `4dbc8c83f` | feat: improve community settings and database backup image | 站点/备份 |
| 2026-07-11 | `0b92464bb` | feat: add user subscriptions feature flag | 订阅 |
| 2026-07-11 | `68f8815c0` | test: update settings API contract | 设置测试 |
| 2026-07-13 | `80275b533` | fix: prefer higher account priority values | 账号调度 |
| 2026-07-13 | `71a8c43ea` | fix: preserve applied migration checksum | 迁移 |
| 2026-07-13 | `562193408` | fix: define priority one as highest | 历史语义漂移 |
| 2026-07-13 | `2977d946a` | test: align priority ordering after upstream sync | 账号调度测试 |
| 2026-07-15 | `91d6a3d64` | fix: preserve FlowAI compatibility with v0.1.156 | 兼容性 |
| 2026-07-16 | `eda0be6aa` | feat: improve account testing and registration flow | 账号/注册 |
| 2026-07-16 | `d27697bc6` | test: update Project Mihomo admin service stub | Mihomo测试 |
| 2026-07-16 | `0f97772e0` | test: complete Project Mihomo admin service stub | Mihomo测试 |
| 2026-07-17 | `6bc2766a2` | build: increase frontend heap for Docker images | 构建 |
| 2026-07-18 | `f9957a098` | feat: optionally credit recharge fees to balance | 支付 |
| 2026-07-21 | `e50df5a50` | feat(payment): allow disabling subscription fees | 支付 |
| 2026-07-21 | `696af4af3` | fix(security): upgrade axios to 1.18.1 | 依赖安全 |
| 2026-07-22 | `ea2ed4fcf` | feat(security): add automatic login IP bans | 访问安全 |
| 2026-07-22 | `a3076680d` | fix(security): upgrade x/text for GO-2026-5970 | 依赖安全 |
| 2026-07-22 | `d8c9b6f42` | refactor(security): isolate auth ban Redis counter | 访问安全 |
| 2026-07-23 | `4cf5d6265` | fix(security): redact upstream addresses from client errors | 错误安全 |
| 2026-07-24 | `9e30d4e88` | fix(payment): clarify recharge fee notice | 支付/i18n |
| 2026-07-26 | `ba2dea7ba` | feat(checkin): add secure daily balance rewards | 签到 |
| 2026-07-26 | `843443b7b` | fix(checkin): handle query row cleanup | 签到 |
| 2026-07-26 | `e38b0974a` | feat(checkin): add configurable reward modes and shortcuts | 签到 |
| 2026-07-26 | `4e02a67d6` | refactor(checkin): polish navigation entry points | 签到/UI |
| 2026-07-27 | `8e93556fa` | feat(checkin): separate lucky odds and settle to cents | 签到 |
| 2026-07-28 | `8797a4ef2` | feat(checkin): add weighted lucky reward tiers | 签到 |
| 2026-07-31 | `fa72c3c6b` | feat(checkin): show rewards in balance history | 签到/UI |
| 2026-07-31 | `e0c7e13b3` | feat(checkin): label lucky balance history | 签到/UI |
| 2026-08-02 | `f7a9bb86c` | feat(checkin): reduce rewards for unrecharged users | 签到/风控 |
| 2026-08-03 | `710b1bc92` | feat: add admin email broadcasts | 邮件 |
| 2026-08-03 | `cdffe4a5e` | fix: prevent duplicate email broadcast delivery | 邮件/幂等 |
| 2026-08-03 | `afc7f17cd` | fix: make email broadcast delivery at-most-once | 邮件/幂等 |
| 2026-08-05 | `0bf8327c9` | fix: retry OpenAI model capacity failures | OpenAI调度 |
| 2026-08-08 | `d85df51f4` | feat: add self-service invoice applications | 发票 |
| 2026-08-08 | `79959dba2` | fix: satisfy invoice lint checks | 发票/CI |
| 2026-08-08 | `2b91558e1` | feat: make invoice integration settings editable | 发票 |
| 2026-08-08 | `ae9d2066d` | feat: apply invoice fee payer policy | 发票 |
| 2026-08-08 | `bc1434464` | fix: update nanoid security patch | 依赖安全 |
| 2026-08-08 | `1623194bb` | feat: show invoice status on payment orders | 发票/订单 |
| 2026-08-09 | `bfadfec88` | feat: harden signup and check-in grant controls | 注册/签到 |
| 2026-08-09 | `8674c62c2` | test: check abuse guard interface assertion | 签到测试 |
| 2026-08-12 | `b91f93064` | fix(i18n): restore account threshold translations for 0.1.175 | i18n |
| 2026-08-12 | `96082d695` | chore(i18n): deduplicate merged threshold copy | i18n |
| 2026-08-16 | `ec0198a11` | fix: align Docker Go toolchain with go.mod | 构建 |
| 2026-08-18 | `826326132` | fix(auth): enforce registration email risk policy | 注册 |
| 2026-08-20 | `6a711a76e` | feat: add invoicing and USDT payment flows | 发票/USDT |
| 2026-08-23 | `16517e464` | chore: snapshot sub2api-flowai local changes | USDT配置快照 |
| 2026-08-23 | `69ba15b7a` | feat: consolidate USDT payment integration | USDT |
| 2026-08-23 | `9013ea7f0` | fix: complete merged USDT cashier wiring | USDT |
| 2026-08-23 | `cd6e05c6e` | feat: finalize USDT payment gateway deployment | USDT/部署 |
| 2026-08-24 | `c9d406e28` | fix(usdt): harden payment gateway configuration | USDT |
| 2026-08-24 | `783426c12` | fix: satisfy backend CI checks | CI |
| 2026-08-24 | `d5bf81b58` | fix: make nil assertions lint-safe | CI/测试 |
| 2026-08-24 | `720ffa185` | fix: keep sqlmock cleanup from failing unit tests | CI/测试 |
| 2026-08-24 | `106676f6a` | test: align failover fixtures with priority policy | 调度测试 |
| 2026-08-24 | `87ece0b48` | test: align schedulable projection integration order | 调度/SQL测试 |
| 2026-08-25 | `aeac5e5d3` | fix: preserve production migration checksum | 迁移 |
| 2026-08-25 | `453ffc9fa` | fix: support deny-all user concurrency | 并发/注册 |
| 2026-08-29 | `f6828cd9c` | fix: make account priority 1 highest | 账号调度 |
| 2026-08-29 | `b77636d2c` | feat(payment): replace BEpusdt with GM through EasyPay | 支付/GM |
| 2026-08-30 | `52d8138e2` | test: align schedulable projection with priority policy | 调度/SQL测试 |
<!-- FLOWAI_LEDGER_NON_MERGE_END -->

## 8. 历史合并提交索引

合并提交本身也要记录，因为冲突解决可能改变文件内容，即使没有新增 feature commit。
以下是当前分支基线之后的合并提交；下一次执行上游合并后，必须把新的 merge hash 和
审阅结论加入标记区。`git merge-base HEAD upstream/main` 变化后，以命令输出为准补齐。

<!-- FLOWAI_LEDGER_MERGE_BEGIN -->
| 日期 | 合并提交 | 说明 |
| --- | --- | --- |
| 2026-08-29 | `ea2096d33` | Merge upstream main 0.1.183 updates；保留 FlowAI 支付布局、i18n 和账号调度，吸收上游流式/定价/智谱团队版更新 |
| 2026-08-29 | `828348048` | Merge upstream main 0.1.183 into sub2api-flowai；保留 FlowAI 调度/i18n/支付/并发，并修正邮箱换绑与注册别名规则冲突 |
| 2026-08-25 | `1e1b7e9ed` | Merge upstream main 0.1.181 into sub2api-flowai |
| 2026-08-24 | `bd3b7b205` | merge upstream main 0.1.180 into sub2api-flowai |
| 2026-08-24 | `75c324cf3` | Merge branch Wei-Shaw main into sub2api-flowai |
| 2026-08-24 | `98aa5c4e7` | Merge upstream main for 0.1.179 release |
| 2026-08-23 | `e95e5c793` | Merge origin/sub2api-flowai into sub2api-flowai |
| 2026-08-20 | `80404e162` | Merge branch Wei-Shaw main into sub2api-flowai |
| 2026-08-20 | `2f8e9a3d8` | Merge origin/sub2api-flowai into codex/usdt-bepusdt-integration |
| 2026-08-20 | `ade58371b` | Merge upstream main into codex/usdt-bepusdt-integration |
| 2026-08-18 | `07530219f` | Merge upstream v0.1.178 into sub2api-flowai |
| 2026-08-16 | `5c797fdec` | Merge upstream main into sub2api-flowai |
| 2026-08-13 | `c27407454` | Merge upstream main into sub2api-flowai |
| 2026-08-12 | `ff655f5e0` | Merge upstream main into sub2api-flowai |
| 2026-08-09 | `a3b4d092d` | merge upstream 0.1.173 into flowai |
| 2026-08-08 | `42657445b` | Merge upstream main into sub2api-flowai |
| 2026-08-05 | `c53613da4` | Merge upstream main into sub2api-flowai |
| 2026-08-03 | `906a07793` | Merge upstream main into sub2api-flowai |
| 2026-07-31 | `02bac999e` | Merge upstream v0.1.169 into sub2api-flowai |
| 2026-07-29 | `95cf05b72` | Merge upstream v0.1.168 into sub2api-flowai |
| 2026-07-27 | `7e360d053` | Merge upstream v0.1.166 into sub2api-flowai |
| 2026-07-26 | `f5f370d65` | Merge upstream v0.1.165 into sub2api-flowai |
| 2026-07-23 | `3c49e128b` | Merge upstream v0.1.164 into sub2api-flowai |
| 2026-07-22 | `bdba2af2a` | Merge upstream main into sub2api-flowai |
| 2026-07-22 | `c834f933d` | Merge upstream v0.1.163 into sub2api-flowai |
| 2026-07-20 | `ee2229043` | Merge upstream v0.1.162 into sub2api-flowai |
| 2026-07-19 | `97ada00f1` | Merge upstream v0.1.161 into sub2api-flowai |
| 2026-07-17 | `b72cf8fff` | Merge upstream v0.1.160 into sub2api-flowai |
| 2026-07-17 | `4777921b9` | Merge upstream v0.1.159 into sub2api-flowai |
| 2026-07-16 | `679f63b4c` | Merge upstream v0.1.158 into sub2api-flowai |
| 2026-07-16 | `8a6dacc08` | Merge upstream v0.1.157 into sub2api-flowai |
| 2026-07-15 | `464096cf5` | Merge upstream v0.1.156 into sub2api-flowai |
| 2026-07-14 | `33348dd72` | Merge upstream v0.1.155 into sub2api-flowai |
| 2026-07-13 | `fdf28670b` | Merge upstream v0.1.153 into sub2api-flowai |
| 2026-07-13 | `bbcf05df3` | Merge upstream main into sub2api-flowai |
| 2026-07-10 | `68b10d81d` | Merge upstream main into sub2api-flowai |
| 2026-07-10 | `56b4c23d3` | Merge origin/main into sub2api-flowai |
| 2026-07-10 | `d55704df9` | chore: sync upstream 0.1.150 |
| 2026-07-10 | `2bfb4f4e0` | Merge origin/main into sub2api-flowai |
| 2026-07-08 | `673e7459f` | Merge origin/main into codex/flowai-theme-deploy |
| 2026-07-06 | `5ed689e30` | Merge origin/main into HEAD |
| 2026-07-04 | `43338ffd0` | Merge branch Wei-Shaw main into codex/flowai-theme-deploy |
| 2026-07-03 | `97e7ae573` | Merge origin/main into codex/flowai-theme-deploy |
| 2026-07-01 | `58f311002` | Merge origin/main into flowai theme deploy |
| 2026-06-22 | `ec51b4112` | Merge origin/main into flowai theme deploy |
| 2026-06-11 | `76525d7d7` | Merge origin/main into flowai theme deploy |
| 2026-06-08 | `438a24d56` | Merge upstream main into flowai theme deploy |
| 2026-06-06 | `3939cccd3` | Merge branch Wei-Shaw main into codex/flowai-theme-deploy |
| 2026-06-03 | `efe64b783` | Merge branch Wei-Shaw main into codex/flowai-theme-deploy |
| 2026-05-30 | `5ac674002` | Merge branch Wei-Shaw main into codex/flowai-theme-deploy |
| 2026-05-29 | `7ed6edabe` | Merge branch Wei-Shaw main into codex/flowai-theme-deploy |
| 2026-05-28 | `56e77c9ff` | Merge branch Wei-Shaw main into codex/flowai-theme-deploy |
| 2026-05-27 | `4b8644cb4` | Merge branch Wei-Shaw main into codex/flowai-theme-deploy |
| 2026-05-26 | `85295c2f6` | Merge branch Wei-Shaw main into codex/flowai-theme-deploy |
| 2026-05-24 | `a993d72e4` | Merge origin/main into codex/flowai-theme-deploy |
| 2026-05-22 | `7e63ea74e` | Merge branch Wei-Shaw main into codex/flowai-theme-deploy |
| 2026-05-20 | `78ed94741` | Merge origin/main into codex/flowai-theme-deploy |
<!-- FLOWAI_LEDGER_MERGE_END -->

## 9. 发布记录模板

每次发布在工单、PR 或发布记录中复制以下字段；不要把密钥、JWT、TOTP、支付 secret
或 SSH 私钥写入文档。

```text
日期/时区：
发布人：
分支：sub2api-flowai
发布前 HEAD：
上游基线：
应用版本：
本次 FlowAI 变更（提交 hash + 功能）：
上游合并提交及冲突结论：
契约检查：PASS / FAIL（命令与输出）：
提交台账检查：PASS / FAIL：
测试结果与未运行项：
新增迁移及 checksum/执行结果：
镜像 tag：
镜像 digest：
23 服务器部署目录：
上一版本镜像 tag/digest：
health/settings/public/容器日志：
业务 smoke test：
异常、回滚和未验证风险：
```
