import type { GroupPlatform } from '@/types'

export const HELP_API_PROMPT = 'hi'
export const HELP_API_MAX_TOKENS = 16
export const HELP_API_BODY_LIMIT = 8192
export const HELP_API_TIMEOUT_MS = 30_000

export type HelpApiProtocol = 'messages' | 'responses' | 'chat' | 'gemini'

export interface HelpApiGroupInput {
  platform: GroupPlatform | string
  claude_code_only?: boolean
  allow_messages_dispatch?: boolean
}

export interface HelpApiProtocolSet {
  protocols: HelpApiProtocol[]
  defaultProtocol: HelpApiProtocol | null
}

export interface BuiltHelpApiRequest {
  method: 'POST'
  url: string
  headers: Record<string, string>
  body: Record<string, unknown>
  preview: string
}

const providerAliases: Record<string, string> = {
  anthropic: 'anthropic',
  claude: 'anthropic',
  openai: 'openai',
  chatgpt: 'openai',
  google: 'gemini',
  'google-ai-studio': 'gemini',
  gemini: 'gemini',
  xai: 'grok',
  'x-ai': 'grok',
  grok: 'grok',
  kimi: 'kimi',
  moonshot: 'kimi',
  zhipu: 'zhipu',
  glm: 'zhipu',
  bigmodel: 'zhipu',
  deepseek: 'deepseek',
  minimax: 'minimax'
}

export function detectHelpModelPlatform(model: string): string | null {
  let normalized = model.trim().toLowerCase()
  if (!normalized) return null
  if (normalized.startsWith('models/')) normalized = normalized.slice('models/'.length)
  const slash = normalized.indexOf('/')
  if (slash > 0) {
    const provider = normalized.slice(0, slash).trim()
    const rest = normalized.slice(slash + 1).trim()
    const mapped = providerAliases[provider]
    if (mapped) return mapped
    if (rest) normalized = rest.startsWith('models/') ? rest.slice('models/'.length) : rest
  }
  if (normalized.startsWith('anthropic.claude-') || normalized.startsWith('claude-')) return 'anthropic'
  if (
    normalized.startsWith('gpt-') ||
    normalized.startsWith('chatgpt-') ||
    normalized.startsWith('codex-') ||
    normalized.startsWith('text-embedding-') ||
    normalized.startsWith('text-moderation-') ||
    normalized.startsWith('omni-moderation-') ||
    normalized.startsWith('dall-e-') ||
    normalized.startsWith('gpt-image-') ||
    normalized.startsWith('tts-') ||
    normalized.startsWith('whisper-') ||
    hasOpenAISeriesPrefix(normalized)
  ) return 'openai'
  if (normalized.startsWith('gemini-') || normalized.startsWith('learnlm-')) return 'gemini'
  if (normalized === 'grok' || normalized.startsWith('grok-')) return 'grok'
  if (normalized === 'k3' || normalized === 'k3-256k' || normalized.startsWith('kimi-') || normalized.startsWith('moonshot-')) return 'kimi'
  if (normalized.startsWith('glm-')) return 'zhipu'
  if (normalized.startsWith('deepseek-')) return 'deepseek'
  if (normalized.startsWith('minimax-') || normalized.startsWith('abab5') || normalized.startsWith('abab6') || normalized.startsWith('abab7')) return 'minimax'
  return null
}

function hasOpenAISeriesPrefix(model: string) {
  return ['o1', 'o3', 'o4', 'o5'].some(prefix => model === prefix || model.startsWith(`${prefix}-`))
}

function protocolsFor(platform: string, allowMessagesDispatch: boolean): HelpApiProtocolSet {
  switch (platform) {
    case 'anthropic':
      return { protocols: ['messages', 'responses', 'chat'], defaultProtocol: 'messages' }
    case 'openai':
      return allowMessagesDispatch
        ? { protocols: ['responses', 'chat', 'messages'], defaultProtocol: 'responses' }
        : { protocols: ['responses', 'chat'], defaultProtocol: 'responses' }
    case 'gemini':
      return { protocols: ['gemini', 'responses'], defaultProtocol: 'gemini' }
    case 'antigravity':
      return { protocols: ['messages', 'gemini'], defaultProtocol: 'messages' }
    case 'grok':
      return { protocols: ['responses', 'messages', 'chat'], defaultProtocol: 'responses' }
    case 'deepseek':
    case 'minimax':
    case 'kimi':
    case 'zhipu':
      return { protocols: ['messages', 'responses', 'chat'], defaultProtocol: 'messages' }
    default:
      return { protocols: [], defaultProtocol: null }
  }
}

export function helpApiProtocols(group: HelpApiGroupInput, model = ''): HelpApiProtocolSet {
  if (group.claude_code_only) return { protocols: ['messages'], defaultProtocol: 'messages' }
  if (group.platform === 'composite') {
    const detected = detectHelpModelPlatform(model)
    if (!detected) return { protocols: ['messages', 'responses', 'chat', 'gemini'], defaultProtocol: null }
    return protocolsFor(detected, Boolean(group.allow_messages_dispatch))
  }
  return protocolsFor(group.platform, Boolean(group.allow_messages_dispatch))
}

export function helpApiRoot(baseUrl: string) {
  return baseUrl.trim().replace(/\/+$/, '').replace(/\/v1$/i, '')
}

export function maskApiKey(key: string) {
  const trimmed = key.trim()
  if (trimmed.length <= 4) return '****'
  return `****${trimmed.slice(-4)}`
}

function geminiModelSegment(model: string) {
  return encodeURIComponent(model.trim().replace(/^models\//i, ''))
}

function requestBody(protocol: HelpApiProtocol, model: string): Record<string, unknown> {
  if (protocol === 'messages') {
    return { model, max_tokens: HELP_API_MAX_TOKENS, messages: [{ role: 'user', content: HELP_API_PROMPT }], stream: false }
  }
  if (protocol === 'responses') {
    return { model, input: HELP_API_PROMPT, stream: false }
  }
  if (protocol === 'chat') {
    return { model, messages: [{ role: 'user', content: HELP_API_PROMPT }], max_tokens: HELP_API_MAX_TOKENS, stream: false }
  }
  return { contents: [{ role: 'user', parts: [{ text: HELP_API_PROMPT }] }] }
}

function shellSingleQuote(value: string) {
  return `'${value.replace(/'/g, `'\\''`)}'`
}

function curlPreview(url: string, headers: [string, string][], body: Record<string, unknown>) {
  const lines = [`curl ${shellSingleQuote(url)} \\`]
  for (const [name, value] of headers) lines.push(`  -H ${shellSingleQuote(`${name}: ${value}`)} \\`)
  lines.push(`  -d ${shellSingleQuote(JSON.stringify(body, null, 2))}`)
  return lines.join('\n')
}

export function buildHelpApiRequest(input: {
  baseUrl: string
  platform: string
  protocol: HelpApiProtocol
  model: string
  apiKey: string
}): BuiltHelpApiRequest {
  const root = helpApiRoot(input.baseUrl)
  const antigravity = input.platform === 'antigravity'
  const path = input.protocol === 'messages'
    ? `${antigravity ? '/antigravity' : ''}/v1/messages`
    : input.protocol === 'responses'
      ? '/v1/responses'
      : input.protocol === 'chat'
        ? '/v1/chat/completions'
        : `${antigravity ? '/antigravity' : ''}/v1beta/models/${geminiModelSegment(input.model)}:generateContent`
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const masked = maskApiKey(input.apiKey)
  const previewHeaders: [string, string][] = []
  if (input.protocol === 'messages') {
    headers['x-api-key'] = input.apiKey
    headers['anthropic-version'] = '2023-06-01'
    previewHeaders.push(['x-api-key', masked], ['anthropic-version', '2023-06-01'])
  } else if (input.protocol === 'gemini') {
    headers['x-goog-api-key'] = input.apiKey
    previewHeaders.push(['x-goog-api-key', masked])
  } else {
    headers.Authorization = `Bearer ${input.apiKey}`
    previewHeaders.push(['Authorization', `Bearer ${masked}`])
  }
  previewHeaders.push(['Content-Type', 'application/json'])
  const body = requestBody(input.protocol, input.model)
  const url = `${root}${path}`
  return { method: 'POST', url, headers, body, preview: curlPreview(url, previewHeaders, body) }
}

export async function readHelpApiBody(response: Response, limit = HELP_API_BODY_LIMIT): Promise<{ text: string; truncated: boolean }> {
  const reader = response.body?.getReader()
  if (!reader) {
    const text = await response.text()
    const encoded = new TextEncoder().encode(text)
    if (encoded.length <= limit) return { text, truncated: false }
    return { text: new TextDecoder().decode(encoded.slice(0, limit)), truncated: true }
  }
  const chunks: Uint8Array[] = []
  let total = 0
  let truncated = false
  while (total < limit) {
    const { done, value } = await reader.read()
    if (done) break
    if (!value?.byteLength) continue
    const room = limit - total
    if (value.byteLength > room) {
      chunks.push(value.slice(0, room))
      total += room
      truncated = true
      break
    }
    chunks.push(value)
    total += value.byteLength
  }
  if (!truncated) {
    const extra = await reader.read()
    truncated = !extra.done && Boolean(extra.value?.byteLength)
  }
  await reader.cancel().catch(() => undefined)
  const merged = new Uint8Array(total)
  let offset = 0
  for (const chunk of chunks) {
    merged.set(chunk, offset)
    offset += chunk.byteLength
  }
  return { text: new TextDecoder().decode(merged), truncated }
}
