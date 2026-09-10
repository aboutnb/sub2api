export type SmartRouteMode = 'single' | 'smart'

export type SmartRouteStrategy = 'auto' | 'price' | 'speed' | 'success' | 'custom'

export interface SmartRouteWeights {
  price: number
  speed: number
  success: number
}

export interface SmartRouteRateGuard {
  enabled: boolean
  max_rate_multiplier: number | null
  max_image_rate_multiplier: number | null
}

export interface SmartRouteInput {
  mode: SmartRouteMode
  group_id?: number
  candidate_group_ids?: number[]
  strategy?: SmartRouteStrategy
  weights?: SmartRouteWeights
  rate_guard: SmartRouteRateGuard
}

export interface SmartRouteConfig {
  mode: 'smart'
  platform: string
  subscription_type: string
  candidate_group_ids: number[]
  strategy: SmartRouteStrategy
  weights: SmartRouteWeights
  rate_guard: SmartRouteRateGuard
  updated_at?: string
}

export interface SmartRouteFormState {
  mode: SmartRouteMode
  group_id: number | null
  candidate_group_ids: number[]
  strategy: SmartRouteStrategy
  weights: SmartRouteWeights
  rate_guard: SmartRouteRateGuard
}
