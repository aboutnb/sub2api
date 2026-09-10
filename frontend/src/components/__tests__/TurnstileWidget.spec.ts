import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'

import TurnstileWidget from '@/components/TurnstileWidget.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

describe('TurnstileWidget', () => {
  afterEach(() => {
    document
      .querySelectorAll('script[src*="challenges.cloudflare.com/turnstile/v0/api.js"]')
      .forEach(script => script.remove())
    delete window.turnstile
  })

  it('handles script failure, shared loading, an already loaded SDK, and an empty site key', async () => {
    const failed = mount(TurnstileWidget, { props: { siteKey: 'failed-site-key' } })
    const failedScript = document.querySelector<HTMLScriptElement>(
      'script[src*="challenges.cloudflare.com/turnstile/v0/api.js"]',
    )
    expect(failedScript).not.toBeNull()
    failedScript!.dispatchEvent(new Event('error'))
    await flushPromises()
    expect(failed.find('[role="status"]').exists()).toBe(false)
    expect(failed.emitted('error')).toEqual([[]])
    failed.unmount()

    const renderCalls: string[] = []
    const removedWidgets: string[] = []
    const first = mount(TurnstileWidget, { props: { siteKey: 'first-site-key' } })
    const second = mount(TurnstileWidget, { props: { siteKey: 'second-site-key' } })
    await flushPromises()

    const scripts = document.querySelectorAll<HTMLScriptElement>(
      'script[src*="challenges.cloudflare.com/turnstile/v0/api.js"]',
    )
    expect(scripts).toHaveLength(1)

    window.turnstile = {
      render: (_container, options) => {
        renderCalls.push(options.sitekey)
        return `widget-${renderCalls.length}`
      },
      reset: () => undefined,
      remove: widgetId => {
        if (widgetId) removedWidgets.push(widgetId)
      },
    }
    scripts[0]!.dispatchEvent(new Event('load'))
    await flushPromises()

    expect(renderCalls).toEqual(['first-site-key', 'second-site-key'])
    expect(first.find('[role="status"]').exists()).toBe(false)
    expect(second.find('[role="status"]').exists()).toBe(false)

    first.unmount()
    second.unmount()
    expect(removedWidgets).toEqual(['widget-1', 'widget-2'])

    const loaded = mount(TurnstileWidget, { props: { siteKey: 'loaded-site-key' } })
    await flushPromises()
    expect(renderCalls).toEqual(['first-site-key', 'second-site-key', 'loaded-site-key'])
    loaded.unmount()

    const empty = mount(TurnstileWidget, { props: { siteKey: '' } })
    await flushPromises()
    expect(empty.find('.turnstile-wrapper').exists()).toBe(false)
    expect(empty.find('[role="status"]').exists()).toBe(false)
    empty.unmount()
  })
})
