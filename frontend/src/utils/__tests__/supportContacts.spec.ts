import { describe, expect, it } from 'vitest'
import { parseSupportContacts, serializeSupportContacts, safeContactUrl, validateSupportContacts } from '../supportContacts'

describe('support contacts storage', () => {
  it('preserves legacy multiline text and supports clearing all contacts', () => {
    const legacy = '微信：hello\n工作日 9:00–18:00'
    expect(parseSupportContacts(legacy)[0].account).toBe(legacy)
    expect(validateSupportContacts(legacy)).toBeNull()
    expect(parseSupportContacts('')).toEqual([])
    expect(serializeSupportContacts([])).toBe('')
  })
  it('round trips names, custom tags, icons and optional links in order', () => {
    const entries = [
      { name: '客服', tag: '工作日在线', icon: 'chat' as const, account: 'hello', url: '' },
      { name: '邮件支持', tag: '技术支持', icon: 'mail' as const, account: 'support@example.com', url: 'mailto:support@example.com' }
    ]
    const raw = serializeSupportContacts(entries)
    expect(parseSupportContacts(raw)).toEqual(entries)
    expect(validateSupportContacts(raw)).toBeNull()
    expect(validateSupportContacts(serializeSupportContacts([{ ...entries[0], name: ' ' }]))).toBe('nameRequired')
  })
  it('rejects executable links and tolerates malformed stored fields', () => {
    for (const url of ['javascript:alert(1)', 'data:text/html,test', '//example.com', 'invalid']) expect(safeContactUrl(url)).toBe('')
    expect(safeContactUrl('https://example.com')).toBe('https://example.com/')
    expect(safeContactUrl('tel:+1234')).toBe('tel:+1234')
    const raw = JSON.stringify({ version: 1, contacts: [null, { name: 'test', icon: '<svg>', account: 123, url: 'javascript:alert(1)' }] })
    expect(parseSupportContacts(raw)).toEqual([{ name: 'test', tag: '', icon: 'chat', account: '', url: 'javascript:alert(1)' }])
    expect(validateSupportContacts(raw)).toBe('invalidUrl')
    expect(parseSupportContacts('{malformed')[0].account).toBe('{malformed')
  })
})
