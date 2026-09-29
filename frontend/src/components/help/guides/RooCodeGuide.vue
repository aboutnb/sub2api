<template>
  <div class="client-specific-guide roo-guide" data-guide="roo-code">
    <GuideStep :number="1" id="roo-profile" :title="t('helpCenter.specific.rooProfile')" :description="t('helpCenter.specific.rooProfileHint')"><div class="guide-screen screen-columns"><aside class="screen-nav"><span class="current">Providers</span><span>Auto-Approve</span><span>Browser</span><span>Terminal</span></aside><div class="screen-body"><strong>API Configuration</strong><div class="screen-value">{{ t('helpCenter.walkthrough.site') }} ⌄</div><HelpSettingField :field="field('Provider')" label="API Provider" choice /><HelpSettingField :field="field('Base URL')" /><HelpSettingField :field="field('API Key')" /><HelpSettingField v-if="modelField" :field="modelField" /></div></div></GuideStep>
    <GuideStep :number="2" id="roo-capabilities" :title="t('helpCenter.specific.rooTools')" :description="t('helpCenter.specific.rooToolsHint')"><div class="guide-note">{{ t('helpCenter.scenes.capabilities') }}</div></GuideStep>
    <GuideStep :number="3" id="active-profile" :title="t('helpCenter.specific.rooActivate')" :description="t('helpCenter.specific.rooActivateHint')"><div class="guide-screen"><div class="screen-top"><strong>Roo Code</strong><span>API Configuration</span></div><div class="screen-body"><div class="inline-model"><strong>{{ t('helpCenter.walkthrough.site') }}</strong><code>{{ model }}</code><span>✓</span></div><p class="screen-hint">{{ t('helpCenter.scenes.selectionExample') }}</p></div></div></GuideStep>
    <GuideStep :number="4" id="roo-request" :title="t('helpCenter.specific.rooTest')" :description="t('helpCenter.successBody')"><HelpConversationScene client-id="roo-code" client-name="Roo Code" :model="model" :provider="provider" /></GuideStep>
  </div>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import GuideStep from './GuideStep.vue'
import HelpSettingField from '../HelpSettingField.vue'
import HelpConversationScene from '../HelpConversationScene.vue'
import { useGuideFields, type ClientGuideProps } from './guideFields'
const props = defineProps<ClientGuideProps>(); const { t } = useI18n(); const { field, modelField, model, provider } = useGuideFields(props)
</script>
