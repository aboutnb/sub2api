<template>
  <div class="client-specific-guide workbuddy-guide" data-guide="workbuddy">
    <GuideStep :number="1" id="account-menu" :title="t('helpCenter.specific.workbuddyMenu')" :description="t('helpCenter.specific.workbuddyMenuHint')">
      <div class="wb-account" role="img" :aria-label="t('helpCenter.specific.workbuddyMenu')">
        <div class="wb-popover">
          <div class="wb-userline"><strong>{{ t('helpCenter.specific.yourAccount') }}</strong><span aria-hidden="true">⧉</span></div>
          <div class="wb-row"><span>{{ t('helpCenter.workbuddyUi.trial') }}</span><b>{{ t('helpCenter.workbuddyUi.upgrade') }}</b></div>
          <div class="wb-station">
            <div class="wb-station-head"><span>{{ t('helpCenter.workbuddyUi.buddyStation') }}</span><small>{{ t('helpCenter.workbuddyUi.eventChip') }}</small></div>
            <strong>{{ t('helpCenter.workbuddyUi.issue') }}</strong>
            <p>{{ t('helpCenter.workbuddyUi.dailyPoints') }}</p>
            <div><span>{{ t('helpCenter.workbuddyUi.claimNow') }}</span><span>{{ t('helpCenter.workbuddyUi.tryInspiration') }}</span></div>
          </div>
          <div class="wb-row"><span>{{ t('helpCenter.workbuddyUi.pointsBalance') }}</span><em>{{ t('helpCenter.workbuddyUi.pointsValue') }} ›</em></div>
          <div class="wb-row"><span>{{ t('helpCenter.workbuddyUi.growthPlan') }}</span><em>{{ t('helpCenter.workbuddyUi.growthDetail') }} ›</em></div>
          <div class="wb-row is-current"><span>{{ t('helpCenter.specific.settings') }}</span><em>{{ t('helpCenter.workbuddyUi.clickToEnter') }}</em></div>
          <div class="wb-row"><span>{{ t('helpCenter.workbuddyUi.appearance') }}</span><em>{{ t('helpCenter.workbuddyUi.light') }} {{ t('helpCenter.workbuddyUi.dark') }}</em></div>
          <div class="wb-row"><span>{{ t('helpCenter.workbuddyUi.helpFeedback') }}</span></div>
          <div class="wb-row"><span>{{ t('helpCenter.specific.checkUpdates') }}</span></div>
          <div class="wb-row"><span>{{ t('helpCenter.workbuddyUi.logout') }}</span></div>
        </div>
        <div class="wb-account-foot"><i>W</i><strong>{{ t('helpCenter.specific.yourAccount') }}</strong></div>
      </div>
    </GuideStep>
    <GuideStep :number="2" id="model-settings" :title="t('helpCenter.specific.workbuddyModels')" :description="t('helpCenter.specific.workbuddyModelsHint')">
      <div class="wb-models" role="img" :aria-label="t('helpCenter.specific.workbuddyModels')">
        <aside>
          <div v-for="item in settingsNav" :key="item" :class="{ current: item === 'models' }">{{ t(`helpCenter.workbuddyUi.${item}`) }}</div>
        </aside>
        <section>
          <div class="wb-page-title"><h3>{{ t('helpCenter.model') }}</h3><span aria-hidden="true">×</span></div>
          <strong class="wb-section">{{ t('helpCenter.specific.customModels') }}</strong>
          <div class="model-file">
            <div>
              <strong>{{ t('helpCenter.specific.localConfig') }}</strong>
              <p>{{ t('helpCenter.workbuddyUi.localConfigPrefix') }} <code>{{ t('helpCenter.workbuddyUi.configPath') }}</code></p>
            </div>
            <span class="add-model">+ {{ t('helpCenter.scenes.addModel') }}</span>
          </div>
          <strong class="wb-section">{{ t('helpCenter.workbuddyUi.savedModels') }}</strong>
          <div class="saved-model"><span aria-hidden="true">+</span><div><strong>{{ model }}</strong><small>{{ t('helpCenter.workbuddyUi.custom') }}</small></div><em aria-hidden="true">✎</em><em aria-hidden="true">⌫</em></div>
        </section>
      </div>
    </GuideStep>
    <GuideStep :number="3" id="edit-model" :title="t('helpCenter.scenes.editModel')" :description="t('helpCenter.scenes.workbuddyCustom')">
      <div class="guide-screen workbuddy-edit">
        <div class="screen-top"><strong>{{ t('helpCenter.scenes.editModel') }}</strong><span>×</span></div>
        <div class="screen-body">
          <p class="screen-hint">{{ t('helpCenter.workbuddyUi.openaiOnly') }}</p>
          <HelpSettingField :field="field('Provider')" :label="t('helpCenter.scenes.provider')" choice />
          <HelpSettingField :field="field('URL')" :label="t('helpCenter.scenes.endpoint')" />
          <HelpSettingField :field="field('API Key')" />
          <HelpSettingField :field="field('Model')" :label="t('helpCenter.scenes.modelName')" />
          <strong class="advanced-title">{{ t('helpCenter.scenes.advanced') }}</strong>
          <div class="cap-grid">
            <span class="cap-item"><i class="on">✓</i>{{ t('helpCenter.scenes.workbuddyToolCalling') }}</span>
            <span class="cap-item"><i class="on">✓</i>{{ t('helpCenter.scenes.workbuddyImageInput') }}</span>
            <span class="cap-item"><i class="on">✓</i>{{ t('helpCenter.scenes.workbuddyThinking') }}</span>
            <span class="cap-item"><i class="on">✓</i>{{ t('helpCenter.scenes.workbuddyThinkingOnly') }}</span>
            <span class="cap-item"><i class="on">✓</i>{{ t('helpCenter.scenes.workbuddyAllowDisableThinking') }}</span>
          </div>
          <HelpSettingField :field="field('Custom Protocol')" :label="t('helpCenter.scenes.customProtocol')" toggle :hint="t('helpCenter.scenes.customProtocolHint')" />
          <span class="field-caption">{{ t('helpCenter.scenes.workbuddyDefaultReasoning') }}</span>
          <div class="screen-value">{{ t('helpCenter.scenes.workbuddyAutoReasoning') }} <b aria-hidden="true">⌄</b></div>
          <span class="field-caption">{{ t('helpCenter.scenes.workbuddySupportedReasoning') }}</span>
          <div class="reason-grid">
            <span class="cap-item"><i class="on">✓</i>{{ t('helpCenter.scenes.workbuddyLow') }}</span>
            <span class="cap-item"><i class="on">✓</i>{{ t('helpCenter.scenes.workbuddyMedium') }}</span>
            <span class="cap-item"><i class="on">✓</i>{{ t('helpCenter.scenes.workbuddyHigh') }}</span>
            <span class="cap-item"><i class="on">✓</i>{{ t('helpCenter.scenes.workbuddyVeryHigh') }}</span>
            <span class="cap-item"><i class="on">✓</i>{{ t('helpCenter.scenes.workbuddyExtreme') }}</span>
          </div>
          <div class="field-pair">
            <div><span class="field-caption">{{ t('helpCenter.scenes.workbuddyInput') }}</span><div class="screen-value">-</div></div>
            <div><span class="field-caption">{{ t('helpCenter.scenes.workbuddyOutput') }}</span><div class="screen-value">-</div></div>
          </div>
          <p class="screen-hint">{{ t('helpCenter.scenes.workbuddyAdvancedHint') }}</p>
        </div>
        <div class="screen-bottom"><span>{{ t('common.cancel') }}</span><span class="illustrated-action illustrated-primary">{{ t('common.save') }}</span></div>
      </div>
    </GuideStep>
    <GuideStep :number="4" id="task-model" :title="t('helpCenter.specific.workbuddyTask')" :description="t('helpCenter.specific.workbuddyTaskHint')">
      <div class="wb-chat" role="img" :aria-label="t('helpCenter.specific.workbuddyTask')">
        <div class="wb-menubar"><ClientLogo client="workbuddy" /><strong>WorkBuddy</strong><span>{{ t('helpCenter.workbuddyUi.editMenu') }}</span><span>{{ t('helpCenter.workbuddyUi.windowMenu') }}</span><span>{{ t('helpCenter.workbuddyUi.helpMenu') }}</span><em aria-hidden="true">— □ ×</em></div>
        <div class="workspace-conversation">
          <aside>
            <div class="rail-brand"><ClientLogo client="workbuddy" /><span>WorkBuddy</span></div>
            <div v-for="item in rail" :key="item">{{ t(`helpCenter.workbuddyUi.${item}`) }}</div>
            <small>{{ t('helpCenter.workbuddyUi.tasks') }}</small>
            <div class="rail-current">{{ t('helpCenter.workbuddyUi.connectionTest') }}</div>
            <small>{{ t('helpCenter.workbuddyUi.spaces') }}</small>
            <div>{{ t('helpCenter.workbuddyUi.onboarding') }}</div>
          </aside>
          <main>
            <div class="chat-title">{{ t('helpCenter.workbuddyUi.connectionTest') }}</div>
            <div class="user-bubble">{{ t('helpCenter.workbuddyUi.testPrompt') }}</div>
            <div class="assistant-block">
              <ClientLogo client="workbuddy" />
              <div>
                <strong>WorkBuddy</strong>
                <small>{{ t('helpCenter.workbuddyUi.completed') }}</small>
                <p>{{ t('helpCenter.workbuddyUi.connected') }}</p>
                <p>{{ t('helpCenter.workbuddyUi.connectedModel', { model }) }}</p>
              </div>
            </div>
            <div class="composer">
              <p>{{ t('helpCenter.workbuddyUi.composerPlaceholder') }}</p>
              <div><span>＋</span><span>{{ t('helpCenter.workbuddyUi.defaultPermissions') }}</span><b>{{ model }} ⌄</b><i aria-hidden="true">↑</i></div>
            </div>
            <small class="disclaimer">{{ t('helpCenter.workbuddyUi.aiDisclaimer') }}</small>
          </main>
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
const { field, model } = useGuideFields(props)
const settingsNav = ['accountManagement', 'agentMailbox', 'systemSettings', 'agentSettings', 'shortcuts', 'memory', 'models', 'assistantSettings', 'personalization', 'dataManagement', 'securityCenter', 'helpFeedback']
const rail = ['newTask', 'assistant', 'project', 'experts', 'automation', 'more']
</script>
<style scoped>
.workbuddy-guide { --client-accent: #0f9f78; }
.wb-account { max-width: 390px; margin: 0 auto; padding: .8rem; border-radius: 26px; background: var(--scene-panel); border: 1px solid var(--scene-line); }
.wb-popover { overflow: hidden; border-radius: 22px; background: var(--scene-bg); box-shadow: var(--av-shadow-md); }
.wb-userline, .wb-row, .wb-account-foot { display: flex; align-items: center; justify-content: space-between; gap: .6rem; }
.wb-userline { padding: 1.1rem 1.2rem; border-bottom: 1px solid var(--scene-line-soft); font-size: 1rem; }
.wb-row { margin: .35rem .6rem; padding: .55rem .45rem; font-size: .78rem; }
.wb-row em { color: var(--scene-muted); font-style: normal; font-size: .72rem; }
.wb-row b { border-radius: 8px; background: var(--scene-solid); color: var(--scene-solid-fg); padding: .35rem .7rem; font-weight: 600; }
.wb-row.is-current { border-radius: 10px; background: var(--scene-accent-soft); outline: 2px solid #34d399; }
.wb-station { margin: .4rem .7rem .6rem; overflow: hidden; border: 2px solid #34d399; border-radius: 18px; }
.wb-station-head { display: flex; align-items: center; gap: .45rem; padding: .45rem .7rem; background: #34d399; color: #022c22; font-weight: 700; }
.wb-station-head small { border-radius: 4px; background: var(--scene-accent-soft); color: var(--scene-accent-fg); padding: .05rem .3rem; font-size: .62rem; }
.wb-station strong, .wb-station p { display: block; margin: .55rem .8rem .2rem; }
.wb-station p { color: var(--scene-muted); font-size: .72rem; }
.wb-station > div:last-child { display: grid; grid-template-columns: 1fr 1fr; gap: .4rem; padding: .4rem .8rem .8rem; }
.wb-station > div:last-child span { border-radius: 8px; padding: .45rem; text-align: center; font-size: .72rem; background: var(--scene-solid); color: var(--scene-solid-fg); }
.wb-station > div:last-child span:last-child { background: var(--scene-bg); color: var(--scene-fg); border: 1px solid var(--scene-line); }
.wb-account-foot { padding: .8rem .4rem .2rem; }
.wb-account-foot i { display: grid; place-items: center; width: 36px; height: 36px; border-radius: 50%; background: #34d399; color: #022c22; font-style: normal; font-weight: 700; }
.wb-models { display: grid; grid-template-columns: 168px minmax(0, 1fr); min-height: 420px; overflow: hidden; border: 1px solid var(--scene-line); border-radius: 16px; background: var(--scene-bg); }
.wb-models aside { display: flex; flex-direction: column; gap: .15rem; padding: .7rem; background: var(--scene-panel); border-right: 1px solid var(--scene-line); }
.wb-models aside div { padding: .45rem .5rem; border-radius: 8px; color: var(--scene-muted); font-size: .72rem; }
.wb-models aside .current { background: var(--scene-chip); color: var(--scene-fg); font-weight: 700; }
.wb-models section { min-width: 0; padding: 1.1rem 1.2rem 1.3rem; }
.wb-page-title { display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid var(--scene-line); padding-bottom: .8rem; }
.wb-page-title h3 { margin: 0; font-size: 1.25rem; }
.wb-section { display: block; margin: 1rem 0 .45rem; font-size: .78rem; }
.model-file, .saved-model { display: flex; align-items: center; gap: .8rem; background: var(--scene-panel); border-radius: 8px; padding: .8rem; }
.model-file p, .saved-model small { margin: .25rem 0 0; color: var(--scene-muted); font-size: .68rem; }
.model-file code { color: var(--scene-info-fg); }
.add-model { margin-left: auto; border-radius: 8px; background: var(--scene-bg); padding: .45rem .7rem; font-weight: 700; box-shadow: 0 0 0 2px #34d399; white-space: nowrap; }
.saved-model strong, .saved-model small { display: block; }
.saved-model em { font-style: normal; color: var(--scene-muted); }
.workbuddy-edit { max-width: 640px; margin-left: auto; margin-right: auto; }
.advanced-title, .field-caption { display: block; margin-top: .9rem; font-size: .78rem; }
.cap-grid, .reason-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .55rem .7rem; margin-top: .65rem; }
.reason-grid { grid-template-columns: repeat(5, minmax(0, 1fr)); }
.cap-item { display: flex; align-items: center; gap: .35rem; min-width: 0; font-size: .72rem; }
.cap-item i { display: inline-grid; place-items: center; width: 14px; height: 14px; flex-shrink: 0; border-radius: 3px; font-style: normal; font-size: .62rem; }
.cap-item i.on { background: var(--scene-solid); color: var(--scene-solid-fg); }
.wb-chat { overflow: hidden; border: 1px solid var(--scene-line); border-radius: 16px; background: var(--scene-panel); }
.wb-menubar { display: flex; align-items: center; gap: .8rem; height: 32px; padding: 0 .7rem; background: var(--scene-panel); border-bottom: 1px solid var(--scene-line); font-size: .68rem; }
.wb-menubar em { margin-left: auto; font-style: normal; }
.workspace-conversation { display: flex; min-height: 430px; background: var(--scene-bg); }
.workspace-conversation aside { display: flex; flex-direction: column; gap: .35rem; width: 168px; flex-shrink: 0; padding: .8rem .6rem; background: var(--scene-panel); border-right: 1px solid var(--scene-line); font-size: .68rem; }
.rail-brand { display: flex; align-items: center; gap: .4rem; font-weight: 700; }
.rail-current { border-radius: 8px; background: var(--scene-chip); padding: .4rem .45rem; font-weight: 700; }
.workspace-conversation aside small { color: var(--scene-faint); }
.workspace-conversation main { flex: 1; min-width: 0; padding: 1rem 1.1rem; }
.chat-title { padding-bottom: .7rem; border-bottom: 1px solid var(--scene-line-soft); font-weight: 700; }
.user-bubble { width: fit-content; max-width: 82%; margin-left: auto; margin-top: 1rem; border-radius: 16px 16px 4px 16px; background: var(--scene-panel); padding: .55rem .8rem; }
.assistant-block { display: flex; gap: .7rem; margin-top: 1.2rem; }
.assistant-block small, .disclaimer { display: block; color: var(--scene-faint); font-size: .68rem; }
.composer { max-width: 560px; margin: 1.2rem auto 0; border: 1px solid var(--scene-line); border-radius: 20px; padding: .8rem; box-shadow: 0 0 0 2px #6ee7b780; }
.composer p { min-height: 2.4rem; margin: 0; color: var(--scene-faint); }
.composer > div { display: flex; align-items: center; gap: .55rem; }
.composer b { margin-left: auto; max-width: 46%; overflow-wrap: anywhere; border-radius: 8px; background: var(--scene-accent-soft); color: var(--scene-accent-fg); padding: .25rem .45rem; font-size: .72rem; }
.composer i { display: grid; place-items: center; width: 26px; height: 26px; border-radius: 50%; background: var(--scene-solid); color: var(--scene-solid-fg); font-style: normal; }
.disclaimer { margin-top: .6rem; text-align: center; }
@container help-scenes (max-width: 640px) {
  .wb-models, .workspace-conversation, .cap-grid, .reason-grid { grid-template-columns: minmax(0, 1fr); }
  .workspace-conversation { display: block; }
  .workspace-conversation aside { width: auto; flex-direction: row; flex-wrap: wrap; }
  .model-file, .saved-model { flex-wrap: wrap; }
}
</style>
