<template>
  <div class="client-specific-guide tavernai-guide" data-guide="tavernai">
    <GuideStep :number="1" id="open" :title="t('helpCenter.specific.tavernOpen')" :description="t('helpCenter.specific.tavernOpenHint')">
      <div class="tai-window" role="img" :aria-label="t('helpCenter.tavernUi.title')">
        <div class="tai-top"><span class="plug" aria-hidden="true">⌁</span><strong>{{ t('helpCenter.tavernUi.title') }}</strong></div>
        <p class="screen-hint">{{ t('helpCenter.tavernUi.notOpenAI') }}</p>
      </div>
    </GuideStep>
    <GuideStep :number="2" id="custom" :title="t('helpCenter.specific.tavernCustom')" :description="t('helpCenter.specific.tavernCustomHint')">
      <div class="tai-window">
        <div class="tai-top"><strong>{{ t('helpCenter.tavernUi.title') }}</strong></div>
        <div class="tai-body">
          <HelpSettingField :field="field('Base Provider')" :label="t('helpCenter.tavernUi.baseProvider')" choice />
          <HelpSettingField :field="field('Model Endpoint')" :label="t('helpCenter.tavernUi.modelEndpoint')" choice />
          <HelpSettingField :field="field('API Address')" :label="t('helpCenter.tavernUi.apiAddress')" :hint="t('helpCenter.scenes.endpointHint')" />
          <HelpSettingField :field="field('Use direct API address')" :label="t('helpCenter.tavernUi.directApi')" toggle />
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="3" id="connect" :title="t('helpCenter.specific.tavernConnect')" :description="t('helpCenter.specific.tavernConnectHint')">
      <div class="tai-window">
        <div class="tai-body">
          <div class="load-row"><span>{{ t('helpCenter.tavernUi.providerModels') }}</span><span class="load">{{ t('helpCenter.tavernUi.load') }}</span></div>
          <p class="screen-hint">{{ t('helpCenter.tavernUi.loadEmpty') }}</p>
          <HelpSettingField v-if="modelField" :field="modelField" :label="t('helpCenter.tavernUi.modelId')" />
          <HelpSettingField :field="field('API key')" :label="t('helpCenter.tavernUi.apiKey')" />
          <label class="tai-check"><span class="box" aria-hidden="true" /><span>{{ t('helpCenter.tavernUi.skip') }}</span></label>
          <p class="screen-hint">{{ t('helpCenter.tavernUi.skipHint') }}</p>
          <div class="tai-actions"><span class="tai-button">{{ t('helpCenter.tavernUi.connect') }}</span><span>{{ t('helpCenter.tavernUi.disconnect') }}</span></div>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="4" id="message" :title="t('helpCenter.specific.tavernChat')" :description="t('helpCenter.successBody')">
      <HelpConversationScene client-id="tavernai" client-name="TavernAI" :model="model" provider="Custom" />
    </GuideStep>
  </div>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import GuideStep from './GuideStep.vue'
import HelpSettingField from '../HelpSettingField.vue'
import HelpConversationScene from '../HelpConversationScene.vue'
import { useGuideFields, type ClientGuideProps } from './guideFields'
const props = defineProps<ClientGuideProps>(); const { t } = useI18n(); const { field, modelField, model } = useGuideFields(props)
</script>
<style scoped>
.tai-window { border: 1px solid #3a3128; border-radius: 12px; background: #1c1916; color: #f6f1e8; overflow: hidden; }
.tai-top { display: flex; align-items: center; gap: .6rem; padding: .85rem 1rem; background: #26211c; }
.plug { display: grid; place-items: center; width: 28px; height: 28px; border-radius: 6px; background: #e0a100; color: #2a1c04; font-weight: 700; }
.tai-body { padding: .3rem 1rem 1rem; }
.tavernai-guide :deep(.scene-field) { color: #f6f1e8; }
.tavernai-guide :deep(.field-input) { background: #12100e; border-color: #8a6a22; }
.tavernai-guide :deep(code), .tavernai-guide :deep(.scene-switch.is-on) { color: #f6f1e8; }
.tavernai-guide :deep(.scene-switch.is-on) { background: #e0a100; }
.load-row, .tai-actions, .tai-check { display: flex; align-items: center; gap: .6rem; }
.load { margin-left: auto; padding: .25rem .6rem; border-radius: 6px; background: #3a3128; }
.box { width: 14px; height: 14px; border: 1px solid #8a6a22; border-radius: 3px; }
.tai-button { padding: .4rem .8rem; border-radius: 6px; background: #e0a100; color: #2a1c04; font-weight: 650; }
</style>
