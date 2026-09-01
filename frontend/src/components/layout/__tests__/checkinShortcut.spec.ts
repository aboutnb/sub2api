import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineComponent } from 'vue'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import CheckinShortcut from '@/components/checkin/CheckinShortcut.vue'
import type { CheckinRecord, CheckinStatus } from '@/types'

const {
  getStatus,
  checkIn,
  refreshUser,
  fetchPublicSettings,
  showSuccess,
  showError,
  push,
  authStore,
} = vi.hoisted(() => ({
  getStatus: vi.fn(),
  checkIn: vi.fn(),
  refreshUser: vi.fn(),
  fetchPublicSettings: vi.fn(),
  showSuccess: vi.fn(),
  showError: vi.fn(),
  push: vi.fn(),
  authStore: {
    user: { id: 1, role: 'admin', username: 'admin', email: 'admin@example.com', balance: 10 },
    isSimpleMode: false,
    refreshUser: vi.fn(),
  },
}))

authStore.refreshUser = refreshUser

vi.mock('@/api/checkin', () => ({ checkinAPI: { getStatus, checkIn } }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => authStore }))
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ fetchPublicSettings, showSuccess, showError }),
}))
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

const IconStub = defineComponent({
  props: ['name'],
  template: '<i :data-icon="name" />',
})
const TurnstileStub = defineComponent({
  props: ['siteKey', 'size'],
  emits: ['verify', 'expire', 'error'],
  setup(_, { expose }) {
    expose({ reset: vi.fn() })
  },
  template: '<button data-testid="turnstile-stub" type="button" @click="$emit(\'verify\', \'turnstile-proof\')" />',
})

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

const record: CheckinRecord = {
  id: 7,
  user_id: 1,
  checkin_date: '2026-07-26',
  mode: 'normal',
  reward_type: 'amount',
  random_value: 0.03,
  reward_amount: 0.03,
  balance_before: 10,
  balance_after: 10.03,
  checked_in_at: '2026-07-26T08:00:01+08:00',
}

const wrappers: VueWrapper[] = []

function mountShortcut() {
  const wrapper = mount(CheckinShortcut, {
    global: {
      stubs: {
        Icon: IconStub,
        RouterLink: defineComponent({
          props: ['to'],
          template: '<a :href="to"><slot /></a>',
        }),
        TurnstileWidget: TurnstileStub,
      },
    },
  })
  wrappers.push(wrapper)
  return wrapper
}

describe('check-in header shortcut', () => {
  beforeEach(() => {
    for (const fn of [getStatus, checkIn, refreshUser, fetchPublicSettings, showSuccess, showError, push]) fn.mockReset()
    authStore.user = { id: 1, role: 'admin', username: 'admin', email: 'admin@example.com', balance: 10 }
    authStore.isSimpleMode = false
    getStatus.mockResolvedValue({ ...status })
    checkIn.mockResolvedValue({ newly_checked_in: true, record: { ...record } })
    refreshUser.mockResolvedValue(authStore.user)
    fetchPublicSettings.mockResolvedValue(undefined)
  })

  afterEach(() => {
    while (wrappers.length) wrappers.pop()?.unmount()
  })

  it('is placed immediately before the balance in the app header', () => {
    const dir = dirname(fileURLToPath(import.meta.url))
    const headerSource = readFileSync(resolve(dir, '../AppHeader.vue'), 'utf8')
    const shortcutIndex = headerSource.indexOf('<CheckinShortcut')
    const balanceIndex = headerSource.indexOf('<!-- Balance Display -->')

    expect(shortcutIndex).toBeGreaterThan(-1)
    expect(balanceIndex).toBeGreaterThan(shortcutIndex)
    expect(headerSource).toContain('<CheckinShortcut v-if="user && !authStore.isSimpleMode" />')
  })

  it('shows quick check-in for an eligible administrator', async () => {
    const wrapper = mountShortcut()
    await flushPromises()

    expect(wrapper.get('[data-testid="checkin-shortcut"]').text()).toContain('checkin.quickAction')
    expect(wrapper.get('[data-testid="checkin-shortcut-actions"]').classes()).toContain('xl:flex')
    expect(wrapper.get('[data-testid="quick-checkin-normal"]').text()).toContain('checkin.normal')
    expect(wrapper.get('[data-testid="quick-checkin-lucky"]').text()).toContain('checkin.lucky')
  })

  it('uses the compact borderless header treatment in every state', async () => {
    const wrapper = mountShortcut()
    await flushPromises()

    for (const testId of ['quick-checkin-normal', 'quick-checkin-lucky', 'checkin-shortcut']) {
      const classes = wrapper.get(`[data-testid="${testId}"]`).classes()
      expect(classes).toContain('h-8')
      expect(classes).toContain('rounded-xl')
      expect(classes).not.toContain('border')
    }

    getStatus.mockResolvedValue({
      ...status,
      checked_in_today: true,
      can_check_in: false,
      today_record: { ...record },
    })
    window.dispatchEvent(new Event('focus'))
    await flushPromises()

    const checkedClasses = wrapper.get('[data-testid="checkin-shortcut"]').classes()
    expect(checkedClasses).toContain('h-8')
    expect(checkedClasses).toContain('rounded-xl')
    expect(checkedClasses).not.toContain('border')
  })

  it.each(['normal', 'lucky'] as const)('submits the selected %s mode', async (mode) => {
    const wrapper = mountShortcut()
    await flushPromises()

    await wrapper.get('[data-testid="checkin-shortcut"]').trigger('click')
    await wrapper.get(`[data-testid="quick-checkin-${mode}"]`).trigger('click')
    if (mode === 'lucky') {
      expect(checkIn).not.toHaveBeenCalled()
      document.body.querySelector<HTMLButtonElement>('[data-testid="confirm-lucky-checkin"]')?.click()
    }
    await flushPromises()

    expect(checkIn).toHaveBeenCalledWith(mode, status.business_date)
    expect(refreshUser).toHaveBeenCalledTimes(1)
    expect(showSuccess).toHaveBeenCalledWith('checkin.success')
  })

  it('keeps both modes available from the compact header menu', async () => {
    const wrapper = mountShortcut()
    await flushPromises()

    await wrapper.get('[data-testid="checkin-shortcut"]').trigger('click')
    expect(wrapper.get('[data-testid="quick-checkin-menu-normal"]').text()).toContain('checkin.normal')
    expect(wrapper.get('[data-testid="quick-checkin-menu-lucky"]').text()).toContain('checkin.lucky')
  })

  it('hides disabled modes and hides the shortcut when no mode is enabled', async () => {
    getStatus
      .mockResolvedValueOnce({ ...status, lucky_enabled: false })
      .mockResolvedValueOnce({ ...status, normal_enabled: false, lucky_enabled: false, can_check_in: false, unavailable_reason: 'no_modes_enabled' })
    const wrapper = mountShortcut()
    await flushPromises()

    expect(wrapper.find('[data-testid="quick-checkin-normal"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="quick-checkin-lucky"]').exists()).toBe(false)

    window.dispatchEvent(new Event('focus'))
    await flushPromises()

    expect(wrapper.find('[data-testid="checkin-shortcut"]').exists()).toBe(false)
  })

  it('ignores repeated clicks while a check-in request is pending', async () => {
    let resolveCheckin!: (value: { newly_checked_in: boolean; record: CheckinRecord }) => void
    checkIn.mockReturnValue(new Promise(resolve => { resolveCheckin = resolve }))
    const wrapper = mountShortcut()
    await flushPromises()

    await wrapper.get('[data-testid="checkin-shortcut"]').trigger('click')
    const normal = wrapper.get('[data-testid="quick-checkin-normal"]')
    void normal.trigger('click')
    void normal.trigger('click')

    expect(checkIn).toHaveBeenCalledTimes(1)
    resolveCheckin({ newly_checked_in: true, record: { ...record } })
    await flushPromises()
  })

  it('switches to checked-in state and refreshes the displayed balance source', async () => {
    getStatus
      .mockResolvedValueOnce({ ...status })
      .mockResolvedValue({ ...status, checked_in_today: true, can_check_in: false, today_record: { ...record } })
    const wrapper = mountShortcut()
    await flushPromises()

    await wrapper.get('[data-testid="checkin-shortcut"]').trigger('click')
    await wrapper.get('[data-testid="quick-checkin-normal"]').trigger('click')
    await flushPromises()

    const shortcut = wrapper.get('[data-testid="checkin-shortcut"]')
    expect(shortcut.text()).toContain('checkin.checkedToday')
    expect(shortcut.attributes('aria-expanded')).toBe('false')
    expect(refreshUser).toHaveBeenCalledTimes(1)
    expect(getStatus).toHaveBeenCalledTimes(2)
  })

  it('opens the check-in page instead of resubmitting after today is complete', async () => {
    getStatus.mockResolvedValue({ ...status, checked_in_today: true, can_check_in: false, today_record: { ...record } })
    const wrapper = mountShortcut()
    await flushPromises()

    await wrapper.get('[data-testid="checkin-shortcut"]').trigger('click')

    expect(checkIn).not.toHaveBeenCalled()
    expect(push).toHaveBeenCalledWith('/checkin')
  })

  it('uses the existing security message for a blocked source', async () => {
    checkIn.mockRejectedValue({ reason: 'CHECKIN_SOURCE_LIMITED', message: 'blocked' })
    const wrapper = mountShortcut()
    await flushPromises()

    await wrapper.get('[data-testid="checkin-shortcut"]').trigger('click')
    await wrapper.get('[data-testid="quick-checkin-normal"]').trigger('click')
    await flushPromises()

    expect(showError).toHaveBeenCalledWith('checkin.sourceLimited')
    expect(refreshUser).not.toHaveBeenCalled()
  })

  it('requires a Turnstile token before quick check-in', async () => {
    getStatus.mockResolvedValue({
      ...status,
      turnstile_enabled: true,
      turnstile_site_key: 'site-key',
    })
    const wrapper = mountShortcut()
    await flushPromises()

    await wrapper.get('[data-testid="checkin-shortcut"]').trigger('click')
    expect(wrapper.get('[data-testid="checkin-shortcut-turnstile"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="quick-checkin-menu-normal"]').attributes('disabled')).toBeDefined()

    await wrapper.get('[data-testid="turnstile-stub"]').trigger('click')
    await wrapper.get('[data-testid="quick-checkin-menu-normal"]').trigger('click')
    await flushPromises()

    expect(checkIn).toHaveBeenCalledWith('normal', status.business_date, 'turnstile-proof')
  })
})
