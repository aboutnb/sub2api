<template>
  <div
    class="auth-shell relative flex min-h-screen min-h-dvh items-center justify-center overflow-x-hidden px-4 py-8 sm:px-6 sm:py-10"
  >
    <div class="auth-backdrop pointer-events-none absolute inset-0 overflow-hidden" aria-hidden="true">
      <span class="auth-pixels auth-pixels-top"></span>
      <span class="auth-pixels auth-pixels-bottom"></span>
    </div>

    <button
      type="button"
      class="auth-theme-toggle"
      :aria-label="isDark ? t('nav.lightMode') : t('nav.darkMode')"
      :title="isDark ? t('nav.lightMode') : t('nav.darkMode')"
      @click="toggleTheme"
    >
      <Icon :name="isDark ? 'sun' : 'moon'" size="md" aria-hidden="true" />
    </button>

    <main class="relative z-10 w-full max-w-md">
      <div class="mb-6 min-h-14">
        <div v-if="settingsLoaded" class="auth-brand mx-auto flex w-fit max-w-full items-center gap-3">
          <div class="auth-brand-mark flex shrink-0 items-center justify-center overflow-hidden">
            <img
              :src="siteLogo"
              :alt="`${siteName} logo`"
              class="h-full w-full object-contain"
            />
          </div>
          <div class="min-w-0 text-left">
            <h1 class="auth-brand-name text-xl font-bold leading-tight">
              {{ siteName }}
            </h1>
            <p class="auth-brand-subtitle mt-1 text-xs leading-relaxed">
              {{ siteSubtitle }}
            </p>
          </div>
        </div>
      </div>

      <section class="auth-card p-5 sm:p-7">
        <slot />
      </section>

      <div class="auth-footer mt-6 text-center text-sm">
        <slot name="footer" />
      </div>

      <div class="auth-copyright mt-7 text-center text-xs">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useThemeMode } from '@/composables/useThemeMode'
import { useAppStore } from '@/stores'
import { resolveBrandLogo } from '@/utils/branding'
import Icon from '@/components/icons/Icon.vue'

const appStore = useAppStore()
const { t } = useI18n()
const { isDark, toggleTheme } = useThemeMode()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => resolveBrandLogo(appStore.siteLogo, isDark.value))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
.auth-shell {
  isolation: isolate;
  background: var(--av-paper, #fffaf4);
  color: var(--av-ink, #172033);
}

.auth-backdrop {
  background-image:
    linear-gradient(var(--av-grid, rgba(23, 32, 51, 0.045)) 1px, transparent 1px),
    linear-gradient(90deg, var(--av-grid, rgba(23, 32, 51, 0.045)) 1px, transparent 1px);
  background-size: 32px 32px;
  -webkit-mask-image: linear-gradient(to bottom, black 0%, rgba(0, 0, 0, 0.62) 56%, transparent 100%);
  mask-image: linear-gradient(to bottom, black 0%, rgba(0, 0, 0, 0.62) 56%, transparent 100%);
}

.auth-theme-toggle {
  position: absolute;
  right: max(1rem, env(safe-area-inset-right));
  top: max(1rem, env(safe-area-inset-top));
  z-index: 20;
  display: inline-flex;
  min-width: 44px;
  min-height: 44px;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--av-ink, #172033);
  border-radius: 12px;
  background: var(--av-surface, #ffffff);
  color: var(--av-ink, #172033);
  box-shadow: 3px 3px 0 var(--av-teal, #2fb9aa);
}

.auth-theme-toggle:hover {
  transform: translateY(-1px);
}

.auth-theme-toggle:focus-visible {
  outline: 3px solid var(--av-coral, #ff8a5c);
  outline-offset: 2px;
}

.auth-pixels {
  position: absolute;
  width: 7px;
  height: 7px;
  border: 1px solid var(--av-ink, #172033);
  background: var(--av-teal, #2fb9aa);
  box-shadow:
    11px 0 0 var(--av-coral, #ff8a5c),
    22px 0 0 var(--av-yellow, #f7c95c),
    0 11px 0 var(--av-yellow, #f7c95c),
    11px 11px 0 var(--av-teal, #2fb9aa);
  opacity: 0.7;
}

.auth-pixels-top {
  right: max(22px, 7vw);
  top: 28px;
}

.auth-pixels-bottom {
  bottom: 35px;
  left: max(20px, 6vw);
  transform: rotate(180deg);
}

.auth-brand-mark {
  width: 52px;
  height: 52px;
  border: 2px solid var(--av-ink, #172033);
  border-radius: 13px;
  background: var(--av-surface, #ffffff);
  box-shadow: 4px 4px 0 var(--av-coral, #ff8a5c);
}

.auth-brand-name {
  max-width: 19rem;
  color: var(--av-ink, #172033);
  overflow-wrap: anywhere;
}

.auth-brand-subtitle {
  max-width: 19rem;
  color: var(--av-ink-soft, #566074);
  overflow-wrap: anywhere;
}

.auth-card {
  position: relative;
  border: 2px solid var(--av-ink, #172033);
  border-radius: var(--av-radius-lg, 18px);
  background: var(--av-surface, #ffffff);
  box-shadow: 6px 6px 0 rgba(23, 32, 51, 0.16);
}

.auth-card::before {
  content: '';
  position: absolute;
  top: 10px;
  left: 22px;
  width: 46px;
  height: 4px;
  border-radius: 999px;
  background: linear-gradient(
    90deg,
    var(--av-teal, #2fb9aa) 0 18px,
    transparent 18px 23px,
    var(--av-coral, #ff8a5c) 23px 34px,
    transparent 34px 39px,
    var(--av-yellow, #f7c95c) 39px 46px
  );
}

.auth-footer {
  color: var(--av-ink-soft, #566074);
}

.auth-copyright {
  color: var(--av-ink-soft, #566074);
  opacity: 0.78;
}

:global(.dark) .auth-card {
  box-shadow: 6px 6px 0 rgba(0, 0, 0, 0.38);
}

@media (max-width: 420px) {
  .auth-card {
    box-shadow: 4px 4px 0 rgba(23, 32, 51, 0.16);
  }

  :global(.dark) .auth-card {
    box-shadow: 4px 4px 0 rgba(0, 0, 0, 0.38);
  }
}

@media (prefers-reduced-motion: reduce) {
  .auth-shell,
  .auth-shell :deep(*) {
    scroll-behavior: auto !important;
    animation-duration: 1ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 1ms !important;
  }
}

@media (forced-colors: active) {
  .auth-brand-mark,
  .auth-card {
    border-color: CanvasText;
    box-shadow: none;
  }

  .auth-card::before,
  .auth-pixels {
    display: none;
  }
}
</style>
