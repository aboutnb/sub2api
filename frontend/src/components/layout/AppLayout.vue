<template>
  <div class="app-shell min-h-screen">
    <a class="skip-to-main" href="#app-main-content" @click.prevent="focusMain">
      {{ t('common.skipToMain') === 'common.skipToMain' ? 'Skip to main content' : t('common.skipToMain') }}
    </a>
    <!-- Warm paper and subtle pixel-grid decoration shared by every signed-in page. -->
    <div class="app-shell-backdrop pointer-events-none fixed inset-0" aria-hidden="true"></div>
    <WaterRippleBackdrop />

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="app-workspace relative min-h-screen w-full min-w-0 max-w-full overflow-x-clip transition-all duration-300"
      :class="[
        sidebarCollapsed
          ? 'lg:ml-[72px] lg:w-[calc(100%-72px)]'
          : 'lg:ml-64 lg:w-[calc(100%-16rem)]'
      ]"
      :aria-hidden="mobileSidebarModalOpen ? 'true' : undefined"
      :inert="mobileSidebarModalOpen ? true : undefined"
    >
      <!-- Header -->
      <AppHeader />

      <!-- Main Content -->
      <main
        id="app-main-content"
        ref="mainRef"
        class="app-main min-w-0 max-w-full p-4 focus:outline-none md:p-6 lg:p-8"
        tabindex="-1"
      >
        <div class="app-main-content mx-auto w-full min-w-0 max-w-[1680px]">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'
import WaterRippleBackdrop from './WaterRippleBackdrop.vue'

const appStore = useAppStore()
const authStore = useAuthStore()
const route = useRoute()
const { t } = useI18n()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const mobileOpen = computed(() => appStore.mobileOpen)
const isAdmin = computed(() => authStore.user?.role === 'admin')
const desktopViewportMedia = typeof window === 'undefined'
  ? null
  : window.matchMedia('(min-width: 1024px)')
const isDesktopViewport = ref(desktopViewportMedia?.matches ?? true)
const mobileSidebarModalOpen = computed(() => mobileOpen.value && !isDesktopViewport.value)
const mainRef = ref<HTMLElement | null>(null)

function focusMain() {
  mainRef.value?.focus({ preventScroll: true })
}

function handleDesktopViewportChange(event: MediaQueryListEvent) {
  isDesktopViewport.value = event.matches
}

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  desktopViewportMedia?.addEventListener('change', handleDesktopViewportChange)
  onboardingStore.setReplayCallback(replayTour)
})

onBeforeUnmount(() => {
  desktopViewportMedia?.removeEventListener('change', handleDesktopViewportChange)
})

watch(
  () => route.fullPath,
  async (nextPath, previousPath) => {
    if (nextPath === previousPath) return
    await nextTick()
    focusMain()
  },
  { flush: 'post' },
)

defineExpose({ replayTour })
</script>

<style scoped>
.skip-to-main {
  position: fixed;
  left: 1rem;
  top: 0.75rem;
  z-index: 100000100;
  min-height: 44px;
  max-width: calc(100vw - 2rem);
  padding: 0.625rem 1rem;
  border: 2px solid var(--av-border-strong, #172033);
  border-radius: 0.75rem;
  background: var(--av-surface, #ffffff);
  color: var(--av-ink, #172033);
  font-size: 0.875rem;
  font-weight: 700;
  opacity: 0;
  pointer-events: none;
  transform: translateY(-150%);
}

.skip-to-main:focus-visible {
  opacity: 1;
  pointer-events: auto;
  transform: translateY(0);
  outline: 3px solid var(--av-teal, #2fb9aa);
  outline-offset: 2px;
}
</style>
