<template>
  <div class="client-specific-guide cline-guide" data-guide="cline">
    <GuideStep :number="1" id="cline-provider" :title="t('helpCenter.specific.clineProvider')" :description="t('helpCenter.specific.clineProviderHint')"><div class="guide-screen cline-sidebar"><div class="screen-top"><strong>CLINE / Settings</strong><span>Done</span></div><div class="screen-body"><strong>API Configuration</strong><HelpSettingField :field="field('Provider')" label="API Provider" choice /><HelpSettingField :field="field('Base URL')" /><HelpSettingField :field="field('API Key')" /><HelpSettingField v-if="modelField" :field="modelField" /><p class="screen-hint">{{ t('helpCenter.scenes.capabilities') }}</p></div></div></GuideStep>
    <GuideStep :number="2" id="plan-act" :title="t('helpCenter.specific.clineModes')" :description="t('helpCenter.specific.clineModesHint')"><div class="guide-screen"><div class="screen-body"><div class="mode-tabs"><strong>Plan</strong><span>Act</span></div><div class="inline-model"><strong>OpenAI Compatible</strong><code>{{ model }}</code></div><p class="screen-hint">{{ t('helpCenter.walkthrough.notes.cline') }}</p></div></div></GuideStep>
    <GuideStep :number="3" id="plan-test" :title="t('helpCenter.specific.clineTest')" :description="t('helpCenter.successBody')"><HelpConversationScene client-id="cline" client-name="Cline · Plan" :model="model" :provider="provider" /></GuideStep>
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
<style scoped>.cline-sidebar { max-width: 500px; margin-left: auto; margin-right: auto; }.mode-tabs { display: flex; gap: 1rem; margin-bottom: 1rem; border-bottom: 1px solid var(--scene-line); }.mode-tabs > * { padding: .5rem 1rem; }.mode-tabs strong { border-bottom: 2px solid var(--scene-fg); }</style>
