import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import ConfirmDialog from '../ConfirmDialog.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

describe('ConfirmDialog', () => {
  afterEach(() => {
    document.body.innerHTML = ''
    document.body.classList.remove('modal-open')
  })

  it('prevents duplicate confirmation and cancellation while pending', async () => {
    const wrapper = mount(ConfirmDialog, {
      attachTo: document.body,
      props: {
        show: true,
        title: 'Delete item',
        message: 'This cannot be undone.',
        pending: true
      },
      global: { stubs: { Icon: true } }
    })

    await nextTick()
    const buttons = [...document.body.querySelectorAll<HTMLButtonElement>('.modal-footer button')]
    expect(buttons).toHaveLength(2)
    expect(buttons.every((button) => button.disabled)).toBe(true)
    expect(buttons[1].getAttribute('aria-busy')).toBe('true')

    buttons.forEach((button) => button.click())
    expect(wrapper.emitted('confirm')).toBeUndefined()
    expect(wrapper.emitted('cancel')).toBeUndefined()

    wrapper.unmount()
  })
})
