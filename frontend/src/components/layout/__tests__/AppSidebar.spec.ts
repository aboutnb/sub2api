import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppSidebar.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')

describe('AppSidebar custom SVG styles', () => {
  it('does not override uploaded SVG fill or stroke colors', () => {
    expect(componentSource).toContain('.sidebar-svg-icon {')
    expect(componentSource).toContain('color: currentColor;')
    expect(componentSource).toContain('display: block;')
    expect(componentSource).not.toContain('stroke: currentColor;')
    expect(componentSource).not.toContain('fill: none;')
  })
})

describe('AppSidebar scroll position persistence', () => {
  it('binds a template ref to the sidebar nav element', () => {
    expect(componentSource).toContain('ref="sidebarNavRef"')
    expect(componentSource).toContain('sidebar-nav')
  })

  it('declares sidebarNavRef in script setup', () => {
    expect(componentSource).toContain("const sidebarNavRef = ref<HTMLElement | null>(null)")
  })

  it('saves scroll position on beforeUnmount', () => {
    expect(componentSource).toContain('onBeforeUnmount')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('sidebarNavRef.value.scrollTop')
  })

  it('restores scroll position on mount', () => {
    expect(componentSource).toContain('onMounted')
    expect(componentSource).toContain('appStore.sidebarScrollTop')
    expect(componentSource).toContain('nextTick')
  })
})

describe('AppSidebar header styles', () => {
  it('does not clip the version badge dropdown', () => {
    const sidebarHeaderBlockMatch = styleSource.match(/\.sidebar-header\s*\{[\s\S]*?\n {2}\}/)
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(sidebarHeaderBlockMatch).not.toBeNull()
    expect(sidebarBrandBlockMatch).not.toBeNull()
    expect(sidebarHeaderBlockMatch?.[0]).not.toContain('@apply overflow-hidden;')
    expect(sidebarBrandBlockMatch?.[0]).not.toContain('overflow: hidden;')
  })

  it('keeps the 44px version control beside the brand instead of above the viewport', () => {
    const sidebarBrandBlockMatch = componentSource.match(/\.sidebar-brand\s*\{[\s\S]*?\n\}/)

    expect(componentSource).toContain('class="sidebar-version"')
    expect(sidebarBrandBlockMatch?.[0]).toContain('display: flex;')
    expect(sidebarBrandBlockMatch?.[0]).toContain('align-items: center;')
    expect(componentSource).toContain('.sidebar-version {')
  })

  it('keeps the brand mark at 44px in both expanded and 72px collapsed shells', () => {
    expect(componentSource).toContain('sidebar-logo flex h-11 w-11')
    expect(componentSource).toContain('flex: 0 0 2.75rem;')
    expect(componentSource).toContain('min-width: 2.75rem;')
    expect(componentSource).toContain('padding-left: 0.875rem;')
    expect(componentSource).toContain('padding-right: 0.875rem;')
  })
})

describe('AppSidebar user subscriptions feature flag', () => {
  it('binds My Subscriptions to the public feature switch', () => {
    expect(componentSource).toContain(
      "{ path: '/subscriptions', label: t('nav.mySubscriptions'), icon: CreditCardIcon, hideInSimpleMode: true, featureFlag: flagUserSubscriptions }",
    )
    expect(componentSource).toContain(
      'const flagUserSubscriptions = makeSidebarFlag(FeatureFlags.userSubscriptions)',
    )
  })
})

describe('AppSidebar check-in access', () => {
  it('keeps check-in in the shared personal menu for users and administrators', () => {
    expect(componentSource).toContain(
      "{ path: '/checkin', label: t('nav.checkin'), icon: CheckinCenterIcon, hideInSimpleMode: true, featureFlag: flagCheckin }",
    )
    expect(componentSource).not.toContain('featureFlag: () => withDashboard && flagCheckin()')
  })
})

describe('AppSidebar mobile accessibility', () => {
  it('keeps the off-canvas navigation out of the focus tree while closed', () => {
    expect(componentSource).toContain('id="app-sidebar"')
    expect(componentSource).toContain(':inert="!isDesktopViewport && !mobileOpen ? true : undefined"')
    expect(componentSource).toContain(':aria-hidden="!isDesktopViewport && !mobileOpen ? \'true\' : undefined"')
  })

  it('supports Escape, focus restoration, and a semantic backdrop control', () => {
    expect(componentSource).toContain("event.key === 'Escape'")
    expect(componentSource).toContain('closeMobile(true)')
    expect(componentSource).toContain('[aria-controls="app-sidebar"]')
    expect(componentSource).toContain('type="button"')
  })

  it('keeps keyboard focus inside the open mobile drawer', () => {
    expect(componentSource).toContain(':aria-modal="!isDesktopViewport && mobileOpen ? \'true\' : undefined"')
    expect(componentSource).toContain("if (event.key !== 'Tab') return")
    expect(componentSource).toContain('getMobileDialogFocusables()')
    expect(componentSource).toContain('lastFocusable.focus()')
    expect(componentSource).toContain('firstFocusable.focus()')
  })

  it('exposes collapsible navigation group state', () => {
    expect(componentSource).toContain(':aria-expanded="!sidebarCollapsed && isGroupExpanded(item)"')
    expect(componentSource).toContain(':aria-controls="groupPanelId(item.path)"')
    expect(componentSource).toContain(':id="groupPanelId(item.path)"')
  })
})

describe('AppSidebar security audit group', () => {
  const groupStart = componentSource.indexOf("path: '/admin/security-audit'")
  const groupEnd = componentSource.indexOf("{ path: '/admin/redeem'", groupStart)
  const groupSource = componentSource.slice(groupStart, groupEnd)

  it('groups login protection and audit logs under security audit', () => {
    expect(groupStart).toBeGreaterThan(-1)
    expect(groupEnd).toBeGreaterThan(groupStart)
    expect(groupSource).toContain("path: '/admin/auth-ip-bans'")
    expect(groupSource).toContain("path: '/admin/audit-logs'")
    expect(componentSource.match(/path: '\/admin\/auth-ip-bans'/g)).toHaveLength(1)
    expect(componentSource.match(/path: '\/admin\/audit-logs'/g)).toHaveLength(1)
  })

  it('keeps only login protection visible from the group in simple mode', () => {
    expect(groupSource).toContain(
      "{ path: '/admin/audit-logs', label: t('nav.auditLogs'), icon: ShieldIcon, hideInSimpleMode: true }",
    )
    expect(componentSource).toContain(
      'children: item.children.filter(child => !child.hideInSimpleMode)',
    )
  })
})
