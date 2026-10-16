import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { STUDIO_CHANNEL, isStudioRequest } from '../../../../shared/image-studio'
import { installStudioBridge } from './bridge'
import apiClient from '@/api/client'
vi.mock('@/api/client', () => ({ default: { get: vi.fn(), request: vi.fn() } }))

describe('image studio bridge', () => {
  const token = (userID: number, version = 1) => `header.${btoa(JSON.stringify({ user_id: userID, version }))}.signature`
  function session(userID: number) {
    localStorage.setItem('auth_token', token(userID))
    localStorage.setItem('auth_user', JSON.stringify({ id: userID }))
  }
  beforeEach(() => { session(1) })
  afterEach(() => { vi.clearAllMocks(); localStorage.clear(); document.body.innerHTML = '' })
  it('only permits exact image operations', () => {
    const base = { channel: STUDIO_CHANNEL, id: 'test-1', operation: 'request', method: 'POST', body: '{}' }
    expect(isStudioRequest({ ...base, path: '/image-studio/groups/1/images/generations/async' })).toBe(true)
    expect(isStudioRequest({ ...base, path: '/image-studio/groups/1/prompt', body: '{"prompt":"cat"}' })).toBe(true)
    expect(isStudioRequest({ ...base, path: '/image-studio/groups/1/prompt', body: { prompt: 'cat' } as never })).toBe(false)
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
  it('aborts on a cross-tab account change before the stale store has updated', async () => {
    const frame = document.createElement('iframe')
    document.body.append(frame)
    const post = vi.spyOn(frame.contentWindow!, 'postMessage')
    const changed = vi.fn()
    let resolve!: (value: { data: object }) => void
    vi.mocked(apiClient.get).mockImplementation(() => new Promise((done) => { resolve = done }))
    const dispose = installStudioBridge(frame, () => true, () => ({ theme: 'light', lang: 'zh' }), changed)
    const data = { channel: STUDIO_CHANNEL, id: 'pending', operation: 'bootstrap' }
    window.dispatchEvent(new MessageEvent('message', { data, origin: window.location.origin, source: frame.contentWindow }))
    const signal = vi.mocked(apiClient.get).mock.calls[0]?.[1]?.signal
    session(2)
    window.dispatchEvent(new StorageEvent('storage', { key: 'auth_user', storageArea: localStorage }))
    expect(signal?.aborted).toBe(true)
    expect(changed).toHaveBeenCalledTimes(1)
    resolve({ data: { storage_namespace: 'image-studio-user-1' } })
    await flushPromises()
    expect(post).not.toHaveBeenCalled()
    window.dispatchEvent(new MessageEvent('message', { data: { ...data, id: 'new' }, origin: window.location.origin, source: frame.contentWindow }))
    expect(apiClient.get).toHaveBeenCalledTimes(1)
    dispose()
  })
  it('does not reset the studio for normal token refresh or user profile updates', async () => {
    const frame = document.createElement('iframe')
    document.body.append(frame)
    const changed = vi.fn()
    vi.mocked(apiClient.get).mockResolvedValue({ data: { enabled: true } })
    const dispose = installStudioBridge(frame, () => true, () => ({ theme: 'light', lang: 'zh' }), changed)
    localStorage.setItem('auth_token', token(1, 2))
    localStorage.setItem('auth_user', JSON.stringify({ id: 1, balance: 2 }))
    window.dispatchEvent(new StorageEvent('storage', { key: 'auth_token', storageArea: localStorage }))
    window.dispatchEvent(new MessageEvent('message', { data: { channel: STUDIO_CHANNEL, id: 'new', operation: 'bootstrap' }, origin: window.location.origin, source: frame.contentWindow }))
    await flushPromises()
    expect(changed).not.toHaveBeenCalled()
    expect(apiClient.get).toHaveBeenCalledTimes(1)
    dispose()
  })
  it('blocks a changed credential at the last request transform before the API sends', () => {
    const frame = document.createElement('iframe')
    document.body.append(frame)
    const changed = vi.fn()
    vi.mocked(apiClient.get).mockImplementation(() => new Promise(() => {}))
    const dispose = installStudioBridge(frame, () => true, () => ({ theme: 'light', lang: 'zh' }), changed)
    window.dispatchEvent(new MessageEvent('message', { data: { channel: STUDIO_CHANNEL, id: 'pending', operation: 'bootstrap' }, origin: window.location.origin, source: frame.contentWindow }))
    const config = vi.mocked(apiClient.get).mock.calls[0]?.[1]
    // The token and user profile are separate storage writes. Block the interval between them too.
    localStorage.setItem('auth_token', token(2))
    const transforms = config?.transformRequest as Array<(data: unknown) => unknown>
    expect(() => transforms.at(-1)?.(undefined)).toThrow('Image studio session changed')
    expect(config?.signal?.aborted).toBe(true)
    expect(changed).toHaveBeenCalledOnce()
    dispose()
  })
  it('preserves actionable gateway error codes', async () => {
    const frame = document.createElement('iframe')
    document.body.append(frame)
    const post = vi.spyOn(frame.contentWindow!, 'postMessage')
    vi.mocked(apiClient.request).mockRejectedValue({ status: 403, reason: 'INSUFFICIENT_BALANCE', error: { code: 'balance_exhausted', message: 'Balance exhausted', type: 'billing_error' } })
    const dispose = installStudioBridge(frame, () => true, () => ({ theme: 'light', lang: 'zh' }))
    window.dispatchEvent(new MessageEvent('message', { data: { channel: STUDIO_CHANNEL, id: 'failed', operation: 'request', path: '/image-studio/groups/1/images/generations', method: 'POST', body: '{}' }, origin: window.location.origin, source: frame.contentWindow }))
    await flushPromises()
    expect(post).toHaveBeenCalledWith({ channel: STUDIO_CHANNEL, id: 'failed', status: 403, data: { error: { code: 'balance_exhausted', message: 'Balance exhausted', type: 'billing_error', reason: 'INSUFFICIENT_BALANCE' } } }, window.location.origin)
    dispose()
  })
  it('submits a request only once and aborts only its matching cancellation', async () => {
    const frame = document.createElement('iframe')
    document.body.append(frame)
    const post = vi.spyOn(frame.contentWindow!, 'postMessage')
    let resolve!: (value: { status: number; data: object }) => void
    vi.mocked(apiClient.request).mockImplementation(() => new Promise((done) => { resolve = done }))
    const dispose = installStudioBridge(frame, () => true, () => ({ theme: 'light', lang: 'zh' }))
    const data = { channel: STUDIO_CHANNEL, id: 'generate', operation: 'request', path: '/image-studio/groups/1/images/generations/async', method: 'POST', body: '{}' }
    const dispatch = (message: object) => window.dispatchEvent(new MessageEvent('message', { data: message, origin: window.location.origin, source: frame.contentWindow }))
    dispatch(data)
    dispatch(data)
    expect(apiClient.request).toHaveBeenCalledOnce()
    const signal = vi.mocked(apiClient.request).mock.calls[0]?.[0]?.signal
    dispatch({ channel: STUDIO_CHANNEL, id: 'unrelated', operation: 'cancel' })
    expect(signal?.aborted).toBe(false)
    dispatch({ channel: STUDIO_CHANNEL, id: data.id, operation: 'cancel' })
    expect(signal?.aborted).toBe(true)
    resolve({ status: 200, data: { id: 'imgtask_1' } })
    await flushPromises()
    expect(post).not.toHaveBeenCalled()
    dispose()
  })
})
