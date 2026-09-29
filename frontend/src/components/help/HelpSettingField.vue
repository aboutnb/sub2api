<template>
  <div class="scene-field" :data-field="field.path">
    <div class="field-label"><span>{{ label || field.path }}</span><span v-if="toggle" class="scene-switch" :class="{ 'is-on': field.content === 'ON' }" aria-hidden="true"><i /></span></div>
    <div class="field-input"><code>{{ field.content }}</code><span v-if="choice" aria-hidden="true">⌄</span><button type="button" :aria-label="t('helpCenter.copyConfig') + ' ' + field.path" @click="copyToClipboard(field.content, t('helpCenter.copied'))">{{ t('helpCenter.copyConfig') }}</button></div>
    <small v-if="hint">{{ hint }}</small>
  </div>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useClipboard } from '@/composables/useClipboard'
defineProps<{ field: { path: string; content: string }; label?: string; hint?: string; toggle?: boolean; choice?: boolean }>()
const { t } = useI18n()
const { copyToClipboard } = useClipboard()
</script>
<style scoped>
.scene-field { min-width: 0; margin: 1rem 0; }
.field-label { display: flex; align-items: center; justify-content: space-between; gap: .75rem; font-size: .78rem; font-weight: 600; margin-bottom: .45rem; }
.field-input { display: flex; align-items: center; gap: .65rem; border: 1px solid var(--scene-line); border-radius: 7px; background: var(--scene-input); padding: .65rem .8rem; }
code { flex: 1; min-width: 0; font-size: .76rem; overflow-wrap: anywhere; white-space: pre-wrap; }
button { flex-shrink: 0; padding: .25rem; color: var(--client-accent-ink, var(--av-color-brand-primary)); font-size: .7rem; border-radius: 4px; }
button:hover { background: color-mix(in srgb, var(--scene-muted) 14%, transparent); }
button:focus-visible { outline: 2px solid currentColor; outline-offset: 2px; }
small { display: block; color: var(--scene-muted); font-size: .7rem; line-height: 1.7; margin-top: .4rem; }
.scene-switch { display: flex; width: 30px; padding: 3px; border-radius: 20px; background: var(--scene-track); }
.scene-switch.is-on { background: var(--client-accent, #197363); justify-content: flex-end; }
.scene-switch i { height: 12px; width: 12px; background: white; border-radius: 50%; }
</style>
