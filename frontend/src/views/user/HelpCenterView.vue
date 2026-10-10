<template>
  <AppLayout>
    <div class="tutorial">
      <header class="tutorial-header">
        <p class="tutorial-intro">{{ t('helpCenter.intro') }}</p>
        <RouterLink to="/keys" class="btn btn-primary">{{ t('helpCenter.manageKeys') }} <span aria-hidden="true">↗</span></RouterLink>
      </header>
      <div class="tutorial-toolbar">
        <nav class="topic-tabs" :aria-label="t('helpCenter.menu')">
          <RouterLink to="/help" :class="{ selected: activeDocument?.category === 'quick-start' }">{{ t('helpCenter.quickStart') }}</RouterLink>
          <RouterLink v-for="doc in documents.filter(d => ['api', 'faq'].includes(d.category))" :key="doc.id" :to="articleLink(doc)" :class="{ selected: doc.id === activeDocument?.id }">{{ categoryLabel(doc.category) }}</RouterLink>
        </nav>
        <input v-model="search" type="search" class="input search-input" :placeholder="t('helpCenter.search')" :aria-label="t('helpCenter.search')">
      </div>
      <div v-if="search.trim()" class="search-results card">
        <RouterLink v-for="doc in searchResults" :key="doc.id" :to="articleLink(doc)" @click="search = ''"><strong>{{ doc.title }}</strong><span>{{ doc.summary }}</span></RouterLink>
        <p v-if="!searchResults.length">{{ t('helpCenter.empty') }}</p>
      </div>
      <p v-if="loading" role="status" class="py-12">{{ t('common.loading') }}</p>
      <div v-else-if="error" role="alert" class="card p-8"><p>{{ error }}</p><button class="btn btn-secondary mt-3" @click="load">{{ t('helpCenter.retry') }}</button></div>
      <div v-else-if="!activeDocument" class="card p-8">{{ t('helpCenter.empty') }}</div>
      <template v-else>
        <section v-if="activeDocument.category !== 'api'" class="setup-panel card" :aria-label="t('helpCenter.setup')">
          <nav class="channel-row" :aria-label="t('helpCenter.channel')"><span class="row-label">{{ t('helpCenter.channel') }}</span><div class="choice-row"><button v-for="platform in platforms" :key="platform" :aria-pressed="selectedPlatform === platform" :class="{ selected: selectedPlatform === platform }" @click="choosePlatform(platform)">{{ platformLabel(platform) }}</button></div></nav>
          <nav class="client-row" :aria-label="t('helpCenter.client')"><span class="row-label">{{ t('helpCenter.client') }}</span><div class="choice-row"><button v-for="client in clients" :key="client" :aria-pressed="isClientArticle && selectedClient === client" :class="{ selected: isClientArticle && selectedClient === client }" @click="chooseClient(client)"><ClientLogo :client="client" />{{ clientLabel(client) }}</button></div><span v-if="!clients.length">{{ t('helpCenter.empty') }}</span></nav>
          <p class="navigation-hint">{{ t('helpCenter.channelHint') }}</p>
          <template v-if="isClientArticle">
          <div class="selection-grid">
            <div class="setup-group"><label for="help-group">{{ t('helpCenter.group') }}</label><select id="help-group" v-model="groupId" class="input" @change="changeGroup"><option v-if="!availableGroups.length" value="">{{ t('helpCenter.choose') }}</option><option v-for="group in availableGroups" :key="group.id" :value="String(group.id)">{{ group.name }}</option></select><span class="selection-hint">{{ t('helpCenter.groupHint') }}</span></div>
            <div class="setup-group model-selection"><label for="help-model">{{ t('helpCenter.model') }}</label><select id="help-model" v-model="selectedModel" class="input" :disabled="modelsLoading || !modelOptions.length"><option v-if="!modelOptions.length" value="">{{ t(modelsLoading ? 'common.loading' : 'helpCenter.noModels') }}</option><option v-for="model in modelOptions" :key="model" :value="model">{{ model }}</option></select><span class="selection-hint">{{ t('helpCenter.modelHint') }}</span><button v-if="modelsError" type="button" @click="loadModels">{{ t('helpCenter.retry') }}</button></div>
          </div>
          <div v-if="!isDesktopClient(selectedClient) && !isTavernClient(selectedClient)" class="client-row"><span class="row-label">{{ t('helpCenter.os') }}</span><div class="choice-row"><button v-for="system in systems" :key="system.id" :aria-pressed="os === system.id" :class="{ selected: os === system.id }" @click="changeOS(system.id)">{{ system.label }}</button></div><label class="shell-label">Shell <select v-model="shell" class="input" @change="syncSelection"><option v-for="item in shells" :key="item">{{ item }}</option></select></label></div>
          <p v-if="!compatible" role="status" class="selection-warning">{{ !hasCompatibleGroup ? t('helpCenter.noClientGroup') : t('helpCenter.unsupported') }}</p>
          </template>
        </section>
        <div class="reading-layout">
          <main class="min-w-0">
            <header class="article-heading"><p class="eyebrow">{{ isClientArticle ? platformLabel(selectedPlatform) + ' / ' + clientLabel(selectedClient) : categoryLabel(activeDocument.category) }}</p><h2>{{ activeDocument.title }}</h2><p>{{ activeDocument.summary }}</p></header>
            <section v-if="activeDocument.category === 'quick-start'" class="client-directory">
              <RouterLink v-for="doc in clientDocuments" :key="doc.id" :to="articleLink(doc)"><ClientLogo :client="clientFor(doc)" /><div><strong>{{ clientLabel(clientFor(doc)) }}</strong><p>{{ doc.summary }}</p></div><span aria-hidden="true">↗</span></RouterLink>
              <p v-if="!groups.length" class="text-ink-muted">{{ t('helpCenter.noGroups') }}</p>
            </section>
            <template v-if="isClientArticle">
              <section id="guide-key" class="guide-step"><span class="step-number">01</span><div class="step-body"><h3>{{ t('helpCenter.stepKey') }}</h3><p>{{ t('helpCenter.keyInstructions') }}</p><RouterLink to="/keys" class="text-primary-600 underline">{{ t('helpCenter.manageKeys') }} ↗</RouterLink></div></section>
              <section v-for="(section, index) in sections" :id="'guide-' + index" :key="index" class="guide-step">
                <span class="step-number">{{ String(index + 2).padStart(2, '0') }}</span>
                <div class="step-body">
                  <h3>{{ section.title || activeDocument.title }}</h3>
                  <template v-if="isInstall(section.title) && installCommand">
                    <p>{{ ['claude', 'grok', 'hermes'].includes(selectedClient) ? t('helpCenter.nativeInstall') : t('helpCenter.npmInstall') }}</p>
                    <HelpCodeBlock :content="installCommand" :label="os === 'windows' ? shell : 'Terminal'" />
                    <p>{{ t('helpCenter.installSuccess') }}</p><HelpCodeBlock :content="command + ' --version'" :label="t('helpCenter.verifyInstall')" />
                    <details class="supplement"><summary>{{ t('helpCenter.installNotes') }}</summary><HelpMarkdown :content="section.markdown" :show-toc="false" /></details>
                  </template>
                  <template v-else-if="isConfiguration(section.title)">
                    <template v-if="prepareCommand"><p>{{ t(selectedClient === 'gemini' ? 'helpCenter.cleanGeminiEnv' : 'helpCenter.createDirectory') }}</p><HelpCodeBlock :content="prepareCommand" :label="os === 'windows' ? shell : 'Terminal'" /></template>
                    <p v-if="!isGuiClient">{{ t('helpCenter.writeConfig') }}</p>
                    <div v-if="selectedClient === 'claude'" class="choice-row my-3"><button :class="{ selected: configMode === 'file' }" :aria-pressed="configMode === 'file'" @click="configMode = 'file'">{{ t('helpCenter.persistentConfig') }}</button><button :class="{ selected: configMode === 'terminal' }" :aria-pressed="configMode === 'terminal'" @click="configMode = 'terminal'">{{ t('helpCenter.temporaryConfig') }}</button></div>
                    <p v-if="selectedClient === 'claude'" class="text-sm text-ink-muted">{{ t(configMode === 'file' ? 'helpCenter.persistentHint' : 'helpCenter.temporaryHint') }}</p>
                    <label v-if="selectedGroup?.platform === 'openai' && selectedClient.startsWith('codex')" class="block my-4">{{ t('helpCenter.authMode') }}<select v-model="codexAuthMode" class="input ml-2"><option value="legacy">{{ t('keys.useKeyModal.openai.authModeLegacy') }}</option><option value="api-key">{{ t('keys.useKeyModal.openai.authModeApiKey') }}</option></select></label>
                    <div v-if="selectedClient.startsWith('codex')" class="instruction-note"><p>{{ t('helpCenter.codexCatalogHint') }}</p><RouterLink to="/keys" class="underline">{{ t('helpCenter.openKeys') }} ↗</RouterLink></div>
                    <HelpClientSettings v-if="compatible && isGuiClient" :client-id="selectedClient" :client-name="clientLabel(selectedClient)" :fields="displayedFiles" />
                    <template v-else-if="compatible"><div v-for="file in displayedFiles" :key="file.path"><HelpCodeBlock :content="file.content" :label="file.path" /><p v-if="file.hint && !selectedClient.startsWith('codex')" class="text-sm text-ink-muted">{{ file.hint }}</p></div></template>
                    <div v-if="compatible && selectedClient === 'codex-ws'" class="protocol-branch-note"><strong>WebSocket / Context management</strong><p>启用 WebSocket 时，配置中的 <code>supports_websockets = true</code> 与 features 开关必须和当前 Codex 版本匹配；HTTP/SSE 网关不要套用该分支。上下文窗口和自动压缩只影响客户端管理，不代表模型实际支持更长上下文。</p></div>
                    <div v-else-if="compatible && selectedClient === 'codex'" class="protocol-branch-note"><strong>Responses / Context management</strong><p>标准分支使用 Responses over HTTP。保留 <code>context_window</code>、<code>model_catalog_json</code> 与 provider 的对应关系；模型列表成功不等于推理成功。</p></div>
                    <p v-if="!compatible" class="selection-warning">{{ t('helpCenter.unsupported') }}</p>
                    <div class="instruction-note">{{ t('helpCenter.configSecurityHint') }}</div>
                    <details v-if="isGuiClient" class="supplement"><summary>{{ t('helpCenter.specific.detailNotes') }}</summary><HelpMarkdown :content="section.markdown" :show-toc="false" /></details>
                    <HelpMarkdown v-else :content="section.markdown" :show-toc="false" />
                  </template>
                  <template v-else>
                    <HelpMarkdown :content="section.markdown" :show-toc="false" />
                    <HelpClaudeTerminal v-if="isVerification(section.title) && selectedClient === 'claude' && compatible" :model="selectedModel" :os="os" :shell="shell" :config-mode="configMode" />
                    <HelpOpenCodeTerminal v-if="isVerification(section.title) && selectedClient === 'opencode' && compatible" :model="selectedModel" :os="os" :provider="openCodeProvider" />
                    <HelpTerminalClientScene v-if="isVerification(section.title) && ['codex', 'codex-ws', 'dsh', 'pi', 'grok', 'openclaw', 'hermes'].includes(selectedClient) && compatible" :client="selectedClient === 'codex-ws' ? 'codex' : selectedClient" :model="selectedModel" :os="os" :shell="shell" :platform="selectedPlatform" />
                    <div v-if="isVerification(section.title)" class="success-note"><strong>{{ t('helpCenter.successTitle') }}</strong><p>{{ t('helpCenter.successBody') }}</p><RouterLink to="/usage" class="underline">{{ t('helpCenter.viewUsage') }} ↗</RouterLink></div>
                  </template>
                </div>
              </section>
              <section v-if="!sections.some(s => isConfiguration(s.title))" id="guide-config" class="guide-step"><span class="step-number">＋</span><div class="step-body"><h3>{{ t('helpCenter.generatedConfig') }}</h3><p>{{ t('helpCenter.writeConfig') }}</p><template v-if="compatible"><HelpCodeBlock v-for="file in configFiles" :key="file.path" :content="file.content" :label="file.path" /></template><p v-else>{{ t('helpCenter.unsupported') }}</p></div></section>
            </template>
            <div v-else-if="activeDocument.category === 'faq'" class="faq-list"><details v-for="(section, index) in sections" :key="index" :open="index === 0"><summary>{{ section.title || activeDocument.title }}</summary><HelpMarkdown :content="section.markdown" :show-toc="false" /></details></div>
            <template v-else-if="activeDocument.category === 'api'"><HelpApiProbe :groups="groups" :base-url="baseUrl" /><HelpMarkdown :content="activeDocument.content_markdown" :show-toc="false" /></template>
            <HelpMarkdown v-else :content="activeDocument.content_markdown" :show-toc="false" />
            <footer class="article-footer"><span v-if="activeDocument.published_at">{{ t('helpCenter.updatedAt') }} {{ new Date(activeDocument.updated_at).toLocaleDateString() }}</span><span>v1</span><RouterLink to="/keys">{{ t('helpCenter.openKeys') }} ↗</RouterLink></footer>
          </main>
          <aside class="reading-toc"><h2>{{ t('helpCenter.onThisPage') }}</h2><template v-if="isClientArticle"><a href="#guide-key">01 · {{ t('helpCenter.stepKey') }}</a><a v-for="(section, index) in sections" :key="index" :href="'#guide-' + index">{{ String(index + 2).padStart(2, '0') }} · {{ section.title }}</a><a v-if="!sections.some(s => isConfiguration(s.title))" href="#guide-config">{{ t('helpCenter.generatedConfig') }}</a></template><template v-else-if="activeDocument.category !== 'faq'"><a v-for="heading in articleHeadings" :key="heading.id" :href="'#' + heading.id">{{ heading.text }}</a></template><p>{{ t('helpCenter.helpHint') }}</p></aside>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { supportedClientIds, useClientConfiguration, type ClientId } from '@/composables/useClientConfiguration'
import type { Group, GroupPlatform } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import HelpMarkdown from '@/components/help/HelpMarkdown.vue'
import HelpCodeBlock from '@/components/help/HelpCodeBlock.vue'
import HelpClientSettings from '@/components/help/HelpClientSettings.vue'
import HelpClaudeTerminal from '@/components/help/HelpClaudeTerminal.vue'
import HelpOpenCodeTerminal from '@/components/help/HelpOpenCodeTerminal.vue'
import HelpTerminalClientScene from '@/components/help/HelpTerminalClientScene.vue'
import HelpApiProbe from '@/components/help/HelpApiProbe.vue'
import { desktopClientNames, isDesktopClient } from '@/utils/desktopClients'
import { extendedClients, isExtendedClient } from '@/utils/extendedClients'
import { sortHelpModels } from '@/utils/helpModels'
import ClientLogo from '@/components/common/ClientLogo.vue'
import { matchesHelpSelectors } from '@/utils/helpSelectors'
import { renderHelpMarkdown } from '@/utils/helpMarkdown'
import { helpSections, clientInstall, clientCommand, clientPrepare } from '@/utils/helpGuide'
import helpDocs, { type HelpDocument } from '@/api/helpDocs'
import { userGroupsAPI } from '@/api/groups'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const documents = ref<HelpDocument[]>([])
const groups = ref<Group[]>([])
const loading = ref(true)
const error = ref('')
const search = ref('')
const selectedPlatform = ref('')
const groupId = ref('')
const selectedClient = ref('')
const selectedModel = ref('')
const modelOptions = ref<string[]>([])
const modelsLoading = ref(false)
const modelsError = ref(false)
let modelRequest: AbortController | undefined
async function loadModels() {
  modelRequest?.abort()
  const request = new AbortController()
  modelRequest = request
  selectedModel.value = ''; modelOptions.value = []; modelsError.value = false; modelsLoading.value = false
  if (!selectedGroup.value) return
  modelsLoading.value = true
  try {
    const items = await userGroupsAPI.getHelpModels(selectedGroup.value.id, request.signal)
    if (request.signal.aborted) return
    modelOptions.value = sortHelpModels(items.map(item => item.id))
    selectedModel.value = modelOptions.value[0] || ''
  } catch { if (!request.signal.aborted) modelsError.value = true }
  finally { if (!request.signal.aborted) modelsLoading.value = false }
}
const isGuiClient = computed(() => isDesktopClient(selectedClient.value) || (isExtendedClient(selectedClient.value) && extendedClients[selectedClient.value].gui))
const os = ref('macos')
const shell = ref('bash')
const codexAuthMode = ref<'legacy' | 'api-key'>('legacy')
const manifest = ref('')
const configMode = ref<'file' | 'terminal'>('file')
const systems = [{ id: 'macos', label: 'macOS' }, { id: 'windows', label: 'Windows' }, { id: 'linux', label: 'Linux' }]
const aliases: Record<string, string> = { 'claude-code': 'claude', 'gemini-cli': 'gemini', 'grok-cli': 'grok', codex: 'codex', opencode: 'opencode', ...Object.fromEntries([...Object.keys(desktopClientNames), ...Object.keys(extendedClients)].map(id => [id, id])) }
const activeDocument = computed(() => documents.value.find(d => d.slug === String(route.params.slug || route.query.slug || 'quick-start')) || null)
const isClientArticle = computed(() => Boolean(activeDocument.value && aliases[activeDocument.value.category]))
const searchResults = computed(() => documents.value.filter(d => (d.title + d.summary + d.content_markdown).toLowerCase().includes(search.value.trim().toLowerCase())))
// Native guides stay discoverable before the account receives group access.
const nativePlatforms: GroupPlatform[] = ['anthropic', 'openai', 'gemini', 'grok']
const hiddenHelpPlatforms = new Set<GroupPlatform>(['deepseek', 'minimax', 'kimi', 'zhipu'])
const platforms = computed<string[]>(() => {
  const listed = [...new Set([...nativePlatforms, ...groups.value.map(g => g.platform)])].filter(platform => !hiddenHelpPlatforms.has(platform) && clientsForPlatform(platform).length)
  return clientsForPlatform('tavern').length ? [...listed, 'tavern'] : listed
})
const availableGroups = computed(() => selectedPlatform.value === 'tavern' ? groups.value.filter(g => !g.claude_code_only) : groups.value.filter(g => g.platform === selectedPlatform.value))
const selectedGroup = computed(() => availableGroups.value.find(g => String(g.id) === groupId.value))
const clients = computed(() => clientsForPlatform(selectedPlatform.value))
const clientDocuments = computed(() => documents.value.filter(d => aliases[d.category]))
const hasCompatibleGroup = computed(() => selectedPlatform.value === 'tavern' ? availableGroups.value.length > 0 : availableGroups.value.some(g => supportedClientIds(g.platform, g.allow_messages_dispatch).includes(selectedClient.value as ClientId)))
const baseUrl = computed(() => appStore.cachedPublicSettings?.api_base_url || window.location.origin)
const shells = computed(() => os.value === 'windows' ? (selectedClient.value.startsWith('codex') || selectedClient.value === 'grok' ? ['powershell'] : ['powershell', 'cmd']) : ['bash', 'zsh'])
const activeTab = computed(() => os.value !== 'windows' ? 'unix' : selectedClient.value.startsWith('codex') || selectedClient.value === 'grok' ? 'windows' : shell.value)
const compatible = computed(() => Boolean(selectedGroup.value && clients.value.includes(selectedClient.value as ClientId) && shells.value.includes(shell.value) && matchesHelpSelectors(activeDocument.value?.selector_schema, { platform: selectedPlatform.value, client: selectedClient.value, system: os.value, shell: shell.value })))
const context = { get platform() { return selectedGroup.value?.platform || null }, get baseUrl() { return baseUrl.value }, apiKey: 'YOUR_API_KEY', get model() { return selectedModel.value || undefined }, get allowMessagesDispatch() { return selectedGroup.value?.allow_messages_dispatch } }
const { currentFiles: configFiles } = useClientConfiguration(context, activeTab, selectedClient, codexAuthMode, manifest, t)
const displayedFiles = computed(() => selectedClient.value !== 'claude' ? configFiles.value : configFiles.value.filter(file => configMode.value === 'file' ? file.path.endsWith('settings.json') : !file.path.endsWith('settings.json')))
const openCodeProvider = computed(() => {
  if (selectedClient.value !== 'opencode') return ''
  try {
    const config = JSON.parse(configFiles.value[0]?.content || '{}')
    const providers = config.provider as Record<string, { name?: string }> | undefined
    return Object.entries(providers || {}).map(([id, provider]) => provider.name || id).join(' / ')
  } catch { return '' }
})
const sections = computed(() => helpSections(activeDocument.value?.content_markdown || ''))
const articleHeadings = computed(() => renderHelpMarkdown(activeDocument.value?.content_markdown || '').headings)
function isTavernClient(id: string) { return id === 'sillytavern' || id === 'tavernai' }
function clientsForPlatform(platform: string): ClientId[] {
  if (platform === 'tavern') return (['sillytavern', 'tavernai'] as const).filter(id => documents.value.some(d => clientFor(d) === id && matchesHelpSelectors(d.selector_schema, { platform, client: id })))
  return supportedClientIds((platform || null) as GroupPlatform | null, selectedGroup.value?.platform === platform ? selectedGroup.value.allow_messages_dispatch : groups.value.some(g => g.platform === platform && g.allow_messages_dispatch)).filter(id => documents.value.some(d => clientFor(d) === (id === 'codex-ws' ? 'codex' : id) && matchesHelpSelectors(d.selector_schema, { platform, client: id })))
}
const installCommand = computed(() => clientInstall(selectedClient.value, os.value, shell.value))
const prepareCommand = computed(() => clientPrepare(selectedClient.value, os.value, shell.value))
const command = computed(() => clientCommand(selectedClient.value))
function isInstall(title: string) { return /安装|install/i.test(title) }
function isConfiguration(title: string) { return /认证|配置|configuration|authentication/i.test(title) && !/验证|verification/i.test(title) }
function isVerification(title: string) { return /验证与|成功判据|verification/i.test(title) }
function clientFor(doc: HelpDocument) { return aliases[doc.category] || '' }
function clientLabel(id: string) { return isExtendedClient(id) ? extendedClients[id].name : ({ claude: 'Claude Code', codex: 'Codex', 'codex-ws': 'Codex WebSocket', opencode: 'OpenCode', gemini: 'Gemini CLI', grok: 'Grok CLI', ...desktopClientNames } as Record<string, string>)[id] || id }
function platformLabel(id: string) { return ({ anthropic: 'Claude', openai: 'OpenAI', gemini: 'Gemini', antigravity: 'Antigravity', grok: 'Grok', deepseek: 'DeepSeek', minimax: 'MiniMax', kimi: 'Kimi', zhipu: '智谱 GLM', composite: t('helpCenter.composite'), tavern: t('helpCenter.tavernChannel') } as Record<string, string>)[id] || id }
function chatCompletionGroup(client: string) {
  if (!isTavernClient(client)) return undefined
  return groups.value.find(g => String(g.id) === groupId.value && !g.claude_code_only) || groups.value.find(g => !g.claude_code_only)
}
function categoryLabel(id: string) { return id === 'quick-start' ? t('helpCenter.quickStart') : id === 'faq' ? t('helpCenter.faq') : id === 'api' ? t('helpCenter.api') : clientLabel(aliases[id] || id) }
function safeQuery() { return { platform: selectedPlatform.value, group: groupId.value, client: selectedClient.value, os: os.value, shell: shell.value } }
function articleLink(doc: HelpDocument) {
  const client = clientFor(doc)
  const tavernGroup = chatCompletionGroup(client)
  if (isTavernClient(client)) return { path: '/help/' + doc.slug, query: { ...safeQuery(), platform: 'tavern', group: String(tavernGroup?.id || ''), client } }
  const group = client ? groups.value.find(g => String(g.id) === groupId.value && supportedClientIds(g.platform, g.allow_messages_dispatch).includes(client as ClientId)) || groups.value.find(g => supportedClientIds(g.platform, g.allow_messages_dispatch).includes(client as ClientId)) : selectedGroup.value
  const platform = group?.platform || platforms.value.find(p => clientsForPlatform(p).includes(client as ClientId)) || selectedPlatform.value
  return { path: '/help/' + doc.slug, query: { ...safeQuery(), platform, group: String(group?.id || ''), client } }
}
function restoreSelection() {
  const docClient = activeDocument.value ? clientFor(activeDocument.value) : ''
  selectedClient.value = docClient === 'codex' && route.query.client === 'codex-ws' ? 'codex-ws' : docClient || String(route.query.client || '')
  const fallback = isTavernClient(selectedClient.value) ? groups.value.find(g => !g.claude_code_only) : groups.value.find(g => supportedClientIds(g.platform, g.allow_messages_dispatch).includes(selectedClient.value as ClientId)) || (!docClient ? groups.value[0] : undefined)
  const requestedPlatform = String(route.query.platform || '')
  selectedPlatform.value = platforms.value.includes(requestedPlatform)
    ? requestedPlatform
    : fallback?.platform || platforms.value.find(p => clientsForPlatform(p).includes(selectedClient.value as ClientId)) || platforms.value[0] || ''
  if (isTavernClient(selectedClient.value) && platforms.value.includes('tavern')) selectedPlatform.value = 'tavern'
  const requestedGroup = String(route.query.group || '')
  groupId.value = availableGroups.value.some(g => String(g.id) === requestedGroup) ? requestedGroup : String(availableGroups.value[0]?.id || fallback?.id || '')
  os.value = systems.some(s => s.id === route.query.os) ? String(route.query.os) : 'macos'
  shell.value = String(route.query.shell || (os.value === 'windows' ? 'powershell' : 'bash'))
}
function syncSelection() { void router.replace({ path: route.path, query: safeQuery(), hash: route.hash }) }
function chooseClient(client: string) {
  selectedClient.value = client
  if (!shells.value.includes(shell.value)) shell.value = shells.value[0]!
  const doc = documents.value.find(d => clientFor(d) === (client === 'codex-ws' ? 'codex' : client))
  if (doc) void router.push({ path: '/help/' + doc.slug, query: safeQuery() })
}
function changeGroup() { chooseClient(clients.value.includes(selectedClient.value as ClientId) ? selectedClient.value : clients.value[0] || '') }
function choosePlatform(platform: string) { selectedPlatform.value = platform; groupId.value = String(availableGroups.value[0]?.id || ''); changeGroup() }
function changeOS(system: string) { os.value = system; shell.value = system === 'windows' ? 'powershell' : 'bash'; syncSelection() }
async function load() {
  loading.value = true; error.value = ''
  try {
    documents.value = (await helpDocs.list()).filter(doc => !['read-frog', 'roo-code'].includes(doc.category))
    groups.value = (await userGroupsAPI.getAvailable()).filter(g => g.status === 'active')
    await appStore.fetchPublicSettings()
    restoreSelection()
    if (Object.keys(route.query).some(key => !Object.keys(safeQuery()).includes(key)) || String(route.query.platform || '') !== selectedPlatform.value) syncSelection()
  } catch { error.value = t('helpCenter.loadFailed') } finally { loading.value = false }
}
watch(() => route.fullPath, () => { if (!loading.value) restoreSelection() })
onMounted(load)
watch(() => selectedGroup.value?.id, loadModels)
onBeforeUnmount(() => modelRequest?.abort())
</script>

<style scoped>
.tutorial { max-width: 1280px; margin: 0 auto; color: var(--av-color-text-default); }
.tutorial-header { display: flex; align-items: center; justify-content: space-between; gap: 1rem; padding: .25rem 0 1.25rem; }
.tutorial-intro { margin: 0; max-width: 42rem; color: var(--av-color-text-muted); font-size: .875rem; line-height: 1.7; }
.tutorial-toolbar { border-bottom: 1px solid var(--av-line); display: flex; gap: 1rem; align-items: center; margin-bottom: 1.5rem; }
.topic-tabs { display: flex; overflow-x: auto; gap: .25rem; flex: 1; }
.topic-tabs a, .topic-tabs button { white-space: nowrap; padding: .9rem 1rem; font-size: .875rem; border-bottom: 2px solid transparent; color: var(--av-color-text-muted); }
.topic-tabs .selected { border-color: var(--av-color-brand-primary); color: var(--av-color-brand-primary); font-weight: 650; }
.search-input { width: 180px; margin-bottom: .5rem; }
.setup-panel { padding: 1.25rem 1.5rem; margin-bottom: 2rem; }
.channel-row { display: flex; align-items: center; gap: 1rem; flex-wrap: wrap; }
.navigation-hint { font-size: .75rem; line-height: 1.7; color: var(--av-color-text-muted); margin: 1rem 0; }
.setup-group { display: flex; align-items: center; gap: 1rem; padding-bottom: 1rem; border-bottom: 1px solid var(--av-line); flex-wrap: wrap; }
.selection-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .75rem; margin: .75rem 0 .25rem; }
.selection-grid .setup-group { display: grid; grid-template-columns: auto minmax(0, 1fr); align-items: center; gap: .35rem .65rem; padding: .7rem .8rem .75rem; margin: 0; border: 1px solid var(--av-color-border-default); border-radius: 12px; background: color-mix(in srgb, var(--av-color-bg-surface) 86%, var(--av-color-brand-primary) 14%); }
.selection-grid .setup-group label { min-width: 0; font-size: .7rem; color: var(--av-color-brand-primary); }
.selection-grid .setup-group select { min-width: 0; width: 100%; height: 2.25rem; padding: .25rem .65rem; font-size: .75rem; color-scheme: light; background: var(--av-color-bg-surface); color: var(--av-color-text-strong); border-color: var(--av-color-border-control); }
.selection-grid .setup-group select:disabled { background: var(--av-color-bg-surface-muted); color: var(--av-color-text-muted); }
:global(:root.dark) .selection-grid .setup-group select { color-scheme: dark; }
.selection-hint { grid-column: 1 / -1; overflow: hidden; color: var(--av-color-text-muted); font-size: .65rem; line-height: 1.4; text-overflow: ellipsis; white-space: nowrap; }
.setup-group label, .row-label { font-size: .75rem; font-weight: 600; min-width: 64px; color: var(--av-color-text-muted); }
.setup-group select { min-width: 180px; max-width: 100%; }
.client-row { display: flex; align-items: center; gap: 1rem; margin-top: 1rem; flex-wrap: wrap; }
.choice-row { display: flex; gap: .4rem; flex-wrap: wrap; }
.choice-row button { display: inline-flex; align-items: center; gap: .5rem; }
.model-selection { margin: 0; }
.choice-row button { border-radius: 8px; padding: .65rem 1rem; font-size: .8125rem; border: 1px solid transparent; }
.choice-row button:hover { background: var(--av-color-brand-primary-soft); }
.choice-row .selected { color: var(--av-color-brand-primary); background: var(--av-color-brand-primary-soft); border-color: color-mix(in srgb, var(--av-color-brand-primary) 45%, transparent); font-weight: 650; }
.shell-label { display: flex; align-items: center; gap: .5rem; font-size: .75rem; margin-left: auto; color: var(--av-color-text-muted); }
.shell-label select { width: auto; min-width: 8.5rem; color-scheme: light; background: var(--av-color-bg-surface); color: var(--av-color-text-strong); border-color: var(--av-color-border-control); }
:global(:root.dark) .shell-label select { color-scheme: dark; }
.reading-layout { display: grid; grid-template-columns: minmax(0, 1fr) 190px; gap: 4rem; padding: 0 1rem; }
.article-heading { margin: 0 0 2rem; }
.eyebrow { color: var(--av-color-text-muted); font-size: .7rem; letter-spacing: .12em; margin-bottom: .5rem; }
.article-heading h2 { font-size: 1.65rem; font-weight: 650; letter-spacing: -.025em; }
.article-heading > p:last-child { margin-top: .75rem; color: var(--av-color-text-muted); font-size: .875rem; line-height: 1.8; }
.guide-step { display: flex; gap: 1.2rem; scroll-margin-top: 6rem; padding: 0 0 2rem; margin-bottom: 1rem; border-bottom: 1px solid var(--av-line); }
.step-number { font-size: .7rem; font-family: monospace; color: var(--av-color-brand-primary); background: var(--av-color-brand-primary-soft); border: 1px solid color-mix(in srgb, var(--av-color-brand-primary) 28%, transparent); display: grid; place-content: center; height: 30px; width: 30px; flex-shrink: 0; border-radius: 50%; margin-top: .1rem; }
.step-body { flex: 1; min-width: 0; font-size: .875rem; line-height: 1.9; }
.step-body h3 { font-size: 1.075rem; font-weight: 650; margin-bottom: 1rem; }
.step-body > p { margin-bottom: .75rem; }
.reading-toc { position: sticky; top: 6rem; align-self: start; border-left: 1px solid var(--av-line); padding-left: 1.25rem; }
.reading-toc h2 { font-size: .75rem; font-weight: 650; margin-bottom: 1rem; }
.reading-toc a { display: block; font-size: .75rem; padding: .6rem 0; color: var(--av-color-text-muted); }
.reading-toc a:hover { color: var(--av-color-brand-primary); }
.reading-toc p { margin-top: 1.5rem; font-size: .7rem; line-height: 1.8; color: var(--av-color-text-muted); }
.instruction-note, .success-note, .selection-warning { border-radius: 10px; padding: 1rem; margin: 1rem 0; background: var(--av-color-brand-primary-soft); border: 1px solid color-mix(in srgb, var(--av-color-brand-primary) 28%, transparent); font-size: .8125rem; line-height: 1.8; }
.protocol-branch-note { border-radius: 10px; padding: 1rem; margin: 1rem 0; background: var(--av-color-info-soft); border: 1px solid color-mix(in srgb, var(--av-color-info) 35%, transparent); font-size: .8125rem; line-height: 1.8; }
.protocol-branch-note strong { color: var(--av-color-info); }
.selection-warning { background: var(--av-color-warning-soft); border-color: color-mix(in srgb, var(--av-color-warning) 40%, transparent); color: var(--av-color-text-default); }
.success-note strong { color: var(--av-color-brand-primary); }
.supplement summary { cursor: pointer; font-size: .8125rem; color: var(--av-color-text-muted); padding: .75rem 0; }
.client-directory { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .75rem; margin-bottom: 2rem; }
.client-directory a { display: flex; align-items: center; gap: 1rem; padding: 1.25rem; border: 1px solid var(--av-line); border-radius: 12px; }
.client-directory a:hover { border-color: var(--av-color-brand-primary); background: color-mix(in srgb, var(--av-color-brand-primary-soft) 70%, transparent); }
.client-mark { font-family: monospace; font-size: 1.2rem; color: var(--av-color-brand-primary); }
.client-directory strong { font-size: .875rem; }
.client-directory p { font-size: .75rem; line-height: 1.7; margin-top: .4rem; color: var(--av-color-text-muted); }
.faq-list details { border-bottom: 1px solid var(--av-line); padding: 1.25rem 0; }
.faq-list summary { cursor: pointer; font-weight: 600; margin-bottom: 1rem; }
.article-footer { border-top: 1px solid var(--av-line); margin: 2rem 0; padding-top: 1rem; display: flex; flex-wrap: wrap; gap: 1rem; font-size: .75rem; color: var(--av-color-text-muted); }
.article-footer a { margin-left: auto; }
.search-results { margin-bottom: 1rem; padding: 1rem; }
.search-results a { display: block; padding: .75rem; }
.search-results span { display: block; font-size: .75rem; margin-top: .3rem; }
@media (max-width: 1100px) { .reading-layout { grid-template-columns: minmax(0, 1fr); gap: 0; padding: 0; } .reading-toc { display: none; } }
@media (max-width: 640px) { .tutorial-header { align-items: start; } .tutorial-header .btn { font-size: .75rem; } .tutorial-toolbar { flex-direction: column-reverse; align-items: stretch; gap: .2rem; } .search-input { width: 100%; } .setup-panel { padding: 1rem; } .client-row { gap: .5rem; } .row-label { width: 100%; } .shell-label { margin-left: 0; } .choice-row button { padding: .6rem .75rem; } .selection-grid { grid-template-columns: minmax(0, 1fr); } .selection-grid .setup-group { grid-template-columns: 5rem minmax(0, 1fr); } .guide-step { gap: .65rem; } .client-directory { grid-template-columns: minmax(0, 1fr); } }
</style>
