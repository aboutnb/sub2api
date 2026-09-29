export const STUDIO_CHANNEL = 'sub2api.image-studio.v1'

export interface StudioBootstrap {
  enabled: boolean
  storage_namespace: string
  async_enabled: boolean
  max_body_bytes: number
  groups: Array<{
    id: number
    name: string
    subscription_type: string
    rate_multiplier: number
    models: Array<{ id: string; platform: string; capabilities?: StudioCapabilities }>
  }>
  prompt_models?: StudioPromptModel[]
  prompt_route?: {
    key_id: number
    name: string
    channels: number
    default_model: string
    default_group: string
  }
}

export interface StudioPromptModel {
  group_id: number
  group_name: string
  platform: string
  model: string
}

export interface StudioCapabilities {
  edits: boolean
  mask: boolean
  multiple: boolean
  max_input_images: number
  max_upload_bytes: number
  input_types: string[]
}

export type StudioRequest = {
  channel: typeof STUDIO_CHANNEL
  id: string
  operation: 'bootstrap' | 'request' | 'cancel'
  path?: string
  method?: 'GET' | 'POST'
  body?: string | Array<[string, string | Blob]>
}

export function isStudioRequest(value: unknown): value is StudioRequest {
  if (!value || typeof value !== 'object') return false
  const item = value as StudioRequest
  if (item.channel !== STUDIO_CHANNEL || typeof item.id !== 'string' || !/^[a-zA-Z0-9-]{1,80}$/.test(item.id)) return false
  if (item.operation === 'bootstrap' || item.operation === 'cancel') return true
  if (item.operation !== 'request' || typeof item.path !== 'string') return false
  const group = item.path.match(/^\/image-studio\/groups\/([1-9]\d*)\/(.*)$/)
  if (!group) return false
  const suffix = group[2]
  if (item.method === 'POST' && suffix === 'prompt') return typeof item.body === 'string' && item.body.length <= 20000
  if (!suffix.startsWith('images/')) return false
  const imageSuffix = suffix.slice('images/'.length)
  if (item.method === 'GET') return /^tasks\/imgtask_[a-zA-Z0-9_-]+$/.test(imageSuffix) && item.body === undefined
  if (item.method !== 'POST' || !/^(generations|edits)(\/async)?$/.test(imageSuffix)) return false
  return typeof item.body === 'string' || (Array.isArray(item.body) && item.body.every((entry) =>
    Array.isArray(entry) && entry.length === 2 && typeof entry[0] === 'string' &&
    (typeof entry[1] === 'string' || entry[1] instanceof Blob)))
}
