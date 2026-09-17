<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useClipboard } from '@/composables/useClipboard'
import Icon from '@/components/icons/Icon.vue'
import type { CustomEndpoint } from '@/types'

const props = defineProps<{
  apiBaseUrl: string
  customEndpoints: CustomEndpoint[]
}>()

const { t } = useI18n()
const { copyToClipboard } = useClipboard()
const copiedEndpoint = ref<string | null>(null)

let copiedResetTimer: number | undefined

const allEndpoints = computed(() => {
  const items: Array<{ name: string; endpoint: string; description: string; isDefault: boolean }> = []
  if (props.apiBaseUrl) {
    items.push({
      name: t('keys.endpoints.title'),
      endpoint: props.apiBaseUrl,
      description: '',
      isDefault: true,
    })
  }
  for (const ep of props.customEndpoints) {
    items.push({ ...ep, isDefault: false })
  }
  return items
})

async function copy(url: string) {
  const success = await copyToClipboard(url, t('keys.endpoints.copied'))
  if (!success) return

  copiedEndpoint.value = url
  if (copiedResetTimer !== undefined) {
    window.clearTimeout(copiedResetTimer)
  }
  copiedResetTimer = window.setTimeout(() => {
    if (copiedEndpoint.value === url) {
      copiedEndpoint.value = null
    }
  }, 1800)
}

function tooltipHint(endpoint: string): string {
  return copiedEndpoint.value === endpoint
    ? t('keys.endpoints.copiedHint')
    : t('keys.endpoints.clickToCopy')
}

function speedTestUrl(endpoint: string): string {
  return `https://www.tcptest.cn/http/${encodeURIComponent(endpoint)}`
}

onBeforeUnmount(() => {
  if (copiedResetTimer !== undefined) {
    window.clearTimeout(copiedResetTimer)
  }
})
</script>

<template>
  <div v-if="allEndpoints.length > 0" class="endpoint-list">
    <div
      v-for="(item, index) in allEndpoints"
      :key="index"
      class="endpoint-item"
    >
      <div class="flex min-w-0 items-center gap-1.5">
        <span class="font-medium text-ink [overflow-wrap:anywhere] dark:text-ink-muted">{{ item.name }}</span>
        <span
          v-if="item.isDefault"
          class="rounded bg-primary-50 px-1 py-px text-[10px] font-medium leading-tight text-primary-600 dark:bg-primary-900/30 dark:text-primary-400"
        >{{ t('keys.endpoints.default') }}</span>
      </div>
      <div class="flex min-w-0 items-center gap-1">
        <button
          type="button"
          role="button"
          class="endpoint-copy"
          :class="copiedEndpoint === item.endpoint ? 'text-success' : 'text-brand'"
          :aria-label="tooltipHint(item.endpoint)"
          @click="copy(item.endpoint)"
        >
          <code class="min-w-0 flex-1 truncate text-left font-mono leading-5" :title="item.endpoint">{{ item.endpoint }}</code>
          <Icon :name="copiedEndpoint === item.endpoint ? 'check' : 'copy'" size="sm" class="shrink-0" aria-hidden="true" />
          <span class="sr-only">{{ tooltipHint(item.endpoint) }}</span>
        </button>

        <a
          :href="speedTestUrl(item.endpoint)"
          target="_blank"
          rel="noopener noreferrer"
          class="endpoint-speed flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-ink-muted transition-colors hover:bg-surface-muted hover:text-brand"
          :title="t('keys.endpoints.speedTest')"
          :aria-label="`${t('keys.endpoints.speedTest')}: ${item.name}`"
        >
          <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </a>
      </div>
      <p v-if="item.description" class="endpoint-description text-xs leading-5 text-ink-muted [overflow-wrap:anywhere]">{{ item.description }}</p>
    </div>
  </div>
</template>

<style scoped>
.endpoint-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 20rem), 1fr));
  gap: 0.5rem;
  min-width: 0;
  align-items: start;
}
.endpoint-item {
  @apply grid min-w-0 items-center gap-x-2 rounded-lg border border-line bg-surface px-2.5 py-1 text-xs;
  grid-template-columns: minmax(0, auto) minmax(0, 1fr);
}
.endpoint-copy {
  @apply flex min-h-8 min-w-0 flex-1 items-center gap-1.5 rounded-md px-1 py-0.5 text-xs leading-5 transition-colors hover:bg-brand-soft;
}
.endpoint-description {
  grid-column: 1 / -1;
}
@media (max-width: 479px) {
  .endpoint-item {
    grid-template-columns: minmax(0, 1fr);
  }
}
@media (pointer: coarse) {
  .endpoint-copy,
  .endpoint-speed {
    min-height: 2.75rem;
  }
  .endpoint-speed {
    min-width: 2.75rem;
  }
}
</style>
