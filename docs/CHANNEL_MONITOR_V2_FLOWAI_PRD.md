# 渠道状态 V2/V3 在 `sub2api-flowai` 分支的二开 PRD

> 状态：Draft（用于技术评审）
> 分支：`sub2api-flowai`
> 编写日期：2026-09-02
> 范围：用户端渠道状态展示层；不改变现有 V2 被动聚合数据管线

## 1. 现状与结论

### 1.1 当前分支已经具备的能力

`sub2api-flowai` 当前 HEAD 已包含渠道监控 V2 的后端和用户端 Ops 页面：

- `backend/internal/service/channel_monitor_v2.go`：时间范围、固定桶、指标、健康评分、覆盖进度和隐私字段定义。
- `backend/internal/service/channel_monitor_v2_aggregator.go`：单实例锁、尾部重算、首轮种子、渐进式历史回填和失败退避。
- `backend/internal/repository/channel_monitor_v2_repo.go`：`snapshot`、`matrix`、`models`、`errors`、`users`、`dimensions` 只读查询，以及分组/用户权限裁剪。
- `backend/internal/server/routes/user.go`：用户端 `/channel-monitor-v2/*` 路由仅在监控启用且 `channel_monitor_mode=v2` 时开放。
- `frontend/src/views/user/ChannelStatusV2View.vue`：完整筛选、趋势、模型、错误、用户明细的 Ops 风格 V2 页面。
- `frontend/src/views/admin/ChannelMonitorView.vue` 与 `frontend/src/features/channel-monitor-v2/MonitorSettingsPanel.vue`：V2 配置和 V1 回退入口。

因此，本项目不是从零移植 V2 后端；二开重点是**在既有 V2 数据契约上增加紧凑的 V3 用户展示层**，并保持 V2 页面可回退。

### 1.2 线上参考页面

2026-09-02 通过浏览器观察到的 [Kedaya 线上渠道监控页面](https://kedaya.ai/monitor) 具备以下交互契约：

- 顶部显示更新时间、手动刷新、`90m / 24h / 7d / 30d` 四个范围和总体可用率/缓存率。
- 每个分组一张卡片：提供商图标、分组名、用户倍率、正常/降级/失败状态。
- 卡片展示缓存率、可用率、首 Token（TTFT）三个最新指标。
- 卡片底部展示固定数量的历史脉冲条；90m 为 18 个 5 分钟桶，悬停/聚焦显示时间、可用率、缓存率和 TTFT。
- 页面按响应式网格排列，数据不足时保留空桶占位；回填未完成时需要明确标识。

该页面的 V2 数据聚合实现和 V3 紧凑视图分别可参考 [V2 合并提交](https://github.com/kiss-kedaya/sub2api/commit/909cbb9b1) 与 [紧凑 V3 视图提交](https://github.com/kiss-kedaya/sub2api/commit/568bf90b417fa374ac6a3604c8008b4d196f704b)。

## 2. 目标与非目标

### 2.1 目标

1. 在 FlowAI 的现有 V2 API 上提供与线上页面一致的“分组卡片 + 最新桶指标 + 脉冲时间轴”用户视图。
2. 保留现有完整 V2 Ops 页面，支持无后端开关的诊断和回滚。
3. 复用 FlowAI 现有 provider 图标、分组倍率接口、主题变量和国际化体系，不引入全局视觉重构。
4. 维持现有用户分组权限、吞吐量隐私和服务端特性开关边界。
5. 在数据回填、空数据、请求取消、自动刷新和移动端布局下保持可解释且稳定的状态。

### 2.2 非目标

- 不重新设计或替换 V2 聚合器、rollup 表、健康评分算法和错误分类法。
- 不在用户端增加主动探针、写入监控数据或改变渠道真实流量。
- 不把管理员的错误明细、用户排行、RPM/TPM 透传到普通用户 V3 页面。
- 不直接移植上游 Soft Glass 的全局 `style.css`、Tailwind、部署版本或无关业务改动。
- 不新增数据库迁移；如验证发现现有迁移缺失，另开数据迁移任务处理。

## 3. 产品规格

### 3.1 路由与模式

`frontend/src/views/user/ChannelStatusView.vue` 按以下优先级渲染：

| 条件 | 页面 |
| --- | --- |
| 监控关闭或路由不可用 | 沿用现有路由守卫行为 |
| `channel_monitor_mode=v1` | `ChannelStatusV1View.vue`，行为不变 |
| `channel_monitor_mode=v2` 且无查询参数 | 新增 `ChannelStatusV3View.vue` |
| `channel_monitor_mode=v2` 且 `monitor_view=v2` | 现有 `ChannelStatusV2View.vue`，作为诊断/回滚视图 |

`monitor_view=v2` 只影响展示层，不绕过后端模式守卫；V2 API 不可用时不能通过前端参数强行访问。

### 3.2 顶部工具栏

- 默认范围：`90m`。
- 范围切换：`90m`、`24h`、`7d`、`30d`；切换后重新请求 `snapshot` 和 `matrix`。
- 总览文本：取 `snapshot.trend` 中最新完成桶，展示总体可用率（`1 - error_rate`）和缓存率。
- 显示 `coverage.data_through`；`coverage_complete=false` 或 `bootstrap.active=true` 时显示部分覆盖/回填状态。
- 刷新按钮触发非静默刷新；自动刷新使用服务端 `refresh_interval_seconds`，页面隐藏时暂停请求并继续在可见时恢复。
- 倒计时只反映下一次刷新，不作为数据新鲜度的唯一判断。

### 3.3 分组卡片

卡片数据来自 `matrix?group_by=platform_group`，每行只保留有效 `group_id > 0` 的分组。

字段与语义：

| 字段 | 来源/规则 |
| --- | --- |
| 提供商图标/标签 | 复用 `useChannelMonitorFormat` 和现有 `ProviderIcon.vue` |
| 分组名 | `row.group_name`；缺失时显示本地化“未知分组” |
| 用户倍率 | `/groups/available` 与 `/groups/rates` 合并；自定义倍率优先，接口失败显示 `-`，不阻塞监控 |
| 状态 | 最新桶 `health.overall`：`healthy=正常`、`warning=降级`、`critical=失败`、其他=未知 |
| 缓存率 | 最新桶 `metrics.cache_rate`，按百分比格式化 |
| 可用率 | 最新桶 `1 - metrics.error_rate`，按百分比格式化并按阈值着色 |
| 首 Token | 最新桶 `metrics.ttft.p50_ms`，毫秒/秒自适应格式化 |

重要语义：卡片主指标使用**最新完成桶**，而不是用户选择范围的聚合值；范围只决定时间轴和服务端查询窗口。没有完成桶时回退到 `row.metrics/row.health`，并显示未知或占位值。

### 3.4 历史脉冲时间轴

| 范围 | 目标条数 | 后端桶 |
| --- | ---: | --- |
| `90m` | 18 | 5 分钟 |
| `24h` | 24 | 1 小时 |
| `7d` | 14 | 12 小时 |
| `30d` | 30 | 1 天 |

- 按 `bucket_start` 升序排列，数据不足时在左侧补未知占位条。
- 每条高度代表该桶是否有有效样本；颜色代表可用率阈值，未知数据使用中性灰。
- 桌面端悬停、键盘聚焦均显示 Teleport tooltip；tooltip 至少包含时间、可用率、缓存率、TTFT，并限制在视口内。
- 命中区域尺寸固定，动画只作用于视觉层，避免悬停放大导致相邻条目跳动。
- 低于阈值的颜色分段必须集中在 `monitorFormat.ts` 的纯函数中，禁止在模板内散落数字阈值。

### 3.5 空态、加载和错误

- 首次加载且无行：显示固定高度 skeleton，避免网格抖动。
- 查询成功但无分组：显示空态，不显示误导性的“全部正常”。
- 单次刷新失败：保留上一份成功数据，并显示可关闭错误提示；首次失败显示加载失败空态。
- `AbortError`/Axios cancel 不提示错误；组件卸载时清理请求和定时器。
- 回填进行中时不阻塞卡片渲染，使用覆盖提示说明历史数据可能不完整。

## 4. 数据与权限契约

### 4.1 现有接口复用

用户端只读接口：

- `GET /channel-monitor-v2/snapshot?range=...`
- `GET /channel-monitor-v2/matrix?range=...&group_by=platform_group`
- 可选维度：`GET /channel-monitor-v2/dimensions`
- 用户倍率：`GET /groups/available`、`GET /groups/rates`

前端复用 `frontend/src/api/channelMonitorV2.ts` 的重复 query 参数序列化，不新建第二套 DTO。`MonitorMatrixRow.buckets` 是时间轴唯一数据源；`snapshot.trend` 只用于顶部总体摘要。

### 4.2 服务端安全边界

- 用户请求经过 `channelMonitorModeV2Guard`，必须同时满足监控启用和 `channel_monitor_mode=v2`。
- `ChannelMonitorV2Handler.scopeFilter` 从认证主体取得可用分组，设置 `RestrictGroups=true`；客户端提交的 `group_id` 只能与服务端允许集合求交。
- 服务端根据 `channel_monitor_hide_throughput` 裁剪普通用户 RPM/TPM；V3 不渲染这些字段。
- 错误详情、用户排行和管理员明细只保留在现有完整 V2/管理员页面。
- 无可见分组时返回空 `items`，前端显示空态而非泄漏其他分组存在性。

### 4.3 数据新鲜度和回填

V2 聚合器继续负责 5m/1h/12h/1d rollup、10 分钟尾部重算、首次 2 小时种子和渐进式历史回填。V3 只消费 `coverage`：

- `coverage.data_through`：最后已聚合时间，用于更新时间。
- `coverage.coverage_complete`：所选窗口起点是否已经覆盖。
- `coverage.bootstrap`：首次升级到 30d 产品窗口的回填进度。

## 5. 技术改动清单

### 5.1 新增文件

- `frontend/src/views/user/ChannelStatusV3View.vue`：范围状态、双请求加载、分组行筛选、自动刷新、空态和回填提示。
- `frontend/src/components/user/monitor/ChannelMonitorV3Card.vue`：提供商、分组、倍率、状态和三个最新指标。
- `frontend/src/components/user/monitor/ChannelMonitorV3Timeline.vue`：固定条数、颜色、占位、tooltip、键盘可访问和悬停动画。
- `frontend/src/i18n/locales/zh/channelMonitorV3.ts`、`frontend/src/i18n/locales/en/channelMonitorV3.ts`：标题、状态、范围、tooltip、空态和错误文案。

### 5.2 修改文件

- `frontend/src/views/user/ChannelStatusView.vue`：增加 V3 默认分支和 `monitor_view=v2` 回退判断。
- `frontend/src/features/channel-monitor-v2/monitorFormat.ts`：增加 `availabilityBadgeClass`、`availabilityBarClass`、`availabilityTextClass` 等纯函数及边界测试。
- `frontend/src/i18n/locales/{zh,en}/index.ts`：注册 V3 locale。
- 如需样式，仅在 V3 组件使用 scoped CSS 和现有 CSS 变量；不得改动全局 body、Tailwind 主色或与 FlowAI 无关的页面。

### 5.3 明确复用而不改动的模块

- `frontend/src/api/channelMonitorV2.ts`
- `frontend/src/api/groups.ts`
- `frontend/src/composables/useChannelMonitorFormat.ts`
- `frontend/src/components/user/monitor/ProviderIcon.vue`
- `backend/internal/service/channel_monitor_v2_*`
- `backend/internal/repository/channel_monitor_v2_*`
- `backend/internal/server/routes/user.go` 的 V2 守卫和只读路由

## 6. 配置、迁移与发布

### 6.1 配置

现有设置已覆盖：

- `channel_monitor_enabled`
- `channel_monitor_mode`（合法值 `v1`/`v2`，缺失或非法默认 `v1`）
- `channel_monitor_default_interval_seconds`
- `channel_monitor_hide_throughput`（普通用户默认 fail-closed）
- `channel_monitor_show_quota`

发布顺序：先部署代码并保持 `mode=v1`，确认迁移和静态资源完整后，再由管理员在 V2 配置面板启用 `mode=v2`。V3 默认视图若出现展示回归，可直接使用 `?monitor_view=v2`，无需改数据库。

### 6.2 数据库

本次不新增迁移。上线前必须确认 `194` 至 `206` 中的 V2 表、rollup 权限、5 分钟刷新和隐私默认值已在目标环境执行；`226_channel_monitor_quota_mode.sql` 等后续 FlowAI 迁移按现有部署顺序执行。

### 6.3 回滚

1. 展示层问题：将用户 URL 加上 `monitor_view=v2`，保留 V2 API 和聚合器运行。
2. 业务风险：管理员将 `channel_monitor_mode` 切回 `v1`，后端停止被动聚合并恢复 V1 页面。
3. 数据异常：先查看 `coverage`、聚合器日志和 rollup 水位，不通过清表或重跑迁移修复展示问题。

## 7. 测试与验收

### 7.1 前端自动化

新增或扩展以下测试：

- `ChannelStatusView`：v1、v2+回退参数、v2默认 V3 三种路由分支。
- `ChannelMonitorV3Card`：最新桶优先、健康状态映射、可用率/缓存率/TTFT 格式化、倍率缺失回退。
- `ChannelMonitorV3Timeline`：四种条数、左侧占位、时间排序、阈值颜色、tooltip 文案、键盘聚焦和取消请求。
- `monitorFormat.spec.ts`：可用率边界 `<30/<50/<60/<80/<90/>=90` 的确定性测试。
- 现有 `channelMonitorV2.spec.ts`：保持重复 query 参数和 `platform_group` 请求契约。

执行目录和命令：

```bash
cd frontend
pnpm exec vitest run src/features/channel-monitor-v2/__tests__/monitorFormat.spec.ts
pnpm exec vitest run src/components/user/monitor/__tests__/ChannelMonitorV3Card.spec.ts
pnpm exec vitest run src/components/user/monitor/__tests__/ChannelMonitorV3Timeline.spec.ts
pnpm run typecheck
pnpm run lint:check
```

### 7.2 后端回归

保持现有 V2 handler、repository、aggregator 和路由守卫测试通过；至少执行：

```bash
cd backend
go test ./internal/handler/... ./internal/repository/... ./internal/service/... ./internal/server/routes/...
```

### 7.3 人工验收矩阵

- V1 模式页面无变化，V2 模式默认进入紧凑卡片页面。
- `90m/24h/7d/30d` 条数分别为 18/24/14/30；切换范围后卡片不抖动。
- 卡片主指标与最新完成桶一致；顶部摘要与 `snapshot.trend` 最新桶一致。
- 普通用户只能看到自己可用分组；提交未授权 `group_id` 不会扩大结果集。
- RPM/TPM、错误明细、用户明细不出现在 V3 用户页面。
- 回填未完成、无数据、请求失败、手动刷新、自动刷新、浏览器隐藏/恢复和移动端宽度均有可解释状态。
- 键盘可操作刷新、范围切换和时间轴 tooltip；文本不溢出卡片。

## 8. 风险与待确认项

1. **视觉基准**：附件截图为深色卡片，线上页面为较浅的 Soft Glass 风格；本 PRD 采用“线上布局与数据语义 + FlowAI 现有主题变量”的折中，最终颜色需在 FlowAI 实例截图评审后定稿。
2. **分组排序**：默认按 `group_id` 升序，避免请求量波动造成卡片位置跳动；如产品要求运营优先级，应增加服务端显式排序字段，而不是前端按实时指标排序。
3. **聚合刷新间隔**：前端倒计时跟随 `config.refresh_interval_seconds`，不硬编码线上示例中的秒数。
4. **倍率接口可用性**：倍率是辅助信息，接口失败不能阻断监控主数据；需要确认 `groups/available` 在所有订阅用户角色下的授权行为。
5. **最新桶为空**：需要产品确认是显示 `-`、状态“未知”，还是沿用范围聚合值；本 PRD 采用 `-`/未知并回退到行级字段的保守方案。

## 9. Definition of Done

- PRD 评审确认 V3 默认展示、V2 回退参数和深浅主题策略。
- 新增 V3 组件、国际化和路由分支，未改动 V2 后端数据管线。
- 前端目标测试、类型检查、lint 和后端 V2 回归测试通过。
- 在已配置 V2 的 FlowAI 环境完成桌面/移动端验收，并记录 `coverage`、刷新和权限结果。
- 发布说明明确启用设置、回退 URL 和无新增迁移。
