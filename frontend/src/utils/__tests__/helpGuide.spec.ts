import { describe, expect, it } from 'vitest'
import { clientInstall, clientPrepare } from '../helpGuide'

describe('client tutorial commands', () => {
  it('uses official Grok Build installers for the selected OS', () => {
    expect(clientInstall('grok', 'linux', 'bash')).toBe('curl -fsSL https://x.ai/cli/install.sh | bash')
    expect(clientInstall('grok', 'windows', 'powershell')).toBe('irm https://x.ai/cli/install.ps1 | iex')
  })
  it.each(['codex', 'codex-ws', 'grok'])('creates the correct user directory for %s without replacing files', client => {
    const directory = client === 'grok' ? '.grok' : '.codex'
    expect(clientPrepare(client, 'linux', 'bash')).toBe(`mkdir -p ~/${directory}`)
    expect(clientPrepare(client, 'windows', 'powershell')).toBe(`New-Item -ItemType Directory -Force -Path "$HOME\\${directory}" | Out-Null`)
  })
  it('uses the client install command captured from the live tutorial', () => {
    expect(clientInstall('pi', 'linux', 'zsh')).toBe('npm install -g --ignore-scripts @earendil-works/pi-coding-agent')
    expect(clientInstall('dsh', 'windows', 'powershell')).toBe('npm install -g @deepseek-ai/dsh')
    expect(clientInstall('hermes', 'linux', 'bash')).toContain('hermes-agent/main/scripts/install.sh')
    expect(clientInstall('hermes', 'windows', 'powershell')).toBe('')
    expect(clientPrepare('pi', 'linux', 'bash')).toBe('mkdir -p ~/.pi/agent')
  })
  it('clears conflicting Gemini variables using the selected shell syntax', () => {
    expect(clientPrepare('gemini', 'linux', 'zsh')).toBe('unset GOOGLE_API_KEY GOOGLE_GENAI_USE_VERTEXAI')
    expect(clientPrepare('gemini', 'windows', 'cmd')).toBe('set GOOGLE_API_KEY=\nset GOOGLE_GENAI_USE_VERTEXAI=')
    expect(clientPrepare('gemini', 'windows', 'powershell')).toContain('Remove-Item Env:GOOGLE_API_KEY, Env:GOOGLE_GENAI_USE_VERTEXAI')
  })
})
