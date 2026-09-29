import type { GroupPlatform } from '@/types'
import { clientEndpoints, usesNativeMessages } from './clientProtocol'

export const desktopClientNames = {
  'cherry-studio': 'Cherry Studio',
  cursor: 'Cursor',
  cline: 'Cline',
  'roo-code': 'Roo Code'
} as const

export type DesktopClientId = keyof typeof desktopClientNames
export function isDesktopClient(id: string): id is DesktopClientId {
  return Object.prototype.hasOwnProperty.call(desktopClientNames, id)
}

/** GUI fields, not a configuration file. No client-specific endpoint override. */
export function desktopClientFields(client: DesktopClientId, baseUrl: string, apiKey: string, model = 'YOUR_MODEL_ID', platform?: GroupPlatform | null) {
  const endpoints = clientEndpoints(baseUrl, platform)
  const native = client === 'cherry-studio' && usesNativeMessages(platform)
  if (client === 'cursor') return [
    { path: 'API Key', content: apiKey },
    { path: 'Override OpenAI Base URL', content: endpoints.openai },
    { path: 'Model ID', content: model }
  ]
  return [
    { path: 'Provider', content: native ? 'Anthropic' : client === 'cline' || client === 'roo-code' ? 'OpenAI Compatible' : 'OpenAI' },
    { path: client === 'cherry-studio' ? 'API Host' : 'Base URL', content: native ? endpoints.messagesRoot : endpoints.openai },
    { path: 'API Key', content: apiKey },
    { path: 'Model ID', content: model }
  ]
}
