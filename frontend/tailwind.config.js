const aivozaTeal = {
  50: '#edfbf7',
  100: '#dff8f1',
  200: '#b8eee3',
  300: '#7edfd1',
  400: '#2fb9aa',
  500: '#087c74',
  600: '#08645e',
  700: '#0a514d',
  800: '#08433f',
  900: '#063935',
  950: '#032827'
}

const aivozaCoral = {
  50: '#fff7f2',
  100: '#fff0e8',
  200: '#ffd9c7',
  300: '#ffb99a',
  400: '#ff9c72',
  500: '#ff8a5c',
  600: '#eb673b',
  700: '#c94f2b',
  800: '#a94126',
  900: '#893922',
  950: '#4b1a0e'
}

const aivozaDark = {
  50: '#f7f5f2',
  100: '#e7e3de',
  200: '#c8c4bf',
  300: '#a6a8ad',
  400: '#858b99',
  500: '#657084',
  600: '#4a566c',
  700: '#344056',
  800: '#263147',
  900: '#172033',
  950: '#0d1422'
}

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Product components consume mode-aware semantic roles. Tailwind's stock
        // gray, blue, teal and other palettes remain untouched for explicit data,
        // provider-brand and status exceptions.
        canvas: 'rgb(var(--av-rgb-bg-canvas) / <alpha-value>)',
        surface: {
          DEFAULT: 'rgb(var(--av-rgb-bg-surface) / <alpha-value>)',
          muted: 'rgb(var(--av-rgb-bg-surface-muted) / <alpha-value>)',
          elevated: 'rgb(var(--av-rgb-bg-surface-elevated) / <alpha-value>)'
        },
        ink: {
          DEFAULT: 'rgb(var(--av-rgb-text-default) / <alpha-value>)',
          strong: 'rgb(var(--av-rgb-text-strong) / <alpha-value>)',
          muted: 'rgb(var(--av-rgb-text-muted) / <alpha-value>)',
          inverse: 'rgb(var(--av-rgb-text-inverse) / <alpha-value>)'
        },
        line: {
          DEFAULT: 'rgb(var(--av-rgb-border-default) / <alpha-value>)',
          strong: 'rgb(var(--av-rgb-border-strong) / <alpha-value>)',
          control: 'rgb(var(--av-rgb-border-control) / <alpha-value>)'
        },
        brand: {
          DEFAULT: 'rgb(var(--av-rgb-brand-primary) / <alpha-value>)',
          soft: 'rgb(var(--av-rgb-brand-primary-soft) / <alpha-value>)',
          decorative: 'rgb(var(--av-rgb-brand-decorative) / <alpha-value>)',
          accent: 'rgb(var(--av-rgb-brand-accent) / <alpha-value>)'
        },
        action: {
          DEFAULT: 'rgb(var(--av-rgb-action) / <alpha-value>)',
          hover: 'rgb(var(--av-rgb-action-hover) / <alpha-value>)',
          soft: 'rgb(var(--av-rgb-action-soft) / <alpha-value>)',
          foreground: 'rgb(var(--av-rgb-action-foreground) / <alpha-value>)'
        },
        info: 'rgb(var(--av-rgb-info) / <alpha-value>)',
        success: 'rgb(var(--av-rgb-success) / <alpha-value>)',
        warning: 'rgb(var(--av-rgb-warning) / <alpha-value>)',
        danger: 'rgb(var(--av-rgb-danger) / <alpha-value>)',
        primary: aivozaTeal,
        accent: aivozaCoral,
        // Kept temporarily for legacy dark:* utilities while page-owned styles
        // migrate to surface/ink/line roles. It does not override stock palettes.
        dark: aivozaDark
      },
      fontFamily: {
        sans: [
          'Avenir Next',
          'Trebuchet MS',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'sans-serif'
        ],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      boxShadow: {
        glass: 'var(--av-shadow-lg)',
        'glass-sm': 'var(--av-shadow-md)',
        glow: 'var(--av-shadow-glow)',
        card: 'var(--av-shadow-md)',
        'card-hover': 'var(--av-shadow-lg)',
        pixel: 'var(--av-shadow-pixel)',
        'pixel-sm': 'var(--av-shadow-pixel-sm)',
        'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.1)'
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 2s linear infinite',
        glow: 'glow 2s ease-in-out infinite alternate'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' }
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' }
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' }
        },
        glow: {
          '0%': { boxShadow: '3px 3px 0 rgba(47, 185, 170, 0.18)' },
          '100%': { boxShadow: '5px 5px 0 rgba(47, 185, 170, 0.28)' }
        }
      },
      backdropBlur: {
        xs: '2px'
      },
      borderRadius: {
        '4xl': '2rem'
      }
    }
  },
  plugins: []
}
