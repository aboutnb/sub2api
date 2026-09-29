import { computed, ref } from 'vue'
import { describe, expect, it } from 'vitest'
import { desktopClientFields } from '../desktopClients'
import { supportedClientIds, useClientConfiguration } from '@/composables/useClientConfiguration'

describe('desktop client configuration', () => {
  it.each(['https://site.example', 'https://site.example/', 'https://site.example/v1', 'https://site.example/v1/'])('normalizes the default endpoint without duplicate protocol paths: %s', baseUrl => {
    expect(desktopClientFields('cline', baseUrl, 'PLACEHOLDER')[1]?.content).toBe('https://site.example/v1')
  })
  it('shares the exact fields between the key modal and help generator for every new client', () => {
    for (const client of ['cherry-studio', 'cursor', 'cline', 'roo-code'] as const) {
      const context = { platform: 'openai' as const, baseUrl: 'https://site.example/gateway/v1/', apiKey: 'YOUR_API_KEY' }
      const { currentFiles } = useClientConfiguration(context, computed(() => 'unix'), ref(client), ref('legacy'), ref(''), key => key)
      expect(currentFiles.value).toEqual(desktopClientFields(client, context.baseUrl, context.apiKey))
      expect(currentFiles.value.find(file => file.path === 'Base URL' || file.path === 'API Host' || file.path === 'Override OpenAI Base URL')?.content).toBe('https://site.example/gateway/v1')
      if (client === 'cursor') expect(currentFiles.value.map(file => file.path)).not.toContain('Provider')
    }
    expect(supportedClientIds(null)).toEqual([])
    expect(supportedClientIds('grok')).not.toContain('cursor')
    expect(supportedClientIds('openai')).toContain('cursor')
  })
  it('keeps native Messages and OpenAI address rules consistent in shared configuration', () => {
    for (const platform of ['anthropic', 'antigravity', 'deepseek', 'minimax', 'openai', 'grok', 'gemini', 'composite'] as const) {
      const context = { platform, baseUrl: 'https://site.example/v1/', apiKey: 'YOUR_API_KEY', model: 'selected' }
      const { currentFiles } = useClientConfiguration(context, computed(() => 'unix'), ref('cherry-studio'), ref('legacy'), ref(''), key => key)
      expect(currentFiles.value).toEqual(desktopClientFields('cherry-studio', context.baseUrl, context.apiKey, context.model, platform))
      const native = ['anthropic', 'antigravity', 'deepseek', 'minimax'].includes(platform)
      expect(currentFiles.value[0]?.content).toBe(native ? 'Anthropic' : 'OpenAI')
      expect(currentFiles.value[1]?.content).toBe(native ? 'https://site.example' + (platform === 'antigravity' ? '/antigravity' : '') : 'https://site.example/v1')
    }
  })
})
