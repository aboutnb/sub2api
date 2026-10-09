<template>
  <div class="space-y-3">
    <div>
      <label class="mb-1.5 block text-xs font-medium text-ink-muted dark:text-ink-muted">
        {{ t('payment.customAmount') }}
      </label>
      <div class="relative">
        <span class="absolute left-4 top-1/2 -translate-y-1/2 text-lg font-medium text-ink-muted dark:text-ink-muted">
          {{ currencySymbol }}
        </span>
        <input
          type="text"
          inputmode="decimal"
          :value="customText"
          :placeholder="placeholderText"
          class="input h-14 w-full rounded-md pl-9 pr-4 text-xl font-semibold tabular-nums"
          @input="handleInput"
        />
      </div>
    </div>

    <div>
      <label class="mb-1.5 block text-xs font-medium text-ink-muted dark:text-ink-muted">
        {{ t('payment.quickAmounts') }}
      </label>
      <div data-testid="quick-amounts" class="scrollbar-hide grid grid-flow-col auto-cols-[minmax(4.5rem,1fr)] gap-1.5 overflow-x-auto pb-0.5 pr-1 pt-2">
        <button
          v-for="amt in filteredAmounts"
          :key="amt"
          type="button"
          :aria-pressed="modelValue === amt"
          :aria-label="amountAriaLabel(amt)"
          :class="[
            'relative min-h-9 rounded-md border px-2 text-center text-sm font-semibold tabular-nums transition-colors',
            modelValue === amt
              ? 'border-primary-500 bg-primary-50/70 text-primary-700 ring-1 ring-primary-500/20 dark:border-primary-400 dark:bg-canvas dark:text-primary-300'
              : 'border-line-strong bg-transparent text-ink hover:border-primary-400 dark:border-line-strong dark:text-gray-200 dark:hover:border-primary-400',
          ]"
          :data-testid="`quick-amount-${amt}`"
          @click="selectAmount(amt)"
        >
          {{ currencySymbol }}{{ amt }}
          <span
            v-if="amountBadges[amt] || quoteFor(amt).percent > 0"
            aria-hidden="true"
            data-testid="quick-amount-badge"
            class="absolute right-0 top-0 -translate-y-1/2 translate-x-1/4 whitespace-nowrap rounded-sm bg-emerald-600 px-1.5 py-0.5 text-[10px] font-semibold leading-none text-white shadow-sm dark:bg-emerald-500 dark:text-emerald-950"
          >
            {{ amountBadges[amt] || badgeText(amt) }}
          </span>
          <span v-if="showSecondLine" class="block text-[10px] font-normal" data-testid="quick-amount-credited">{{ secondLine(amt) }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { RechargeBonusTier } from '@/types/payment'
import { formatRechargeBonusNumber, quoteRechargeBonus, type RechargeBonusMode } from '@/utils/rechargeBonus'
import { formatPaymentAmount } from './currency'

const props = withDefaults(defineProps<{
  amounts?: number[]
  modelValue: number | null
  min?: number
  max?: number
  currencySymbol?: string
  amountBadges?: Record<number, string>
  /** 充值优惠阶梯（按 min_amount 升序）；为空时不显示价签与第二行 */
  bonusTiers?: RechargeBonusTier[]
  /** 阶梯模式：bonus 赠金 / discount 折扣 */
  bonusMode?: RechargeBonusMode
  /** 充值倍率（1 支付币种 = multiplier USD），用于计算到账金额 */
  multiplier?: number
  /** 支付币种（折扣模式第二行实付金额的币种与精度） */
  currency?: string
}>(), {
  amounts: () => [10, 20, 50, 100, 200, 500],
  min: 0,
  max: 0,
  currencySymbol: '$',
  amountBadges: () => ({}),
  bonusTiers: () => [],
  bonusMode: 'bonus',
  multiplier: 1,
  currency: undefined,
})

const emit = defineEmits<{
  'update:modelValue': [value: number | null]
}>()

const { t } = useI18n()

const customText = ref('')

// 0 = no limit
const filteredAmounts = computed(() =>
  props.amounts.filter((a) => (props.min <= 0 || a >= props.min) && (props.max <= 0 || a <= props.max))
)

const showSecondLine = computed(() => props.bonusTiers.length > 0)

function currencyDigits(): number {
  if (!props.currency) return 2
  try {
    return new Intl.NumberFormat(undefined, { style: 'currency', currency: props.currency }).resolvedOptions().maximumFractionDigits ?? 2
  } catch {
    return 2
  }
}

function quoteFor(amt: number) {
  return quoteRechargeBonus(props.bonusTiers, amt, {
    multiplier: props.multiplier,
    mode: props.bonusMode,
    currencyDigits: currencyDigits(),
  })
}

// 价签文案：赠金「+20%」，折扣「20% OFF」
function badgeText(amt: number): string {
  const percent = formatRechargeBonusNumber(quoteFor(amt).percent)
  return props.bonusMode === 'discount' ? `${percent}% OFF` : `+${percent}%`
}

function secondLine(amt: number): string {
  const quote = quoteFor(amt)
  if (props.bonusMode === 'discount') {
    return t('payment.rechargeBonus.payShort', { amount: formatPaymentAmount(quote.payBase, props.currency) })
  }
  return t('payment.rechargeBonus.creditedShort', { amount: '$' + quote.credited.toFixed(2) })
}

const placeholderText = computed(() => {
  if (props.min > 0 && props.max > 0) return `${props.min} - ${props.max}`
  if (props.min > 0) return `≥ ${props.min}`
  if (props.max > 0) return `≤ ${props.max}`
  return t('payment.enterAmount')
})

const AMOUNT_PATTERN = /^\d*(\.\d{0,2})?$/

function selectAmount(amt: number) {
  customText.value = String(amt)
  emit('update:modelValue', amt)
}

function amountAriaLabel(amt: number): string {
  const badge = props.amountBadges[amt]
  return badge ? `${props.currencySymbol}${amt}, ${badge}` : `${props.currencySymbol}${amt}`
}

function handleInput(e: Event) {
  const input = e.target as HTMLInputElement
  const val = input.value
  if (!AMOUNT_PATTERN.test(val)) {
    input.value = customText.value
    return
  }
  customText.value = val
  if (val === '') {
    emit('update:modelValue', null)
    return
  }
  const num = parseFloat(val)
  if (!isNaN(num) && num > 0) {
    emit('update:modelValue', num)
  } else {
    emit('update:modelValue', null)
  }
}

watch(() => props.modelValue, (v) => {
  if (v !== null && String(v) !== customText.value) {
    customText.value = String(v)
  }
}, { immediate: true })
</script>
