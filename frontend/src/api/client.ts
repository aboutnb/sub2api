/**
 * Axios HTTP Client Configuration
 * Base client with interceptors for authentication, token refresh, and error handling
 */

import axios, { AxiosInstance, AxiosError, InternalAxiosRequestConfig, AxiosResponse } from 'axios'
import type { ApiResponse } from '@/types'
import { getLocale } from '@/i18n'
import { getAPIBaseURL } from './url'
export { buildApiUrl, buildGatewayUrl } from './url'

// ==================== Axios Instance Configuration ====================

const DEFAULT_PUBLIC_ACCESS_HEADER = 'x-sub2api-publish-key'

function getPublicAccessPublishKey(): string {
  if (typeof window === 'undefined') {
    return ''
  }
  const config = window.__APP_CONFIG__
  if (config?.public_access_guard_enabled !== true) {
    return ''
  }
  return (config.public_access_publish_key || '').trim()
}

function getPublicAccessHeaderName(): string {
  if (typeof window === 'undefined') {
    return DEFAULT_PUBLIC_ACCESS_HEADER
  }
  return (window.__APP_CONFIG__?.public_access_header_name || DEFAULT_PUBLIC_ACCESS_HEADER).trim() || DEFAULT_PUBLIC_ACCESS_HEADER
}

function withPublicAccessHeader(headers: Record<string, string> = {}): Record<string, string> {
  const publishKey = getPublicAccessPublishKey()
  if (publishKey) {
    headers[getPublicAccessHeaderName()] = publishKey
  }
  return headers
}

export const apiClient: AxiosInstance = axios.create({
  baseURL: getAPIBaseURL(),
  withCredentials: true,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// ==================== Registration Challenge ====================

type RegistrationChallengeAction =
  | 'register'
  | 'send_verify_code'
  | 'oauth_pending_send_verify_code'
  | 'oauth_pending_create_account'

interface RegistrationChallengeResponse {
  token: string
  issued_at: number
  expires_at: number
  min_elapsed_ms: number
  trap_field: string
  salt: string
}

interface RegistrationChallengeSubmission {
  token: string
  completed_at: number
  proof: string
  trap_field: string
  trap_value: string
}

const REGISTRATION_CHALLENGE_ENDPOINTS: Array<{
  suffix: string
  action: RegistrationChallengeAction
}> = [
  { suffix: '/auth/register', action: 'register' },
  { suffix: '/auth/send-verify-code', action: 'send_verify_code' },
  {
    suffix: '/auth/oauth/pending/send-verify-code',
    action: 'oauth_pending_send_verify_code'
  },
  {
    suffix: '/auth/oauth/pending/create-account',
    action: 'oauth_pending_create_account'
  }
]

let cachedRegistrationChallenge: RegistrationChallengeResponse | null = null
let pendingRegistrationChallenge: Promise<RegistrationChallengeResponse> | null = null

function getRegistrationChallengeAction(config: InternalAxiosRequestConfig): RegistrationChallengeAction | null {
  if (String(config.method || '').toLowerCase() !== 'post') return null
  const rawURL = String(config.url || '')
  if (!rawURL) return null
  let pathname = rawURL
  try {
    pathname = new URL(rawURL, 'https://sub2api.local').pathname
  } catch {
    pathname = rawURL.split('?')[0] || rawURL
  }
  const match = REGISTRATION_CHALLENGE_ENDPOINTS.find(endpoint => pathname.endsWith(endpoint.suffix))
  return match?.action || null
}

function isRegistrationChallengeFresh(challenge: RegistrationChallengeResponse | null): challenge is RegistrationChallengeResponse {
  return !!challenge && Date.now() < challenge.expires_at - 30_000
}

async function fetchRegistrationChallenge(): Promise<RegistrationChallengeResponse> {
  if (isRegistrationChallengeFresh(cachedRegistrationChallenge)) {
    return cachedRegistrationChallenge
  }
  if (pendingRegistrationChallenge) {
    return pendingRegistrationChallenge
  }

  pendingRegistrationChallenge = axios
    .get<ApiResponse<RegistrationChallengeResponse>>('/auth/registration-challenge', {
      baseURL: getAPIBaseURL(),
      withCredentials: true,
      headers: withPublicAccessHeader({
        'Accept-Language': getLocale()
      })
    })
    .then(response => {
      const payload = response.data
      if (!payload || payload.code !== 0 || !payload.data?.token) {
        throw new Error(payload?.message || 'Failed to initialize registration challenge')
      }
      cachedRegistrationChallenge = payload.data
      return payload.data
    })
    .finally(() => {
      pendingRegistrationChallenge = null
    })

  return pendingRegistrationChallenge
}

function getRegistrationTrapValue(): string {
  if (typeof document === 'undefined') return ''
  const inputs = Array.from(document.querySelectorAll<HTMLInputElement>('[data-registration-trap]'))
  for (const input of inputs) {
    const value = input.value || ''
    if (value.trim()) return value
  }
  return ''
}

function normalizeChallengeEmail(email: unknown): string {
  return typeof email === 'string' ? email.trim().toLowerCase() : ''
}

function registrationChallengeProofSource(
  token: string,
  email: string,
  action: RegistrationChallengeAction,
  completedAt: number,
  trapField: string,
  salt: string
): string {
  return [token, email, action, String(completedAt), trapField, salt].join('\n')
}

async function sha256Hex(value: string): Promise<string> {
  if (typeof crypto === 'undefined' || !crypto.subtle) {
    throw new Error('Web Crypto is unavailable')
  }
  const encoded = new TextEncoder().encode(value)
  const digest = await crypto.subtle.digest('SHA-256', encoded)
  return Array.from(new Uint8Array(digest))
    .map(byte => byte.toString(16).padStart(2, '0'))
    .join('')
}

function fnv1a64Hex(value: string): string {
  let hash = 0xcbf29ce484222325n
  const prime = 0x100000001b3n
  const mask = 0xffffffffffffffffn
  for (let i = 0; i < value.length; i += 1) {
    hash ^= BigInt(value.charCodeAt(i))
    hash = (hash * prime) & mask
  }
  return `fnv1a:${hash.toString(16).padStart(16, '0')}`
}

async function buildRegistrationChallengeSubmission(
  challenge: RegistrationChallengeResponse,
  action: RegistrationChallengeAction,
  email: string
): Promise<RegistrationChallengeSubmission> {
  const waitMs = challenge.min_elapsed_ms - (Date.now() - challenge.issued_at)
  if (waitMs > 0 && waitMs < 5_000) {
    await new Promise(resolve => setTimeout(resolve, waitMs))
  }

  const completedAt = Date.now()
  const proofSource = registrationChallengeProofSource(
    challenge.token,
    email,
    action,
    completedAt,
    challenge.trap_field,
    challenge.salt
  )

  let proof: string
  try {
    proof = await sha256Hex(proofSource)
  } catch {
    proof = fnv1a64Hex(proofSource)
  }

  return {
    token: challenge.token,
    completed_at: completedAt,
    proof,
    trap_field: challenge.trap_field,
    trap_value: getRegistrationTrapValue()
  }
}

function parseRequestBody(data: unknown): { body: Record<string, unknown>; stringify: boolean } | null {
  if (!data) return { body: {}, stringify: false }
  if (typeof data === 'string') {
    try {
      const parsed = JSON.parse(data)
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        return { body: parsed as Record<string, unknown>, stringify: true }
      }
    } catch {
      return null
    }
    return null
  }
  if (typeof data === 'object' && !Array.isArray(data)) {
    return { body: data as Record<string, unknown>, stringify: false }
  }
  return null
}

async function attachRegistrationChallenge(config: InternalAxiosRequestConfig): Promise<void> {
  const action = getRegistrationChallengeAction(config)
  if (!action) return

  const parsed = parseRequestBody(config.data)
  if (!parsed) return
  if (parsed.body.registration_challenge) return

  const challenge = await fetchRegistrationChallenge()
  cachedRegistrationChallenge = null
  parsed.body.registration_challenge = await buildRegistrationChallengeSubmission(
    challenge,
    action,
    normalizeChallengeEmail(parsed.body.email)
  )
  config.data = parsed.stringify ? JSON.stringify(parsed.body) : parsed.body
}

// ==================== Token Refresh State ====================

// Track if a token refresh is in progress to prevent multiple simultaneous refresh requests
let isRefreshing = false
// Queue of requests waiting for token refresh
let refreshSubscribers: Array<(token: string) => void> = []

/**
 * Subscribe to token refresh completion
 */
function subscribeTokenRefresh(callback: (token: string) => void): void {
  refreshSubscribers.push(callback)
}

/**
 * Notify all subscribers that token has been refreshed
 */
function onTokenRefreshed(token: string): void {
  refreshSubscribers.forEach((callback) => callback(token))
  refreshSubscribers = []
}

// ==================== Request Interceptor ====================

// Get user's timezone
const getUserTimezone = (): string => {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone
  } catch {
    return 'UTC'
  }
}

apiClient.interceptors.request.use(
  async (config: InternalAxiosRequestConfig) => {
    // Attach token from localStorage
    const token = localStorage.getItem('auth_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // Attach locale for backend translations
    if (config.headers) {
      config.headers['Accept-Language'] = getLocale()
      const publishKey = getPublicAccessPublishKey()
      if (publishKey) {
        config.headers[getPublicAccessHeaderName()] = publishKey
      }
    }

    // Attach timezone for all GET requests (backend may use it for default date ranges)
    if (config.method === 'get') {
      if (!config.params) {
        config.params = {}
      }
      config.params.timezone = getUserTimezone()
    }

    await attachRegistrationChallenge(config)

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// ==================== Response Interceptor ====================

apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    // Unwrap standard API response format { code, message, data }
    const apiResponse = response.data as ApiResponse<unknown>
    if (apiResponse && typeof apiResponse === 'object' && 'code' in apiResponse) {
      if (apiResponse.code === 0) {
        // Success - return the data portion
        response.data = apiResponse.data
      } else {
        // API error
        const resp = apiResponse as unknown as Record<string, unknown>
        return Promise.reject({
          status: response.status,
          code: apiResponse.code,
          message: apiResponse.message || 'Unknown error',
          reason: resp.reason,
          metadata: resp.metadata,
        })
      }
    }
    return response
  },
  async (error: AxiosError<ApiResponse<unknown>>) => {
    // Request cancellation: keep the original axios cancellation error so callers can ignore it.
    // Otherwise we'd misclassify it as a generic "network error".
    if (error.code === 'ERR_CANCELED' || axios.isCancel(error)) {
      return Promise.reject(error)
    }

    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    // Handle common errors
    if (error.response) {
      const { status, data } = error.response
      const url = String(error.config?.url || '')

      // Validate `data` shape to avoid HTML error pages breaking our error handling.
      const apiData = (typeof data === 'object' && data !== null ? data : {}) as Record<string, any>

      // Ops monitoring disabled: treat as feature-flagged 404, and proactively redirect away
      // from ops pages to avoid broken UI states.
      if (status === 404 && apiData.message === 'Ops monitoring is disabled') {
        try {
          localStorage.setItem('ops_monitoring_enabled_cached', 'false')
        } catch {
          // ignore localStorage failures
        }
        try {
          window.dispatchEvent(new CustomEvent('ops-monitoring-disabled'))
        } catch {
          // ignore event failures
        }

        if (window.location.pathname.startsWith('/admin/ops')) {
          window.location.href = '/admin/settings'
        }

        return Promise.reject({
          status,
          code: 'OPS_DISABLED',
          message: apiData.message || error.message,
          url
        })
      }

      if (status === 423 && apiData.code === 'ADMIN_COMPLIANCE_ACK_REQUIRED') {
        try {
          window.dispatchEvent(new CustomEvent('admin-compliance-required', {
            detail: apiData.metadata || {}
          }))
        } catch {
          // ignore event failures
        }

        return Promise.reject({
          status,
          code: apiData.code,
          message: apiData.message || error.message,
          metadata: apiData.metadata,
        })
      }

      // 401: Try to refresh the token if we have a refresh token
      // This handles TOKEN_EXPIRED, INVALID_TOKEN, TOKEN_REVOKED, etc.
      if (status === 401 && !originalRequest._retry) {
        const refreshToken = localStorage.getItem('refresh_token')
        const isAuthEndpoint =
          url.includes('/auth/login') || url.includes('/auth/register') || url.includes('/auth/refresh')

        // If we have a refresh token and this is not an auth endpoint, try to refresh
        if (refreshToken && !isAuthEndpoint) {
          if (isRefreshing) {
            // Wait for the ongoing refresh to complete
            return new Promise((resolve, reject) => {
              subscribeTokenRefresh((newToken: string) => {
                if (newToken) {
                  // Mark as retried to prevent infinite loop if retry also returns 401
                  originalRequest._retry = true
                  if (originalRequest.headers) {
                    originalRequest.headers.Authorization = `Bearer ${newToken}`
                  }
                  resolve(apiClient(originalRequest))
                } else {
                  // Refresh failed, reject with original error
                  reject({
                    status,
                    code: apiData.code,
                    message: apiData.message || apiData.detail || error.message
                  })
                }
              })
            })
          }

          originalRequest._retry = true
          isRefreshing = true

          try {
            // Call refresh endpoint directly to avoid circular dependency
            const refreshResponse = await axios.post(
              `${getAPIBaseURL()}/auth/refresh`,
              { refresh_token: refreshToken },
              { headers: withPublicAccessHeader({ 'Content-Type': 'application/json' }) }
            )

            const refreshData = refreshResponse.data as ApiResponse<{
              access_token: string
              refresh_token: string
              expires_in: number
            }>

            if (refreshData.code === 0 && refreshData.data) {
              const { access_token, refresh_token: newRefreshToken, expires_in } = refreshData.data

              // Update tokens in localStorage (convert expires_in to timestamp)
              localStorage.setItem('auth_token', access_token)
              localStorage.setItem('refresh_token', newRefreshToken)
              localStorage.setItem('token_expires_at', String(Date.now() + expires_in * 1000))

              // Notify subscribers with new token
              onTokenRefreshed(access_token)

              // Retry the original request with new token
              if (originalRequest.headers) {
                originalRequest.headers.Authorization = `Bearer ${access_token}`
              }

              isRefreshing = false
              return apiClient(originalRequest)
            }

            // Refresh response was not successful, fall through to clear auth
            throw new Error('Token refresh failed')
          } catch (refreshError) {
            // Refresh failed - notify subscribers with empty token
            onTokenRefreshed('')
            isRefreshing = false

            // Clear tokens and redirect to login
            localStorage.removeItem('auth_token')
            localStorage.removeItem('refresh_token')
            localStorage.removeItem('auth_user')
            localStorage.removeItem('token_expires_at')
            sessionStorage.setItem('auth_expired', '1')

            if (!window.location.pathname.includes('/login')) {
              window.location.href = '/login'
            }

            return Promise.reject({
              status: 401,
              code: 'TOKEN_REFRESH_FAILED',
              message: 'Session expired. Please log in again.'
            })
          }
        }

        // No refresh token or is auth endpoint - clear auth and redirect
        const hasToken = !!localStorage.getItem('auth_token')
        const headers = error.config?.headers as Record<string, unknown> | undefined
        const authHeader = headers?.Authorization ?? headers?.authorization
        const sentAuth =
          typeof authHeader === 'string'
            ? authHeader.trim() !== ''
            : Array.isArray(authHeader)
              ? authHeader.length > 0
              : !!authHeader

        localStorage.removeItem('auth_token')
        localStorage.removeItem('refresh_token')
        localStorage.removeItem('auth_user')
        localStorage.removeItem('token_expires_at')
        if ((hasToken || sentAuth) && !isAuthEndpoint) {
          sessionStorage.setItem('auth_expired', '1')
        }
        // Only redirect if not already on login page
        if (!window.location.pathname.includes('/login')) {
          window.location.href = '/login'
        }
      }

      // Return structured error
      return Promise.reject({
        status,
        code: apiData.code,
        reason: apiData.reason,
        error: apiData.error,
        message: apiData.message || apiData.detail || error.message,
        metadata: apiData.metadata,
      })
    }

    // Network error
    return Promise.reject({
      status: 0,
      message: 'Network error. Please check your connection.'
    })
  }
)

export default apiClient
