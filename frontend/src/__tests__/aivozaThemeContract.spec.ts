import { describe, expect, it } from 'vitest'
import { readFileSync, readdirSync } from 'node:fs'
import { dirname, extname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentDirectory = dirname(fileURLToPath(import.meta.url))
const frontendRoot = resolve(currentDirectory, '../..')
const sourceRoot = resolve(frontendRoot, 'src')

function sourceFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) return sourceFiles(path)
    return ['.vue', '.css', '.ts'].includes(extname(entry.name)) ? [path] : []
  })
}

const files = sourceFiles(sourceRoot).filter(
  (path) => !path.includes('/__tests__/') && !path.endsWith('.spec.ts'),
)

function rgbToken(block: string, token: string): [number, number, number] {
  const match = block.match(new RegExp(`--${token}:\\s*(\\d+)\\s+(\\d+)\\s+(\\d+);`))
  if (!match) throw new Error(`Missing RGB token: ${token}`)
  return [Number(match[1]), Number(match[2]), Number(match[3])]
}

function contrastRatio(first: [number, number, number], second: [number, number, number]) {
  const luminance = ([red, green, blue]: [number, number, number]) => {
    const [r, g, b] = [red, green, blue]
      .map((channel) => channel / 255)
      .map((channel) => channel <= 0.04045
        ? channel / 12.92
        : ((channel + 0.055) / 1.055) ** 2.4)
    return 0.2126 * r + 0.7152 * g + 0.0722 * b
  }
  const values = [luminance(first), luminance(second)].sort((a, b) => b - a)
  return (values[0] + 0.05) / (values[1] + 0.05)
}

describe('Aivoza theme contract', () => {
  it('exposes semantic Tailwind roles without overriding stock palettes or shadows', () => {
    const config = readFileSync(resolve(frontendRoot, 'tailwind.config.js'), 'utf8')
    const style = readFileSync(resolve(frontendRoot, 'src/style.css'), 'utf8')

    for (const role of ['canvas', 'surface', 'ink', 'line', 'brand', 'action']) {
      expect(config).toMatch(new RegExp(`\\b${role}:`))
    }
    expect(config).not.toMatch(/\b(?:gray|blue|teal):\s*aivoza/)
    expect(config).not.toMatch(/boxShadow:\s*\{[\s\S]*?\b(?:sm|DEFAULT|md|lg|xl|'2xl'):\s*['"]/)
    expect(style).toContain('Layer 1: fixed primitives')
    expect(style).toContain('Layer 2: mode-aware semantic color roles')
    expect(style).toContain('Layer 3: component contracts')
  })

  it('keeps control boundaries and focus indicators at 3:1 in both themes', () => {
    const style = readFileSync(resolve(frontendRoot, 'src/style.css'), 'utf8')
    const light = style.match(/:root\s*\{([\s\S]*?)\n\s*\}/)?.[1] ?? ''
    const dark = style.match(/:root\.dark\s*\{([\s\S]*?)\n\s*\}/)?.[1] ?? ''

    for (const block of [light, dark]) {
      const surface = rgbToken(block, 'av-rgb-bg-surface')
      const elevatedSurface = rgbToken(block, 'av-rgb-bg-surface-elevated')
      expect(contrastRatio(rgbToken(block, 'av-rgb-border-control'), surface)).toBeGreaterThanOrEqual(3)
      expect(contrastRatio(rgbToken(block, 'av-rgb-border-control'), elevatedSurface)).toBeGreaterThanOrEqual(3)
      expect(contrastRatio(rgbToken(block, 'av-rgb-focus'), surface)).toBeGreaterThanOrEqual(3)
      expect(contrastRatio(rgbToken(block, 'av-rgb-focus'), elevatedSurface)).toBeGreaterThanOrEqual(3)
    }

    expect(style).toContain("[tabindex='0']):focus-visible")
    expect(style).toContain('outline: 3px solid var(--av-focus) !important')
  })

  it('does not use poster-sized utility classes on headings', () => {
    const violations = files.flatMap((path) => {
      const source = readFileSync(path, 'utf8')
      const headings = source.match(/<h[1-6]\b[\s\S]*?>/g) || []
      return headings
        .filter((heading) => /\btext-(?:4xl|5xl|6xl|7xl|8xl|9xl)\b/.test(heading))
        .map((heading) => `${relative(frontendRoot, path)}: ${heading.replace(/\s+/g, ' ')}`)
    })

    expect(violations).toEqual([])

    const legalDocument = readFileSync(
      resolve(frontendRoot, 'src/views/public/LegalDocumentView.vue'),
      'utf8',
    )
    const renderedH1Rule = legalDocument.match(/\.legal-document-content\s+:deep\(h1\)\s*\{([\s\S]*?)\}/)?.[1] ?? ''
    expect(renderedH1Rule).toContain('text-2xl')
    expect(renderedH1Rule).not.toMatch(/text-(?:3xl|4xl|5xl|6xl|7xl|8xl|9xl)/)
  })

  it('keeps decorative gradients out of product surfaces', () => {
    const approvedDataVisualizations = new Set([
      'src/components/admin/usage/UsageTable.vue'
    ])
    const violations = files
      .filter((path) => readFileSync(path, 'utf8').includes('bg-gradient'))
      .map((path) => relative(frontendRoot, path))
      .filter((path) => !approvedDataVisualizations.has(path))

    expect(violations).toEqual([])
  })

  it('keeps cards below the oversized 3xl corner radius', () => {
    const violations = files
      .filter((path) => readFileSync(path, 'utf8').includes('rounded-3xl'))
      .map((path) => relative(frontendRoot, path))

    expect(violations).toEqual([])
  })

  it('does not leave malformed semantic utility names after token migration', () => {
    const malformed = files.flatMap((path) => {
      const source = readFileSync(path, 'utf8')
      return Array.from(source.matchAll(/(?:line|surface-muted|line-strong)0(?:\/\d+)?/g), (match) =>
        `${relative(frontendRoot, path)}: ${match[0]}`
      )
    })

    expect(malformed).toEqual([])
  })

  it('uses the shared public shell on standalone public routes', () => {
    const publicRoutes = [
      'src/views/KeyUsageView.vue',
      'src/views/ModelPlazaView.vue',
      'src/views/public/LegalDocumentView.vue',
      'src/views/setup/SetupWizardView.vue',
      'src/views/auth/OAuthCallbackView.vue',
      'src/views/auth/WechatPaymentCallbackView.vue',
      'src/views/user/PaymentResultView.vue',
      'src/views/user/StripePopupView.vue'
    ]

    const violations = publicRoutes.filter(
      (path) => !readFileSync(resolve(frontendRoot, path), 'utf8').includes('public-shell')
    )
    expect(violations).toEqual([])
  })

  it('keeps the fixed homepage template connected to the configured logo', () => {
    const template = readFileSync(resolve(frontendRoot, '../aivoza-home-pixel.html'), 'utf8')
    expect(template.match(/\{\{SITE_LOGO\}\}/g)).toHaveLength(1)
    expect(template).toContain('src="{{SITE_LOGO}}"')
    expect(template).not.toMatch(/<img[^>]+src=["']data:image\//i)
    expect(template).toContain('.dark .aivoza-home')

    const declaredPixelSizes = Array.from(template.matchAll(/font-size:\s*([^;]+);/g))
      .flatMap(([, declaration]) =>
        Array.from(declaration.matchAll(/(\d+(?:\.\d+)?)px/g), (match) => Number(match[1]))
      )
    expect(Math.max(...declaredPixelSizes)).toBeLessThanOrEqual(40)
  })

  it('keeps executable content out of the build-time homepage asset', () => {
    const template = readFileSync(resolve(frontendRoot, '../aivoza-home-pixel.html'), 'utf8')
    expect(template).not.toMatch(/<script\b/i)
    expect(template).not.toMatch(/\son[a-z]+\s*=/i)
    expect(template).not.toMatch(/javascript\s*:/i)
    expect(template).not.toMatch(/<iframe\b/i)
  })

  it('loads the homepage from the reviewed asset instead of public settings', () => {
    const homeView = readFileSync(resolve(frontendRoot, 'src/views/HomeView.vue'), 'utf8')
    expect(homeView).toContain("../../../aivoza-home-pixel.html?raw")
    expect(homeView).not.toContain('home_content')
    expect(homeView).not.toContain('compact_home_enabled')
  })
})
