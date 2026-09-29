<template>
  <section class="settings-preview" :style="{ '--client-accent': accent }" :aria-label="t('helpCenter.settingsPreview')">
    <header class="preview-heading"><ClientLogo :client="clientId" /><strong>{{ clientName }}</strong><span>{{ t('helpCenter.settingsPreview') }}</span></header>
    <p class="preview-caption">{{ t('helpCenter.walkthrough.caption') }}</p>
    <component :is="clientGuideComponents[clientId]" v-if="clientGuideComponents[clientId]" :client-id="clientId" :client-name="clientName" :fields="fields" />
    <p v-else role="status">{{ t('helpCenter.unsupported') }}</p>
    <p class="verification-note">{{ t('helpCenter.walkthrough.notTested') }} <RouterLink to="/usage">{{ t('helpCenter.viewUsage') }} →</RouterLink></p>
  </section>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import ClientLogo from '@/components/common/ClientLogo.vue'
import { clientWalkthroughs } from '@/utils/helpClientWalkthrough'
import { clientGuideComponents } from './guides/guideRegistry'
import './guides/guideScreens.css'
const props = defineProps<{ clientId: string; clientName: string; fields: { path: string; content: string }[] }>()
const { t } = useI18n()
const accent = computed(() => clientWalkthroughs[props.clientId]?.accent || '#187767')
</script>
<style scoped>
.settings-preview { margin: 1.25rem 0; min-width: 0; }
.preview-heading { display: flex; align-items: center; gap: .65rem; flex-wrap: wrap; }
.preview-heading > span { margin-left: auto; font-size: .75rem; color: var(--av-color-text-muted); }
.preview-caption, .verification-note { font-size: .8rem; line-height: 1.85; margin: .8rem 0; }
.verification-note a { color: var(--client-accent-ink, var(--client-accent)); text-decoration: underline; }
</style>
