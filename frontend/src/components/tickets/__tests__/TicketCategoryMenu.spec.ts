import { afterEach, describe, expect, it } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { createTicketI18n } from './i18n'
import TicketCategoryMenu from '../TicketCategoryMenu.vue'

let wrapper: VueWrapper | undefined
afterEach(() => { wrapper?.unmount(); document.body.innerHTML = '' })

function render(props: { modelValue: string; includeAll?: boolean }, locale = 'zh') {
  wrapper = mount(TicketCategoryMenu, {
    props,
    attachTo: document.body,
    global: {
      plugins: [createTicketI18n(locale)],
      stubs: { Icon: true }
    }
  })
  return wrapper
}

describe('TicketCategoryMenu', () => {
  it('uses the system Select and preserves category order with no search field', async () => {
    const view = render({ modelValue: '' })
    expect(view.get('[role="combobox"]').text()).toBe('全部分类')
    expect(document.querySelectorAll('[role="option"]')).toHaveLength(0)
    await view.get('[role="combobox"]').trigger('click')
    expect(Array.from(document.querySelectorAll('[role="option"]'), item => item.textContent?.trim())).toEqual([
      '全部分类', '充值与账单', '接口与模型', '账号与权限', '其他问题'
    ])
    expect(document.querySelector('.select-search-input')).toBeNull()
    const apiOption = document.querySelectorAll<HTMLElement>('[role="option"]')[2]
    apiOption.click()
    expect(view.emitted('update:modelValue')?.[0]).toEqual(['api_models'])
  })

  it('emits an empty string when selecting all and accepts the all filter alias', async () => {
    const view = render({ modelValue: 'all' })
    expect(view.get('[role="combobox"]').text()).toBe('全部分类')
    await view.get('[role="combobox"]').trigger('click')
    document.querySelector<HTMLElement>('[role="option"]')?.click()
    expect(view.emitted('update:modelValue')?.[0]).toEqual([''])
  })

  it('excludes all categories from the creation menu', async () => {
    const view = render({ modelValue: 'billing', includeAll: false })
    await view.get('[role="combobox"]').trigger('click')
    expect(Array.from(document.querySelectorAll('[role="option"]'), item => item.textContent?.trim())).toEqual([
      '充值与账单', '接口与模型', '账号与权限', '其他问题'
    ])
  })

  it('renders translated category labels in English', async () => {
    const view = render({ modelValue: '' }, 'en')
    await view.get('[role="combobox"]').trigger('click')
    expect(Array.from(document.querySelectorAll('[role="option"]'), item => item.textContent?.trim())).toEqual([
      'All Categories', 'Recharge & Billing', 'API & Models', 'Account & Permissions', 'Other Issues'
    ])
  })
})
