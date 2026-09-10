import { computed } from 'vue'
import { useThemeMode } from '@/composables/useThemeMode'

const SERIES_COUNT = 12

const lightFallback = {
  text: '#3c465a',
  mutedText: '#566074',
  grid: 'rgb(86 96 116 / 0.18)',
  tooltipBackground: '#ffffff',
  tooltipTitle: '#172033',
  tooltipBody: '#3c465a',
  tooltipBorder: '#eadfd4',
  series: [
    '#4964b7', '#087c74', '#c94f2b', '#a33a57', '#a16207', '#6d28d9',
    '#0369a1', '#047857', '#be123c', '#4d7c0f', '#0e7490', '#7e22ce',
  ],
  seriesSoft: [
    'rgb(73 100 183 / 0.12)', 'rgb(8 124 116 / 0.12)', 'rgb(201 79 43 / 0.12)',
    'rgb(163 58 87 / 0.12)', 'rgb(161 98 7 / 0.12)', 'rgb(109 40 217 / 0.12)',
    'rgb(3 105 161 / 0.12)', 'rgb(4 120 87 / 0.12)', 'rgb(190 18 60 / 0.12)',
    'rgb(77 124 15 / 0.12)', 'rgb(14 116 144 / 0.12)', 'rgb(126 34 206 / 0.12)',
  ],
}

const darkFallback = {
  text: '#e7e3de',
  mutedText: '#a6a8ad',
  grid: 'rgb(166 168 173 / 0.22)',
  tooltipBackground: '#263147',
  tooltipTitle: '#f7f5f2',
  tooltipBody: '#e7e3de',
  tooltipBorder: '#344056',
  series: [
    '#7893ec', '#4acbbb', '#ff9c72', '#fba8c0', '#f7c95c', '#a78bfa',
    '#38bdf8', '#34d399', '#fb7185', '#a3e635', '#22d3ee', '#c084fc',
  ],
  seriesSoft: [
    'rgb(120 147 236 / 0.16)', 'rgb(74 203 187 / 0.16)', 'rgb(255 156 114 / 0.16)',
    'rgb(251 168 192 / 0.16)', 'rgb(247 201 92 / 0.16)', 'rgb(167 139 250 / 0.16)',
    'rgb(56 189 248 / 0.16)', 'rgb(52 211 153 / 0.16)', 'rgb(251 113 133 / 0.16)',
    'rgb(163 230 53 / 0.16)', 'rgb(34 211 238 / 0.16)', 'rgb(192 132 252 / 0.16)',
  ],
}

export interface ChartThemePalette {
  text: string
  mutedText: string
  grid: string
  tooltipBackground: string
  tooltipTitle: string
  tooltipBody: string
  tooltipBorder: string
  series: string[]
  seriesSoft: string[]
}

function readToken(styles: CSSStyleDeclaration | null, token: string, fallback: string): string {
  return styles?.getPropertyValue(token).trim() || fallback
}

/**
 * Reactive, semantic color source for every Chart.js/ECharts wrapper.
 * Reading `isDark` makes the computed palette invalidate when the root theme
 * class changes, while CSS custom properties remain the single visual source.
 */
export function useChartTheme() {
  const { isDark } = useThemeMode()

  const chartTheme = computed<ChartThemePalette>(() => {
    const fallback = isDark.value ? darkFallback : lightFallback
    const styles = typeof window === 'undefined' || typeof document === 'undefined'
      ? null
      : window.getComputedStyle(document.documentElement)

    return {
      text: readToken(styles, '--av-chart-text', fallback.text),
      mutedText: readToken(styles, '--av-chart-muted-text', fallback.mutedText),
      grid: readToken(styles, '--av-chart-grid', fallback.grid),
      tooltipBackground: readToken(styles, '--av-chart-tooltip-background', fallback.tooltipBackground),
      tooltipTitle: readToken(styles, '--av-chart-tooltip-title', fallback.tooltipTitle),
      tooltipBody: readToken(styles, '--av-chart-tooltip-body', fallback.tooltipBody),
      tooltipBorder: readToken(styles, '--av-chart-tooltip-border', fallback.tooltipBorder),
      series: Array.from({ length: SERIES_COUNT }, (_, index) =>
        readToken(styles, `--av-chart-series-${index + 1}`, fallback.series[index])
      ),
      seriesSoft: Array.from({ length: SERIES_COUNT }, (_, index) =>
        readToken(styles, `--av-chart-series-${index + 1}-soft`, fallback.seriesSoft[index])
      ),
    }
  })

  const seriesColor = (index: number): string => {
    const colors = chartTheme.value.series
    return colors[((index % colors.length) + colors.length) % colors.length]
  }

  const seriesSoftColor = (index: number): string => {
    const colors = chartTheme.value.seriesSoft
    return colors[((index % colors.length) + colors.length) % colors.length]
  }

  return {
    isDark,
    chartTheme,
    seriesColor,
    seriesSoftColor,
  }
}
