import { useState } from 'react'
import type { StudioBootstrap } from '../../../shared/image-studio'
import type { ApiProfile } from '../types'
import {
  selectedPromptOptimizer,
  setStudioPromptOptimizationEnabled,
  studioPromptOptimizationEnabled,
  studioStorageNamespace,
} from '../lib/studioBridge'
import { studioText } from '../lib/studioLocale'

interface Props {
  config: StudioBootstrap
  profiles: ApiProfile[]
  activeProfileId: string
  onSelect: (id: string) => void
}

export default function StudioModelPanel({ config, profiles, activeProfileId, onSelect }: Props) {
  const [optimizationEnabled, setOptimizationEnabled] = useState(() => studioPromptOptimizationEnabled())
  const [optimizer, setOptimizer] = useState(() => {
    const stored = localStorage.getItem(`${studioStorageNamespace()}:prompt-optimizer`) || ''
    if (stored === 'off') return ''
    const choice = selectedPromptOptimizer()
    return choice ? `${choice.group_id}:${encodeURIComponent(choice.model)}` : stored
  })
  const imageGroup = config.groups.find((group) => profiles.some((profile) =>
    profile.id === activeProfileId && profile.id.startsWith(`studio:${group.id}:`)))
  const promptGroups = new Map<number, { name: string; models: string[] }>()
  for (const item of config.prompt_models ?? []) {
    const group = promptGroups.get(item.group_id) ?? { name: item.group_name, models: [] }
    group.models.push(item.model)
    promptGroups.set(item.group_id, group)
  }
  const validOptimizer = !optimizer || optimizer === 'off' || (config.prompt_models ?? []).some((item) => `${item.group_id}:${encodeURIComponent(item.model)}` === optimizer)
  const promptChannel = !validOptimizer ? 'unavailable' : optimizer.includes(':') ? optimizer.split(':')[0] : optimizer
  const promptGroup = promptGroups.get(Number(promptChannel))
  const chooseOptimizer = (value: string) => {
    if (value) localStorage.setItem(`${studioStorageNamespace()}:prompt-optimizer`, value)
    else localStorage.removeItem(`${studioStorageNamespace()}:prompt-optimizer`)
    setOptimizer(value)
  }
  const disabled = !config.enabled || !profiles.length
  const toggleOptimization = () => {
    const next = !optimizationEnabled
    setStudioPromptOptimizationEnabled(next)
    setOptimizationEnabled(next)
  }

  return <section className="studio-model-panel" aria-label={studioText('模型与渠道')}>
    <div className="studio-model-section">
      <div className="studio-section-heading">
        <h2>{studioText('生图模型')}</h2>
        <span className="studio-mode-badge">{studioText(config.async_enabled ? '支持后台生成' : '在线生成')}</span>
      </div>
      <div className="studio-model-controls">
        <label className="studio-model-field" htmlFor="studio-image-channel">
          <span className="studio-model-label">{studioText('生图渠道')}</span>
          <select id="studio-image-channel" value={imageGroup?.id ?? ''} disabled={disabled} className="studio-model-select"
            onChange={(event) => {
              const group = config.groups.find((item) => item.id === Number(event.target.value))
              const currentModel = profiles.find((item) => item.id === activeProfileId)?.model
              const choices = profiles.filter((item) => item.id.startsWith(`studio:${group?.id}:`))
              const next = choices.find((item) => item.model === currentModel) ?? choices[0]
              if (next) onSelect(next.id)
            }}>
            {!imageGroup && <option value="">{studioText('暂无可用渠道')}</option>}
            {config.groups.filter((group) => group.models.length).map((group) => <option key={group.id} value={group.id}>{group.name}</option>)}
          </select>
        </label>
        <label className="studio-model-field" htmlFor="studio-model">
          <span className="studio-model-label">{studioText('生图模型')}</span>
          <select id="studio-model" value={activeProfileId} disabled={disabled} className="studio-model-select" onChange={(event) => onSelect(event.target.value)}>
            {!profiles.length && <option value="">{studioText('暂无可用模型')}</option>}
            {profiles.filter((profile) => profile.id.startsWith(`studio:${imageGroup?.id}:`)).map((profile) => <option key={profile.id} value={profile.id}>{profile.model}</option>)}
          </select>
        </label>
      </div>
      <p className="studio-model-hint">{imageGroup
        ? `${imageGroup.name} · ${studioText(imageGroup.subscription_type === 'subscription' ? '订阅分组' : '余额计费')} · ${imageGroup.rate_multiplier}×`
        : studioText('当前账号没有可用的生图模型')}</p>
    </div>
    <div className="studio-model-section">
      <div className="studio-section-heading">
        <h2>{studioText('提示词优化')}</h2>
        <div className="studio-optimizer-toggle">
          <span id="studio-optimizer-toggle-label" className="studio-model-label">{studioText(optimizationEnabled ? '已开启' : '已关闭')}</span>
          <button
            type="button"
            role="switch"
            aria-checked={optimizationEnabled}
            aria-label={studioText('提示词优化开关')}
            disabled={disabled}
            className={`studio-switch${optimizationEnabled ? ' is-on' : ''}`}
            onClick={toggleOptimization}
          >
            <span className="studio-switch-track"><span className="studio-switch-thumb" /></span>
          </button>
        </div>
      </div>
      <div className="studio-model-controls">
        <label className="studio-model-field" htmlFor="studio-prompt-channel">
          <span className="studio-model-label">{studioText('优化渠道')}</span>
          <select id="studio-prompt-channel" value={promptChannel} disabled={disabled || !optimizationEnabled} className="studio-model-select" onChange={(event) => {
            const value = event.target.value
            const group = promptGroups.get(Number(value))
            const model = group?.models.find((item) => item.toLowerCase() === 'gpt-5.6-sol') ?? group?.models[0]
            chooseOptimizer(model ? `${value}:${encodeURIComponent(model)}` : value)
          }}>
            {!validOptimizer && <option value="unavailable" disabled>{studioText('原优化模型已不可用，请重新选择')}</option>}
            <option value="">{studioText('自动 · OpenAI 优先')}</option>
            {[...promptGroups].map(([id, group]) => <option key={id} value={id}>{group.name}</option>)}
          </select>
        </label>
        <label className="studio-model-field" htmlFor="studio-optimizer">
          <span className="studio-model-label">{studioText('优化模型')}</span>
          <select id="studio-optimizer" value={promptGroup ? optimizer : ''} disabled={disabled || !optimizationEnabled || !promptGroup} className="studio-model-select" onChange={(event) => chooseOptimizer(event.target.value)}>
            {!promptGroup && <option value="">{!validOptimizer ? studioText('暂无可用模型') : config.prompt_route?.default_model || studioText('按可用渠道自动选择')}</option>}
            {promptGroup?.models.map((model) => <option key={model} value={`${promptChannel}:${encodeURIComponent(model)}`}>{model}</option>)}
          </select>
        </label>
      </div>
      <p className="studio-model-hint">{studioText(!optimizationEnabled
        ? '按原文生成，不产生提示词优化调用。'
        : '优化按所选渠道计费；任务中可查看前后提示词及模型。')}</p>
    </div>
    {config.prompt_route && <p className="studio-route-note">{studioText('已创建生图智能路由，可在密钥页修改渠道')}</p>}
    {!config.enabled && <p role="status" className="studio-notice">{studioText('AI 绘图暂未开放，已有作品仍可查看和下载。')}</p>}
  </section>
}
