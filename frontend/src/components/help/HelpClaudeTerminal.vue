<template>
  <figure class="claude-terminal-example" :aria-label="t('helpCenter.claudeTerminal.title')">
    <figcaption><strong>{{ t('helpCenter.claudeTerminal.title') }}</strong><span>{{ t('helpCenter.walkthrough.exampleOnly') }}</span></figcaption>
    <p class="terminal-instruction">{{ t('helpCenter.claudeTerminal.instruction') }}</p>
    <div class="claude-terminal">
      <div class="terminal-titlebar"><div class="traffic-lights" aria-hidden="true"><i /><i /><i /></div><span>{{ t('helpCenter.claudeTerminal.terminal') }}</span></div>
      <div class="terminal-body">
        <div class="launch-command"><span class="muted">{{ projectPath }} &gt;</span> <strong>claude</strong></div>
        <div class="terminal-divider"><span>Claude Code</span><small>{{ t('helpCenter.claudeTerminal.versionHint') }}</small></div>
        <div class="welcome-panel">
          <div class="welcome-identity">
            <strong>Welcome back!</strong>
            <div class="claude-mascot" aria-hidden="true"><i v-for="(pixel, index) in pixels" :key="index" :class="{ body: pixel === '1', eye: pixel === '2' }" /></div>
            <div class="model-summary"><span class="terminal-model">{{ displayModel }}</span> · API Usage Billing</div>
            <div class="project-directory">{{ projectPath }}</div>
          </div>
          <div class="welcome-tips">
            <strong>Tips for getting started</strong>
            <p>Run /init to create a CLAUDE.md file with instructions for Claude</p>
            <hr>
            <strong>Recent activity</strong>
            <p>No recent activity</p>
          </div>
        </div>
        <div class="model-status">Using <span class="terminal-model">{{ displayModel }}</span> <span>{{ t('helpCenter.claudeTerminal.' + (configMode === 'terminal' ? 'environmentSource' : 'fileSource')) }}</span> · /model to change</div>
        <div class="terminal-message user-prompt"><span aria-hidden="true">❯</span><strong>hi</strong></div>
        <div class="terminal-message assistant-reply"><span class="reply-dot" aria-hidden="true">●</span><span>Hey! What can I help you with?</span></div>
        <div class="terminal-message input-prompt" aria-hidden="true"><span>❯</span><i class="terminal-cursor" /></div>
        <footer>? for shortcuts</footer>
      </div>
    </div>
    <p class="terminal-disclaimer">{{ t('helpCenter.claudeTerminal.disclaimer') }}</p>
  </figure>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ model: string; os: string; shell: string; configMode: 'file' | 'terminal' }>()
const { t } = useI18n()
const displayModel = computed(() => props.model || 'YOUR_MODEL_ID')
const projectPath = computed(() => props.os === 'windows' && props.shell !== 'bash' ? 'C:\\projects\\my-app' : '~/projects/my-app')
// Eight-column pixel mascot from the supplied terminal reference.
const pixels = ['00111100', '01111110', '11211211', '11111111', '01111110', '10111101', '10011001'].join('').split('')
</script>

<style scoped>
.claude-terminal-example { margin: 1.5rem 0; min-width: 0; }
figcaption { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: .5rem; font-size: .85rem; }
figcaption > span { font-size: .7rem; border: 1px solid var(--scene-line); border-radius: 4px; padding: .2rem .5rem; }
.terminal-instruction, .terminal-disclaimer { font-size: .8rem; line-height: 1.8; margin: .75rem 0; }
.claude-terminal { --claude-accent: #b36440; max-width: 42rem; margin: 1rem auto; border: 1px solid var(--scene-line); border-radius: 12px; overflow: hidden; background: var(--scene-bg); color: var(--scene-fg); box-shadow: var(--scene-shadow); font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', monospace; font-size: 12px; line-height: 1.65; }
:global(:root.dark) .claude-terminal { --claude-accent: #e7b5a4; }
.terminal-titlebar { display: flex; align-items: center; justify-content: space-between; gap: .75rem; padding: .65rem 1rem; border-bottom: 1px solid var(--scene-line-soft); }
.terminal-titlebar > span { font-size: 10px; color: var(--scene-muted); }
.traffic-lights { display: flex; gap: 8px; }
.traffic-lights i { width: 12px; height: 12px; border-radius: 50%; background: #ff5f57; }
.traffic-lights i:nth-child(2) { background: #febc2e; }
.traffic-lights i:nth-child(3) { background: #28c840; }
.terminal-body { padding: 1.25rem; }
.launch-command, .model-status, .model-summary, .project-directory { overflow-wrap: anywhere; }
.muted { color: var(--scene-muted); }
.terminal-divider { display: flex; align-items: center; flex-wrap: wrap; gap: .5rem; margin-top: .75rem; }
.terminal-divider::before, .terminal-divider::after { content: ''; flex: 1; border-top: 1px dashed var(--scene-line); }
.terminal-divider > span { color: var(--claude-accent); }
.terminal-divider > small { color: var(--scene-muted); font-size: 10px; }
.welcome-panel { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); border: 1px dashed var(--claude-accent); border-radius: 4px; margin-top: .75rem; }
.welcome-identity { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 1.25rem; border-right: 1px dashed var(--claude-accent); text-align: center; }
.welcome-identity > strong { font-size: 13px; }
.claude-mascot { display: grid; grid-template-columns: repeat(8, 6px); margin: 1rem 0; }
.claude-mascot i { height: 6px; width: 6px; }
.claude-mascot .body { background: #c2714f; }
.claude-mascot .eye { background: #1a1a1a; }
.model-summary, .project-directory { font-size: 11px; color: var(--scene-muted); max-width: 100%; }
.welcome-tips { padding: 1rem; font-size: 11px; line-height: 1.8; }
.welcome-tips strong { color: var(--claude-accent); }
.welcome-tips p { margin-top: .4rem; }
.welcome-tips hr { margin: .75rem 0; border: 0; border-top: 1px dashed var(--claude-accent); }
.model-status { margin: .75rem 0 .4rem; font-size: 11px; color: var(--scene-muted); }
.model-status .terminal-model { color: var(--scene-fg); }
.terminal-message { display: flex; align-items: flex-start; gap: .5rem; padding: .5rem 0; }
.user-prompt, .input-prompt { border-top: 1px solid var(--scene-line-soft); }
.reply-dot { color: var(--claude-accent); }
.terminal-cursor { display: inline-block; width: 7px; height: 14px; margin-top: 3px; background: var(--scene-fg); animation: terminal-blink 1.2s step-end infinite; }
footer { margin-top: .75rem; padding-top: .5rem; border-top: 1px dotted var(--scene-line); font-size: 10px; color: var(--scene-muted); }
@keyframes terminal-blink { 50% { opacity: 0; } }
@media (prefers-reduced-motion: reduce) { .terminal-cursor { animation: none; } }
@media (max-width: 540px) {
  .terminal-body { padding: .85rem; }
  .welcome-panel { grid-template-columns: minmax(0, 1fr); }
  .welcome-identity { border-right: 0; border-bottom: 1px dashed var(--claude-accent); }
  .terminal-divider { gap: .35rem; }
  .terminal-divider > small { flex-basis: 100%; text-align: center; order: 1; }
}
</style>
