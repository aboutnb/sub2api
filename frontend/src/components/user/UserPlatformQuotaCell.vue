<template>
  <span v-if="props.quotas === undefined" class="text-xs text-ink-muted dark:text-ink-muted">…</span>
  <span v-else-if="configured.length === 0" class="text-xs text-ink-muted dark:text-ink-muted">
    {{ t('admin.users.platformQuota.cellNotConfigured') }}
  </span>
  <div v-else class="space-y-0.5 text-xs">
    <div
      v-for="row in configured"
      :key="row.platform"
      class="flex items-center gap-2 whitespace-nowrap"
    >
      <span class="w-20 shrink-0 font-mono text-ink dark:text-ink-muted">{{ row.platform }}</span>
      <span class="text-ink-muted dark:text-ink-muted">
        {{ t('admin.users.platformQuota.windowDaily') }}
        <span class="text-ink-strong dark:text-white">{{ fmtUsd(row.daily_usage_usd) }}/{{ fmtLimit(row.daily_limit_usd) }}</span>
      </span>
      <span class="text-ink-muted dark:text-ink-muted">
        {{ t('admin.users.platformQuota.windowWeekly') }}
        <span class="text-ink-strong dark:text-white">{{ fmtUsd(row.weekly_usage_usd) }}/{{ fmtLimit(row.weekly_limit_usd) }}</span>
      </span>
      <span class="text-ink-muted dark:text-ink-muted">
        {{ t('admin.users.platformQuota.windowMonthly') }}
        <span class="text-ink-strong dark:text-white">{{ fmtUsd(row.monthly_usage_usd) }}/{{ fmtLimit(row.monthly_limit_usd) }}</span>
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PlatformQuotaItem, PlatformQuotaPlatform } from '@/api/admin/users'

const props = defineProps<{ quotas?: PlatformQuotaItem[] }>()
const { t } = useI18n()

const PLATFORM_ORDER: PlatformQuotaPlatform[] = ['anthropic', 'openai', 'gemini', 'antigravity', 'grok']

// 仅展示「至少一档限额非空」的平台（配额列，非用量列）
const configured = computed(() => {
  if (!props.quotas) return []
  return props.quotas
    .filter(
      (q) =>
        q.daily_limit_usd != null ||
        q.weekly_limit_usd != null ||
        q.monthly_limit_usd != null
    )
    .slice()
    .sort((a, b) => PLATFORM_ORDER.indexOf(a.platform) - PLATFORM_ORDER.indexOf(b.platform))
})

// 去尾零、最多 2 位小数：100→"100"，90.5→"90.5"，0.42→"0.42"
function fmtUsd(n: number): string {
  if (n == null || Number.isNaN(n)) return '0'
  return String(Math.round(n * 100) / 100)
}
function fmtLimit(n: number | null): string {
  return n == null ? '—' : fmtUsd(n)
}
</script>
