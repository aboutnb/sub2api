import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AuthLayout from '../AuthLayout.vue'

const { appStore, resolveBrandLogoMock, toggleThemeMock } = vi.hoisted(() => ({
  appStore: {
    siteName: 'Aivoza',
    siteLogo: '/custom-logo.svg',
    cachedPublicSettings: {
      site_subtitle: 'One account, every API',
    } as Record<string, unknown>,
    publicSettingsLoaded: true,
    fetchPublicSettings: vi.fn(),
  },
  resolveBrandLogoMock: vi.fn(() => '/resolved-logo.svg'),
  toggleThemeMock: vi.fn(),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStore,
}))

vi.mock('@/composables/useThemeMode', async () => {
  const { ref } = await import('vue')
  return {
    useThemeMode: () => ({ isDark: ref(false), toggleTheme: toggleThemeMock }),
  }
})

vi.mock('@/utils/branding', () => ({
  resolveBrandLogo: resolveBrandLogoMock,
}))

describe('AuthLayout', () => {
  beforeEach(() => {
    appStore.siteName = 'Aivoza'
    appStore.siteLogo = '/custom-logo.svg'
    appStore.cachedPublicSettings = { site_subtitle: 'One account, every API' }
    appStore.publicSettingsLoaded = true
    appStore.fetchPublicSettings.mockReset()
    resolveBrandLogoMock.mockClear()
    toggleThemeMock.mockReset()
  })

  it('offers an accessible theme control outside the auth card', async () => {
    const wrapper = mount(AuthLayout, { global: { stubs: { Icon: true } } })
    const themeToggle = wrapper.get('.auth-theme-toggle')

    expect(themeToggle.attributes('type')).toBe('button')
    expect(themeToggle.attributes('aria-label')).toBe('nav.darkMode')
    expect(wrapper.get('.auth-card').find('.auth-theme-toggle').exists()).toBe(false)

    await themeToggle.trigger('click')
    expect(toggleThemeMock).toHaveBeenCalledOnce()
  })

  it('keeps dynamic branding and both content slots in the compact auth shell', () => {
    const wrapper = mount(AuthLayout, {
      slots: {
        default: '<form data-testid="auth-form">Sign in form</form>',
        footer: '<a data-testid="auth-footer">Create account</a>',
      },
    })

    expect(wrapper.get('h1').text()).toBe('Aivoza')
    expect(wrapper.get('.auth-brand-subtitle').text()).toBe('One account, every API')
    expect(wrapper.get('img').attributes()).toMatchObject({
      src: '/resolved-logo.svg',
      alt: 'Aivoza logo',
    })
    expect(wrapper.get('[data-testid="auth-form"]').text()).toBe('Sign in form')
    expect(wrapper.get('[data-testid="auth-footer"]').text()).toBe('Create account')
    expect(wrapper.get('.auth-card').classes()).toContain('p-5')
    expect(resolveBrandLogoMock).toHaveBeenCalledWith('/custom-logo.svg', false)
    expect(appStore.fetchPublicSettings).toHaveBeenCalledOnce()
  })

  it('does not show stale branding before public settings finish loading', () => {
    appStore.publicSettingsLoaded = false

    const wrapper = mount(AuthLayout)

    expect(wrapper.find('.auth-brand').exists()).toBe(false)
    expect(wrapper.find('.auth-card').exists()).toBe(true)
  })
})
