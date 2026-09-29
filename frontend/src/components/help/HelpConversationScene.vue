<template>
  <div class="conversation-scene" :class="[clientId, { translation: isTranslation }]">
    <header><ClientLogo :client="clientId" /><strong>{{ clientId === 'cherry-studio' ? 'Default Assistant' : clientName }}</strong><span>{{ t('helpCenter.walkthrough.exampleOnly') }}</span></header>
    <div v-if="isTranslation" class="translation-page"><div class="translation-toolbar">{{ t('helpCenter.scenes.translationService') }} · {{ provider }}<code>{{ model }}</code></div><article><small>{{ t('helpCenter.scenes.sourceText') }}</small><p>Hello, world.</p><div class="translated-paragraph"><small>{{ t('helpCenter.walkthrough.exampleReply') }}</small><p>{{ t('helpCenter.walkthrough.translation') }}</p></div></article></div>
    <div v-else class="workspace-conversation">
      <aside v-if="['workbuddy', 'zcode'].includes(clientId)"><ClientLogo :client="clientId" /><span>＋</span><span>{{ clientId === 'zcode' ? 'Workspace' : t('helpCenter.scenes.newTask') }}</span></aside>
      <main><div v-if="clientId === 'cherry-studio'" class="assistant-banner">{{ t('helpCenter.scenes.assistantReady') }}<code>{{ model }}</code></div><div class="user-bubble">{{ t('helpCenter.walkthrough.prompt') }}</div><div class="assistant-bubble"><ClientLogo :client="clientId" /><div><small>{{ clientId === 'trae' ? 'SOLO' : clientName }} · {{ t('helpCenter.walkthrough.exampleReply') }}</small><p>OK</p></div></div><div class="composer"><span>{{ t('helpCenter.walkthrough.inClient') }}</span><div><span aria-hidden="true">＋</span><code>{{ model }}</code><span aria-hidden="true">⌄</span><span class="send-icon" aria-hidden="true">↑</span></div></div></main>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import ClientLogo from '@/components/common/ClientLogo.vue'
const props = defineProps<{ clientId: string; clientName: string; model: string; provider: string }>()
const { t } = useI18n()
const isTranslation = computed(() => ['read-frog', 'kiss-translator', 'immersive-translate'].includes(props.clientId))
</script>
<style scoped>
.conversation-scene { border: 1px solid var(--scene-line); border-radius: 12px; background: var(--scene-bg); color: var(--scene-fg); font-size: .8rem; overflow: hidden; }
header { display: flex; align-items: center; gap: .6rem; padding: .85rem 1rem; border-bottom: 1px solid var(--scene-line); flex-wrap: wrap; }
header > span { margin-left: auto; font-size: .65rem; color: var(--scene-muted); }
.workspace-conversation { display: flex; min-height: 260px; }
aside { display: flex; flex-direction: column; align-items: center; gap: 1.25rem; width: 84px; flex-shrink: 0; padding: 1rem .5rem; background: var(--scene-panel); border-right: 1px solid var(--scene-line); font-size: .65rem; }
main { flex: 1; min-width: 0; padding: 1.25rem; }
.assistant-banner { text-align: center; font-size: .75rem; color: var(--scene-muted); margin: .5rem 0 1.5rem; }
.assistant-banner code { display: block; margin-top: .5rem; }
.user-bubble { width: fit-content; max-width: 90%; margin-left: auto; padding: .6rem .9rem; background: var(--scene-panel); border-radius: 10px 10px 2px 10px; }
.assistant-bubble { display: flex; gap: .7rem; margin: 1.25rem 0; }
small { font-size: .65rem; color: var(--scene-muted); }
.assistant-bubble p { margin-top: .4rem; }
.composer { border: 1px solid var(--scene-line); border-radius: 10px; padding: .75rem; color: var(--scene-muted); font-size: .72rem; }
.composer > div { display: flex; align-items: center; gap: .5rem; margin-top: 1.5rem; }
.composer code { margin-left: auto; }
code { overflow-wrap: anywhere; min-width: 0; font-size: .7rem; }
.send-icon { display: grid; place-items: center; background: var(--scene-solid); color: var(--scene-solid-fg); width: 23px; height: 23px; flex-shrink: 0; border-radius: 5px; }
.trae main { background: var(--scene-panel); }
.trae .composer { background: var(--scene-bg); border-color: var(--scene-accent-fg); }
.translation-toolbar { background: var(--scene-panel); border-bottom: 1px solid var(--scene-line); padding: .75rem 1rem; font-size: .72rem; }
.translation-toolbar code { display: block; margin-top: .4rem; }
article { margin: 1.25rem; padding: 1rem; border: 1px solid var(--scene-line); border-radius: 8px; }
article p { margin: .5rem 0 1rem; font-size: 1rem; }
.translated-paragraph { border-left: 3px solid var(--client-accent); padding-left: 1rem; }
@media (max-width: 540px) { aside { display: none; } main { padding: 1rem; } }
@container help-scenes (max-width: 540px) { aside { display: none; } main { padding: 1rem; } }
</style>
