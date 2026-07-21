export default {
  authIPBan: {
    title: '登录防护',
    description: '认证入口自动封禁与来源审查。浏览器和自动化客户端采用不同阈值，API Key 网关请求不参与统计。',
    loadFailed: '加载登录封禁记录失败',
    empty: '暂无登录封禁记录',
    refresh: '刷新',
    searchPlaceholder: 'IP / User-Agent / 目标账号 / 原因',
    filters: {
      status: '状态',
      all: '全部',
      active: '封禁中',
      expired: '已过期',
      released: '已解除'
    },
    columns: {
      source: '来源',
      fingerprint: '客户端识别',
      target: '目标与触发',
      failures: '失败次数',
      timeline: '封禁时间',
      status: '状态'
    },
    category: {
      browser: '浏览器',
      automation: '自动化脚本',
      empty: '空 UA',
      other: '其他客户端'
    },
    scope: {
      ip: '整 IP',
      ip_ua: 'IP + UA'
    },
    reason: {
      turnstile_token_missing: '缺少站点验证令牌',
      turnstile_verification_failed: '站点验证失败',
      credentials_rejected: '账号或密码校验失败',
      invalid_login_request: '登录请求格式无效',
      invalid_2fa_request: '二次验证请求无效',
      two_factor_session_rejected: '二次验证会话无效',
      two_factor_code_rejected: '二次验证码错误',
      login_policy_rejected: '登录策略拒绝',
      auth_request_rejected: '认证请求被拒绝'
    },
    policy: {
      title: '当前自动封禁阈值',
      summary: '{threshold} 次 / {window} 分钟，限制 {duration}',
      minutes: '{count} 分钟',
      hours: '{count} 小时'
    },
    detail: {
      title: '封禁来源详情',
      ip: 'IP 地址',
      userAgent: 'User-Agent',
      category: 'UA 分类',
      scope: '限制范围',
      target: '目标账号',
      reason: '触发原因',
      path: '触发接口',
      firstSeen: '首次失败',
      lastSeen: '最近失败',
      bannedAt: '封禁时间',
      expiresAt: '到期时间',
      releasedAt: '解除时间',
      releasedBy: '解除人',
      banCount: '累计封禁次数',
      releaseNote: '解除说明'
    },
    actions: {
      detail: '查看',
      release: '解除',
      releaseTitle: '解除登录封禁',
      releaseMessage: '确认解除 IP {ip} 的当前限制？记录会保留在后台审计中。',
      releaseSuccess: '已解除该登录封禁',
      releaseFailed: '解除登录封禁失败'
    }
  }
}
