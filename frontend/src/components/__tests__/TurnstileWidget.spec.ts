import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'

import TurnstileWidget from '@/components/TurnstileWidget.vue'

describe('TurnstileWidget', () => {
  afterEach(() => {
    document.querySelectorAll('script[src*="challenges.cloudflare.com/turnstile/v0/api.js"]').forEach(script => script.remove())
    delete window.turnstile
  })

  it('shares one script load across multiple widget instances', async () => {
    const renderCalls: string[] = []
    const removedWidgets: string[] = []
    const first = mount(TurnstileWidget, { props: { siteKey: 'first-site-key' } })
    const second = mount(TurnstileWidget, { props: { siteKey: 'second-site-key' } })
    await flushPromises()

    const scripts = document.querySelectorAll<HTMLScriptElement>('script[src*="challenges.cloudflare.com/turnstile/v0/api.js"]')
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
    scripts[0].dispatchEvent(new Event('load'))
    await flushPromises()

    expect(renderCalls).toEqual(['first-site-key', 'second-site-key'])

    first.unmount()
    second.unmount()
    expect(removedWidgets).toEqual(['widget-1', 'widget-2'])
  })
})
