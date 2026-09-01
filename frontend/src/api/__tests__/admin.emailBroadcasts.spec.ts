import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const { post } = vi.hoisted(() => ({ post: vi.fn() }))

vi.mock('@/api/client', () => ({ apiClient: { post } }))

import { create, estimate, type EmailBroadcastPayload } from '@/api/admin/emailBroadcasts'

const payload: EmailBroadcastPayload = {
  title: 'August maintenance',
  subject_zh: '服务器升级通知',
  heading_zh: '服务器升级',
  body_zh: '服务可能短暂不可用。',
  action_zh: '请提前保存工作。',
  subject_en: 'Server upgrade notice',
  heading_en: 'Server upgrade',
  body_en: 'The service may be briefly unavailable.',
  action_en: 'Please save ongoing work.',
  audience: { mode: 'all' }
}

describe('admin email broadcast API', () => {
  beforeEach(() => {
    sessionStorage.clear()
    post.mockReset()
    post.mockResolvedValue({ data: { id: 42, status: 'pending' } })
    vi.spyOn(globalThis.crypto, 'randomUUID').mockReturnValue('11111111-1111-4111-8111-111111111111')
  })

  afterEach(() => vi.restoreAllMocks())

  it('requires a stable idempotency key for task creation', async () => {
    const task = await create(payload)

    expect(post).toHaveBeenCalledTimes(1)
    expect(post.mock.calls[0][0]).toBe('/admin/email-broadcasts')
    expect(post.mock.calls[0][1]).toEqual(payload)
    expect(post.mock.calls[0][2].headers['Idempotency-Key']).toMatch(
      /^email-broadcast-[a-z0-9]+-11111111-1111-4111-8111-111111111111$/
    )
    expect(task).toEqual({ id: 42, status: 'pending' })
    expect(sessionStorage.length).toBe(0)
  })

  it('reuses the same key after an ambiguous network failure', async () => {
    post.mockRejectedValueOnce(new Error('network timeout'))
    await expect(create(payload)).rejects.toThrow('network timeout')
    const firstHeaders = post.mock.calls[0][2].headers

    post.mockResolvedValueOnce({ data: { id: 42, status: 'pending' } })
    await create(payload)

    expect(post.mock.calls[1][2].headers).toEqual(firstHeaders)
    expect(sessionStorage.length).toBe(0)
  })

  it('posts the audience when estimating a filtered broadcast', async () => {
    post.mockResolvedValueOnce({ data: { eligible_recipients: 12 } })

    await expect(estimate({ mode: 'role', roles: ['user'] })).resolves.toEqual({ eligible_recipients: 12 })
    expect(post).toHaveBeenCalledWith('/admin/email-broadcasts/estimate', {
      audience: { mode: 'role', roles: ['user'] }
    })
  })

  it('submits reactivation events and the new audience filters', async () => {
    const reactivationPayload: EmailBroadcastPayload = {
      ...payload,
      event: 'system.reactivation',
      audience: { mode: 'inactive', inactive_days: 7 }
    }

    await create(reactivationPayload)
    expect(post.mock.calls[0][1]).toEqual(reactivationPayload)

    post.mockResolvedValueOnce({ data: { eligible_recipients: 3 } })
    await estimate({ mode: 'inactive', inactive_days: 7 })
    expect(post).toHaveBeenLastCalledWith('/admin/email-broadcasts/estimate', {
      audience: { mode: 'inactive', inactive_days: 7 }
    })
  })
})
