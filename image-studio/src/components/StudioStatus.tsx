import { studioText, useStudioLocale } from '../lib/studioLocale'

export default function StudioStatus({ error }: { error?: string }) {
  useStudioLocale()
  return <main className="studio-startup" role={error ? 'alert' : 'status'}>
    <div className="studio-startup-card">
      <h1>{studioText(error ? '工作台加载失败' : '加载中...')}</h1>
      {error && <>
        <p>{error}</p>
        <button type="button" onClick={() => window.location.reload()}>{studioText('重新加载')}</button>
      </>}
    </div>
  </main>
}
