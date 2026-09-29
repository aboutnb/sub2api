import { describe, expect, it } from 'vitest'
import { extendedClients, extendedClientFiles, extendedClientIds } from '../extendedClients'
import { sortHelpModels } from '../helpModels'
import { readFileSync, existsSync } from 'node:fs'
import { resolve } from 'node:path'

describe('extended client catalog', () => {
  it('covers all ten missing clients with model and default endpoint in each configuration', () => {
    expect(Object.keys(extendedClients)).toHaveLength(12)
    expect(extendedClientIds('openai')).not.toEqual(expect.arrayContaining(['sillytavern', 'tavernai']))
    for (const client of extendedClientIds('openai')) {
      const files = extendedClientFiles(client, 'openai', 'https://site.example/gateway/v1/', 'YOUR_API_KEY', 'gpt-6-astra')
      const text = files.map(f => f.content).join('\n')
      expect(text).toContain('https://site.example/gateway/v1')
      expect(text).toContain('gpt-6-astra')
      expect(text).toContain('YOUR_API_KEY')
      expect(text).not.toContain('/v1/v1')
      expect(text).not.toContain('krill')
      if (client === 'pi' || client === 'openclaw') expect(() => JSON.parse(files[0]!.content)).not.toThrow()
    }
  })
  it('distinguishes full endpoint URLs, provider base URLs, and Anthropic paths', () => {
    expect(extendedClientFiles('trae', 'anthropic', 'https://site.example/v1', 'KEY')[1]?.content).toBe('https://site.example')
    expect(extendedClientFiles('trae', 'antigravity', 'https://site.example', 'KEY')[1]?.content).toBe('https://site.example/antigravity')
    expect(extendedClientFiles('trae', 'deepseek', 'https://site.example/v1', 'KEY')[0]?.content).toBe('Anthropic Messages')
    expect(extendedClientFiles('trae', 'grok', 'https://site.example', 'KEY').find(file => file.path === '完整 URL')?.content).toBe('OFF')
    expect(extendedClientFiles('workbuddy', 'openai', 'https://site.example', 'KEY')[1]?.content).toBe('https://site.example/v1/chat/completions')
    expect(extendedClientFiles('workbuddy', 'openai', 'https://site.example', 'KEY').find(file => file.path === 'Custom Protocol')?.content).toBe('OFF')
    expect(extendedClientFiles('zcode', 'anthropic', 'https://site.example/v1', 'KEY').find(file => file.path === 'Base URL')?.content).toBe('https://site.example/v1')
    expect(extendedClientFiles('zcode', 'openai', 'https://site.example', 'KEY').find(file => file.path === 'Provider')?.content).toBe('site')
    expect(extendedClientFiles('read-frog', 'openai', 'https://site.example', 'KEY')[1]?.content).toBe('https://site.example/v1')
    expect(extendedClientIds('grok')).not.toContain('pi')
    for (const platform of ['openai', 'anthropic', 'antigravity', 'gemini', 'grok', 'deepseek'] as const) {
      expect(extendedClientFiles('sillytavern', platform, 'https://site.example/v1/', 'YOUR_API_KEY', 'selected').find(file => file.path === 'Custom Endpoint (Base URL)')?.content).toBe('https://site.example/v1')
      const tavern = extendedClientFiles('tavernai', platform, 'https://site.example', 'YOUR_API_KEY', 'selected')
      expect(tavern.find(file => file.path === 'API Address')?.content).toBe('https://site.example/v1')
      expect(tavern.find(file => file.path === 'Model Endpoint')?.content).toBe('Chat Completions')
      expect(tavern.map(file => file.content).join('\n')).not.toContain('/chat/completions')
    }
  })

  it('selects OpenClaw and Hermes protocols from the current platform', () => {
    const openclaw = (platform: 'openai' | 'anthropic' | 'antigravity' | 'grok') =>
      JSON.parse(extendedClientFiles('openclaw', platform, 'https://site.example/v1/', 'YOUR_API_KEY', 'selected')[0]!.content)
    expect(openclaw('openai').models.providers.site).toMatchObject({ api: 'openai-responses', baseUrl: 'https://site.example/v1' })
    expect(openclaw('anthropic').models.providers.site).toMatchObject({ api: 'anthropic-messages', baseUrl: 'https://site.example' })
    expect(openclaw('antigravity').models.providers.site.baseUrl).toBe('https://site.example/antigravity')
    expect(openclaw('grok').models.providers.site.api).toBe('openai-completions')

    const hermes = (platform: 'openai' | 'anthropic' | 'grok') => extendedClientFiles('hermes', platform, 'https://site.example/v1', 'YOUR_API_KEY', 'selected').map(file => file.content).join('\n')
    expect(hermes('openai')).toContain('api_mode: codex_responses')
    expect(hermes('openai')).toContain('https://site.example/v1')
    expect(hermes('anthropic')).toContain('api_mode: anthropic_messages')
    expect(hermes('anthropic')).toContain('base_url: "https://site.example"')
    expect(hermes('anthropic')).not.toContain('/v1')
    expect(hermes('grok')).not.toContain('api_mode')
    expect(hermes('grok')).toContain('YOUR_API_KEY')
  })
  it('sorts only safe text-model identifiers without mutation or fictional fallback models', () => {
    const ids = ['gpt-5.5', 'gpt-6-astra', 'gpt-image-2', 'gpt-6-astra', 'bad"\ncommand']
    expect(sortHelpModels(ids)).toEqual(['gpt-6-astra', 'gpt-5.5'])
    expect(ids).toHaveLength(5)
    expect(sortHelpModels([])).toEqual([])
  })
  it('uses Anthropic translation providers only on explicit native Messages platforms', () => {
    for (const platform of ['anthropic', 'antigravity', 'deepseek', 'minimax', 'openai', 'grok', 'gemini', 'composite'] as const) {
      const native = ['anthropic', 'antigravity', 'deepseek', 'minimax'].includes(platform)
      for (const client of ['read-frog', 'kiss-translator', 'immersive-translate'] as const) {
        const fields = extendedClientFiles(client, platform, 'https://site.example/v1/', 'PLACEHOLDER', 'selected')
        expect(fields[0]?.content).toBe(native ? client === 'read-frog' ? 'Anthropic' : 'Claude' : client === 'read-frog' ? 'OpenAI Compatible' : 'OpenAI')
        const prefix = 'https://site.example' + (native && platform === 'antigravity' ? '/antigravity' : '')
        expect(fields[1]?.content).toBe(prefix + (client === 'read-frog' ? '/v1' : native ? '/v1/messages' : '/v1/chat/completions'))
      }
    }
  })
  it('bundles every referenced logo locally with no scripts or remote image dependency', () => {
    const component = readFileSync(resolve('src/components/common/ClientLogo.vue'), 'utf8')
    const assets = [...component.matchAll(/'([^']+\.(?:svg|png))'/g)].map(match => match[1]!)
    expect(new Set(assets).size).toBe(21)
    for (const asset of assets) {
      const path = resolve('public/client-logos', asset)
      expect(existsSync(path)).toBe(true)
      if (asset.endsWith('.svg')) expect(readFileSync(path, 'utf8')).not.toMatch(/<script|onload=|javascript:/i)
    }
  })
})
