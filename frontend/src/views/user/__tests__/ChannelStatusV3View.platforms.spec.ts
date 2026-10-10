import { defineComponent, h } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ChannelStatusV3View from '../ChannelStatusV3View.vue'

const { getMatrix } = vi.hoisted(() => ({ getMatrix: vi.fn() }))
vi.mock('@/api/channelMonitorV2', () => ({
  getMatrix,
  getSnapshot: vi.fn().mockResolvedValue({ coverage: {}, metrics: {}, trend: [], config: { refresh_interval_seconds: 300 } }),
}))
vi.mock('@/api/groups', () => ({ default: {
  getAvailable: vi.fn().mockResolvedValue([]), getUserGroupRates: vi.fn().mockResolvedValue({}),
} }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError: vi.fn() }) }))
vi.mock('vue-i18n', async importOriginal => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key, locale: { value: 'en' } }),
}))

describe('channel status platform cards', () => {
  it('keeps separate card instances when the same composite group serves multiple platforms', async () => {
    getMatrix.mockResolvedValue({ items: [
      { platform: 'openai', group_id: 7 }, { platform: 'deepseek', group_id: 7 },
    ] })
    const card = defineComponent({
      props: ['row'],
      setup(props) {
        const originalPlatform = props.row.platform
        return () => h('div', { 'data-platform': props.row.platform, 'data-original': originalPlatform })
      },
    })
    const wrapper = mount(ChannelStatusV3View, { global: { stubs: {
      AppLayout: { template: '<div><slot /></div>' }, Icon: true, EmptyState: true, ChannelMonitorV3Card: card,
    } } })
    await flushPromises()
    expect(wrapper.findAll('[data-platform]')).toHaveLength(2)
    getMatrix.mockResolvedValue({ items: [{ platform: 'deepseek', group_id: 7 }] })
    await wrapper.get('button[title="common.refresh"]').trigger('click')
    await flushPromises()
    const remaining = wrapper.get('[data-platform="deepseek"]')
    expect(remaining.attributes('data-original')).toBe('deepseek')
    expect(wrapper.findAll('[data-platform]')).toHaveLength(1)
    wrapper.unmount()
  })
})
