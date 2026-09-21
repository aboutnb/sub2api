import { afterEach, describe, expect, it, vi } from 'vitest'
import { STUDIO_CHANNEL, isStudioRequest } from '../../../../shared/image-studio'
import { installStudioBridge } from './bridge'
import apiClient from '@/api/client'
vi.mock('@/api/client', () => ({ default: { get: vi.fn(), request: vi.fn() } }))

describe('image studio bridge', () => {
  afterEach(() => { vi.clearAllMocks(); document.body.innerHTML = '' })
  it('only permits exact image operations', () => {
    const base = { channel: STUDIO_CHANNEL, id: 'test-1', operation: 'request', method: 'POST', body: '{}' }
    expect(isStudioRequest({ ...base, path: '/image-studio/groups/1/images/generations/async' })).toBe(true)
    for (const path of ['/admin/users', 'https://external.test/', '/image-studio/groups/1/images/../responses', '/image-studio/groups/1/images/generations?key=secret']) {
      expect(isStudioRequest({ ...base, path })).toBe(false)
    }
    expect(isStudioRequest({ ...base, path: '/image-studio/groups/1/images/tasks/imgtask_a', method: 'GET', body: undefined })).toBe(true)
  })
  it('rejects spoofed origins and sources, and disposes listeners', async () => {
    const frame = document.createElement('iframe')
    document.body.append(frame)
    const dispose = installStudioBridge(frame, () => true, () => ({ theme: 'light', lang: 'zh' }))
    const data = { channel: STUDIO_CHANNEL, id: 'init', operation: 'bootstrap' }
    window.dispatchEvent(new MessageEvent('message', { data, origin: 'https://external.test', source: frame.contentWindow }))
    window.dispatchEvent(new MessageEvent('message', { data, origin: window.location.origin, source: window }))
    expect(apiClient.get).not.toHaveBeenCalled()
    vi.mocked(apiClient.get).mockResolvedValue({ data: { enabled: true } })
    window.dispatchEvent(new MessageEvent('message', { data, origin: window.location.origin, source: frame.contentWindow }))
    await Promise.resolve()
    expect(apiClient.get).toHaveBeenCalledTimes(1)
    dispose()
    window.dispatchEvent(new MessageEvent('message', { data: { ...data, id: 'again' }, origin: window.location.origin, source: frame.contentWindow }))
    expect(apiClient.get).toHaveBeenCalledTimes(1)
  })
  it('does not call the API after account change', () => {
    const frame = document.createElement('iframe')
    document.body.append(frame)
    const dispose = installStudioBridge(frame, () => false, () => ({ theme: 'light', lang: 'en' }))
    window.dispatchEvent(new MessageEvent('message', { data: { channel: STUDIO_CHANNEL, id: 'init', operation: 'bootstrap' }, origin: window.location.origin, source: frame.contentWindow }))
    expect(apiClient.get).not.toHaveBeenCalled()
    dispose()
  })
  it('aborts in-flight requests on logout and never delivers their results', async () => {
    const frame = document.createElement('iframe')
    document.body.append(frame)
    const post = vi.spyOn(frame.contentWindow!, 'postMessage')
    let resolve!: (value: { data: object }) => void
    vi.mocked(apiClient.get).mockImplementation(() => new Promise((done) => { resolve = done }))
    const dispose = installStudioBridge(frame, () => true, () => ({ theme: 'light', lang: 'zh' }))
    window.dispatchEvent(new MessageEvent('message', { data: { channel: STUDIO_CHANNEL, id: 'pending', operation: 'bootstrap' }, origin: window.location.origin, source: frame.contentWindow }))
    const signal = vi.mocked(apiClient.get).mock.calls[0]?.[1]?.signal
    expect(signal?.aborted).toBe(false)
    dispose()
    expect(signal?.aborted).toBe(true)
    resolve({ data: { enabled: true } })
    await Promise.resolve()
    expect(post).not.toHaveBeenCalled()
  })
})
