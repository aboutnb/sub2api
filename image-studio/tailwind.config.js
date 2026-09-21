/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}', './node_modules/streamdown/dist/*.js'],
  theme: {
    extend: {
      colors: {
        background: 'rgb(var(--av-rgb-bg-canvas) / <alpha-value>)',
        border: 'rgb(var(--av-rgb-border-default) / <alpha-value>)',
        gray: Object.fromEntries([50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950].map((step) => [step, `rgb(var(--studio-gray-${step}) / <alpha-value>)`])),
        blue: Object.fromEntries([50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950].map((step) => [step, `rgb(var(--av-rgb-${step <= 100 ? 'brand-primary-soft' : 'brand-primary'}) / <alpha-value>)`])),
        foreground: 'rgb(var(--av-rgb-text-strong) / <alpha-value>)',
        input: 'rgb(var(--av-rgb-border-control) / <alpha-value>)',
        muted: {
          DEFAULT: 'rgb(var(--av-rgb-bg-surface-muted) / <alpha-value>)',
          foreground: 'rgb(var(--av-rgb-text-muted) / <alpha-value>)',
        },
        primary: {
          DEFAULT: 'rgb(var(--av-rgb-brand-primary) / <alpha-value>)',
          foreground: 'rgb(var(--av-rgb-text-inverse) / <alpha-value>)',
        },
        sidebar: {
          DEFAULT: 'hsl(var(--sidebar) / <alpha-value>)',
          foreground: 'hsl(var(--sidebar-foreground) / <alpha-value>)',
        },
      },
      fontFamily: {
        sans: ['var(--font-ui-sans)'],
        mono: ['var(--font-mono)'],
      },
    },
  },
  plugins: [],
}
