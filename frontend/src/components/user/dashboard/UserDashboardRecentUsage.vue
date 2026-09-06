<template>
  <section class="card overflow-hidden" :aria-label="t('dashboard.recentUsage')">
    <div class="flex flex-wrap items-center justify-between gap-3 border-b-2 border-line px-4 py-4 dark:border-line sm:px-5">
      <div class="flex items-center gap-2">
        <span class="h-2.5 w-2.5 rounded-sm bg-primary-400" aria-hidden="true" />
        <h2 class="text-base font-bold text-ink-strong dark:text-white sm:text-lg">
          {{ t('dashboard.recentUsage') }}
        </h2>
      </div>
      <span class="badge badge-warning border-2 border-amber-300 dark:border-amber-700">
        {{ t('dashboard.last7Days') }}
      </span>
    </div>
    <div class="p-4 sm:p-5">
      <div v-if="loading" class="flex items-center justify-center py-12" aria-live="polite">
        <LoadingSpinner size="lg" />
      </div>
      <div v-else-if="data.length === 0" class="py-8">
        <EmptyState :title="t('dashboard.noUsageRecords')" :description="t('dashboard.startUsingApi')" />
      </div>
      <div v-else class="space-y-3">
        <article
          v-for="log in data"
          :key="log.id"
          class="flex flex-col gap-3 rounded-xl border-2 border-line bg-surface-muted/80 p-3 shadow-sm transition-colors duration-150 hover:border-primary-300 hover:bg-primary-50/70 dark:border-line-strong dark:bg-canvas/35 dark:hover:border-primary-700 dark:hover:bg-primary-900/10 sm:flex-row sm:items-center sm:justify-between sm:p-4"
        >
          <div class="flex min-w-0 items-center gap-3">
            <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-primary-300 bg-primary-100 dark:border-primary-700 dark:bg-primary-900/30">
              <Icon name="beaker" size="md" class="text-primary-600 dark:text-primary-400" />
            </div>
            <div class="min-w-0">
              <p class="truncate text-sm font-bold text-ink-strong dark:text-white" :title="log.model">
                {{ log.model }}
              </p>
              <p class="mt-0.5 text-xs leading-5 text-ink dark:text-ink-muted">
                {{ formatDateTime(log.created_at) }}
              </p>
            </div>
          </div>
          <div class="border-l-2 border-accent-300 pl-3 text-left dark:border-accent-700 sm:border-l-0 sm:pl-0 sm:text-right">
            <p class="text-sm font-bold tabular-nums">
              <span class="text-primary-700 dark:text-primary-300" :title="t('dashboard.actual')">
                ${{ formatCost(log.actual_cost) }}
              </span>
              <span class="font-medium text-ink-muted dark:text-ink-muted" :title="t('dashboard.standard')">
                / ${{ formatCost(log.total_cost) }}
              </span>
            </p>
            <p class="mt-0.5 text-xs tabular-nums text-ink dark:text-ink-muted">
              {{ (log.input_tokens + log.output_tokens).toLocaleString() }} tokens
            </p>
          </div>
        </article>

        <router-link
          to="/usage"
          class="flex min-h-11 items-center justify-center gap-2 rounded-xl border-2 border-line bg-white px-4 py-2 text-sm font-bold text-primary-700 shadow-sm transition-[border-color,background-color,color] duration-150 hover:border-primary-400 hover:bg-primary-50 dark:border-line-strong dark:bg-surface dark:text-primary-300 dark:hover:border-primary-500 dark:hover:bg-primary-900/15"
        >
          {{ t('dashboard.viewAllUsage') }}
          <Icon name="arrowRight" size="sm" aria-hidden="true" />
        </router-link>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import type { UsageLog } from '@/types'

defineProps<{
  data: UsageLog[]
  loading: boolean
}>()
const { t } = useI18n()
const formatCost = (c: number) => c.toFixed(4)
</script>
