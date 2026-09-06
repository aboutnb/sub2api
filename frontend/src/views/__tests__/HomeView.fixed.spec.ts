import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import HomeView from '../HomeView.vue'

const { appStore } = vi.hoisted(() => ({
  appStore: {
    cachedPublicSettings: {} as Record<string, unknown>,
    siteLogo: '',
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStore,
}))

function mountHome(settings: Record<string, unknown> = {}) {
  appStore.cachedPublicSettings = settings
  return mount(HomeView)
}

describe('HomeView fixed homepage', () => {
  beforeEach(() => {
    appStore.siteLogo = ''
    appStore.cachedPublicSettings = {}
    localStorage.clear()
    document.documentElement.classList.remove('dark')
  })

  it('always renders the reviewed homepage asset and ignores legacy database content', () => {
    const wrapper = mountHome({
      home_content: '<section id="legacy-home">Legacy home</section>',
      compact_home_enabled: true,
    })

    expect(wrapper.get('[data-testid="fixed-home"] .aivoza-home').exists()).toBe(true)
    expect(wrapper.find('#legacy-home').exists()).toBe(false)
    expect(wrapper.find('iframe').exists()).toBe(false)
    expect(wrapper.get('#av-hero-title').text()).toContain('一个 API')
    expect(wrapper.findAll('.av-feature-card')).toHaveLength(3)
  })

  it('uses the configured official logo and escapes it before template interpolation', () => {
    const configuredLogo = `data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg"><text>&'</text></svg>`
    const wrapper = mountHome({ site_logo: configuredLogo })
    const logo = wrapper.get('.av-brand-mark')

    expect(logo.attributes('src')).toBe(configuredLogo)
    expect(wrapper.html()).not.toContain('{{SITE_LOGO}}')
    expect(wrapper.html()).toContain('&quot;http://www.w3.org/2000/svg&quot;')
    expect(wrapper.html()).toContain('&amp;\'')
  })

  it('uses the theme-managed bundled logo when no custom logo is configured', async () => {
    const wrapper = mountHome()
    expect(wrapper.get('.av-brand-mark').attributes('src')).toBe('/flowai-logo-mark-light.svg')

    document.documentElement.classList.add('dark')
    await vi.waitFor(() => {
      expect(wrapper.get('.av-brand-mark').attributes('src')).toBe('/flowai-logo-mark-dark.svg')
    })
  })

  it('keeps the static navigation and console entry links intact', () => {
    const wrapper = mountHome()

    expect(wrapper.get('.av-brand').attributes('href')).toBe('/home')
    expect(wrapper.findAll('a[href="/login"]')).toHaveLength(4)
    expect(wrapper.get('.av-skip-link').attributes('href')).toBe('#av-main')
  })
})
