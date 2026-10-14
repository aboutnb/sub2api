# 站内工单安全冒烟报告

日期：2026-10-14。范围：用户与管理员工单功能、顶部签到和工单入口、头像菜单客服、管理员公告及兑换菜单。

最终结果：**133 项真实 API 检查、47 项浏览器断言全部通过**。发现的输入边界、资源限制及键盘交互问题已修复，本地 Docker 已更新并通过复核。本报告不代表生产环境已完成安全验收。

## 环境与数据保护

- 最新主线工作目录：`/Users/xiaobo/develop/sub2api-v025-release`，基线 `c919d8d2403f165173af283d1f679a0b38c2709f`，叠加本地未提交工单、导航和安全修复。
- 修复已按功能同步到 `/Users/xiaobo/develop/sub2api`；保留该目录与最新主线的其他历史差异。
- 攻击样例和并发写入使用 `127.0.0.1:18081` 的独立应用、全新 PostgreSQL 16、独立 Redis 与测试账号。未复制现有渠道密钥或用户数据，未调用生产网站、支付或模型供应商。
- 普通用户、第二普通用户及管理员均使用真实数据库账号和实际登录会话；没有通过浏览器修改角色来替代授权验证。
- 现有应用 `127.0.0.1:8080` 更新为 `sub2api-local:ticket-security-20261014`，31 项容器配置核验全部一致，健康检查通过。
- 更新前后用户数、工单数、消息数、设置数及余额汇总一致。现有应用复核只执行登录、读取和非法查询，未写入测试工单。
- 私有数据库、数据目录和配置备份：`/Users/xiaobo/.codex/backups/sub2api-local/20261014-security-smoke/runtime`。上一应用保留为 `sub2api-image-studio-prompt-before-ticket-security-20261014-195935`。报告与公开测试结果不含密码、JWT 或管理密钥。

## 发现与修复

| 项目 | 修复前实测或源码发现 | 修复后行为 |
| --- | --- | --- |
| 分页与搜索边界 | 极大页码造成 OFFSET 溢出；搜索包含 NUL 字符均返回 500 | 400；严格解析整数，每页最多 100 条、页码最多 1000000、搜索最多 200 字符，拒绝非法 UTF-8 和 NUL |
| 工单文本 | 主题或正文包含 NUL 字符返回数据库异常 500 | 入库前返回 400，保留正常中文、表情与换行 |
| 请求体大小与结构 | 带 1 MB 无关字段的创建请求返回 200 | 创建、用户/管理员回复和管理员状态 JSON 最多 64 KiB，超限返回 413；未知字段、尾随 JSON、非法 UTF-8 返回 400 |
| 持续写入滥用 | 已有全局限流及 10 个未关闭上限，但创建后关闭可持续写入；缺少工单专属窗口 | 按认证用户累计：创建 5 次/分钟、20 次/小时，回复 20 次/分钟、120 次/小时。更换 IP、工单编号或关闭工单不能重置额度；关闭全局限流后仍有效 |
| 长会话资源放大 | 消息总量无上限，详情及每次回复返回全部会话 | 每工单最多 200 条消息，事务锁内计数。真实数据库并发争抢最后一个名额时仅一次成功；管理员同样受约束 |
| 限流故障 | 全局面板限流采用 Redis 故障放行策略 | 工单创建/回复在 Redis 不可用时返回 503，读取继续可用；Redis 恢复后可正常写入 |
| 签到菜单键盘操作 | 焦点在触发按钮时 Esc 无法关闭菜单 | 根节点处理 Esc；组件回归与实际浏览器均确认可关闭 |

限流记录尝试次数；超限返回 429 和 `Retry-After`。管理员与管理员 API Key 都受工单回复限流。达到 200 条后可继续读取和关闭，应创建新工单继续会话。新错误均有中文、英文文案；迁移文件及其校验和未改动。

## API 验证

133 项检查通过，逐项结果保存在 [主 API 结果](../output/security-smoke/api-results.json) 与 [管理员 Key 补充结果](../output/security-smoke/admin-key-results.json)。

- 未认证的全部工单入口返回 401；普通用户的全部管理员工单入口返回 403。
- 用户 B 读取、回复、关闭用户 A 工单均返回 404，与不存在工单响应一致；列表不能通过 `user_id` 查询他人工单。普通用户详情不附带邮箱，管理员可按用户邮箱查询。
- 无签名、篡改、过期和已撤销 JWT 被拒绝；即使以测试签名生成含管理员角色的用户 JWT，数据库中的普通用户角色仍阻止管理员访问。URL、Cookie 和网关 Key 均不能替代用户 Bearer 认证。
- 管理员 Key 可正常读取和回复，作者身份取真实管理员；错误 Key 返回 401，合法管理 Key不能用于用户 JWT 路由，管理员 Key 不能绕过工单限流。
- JSON 不能伪造工单归属、状态、作者 ID 或客服身份；未知字段直接拒绝。
- 四种分类、队列切换、用户回复、管理员回复、关闭、重复关闭幂等、管理员修改状态及回复重开全部通过。
- 中文 80 字主题、4000 个表情正文成功；超出一个字符、空白必填、非法分类/状态、非法 ID、NUL、错误 UTF-8、超大请求及尾随 JSON 均按预期拒绝。
- SQL 注入字符串不能扩大结果或执行 SQL；`%`、`_`、反斜杠按字面查询；用户邮箱搜索不能泄露其他数据。工单内容通过参数化 SQL 保存。
- 并发创建没有突破未关闭上限；真实 PostgreSQL 并发回复没有突破 200 条上限；关闭与回复竞争保持最终状态、时间戳一致。
- 分钟、小时、全局用户限额及不同用户隔离均有效；关闭全局限流后工单专属限流仍有效。Redis 故障与恢复验证通过。
- 禁用、修改密码和删除测试账号后，旧会话均不能访问工单。Cookie 跨站写入失败；跨域响应不会暴露携带凭据的通配访问；响应具备 `nosniff` 和框架嵌入防护。

## 浏览器与回归验证

47 项明确记录的浏览器断言通过，另外验证了签到菜单 Esc 关闭、明暗主题及第二真实用户隔离。

- 用户实际创建、回复、关闭工单；关闭后回复表单隐藏。管理员实际回复、关闭、通过回复重开；队列正确切换。
- 存储的 `<script>`、`<img onerror>`、`<svg onload>` 在用户和管理员会话中保持文本，没有生成攻击元素或执行脚本。
- 客服名称中的恶意 HTML 保持文本，`javascript:` 联系链接不生成可点击链接，合法邮件链接正常显示；弹窗 Esc 关闭并恢复焦点。
- 375、768、1024、1440 宽度无横向溢出；顶部工单入口可见、当前路由状态正确；用户侧栏无重复签到及工单入口；客服在头像菜单 API 密钥之后。
- 普通用户不能进入管理员工单页面；管理员“公告管理”“兑换管理”及两个兑换子菜单路由正确。
- 现有本地 Docker 的真实管理员登录、管理员工单页面、顶部工单入口及用户工单页面正常。

| 检查 | 最终结果 | 证据 |
| --- | --- | --- |
| 后端目标包测试与 race | 7 个包通过，无竞态报告 | [日志](../output/security-smoke/backend-fixed-race.log) |
| PostgreSQL 工单集成测试 | 6 项通过，含实际并发与事务回滚 | [日志](../output/security-smoke/postgres-integration-fixed.log) |
| 工单、导航、签到、i18n 前端测试 | 63 项通过 | [日志](../output/security-smoke/frontend-tests-final.log) |
| 客服链接与布局无障碍回归 | 17 项通过 | [日志](../output/security-smoke/contact-a11y.log) |
| typecheck、目标 ESLint、go vet | 通过 | [类型检查](../output/security-smoke/typecheck-final.log)、[ESLint](../output/security-smoke/eslint-final.log)、[go vet](../output/security-smoke/go-vet.log) |
| 原工作目录同步验证 | 后端目标包、27 项前端测试及目标 ESLint 通过 | [后端](../output/security-smoke/original-backend.log)、[前端](../output/security-smoke/original-frontend.log) |
| 最终 Docker 构建 | 成功，更新后 healthy | [构建日志](../output/security-smoke/docker-build-final.log) |

浏览器证据：[用户](../output/security-smoke/browser-user-results.json)、[管理员](../output/security-smoke/browser-admin-results.json)、[第二用户](../output/security-smoke/browser-isolation-results.json)、[现有本地应用](../output/security-smoke/runtime-browser-results.json)。示例截图：[手机客服](../output/playwright/ticket-security-contact-375.png)、[手机深色主题](../output/playwright/ticket-security-mobile-dark.png)、[管理员恶意文本](../output/playwright/ticket-security-admin-xss.png)。

## 边界与清理

本轮覆盖工单功能及相关导航在本地的安全与功能冒烟，不是全站渗透测试或生产攻击测试，也未运行全部仓库测试。真实恶意输入、账号修改、并发写入和 Redis 停机仅发生在隔离实例。测试结束后移除本轮的应用、PostgreSQL、Redis 和独立网络；保留报告、证据、本地应用、现有数据库、数据备份与回滚容器。源码未提交或推送，生产环境未更新。
