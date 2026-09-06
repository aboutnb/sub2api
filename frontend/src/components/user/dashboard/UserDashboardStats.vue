<template>
  <div class="space-y-5">
    <!-- Row 1: Core Stats -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <!-- Balance -->
      <div v-if="!isSimple" class="card relative min-w-0 overflow-hidden p-4">
        <span class="absolute left-4 top-0 h-1 w-10 rounded-b-sm bg-primary-400" aria-hidden="true" />
        <div class="flex items-start gap-3">
          <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-primary-300 bg-primary-100 shadow-pixel-sm dark:border-primary-700 dark:bg-primary-900/30">
            <svg
              class="h-5 w-5 text-primary-700 dark:text-primary-300"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z" />
            </svg>
          </div>
          <div class="min-w-0">
            <p class="text-xs font-bold uppercase tracking-wide text-ink dark:text-ink-muted">
              {{ t('dashboard.balance') }}
            </p>
            <p class="mt-1 text-xl font-extrabold tabular-nums text-primary-700 dark:text-primary-300">
              ${{ formatBalance(balance) }}
            </p>
            <p class="mt-1 text-xs text-ink dark:text-ink-muted">{{ t('common.available') }}</p>
          </div>
        </div>
      </div>

      <!-- API Keys -->
      <div class="card relative min-w-0 overflow-hidden p-4">
        <span class="absolute left-4 top-0 h-1 w-10 rounded-b-sm bg-accent-400" aria-hidden="true" />
        <div class="flex items-start gap-3">
          <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-accent-300 bg-accent-100 shadow-pixel-sm dark:border-accent-700 dark:bg-accent-900/25">
            <Icon name="key" size="md" class="text-accent-700 dark:text-accent-300" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="text-xs font-bold uppercase tracking-wide text-ink dark:text-ink-muted">
              {{ t('dashboard.apiKeys') }}
            </p>
            <p class="mt-1 text-xl font-extrabold tabular-nums text-ink-strong dark:text-white">
              {{ stats?.total_api_keys || 0 }}
            </p>
            <p class="mt-1 text-xs font-semibold text-primary-700 dark:text-primary-300">
              {{ stats?.active_api_keys || 0 }} {{ t('common.active') }}
            </p>
          </div>
        </div>
      </div>

      <!-- Today Requests -->
      <div class="card relative min-w-0 overflow-hidden p-4">
        <span class="absolute left-4 top-0 h-1 w-10 rounded-b-sm bg-amber-400" aria-hidden="true" />
        <div class="flex items-start gap-3">
          <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-amber-300 bg-amber-100 shadow-pixel-sm dark:border-amber-700 dark:bg-amber-900/25">
            <Icon name="chart" size="md" class="text-amber-700 dark:text-amber-300" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="text-xs font-bold uppercase tracking-wide text-ink dark:text-ink-muted">
              {{ t('dashboard.todayRequests') }}
            </p>
            <p class="mt-1 text-xl font-extrabold tabular-nums text-ink-strong dark:text-white">
              {{ stats?.today_requests || 0 }}
            </p>
            <p class="mt-1 text-xs tabular-nums text-ink dark:text-ink-muted">
              {{ t('common.total') }}: {{ formatNumber(stats?.total_requests || 0) }}
            </p>
          </div>
        </div>
      </div>

      <!-- Today Cost -->
      <div class="card relative min-w-0 overflow-hidden p-4">
        <span class="absolute left-4 top-0 h-1 w-10 rounded-b-sm bg-accent-400" aria-hidden="true" />
        <div class="flex items-start gap-3">
          <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-accent-300 bg-accent-100 shadow-pixel-sm dark:border-accent-700 dark:bg-accent-900/25">
            <Icon name="dollar" size="md" class="text-accent-700 dark:text-accent-300" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="text-xs font-bold uppercase tracking-wide text-ink dark:text-ink-muted">
              {{ t('dashboard.todayCost') }}
            </p>
            <p class="mt-1 flex flex-wrap items-baseline gap-x-1 text-xl font-extrabold tabular-nums text-ink-strong dark:text-white">
              <span class="text-accent-700 dark:text-accent-300" :title="t('dashboard.actual')">
                ${{ formatCost(stats?.today_actual_cost || 0) }}
              </span>
              <span class="text-sm font-medium text-ink-muted dark:text-ink-muted" :title="t('dashboard.standard')">
                / ${{ formatCost(stats?.today_cost || 0) }}
              </span>
            </p>
            <p class="mt-1 text-xs tabular-nums">
              <span class="text-ink dark:text-ink-muted">{{ t('common.total') }}: </span>
              <span class="font-semibold text-accent-700 dark:text-accent-300" :title="t('dashboard.actual')">
                ${{ formatCost(stats?.total_actual_cost || 0) }}
              </span>
              <span class="text-ink-muted dark:text-ink-muted" :title="t('dashboard.standard')">
                / ${{ formatCost(stats?.total_cost || 0) }}
              </span>
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Row 2: Token Stats -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <!-- Today Tokens -->
      <div class="card relative min-w-0 overflow-hidden p-4">
        <span class="absolute left-4 top-0 h-1 w-10 rounded-b-sm bg-amber-400" aria-hidden="true" />
        <div class="flex items-start gap-3">
          <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-amber-300 bg-amber-100 shadow-pixel-sm dark:border-amber-700 dark:bg-amber-900/25">
            <Icon name="cube" size="md" class="text-amber-700 dark:text-amber-300" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="text-xs font-bold uppercase tracking-wide text-ink dark:text-ink-muted">
              {{ t('dashboard.todayTokens') }}
            </p>
            <p class="mt-1 text-xl font-extrabold tabular-nums text-ink-strong dark:text-white">
              {{ formatTokens(stats?.today_tokens || 0) }}
            </p>
            <p class="mt-1 text-xs leading-5 tabular-nums text-ink dark:text-ink-muted">
              {{ t('dashboard.input') }}: {{ formatTokens(stats?.today_input_tokens || 0) }} /
              {{ t('dashboard.output') }}: {{ formatTokens(stats?.today_output_tokens || 0) }} /
              {{ t('dashboard.cache') }}: {{ formatTokens((stats?.today_cache_creation_tokens || 0) + (stats?.today_cache_read_tokens || 0)) }}
            </p>
          </div>
        </div>
      </div>

      <!-- Total Tokens -->
      <div class="card relative min-w-0 overflow-hidden p-4">
        <span class="absolute left-4 top-0 h-1 w-10 rounded-b-sm bg-primary-400" aria-hidden="true" />
        <div class="flex items-start gap-3">
          <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-primary-300 bg-primary-100 shadow-pixel-sm dark:border-primary-700 dark:bg-primary-900/30">
            <Icon name="database" size="md" class="text-primary-700 dark:text-primary-300" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="text-xs font-bold uppercase tracking-wide text-ink dark:text-ink-muted">
              {{ t('dashboard.totalTokens') }}
            </p>
            <p class="mt-1 text-xl font-extrabold tabular-nums text-ink-strong dark:text-white">
              {{ formatTokens(stats?.total_tokens || 0) }}
            </p>
            <p class="mt-1 text-xs leading-5 tabular-nums text-ink dark:text-ink-muted">
              {{ t('dashboard.input') }}: {{ formatTokens(stats?.total_input_tokens || 0) }} /
              {{ t('dashboard.output') }}: {{ formatTokens(stats?.total_output_tokens || 0) }} /
              {{ t('dashboard.cache') }}: {{ formatTokens((stats?.total_cache_creation_tokens || 0) + (stats?.total_cache_read_tokens || 0)) }}
            </p>
          </div>
        </div>
      </div>

      <!-- Performance (RPM/TPM) -->
      <div class="card relative min-w-0 overflow-hidden p-4">
        <span class="absolute left-4 top-0 h-1 w-10 rounded-b-sm bg-accent-400" aria-hidden="true" />
        <div class="flex items-start gap-3">
          <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-accent-300 bg-accent-100 shadow-pixel-sm dark:border-accent-700 dark:bg-accent-900/25">
            <Icon name="bolt" size="md" class="text-accent-700 dark:text-accent-300" :stroke-width="2" />
          </div>
          <div class="min-w-0 flex-1">
            <p class="text-xs font-bold uppercase tracking-wide text-ink dark:text-ink-muted">
              {{ t('dashboard.performance') }}
            </p>
            <div class="mt-1 flex items-baseline gap-2">
              <p class="text-xl font-extrabold tabular-nums text-ink-strong dark:text-white">
                {{ formatTokens(stats?.rpm || 0) }}
              </p>
              <span class="text-xs font-semibold text-ink dark:text-ink-muted">RPM</span>
            </div>
            <div class="mt-1 flex items-baseline gap-2">
              <p class="text-sm font-bold tabular-nums text-accent-700 dark:text-accent-300">
                {{ formatTokens(stats?.tpm || 0) }}
              </p>
              <span class="text-xs font-semibold text-ink dark:text-ink-muted">TPM</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Avg Response Time -->
      <div class="card relative min-w-0 overflow-hidden p-4">
        <span class="absolute left-4 top-0 h-1 w-10 rounded-b-sm bg-primary-400" aria-hidden="true" />
        <div class="flex items-start gap-3">
          <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-primary-300 bg-primary-100 shadow-pixel-sm dark:border-primary-700 dark:bg-primary-900/30">
            <Icon name="clock" size="md" class="text-primary-700 dark:text-primary-300" :stroke-width="2" />
          </div>
          <div class="min-w-0">
            <p class="text-xs font-bold uppercase tracking-wide text-ink dark:text-ink-muted">
              {{ t('dashboard.avgResponse') }}
            </p>
            <p class="mt-1 text-xl font-extrabold tabular-nums text-ink-strong dark:text-white">
              {{ formatDuration(stats?.average_duration_ms || 0) }}
            </p>
            <p class="mt-1 text-xs text-ink dark:text-ink-muted">{{ t('dashboard.averageTime') }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Row 3: Per-platform breakdown -->
    <section v-if="!isSimple && platformCards.length > 0" class="card relative overflow-hidden p-4 sm:p-5">
      <span class="absolute left-4 top-0 h-1 w-12 rounded-b-sm bg-amber-400 sm:left-5" aria-hidden="true" />
      <div class="mb-4 flex flex-wrap items-center justify-between gap-2">
        <div class="flex items-center gap-2">
          <span class="h-2.5 w-2.5 rounded-sm bg-amber-400" aria-hidden="true" />
          <h3 class="text-base font-bold text-ink-strong dark:text-white">
            {{ t('dashboard.platformBreakdown') }}
          </h3>
        </div>
        <span class="badge badge-gray border-2 border-line dark:border-line-strong">
          {{ t('dashboard.platformCount', { count: sortedPlatforms.length }) }}
        </span>
      </div>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <article
          v-for="item in platformCards"
          :key="item.platform"
          :class="[
            'rounded-xl border-2 p-4 shadow-pixel-sm',
            item.isOther
              ? 'border-dashed border-line-strong bg-surface-muted dark:border-line-strong dark:bg-surface-muted/30'
              : 'border-line bg-white dark:border-line-strong dark:bg-canvas/30'
          ]"
        >
          <div class="flex items-start justify-between gap-3">
            <span class="min-w-0 break-words text-sm font-bold text-ink-strong dark:text-white">
              {{ item.isOther ? t('dashboard.platformOther') : platformLabel(item.platform) }}
            </span>
            <span class="flex-shrink-0 font-mono text-sm font-bold tabular-nums text-primary-700 dark:text-primary-300" :title="t('dashboard.actual')">
              ${{ formatCost(item.total_actual_cost) }}
            </span>
          </div>
          <dl class="mt-3 space-y-2 border-l-2 border-accent-300 pl-3 text-xs dark:border-accent-700">
            <div class="flex items-center justify-between gap-3">
              <dt class="text-ink dark:text-ink-muted">{{ t('dashboard.todayCost') }}</dt>
              <dd class="font-mono tabular-nums text-ink-strong dark:text-white">${{ formatCost(item.today_actual_cost) }}</dd>
            </div>
            <div class="flex items-center justify-between gap-3">
              <dt class="text-ink dark:text-ink-muted">{{ t('dashboard.requests') }}</dt>
              <dd class="font-mono tabular-nums text-ink dark:text-gray-200">
                {{ item.total_requests > 0 ? formatNumber(item.total_requests) : '-' }}
              </dd>
            </div>
            <div class="flex items-center justify-between gap-3">
              <dt class="text-ink dark:text-ink-muted">{{ t('dashboard.tokens') }}</dt>
              <dd class="font-mono tabular-nums text-ink dark:text-gray-200">
                {{ item.total_tokens > 0 ? formatTokens(item.total_tokens) : '-' }}
              </dd>
            </div>
          </dl>

          <!-- Quota 区：仅当 quota 配置存在、非 __other__ 且至少有一个窗口配了 limit 时显示 -->
          <div v-if="hasAnyLimit(item.quota) && !item.isOther" class="mt-4 space-y-2 border-t-2 border-line pt-3 dark:border-line">
            <p class="text-xs font-bold uppercase tracking-wide text-ink-muted dark:text-ink-muted">
              {{ t('dashboard.platformQuota.title') }}
            </p>
            <template v-for="w in (['daily', 'weekly', 'monthly'] as const)" :key="w">
              <div v-if="quotaVal(item.quota, `${w}_limit_usd`) != null" class="space-y-1">
                <!-- limit=0：完全禁用 -->
                <template v-if="(quotaVal(item.quota, `${w}_limit_usd`) as number) === 0">
                  <div class="flex items-center justify-between gap-3 text-xs">
                    <span class="text-ink dark:text-ink-muted">{{ t(`dashboard.platformQuota.${w}`) }}</span>
                    <span class="font-mono font-bold text-red-600 dark:text-red-400">{{ t('dashboard.platformQuota.disabled') }}</span>
                  </div>
                  <div class="h-2 w-full overflow-hidden rounded-sm border border-red-300 bg-red-100 dark:border-red-800 dark:bg-red-950/40">
                    <div class="h-full w-full bg-red-500" />
                  </div>
                </template>
                <!-- limit>0：正常用量进度条 -->
                <template v-else>
                  <div class="flex items-center justify-between gap-3 text-xs">
                    <span class="text-ink dark:text-ink-muted">{{ t(`dashboard.platformQuota.${w}`) }}</span>
                    <span class="font-mono tabular-nums text-ink dark:text-gray-200">
                      ${{ formatUsd((quotaVal(item.quota, `${w}_usage_usd`) as number) ?? 0) }} / ${{ formatUsd(quotaVal(item.quota, `${w}_limit_usd`) as number) }}
                    </span>
                  </div>
                  <div class="h-2 w-full overflow-hidden rounded-sm border border-line-strong bg-surface-muted dark:border-line-strong dark:bg-surface-muted">
                    <div
                      class="h-full transition-[width] duration-200 motion-reduce:transition-none"
                      :class="quotaBarClass(calcPercent((quotaVal(item.quota, `${w}_usage_usd`) as number) ?? 0, quotaVal(item.quota, `${w}_limit_usd`) as number))"
                      :style="{ width: calcPercent((quotaVal(item.quota, `${w}_usage_usd`) as number) ?? 0, quotaVal(item.quota, `${w}_limit_usd`) as number) + '%' }"
                    />
                  </div>
                  <p v-if="quotaVal(item.quota, `${w}_window_resets_at`)" class="text-xs leading-5 text-ink-muted dark:text-ink-muted">
                    {{ t('dashboard.platformQuota.resetsAt', { time: formatResetTime(quotaVal(item.quota, `${w}_window_resets_at`) as string) }) }}
                  </p>
                </template>
              </div>
            </template>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'
import type { PlatformQuotaItem } from '@/types'

interface FusedPlatformCard {
  platform: string
  total_actual_cost: number
  today_actual_cost: number
  total_requests: number
  total_tokens: number
  isOther?: boolean
  quota?: PlatformQuotaItem
}

const props = defineProps<{
  stats: UserStatsType
  balance: number
  isSimple: boolean
  platformQuotas?: PlatformQuotaItem[] | null
}>()
const { t } = useI18n()

const PLATFORM_LABELS: Record<string, string> = {
  anthropic: 'Claude',
  openai: 'OpenAI',
  gemini: 'Gemini',
  antigravity: 'Antigravity'
}

const platformLabel = (p: string) => PLATFORM_LABELS[p] ?? p

const sortedPlatforms = computed(() => {
  const list = props.stats?.by_platform ?? []
  return [...list].sort((a, b) => b.total_actual_cost - a.total_actual_cost)
})

// 处理"各平台之和 < 总值"的差值：后端按平台聚合时过滤了无法归属平台的行
// （group 与 account 都缺 platform）。这里把差值作为"其他"卡片显式展示，
// 避免 Row 1 总值与 Row 3 平台拆分加总对不上、用户困惑。
const OTHER_THRESHOLD = 0.0001
const platformCards = computed<FusedPlatformCard[]>(() => {
  // 建立 by_platform Map
  const byPlat = new Map<string, (typeof sortedPlatforms.value)[number]>()
  for (const item of props.stats?.by_platform ?? []) byPlat.set(item.platform, item)

  // 建立 quota Map
  const byQuota = new Map<string, PlatformQuotaItem>()
  for (const q of props.platformQuotas ?? []) byQuota.set(q.platform, q)

  // union 平台集合。后端 by_platform / quota 接口均不会返回 platform='__other__'，
  // 无需显式排除；__other__ 由下方差值补差逻辑单独追加。
  const platforms = new Set<string>([...byPlat.keys(), ...byQuota.keys()])

  const PLATFORM_ORDER = ['anthropic', 'openai', 'gemini', 'antigravity', 'grok']
  const cards: FusedPlatformCard[] = []

  for (const p of platforms) {
    const stat = byPlat.get(p)
    cards.push({
      platform: p,
      total_actual_cost: stat?.total_actual_cost ?? 0,
      today_actual_cost: stat?.today_actual_cost ?? 0,
      total_requests: stat?.total_requests ?? 0,
      total_tokens: stat?.total_tokens ?? 0,
      quota: byQuota.get(p),
    })
  }

  // 排序：按 PLATFORM_ORDER，未知平台按名称排序
  cards.sort((a, b) => {
    const ai = PLATFORM_ORDER.indexOf(a.platform)
    const bi = PLATFORM_ORDER.indexOf(b.platform)
    if (ai === -1 && bi === -1) return a.platform.localeCompare(b.platform)
    if (ai === -1) return 1
    if (bi === -1) return -1
    return ai - bi
  })

  // __other__ 补差逻辑：只对 by_platform 有 usage 数据的总和计算
  const total = props.stats?.total_actual_cost ?? 0
  const today = props.stats?.today_actual_cost ?? 0
  const sumTotal = cards.reduce((s, c) => s + c.total_actual_cost, 0)
  const sumToday = cards.reduce((s, c) => s + c.today_actual_cost, 0)
  const diffTotal = Math.max(0, total - sumTotal)
  const diffToday = Math.max(0, today - sumToday)

  if (diffTotal > OTHER_THRESHOLD || diffToday > OTHER_THRESHOLD) {
    cards.push({
      platform: '__other__',
      total_actual_cost: diffTotal,
      today_actual_cost: diffToday,
      total_requests: 0,
      total_tokens: 0,
      isOther: true,
    })
  }

  return cards
})

// Quota helpers

type QuotaWindow = 'daily' | 'weekly' | 'monthly'
type QuotaField = `${QuotaWindow}_limit_usd` | `${QuotaWindow}_usage_usd` | `${QuotaWindow}_window_resets_at`

function quotaVal(q: PlatformQuotaItem | undefined, key: QuotaField): PlatformQuotaItem[QuotaField] {
  return q?.[key]
}

function hasAnyLimit(q: PlatformQuotaItem | undefined): boolean {
  if (!q) return false
  return q.daily_limit_usd != null || q.weekly_limit_usd != null || q.monthly_limit_usd != null
}

function calcPercent(usage: number, limit: number): number {
  if (!limit || limit <= 0) return 0
  return Math.min(100, Math.max(0, Math.round((usage / limit) * 100)))
}

function quotaBarClass(p: number): string {
  if (p >= 95) return 'bg-red-500'
  if (p >= 75) return 'bg-amber-500'
  return 'bg-primary-500'
}

// 与 formatBalance 一致使用 Intl.NumberFormat 做半偶舍入，避免 toFixed 在不同 JS 引擎
// 下偶发截断而非四舍五入（与后端展示精度不一致）。
const usdFormatter = new Intl.NumberFormat('en-US', {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
})
function formatUsd(n: number): string {
  if (!Number.isFinite(n)) return '0.00'
  return usdFormatter.format(n)
}

function formatResetTime(iso: string | null | undefined): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString(undefined, {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

const formatBalance = (b: number) =>
  new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(b)

const formatNumber = (n: number) => n.toLocaleString()
const formatCost = (c: number) => c.toFixed(4)
const formatTokens = (t: number) => {
  if (t >= 1_000_000) return `${(t / 1_000_000).toFixed(1)}M`
  if (t >= 1000) return `${(t / 1000).toFixed(1)}K`
  return t.toString()
}
const formatDuration = (ms: number) => ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${ms.toFixed(0)}ms`
</script>
