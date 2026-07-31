import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import RedeemView from '@/views/user/RedeemView.vue'

const { getHistory, getPublicSettings } = vi.hoisted(() => ({
  getHistory: vi.fn(),
  getPublicSettings: vi.fn(),
}))

vi.mock('@/api', () => ({
  redeemAPI: { getHistory, redeem: vi.fn() },
  authAPI: { getPublicSettings },
}))
vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: { id: 1, balance: 10, concurrency: 2 },
    refreshUser: vi.fn(),
  }),
}))
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showSuccess: vi.fn(), showError: vi.fn(), showWarning: vi.fn() }),
}))
vi.mock('@/stores/subscriptions', () => ({
  useSubscriptionStore: () => ({ fetchActiveSubscriptions: vi.fn() }),
}))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})
const AppLayoutStub = defineComponent({ template: '<main><slot /></main>' })
const IconStub = defineComponent({ props: ['name'], template: '<i :data-icon="name" />' })

describe('user redeem check-in history tags', () => {
  beforeEach(() => {
    getHistory.mockReset()
    getPublicSettings.mockReset()
    getHistory.mockResolvedValue([
      {
        id: 1,
        code: 'SYS-CHECKIN-1',
        type: 'checkin',
        value: 0.2,
        status: 'used',
        used_at: '2026-07-31T08:00:00Z',
        created_at: '2026-07-31T08:00:00Z',
        checkin_mode: 'lucky',
      },
      {
        id: 2,
        code: 'SYS-CHECKIN-2',
        type: 'checkin',
        value: 0.05,
        status: 'used',
        used_at: '2026-07-30T08:00:00Z',
        created_at: '2026-07-30T08:00:00Z',
        checkin_mode: 'normal',
      },
    ])
    getPublicSettings.mockResolvedValue({ contact_info: '' })
  })

  it('labels only lucky check-ins', async () => {
    const wrapper = mount(RedeemView, {
      global: { stubs: { AppLayout: AppLayoutStub, Icon: IconStub } },
    })
    await flushPromises()

    const tags = wrapper.findAll('[data-testid="lucky-checkin-tag"]')
    expect(tags).toHaveLength(1)
    expect(tags[0].text()).toBe('checkin.lucky')
    expect(wrapper.text()).toContain('+$0.05')
  })
})
