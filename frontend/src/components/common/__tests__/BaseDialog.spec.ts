import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import BaseDialog from '../BaseDialog.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key })
}))

describe('BaseDialog', () => {
  afterEach(() => {
    document.body.innerHTML = ''
    document.body.classList.remove('modal-open')
  })

  it('resets body scroll position when reopened', async () => {
    const wrapper = mount(BaseDialog, {
      attachTo: document.body,
      props: { show: false, title: 'Details' },
      slots: { default: '<div style="height: 2000px">content</div>' },
      global: { stubs: { Icon: true } }
    })

    await wrapper.setProps({ show: true })
    await nextTick()
    const body = document.body.querySelector<HTMLElement>('.modal-body')
    expect(body).not.toBeNull()
    body!.scrollTop = 480

    await wrapper.setProps({ show: false })
    await wrapper.setProps({ show: true })
    await nextTick()

    expect(document.body.querySelector<HTMLElement>('.modal-body')?.scrollTop).toBe(0)
    wrapper.unmount()
  })

  it('keeps keyboard focus inside the topmost dialog', async () => {
    const wrapper = mount(BaseDialog, {
      attachTo: document.body,
      props: { show: true, title: 'Keyboard dialog', showCloseButton: false },
      slots: {
        default: '<button id="first-action">First</button>',
        footer: '<button id="last-action">Last</button>'
      },
      global: { stubs: { Icon: true } }
    })

    await nextTick()
    const first = document.body.querySelector<HTMLButtonElement>('#first-action')!
    const last = document.body.querySelector<HTMLButtonElement>('#last-action')!
    expect(document.activeElement).toBe(first)

    last.focus()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }))
    expect(document.activeElement).toBe(first)

    document.dispatchEvent(
      new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true })
    )
    expect(document.activeElement).toBe(last)
    wrapper.unmount()
  })

  it('keeps body scroll locked until the final stacked dialog closes', async () => {
    const first = mount(BaseDialog, {
      attachTo: document.body,
      props: { show: true, title: 'First dialog' },
      global: { stubs: { Icon: true } }
    })
    const second = mount(BaseDialog, {
      attachTo: document.body,
      props: { show: true, title: 'Second dialog' },
      global: { stubs: { Icon: true } }
    })

    await nextTick()
    const dialogs = document.body.querySelectorAll<HTMLElement>('[role="dialog"]')
    expect(dialogs).toHaveLength(2)
    expect(Number(dialogs[1].style.zIndex)).toBeGreaterThan(Number(dialogs[0].style.zIndex))
    expect(dialogs[0].getAttribute('aria-hidden')).toBe('true')
    expect(document.body.classList.contains('modal-open')).toBe(true)

    await second.setProps({ show: false })
    expect(document.body.classList.contains('modal-open')).toBe(true)
    expect(first.emitted('close')).toBeUndefined()

    await first.setProps({ show: false })
    expect(document.body.classList.contains('modal-open')).toBe(false)

    second.unmount()
    first.unmount()
  })

  it('only lets the topmost dialog handle Escape', async () => {
    const first = mount(BaseDialog, {
      attachTo: document.body,
      props: { show: true, title: 'First dialog' },
      global: { stubs: { Icon: true } }
    })
    const second = mount(BaseDialog, {
      attachTo: document.body,
      props: { show: true, title: 'Second dialog' },
      global: { stubs: { Icon: true } }
    })

    await nextTick()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))

    expect(first.emitted('close')).toBeUndefined()
    expect(second.emitted('close')).toHaveLength(1)

    second.unmount()
    first.unmount()
  })

  it('blocks every dismissal path while pending', async () => {
    const wrapper = mount(BaseDialog, {
      attachTo: document.body,
      props: {
        show: true,
        title: 'Saving',
        pending: true,
        closeOnClickOutside: true
      },
      global: { stubs: { Icon: true } }
    })

    await nextTick()
    const overlay = document.body.querySelector<HTMLElement>('[role="dialog"]')!
    expect(overlay.getAttribute('aria-busy')).toBe('true')
    expect(overlay.querySelector<HTMLButtonElement>('button')?.disabled).toBe(true)

    overlay.click()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    expect(wrapper.emitted('close')).toBeUndefined()

    wrapper.unmount()
  })

  it('restores focus to the opener when the dialog closes or unmounts', async () => {
    const opener = document.createElement('button')
    opener.textContent = 'Open dialog'
    document.body.appendChild(opener)
    opener.focus()

    const wrapper = mount(BaseDialog, {
      attachTo: document.body,
      props: { show: false, title: 'Focus return' },
      global: { stubs: { Icon: true } }
    })

    await wrapper.setProps({ show: true })
    await nextTick()
    expect(document.activeElement).not.toBe(opener)

    await wrapper.setProps({ show: false })
    await nextTick()
    expect(document.activeElement).toBe(opener)

    opener.focus()
    await wrapper.setProps({ show: true })
    await nextTick()
    wrapper.unmount()
    expect(document.activeElement).toBe(opener)
  })
})
