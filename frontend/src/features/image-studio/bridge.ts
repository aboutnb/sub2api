import apiClient from '@/api/client'
import axios, { CanceledError } from 'axios'
import { STUDIO_CHANNEL, isStudioRequest } from '../../../../shared/image-studio'
import type { StudioThemeTokens } from '../../../../shared/image-studio-theme'

function storedUserID(): number | null {
  try {
    const token = localStorage.getItem('auth_token')
    if (!token) return null
    const user = JSON.parse(localStorage.getItem('auth_user') || 'null')
    const claims = JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')))
    // This checks account continuity only; the server still authenticates every JWT.
    return Number.isSafeInteger(user?.id) && user.id > 0 && claims.user_id === user.id ? user.id : null
  } catch {
    return null
  }
}

export function installStudioBridge(
  frame: HTMLIFrameElement,
  active: () => boolean,
  context: () => { theme: string; lang: string; tokens?: StudioThemeTokens },
  onSessionChange: () => void = () => {},
) {
  const userID = storedUserID()
  const requests = new Map<string, AbortController>()
  const seen = new Set<string>()
  let disposed = false
  function dispose() {
    disposed = true
    window.removeEventListener('message', receive)
    window.removeEventListener('storage', checkStoredSession)
    for (const request of requests.values()) request.abort()
    requests.clear()
  }
  function sessionIsCurrent() {
    if (disposed || !active()) return false
    if (userID !== null && storedUserID() === userID) return true
    dispose()
    onSessionChange()
    return false
  }
  function checkStoredSession(event: StorageEvent) {
    if (event.storageArea && event.storageArea !== localStorage) return
    if (event.key === null || event.key === 'auth_user' || event.key === 'auth_token') sessionIsCurrent()
  }
  const defaultTransforms = axios.defaults.transformRequest
  const transformRequest = [
    ...(Array.isArray(defaultTransforms) ? defaultTransforms : defaultTransforms ? [defaultTransforms] : []),
    (data: unknown) => {
      // Axios attaches the current token asynchronously. Recheck immediately before sending.
      if (!sessionIsCurrent()) throw new CanceledError('Image studio session changed')
      return data
    },
  ]
  async function receive(event: MessageEvent) {
    if (event.origin !== window.location.origin || event.source !== frame.contentWindow || !isStudioRequest(event.data) || !sessionIsCurrent()) return
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
      if (sessionIsCurrent() && !controller.signal.aborted && target === frame.contentWindow) {
        target?.postMessage({ channel: STUDIO_CHANNEL, id: request.id, ...data as object }, window.location.origin)
      }
    }
    try {
      if (request.operation === 'bootstrap') {
        const { data } = await apiClient.get('/image-studio/bootstrap', { signal: controller.signal, transformRequest })
        send({ status: 200, data: { ...data, ...context() } })
        return
      }
      let body: string | FormData | undefined
      if (Array.isArray(request.body)) {
        body = new FormData()
        for (const [name, value] of request.body) body.append(name, value)
      } else body = request.body
      const result = await apiClient.request({
        url: request.path, method: request.method, data: body, signal: controller.signal, transformRequest,
        timeout: 31 * 60 * 1000,
        // Preserve Images error bodies, while retaining the site's session refresh on 401.
        validateStatus: (status) => status !== 401,
        headers: { 'Content-Type': body instanceof FormData ? undefined : 'application/json' },
      })
      send({ status: result.status, data: result.data })
    } catch (error) {
      const failure = error as { status?: number; message?: string; code?: string; reason?: string; error?: { message?: string; code?: string; type?: string } }
      send({
        status: failure.status && failure.status >= 400 ? failure.status : 502,
        data: { error: {
          message: failure.error?.message || failure.message || 'Image request failed',
          code: failure.error?.code || failure.code,
          type: failure.error?.type,
          reason: failure.reason,
        } },
      })
    } finally { requests.delete(request.id) }
  }
  window.addEventListener('message', receive)
  window.addEventListener('storage', checkStoredSession)
  return dispose
}
