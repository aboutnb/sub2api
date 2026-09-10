<template>
  <div class="space-y-5">
    <!-- Date Range Filter -->
    <div class="card relative overflow-hidden p-4 sm:p-5">
      <div
        class="absolute left-4 top-0 h-1 w-12 rounded-b-sm bg-accent-400 sm:left-5"
        aria-hidden="true"
      />
      <div class="flex flex-col gap-4 lg:flex-row lg:items-end">
        <div class="min-w-0 flex-1">
          <span class="mb-2 block text-xs font-bold uppercase tracking-wide text-ink dark:text-ink-muted">
            {{ t('dashboard.timeRange') }}
          </span>
          <DateRangePicker
            :start-date="startDate"
            :end-date="endDate"
            @update:startDate="$emit('update:startDate', $event)"
            @update:endDate="$emit('update:endDate', $event)"
            @change="$emit('dateRangeChange', $event)"
          />
        </div>
        <div class="min-w-0 sm:w-40">
          <span class="mb-2 block text-xs font-bold uppercase tracking-wide text-ink dark:text-ink-muted">
            {{ t('dashboard.granularity') }}
          </span>
          <div class="w-full">
            <Select
              :model-value="granularity"
              :options="[
                { value: 'day', label: t('dashboard.day') },
                { value: 'hour', label: t('dashboard.hour') }
              ]"
              @update:model-value="$emit('update:granularity', $event)"
              @change="$emit('granularityChange')"
            />
          </div>
        </div>
        <button
          type="button"
          class="btn btn-secondary w-full sm:w-auto"
          :disabled="loading"
          @click="$emit('refresh')"
        >
          {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <!-- Charts Grid -->
    <div class="grid grid-cols-1 gap-5 lg:grid-cols-2">
      <!-- Model Distribution Chart -->
      <section
        class="card relative overflow-hidden p-4 sm:p-5"
        :aria-busy="loading"
        :aria-label="t('dashboard.modelDistribution')"
      >
        <div
          class="absolute left-4 top-0 h-1 w-12 rounded-b-sm bg-primary-400 sm:left-5"
          aria-hidden="true"
        />
        <div
          v-if="loading"
          class="absolute inset-0 z-10 flex items-center justify-center bg-white/80 backdrop-blur-sm dark:bg-surface/80"
        >
          <LoadingSpinner size="md" />
        </div>
        <div class="mb-4 flex items-center gap-2">
          <span class="h-2.5 w-2.5 rounded-sm bg-primary-400" aria-hidden="true" />
          <h3 class="text-base font-bold text-ink-strong dark:text-white">
            {{ t('dashboard.modelDistribution') }}
          </h3>
        </div>
        <div class="flex flex-col items-center gap-5 sm:flex-row sm:items-start sm:gap-6">
          <div class="h-44 w-44 shrink-0 sm:h-48 sm:w-48">
            <Doughnut v-if="modelData" :data="modelData" :options="doughnutOptions" />
            <div
              v-else
              class="flex h-full items-center justify-center rounded-xl border-2 border-dashed border-line bg-surface-muted/70 px-4 text-center text-sm text-ink-muted dark:border-line-strong dark:bg-canvas/30 dark:text-ink-muted"
            >
              {{ t('dashboard.noDataAvailable') }}
            </div>
          </div>
          <div class="max-h-48 w-full min-w-0 flex-1 overflow-auto rounded-xl border-2 border-line dark:border-line-strong">
            <table class="min-w-[30rem] w-full text-xs">
              <caption class="sr-only">
                {{ t('dashboard.modelDistribution') }}
              </caption>
              <thead class="sticky top-0 bg-surface-muted dark:bg-canvas">
                <tr class="text-ink dark:text-ink-muted">
                  <th class="px-3 py-2 text-left font-bold">{{ t('dashboard.model') }}</th>
                  <th class="px-3 py-2 text-right font-bold">{{ t('dashboard.requests') }}</th>
                  <th class="px-3 py-2 text-right font-bold">{{ t('dashboard.tokens') }}</th>
                  <th class="px-3 py-2 text-right font-bold">{{ t('dashboard.actual') }}</th>
                  <th class="px-3 py-2 text-right font-bold">{{ t('dashboard.standard') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y-2 divide-line dark:divide-line">
                <tr v-for="(model, index) in models" :key="model.model">
                  <td
                    class="max-w-[9rem] px-3 py-2 font-semibold text-ink-strong dark:text-white"
                    :title="model.model"
                  >
                    <span class="flex min-w-0 items-center gap-2">
                      <span
                        class="h-2.5 w-2.5 flex-shrink-0 rounded-sm border border-ink-strong/20"
                        :style="{ backgroundColor: modelColor(index) }"
                        aria-hidden="true"
                      />
                      <span class="truncate">{{ model.model }}</span>
                    </span>
                  </td>
                  <td class="px-3 py-2 text-right tabular-nums text-ink dark:text-ink-muted">
                    {{ formatNumber(model.requests) }}
                  </td>
                  <td class="px-3 py-2 text-right tabular-nums text-ink dark:text-ink-muted">
                    {{ formatTokens(model.total_tokens) }}
                  </td>
                  <td class="px-3 py-2 text-right font-semibold tabular-nums text-primary-600 dark:text-primary-300">
                    ${{ formatCost(model.actual_cost) }}
                  </td>
                  <td class="px-3 py-2 text-right tabular-nums text-ink-muted dark:text-ink-muted">
                    ${{ formatCost(model.cost) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <!-- Token Usage Trend Chart -->
      <TokenUsageTrend :trend-data="trend" :loading="loading" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import { Doughnut } from 'vue-chartjs'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import { useChartTheme } from '@/composables/useChartTheme'
import type { TrendDataPoint, ModelStat } from '@/types'
import { formatCostFixed as formatCost, formatNumberLocaleString as formatNumber, formatTokensK as formatTokens } from '@/utils/format'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler } from 'chart.js'
ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{ loading: boolean, startDate: string, endDate: string, granularity: string, trend: TrendDataPoint[], models: ModelStat[] }>()
defineEmits(['update:startDate', 'update:endDate', 'update:granularity', 'dateRangeChange', 'granularityChange', 'refresh'])
const { t } = useI18n()
const { chartTheme } = useChartTheme()

const modelColor = (index: number) =>
  chartTheme.value.series[index % chartTheme.value.series.length]

const modelData = computed(() => !props.models?.length ? null : {
  labels: props.models.map((m: ModelStat) => m.model),
  datasets: [{
    data: props.models.map((m: ModelStat) => m.total_tokens),
    backgroundColor: props.models.map((_: ModelStat, index: number) => modelColor(index)),
    borderWidth: 0
  }]
})

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: chartTheme.value.tooltipBackground,
      titleColor: chartTheme.value.tooltipTitle,
      bodyColor: chartTheme.value.tooltipBody,
      borderColor: chartTheme.value.tooltipBorder,
      borderWidth: 1,
      callbacks: {
        label: (context: any) => `${context.label}: ${formatTokens(context.parsed)} tokens`
      }
    }
  }
}))
</script>
