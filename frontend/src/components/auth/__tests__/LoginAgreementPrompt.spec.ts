import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

import LoginAgreementPrompt from '../LoginAgreementPrompt.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const documents = [
  { id: 'terms', title: 'Terms', content_md: '# Terms' },
  { id: 'privacy', title: 'Privacy Policy', content_md: '# Privacy' },
]

describe('LoginAgreementPrompt dialog', () => {
  afterEach(() => {
    document.body.innerHTML = ''
    document.body.classList.remove('modal-open')
    document.body.style.paddingRight = ''
  })

  it('exposes modal semantics, traps focus and restores it after Escape rejection', async () => {
    const opener = document.createElement('button')
    opener.textContent = 'Open agreement'
    document.body.appendChild(opener)
    opener.focus()

    const wrapper = mount(LoginAgreementPrompt, {
      attachTo: document.body,
      props: {
        accepted: false,
        documents,
        mode: 'modal',
        updatedAt: '2026-09-05',
        visible: true,
      },
      global: {
        stubs: {
          Icon: true,
          RouterLink: {
            props: ['to'],
            template: '<a href="#"><slot /></a>',
          },
        },
      },
    })
    await nextTick()

    const dialog = document.body.querySelector<HTMLElement>('[role="dialog"]')!
    expect(dialog.getAttribute('aria-modal')).toBe('true')
    expect(dialog.getAttribute('aria-labelledby')).toBeTruthy()
    expect(dialog.getAttribute('aria-describedby')).toBeTruthy()
    expect(dialog.contains(document.activeElement)).toBe(true)
    expect(document.body.classList.contains('modal-open')).toBe(true)

    const focusable = dialog.querySelectorAll<HTMLElement>('a[href], button')
    const first = focusable[0]
    const last = focusable[focusable.length - 1]
    last.focus()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }))
    expect(document.activeElement).toBe(first)

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    expect(wrapper.emitted('reject')).toHaveLength(1)
    await wrapper.setProps({ visible: false })
    await nextTick()
    expect(document.activeElement).toBe(opener)
    expect(document.body.classList.contains('modal-open')).toBe(false)

    wrapper.unmount()
    opener.remove()
  })
})
