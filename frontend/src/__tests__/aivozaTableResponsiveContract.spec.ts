import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const srcRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const source = (relativePath: string) => readFileSync(resolve(srcRoot, relativePath), 'utf8')

const responsiveTableSurfaces = [
  'components/common/DataTable.vue',
  'components/account/AccountUsageCell.vue',
  'views/admin/AccountsView.vue',
  'views/admin/ops/components/OpsAlertEventsCard.vue',
  'views/admin/ops/components/OpsAlertRulesCard.vue',
  'views/admin/ops/components/OpsOpenAITokenStatsCard.vue'
]

const denseDataTableSurfaces = [
  'components/admin/usage/UsageTable.vue',
  'components/user/UserErrorRequestsTable.vue',
  'views/admin/AuditLogView.vue',
  'views/admin/ops/components/OpsErrorLogTable.vue'
]

const denseNativeTableSurfaces = [
  'features/prompt-audit/components/EventWorkspace.vue',
  'views/admin/ops/components/OpsRequestDetailsModal.vue',
  'views/admin/ops/components/OpsSystemLogTable.vue'
]

describe('Aivoza responsive table contract', () => {
  it('uses 1024px as the single table/card breakpoint', () => {
    for (const relativePath of responsiveTableSurfaces) {
      const componentSource = source(relativePath)
      expect(componentSource, relativePath).toContain('(min-width: 1024px)')
      expect(componentSource, relativePath).not.toContain('(min-width: 768px)')
    }

    expect(source('components/layout/TablePageLayout.vue')).toContain('(max-width: 1023px)')
  })

  it('keeps dense DataTable surfaces as locally scrollable tables', () => {
    for (const relativePath of denseDataTableSurfaces) {
      expect(source(relativePath), relativePath).toContain('responsive-mode="scroll"')
    }

    const dataTableSource = source('components/common/DataTable.vue')
    expect(dataTableSource).toContain("responsiveMode?: 'cards' | 'scroll'")
    expect(dataTableSource).toMatch(/\.table-wrapper\s*\{[^}]*overflow-x:\s*auto/s)
  })

  it('contains native log tables inside named keyboard-scrollable regions', () => {
    for (const relativePath of denseNativeTableSurfaces) {
      const componentSource = source(relativePath)
      expect(componentSource, relativePath).toMatch(/overflow-(?:x-auto|auto)/)
      expect(componentSource, relativePath).toContain('role="region"')
      expect(componentSource, relativePath).toContain('tabindex="0"')
      expect(componentSource, relativePath).toContain('<caption class="sr-only">')
      expect(componentSource, relativePath).toMatch(/min-w-\[\d+px\]/)
    }
  })

  it('keeps compact table actions at least 36px high', () => {
    expect(source('style.css')).toMatch(/\.btn-sm\s*\{[^}]*min-h-9/s)

    const compactActionSurfaces = [
      'components/admin/usage/UsageTable.vue',
      'features/prompt-audit/components/EventWorkspace.vue',
      'views/admin/AuditLogView.vue',
      'views/admin/ops/components/OpsErrorLogTable.vue'
    ]
    for (const relativePath of compactActionSurfaces) {
      const componentSource = source(relativePath)
      expect(componentSource, relativePath).toMatch(/(?:min-h-9|h-9 w-9|btn-sm)/)
    }
  })
})
