import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import messages from '@/i18n/locales/zh/helpCenter'
import HelpClaudeTerminal from '../HelpClaudeTerminal.vue'

vi.mock('vue-i18n', () => ({ useI18n: () => ({
  t: (key: string) => key.split('.').reduce<unknown>((value, part) => (value as Record<string, unknown>)[part], messages)
}) }))

describe('Claude terminal illustration', () => {
  const render = () => mount(HelpClaudeTerminal, {
    props: { model: 'example-model', os: 'macos', shell: 'bash', configMode: 'file' }
  })

  it('recreates the supplied terminal without claiming a real response or fixed version', () => {
    const wrapper = render()
    expect(wrapper.text()).toContain('Welcome back!')
    expect(wrapper.text()).toContain('示例，非实际调用')
    expect(wrapper.text()).toContain('不是当前会话或真实调用结果')
    expect(wrapper.text()).not.toContain('v2.1.118')
    expect(wrapper.text()).not.toContain('deepseek-v4-pro')
    expect(wrapper.findAll('.claude-mascot i')).toHaveLength(56)
    expect(wrapper.find('input, button, iframe').exists()).toBe(false)
  })

  it('reacts to model, OS and configuration mode without injecting HTML', async () => {
    const wrapper = render()
    expect(wrapper.get('.launch-command').text()).toContain('~/projects/my-app')
    await wrapper.setProps({ model: '<img src=x onerror=alert(1)>', os: 'windows', shell: 'powershell', configMode: 'terminal' })
    expect(wrapper.findAll('.terminal-model').every(node => node.text() === '<img src=x onerror=alert(1)>')).toBe(true)
    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.get('.launch-command').text()).toContain('C:\\projects\\my-app')
    expect(wrapper.get('.model-status').text()).toContain('当前终端环境变量')
    expect(wrapper.get('.model-status').text()).not.toContain('settings.json')
    await wrapper.setProps({ model: '' })
    expect(wrapper.get('.terminal-model').text()).toBe('YOUR_MODEL_ID')
  })
})
