import { describe, expect, it } from 'vitest'
import { buildHelpApiRequest, detectHelpModelPlatform, helpApiProtocols, maskApiKey, readHelpApiBody } from '../helpApiProbe'

describe('help API probe protocols', () => {
  it('detects composite model platforms without guessing', () => {
    expect(detectHelpModelPlatform('claude-sonnet-4')).toBe('anthropic')
    expect(detectHelpModelPlatform('models/gemini-2.5-pro')).toBe('gemini')
    expect(detectHelpModelPlatform('openai/gpt-4o')).toBe('openai')
    expect(detectHelpModelPlatform('google/gemini-2.5-pro')).toBe('gemini')
    expect(detectHelpModelPlatform('o3-mini')).toBe('openai')
    expect(detectHelpModelPlatform('grok-4')).toBe('grok')
    expect(detectHelpModelPlatform('kimi-k2')).toBe('kimi')
    expect(detectHelpModelPlatform('glm-4')).toBe('zhipu')
    expect(detectHelpModelPlatform('deepseek-chat')).toBe('deepseek')
    expect(detectHelpModelPlatform('abab6.5')).toBe('minimax')
    expect(detectHelpModelPlatform('custom-model')).toBeNull()
    expect(detectHelpModelPlatform('')).toBeNull()
  })

  it.each([
    ['anthropic', false, ['messages', 'responses', 'chat'], 'messages'],
    ['openai', false, ['responses', 'chat'], 'responses'],
    ['openai', true, ['responses', 'chat', 'messages'], 'responses'],
    ['gemini', false, ['gemini', 'responses'], 'gemini'],
    ['antigravity', false, ['messages', 'gemini'], 'messages'],
    ['grok', false, ['responses', 'messages', 'chat'], 'responses'],
    ['deepseek', false, ['messages', 'responses', 'chat'], 'messages'],
    ['minimax', false, ['messages', 'responses', 'chat'], 'messages'],
    ['kimi', false, ['messages', 'responses', 'chat'], 'messages'],
    ['zhipu', false, ['messages', 'responses', 'chat'], 'messages']
  ] as const)('selects %s protocols', (platform, allowMessages, protocols, fallback) => {
    expect(helpApiProtocols({ platform, allow_messages_dispatch: allowMessages })).toEqual({ protocols, defaultProtocol: fallback })
  })

  it('keeps claude code groups on Messages only', () => {
    expect(helpApiProtocols({ platform: 'openai', claude_code_only: true, allow_messages_dispatch: true })).toEqual({ protocols: ['messages'], defaultProtocol: 'messages' })
    expect(helpApiProtocols({ platform: 'composite', claude_code_only: true }, 'gpt-5')).toEqual({ protocols: ['messages'], defaultProtocol: 'messages' })
  })

  it('uses the detected composite platform and does not guess an unknown model', () => {
    expect(helpApiProtocols({ platform: 'composite' }, 'claude-sonnet-4').defaultProtocol).toBe('messages')
    expect(helpApiProtocols({ platform: 'composite' }, 'gpt-5').protocols).toEqual(['responses', 'chat'])
    expect(helpApiProtocols({ platform: 'composite', allow_messages_dispatch: true }, 'gpt-5').protocols).toEqual(['responses', 'chat', 'messages'])
    expect(helpApiProtocols({ platform: 'composite' }, 'gemini-2.5-pro').defaultProtocol).toBe('gemini')
    expect(helpApiProtocols({ platform: 'composite' }, 'custom-model')).toEqual({
      protocols: ['messages', 'responses', 'chat', 'gemini'],
      defaultProtocol: null
    })
  })
})

describe('help API probe requests', () => {
  const key = 'sk-secret-value-1234'

  it('builds a masked non-streaming Messages request', () => {
    const request = buildHelpApiRequest({ baseUrl: 'https://gateway.example/v1/', platform: 'anthropic', protocol: 'messages', model: 'claude-sonnet', apiKey: key })
    expect(request.url).toBe('https://gateway.example/v1/messages')
    expect(request.headers['x-api-key']).toBe(key)
    expect(request.headers['anthropic-version']).toBe('2023-06-01')
    expect(request.headers.Authorization).toBeUndefined()
    expect(request.body).toEqual({ model: 'claude-sonnet', max_tokens: 16, messages: [{ role: 'user', content: 'hi' }], stream: false })
    expect(request.preview).toContain('****1234')
    expect(request.preview).not.toContain(key)
  })

  it('builds Responses, Chat, Gemini, and Antigravity paths', () => {
    const responses = buildHelpApiRequest({ baseUrl: 'https://gateway.example/gateway/v1', platform: 'openai', protocol: 'responses', model: 'gpt-5', apiKey: key })
    expect(responses.url).toBe('https://gateway.example/gateway/v1/responses')
    expect(responses.headers.Authorization).toBe(`Bearer ${key}`)
    expect(responses.body).toEqual({ model: 'gpt-5', input: 'hi', stream: false })
    expect(responses.preview).toContain('Bearer ****1234')
    expect(responses.preview).not.toContain(key)

    const chat = buildHelpApiRequest({ baseUrl: 'https://gateway.example', platform: 'deepseek', protocol: 'chat', model: 'deepseek-chat', apiKey: key })
    expect(chat.url).toBe('https://gateway.example/v1/chat/completions')
    expect(chat.body).toMatchObject({ max_tokens: 16, stream: false })

    const gemini = buildHelpApiRequest({ baseUrl: 'https://gateway.example', platform: 'gemini', protocol: 'gemini', model: 'models/gemini-2.5-pro', apiKey: key })
    expect(gemini.url).toBe('https://gateway.example/v1beta/models/gemini-2.5-pro:generateContent')
    expect(gemini.headers['x-goog-api-key']).toBe(key)
    expect(gemini.body).toEqual({ contents: [{ role: 'user', parts: [{ text: 'hi' }] }] })
    expect(gemini.preview).not.toContain(key)

    const antigravityMessages = buildHelpApiRequest({ baseUrl: 'https://gateway.example', platform: 'antigravity', protocol: 'messages', model: 'claude-sonnet', apiKey: key })
    expect(antigravityMessages.url).toBe('https://gateway.example/antigravity/v1/messages')
    const antigravityGemini = buildHelpApiRequest({ baseUrl: 'https://gateway.example', platform: 'antigravity', protocol: 'gemini', model: 'foo/bar', apiKey: key })
    expect(antigravityGemini.url).toBe('https://gateway.example/antigravity/v1beta/models/foo%2Fbar:generateContent')
  })

  it('masks short keys and quotes model text', () => {
    expect(maskApiKey('abcd')).toBe('****')
    const request = buildHelpApiRequest({ baseUrl: 'https://gateway.example', platform: 'openai', protocol: 'responses', model: "o'hara", apiKey: 'abcd' })
    expect(request.preview).toContain(`o'\\''hara`)
    expect(request.preview).not.toContain('abcd')
  })

  it('truncates response bodies at 8KB', async () => {
    const full = await readHelpApiBody(new Response('ok'))
    expect(full).toEqual({ text: 'ok', truncated: false })
    const limited = await readHelpApiBody(new Response('x'.repeat(9000)), 8)
    expect(limited.truncated).toBe(true)
    expect(limited.text).toBe('xxxxxxxx')
  })
})
