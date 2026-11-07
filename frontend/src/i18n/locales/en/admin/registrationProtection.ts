export default {
  registrationProtection: {
    releaseSourceHint: 'Release this temporary source restriction and reset its failure counter. Successful signup quotas and other active restrictions remain. The action is audited.',
    title: 'Registration Protection', description: 'Control bulk signups by source and review risk records and account restrictions. Existing logins and balances are unaffected.',
    tabs: { settings: 'Protection settings', sources: 'Source records', accounts: 'Risk accounts' },
    enabled: 'Enable enhanced registration protection', quotaHint: 'Only created accounts consume quotas; failed attempts do not. Email and OAuth signups share quotas. Disabling protection retains basic throttling, CAPTCHA and existing signup benefit decisions.',
    observationHint: 'Source quotas, active source blocks, or matching browser features reaching both account and distinct-IP thresholds reject signup before email delivery and persist an email block. Blocked inbox aliases cannot retry from another IP. Features or missing identifiers alone do not trigger a block.',
    fields: {
      browser_success_limit: 'Successful signups per signed browser', observe_fingerprint_limit: 'Cross-IP signup blocking account threshold', fingerprint_window_minutes: 'Browser feature window (minutes)', fingerprint_distinct_ips: 'Minimum distinct IPs for signup blocking',
      identity_success_limit: 'Successful signup limit per IP + UA', ip_success_limit: 'Total successful signup limit per IP', success_window_hours: 'Successful signup window (hours)',
      identity_failure_limit: 'Verification failure limit per IP + UA', ip_failure_limit: 'Total verification failure limit per IP', failure_window_minutes: 'Verification failure window (minutes)',
      block_minutes: 'Temporary source restriction (minutes)', observe_identity_success_limit: 'IP + UA observation threshold'
    },
    loading: 'Loading…', save: 'Save settings', retry: 'Reload', saved: 'Registration protection settings saved', recordType: 'Record type',
    kinds: { sources: 'Successful signup sources', events: 'Risk events', blocks: 'Temporary source restrictions' },
    search: 'Search', searchHint: 'IP / UA / email', status: 'Status', all: 'All', empty: 'No matching records', source: 'Signup source', account: 'Account', reason: 'Reason and status',
    created: 'Recorded at', actions: 'Actions', details: 'View details', concurrency: 'Current concurrency', release: 'Review release', restrict: 'Restrict usage',
    total: '{total} records · Page {page}', previous: 'Previous', next: 'Next', review: 'Review risk account',
    reviewHint: 'Restriction sets concurrency to -1. Release restores the saved value; if it was already -1, API usage remains blocked. Refresh and investigate if another operation changed the account state.',
    note: 'Review note (required)', cancel: 'Cancel', confirm: 'Confirm review', reviewed: 'Account review completed', failed: 'Operation failed. Please try again later.',
    statuses: { observed: 'Needs review', restricted: 'Usage restricted', released: 'Released', active: 'Active', expired: 'Expired' },
    reasons: { browser_features_registration_observed: 'Burst of signups with similar browser features across IPs; needs review', duplicate_signup_identity: 'Repeated signup source restricted by existing benefit policy', admin_review: 'Administrator review', identity_registration_observed: 'Multiple signups from the same source', registration_failure_limit: 'Too many failed registration verifications', registration_source_quota: 'Successful signup source quota reached' },
    detailFields: { browser_hash: 'Signed browser identifier digest', feature_hash: 'Browser feature digest (auxiliary signal)', ip_address: 'IP address', user_agent: 'User-Agent', email: 'Email', user_id: 'User ID', reason: 'Risk reason', status: 'Status', concurrency: 'Current concurrency', previous_concurrency: 'Concurrency restored on release', trigger_path: 'Trigger endpoint', created_at: 'Recorded at', expires_at: 'Expires at', reviewed_at: 'Reviewed at', reviewed_by: 'Reviewer ID', review_note: 'Review note', release_note: 'Release note' }
  }
}
