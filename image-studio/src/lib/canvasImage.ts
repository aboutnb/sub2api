import { assertUsableMaskCoverage, classifyMaskAlpha, type MaskCoverage } from './mask'
import { dataUrlToBytes } from './dataUrl'

export async function loadImage(dataUrl: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error('图片加载失败'))
    image.src = dataUrl
  })
}

export async function dataUrlToBlob(dataUrl: string, fallbackType = 'image/png'): Promise<Blob> {
  // 本地图片不经过 fetch，避免被生产环境的 connect-src CSP 拦截。
  if (dataUrl.startsWith('data:')) {
    const separator = dataUrl.indexOf(',')
    if (separator < 0) throw new Error('图片数据格式无效')
    const metadata = dataUrl.slice(5, separator)
    const type = metadata.split(';')[0] || fallbackType
    if (/;base64$/i.test(metadata)) {
      const { bytes } = dataUrlToBytes(`data:${type};base64,${dataUrl.slice(separator + 1)}`)
      return new Blob([new Uint8Array(bytes)], { type })
    }
    const payload = dataUrl.slice(separator + 1)
    const chunks: Uint8Array[] = []
    // 百分号编码可能包含非 UTF-8 图片字节，不能通过 decodeURIComponent 转换。
    for (let offset = 0; offset < payload.length;) {
      if (payload[offset] === '%' && /^[0-9a-f]{2}$/i.test(payload.slice(offset + 1, offset + 3))) {
        chunks.push(Uint8Array.of(parseInt(payload.slice(offset + 1, offset + 3), 16)))
        offset += 3
      } else {
        const point = String.fromCodePoint(payload.codePointAt(offset)!)
        chunks.push(new TextEncoder().encode(point))
        offset += point.length
      }
    }
    return new Blob(chunks as BlobPart[], { type })
  }
  const response = await fetch(dataUrl)
  const blob = await response.blob()
  return blob.type ? blob : new Blob([await blob.arrayBuffer()], { type: fallbackType })
}

export async function imageDataUrlToPngBlob(dataUrl: string): Promise<Blob> {
  const image = await loadImage(dataUrl)
  const canvas = document.createElement('canvas')
  canvas.width = image.naturalWidth
  canvas.height = image.naturalHeight
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('当前浏览器不支持 Canvas')
  ctx.drawImage(image, 0, 0)
  return canvasToBlob(canvas, 'image/png')
}

export async function maskDataUrlToPngBlob(maskDataUrl: string): Promise<Blob> {
  const blob = await dataUrlToBlob(maskDataUrl, 'image/png')
  if (blob.type !== 'image/png') {
    return imageDataUrlToPngBlob(maskDataUrl)
  }
  return blob
}

export async function canvasToBlob(canvas: HTMLCanvasElement, type = 'image/png', quality?: number): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (!blob) reject(new Error('图片导出失败'))
      else resolve(blob)
    }, type, quality)
  })
}

export async function validateMaskMatchesImage(maskDataUrl: string, imageDataUrl: string): Promise<MaskCoverage> {
  const [maskImage, sourceImage] = await Promise.all([loadImage(maskDataUrl), loadImage(imageDataUrl)])
  if (maskImage.naturalWidth !== sourceImage.naturalWidth || maskImage.naturalHeight !== sourceImage.naturalHeight) {
    throw new Error('遮罩尺寸与遮罩主图不一致，请重新绘制遮罩')
  }

  const canvas = document.createElement('canvas')
  canvas.width = maskImage.naturalWidth
  canvas.height = maskImage.naturalHeight
  const ctx = canvas.getContext('2d', { willReadFrequently: true })
  if (!ctx) throw new Error('当前浏览器不支持 Canvas')
  ctx.drawImage(maskImage, 0, 0)
  const coverage = classifyMaskAlpha(ctx.getImageData(0, 0, canvas.width, canvas.height))
  assertUsableMaskCoverage(coverage)
  return coverage
}

export async function createMaskPreviewDataUrl(imageDataUrl: string, maskDataUrl: string): Promise<string> {
  const [image, mask] = await Promise.all([loadImage(imageDataUrl), loadImage(maskDataUrl)])
  if (image.naturalWidth !== mask.naturalWidth || image.naturalHeight !== mask.naturalHeight) {
    throw new Error('遮罩尺寸与遮罩主图不一致，请重新绘制遮罩')
  }

  const canvas = document.createElement('canvas')
  canvas.width = image.naturalWidth
  canvas.height = image.naturalHeight
  const ctx = canvas.getContext('2d', { willReadFrequently: true })
  if (!ctx) throw new Error('当前浏览器不支持 Canvas')

  ctx.drawImage(image, 0, 0)

  const maskCanvas = document.createElement('canvas')
  maskCanvas.width = mask.naturalWidth
  maskCanvas.height = mask.naturalHeight
  const maskCtx = maskCanvas.getContext('2d', { willReadFrequently: true })
  if (!maskCtx) throw new Error('当前浏览器不支持 Canvas')
  maskCtx.drawImage(mask, 0, 0)
  const maskPixels = maskCtx.getImageData(0, 0, maskCanvas.width, maskCanvas.height)

  const overlay = ctx.createImageData(canvas.width, canvas.height)
  for (let i = 0; i < maskPixels.data.length; i += 4) {
    const editStrength = 255 - maskPixels.data[i + 3]
    overlay.data[i] = 59
    overlay.data[i + 1] = 130
    overlay.data[i + 2] = 246
    overlay.data[i + 3] = Math.round(editStrength * 0.58)
  }

  const overlayCanvas = document.createElement('canvas')
  overlayCanvas.width = canvas.width
  overlayCanvas.height = canvas.height
  const overlayCtx = overlayCanvas.getContext('2d')
  if (!overlayCtx) throw new Error('当前浏览器不支持 Canvas')
  overlayCtx.putImageData(overlay, 0, 0)
  ctx.drawImage(overlayCanvas, 0, 0)
  return canvas.toDataURL('image/png')
}
