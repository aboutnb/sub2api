<template>
  <div
    class="not-found-shell relative flex min-h-screen min-h-dvh items-center justify-center overflow-hidden px-4 py-10"
  >
    <div class="not-found-grid pointer-events-none absolute inset-0" aria-hidden="true"></div>

    <main class="not-found-card relative z-10 w-full max-w-md text-center">
      <div class="mb-5 flex items-center justify-center gap-3" aria-hidden="true">
        <span class="not-found-code">404</span>
        <span class="not-found-icon flex h-12 w-12 items-center justify-center rounded-xl">
          <Icon name="exclamationTriangle" size="lg" />
        </span>
      </div>

      <div class="mb-7">
        <h1 class="mb-3 text-2xl font-extrabold text-ink-strong dark:text-white">
          {{ t('errors.pageNotFound') }}
        </h1>
        <p class="text-sm leading-6 text-ink dark:text-ink">
          The page you are looking for doesn't exist or has been moved.
        </p>
      </div>

      <div class="flex flex-col justify-center gap-3 sm:flex-row">
        <button @click="goBack" class="btn btn-secondary">
          <Icon name="arrowLeft" size="md" class="mr-2" />
          Go Back
        </button>
        <router-link to="/dashboard" class="btn btn-primary">
          <Icon name="home" size="md" class="mr-2" />
          Go to Dashboard
        </router-link>
      </div>

      <p class="mt-7 text-sm text-ink-muted dark:text-ink-muted">
        Need help?
        <a
          href="#"
          class="text-primary-600 transition-colors hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
        >
          Contact support
        </a>
      </p>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const router = useRouter()

function goBack(): void {
  router.back()
}
</script>

<style scoped>
.not-found-shell {
  isolation: isolate;
  background: var(--av-paper, #fffaf4);
}

.not-found-grid {
  z-index: -1;
  background-image:
    radial-gradient(circle at 12% 8%, rgba(255, 138, 92, 0.14), transparent 22rem),
    radial-gradient(circle at 88% 4%, rgba(47, 185, 170, 0.15), transparent 24rem),
    linear-gradient(var(--av-grid, rgba(23, 32, 51, 0.035)) 1px, transparent 1px),
    linear-gradient(90deg, var(--av-grid, rgba(23, 32, 51, 0.035)) 1px, transparent 1px);
  background-size: auto, auto, 32px 32px, 32px 32px;
}

.not-found-card {
  border: 2px solid var(--av-ink, #172033);
  border-radius: var(--av-radius-lg, 18px);
  padding: 2rem;
  background: var(--av-surface, #ffffff);
  box-shadow: 6px 6px 0 rgba(23, 32, 51, 0.16);
}

.not-found-card::before {
  position: absolute;
  top: 11px;
  left: 18px;
  width: 38px;
  height: 4px;
  background: linear-gradient(90deg, var(--av-coral, #ff8a5c) 0 45%, transparent 45% 56%, var(--av-teal, #2fb9aa) 56% 100%);
  content: '';
}

.not-found-code {
  color: var(--av-ink, #172033);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 2.25rem;
  font-weight: 900;
  letter-spacing: -0.06em;
}

.not-found-icon {
  border: 2px solid var(--av-ink, #172033);
  background: var(--av-yellow, #f7c95c);
  color: #172033;
  box-shadow: 3px 3px 0 rgba(23, 32, 51, 0.18);
}

:global(.dark) .not-found-card {
  box-shadow: 6px 6px 0 rgba(0, 0, 0, 0.38);
}

@media (max-width: 420px) {
  .not-found-card {
    padding: 1.5rem;
    box-shadow: 4px 4px 0 rgba(23, 32, 51, 0.16);
  }
}
</style>
