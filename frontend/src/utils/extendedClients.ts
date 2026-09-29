import type { GroupPlatform } from '@/types'
import { clientEndpoints, usesNativeMessages } from './clientProtocol'

export const extendedClients = {
  dsh: { name: 'dsh', gui: true, source: 'https://github.com/deepseek-ai/deepseek-harness/blob/master/docs/user/guide/providers.md' },
  pi: { name: 'Pi', gui: false, source: 'https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/models.md' },
  openclaw: { name: 'OpenClaw', gui: false, source: 'https://docs.openclaw.ai/gateway/config-tools/custom-providers' },
  hermes: { name: 'Hermes', gui: false, source: 'https://hermes-agent.nousresearch.com/docs/integrations/providers' },
  workbuddy: { name: 'WorkBuddy', gui: true, source: 'https://www.workbuddy.ai/docs/workbuddy/From-Beginner-to-Expert-Guide/Function-Description/Model' },
  zcode: { name: 'ZCode', gui: true, source: 'https://zcode.z.ai/cn/docs/configuration' },
  trae: { name: 'TRAE SOLO', gui: true, source: 'https://www.trae.ai/' },
  'read-frog': { name: 'Read Frog', gui: true, source: 'https://www.readfrog.app/zh/docs/providers/openai-compatible-providers' },
  'kiss-translator': { name: 'Kiss Translator', gui: true, source: 'https://github.com/fishjar/kiss-translator' },
  'immersive-translate': { name: '沉浸式翻译', gui: true, source: 'https://immersivetranslate.com/docs/services/openai/' },
  sillytavern: { name: 'SillyTavern', gui: true, source: 'https://docs.sillytavern.app/usage/api-connections/openai/' },
  tavernai: { name: 'TavernAI', gui: true, source: 'https://tavernai.net/docs/quick-start/' }
} as const
export type ExtendedClientId = keyof typeof extendedClients
export function isExtendedClient(id: string): id is ExtendedClientId {
  return Object.prototype.hasOwnProperty.call(extendedClients, id)
}
const channelScopedClients = new Set<ExtendedClientId>(['sillytavern', 'tavernai'])
export function extendedClientIds(platform: GroupPlatform): ExtendedClientId[] {
  return (Object.keys(extendedClients) as ExtendedClientId[]).filter(id => !channelScopedClients.has(id) && (id !== 'pi' || platform === 'openai'))
}

export function extendedClientFiles(client: ExtendedClientId, platform: GroupPlatform | null, baseUrl: string, key: string, model = 'YOUR_MODEL_ID') {
  const root = baseUrl.replace(/\/+$/, '').replace(/\/v1$/, '')
  const base = `${root}/v1`
  const field = (path: string, content: string) => ({ path, content })
  const json = (value: unknown) => JSON.stringify(value, null, 2)
  const anthropicRoot = platform === 'antigravity' ? `${root}/antigravity` : root
  const nativeMessages = usesNativeMessages(platform)
  const endpoints = clientEndpoints(baseUrl, platform)
  const nativeTranslation = usesNativeMessages(platform)
  const openclaw = openclawConnection(platform, base, anthropicRoot)
  const hermesMode = hermesApiMode(platform)
  const hermesBase = hermesMode === 'anthropic_messages' ? anthropicRoot : base
  if (client === 'pi') return [
    field('~/.pi/agent/models.json', json({ providers: { site: { name: 'site', baseUrl: base, api: 'openai-responses', apiKey: key, models: [{ id: model, name: model }] } } })),
    field('Terminal', `mkdir -p ~/.pi/agent\nnano ~/.pi/agent/models.json\npi --provider site --model ${model}`)
  ]
  if (client === 'openclaw') return [field('~/.openclaw/openclaw.json', json({
    models: { mode: 'merge', providers: { site: { baseUrl: openclaw.baseUrl, apiKey: key, api: openclaw.api, models: [{ id: model, name: model }] } } },
    agents: { defaults: { model: { primary: `site/${model}` } } }
  }))]
  if (client === 'hermes') return [
    field('~/.hermes/config.yaml', hermesConfig(hermesBase, key, model, hermesMode)),
    field('Terminal', hermesCommands(hermesBase, model, hermesMode))
  ]
  if (client === 'dsh') return [field('Provider ID', 'site'), field('Base URL', base), field('API', 'openai-responses'), field('API Key', key), field('Model ID', model)]
  if (client === 'trae') return [field('API 格式', nativeMessages ? 'Anthropic Messages' : 'OpenAI Chat Completions'), field('自定义请求地址', nativeMessages ? anthropicRoot : base), field('完整 URL', 'OFF'), field('模型 ID', model), field('API 密钥', key)]
  if (client === 'workbuddy') return [field('Provider', '自定义 / Custom'), field('URL', `${base}/chat/completions`), field('Custom Protocol', 'OFF'), field('API Key', key), field('Model', model)]
  if (client === 'zcode') return [field('Provider', 'site'), field('Base URL', base), field('API Key', key), field('Model ID', model)]
  if (client === 'kiss-translator' || client === 'immersive-translate') return [field('Provider', nativeTranslation ? 'Claude' : 'OpenAI'), field('API URL', nativeTranslation ? endpoints.messages : `${base}/chat/completions`), field('API Key', key), field('Model', model)]
  if (client === 'read-frog' && nativeTranslation) return [field('Provider', 'Anthropic'), field('Base URL', `${endpoints.messagesRoot}/v1`), field('API Key', key), field('Model ID', model)]
  if (client === 'sillytavern' || client === 'tavernai') return tavernChatFields(client, base, key, model)
  return [field('Provider', 'OpenAI Compatible'), field('Base URL', base), field('API Key', key), field('Model ID', model)]
}

export function openclawConnection(platform: GroupPlatform | null, openaiBase: string, anthropicRoot: string) {
  if (usesNativeMessages(platform)) return { baseUrl: anthropicRoot, api: 'anthropic-messages' }
  if (platform === 'openai') return { baseUrl: openaiBase, api: 'openai-responses' }
  return { baseUrl: openaiBase, api: 'openai-completions' }
}

export function hermesApiMode(platform: GroupPlatform | null) {
  if (usesNativeMessages(platform)) return 'anthropic_messages'
  if (platform === 'openai') return 'codex_responses'
  return ''
}

function hermesConfig(base: string, key: string, model: string, apiMode: string) {
  const mode = apiMode ? `\n  api_mode: ${apiMode}` : ''
  return `model:\n  provider: custom\n  default: ${JSON.stringify(model)}\n  base_url: ${JSON.stringify(base)}\n  api_key: ${JSON.stringify(key)}${mode}`
}

function tavernChatFields(client: 'sillytavern' | 'tavernai', base: string, key: string, model: string) {
  const entry = (path: string, content: string) => ({ path, content })
  if (client === 'sillytavern') return [
    entry('API', 'Chat Completion'),
    entry('Chat Completion Source', 'Custom (OpenAI-compatible)'),
    entry('Custom Endpoint (Base URL)', base),
    entry('Custom API Key', key),
    entry('Model ID', model)
  ]
  return [
    entry('Base Provider', 'Custom'),
    entry('Model Endpoint', 'Chat Completions'),
    entry('API Address', base),
    entry('Use direct API address', 'ON'),
    entry('API key', key),
    entry('Model ID', model)
  ]
}

function hermesCommands(base: string, model: string, apiMode: string) {
  const mode = apiMode ? `\nhermes config set model.api_mode ${apiMode}` : ''
  return `hermes config set model.provider custom\nhermes config set model.base_url ${base}\nhermes config set model.default ${model}${mode}\nhermes chat -m ${model}`
}
