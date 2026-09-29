<template>
  <div class="client-specific-guide kiss-guide" data-guide="kiss-translator">
    <GuideStep :number="1" id="add-service" :title="t('helpCenter.specific.kissAdd')" :description="t('helpCenter.specific.kissAddHint', { provider })">
      <div class="guide-screen">
        <div class="screen-top"><strong>{{ t('helpCenter.scenes.translationService') }}</strong><span class="illustrated-action">Add</span></div>
        <div class="screen-body"><HelpSettingField :field="field('Provider')" choice /></div>
      </div>
    </GuideStep>
    <GuideStep :number="2" id="service-form" :title="t('helpCenter.specific.kissForm')" :description="t('helpCenter.specific.kissFormHint')">
      <div class="guide-screen kiss-form">
        <div class="screen-top"><strong>[{{ provider }}] {{ t('helpCenter.walkthrough.site') }}</strong><span aria-hidden="true">⌃</span></div>
        <div class="screen-body">
          <div class="field-pair">
            <div class="float-field"><span>API Name</span><div class="screen-value">{{ t('helpCenter.walkthrough.site') }}</div></div>
            <div class="float-field"><span>Sort Order</span><div class="screen-value">0</div><small>{{ t('helpCenter.scenes.sortEarlier') }}</small></div>
          </div>
          <HelpSettingField :field="field('API URL')" label="URL" :hint="t('helpCenter.specific.fullURLHint')" />
          <HelpSettingField :field="field('API Key')" label="Key" :hint="t('helpCenter.scenes.kissKeyHint')" />
          <div class="kiss-options">
            <HelpSettingField :field="field('Model')" label="Model" />
            <div class="float-field"><span>Translation style</span><div class="screen-value">formal</div></div>
            <div class="float-field"><span>Temperature (0.0-2.0)</span><div class="screen-value">0</div></div>
            <div class="float-field"><span>Max Tokens (0-1000000)</span><div class="screen-value">20480</div></div>
          </div>
        </div>
        <div class="screen-bottom"><span class="illustrated-action illustrated-primary">Save</span></div>
      </div>
    </GuideStep>
    <GuideStep :number="3" id="refresh-translate" :title="t('helpCenter.specific.kissTranslate')" :description="t('helpCenter.specific.kissTranslateHint')">
      <HelpConversationScene client-id="kiss-translator" client-name="Kiss Translator" :model="model" :provider="provider" />
    </GuideStep>
  </div>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import GuideStep from './GuideStep.vue'
import HelpSettingField from '../HelpSettingField.vue'
import HelpConversationScene from '../HelpConversationScene.vue'
import { useGuideFields, type ClientGuideProps } from './guideFields'
const props = defineProps<ClientGuideProps>()
const { t } = useI18n()
const { field, model, provider } = useGuideFields(props)
</script>
<style scoped>
.kiss-guide { --client-accent: #2563eb; }
.kiss-form { max-width: 760px; margin-left: auto; margin-right: auto; border-radius: 8px; }
.float-field span { display: block; margin-bottom: .35rem; color: var(--scene-muted); font-size: .72rem; }
.kiss-options { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: .75rem; align-items: start; }
.kiss-guide :deep(.field-input) { border: 2px solid #60a5fa; background: var(--scene-info-soft); }
.kiss-options :deep(.scene-field) { margin-top: 0; }
@container help-scenes (max-width: 720px) { .kiss-options { grid-template-columns: minmax(0, 1fr); } }
</style>
