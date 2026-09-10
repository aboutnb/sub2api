import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { stopThemeMode, useThemeMode } from '@/composables/useThemeMode'
import { useChartTheme } from '@/composables/useChartTheme'

const chartComponentSources = import.meta.glob('../../**/*.vue', {
  eager: true,
  query: '?raw',
  import: 'default',
}) as Record<string, string>

const tokenValues = {
  light: {
    '--av-chart-text': '#111111',
    '--av-chart-muted-text': '#222222',
    '--av-chart-grid': '#333333',
    '--av-chart-tooltip-background': '#444444',
    '--av-chart-tooltip-title': '#555555',
    '--av-chart-tooltip-body': '#666666',
    '--av-chart-tooltip-border': '#777777',
    '--av-chart-series-1': '#123456',
    '--av-chart-series-1-soft': 'rgb(18 52 86 / 0.12)',
  },
  dark: {
    '--av-chart-text': '#eeeeee',
    '--av-chart-muted-text': '#dddddd',
    '--av-chart-grid': '#cccccc',
    '--av-chart-tooltip-background': '#bbbbbb',
    '--av-chart-tooltip-title': '#aaaaaa',
    '--av-chart-tooltip-body': '#999999',
    '--av-chart-tooltip-border': '#888888',
    '--av-chart-series-1': '#abcdef',
    '--av-chart-series-1-soft': 'rgb(171 205 239 / 0.16)',
  },
} as const

describe('useChartTheme', () => {
  beforeEach(() => {
    stopThemeMode()
    document.documentElement.classList.remove('dark')
    window.localStorage.clear()

    vi.spyOn(window, 'getComputedStyle').mockImplementation(() => {
      const mode = document.documentElement.classList.contains('dark') ? 'dark' : 'light'
      return {
        getPropertyValue: (name: string) =>
          tokenValues[mode][name as keyof typeof tokenValues.light] ?? '',
      } as CSSStyleDeclaration
    })
  })

  afterEach(() => {
    stopThemeMode()
    document.documentElement.classList.remove('dark')
    window.localStorage.clear()
    vi.restoreAllMocks()
  })

  it('reads semantic chart tokens and recomputes immediately after a theme change', async () => {
    const Harness = defineComponent({
      setup() {
        const { chartTheme } = useChartTheme()
        const { setTheme } = useThemeMode()
        return { chartTheme, setTheme }
      },
      template: `
        <button type="button" @click="setTheme(true)">dark</button>
        <output
          :data-text="chartTheme.text"
          :data-grid="chartTheme.grid"
          :data-series="chartTheme.series[0]"
          :data-soft="chartTheme.seriesSoft[0]"
        />
      `,
    })

    const wrapper = mount(Harness)
    const output = () => wrapper.get('output')

    expect(output().attributes('data-text')).toBe('#111111')
    expect(output().attributes('data-grid')).toBe('#333333')
    expect(output().attributes('data-series')).toBe('#123456')
    expect(output().attributes('data-soft')).toBe('rgb(18 52 86 / 0.12)')

    await wrapper.get('button').trigger('click')
    await nextTick()

    expect(document.documentElement.classList.contains('dark')).toBe(true)
    expect(output().attributes('data-text')).toBe('#eeeeee')
    expect(output().attributes('data-grid')).toBe('#cccccc')
    expect(output().attributes('data-series')).toBe('#abcdef')
    expect(output().attributes('data-soft')).toBe('rgb(171 205 239 / 0.16)')
  })

  it('routes every Chart.js component through the shared composable', () => {
    const chartComponents = Object.entries(chartComponentSources).filter(([, source]) =>
      /from ['"]chart\.js['"]/.test(source)
    )

    expect(chartComponents.length).toBeGreaterThan(0)
    for (const [path, source] of chartComponents) {
      expect(source, path).toContain('useChartTheme')
      expect(source, path).not.toMatch(/classList\.contains\(['"]dark['"]\)/)
    }
  })
})
