import { Lexer } from 'marked'

export function helpSections(markdown: string) {
  const sections: { title: string; markdown: string }[] = []
  for (const token of Lexer.lex(markdown)) {
    if (token.type === 'heading' && token.depth === 2) sections.push({ title: token.text, markdown: '' })
    else {
      if (!sections.length) sections.push({ title: '', markdown: '' })
      sections[sections.length - 1]!.markdown += token.raw
    }
  }
  return sections.filter(section => section.title || section.markdown.trim())
}

export function clientInstall(client: string, os: string, shell: string) {
  if (client === 'grok') return os === 'windows' ? 'irm https://x.ai/cli/install.ps1 | iex' : 'curl -fsSL https://x.ai/cli/install.sh | bash'
  if (client === 'hermes' && os !== 'windows') return 'curl -fsSL https://raw.githubusercontent.com/NousResearch/hermes-agent/main/scripts/install.sh | bash'
  if (client === 'claude') return os === 'windows'
    ? shell === 'cmd' ? 'curl -fsSL https://claude.ai/install.cmd -o install.cmd && install.cmd && del install.cmd' : 'irm https://claude.ai/install.ps1 | iex'
    : 'curl -fsSL https://claude.ai/install.sh | bash'
  return ({
    codex: 'npm install -g @openai/codex', 'codex-ws': 'npm install -g @openai/codex', opencode: 'npm install -g opencode-ai', gemini: 'npm install -g @google/gemini-cli',
    pi: 'npm install -g --ignore-scripts @earendil-works/pi-coding-agent', dsh: 'npm install -g @deepseek-ai/dsh'
  } as Record<string, string>)[client] || ''
}

export function clientPrepare(client: string, os: string, shell: string) {
  if (client === 'pi') {
    return os === 'windows'
      ? 'New-Item -ItemType Directory -Force -Path "$HOME\\.pi\\agent" | Out-Null'
      : 'mkdir -p ~/.pi/agent'
  }
  if (client.startsWith('codex') || client === 'grok') {
    const directory = client === 'grok' ? '.grok' : '.codex'
    return os === 'windows'
      ? `New-Item -ItemType Directory -Force -Path "$HOME\\${directory}" | Out-Null`
      : `mkdir -p ~/${directory}`
  }
  if (client === 'gemini') {
    if (os !== 'windows') return 'unset GOOGLE_API_KEY GOOGLE_GENAI_USE_VERTEXAI'
    return shell === 'cmd' ? 'set GOOGLE_API_KEY=\nset GOOGLE_GENAI_USE_VERTEXAI=' : 'Remove-Item Env:GOOGLE_API_KEY, Env:GOOGLE_GENAI_USE_VERTEXAI -ErrorAction SilentlyContinue'
  }
  return ''
}

export function clientCommand(client: string) {
  return ({ claude: 'claude', codex: 'codex', 'codex-ws': 'codex', opencode: 'opencode', gemini: 'gemini', grok: 'grok', pi: 'pi', dsh: 'dsh', hermes: 'hermes', openclaw: 'openclaw' } as Record<string, string>)[client] || ''
}
