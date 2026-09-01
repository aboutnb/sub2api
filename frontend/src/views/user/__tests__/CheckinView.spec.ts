import { defineComponent } from 'vue'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import CheckinView from '@/views/user/CheckinView.vue'
import LuckyCheckinConfirmDialog from '@/components/checkin/LuckyCheckinConfirmDialog.vue'
import type { CheckinRecord, CheckinStatus } from '@/types'

const { getStatus, getRecords, checkIn, refreshUser, showSuccess, showError, authStore } = vi.hoisted(() => ({
  getStatus: vi.fn(),
  getRecords: vi.fn(),
  checkIn: vi.fn(),
  refreshUser: vi.fn(),
  showSuccess: vi.fn(),
  showError: vi.fn(),
  authStore: {
    user: { id: 1, role: 'user', username: 'tester', email: 'tester@example.com', balance: 9.5639 },
    refreshUser: vi.fn(),
  },
}))

authStore.refreshUser = refreshUser

vi.mock('@/api/checkin', () => ({ checkinAPI: { getStatus, getRecords, checkIn } }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => authStore }))
vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showSuccess, showError }) }))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key, locale: { value: 'zh' } }),
  }
})

const AppLayoutStub = defineComponent({ template: '<main><slot /></main>' })
const IconStub = defineComponent({ props: ['name'], template: '<i :data-icon="name" />' })
const TurnstileStub = defineComponent({
  props: ['siteKey', 'size'],
  emits: ['verify', 'expire', 'error'],
  setup(_, { expose }) {
    expose({ reset: vi.fn() })
  },
  template: '<button data-testid="turnstile-stub" type="button" @click="$emit(\'verify\', \'turnstile-proof\')" />',
})

const negativeRecord: CheckinRecord = {
  id: 1,
  user_id: 1,
  checkin_date: '2026-07-26',
  mode: 'lucky',
  reward_type: 'multiplier',
  random_value: -0.04661,
  reward_amount: -0.4661,
  balance_before: 10,
  balance_after: 9.5339,
  checked_in_at: '2026-07-26T06:24:00+08:00',
}

const positiveRecord: CheckinRecord = {
  ...negativeRecord,
  id: 2,
  checkin_date: '2026-07-25',
  mode: 'normal',
  random_value: 0.03,
  reward_amount: 0.03,
  balance_before: 9.97,
  balance_after: 10,
  checked_in_at: '2026-07-25T08:00:00+08:00',
}

const status: CheckinStatus = {
  enabled: true,
  normal_enabled: true,
  lucky_enabled: true,
  lucky_reward_type: 'multiplier',
  lucky_min_multiplier: -0.05,
  lucky_max_multiplier: 0.1,
  eligible: true,
  can_check_in: true,
  unavailable_reason: '',
  business_date: '2026-07-26',
  timezone: 'Asia/Shanghai',
  server_time: '2026-07-26T08:00:00+08:00',
  next_reset_at: '2026-07-27T00:00:00+08:00',
  checked_in_today: false,
  today_record: null,
  days_in_month: 31,
  first_weekday: 2,
}

const wrappers: VueWrapper[] = []

function mountView() {
  const wrapper = mount(CheckinView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        Icon: IconStub,
        LoadingSpinner: true,
        TurnstileWidget: TurnstileStub,
      },
    },
  })
  wrappers.push(wrapper)
  return wrapper
}

describe('user CheckinView', () => {
  beforeEach(() => {
    for (const fn of [getStatus, getRecords, checkIn, refreshUser, showSuccess, showError]) fn.mockReset()
    getStatus.mockResolvedValue({ ...status })
    getRecords.mockResolvedValue({ items: [negativeRecord, positiveRecord], total: 2, page: 1, page_size: 100, pages: 1 })
    checkIn.mockResolvedValue({ newly_checked_in: true, record: positiveRecord })
    refreshUser.mockResolvedValue(authStore.user)
  })

  afterEach(() => {
    while (wrappers.length) wrappers.pop()?.unmount()
  })

  it('renders normal and lucky modes as one compact action row', async () => {
    const wrapper = mountView()
    await flushPromises()

    const actions = wrapper.get('[data-testid="checkin-mode-actions"]')
    expect(wrapper.get('[data-testid="checkin-summary"]').element.contains(actions.element)).toBe(true)
    expect(actions.classes()).toContain('grid-cols-2')
    expect(actions.findAll('button')).toHaveLength(2)
    expect(actions.text()).not.toContain('checkin.normalHint')
    expect(actions.text()).not.toContain('checkin.luckyHint')
  })

  it('shows each recorded balance change in its calendar day', async () => {
    const wrapper = mountView()
    await flushPromises()

    const luckyDay = wrapper.get('[data-testid="calendar-day-26"]')
    const normalDay = wrapper.get('[data-testid="calendar-day-25"]')
    expect(luckyDay.text()).toContain('-$0.47')
    expect(luckyDay.text()).toContain('-0.05x')
    expect(luckyDay.attributes('data-mode')).toBe('lucky')
    expect(luckyDay.find('[data-icon="sparkles"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="history-multiplier-1"]').text()).toBe('-0.05x')
    expect(normalDay.text()).toContain('+$0.03')
    expect(normalDay.attributes('data-mode')).toBe('normal')
    expect(normalDay.find('[data-icon="gift"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="calendar-day-24"]').text()).not.toContain('$')
  })

  it('summarizes monthly check-ins by mode below the calendar', async () => {
    const wrapper = mountView()
    await flushPromises()

    const summary = wrapper.get('[data-testid="calendar-month-summary"]')
    expect(summary.text()).toContain('checkin.daysCount')
    expect(summary.text()).toContain('-$0.44')
    expect(summary.text()).toContain('+$0.03')
    expect(summary.text()).toContain('-$0.47')
  })

  it('compacts large calendar amounts so they cannot overflow a day cell', async () => {
    getRecords.mockResolvedValue({
      items: [{ ...positiveRecord, id: 3, checkin_date: '2026-07-23', random_value: 11375097247.43, reward_amount: 11375097247.43 }],
      total: 1,
      page: 1,
      page_size: 100,
      pages: 1,
    })
    const wrapper = mountView()
    await flushPromises()

    const day = wrapper.get('[data-testid="calendar-day-23"]')
    expect(day.text()).toContain('+$113.75亿')
    expect(day.text()).not.toContain('11375097247.43')
  })

  it('requires confirmation and shows the possible multiplier before lucky check-in', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="checkin-mode-lucky"]').trigger('click')
    await flushPromises()

    expect(checkIn).not.toHaveBeenCalled()
    expect(document.body.querySelector('[data-testid="lucky-multiplier-range"]')?.textContent).toContain('-0.05x ~ +0.10x')

    const confirm = document.body.querySelector<HTMLButtonElement>('[data-testid="confirm-lucky-checkin"]')
    expect(confirm).not.toBeNull()
    confirm?.click()
    await flushPromises()

    expect(checkIn).toHaveBeenCalledWith('lucky', status.business_date)
  })

  it('shows the settled multiplier after a lucky check-in succeeds', async () => {
    checkIn.mockResolvedValue({ newly_checked_in: true, record: negativeRecord })
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-testid="checkin-mode-lucky"]').trigger('click')
    await flushPromises()
    getStatus.mockResolvedValue({
      ...status,
      can_check_in: false,
      checked_in_today: true,
      today_record: negativeRecord,
    })
    wrapper.getComponent(LuckyCheckinConfirmDialog).vm.$emit('confirm')
    await flushPromises()

    expect(wrapper.text()).toContain('checkin.success')
    const multiplier = wrapper.get('[data-testid="checkin-result-multiplier"]')
    expect(multiplier.text()).toContain('checkin.resultMultiplier')
    expect(multiplier.text()).toContain('-0.05x')
  })

  it('does not label fixed lucky rewards as multipliers', async () => {
    const fixedLuckyRecord: CheckinRecord = {
      ...negativeRecord,
      reward_type: 'amount',
      random_value: 0.08,
      reward_amount: 0.08,
    }
    getStatus.mockResolvedValue({ ...status, lucky_reward_type: 'amount' })
    getRecords.mockResolvedValue({ items: [fixedLuckyRecord], total: 1, page: 1, page_size: 100, pages: 1 })
    checkIn.mockResolvedValue({ newly_checked_in: true, record: fixedLuckyRecord })
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('[data-testid="calendar-day-26"]').text()).not.toContain('x')
    expect(wrapper.find('[data-testid="history-multiplier-1"]').exists()).toBe(false)

    await wrapper.get('[data-testid="checkin-mode-lucky"]').trigger('click')
    await flushPromises()
    getStatus.mockResolvedValue({
      ...status,
      lucky_reward_type: 'amount',
      can_check_in: false,
      checked_in_today: true,
      today_record: fixedLuckyRecord,
    })
    wrapper.getComponent(LuckyCheckinConfirmDialog).vm.$emit('confirm')
    await flushPromises()

    expect(wrapper.text()).toContain('checkin.success')
    expect(wrapper.find('[data-testid="checkin-result-multiplier"]').exists()).toBe(false)
  })

  it('renders only the enabled mode and uses a single-column action layout', async () => {
    getStatus.mockResolvedValue({ ...status, lucky_enabled: false })
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('[data-testid="checkin-mode-actions"]').classes()).toContain('grid-cols-1')
    expect(wrapper.find('[data-testid="checkin-mode-normal"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="checkin-mode-lucky"]').exists()).toBe(false)
  })

  it('shows an unavailable state when both modes are disabled', async () => {
    getStatus.mockResolvedValue({
      ...status,
      normal_enabled: false,
      lucky_enabled: false,
      can_check_in: false,
      unavailable_reason: 'no_modes_enabled',
    })
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('checkin.noModesAvailable')
    expect(wrapper.find('[data-testid="checkin-mode-actions"]').exists()).toBe(false)
  })

  it('requires and submits a Turnstile token when check-in protection is enabled', async () => {
    getStatus.mockResolvedValue({
      ...status,
      turnstile_enabled: true,
      turnstile_site_key: 'site-key',
    })
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('[data-testid="checkin-turnstile"]').exists()).toBe(true)
    expect(checkIn).not.toHaveBeenCalled()
    await wrapper.get('[data-testid="turnstile-stub"]').trigger('click')
    await wrapper.get('[data-testid="checkin-mode-normal"]').trigger('click')
    await flushPromises()

    expect(checkIn).toHaveBeenCalledWith('normal', status.business_date, 'turnstile-proof')
  })
})
