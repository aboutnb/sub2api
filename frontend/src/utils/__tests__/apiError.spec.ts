import { describe, expect, it } from 'vitest'
import { extractApiErrorCode, extractI18nErrorMessage } from '@/utils/apiError'

describe('API error localization', () => {
  it('prefers the semantic reason over a numeric HTTP code', () => {
    expect(extractApiErrorCode({
      code: 503,
      reason: 'IDENTITY_VERIFICATION_REQUIRED',
    })).toBe('IDENTITY_VERIFICATION_REQUIRED')
  })

  it('shows the localized invoice reason instead of the generic upstream message', () => {
    const messages: Record<string, string> = {
      'payment.errors.IDENTITY_VERIFICATION_REQUIRED': '开票服务尚未完成实名认证，请联系站点管理员处理。',
    }
    const t = (key: string) => messages[key] ?? key

    expect(extractI18nErrorMessage(
      {
        code: 503,
        reason: 'IDENTITY_VERIFICATION_REQUIRED',
        message: 'invoice service request failed',
        metadata: { request_id: 'req-owner-verification' },
      },
      t,
      'payment.errors',
      '开票订单校验失败',
    )).toBe('开票服务尚未完成实名认证，请联系站点管理员处理。')
  })
})
