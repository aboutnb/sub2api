const nonChat = /image|embedding|whisper|tts|audio|video|moderation|rerank|realtime|auto-review/i
export function sortHelpModels(ids: string[]): string[] {
  const version = (id: string) => id.match(/(?:gpt|claude|gemini|grok|deepseek|glm|kimi|minimax)[^0-9]*([0-9]+(?:[.-][0-9]+)?)/i)?.[1]?.replace('-', '.') || ''
  return [...new Set(ids)].filter(id => /^[a-zA-Z0-9][a-zA-Z0-9_./:+-]{0,159}$/.test(id) && !nonChat.test(id)).sort((a, b) => {
    const av = version(a), bv = version(b)
    if (av && bv) {
      const result = bv.localeCompare(av, 'en', { numeric: true })
      if (result) return result
    }
    return b.localeCompare(a, 'en', { numeric: true })
  })
}
