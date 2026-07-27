import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import LuckyCheckinConfirmDialog from '@/components/checkin/LuckyCheckinConfirmDialog.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

const BaseDialogStub = defineComponent({
  props: ['show', 'title'],
  emits: ['close'],
  template: '<section v-if="show"><h2>{{ title }}</h2><slot /><footer><slot name="footer" /></footer></section>',
})

function mountDialog(rewardType: 'multiplier' | 'amount') {
  return mount(LuckyCheckinConfirmDialog, {
    props: {
      show: true,
      rewardType,
      minMultiplier: -0.05,
      maxMultiplier: 0.1,
      submitting: false,
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        Icon: true,
      },
    },
  })
}

describe('LuckyCheckinConfirmDialog', () => {
  it('shows the configured multiplier range and emits confirmation', async () => {
    const wrapper = mountDialog('multiplier')

    expect(wrapper.get('[data-testid="lucky-multiplier-range"]').text()).toContain('-0.05x ~ +0.10x')
    expect(wrapper.find('[data-testid="lucky-positive-probability"]').exists()).toBe(false)
    await wrapper.get('[data-testid="confirm-lucky-checkin"]').trigger('click')
    expect(wrapper.emitted('confirm')).toHaveLength(1)
  })

  it('does not expose an amount range for fixed-amount lucky rewards', () => {
    const wrapper = mountDialog('amount')

    expect(wrapper.find('[data-testid="lucky-multiplier-range"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="lucky-positive-probability"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('checkin.luckyAmountRisk')
  })
})
