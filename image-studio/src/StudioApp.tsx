import { studioText } from './lib/studioLocale'
import { Fragment, useEffect, useMemo, useState } from 'react'
import { useStudioLocale } from './lib/studioLocale'
import { initStore, useStore } from './store'
import { createDefaultOpenAIProfile } from './lib/apiProfiles'
import { studioConfig, studioStorageNamespace } from './lib/studioBridge'
import SearchBar from './components/SearchBar'
import TaskGrid from './components/TaskGrid'
import InputBar from './components/InputBar'
import DetailModal from './components/DetailModal'
import Lightbox from './components/Lightbox'
import ConfirmDialog from './components/ConfirmDialog'
import Toast from './components/Toast'
import MaskEditorModal from './components/MaskEditorModal'
import ImageContextMenu from './components/ImageContextMenu'
import { FavoriteCollectionPickerModal, FavoriteCollectionsView, ManageCollectionsModal } from './components/FavoriteCollections'
import { useGlobalClickSuppression } from './lib/clickSuppression'

const config = studioConfig()!
const profiles = config.groups.flatMap((group) => group.models.map((model) => ({
  ...createDefaultOpenAIProfile(),
  id: `studio:${group.id}:${encodeURIComponent(model.id)}:${config.async_enabled ? 'async' : 'sync'}`,
  name: `${group.name} / ${model.id}`,
  provider: config.async_enabled ? 'sb2api-async' : 'openai',
  baseUrl: `${window.location.origin}/image-studio/groups/${group.id}/`,
  apiKey: '', model: model.id, apiMode: 'images' as const,
  timeout: 1800, apiProxy: false, streamImages: false,
})))

let initialization: Promise<void> | undefined
function initialize() {
  if (initialization) return initialization
  const state = useStore.getState()
  // 保留已移除模型的无凭证配置用于历史任务查询，但不放入模型选择列表。
  const historical = state.settings.profiles.flatMap((profile) => {
    const match = /^studio:([1-9]\d*):([^:]+):(async|sync)$/.exec(profile.id)
    if (!match || profiles.some((item) => item.id === profile.id)) return []
    return [{ ...createDefaultOpenAIProfile(), id: profile.id, name: profile.name,
      provider: match[3] === 'async' ? 'sb2api-async' : 'openai', apiKey: '',
      baseUrl: `${window.location.origin}/image-studio/groups/${match[1]}/`, model: decodeURIComponent(match[2]),
      apiMode: 'images' as const, apiProxy: false, streamImages: false }]
  })
  state.setSettings({ profiles: [...profiles, ...historical], customProviders: [],
    activeProfileId: profiles.some((item) => item.id === state.settings.activeProfileId) ? state.settings.activeProfileId : profiles[0]?.id,
    agentApiConfigMode: 'off', reuseTaskApiProfileTemporarily: false })
  useStore.setState({ appMode: 'gallery', showSettings: false })
  initialization = initStore()
  return initialization
}

export default function StudioApp() {
  const locale = useStudioLocale()
  const settings = useStore((state) => state.settings)
  const filterFavorite = useStore((state) => state.filterFavorite)
  const activeCollection = useStore((state) => state.activeFavoriteCollectionId)
  const [ready, setReady] = useState(false)
  const [error, setError] = useState('')
  const [optimizer, setOptimizer] = useState('')
  useGlobalClickSuppression()
  useEffect(() => { initialize().then(() => setReady(true)).catch((error: Error) => setError(error.message)) }, [])
  useEffect(() => {
    if (!ready) return
    setOptimizer(localStorage.getItem(`${studioStorageNamespace()}:prompt-optimizer`) || '')
  }, [ready])
  const promptGroups = useMemo(() => {
    const models = config.prompt_models ?? []
    const groups = new Map<number, { name: string; models: typeof models }>()
    for (const item of models) {
      const group = groups.get(item.group_id) ?? { name: item.group_name, models: [] }
      group.models.push(item)
      groups.set(item.group_id, group)
    }
    return [...groups.entries()]
  }, [])
  const chooseOptimizer = (value: string) => {
    setOptimizer(value)
    const key = `${studioStorageNamespace()}:prompt-optimizer`
    if (value) localStorage.setItem(key, value)
    else localStorage.removeItem(key)
  }
  useEffect(() => {
    if (!optimizer) return
    const available = promptGroups.some(([, group]) => group.models.some((item) => `${item.group_id}:${encodeURIComponent(item.model)}` === optimizer))
    if (!available) chooseOptimizer('')
  }, [optimizer, promptGroups])
  if (error) return <p role="alert" className="p-6 text-red-600">{error}</p>
  if (!ready) return <p role="status" className="p-6">{studioText("加载中...")}</p>
  return <Fragment key={locale}>
    <header className="studio-toolbar safe-area-x">
      <label className="studio-model-field">
        <span className="studio-model-label" id="studio-model-label">{studioText('生图模型')}</span>
        <select id="studio-model" aria-labelledby="studio-model-label" value={settings.activeProfileId} disabled={!config.enabled || !profiles.length}
          className="studio-model-select"
          onChange={(event) => useStore.getState().setSettings({ activeProfileId: event.target.value })}>
          {config.groups.map((group) => <optgroup key={group.id} label={`${group.name} · ${group.rate_multiplier}x`}>
            {group.models.map((model) => {
              const id = `studio:${group.id}:${encodeURIComponent(model.id)}:${config.async_enabled ? 'async' : 'sync'}`
              const rate = Number.isFinite(group.rate_multiplier) ? `${group.rate_multiplier}x` : ''
              const label = [group.name, model.id, rate].filter(Boolean).join(' · ')
              return <option key={id} value={id}>{label}</option>
            })}
          </optgroup>)}
        </select>
      </label>
      <label className="studio-model-field">
        <span className="studio-model-label" id="studio-optimizer-label">{studioText('优化模型')}</span>
        <select id="studio-optimizer" aria-labelledby="studio-optimizer-label" value={optimizer} className="studio-model-select" onChange={(event) => chooseOptimizer(event.target.value)}>
          <option value="">{studioText('自动 · OpenAI 优先')}{config.prompt_route?.default_model ? ` · ${config.prompt_route.default_model}` : ''}</option>
          {promptGroups.map(([groupID, group]) => <optgroup key={groupID} label={group.name}>
            {group.models.map((item) => {
              const value = `${item.group_id}:${encodeURIComponent(item.model)}`
              return <option key={value} value={value}>{`${group.name} · ${item.model}`}</option>
            })}
          </optgroup>)}
        </select>
      </label>
      {config.prompt_route && <span className="studio-route-note">{studioText('已创建生图智能路由，可在密钥页修改渠道')}</span>}
      {!config.enabled && <span role="status">{studioText("AI 绘图暂未开放")}</span>}
      {config.enabled && !profiles.length && <span role="status">{studioText("当前账号没有可用的生图模型")}</span>}
    </header>
    <main data-home-main data-drag-select-surface className="studio-gallery">
      <div className="safe-area-x studio-gallery-inner"><SearchBar />
        {filterFavorite && !activeCollection ? <FavoriteCollectionsView /> : <TaskGrid />}
      </div>
    </main>
    {config.enabled && profiles.length > 0 && <InputBar />}
    <DetailModal /><Lightbox /><ConfirmDialog /><Toast /><MaskEditorModal /><ImageContextMenu />
    <FavoriteCollectionPickerModal /><ManageCollectionsModal />
  </Fragment>
}
