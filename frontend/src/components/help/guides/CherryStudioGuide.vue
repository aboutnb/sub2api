<template>
  <div class="client-specific-guide cherry-guide" data-guide="cherry-studio">
    <GuideStep :number="1" id="provider" :title="t('helpCenter.specific.cherryCreate')" :description="t('helpCenter.specific.cherryCreateHint', { provider })">
      <div class="guide-screen cherry-add">
        <div class="screen-top"><strong>{{ t('helpCenter.scenes.addProvider') }}</strong></div>
        <div class="screen-body">
          <div class="cherry-avatar" aria-hidden="true">{{ siteInitial }}</div>
          <span class="field-caption">{{ t('helpCenter.walkthrough.providerName') }}</span>
          <div class="name-value">{{ t('helpCenter.walkthrough.site') }}</div>
          <HelpSettingField :field="field('Provider')" :label="t('helpCenter.scenes.providerType')" choice />
        </div>
        <div class="screen-bottom"><span>{{ t('common.cancel') }}</span><span class="confirm">{{ t('common.confirm') }}</span></div>
      </div>
    </GuideStep>
    <GuideStep :number="2" id="credentials" :title="t('helpCenter.specific.cherryCredentials')" :description="t('helpCenter.specific.cherryCredentialsHint')">
      <div class="cherry-api">
        <div class="api-block">
          <div class="label-row"><strong>{{ t('helpCenter.scenes.apiKey') }}</strong><span aria-hidden="true">⋯</span></div>
          <div class="key-row">
            <HelpSettingField :field="field('API Key')" :label="t('helpCenter.scenes.apiKey')" />
            <span class="ghost">{{ t('helpCenter.cherryUi.detect') }}</span>
          </div>
          <p class="fine">{{ t('helpCenter.cherryUi.multipleKeys') }}</p>
        </div>
        <div class="api-block">
          <strong>{{ t('helpCenter.scenes.apiAddress') }}</strong>
          <HelpSettingField :field="field('API Host')" :label="t('helpCenter.scenes.apiAddress')" />
          <p class="fine">{{ t('helpCenter.cherryUi.preview', { path: previewPath }) }}</p>
        </div>
        <div class="model-tools">
          <strong>{{ t('helpCenter.model') }} <em>0</em></strong>
          <span class="ghost">↻ {{ t('helpCenter.scenes.fetchModels') }}</span>
        </div>
      </div>
    </GuideStep>
    <GuideStep :number="3" id="catalog" :title="t('helpCenter.specific.cherryCatalog')" :description="t('helpCenter.specific.cherryCatalogHint')">
      <div class="cherry-models">
        <div class="models-title"><strong>{{ t('helpCenter.cherryUi.listTitle', { name: t('helpCenter.walkthrough.site') }) }}</strong><span aria-hidden="true">×</span></div>
        <div class="search-row"><span class="search">⌕ {{ t('helpCenter.cherryUi.search') }}</span><span aria-hidden="true">↻</span><span aria-hidden="true">▾</span></div>
        <div class="model-filters">
          <span class="active">{{ t('helpCenter.scenes.allModels') }}</span>
          <span>{{ t('helpCenter.scenes.reasoning') }}</span>
          <span>{{ t('helpCenter.scenes.vision') }}</span>
          <span>{{ t('helpCenter.cherryUi.tabWeb') }}</span>
          <span>{{ t('helpCenter.cherryUi.tabFree') }}</span>
          <span>{{ t('helpCenter.cherryUi.tabEmbedding') }}</span>
          <span>{{ t('helpCenter.cherryUi.tabRerank') }}</span>
          <span>{{ t('helpCenter.scenes.tools') }}</span>
        </div>
        <HelpSettingField v-if="modelField" :field="modelField" :label="t('helpCenter.model')" />
      </div>
    </GuideStep>
    <GuideStep :number="4" id="chat" :title="t('helpCenter.specific.cherryChat')" :description="t('helpCenter.specific.cherryChatHint')">
      <div class="cherry-reply">
        <div class="reply-bar"><span class="face" aria-hidden="true">:)</span><strong>{{ t('helpCenter.cherryUi.defaultAssistant') }}</strong><span class="model-pill">{{ model }}</span></div>
        <p class="intro">{{ t('helpCenter.cherryUi.defaultMessage') }}</p>
        <div class="turn"><i>{{ t('helpCenter.cherryUi.user') }}</i><div><small>hi</small><p>{{ t('helpCenter.cherryUi.hi') }}</p></div></div>
        <div class="turn assistant"><i>{{ providerInitial }}</i><div><small>{{ t('helpCenter.cherryUi.thinking') }}</small><p>{{ t('helpCenter.cherryUi.reply') }}</p><span class="text-action">{{ t('helpCenter.cherryUi.textAction') }}</span></div></div>
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
const { field, modelField, model, provider } = useGuideFields(props)
const previewPath = computed(() => provider.value === 'Anthropic' ? '/v1/messages' : '/chat/completions')
const siteInitial = computed(() => t('helpCenter.walkthrough.site').trim().charAt(0).toUpperCase() || 'A')
const providerInitial = computed(() => (provider.value || 'M').charAt(0).toUpperCase())
</script>
<style scoped>
.cherry-guide { --client-accent: #16a34a; }
.cherry-add { max-width: 360px; margin-left: auto; margin-right: auto; }
.cherry-avatar { display: grid; place-items: center; width: 54px; height: 54px; margin: 0 auto .8rem; border-radius: 50%; background: var(--scene-solid); color: var(--scene-solid-fg); font-weight: 750; }
.field-caption { display: block; margin-bottom: .35rem; font-size: .78rem; }
.name-value { margin-bottom: .8rem; border: 2px solid #22c55e; border-radius: 8px; padding: .55rem .7rem; }
.confirm { border-radius: 8px; background: #16a34a; color: #022c22; padding: .35rem .9rem; }
.cherry-api, .cherry-models, .cherry-reply { max-width: 460px; margin: 0 auto; overflow: hidden; border: 1px solid var(--scene-line); border-radius: 16px; background: var(--scene-bg); box-shadow: var(--scene-shadow); }
.cherry-api, .cherry-models { padding: 1rem 1.1rem 1.15rem; }
.api-block { margin-bottom: 1rem; }
.label-row, .key-row, .model-tools, .models-title, .search-row { display: flex; align-items: center; justify-content: space-between; gap: .6rem; }
.key-row { align-items: flex-end; }
.key-row :deep(.scene-field) { flex: 1; min-width: 0; }
.ghost { border: 1px solid var(--scene-line); border-radius: 8px; padding: .4rem .65rem; color: var(--scene-muted); font-size: .75rem; white-space: nowrap; }
.fine { margin: .35rem 0 0; color: var(--scene-faint); font-size: .72rem; text-align: right; }
.model-tools { margin-top: .2rem; }
.model-tools em { border-radius: 999px; background: var(--scene-panel); color: var(--scene-muted); font-style: normal; font-size: .68rem; padding: .05rem .4rem; }
.models-title { margin-bottom: .7rem; }
.search { flex: 1; border: 1px solid var(--scene-line); border-radius: 8px; padding: .5rem .7rem; color: var(--scene-faint); }
.search-row > span:not(.search) { display: grid; place-items: center; width: 34px; height: 34px; border: 1px solid var(--scene-line); border-radius: 8px; }
.model-filters { display: flex; gap: .8rem; overflow-x: auto; margin: .75rem 0; color: var(--scene-muted); font-size: .75rem; }
.model-filters .active { color: var(--scene-accent-fg); font-weight: 700; border-bottom: 2px solid #22c55e; padding-bottom: .3rem; }
.cherry-reply { max-width: 640px; }
.reply-bar { display: flex; align-items: center; gap: .5rem; padding: .7rem 1rem; border-bottom: 1px solid var(--scene-line-soft); }
.face { display: grid; place-items: center; width: 24px; height: 24px; border-radius: 50%; background: var(--scene-warm-soft); font-size: .68rem; }
.model-pill { border-radius: 999px; background: var(--scene-info-soft); color: var(--scene-info-fg); padding: .15rem .5rem; font-size: .75rem; }
.intro { margin: .8rem 1rem; border-radius: 12px; background: var(--scene-panel); padding: .6rem .8rem; color: var(--scene-muted); }
.turn { display: flex; gap: .7rem; padding: 0 1rem .9rem; }
.turn i { display: grid; place-items: center; width: 36px; height: 36px; flex-shrink: 0; border-radius: 50%; background: #14b8a6; color: #042f2e; font-style: normal; font-size: .72rem; }
.turn.assistant i { background: #3b82f6; }
.turn small { color: var(--scene-faint); }
.turn p { margin: .2rem 0; }
.text-action { display: inline-grid; place-items: center; width: 22px; height: 22px; border: 1px solid var(--scene-line); border-radius: 6px; font-size: .72rem; }
.cherry-guide :deep(.field-input) { border-radius: 8px; background: var(--scene-panel); }
</style>
