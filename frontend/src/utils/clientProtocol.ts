import type { GroupPlatform } from '@/types'

// These project platforms expose native Messages routes. Other platforms keep
// the OpenAI-compatible adapter unless a client has a more specific rule.
export function usesNativeMessages(platform: GroupPlatform | null | undefined) {
  return ['anthropic', 'antigravity', 'deepseek', 'minimax'].includes(platform || '')
}

export function clientEndpoints(baseUrl: string, platform?: GroupPlatform | null) {
  const root = baseUrl.replace(/\/+$/, '').replace(/\/v1$/, '')
  const messagesRoot = platform === 'antigravity' ? `${root}/antigravity` : root
  return { root, openai: `${root}/v1`, messagesRoot, messages: `${messagesRoot}/v1/messages` }
}
