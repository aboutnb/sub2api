import { describe, expect, it } from 'vitest'
import { readFileSync, readdirSync } from 'node:fs'
import { dirname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const testDirectory = dirname(fileURLToPath(import.meta.url))
const frontendRoot = resolve(testDirectory, '../..')
const repoRoot = resolve(frontendRoot, '..')
const viewsRoot = resolve(frontendRoot, 'src/views')
const componentsRoot = resolve(frontendRoot, 'src/components')
const featuresRoot = resolve(frontendRoot, 'src/features')
const ledger = readFileSync(
  resolve(repoRoot, 'design-system/aivoza-flowai/COVERAGE.md'),
  'utf8',
)

function vueFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return vueFiles(path)
    return entry.name.endsWith('.vue') ? [path] : []
  })
}

const views = vueFiles(viewsRoot)
const featureSurfaces = vueFiles(featuresRoot)
const productionSurfaces = [
  ...views,
  ...vueFiles(componentsRoot),
  ...featureSurfaces,
]
const sharedOverlays = [...vueFiles(componentsRoot), ...featureSurfaces].filter((path) =>
  /<Teleport|BaseDialog|ConfirmDialog|role=["']dialog["']|aria-modal/.test(
    readFileSync(path, 'utf8'),
  ),
)

describe('Aivoza visual coverage', () => {
  it('keeps all current view-owned surfaces in the release ledger', () => {
    expect(views).toHaveLength(90)

    const undocumented = views
      .map((path) => `frontend/src/views/${relative(viewsRoot, path)}`)
      .filter((path) => !ledger.includes(`\`${path}\``))

    expect(undocumented).toEqual([])
  })

  it('keeps every route and hidden feature surface in the release ledger', () => {
    const featureGaps = featureSurfaces
      .map((path) => `frontend/src/features/${relative(featuresRoot, path)}`)
      .filter((path) => !ledger.includes(`\`${path}\``))

    const router = readFileSync(resolve(frontendRoot, 'src/router/index.ts'), 'utf8')
    const routeSurfaces = Array.from(router.matchAll(/import\(['"]@\/([^'"]+\.vue)['"]\)/g), (
      match,
    ) => `frontend/src/${match[1]}`)
    const routeGaps = routeSurfaces.filter((path) => !ledger.includes(`\`${path}\``))

    expect(featureGaps).toEqual([])
    expect(routeGaps).toEqual([])
  })

  it('archives every shared overlay and Portal surface', () => {
    const undocumented = sharedOverlays
      .map((path) => `frontend/src/${relative(resolve(frontendRoot, 'src'), path)}`)
      .filter((path) => !ledger.includes(`\`${path}\``))

    expect(sharedOverlays).toHaveLength(70)
    expect(undocumented).toEqual([])
  })

  it('connects every view to a dark-aware surface', () => {
    const themeAwareDelegators = new Set([
      'admin/affiliates/AdminAffiliateInvitesView.vue',
      'admin/affiliates/AdminAffiliateRebatesView.vue',
      'admin/affiliates/AdminAffiliateTransfersView.vue',
      'user/ChannelStatusV1View.vue',
      'user/ChannelStatusView.vue',
      'user/DashboardView.vue',
    ])

    const disconnected = views.flatMap((path) => {
      const source = readFileSync(path, 'utf8')
      const name = relative(viewsRoot, path)
      const isHomeTemplateHost = name === 'HomeView.vue'
      const usesPublicShell = source.includes('public-shell')
      return source.includes('dark:') || usesPublicShell || themeAwareDelegators.has(name) || isHomeTemplateHost
        ? []
        : [name]
    })

    expect(disconnected).toEqual([])
  })

  it('keeps every production h1 at or below 24px', () => {
    const violations = productionSurfaces.flatMap((path) => {
      const source = readFileSync(path, 'utf8')
      const headings = source.match(/<h1\b[\s\S]*?>/g) || []
      return headings
        .filter((heading) => /\btext-(?:3xl|4xl|5xl|6xl|7xl|8xl|9xl)\b|text-\[(?:2[5-9]|[3-9]\d)px\]/.test(heading))
        .map(() => relative(resolve(frontendRoot, 'src'), path))
    })

    expect(violations).toEqual([])
  })

  it('gives every Toggle callsite an explicit accessible name', () => {
    const unnamed = productionSurfaces.flatMap((path) => {
      const source = readFileSync(path, 'utf8')
      const tags = source.match(/<Toggle\b(?:"[^"]*"|'[^']*'|[^'">])*>/g) || []

      return tags
        .filter((tag) => !/\baria-(?:label|labelledby)\s*=/.test(tag))
        .map((tag) => `${relative(resolve(frontendRoot, 'src'), path)}: ${tag.replace(/\s+/g, ' ')}`)
    })

    expect(unnamed).toEqual([])
  })
})
