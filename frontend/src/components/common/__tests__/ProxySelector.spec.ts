import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ProxySelector from '../ProxySelector.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

afterEach(() => {
  document.body.innerHTML = ''
})

describe('ProxySelector focus behavior', () => {
  it('does not steal focus on unrelated clicks and keeps pointer-dismiss focus outside', async () => {
    const host = document.createElement('div')
    const outside = document.createElement('input')
    document.body.append(host, outside)

    const wrapper = mount(ProxySelector, {
      attachTo: host,
      props: { modelValue: null, proxies: [] },
      global: { stubs: { Icon: true } },
    })

    outside.focus()
    outside.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()
    expect(document.activeElement).toBe(outside)

    await wrapper.get('button').trigger('click')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)

    outside.focus()
    outside.dispatchEvent(new MouseEvent('click', { bubbles: true }))
    await flushPromises()
    expect(wrapper.get('button').attributes('aria-expanded')).toBe('false')
    expect(document.activeElement).toBe(outside)

    wrapper.unmount()
  })

  it('returns focus to the trigger when Escape closes the open listbox', async () => {
    const host = document.createElement('div')
    document.body.append(host)
    const wrapper = mount(ProxySelector, {
      attachTo: host,
      props: { modelValue: null, proxies: [] },
      global: { stubs: { Icon: true } },
    })

    const trigger = wrapper.get('button')
    await trigger.trigger('click')
    await flushPromises()
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)

    await wrapper.get('[role="dialog"]').trigger('keydown', { key: 'Escape' })
    await flushPromises()
    expect(trigger.attributes('aria-expanded')).toBe('false')
    expect(document.activeElement).toBe(trigger.element)

    wrapper.unmount()
  })

  it('uses a dialog picker when rows include independent test actions', async () => {
    const wrapper = mount(ProxySelector, {
      props: {
        modelValue: null,
        proxies: [{ id: 7, name: 'Proxy 7', protocol: 'http', host: '127.0.0.1', port: 8080 }] as any,
      },
      global: { stubs: { Icon: true } },
    })

    await wrapper.get('button').trigger('click')
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
    expect(wrapper.get('[role="dialog"]').findAll('button')).toHaveLength(4)
    expect(wrapper.find('button[aria-pressed="true"]').exists()).toBe(true)
    wrapper.unmount()
  })
})
