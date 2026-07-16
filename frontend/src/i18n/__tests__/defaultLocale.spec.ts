import { beforeEach, describe, expect, it, vi } from 'vitest'

describe('default locale', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.resetModules()
  })

  it('没有保存语言偏好时默认使用中文', async () => {
    vi.spyOn(window.navigator, 'language', 'get').mockReturnValue('en-US')

    const { getLocale } = await import('../index')

    expect(getLocale()).toBe('zh')
  })

  it('保留用户主动选择的英文', async () => {
    localStorage.setItem('sub2api_locale', 'en')

    const { getLocale } = await import('../index')

    expect(getLocale()).toBe('en')
  })
})
