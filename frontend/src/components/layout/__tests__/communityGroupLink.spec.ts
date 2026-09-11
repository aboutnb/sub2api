import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const headerSource = readFileSync(resolve(dir, '../AppHeader.vue'), 'utf8')

describe('community group header link', () => {
  it('renders immediately after the announcement bell', () => {
    const announcementIndex = headerSource.indexOf('<AnnouncementBell v-if="user" />')
    const communityIndex = headerSource.indexOf('<!-- Community Group Link -->')
    const docsIndex = headerSource.indexOf('<!-- Docs Link -->')

    expect(announcementIndex).toBeGreaterThan(-1)
    expect(communityIndex).toBeGreaterThan(announcementIndex)
    expect(communityIndex).toBeLessThan(docsIndex)
  })

  it('keeps the group link in the frequent action row and out of More', () => {
    const primaryIndex = headerSource.indexOf('data-testid="header-primary-actions"')
    const communityIndex = headerSource.indexOf('<!-- Community Group Link -->')
    const moreMenuIndex = headerSource.indexOf('id="app-more-menu"')

    expect(communityIndex).toBeGreaterThan(primaryIndex)
    expect(communityIndex).toBeLessThan(moreMenuIndex)
    expect(headerSource.match(/<!-- Community Group Link -->/g)).toHaveLength(1)
    expect(headerSource).toContain("const hasSecondaryActions = computed(() => Boolean(\n  docUrl.value")
  })

  it('keeps desktop action blocks compact while preserving coarse-pointer sizing', () => {
    expect(headerSource).toContain('@media (pointer: fine)')
    expect(headerSource).toContain('height: 2.25rem')
    expect(headerSource).toContain('min-height: 2.25rem')
    expect(headerSource).toContain(':deep(button)')
    expect(headerSource).toContain(':deep(a[href])')
  })

  it('sanitizes the configured URL and opens it safely', () => {
    expect(headerSource).toContain('sanitizeUrl(appStore.communityGroupUrl)')
    expect(headerSource).toContain(':href="communityGroupUrl"')
    expect(headerSource).toContain('target="_blank"')
    expect(headerSource).toContain('rel="noopener noreferrer"')
  })

  it('uses the configured name and sanitizes the optional icon', () => {
    expect(headerSource).toContain("appStore.communityGroupName.trim() || t('nav.communityGroup')")
    expect(headerSource).toContain('sanitizeUrl(appStore.communityGroupIcon')
    expect(headerSource).toContain('allowDataUrl: true')
    expect(headerSource).toContain('v-if="communityGroupIcon"')
  })

  it('keeps the supplied contact icon as the fallback', () => {
    expect(headerSource).toContain('M928 585.344c0-67.328')
    expect(headerSource).toContain('fill="currentColor"')
  })

  it('uses a quiet header treatment while keeping mobile icon-only layout', () => {
    expect(headerSource).not.toContain('border-cyan-200 bg-cyan-50')
    expect(headerSource).toContain('text-ink transition-all')
    expect(headerSource).toContain('hover:bg-surface-muted')
    expect(headerSource).toContain('class="hidden max-w-28 truncate leading-tight sm:inline"')
  })
})
