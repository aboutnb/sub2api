<template><img :src="'/client-logos/' + asset" alt="" width="22" height="22" class="client-logo" aria-hidden="true"></template>
<script setup lang="ts">
import { computed } from 'vue'
const props = defineProps<{ client: string }>()
const assets: Record<string, string> = {
  claude: 'claudecode-color.svg', codex: 'codex-color.svg', 'codex-ws': 'codex-color.svg', gemini: 'geminicli-color.svg', grok: 'grok.svg', opencode: 'opencode.svg',
  'cherry-studio': 'cherrystudio-color.svg', cursor: 'cursor.svg', cline: 'cline.svg', 'roo-code': 'roocode.svg',
  dsh: 'deepseek-color.svg', pi: 'pi.svg', openclaw: 'openclaw-color.svg', hermes: 'hermesagent.svg', workbuddy: 'workbuddy.svg', zcode: 'zcode.png', trae: 'trae-color.svg',
  'read-frog': 'read-frog.png', 'kiss-translator': 'kiss-translator.png', 'immersive-translate': 'immersive-translate.svg',
  sillytavern: 'sillytavern.svg', tavernai: 'tavernai.svg'
}
const asset = computed(() => assets[props.client] || 'opencode.svg')
</script>
<style scoped>.client-logo { display: inline-block; flex: 0 0 22px; width: 22px; height: 22px; object-fit: contain; vertical-align: middle; background: #fff; border-radius: 4px; padding: 2px; }</style>
