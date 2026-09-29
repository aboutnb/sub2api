<template>
  <div class="client-specific-guide trae-guide" data-guide="trae">
    <GuideStep :number="1" id="custom-model" :title="t('helpCenter.specific.traeForm')" :description="t('helpCenter.specific.traeFormHint')">
      <div class="trae-window">
        <div class="trae-title"><strong>{{ t('helpCenter.scenes.addModel') }}</strong><span aria-hidden="true">×</span></div>
        <div class="trae-sheet">
          <div class="custom-config">
            <div class="custom-header"><span class="cube" aria-hidden="true" />{{ t('helpCenter.scenes.customConfig') }}<b aria-hidden="true">⌃</b></div>
            <div class="custom-fields">
              <HelpSettingField :field="field('API 格式')" :label="requiredLabel('API 格式')" choice />
              <HelpSettingField :field="field('自定义请求地址')" :label="requiredLabel('自定义请求地址')" />
              <HelpSettingField :field="field('完整 URL')" toggle />
              <p class="endpoint-note">{{ t('helpCenter.specific.traeEndpointHint', { format: field('API 格式').content, path: appendedPath }) }}</p>
              <div class="model-head"><span class="multimodal"><i class="unchecked" aria-hidden="true" />{{ t('helpCenter.specific.multimodal') }}<small>{{ t('helpCenter.specific.verifyCapability') }}</small></span></div>
              <HelpSettingField :field="field('模型 ID')" :label="requiredLabel('模型 ID')" />
              <HelpSettingField :field="field('API 密钥')" :label="requiredLabel('API 密钥')" />
              <div class="trae-actions"><span>{{ t('helpCenter.traeUi.cancel') }}</span><span class="submit">{{ t('helpCenter.specific.submit') }}</span></div>
            </div>
          </div>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="2" id="solo-model" :title="t('helpCenter.specific.traeSelect')" :description="t('helpCenter.specific.traeSelectHint')">
      <div class="trae-chat">
        <div class="solo-mark"><i>S</i>{{ t('helpCenter.traeUi.solo') }}</div>
        <div class="composer">
          <p>{{ t('helpCenter.traeUi.taskSummary') }}</p>
          <div><span aria-hidden="true">▢</span><b>{{ model }} <i aria-hidden="true">⌄</i></b><span class="send" aria-hidden="true">↑</span></div>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="3" id="solo-test" :title="t('helpCenter.specific.traeTest')" :description="t('helpCenter.successBody')">
      <div class="trae-chat">
        <div class="user-bubble">{{ t('helpCenter.traeUi.hi') }}</div>
        <div class="solo-mark"><i>S</i>{{ t('helpCenter.traeUi.solo') }}</div>
        <p class="reply">{{ t('helpCenter.traeUi.reply') }}</p>
        <div class="task-done"><i aria-hidden="true">✓</i>{{ t('helpCenter.traeUi.taskDone') }}</div>
        <div class="composer">
          <p>{{ t('helpCenter.traeUi.taskSummary') }}</p>
          <div><span aria-hidden="true">▢</span><b>{{ model }} <i aria-hidden="true">⌄</i></b><span class="send" aria-hidden="true">↑</span></div>
        </div>
      </div>
    </GuideStep>
  </div>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GuideStep from './GuideStep.vue'
import HelpSettingField from '../HelpSettingField.vue'
import { useGuideFields, type ClientGuideProps } from './guideFields'
const props = defineProps<ClientGuideProps>()
const { t } = useI18n()
const { field, model } = useGuideFields(props)
const appendedPath = computed(() => field('API 格式').content.startsWith('Anthropic') ? '/v1/messages' : '/chat/completions')
const requiredLabel = (path: string) => `* ${path}`
</script>
<style scoped>
.trae-guide { --client-accent: #6d28d9; }
.trae-window, .trae-chat { overflow: hidden; border: 1px solid var(--scene-line); border-radius: 16px; background: var(--scene-bg); box-shadow: var(--scene-shadow); }
.trae-window { max-width: 680px; margin: 0 auto; }
.trae-title { display: flex; align-items: center; justify-content: space-between; padding: 1rem 1.2rem; border-bottom: 1px solid var(--scene-line); font-size: 1.05rem; }
.trae-sheet { padding: .8rem 1.1rem 1.1rem; background: var(--scene-panel); }
.custom-config { border: 1px solid var(--scene-line); border-radius: 10px; background: var(--scene-panel); }
.custom-header { display: flex; align-items: center; gap: .7rem; padding: .8rem 1rem; border-bottom: 1px solid var(--scene-line); font-weight: 750; }
.cube { width: 28px; height: 28px; border-radius: 7px; background: var(--scene-bg); box-shadow:inset 0 0 0 1.5px var(--scene-fg); }
.custom-header b { margin-left: auto; }
.custom-fields { padding: .3rem 1rem 1rem; }
.url-head, .model-head, .full-url, .multimodal { display: flex; align-items: center; gap: .45rem; }
.url-head, .model-head { justify-content: space-between; margin-top: .2rem; font-size: .78rem; font-weight: 700; }
.full-url { font-weight: 650; }
.unchecked { width: 14px; height: 14px; border: 1px solid var(--scene-faint); border-radius: 3px; }
.multimodal small { color: var(--scene-muted); font-weight: 500; }
.endpoint-note { border: 1px solid var(--scene-info-soft); border-radius: 8px; background: var(--scene-info-soft); padding: .7rem .8rem; color: var(--scene-fg); font-size: .75rem; line-height: 1.7; }
.trae-actions { display: flex; justify-content: flex-end; gap: .6rem; padding-top: .4rem; }
.trae-actions span { border-radius: 8px; padding: .45rem 1.1rem; font-weight: 750; background: var(--scene-chip); }
.trae-actions .submit { background: var(--scene-solid); color: var(--scene-solid-fg); }
.trae-chat { padding: 1.1rem 1.2rem 1.2rem; }
.user-bubble { width: fit-content; margin-left: auto; border-radius: 12px; background: var(--scene-panel); padding: .7rem .9rem; font-weight: 650; }
.solo-mark { display: flex; align-items: center; gap: .45rem; margin: .9rem 0 .4rem; color: var(--scene-muted); font-weight: 700; }
.solo-mark i, .send { display: grid; place-items: center; width: 24px; height: 24px; border-radius: 6px; background: #5d42a7; color: #fff; font-style: normal; }
.reply { margin: .2rem 0 .8rem; font-size: 1.05rem; }
.task-done { display: flex; align-items: center; gap: .4rem; color: var(--scene-muted); font-weight: 750; }
.task-done i { display: grid; place-items: center; width: 18px; height: 18px; border-radius: 50%; background: #10b981; color: #022c22; font-style: normal; font-size: .7rem; }
.composer { margin-top: .9rem; border: 1px solid var(--scene-info-soft); border-radius: 16px; padding: .9rem 1rem; box-shadow: var(--scene-shadow); }
.composer p { min-height: 3rem; margin: 0; color: var(--scene-faint); }
.composer > div { display: flex; align-items: center; gap: .7rem; }
.composer b { margin-left: auto; max-width: 55%; overflow-wrap: anywhere; font-size: .85rem; }
.send { width: 34px; height: 34px; border-radius: 8px; background: var(--scene-violet-soft); color: var(--scene-violet-fg); }
.trae-guide :deep(.field-label span) { font-weight: 750; }
.trae-guide :deep(.field-input) { border-radius: 8px; background: var(--scene-panel); }
@container help-scenes (max-width: 640px) {
  .url-head, .model-head { align-items: flex-start; flex-direction: column; }
}
</style>
