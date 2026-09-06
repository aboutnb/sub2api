<template>
  <div
    class="stat-card !min-h-[6.5rem] !p-4"
    :title="title || undefined"
  >
    <div
      v-if="state"
      class="mt-1 h-2 w-2 shrink-0 rounded-full"
      :class="dotClass"
      aria-hidden="true"
    ></div>
    <div class="min-w-0 flex-1">
      <span class="stat-label text-[10px] font-bold uppercase tracking-wider text-ink-muted">{{ label }}</span>
      <strong
        class="stat-value mt-1 block overflow-visible text-xl tabular-nums leading-tight !text-clip !whitespace-normal"
        :class="stateClass"
      >{{ value }}</strong>
      <div
        v-if="detailParts.length > 1"
        class="mt-1.5 flex flex-wrap gap-x-2 gap-y-0.5 text-[11px] leading-snug text-ink-muted dark:text-ink-muted"
      >
        <span
          v-for="(part, index) in detailParts"
          :key="`${index}:${part}`"
          class="whitespace-nowrap tabular-nums"
        >{{ part }}</span>
      </div>
      <small
        v-else-if="detail"
        class="mt-1.5 block text-[11px] leading-snug text-ink-muted dark:text-ink-muted"
      >{{ detail }}</small>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { HealthState } from '@/api/channelMonitorV2'

const props = defineProps<{
  label: string
  value: string
  detail: string
  state?: HealthState
  /** Exact numeric tooltip (e.g. uncompacted RPM/TPM). */
  title?: string
}>()

/** Split "AVG 475ms · P90 800ms" into chips so nothing is ellipsized. */
const detailParts = computed(() => {
  const raw = (props.detail || '').trim()
  if (!raw || raw === '-') return []
  return raw
    .split(/\s*[·|]\s*/)
    .map((part) => part.trim())
    .filter(Boolean)
})

const stateClass = computed(() => {
  if (!props.state) return 'text-ink-strong dark:text-white'
  if (props.state === 'healthy') return 'text-emerald-600 dark:text-emerald-400'
  if (props.state === 'warning') return 'text-amber-600 dark:text-amber-400'
  if (props.state === 'critical') return 'text-red-600 dark:text-red-400'
  return 'text-ink-muted dark:text-ink-muted'
})

const dotClass = computed(() => {
  if (props.state === 'healthy') return 'bg-emerald-500'
  if (props.state === 'warning') return 'bg-amber-500'
  if (props.state === 'critical') return 'bg-red-500'
  return 'bg-line-strong dark:bg-line-strong'
})
</script>
