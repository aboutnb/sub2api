<template>
  <div class="model-picker-scene" :class="clientId">
    <header><strong>{{ title }}</strong><span aria-hidden="true">×</span></header>
    <div class="picker-body">
      <p>{{ location }}</p>
      <div class="search-example">{{ t('helpCenter.scenes.searchModels') }}</div>
      <div v-if="clientId === 'cherry-studio'" class="model-filters"><span>{{ t('helpCenter.scenes.allModels') }}</span><small>{{ t('helpCenter.scenes.capabilitiesNotAssumed') }}</small></div>
      <div class="provider-heading"><ClientLogo :client="clientId" /><strong>{{ t('helpCenter.walkthrough.site') }}</strong><span aria-hidden="true">⌄</span></div>
      <div class="model-choice"><code>{{ model }}</code><span aria-hidden="true">✓</span></div>
      <small class="selection-hint">{{ t('helpCenter.scenes.selectionExample') }}</small>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import ClientLogo from '@/components/common/ClientLogo.vue'
const props = defineProps<{ clientId: string; model: string; location: string }>()
const { t } = useI18n()
const title = computed(() => props.clientId === 'zcode' ? 'Choose model' : props.clientId === 'cherry-studio' ? t('helpCenter.scenes.providerModels') : t('helpCenter.walkthrough.selectModel'))
</script>
<style scoped>
.model-picker-scene { border: 1px solid var(--scene-line); border-radius: 10px; background: var(--scene-bg); color: var(--scene-fg); min-width: 0; overflow: hidden; }
header { display: flex; justify-content: space-between; padding: 1rem; border-bottom: 1px solid var(--scene-line); font-size: .8rem; }
.picker-body { padding: 1rem; }
p, small { font-size: .7rem; color: var(--scene-muted); line-height: 1.8; }
p { margin: 0 0 .75rem; }
.search-example { border: 1px solid var(--scene-line); padding: .65rem; border-radius: 6px; font-size: .7rem; color: var(--scene-muted); }
.model-filters { display: flex; align-items: center; flex-wrap: wrap; gap: .5rem; margin: .75rem 0; }
.model-filters > span { padding: .2rem .6rem; border-radius: 20px; background: var(--scene-accent-soft); font-size: .68rem; }
.provider-heading { display: flex; align-items: center; gap: .5rem; margin: 1rem 0 .5rem; font-size: .75rem; }
.provider-heading > span { margin-left: auto; }
.model-choice { display: flex; align-items: center; gap: .5rem; border: 1px solid var(--client-accent); padding: .65rem .75rem; border-radius: 6px; background: var(--scene-panel); }
code { flex: 1; min-width: 0; overflow-wrap: anywhere; font-size: .75rem; }
.model-choice > span { color: var(--client-accent-ink, var(--client-accent)); }
.selection-hint { display: block; margin-top: .8rem; }
.zcode { border-radius: 6px; }
.zcode .model-choice { border-left-width: 3px; }
</style>
