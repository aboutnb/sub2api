<template>
  <figure class="opencode-example">
    <figcaption><strong>OpenCode · {{ t('helpCenter.scenes.terminalTitle') }}</strong><span>{{ t('helpCenter.walkthrough.exampleOnly') }}</span></figcaption>
    <p>{{ t('helpCenter.scenes.opencodeIntro') }}</p>
    <div class="opencode-window">
      <header><span class="lights" aria-hidden="true"><i /><i /><i /></span><strong>opencode</strong><span>{{ t('helpCenter.walkthrough.illustration') }}</span></header>
      <div class="session-title"><strong># Welcome to OpenCode</strong><small>{{ t('helpCenter.scenes.noLiveMetrics') }}</small></div>
      <div class="terminal-output"><p>{{ t('helpCenter.scenes.opencodeWelcome') }}</p><h5>## {{ t('helpCenter.scenes.capabilities') }}</h5><ul class="capability-list"><li><strong>Read &amp; edit files</strong><span>Navigate and modify the project</span></li><li><strong>Run commands</strong><span>Execute commands and review output</span></li><li><strong>Answer questions</strong><span>Explain and debug code</span></li></ul><h5>## {{ t('helpCenter.scenes.nextSteps') }}</h5><p><code>/models</code> · {{ t('helpCenter.scenes.checkModel') }}</p><p><code>/help</code> · {{ t('helpCenter.scenes.commandHelp') }}</p><div class="model-dialog"><strong>{{ t('helpCenter.walkthrough.selectModel') }}</strong><div class="model-search">{{ t('helpCenter.scenes.searchModels') }}</div><small>{{ provider }}</small><div class="model-result"><span aria-hidden="true">›</span><code>{{ displayModel }}</code><span aria-hidden="true">✓</span></div></div><div class="test-prompt">› {{ t('helpCenter.walkthrough.prompt') }}</div><p class="test-reply">OK <small>· {{ t('helpCenter.walkthrough.exampleReply') }}</small></p></div>
      <div class="terminal-input" aria-hidden="true">&gt; <i /></div>
      <div class="terminal-status"><span>enter send</span><span>{{ provider }} · <code>{{ displayModel }}</code></span></div>
      <footer><span>{{ os === 'windows' ? 'C:\\projects\\my-app' : '~/projects/my-app' }}</span><span>BUILD</span></footer>
    </div>
    <p class="terminal-disclaimer">{{ t('helpCenter.walkthrough.notTested') }}</p>
  </figure>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
const props = defineProps<{ model: string; os: string; provider: string }>()
const { t } = useI18n()
const displayModel = computed(() => props.model || 'YOUR_MODEL_ID')
</script>
<style scoped>
.opencode-example { margin: 1.5rem 0; min-width: 0; }
figcaption { display: flex; justify-content: space-between; gap: .75rem; flex-wrap: wrap; font-size: .85rem; }
figcaption span { font-size: .7rem; }
.opencode-example > p { margin: .75rem 0; font-size: .8rem; line-height: 1.8; }
.opencode-window { max-width: 44rem; margin: 1rem auto; border: 1px solid #383843; border-radius: 12px; background: #11121b; color: #d5d7e2; overflow: hidden; font: 12px/1.8 'SFMono-Regular', Consolas, 'Liberation Mono', monospace; }
header { display: flex; justify-content: space-between; align-items: center; gap: 1rem; background: #20212c; border-bottom: 1px solid #383843; padding: .8rem 1rem; }
header > span:last-child { font-size: 10px; color: #b3b5c5; }
.lights { display: flex; gap: 7px; }
.lights i { width: 10px; height: 10px; border-radius: 50%; background: #ff6058; }
.lights i:nth-child(2) { background: #febd2e; }
.lights i:nth-child(3) { background: #27c940; }
.session-title { background: #191a25; padding: 1rem 1.25rem; display: flex; gap: .5rem; justify-content: space-between; flex-wrap: wrap; }
small { color: #a0a5b7; font-size: 10px; }
.terminal-output { padding: 1.25rem; }
.terminal-output p { margin: .65rem 0; }
.capability-list { display: grid; gap: .35rem; margin: .7rem 0 1rem; padding: 0; list-style: none; color: #bac2de; }
.capability-list li { display: flex; gap: .45rem; flex-wrap: wrap; }
.capability-list li::before { content: '-'; color: #a6e3a1; }
.capability-list strong { color: #cdd6f4; }
h5 { color: #a9d69f; margin-top: 1rem; font-weight: 600; }
code { overflow-wrap: anywhere; white-space: pre-wrap; }
.model-dialog { margin: 1.25rem 0; padding: 1rem; border: 1px solid #4c5062; border-radius: 5px; background: #20212c; }
.model-search { padding: .4rem .6rem; margin: .7rem 0; color: #a0a5b7; border: 1px solid #4c5062; }
.model-result { display: flex; align-items: center; gap: .6rem; padding: .5rem; margin-top: .3rem; background: #33364a; }
.model-result code { flex: 1; min-width: 0; }
.test-prompt { border-top: 1px solid #383843; padding-top: .9rem; }
.test-reply { color: #a9d69f; }
.terminal-input { padding: .8rem 1.25rem; border-top: 1px solid #383843; background: #191a25; }
.terminal-input i { display: inline-block; width: 7px; height: 13px; vertical-align: middle; background: #edaf85; }
.terminal-status, footer { display: flex; gap: .8rem; justify-content: space-between; flex-wrap: wrap; padding: .5rem 1.25rem; background: #191a25; border-top: 1px solid #383843; font-size: 10px; }
.terminal-status > span { min-width: 0; overflow-wrap: anywhere; }
footer { background: #20212c; color: #b3b5c5; }
@media (max-width: 540px) { .terminal-output { padding: 1rem; } header { flex-wrap: wrap; } .session-title { flex-direction: column; } }
</style>
