<template>
  <div class="card p-4">
    <h3 class="mb-4 text-sm font-semibold text-ink-strong dark:text-white">
      {{ t('payment.admin.dailyRevenue') }}
    </h3>
    <div class="h-64">
      <div v-if="loading" class="flex h-full items-center justify-center">
        <LoadingSpinner size="md" />
      </div>
      <Line v-else-if="chartData" :data="chartData" :options="chartOptions" />
      <div
        v-else
        class="flex h-full items-center justify-center text-sm text-ink-muted dark:text-ink-muted"
      >
        {{ t('payment.admin.noData') }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { useChartTheme } from '@/composables/useChartTheme'
import type { DailyPaymentStats } from '@/types/payment'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

const { t } = useI18n()
const { chartTheme } = useChartTheme()

const props = defineProps<{
  data: DailyPaymentStats[]
  loading?: boolean
}>()

const chartData = computed(() => {
  if (!props.data || props.data.length === 0) return null
  const currencies = [...new Set(props.data.flatMap(day => Object.keys(day.amount)))].sort()
  return {
    labels: props.data.map(d => d.date),
    datasets: [
      ...currencies.map((currency, index) => {
        const paletteIndex = index % chartTheme.value.series.length
        return {
          label: `${currency} ${t('payment.admin.revenue')}`,
          data: props.data.map(day => day.amount[currency] || 0),
          borderColor: chartTheme.value.series[paletteIndex],
          backgroundColor: chartTheme.value.seriesSoft[paletteIndex],
          fill: true,
          tension: 0.3,
          pointRadius: 3,
          pointHoverRadius: 5,
        }
      }),
      {
        label: t('payment.admin.orderCount'),
        data: props.data.map(d => d.count),
        borderColor: chartTheme.value.series[7],
        backgroundColor: chartTheme.value.seriesSoft[7],
        fill: false,
        tension: 0.3,
        pointRadius: 3,
        pointHoverRadius: 5,
        yAxisID: 'y1',
      }
    ]
  }
})

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: { mode: 'index' as const, intersect: false },
  scales: {
    x: {
      grid: { color: chartTheme.value.grid },
      ticks: { color: chartTheme.value.text },
    },
    y: {
      type: 'linear' as const,
      display: true,
      position: 'left' as const,
      title: { display: true, text: t('payment.admin.revenue'), color: chartTheme.value.text },
      grid: { color: chartTheme.value.grid },
      ticks: { color: chartTheme.value.text },
    },
    y1: {
      type: 'linear' as const,
      display: true,
      position: 'right' as const,
      title: { display: true, text: t('payment.admin.orderCount'), color: chartTheme.value.text },
      grid: { drawOnChartArea: false },
      ticks: { color: chartTheme.value.series[7] },
    }
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: { color: chartTheme.value.text },
    },
    tooltip: {
      backgroundColor: chartTheme.value.tooltipBackground,
      titleColor: chartTheme.value.tooltipTitle,
      bodyColor: chartTheme.value.tooltipBody,
      borderColor: chartTheme.value.tooltipBorder,
      borderWidth: 1,
    },
  }
}))
</script>
