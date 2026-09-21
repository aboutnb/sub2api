<template>
  <AppLayout>
    <iframe v-if="auth.user && auth.token" :key="sessionKey" ref="frame" src="/image-studio-app/" :title="t('nav.imageStudio')" class="studio-frame" :style="frameHeight ? { height: `${frameHeight}px` } : undefined" @load="syncAppearance" />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useThemeMode } from '@/composables/useThemeMode'
import { installStudioBridge } from '@/features/image-studio/bridge'
import { STUDIO_CHANNEL } from '../../../../shared/image-studio'
import { STUDIO_THEME_ROLES, validStudioColor } from '../../../../shared/image-studio-theme'

const auth = useAuthStore()
const { t, locale } = useI18n()
const theme = useThemeMode()
const frame = ref<HTMLIFrameElement>()
const frameHeight = ref(0)
let resizeObserver: ResizeObserver | undefined
function fitFrame() {
  if (!frame.value) return
  const main = frame.value.closest('main')
  const bottomPadding = main ? parseFloat(getComputedStyle(main).paddingBottom) : 16
  frameHeight.value = Math.max(320, (window.visualViewport?.height ?? window.innerHeight) - frame.value.getBoundingClientRect().top - bottomPadding)
}
const sessionKey = computed(() => `${auth.user?.id ?? ''}:${Boolean(auth.token)}`)
const appearance = () => {
  const styles = getComputedStyle(document.documentElement)
  const tokens = Object.fromEntries(STUDIO_THEME_ROLES.flatMap((role) => {
    const value = styles.getPropertyValue(`--av-rgb-${role}`).trim().replace(/\s+/g, ' ')
    return validStudioColor(value) ? [[role, value]] : []
  }))
  return { theme: theme.isDark.value ? 'dark' : 'light', lang: locale.value, tokens }
}
let dispose: (() => void) | undefined
watch(frame, (element) => {
  dispose?.()
  resizeObserver?.disconnect()
  if (!element) return
  const userID = auth.user?.id
  dispose = installStudioBridge(element, () => !!auth.token && auth.user?.id === userID, appearance)
  fitFrame()
  const header = element.closest('.app-workspace')?.querySelector('header')
  if (header) { resizeObserver = new ResizeObserver(fitFrame); resizeObserver.observe(header) }
}, { flush: 'post' })
function syncAppearance() {
  fitFrame()
  frame.value?.contentWindow?.postMessage({ channel: STUDIO_CHANNEL, operation: 'appearance', ...appearance() }, window.location.origin)
}
watch([() => theme.isDark.value, locale], syncAppearance, { flush: 'post' })
watch(sessionKey, () => dispose?.(), { flush: 'sync' })
onMounted(() => {
  window.addEventListener('resize', fitFrame)
  window.visualViewport?.addEventListener('resize', fitFrame)
})
onBeforeUnmount(() => {
  dispose?.()
  resizeObserver?.disconnect()
  window.removeEventListener('resize', fitFrame)
  window.visualViewport?.removeEventListener('resize', fitFrame)
})
</script>

<style scoped>
.studio-frame { display: block; width: 100%; height: calc(100dvh - 9rem); min-height: 320px; border: 0; }
</style>
