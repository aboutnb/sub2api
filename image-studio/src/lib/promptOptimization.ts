export interface StudioPromptOptimization {
  prompt: string
  optimized: boolean
  model?: string
  group?: string
}

function text(value: unknown) {
  return typeof value === 'string' && value.trim() ? value.trim() : undefined
}

export function parseStudioPromptOptimization(original: string, status: number, data: unknown): StudioPromptOptimization {
  const body = data && typeof data === 'object' ? data as { prompt?: unknown; optimized?: unknown; model?: unknown; group?: unknown; error?: { message?: unknown } } : {}
  const returned = typeof body.prompt === 'string' ? body.prompt.trim() : ''
  if (status !== 200 || !returned) {
    const message = text(body.error?.message)
    throw new Error(message || '提示词优化失败')
  }
  const optimized = returned !== original.trim() && body.optimized !== false
  return {
    prompt: optimized ? returned : original,
    optimized,
    model: optimized ? text(body.model) : undefined,
    group: optimized ? text(body.group) : undefined,
  }
}
