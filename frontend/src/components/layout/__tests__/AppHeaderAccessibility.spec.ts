import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppHeader.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const subscriptionSource = readFileSync(
  resolve(dirname(fileURLToPath(import.meta.url)), '../../common/SubscriptionProgressMini.vue'),
  'utf8',
)

describe('AppHeader accessibility', () => {
  it('connects the mobile menu trigger to the off-canvas sidebar', () => {
    expect(componentSource).toContain('aria-controls="app-sidebar"')
    expect(componentSource).toContain(':aria-expanded="mobileOpen"')
    expect(componentSource).toContain('lg:hidden')
    expect(componentSource).toContain('{{ pageTitle }}')
  })

  it('exposes and restores focus for the user disclosure', () => {
    expect(componentSource).toContain('id="app-user-menu-trigger"')
    expect(componentSource).toContain(':aria-expanded="dropdownOpen"')
    expect(componentSource).toContain('aria-controls="app-user-menu"')
    expect(componentSource).not.toMatch(/app-user-menu-trigger[\s\S]{0,500}aria-haspopup/)
    expect(componentSource).toContain("event.key === 'Escape'")
    expect(componentSource).toContain('userMenuButtonRef.value?.focus()')
  })

  it('makes the balance details available beyond pointer hover', () => {
    expect(componentSource).toContain('data-testid="header-wallet"')
    expect(componentSource).toContain('id="header-balance-trigger"')
    expect(componentSource).toContain('type="button"')
    expect(componentSource).toContain('data-testid="header-balance"')
    expect(componentSource).toContain('class="balance-focus header-balance relative')
    expect(componentSource).toContain('role="group"')
    expect(componentSource).toContain(':aria-expanded="balanceDetailsOpen"')
    expect(componentSource).toContain('aria-controls="header-balance-details"')
    expect(componentSource).toContain('aria-labelledby="header-balance-trigger"')
    expect(componentSource).toContain('closeBalanceDetails(true)')
    expect(componentSource).toContain('balanceButtonRef.value?.focus()')
    expect(componentSource).toContain('@focusout="handleWalletFocusout"')
    expect(componentSource).toContain('.balance-focus:focus-visible')
  })

  it('integrates a compact recharge action into the balance control', () => {
    expect(componentSource).toContain('data-testid="header-recharge-shortcut"')
    expect(componentSource).toContain("v-if=\"showRechargeShortcut\"")
    expect(componentSource).toContain('to="/purchase"')
    expect(componentSource).not.toContain('to="/payment"')
    expect(componentSource).toContain('paymentStore.configLoaded')
    expect(componentSource).toContain("paymentStore.config?.enabled === true")
    expect(componentSource).toContain("paymentStore.config.balance_disabled === false")
    expect(componentSource).toContain('!authStore.isSimpleMode')
    expect(componentSource).toContain('class="hidden whitespace-nowrap sm:inline"')
    expect(componentSource).toMatch(/data-testid="header-recharge-shortcut"[\s\S]{0,500}min-h-11 min-w-11/)
    expect(componentSource).toContain('.header-recharge-link:focus-visible')
  })

  it('keeps the five frequent actions directly visible and renders each only once', () => {
    expect(componentSource.match(/<AnnouncementBell/g)).toHaveLength(1)
    expect(componentSource.match(/<SubscriptionProgressMini/g)).toHaveLength(1)
    expect(componentSource.match(/<CheckinShortcut/g)).toHaveLength(1)
    expect(componentSource.match(/<LocaleSwitcher/g)).toHaveLength(1)
    expect(componentSource.match(/data-testid="header-balance"/g)).toHaveLength(1)
    expect(componentSource).toContain('data-testid="header-primary-actions"')
    expect(componentSource).toContain('role="group"')
    expect(componentSource).toContain("grid-template-areas:\n      'leading user more'\n      'primary primary primary'")
    expect(componentSource).toContain('@media (max-width: 1279.98px)')
    expect(componentSource).toContain('flex-wrap: wrap')
  })

  it('collapses only secondary links into a keyboard-operable More disclosure below 1536px', () => {
    expect(componentSource).toContain("const desktopHeaderQuery = '(min-width: 1536px)'")
    expect(componentSource).toContain('<template v-if="isDesktopHeader">')
    expect(componentSource).toContain('id="app-more-menu-trigger"')
    expect(componentSource).toContain('v-if="!isDesktopHeader && hasSecondaryActions"')
    expect(componentSource).toContain(':aria-expanded="moreOpen"')
    expect(componentSource).toContain('aria-controls="app-more-menu"')
    expect(componentSource).not.toMatch(/app-more-menu-trigger[\s\S]{0,500}aria-haspopup/)
    expect(componentSource).toContain("if (event.key === 'ArrowDown')")
    expect(componentSource).toContain("if (event.key === 'Home')")
    expect(componentSource).toContain('moreMenuButtonRef.value?.focus()')
    expect(componentSource).toContain('@focusout="handleMoreFocusout"')
    expect(componentSource).not.toContain("if (event.key === 'Tab')")
    expect(componentSource).not.toContain('openMobileAnnouncements')
    expect(componentSource).not.toContain('mobileAnnouncementRef')
  })

  it('truncates long route titles without crowding header actions', () => {
    expect(componentSource).toContain('flex min-w-0 flex-1 items-center')
    expect(componentSource).toContain('hidden min-w-0 max-w-[28rem] lg:block')
    expect(componentSource).toContain('truncate text-lg font-semibold')
    expect(componentSource).toContain('hidden min-w-0 max-w-28 text-left md:block')
  })

  it('does not expose subscription navigation in simple mode', () => {
    expect(componentSource).toContain('v-if="user && !authStore.isSimpleMode && userSubscriptionsEnabled"')
    expect(componentSource).toMatch(
      /<SubscriptionProgressMini\s+v-if="user && !authStore\.isSimpleMode && userSubscriptionsEnabled"/,
    )
  })

  it('keeps the subscription destination visible before a plan is active', () => {
    expect(subscriptionSource).toContain('v-if="hasActiveSubscriptions"')
    expect(subscriptionSource).toContain('data-testid="subscription-shortcut-empty"')
    expect(subscriptionSource).toContain('v-else\n      to="/subscriptions"')
  })

  it('keeps active subscription details keyboard dismissible', () => {
    expect(subscriptionSource).toContain(':aria-expanded="tooltipOpen"')
    expect(subscriptionSource).toContain('aria-controls="subscription-progress-popover"')
    expect(subscriptionSource).toContain("event.key !== 'Escape'")
    expect(subscriptionSource).toContain('triggerRef.value?.focus()')
    expect(subscriptionSource).toContain('@focusout="handleFocusout"')
  })
})
