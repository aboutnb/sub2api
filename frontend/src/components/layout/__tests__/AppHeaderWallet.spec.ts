import { defineComponent, nextTick } from 'vue'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import AppHeader from '../AppHeader.vue'

const state = vi.hoisted(() => ({
  flags: {
    modelPlaza: false,
    userSubscriptions: true,
    checkin: true,
    payment: true,
  },
  route: {
    name: 'Dashboard',
    params: {},
    meta: { titleKey: 'dashboard.title' },
  },
  router: {
    push: vi.fn(),
  },
  appStore: {
    mobileOpen: false,
    contactInfo: '',
    communityGroupName: '',
    communityGroupIcon: '',
    communityGroupUrl: '',
    docUrl: '',
    cachedPublicSettings: { custom_menu_items: [] },
    toggleMobileSidebar: vi.fn(),
  },
  authStore: {
    user: {
      username: 'wallet-user',
      email: 'wallet@example.test',
      role: 'user',
      balance: 9.06,
      frozen_balance: 1.25,
      avatar_url: '',
    } as Record<string, unknown> | null,
    isSimpleMode: false,
    isAdmin: false,
    logout: vi.fn(),
  },
  onboardingStore: {
    replay: vi.fn(),
  },
  adminSettingsStore: {
    customMenuItems: [],
  },
  paymentStore: {
    configLoaded: true,
    configLoading: false,
    config: {
      enabled: true,
      balance_disabled: false,
    } as Record<string, unknown> | null,
    fetchConfig: vi.fn(),
  },
}))

vi.mock('vue-router', () => ({
  useRoute: () => state.route,
  useRouter: () => state.router,
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => ({
        'common.actions': '常用操作',
        'common.availableBalance': '可用余额',
        'common.frozenBalance': '冻结金额',
        'common.totalBalance': '总余额',
        'payment.tabTopUp': '充值',
      })[key] ?? key,
    }),
  }
})

vi.mock('@/stores', () => ({
  useAppStore: () => state.appStore,
  useAuthStore: () => state.authStore,
  useOnboardingStore: () => state.onboardingStore,
  usePaymentStore: () => state.paymentStore,
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => state.adminSettingsStore,
}))

vi.mock('@/utils/featureFlags', () => ({
  FeatureFlags: {
    modelPlaza: 'modelPlaza',
    userSubscriptions: 'userSubscriptions',
    checkin: 'checkin',
    payment: 'payment',
  },
  isFeatureFlagEnabled: (flag: keyof typeof state.flags) => state.flags[flag],
}))

const RouterLinkStub = defineComponent({
  name: 'RouterLink',
  inheritAttrs: true,
  props: {
    to: {
      type: [String, Object],
      required: true,
    },
  },
  computed: {
    href(): string {
      return typeof this.to === 'string'
        ? this.to
        : String((this.to as { path?: string }).path ?? '')
    },
  },
  template: '<a :href="href"><slot /></a>',
})

const childStubs = {
  AnnouncementBell: defineComponent({ template: '<span data-testid="announcement-stub" />' }),
  SubscriptionProgressMini: defineComponent({ template: '<span data-testid="subscription-stub" />' }),
  CheckinShortcut: defineComponent({ template: '<span data-testid="checkin-stub" />' }),
  LocaleSwitcher: defineComponent({ template: '<span data-testid="locale-stub" />' }),
  Icon: defineComponent({ template: '<span aria-hidden="true" />' }),
  RouterLink: RouterLinkStub,
}

let wrapper: VueWrapper | null = null

function mountHeader() {
  wrapper = mount(AppHeader, {
    attachTo: document.body,
    global: { stubs: childStubs },
  })
  return wrapper
}

beforeEach(() => {
  state.flags.modelPlaza = false
  state.flags.userSubscriptions = true
  state.flags.checkin = true
  state.flags.payment = true
  state.authStore.user = {
    username: 'wallet-user',
    email: 'wallet@example.test',
    role: 'user',
    balance: 9.06,
    frozen_balance: 1.25,
    avatar_url: '',
  }
  state.authStore.isSimpleMode = false
  state.authStore.isAdmin = false
  state.paymentStore.configLoaded = true
  state.paymentStore.configLoading = false
  state.paymentStore.config = { enabled: true, balance_disabled: false }
  state.paymentStore.fetchConfig.mockReset()
  state.paymentStore.fetchConfig.mockResolvedValue(state.paymentStore.config)
  state.router.push.mockReset()
})

afterEach(() => {
  wrapper?.unmount()
  wrapper = null
  document.body.innerHTML = ''
  vi.restoreAllMocks()
})

describe('AppHeader wallet and recharge control', () => {
  it('keeps recharge hidden while payment configuration is loading or failed', async () => {
    let finishRequest: (value: null) => void = () => undefined
    state.paymentStore.configLoaded = false
    state.paymentStore.config = null
    state.paymentStore.fetchConfig.mockReturnValue(new Promise<null>((resolve) => {
      finishRequest = resolve
    }))

    const mounted = mountHeader()

    expect(state.paymentStore.fetchConfig).toHaveBeenCalledOnce()
    expect(mounted.find('[data-testid="header-recharge-shortcut"]').exists()).toBe(false)

    finishRequest(null)
    await flushPromises()

    expect(mounted.find('[data-testid="header-recharge-shortcut"]').exists()).toBe(false)
  })

  it.each([
    ['the global payment flag is disabled', { paymentFlag: false }],
    ['payment configuration is not loaded', { configLoaded: false }],
    ['the payment system is disabled', { enabled: false }],
    ['balance recharge is disabled', { balanceDisabled: true }],
    ['the account uses simple mode', { simpleMode: true }],
  ])('hides recharge when %s', (_description, override) => {
    state.flags.payment = override.paymentFlag ?? true
    state.paymentStore.configLoaded = override.configLoaded ?? true
    state.paymentStore.config = {
      enabled: override.enabled ?? true,
      balance_disabled: override.balanceDisabled ?? false,
    }
    state.authStore.isSimpleMode = override.simpleMode ?? false

    const mounted = mountHeader()

    expect(mounted.find('[data-testid="header-recharge-shortcut"]').exists()).toBe(false)
  })

  it('links the enabled recharge action to the real purchase route with a 44px accessible target', () => {
    const mounted = mountHeader()
    const wallet = mounted.get('[data-testid="header-wallet"]')
    const recharge = mounted.get('[data-testid="header-recharge-shortcut"]')

    expect(recharge.element.tagName).toBe('A')
    expect(recharge.attributes('href')).toBe('/purchase')
    expect(recharge.attributes('aria-label')).toBe('充值')
    expect(recharge.attributes('title')).toBe('充值')
    expect(recharge.classes()).toEqual(expect.arrayContaining(['min-h-11', 'min-w-11']))
    expect(wallet.attributes('aria-label')).toBe('可用余额: $9.06; 充值')
  })

  it('toggles the labelled balance details and reports disclosure state', async () => {
    const mounted = mountHeader()
    const trigger = mounted.get('[data-testid="header-balance"]')

    expect(trigger.attributes('aria-expanded')).toBe('false')
    expect(trigger.attributes('aria-controls')).toBe('header-balance-details')
    expect(mounted.find('#header-balance-details').exists()).toBe(false)

    await trigger.trigger('click')

    expect(trigger.attributes('aria-expanded')).toBe('true')
    const details = mounted.get('#header-balance-details')
    expect(details.attributes('role')).toBe('region')
    expect(details.attributes('aria-labelledby')).toBe('header-balance-trigger')
    expect(details.text()).toContain('$9.06')
    expect(details.text()).toContain('$1.25')
    expect(details.text()).toContain('$10.31')

    await trigger.trigger('click')
    expect(trigger.attributes('aria-expanded')).toBe('false')
    expect(mounted.find('#header-balance-details').exists()).toBe(false)
  })

  it('closes balance details with Escape and restores focus to the trigger', async () => {
    vi.spyOn(window, 'requestAnimationFrame').mockImplementation((callback) => {
      callback(0)
      return 1
    })
    const mounted = mountHeader()
    const trigger = mounted.get<HTMLButtonElement>('[data-testid="header-balance"]')

    await trigger.trigger('click')
    const details = mounted.get('#header-balance-details')
    await details.trigger('keydown', { key: 'Escape' })
    await nextTick()

    expect(mounted.find('#header-balance-details').exists()).toBe(false)
    expect(trigger.attributes('aria-expanded')).toBe('false')
    expect(document.activeElement).toBe(trigger.element)
  })
})
