import type { StudioBootstrap } from '../../../shared/image-studio'
import type { ApiProfile } from '../types'
import { createDefaultOpenAIProfile } from './apiProfiles'

export function studioProfiles(config: StudioBootstrap, origin: string, saved: ApiProfile[] = [], activeID = '') {
  const available: ApiProfile[] = config.groups.flatMap((group) => group.models.map((model) => ({
    ...createDefaultOpenAIProfile(),
    id: `studio:${group.id}:${encodeURIComponent(model.id)}:${config.async_enabled ? 'async' : 'sync'}`,
    name: `${group.name} / ${model.id}`,
    provider: config.async_enabled ? 'sb2api-async' : 'openai',
    baseUrl: `${origin}/image-studio/groups/${group.id}/`,
    apiKey: '', model: model.id, apiMode: 'images',
    timeout: 1800, apiProxy: false, streamImages: false,
  })))
  // 旧配置只用于轮询原任务，不能重新进入可选模型列表。
  const historical: ApiProfile[] = saved.flatMap((profile) => {
    const match = /^studio:([1-9]\d*):([^:]+):(async|sync)$/.exec(profile.id)
    if (!match || available.some((item) => item.id === profile.id)) return []
    try {
      return [{ ...createDefaultOpenAIProfile(), id: profile.id, name: profile.name,
        provider: match[3] === 'async' ? 'sb2api-async' : 'openai', apiKey: '',
        baseUrl: `${origin}/image-studio/groups/${match[1]}/`, model: decodeURIComponent(match[2]),
        apiMode: 'images', timeout: 1800, apiProxy: false, streamImages: false }]
    } catch {
      return []
    }
  })
  // 同步/异步开关变化不应改变用户选择的渠道和模型。
  const identity = activeID.replace(/:(async|sync)$/, '')
  const active = available.find((item) => item.id.replace(/:(async|sync)$/, '') === identity) ?? available[0]
  return { available, profiles: [...available, ...historical], activeProfileId: active?.id ?? '' }
}
