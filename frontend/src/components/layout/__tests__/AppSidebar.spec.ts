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

describe('AppSidebar collapsible groups', () => {
  it('lets the user collapse a group even while a child route is active', () => {
    // The expand state must come from the user's override first, falling back
    // to the active-route heuristic only when the user has not clicked yet.
    expect(componentSource).toContain('const groupExpandOverrides = ref<Map<string, boolean>>(new Map())')
    expect(componentSource).not.toContain('expandedGroups.value.has(item.path) || isGroupActive(item)')
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
