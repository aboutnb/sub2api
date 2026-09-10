import type { ApiKey } from '@/types'
import type {
  SmartRouteFormState,
  SmartRouteInput,
  SmartRouteStrategy,
  SmartRouteWeights
} from '@/types/smart-routing'

const PRESET_WEIGHTS: Record<Exclude<SmartRouteStrategy, 'custom'>, SmartRouteWeights> = {
  auto: { price: 40, speed: 30, success: 30 },
  price: { price: 100, speed: 0, success: 0 },
  speed: { price: 15, speed: 70, success: 15 },
  success: { price: 15, speed: 15, success: 70 }
}

export function smartRoutePresetWeights(strategy: SmartRouteStrategy): SmartRouteWeights {
  const value = strategy === 'custom' ? PRESET_WEIGHTS.auto : PRESET_WEIGHTS[strategy]
  return { ...value }
}

export function createSmartRouteForm(groupId: number | null = null): SmartRouteFormState {
  return {
    mode: 'single',
    group_id: groupId,
    candidate_group_ids: [],
    strategy: 'auto',
    weights: smartRoutePresetWeights('auto'),
    rate_guard: {
      enabled: true,
      max_rate_multiplier: 1,
      max_image_rate_multiplier: null
    }
  }
}

export function smartRouteFormFromKey(key: ApiKey): SmartRouteFormState {
  if (!key.routing) return createSmartRouteForm(key.group_id)
  return {
    mode: 'smart',
    group_id: null,
    candidate_group_ids: [...key.routing.candidate_group_ids],
    strategy: key.routing.strategy,
    weights: { ...key.routing.weights },
    rate_guard: { ...key.routing.rate_guard }
  }
}

export function smartRouteFormError(form: SmartRouteFormState): 'group' | 'candidates' | 'weights' | 'rate' | null {
  if (form.mode === 'single') return form.group_id === null ? 'group' : null
  if (form.candidate_group_ids.length < 1 || form.candidate_group_ids.length > 20) return 'candidates'
  const { price, speed, success } = form.weights
  if (
    form.strategy === 'custom' &&
    (![price, speed, success].every((value) => Number.isInteger(value) && value >= 0 && value <= 100) || price + speed + success !== 100)
  ) return 'weights'
  if (
    form.rate_guard.enabled &&
    [form.rate_guard.max_rate_multiplier, form.rate_guard.max_image_rate_multiplier]
      .some((value) => value !== null && (!Number.isFinite(value) || value < 0))
  ) return 'rate'
  return null
}

export function smartRoutePayload(form: SmartRouteFormState): SmartRouteInput {
  if (form.mode === 'single') {
    return {
      mode: 'single',
      group_id: form.group_id ?? undefined,
      rate_guard: { ...form.rate_guard }
    }
  }
  return {
    mode: 'smart',
    candidate_group_ids: [...form.candidate_group_ids],
    strategy: form.strategy,
    weights: form.strategy === 'custom' ? { ...form.weights } : smartRoutePresetWeights(form.strategy),
    rate_guard: { ...form.rate_guard }
  }
}
