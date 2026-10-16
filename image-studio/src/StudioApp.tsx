import { studioText } from './lib/studioLocale'
import { Fragment, useEffect, useState } from 'react'
import { useStudioLocale } from './lib/studioLocale'
import { initStore, useStore } from './store'
import { studioConfig } from './lib/studioBridge'
import { studioProfiles } from './lib/studioProfiles'
import StudioModelPanel from './components/StudioModelPanel'
import StudioStatus from './components/StudioStatus'
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
const profiles = studioProfiles(config, window.location.origin).available

let initialization: Promise<void> | undefined
function initialize() {
  if (initialization) return initialization
  const state = useStore.getState()
  const restored = studioProfiles(config, window.location.origin, state.settings.profiles, state.settings.activeProfileId)
  state.setSettings({ profiles: restored.profiles, customProviders: [],
    activeProfileId: restored.activeProfileId,
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
  if (error) return <StudioStatus error={error} />
  if (!ready) return <StudioStatus />
  return <Fragment key={locale}>
    <header className="studio-toolbar safe-area-x">
      <div className="studio-heading">
        <div><h1>{studioText('AI 绘图')}</h1><p>{studioText('从一句想法，到一张作品。')}</p></div>
        <span className="studio-local-note">{studioText('历史保存在当前浏览器')}</span>
      </div>
      <StudioModelPanel config={config} profiles={profiles} activeProfileId={settings.activeProfileId}
        onSelect={(id) => useStore.getState().setSettings({ activeProfileId: id })} />
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
