// Browser properties are auxiliary abuse signals, never proof of identity.
let pendingDigest: Promise<string> | undefined
export function registrationFeatureDigest(): Promise<string> {
  pendingDigest ??= collectDigest()
  return pendingDigest
}

async function collectDigest(): Promise<string> {
  try {
    const data = JSON.stringify({
      v: 1,
      platform: navigator.platform,
      languages: Array.from(navigator.languages || []).slice(0, 4),
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      screen: [Math.min(screen.width, screen.height), Math.max(screen.width, screen.height), screen.colorDepth],
      touch: navigator.maxTouchPoints,
      cores: navigator.hardwareConcurrency
    })
    const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(data))
    return Array.from(new Uint8Array(digest), byte => byte.toString(16).padStart(2, '0')).join('')
  } catch {
    // Privacy restrictions and unsupported browsers retain existing verification.
    return ''
  }
}
