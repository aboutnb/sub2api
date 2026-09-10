<template>
  <section class="card overflow-hidden" :aria-label="t('dashboard.quickActions')">
    <div class="flex items-center gap-2 border-b-2 border-line px-4 py-4 dark:border-line sm:px-5">
      <span class="h-2.5 w-2.5 rounded-sm bg-accent-400" aria-hidden="true" />
      <h2 class="text-base font-bold text-ink-strong dark:text-white sm:text-lg">
        {{ t('dashboard.quickActions') }}
      </h2>
    </div>
    <div class="space-y-3 p-4 sm:p-5">
      <button
        type="button"
        class="group flex min-h-[4.5rem] w-full items-center gap-3 rounded-xl border-2 border-line bg-surface-muted/80 p-3 text-left shadow-sm transition-[border-color,background-color,box-shadow] duration-150 hover:border-primary-400 hover:bg-primary-50 hover:shadow-card dark:border-line-strong dark:bg-canvas/35 dark:hover:border-primary-400 dark:hover:bg-primary-900/15"
        @click="router.push('/keys')"
      >
        <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-primary-300 bg-primary-100 dark:border-primary-700 dark:bg-primary-900/30">
          <Icon name="key" size="lg" class="text-primary-600 dark:text-primary-400" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-bold text-ink-strong dark:text-white">{{ t('dashboard.createApiKey') }}</p>
          <p class="mt-0.5 text-xs leading-5 text-ink dark:text-ink-muted">{{ t('dashboard.generateNewKey') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="flex-shrink-0 text-ink-muted transition-colors group-hover:text-primary-600 dark:text-ink-muted dark:group-hover:text-primary-300"
          aria-hidden="true"
        />
      </button>

      <button
        type="button"
        class="group flex min-h-[4.5rem] w-full items-center gap-3 rounded-xl border-2 border-line bg-surface-muted/80 p-3 text-left shadow-sm transition-[border-color,background-color,box-shadow] duration-150 hover:border-accent-400 hover:bg-accent-50 hover:shadow-card dark:border-line-strong dark:bg-canvas/35 dark:hover:border-accent-400 dark:hover:bg-accent-900/10"
        @click="router.push('/usage')"
      >
        <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-accent-300 bg-accent-100 dark:border-accent-700 dark:bg-accent-900/25">
          <Icon name="chart" size="lg" class="text-accent-700 dark:text-accent-300" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-bold text-ink-strong dark:text-white">{{ t('dashboard.viewUsage') }}</p>
          <p class="mt-0.5 text-xs leading-5 text-ink dark:text-ink-muted">{{ t('dashboard.checkDetailedLogs') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="flex-shrink-0 text-ink-muted transition-colors group-hover:text-accent-700 dark:text-ink-muted dark:group-hover:text-accent-300"
          aria-hidden="true"
        />
      </button>

      <button
        v-if="canUseBatchImage"
        type="button"
        class="group flex min-h-[4.5rem] w-full items-center gap-3 rounded-xl border-2 border-line bg-surface-muted/80 p-3 text-left shadow-sm transition-[border-color,background-color,box-shadow] duration-150 hover:border-primary-400 hover:bg-primary-50 hover:shadow-card dark:border-line-strong dark:bg-canvas/35 dark:hover:border-primary-400 dark:hover:bg-primary-900/15"
        @click="router.push('/batch-image')"
      >
        <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-primary-300 bg-primary-100 dark:border-primary-700 dark:bg-primary-900/30">
          <Icon name="sparkles" size="lg" class="text-primary-700 dark:text-primary-300" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-bold text-ink-strong dark:text-white">{{ t('dashboard.batchImageAgent') }}</p>
          <p class="mt-0.5 text-xs leading-5 text-ink dark:text-ink-muted">{{ t('dashboard.batchImageAgentDesc') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="flex-shrink-0 text-ink-muted transition-colors group-hover:text-primary-600 dark:text-ink-muted dark:group-hover:text-primary-300"
          aria-hidden="true"
        />
      </button>

      <button
        type="button"
        class="group flex min-h-[4.5rem] w-full items-center gap-3 rounded-xl border-2 border-line bg-amber-50/80 p-3 text-left shadow-sm transition-[border-color,background-color,box-shadow] duration-150 hover:border-amber-400 hover:bg-amber-100 hover:shadow-card dark:border-line-strong dark:bg-amber-900/10 dark:hover:border-amber-400 dark:hover:bg-amber-900/20"
        @click="router.push('/redeem')"
      >
        <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-amber-300 bg-amber-100 dark:border-amber-700 dark:bg-amber-900/30">
          <Icon name="gift" size="lg" class="text-amber-600 dark:text-amber-400" />
        </div>
        <div class="min-w-0 flex-1">
          <p class="text-sm font-bold text-ink-strong dark:text-white">{{ t('dashboard.redeemCode') }}</p>
          <p class="mt-0.5 text-xs leading-5 text-ink dark:text-ink-muted">{{ t('dashboard.addBalanceWithCode') }}</p>
        </div>
        <Icon
          name="chevronRight"
          size="md"
          class="flex-shrink-0 text-ink-muted transition-colors group-hover:text-amber-700 dark:text-ink-muted dark:group-hover:text-amber-300"
          aria-hidden="true"
        />
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useBatchImageAccess } from '@/composables/useBatchImageAccess'
const router = useRouter()
const { t } = useI18n()
const { canUseBatchImage, refreshBatchImageAccess } = useBatchImageAccess()

onMounted(() => {
  void refreshBatchImageAccess()
})
</script>
