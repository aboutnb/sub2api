import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import AmountInput from '../AmountInput.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

describe('AmountInput quick amount badges', () => {
  it('renders stable top-right badges without changing the amount labels', () => {
    const wrapper = mount(AmountInput, {
      props: {
        modelValue: null,
        amounts: [10, 50, 100],
        amountBadges: { 50: '赠5%', 100: '赠10%' },
      },
    })

    const badges = wrapper.findAll('[data-testid="quick-amount-badge"]')
    expect(badges.map(badge => badge.text())).toEqual(['赠5%', '赠10%'])
    expect(wrapper.findAll('button').map(button => button.text())).toEqual(['$10', '$50 赠5%', '$100 赠10%'])
    expect(wrapper.findAll('button')[1].attributes('aria-label')).toBe('$50, 赠5%')
  })
})
