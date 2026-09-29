<template>
  <div class="space-y-4">
    <nav v-if="showToc && rendered.headings.length" :aria-label="t('helpCenter.onThisPage')" class="flex flex-wrap gap-x-4 gap-y-2 border-b border-line pb-4">
      <a v-for="heading in rendered.headings" :key="heading.id" :href="`#${heading.id}`" class="text-sm text-primary-600">{{ heading.text }}</a>
    </nav>
    <template v-for="(part, index) in parts" :key="index">
      <HelpCodeBlock v-if="part.code !== undefined" :content="part.code" :label="part.language" />
      <div v-else class="help-markdown min-w-0" v-html="part.html" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import HelpCodeBlock from './HelpCodeBlock.vue'
import { renderHelpMarkdown } from '@/utils/helpMarkdown'
const props = withDefaults(defineProps<{ content: string; showToc?: boolean }>(), { showToc: true })
const { t } = useI18n()
const rendered = computed(() => renderHelpMarkdown(props.content))
const parts = computed(() => rendered.value.html.split(/(<pre\b[\s\S]*?<\/pre>)/i).filter(Boolean).map(html => {
  if (!html.startsWith('<pre')) return { html }
  const template = document.createElement('template')
  template.innerHTML = html
  const code = template.content.querySelector('code')
  return { code: template.content.textContent || '', language: code?.className.replace('language-', '') || 'code' }
}))
</script>

<style scoped>
.help-markdown :deep(h1), .help-markdown :deep(h2), .help-markdown :deep(h3) { font-weight: 650; margin: 1.6em 0 .7em; scroll-margin-top: 6rem; }
.help-markdown :deep(h2) { font-size: 1.3rem; }
.help-markdown :deep(p), .help-markdown :deep(li) { line-height: 1.8; overflow-wrap: anywhere; }
.help-markdown :deep(p) { margin-bottom: 1em; }
.help-markdown :deep(ul), .help-markdown :deep(ol) { padding-left: 1.5rem; list-style: revert; }
.help-markdown :deep(a) { text-decoration: underline; }
.help-markdown :deep(img) { max-width: 100%; height: auto; }
.help-markdown :deep(table) { width: 100%; border-collapse: collapse; display: block; overflow-x: auto; font-size: .875rem; }
.help-markdown :deep(th), .help-markdown :deep(td) { border-bottom: 1px solid var(--av-line); padding: .7rem .75rem; text-align: left; vertical-align: top; line-height: 1.7; }
.help-markdown :deep(th) { color: var(--av-color-text-muted); font-size: .75rem; font-weight: 650; }
.help-markdown :deep(code) { padding: .2em .4em; border-radius: 6px; border: 1px solid var(--av-color-border-strong); background: var(--av-color-bg-surface-muted); color: var(--av-color-text-strong); font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace; font-size: 85%; }
</style>
