import { describe, expect, it } from 'vitest'
import { githubHighlighterReady, highlightGithubCode, inferCodeLanguage } from '../githubCode'

describe('github code highlighting', () => {
  it('infers a language from a filename or label', () => {
    expect(inferCodeLanguage('~/.codex/config.toml')).toBe('toml')
    expect(inferCodeLanguage('PowerShell')).toBe('powershell')
    expect(inferCodeLanguage('curl')).toBe('shell')
    expect(inferCodeLanguage('Command Prompt')).toBe('cmd')
    expect(inferCodeLanguage('code')).toBe('text')
  })

  it('uses GitHub light and dark token colors and escapes markup', async () => {
    await githubHighlighterReady()
    const html = highlightGithubCode('{\n  "model": "<gpt>",\n  "ok": true\n}\n', 'json')
    expect(html).toContain('--shiki-light:#116329')
    expect(html).toContain('--shiki-dark:#7EE787')
    expect(html).toContain('--shiki-light:#0A3069')
    expect(html).toContain('--shiki-dark:#A5D6FF')
    expect(html).toContain('gpt')
    expect(html).not.toContain('<gpt>')
  })

  it('highlights shell commands without treating quoted text as comments', async () => {
    await githubHighlighterReady()
    const html = highlightGithubCode('echo "keep # this"\n# comment\n', 'sh')
    expect(html).toContain('keep # this')
    expect(html).toMatch(/--shiki-light:#0A3069[^>]*>\s*"keep # this"/)
    expect(html).toMatch(/--shiki-light:#6E7781[^>]*># comment/)
  })
})
