import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../VersionBadge.vue')
const componentSource = readFileSync(componentPath, 'utf8')

describe('VersionBadge compact trigger', () => {
  it('keeps a 44px target around a visually compact 24px badge', () => {
    expect(componentSource).toContain('min-h-11 min-w-11')
    expect(componentSource).toContain('data-version-chip')
    expect(componentSource).toContain('class="flex h-6 items-center')
  })
})
