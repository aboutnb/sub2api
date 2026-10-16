import 'core-js/actual/array/at'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { initializeStudio } from './lib/studioBridge'
import './index.css'
import './studioTheme.css'
import { installMobileViewportGuards } from './lib/viewport'
import StudioStatus from './components/StudioStatus'

installMobileViewportGuards()

initializeStudio().then(async () => {
  const { default: StudioApp } = await import('./StudioApp')
  createRoot(document.getElementById('root')!).render(<StrictMode><StudioApp /></StrictMode>)
}).catch((error: Error) => {
  createRoot(document.getElementById('root')!).render(<StudioStatus error={error.message} />)
})
