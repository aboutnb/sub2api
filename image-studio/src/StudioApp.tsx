import { studioText } from './lib/studioLocale'
import { Fragment, useEffect, useState } from 'react'
import { useStudioLocale } from './lib/studioLocale'
import { initStore, useStore } from './store'
import { createDefaultOpenAIProfile } from './lib/apiProfiles'
import { studioConfig } from './lib/studioBridge'
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
  useGlobalClickSuppression()
  useEffect(() => { initialize().then(() => setReady(true)).catch((error: Error) => setError(error.message)) }, [])
  if (error) return <p role="alert" className="p-6 text-red-600">{error}</p>
  if (!ready) return <p role="status" className="p-6">{studioText("加载中...")}</p>
  return <Fragment key={locale}>
    <header className="studio-toolbar safe-area-x">
      <label className="studio-model-label" htmlFor="studio-model">{studioText('生图模型')}</label>
      <select id="studio-model" aria-label={studioText("生图模型")} value={settings.activeProfileId} disabled={!config.enabled || !profiles.length}
        className="studio-model-select"
        onChange={(event) => useStore.getState().setSettings({ activeProfileId: event.target.value })}>
        {profiles.map((profile) => <option key={profile.id} value={profile.id}>{profile.name}</option>)}
      </select>
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
