import { afterEach, expect, it, vi } from 'vitest'
import { unzipSync } from 'fflate'
import { dataUrlToBlob } from './canvasImage'
import { downloadImageIds, downloadImageEntriesAsZip } from './downloadImages'
import { ensureImageCached } from './imageCache'

vi.mock('./imageCache', () => ({ ensureImageCached: vi.fn() }))
afterEach(() => { vi.unstubAllGlobals(); vi.restoreAllMocks() })

it('decodes local base64 and percent-encoded bytes without fetch', async () => {
  const fetch = vi.fn(() => { throw new Error('Blocked by CSP') })
  vi.stubGlobal('fetch', fetch)
  const blob = await dataUrlToBlob('data:image/png;base64,iVBORw0KGgo=')
  expect(blob.type).toBe('image/png')
  expect([...new Uint8Array(await blob.arrayBuffer())]).toEqual([137, 80, 78, 71, 13, 10, 26, 10])
  const encoded = await dataUrlToBlob('data:image/png,%89PNG%0D%0A%1A%0A')
  expect(await encoded.arrayBuffer()).toEqual(await blob.arrayBuffer())
  expect(fetch).not.toHaveBeenCalled()
})

it('downloads cached images and ZIP files under a restrictive connect-src policy', async () => {
  vi.mocked(ensureImageCached).mockResolvedValue('data:image/png;base64,iVBORw0KGgo=')
  const fetch = vi.fn(() => { throw new Error('Blocked by CSP') })
  vi.stubGlobal('fetch', fetch)
  const blobs: Blob[] = []
  vi.spyOn(URL, 'createObjectURL').mockImplementation((blob) => { blobs.push(blob as Blob); return 'blob:test' })
  const click = vi.fn()
  vi.stubGlobal('document', { createElement: () => ({ click }), body: { appendChild: vi.fn(), removeChild: vi.fn() } })
  vi.stubGlobal('window', { setTimeout: vi.fn() })
  expect(await downloadImageIds(['cached-image'])).toEqual({ successCount: 1, failCount: 0 })
  expect(await downloadImageEntriesAsZip([{ imageId: 'cached-image' }])).toEqual({ successCount: 1, failCount: 0 })
  expect(click).toHaveBeenCalledTimes(2)
  expect(Object.keys(unzipSync(new Uint8Array(await blobs[1].arrayBuffer())))).toEqual(['image-01.png'])
  expect(fetch).not.toHaveBeenCalled()
})
