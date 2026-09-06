import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'

import AnnouncementBell from '../AnnouncementBell.vue'
import { useAnnouncementStore } from '@/stores/announcements'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const announcement = {
  id: 7,
  title: 'Keyboard announcement',
  content: 'Announcement body',
  notify_mode: 'none' as const,
  created_at: '2026-07-24T07:30:00Z',
  updated_at: '2026-07-24T07:30:00Z',
}

describe('AnnouncementBell dialogs', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  afterEach(() => {
    document.body.innerHTML = ''
    document.body.classList.remove('modal-open')
    document.body.style.paddingRight = ''
  })

  it('uses keyboard-operable announcement rows and manages nested dialog focus', async () => {
    const store = useAnnouncementStore()
    store.announcements = [{ ...announcement }]
    vi.spyOn(store, 'markAsRead').mockResolvedValue()

    const wrapper = mount(AnnouncementBell, {
      attachTo: document.body,
      global: { stubs: { Icon: true } },
    })
    const opener = wrapper.get<HTMLButtonElement>('button[aria-label="announcements.title"]')
    opener.element.focus()
    await opener.trigger('click')
    await nextTick()

    let dialogs = document.body.querySelectorAll<HTMLElement>('[role="dialog"]')
    expect(dialogs).toHaveLength(1)
    expect(dialogs[0].getAttribute('aria-modal')).toBe('true')
    expect(dialogs[0].getAttribute('aria-labelledby')).toBeTruthy()
    expect(document.body.classList.contains('modal-open')).toBe(true)

    const row = dialogs[0].querySelector<HTMLButtonElement>('button.group')!
    expect(row.tagName).toBe('BUTTON')
    row.focus()
    row.click()
    await nextTick()
    await nextTick()

    dialogs = document.body.querySelectorAll<HTMLElement>('[role="dialog"]')
    expect(dialogs).toHaveLength(2)
    expect(dialogs[0].getAttribute('aria-hidden')).toBe('true')
    expect(dialogs[0].hasAttribute('inert')).toBe(true)
    expect(Number(dialogs[1].style.zIndex)).toBeGreaterThan(Number(dialogs[0].style.zIndex))
    expect(dialogs[1].contains(document.activeElement)).toBe(true)

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await nextTick()
    await nextTick()
    dialogs = document.body.querySelectorAll<HTMLElement>('[role="dialog"]')
    // The leaving detail node can remain briefly for its CSS transition, but it
    // has already left the stack and the list is the sole interactive dialog.
    expect(dialogs[0].hasAttribute('inert')).toBe(false)
    expect(document.activeElement).toBe(row)
    expect(document.body.classList.contains('modal-open')).toBe(true)

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await nextTick()
    await nextTick()
    expect(document.activeElement).toBe(opener.element)
    expect(document.body.classList.contains('modal-open')).toBe(false)

    wrapper.unmount()
  })
})
