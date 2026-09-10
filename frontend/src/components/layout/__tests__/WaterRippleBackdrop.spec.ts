import { mount, type VueWrapper } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import WaterRippleBackdrop from '../WaterRippleBackdrop.vue'

interface MediaOptions {
  finePointer?: boolean
  reducedMotion?: boolean
}

function stubMatchMedia({ finePointer = true, reducedMotion = false }: MediaOptions = {}) {
  vi.stubGlobal('matchMedia', vi.fn((query: string) => ({
    matches: query.includes('prefers-reduced-motion') ? reducedMotion : finePointer,
    media: query,
    onchange: null,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })))
}

function createWorkspace() {
  const workspace = document.createElement('div')
  workspace.className = 'app-workspace'

  const content = document.createElement('main')
  content.dataset.testid = 'workspace-content'
  workspace.appendChild(content)

  const header = document.createElement('header')
  header.className = 'app-header'
  workspace.appendChild(header)

  document.body.appendChild(workspace)
  return { workspace, content, header }
}

function movePointer(target: Element, x: number, y: number) {
  target.dispatchEvent(new MouseEvent('pointermove', {
    bubbles: true,
    clientX: x,
    clientY: y,
  }))
}

let wrapper: VueWrapper | null = null

beforeEach(() => {
  vi.useFakeTimers()
  vi.setSystemTime(new Date('2026-09-05T00:00:00Z'))
})

afterEach(() => {
  wrapper?.unmount()
  wrapper = null
  document.body.innerHTML = ''
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe('WaterRippleBackdrop', () => {
  it('emits throttled ripples only inside the workspace content', async () => {
    stubMatchMedia()
    const { workspace, content, header } = createWorkspace()
    wrapper = mount(WaterRippleBackdrop, { attachTo: workspace })

    movePointer(header, 40, 40)
    expect(wrapper.findAll('.water-ripple')).toHaveLength(0)

    movePointer(content, 80, 120)
    await wrapper.vm.$nextTick()
    expect(wrapper.findAll('.water-ripple')).toHaveLength(1)

    vi.advanceTimersByTime(30)
    movePointer(content, 140, 180)
    await wrapper.vm.$nextTick()
    expect(wrapper.findAll('.water-ripple')).toHaveLength(1)

    vi.advanceTimersByTime(80)
    movePointer(content, 84, 124)
    await wrapper.vm.$nextTick()
    expect(wrapper.findAll('.water-ripple')).toHaveLength(1)

    movePointer(content, 220, 240)
    await wrapper.vm.$nextTick()
    expect(wrapper.findAll('.water-ripple')).toHaveLength(2)
  })

  it('caps active ripples and removes them after their animation lifetime', async () => {
    stubMatchMedia()
    const { workspace, content } = createWorkspace()
    wrapper = mount(WaterRippleBackdrop, { attachTo: workspace })

    for (let index = 0; index < 10; index += 1) {
      movePointer(content, 40 + index * 36, 100 + index * 32)
      vi.advanceTimersByTime(80)
    }
    await wrapper.vm.$nextTick()

    expect(wrapper.findAll('.water-ripple')).toHaveLength(7)

    vi.advanceTimersByTime(1400)
    await wrapper.vm.$nextTick()
    expect(wrapper.findAll('.water-ripple')).toHaveLength(0)
  })

  it.each([
    ['reduced motion is requested', { reducedMotion: true }],
    ['the primary pointer is coarse', { finePointer: false }],
  ])('stays inactive when %s', async (_description, media) => {
    stubMatchMedia(media)
    const { workspace, content } = createWorkspace()
    wrapper = mount(WaterRippleBackdrop, { attachTo: workspace })

    movePointer(content, 100, 160)
    await wrapper.vm.$nextTick()

    expect(wrapper.findAll('.water-ripple')).toHaveLength(0)
  })
})
