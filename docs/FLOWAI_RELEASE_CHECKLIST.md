# FlowAI 合并与发布核对清单

> 2026-09-11 主线迁移：用户已授权以 `main` 为 Aivoza/FlowAI 唯一正式主线。
> `upstream/main` 保留上游身份；旧 `sub2api-flowai` 与主题分支作为历史归档。
> GitHub 测试通过后构建 `ghcr.io/aboutnb/aivoza-sub2api:sha-<sha12>`，生产固定 digest。
> 本次要求不停机：新旧容器并存、健康验证、代理热切换、旧连接排空；不得直接执行旧的重建发布脚本。
> 下文历史快照中的分支和旧镜像仅用于追溯，不再授权旧分支发布。


这份清单用于 main 的每次上游合并、功能发布和服务器更新。勾选项必须有
命令输出、测试结果、镜像标签或截图等证据；没有证据就视为未完成。

## 1. 发布前冻结和识别

- [ ] 当前工作目录是 /Users/xiaobo/develop/sub2api（或已确认的同一仓库副本）。
- [ ] 开发位于 feature/*、fix/* 或 sync/*；正式发布源为已通过 PR 和完整 CI 的 main。
- [ ] 已记录发布前的 HEAD、版本号、origin/main 和 upstream/main。
- [ ] 已确认 .playwright-cli/、data/、本地素材等未跟踪文件的保留/排除清单，未把它们
      顺手加入发布提交。
- [ ] 已完成数据库、Redis AOF、/app/data 和 Mihomo 状态备份或确认本次不涉及生产数据。

建议命令：

~~~bash
git switch -c sync/upstream-<version> origin/main
git status --short --branch
git log -1 --oneline --decorate
git rev-parse HEAD
git rev-parse --verify upstream/main
grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' backend/cmd/server/VERSION
~~~

## 2. 上游合并前核对

本清单必须与 `docs/FLOWAI_CHANGELOG.md`（变更台账）和
`docs/FLOWAI_BRANCH_CONTRACT.md`（行为契约）一起使用；只看上游 release note
不能替代分支台账核对。

- [ ] 已阅读变更台账的当前行为总表、受保护路径和上游合并冲突矩阵。
- [ ] 已确认本次功能提交已先写入“功能与提交台账”；仅治理文档提交才可使用
      `[flowai-governance]` 标记，且不得夹带业务代码。
- [ ] 已确认上游变更没有覆盖 FlowAI 的调度、并发、i18n、支付、签到、邮件、迁移或部署
      行为；有冲突时已在发布记录中写明保留哪一方及测试证据。

- [ ] 上游引用已更新，但没有切换或修改本地 main。
- [ ] 如果 `upstream/main` 已前进，已完成预审并用完整 SHA 显式确认；未确认时不得合并或发布。
- [ ] 已查看 git diff upstream/main...HEAD 的文件清单，标出 FlowAI 专属文件。
- [ ] 已查看上游新增迁移，并确认没有同名文件覆盖或丢失。
- [ ] 已阅读分支契约第 2 至第 6 节，尤其是账号优先级、-1/0 并发、邮件防重、i18n、支付
      和迁移 checksum 条款。
- [ ] 冲突解决保留了 FlowAI 行为；没有用上游文件整文件覆盖本分支实现。

建议命令：

~~~bash
git fetch upstream main
make review-flowai-upstream
# 人工核对报告后才允许确认；SHA 必须是本次实际 upstream/main 的完整值
FLOWAI_UPSTREAM_REVIEW_ACK=<upstream-main-full-sha> make review-flowai-upstream
git diff --stat upstream/main...HEAD
git diff --name-status upstream/main...HEAD
git log --oneline --no-merges $(git merge-base HEAD upstream/main)..HEAD
~~~

上游合并完成后，先保存冲突结论和新合并 hash，再更新
`docs/FLOWAI_CHANGELOG.md` 的合并提交标记区。功能代码提交不能用“以后补文档”放行；
若门禁提示提交未登记，停止发布并先补齐行为、路径、测试和迁移记录。

如果预审报告出现冲突、受保护路径变更或无法解释的删除，保持当前分支不变；不要用
`git checkout upstream/main -- .`、整树覆盖或“先合并再看测试”的方式处理。

确认无误后才允许执行：

~~~bash
git merge --no-ff upstream/main -m "merge reviewed upstream into Aivoza candidate"
~~~

若发生冲突，先暂停发布，按功能块解决并重新查看：

~~~bash
git diff --name-only --diff-filter=U
git diff --check
~~~

## 3. 自动契约检查

- [ ] make check-flowai-contract 通过。
- [ ] 发布候选使用 `make check-flowai-contract-strict`；该命令必须在工作树无暂存、未暂存
      和未跟踪文件时通过。
- [ ] 检查报告确认当前分支包含已获取的 upstream/main，与 .github/aivoza-upstream-ref 的已审 SHA 一致。
- [ ] 账号调度仍为 1 最高、数值越小越优先；没有把错误透传规则的优先级规则混入账号调度。
- [ ] -1 在用户槽位和等待队列前立即拒绝，0 保持不限制，小于 -1 不会写入。
- [ ] 风险注册在授权判断前仍为 -1，授权成功后才应用实际并发。
- [ ] i18n、FlowAI 迁移、预构建镜像脚本和工作流文件均存在。
- [ ] 已执行迁移的 checksum 保护代码仍存在。
- [ ] API key 智能路由默认关闭；启用时候选组不超过 20 个、平台/计费类型一致，且无效
      权重、倍率、模型或账号状态会被拒绝或安全排除。
- [ ] 渠道监控 V3 只改变用户端展示层；V2 API/权限裁剪不变，`monitor_view=v2` 回退路径
      可用且不会绕过后端模式守卫。

本地默认检查允许存在未提交的跟踪文件改动，但会明确警告。正式发布前使用严格模式：

~~~bash
make check-flowai-contract-strict
~~~

`make check-flowai-contract` 适合开发过程中保留本地素材时进行语义检查；正式合并或发布
使用 `make check-flowai-contract-strict`。严格模式发现未跟踪文件时，先确认它们不属于发布
提交，不要用删除本地素材的方式清理工作树。

如果检查因有意的重构失败，先更新测试和本契约，再继续；不要通过删除检查项来放行。

## 4. 测试门槛

- [ ] git diff --check 通过。
- [ ] 后端并发、风险注册、调度、邮件广播、支付、签到和迁移相关测试通过。
- [ ] 签到 Turnstile 关闭时保持原流程；启用时缺 token、验证失败和验证依赖不可用均不结算，
      当天已结算请求仍幂等返回；关闭可编辑风控也不会关闭硬性同源/指纹保护。
- [ ] 邮件召回的未活跃天数筛选、事件模板和受众快照正确；`sent`/`sending` 防重语义未改变。
- [ ] 前端 lint、typecheck、关键 Vitest 和 i18n 测试通过。
- [ ] Docker Compose 配置渲染通过，环境变量没有把生产密钥写入仓库。
- [ ] 若涉及上游协议或支付回调，已做真实配置下的连通性/回调状态核对；没有把 HTTP
      200、钱包存在或浏览器跳转当作链上结算证据。
- [ ] GM 保持精确金额匹配、最小精度递增和严格 `chain_tokens.min_amount`；不存在
      `system.usdt_underpayment_tolerance`、固定网络费或少到账补偿逻辑。
- [ ] 充值赠送默认档位为 50 赠 5%、100 赠 10%，后台可增删配置且空数组可关闭；快捷
      金额角标、到账预览和 `/api/v1/payment/checkout-info` 返回值一致。
- [ ] 已验证门槛前后、最高档位、手续费不参与赠送、余额倍率换算、重复回调幂等和配置
      修改不影响既有订单。

建议命令：

~~~bash
git diff --check
make check-flowai-contract
make test-backend
make test-frontend
docker compose --env-file deploy/.env.preview -f deploy/docker-compose.preview.yml config --quiet
~~~

资源不足时可以先跑目标测试，但正式发布记录必须写明未运行的测试和剩余风险。

## 5. 镜像构建和取证

- [ ] GitHub Actions 的提交 SHA 与准备发布的 HEAD 相同。
- [ ] .github/workflows/preview-image.yml 的契约检查通过后才开始构建。
- [ ] GHCR 中确认 aivoza-sub2api:sha-<sha12> 标签；生产必须固定 digest。
- [ ] 记录镜像 digest、构建时间和提交 SHA，不只记录 latest 或可变标签。
- [ ] 确认镜像包含当前前端 locale、迁移文件和后端版本。
- [ ] 已确认 `backend/migrations/234_api_key_smart_routing.sql` 在镜像中，且
      `smart_routing_enabled` 默认值为 `false`；未启用前不改变既有 API key 路由。
- [ ] 23 服务器发布命令显式传入 `ghcr.io/aboutnb/aivoza-sub2api@sha256:<digest>`，没有依赖可变分支标签。

不要在 23 服务器上运行 docker compose build、pnpm build 或 go build 作为正式发布
步骤；服务器只拉取已验证镜像。

## 6. 23 服务器发布（Termius）

- [ ] 通过 Termius 连接目标 23 服务器，确认主机、用户和部署目录无误。
- [ ] 发布前查看容器、磁盘、最近错误日志和当前镜像标签。
- [ ] 已确认 23 服务器 `/root/flowai/deploy/.env` 仍使用生产配置，未用示例文件覆盖它。
- [ ] 使用已测试镜像 digest 执行 deploy/deploy-aivoza-bluegreen.py prepare；候选验证后再 promote。
- [ ] 候选与旧应用并存，启用 AIVOZA_ROLLING_DEPLOY=true；代理热切换期间连续健康探测无失败；依赖未重启。
- [ ] 迁移执行成功；检查日志中没有 checksum mismatch、migration failed 或数据库连接错误。
- [ ] 发布后核对健康、公开设置、登录、账号调度、-1 并发拒绝和一个普通正数并发请求。
- [ ] 若本版本涉及支付/签到/邮件，完成对应的非破坏性 smoke test，并记录数据库/日志证据。
- [ ] Sub2API 应用仍使用 `flowai-app`、`flowai-postgres`、`flowai-redis` 和
      `/root/flowai-data/{app,postgres,redis}`；不得误用通用 `docker-compose.yml` 创建
      另一套 `sub2api*` 容器或空命名卷。
- [ ] BEpusdt 生产容器和活跃目录不存在，Sub2API 镜像中没有独立 BEpusdt 配置、路由或
      服务；Caddy 当前配置不再包含 `bepusdt` 上游，历史 206 迁移文件未被改写。
- [ ] GM 使用独立 `gm-epusdt` Compose、`flowai-gm` 容器和独立数据目录；更新 GM 时不得
      重建 `flowai-app` 或复用 BEpusdt 数据目录。
- [ ] 启用 GM/EasyPay USDT 前，`/payments/gmpay/v1/config` 返回非空
      `supported_assets`，钱包、RPC/监听和回调 smoke test 均有证据；为空时保持前台方式禁用。
- [ ] `/api/v1/payment/limits` 中 USDT 方式的 `single_min` 与后台
      `payment_usdt_min_amount` 一致（默认 50），低于下限的直接创建订单请求也被后端拒绝。
- [ ] 桌面端 GM checkout 弹窗约为 `625x900` 且居中；小屏会自适应缩小，父页面订单轮询
      和服务端回调同步未受影响。

旧重建命令仅作历史回滚参考（本次禁止使用；连续服务发布采用契约第 6 节的 prepare/promote）：

~~~bash
cd /root/flowai/deploy
docker compose --env-file .env -f docker-compose.preview.yml ps
docker compose --env-file .env -f docker-compose.preview.yml logs --tail 200 sub2api
ENV_FILE=.env \
SUB2API_IMAGE=ghcr.io/aboutnb/sub2api:sub2api-flowai-<sha12> \
SERVICE=sub2api \
  ./deploy-preview-image.sh
curl -fsS https://aivoza.com/health
curl -fsS https://aivoza.com/api/v1/settings/public
~~~

实际域名、目录和端口以服务器 `.env`、Caddy 配置及变更记录为准；不要把密码、
JWT、TOTP、支付密钥或 SSH 私钥写入本清单。

## 7. 回滚和异常处理

- [ ] 已记录上一版本不可变镜像标签和 digest。
- [ ] 健康检查失败时先保留 docker compose ps、应用日志和迁移日志。
- [ ] 应用镜像可以切回上一版本；数据库迁移不执行盲目逆向，按备份和兼容性方案处理。
- [ ] 邮件任务出现 sending 时，先查 SMTP/provider 结果再决定是否人工处理，禁止全量
      重置为 pending。
- [ ] 支付订单、签到账务或迁移状态异常时暂停后续重试，保留订单号、任务 ID、迁移名和
      时间戳，避免重复结算或重复发送。

## 8. 发布记录模板

~~~text
日期/时区：
发布人：
分支：main
HEAD：
上游基线（upstream/main）：
应用版本：
FlowAI 专属变更：
契约检查：PASS / FAIL（命令与输出位置）
测试结果：
新增迁移及执行结果：
镜像 tag：
镜像 digest：
服务器：23
部署目录：
发布后 health/settings/public：
业务 smoke test：
上一版本回滚 tag：
异常/未验证风险：
~~~
