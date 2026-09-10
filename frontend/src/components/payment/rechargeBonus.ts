import type { RechargeBonusTier } from '@/types/payment'

function isFinitePositive(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value) && value > 0
}

function roundBalance(value: number): number {
  return Math.round((value + Number.EPSILON) * 100) / 100
}

export function normalizeRechargeBonusTiers(value: unknown): RechargeBonusTier[] {
  if (!Array.isArray(value)) return []
  return value
    .filter((tier): tier is RechargeBonusTier => (
      tier !== null
      && typeof tier === 'object'
      && isFinitePositive((tier as RechargeBonusTier).min_amount)
      && isFinitePositive((tier as RechargeBonusTier).bonus_percent)
      && (tier as RechargeBonusTier).bonus_percent <= 100
    ))
    .map(tier => ({ ...tier }))
    .sort((a, b) => a.min_amount - b.min_amount)
}

export function resolveRechargeBonusPercent(amount: number, tiers: RechargeBonusTier[]): number {
  if (!isFinitePositive(amount)) return 0
  let bestThreshold = 0
  let bonusPercent = 0
  for (const tier of tiers) {
    if (amount >= tier.min_amount && tier.min_amount >= bestThreshold) {
      bestThreshold = tier.min_amount
      bonusPercent = tier.bonus_percent
    }
  }
  return bonusPercent
}

export function calculateRechargeBonusCredit(
  rechargePrincipal: number,
  multiplier: number,
  tiers: RechargeBonusTier[],
): number {
  const bonusPercent = resolveRechargeBonusPercent(rechargePrincipal, tiers)
  if (bonusPercent <= 0) return 0
  const effectiveMultiplier = isFinitePositive(multiplier) ? multiplier : 1
  return roundBalance(rechargePrincipal * bonusPercent / 100 * effectiveMultiplier)
}

export function calculateRechargeCreditedAmount(
  rechargePrincipal: number,
  paidAmount: number,
  multiplier: number,
  feeCredited: boolean,
  tiers: RechargeBonusTier[],
): number {
  if (!isFinitePositive(rechargePrincipal)) return 0
  const effectiveMultiplier = isFinitePositive(multiplier) ? multiplier : 1
  const creditBase = feeCredited ? paidAmount : rechargePrincipal
  const bonusPercent = resolveRechargeBonusPercent(rechargePrincipal, tiers)
  const bonusPrincipal = rechargePrincipal * bonusPercent / 100
  return roundBalance((creditBase + bonusPrincipal) * effectiveMultiplier)
}

export function formatRechargeBonusPercent(value: number): string {
  return value.toFixed(2).replace(/\.00$/, '').replace(/(\.\d)0$/, '$1')
}
