# FlowAI 分支契约

> 适用分支：sub2api-flowai
>
> 这份文档记录 FlowAI 分支相对于上游 main 的业务约定、必须保留的功能和
> 发布边界。它不是上游项目的通用说明。每次合并上游或发布前，必须先阅读
> docs/FLOWAI_CHANGELOG.md，再执行 make check-flowai-contract，并按发布核对清单留下证据。

## 1. 分支关系和所有权

| 引用 | 作用 | 规则 |
| --- | --- | --- |
| upstream/main | Wei-Shaw/sub2api 的上游代码 | 可以定期合并；合并前必须审阅差异 |
| origin/sub2api-flowai | FlowAI 发布源 | 只从该分支构建和发布 |
| 本地 main | 上游镜像/本地工作引用 | 不作为 FlowAI 合并目标，不在发布流程中修改 |

必须遵守以下边界：

1. FlowAI 的发布分支是 sub2api-flowai，不是本地 main，也不是临时
   codex/* 工作分支。
2. 合并上游使用明确的 upstream/main 合并提交。禁止用
   git checkout upstream/main -- .、整树覆盖或无审阅的 rebase 丢弃本分支提交。
3. 冲突解决后按“上游新增行为”和“FlowAI 保留行为”逐项归类；不能只看最终能否编译。
4. 上游迁移文件和 FlowAI 迁移文件都必须保留。已有迁移文件不能为了消除冲突而重写，
   需要变更时新增迁移文件。
5. 若业务约定需要改变，先修改本契约、测试和发布清单，再改实现；不能在一次上游
   合并中顺手改变语义。
6. 每个新增或修改的功能提交都必须在变更台账中登记提交 hash、行为、受保护路径和
   验证方式；合并提交还必须记录冲突结论。没有台账记录的提交不得发布。

当前代码检查基线是最后已审并合入的上游 0.1.181（`3b7753a8e`）。上游引用可以暂时
前进，但在预审、冲突结论和契约检查完成前，不能把新上游当作已同步版本发布。版本号会
继续变化，行为契约不会因为版本号变化而自动变化。

## 2. 账号调度优先级

### 2.1 账号优先级的当前约定

FlowAI 账号调度采用**1 为最高优先级，数值越小越优先**：例如优先级 20 的账号
先于优先级 80 的账号进入调度。前端账号表单的有效起始值为 1。

在优先级相同的情况下，具体调度器再按各自策略比较负载、排队深度、最近使用时间、
OAuth 类型、重置时间或其他评分因子。不能把这些平局规则误写成优先级方向。

当前必须保持数值越小优先的代码路径包括：

- backend/internal/service/gateway_scheduling.go 的分层筛选、普通排序和混合调度；
- backend/internal/service/openai_account_scheduler.go 的评分、Top-K 和 compact 重试；
- backend/internal/service/openai_gateway_scheduling.go 的传统 OpenAI 选择；
- backend/internal/service/gemini_messages_compat_service.go 的 Gemini 选择；
- backend/internal/service/batch_image_public.go 的图片批量账号排序；
- backend/internal/repository/account_repo.go 中账号字段的 ASC 排序。

账号优先级和分组成员优先级是两个字段：查询中的 ag.priority ASC 是分组成员顺序，
a.priority ASC 是账号自身调度优先级。合并 SQL 时必须分别核对，不能因为看到不同
字段就把账号排序改反。

### 2.2 不同规则的优先级不能混用

错误透传规则的 priority 是另一套规则，当前设计是数值越小越先匹配，相关文案位于
frontend/src/i18n/locales/{zh,en}/admin/settings.ts 的 errorPassthrough 区域。
它不控制账号调度，不能用它推导账号优先级。

历史提交 562193408 曾将账号调度改成“1 最高”，后续上游合并提交
bd3b7b205 又恢复为 DESC。本分支现再次明确采用“1 最高、数值越小越优先”；后续
若出现 DESC 账号排序，必须停止发布并核对冲突，而不是按上游结果直接接受。

## 3. 用户和账号并发

用户并发值的含义固定如下：

| 值 | 行为 |
| --- | --- |
| -1 | 立即拒绝请求；不写入/读取用户 Redis 槽位，不进入等待队列 |
| 0 | 不限制用户并发，保持兼容逻辑；不访问用户槽位 Redis |
| 正数 | 作为用户最大并发数；占满后按现有等待策略处理 |
| 小于 -1 | 输入层拒绝，或在默认配置/风险路径安全归一化为 -1；不得写入无效值 |

这个约定适用于：

- 系统默认用户并发；
- 注册来源（email、LinuxDo、OIDC、微信、GitHub、Google、钉钉等）的默认并发；
- 管理员编辑用户；
- 管理员批量设置用户并发；
- 网关进入等待队列前的用户并发检查。

账号槽位也保留 0 不限制和负数拒绝的保护逻辑，但不要把账号并发、用户并发、
图片并发和 RPM 当作同一个配置。

风险注册的安全顺序不可改变：带有服务端风险身份的注册请求先以余额 0 和并发
-1 创建/进入授权判断，只有 ClaimSignupGrant 成功后才写入实际赠送并发（包括
0 不限制）。风险存储不可用时 fail closed，不能退回无限并发。

图片并发、RPM 限制和现有 API 路径不是本次用户并发功能的改动范围。上游合并如果
触碰这些区域，必须单独记录原因和测试，不得借并发改动顺手修改。

## 4. FlowAI 专属功能清单

以下功能来自 FlowAI 分支提交或其配套迁移，发布时必须保留后端、前端、文案、测试和
数据库迁移的完整链路。

### 4.1 USDT、GM/EasyPay 和发票

- BEpusdt 的独立应用链路已经移除，不再保留专用配置、路由、后台设置、收银台或服务端
  `usdtpayment` 包；上游合并不得重新引入这条链路。
- `206_add_usdt_payments.sql` 作为历史数据库迁移保持不变，既有表和数据不在本次代码清理中
  删除或改写；后续历史数据处置必须单独设计迁移。
- 当前生产 USDT 方案是独立项目 GM：单独仓库和 `gm-epusdt` 分支、单独镜像、Compose、
  容器和数据目录。不得把 GM 文件覆盖到 BEpusdt 目录，也不得把 BEpusdt 数据目录直接
  当作 GM 数据目录使用。
- Sub2API 通过现有 EasyPay provider registry 接入 GM。自定义支付方式必须显式映射 GM
  selector，例如前台 `usdt_trc20` 映射上游
  `usdt.tron`；回调继续使用 /api/v1/payment/webhook/easypay。
- GM 只有在公开配置返回非空 `supported_assets`，对应钱包、RPC 和链监听均通过检查后，
  才能在 Sub2API 启用 EasyPay USDT 方式。容器健康或页面 HTTP 200 不能代替收款就绪。
- 发票申请使用本地 invoice_applications 状态和 XZNOAuth 外部服务；订单校验、税费
  状态、申请、取消和 PDF 下载流程必须一起保留。
- 迁移入口至少包括 205_add_invoice_applications.sql 和历史保留的
  206_add_usdt_payments.sql；支付配置说明见 docs/PAYMENT.md 和 docs/PAYMENT_CN.md。

钱包存在、支付容器运行或浏览器跳转成功，都不等于链上结算成功。发布验收只能把 GM
公开资产/RPC 状态、回调 HTTP 状态、两端订单状态、provider 日志和链上证明分别记录，
不能用其中一项代替全部证据。

### 4.2 每日签到和奖励

- 用户每天每个业务日期最多结算一次，普通签到和运气签到共享唯一性约束。
- 普通/运气模式、正向概率、倍率/固定金额、正向阶梯、两位小数结算和未充值用户
  奖励策略均可由管理员配置。
- 余额结算、历史记录、配置版本、审计和防批量/防滥用检查必须保持事务和幂等性。
- 风控配置读取或参数损坏时 fail closed；关闭某个模式不能绕过每日唯一和账务约束。
- 相关迁移为 191_daily_checkin.sql 至 202_checkin_unrecharged_reward_policy.sql
  中的签到文件（迁移编号与上游同编号文件交错存在），产品规则见
  docs/DAILY_CHECK_IN_PRD.md。

### 4.3 管理员邮件广播

- 广播任务和收件人快照持久化在 email_broadcast_tasks 与
  email_broadcast_recipients，创建时固定受众、模板快照和变量。
- 领取使用事务、FOR UPDATE SKIP LOCKED 和 pending -> sending 状态，多个实例不能
  领取同一个收件人。
- 只有处于 sending 的收件人可以转为 sent 或 failed；sent 收件人不会再次进入
  领取查询。
- SMTP 已接受但客户端观察到错误时，不能自动重试该收件人，因为结果不确定；该收件人
  进入失败/人工核对路径。人工重试只允许明确的 failed 收件人，不能把已发送记录批量
  重置为 pending。
- 进程在 SMTP 成功后、数据库完成前崩溃时，可能留下 sending 状态。这是防重复优先的
  保护状态，处理前必须结合 SMTP/provider 日志核对，不能为追求计数而直接重发。
- 相关迁移为 203_add_email_broadcast_tasks.sql 和
  204_generalize_email_broadcasts.sql；重复发送保护实现位于
  backend/internal/repository/email_broadcast_repo.go 和
  backend/internal/service/email_broadcast_service.go。

### 4.4 注册风控和安全控制

- 注册 challenge 绑定请求指纹、最短耗时、陷阱字段和一次性消费，并使用 Redis 频率控制。
- 注册邮箱策略覆盖域名白名单、可注册主域归一化、明显别名和一次性邮箱域名。
- 登录失败可按 IP 或 IP+UA 作用域形成自动封禁，管理员可以审阅和释放记录。
- 风险身份的一次性注册赠送记录使用服务端 HMAC 指纹，不保存原始 IP/UA 作为业务身份。
- 208_signup_risk_grant_guard.sql、185_auth_ip_bans.sql、认证路由和相关测试必须
  与前端注册流程一起核对。

### 4.5 Project Mihomo / FlowAI 部署配置

- 管理端 Project Mihomo 支持订阅源、静态 YAML、节点测试、同步和多端口 listener 配置。
- 应用通过 mihomo-sub2api 网络别名访问 Mihomo controller；运行时状态持久化在
  /app/data/mihomo，不能在发布时用空目录覆盖。
- 相关管理路由为 /api/v1/admin/proxies/project-mihomo 及其 sync、节点测试接口。
- deploy/docker-compose.preview.yml、deploy/Caddyfile.flowai 和对应环境变量是
  FlowAI 预构建发布的一部分，不能只更新应用镜像而漏掉部署配置。

### 4.6 i18n 和前端契约

- 中文和英文 locale 必须同时维护；新增功能不得以另一语言的硬编码文本替代 locale key。
- 账号优先级文案必须表达“1 为最高优先级，数值越小越优先使用”；用户并发文案必须明确 -1 拒绝、0
  不限制、正数为上限。
- 错误透传规则的“小数值优先”文案只允许出现在其自身 namespace，不得污染账号管理文案。
- i18n 目录、locale 注册入口、相关测试和构建产物均属于发布文件。合并冲突时不能因为
  上游文件同名就删除 FlowAI key；应先做 key 级合并，再执行 locale 测试。

## 5. 数据库迁移和不可变性

1. 应用启动时通过 backend/internal/repository/migrations_runner.go 按文件名执行迁移，
   已执行迁移会校验 SHA-256 checksum。
2. 已执行迁移内容原则上不可编辑；新需求必须新增迁移文件。
3. 代码中的 migrationChecksumCompatibilityRules 只允许明确登记的历史误改兼容，
   不是修改任意旧迁移的通行证。
4. 001_init.sql 中关于优先级的旧注释曾为保留生产 checksum 被调整过；它不是运行时
   调度实现的权威来源。判断优先级必须看本契约、调度代码和测试。
5. 发布前备份 PostgreSQL、应用数据、Redis AOF 和 Mihomo 状态；迁移失败时先保留日志
   和数据库状态，不要删除 schema_migrations 或手工改 checksum。
6. 迁移编号可能与上游同编号但不同文件名并存，冲突解决必须以完整文件名为键。

## 6. 预构建镜像发布约定

FlowAI 使用 GitHub Actions 在构建机打包，服务器只拉取镜像，不在生产机执行前端/Go 构建：

- 工作流：.github/workflows/preview-image.yml；触发分支：sub2api-flowai。
- 镜像：ghcr.io/aboutnb/sub2api:sub2api-flowai（可变发布标签）以及带提交短 SHA 的
  不可变标签 sub2api-flowai-<sha12>。
- Compose：deploy/docker-compose.preview.yml。
- 服务器脚本：deploy/deploy-preview-image.sh，执行 pull、只重建 sub2api 应用服务、
  等待 /health，不删除 PostgreSQL、Redis、Mihomo 数据目录。
- 默认持久化路径包括 /root/flowai-preview-data/app、postgres、redis 和 Caddy
  目录；实际生产路径以服务器 .env.preview 为准。

通过 Termius 连接服务器时，发布命令应指向已验证的 SHA 标签，例如：

~~~bash
cd /root/flowai-preview/deploy
SUB2API_IMAGE=ghcr.io/aboutnb/sub2api:sub2api-flowai-<sha12> \
  ./deploy-preview-image.sh
~~~

发布后至少核对 /health、/api/v1/settings/public、容器状态、应用日志、迁移结果和
关键业务 smoke test。回滚只切回上一个已验证的不可变镜像标签；数据库迁移不能靠简单
回滚镜像逆向撤销，必须按迁移兼容性和备份方案处理。

当 `upstream/main` 已前进时，`make review-flowai-upstream` 会列出新增提交、重叠路径、
受保护路径和模拟冲突，并默认阻断后续操作。只有人工完成冲突矩阵核对后，才可以用本次
上游的**完整 SHA**显式确认：

```bash
FLOWAI_UPSTREAM_REVIEW_ACK=<upstream-main-full-sha> make review-flowai-upstream
```

这个确认值只用于把审阅对象锁定到不可变提交，不代表可以跳过行为判断、测试或迁移检查。
上游 SHA 变化后必须重新审阅，不能复用旧确认值。

## 7. 契约变更规则

当业务确实要改变本契约时，变更必须同时包含：

- 本文件中对应条目的更新；
- 后端和前端实现；
- 正向测试和回归测试；
- 迁移/数据兼容说明（如涉及数据）；
- 发布清单中的验收项和回滚说明；
- 一条明确的提交说明，说明这是有意的 FlowAI 行为变更，而不是上游合并副作用。

只修改代码、不更新本文件或变更台账的提交，不能作为 FlowAI 发布候选。
