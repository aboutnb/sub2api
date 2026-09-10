import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const { post } = vi.hoisted(() => ({
  post: vi.fn()
}))

vi.mock('@/api/client', () => ({
  apiClient: { post }
}))

import { checkIn } from '@/api/checkin'

describe('check-in API security', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
    localStorage.setItem('auth_user', JSON.stringify({ id: 7 }))
    post.mockReset()
    post.mockResolvedValue({ data: { newly_checked_in: true, record: { id: 1 } } })
    vi.spyOn(globalThis.crypto, 'randomUUID').mockReturnValue('11111111-1111-4111-8111-111111111111')
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('sends an idempotency key scoped to the business date', async () => {
    await checkIn('normal', '2026-07-26')

    expect(post).toHaveBeenCalledWith('/user/checkin', { mode: 'normal' }, {
      headers: {
        'Idempotency-Key': 'checkin-2026-07-26-11111111-1111-4111-8111-111111111111'
      }
    })
    expect(sessionStorage.length).toBe(0)
  })

  it('sends the optional Turnstile token in a request header', async () => {
    await checkIn('normal', '2026-07-26', 'turnstile-proof')

    expect(post).toHaveBeenCalledWith('/user/checkin', { mode: 'normal' }, {
      headers: {
        'Idempotency-Key': 'checkin-2026-07-26-11111111-1111-4111-8111-111111111111',
        'X-Turnstile-Token': 'turnstile-proof'
      }
    })
  })

  it('reuses the same key after an ambiguous failure and a page reload', async () => {
    post.mockRejectedValueOnce(new Error('network timeout'))
    await expect(checkIn('lucky', '2026-07-26')).rejects.toThrow('network timeout')
    const firstHeaders = post.mock.calls[0][2].headers

    vi.resetModules()
    post.mockResolvedValueOnce({ data: { newly_checked_in: true, record: { id: 1 } } })
    const { checkIn: checkInAfterReload } = await import('@/api/checkin')
    await checkInAfterReload('lucky', '2026-07-26')

    expect(post.mock.calls[1][2].headers).toEqual(firstHeaders)
    expect(sessionStorage.length).toBe(0)
  })

  it('does not reuse a key on the next business date', async () => {
    post.mockRejectedValueOnce(new Error('network timeout'))
    await expect(checkIn('normal', '2026-07-26')).rejects.toThrow('network timeout')
    const firstHeaders = post.mock.calls[0][2].headers

    vi.mocked(globalThis.crypto.randomUUID).mockReturnValueOnce('22222222-2222-4222-8222-222222222222')
    await checkIn('normal', '2026-07-27')

    expect(post.mock.calls[1][2].headers).not.toEqual(firstHeaders)
  })
})
