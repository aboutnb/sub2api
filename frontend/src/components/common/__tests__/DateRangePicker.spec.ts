import { afterEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, mount } from '@vue/test-utils'
import { nextTick, ref } from 'vue'

import DateRangePicker from '../DateRangePicker.vue'

enableAutoUnmount(afterEach)
afterEach(() => vi.restoreAllMocks())

const messages: Record<string, string> = {
  'dates.today': 'Today',
  'dates.yesterday': 'Yesterday',
  'dates.last24Hours': 'Last 24 Hours',
  'dates.last7Days': 'Last 7 Days',
  'dates.last14Days': 'Last 14 Days',
  'dates.last30Days': 'Last 30 Days',
  'dates.thisMonth': 'This Month',
  'dates.lastMonth': 'Last Month',
  'dates.startDate': 'Start Date',
  'dates.endDate': 'End Date',
  'dates.apply': 'Apply',
  'dates.selectDateRange': 'Select date range'
}

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => messages[key] ?? key,
    locale: ref('en')
  })
}))

const formatLocalDate = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

describe('DateRangePicker', () => {
  it('escapes clipping ancestors, stays within the viewport and returns keyboard focus', async () => {
    const wrapper = mount(DateRangePicker, {
      attachTo: document.body,
      props: { startDate: '2026-09-01', endDate: '2026-09-11' },
      global: { stubs: { Icon: true } }
    })
    const trigger = wrapper.get<HTMLButtonElement>('.date-picker-trigger')
    vi.spyOn(trigger.element, 'getBoundingClientRect').mockReturnValue({
      x: 300, y: 600, left: 300, top: 600, right: 360, bottom: 644, width: 60, height: 44,
      toJSON: () => ({})
    })
    vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(375)
    vi.spyOn(window, 'innerHeight', 'get').mockReturnValue(700)
    vi.spyOn(HTMLElement.prototype, 'scrollHeight', 'get').mockReturnValue(350)
    await trigger.trigger('click')
    await nextTick()
    const popup = document.getElementById(trigger.attributes('aria-controls'))!
    expect(popup.parentElement).toBe(document.body)
    expect(popup.style.position).toBe('fixed')
    expect(Number.parseFloat(popup.style.left)).toBeGreaterThanOrEqual(8)
    expect(Number.parseFloat(popup.style.left) + Number.parseFloat(popup.style.width)).toBeLessThanOrEqual(367)
    expect(Number.parseFloat(popup.style.top)).toBeLessThan(600)
    const controls = popup.querySelectorAll<HTMLElement>('button, input')
    controls[0].focus()
    controls[0].dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true }))
    expect(document.activeElement).toBe(controls[controls.length - 1])
    popup.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await nextTick()
    expect(trigger.attributes('aria-expanded')).toBe('false')
    await vi.waitFor(() => expect(document.getElementById(popup.id)).toBeNull())
    expect(document.activeElement).toBe(trigger.element)
  })

  it('uses last 24 hours as the default recognized preset', () => {
    const now = new Date()
    const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000)

    const wrapper = mount(DateRangePicker, {
      props: {
        startDate: formatLocalDate(yesterday),
        endDate: formatLocalDate(now)
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true
        }
      }
    })

    expect(wrapper.text()).toContain('Last 24 Hours')
  })

  it('emits range updates with last24Hours preset when applied', async () => {
    const now = new Date()
    const today = formatLocalDate(now)

    const wrapper = mount(DateRangePicker, {
      props: {
        startDate: today,
        endDate: today
      },
      global: {
        stubs: {
          Icon: true,
          Teleport: true
        }
      }
    })

    await wrapper.find('.date-picker-trigger').trigger('click')
    const presetButton = wrapper.findAll('.date-picker-preset').find((node) =>
      node.text().includes('Last 24 Hours')
    )
    expect(presetButton).toBeDefined()

    await presetButton!.trigger('click')
    await wrapper.find('.date-picker-apply').trigger('click')

    const nowAfterClick = new Date()
    const yesterdayAfterClick = new Date(nowAfterClick.getTime() - 24 * 60 * 60 * 1000)
    const expectedStart = formatLocalDate(yesterdayAfterClick)
    const expectedEnd = formatLocalDate(nowAfterClick)

    expect(wrapper.emitted('update:startDate')?.[0]).toEqual([expectedStart])
    expect(wrapper.emitted('update:endDate')?.[0]).toEqual([expectedEnd])
    expect(wrapper.emitted('change')?.[0]).toEqual([
      {
        startDate: expectedStart,
        endDate: expectedEnd,
        preset: 'last24Hours'
      }
    ])
  })
})
