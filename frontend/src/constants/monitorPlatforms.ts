import type { MonitorConfig } from '@/api/channelMonitorV2'
import { CONCRETE_PLATFORM_OPTIONS } from './platforms'

// Passive monitoring also retains the legacy Kiro dimension.
function monitorPlatformOptions() {
  return [...CONCRETE_PLATFORM_OPTIONS, { value: 'kiro', label: 'Kiro' }]
}

export function monitorPlatformLabel(value: string): string {
  if (value === 'anthropic') return 'Claude'
  if (value === 'composite') return 'Composite'
  return monitorPlatformOptions().find(option => option.value === value)?.label || value
}

export function completeMonitorPlatforms(platforms: MonitorConfig['platforms']): MonitorConfig['platforms'] {
  const result = platforms.map(platform => ({ ...platform, models: [...platform.models] }))
  const existing = new Set(result.map(platform => platform.platform))
  for (const option of monitorPlatformOptions()) {
    // Missing future platforms remain opt-in until the operator enables them.
    if (!existing.has(option.value)) result.push({ platform: option.value, enabled: false, models: [] })
  }
  return result
}
