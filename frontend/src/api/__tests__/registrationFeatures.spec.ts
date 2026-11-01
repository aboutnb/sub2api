import { beforeEach, describe, expect, it, vi } from 'vitest'

describe('registrationFeatureDigest', () => {
  beforeEach(() => { vi.resetModules(); vi.unstubAllGlobals() })
  it('hashes features once without storing raw device properties', async () => {
    const digest = vi.fn().mockResolvedValue(new Uint8Array(32).fill(7).buffer)
    vi.stubGlobal('crypto', { subtle: { digest } })
    const { registrationFeatureDigest } = await import('../registrationFeatures')
    expect(await registrationFeatureDigest()).toBe('07'.repeat(32))
    expect(await registrationFeatureDigest()).toBe('07'.repeat(32))
    expect(digest).toHaveBeenCalledTimes(1)
  })
  it('keeps registration available when browser features are inaccessible', async () => {
    vi.stubGlobal('crypto', undefined)
    const { registrationFeatureDigest } = await import('../registrationFeatures')
    expect(await registrationFeatureDigest()).toBe('')
  })
})
