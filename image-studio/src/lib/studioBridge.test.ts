import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { STUDIO_CHANNEL } from '../../../shared/image-studio'

const config = {
  enabled: true, storage_namespace: 'image-studio-user-7', async_enabled: true,
  max_body_bytes: 4096, theme: 'dark', lang: 'en',
  groups: [{ id: 3, name: 'Images', subscription_type: 'standard', rate_multiplier: 1,
    models: [{ id: 'drawing', platform: 'openai', capabilities: { edits: true, mask: true, multiple: true, max_input_images: 2, max_upload_bytes: 100, input_types: ['image/png'] } }] }],
}

describe('embedded studio client', () => {
  let surface: EventTarget
  let parent: { postMessage: ReturnType<typeof vi.fn> }
  function reply(id: string, data: unknown, status = 200, source: unknown = parent, origin = 'https://site.test') {
    surface.dispatchEvent(Object.assign(new Event('message'), { source, origin, data: { channel: STUDIO_CHANNEL, id, status, data } }))
  }
  beforeEach(() => {
    vi.resetModules()
    surface = new EventTarget()
    parent = { postMessage: vi.fn() }
    vi.stubGlobal('window', Object.assign(surface, { parent, location: { origin: 'https://site.test' } }))
    vi.stubGlobal('document', { documentElement: { classList: { toggle: vi.fn() }, style: { setProperty: vi.fn(), removeProperty: vi.fn() }, lang: '' } })
  })
  afterEach(() => { vi.unstubAllGlobals() })
  async function initialize() {
    const bridge = await import('./studioBridge')
    parent.postMessage.mockImplementation((request) => { if (request.operation === 'bootstrap') reply(request.id, config) })
    await bridge.initializeStudio()
    return bridge
  }
  it('ignores forged bootstrap replies and uses only the server user namespace', async () => {
    const bridge = await import('./studioBridge')
    const pending = bridge.initializeStudio()
    const id = parent.postMessage.mock.calls[0][0].id
    reply(id, config, 200, {}, 'https://site.test')
    reply(id, config, 200, parent, 'https://attacker.test')
    await Promise.resolve()
    expect(bridge.studioConfig()).toBeUndefined()
    reply(id, config)
    await pending
    expect(bridge.studioStorageNamespace()).toBe('image-studio-user-7')
    expect(document.documentElement.lang).toBe('en')
  })
  it('rejects arbitrary URLs and preserves asynchronous failure without resubmitting', async () => {
    const bridge = await initialize()
    parent.postMessage.mockClear()
    parent.postMessage.mockImplementation((request) => reply(request.id, { error: { message: 'Insufficient balance' } }, 403))
    await expect(bridge.studioFetch('https://other.test/images/generations', { method: 'POST', body: '{}' })).rejects.toThrow('Unsupported image endpoint')
    const response = await bridge.studioFetch('/image-studio/groups/3/images/generations/async', { method: 'POST', body: '{"model":"drawing"}' })
    expect(response.status).toBe(403)
    expect(parent.postMessage).toHaveBeenCalledTimes(1)
    expect(parent.postMessage.mock.calls[0][0].path).toContain('/async')
  })
  it('validates upload constraints and forwards no credentials', async () => {
    const bridge = await initialize()
    const form = new FormData()
    form.append('model', 'drawing')
    form.append('image[]', new Blob(['image'], { type: 'image/png' }), 'input.png')
    parent.postMessage.mockImplementation((request) => reply(request.id, { data: [] }))
    await bridge.studioFetch('/image-studio/groups/3/images/edits', { method: 'POST', body: form, headers: { Authorization: 'must-not-forward' } })
    expect(JSON.stringify(parent.postMessage.mock.lastCall?.[0])).not.toContain('must-not-forward')
    form.append('image[]', new Blob(['bad'], { type: 'image/svg+xml' }))
    await expect(bridge.studioFetch('/image-studio/groups/3/images/edits', { method: 'POST', body: form })).rejects.toThrow('PNG')
  })
  it('sends cancellation and does not accept results after cancellation', async () => {
    const bridge = await initialize()
    parent.postMessage.mockClear()
    parent.postMessage.mockImplementation(() => {})
    const controller = new AbortController()
    const request = bridge.studioFetch('/image-studio/groups/3/images/tasks/imgtask_pending', { signal: controller.signal })
    controller.abort()
    await expect(request).rejects.toThrow('Aborted')
    expect(parent.postMessage.mock.lastCall?.[0].operation).toBe('cancel')
  })
  it('applies only allowlisted RGB theme tokens from the parent', async () => {
    await initialize()
    surface.dispatchEvent(Object.assign(new Event('message'), { source: parent, origin: 'https://site.test', data: {
      channel: STUDIO_CHANNEL, operation: 'appearance', theme: 'light', lang: 'zh',
      tokens: { 'bg-canvas': '255 250 244', 'action': 'url(https://other.test)', 'text-muted': '999 0 0', unknown: '1 2 3' },
    } }))
    expect(document.documentElement.style.setProperty).toHaveBeenCalledExactlyOnceWith('--av-rgb-bg-canvas', '255 250 244')
    expect(document.documentElement.style.removeProperty).toHaveBeenCalledWith('--av-rgb-action')
  })
})
