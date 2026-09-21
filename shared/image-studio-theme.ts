export const STUDIO_THEME_ROLES = [
  'bg-canvas', 'bg-surface', 'bg-surface-muted', 'bg-surface-elevated',
  'text-strong', 'text-default', 'text-muted', 'text-inverse',
  'border-default', 'border-strong', 'border-control', 'focus',
  'brand-primary', 'brand-primary-soft', 'brand-decorative',
  'action', 'action-hover', 'action-soft', 'action-foreground',
  'info', 'success', 'warning', 'danger',
] as const

export type StudioThemeTokens = Partial<Record<typeof STUDIO_THEME_ROLES[number], string>>

export function validStudioColor(value: unknown): value is string {
  return typeof value === 'string' && /^\d{1,3} \d{1,3} \d{1,3}$/.test(value) && value.split(' ').every((part) => Number(part) <= 255)
}
