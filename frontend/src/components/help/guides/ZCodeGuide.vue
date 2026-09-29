<template>
  <div class="client-specific-guide zcode-guide" data-guide="zcode">
    <GuideStep :number="1" id="add-provider" :title="t('helpCenter.specific.zcodeProvider')" :description="t('helpCenter.specific.zcodeProviderHint')">
      <div class="zcode-window" role="img" :aria-label="t('helpCenter.zcodeUi.settings')">
        <div class="z-titlebar"><ClientLogo client="zcode" /><span class="z-controls" aria-hidden="true"><i /><i /><i /></span></div>
        <div class="z-layout">
          <aside>
            <div class="back">← {{ t('helpCenter.zcodeUi.back') }}</div>
            <div v-for="item in nav" :key="item" class="nav-item" :class="{ current: item === 'settings' }">{{ t(`helpCenter.zcodeUi.${item}`) }}</div>
            <div class="onboard">{{ t('helpCenter.zcodeUi.onboard') }}</div>
            <div class="connect"><span class="avatar" aria-hidden="true" />{{ t('helpCenter.zcodeUi.connect') }}<b aria-hidden="true">⚙</b></div>
          </aside>
          <main>
            <div class="settings-head"><h3>{{ t('helpCenter.zcodeUi.settings') }}</h3><span aria-hidden="true">↻</span></div>
            <p>{{ t('helpCenter.zcodeUi.description') }}</p>
            <div class="settings-grid">
              <div class="provider-list">
                <strong>{{ t('helpCenter.zcodeUi.providers') }}</strong>
                <div class="builtin"><span class="diamond" aria-hidden="true" />BigModel</div>
                <strong>{{ t('helpCenter.zcodeUi.customProviders') }}</strong>
                <div class="custom-provider"><ClientLogo client="zcode" /><span>{{ provider }}</span><i /></div>
                <div class="add-provider">+ {{ t('helpCenter.zcodeUi.addProvider') }}</div>
              </div>
              <div class="provider-editor">
                <div class="editor-title"><ClientLogo client="zcode" /><strong>{{ provider }}</strong><span>{{ t('helpCenter.zcodeUi.enabled') }}</span></div>
                <div class="field-pair">
                  <HelpSettingField :field="field('Provider')" :label="t('helpCenter.zcodeUi.providerName')" />
                  <HelpSettingField :field="field('API Key')" label="API Key" />
                </div>
                <HelpSettingField :field="field('Base URL')" label="API Base URL" />
                <div class="model-box">
                  <div><strong>{{ t('helpCenter.zcodeUi.models') }}</strong><span class="add-model">+ {{ t('helpCenter.zcodeUi.addModel') }}</span></div>
                  <div class="model-line"><i /><code>{{ model }}</code></div>
                </div>
                <div class="save-row"><span>{{ t('helpCenter.zcodeUi.saveProvider') }}</span></div>
              </div>
            </div>
          </main>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="2" id="add-model" :title="t('helpCenter.specific.zcodeModel')" :description="t('helpCenter.specific.zcodeModelHint')">
      <div class="zcode-dialog">
        <div class="dialog-title"><strong>{{ t('helpCenter.zcodeUi.addModel') }}</strong><span aria-hidden="true">×</span></div>
        <HelpSettingField v-if="modelField" :field="modelField" :label="t('helpCenter.zcodeUi.modelId')" />
        <span class="context-label">{{ t('helpCenter.zcodeUi.contextWindow') }}</span>
        <div class="screen-value context-value">-</div>
        <p class="screen-hint">{{ t('helpCenter.specific.actualCapacity') }}</p>
        <div class="dialog-actions"><span>{{ t('helpCenter.zcodeUi.cancel') }}</span><span class="save-highlight">{{ t('helpCenter.zcodeUi.save') }}</span></div>
      </div>
    </GuideStep>
    <GuideStep :number="3" id="choose-model" :title="t('helpCenter.specific.zcodeChoose')" :description="t('helpCenter.specific.zcodeChooseHint')">
      <div class="zcode-picker">
        <div class="project">{{ t('helpCenter.zcodeUi.project') }} <span aria-hidden="true">⌄</span></div>
        <div class="composer">
          <p>{{ t('helpCenter.zcodeUi.ask') }}</p>
          <span class="plus" aria-hidden="true">+</span>
          <div class="picker-menu">
            <div class="floating-model">{{ model }}</div>
            <div class="provider-menu"><ClientLogo client="zcode" /><span>{{ provider }}</span><b aria-hidden="true">›</b></div>
            <div class="manage">{{ t('helpCenter.zcodeUi.manageModels') }}</div>
          </div>
          <div class="composer-actions"><span>Choose model <i aria-hidden="true">⌄</i></span><span class="send" aria-hidden="true">↑</span></div>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="4" id="workspace" :title="t('helpCenter.specific.zcodeChat')" :description="t('helpCenter.successBody')">
      <div class="zcode-chat">
        <div class="user-message"><span>{{ t('helpCenter.zcodeUi.hi') }}</span><small>02:23 PM</small></div>
        <p class="worked">{{ t('helpCenter.zcodeUi.worked') }}</p>
        <p class="reply">{{ t('helpCenter.zcodeUi.reply') }}</p>
        <div class="followup">
          <p>{{ t('helpCenter.zcodeUi.followup') }}</p>
          <div><span>{{ t('helpCenter.zcodeUi.askBefore') }} <i aria-hidden="true">⌄</i></span><code>{{ model }}</code><span>{{ t('helpCenter.zcodeUi.medium') }} <i aria-hidden="true">⌄</i></span><span class="send" aria-hidden="true">↑</span></div>
        </div>
      </div>
    </GuideStep>
  </div>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import GuideStep from './GuideStep.vue'
import HelpSettingField from '../HelpSettingField.vue'
import ClientLogo from '@/components/common/ClientLogo.vue'
import { useGuideFields, type ClientGuideProps } from './guideFields'
const props = defineProps<ClientGuideProps>()
const { t } = useI18n()
const { field, model, modelField, provider } = useGuideFields(props)
const nav = ['general', 'preview', 'settings', 'skills', 'subagents', 'mcp', 'plugins', 'commands', 'index', 'usage']
</script>
<style scoped>
.zcode-guide { --client-accent: #111827; }
.zcode-window, .zcode-dialog, .zcode-picker, .zcode-chat { border: 1px solid var(--scene-line); border-radius: 16px; background: var(--scene-panel); color: var(--scene-fg); overflow: hidden; }
.z-titlebar { display: flex; align-items: center; height: 36px; padding: 0 .8rem; border-bottom: 1px solid var(--scene-line); background: var(--scene-panel); }
.z-controls { display: flex; gap: .7rem; margin-left: auto; color: var(--scene-muted); }
.z-controls i { width: 10px; height: 10px; border: 1.5px solid currentColor; border-radius: 1px; }
.z-controls i:first-child { border: 0; border-bottom: 1.5px solid currentColor; height: 6px; }
.z-layout { display: grid; grid-template-columns: 168px minmax(0, 1fr); min-height: 520px; }
aside { display: flex; flex-direction: column; gap: .15rem; padding: .7rem; border-right: 1px solid var(--scene-line); background: var(--scene-panel); }
.back, .nav-item, .onboard, .connect { display: flex; align-items: center; gap: .4rem; padding: .45rem .5rem; border-radius: 8px; font-size: .72rem; }
.nav-item.current { background: var(--scene-chip); font-weight: 650; }
.onboard { margin-top: .4rem; border: 1px dashed var(--scene-line); }
.connect { margin-top: auto; font-weight: 650; }
.avatar { width: 22px; height: 22px; border: 1px solid var(--scene-line); border-radius: 50%; background: var(--scene-bg); }
.connect b { margin-left: auto; font-weight: 500; }
main { min-width: 0; padding: 1.1rem 1.2rem 1.2rem; }
.settings-head { display: flex; align-items: center; justify-content: space-between; }
h3 { margin: 0; font-size: 1.35rem; }
main > p { margin: .35rem 0 .9rem; color: var(--scene-muted); font-size: .75rem; }
.settings-grid { display: grid; grid-template-columns: 150px minmax(0, 1fr); min-height: 360px; overflow: hidden; border: 1px solid var(--scene-line); border-radius: 12px; background: var(--scene-bg); }
.provider-list { display: flex; flex-direction: column; gap: .45rem; padding: .8rem; border-right: 1px solid var(--scene-line); }
.provider-list strong { color: var(--scene-faint); font-size: .68rem; }
.builtin, .custom-provider, .add-provider { display: flex; align-items: center; gap: .4rem; padding: .45rem; border-radius: 8px; font-size: .75rem; }
.builtin { border: 1px solid var(--scene-line); }
.diamond { width: 14px; height: 14px; rotate: 45deg; background: linear-gradient(135deg, #818cf8, #2563eb); }
.custom-provider { background: var(--scene-panel); font-weight: 650; box-shadow: 0 0 0 2px #34d399; }
.custom-provider span { min-width: 0; overflow-wrap: anywhere; }
.custom-provider i, .model-line i { width: 7px; height: 7px; margin-left: auto; border-radius: 50%; background: #10b981; flex-shrink: 0; }
.provider-editor { min-width: 0; padding: .9rem; }
.editor-title { display: flex; align-items: center; gap: .45rem; padding-bottom: .7rem; border-bottom: 1px solid var(--scene-line-soft); }
.editor-title span { border-radius: 999px; background: #0a6650; color: #fff; padding: .1rem .45rem; font-size: .65rem; }
.model-box { margin-top: .8rem; padding: .75rem; border: 1px solid var(--scene-line); border-radius: 12px; background: var(--scene-panel); }
.model-box > div:first-child, .save-row, .dialog-actions, .composer-actions, .followup > div { display: flex; align-items: center; justify-content: space-between; gap: .5rem; }
.add-model, .save-row span, .save-highlight, .send { border-radius: 8px; background: var(--scene-solid); color: var(--scene-solid-fg); }
.add-model { padding: .35rem .55rem; box-shadow: 0 0 0 2px #34d399; font-size: .72rem; }
.model-line { display: flex; align-items: center; gap: .45rem; margin-top: .7rem; padding: .55rem .7rem; border-radius: 8px; background: var(--scene-bg); }
.model-line code, .followup code { min-width: 0; overflow-wrap: anywhere; font-size: .75rem; }
.save-row { justify-content: flex-end; margin-top: .8rem; }
.save-row span, .save-highlight { padding: .45rem .8rem; font-size: .75rem; }
.zcode-dialog { max-width: 640px; margin: 0 auto; padding: 1rem 1.1rem 1.1rem; background: var(--scene-bg); }
.dialog-title { display: flex; justify-content: space-between; font-size: 1.05rem; }
.context-label { display: block; margin-top: .8rem; color: var(--scene-muted); font-size: .8rem; }
.context-value { color: var(--scene-faint); }
.save-highlight { box-shadow: 0 0 0 2px #34d399; }
.zcode-picker { padding: .8rem; background: var(--scene-panel); }
.project { display: flex; align-items: center; gap: .35rem; margin-bottom: .5rem; font-size: .8rem; }
.composer { position: relative; min-height: 180px; padding: 1rem; border-radius: 16px; background: var(--scene-bg); }
.composer p, .followup p { margin: 0; color: var(--scene-faint); }
.plus { position: absolute; left: 1rem; bottom: 1rem; color: var(--scene-muted); font-size: 1.4rem; }
.picker-menu { position: absolute; right: 4.5rem; bottom: 3.2rem; width: min(220px, 70%); }
.floating-model, .provider-menu, .manage { margin-bottom: .35rem; padding: .45rem .6rem; border-radius: 10px; background: var(--scene-bg); box-shadow: var(--av-shadow-md); font-size: .75rem; }
.provider-menu { display: flex; align-items: center; gap: .4rem; }
.provider-menu b { margin-left: auto; }
.composer-actions { position: absolute; right: 1rem; bottom: .8rem; }
.composer-actions span:first-child { padding: .4rem .6rem; border-radius: 8px; background: var(--scene-panel); font-size: .75rem; }
.send { display: grid; place-items: center; width: 32px; height: 32px; }
.zcode-chat { display: flex; flex-direction: column; min-height: 460px; padding: 1.2rem; background: var(--scene-panel); }
.user-message { align-self: flex-end; text-align: right; }
.user-message span { display: inline-block; padding: .55rem .8rem; border-radius: 12px; background: var(--scene-panel); }
.user-message small, .worked { color: var(--scene-faint); font-size: .72rem; }
.reply { margin: .8rem 0 1.5rem; padding-top: .8rem; border-top: 1px solid var(--scene-line); }
.followup { margin-top: auto; padding: .9rem; border: 1px solid var(--scene-line); border-radius: 16px; background: var(--scene-bg); }
.followup > div { margin-top: 1.5rem; font-size: .75rem; }
.zcode-guide :deep(.field-input) { border-color: var(--scene-line); background: var(--scene-bg); }
@container help-scenes (max-width: 720px) {
  .z-layout, .settings-grid, .field-pair { grid-template-columns: minmax(0, 1fr); }
  aside { flex-direction: row; flex-wrap: wrap; border-right: 0; border-bottom: 1px solid var(--scene-line); }
  .connect { margin-top: 0; }
  .picker-menu { position: static; width: auto; margin-top: .8rem; }
  .composer { min-height: 0; padding-bottom: 3.5rem; }
}
</style>
