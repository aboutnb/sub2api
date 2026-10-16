// @vitest-environment jsdom
import { act } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { StudioBootstrap } from '../../../shared/image-studio'
import { studioProfiles } from '../lib/studioProfiles'
import StudioModelPanel from './StudioModelPanel'

vi.mock('../lib/studioBridge', () => ({
  studioStorageNamespace: () => 'image-studio-user-8',
  selectedPromptOptimizer: () => undefined,
  studioPromptOptimizationEnabled: () => true,
  setStudioPromptOptimizationEnabled: vi.fn(),
}))

const config: StudioBootstrap = {
  enabled: true, async_enabled: true, storage_namespace: 'image-studio-user-8', max_body_bytes: 4096,
  groups: [
    { id: 1, name: 'Images A', rate_multiplier: 1, subscription_type: 'standard', models: [{ id: 'shared', platform: 'openai' }] },
    { id: 2, name: 'Images B', rate_multiplier: 2, subscription_type: 'standard', models: [{ id: 'only-b', platform: 'grok' }, { id: 'shared', platform: 'grok' }] },
  ],
  prompt_models: [
    { group_id: 3, group_name: 'Text A', platform: 'openai', model: 'gpt-mini' },
    { group_id: 3, group_name: 'Text A', platform: 'openai', model: 'gpt-5.6-sol' },
    { group_id: 4, group_name: 'Text B', platform: 'gemini', model: 'gemini-flash' },
  ],
}

describe('studio channel and model picker', () => {
  let container: HTMLDivElement
  let root: Root
  const onSelect = vi.fn()
  beforeEach(() => {
    vi.stubGlobal('IS_REACT_ACT_ENVIRONMENT', true)
    localStorage.clear()
    container = document.createElement('div')
    document.body.append(container)
    root = createRoot(container)
    onSelect.mockClear()
  })
  afterEach(() => {
    act(() => root.unmount())
    container.remove()
    vi.unstubAllGlobals()
  })
  function render(value = config) {
    const profiles = studioProfiles(value, 'https://site.test').available
    act(() => root.render(<StudioModelPanel config={value} profiles={profiles} activeProfileId={profiles[0]?.id ?? ''} onSelect={onSelect} />))
  }
  function select(id: string, value: string) {
    const field = container.querySelector<HTMLSelectElement>(`#${id}`)!
    act(() => { field.value = value; field.dispatchEvent(new Event('change', { bubbles: true })) })
  }
  it('keeps model identity when a user explicitly switches image channels', () => {
    render()
    select('studio-image-channel', '2')
    expect(onSelect).toHaveBeenCalledExactlyOnceWith('studio:2:shared:async')
    expect(container.querySelector('#studio-model')?.textContent).not.toContain('only-b')
  })
  it('lists optimizer models from the selected channel and defaults to gpt-5.6-sol', () => {
    render()
    select('studio-prompt-channel', '3')
    const model = container.querySelector<HTMLSelectElement>('#studio-optimizer')!
    expect(model.value).toBe('3:gpt-5.6-sol')
    expect(model.textContent).not.toContain('gemini')
    select('studio-prompt-channel', '4')
    expect(model.value).toBe('4:gemini-flash')
    expect(localStorage.getItem('image-studio-user-8:prompt-optimizer')).toBe('4:gemini-flash')
    const toggle = container.querySelector<HTMLButtonElement>('[role="switch"]')!
    expect(toggle.getAttribute('aria-checked')).toBe('true')
    act(() => toggle.click())
    expect(toggle.getAttribute('aria-checked')).toBe('false')
    expect(container.querySelector<HTMLSelectElement>('#studio-prompt-channel')?.disabled).toBe(true)
    expect(model.disabled).toBe(true)
    expect(localStorage.getItem('image-studio-user-8:prompt-optimizer')).toBe('4:gemini-flash')
  })
  it('defaults optimization on and restores the saved model after toggling back on', () => {
    localStorage.setItem('image-studio-user-8:prompt-optimizer', '4:gemini-flash')
    render()
    const toggle = container.querySelector<HTMLButtonElement>('[role="switch"]')!
    expect(toggle.getAttribute('aria-checked')).toBe('true')
    act(() => toggle.click())
    act(() => toggle.click())
    expect(toggle.getAttribute('aria-checked')).toBe('true')
    expect(container.querySelector<HTMLSelectElement>('#studio-prompt-channel')?.value).toBe('4')
    expect(container.querySelector<HTMLSelectElement>('#studio-optimizer')?.value).toBe('4:gemini-flash')
  })
  it('shows stale optimizer selection without silently replacing its billing route', () => {
    localStorage.setItem('image-studio-user-8:prompt-optimizer', '9:removed')
    render()
    expect(container.querySelector<HTMLSelectElement>('#studio-prompt-channel')?.value).toBe('unavailable')
    expect(container.textContent).toContain('原优化模型已不可用')
    expect(localStorage.getItem('image-studio-user-8:prompt-optimizer')).toBe('9:removed')
  })
  it('disables creation controls when there are no image models or the feature is off', () => {
    render({ ...config, groups: [] })
    expect([...container.querySelectorAll('select')].every((item) => item.disabled)).toBe(true)
    render({ ...config, enabled: false })
    expect(container.querySelector('[role="status"]')?.textContent).toContain('已有作品仍可查看')
    expect([...container.querySelectorAll('select')].every((item) => item.disabled)).toBe(true)
  })
})
