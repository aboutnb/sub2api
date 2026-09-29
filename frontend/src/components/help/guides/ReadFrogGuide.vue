<template>
  <div class="client-specific-guide frog-guide" data-guide="read-frog">
    <GuideStep :number="1" id="provider-settings" :title="t('helpCenter.specific.frogProvider')" :description="t('helpCenter.specific.frogProviderHint', { provider })">
      <div class="guide-screen frog-settings">
        <div class="frog-page-head"><strong>API Providers</strong><p>Configure the provider for this channel. Built-in providers stay available.</p></div>
        <div class="screen-columns">
          <aside class="screen-nav">
            <div class="add-provider">+ Add Provider</div>
            <div class="provider-row current">
              <span class="provider-mark" aria-hidden="true">A</span>
              <span class="provider-name">{{ provider }}</span>
              <span class="mini-switch is-on" aria-hidden="true"><i /></span>
            </div>
          </aside>
          <div class="screen-body">
            <div class="provider-heading">
              <span><span class="provider-mark" aria-hidden="true">A</span><strong>{{ provider }}</strong></span>
              <span class="docs-link">How to configure?</span>
            </div>
            <HelpSettingField :field="field('Provider')" label="Name" />
            <span class="field-caption">Description</span>
            <div class="screen-value muted">{{ providerDescription }}</div>
            <div class="label-row">
              <span>API Key</span>
              <span class="illustrated-action">Test Connection</span>
            </div>
            <HelpSettingField :field="field('API Key')" />
            <div class="check-line"><span class="check on" aria-hidden="true">✓</span>Show API Key</div>
            <span class="field-caption">Base URL <em>(Optional)</em></span>
            <HelpSettingField :field="field('Base URL')" :hint="t('helpCenter.specific.frogBaseHint')" />
            <div class="label-row">
              <span>Model</span>
              <span class="gear" aria-hidden="true">⚙</span>
            </div>
            <HelpSettingField v-if="modelField" :field="modelField" label="Model" />
            <div class="check-line"><span class="check on" aria-hidden="true">✓</span>Enter the name of the custom model</div>
            <div class="collapsed-row">▸ Feature Providers</div>
            <div class="collapsed-row">▸ Advanced Options</div>
          </div>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="2" id="connection-test" :title="t('helpCenter.specific.frogTest')" :description="t('helpCenter.specific.frogTestHint')">
      <div class="guide-screen">
        <div class="screen-body">
          <span class="illustrated-action">Test Connection</span>
          <p class="screen-hint">{{ t('helpCenter.scenes.testNotRun') }}</p>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="3" id="page-translate" :title="t('helpCenter.specific.frogFeature')" :description="t('helpCenter.specific.frogFeatureHint')">
      <HelpConversationScene client-id="read-frog" client-name="Read Frog" :model="model" :provider="provider" />
    </GuideStep>
  </div>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GuideStep from './GuideStep.vue'
import HelpSettingField from '../HelpSettingField.vue'
import HelpConversationScene from '../HelpConversationScene.vue'
import { useGuideFields, type ClientGuideProps } from './guideFields'
const props = defineProps<ClientGuideProps>()
const { t } = useI18n()
const { field, model, modelField, provider } = useGuideFields(props)
const providerDescription = computed(() => provider.value === 'Anthropic'
  ? 'Claude models known for safety and reasoning excellence'
  : 'OpenAI-compatible models for this channel')
</script>
<style scoped>
.frog-guide { --client-accent: #0f766e; }
.frog-settings { max-width: 760px; margin-left: auto; margin-right: auto; }
.frog-page-head { padding: .9rem 1.1rem .2rem; }
.frog-page-head p { margin: .25rem 0 .4rem; color: var(--scene-faint); font-size: .72rem; line-height: 1.6; }
.add-provider { margin: .2rem .35rem .7rem; border: 1px solid var(--scene-line); border-radius: 999px; padding: .45rem .6rem; text-align: center; color: var(--scene-muted); font-size: .72rem; }
.provider-row { display: flex; align-items: center; gap: .45rem; padding: .55rem .6rem; border-radius: 10px; border-left: 3px solid transparent; }
.provider-row.current { border-left-color: #14b8a6; background: var(--scene-accent-soft); }
.provider-mark { display: inline-grid; place-items: center; width: 22px; height: 22px; border-radius: 50%; background: #d4a27f; color: #3f2a1d; font-size: .72rem; font-weight: 700; }
.provider-name { min-width: 0; overflow-wrap: anywhere; font-size: .75rem; font-weight: 650; }
.mini-switch { width: 28px; height: 16px; margin-left: auto; border-radius: 999px; background: var(--scene-chip); padding: 2px; display: flex; }
.mini-switch.is-on { justify-content: flex-end; background: #14b8a6; }
.mini-switch i { width: 12px; height: 12px; border-radius: 50%; background: white; }
.provider-heading, .label-row { display: flex; align-items: center; justify-content: space-between; gap: .75rem; }
.provider-heading span:first-child, .provider-heading strong { display: inline-flex; align-items: center; gap: .45rem; }
.docs-link { color: var(--scene-info-fg); font-size: .72rem; }
.field-caption { display: block; margin-top: .85rem; font-size: .78rem; font-weight: 650; }
.field-caption em { color: var(--scene-faint); font-style: normal; font-weight: 500; }
.muted { color: var(--scene-muted); }
.check-line, .collapsed-row { display: flex; align-items: center; gap: .4rem; margin: .45rem 0; color: var(--scene-muted); font-size: .75rem; }
.check { display: inline-grid; place-items: center; width: 14px; height: 14px; border-radius: 3px; font-size: .65rem; }
.check.on { background: #14b8a6; color: #042f2e; }
.gear { color: #14b8a6; }
.frog-guide :deep(.field-input) { border: 2px solid #5eead4; background: var(--scene-accent-soft); }
@media (max-width: 540px) { .provider-heading, .label-row { align-items: flex-start; flex-direction: column; } }
</style>
