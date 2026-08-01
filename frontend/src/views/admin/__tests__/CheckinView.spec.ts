import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import CheckinView from '@/views/admin/CheckinView.vue'

const {
  getConfig,
  getOverview,
  getRecords,
  updateConfig,
  fetchPublicSettings,
  showSuccess,
  showError,
} = vi.hoisted(() => ({
  getConfig: vi.fn(),
  getOverview: vi.fn(),
  getRecords: vi.fn(),
  updateConfig: vi.fn(),
  fetchPublicSettings: vi.fn(),
  showSuccess: vi.fn(),
  showError: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    checkin: { getConfig, getOverview, getRecords, updateConfig },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ fetchPublicSettings, showSuccess, showError }),
}))

vi.mock('@/composables/useStepUp', () => ({
  useStepUp: () => ({
    visible: { value: false },
    blockedReason: { value: '' },
    run: (action: () => Promise<unknown>) => action(),
    prompt: vi.fn(),
    onVerified: vi.fn(),
    onCancel: vi.fn(),
  }),
  isStepUpBlocked: () => false,
  isStepUpCancelled: () => false,
  stepUpBlockReason: () => '',
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key, locale: { value: 'zh' } }),
  }
})

const AppLayoutStub = defineComponent({ template: '<main><slot /></main>' })

const config = {
  enabled: false,
  normal_enabled: true,
  lucky_enabled: true,
  normal_min: '0.01',
  normal_max: '0.05',
  lucky_reward_type: 'multiplier' as const,
  lucky_positive_probability: '65',
  lucky_multiplier_positive_tiers: [
    { min: '0.01', max: '0.10', weight: '70' },
    { min: '0.11', max: '0.15', weight: '20' },
    { min: '0.16', max: '0.20', weight: '10' },
  ],
  lucky_amount_positive_tiers: [{ min: '0.01', max: '0.10', weight: '100' }],
  lucky_min_multiplier: '-0.08',
  lucky_max_multiplier: '0.20',
  lucky_amount_min: '-0.05',
  lucky_amount_max: '0.10',
  risk_control_enabled: true,
  min_account_age_hours: 24,
  ip_window_minutes: 10,
  ip_max_users: 20,
  unrecharged_reduction_enabled: true,
  unrecharged_checkin_threshold: 3,
  unrecharged_normal_reward_percent: '50',
  config_version: 1,
  updated_at: '2026-07-26T00:00:00Z',
}

function mountView() {
  return mount(CheckinView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        Icon: true,
        LoadingSpinner: true,
        Pagination: true,
        Toggle: true,
        TotpStepUpDialog: true,
      },
    },
  })
}

describe('CheckinView configuration errors', () => {
  beforeEach(() => {
    for (const fn of [getConfig, getOverview, getRecords, updateConfig, fetchPublicSettings, showSuccess, showError]) fn.mockReset()
    getConfig.mockResolvedValue({ ...config })
    getOverview.mockResolvedValue({ business_date: '2026-07-26', total: 0, normal_count: 0, lucky_count: 0, positive_total: 0, negative_total: 0 })
    getRecords.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 20, pages: 1 })
    updateConfig.mockResolvedValue({ ...config, enabled: true, config_version: 2 })
    fetchPublicSettings.mockResolvedValue(undefined)
  })

  it('serializes edited decimal inputs before saving', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="lucky-positive-probability"]').setValue('65')
    await wrapper.get('[data-testid="lucky-min-multiplier"]').setValue('-0.5')
    await wrapper.get('[data-testid="lucky-max-multiplier"]').setValue('2')
    await wrapper.get('[data-testid="positive-tier-2"]').findAll('input')[1].setValue('2')
    await wrapper.get('textarea').setValue(' test change ')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(updateConfig).toHaveBeenCalledWith(expect.objectContaining({
      normal_enabled: true,
      lucky_enabled: true,
      normal_min: '0.01',
      normal_max: '0.05',
      lucky_positive_probability: '65',
      lucky_multiplier_positive_tiers: [
        { min: '0.01', max: '0.10', weight: '70' },
        { min: '0.11', max: '0.15', weight: '20' },
        { min: '0.16', max: '2', weight: '10' },
      ],
      lucky_amount_positive_tiers: [{ min: '0.01', max: '0.10', weight: '100' }],
      lucky_min_multiplier: '-0.5',
      lucky_max_multiplier: '2',
      lucky_amount_min: '-0.05',
      lucky_amount_max: '0.10',
      unrecharged_reduction_enabled: true,
      unrecharged_checkin_threshold: 3,
      unrecharged_normal_reward_percent: '50',
      change_reason: 'test change',
      expected_config_version: 1,
    }))
    wrapper.unmount()
  })

  it('calculates effective tier probabilities from the positive total', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('[data-testid="positive-tier-probability-0"]').text()).toBe('45.5%')
    expect(wrapper.get('[data-testid="positive-tier-probability-1"]').text()).toBe('13%')
    expect(wrapper.get('[data-testid="positive-tier-probability-2"]').text()).toBe('6.5%')
    expect(wrapper.get('[data-testid="positive-tier-total"]').text()).toBe('65%')
    wrapper.unmount()
  })

  it('disables a mode configuration area when its independent switch is off', async () => {
    getConfig.mockResolvedValueOnce({ ...config, normal_enabled: false, lucky_enabled: true })
    const wrapper = mountView()
    await flushPromises()

    const decimalInputs = wrapper.findAll<HTMLInputElement>('input[type="number"]')
    expect(decimalInputs[0].attributes('disabled')).toBeDefined()
    expect(decimalInputs[1].attributes('disabled')).toBeDefined()
    expect(wrapper.get('[data-testid="lucky-positive-probability"]').attributes('disabled')).toBeUndefined()
    expect(wrapper.get('[data-testid="lucky-min-multiplier"]').attributes('disabled')).toBeUndefined()
    expect(wrapper.get('[data-testid="lucky-max-multiplier"]').attributes('disabled')).toBeUndefined()
    wrapper.unmount()
  })

  it('shows save failures as a toast without reporting a load failure', async () => {
    updateConfig.mockRejectedValueOnce({ reason: 'CHECKIN_CONFIG_INVALID', message: 'invalid configuration' })
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('textarea').setValue('test change')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(showError).toHaveBeenCalledWith('admin.checkin.configInvalid')
    expect(wrapper.text()).not.toContain('admin.checkin.loadFailed')
    wrapper.unmount()
  })

  it('blocks an empty change reason before sending a request', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(updateConfig).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledWith('admin.checkin.reasonRequired')
    wrapper.unmount()
  })
})
