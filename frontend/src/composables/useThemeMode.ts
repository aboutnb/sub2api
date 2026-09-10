import { onMounted, readonly, ref } from 'vue'

const isDark = ref(false)

let observer: MutationObserver | null = null

function syncThemeFromDom(): void {
  if (typeof document === 'undefined') {
    return
  }
  isDark.value = document.documentElement.classList.contains('dark')
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
  ensureThemeTracking()
  document.documentElement.classList.toggle('dark', nextDark)
  localStorage.setItem('theme', nextDark ? 'dark' : 'light')
  syncThemeFromDom()
}

function toggleTheme(): void {
  setTheme(!isDark.value)
}

export function initThemeMode(): void {
  ensureThemeTracking()
}

export function useThemeMode() {
  onMounted(() => {
    ensureThemeTracking()
  })

  return {
    isDark: readonly(isDark),
    setTheme,
    toggleTheme,
    syncThemeFromDom,
  }
}
