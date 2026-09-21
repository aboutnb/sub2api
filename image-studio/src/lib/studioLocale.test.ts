import { expect, it } from 'vitest'
import { setStudioLocale, studioText } from './studioLocale'

it('switches primary controls and keeps unknown user content unchanged', () => {
  setStudioLocale('en-US')
  expect(studioText('生成图像')).toBe('Generate image')
  expect(studioText('my prompt')).toBe('my prompt')
  setStudioLocale('zh-CN')
  expect(studioText('生成图像')).toBe('生成图像')
})
