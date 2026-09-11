export default {
  registrationProtection: {
    releaseSourceHint: '解除本条来源临时限制并清除对应失败计数。成功注册配额和其他有效限制保持不变，操作会记录在安全审计中。',
    title: '注册防护', description: '按注册来源控制批量建号，集中查看风险记录与账号限制。不会限制已有用户登录或更改余额。',
    tabs: { settings: '防护配置', sources: '来源记录', accounts: '风险账号' },
    enabled: '启用注册增强防护', quotaHint: '数量按成功创建账号统计，失败重试不占名额。普通注册和 OAuth 新建账号共用额度。关闭增强防护仍保留基础限流、人机验证和既有注册赠送判定。',
    observationHint: '观察阈值只记录风险，不直接限制账号。低于观察阈值的成功注册上限会优先生效。UA 相同不代表同一人。',
    fields: {
      identity_success_limit: '同 IP＋UA 成功注册上限', ip_success_limit: '同 IP 成功注册总上限', success_window_hours: '成功注册统计窗口（小时）',
      identity_failure_limit: '同 IP＋UA 验证失败上限', ip_failure_limit: '同 IP 验证失败总上限', failure_window_minutes: '验证失败统计窗口（分钟）',
      block_minutes: '来源临时限制时长（分钟）', observe_identity_success_limit: '同 IP＋UA 风险观察数量'
    },
    loading: '加载中…', save: '保存配置', retry: '重新加载', saved: '注册防护配置已保存', recordType: '记录类型',
    kinds: { sources: '成功注册来源', events: '风险事件', blocks: '来源临时限制' },
    search: '查询', searchHint: 'IP / UA / 邮箱', status: '状态', all: '全部', empty: '暂无匹配记录', source: '注册来源', account: '账号', reason: '原因与状态',
    created: '记录时间', actions: '操作', details: '查看详情', concurrency: '当前并发', release: '审核解除', restrict: '限制使用',
    total: '共 {total} 条 · 第 {page} 页', previous: '上一页', next: '下一页', review: '风险账号审核',
    reviewHint: '限制使用会将并发设为 -1。解除时恢复记录中保存的并发值；若原值就是 -1，解除风险标记后仍不能调用。账号状态已被其他操作修改时，请刷新后重新核查。',
    note: '审核说明（必填）', cancel: '取消', confirm: '确认审核', reviewed: '风险账号审核已完成', failed: '操作失败，请稍后重试',
    statuses: { observed: '待审查', restricted: '已限制使用', released: '已解除', active: '限制中', expired: '已到期' },
    reasons: { duplicate_signup_identity: '重复注册来源，既有赠送风控限制', admin_review: '管理员审核', identity_registration_observed: '同来源多账号注册', registration_failure_limit: '注册验证失败过多', registration_source_quota: '成功注册来源额度已满' },
    detailFields: { ip_address: 'IP 地址', user_agent: 'User-Agent', email: '邮箱', user_id: '用户 ID', reason: '风险原因', status: '处理状态', concurrency: '当前并发', previous_concurrency: '解除后恢复并发', trigger_path: '触发接口', created_at: '记录时间', expires_at: '到期时间', reviewed_at: '审核时间', reviewed_by: '审核人 ID', review_note: '审核说明', release_note: '解除说明' }
  }
}
