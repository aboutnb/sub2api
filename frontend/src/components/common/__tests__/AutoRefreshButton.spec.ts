import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import AutoRefreshButton from '../AutoRefreshButton.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

const defaultProps = {
  enabled: false,
  intervalSeconds: 30,
  countdown: 18,
  intervals: [15, 30, 60] as const
}

describe('AutoRefreshButton', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('exposes its popup state and restores focus after Escape', async () => {
    const wrapper = mount(AutoRefreshButton, {
      attachTo: document.body,
      props: defaultProps
    })
    const trigger = wrapper.get('button[aria-haspopup="menu"]')

    await trigger.trigger('click')
    expect(trigger.attributes('aria-expanded')).toBe('true')
    expect(document.body.querySelector('[role="menu"]')).not.toBeNull()

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await nextTick()
    expect(trigger.attributes('aria-expanded')).toBe('false')
    expect(document.activeElement).toBe(trigger.element)
    wrapper.unmount()
  })

  it('emits the chosen interval and closes the popup', async () => {
    const wrapper = mount(AutoRefreshButton, {
      attachTo: document.body,
      props: defaultProps
    })

    await wrapper.get('button[aria-haspopup="menu"]').trigger('click')
    const intervalItems = wrapper.findAll('[role="menuitemradio"]')
    await intervalItems[2].trigger('click')

    expect(wrapper.emitted('update:interval')).toEqual([[60]])
    expect(wrapper.find('[role="menu"]').exists()).toBe(false)
    wrapper.unmount()
  })
})
