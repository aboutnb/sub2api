import { STUDIO_CHANNEL, isStudioRequest } from '../../../shared/image-studio'
import type { StudioBootstrap, StudioRequest } from '../../../shared/image-studio'
import { setStudioLocale } from './studioLocale'
import { STUDIO_THEME_ROLES, validStudioColor, type StudioThemeTokens } from '../../../shared/image-studio-theme'

type Bootstrap = StudioBootstrap & { theme: string; lang: string; tokens?: StudioThemeTokens }
let bootstrap: Bootstrap | undefined
const pending = new Map<string, { resolve: (value: { status: number; data: unknown }) => void; reject: (error: Error) => void }>()

function appearance(theme: string, lang: string, tokens?: StudioThemeTokens) {
  document.documentElement.classList.toggle('dark', theme === 'dark')
  document.documentElement.style.colorScheme = theme === 'dark' ? 'dark' : 'light'
  document.documentElement.lang = lang
  setStudioLocale(lang)
  for (const role of STUDIO_THEME_ROLES) {
    const value = tokens?.[role]
    if (validStudioColor(value)) document.documentElement.style.setProperty(`--av-rgb-${role}`, value)
    else document.documentElement.style.removeProperty(`--av-rgb-${role}`)
  }
}

if (typeof window !== 'undefined') window.addEventListener('message', (event) => {
  if (event.source !== window.parent || event.origin !== window.location.origin || event.data?.channel !== STUDIO_CHANNEL) return
  const message = event.data
  if (message.operation === 'appearance') {
    appearance(message.theme, message.lang, message.tokens)
    return
  }
  const request = pending.get(message.id)
  if (!request || !Number.isInteger(message.status)) return
  request.resolve({ status: message.status, data: message.data })
})

async function rpc(input: Omit<StudioRequest, 'channel' | 'id'>, signal?: AbortSignal) {
  if (window.parent === window) throw new Error('请从本站 AI 绘图入口打开')
  const id = crypto.randomUUID()
  if (signal?.aborted) throw new DOMException('Aborted', 'AbortError')
  let abort: () => void = () => {}
  let timeout: ReturnType<typeof setTimeout>
  try {
    return await new Promise<{ status: number; data: unknown }>((resolve, reject) => {
      pending.set(id, { resolve, reject })
      abort = () => {
        window.parent.postMessage({ channel: STUDIO_CHANNEL, id, operation: 'cancel' }, window.location.origin)
        reject(new DOMException('Aborted', 'AbortError'))
      }
      signal?.addEventListener('abort', abort, { once: true })
      timeout = setTimeout(abort, input.operation === 'bootstrap' ? 30000 : 31 * 60 * 1000)
      window.parent.postMessage({ ...input, channel: STUDIO_CHANNEL, id }, window.location.origin)
    })
  } finally {
    clearTimeout(timeout!)
    signal?.removeEventListener('abort', abort)
    pending.delete(id)
  }
}

export async function initializeStudio() {
  const result = await rpc({ operation: 'bootstrap' })
  if (result.status !== 200) throw new Error('无法加载绘图配置，请重新登录本站')
  const value = result.data as Bootstrap
  if (!/^image-studio-user-\d+$/.test(value.storage_namespace) || !Array.isArray(value.groups)) throw new Error('绘图配置无效')
  bootstrap = value
  appearance(value.theme, value.lang, value.tokens)
  return value
}

export function studioConfig() { return bootstrap }
export function studioStorageNamespace() { return bootstrap?.storage_namespace ?? 'image-studio-uninitialized' }

export function studioCapabilities(profile: { id: string; model: string }) {
  const groupID = Number(/^studio:(\d+):/.exec(profile.id)?.[1])
  return bootstrap?.groups.find((group) => group.id === groupID)?.models.find((model) => model.id === profile.model)?.capabilities
}

// API 请求只通过父页面；不会转发自定义域名或客户端凭证。
export async function studioFetch(input: string | URL | Request, init?: RequestInit): Promise<Response> {
  if (!bootstrap) return globalThis.fetch(input, init)
  const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.origin)
  if (url.origin !== window.location.origin || url.search || url.hash) throw new Error('Unsupported image endpoint')
  const body = init?.body instanceof FormData ? [...init.body.entries()] : init?.body
  const request = { channel: STUDIO_CHANNEL, id: 'validation', operation: 'request', path: url.pathname, method: init?.method || 'GET', body }
  if (!isStudioRequest(request)) throw new Error('Unsupported image request')
  if (request.method === 'POST') {
    const model = typeof request.body === 'string' ? JSON.parse(request.body).model : request.body?.find(([name]) => name === 'model')?.[1]
    const groupID = Number(/^\/image-studio\/groups\/(\d+)\//.exec(request.path)?.[1])
    const capabilities = bootstrap.groups.find((group) => group.id === groupID)?.models.find((item) => item.id === model)?.capabilities
    if (Array.isArray(request.body) && capabilities) {
      const files = request.body.filter((entry): entry is [string, Blob] => entry[1] instanceof Blob)
      if (files.filter(([name]) => name !== 'mask').length > capabilities.max_input_images) throw new Error('参考图数量超过当前模型限制')
      for (const [, file] of files) {
        if (!(file instanceof Blob)) continue
        if (file.size > capabilities.max_upload_bytes) throw new Error('单张图片超过本站上传限制')
        if (!capabilities.input_types.includes(file.type)) throw new Error('当前模型仅支持 PNG、JPEG、WebP 图片')
      }
    }
  }
  if (bootstrap?.max_body_bytes && request.body) {
    const size = typeof request.body === 'string' ? new Blob([request.body]).size : request.body.reduce((sum, [, value]) => sum + (typeof value === 'string' ? new Blob([value]).size : value.size) + 256, 1024)
    if (size > bootstrap.max_body_bytes) throw new Error('图片请求超过本站上传限制')
  }
  const result = await rpc(request, init?.signal ?? undefined)
  return new Response(JSON.stringify(result.data), { status: result.status, headers: { 'Content-Type': 'application/json' } })
}
