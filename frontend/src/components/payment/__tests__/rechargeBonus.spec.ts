import { describe, expect, it } from 'vitest'
import {
  calculateRechargeBonusCredit,
  calculateRechargeCreditedAmount,
  formatRechargeBonusPercent,
  normalizeRechargeBonusTiers,
  resolveRechargeBonusPercent,
} from '../rechargeBonus'

const tiers = [
  { min_amount: 50, bonus_percent: 5 },
  { min_amount: 100, bonus_percent: 10 },
]

describe('recharge bonus calculations', () => {
  it('normalizes valid tiers and drops invalid entries', () => {
    expect(normalizeRechargeBonusTiers([
      { min_amount: 100, bonus_percent: 10 },
      { min_amount: 0, bonus_percent: 5 },
      { min_amount: 50, bonus_percent: 5 },
    ])).toEqual(tiers)
  })

  it('uses the highest threshold reached', () => {
    expect(resolveRechargeBonusPercent(49.99, tiers)).toBe(0)
    expect(resolveRechargeBonusPercent(50, tiers)).toBe(5)
    expect(resolveRechargeBonusPercent(55, tiers)).toBe(5)
    expect(resolveRechargeBonusPercent(60, tiers)).toBe(5)
    expect(resolveRechargeBonusPercent(70, tiers)).toBe(5)
    expect(resolveRechargeBonusPercent(99.99, tiers)).toBe(5)
    expect(resolveRechargeBonusPercent(100, tiers)).toBe(10)
    expect(resolveRechargeBonusPercent(500, tiers)).toBe(10)
  })

  it('does not apply the bonus percentage to a credited fee', () => {
    expect(calculateRechargeCreditedAmount(100, 102, 1, true, tiers)).toBe(112)
    expect(calculateRechargeBonusCredit(100, 1, tiers)).toBe(10)
  })

  it('applies the balance multiplier after adding the principal bonus', () => {
    expect(calculateRechargeCreditedAmount(100, 102, 0.14, true, tiers)).toBe(15.68)
    expect(calculateRechargeBonusCredit(100, 0.14, tiers)).toBe(1.4)
  })

  it('formats compact percentage labels', () => {
    expect(formatRechargeBonusPercent(5)).toBe('5')
    expect(formatRechargeBonusPercent(6.5)).toBe('6.5')
    expect(formatRechargeBonusPercent(6.25)).toBe('6.25')
  })
})
