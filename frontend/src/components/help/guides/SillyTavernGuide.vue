<template>
  <div class="client-specific-guide sillytavern-guide" data-guide="sillytavern">
    <GuideStep :number="1" id="api" :title="t('helpCenter.specific.sillyApi')" :description="t('helpCenter.specific.sillyApiHint')">
      <div class="st-window" role="img" :aria-label="t('helpCenter.sillyUi.api')">
        <div class="st-title"><span>{{ t('helpCenter.sillyUi.api') }}</span><span class="st-status">{{ t('helpCenter.sillyUi.notConnected') }}</span></div>
        <div class="st-body">
          <HelpSettingField :field="field('API')" label="API" choice />
          <HelpSettingField :field="field('Chat Completion Source')" :label="t('helpCenter.sillyUi.source')" choice />
          <p class="screen-hint">{{ t('helpCenter.sillyUi.notBuiltin') }}</p>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="2" id="endpoint" :title="t('helpCenter.specific.sillyEndpoint')" :description="t('helpCenter.specific.sillyEndpointHint')">
      <div class="st-window">
        <div class="st-title"><span>{{ t('helpCenter.sillyUi.custom') }}</span></div>
        <div class="st-body">
          <HelpSettingField :field="field('Custom Endpoint (Base URL)')" :label="t('helpCenter.sillyUi.endpoint')" :hint="t('helpCenter.sillyUi.endpointNote')" />
          <HelpSettingField :field="field('Custom API Key')" :label="t('helpCenter.sillyUi.apiKey')" :hint="t('helpCenter.sillyUi.keyRequired')" />
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="3" id="model" :title="t('helpCenter.specific.sillyModel')" :description="t('helpCenter.specific.sillyModelHint')">
      <div class="st-window">
        <div class="st-body">
          <HelpSettingField v-if="modelField" :field="modelField" :label="t('helpCenter.sillyUi.modelId')" />
          <label class="st-check"><span class="box" aria-hidden="true" /><span>{{ t('helpCenter.sillyUi.bypass') }}</span></label>
          <p class="screen-hint">{{ t('helpCenter.sillyUi.bypassHint') }}</p>
          <div class="st-actions"><span class="st-button primary">{{ t('helpCenter.sillyUi.connect') }}</span><span class="st-button">{{ t('helpCenter.sillyUi.test') }}</span></div>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="4" id="reply" :title="t('helpCenter.specific.sillyChat')" :description="t('helpCenter.successBody')">
      <HelpConversationScene client-id="sillytavern" client-name="SillyTavern" :model="model" provider="Custom" />
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
.st-window { border: 1px solid #3a3358; border-radius: 12px; background: #221c33; color: #f4f0ff; overflow: hidden; }
.st-title { display: flex; justify-content: space-between; gap: .75rem; padding: .8rem 1rem; background: #2c2544; font-weight: 650; }
.st-status { color: #c8bddf; font-size: .72rem; font-weight: 500; }
.st-body { padding: .4rem 1rem 1rem; }
.sillytavern-guide :deep(.scene-field) { color: #f4f0ff; }
.sillytavern-guide :deep(.field-input) { background: #161226; border-color: #6d5cae; }
.sillytavern-guide :deep(code) { color: #f7f3ff; }
.st-check { display: flex; align-items: center; gap: .55rem; margin: .2rem 0; font-size: .78rem; }
.box { width: 14px; height: 14px; border: 1px solid #8d82b5; border-radius: 3px; background: #161226; }
.st-actions { display: flex; flex-wrap: wrap; gap: .5rem; margin-top: .8rem; }
.st-button { padding: .45rem .8rem; border-radius: 6px; background: #3a3358; }
.st-button.primary { background: #6d5cae; }
</style>
