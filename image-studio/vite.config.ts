import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { readFileSync } from 'node:fs'

const pkg = JSON.parse(readFileSync('./package.json', 'utf-8'))

export default defineConfig(({ mode }) => ({
  plugins: [react(), {
    name: 'studio-license',
    generateBundle() {
      this.emitFile({ type: 'asset', fileName: 'LICENSE.txt', source: readFileSync('./LICENSE', 'utf-8') })
    },
  }],
  base: '/image-studio-app/',
  build: { outDir: '../backend/internal/web/dist/image-studio-app', emptyOutDir: true },
  define: {
    __APP_VERSION__: JSON.stringify(pkg.version),
    __DEV_PROXY_CONFIG__: 'null',
    ...(mode === 'test' ? {} : {
      'import.meta.env.VITE_DEFAULT_API_URL': '""',
      'import.meta.env.VITE_API_PROXY_AVAILABLE': '"false"',
      'import.meta.env.VITE_API_PROXY_LOCKED': '"false"',
      'import.meta.env.VITE_SHOW_PRESET_CONFIG_ONLY': '"false"',
      'import.meta.env.VITE_SHOW_DEFAULT_CONFIG_ONLY': '"false"',
      'import.meta.env.VITE_LOCK_PRESET_CONFIG_PARAMS': '"false"',
      'import.meta.env.VITE_PREVENT_PRESET_CONFIG_DELETION': '"false"',
    }),
  },
  server: {
    host: true,
    fs: { allow: ['..'] },
  },
}))
