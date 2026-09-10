import { describe, expect, it } from 'vitest'

import zhAdminCheckin from '../locales/zh/admin/checkin'
import zhCheckin from '../locales/zh/checkin'

describe('check-in Chinese terminology', () => {
  it('uses 运气签到 consistently in user and admin copy', () => {
    expect(zhCheckin.checkin.lucky).toBe('运气签到')
    expect(zhCheckin.checkin.luckyConfirmTitle).toBe('确认运气签到')
    expect(zhAdminCheckin.checkin.lucky).toBe('运气签到')
    expect(zhAdminCheckin.checkin.luckyEnabled).toBe('启用运气签到')

    const allCheckinCopy = JSON.stringify([zhCheckin, zhAdminCheckin])
    expect(allCheckinCopy).not.toContain('幸运签到')
  })
})
