# FlowAI 分支变更台账

### 2026-10-29 v0.2.19 发布候选（尚未部署）

- 已审阅并合并上游 `a3298b46ebe239fb0720b068a1eb3383955bdac2`（67 个提交、130 个变更路径）；上游版本为 `0.2.19`。新增 Grok 账号模型目录同步/诊断、Gemini 同账号重试与模型目录更新、OpenAI OAuth 非流式兼容、分页溢出保护、图片上传大小校验、国产 Coding Plan 频率型 429 冷却及多项计费/协议修正；本次上游无数据库迁移。
- 保留 FlowAI 主题、教程、生图智能路由、图片工作台、站内工单、充值签到、渠道监控、优先级/并发契约及 GMPay 主动查单。四处冲突按模块合并：账号仓储保留优先级排序并接入 Grok 管理目录持久化；Gemini Messages 合并模型/缓存安全与可配置同账号重试；账号编辑弹窗合并 Grok 诊断和自定义标识；密钥使用弹窗保留客户端配置/教程并接入分组模型列表。
- 保留本地 GMPay 查单修复：GMPay 仅通过 POST 表单传参，避免 URL 与表单重复参数；其他 PHP 易支付渠道行为不变。此项只影响主动查单请求，不修改订单余额或支付状态。
- 后端 `go test ./...`、`go vet ./...` 全部通过；`make -C backend test` 随后在 `golangci-lint run ./...` 阶段停止，因为本机未安装 `golangci-lint`。前端 Vue 431 文件/3,354 项测试、lint、typecheck、i18n 检查和生产构建通过；图片工作台 39 文件/599 项测试与生产构建通过。GitHub 同源 CI、安全扫描和不可变镜像构建仍为发布前提。当前尚未推送或部署。

### 2026-10-29 GMPay 查单重复参数修复（尚未部署）

- 订单 `sub2_20261028UmdhK0Qc` 的回调已到达，但主动查单收到 `invalid request`。PHP 查单兼容改动同时发送 URL 与表单参数，与 GMPay 拒绝重复参数的规则冲突。
- GMPay 标准查单路径恢复仅 POST 表单传参，其他 EasyPay/PHP 通道保持原行为。保留上游签名、商户、订单、交易号、金额、支付类型、链上哈希校验及本地幂等入账，不放宽 GMPay 参数校验。
- GMPay 模拟服务改为检查合并后的参数唯一性，覆盖真实服务行为，防止双份参数再次漏测。尚未补确认生产订单，未改动余额。

### 2026-10-26 v0.2.18 已发布至 23 服务器

- 发布源码 `c79fe06716a52314957be6d32b89b1e6895292b6`；同源 CI `40303716905`、Security Scan `40303716906`、Aivoza Image `40305339104`（#148）均成功。生产候选固定为 `ghcr.io/aboutnb/aivoza-sub2api@sha256:1d5c95e4a2f665cdb2b07f45522702da42a587bcdd815059311be873e02ca9c7`。
- 通过 Termius 确认 23 服务器 `/root/flowai/deploy`；蓝绿脚本 SHA256 与源码一致。候选 `flowai-app-next-c79fe06716a5` 在回环端口 3017 启动，`state/aivoza-v0218-c79fe06716a5.json` 返回 ready，容器 healthy，二进制与公开设置均为 `0.2.18`，镜像 revision 与发布源码一致。
- 发布前备份 `/root/flowai/backups/aivoza-v0218-20261026`：PostgreSQL custom dump 596578426 字节并通过 TOC 校验，app-data 29514 字节并通过 tar 校验，Redis RDB 1743685 字节并通过 redis-check-rdb；生产配置、部署脚本、历史发布状态和容器元数据均已保存。应用日志不纳入 app-data 归档，原数据与依赖服务保留。
- 三条新增迁移已应用，数据库 checksum 与源码 TrimSpace 后 SHA256 逐条一致：`246_enable_gemini_image_generation_groups.sql`、`247_scoped_admin_api_keys.sql`、`248_group_image_count_billing.sql`。候选启动日志未检出 panic/fatal/migration failed/checksum mismatch。
- 蓝绿状态已确认 promoted，当前 `flowai-app`（`b4444df731db`，端口 3017）与状态文件 new_id 一致，Caddy 配置与预期 after 完全一致。上一版价格修正镜像 `4df9eb40ddf9...` 保留为 `flowai-app-rollback-e9df5963e1a6`（3016）；两者均 healthy。
- 候选首页、教程、工单、图片工作台、公开设置和价格 API 均 200；未认证用户/管理员工单 API 均 401。独立公网检查确认两个域名 `/health` 正常且公开版本均为 `0.2.18`，主站页面入口正常，价格 API 返回 success、schema 1.1、CNY 和 31 个模型；美国 API 域名页面 404 符合既有隔离配置。
- 服务器双域名连续探测覆盖北京时间 18:37:51 至 18:44:53，共 360 次（每域名 180 次）全部 HTTP 200，失败 0；任务正常完成，原始结果保存在备份目录 `switch-probes.txt`。PostgreSQL、Redis、GM、Mihomo、Caddy、Guardian 和 workbuddy2api 的容器 ID 与启动时间均未变化；Guardian 到 `http://flowai-app:8080/health` 在清理前后均正常。
- 更早回滚容器 `flowai-app-rollback-e1ce49badf70` 入站连接为 0，Caddy 无旧别名、容器配置无旧 IP 引用；优雅停止以 exit 0 退出后删除。仅在检查全部运行及停止容器均无引用后删除其 `f3e3b198ec7d...` 镜像；最终仅保留当前与上一版两个应用容器，数据卷、备份、回滚状态及无关服务均保留。最近应用日志未检出 panic/fatal/migration failed/checksum mismatch。
- 未执行真实付费生成、支付或生产写入验收；Guardian 管理员认证未在本次重新验收。

### 2026-10-26 v0.2.18 镜像构建修复

- 修正 `deploy/Dockerfile` Go builder 镜像版本为 1.27.3，与 `backend/go.mod` 的最低 Go 版本一致；此前 v0.2.18 候选镜像在 `go mod download` 阶段因 builder 为 1.27.2 确定性失败。
- 发布前需确认同源 CI、安全扫描及 Aivoza Image 成功，再使用不可变 digest 通过 23 服务器蓝绿流程部署。
- 修复提交：`f8daec67f`。

### 2026-10-26 v0.2.18 发布候选

- 已审并合并上游 `ba81c7fd758e425d67f856d601147ef0c5263d3f`（相对已发布 v0.2.17 共 54 个提交、244 个文件），合并提交 `5779f635a1f89a95babdc484dc9380c4dc0e9fc7`；版本为 `0.2.18`。
- 接入分范围管理员 API 密钥及审计/限流/事务保护、Vertex 账号配额上限、Gemini 分组生图开关、按实际返回图片数计费、OpenAI 请求级 Retry-After 退避、流式断开后的用量计费和调度快照优化；保留 Aivoza 主题、支付确认、教程、生图智能路由、站内工单、账号优先级与并发契约。
- 新增迁移 `246_enable_gemini_image_generation_groups.sql`、`247_scoped_admin_api_keys.sql`、`248_group_image_count_billing.sql`，均为增量文件，未改写既有迁移；旧镜像回滚不逆向删除新列/表或已入账数据。
- 图片工作台生产依赖升级至 `streamdown 2.7.1`、`@streamdown/math 1.0.3`、`katex 0.19.1`、`dompurify 3.4.17`，消除 Security Scan 报告的高危依赖；`npm ci`、生产 audit 结果为 0 漏洞，React 工作台 599 项测试和生产构建通过。
- 本地后端 `go test ./...` 与 `go test -tags=unit ./...`、Vue 账号/设置/管理员密钥专项 187 项、前端 lint/typecheck、React 图片工作台生产构建通过；本机未安装 `golangci-lint`，该门禁待同源 GitHub CI 验证。尚未部署 23 服务器。

### 2026-10-17 公开价格 API 报价修正已发布至 23 服务器

- 版本保持 `0.2.17`；origin/main 与 upstream/main 无新增业务提交，复用现有 main 整合原工作区 4 个价格 API 文件，未改动原工作区。
- 公开报价改为官方模型价格数字按站内 CNY 余额 1:1 口径乘公开分组倍率一次；官方目录优先，无官方价格的自定义模型回退分组/渠道配置。移除固定 7.5 汇率，不叠加用户专属或高峰倍率；不修改实际网关计费及生产分组配置。
- 图片按次报价遵循独立图片倍率，普通按次报价用分组倍率；token 保留一位小数，按次保留两位小数，避免 0.06 等低价舍入失真。无效、非有限或舍入为零的必要价格继续省略，独占组排除、禁用条目和 30 分钟缓存保持原规则。
- 受影响路径：`backend/internal/service/provider_pricing.go` 及 service/handler 测试、`docs/PROVIDER_PRICING_API_PRD.md`；无数据库迁移、前端或支付修改。
- 本地 ProviderPricing 服务/处理器专项、完整后端 unit、golangci-lint（0 issues）和严格契约（0 failure/0 warning）通过；同源 CI `38949586873`（含集成与前端全量）、Security Scan `38949586853`、Aivoza Image `38950943175` 均成功。
- 发布源码 `487cc0e94f05c1461c410bff9495baad4c489284`，固定镜像 `ghcr.io/aboutnb/aivoza-sub2api@sha256:4df9eb40ddf98c62259144c0c832f9437f6944996091bfcb1f1cf80cef1206d2`。蓝绿状态 `/root/flowai/deploy/state/aivoza-pricing-487cc0e94f05.json` 为 promoted；当前 `flowai-app` 端口 3016，上一版 `e4ea87bf7` 保留为 `flowai-app-rollback-e1ce49badf70`（3015）。
- 两公网域名健康与价格 API 返回 200，均导出 33 个模型；Pro GPT-5.5 输出报价为 7.5 CNY/1M tokens。当前 gpt-image2 渠道实际模式为 `per_request`、基础价 0.08、分组倍率 0.2，按次导出 0.02；文档中的 0.06 是精度示例，不是生产配置期望值。未修改线上计费模式或倍率，未执行付费生成。
- 首页/帮助页 200，未认证用户/管理员工单 API 401；Caddy 配置与发布状态一致，应用日志无 panic/fatal/迁移失败/checksum mismatch。数据库、Redis、GM、Mihomo、Caddy、Guardian 容器 ID 与启动时间保持不变；Guardian 到稳定 flowai-app 地址的 health 返回正常。
- 备份 `/root/flowai/backups/aivoza-pricing-20261017`：PostgreSQL custom dump 528630430 字节并通过 TOC 校验，app-data 29178 字节并通过 tar 校验，Redis RDB 1262578 字节并通过 redis-check-rdb；配置与容器元数据已保存。Redis 备份按实际无密码配置在单独进程清除继承的 REDISCLI_AUTH 后成功，未改 Redis 服务配置。

- 服务器双域名连续健康探测覆盖北京时间 02:53:41–03:11:35，共 798 次（每域名 399 次）全部 HTTP 200，失败 0；探测已停止，原始记录及 `probe-summary.json` 保留在备份目录。
- 更早 v0.2.16 `flowai-app-rollback-636b3301667b` 与 v0.2.15 `flowai-app-rollback-5f6b13d3ec6f` 入站连接均为 0，核对依赖无旧目标并备份元数据后优雅停止、删除，随后仅删除全部容器均无引用的对应镜像。最终当前与上一版共 2 个应用容器健康，数据卷、备份与其他服务保留。

### 2026-10-16 v0.2.17 与最新本地功能已发布至 23 服务器

- 已审上游 `b0a02a3304a0a78cdd003ddda1dd23ba02c03084`（47 提交、150 文件），合并提交 `0477bb8bfb648eddbd051bd6f6a59cceed0085fe`。接入 Gemini 4 Argon/思考档位、Vertex 429 重试、OpenAI 搜索模拟与模型退役、Responses created_at、退款 pending 扣减保留和 Stripe 本地快照。无新增或修改数据库迁移。
- 6 处冲突逐块整合：支付确认统一使用同一 provider，保留 USDT 签名/PID/订单/金额/币种/链上交易校验，普通易支付增加上游确认；查单兼容 GET 参数和 PHP 字符串状态，拒绝异常金额，失败保持未入账并返回 503。保留账号折叠筛选、可访问性测试及 Aivoza 主题；OpenAI API key 接入搜索模拟开关。
- 整合原工作目录最新生图更新：渠道与模型分开选择、提示词优化独立开关、失效选择显式报错、图片/遮罩预校验、提交锁、取消后不再提交生成、重试复用原异步任务；桥接检查跨标签账号连续性，路由初始化加行锁并保留用户编辑。原工作目录与未提交工作不改写。
- 新增公开只读 `GET /api/provider/pricing`，仅非独占分组，遵循 v1.1 协议、CNY 固定换算与 30 分钟缓存，受公开 IP 限流保护；按 PRD 不应用用户/分组倍率，不公开凭据，省略视频价格。Wire 已重新生成。
- 提交 `67054ae31` 接入最新图片工作台控件与公开价格 API；提交 `c76322ded` 在其他集成套件前清理工单测试夹具；提交 `97a9ba465` 修正工单断言并关闭仓储查询行，均已纳入本候选台账。
- 本地支付专项、供应商价格/生图/智能路由专项通过；Vue 427 文件/3205 项、React 39 文件/599 项、lint/typecheck 和两套构建通过。发布源码为 `e4ea87bf76e1cdb10e5994f1a9787961c2223354`，版本 `0.2.17`；同源 CI `38880178280`、Security Scan `38880178255`、Aivoza Image `38881413389` 均成功。
- 生产固定镜像为 `ghcr.io/aboutnb/aivoza-sub2api@sha256:f3e3b198ec7d4a66e0050a0987aeafa4a24054d4c5dd419bd8b6c0dbe1f0c969`。蓝绿 prepare/promote 已完成，当前 `flowai-app` 使用回环端口 3015，发布状态 `/root/flowai/deploy/state/aivoza-v0217-e4ea87bf76e1.json` 为 `promoted`；上一版 v0.2.16 保留为 `flowai-app-rollback-636b3301667b`。Caddy 两个应用上游均已切换到 `aivoza-e4ea87bf76e1:8080`，PostgreSQL、Redis、GM、Mihomo、Caddy 保持运行。
- 发布前备份位于 `/root/flowai/backups/aivoza-v0217-20261016`，PostgreSQL dump、app-data、Redis RDB、容器/配置/依赖状态均已完成并校验；本次无新增数据库迁移，未删除数据卷、备份或依赖服务。
- `aivoza.com` 与 `us.aivoza.com` `/health` 均返回 200；切流连续探测共 3728 次（每域名 1864 次），失败 0。公开价格 API 返回 HTTP 200、schema `1.1`、33 个模型并带 30 分钟公共缓存；未认证用户/管理员工单接口返回 401；首页、帮助页返回 200，图片工作台入口保持既有 301 跳转。启动日志未发现 migration failed、checksum mismatch、panic 或 fatal。

### 2026-10-14 v0.2.16 与站内工单已发布至 23 服务器

- 发布源码 `61579cd85606882feacec8198c3db8597f858742`，版本 `0.2.16`；固定 GitHub 镜像 `ghcr.io/aboutnb/aivoza-sub2api@sha256:b1145eb844046ce21753227762b7c0de9baf0f39019360ad60ebe45d92f4341d`。CI `38567146664`（含完整集成）、Security Scan `38567146685`、Aivoza Image `38568371535` 均成功。工单集成测试清理顺序已修复，安全扫描依赖下载错误诊断后重跑成功。
- 蓝绿 prepare/promote 完成，当前 `flowai-app`（3014），发布状态 `/root/flowai/deploy/state/aivoza-v0216-61579cd85606.json` 为 promoted；Caddy 路由新版本并热加载成功。上一版支付补丁 v0.2.15（`4daabff0855a9b9a9d261fe46dfad56cc1214c8b`）保留为 `flowai-app-rollback-5f6b13d3ec6f`（3013）。三个更早版本连接排空后优雅停止并删除，只删除无容器引用的对应镜像；数据卷、备份与无关服务保留。
- 两域名 `aivoza.com` / `us.aivoza.com` 健康与公开版本检查通过；切流连续探测 1902 次、失败 0（切流前 1628，切流后 274；每域名 951），证据 `probe-summary.json` / `switch-probes.jsonl` 位于下述备份目录。工单、管理员工单、教程、图片工作台与监控页面 curl 均 200；未认证的用户/管理员工单 API 均 401。未在生产创建真实工单或执行攻击测试。
- 5 条新增迁移（242_group_i18n、243/244/245_gemini_cached_contents 系列、256_add_support_tickets）已应用，并按应用 TrimSpace 后 SHA256 规则逐条核对数据库 checksum；证据 `migration-checksums.txt`。应用源码 label、发布状态、依赖启动时间核验一致，最终应用/数据库/Redis 健康，应用日志未检出 panic/fatal/迁移失败/checksum mismatch。
- Guardian 旧 hosts 指向旧应用 IP；备份元数据后使用同镜像/权限/端口/数据卷重建，仅修改 BASE_URL 为 `http://flowai-app:8080`。健康、DNS 与上游 0.2.16 版本读取通过；旧停止容器已删除，配置备份 `/root/flowai/backups/guardian-v0216-20261014` 保留。原持久监控连接 enabled=false，保存的管理密钥返回 401 Invalid admin API key；保留禁用状态，未修改凭据或声称管理员认证通过。
- 发布前备份 `/root/flowai/backups/aivoza-v0216-20261014`：PostgreSQL custom dump 515292958 字节并通过 TOC 校验，应用数据 29607 字节并通过 tar 校验，Redis RDB 1061114 字节并通过 redis-check-rdb；配置/容器/依赖启动时间已保存。持续写入日志单独保留在原目录，不纳入数据归档。

- 已审并合入上游 `9c2cca9e627291bc719bada3d57de23b625a4180`，219 文件，新增 4 条迁移。保留工单、教程、生图智能路由、主题和支付回调主动确认；Ent/Wire 生成保留全部定制 provider。
- 8 个冲突逐块处理，接入分组多语言、Vertex API Key、Gemini 缓存与运营账号名。API key 带路由的返回也执行请求语言本地化。本地后端全量 unit、Vue 427 文件/3201 项测试、React 37 文件/567 项测试、前端 lint/typecheck 及两套构建通过。全仓 PostgreSQL 集成、golangci-lint v2.14（0 issues）、严格契约（0 failure/0 warning）也通过；生产部署证据见上文。

### 2026-10-14 站内工单安全冒烟与本地加固（随本次 v0.2.16 发布）

- 工单采用 JWT 与管理员中间件、按用户隔离、参数化查询、事务锁及最多 10 个未关闭工单；工单正文与主题以纯文本显示。
- 本地实测后修复非法空字符及极大分页触发 500、超大无关 JSON 字段可被接受的问题。工单 JSON 写请求上限 64 KiB，拒绝未知字段、非法 UTF-8 和尾随 JSON；查询限制每页 100 条、页码 1000000、搜索 200 字符。
- 工单独立按用户限流：创建 5 次/分钟、20 次/小时，回复 20 次/分钟、120 次/小时；不受全局面板限流开关或管理员豁免影响。Redis 不可用时暂停工单创建/回复、保留读取；每工单最多 200 条消息，以事务内锁和计数保护并发边界。
- 修复顶部签到菜单焦点在触发按钮时 Esc 无法关闭的问题。新增错误文案保持中英文一致。
- 独立本地数据库及真实账号完成 133 项 API 检查、47 项浏览器断言，后端目标包及 race、6 项 PostgreSQL 集成测试、80 项前端测试、typecheck、ESLint、go vet 通过；具体范围、证据与限制见 [工单安全冒烟报告](SUPPORT_TICKETS_SECURITY_SMOKE.md)。
- 本地镜像 `sub2api-local:ticket-security-20261014` 已更新并通过健康、配置与数据计数核验。该记录原为本地验收；源码现已合入 main 并随上述 v0.2.16 发布，未执行线上攻击测试。

### 2026-10-14 EasyPay USDT 回调主动确认修复（GMPay 已部署，Sub2API 尚未部署）

- 复盘订单 `sub2_20261013lGNRzAmc` 确认线上无对应 GM 订单或链上收款，攻击者利用 EasyPay
  创建请求与回调规范串相同的缺陷复用签名。第一层修复已拒绝未知/重复/缺失字段、错误 PID、
  非 MD5 类型、未配置支付类型及 `return_url` userinfo 注入，并把回调支付类型绑定到本地订单。
- 第二层将 USDT 回调降级为触发信号：使用验证回调的同一 provider 实例主动查询 GM，只有响应
  签名、PID、`out_trade_no`、`trade_no`、支付类型、金额、成功状态和链上交易哈希全部匹配才入账。
  查单失败或不一致保持 `PENDING`，追加 `PAYMENT_UPSTREAM_CONFIRMATION_FAILED` 审计并返回 503
  触发 GM 重试；支付宝/微信 EasyPay 回调保持原路径。
- 配套 GMPay 提交 `20f52b07afbccfcca4b8c6a0ae10f34ec641faf8` 在 `gm-epusdt` 分支新增受 PID/密钥、IP 白名单和订单所属 API Key 保护的只读
  `/payments/epay/v1/order/create-transaction/api.php`，完整响应使用商户密钥签名。发布顺序固定为
  GM 先、Sub2API 后。GitHub Actions `docker-gm-epusdt` run `38441514491` 已成功构建
  `gm-epusdt-sha-20f52b0`，多架构镜像 digest 为
  `sha256:4c92dfc419f220e8bf53fae28a0907e5d4207deefe3aee421b2c50b2636a5cd2`。23 服务器的独立
  `flowai-gm` 已固定到该 digest 且健康，查单端点对非法凭据返回 HTTP 200 的安全失败响应；
  `flowai-app`、Caddy、PostgreSQL、Redis 未重建且保持健康。当前尚未执行真实链上支付。
- 本地已通过 GMPay 全仓 Go 测试、Sub2API 全仓 Go 测试、完整 unit service/handler/provider、
  漏洞碰撞、主动确认成功和九类不一致拒绝、RMB 旁路、待支付状态与审计专项测试；FlowAI
  严格契约检查以 0 failure/0 warning 通过。GMPay 镜像构建已通过，Sub2API GitHub CI、镜像构建
  和两端生产验证待补齐。

### 2026-10-10 v0.2.15 已切流至 23 服务器（Guardian 验收未完成）

- 发布源码 `fa80f638880a57ecb019c3b1a8ab6bda8ee059a1`，线上版本 `0.2.15`；同源 CI `38031782768`、Security Scan `38031782778` 和 Aivoza Image `38032373407` 成功。后端全量单测、前端 414 文件 / 3106 项测试、lint、类型检查及构建通过。
- 生产固定镜像 `ghcr.io/aboutnb/aivoza-sub2api@sha256:25e2b77ee91d7a16f81a31669ef27ba37950827b523bed76729643513252a6fd`；蓝绿状态 `state/aivoza-v0215-fa80f638880a.json` 为 `promoted`。当前 `flowai-app` 回环端口 3012，上一版 v0.2.14 保留为 `flowai-app-rollback-7dd817f3af17`（3011）。
- `242_drop_platform_check_constraints.sql`、`255_channel_monitor_platform_catalog.sql`、`255_remove_retired_help_clients.sql`、`256_channel_monitor_v0215_platforms.sql` 均已执行，数据库 checksum 与源码一致。监控配置从 6 扩充至 14 个平台，保留原有开关、模型及范围；后台实际显示 14 个平台选项，公网监控已有 Gemini、DeepSeek 等数据。
- 候选启动日志无 checksum 错误、migration failed、panic 或 fatal；Caddy 校验成功（仅既存格式告警），PostgreSQL、Redis、GM、Mihomo、Caddy 启动时间未变化。两公网域名健康及版本接口均正常；本地续接验收再次直连确认两域名返回 `status=ok`、版本 `0.2.15`。登录、帮助、监控、生图页面 200，未认证发票接口 401。切换前 26 次服务器探测全部 200；切换后连续探测最终统计尚未取回，不宣称全程零失败。
- 备份 `/root/flowai/backups/aivoza-v0215-20261010` 内 PostgreSQL、应用数据、Redis 已校验，并保留监控配置快照、探测和发布日志。
- 续接发现 Guardian 的 `SUB2API_BASE_URL=http://sub2api:8080` 无法解析；其健康状态不能证明上游连通。Guardian 内 `http://flowai-app:8080/health` 实测返回正常，当前应用 IP 为 `172.19.0.11`。以默认用户写入临时 hosts 映射被权限拒绝；后续修复输入没有可靠成功输出，不能视为已修复。应将 Guardian 持久配置指向蓝绿切流后稳定容器名称，并验证上游认证请求，避免仅写临时 IP 或中断当前应用网络。
- 旧 v0.2.10 `flowai-app-rollback-c8734dbe73a6`（3010）入站连接曾核验为 0，但在 Guardian 路由修复并复核前继续保留；尚未删除其容器或镜像。Termius 输入控制未可靠响应，直接 SSH BatchMode 认证被拒绝，收尾受阻。待恢复控制后结束/统计 `switch-probes.jsonl`、核实 Guardian、优雅清理 v0.2.10 及无引用镜像；保留 v0.2.14 回滚、备份和所有数据卷。

### 2026-10-10 v0.2.15 合并记录（已部署）

- 发布前复核原历史工作目录，发现 10-09 下午及 10-10 新增的发票/教程代码。按文件差异整合：发票实名认证语义 403 且不刷新 Token、保留错误码/request_id、取消固定 20 单限制、先保存申请再税费续付并刷新同一记录；保留主线支付限流、赠送/折扣字段、分页重置与鉴权回归。已付待确认记录不再显示第二次提交表单或收款入口。
- 教程清理 Read Frog/Roo Code、代码块行高及展示版号 v1；保留酒馆两种分类和现有共享配置。新增完整文件名 `255_remove_retired_help_clients.sql` 只清理两个退役教程，既有同编号监控迁移不动；发布前数据库备份可恢复相关内容。
- 审阅上游 `3a6fd1c9db07203ca308aaba69e502bc1f35b307`（22 个提交，含 2 个合并；140 文件），吸收平台清单、多协议 profile、Cline/Command Code、模型同步能力控制、Claude 计费/钱包冷却和 OpenAI 搜索历史修复。
- 11 个冲突文件逐块合并：保留 Aivoza 主题、教程共享客户端配置、Codex 模型发现、路由设置入口和 PostgreSQL 客户端构建；Go 保持 1.27.2、x/tools 0.51.0，接入 validator 直接依赖，lint 使用 v2.14.0。调度优先级、并发、支付/签到/邮件和旧迁移保持原语义。
- 新增上游 `242_drop_platform_check_constraints.sql`（平台校验转到应用目录）及本地 `256_channel_monitor_v0215_platforms.sql`（追加 Cline/Command Code）；已提交 255 文件不改写。监控设置按动态目录读取，未来平台不会因初始快照遗漏。验证和生产证据见上方发布记录。

### 2026-10-10 渠道监控平台目录修复（已随 v0.2.15 部署）

- 被动监控 V2/V3 的配置仍沿用最早六个平台，导致后续新增平台被后端配置范围排除；新增 `255_channel_monitor_platform_catalog.sql`，补齐 Kimi、Zhipu GLM、DeepSeek、MiniMax、OpenCode、TypeSafe / Jev。只追加缺失项，保留明确关闭的平台、模型列表、全局开关和分组范围；显式空平台列表不自动开启，重复执行不重复追加或增加版本号。新安装默认包含全部 12 个监控维度（含历史 Kiro）。
- 后台监控设置从共享平台目录补齐可选项，未知自定义平台继续保留；未来缺失项在界面默认关闭，管理员可自行启用。平台名称统一展示；V3 卡片使用平台和分组组合键，修复同一复合分组跨平台时刷新复用错误卡片的问题。
- 前端监控相关 13 个文件 / 55 项测试、类型检查和改动文件 ESLint 通过；后端监控相关 unit 通过，新增仓储 integration 测试编译通过。PostgreSQL PGlite 实际执行迁移验证通过，覆盖旧配置保留、显式空配置、幂等和新安装默认值。本机 Docker daemon 不可用，完整原生集成测试未执行；后续已推送并随 v0.2.15 发布到 23 服务器，见上方记录。

### 2026-10-09 v0.2.14 已完成 23 服务器发布

- 发布源码 `ab520b4c5c5dbe336b0f147f1adaa3c3074ba4a8`，版本 0.2.14；同源 GitHub CI `37875387481`（含集成测试）、Security Scan `37875387480`、Aivoza Image `37876158329` 全部成功。安全补丁后的本地完整 Go unit、golangci v2.14、严格契约均通过。
- 镜像 `ghcr.io/aboutnb/aivoza-sub2api:sha-ab520b4c5c5d`，固定 digest `sha256:ea1f175db7428423b890ca0be2ed2379ccbad8630c28451b32158cbfd0f219d7`；服务器核验 amd64、完整源码标签与发布状态一致。
- 2026-10-09 03:42:24 UTC 开始 promote，当前 flowai-app 在回环 3011；上一版 0.2.10 `flowai-app-rollback-c8734dbe73a6` 在 3010。Caddy 路由匹配，PostgreSQL、Redis、GM、Mihomo、Caddy 启动时间未变化，Guardian 健康且实际连接域名仍为 aivoza.com。
- 切换前后服务器 curl 探测覆盖 03:40:38–03:46:10 UTC，两个公网域名各 269 次全部 200。公网版本均为 0.2.14；主站登录/工作台入口 200，美国 API 域名对应页面 404（既有隔离），未认证 auth/me 与 bootstrap 均 401。一次本机代理链路 TLS 连接失败，随后主站直连版本验证成功；不将该本机失败混入服务器连续探测统计。
- 两条新迁移校验和与源码一致：241_add_payment_order_bonus_amount.sql `18b6a524a9873d3e4a45e6f9bd03b6e69f4384b701825c659817e017f2e8244f`，241_add_typesafe_platform.sql `b6559525bf8d0b5d7c617e7415944f0d1fae8c408e9131ace84139795f66dcd2`。充值赠送档位、手续费及工作台/路由数据库开关未修改；启动及切流后日志无 panic/fatal/迁移错误。未执行真实生图或支付。
- 完整备份位于 `/root/flowai/backups/aivoza-v0214-20261009`，回滚状态 `state/aivoza-v0214-ab520b4c5c5d.json`。更早 0.2.8 容器入站连接为 0 后优雅停止并移除，无容器引用后删除旧镜像；保留数据卷、备份和无关服务。终端控制中断前只准备候选，恢复后重新启动探测并切流；一次重复排队 prepare 因容器名冲突失败，原 ready 状态经复核完好。

### 2026-10-09 v0.2.14 合并记录（已部署）

- 恢复后的 Security Scan `37874088635` 拦截了上游自带的 Go/HTTP2 漏洞（如 GO-2026-6617）；前端安全检查成功。按漏洞数据库固定修复版本升级 Go 1.27.2、golang.org/x/net v0.60.0 及其最小依赖集，同步 CI/Release/Docker 构建版本。官方 Alpine 镜像 amd64 可用，本地 govulncheck 可达漏洞为 0；7 条仅位于未调用依赖模块，未新增豁免。后端测试及新同源 GitHub 门禁继续验证。

- Go 1.27.2 导出格式升级要求 x/tools v0.51.0 与 golangci-lint v2.14；上游出站 HTTP/2 健康探测改用标准库 HTTP2Config，协议协商与代理配置回归通过。服务端保留 x/net 兼容层以维持 H1/H2 独立 idle timeout，GOAWAY 旧类型识别保留重试；仅对这两处弃用告警作说明，不豁免漏洞。修补前 GitHub CI `37874088611` 全部成功（含集成测试），修补后须重新验证同源门禁。

- 合并提交 `09ab7ed2ef7d219576c00b1fc8614a1fe2fbfd06`，候选 `79b0c8c37a55ab5e21f13a6e3d1b5250a2363dbd` 与后续治理提交已推送；GitHub API 确认主线一致，但无 Actions run/check suite。最初怀疑上游四个 VERSION 提交含 `[skip ci]`；进一步用 Actions 页面确认真正阻塞为 fork 因 Actions 使用情况被 GitHub 暂停，需维护者重新启用。不能仅靠 workflow API 的 active 状态认定仓库可运行。已通过现有 Chrome 维护者会话恢复，页面明确显示 Actions Enabled；本次治理提交触发恢复后的完整门禁。
- 23 服务器发布前备份 `/root/flowai/backups/aivoza-v0214-20261009` 已完成：PostgreSQL custom dump 485 MB 且 TOC 校验通过、应用/Mihomo 状态 29 KB、Redis RDB 956 KB 且 redis-check-rdb 通过、配置及旧容器元数据；尚未启动候选或切换流量。
- 本地嵌入静态资源的 Go 构建通过，pnpm 生产依赖审计例外校验通过。原历史工作区跟踪改动未发现上次发布后更新；保留全部未提交文件，不回抄旧实现。

- 复用现有 main 工作区，从 `856a7146a2e66fa855f45ace7a6e67cbc5e3657b` 合并已审上游 `5fc0e486c3f6a8a191b8bd140f39b60457f611cf`（VERSION 0.2.14，包含 v0.2.14 标签后的修复）；上游增量 194 提交、389 文件。历史脏工作区及素材保留。
- 43 个冲突文件按功能块合并：支付保持 GM/USDT、发票、默认赠送、手续费选项，并接入上游赠送/折扣、活动文案、bonus_amount 返利剔除及 EasyPay 回调安全修复；邮件保留注册上下文与错误分类，吸收原子尝试计数及哈希重置令牌；Ent/Wire 重新生成。
- 前端保留 Aivoza 主题、弹窗层级与 pending 防关闭、金额布局、教程共享配置和工作台智能路由；补入 TypeSafe/System One、Codex 远程目录与新模型、TPS 和额度展示。测试跟随新平台、协议规则和控件可访问性更新，保留原行为断言。
- 赠送配置读写统一支持上游范围；缺失保持旧默认，[] 明确关闭。报价一次计算，手续费不参与赠送；新增组合回归测试。密码重置旧格式链接需重新申请，记录在契约与发布核对表。
- 新增两条上游 241 迁移（支付免费额度、TypeSafe 平台约束），已有迁移均不改写。数据库回滚通过旧镜像兼容新增列，不撤销数据。
- 本地后端完整 unit、后续支付专项测试、golangci-lint v2.13（0 issues）通过；Vue 408 文件/3046 测试与后续客户端专项 32 测试、React 37 文件/567 测试、前端 lint/类型检查及 Vue→React 生产构建通过。本机 Docker daemon 不可用，集成测试交由 GitHub CI 门禁；生产证据待补齐。


### 2026-09-29 v0.2.10 已完成 23 服务器发布

- 发布源码 `b55e31bd98688cdb069227ab7734882176721ddd`，版本 `0.2.10`。GitHub CI `36567062039`、Security Scan `36567061892`、Aivoza Image `36568315666` 均成功且 SHA 一致。
- 镜像 `ghcr.io/aboutnb/aivoza-sub2api:sha-b55e31bd9868`，生产固定 digest `sha256:3857c4d4ffcb8333b042af3f85ff98eeb5c98db8de20a2379cc8eeaa1f3bb40f`；服务器核验 amd64 与完整源码标签一致。
- 服务器 23 通过蓝绿 prepare/promote 发布，当前 `flowai-app` 在回环端口 3010，上一版 0.2.8 `flowai-app-rollback-8f0253ae4f4f` 在 3009。旧 0.2.7 回滚容器及未引用镜像已优雅清理；数据卷、备份和依赖服务未动。
- 发布前备份 `/root/flowai/backups/aivoza-v0210-20260929`：PostgreSQL custom dump 454 MB（TOC 校验通过）、应用/Mihomo 状态、Redis RDB 1.44 MB（4502 keys、redis-check-rdb 通过）、配置和容器元数据。3010 候选健康，启动日志无 panic/fatal/迁移或 checksum 错误。
- `aivoza.com` 与 `us.aivoza.com` 健康及公开版本均为 `0.2.10`；主站连续 20 次健康探测通过。工作台入口返回 200，未认证 `/api/v1/auth/me` 与工作台 bootstrap 返回 401；未执行真实生图、支付或计费请求。

### 2026-09-29 v0.2.10 合并候选（已部署）

- 本地功能增量：使用教程（客户端配置、模型选择、API 探测示例、Markdown 安全渲染）、后台教程草稿/预览/发布/回滚；生图提示词优化、多渠道智能路由与工作台细节。顶部旧社群入口按本地实现替换为教程，原客服联系方式保留。
- 采用原工作区 HEAD、本地文件、已发布 main 的三方合并；保留主线注册防护、复合图片模型、端点布局、订阅开关及上游客户端修复。原工作区和 data/、素材不改动。
- 新增 240_help_documents.sql 至 253_help_api_probe.sql，以及 254_smart_route_cross_platform.sql。已有迁移不改写；新表/列为增量，旧应用可继续运行，镜像回滚不会撤销数据库迁移。教程种子更新仅处理未被管理员修改的文档。
- 智能路由仅工作台托管 key 可跨平台，普通 key 保留同平台和全局开关规则；候选授权、计费类型、模型及账号可用性继续验证。托管 key 可在密钥页编辑路由但不暴露内部凭据，公网入口仍拒绝其凭据。初始化使用数据库事务及 key 行锁，不覆盖已有配置、不保留全局无限缓存；排除按分组名称绕过独占组授权的旧代码。
- 迁移 Ent/Wire 由合并后的 schema/provider 生成。修复教程 i18n 转义、酒馆文档分类编辑和客户端抽取中的 URL、Windows 目录及模型目录回归。后端 unit、Vue 384 文件/2826 测试、React 37 文件/567 测试、lint/类型检查、两套构建已通过；本机 Docker 不可用，集成测试明确跳过，GitHub CI 是部署门禁。

- 审阅并合并上游 `a60a29549f488a854966aaec9541abbe006cac22`，相对已发布 0.2.8 的上游基线共 103 个提交、214 个文件；无上游 SQL 迁移变化。跨分支整树差异不代表本次删除量。
- 五处冲突逐块处理：Wire 重新生成并保留定制服务；账号用量保留主题并接入稳定的 Claude 重置额度组件；账号编辑保留布局并传入模型映射；模型广场吸收视频独立倍率；后台趋势吸收 token/消费切换并保留主题。
- 保留账号优先级 1 最高、负并发拒绝、注册防护、GM、充值赠送、签到、邮件防重与既有迁移。Wire 生成补齐工具依赖 checksum，不升级依赖。
- 已通过五个相关前端测试文件共 132 项。完整验证、教程/工作台增量、GitHub 门禁和生产发布尚未完成。
- 合并提交 `2bc658f4e`；合并前基线治理提交 `d75d18387` 已纳入主线记录。后续功能提交继续按本台账登记。
- CI 首轮阻止了教程演示中的全角空格和教程仓储未检查 rollback 返回值；修复后重新跑同源门禁，不跳过检查、不使用空提交重试。严格契约已补齐 15 条新迁移和功能提交的完整登记。

### 2026-09-24 v0.2.8 已完成 23 服务器发布

- 发布源码 `6c9a9edb91be6c6fc71eef41bbe46e7685d66e99`；GitHub CI `35963082397`（含集成测试与 lint）、Security Scan `35963082357`、Aivoza Image `35963859072` 均成功且 SHA 一致。本地后端完整 unit、前端 2712 测试、图片工作台 565 测试、两套生产构建及严格契约通过。
- 镜像 `ghcr.io/aboutnb/aivoza-sub2api:sha-6c9a9edb91be`，生产固定 `sha256:6ef8397aedb4733ad12cfcee106a060342de2940a6c785eb7f9375824301b5ea`；已核验 amd64 与完整源码标签。
- Termius 蓝绿 prepare/promote 完成，当前 flowai-app 在回环端口 3009，上一版 `flowai-app-rollback-1c0133f0d69a` 在 3008；发布状态、实际镜像及 Caddy 路由一致。PostgreSQL、Redis、GM、Mihomo、Caddy 启动时间均未变化，新旧应用健康。
- 238b/239/240 三条新增迁移已登记，SHA256 分别为 `420af91eec85bdb60474640612a0f4d069db6ddfb21e6ed702b89713f3599287`、`66feb546785dfa385d8efd268a40276cd8ffdb0cead7026cd79e339a3b69edf1`、`3823bfea5f64feb58f5fcebc341c6877eaefc97834b4cd652c8e83ad08ed78da`，与源码去首尾空白后的校验和一致。启动前再次确认旧分组/渠道倍率命中均为 0；镜像回滚不撤销迁移，蓝绿并存期不调整推理倍率。
- 两个公网域名均返回 0.2.8；主站登录 200、美国 API 域名登录 404（既有配置）、未认证用户接口均 401。候选工作台入口 200、缺失资源 404、未认证 bootstrap 401，公网工作台入口 200。生产工作台开关实查为 true，本次未修改开关，未执行真实生成或支付。
- 切换期间 Python urllib 对主域名 30 次均被 Cloudflare 返回 403/1010，美国域名 30 次均为 200；不能宣称全程探测零失败。随后同服务器 curl 对两个公网域名连续 60 次全部 200，主域名源站 HTTPS 直连健康正常，本机另行确认公网版本 0.2.8。未更改 Cloudflare 安全策略。新应用日志未匹配到 panic/fatal、迁移失败或校验和错误。
- 备份 `/root/flowai/backups/aivoza-v028-20260924` 包含 450 MB PostgreSQL custom dump（TOC 校验通过）、应用/Mihomo 状态、Redis RDB（4437 keys，redis-check-rdb 通过）、配置和旧容器元数据。初次含活跃日志的 tar 返回文件变化提示，随后排除日志重新生成并校验 `app-state.tar.gz`。回滚状态 `state/aivoza-v028-6c9a9edb91be.json` 含敏感配置，不公开输出。
- 更早的 `7f5910da3385` 容器入站 HTTP 连接为 0 后已无限等待优雅停止并移除，确认无容器引用后删除 `e50e929269db` 旧应用镜像。仅保留当前及上一版，不动备份、数据卷和无关服务。

### 2026-09-24 v0.2.8 发布候选（尚未部署）

- 从 main `9b6b7259c3d116d784347cd44823f4457c70dc32` 审阅上游 `a3eb7ef302961cba716dc78b39b93b60c467db0e`，237 个新增提交、473 个文件变化。复用现有 main 工作区，原历史工作区未提交修改不纳入本次发布。
- 吸收 OpenCode Go 用量、模型目录、Claude 版本同步、备份归档、可配置推理倍率、线下提现幂等、调度及协议兼容修复。保留优先级 1 最高、-1 并发拒绝、GM/EasyPay、充值赠送、签到和邮件防重、图片工作台默认/现有开关以及 Aivoza 主题。
- 13 个文本冲突逐块处理：release 工作流保留 fork 禁用及不回写主线版本；Wire 同时接入定制与上游服务；图片模型保留 Studio 请求边界并吸收复合模型解析；dev compose 保留发票变量；界面保留主题、弹窗堆叠及可访问性，并吸收用量、空订单、备份和审核功能；测试保留模型候选模拟。
- 三条新增迁移不改既有文件：238b 审核来源元数据、239 推理倍率、240 提现 operation_id。239 会移除分组 JSON 旧倍率键，镜像回滚不等于数据回滚；23 只读统计旧分组/渠道倍率命中均为 0，正式启动前需再次核验并备份，蓝绿并存期不调整推理倍率配置。
- 合并提交 `002f9da298b0fa1d014826e4bc8d938dd9a9d7b5`；前端 373 文件 / 2712 测试及图片工作台 36 文件 / 565 测试通过。已修复上游用例与定制 i18n/Teleport/订阅开关不兼容及 Select 双 listbox。后端服务、处理器、仓储、迁移测试通过；本地 Docker 不可用，旧 lint 二进制不支持 Go 1.27，集成测试和 lint 必须由 GitHub CI 验证。GitHub CI、安全扫描、镜像及生产验收仍为发布前提。

### 2026-09-21 图片工作台代码已发布到 23

- 发布源码 `5d6884aa6f7fa7268909ee0f99253d71d8a2b5ce`，版本保持 0.2.7。GitHub CI `35614245418`、Security Scan `35614245343`、Aivoza Image `35615350414` 均成功；本地 golangci-lint 零问题。
- 镜像 `ghcr.io/aboutnb/aivoza-sub2api:sha-5d6884aa6f7f`，生产固定 `sha256:2fda538d38c716e72f6341c958b68def63f6da244a7e8853e644f13cbc0556af`；已核验 amd64 和源码标签。
- Termius 蓝绿 prepare/promote 完成，当前 flowai-app 在 3008，上一版 `flowai-app-rollback-7f5910da3385` 在 3007；Caddy 路由与发布状态匹配，依赖服务启动时间未变化。
- 239 迁移已登记，规范化 SQL 校验和为 `553fc522dbcc6190fa34a422a08b648b5b49d99be7c9a37d8913c11a530e0ccb`；purpose 默认 standard 和唯一索引均已核实。工作台开关仍为 false，未进行真实生成、支付或计费验收。
- 公网主站登录、工作台 index 和构建产物均 200；缺失静态资源 404、未登录 bootstrap 401；工作台返回 SAMEORIGIN 与 frame-ancestors self。两个域名版本均为 0.2.7。
- 本机代理探测 90 次中 83 次 HTTP 200、7 次 TLS/连接失败，不能据此宣称全程探测零失败；切换后从服务器直接连续探测两个公网域名共 30 次全部 200。候选与回滚均健康，启动日志未匹配到迁移失败或致命错误。
- 备份 `/root/flowai/backups/aivoza-studio-20260921` 包含 447 MB PostgreSQL dump（TOC 已校验）、应用/Mihomo、Redis RDB、配置和旧容器元数据；状态 `state/aivoza-studio-5d6884aa6f7f.json` 包含敏感配置，不公开输出。
- 更早的 e11335eb47f2 容器连接排空后已优雅移除，无引用的 748d3873238d 镜像已清理；仅保留当前及上一版，不动备份、数据卷和依赖服务。

### 2026-09-21 本地图片工作台发布候选

- 用户确认将旧工作目录中的新增代码整理到 main 发布，版本保持 0.2.7，不额外合入未经审阅的上游提交。
- 三方比对保留已发布的客服、订阅开关、API 端点和界面修复；注册防护差异为 main 已有修复，不用旧代码覆盖。原工作目录和未提交素材/数据不改动。
- 功能提交 `bfa65c190faa1c2124c719725435013eff93e0cb`：内嵌图片工作台、受限同源消息桥接、专用内部密钥、既有图片网关认证/调度/计费链路和后台独立开关。默认关闭，未授权真实付费生成，不在发布时自动启用。
- 新增 `239_image_studio_keys.sql`，为普通密钥增加默认 standard 的 purpose 字段，为工作台内部密钥增加有效用户/分组唯一索引；发布前备份，旧应用兼容新增字段，不逆向删表/字段。
- 本地 Vue 338 文件 / 2458 测试、React 36 文件 / 565 测试、两套前端生产构建、Vue lint、后端完整 unit/integration 和嵌入资源测试通过。CI 新增 React 测试/构建及生产依赖审计；本地生产依赖仅两项 moderate（未被 StudioApp 引用的 Agent 依赖链），无 high/critical。
- GitHub 完整门禁、镜像和线上迁移核验仍为发布前提，本候选记录不代表已经部署。

### 2026-09-20 v0.2.7 已完成 23 服务器发布

- 已发布源码 `acb2633ad064eba87e8a664afd82ba828801f0ca`；GitHub CI `35454631025`（含 Docker 集成测试）、Security Scan `35454631069`、Aivoza Image `35455222658` 均成功且源码一致。本地后端完整单元测试、前端 2454 测试、lint/构建及严格契约通过。
- 镜像 `ghcr.io/aboutnb/aivoza-sub2api:sha-acb2633ad064`，生产固定 `sha256:e50e929269dbadf33da95868c790cb76b243c691f9484cbc3ac35c47f66c1b84`；已核验 amd64 和完整源码标签。
- Termius prepare/promote 蓝绿发布完成，当前 flowai-app 位于回环端口 3007；上一版 `flowai-app-rollback-e11335eb47f2` 保留在 3006。数据库、Redis、GM、Mihomo、Caddy 启动时间未变化。
- 两个公网域名均返回 0.2.7，切换期间 60 次健康检查全部 200；主站登录 200、两个域名未认证用户接口均 401。us.aivoza.com 原配置仅提供 API，登录页 404 为既有预期。启动日志未匹配到 panic/fatal、迁移失败或校验和错误；未执行真实支付或付费模型请求。
- 发布前备份 `/root/flowai/backups/aivoza-v027-20260920`：PostgreSQL custom dump 446 MB（TOC 校验通过）、应用/Mihomo 数据、Redis RDB、配置及旧容器元数据。本次无 SQL 迁移变化。回滚状态 `state/aivoza-v027-acb2633ad064.json` 含敏感配置，不公开输出。
- 确认更早的 4263d8d81dca 容器无入站 HTTP 连接后优雅停止并移除；确认所有容器均无引用后删除 ac1658681f50 历史应用镜像。仅保留当前及上一版，不清理备份或数据卷。

### 2026-09-20 v0.2.7 发布候选

- 合并已审上游 `1a9d49e16f7a22c432b428fce4af8d731f1fa364`，包括 v0.2.7 标签及其后的版本号同步提交；无 SQL 迁移变化。
- 新增 Seedance Ark 视频适配和插件宿主能力，吸收协议兼容修复；未安装或启用额外插件。
- 两处冲突：媒体调度保留本分支四返回值接口并采用上游平台参数；顶部模型广场保留 Aivoza 配色，吸收移动端入口及可访问性标签。
- 保留优先级 1 最高、-1 并发立即拒绝、GM/EasyPay、充值赠送、签到和邮件防重、既有迁移校验和及已发布界面修复。
- 本地前端 337 文件 / 2454 测试、lint、类型检查及生产构建通过。本机 Docker daemon 不可用，集成测试以 GitHub CI 为发布门禁；镜像构建和生产验证尚待完成。

### 2026-09-18 23 服务器发布完成

- 已发布源码 `a66b14764ca4dcc796700cb7c5ca7ae008d7532b`，版本保持 0.2.5；包含本次界面修复及已审上游 efe9aab1e 增量。
- GitHub CI `35251518130`（含单元、集成、lint、前端完整测试/构建及严格契约）、Security Scan `35251518007`、Aivoza Image `35252637169` 均成功且 SHA 一致。
- 镜像 `ghcr.io/aboutnb/aivoza-sub2api:sha-a66b14764ca4`；生产固定 digest `sha256:748d3873238d0057055ef5e2e132da6d40c921e7a8b6bb9ef1ee1b79b2881187`，amd64 和源码标签核验一致。
- Termius 蓝绿 prepare/promote 完成，新 flowai-app 在回环端口 3006；上一版 flowai-app-rollback-4263d8d81dca 在 3005，保留 9109c663901d 对应镜像。Caddy 两个应用上游均指向 aivoza-a66b14764ca4:8080，依赖服务启动时间未变化。
- 切换期间 aivoza.com 与 us.aivoza.com 共 60 次健康探测全部 200；登录页 200、未认证 /api/v1/auth/me 为 401，公网主 JS 为 index-BcmwBKda.js，与本次构建一致。新旧应用健康，启动日志未见迁移/校验和/致命错误。本次无 SQL 迁移，未执行真实支付或付费模型请求。
- 备份 `/root/flowai/backups/aivoza-20260918`：PostgreSQL custom dump 442 MB 且 TOC 可读取，应用/Mihomo 数据、Redis RDB、.env、Caddyfile 和旧容器元数据。回滚状态 `/root/flowai/deploy/state/aivoza-a66b14764ca4.json` 含敏感配置，不公开输出。
- 已确认更早回滚容器 0708e8bab0dc 无入站 HTTP 连接，优雅停止并移除；核对所有容器无引用后删除 c4f2525fa369 历史镜像。仅保留当前及上一版，不动数据卷、备份或依赖服务。

### 2026-09-18 界面修复与上游增量发布候选

- 本地界面修复提交 `81ecaccffffe8ed38902a0271310cf15e6497253`：签到验证码铺满、统一开关、紧凑模型行及 API 端点；保持 0.2.5。
- 审阅并合并上游 `efe9aab1e4ec89a42ba45e8dac20e882c5409a6a`，合并提交 `8763b72af0a325f17f2656429ced6b60370118e9`。包含暂停账号 OAuth 刷新、分组用量查询、协议兼容、兑换分页及界面错误修复，无新增或修改 SQL 迁移。
- 冲突处理：BaseDialog 保留 Vue UID、弹窗堆叠和 i18n；充值金额测试同时保留赠送角标及上游非法输入回退；监控测试保留 PROVIDERS.length；订单页保留 Aivoza 布局与发票，仅同步筛选重置页码。
- 自动合并核对：模型输入紧凑样式与空输入 Tab 导航均保留；注册优惠码初始化不改风险授权；OAuth 刷新移除暂停过滤不改变账号调度优先级；GM、充值赠送、-1 并发、签到结算、邮件防重和既有迁移未改动。
- 合并前完整前端 325 文件 / 2387 测试、lint、类型检查及生产构建通过；合并后的完整验证及 GitHub 镜像为发布前门禁，不以候选记录代替生产完成证据。
- 合并后完整前端 337 文件 / 2452 测试通过，lint、类型检查与生产构建通过。新增上游测试已适配现有 i18n、跳转按钮与分页响应；本机 Docker daemon 不可用，集成测试必须由 GitHub CI 的 Docker 环境验证通过才允许发布。

> 2026-09-11 主线迁移：用户已授权以 `main` 为 Aivoza/FlowAI 唯一正式主线。
> `upstream/main` 保留上游身份；旧 `sub2api-flowai` 与主题分支作为历史归档。
> GitHub 测试通过后构建 `ghcr.io/aboutnb/aivoza-sub2api:sha-<sha12>`，生产固定 digest。
> 本次要求不停机：新旧容器并存、健康验证、代理热切换、旧连接排空；不得直接执行旧的重建发布脚本。
> 下文历史快照中的分支和旧镜像仅用于追溯，不再授权旧分支发布。


> 适用分支：`main`
>
> 这不是上游项目的 release note，而是 FlowAI 分支的业务上下文和所有权记录。
> 上游合并、功能调整和正式发布都必须以本台账与
> [FLOWAI_BRANCH_CONTRACT.md](./FLOWAI_BRANCH_CONTRACT.md) 为准。

## 1. 台账目的

### 2026-09-18 控件与模型密度修复（已发布）

- 创建及编辑账号的“临时不可调度”统一复用创建分组使用的公共 Toggle，保持原有布尔值绑定，不修改调度逻辑。
- ModelWhitelistSelector 已选模型行高在桌面收紧至 28px，下拉选择项同步收紧；删除按钮保留可访问名称，触屏点击区域至少 44px，长模型名可悬浮查看。
- 渠道 ModelTagInput 空输入框默认 44px，与标准 Select 等高，不再叠加全局输入最小高度和容器内边距；模型标签紧凑展示，回车添加、删除、粘贴去重逻辑不变。
- EndpointPopover 在桌面把端点名称、地址、复制和测速放在同一行，说明保留在下一行；窄屏换行，不改变端点配置及复制、测速行为。
- 受保护路径：frontend/src/components/account/{CreateAccountModal,EditAccountModal,ModelWhitelistSelector}.vue、frontend/src/components/admin/channel/ModelTagInput.vue、frontend/src/components/keys/EndpointPopover.vue。
- 验证：7 个相关 Vitest 文件共 121 项通过，包含新 ModelTagInput 测试、账号创建编辑及主题契约；ESLint 和 vue-tsc 通过。Playwright 使用真实组件核对 320/375/768/1440/2560px 无横向溢出、模型行高 28px、渠道空输入框 44px、开关状态切换。无后端、支付、迁移或服务器变更。

### 2026-09-17 签到验证码横向自适应（已于 2026-09-18 发布）

- 顶部快捷普通签到由 compact 改为 flexible，与签到页面及幸运签到保持一致。
- 共享确认弹窗移除验证码区域的 300px 最大宽度，默认铺满内容区；小于 Cloudflare 300px 最小宽度时按可用宽度等比例缩放，窗口变化无需重建验证组件或清空 token。
- 受保护路径：frontend/src/components/checkin/CheckinShortcut.vue、frontend/src/components/checkin/LuckyCheckinConfirmDialog.vue。仅调整呈现，不涉及签到结算、验证回调、接口或数据库迁移。
- 验证：4 个相关 Vitest 文件共 29 项通过，修改文件 ESLint、vue-tsc 通过；Playwright 使用模拟 Cloudflare 最小宽度的 iframe，在 280/320/375/445/768/1440/2560px 窗口下验证组件与内容区等宽且无横向溢出。未执行真实验证码或签到请求。

### 2026-09-17 固定 main 流程与历史版本保留

- 用户确认：普通修改、修复、上游同步及发布直接复用 main 和已有工作区，不再每次创建分支。
- 分支删除必须先查职责、独有提交、PR 与本地工作区，提交清单并取得用户确认。本次 6 个临时分支均已完整进入 main，无独有提交、无待合并 PR；本地工作区不删除。
- 本次经用户确认后已清理以下 6 个远端分支，删除时核对了完整 SHA 并使用原子推送和引用保护。其提交仍在 main 历史中，可据记录恢复引用；`sub2api-flowai` 和 `sub2api-flowai-theme` 暂留历史参照。

| 分支 | 最后提交 | 已完成职责 |
| --- | --- | --- |
| feature/aivoza-theme | 8d9dac698 | 主题整合、蓝绿部署与并发保护 |
| migrate/aivoza-main | 908e9424a | 主线迁移与 GitHub 镜像门禁 |
| fix/aivoza-controls-release | 76f3f2594 | 控件布局、筛选、客服入口 |
| fix/registration-protection-release | f391efeed | 注册保护及来源配额 |
| fix/support-contacts-release | b14afc5af | 客服配置整合发布 |
| sync/upstream-v0.2.5 | 9109c6639 | 已上线 v0.2.5 的上游同步 |

- 固定顺序：更新 origin/main、拉取上游、审阅差异和迁移、在 main 合并明确 SHA、登记台账、测试、推送 main、等待同 SHA 的 CI/安全/镜像、Termius 蓝绿发布。
- CI 失败先诊断再重跑工作流，不用空提交反复触发；纯文档维护不重新替换健康生产应用。
- 23 只保留当前应用与上一个回滚版本。更早的应用容器排空后优雅停止并移除，再清理无任何容器引用的旧应用镜像；不删除数据卷、备份或依赖服务。
- 已移除 4 个历史回滚容器：81b5412354e0、14da6bdbec46、0f4b25bca7fe、36ad45263a8e；逐个检查入站 HTTP 连接为 0，使用无限等待的优雅停止，再无强制、无卷删除地移除容器。
- 已清理 31 个无任何运行或停止容器引用的旧应用镜像，包括旧 sub2api-flowai 的 0.1.178 USDT 镜像。当前 ac1658681f50 和上一版 c4f2525fa369 镜像保留。
- Docker 镜像占用从 13.28 GB 降至 3.458 GB，减少约 9.822 GB；卷和构建缓存未清理。清理后仅有当前及上一版两个应用容器，两者健康，aivoza.com 与 us.aivoza.com 的 /health 均为 200，依赖服务未重启。
- 清理前发现 guardian 仍以 http://sub2api:8080 连接最早的历史容器。已在线备份其 SQLite 到 /root/flowai/backups/guardian-before-retention-20260917.sqlite，仅把持久连接地址热更新为 https://aivoza.com；密钥和自动调度关闭状态保持不变，guardian 未重启，旧连接排空后才移除旧应用。
- guardian 原有管理密钥在旧应用、新应用和正式域名均返回 401；这是清理前已存在的认证异常，本次未更换凭据。主站健康检查、当前及上一版应用健康检查通过。

### 2026-09-17 v0.2.5 已完成发布

- 已发布源码 `9109c663901dd50be6c6cf4d1e0431b7fd8df691`，应用版本 `0.2.5`。
- GitHub CI #152、Security Scan #162、Aivoza Image #112 均成功。
- 镜像 `ghcr.io/aboutnb/aivoza-sub2api:sha-9109c663901d`；生产固定 digest `sha256:ac1658681f50ff93d3b7ef1572899eec87d11d98ecc9919eb8a66f12b957a685`。
- 23 上新容器为 flowai-app，回环端口 3005；上一个版本为 flowai-app-rollback-0708e8bab0dc，端口 3004。下次发布重新检查端口，不照搬本次值。
- 回滚状态 `/root/flowai/deploy/state/aivoza-v025-9109c663901d.json`；数据与配置备份 `/root/flowai/backups/aivoza-v025-9109c663901d`。
- 两条上游 238 迁移执行成功，既有注册保护迁移保留；新旧应用健康，依赖启动时间未变化。
- aivoza.com 和 us.aivoza.com 均返回 0.2.5，切换期间共 60 次健康探测全部 200；登录页 200，未认证用户接口 401。未执行真实付费请求或支付交易。

### 2026-09-16 上游 v0.2.5 同步候选

从 Aivoza `main` 基线 `a92d3fc2eb129ad652c876d48f2486fb2fa897b9` 合并
`upstream/main` `881f3202694c6bc932446931a30c27d9675178b9`（tag `v0.2.5`
为 `86f93c28e`，主线另有 VERSION 同步）。逐块解决 51 个冲突文件，吸收
OpenCode Zen/GO、订阅与密钥批量操作、站点计费模式、精度和 WebSocket 修复；
保留 Aivoza 主题及 i18n、账号优先级 1 最高、-1 并发拒绝、GM/EasyPay、充值
赠送、智能路由、签到与邮件防重、注册防护、订阅有效期设置及滚动部署契约。
两条上游 238 迁移与既有 `238_registration_protection.sql` 以完整文件名共存；
清理空限额行前需确认生产行数并备份。验证为完整前后端测试、lint、类型检查、
生产构建、严格分支契约、GitHub CI/安全扫描及候选镜像健康检查；正式切换前
老容器保持服务。

### 2026-09-12 客服联系方式发布

客服配置沿用 contact_info 字段，支持版本化多项名称、标签、图标、账号和链接；兼容旧纯文本配置。顶部、个人资料和兑换页统一展示，链接仅接受 http/https/mailto/tel。保留注册防护及原有后台配置，版本保持 0.2.4，无数据库迁移。验证使用客服解析与编辑器测试、主题覆盖、完整前端检查和 GitHub CI。

### 2026-09-12 注册防护发布

新增注册来源成功配额、验证失败封禁、风险账号审核与后台注册防护页面。默认开启，24 小时同 IP 成功上限 10、同 IP/UA 上限 2；成功计数与用户创建同事务提交。保留风险授权前并发 -1、现有验证码和请求限流，邮件及 OAuth 注册入口统一核验。受保护路径为 registration_protection、registration_source_quota、auth_registration、signup_risk 及对应前端 API/页面和中英文 locale。238 迁移仅新增表和索引，旧镜像可共存，不回退删除表。验证覆盖原子配额集成测试、注册/验证码/封禁单测、前端测试、类型检查及 CI。

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

1. 复用已有 main 工作区，先 fetch origin 并快进同步 origin/main。`git fetch upstream main` 后，记录 `upstream/main` 的完整 hash，然后执行 `make review-flowai-upstream`。若有待合入提交，必须先完成人工
   核对，再用 `FLOWAI_UPSTREAM_REVIEW_ACK=<完整 hash>` 重跑并取得通过结果。
2. 合并前查看 `git diff --name-status upstream/main...HEAD`，并单独查看上游新增迁移、
   调度代码、i18n 聚合入口、部署 compose 和锁文件。
3. 在 main 使用 `git merge --no-ff <已审上游完整SHA>`，并更新 `.github/aivoza-upstream-ref`。冲突时按第 6 节矩阵逐项解决，禁止用
   上游文件整文件覆盖 FlowAI 实现。
4. 合并完成后，把合并提交 hash、每个冲突区域的保留结论和验证方式写入本台账；新增
   功能提交同时登记行为、受保护路径、测试和迁移影响。
5. 先运行 `make check-flowai-contract`，再运行目标测试；发布候选最后运行
   `make check-flowai-contract-strict`。任一门禁失败都不能生成发布镜像或切换生产流量；允许只读核对生产状态。

未来新增功能条目至少包含以下信息：

```text
提交 hash / 日期：
功能和用户可见行为：
必须保留的实现路径与 i18n key：
数据库迁移及是否可回滚：
正向测试、回归测试和未验证项：
与上游同名文件的冲突处理：
```

## 1.2 Aivoza 主线迁移（2026-09-11）

- 用户授权正式主线迁移到 main；旧分支归档，未重写其历史。
- 上游锁定 `98d86915becae9fe9491a91ffc6defd5235c8d2b`，版本 0.2.4；已审新增生图修复和版本同步，无新增上游迁移。
- CI 对 main 和候选 PR 做语义契约、完整前端、后端单元/集成、lint；同 SHA 的 CI 和 Security Scan 成功后才能构建 `ghcr.io/aboutnb/aivoza-sub2api:sha-<sha12>`。
- CI 使用 `.github/aivoza-upstream-ref` 固定已审上游，避免排队期间上游移动导致不可复现；下次同步 PR 必须更新该引用。
- GitHub PR 的最终 merge commit 只有在第一父提交已包含于第二父提交、文件树与第二父提交完全相同且来自本仓库命名候选分支时，才继承候选的台账；有冲突或树变化的合并仍必须显式登记。
- 旧 Release 不再在 fork 回写默认分支；不发布上游的通用二进制替换自定义版。
- 本次尚未切换生产；需完成主题 PR、镜像构建、迁移演练、灰度健康与代理热切换。

## 1.3 主题整合和连续服务发布（2026-09-11）

- 保留 `79c83f43d` 全部主题历史；424 文件改动含首页/图表/明暗模式、无障碍控件、订阅策略与单独 SQL 迁移，未将其误当作纯 CSS。
- 12 文件冲突按功能解决：`setting_public.go` 同时保留排名隐私与订阅设置；`GroupsView` 使用新 modelAllowlist 和 Toggle；登录受公开设置守卫；充值说明保留已净化 Markdown；MiniMax 平台分支与监控缺失值语义保留。
- 后端订阅策略在开关开启时与旧版本一致；默认 true。关闭后仅取消有效期判断，不能绕过 suspended/expired 状态或日周月额度；新旧应用并存期间不得切换此设置。
- 本地完整前端 304 文件/2179 测试通过；lint、类型、生产构建通过；后端 unit/integration 通过；新增滚动并发保护测试通过。GitHub 对最终提交仍需重新验证。
- `deploy/Dockerfile` 已补充主题首页资源；生产不使用本地构建镜像。
- `deploy/deploy-aivoza-bluegreen.py` 只接受 `ghcr.io/aboutnb/aivoza-sub2api@sha256:<digest>`。prepare 不改活动路由，promote 校验前置状态并热加载 Caddy，旧容器保持运行；依赖容器不重建。旧的 force-recreate 脚本不适用于本次连续服务发布。

## 1.4 控件和筛选布局修复（2026-09-11）

- `6c56e55eaffb7ebabcc6ba5c40bcdcfbe3e4a6cf` 发布现有本地前端修复；应用版本保持 0.2.4，无后端改动和数据库迁移。
- 筛选条件使用响应式网格，操作按钮独立排列；图表切换和 API key 开关使用统一控件，保留原事件和数据语义。
- DateRangePicker 弹层挂载到 body，按视口定位，保留 Escape、Tab 和关闭后焦点恢复；相关交互回归测试必须通过。
- 顶部社群入口保留在常用操作区，客服入口读取既有 contactInfo，通过弹窗展示和复制；兑换页面去除硬编码联系地址。
- 控件边框、桌面密度和键盘焦点统一；触摸按钮保持独立尺寸规则。必须保留 design-system/aivoza-flowai 文档及主题、日期和社群回归测试。
- 验证：完整前端 Vitest、lint、typecheck、生产构建和严格分支契约；正式镜像由 main 的 GitHub CI 与 Security Scan 成功后构建。

## 2. 最近一次生产发布快照

| 项目 | 值 |
| --- | --- |
| 开发及发布分支 | main |
| 核对日期 | 2026-09-21（Asia/Shanghai） |
| 已发布提交 | 5d6884aa6f7fa7268909ee0f99253d71d8a2b5ce |
| 已审上游基线 | 1a9d49e16f7a22c432b428fce4af8d731f1fa364 |
| 应用版本 | 0.2.7 |
| 发布镜像 | ghcr.io/aboutnb/aivoza-sub2api:sha-5d6884aa6f7f，部署固定 digest |
| 目标与保留规则 | 23 蓝绿部署，仅当前版本及上一个回滚版本 |

以下 0.2.3 表格及后续旧发布记录保留用于历史追溯，不能用作当前发布指令。

### 2.0 2026-09-09 历史快照

| 项目 | 值 |
| --- | --- |
| 发布分支 | `sub2api-flowai` |
| 本次核对日期 | 2026-09-09（Asia/Shanghai） |
| 业务代码基线 HEAD（本次上游合并前） | `04661fec1888ebdf93ab18d2736f9ab3329687b0` |
| 本次核对 HEAD（上游合并后） | `76b125bb16a5e5baf72e384f9372f0430e2d7f25` |
| FlowAI 0.2.3 版本提交 | `04661fec1888ebdf93ab18d2736f9ab3329687b0` |
| 最后已审并合入的上游基线 | `upstream/main` = `270eac6973049fe1b50eb75560a74a029e82884c`（0.2.3） |
| 上游合并提交 | `fa931c3458f91bb8cdc865d0ed77f791d22d34cf` |
| 应用版本 | `0.2.3` |
| 相对上游的非合并提交 | 当前业务修复提交后 114 个，其中 17 个治理提交按受限规则动态豁免 |
| 相对上游的文件差异 | 业务修复提交后 502 个文件，约 55002 行新增、1984 行删除 |
| 发布镜像 | `ghcr.io/aboutnb/sub2api:sub2api-flowai-<sha12>` |
| 生产发布目标 | 23 服务器，使用预构建镜像 |

快照值会变化，当前行为契约不会因上游版本号变化而自动变化。下一次发布必须重新
记录 HEAD、上游基线、版本、镜像 digest 和实际验收结果。

### 2.1 FlowAI 0.2.3 上游合并记录（已完成）

2026-09-09，GitHub API 确认 `upstream/main` 已更新到
`270eac6973049fe1b50eb75560a74a029e82884c`，已通过 SSH over 443 抓取并完成预审。
上游相对合并基点新增 260 个提交，已由明确的非快进合并提交
`fa931c3458f91bb8cdc865d0ed77f791d22d34cf` 合入；本地 `main` 未修改。

- 版本提交 `04661fec1` 已将 `backend/cmd/server/VERSION` 更新为 0.2.3；合并时必须保留
  上游同一版本的状态，不能用本地版本提交跳过上游代码审阅。
- 合并保留账号 priority、并发、GM/EasyPay、充值赠送、签到、邮件、i18n 与部署边界；
  账号优先级仍为 1 最高、数值越小越优先。
- 上游新增迁移按完整文件名保留：232/233 upstream request id、234 reasoning multiplier、
  234 Codex manifest、235/236 model allowlist、237 MiniMax platform；与 FlowAI 同编号
  迁移并存，未编辑已执行迁移。
- 冲突处理：网关路由、上游错误归因、WebSocket 失败事件、设置公开字段、wire 注入、支付
  页面和账号创建请求头均按功能块合并；旧 `ModelsListConfig` 已适配为上游
  `GroupModelAllowlist`，没有恢复已删除的旧 schema。
- 验证：后端 handler/server/service/cmd 测试通过；前端 Turnstile 与 PaymentView 25 项测试
  通过；完整契约、全量测试和 GitHub Actions 仍是发布前置门禁。
- 合并后的兼容提交 `af997094b` 仅调整上游路由测试断言、补齐 Pinia/API mock 夹具，以及补充
  英文账号管理表单/筛选 i18n key；不改变运行时调度、支付、并发或迁移语义。后端 `go test ./...`
  通过；前端 286 个测试文件、2091 项测试通过，`lint:check`、`typecheck` 和生产构建通过。
- 后续兼容提交 `76b125bb1` 仅补齐上游新增 `AdminService` 接口、`NewAdminService` 和
  `NewProxyHandler` 参数对应的单元测试夹具；`backend make test-unit` 已通过，不改变运行时行为。
- 仍只允许使用 GitHub Actions 为 `sub2api-flowai-<sha12>` 生成的不可变 GHCR 镜像；23
  服务器仅重建 `flowai-app`，不重建任何依赖服务。

### 2.2 本次上游合并审阅（已完成）

2026-09-02 刷新并审阅 `upstream/main` 完整提交
`5097b31457e6dc9f49e5f5c9c72b925ce79543b3`（0.2.0）。本次上游相对上一已审基线
`0d27f45ead1b58908548ec21afd923ecaf7339bc` 共新增 53 个提交，以
`88ad9b765e9a3fbeb7ea3265c9ef1f2874d4c037` 合并进入 `sub2api-flowai`。

- 上游的 53 个提交已完整进入一个明确的合并提交；本地 `main` 未修改。上游相对该基线实际
  变更 125 个文件（新增 4008 行、删除 433 行），没有 incoming 删除。新增的 4 个 SQL
  文件使用独立完整文件名保留，与 FlowAI 的 `232_checkin_turnstile.sql` 不冲突：
  `232_channel_cache_write_1h_pricing.sql`、`232_group_force_openai_fast.sql`、
  `232_group_reasoning_effort_over_limit.sql` 和 `233_group_free_openai_fast.sql`。
- 上游 0.2.0 的分组模型定价、按模型推理强度策略、OpenAI Fast/free Fast 控制、缓存定价、
  API key 缓存身份、WebSocket/Responses 转发和模型映射修复已合入；相关 Ent、handler、
  service、前端管理页和中英文 locale 均按 key/逻辑块合并。
- 合并没有文本冲突，也没有使用上游整树覆盖。重叠路径（Ent group、网关、DTO、迁移测试、
  dashboard/overview locale 等）已逐项核对；`deploy/README.md` 同时保留 FlowAI 预构建镜像/
  固定数据目录说明和上游数据库启动恢复说明。
- 提交 `213b3fcf7` 的签到 Turnstile、硬性批量防护和未活跃用户召回广播，以及
  `4dc03354a` 的普通/幸运签到确认弹窗校验，均在合并前独立提交并登记；邮件
  `sent`/`sending` 防重、签到幂等和迁移 checksum 约束没有改变。
- 账号调度仍为 priority 升序，即 1 最高优先级；`-1` 并发拒绝、风险注册、GM/EasyPay、
  充值赠送、签到、邮件、i18n 和独立部署边界未被上游覆盖。
- 本次发布候选提交 `beed8a2f4` 增加 API key 智能路由和渠道监控 V3 展示层；该提交未改变
  账号优先级、用户并发、支付结算或历史邮件/签到状态机。

合并前预审确认命令：

```bash
FLOWAI_UPSTREAM_REVIEW_ACK=5097b31457e6dc9f49e5f5c9c72b925ce79543b3 \
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
| API key 智能路由 | 默认保持单分组兼容模式；显式开启 `smart_routing_enabled` 后支持最多 20 个同平台/计费类型候选组，按价格、速度、成功率或自定义权重评分，并保留速率护栏 | `backend/internal/service/smart_route.go`、`backend/internal/server/middleware/smart_route_resolver.go`、`SmartRouteEditor.vue` | `smart_route_test.go`、`smart_route_resolver_test.go`、API key handler 测试 |
| 渠道监控 V3 | V2 后端聚合和权限边界不变；V2 模式默认使用紧凑卡片视图，可用 `monitor_view=v2` 回退完整 V2 页面 | `ChannelStatusV3View.vue`、`ChannelMonitorV3Card.vue`、`ChannelMonitorV3Timeline.vue` | `ChannelStatusView.mode.spec.ts`、V3 组件测试、`monitorFormat` 测试 |
| Project Mihomo | 支持 URL/静态源、多源、兼容请求头、节点测速、筛选、自动路由、多 listener 和账号池分配；状态持久化 | `backend/internal/service/project_mihomo_service.go`、admin handler、`deploy/docker-compose.preview.yml` | `project_mihomo_service_test.go`、admin handler 测试 |
| 账号导入/批量测试 | 导入可指定分组/代理池，返回新账号集合；批量测试需人工点击开始，不能自动误发请求 | `backend/internal/handler/admin/account_data.go`、`BatchAccountTestModal.vue`、`AccountsView.vue` | `data-import.spec.ts`、`BatchAccountTestModal.spec.ts`、账号测试 i18n 测试 |
| USDT/发票 | 独立 GM 通过 EasyPay 接入；USDT 最低金额默认 50；GM 保持精确金额匹配且不做网络费补偿；checkout 弹窗为 625x900；余额充值默认 50 赠 5%、100 赠 10%；BEpusdt 专用代码已移除；发票流程保持不变 | `payment_config_service.go`、`payment_recharge_bonus.go`、`payment_config_limits.go`、`payment_order.go`、`providerConfig.ts`、`rechargeBonus.ts`、EasyPay provider、`invoice_service.go` | USDT minimum/method limit、recharge bonus、popup、EasyPay custom method、invoice service 和支付 API 测试 |
| 签到/奖励 | 每用户每业务日最多结算一次；普通/幸运模式、概率、倍率/固定金额、阶梯、精度和未充值策略均受事务与风控约束；可单独启用 Turnstile，硬性同源/指纹防护不能被后台开关关闭 | `backend/internal/service/checkin_service.go`、`checkin_record_repo.go`、`TurnstileWidget.vue` | `checkin_service_test.go`、Turnstile widget、精度/安全/handler 测试 |
| 邮件广播 | 受众和模板快照持久化；`FOR UPDATE SKIP LOCKED` 领取；发送结果不确定时不自动重试；召回广播可按未活跃天数固定受众快照 | `email_broadcast_repo.go`、`email_broadcast_service.go` | repository integration tests、service tests、email broadcast API tests |
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
- 余额充值赠送档位由 `RECHARGE_BONUS_TIERS` 配置，默认 50 赠 5%、100 赠 10%，空数组
  关闭。按充值本金达到的最高门槛计算，手续费不参与赠送；最终到账金额在创建订单时固化，
  前端角标和预览只展示后端公开的同一档位配置。
- XZNOAuth 发票流程保存本地 `invoice_applications` 状态，覆盖草稿、校验、税费、
  申请、取消和 PDF 下载；client secret 只能留在服务端。
- 2026-08-29 起，23 服务器停止生产 BEpusdt；Sub2API 中专用 quote/order/reconcile/webhook
  代码已移除，`206_add_usdt_payments.sql` 保持原样用于历史数据兼容，不代表生产入口仍启用。
- GM 是独立项目：使用 `gm-epusdt` 分支的不可变 GHCR 镜像，独立 Compose、容器、配置
  和数据目录；域名由 Caddy 直接代理到 GM，不在 BEpusdt 目录上改造。
- Sub2API 使用普通 EasyPay provider 实例接入 GM。USDT 自定义方式需要把前台方式映射为
  GM 可用 selector，例如 `usdt_trc20 -> usdt.tron`，并使用 EasyPay 回调
  `/api/v1/payment/webhook/easypay`。
- GM checkout 使用 `popup` 模式时，桌面端在支付按钮点击事件内预打开窗口，订单创建后
  导航到 GM checkout；父页面立即并持续轮询本地订单状态，只有验签回调完成服务端入账后
  才显示成功。移动端、二维码/路由方式和恢复流程不自动创建窗口。
- 功能变更（2026-08-30）：Sub2API 新增 `payment_usdt_min_amount` 管理配置，后端键
  `USDT_MIN_RECHARGE_AMOUNT` 默认 50，`0` 可关闭；method limits 与创建订单后端共同拦截。
  GM checkout 桌面弹窗宽度由 1250 缩为 625，高度保持 900。
- 2026-08-31 取消 GM 网络费/少到账处理：`gm-epusdt` 恢复原有精确金额匹配、最小精度
  递增和严格 `chain_tokens.min_amount` 判断，不保留 `system.usdt_underpayment_tolerance`。
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
- 可编辑风控关闭时仍执行硬性批量保护：10 分钟同源最多 5 个用户、24 小时同指纹最多
  1 个用户；管理员配置只允许收紧窗口或人数，不能放宽到硬下限之外。
- `checkin_turnstile_enabled` 默认关闭。启用后，新签到结算必须提交并服务端验证 Turnstile
  token；密钥未配置或验证依赖异常时不结算。当天已完成的幂等请求不重复消费 token。

### 4.7 邮件广播和通知

- 任务创建时固定受众、变量和中英文模板快照；发送状态由数据库驱动。
- 领取必须使用行锁和 `SKIP LOCKED`，状态路径为 `pending -> sending -> sent/failed`。
  已发送记录不能重置成 `pending`。
- SMTP 已接受但客户端得到错误时结果不确定，不能自动重发；`sending` 残留需先查
  provider/SMTP 证据再人工处理。
- 通知邮件模板、维护公告、登录/注册邮件和公告弹窗均属于用户可见功能，不能只合并
  后端发送逻辑而丢掉模板或 locale。
- `system.reactivation` 召回模板使用独立中英文官方模板；`inactive` 受众以最后活跃、最后
  登录、创建时间的首个非空值计算未活跃天数。任务创建后仍使用固定收件人/模板快照，
  不允许因受众类型不同而重发 `sent` 或结果不确定的 `sending` 收件人。

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

### 4.10 API key 智能路由与渠道监控 V3

- API key 未配置智能路由时继续使用原有单分组路径；智能模式只有在
  `smart_routing_enabled=true` 时可创建或更新，设置默认值为 `false`，因此本次发布不会
  自动改变现有 key 的路由行为。
- 智能模式候选组最多 20 个，必须属于当前用户可用的活跃非 composite 分组，且平台和
  计费类型一致。支持 `auto`、`price`、`speed`、`success` 和 `custom` 策略；自定义权重
  必须为 0-100 且总和为 100。速率护栏拒绝无效或超限倍率，不能用猜测值绕过价格约束。
- 候选组会同时检查订阅额度、图片能力、账号可调度性和请求模型支持；没有可用候选时
  fail closed。历史图片任务/批量任务读取按用户与 API key 所有权及已记录分组解析，不能
  因智能路由配置删除或缓存失效而使历史结果不可读。
- 智能路由请求只覆盖已声明的文本、消息、Responses、聊天、嵌入、Gemini 内容和图片
  入口；视频、音频、实时、搜索及其他未支持端点必须返回协议兼容错误，不得静默改路由。
- 渠道监控 V3 只替换用户端展示层，复用现有 V2 API、权限裁剪、聚合器和 locale；V2 模式
  默认显示 V3，`?monitor_view=v2` 仅回退展示，不绕过后端模式守卫。V1 模式和 V2 后端
  数据管线保持不变。
- 本次新增迁移为 `234_api_key_smart_routing.sql`，只新增智能路由配置表、候选组表和
  默认关闭的设置；已执行迁移仍按 checksum 保护，回滚应用镜像不逆向删除数据。

### 4.11 构建与发布

- `.github/workflows/preview-image.yml` 只从通过同 SHA CI 和 Security Scan 的 main
  构建 linux/amd64 镜像 aivoza-sub2api:sha-<sha12>，服务器固定对应 digest。
- 23 服务器通过 `deploy/deploy-aivoza-bluegreen.py` prepare/promote 连续服务发布；
  PostgreSQL、Redis、Mihomo、GM、Caddy 和 /app/data 不得被重建、删除或重置。
- 验收后仅保留当前及上一个回滚版本，更早的容器排空后清理；旧应用镜像无引用才删除。
- 服务器不得作为正式发布构建机执行 `docker compose build`、`pnpm build` 或 `go build`。
  回滚只切换已验证的镜像 tag；数据库迁移不能用回滚镜像盲目逆向。

## 5. FlowAI 迁移台账

以下文件是当前分支相对上游新增、修改或必须保护的迁移。数字前缀可能与上游其他功能
重复，必须以完整文件名识别，不能因为编号相同而覆盖或删除。

<!-- FLOWAI_MIGRATION_LEDGER_BEGIN -->

| `backend/migrations/245_gemini_cached_contents_mutation_lease.sql` | v0.2.16 上游分组译文/显式缓存 | 新增完整文件名，既有同编号迁移不改写；生产备份后执行 |

| `backend/migrations/244_gemini_cached_contents_channel_usage.sql` | v0.2.16 上游分组译文/显式缓存 | 新增完整文件名，既有同编号迁移不改写；生产备份后执行 |

| `backend/migrations/243_gemini_cached_contents.sql` | v0.2.16 上游分组译文/显式缓存 | 新增完整文件名，既有同编号迁移不改写；生产备份后执行 |

| `backend/migrations/242_group_i18n.sql` | v0.2.16 上游分组译文/显式缓存 | 新增完整文件名，既有同编号迁移不改写；生产备份后执行 |

| `backend/migrations/256_add_support_tickets.sql` | 站内工单与消息表 | 新增完整文件名，保留其他同编号迁移；外键、索引和作者约束 |
| `backend/migrations/255_remove_retired_help_clients.sql` | 清理 Read Frog/Roo Code 教程 | 新增；删除两类教程及对应 slug，发布前备份，可从备份恢复；不改历史种子与同编号监控迁移 |
| `backend/migrations/242_drop_platform_check_constraints.sql` | 上游平台约束转应用校验 | 新增；配额和复合路由平台由应用/Ent 目录验证，监控主动探测约束不动；回滚不逆向恢复旧白名单 |
| `backend/migrations/256_channel_monitor_v0215_platforms.sql` | v0.2.15 被动监控 Cline/Command Code | 新增；仅追加缺失项，保留禁用/模型/分组/空配置，默认 14 维度；不改 255 |
| `backend/migrations/255_channel_monitor_platform_catalog.sql` | 被动监控新增平台目录补齐 | 只追加六个缺失平台并更新版本；保留禁用状态、模型、分组及显式空列表，幂等；旧迁移不改写 |
| `backend/migrations/241_add_payment_order_bonus_amount.sql` | 支付免费额度字段 | 新增文件；旧订单默认 0，旧应用兼容，不逆向删列 |
| `backend/migrations/241_add_typesafe_platform.sql` | TypeSafe 平台约束 | 新增文件；允许平台集合扩展，旧应用兼容 |
| `backend/migrations/240_help_documents.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/241_help_client_guides.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/242_help_desktop_clients.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/243_help_extended_clients.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/244_help_protocol_specific_guides.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/245_help_channel_protocols.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/246_help_gui_client_guides.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/247_help_translation_client_guides.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/248_help_zcode_window_guide.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/249_help_tavern_clients.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/250_help_tutorial_review.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/251_help_faq_troubleshooting.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/252_help_quick_start.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/253_help_api_probe.sql` | 教程表或内置教程种子增量 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/254_smart_route_cross_platform.sql` | 托管工作台跨平台路由字段与唯一索引 | 新增文件；保留已有同编号迁移，发布前备份，不逆向删除 |
| `backend/migrations/239_image_studio_keys.sql` | 图片工作台专用密钥 purpose 字段和有效用户/分组唯一索引 | 仅新增，普通密钥默认 standard；回滚应用前关闭工作台，不逆向删除字段或密钥 |
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
| `backend/migrations/231_add_usage_log_native_compaction_v2.sql` | 标记原生 OpenAI remote compaction v2 请求 | 本次上游 0.1.184 新增；默认 false，不改历史请求状态 |
| `backend/migrations/232_checkin_turnstile.sql` | 签到独立 Turnstile 开关 | 默认 false；只新增设置，不覆盖现有全局 Turnstile 配置或签到状态 |
| `backend/migrations/232_channel_cache_write_1h_pricing.sql` | 上游 0.2.0 渠道缓存写入定价 | 与同编号 FlowAI 签到迁移按完整文件名并存，执行前核对 checksum |
| `backend/migrations/232_group_force_openai_fast.sql` | 上游 0.2.0 分组 OpenAI Fast 强制策略 | 新增字段迁移，不改 FlowAI 账号优先级或并发语义 |
| `backend/migrations/232_group_reasoning_effort_over_limit.sql` | 上游 0.2.0 分组推理强度超限策略 | 新增字段迁移，按完整文件名执行并保留历史迁移不可变性 |
| `backend/migrations/233_group_free_openai_fast.sql` | 上游 0.2.0 分组免费 OpenAI Fast 策略 | 新增字段迁移，不覆盖支付/充值赠送配置 |
| `backend/migrations/234_api_key_smart_routing.sql` | API key 智能路由配置、候选组和功能开关 | 新增表和 `smart_routing_enabled=false` 默认值；只追加迁移，回滚镜像不得逆向删除 |
| `backend/migrations/232_add_usage_log_upstream_request_id.sql` | 上游请求 ID 记录 | 上游 0.2.3 新增；只追加字段，不改历史迁移 |
| `backend/migrations/233_add_usage_log_upstream_request_id_index_notx.sql` | 上游请求 ID 索引 | 上游 0.2.3 新增；按完整文件名执行并核对 checksum |
| `backend/migrations/234_channel_max_reasoning_effort_multiplier.sql` | 渠道推理强度倍率 | 上游 0.2.3 新增；不改变 FlowAI 账号调度语义 |
| `backend/migrations/234_group_codex_models_manifest_config.sql` | Codex models manifest 配置 | 上游 0.2.3 新增；保留与 FlowAI 智能路由的兼容处理 |
| `backend/migrations/235_group_model_allowlist.sql` | 分组模型白名单 | 上游 0.2.3 新增；替代旧模型列表配置并由新 Ent 字段承载 |
| `backend/migrations/236_group_model_allowlist_repair.sql` | 旧模型列表字段修复迁移 | 上游 0.2.3 新增；只做兼容性迁移，不删除生产数据 |
| `backend/migrations/237_add_minimax_platform.sql` | MiniMax 平台支持 | 上游 0.2.3 新增；平台列表和迁移按完整文件名保留 |
| `backend/migrations/235_subscription_expiration_enabled.sql` | 新增 | 仅在不存在时写入 true，保留既有配置，旧版本忽略此设置；不修改已执行迁移 | 迁移及订阅策略测试 |
| `backend/migrations/238_registration_protection.sql` | 注册来源、风险事件、封禁及风险账号表 | 新增结构；旧实例兼容，不做逆向删除 |
| `backend/migrations/238_opencode_go_platform.sql` | OpenCode GO 平台约束 | 新增平台值，保留 MiniMax；与同编号注册防护迁移按完整文件名分别执行 |
| `backend/migrations/238_purge_unlimited_user_platform_quotas.sql` | 清理三档限额均为空的记录 | 上线前确认受影响行数并备份；仅删除等价于不限额的无效记录，不改已有额度 |
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
通过”替代业务语义审阅。当前没有未完成的上游待审快照。

## 7. 历史提交索引（非合并提交）

下面的索引覆盖当前快照中相对 `upstream/main` 的全部 112 个非合并提交，其中 17 个治理
文档提交按上面的受限规则动态豁免，但仍会被路径检查；脚本会逐个检查功能提交 hash
是否存在于标记区，新增代码提交未登记时，CI/发布门禁失败。

<!-- FLOWAI_LEDGER_NON_MERGE_BEGIN -->

| 2026-10-26 | `f8daec67f` | 修正 v0.2.18 部署镜像 Go builder 为 1.27.3，匹配 backend/go.mod | 镜像构建修复 |

| 2026-10-26 | `b65d97539` | 对齐 v0.2.18 新增管理员服务/代理处理器/支付路由参数的 unit 测试夹具；`go test -tags=unit ./...` 全量通过 | CI 修复 |

| 2026-10-26 | `c83a78b88` | 图片工作台 Markdown 依赖安全升级：streamdown/math、KaTeX、DOMPurify 锁定修复版本并通过生产 audit/599 项测试 | 安全修复 |

| 2026-10-26 | `823e1ca7d` | 登记 v0.2.18 上游合并候选、版本同步与本地验证结果；待 GitHub 门禁和 23 服务器发布 | 发布候选台账 |

| 2026-10-17 | `d4d3da27dcb0be57b5f3d601ebfaad5cdab0a2c0` | 公开价格按官方目录与分组倍率报价，图片独立倍率及按次两位小数；保留公开范围、缓存和鉴权边界 | 公开价格修正 |

| 2026-10-16 | `3ae19569fb96ddf58e2e7047b917bf877130ac1a` | 记录 v0.2.17 生产镜像、蓝绿状态、备份、探测和线上验收证据 | 发布证据 |

| 2026-10-16 | `67054ae3170f7cfebc9238e1589d5e9d280a7c92` | 最新生图渠道/优化开关/取消与提交保护、用户隔离桥接、原子路由初始化及公开价格 API；Wire 重生成 | 生图/价格接口 |
| 2026-10-16 | `cec41bb961868af820cd08a41124f5b4196b96e6` | 清理导入生图源码尾随空白，不改变行为 | 格式 |
| 2026-10-16 | `b0a02a3304a0a78cdd003ddda1dd23ba02c03084` | 已审上游 v0.2.17：chore: sync VERSION to 0.2.17 [skip ci]；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `71e6b6a9717064b61826d7e5bc480a5f4b50d9cf` | 已审上游 v0.2.17：test(responses): 模型改写转发测试夹具补齐上游 created_at；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `f2e5b55f3e8069df1c02a9a27af537b5fdefa64e` | 已审上游 v0.2.17：test(responses): 模型改写转发测试夹具补齐上游 created_at；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `bdfe215cbb06f5089112873b14c34c538fc37118` | 已审上游 v0.2.17：fix(responses): 网关送出的 Responses 对象一律带 created_at，修复 grok-build 反序列化失败；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `1b92038eaf60674e622dcd90d13d79642c422a6f` | 已审上游 v0.2.17：fix(payment): fix EasyPay order queries and confirm notifications before crediting；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `86f844304591800be66bafa4454c4a5bc8e02514` | 已审上游 v0.2.17：fix(openai): evaluate retirement after public alias projection；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `f37b69b77934ab2479b41178ef65ea337cc18041` | 已审上游 v0.2.17：fix(payment): treat Alipay refund query errors as pending and delay not-found failure；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `83337a0c6547037fbd122a6dcf0f6cc71679db08` | 已审上游 v0.2.17：fix(payment): query Alipay refund status and close unsupported order types；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `44e35fbf844059c9937c8d910f3da5fdfe0810dc` | 已审上游 v0.2.17：fix(payment): finalize failed pending refunds atomically；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `f996b0bfaae10d0894aaf7edced956805ad4229f` | 已审上游 v0.2.17：fix(payment): load Stripe client secret from the local order snapshot；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `7c6f70680d76faf2b8e0a2df3669b8966c410f97` | 已审上游 v0.2.17：fix(payment): validate order types and keep refund deductions while pending；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `26aab6dc076059efb919d8b88ce33b2b27fde246` | 已审上游 v0.2.17：fix(openai): respect GPT-5.5 retirement on ChatGPT upstreams；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `c2cca510f2dc46f400bd8f9dcc4a5de55859cf51` | 已审上游 v0.2.17：revert(accounts): restore inline account filters (#7853)；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `64211de8defd7144c1e99f72770278adeae3c2a5` | 已审上游 v0.2.17：feat(gemini): 兼容入口转发思考档位并流式返回思考摘要；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `9d3cc4507ff773e8f52e7340f20a5ffcfe8df22f` | 已审上游 v0.2.17：fix(antigravity): keep legacy default passthroughs and bill Claude 5.5 tier IDs at base rates；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `d19265ddb218c79e951011e6dec54fe17a215223` | 已审上游 v0.2.17：fix(gateway): 拒绝含 OpenAI 网关 OAuth 的 composite 歧义路由，防止凭据误入 Anthropic 管线；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `1774b79dcbf9ba25a007772c015d29758b7fdbc2` | 已审上游 v0.2.17：fix(zhipu): 修正 GLM Coding Plan 5h/weekly 用量窗口标反；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `92c9433d40db69b00f8e158fc573e2b796c4a678` | 已审上游 v0.2.17：chore: update sponsors；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `feab6d80dc09fa2d64894d20cdca64fb4e460931` | 已审上游 v0.2.17：fix(antigravity): preserve tool_result images, native Gemini functionCall/thoughtSignature, and composite /v1beta routing；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `2a8cb088af205586b514a27f6a63e9bcf7adb4e6` | 已审上游 v0.2.17：feat(apicompat): configure minimum output token floor；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `373d43533a0fa06e1993af1eedd61154db8ebfb0` | 已审上游 v0.2.17：feat(antigravity): support Gemini 4 Argon and accurate upstream model catalogs；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `16611a8ee0b459ee002d50efb5a90d9d06d92ff0` | 已审上游 v0.2.17：fix(openai): isolate Codex imports by workspace；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `7de15113c77cf6d3c431b1840c6997d561cc634d` | 已审上游 v0.2.17：fix(openai): exclude OAuth accounts from web search emulation；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `26e55cf332e73ae25ae0e17f1e0ef90bcbf03e54` | 已审上游 v0.2.17：fix(openai): emulate web_search on /v1/messages for OpenAI API Key accounts；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `867563ae76f3ff99ed20dfed93ee216a7939ce3f` | 已审上游 v0.2.17：fix(gateway): honor Anthropic protocol base URLs；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `c6c9935ae37035ba69555701e2fb103ab4993f5e` | 已审上游 v0.2.17：feat(gemini): 补齐 Gemini 4 Argon 计费；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `14cd5871752dabfc38069e9a7d87f282da0c9eda` | 已审上游 v0.2.17：fix(gemini): 每个上游错误响应只检查一次错误策略；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `20c4bf5ae59e4cb94669165436bb57ee16c41cb3` | 已审上游 v0.2.17：fix(gemini): Vertex SA 已不存在时账号置为 error 并换号；随 0477bb8bf 合并 | 上游同步 |
| 2026-10-16 | `89ad4c2caa0a2c9f5a393f9d1c9c7e3c73f7c3b3` | 已审上游 v0.2.17：fix(gemini): Vertex 共享容量不足的 429 不冷却账号，改为同号退避重试；随 0477bb8bf 合并 | 上游同步 |

| 2026-10-14 | `c76322ded955b66ad10063cf168c430d5a15abbb` | 工单 PostgreSQL 测试按外键顺序清理消息、工单和用户；避免新非级联外键阻断后续用户/订阅套件清理 | 工单集成测试隔离 |

| 2026-10-14 | `b219ac6b4675e4e7c82ea82907cd990ee957367b` | 工单上限 metadata 断言检查类型转换 | 工单测试 |

| 2026-10-14 | `29eecc52dc734a22fd1e58379b07986074dca312` | 用户工单列表及回复断言检查实际响应结构 | 工单测试 |

| 2026-10-14 | `baff4612d6217f54e72d33716a97cc0257330e0f` | 补齐用户与管理员工单剩余响应类型检查，保持实际响应断言 | 工单测试 |

| 2026-10-14 | `97a9ba465d775896c99f9afab90acfa2f9509704` | 工单 SQL rows 关闭返回值按仓储惯例处理，管理员测试断言检查类型转换；针对仓储/管理员重新测试及完整 lint | 工单发布质量门禁 |

| 2026-10-14 | `8ac38996b738dfa2263c01fe4a58bf44c17e110e` | 工单主题覆盖登记、运营移动卡片与桌面键盘滚动区并存；监控测试验证动态目录默认关闭、配置保留及失败恢复，含 Kiro 维度 | 工单/监控/运营主题回归 |

| 2026-10-14 | `9c2cca9e627291bc719bada3d57de23b625a4180` | 已审上游 v0.2.16：chore: update sponsors；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `4d9ac3e482106d2da3a39fe3e546d8d6801f1cc1` | 已审上游 v0.2.16：chore: sync VERSION to 0.2.16 [skip ci]；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `b669dc6daf1a9909b02ecf1f9cb692d88875e086` | 已审上游 v0.2.16：test(frontend): 将模型白名单回归用例纳入 CI；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `c4bdc484753acc95f9a775854af303146389286c` | 已审上游 v0.2.16：fix(haiku): 修正采样边界、默认模型配置与 lint；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `26f70c6a873bbeed126f1944068a7eb183c89f16` | 已审上游 v0.2.16：fix(gateway): wire Claude Haiku 5.5 into Responses bridge；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `cfbbfd6b3b41bbdeee748e25d8aa3de866b811c7` | 已审上游 v0.2.16：fix(antigravity): isolate image model rate limits；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `43a3c197cb7663582ad1b6b21c761b84776193c9` | 已审上游 v0.2.16：fix(gemini): 显式缓存恢复流程区分账号已删除与暂不支持缓存，上游必已过期后作废意图；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `2023c290e1bbfa63e3f05bdddeeae1f264b8e637` | 已审上游 v0.2.16：fix(account): 计数、Bedrock 与重试耗尽入口把上游模型交给错误处理；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `492c63059980ea5797cc2c34ba642ffee9a56aee` | 已审上游 v0.2.16：fix(gemini): satisfy test lint；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `ba2a71733272b75e948c28a6fb41ffa506a40edc` | 已审上游 v0.2.16：fix(gemini): preserve images from tool results；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `8b74741efe35a42c67bbb35b821af9a12aca2c24` | 已审上游 v0.2.16：fix(groups): 广场 Vary 保留 Origin，分组译文语言代码统一大小写；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `fc82e831026a15eaf609cc1f7cd4d4a3690e5cad` | 已审上游 v0.2.16：fix(account): 显式模型级 401 规则在网关故障转移入口生效；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `26b31a83e4190a3b38bce112fc597defbfb9515b` | 已审上游 v0.2.16：fix(gemini): 显式缓存修改串行化、失败回滚，确定金额费用按 Key / 订阅预留；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `43768680d38ccdbd8fffc2cb91d3c1fc778a61c6` | 已审上游 v0.2.16：fix(billing): 取消 GPT-6.1 Sol 订阅长上下文附加费；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `07185812655ec2e0740ca76882879c0b9ae5ee0a` | 已审上游 v0.2.16：fix(codex): raise GPT-6.1 Sol generated context default to 500k；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `daebb7a80c720763a01e5cef8799c030aed0b058` | 已审上游 v0.2.16：fix(gemini): 模型列表与模型详情接口支持 Agent Platform (Vertex) 账号；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `4426593e9d7dc38a009f28da038ac2906a57c17e` | 已审上游 v0.2.16：fix(gemini): 3.6 Flash 兜底价改为官方优惠价；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `f0ee427f0ad2f7eab6ab8f095ba05041af3a48f8` | 已审上游 v0.2.16：feat(gemini): 显式缓存存储价内置官方价并按 Pro / Flash 系列兜底；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `05d7946af84228589e5a6a8d8b6bc08d686ba778` | 已审上游 v0.2.16：fix(gemini): 显式缓存错误透传收窄到上游网关账号，延长沿用创建时计费口径；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `c9e4b267b6fac1f00060eab983e606f7a2b2d4fc` | 已审上游 v0.2.16：fix(account): 同步上游模型失败时显示后端返回的错误原因；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `6f9b02f12c3d414ed324460a0e82d7793a471f72` | 已审上游 v0.2.16：feat(gemini): Vertex 服务账号支持同步上游模型列表；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `1c68f9996df8e0bd7933d94d4ca826bc3dc570c7` | 已审上游 v0.2.16：feat(gemini): 适配 AI Studio AQ. 格式 Key；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `2588675ce7c0fa13189379489cc4edabe48f6fcd` | 已审上游 v0.2.16：fix(gemini): 第三方上游 API Key 与 Vertex 不再套用 AI Studio 等级和模拟配额；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `81e6ec5346c645fa03d794495198e32e4b910598` | 已审上游 v0.2.16：feat(gemini): Vertex 账号支持 API Key 认证；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `64331a97921bd5a43e7b54ba6a517476c985e8e5` | 已审上游 v0.2.16：ui(gemini): Vertex 显示名改为 Agent Platform (Vertex)；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `3abd64658297f5ec47ad09120a0a26d6081d42de` | 已审上游 v0.2.16：fix(ops): Gemini 原生路由的网关本地错误按错误类型归类；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `d5a802f84a6ae3405bee2a6e0160675785abf237` | 已审上游 v0.2.16：feat(gemini): 支持显式上下文缓存（AI Studio / Vertex / 自建网关上游）；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `3bfd40f5afd820ff7fae32f3f26c75fe3e0dd504` | 已审上游 v0.2.16：fix(gemini): 引用缓存不存在的 403 不再停用账号并换号；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `0f189af8de8374cca7ddd4095a39135fe4725814` | 已审上游 v0.2.16：fix(groups): ignore stale RPM override loads；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `ce6242fca1ef84da4855a92cd80117028b3a99c0` | 已审上游 v0.2.16：fix(quota): recognize international coding endpoints；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `66b7ce08e710ec5f9c5206c292ffcf3972fbe8cc` | 已审上游 v0.2.16：fix(frontend): show full account group names；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `29b2d721f46bddf5dc2555aad4adc063e0181d3b` | 已审上游 v0.2.16：test(web): use shipped SVG asset in embedded frontend tests；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `5753552b958084a7695fec7dfd923e30304ee108` | 已审上游 v0.2.16：fix(gemini): preserve authorization URL error messages；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `8e3d75b57b59b6bf8e9e695af27dab70ac360109` | 已审上游 v0.2.16：fix(monitor): refresh cached availability on automatic reload；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `bd2f53b855b79c6c20f9dee2187c78fa4ab7158c` | 已审上游 v0.2.16：fix(groups): reject nonpositive rate multiplier edits；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `d01628975d9f4d2d3ac9f9792df1bc5946a87061` | 已审上游 v0.2.16：fix(select): preserve keyboard input during IME composition；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `2bfc49477c0860058d2ff27d8ca40b0c3a1ca31c` | 已审上游 v0.2.16：fix(ops): initialize custom time ranges in local time；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `9a51333101a849bb948abf94c5ed53682e373edd` | 已审上游 v0.2.16：feat(groups): 分组名称与描述支持多语言；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `d56f5e787976c8b2a7bd471b9a3c30b0ccee50c1` | 已审上游 v0.2.16：feat(gateway): adapt Claude Haiku 5.5；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `ba280b3d5ea4d71e2f15fdb56ce60c4a05f422c1` | 已审上游 v0.2.16：feat(ops): show account names in request details table；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `daa79157491f25939b502f5c040eb11768a46e30` | 已审上游 v0.2.16：fix(deps): update SheetJS past expired security exceptions；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `63800c50d462efc6436e96b9dd46ae8f98a054bc` | 已审上游 v0.2.16：fix(gateway): guard structured-input fallback against unsafe retries；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `f7f1112cdab09e9bee982d77ae253d5d20a56e73` | 已审上游 v0.2.16：fix(deps): update Vue and source map security patches；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `c15d858eca81b269c41ba3bc161f923b179a860c` | 已审上游 v0.2.16：fix(gateway): preserve Responses semantics in chat fallback；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `a46249aac5611fca71fb0d955ad19be7602aca71` | 已审上游 v0.2.16：test(account): 补充临时不可调度作用范围单元测试；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `5a25507dcbb33fc40e3ef0659af210f206229a31` | 已审上游 v0.2.16：feat(account): 临时不可调度规则新增作用范围 scope 选项；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `983cf297da59e3be23b5e3f2c4eded945b53f408` | 已审上游 v0.2.16：fix(gateway): retry rejected structured input through chat endpoint；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `69d6d106fd365d8ed23b5a1c82eb3a968acc1454` | 已审上游 v0.2.16：fix(monitor): allow adding missing platforms in v2 settings；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `c4224045dc5cbf16f40c7b31aeb639aa0e211101` | 已审上游 v0.2.16：fix(payment): calculate QR countdowns from elapsed time；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `27de27ac6496e647c59866d9ba57de61ac356771` | 已审上游 v0.2.16：fix(risk-control): remove deleted groups from moderation scope；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `a56720855cf0c3ca38aea6da2a24c73e57ec1342` | 已审上游 v0.2.16：fix(ui): refresh countdown when tab becomes visible；随 1905ee5ce 合并并跑完整回归 | 上游同步 |
| 2026-10-14 | `f48866958b1ddcda57dab6844b6a3ade7783bc2d` | 已审上游 v0.2.16：fix(openai): only inject the Codex image generation bridge for OpenAI accounts；随 1905ee5ce 合并并跑完整回归 | 上游同步 |

| 2026-10-14 | `68ebc5b1235ba937df54703f5538d93ce24901e8` | 站内工单用户隔离、管理员队列、事务并发边界、严格 JSON、独立 Redis 限流和顶部导航；本地安全冒烟报告见 SUPPORT_TICKETS_SECURITY_SMOKE.md | 工单/导航 |
| 2026-10-14 | `92ed23164b716dc19015519dbc3addb74fafb6eb` | 支付查单测试使用 any 满足仓库 gofmt rewrite 规则；同版本 golangci-lint 全仓 0 issues，provider unit 通过 | 支付安全 CI |
| 2026-10-14 | `910c44bc2af2250760ee24fbc13dd8e6147f01eb` | EasyPay USDT 回调通过同一 provider 实例主动查单；签名、PID、订单号、GM 流水、类型、金额、支付状态和链上哈希全部匹配后才入账；失败保持 PENDING、写审计并返回 503；RMB EasyPay 保持原路径 | GM/EasyPay USDT 回调安全 |
| 2026-10-14 | `9d9e982f62ef395c8248fd21f56bf83ac21a173d` | 修复 EasyPay 创建签名重组为成功回调：拒绝未知/重复/缺失字段、错误 PID/签名类型/支付类型，绑定订单支付方式并拒绝 return_url userinfo 注入；真实攻击载荷回归通过 | EasyPay 回调防重放 |
| 2026-10-10 | `e13f039def8150f4c0f5d9b2cb9ed272bded612e` | 整合上次发布后的发票实名认证/先保存后税费续付、移除固定单数上限与最新教程调整；保留支付限流、赠送与主线分页/鉴权/酒馆；新增 255 退役教程迁移；89 项相关前端回归及后端发票/教程 unit 通过 | 最新本地代码/发票/教程 |
| 2026-10-10 | `de64c5d652adbb3b1b61e4e6444242ee1c933f0b` | 验证 256 迁移的禁用平台/模型/空范围保留、幂等与 14 维度默认值；账号筛选测试等待延迟重载并卸载组件，避免测试结束后遗留请求；专项 6 测试通过，集成测试编译通过 | 发布回归测试 |
| 2026-10-10 | `9ca791252b453f8fa7d793880e776333a2f8c2fc` | 被动监控平台目录补齐、后台完整开关及 V3 跨平台分组卡片键；保留禁用/模型/分组/空配置；55 项前端监控测试、后端监控 unit、类型/lint 和 PostgreSQL SQL 验证通过 | 渠道监控 |
| 2026-10-09 | `f3c44168d41b3d8e7be60ac42724b716ae08a7f5` | Go 1.27.2、x/net 0.60.0 安全修复与工具链适配；出站 HTTP2Config 健康探测保留协议/代理行为，服务端独立超时和旧 GOAWAY 重试兼容；govulncheck 零可达漏洞、golangci v2.14 零问题、协议及内存专项通过 | 安全/HTTP2/CI |
| 2026-09-29 | `856a7146a2e66fa855f45ace7a6e67cbc5e3657b` | 登记 v0.2.10 服务器发布、镜像与备份；无业务改动 | 发布记录 |
| 2026-09-29 | `173eb0aa3034c053972d2590ceb04e1d28f0dea4` | 修复教程组件全角空格和事务 rollback errcheck；保持发布与权限语义 | CI 修复 |
| 2026-09-29 | `0b7ab5f1cec997440fd851282ab8558c9e3c6e1a` | 教程发布管理、共享客户端配置、工作台提示词优化和跨平台托管路由；初始化事务保护，保留授权/计费及上游修复；新增 240_help_documents 至 254_smart_route_cross_platform，验证见本日候选记录 | 功能整合 |
| 2026-09-24 | `d75d183871a2d7e08e7a7fe531f395e5dac2a6c9` | 记录 0.2.8 生产镜像、备份、蓝绿切换与验收；无业务代码变更 | 发布记录 |
| 2026-09-21 | `d2a747d7e7ba24bf9d6ce9e4a4dfee32aa3bdfbc` | 显式忽略 multipart 校验后临时文件清理错误，保持原行为并通过 errcheck | 图片工作台/CI |
| 2026-09-21 | `bfa65c190faa1c2124c719725435013eff93e0cb` | 内嵌图片工作台、受限桥接和隔离内部密钥；保留既有网关计费与优先级；新增 239 迁移和两套前端 CI | 图片网关/密钥/前端/部署 |
| 2026-09-18 | `fa40c17b2003bb3c0db38de6b0ba31a0e3147cd2` | 上游测试适配 Aivoza i18n 弹窗、分页跳转按钮及签到历史分页响应；保留行为断言 | 测试 |
| 2026-09-18 | `81ecaccffffe8ed38902a0271310cf15e6497253` | 签到验证码自适应、统一开关、模型列表及端点紧凑布局 | 前端/主题 |
| 2026-09-17 | `502c64c966e4a8064f5c3cbb3dbfc7a86901cd1b` | 触发 v0.2.5 发布候选的完整 CI 重跑；不改变代码树和业务行为 | CI/发布 |
| 2026-09-17 | `567959c0433a1d402734c7303acf164c847d981e` | 修复 v0.2.5 CI 测试夹具参数并升级 gRPC 依赖至无已知可修复漏洞版本；不改变业务行为 | CI/安全 |
| 2026-09-12 | `76f3f2594dc44790dcb6201231ac8f4e41a3bbdb` | 多项客服配置、旧文本兼容、统一客服弹窗和链接协议校验 | 前端/客服 |
| 2026-09-12 | `4d4419f1b7dbaf2b337c808f0b8300c80083a792` | 处理配额查询 Close 错误及缓存类型异常，测试连接清理显式处理；通过 errcheck | 注册/质量 |
| 2026-09-12 | `ce3d39a4d9eac521bbb73625d259c3289f0f58a9` | 登记新增注册防护页面，视图覆盖从 90 更新至 91，保留完整性断言 | 测试/主题 |
| 2026-09-12 | `ac36c2401a3bf6c2bd502fa557bc17ed80740ad8` | 注册防护、原子来源配额和管理页面；包含上次界面快照 | 注册/风控 |
| 日期 | 提交 | 说明 | 责任域 |
| --- | --- | --- | --- |
| 2026-09-11 | `6c56e55eaffb7ebabcc6ba5c40bcdcfbe3e4a6cf` | 统一筛选/控件布局，日期弹层视口定位及焦点恢复，顶部客服与社群入口 | 前端/主题 |
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
| 2026-08-30 | `e3fd418b7` | feat(payment): open GM checkout popup and sync status | 支付/GM |
| 2026-08-31 | `444c961a4` | feat(payment): add recharge bonus tiers and USDT limits | 支付/GM |
| 2026-09-01 | `213b3fcf7` | feat(flowai): harden check-in and add reactivation broadcasts | 签到风控/邮件召回 |
| 2026-09-02 | `4dc03354a` | feat(flowai): verify check-in modes in confirmation dialog；普通/幸运签到确认弹窗分别完成 Turnstile 校验 | 签到/i18n |
| 2026-09-03 | `beed8a2f4` | feat(flowai): add API key smart routing；新增智能路由、渠道监控 V3 展示层和中英文 locale | API key/渠道监控/i18n |
| 2026-09-03 | `46d3b617e` | fix(flowai): satisfy smart routing lint；修正 De Morgan 表达式和测试类型断言错误检查，不改变运行时路由语义 | CI/智能路由 |
| 2026-09-03 | `6c16703c7` | fix(flowai): update settings API contract；补齐 `smart_routing_enabled=false` 的公开设置契约测试 | CI/智能路由 |
| 2026-09-09 | `04661fec1` | chore: release FlowAI 0.2.3；仅更新嵌入式应用版本，不改变上游基线或 FlowAI 业务契约 | 版本/发布 |
| 2026-09-09 | `af997094b` | test: align upstream 0.2.3 fixtures；适配上游路由认证链、Pinia/API mock 和英文账号管理 i18n key，不改变运行时业务契约 | CI/测试/i18n |
| 2026-09-09 | `76b125bb1` | test: update upstream 0.2.3 service fixtures；补齐上游接口变更后的单元测试夹具，不改变运行时业务契约 | CI/测试 |
| `ef38fe7a6eb19ed15171af535e941c0176da765e` | Revert "chore: sync VERSION to 0.1.163 [skip ci]" | 历史版本同步及回退，最终版本以 0.2.4 为准 | 契约及完整 CI |
| `e63167456673345a31e0c63eed9ef91bee7ceba1` | chore: sync VERSION to 0.1.163 [skip ci] | 历史版本同步及回退，最终版本以 0.2.4 为准 | 契约及完整 CI |
| `79c83f43d4261420fe92e47092c5a4e3c1ed2d29` | Aivoza 主题、首页、明暗模式、可访问性、订阅有效期开关；保留全部素材和独立迁移 | 前后端/部署测试与迁移验证 |
| `fc9632b9038aa583fbaa30b996e0aa0cc1ab9b7b` | 合并适配：开关可访问名称、订阅测试夹具、中英状态 key、实际发布 Dockerfile 首页资源 | 前后端/部署测试与迁移验证 |
| `cb4d0bf57b8b9fb1ac44764e9db25200c7f29684` | AIVOZA_ROLLING_DEPLOY=true 跳过启动时跨进程并发槽清扫，TTL 清理保留；保护滚动部署中的旧请求 | 前后端/部署测试与迁移验证 |
| `b4a79197bf011474137a633cdb0d7ce6be689996` | 两阶段蓝绿发布：验证 digest 和健康后热切 Caddy，只替换应用，保留旧容器并记录回滚状态 | 前后端/部署测试与迁移验证 |
<!-- FLOWAI_LEDGER_NON_MERGE_END -->

## 8. 历史合并提交索引

合并提交本身也要记录，因为冲突解决可能改变文件内容，即使没有新增 feature commit。
以下是当前分支基线之后的合并提交；下一次执行上游合并后，必须把新的 merge hash 和
审阅结论加入标记区。`git merge-base HEAD upstream/main` 变化后，以命令输出为准补齐。

<!-- FLOWAI_LEDGER_MERGE_BEGIN -->

| 2026-10-29 | `234c14a61753c7438450fd6053178014b40a9b93` | 合并已审上游 a3298b46 v0.2.19；四处冲突逐块整合 Grok 模型目录诊断、Gemini 同账号重试、分组模型教程及客户端配置，同时保留 Aivoza 教程、图片工作台、站内工单、支付/签到/邮件、渠道监控和调度并发契约；无新增上游 SQL 迁移，GMPay 查单重复参数修复随候选保留 | 上游同步/发布候选 |

| 2026-10-26 | `5779f635a1f89a95babdc484dc9380c4dc0e9fc7` | 合并已审上游 ba81c7fd7 v0.2.18；4 处冲突保留 Aivoza 主题、请求级容量判定与旧管理员密钥兼容入口，接入 Retry-After 退避、Vertex 配额、分范围管理员密钥、Gemini 分组生图开关及按返回图片数计费；新增 246/247/248 增量迁移 | 上游同步/发布候选 |

| 2026-10-16 | `0477bb8bfb648eddbd051bd6f6a59cceed0085fe` | 合并已审上游 b0a02a330 v0.2.17；支付确认保留 USDT 强校验并接入普通易支付查单，保留账号筛选/主题；其余模型协议、退款和计费修复完整接入 |

| 2026-10-14 | `1905ee5ce3359087262204b653430b935a7565c3` | 合并已审上游 9c2cca9e6 v0.2.16；8 个冲突保留智能路由、错误清洗、主题及拖拽 JSON，并接入分组本地化、Vertex API Key、运营账号名称；Ent/Wire 重新生成 |

| 2026-10-14 | `4c969884663455fddf57270a5d0b46700239feea` | 工单主线合入远端支付回调主动查单修复；仅台账追加冲突，保留双方记录 |
| 2026-10-14 | `5e6e51570088647627cc7339ce9f2f65f7867cc9` | 合入当前 Aivoza `origin/main` 的 v0.2.15 生产治理记录；仅 `FLOWAI_CHANGELOG.md` 冲突，保留既有部署/Guardian 证据并追加 GMPay 已部署、Sub2API 待发布状态，支付业务代码无冲突 | 支付安全候选/主线同步 |
| 2026-10-10 | `5481cb33749398679c91cf4a4e8c000e505af2eb` | 合并已审上游 3a6fd1c9d v0.2.15；11 个冲突逐块保留主题、教程共享配置、模型发现、设置路由和较新工具链；接入平台目录及两新平台，新增 242/256 迁移；审阅结论见本日候选 | 上游同步/渠道监控 |
| 2026-10-09 | `09ab7ed2ef7d219576c00b1fc8614a1fe2fbfd06` | 合并上游 5fc0e486c3 v0.2.14；43 个冲突文件逐块保留 Aivoza 支付/邮件/主题/教程/工作台，统一优惠报价与校验，吸收安全修复、TypeSafe 和模型目录；两条增量迁移、测试结果及冲突结论见本日候选记录 | 上游同步 |
| 2026-09-29 | `2bc658f4e2ecee46b7766047f3891870570e56c3` | 合并已审上游 0.2.10，五处冲突逐块保留定制服务与主题；无新增上游迁移，客户端配置验证通过 | 上游同步 |
| 2026-09-24 | `002f9da298b0fa1d014826e4bc8d938dd9a9d7b5` | 合并已审上游 v0.2.8；保留 Aivoza 主题、Studio、支付/签到/邮件/并发契约；逐块处理 13 个冲突及测试适配，新增 238b/239/240 上游迁移，生产旧倍率配置命中为 0 | 上游同步 |
| 2026-09-20 | `e8821fe0db36d24edea75b2b65f9976e076ab094` | 合并已审上游 v0.2.7；保留定制调度接口与 Aivoza 主题，吸收平台参数和移动端模型广场入口；无 SQL 迁移 | 上游同步 |
| 2026-09-18 | `8763b72af0a325f17f2656429ced6b60370118e9` | 合并已审上游 efe9aab1e；四处冲突保留 Aivoza 弹窗、充值赠送、监控和发票，吸收增量修复；无迁移 | 上游同步 |
| 2026-09-16 | `76f1cec2eae3d7a9d8026026fb011f4cbf24a75a` | 合入上游 v0.2.5；逐块解决 51 个冲突文件，保留 Aivoza 主题、优先级 1 最高、-1 并发拒绝、GM/EasyPay、充值赠送、智能路由、签到/邮件防重、注册防护和滚动发布契约；新增两条 238 迁移按完整文件名登记 |
| 2026-09-12 | `469d3af9146bb364c043dc7ca0f12a6a8224d468` | 合入 fix/aivoza-controls-release 新客服功能，无冲突；保留注册防护和视图登记 |
| 2026-09-12 | `081aa696696108b6c59643336350228bf04e2424` | PR 4 合并注册防护及原子配额，保留 FlowAI 契约，238 迁移已上线 |
| 2026-09-12 | `5eb066016e84f0abfff653a5c4b0fa61d54660fd` | 合并已发布 main，无冲突；保留注册功能和上次界面变更 |
| 2026-09-11 | `f7518913e4c56f69102d95715fb1ebf0d0f819c5` | PR 3 合并控件布局发布；保留 FlowAI 行为 |
| 日期 | 合并提交 | 说明 |
| --- | --- | --- |
| 2026-09-09 | `fa931c345` | Merge upstream main 0.2.3；完整合入 260 个上游提交，逐项解决 13 个冲突文件；保留 FlowAI 账号优先级 1 最高、-1 并发拒绝、GM/EasyPay、充值赠送、签到/邮件防重、i18n 和预构建部署，适配上游 GroupModelAllowlist、Codex manifest、MiniMax、上游请求 ID 与渠道策略 |
| 2026-09-02 | `88ad9b765` | Merge upstream main 0.2.0；完整合入 53 个上游提交，无文本冲突和 incoming 删除；保留 FlowAI 优先级、并发、GM/赠送、签到 Turnstile、邮件防重、i18n 和预构建部署，吸收分组定价、推理强度、Fast/free Fast、缓存定价及 WebSocket/Responses 修复 |
| 2026-09-01 | `651b92209` | Merge upstream main 0.1.185 hotfixes；保留 FlowAI 错误脱敏、Project Mihomo、优先级、并发和支付行为，吸收 Kimi 原生 Responses 与 Anthropic fallbacks 清理 |
| 2026-09-01 | `41fda3b26` | Merge upstream main 0.1.185；保留 FlowAI 优先级、并发、GM/赠送、签到 Turnstile、邮件召回和 i18n，吸收定价目录、数据库重试及 Codex/WebSocket 修复 |
| 2026-08-31 | `03c0c8289` | Merge upstream main 0.1.184；保留 FlowAI 优先级、并发、GM/充值赠送和 i18n，吸收 TTFT、compaction、Ollama 与稳定性更新 |
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
| `d9af7dfe5b744b034cc353d43da68b0aade1538d` | merge: sync upstream 0.2.4 before Aivoza mainline migration | 上游 98d86915b 仅更新版本为 0.2.4，无冲突、无新增迁移 | 契约及完整 CI |
| `70f1706307d26bf7373e217db6131f195522080d` | merge: preserve fork main history for Aivoza mainline | 合入 fork main 历史，无文本冲突、最终文件树未变化 | 契约及完整 CI |
| `511ee924b154c0dbb2e9742b543e8588522fbb1a` | Merge branch 'Wei-Shaw:main' into sub2api-flowai | 保留 Image 2.5/OAuth 生图修复，与优先级及支付无关 | 契约及完整 CI |
| `1239bc23053c9c76492b58d8a32c2a831f5d808b` | Merge branch 'Wei-Shaw:main' into main | 保留既有分支历史；以本次完整契约和 CI 验证最终树 | 契约及完整 CI |
| `ba457492f18fea409c6210d7859b834a4afd9d66` | Merge branch 'Wei-Shaw:main' into main | 保留既有分支历史；以本次完整契约和 CI 验证最终树 | 契约及完整 CI |
| `5bf0543409ee6e8df716774defedcec96e7a62f0` | Merge branch 'Wei-Shaw:main' into main | 保留既有分支历史；以本次完整契约和 CI 验证最终树 | 契约及完整 CI |
| `a67681f5c4a6740d0b4a73567b2e1d782ea2beef` | Merge branch 'Wei-Shaw:main' into main | 保留既有分支历史；以本次完整契约和 CI 验证最终树 | 契约及完整 CI |
| `db0955f68d2494305644c6528229b2ef37635a3b` | Merge branch 'Wei-Shaw:main' into main | 保留既有分支历史；以本次完整契约和 CI 验证最终树 | 契约及完整 CI |
| `199f886533a997fc667e4d2eb27cf25e7c230557` | Merge branch 'Wei-Shaw:main' into main | 保留既有分支历史；以本次完整契约和 CI 验证最终树 | 契约及完整 CI |
| `87b04f16f427d4fcd82686c88083df3cb1897795` | Merge branch 'Wei-Shaw:main' into main | 保留既有分支历史；以本次完整契约和 CI 验证最终树 | 契约及完整 CI |
| `1f2711f042a468719afa478d5cc3cbb2a87223d1` | Merge branch 'Wei-Shaw:main' into main | 保留既有分支历史；以本次完整契约和 CI 验证最终树 | 契约及完整 CI |
| `6c7fcef9a514648fe2aec1715c5e230721ffca8a` | Merge branch 'Wei-Shaw:main' into main | 保留既有分支历史；以本次完整契约和 CI 验证最终树 | 契约及完整 CI |
| `74468f31938e205290e9716c4bf734dab9f3766a` | 主题合并到上游 0.2.4；12 文件逐块解决，保留模型白名单/Toggle/菜单滚动/注册开关/Markdown/MiniMax/排名隐私开关，套用主题样式 | 完整 CI 和严格契约 |
| `0efaa80cefa284cf0f7d423bdfa348f6b95bd6bb` | 合入第一份 PR 的 GitHub 临时 merge ref 检查；无业务变更 | 完整 CI 和严格契约 |
| `3ab944445a425e010cc2b25cd4fe78d488b0dd92` | 合入已通过 GitHub CI 的 main PR #1；无文件变化，保留完整分支历史 | 完整 CI 和严格契约 |
<!-- FLOWAI_LEDGER_MERGE_END -->

## 9. 发布记录模板

每次发布在工单、PR 或发布记录中复制以下字段；不要把密钥、JWT、TOTP、支付 secret
或 SSH 私钥写入文档。

```text
日期/时区：
发布人：
分支：main
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
