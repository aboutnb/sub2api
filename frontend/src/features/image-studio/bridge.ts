import apiClient from '@/api/client'
import { STUDIO_CHANNEL, isStudioRequest } from '../../../../shared/image-studio'
import type { StudioThemeTokens } from '../../../../shared/image-studio-theme'

export function installStudioBridge(frame: HTMLIFrameElement, active: () => boolean, context: () => { theme: string; lang: string; tokens?: StudioThemeTokens }) {
  const requests = new Map<string, AbortController>()
  const seen = new Set<string>()
  let disposed = false
  async function receive(event: MessageEvent) {
    if (disposed || !active() || event.origin !== window.location.origin || event.source !== frame.contentWindow || !isStudioRequest(event.data)) return
    const request = event.data
    if (request.operation === 'cancel') { requests.get(request.id)?.abort(); return }
    if (seen.has(request.id)) return
    seen.add(request.id)
    if (requests.size >= 32) {
      frame.contentWindow?.postMessage({ channel: STUDIO_CHANNEL, id: request.id, status: 429, data: { error: { message: 'Too many image requests' } } }, window.location.origin)
      return
    }
    const controller = new AbortController()
    requests.set(request.id, controller)
    const target = frame.contentWindow
    const send = (data: unknown) => {
      if (!disposed && active() && !controller.signal.aborted && target === frame.contentWindow) {
        target?.postMessage({ channel: STUDIO_CHANNEL, id: request.id, ...data as object }, window.location.origin)
      }
    }
    try {
      if (request.operation === 'bootstrap') {
        const { data } = await apiClient.get('/image-studio/bootstrap', { signal: controller.signal })
        send({ status: 200, data: { ...data, ...context() } })
        return
      }
      let body: string | FormData | undefined
      if (Array.isArray(request.body)) {
        body = new FormData()
        for (const [name, value] of request.body) body.append(name, value)
      } else body = request.body
      const result = await apiClient.request({
        url: request.path, method: request.method, data: body, signal: controller.signal,
        timeout: 31 * 60 * 1000,
        // Preserve Images error bodies, while retaining the site's session refresh on 401.
        validateStatus: (status) => status !== 401,
        headers: { 'Content-Type': body instanceof FormData ? undefined : 'application/json' },
      })
      send({ status: result.status, data: result.data })
    } catch (error) {
      const failure = error as { status?: number; message?: string }
      send({ status: failure.status || 502, data: { error: { message: failure.message || 'Image request failed' } } })
    } finally { requests.delete(request.id) }
  }
  window.addEventListener('message', receive)
  return () => {
    disposed = true
    window.removeEventListener('message', receive)
    for (const request of requests.values()) request.abort()
    requests.clear()
  }
}
