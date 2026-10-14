import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import Select from '@/components/common/Select.vue'
import MonitorSettingsPanel from '../MonitorSettingsPanel.vue'
import { CONCRETE_PLATFORM_OPTIONS } from '@/constants/platforms'
import type { MonitorConfig } from '@/api/channelMonitorV2'

const mocks = vi.hoisted(() => ({
  getConfig: vi.fn(),
  updateConfig: vi.fn(),
  groups: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))
vi.mock('@/api/channelMonitorV2', () => ({
  getConfig: mocks.getConfig,
  updateConfig: mocks.updateConfig,
  MONITOR_ERROR_CATEGORIES: ['authentication', 'network'],
}))
vi.mock('@/api/admin', () => ({ adminAPI: { groups: { getAllIncludingInactive: mocks.groups } } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({
  cachedPublicSettings: { channel_monitor_enabled: true },
  showError: mocks.showError,
  showSuccess: mocks.showSuccess,
}) }))
vi.mock('@/utils/featureFlags', () => ({
  getChannelMonitorMode: () => 'v2',
  isChannelMonitorV2Mode: () => true,
}))
vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key, te: () => false }),
}))

function config(): MonitorConfig {
  return {
    version: 2,
    enabled: true,
    refresh_interval_seconds: 60,
    platforms: [
      { platform: 'openai', enabled: false, models: ['gpt-4'] },
      { platform: 'custom-provider', enabled: true, models: ['custom-model'] },
    ],
    group_ids: [42],
    ignored_error_categories: [],
    health_thresholds: {
      minimum_sample: 70, warning_error_rate: 0.1, critical_error_rate: 0.3,
      target_ttft_ms: 1500, warning_ttft_ms: 2000, critical_ttft_ms: 5000,
      warning_cache_rate: 0.7, critical_cache_rate: 0.4,
      error_weight: 0.5, ttft_weight: 0.3, cache_weight: 0.2,
    },
  }
}

async function panel(value = config()) {
  mocks.getConfig.mockResolvedValue(value)
  const wrapper = mount(MonitorSettingsPanel, { global: { stubs: { RouterLink: true, teleport: true } } })
  await flushPromises()
  return wrapper
}

function button(wrapper: Awaited<ReturnType<typeof panel>>, key: string) {
  return wrapper.findAll('button').find((item) => item.text() === 'channelMonitorV2.settings.' + key)!
}

beforeEach(() => {
  vi.clearAllMocks()
  mocks.groups.mockResolvedValue([])
  mocks.updateConfig.mockImplementation(async (value) => JSON.parse(JSON.stringify(value)))
})

describe('MonitorSettingsPanel platform configuration', () => {
  it('shows the complete catalog disabled by default and preserves existing settings', async () => {
    const original = config(), wrapper = await panel(original)
    expect(wrapper.findComponent(Select).exists()).toBe(false)
    const inputs = wrapper.findAll('input[type="text"]')
    expect(inputs.length).toBeGreaterThanOrEqual(CONCRETE_PLATFORM_OPTIONS.length)
    await inputs[0].setValue('gpt-new, gpt-new'); await inputs[0].trigger('change')
    await button(wrapper, 'save').trigger('click'); await flushPromises()
    const saved = mocks.updateConfig.mock.calls[0][0] as MonitorConfig
    expect(saved.platforms[0]).toEqual({platform:'openai',enabled:false,models:['gpt-new']})
    expect(saved.platforms[1]).toEqual(original.platforms[1])
    expect(saved.platforms.filter(p => !original.platforms.some(item => item.platform === p.platform)).every(p => !p.enabled && p.models.length === 0)).toBe(true)
    expect(saved.group_ids).toEqual([42]); expect(original.platforms).toHaveLength(2)
    wrapper.unmount()
  })
  it('offers all providers disabled for an explicitly empty config', async () => {
    const wrapper = await panel({...config(), platforms:[]})
    expect(wrapper.findAll('input[type="text"]')).toHaveLength(CONCRETE_PLATFORM_OPTIONS.length + 1)
    expect(button(wrapper, 'save').attributes('disabled')).toBeDefined(); wrapper.unmount()
  })
  it('restores server configuration when saving fails', async () => {
    const wrapper = await panel(); mocks.updateConfig.mockRejectedValueOnce(new Error('save failed'))
    const input = wrapper.findAll('input[type="text"]')[0]
    await input.setValue('changed-model'); await input.trigger('change')
    await button(wrapper, 'save').trigger('click'); await flushPromises()
    expect(mocks.showError).toHaveBeenCalledOnce()
    expect((wrapper.findAll('input[type="text"]')[0].element as HTMLInputElement).value).toBe('gpt-4')
    expect(button(wrapper, 'save').attributes('disabled')).toBeDefined(); wrapper.unmount()
  })
})
