import { describe, expect, it } from 'vitest'
import {
  createSmartRouteForm,
  smartRouteFormError,
  smartRoutePayload,
  smartRoutePresetWeights
} from '@/composables/useSmartRouteForm'

describe('useSmartRouteForm', () => {
  it('uses the documented preset weights', () => {
    expect(smartRoutePresetWeights('auto')).toEqual({ price: 40, speed: 30, success: 30 })
    expect(smartRoutePresetWeights('price')).toEqual({ price: 100, speed: 0, success: 0 })
    expect(smartRoutePresetWeights('speed')).toEqual({ price: 15, speed: 70, success: 15 })
    expect(smartRoutePresetWeights('success')).toEqual({ price: 15, speed: 15, success: 70 })
  })

  it('creates smart payloads without legacy group_id', () => {
    const form = createSmartRouteForm()
    form.mode = 'smart'
    form.candidate_group_ids = [1, 2]
    const payload = smartRoutePayload(form)
    expect(payload.mode).toBe('smart')
    expect(payload.candidate_group_ids).toEqual([1, 2])
    expect(payload.group_id).toBeUndefined()
  })

  it('validates candidates, custom weights, and rate limits', () => {
    const form = createSmartRouteForm()
    form.mode = 'smart'
    expect(smartRouteFormError(form)).toBe('candidates')
    form.candidate_group_ids = [1]
    form.strategy = 'custom'
    form.weights = { price: 40, speed: 30, success: 20 }
    expect(smartRouteFormError(form)).toBe('weights')
    form.weights.success = 30
    form.rate_guard.max_rate_multiplier = -1
    expect(smartRouteFormError(form)).toBe('rate')
  })
})
