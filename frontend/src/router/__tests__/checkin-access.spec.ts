import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const dir = dirname(fileURLToPath(import.meta.url))
const routerSource = readFileSync(resolve(dir, '../index.ts'), 'utf8')

describe('check-in route access', () => {
  it('does not redirect administrators away from the personal check-in page', () => {
    expect(routerSource).toContain("path: '/checkin'")
    expect(routerSource).not.toMatch(/to\.path === ['"]\/checkin['"]\s*&&\s*authStore\.isAdmin/)
  })
})
