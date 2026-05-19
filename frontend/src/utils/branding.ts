import { sanitizeUrl } from '@/utils/url'

export const FLOWAI_LIGHT_LOGO = '/flowai-logo-mark-light.svg'
export const FLOWAI_DARK_LOGO = '/flowai-logo-mark-dark.svg'

const THEME_MANAGED_LOGOS = new Set([
  '',
  '/flowai-logo-mark.svg',
  FLOWAI_LIGHT_LOGO,
  FLOWAI_DARK_LOGO,
])

export function resolveBrandLogo(logoUrl: string | null | undefined, isDark: boolean): string {
  const sanitizedLogo = sanitizeUrl(logoUrl ?? '', {
    allowRelative: true,
    allowDataUrl: true,
  })

  if (!sanitizedLogo || THEME_MANAGED_LOGOS.has(sanitizedLogo)) {
    return isDark ? FLOWAI_DARK_LOGO : FLOWAI_LIGHT_LOGO
  }

  return sanitizedLogo
}
