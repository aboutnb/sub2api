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
      <div class="flex min-w-0 flex-wrap items-center gap-1.5">
      <span class="font-medium text-ink dark:text-ink-muted">{{ item.name }}</span>
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
          <code class="min-w-0 flex-1 whitespace-normal text-left font-mono leading-5 [overflow-wrap:anywhere]">{{ item.endpoint }}</code>
          <Icon :name="copiedEndpoint === item.endpoint ? 'check' : 'copy'" size="sm" class="shrink-0" aria-hidden="true" />
          <span class="sr-only">{{ tooltipHint(item.endpoint) }}</span>
        </button>

        <a
          :href="speedTestUrl(item.endpoint)"
          target="_blank"
          rel="noopener noreferrer"
          class="btn-icon btn-sm rounded-lg text-ink-muted transition-colors hover:bg-surface-muted hover:text-brand"
          :title="t('keys.endpoints.speedTest')"
          :aria-label="`${t('keys.endpoints.speedTest')}: ${item.name}`"
        >
          <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </a>
      </div>
      <p v-if="item.description" class="text-xs leading-5 text-ink-muted [overflow-wrap:anywhere]">{{ item.description }}</p>
    </div>
  </div>
</template>

<style scoped>
.endpoint-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 20rem), 1fr));
  gap: 0.5rem;
  min-width: 0;
}
.endpoint-item {
  @apply min-w-0 rounded-xl border border-line bg-surface px-3 py-2 text-xs;
}
.endpoint-copy {
  @apply flex min-h-9 min-w-0 flex-1 items-center gap-2 rounded-lg px-1 py-1 text-xs transition-colors hover:bg-brand-soft;
}
</style>
