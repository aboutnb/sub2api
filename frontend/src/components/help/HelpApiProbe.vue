<template>
  <section class="api-probe" :aria-label="t('helpCenter.api')">
    <div class="probe-grid">
      <div class="probe-field">
        <label for="help-api-group">{{ t('helpCenter.group') }}</label>
        <select id="help-api-group" v-model="groupId" class="input">
          <option v-if="!groups.length" value="">{{ t('helpCenter.choose') }}</option>
          <option v-for="item in groups" :key="item.id" :value="String(item.id)">{{ item.name }}</option>
        </select>
      </div>
      <div class="probe-field">
        <label for="help-api-model">{{ t('helpCenter.model') }}</label>
        <select id="help-api-model" v-model="selectedModel" class="input" :disabled="modelsLoading || !modelOptions.length">
          <option v-if="!modelOptions.length" value="">{{ t(modelsLoading ? 'common.loading' : 'helpCenter.noModels') }}</option>
          <option v-for="model in modelOptions" :key="model" :value="model">{{ model }}</option>
        </select>
      </div>
      <div class="probe-field">
        <label for="help-api-key">{{ t('helpCenter.probe.key') }}</label>
        <select id="help-api-key" v-model="keyId" class="input" :disabled="keysLoading || !keys.length">
          <option v-if="!keys.length" value="">{{ t(keysLoading ? 'common.loading' : 'helpCenter.probe.noKey') }}</option>
          <option v-for="item in keys" :key="item.id" :value="String(item.id)">{{ item.name }} · {{ maskApiKey(item.key) }}</option>
        </select>
      </div>
    </div>
    <p v-if="!groups.length" class="probe-note">{{ t('helpCenter.noGroups') }} <RouterLink to="/keys">{{ t('helpCenter.openKeys') }}</RouterLink></p>
    <p v-else-if="!keysLoading && !keysError && !keys.length" class="probe-note">{{ t('helpCenter.probe.noKey') }} <RouterLink to="/keys">{{ t('helpCenter.openKeys') }}</RouterLink></p>
    <p v-if="keyTotal > keys.length" class="probe-note">{{ t('helpCenter.probe.keyLimit') }}</p>
    <button v-if="modelsError || keysError" type="button" class="btn btn-secondary" @click="reload">{{ t('helpCenter.retry') }}</button>
    <nav v-if="choice.protocols.length > 1" class="probe-protocols" :aria-label="t('helpCenter.probe.protocol')">
      <button v-for="item in choice.protocols" :key="item" type="button" :aria-pressed="protocol === item" :class="{ selected: protocol === item }" @click="chooseProtocol(item)">{{ t(`helpCenter.probe.${item}`) }}</button>
    </nav>
    <p v-if="choice.protocols.length > 1 && !protocol" class="probe-note">{{ t('helpCenter.probe.chooseProtocol') }}</p>
    <div class="probe-actions">
      <button type="button" class="btn btn-primary" :disabled="!sending && !canSend" @click="send">{{ sending ? t('helpCenter.probe.stop') : t('helpCenter.probe.send') }}</button>
      <p>{{ t('helpCenter.probe.quota') }}</p>
    </div>
    <HelpCodeBlock v-if="previewView" :content="previewView.preview" :label="previewView.label" language="shell" />
    <div v-if="result" class="probe-result" aria-live="polite">
      <div class="probe-meta">
        <span v-if="result.kind === 'http'">{{ t('helpCenter.probe.status') }} <strong>{{ result.status }}</strong></span>
        <span v-else>{{ kindLabel }}</span>
        <span>{{ t('helpCenter.probe.latency') }} <strong>{{ result.latencyMs }} ms</strong></span>
        <span v-if="result.requestId">{{ t('helpCenter.probe.requestId') }} <strong>{{ result.requestId }}</strong></span>
      </div>
      <pre v-if="result.body" class="probe-body">{{ result.body }}</pre>
      <p v-if="result.truncated" class="probe-note">{{ t('helpCenter.probe.truncated') }}</p>
    </div>
    <p v-else class="probe-note">{{ t('helpCenter.probe.idle') }}</p>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import HelpCodeBlock from './HelpCodeBlock.vue'
import { keysAPI } from '@/api/keys'
import { userGroupsAPI } from '@/api/groups'
import type { ApiKey, Group } from '@/types'
import { sortHelpModels } from '@/utils/helpModels'
import { HELP_API_TIMEOUT_MS, buildHelpApiRequest, helpApiProtocols, maskApiKey, readHelpApiBody, type HelpApiProtocol } from '@/utils/helpApiProbe'

const props = defineProps<{ groups: Group[]; baseUrl: string }>()
const { t } = useI18n()
const groupId = ref('')
const modelOptions = ref<string[]>([])
const selectedModel = ref('')
const modelsLoading = ref(false)
const modelsError = ref(false)
const keys = ref<ApiKey[]>([])
const keyTotal = ref(0)
const keyId = ref('')
const keysLoading = ref(false)
const keysError = ref(false)
const protocol = ref<HelpApiProtocol | null>(null)
const sending = ref(false)
const result = ref<ProbeView | null>(null)
let loadToken = 0
let loadAbort: AbortController | undefined
let probeToken = 0
let probeAbort: AbortController | undefined

interface ProbeView {
  kind: 'http' | 'timeout' | 'aborted' | 'network'
  status?: number
  latencyMs: number
  requestId?: string
  body?: string
  truncated?: boolean
}

const group = computed(() => props.groups.find(item => String(item.id) === groupId.value))
const choice = computed(() => group.value ? helpApiProtocols(group.value, selectedModel.value) : { protocols: [], defaultProtocol: null })
const selectedKey = computed(() => keys.value.find(item => String(item.id) === keyId.value))
const canSend = computed(() => Boolean(group.value && selectedModel.value && selectedKey.value && protocol.value && !modelsLoading.value && !keysLoading.value && !modelsError.value && !keysError.value))
const previewView = computed(() => {
  if (!group.value || !protocol.value || !selectedModel.value) return null
  const request = buildHelpApiRequest({ baseUrl: props.baseUrl, platform: group.value.platform, protocol: protocol.value, model: selectedModel.value, apiKey: selectedKey.value?.key || '' })
  return { preview: request.preview, label: requestLabel(request.url) }
})
const kindLabel = computed(() => {
  if (result.value?.kind === 'timeout') return t('helpCenter.probe.timeout')
  if (result.value?.kind === 'aborted') return t('helpCenter.probe.aborted')
  if (result.value?.kind === 'network') return t('helpCenter.probe.network')
  return ''
})

function requestLabel(url: string) {
  try { return `POST ${new URL(url, 'https://placeholder.local').pathname}` } catch { return 'POST' }
}
function chooseProtocol(item: HelpApiProtocol) { protocol.value = item }
function resetProbe() {
  probeToken += 1
  probeAbort?.abort()
  probeAbort = undefined
  sending.value = false
  result.value = null
}
function formatProbeBody(text: string, truncated: boolean) {
  if (truncated) return text
  try { return JSON.stringify(JSON.parse(text), null, 2) } catch { return text }
}
async function reload() {
  resetProbe()
  loadAbort?.abort()
  const controller = new AbortController()
  loadAbort = controller
  const token = ++loadToken
  modelOptions.value = []
  selectedModel.value = ''
  keys.value = []
  keyId.value = ''
  keyTotal.value = 0
  modelsError.value = false
  keysError.value = false
  const current = group.value
  protocol.value = current ? helpApiProtocols(current, '').defaultProtocol : null
  if (!current) return
  modelsLoading.value = true
  keysLoading.value = true
  try {
    const [modelResult, keyResult] = await Promise.allSettled([
      userGroupsAPI.getHelpModels(current.id, controller.signal),
      keysAPI.list(1, 100, { group_id: current.id, status: 'active' }, { signal: controller.signal })
    ])
    if (token !== loadToken || controller.signal.aborted) return
    if (modelResult.status === 'fulfilled') {
      modelOptions.value = sortHelpModels(modelResult.value.map(item => item.id))
      selectedModel.value = modelOptions.value[0] || ''
    } else modelsError.value = true
    if (keyResult.status === 'fulfilled') {
      keys.value = keyResult.value.items || []
      keyTotal.value = keyResult.value.total || keys.value.length
      keyId.value = String(keys.value[0]?.id || '')
    } else keysError.value = true
    protocol.value = helpApiProtocols(current, selectedModel.value).defaultProtocol
  } finally {
    if (token === loadToken) {
      modelsLoading.value = false
      keysLoading.value = false
    }
  }
}
async function send() {
  if (sending.value) {
    probeAbort?.abort()
    return
  }
  if (!canSend.value || !group.value || !protocol.value || !selectedKey.value) return
  const token = ++probeToken
  const controller = new AbortController()
  probeAbort = controller
  let timedOut = false
  const timer = window.setTimeout(() => {
    timedOut = true
    controller.abort()
  }, HELP_API_TIMEOUT_MS)
  sending.value = true
  result.value = null
  const started = performance.now()
  try {
    const request = buildHelpApiRequest({
      baseUrl: props.baseUrl,
      platform: group.value.platform,
      protocol: protocol.value,
      model: selectedModel.value,
      apiKey: selectedKey.value.key
    })
    const response = await fetch(request.url, {
      method: request.method,
      headers: request.headers,
      body: JSON.stringify(request.body),
      signal: controller.signal,
      credentials: 'omit',
      cache: 'no-store'
    })
    const limited = await readHelpApiBody(response)
    if (token !== probeToken) return
    result.value = {
      kind: 'http',
      status: response.status,
      latencyMs: Math.round(performance.now() - started),
      requestId: response.headers.get('x-request-id') || '',
      body: formatProbeBody(limited.text, limited.truncated),
      truncated: limited.truncated
    }
  } catch (error) {
    if (token !== probeToken) return
    const latencyMs = Math.round(performance.now() - started)
    if (timedOut) result.value = { kind: 'timeout', latencyMs }
    else if (error instanceof DOMException && error.name === 'AbortError') result.value = { kind: 'aborted', latencyMs }
    else result.value = { kind: 'network', latencyMs }
  } finally {
    window.clearTimeout(timer)
    if (token === probeToken) {
      sending.value = false
      probeAbort = undefined
    }
  }
}

watch(() => props.groups.map(item => item.id).join(','), () => {
  if (!props.groups.some(item => String(item.id) === groupId.value)) groupId.value = String(props.groups[0]?.id || '')
}, { immediate: true })
watch(groupId, () => { void reload() }, { immediate: true })
watch(selectedModel, () => {
  if (group.value) {
    const next = helpApiProtocols(group.value, selectedModel.value)
    if (!protocol.value || !next.protocols.includes(protocol.value)) protocol.value = next.defaultProtocol
  }
  resetProbe()
})
watch(protocol, () => resetProbe())
onBeforeUnmount(() => {
  loadAbort?.abort()
  resetProbe()
})
</script>

<style scoped>
.api-probe { margin: 0 0 2rem; }
.probe-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: .75rem; }
.probe-field { display: grid; grid-template-columns: auto minmax(0, 1fr); align-items: center; gap: .35rem .65rem; padding: .7rem .8rem .75rem; border: 1px solid var(--av-color-border-default); border-radius: 12px; background: color-mix(in srgb, var(--av-color-bg-surface) 86%, var(--av-color-brand-primary) 14%); }
.probe-field label { font-size: .7rem; font-weight: 650; color: var(--av-color-brand-primary); }
.probe-field select { min-width: 0; width: 100%; height: 2.25rem; padding: .25rem .65rem; font-size: .75rem; color-scheme: light; background: var(--av-color-bg-surface); color: var(--av-color-text-strong); border-color: var(--av-color-border-control); }
.probe-field select:disabled { background: var(--av-color-bg-surface-muted); color: var(--av-color-text-muted); }
:global(:root.dark) .probe-field select { color-scheme: dark; }
.probe-protocols { display: flex; flex-wrap: wrap; gap: .4rem; margin-top: .75rem; }
.probe-protocols button { border-radius: 8px; padding: .65rem 1rem; font-size: .8125rem; border: 1px solid transparent; }
.probe-protocols button:hover { background: var(--av-color-brand-primary-soft); }
.probe-protocols .selected { color: var(--av-color-brand-primary); background: var(--av-color-brand-primary-soft); border-color: color-mix(in srgb, var(--av-color-brand-primary) 45%, transparent); font-weight: 650; }
.probe-actions { display: flex; align-items: center; flex-wrap: wrap; gap: .75rem 1rem; margin-top: .9rem; }
.probe-actions p, .probe-note { margin: .75rem 0 0; color: var(--av-color-text-muted); font-size: .75rem; line-height: 1.6; }
.probe-actions p { margin: 0; }
.probe-result { margin-top: 1rem; padding: 1rem; border: 1px solid var(--av-line); border-radius: 12px; }
.probe-meta { display: flex; flex-wrap: wrap; gap: .5rem 1.25rem; color: var(--av-color-text-muted); font-size: .75rem; }
.probe-meta strong { color: var(--av-color-text-default); font-weight: 650; }
.probe-body { margin: .75rem 0 0; overflow-x: auto; white-space: pre-wrap; color: var(--av-color-text-default); font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace; font-size: 12px; line-height: 1.6; }
@media (max-width: 800px) { .probe-grid { grid-template-columns: minmax(0, 1fr); } }
</style>
