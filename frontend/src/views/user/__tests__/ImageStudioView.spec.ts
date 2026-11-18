import { flushPromises, mount } from '@vue/test-utils'
import { nextTick, reactive, ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ImageStudioView from '../ImageStudioView.vue'

const mocks = vi.hoisted(() => ({
  bridge: vi.fn(),
  dispose: vi.fn(),
  auth: null as unknown as { user: { id: number } | null; token: string | null },
}))

vi.mock('@/features/image-studio/bridge', () => ({ installStudioBridge: mocks.bridge }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => mocks.auth }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key, locale: ref('zh') }) }))
vi.mock('@/composables/useThemeMode', () => ({ useThemeMode: () => ({ isDark: ref(false) }) }))
vi.mock('@/components/layout/AppLayout.vue', () => ({ default: { template: '<main style="padding-bottom: 16px"><slot /></main>' } }))

beforeEach(() => {
  vi.clearAllMocks()
  mocks.auth = reactive({ user: { id: 1 }, token: 'session-1' })
  mocks.bridge.mockReturnValue(mocks.dispose)
})

describe('ImageStudioView session lifecycle', () => {
  it('reuses the iframe during token refresh and replaces it on account changes', async () => {
    const wrapper = mount(ImageStudioView, { attachTo: document.body })
    await flushPromises()
    const firstFrame = wrapper.get('iframe').element
    const active = mocks.bridge.mock.calls[0][1] as () => boolean
    expect(active()).toBe(true)

    mocks.auth.token = 'refreshed-session-1'
    await nextTick()
    expect(wrapper.get('iframe').element).toBe(firstFrame)
    expect(mocks.dispose).not.toHaveBeenCalled()

    mocks.auth.user = { id: 2 }
    expect(mocks.dispose).toHaveBeenCalled()
    expect(active()).toBe(false)
    await flushPromises()
    expect(wrapper.get('iframe').element).not.toBe(firstFrame)
    expect(mocks.bridge).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })

  it('destroys the iframe on cross-tab session invalidation and explains recovery', async () => {
    const wrapper = mount(ImageStudioView, { attachTo: document.body })
    await flushPromises()
    const invalidate = mocks.bridge.mock.calls[0][3] as () => void
    invalidate()
    await nextTick()
    expect(wrapper.find('iframe').exists()).toBe(false)
    expect(wrapper.get('[role="status"]').text()).toContain('imageStudio.sessionChanged')
    expect(wrapper.get('button').text()).toBe('imageStudio.reload')
    expect(mocks.dispose).toHaveBeenCalled()
    wrapper.unmount()
  })

  it('removes the iframe and aborts its bridge immediately on logout', async () => {
    const wrapper = mount(ImageStudioView, { attachTo: document.body })
    await flushPromises()
    mocks.auth.token = null
    expect(mocks.dispose).toHaveBeenCalled()
    await nextTick()
    expect(wrapper.find('iframe').exists()).toBe(false)
    wrapper.unmount()
  })
})
