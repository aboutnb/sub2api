import { getCurrentInstance, onMounted, readonly, ref } from 'vue'

const isDark = ref(false)
const themePreference = ref<ThemePreference>('system')

export type ThemePreference = 'light' | 'dark' | 'system'

const THEME_STORAGE_KEY = 'theme'
const SYSTEM_DARK_QUERY = '(prefers-color-scheme: dark)'

let observer: MutationObserver | null = null
let mediaQuery: MediaQueryList | null = null
let listenersAttached = false

function getStoredPreference(): ThemePreference {
  if (typeof window === 'undefined') return 'system'

  try {
    const stored = window.localStorage.getItem(THEME_STORAGE_KEY)
    return stored === 'light' || stored === 'dark' ? stored : 'system'
  } catch {
    return 'system'
  }
}

function getSystemDarkPreference(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
    return false
  }
  return window.matchMedia(SYSTEM_DARK_QUERY).matches
}

function resolveDark(preference: ThemePreference): boolean {
  if (preference === 'dark') return true
  if (preference === 'light') return false
  return getSystemDarkPreference()
}

function syncThemeFromDom(): void {
  if (typeof document === 'undefined') {
    return
  }

  const nextDark = document.documentElement.classList.contains('dark')
  isDark.value = nextDark
  document.documentElement.style.colorScheme = nextDark ? 'dark' : 'light'
}

function applyTheme(preference: ThemePreference): void {
  if (typeof document === 'undefined') return

  document.documentElement.classList.toggle('dark', resolveDark(preference))
  syncThemeFromDom()
}

function persistPreference(preference: ThemePreference): void {
  if (typeof window === 'undefined') return

  try {
    if (preference === 'system') {
      window.localStorage.removeItem(THEME_STORAGE_KEY)
    } else {
      window.localStorage.setItem(THEME_STORAGE_KEY, preference)
    }
  } catch {
    // The DOM theme remains usable even if persistence is unavailable.
  }
}

function handleSystemThemeChange(event: MediaQueryListEvent): void {
  if (themePreference.value !== 'system' || typeof document === 'undefined') return
  document.documentElement.classList.toggle('dark', event.matches)
  syncThemeFromDom()
}

function handleStorageChange(event: StorageEvent): void {
  if (event.key !== THEME_STORAGE_KEY && event.key !== null) return
  themePreference.value = event.newValue === 'light' || event.newValue === 'dark'
    ? event.newValue
    : 'system'
  applyTheme(themePreference.value)
}

function attachGlobalListeners(): void {
  if (listenersAttached || typeof window === 'undefined') return

  listenersAttached = true
  window.addEventListener('storage', handleStorageChange)

  if (typeof window.matchMedia === 'function') {
    mediaQuery = window.matchMedia(SYSTEM_DARK_QUERY)
    if (typeof mediaQuery.addEventListener === 'function') {
      mediaQuery.addEventListener('change', handleSystemThemeChange)
    } else {
      mediaQuery.addListener(handleSystemThemeChange)
    }
  }
}

function ensureThemeTracking(): void {
  if (typeof window === 'undefined' || typeof document === 'undefined') {
    return
  }

  syncThemeFromDom()

  if (observer) {
    return
  }

  observer = new MutationObserver(() => {
    syncThemeFromDom()
  })

  observer.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['class'],
  })
}

function setTheme(nextDark: boolean): void {
  setThemePreference(nextDark ? 'dark' : 'light')
}

function setThemePreference(preference: ThemePreference): void {
  themePreference.value = preference
  persistPreference(preference)
  applyTheme(preference)
  ensureThemeTracking()
  attachGlobalListeners()
}

function toggleTheme(): void {
  setTheme(!isDark.value)
}

export function initThemeMode(): void {
  themePreference.value = getStoredPreference()
  applyTheme(themePreference.value)
  ensureThemeTracking()
  attachGlobalListeners()
}

export function stopThemeMode(): void {
  observer?.disconnect()
  observer = null

  if (typeof window !== 'undefined') {
    window.removeEventListener('storage', handleStorageChange)
  }

  if (mediaQuery) {
    if (typeof mediaQuery.removeEventListener === 'function') {
      mediaQuery.removeEventListener('change', handleSystemThemeChange)
    } else {
      mediaQuery.removeListener(handleSystemThemeChange)
    }
  }

  mediaQuery = null
  listenersAttached = false
}

export function useThemeMode() {
  if (getCurrentInstance()) {
    onMounted(() => {
      ensureThemeTracking()
    })
  } else {
    ensureThemeTracking()
  }

  return {
    isDark: readonly(isDark),
    themePreference: readonly(themePreference),
    setTheme,
    setThemePreference,
    toggleTheme,
    syncThemeFromDom,
  }
}
