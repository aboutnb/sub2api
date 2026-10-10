import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import MonitorSettingsPanel from '../MonitorSettingsPanel.vue'
import { CONCRETE_PLATFORM_OPTIONS } from '@/constants/platforms'
import { monitorPlatformLabel } from '@/constants/monitorPlatforms'
import type { MonitorConfig } from '@/api/channelMonitorV2'

const { getConfig, updateConfig } = vi.hoisted(() => ({ getConfig: vi.fn(), updateConfig: vi.fn() }))
vi.mock('@/api/channelMonitorV2', async importOriginal => ({
  ...await importOriginal<typeof import('@/api/channelMonitorV2')>(),
  getConfig,
  updateConfig,
}))
vi.mock('@/api/admin', () => ({ adminAPI: { groups: { getAllIncludingInactive: vi.fn().mockResolvedValue([]) } } }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn() }) }))
vi.mock('@/utils/featureFlags', () => ({ getChannelMonitorMode: () => 'v2', isChannelMonitorV2Mode: () => true }))
vi.mock('vue-i18n', async importOriginal => ({
  ...await importOriginal<typeof import('vue-i18n')>(),
  useI18n: () => ({ t: (key: string) => key, te: () => false }),
}))

describe('passive monitor platform settings', () => {
  beforeEach(() => {
    getConfig.mockResolvedValue({
      version: 7, enabled: true, refresh_interval_seconds: 300,
      platforms: [
        { platform: 'openai', enabled: true, models: ['gpt-custom'] },
        { platform: 'deepseek', enabled: false, models: ['deepseek-custom'] },
        { platform: 'custom_provider', enabled: false, models: ['private-model'] },
      ],
      group_ids: [42], ignored_error_categories: [], health_thresholds: {},
    })
    updateConfig.mockReset().mockImplementation(async value => value)
  })

  it('shows every platform for an old config and saves newly enabled platforms without changing existing settings', async () => {
    const wrapper = mount(MonitorSettingsPanel, { global: { stubs: { Icon: true, RouterLink: true } } })
    await flushPromises()
    for (const option of CONCRETE_PLATFORM_OPTIONS) {
      expect(wrapper.find(`[role="switch"][aria-label="${monitorPlatformLabel(option.value)} channelMonitorV2.settings.enableTitle"]`).exists()).toBe(true)
    }
    const deepseek = wrapper.get('[role="switch"][aria-label^="DeepSeek "]')
    expect(deepseek.attributes('aria-checked')).toBe('false')
    const typesafe = wrapper.get('[role="switch"][aria-label^="TypeSafe / Jev "]')
    expect(typesafe.attributes('aria-checked')).toBe('false')
    await typesafe.trigger('click')
    const save = wrapper.findAll('button').find(button => button.text().includes('channelMonitorV2.settings.save'))!
    await save.trigger('click')
    await flushPromises()
    const saved = updateConfig.mock.calls[0][0] as MonitorConfig
    expect(saved.platforms.find(p => p.platform === 'typesafe')).toEqual({ platform: 'typesafe', enabled: true, models: [] })
    expect(saved.platforms.find(p => p.platform === 'openai')).toEqual({ platform: 'openai', enabled: true, models: ['gpt-custom'] })
    expect(saved.platforms.find(p => p.platform === 'deepseek')).toEqual({ platform: 'deepseek', enabled: false, models: ['deepseek-custom'] })
    expect(saved.platforms.find(p => p.platform === 'custom_provider')).toEqual({ platform: 'custom_provider', enabled: false, models: ['private-model'] })
    expect(saved.group_ids).toEqual([42])
    expect(saved.ignored_error_categories).toEqual([])
    expect(saved.version).toBe(7)
    wrapper.unmount()
  })
})
