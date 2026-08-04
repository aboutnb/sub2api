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

export function withPublicAccessHeader(headers: Record<string, string> = {}): Record<string, string> {
  const publishKey = getPublicAccessPublishKey()
  if (publishKey) {
    headers[getPublicAccessHeaderName()] = publishKey
  }
  return headers
}
