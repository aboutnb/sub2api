export default {
  authIPBan: {
    title: 'Login Protection',
    description: 'Automatic authentication bans and source review. Browser and automated clients use separate thresholds; API-key gateway traffic is excluded.',
    loadFailed: 'Failed to load login ban records',
    empty: 'No login ban records',
    refresh: 'Refresh',
    searchPlaceholder: 'IP / User-Agent / target / reason',
    filters: {
      status: 'Status',
      all: 'All',
      active: 'Active',
      expired: 'Expired',
      released: 'Released'
    },
    columns: {
      source: 'Source',
      fingerprint: 'Client',
      target: 'Target and Trigger',
      failures: 'Failures',
      timeline: 'Ban Window',
      status: 'Status'
    },
    category: {
      browser: 'Browser',
      automation: 'Automation',
      empty: 'Empty UA',
      other: 'Other Client'
    },
    scope: {
      ip: 'Entire IP',
      ip_ua: 'IP + UA'
    },
    reason: {
      turnstile_token_missing: 'Missing site verification token',
      turnstile_verification_failed: 'Site verification failed',
      credentials_rejected: 'Account or password rejected',
      invalid_login_request: 'Invalid login request',
      invalid_2fa_request: 'Invalid two-factor request',
      two_factor_session_rejected: 'Invalid two-factor session',
      two_factor_code_rejected: 'Incorrect two-factor code',
      login_policy_rejected: 'Login policy rejected',
      auth_request_rejected: 'Authentication request rejected'
    },
    policy: {
      title: 'Automatic ban thresholds',
      summary: '{threshold} attempts / {window} min, restrict for {duration}',
      minutes: '{count} min',
      hours: '{count} hr'
    },
    detail: {
      title: 'Ban Source Details',
      ip: 'IP Address',
      userAgent: 'User-Agent',
      category: 'UA Category',
      scope: 'Restriction Scope',
      target: 'Target Account',
      reason: 'Trigger Reason',
      path: 'Trigger Endpoint',
      firstSeen: 'First Failure',
      lastSeen: 'Latest Failure',
      bannedAt: 'Banned At',
      expiresAt: 'Expires At',
      releasedAt: 'Released At',
      releasedBy: 'Released By',
      banCount: 'Total Ban Count',
      releaseNote: 'Release Note'
    },
    actions: {
      detail: 'View',
      release: 'Release',
      releaseTitle: 'Release Login Ban',
      releaseMessage: 'Release the current restriction for IP {ip}? The record remains available for audit.',
      releaseSuccess: 'Login ban released',
      releaseFailed: 'Failed to release login ban'
    }
  }
}
