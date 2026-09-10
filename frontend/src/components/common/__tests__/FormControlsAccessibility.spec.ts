import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import Input from '../Input.vue'
import SearchInput from '../SearchInput.vue'
import TextArea from '../TextArea.vue'
import Toggle from '../Toggle.vue'

describe('common form control semantics', () => {
  it('generates an associated input label and announces inline errors', () => {
    const wrapper = mount(Input, {
      props: { modelValue: '', label: 'Email', required: true, error: 'Required' }
    })

    const input = wrapper.get('input')
    expect(input.attributes('id')).toBeTruthy()
    expect(wrapper.get('label').attributes('for')).toBe(input.attributes('id'))
    expect(input.attributes('aria-required')).toBe('true')
    expect(input.attributes('aria-describedby')).toBe(wrapper.get('[role="alert"]').attributes('id'))
  })

  it('generates an associated textarea label and merges external help text', () => {
    const wrapper = mount(TextArea, {
      props: {
        modelValue: '',
        label: 'Notes',
        hint: 'Optional context',
        ariaDescribedby: 'external-help'
      }
    })

    const textarea = wrapper.get('textarea')
    expect(wrapper.get('label').attributes('for')).toBe(textarea.attributes('id'))
    expect(textarea.attributes('aria-describedby')).toContain('external-help')
    expect(textarea.attributes('aria-describedby')).toContain(wrapper.get('.input-hint').attributes('id'))
  })

  it('uses placeholder text as a fallback name only when no visible label is supplied', () => {
    const input = mount(Input, {
      props: { modelValue: '', placeholder: 'Account email' }
    })
    const textarea = mount(TextArea, {
      props: { modelValue: '', placeholder: 'Additional notes' }
    })

    expect(input.get('input').attributes('aria-label')).toBe('Account email')
    expect(textarea.get('textarea').attributes('aria-label')).toBe('Additional notes')
  })

  it('uses a semantic search field with either a visible or fallback name', () => {
    const labelled = mount(SearchInput, {
      props: { modelValue: '', label: 'Search users', placeholder: 'Type a name' },
      global: { stubs: { Icon: true } }
    })
    const labelledInput = labelled.get('input')
    expect(labelledInput.attributes('type')).toBe('search')
    expect(labelled.get('label').attributes('for')).toBe(labelledInput.attributes('id'))

    const fallback = mount(SearchInput, {
      props: { modelValue: '', placeholder: 'Search accounts' },
      global: { stubs: { Icon: true } }
    })
    expect(fallback.get('input').attributes('aria-label')).toBe('Search accounts')
  })

  it('exposes switch state and does not emit while disabled', async () => {
    const wrapper = mount(Toggle, {
      props: { modelValue: true, disabled: true, ariaLabel: 'Enable alerts' }
    })

    const toggle = wrapper.get('button')
    expect(toggle.attributes('role')).toBe('switch')
    expect(toggle.attributes('aria-checked')).toBe('true')
    expect(toggle.attributes('aria-label')).toBe('Enable alerts')
    expect((toggle.element as HTMLButtonElement).disabled).toBe(true)
    expect(toggle.classes()).toContain('h-11')

    await toggle.trigger('click')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })
})
