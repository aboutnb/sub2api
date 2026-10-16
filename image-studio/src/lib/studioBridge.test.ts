import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { STUDIO_CHANNEL } from '../../../shared/image-studio'

vi.mock('./canvasImage', async (importOriginal) => ({
  ...await importOriginal<typeof import('./canvasImage')>(),
  validateMaskMatchesImage: vi.fn(async () => 'partial' as const),
  imageDataUrlToPngBlob: vi.fn(async () => new Blob(['image'], { type: 'image/png' })),
}))

const config = {
  enabled: true, storage_namespace: 'image-studio-user-7', async_enabled: true,
  max_body_bytes: 4096, theme: 'dark', lang: 'en',
  prompt_models: [{ group_id: 3, group_name: 'Images', platform: 'openai', model: 'gpt-5.6-sol' }],
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
    const storage = new Map<string, string>()
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => storage.get(key) ?? null,
      setItem: (key: string, value: string) => storage.set(key, value),
      removeItem: (key: string) => storage.delete(key),
    })
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
  it('uses the selected channel model and allows explicitly skipping optimization', async () => {
    const bridge = await initialize()
    parent.postMessage.mockClear()
    parent.postMessage.mockImplementation((request) => reply(request.id, { prompt: 'optimized prompt', optimized: true, model: 'gpt-5.6-sol' }))
    localStorage.setItem(`${bridge.studioStorageNamespace()}:prompt-optimizer`, '3:gpt-5.6-sol')
    await bridge.optimizeStudioPrompt('studio:3:drawing:async', 'original prompt')
    expect(JSON.parse(parent.postMessage.mock.lastCall?.[0].body)).toEqual({ prompt: 'original prompt', group_id: 3, model: 'gpt-5.6-sol' })
    parent.postMessage.mockClear()
    localStorage.setItem(`${bridge.studioStorageNamespace()}:prompt-optimizer`, 'off')
    expect(await bridge.optimizeStudioPrompt('studio:3:drawing:async', 'original prompt')).toEqual({ prompt: 'original prompt', optimized: false })
    expect(parent.postMessage).not.toHaveBeenCalled()
  })
  it('defaults optimization on and scopes the explicit switch to the current user', async () => {
    const bridge = await initialize()
    expect(bridge.studioPromptOptimizationEnabled()).toBe(true)
    localStorage.setItem(`${bridge.studioStorageNamespace()}:prompt-optimizer-enabled`, '0')
    expect(bridge.studioPromptOptimizationEnabled()).toBe(false)
    parent.postMessage.mockClear()
    await expect(bridge.optimizeStudioPrompt('studio:3:drawing:async', 'original prompt')).resolves.toEqual({ prompt: 'original prompt', optimized: false })
    expect(parent.postMessage).not.toHaveBeenCalled()
    localStorage.setItem(`${bridge.studioStorageNamespace()}:prompt-optimizer-enabled`, '1')
    expect(bridge.studioPromptOptimizationEnabled()).toBe(true)
  })
  it('migrates the legacy off value when the user turns optimization back on', async () => {
    const bridge = await initialize()
    localStorage.setItem(`${bridge.studioStorageNamespace()}:prompt-optimizer`, 'off')
    bridge.setStudioPromptOptimizationEnabled(true)
    expect(localStorage.getItem(`${bridge.studioStorageNamespace()}:prompt-optimizer`)).toBeNull()
    parent.postMessage.mockImplementation((request) => reply(request.id, { prompt: 'optimized prompt', optimized: true, model: 'gpt-5.6-sol' }))
    await bridge.optimizeStudioPrompt('studio:3:drawing:async', 'original prompt')
    expect(parent.postMessage.mock.lastCall?.[0].path).toContain('/prompt')
  })
  it.each(['3:%E0%A4%A', '9:gpt-5.6-sol', '3:removed-model', 'invalid'])('rejects a stale or invalid optimizer without silently choosing another channel: %s', async (value) => {
    const bridge = await initialize()
    parent.postMessage.mockClear()
    localStorage.setItem(`${bridge.studioStorageNamespace()}:prompt-optimizer`, value)
    expect(bridge.selectedPromptOptimizer()).toBeUndefined()
    await expect(bridge.optimizeStudioPrompt('studio:3:drawing:async', 'original prompt')).rejects.toThrow('重新选择')
    expect(parent.postMessage).not.toHaveBeenCalled()
  })
  it('keeps optimizer preferences scoped to the signed in user', async () => {
    const bridge = await initialize()
    localStorage.setItem('image-studio-user-8:prompt-optimizer', '3:gpt-5.6-sol')
    expect(bridge.selectedPromptOptimizer()).toBeUndefined()
  })
  it.each([
    { inputs: ['data:image/svg+xml,%3Csvg%3E'], expected: 'PNG' },
    { inputs: ['data:image/png,' + 'x'.repeat(101)], expected: '单张图片' },
    { inputs: Array(3).fill('data:image/png;base64,aW1hZ2U='), expected: '参考图数量' },
  ])('validates input uploads before a prompt request: $expected', async ({ inputs, expected }) => {
    const bridge = await initialize()
    parent.postMessage.mockClear()
    await expect(bridge.validateStudioImageInputs({ id: 'studio:3:drawing:async', model: 'drawing' }, inputs)).rejects.toThrow(expected)
    expect(parent.postMessage).not.toHaveBeenCalled()
  })
  it('validates mask dimensions and converted PNG size before a prompt request', async () => {
    const bridge = await initialize()
    const canvas = await import('./canvasImage')
    const profile = { id: 'studio:3:drawing:async', model: 'drawing' }
    const image = 'data:image/png;base64,aW1hZ2U='
    parent.postMessage.mockClear()
    vi.mocked(canvas.validateMaskMatchesImage).mockRejectedValueOnce(new Error('遮罩尺寸不一致'))
    await expect(bridge.validateStudioImageInputs(profile, [image], image)).rejects.toThrow('遮罩尺寸')
    vi.mocked(canvas.imageDataUrlToPngBlob).mockResolvedValueOnce(new Blob(['x'.repeat(101)], { type: 'image/png' }))
    await expect(bridge.validateStudioImageInputs(profile, [image], image)).rejects.toThrow('单张图片')
    expect(parent.postMessage).not.toHaveBeenCalled()
  })
  it('rejects unsupported edit and mask capabilities before a prompt request', async () => {
    const bridge = await import('./studioBridge')
    parent.postMessage.mockImplementation((request) => reply(request.id, {
      ...config,
      groups: [{ ...config.groups[0], models: [{ ...config.groups[0].models[0], capabilities: { ...config.groups[0].models[0].capabilities, edits: false, mask: false, multiple: false } }] }],
    }))
    await bridge.initializeStudio()
    parent.postMessage.mockClear()
    const profile = { id: 'studio:3:drawing:async', model: 'drawing' }
    await expect(bridge.validateStudioImageInputs(profile, ['data:image/png;base64,aW1hZ2U='])).rejects.toThrow('不支持参考图')
    await expect(bridge.validateStudioImageInputs(profile, [], undefined, 2)).rejects.toThrow('仅支持单图')
    expect(parent.postMessage).not.toHaveBeenCalled()
  })
})
