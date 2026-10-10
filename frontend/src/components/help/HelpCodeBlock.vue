<template>
  <div class="help-code">
    <div class="help-code-bar">
      <span v-if="displayLabel">{{ displayLabel }}</span>
      <span v-else class="help-code-spacer" />
      <button type="button" :aria-label="copyLabel" @click="copyToClipboard(content, t('helpCenter.copied'))">{{ t('helpCenter.copyConfig') }}</button>
    </div>
    <pre class="shiki github-light-default github-dark-default" translate="no"><code class="shiki" v-html="highlighted" /></pre>
  </div>
</template>
<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useClipboard } from '@/composables/useClipboard'
import { highlightGithubCode, onGithubHighlighter } from '@/utils/githubCode'

const props = defineProps<{ content: string; label?: string; language?: string }>()
const { t } = useI18n()
const { copyToClipboard } = useClipboard()
const displayLabel = computed(() => {
  const label = props.label?.trim()
  return label && label !== 'code' ? label : ''
})
const copyLabel = computed(() => displayLabel.value ? `${t('helpCenter.copyConfig')} ${displayLabel.value}` : t('helpCenter.copyConfig'))
const revision = ref(0)
const stop = onGithubHighlighter(() => {
  revision.value += 1
})
onUnmounted(stop)
const highlighted = computed(() => {
  void revision.value
  return highlightGithubCode(props.content, props.language || props.label)
})
</script>
<style scoped>
.help-code {
  min-width: 0;
  overflow: hidden;
  margin: 1rem 0;
  border: 1px solid var(--gh-line);
  border-radius: 6px;
  background: var(--gh-bg);
  color: var(--gh-fg);
  font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace;
}
.help-code-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  min-height: 2.5rem;
  padding: 0.4rem 0.75rem;
  border-bottom: 1px solid var(--gh-line);
  background: var(--gh-bar);
  color: var(--gh-muted);
  font-size: 12px;
  line-height: 1.4;
}
.help-code-bar span { min-width: 0; overflow-wrap: anywhere; }
.help-code-spacer { flex: 1; }
.help-code-bar button {
  flex-shrink: 0;
  min-height: 2rem;
  padding: 0.2rem 0.7rem;
  border: 1px solid var(--gh-line);
  border-radius: 6px;
  background: var(--gh-btn);
  color: var(--gh-fg);
  font-size: 12px;
  line-height: 1.4;
}
.help-code-bar button:hover { background: var(--gh-btn-hover); }
.help-code-bar button:focus-visible { outline: 2px solid var(--gh-focus); outline-offset: 2px; }
pre {
  margin: 0;
  padding: 16px;
  overflow-x: auto;
  background: var(--gh-bg);
  color: var(--gh-fg);
  font-size: 12px;
  line-height: 1.5;
  tab-size: 2;
}
code { font-family: inherit; }
@media (pointer: coarse) {
  .help-code-bar button { min-height: 2.75rem; min-width: 2.75rem; }
}
</style>

<style>
.help-code {
  --gh-bg: #ffffff;
  --gh-fg: #1f2328;
  --gh-bar: #f6f8fa;
  --gh-muted: #59636e;
  --gh-line: #d0d7de;
  --gh-btn: #f6f8fa;
  --gh-btn-hover: #f3f4f6;
  --gh-focus: #0969da;
  color-scheme: light;
}
:root.dark .help-code,
.dark .help-code {
  --gh-bg: #0d1117;
  --gh-fg: #e6edf3;
  --gh-bar: #161b22;
  --gh-muted: #8b949e;
  --gh-line: #30363d;
  --gh-btn: #21262d;
  --gh-btn-hover: #30363d;
  --gh-focus: #4493f8;
  color-scheme: dark;
}
.help-code pre.shiki {
  --shiki-light: #1f2328;
  --shiki-dark: #e6edf3;
  --shiki-light-bg: #ffffff;
  --shiki-dark-bg: #0d1117;
  background: var(--gh-bg);
  color: var(--gh-fg);
}
.help-code pre.shiki span {
  color: var(--shiki-light);
  font-style: var(--shiki-light-font-style);
  font-weight: var(--shiki-light-font-weight);
  text-decoration: var(--shiki-light-text-decoration);
}
:root.dark .help-code pre.shiki span,
.dark .help-code pre.shiki span {
  color: color-mix(in srgb, var(--shiki-dark, var(--shiki-light)) 80%, #ffffff);
  font-style: var(--shiki-dark-font-style);
  font-weight: var(--shiki-dark-font-weight);
  text-decoration: var(--shiki-dark-text-decoration);
}
.help-code pre.shiki .line { display: inline; }
</style>
