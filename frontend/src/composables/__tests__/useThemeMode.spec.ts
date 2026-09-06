import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import {
  initThemeMode,
  stopThemeMode,
  useThemeMode,
} from '@/composables/useThemeMode'

type ThemeListener = (event: MediaQueryListEvent) => void

function installMatchMedia(initialDark = false) {
  let matches = initialDark
  const listeners = new Set<ThemeListener>()
  const media = {
    media: '(prefers-color-scheme: dark)',
    get matches() {
      return matches
    },
    onchange: null,
    addEventListener: (_type: string, listener: ThemeListener) => listeners.add(listener),
    removeEventListener: (_type: string, listener: ThemeListener) => listeners.delete(listener),
    addListener: (listener: ThemeListener) => listeners.add(listener),
    removeListener: (listener: ThemeListener) => listeners.delete(listener),
    dispatchEvent: () => true,
  } as MediaQueryList

  vi.stubGlobal('matchMedia', vi.fn(() => media))

  return {
    setSystemDark(nextDark: boolean) {
      matches = nextDark
      const event = { matches: nextDark, media: media.media } as MediaQueryListEvent
      listeners.forEach((listener) => listener(event))
    },
  }
}

describe('useThemeMode', () => {
  beforeEach(() => {
    stopThemeMode()
    document.documentElement.classList.remove('dark')
    document.documentElement.style.colorScheme = ''
    window.localStorage.clear()
  })

  afterEach(() => {
    stopThemeMode()
    document.documentElement.classList.remove('dark')
    document.documentElement.style.colorScheme = ''
    window.localStorage.clear()
    vi.unstubAllGlobals()
  })

  it('follows the operating-system preference when no explicit choice exists', async () => {
    const media = installMatchMedia(true)
    initThemeMode()

    expect(document.documentElement.classList.contains('dark')).toBe(true)
    expect(document.documentElement.style.colorScheme).toBe('dark')
    expect(useThemeMode().themePreference.value).toBe('system')

    media.setSystemDark(false)
    await nextTick()

    expect(document.documentElement.classList.contains('dark')).toBe(false)
    expect(document.documentElement.style.colorScheme).toBe('light')
  })

  it('persists an explicit choice and no longer follows system changes', async () => {
    const media = installMatchMedia(false)
    initThemeMode()
    const { setThemePreference } = useThemeMode()

    setThemePreference('dark')
    expect(window.localStorage.getItem('theme')).toBe('dark')
    expect(document.documentElement.classList.contains('dark')).toBe(true)

    media.setSystemDark(false)
    await nextTick()
    expect(document.documentElement.classList.contains('dark')).toBe(true)

    setThemePreference('system')
    expect(window.localStorage.getItem('theme')).toBeNull()
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('synchronizes explicit and system preferences from another tab', () => {
    installMatchMedia(false)
    initThemeMode()

    window.dispatchEvent(new StorageEvent('storage', { key: 'theme', newValue: 'dark' }))
    expect(document.documentElement.classList.contains('dark')).toBe(true)
    expect(useThemeMode().themePreference.value).toBe('dark')

    window.dispatchEvent(new StorageEvent('storage', { key: 'theme', newValue: null }))
    expect(document.documentElement.classList.contains('dark')).toBe(false)
    expect(useThemeMode().themePreference.value).toBe('system')
  })
})
