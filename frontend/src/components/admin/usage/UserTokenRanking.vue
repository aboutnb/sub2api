<template>
  <!-- 用量页"用户排行"tab 内容：无卡片外观，依赖父级统一卡片；筛选/时间范围复用页面级筛选栏 -->
  <div>
    <!-- Toolbar -->
    <div class="flex flex-wrap items-center justify-between gap-3 border-b border-line px-4 py-3 dark:border-line/50 sm:px-6">
      <p class="text-xs text-ink-muted dark:text-ink-muted">{{ t('admin.usage.tokenRanking.subtitle') }}</p>
      <div class="flex items-center gap-3">
        <span v-if="!loading && items.length > 0" class="text-xs text-ink-muted dark:text-ink-muted">
          {{ t('admin.usage.tokenRanking.userCount', { count: items.length }) }}
        </span>
        <div class="w-28">
          <Select
            v-model="limit"
            :options="limitOptions"
            :aria-label="t('admin.usage.tokenRanking.userCount', { count: limit })"
            @change="load"
          />
        </div>
      </div>
    </div>

    <!-- Table -->
    <div
      class="ranking-scroll-region overflow-x-auto"
      role="region"
      tabindex="0"
      :aria-label="t('admin.usage.tokenRanking.subtitle')"
    >
      <table class="w-full min-w-max divide-y divide-line dark:divide-line">
        <thead class="bg-surface-muted dark:bg-surface">
          <tr>
            <th class="w-16 px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-ink-muted dark:text-ink-muted sm:px-6">#</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-ink-muted dark:text-ink-muted">
              {{ t('admin.usage.tokenRanking.columns.user') }}
            </th>
            <th
              v-for="col in sortableColumns"
              :key="col.key"
              class="select-none whitespace-nowrap px-2 py-1 text-right text-xs font-medium uppercase tracking-wider"
              :class="sortBy === col.key ? 'text-primary-600 dark:text-primary-400' : 'text-ink-muted dark:text-ink-muted'"
              :aria-sort="sortBy === col.key ? 'descending' : 'none'"
            >
              <button
                type="button"
                data-touch-target="compact-visualization"
                class="inline-flex min-h-9 w-full items-center justify-end gap-1 rounded-lg px-2 py-1 text-right transition-colors hover:bg-surface-muted focus-visible:bg-surface-muted dark:hover:bg-dark-700 dark:focus-visible:bg-dark-700"
                @click="setSort(col.key)"
              >
                {{ t(col.label) }}
                <span v-if="sortBy === col.key" aria-hidden="true">↓</span>
              </button>
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-line bg-white dark:divide-line dark:bg-canvas">
          <tr v-if="loading">
            <td :colspan="sortableColumns.length + 2" class="py-12 text-center">
              <LoadingSpinner />
            </td>
          </tr>
          <tr v-else-if="items.length === 0">
            <td :colspan="sortableColumns.length + 2" class="py-12 text-center text-sm text-ink-muted">
              {{ t('admin.dashboard.noDataAvailable') }}
            </td>
          </tr>
          <tr
            v-for="(item, index) in items"
            v-else
            :key="item.user_id"
            class="transition-colors hover:bg-surface-muted dark:hover:bg-dark-700/40"
          >
            <td class="px-4 py-3 sm:px-6">
              <span
                v-if="index < 3"
                class="inline-flex h-6 w-6 items-center justify-center rounded-full text-xs font-semibold"
                :class="RANK_BADGE_CLASSES[index]"
              >{{ index + 1 }}</span>
              <span v-else class="inline-block w-6 text-center text-sm tabular-nums text-ink-muted">{{ index + 1 }}</span>
            </td>
            <td class="max-w-[260px] px-2 py-1 text-sm font-medium text-ink dark:text-gray-200" :title="item.email">
              <button
                type="button"
                data-touch-target="compact-visualization"
                class="flex min-h-9 max-w-full items-center rounded-lg px-2 text-left hover:bg-brand-soft focus-visible:bg-brand-soft"
                :aria-label="`${t('admin.usage.tokenRanking.rowHint')}: ${item.email || `User #${item.user_id}`}`"
                @click="emit('select-user', item.user_id, item.email)"
              >
                <span class="truncate">{{ item.email || `User #${item.user_id}` }}</span>
                <span class="ml-1 flex-shrink-0 font-normal text-ink-muted dark:text-ink-muted">#{{ item.user_id }}</span>
              </button>
            </td>
            <td class="whitespace-nowrap px-4 py-3 text-right text-sm tabular-nums text-ink-muted dark:text-ink-muted">{{ item.requests.toLocaleString() }}</td>
            <td class="whitespace-nowrap px-4 py-3 text-right text-sm tabular-nums text-ink-muted dark:text-ink-muted">{{ fmtTokens(item.input_tokens) }}</td>
            <td class="whitespace-nowrap px-4 py-3 text-right text-sm tabular-nums text-ink-muted dark:text-ink-muted">{{ fmtTokens(item.output_tokens) }}</td>
            <td class="whitespace-nowrap px-4 py-3 text-right text-sm tabular-nums text-ink-muted dark:text-ink-muted">{{ fmtTokens(item.cache_tokens) }}</td>
            <td class="whitespace-nowrap px-4 py-3 text-right text-sm font-medium tabular-nums text-ink-strong dark:text-gray-100">{{ fmtTokens(item.total_tokens) }}</td>
            <td class="whitespace-nowrap px-4 py-3 text-right text-sm font-medium tabular-nums text-green-600 dark:text-green-400">${{ fmtCost(item.actual_cost) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getUserBreakdown, type UserBreakdownParams } from '@/api/admin/dashboard'
import { formatCompactNumber, formatCostFixed } from '@/utils/format'
import type { UserBreakdownItem } from '@/types'
import Select from '@/components/common/Select.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const props = defineProps<{
  startDate: string
  endDate: string
  filters: Record<string, unknown>
  model?: string
}>()

const emit = defineEmits<{ (e: 'select-user', userId: number, email: string): void }>()

const { t } = useI18n()

type SortKey = NonNullable<UserBreakdownParams['sort_by']>
const sortableColumns: { key: SortKey; label: string }[] = [
  { key: 'requests', label: 'admin.usage.tokenRanking.columns.requests' },
  { key: 'input_tokens', label: 'admin.usage.tokenRanking.columns.inputTokens' },
  { key: 'output_tokens', label: 'admin.usage.tokenRanking.columns.outputTokens' },
  { key: 'cache_tokens', label: 'admin.usage.tokenRanking.columns.cacheTokens' },
  { key: 'total_tokens', label: 'admin.usage.tokenRanking.columns.totalTokens' },
  { key: 'actual_cost', label: 'admin.usage.tokenRanking.columns.cost' },
]

const limitOptions = [
  { value: 20, label: 'Top 20' },
  { value: 50, label: 'Top 50' },
  { value: 100, label: 'Top 100' },
  { value: 200, label: 'Top 200' },
]

// 前三名金/银/铜徽章
const RANK_BADGE_CLASSES = [
  'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-400',
  'bg-line text-ink dark:bg-surface-muted/20 dark:text-ink-muted',
  'bg-orange-100 text-orange-700 dark:bg-orange-500/20 dark:text-orange-400',
]

const items = ref<UserBreakdownItem[]>([])
const loading = ref(false)
const sortBy = ref<SortKey>('total_tokens')
const limit = ref(50)
let reqSeq = 0

const fmtTokens = (v: number) => formatCompactNumber(v)
const fmtCost = (v: number) => formatCostFixed(v, 4)

const setSort = (key: SortKey) => {
  if (sortBy.value === key) return
  sortBy.value = key
  load()
}

const load = async () => {
  const seq = ++reqSeq
  loading.value = true
  try {
    const params: UserBreakdownParams = {
      ...props.filters,
      start_date: props.startDate,
      end_date: props.endDate,
      sort_by: sortBy.value,
      limit: limit.value,
    }
    if (props.model) params.model = props.model
    const res = await getUserBreakdown(params)
    if (seq !== reqSeq) return
    items.value = res.users || []
  } catch {
    if (seq !== reqSeq) return
    items.value = []
  } finally {
    if (seq === reqSeq) loading.value = false
  }
}

// Reload when the shared filters / date range / model change.
watch(
  () => [props.startDate, props.endDate, props.model, JSON.stringify(props.filters)],
  () => load(),
  { immediate: true }
)

defineExpose({ reload: load })
</script>

<style scoped>
.ranking-scroll-region:focus-visible {
  outline: 3px solid var(--av-focus);
  outline-offset: -3px;
}
</style>
