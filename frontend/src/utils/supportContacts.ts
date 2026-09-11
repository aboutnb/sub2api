/** Versioned contact_info payload. Plain text from older installations remains readable. */
export interface SupportContactEntry {
  name: string
  tag: string
  icon: typeof supportIcons[number]
  account: string
  url: string
}
export const supportIcons = ['chat', 'mail', 'users', 'globe', 'link'] as const
export function safeContactUrl(value: string): string {
  try {
    const url = new URL(value.trim())
    return ['https:', 'http:', 'mailto:', 'tel:'].includes(url.protocol) ? url.href : ''
  } catch { return '' }
}
export function parseSupportContacts(raw: string): SupportContactEntry[] {
  if (!raw.trim()) return []
  try {
    const data = JSON.parse(raw)
    if (data?.version === 1 && Array.isArray(data.contacts)) {
      return data.contacts.filter((entry: unknown) => entry && typeof entry === 'object').map((entry: Record<string, unknown>) => ({
        name: typeof entry.name === 'string' ? entry.name : '',
        tag: typeof entry.tag === 'string' ? entry.tag : '',
        icon: supportIcons.includes(entry.icon as typeof supportIcons[number]) ? entry.icon as typeof supportIcons[number] : 'chat',
        account: typeof entry.account === 'string' ? entry.account : '',
        url: typeof entry.url === 'string' ? entry.url : ''
      }))
    }
  } catch { /* Legacy plain text. */ }
  return [{ name: '', tag: '', icon: 'chat', account: raw, url: '' }]
}
export function serializeSupportContacts(contacts: SupportContactEntry[]): string {
  return contacts.length ? JSON.stringify({ version: 1, contacts }) : ''
}

export function validateSupportContacts(raw: string): 'nameRequired' | 'invalidUrl' | null {
  const entries = parseSupportContacts(raw)
  // Plain text has no name field and must not block unrelated settings saves.
  if (entries.some(entry => entry.url.trim() && !safeContactUrl(entry.url))) return 'invalidUrl'
  if (raw.trim().startsWith('{')) {
    try {
      const data = JSON.parse(raw)
      if (data?.version === 1 && Array.isArray(data.contacts) && entries.some(entry => !entry.name.trim())) return 'nameRequired'
    } catch { /* Legacy text. */ }
  }
  return null
}
