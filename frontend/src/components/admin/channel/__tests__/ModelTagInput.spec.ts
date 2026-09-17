import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ModelTagInput from '../ModelTagInput.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

describe('ModelTagInput', () => {
  it('keeps an empty input compact and adds a model on Enter', async () => {
    const wrapper = mount(ModelTagInput, { props: { models: [] } })
    const input = wrapper.get('input')
    expect(input.classes()).toContain('min-h-0')
    expect(input.classes()).toContain('h-7')
    await input.setValue('gpt-5.5')
    await input.trigger('keydown.enter')
    expect(wrapper.emitted('update:models')).toEqual([[['gpt-5.5']]])
  })

  it('retains accessible removal and deduplicated bulk paste', async () => {
    const wrapper = mount(ModelTagInput, { props: { models: ['gpt-5.5'] } })
    await wrapper.get('button[aria-label="common.remove gpt-5.5"]').trigger('click')
    expect(wrapper.emitted('update:models')?.[0]).toEqual([[]])
    await wrapper.get('input').trigger('paste', {
      clipboardData: { getData: () => 'gpt-5.5\ngpt-6-astra,gpt-6-astra' },
    })
    expect(wrapper.emitted('update:models')?.[1]).toEqual([['gpt-5.5', 'gpt-6-astra']])
  })
})
