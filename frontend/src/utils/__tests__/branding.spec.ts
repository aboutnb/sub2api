import { describe, expect, it } from 'vitest'

import {
  FLOWAI_DARK_LOGO,
  FLOWAI_LIGHT_LOGO,
  resolveBrandLogo,
} from '@/utils/branding'

describe('resolveBrandLogo', () => {
  it('uses the light FlowAI logo by default', () => {
    expect(resolveBrandLogo('', false)).toBe(FLOWAI_LIGHT_LOGO)
  })

  it('switches the bundled FlowAI mark in dark mode', () => {
    expect(resolveBrandLogo('/flowai-logo-mark.svg', true)).toBe(FLOWAI_DARK_LOGO)
  })

  it('keeps a custom configured logo unchanged', () => {
    expect(resolveBrandLogo('/uploads/brand.svg', true)).toBe('/uploads/brand.svg')
  })

  it('falls back to the themed FlowAI logo for invalid URLs', () => {
    expect(resolveBrandLogo('javascript:alert(1)', false)).toBe(FLOWAI_LIGHT_LOGO)
  })
})
