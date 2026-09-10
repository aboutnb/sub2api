import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const sidebarSource = readFileSync(resolve(dir, '../AppSidebar.vue'), 'utf8')
const homeViewSource = readFileSync(resolve(dir, '../../../views/HomeView.vue'), 'utf8')
const keyUsageViewSource = readFileSync(resolve(dir, '../../../views/KeyUsageView.vue'), 'utf8')
const brandingSource = readFileSync(resolve(dir, '../../../utils/branding.ts'), 'utf8')

describe('site_logo sanitization', () => {
  it('AppSidebar resolves siteLogo through the shared brand helper', () => {
    expect(sidebarSource).toContain("import { resolveBrandLogo } from '@/utils/branding'")
    expect(sidebarSource).toContain('resolveBrandLogo(appStore.siteLogo')
  })

  it('HomeView resolves siteLogo through the shared brand helper', () => {
    expect(homeViewSource).toContain('resolveBrandLogo(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo')
  })

  it('KeyUsageView resolves siteLogo through the shared brand helper', () => {
    expect(keyUsageViewSource).toContain('resolveBrandLogo(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo')
  })

  it('the shared brand helper sanitizes relative and data URLs', () => {
    expect(brandingSource).toContain("sanitizeUrl(logoUrl ?? ''")
    expect(brandingSource).toContain('allowRelative: true')
    expect(brandingSource).toContain('allowDataUrl: true')
  })
})
