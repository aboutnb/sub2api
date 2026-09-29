import { describe, expect, it } from 'vitest'
import { matchesHelpSelectors } from '../helpSelectors'

describe('matchesHelpSelectors', () => {
  it('allows unrestricted documents and ignores empty selectors', () => {
    expect(matchesHelpSelectors({}, { platform: 'openai', client: 'codex' })).toBe(true)
    expect(matchesHelpSelectors({ clients: [] }, { client: 'codex' })).toBe(true)
  })

  it('matches every declared selector dimension', () => {
    const schema = { platforms: ['openai'], clients: ['codex'], systems: ['windows'], shells: ['powershell'] }
    expect(matchesHelpSelectors(schema, { platform: 'openai', client: 'codex', system: 'windows', shell: 'powershell' })).toBe(true)
    expect(matchesHelpSelectors(schema, { platform: 'openai', client: 'claude', system: 'windows', shell: 'powershell' })).toBe(false)
    expect(matchesHelpSelectors(schema, { platform: 'openai', client: 'codex', system: 'linux', shell: 'bash' })).toBe(false)
  })

  it('does not treat malformed selector values as a match', () => {
    expect(matchesHelpSelectors({ clients: ['codex', 42] }, { client: 'claude' })).toBe(false)
    expect(matchesHelpSelectors({ clients: 'codex' }, { client: 'codex' })).toBe(true)
  })
})
