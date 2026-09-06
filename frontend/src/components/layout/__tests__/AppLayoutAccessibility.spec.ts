import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../AppLayout.vue')
const componentSource = readFileSync(componentPath, 'utf8')

describe('AppLayout mobile drawer accessibility', () => {
  it('keeps the obscured workspace out of the accessibility and focus trees', () => {
    expect(componentSource).toContain(':inert="mobileSidebarModalOpen ? true : undefined"')
    expect(componentSource).toContain(':aria-hidden="mobileSidebarModalOpen ? \'true\' : undefined"')
    expect(componentSource).toContain("window.matchMedia('(min-width: 1024px)')")
  })

  it('provides a skip link and moves focus to main after route changes', () => {
    expect(componentSource).toContain('href="#app-main-content"')
    expect(componentSource).toContain('id="app-main-content"')
    expect(componentSource).toContain('tabindex="-1"')
    expect(componentSource).toContain('() => route.fullPath')
    expect(componentSource).toContain('mainRef.value?.focus({ preventScroll: true })')
  })

  it('constrains the desktop workspace after applying the sidebar offset', () => {
    expect(componentSource).toContain('overflow-x-clip')
    expect(componentSource).toContain('lg:w-[calc(100%-72px)]')
    expect(componentSource).toContain('lg:w-[calc(100%-16rem)]')
    expect(componentSource).toContain('app-main min-w-0 max-w-full')
  })
})
