import { extractApiErrorCode, extractApiErrorMessage } from '@/utils/apiError'

type TranslateFn = (key: string) => string

const errorTranslationKeys: Record<string, string> = {
  CHECKIN_SOURCE_LIMITED: 'checkin.sourceLimited',
  CHECKIN_RATE_LIMITED: 'checkin.rateLimited',
  CHECKIN_RISK_UNAVAILABLE: 'checkin.riskUnavailable',
  CHECKIN_SECURITY_UNAVAILABLE: 'checkin.securityUnavailable',
  CHECKIN_ENTROPY_UNAVAILABLE: 'checkin.entropyUnavailable',
  CHECKIN_ACCOUNT_TOO_NEW: 'checkin.accountTooNew',
  CHECKIN_NEGATIVE_BALANCE: 'checkin.negativeBalance',
  CHECKIN_GRANT_RESTRICTED: 'checkin.grantRestricted',
  CHECKIN_MODE_DISABLED: 'checkin.modeDisabled',
  CHECKIN_INVALID_REQUEST: 'checkin.invalidRequest',
}

export function checkinErrorMessage(error: unknown, t: TranslateFn, fallback: string): string {
  const code = extractApiErrorCode(error)
  if (code?.startsWith('IDEMPOTENCY_')) return t('checkin.invalidRequest')
  if (code && errorTranslationKeys[code]) return t(errorTranslationKeys[code])
  return extractApiErrorMessage(error, fallback)
}
