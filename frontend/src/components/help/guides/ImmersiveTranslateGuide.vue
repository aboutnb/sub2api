<template>
  <div class="client-specific-guide immersive-guide" data-guide="immersive-translate">
    <GuideStep :number="1" id="custom-service" :title="t('helpCenter.specific.immersiveService')" :description="t('helpCenter.specific.immersiveServiceHint', { provider })">
      <div class="guide-screen immersive-settings">
        <div class="screen-columns">
          <aside class="screen-nav">
            <span class="nav-label">{{ t('helpCenter.scenes.customService') }}</span>
            <div class="service-item current">{{ provider }} 1</div>
          </aside>
          <div class="screen-body">
            <HelpSettingField :field="field('Provider')" :label="t('helpCenter.scenes.translationService')" choice />
            <HelpSettingField :field="field('Model')" :label="t('helpCenter.model')" />
            <div class="check-line"><span class="check on" aria-hidden="true">✓</span>{{ t('helpCenter.scenes.customModel') }}</div>
            <div class="context-option">
              <span>{{ t('helpCenter.specific.aiContext') }}</span>
              <span class="context-switch" aria-hidden="true"><i /></span>
            </div>
            <p class="screen-hint">{{ t('helpCenter.specific.aiContextHint') }}</p>
            <span class="field-caption">{{ t('helpCenter.scenes.strategyHint') }}</span>
            <div class="strategy">{{ t('helpCenter.scenes.general') }} <b aria-hidden="true">⌄</b></div>
            <HelpSettingField :field="field('API URL')" :label="t('helpCenter.scenes.customAddress')" :hint="t('helpCenter.specific.fullURLHint')" />
            <HelpSettingField :field="field('API Key')" label="APIKEY" />
          </div>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="2" id="open-page" :title="t('helpCenter.specific.immersivePanel')" :description="t('helpCenter.specific.immersivePanelHint')">
      <div class="guide-screen browser-page">
        <div class="screen-top"><strong>{{ t('helpCenter.specific.immersiveName') }}</strong><span>{{ provider }} 1</span></div>
        <div class="screen-body">
          <p>Hello, world.</p>
          <div class="translated-line"><small>{{ t('helpCenter.walkthrough.exampleReply') }}</small><p>{{ t('helpCenter.walkthrough.translation') }}</p></div>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="3" id="bilingual-page" :title="t('helpCenter.specific.immersiveVerify')" :description="t('helpCenter.specific.immersiveVerifyHint')">
      <HelpConversationScene client-id="immersive-translate" :client-name="t('helpCenter.specific.immersiveName')" :model="model" :provider="provider" />
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
.immersive-guide { --client-accent: #2563eb; }
.immersive-settings { max-width: 760px; margin-left: auto; margin-right: auto; }
.nav-label { color: var(--scene-faint); font-size: .72rem; padding: .2rem .45rem; }
.service-item { margin-top: .35rem; padding: .55rem .7rem; border-radius: 8px; }
.service-item.current { background: var(--scene-info-soft); color: var(--scene-info-fg); font-weight: 650; }
.check-line { display: flex; align-items: center; gap: .4rem; margin: .35rem 0 .8rem; color: var(--scene-muted); font-size: .75rem; }
.check { display: inline-grid; place-items: center; width: 14px; height: 14px; border-radius: 3px; background: #2b5aa7; color: #fff; font-size: .65rem; }
.context-option { display: flex; align-items: center; justify-content: space-between; gap: .75rem; margin-top: .4rem; padding: .7rem .8rem; border-radius: 8px; background: var(--scene-panel); }
.context-switch { display: flex; width: 36px; height: 20px; flex-shrink: 0; border-radius: 999px; background: var(--scene-chip); padding: 2px; }
.context-switch i { width: 16px; height: 16px; border-radius: 50%; background: white; }
.field-caption { display: block; margin-top: .8rem; color: var(--scene-muted); font-size: .72rem; }
.strategy { display: inline-flex; align-items: center; gap: .35rem; margin: .45rem 0 .2rem; padding: .4rem .7rem; border: 1px solid var(--scene-line); border-radius: 8px; }
.browser-page { max-width: 640px; margin-left: auto; margin-right: auto; }
.translated-line { margin-top: .8rem; padding-top: .8rem; border-top: 1px dashed var(--scene-line); }
.translated-line small { color: var(--scene-muted); }
.immersive-guide :deep(.field-input) { border: 2px solid #60a5fa; background: var(--scene-info-soft); }
</style>
