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

  it('sanitizes the configured URL and opens it safely', () => {
    expect(headerSource).toContain('sanitizeUrl(appStore.communityGroupUrl)')
    expect(headerSource).toContain(':href="communityGroupUrl"')
    expect(headerSource).toContain('target="_blank"')
    expect(headerSource).toContain('rel="noopener noreferrer"')
  })

  it('uses the configured contact icon path with theme-aware color', () => {
    expect(headerSource).toContain('M928 585.344c0-67.328')
    expect(headerSource).toContain('fill="currentColor"')
  })
})
