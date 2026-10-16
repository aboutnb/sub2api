import { describe, expect, it } from 'vitest'
import type { StudioBootstrap } from '../../../shared/image-studio'
import { createDefaultOpenAIProfile } from './apiProfiles'
import { studioProfiles } from './studioProfiles'

const config: StudioBootstrap = {
  enabled: true, async_enabled: false, storage_namespace: 'image-studio-user-1', max_body_bytes: 4096,
  groups: [1, 2].map((id) => ({ id, name: `Channel ${id}`, rate_multiplier: id, subscription_type: 'standard', models: [{ id: 'drawing', platform: 'openai' }] })),
}

describe('studio profile recovery', () => {
  it('keeps channel and model when async availability changes and retains the old poll route', () => {
    const previous = studioProfiles({ ...config, async_enabled: true }, 'https://old.test')
    const result = studioProfiles(config, 'https://site.test', previous.profiles, 'studio:2:drawing:async')
    expect(result.activeProfileId).toBe('studio:2:drawing:sync')
    const historical = result.profiles.find((item) => item.id === 'studio:2:drawing:async')!
    expect(historical.provider).toBe('sb2api-async')
    expect(historical.baseUrl).toBe('https://site.test/image-studio/groups/2/')
    expect(historical.timeout).toBe(1800)
    expect(result.available).toHaveLength(2)
  })
  it('ignores corrupted saved IDs and rebuilds historical profiles without credentials', () => {
    const stored = [
      { ...createDefaultOpenAIProfile(), id: 'studio:2:%broken:async' },
      { ...createDefaultOpenAIProfile(), id: 'studio:7:archived:async', apiKey: 'secret', baseUrl: 'https://untrusted.test' },
    ]
    const result = studioProfiles(config, 'https://site.test', stored)
    expect(result.profiles).toHaveLength(3)
    const historical = result.profiles[result.profiles.length - 1]
    expect(historical.apiKey).toBe('')
    expect(historical.baseUrl).toBe('https://site.test/image-studio/groups/7/')
    expect(result.available.some((item) => item.model === 'archived')).toBe(false)
  })
})
