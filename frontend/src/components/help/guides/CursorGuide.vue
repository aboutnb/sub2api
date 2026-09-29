<template>
  <div class="client-specific-guide cursor-guide" data-guide="cursor">
    <GuideStep :number="1" id="network" :title="t('helpCenter.specific.cursorNetwork')" :description="t('helpCenter.specific.cursorNetworkHint')">
      <div class="guide-screen">
        <div class="screen-top"><strong>Network</strong></div>
        <div class="screen-body">
          <div class="cursor-setting">
            <div>
              <strong>HTTP Compatibility Mode</strong>
              <p>HTTP/2 is recommended for low-latency streaming. In some corporate proxy and VPN environments, the compatibility mode may need to be lowered.</p>
            </div>
            <span class="cursor-select">HTTP/1.1 <b>⌄</b></span>
          </div>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="2" id="api-keys" :title="t('helpCenter.specific.cursorKeys')" :description="t('helpCenter.specific.cursorKeysHint')">
      <div class="guide-screen">
        <div class="screen-top"><strong>⌄ API Keys</strong></div>
        <div class="screen-body">
          <div class="cursor-card">
            <div class="cursor-toggle-row">
              <div>
                <strong>OpenAI API Key</strong>
                <p>You can put in your OpenAI key to use OpenAI models at cost.</p>
              </div>
              <span class="cursor-toggle is-on"><i /></span>
            </div>
            <HelpSettingField :field="field('API Key')" label="OpenAI API Key" />
          </div>
          <div class="cursor-card">
            <div class="cursor-toggle-row">
              <div>
                <strong>Override OpenAI Base URL</strong>
                <p>Change the base URL for OpenAI API requests.</p>
              </div>
              <span class="cursor-toggle is-on"><i /></span>
            </div>
            <HelpSettingField :field="field('Override OpenAI Base URL')" :hint="t('helpCenter.scenes.endpointHint')" />
          </div>
          <div class="screen-hint">{{ t('helpCenter.walkthrough.notes.cursor') }}</div>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="3" id="chat" :title="t('helpCenter.specific.cursorChat')" :description="t('helpCenter.specific.cursorModelHint')">
      <HelpSettingField v-if="modelField" :field="modelField" :label="t('helpCenter.model')" />
      <HelpConversationScene client-id="cursor" client-name="Cursor" :model="model" :provider="provider" />
    </GuideStep>
  </div>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import GuideStep from './GuideStep.vue'
import HelpSettingField from '../HelpSettingField.vue'
import HelpConversationScene from '../HelpConversationScene.vue'
import { useGuideFields, type ClientGuideProps } from './guideFields'
const props = defineProps<ClientGuideProps>(); const { t } = useI18n(); const { field, model, modelField, provider } = useGuideFields(props)
</script>
<style scoped>
.cursor-card { margin: 0 0 .75rem; padding: .2rem .9rem .4rem; border-radius: 12px; background: var(--scene-panel); }
.cursor-toggle-row { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; margin: .8rem 0 .2rem; }
.cursor-toggle-row p, .cursor-setting p { margin-top: .3rem; color: var(--scene-muted); font-size: .72rem; line-height: 1.6; }
.cursor-toggle { display: inline-flex; width: 40px; height: 20px; flex-shrink: 0; border-radius: 999px; background: var(--scene-track); padding: 2px; }
.cursor-toggle.is-on { justify-content: flex-end; background: #16a34a; }
.cursor-toggle i { width: 16px; height: 16px; border-radius: 50%; background: #fff; box-shadow: var(--scene-shadow); }
.cursor-setting { display: flex; align-items: center; justify-content: space-between; gap: 1rem; padding: 1rem; border-radius: 12px; background: var(--scene-panel); }
.cursor-select { display: flex; align-items: center; justify-content: space-between; min-width: 132px; height: 40px; padding: 0 .75rem; border: 2px solid var(--scene-accent-fg); border-radius: 8px; background: var(--scene-accent-soft); color: var(--scene-accent-fg); font-weight: 650; }
.cursor-select b { font-size: .9rem; }
.cursor-guide :deep(.field-input) { border: 2px solid var(--scene-accent-fg); background: var(--scene-accent-soft); border-radius: 10px; }
.cursor-guide :deep(.field-input code) { color: var(--scene-accent-fg); }
@media (max-width: 540px) { .cursor-setting, .cursor-toggle-row { flex-direction: column; } .cursor-select { width: 100%; } }
</style>
