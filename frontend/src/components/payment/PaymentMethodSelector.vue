<template>
  <div>
    <label v-if="!hideLabel" class="mb-3 block text-sm font-medium text-ink dark:text-ink-muted">
      {{ t('payment.paymentMethod') }}
    </label>
    <div
      data-testid="payment-method-grid"
      class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4"
    >
      <button
        v-for="method in sortedMethods"
        :key="method.type"
        type="button"
        :title="methodLabel(method)"
        :aria-pressed="selected === method.type"
        :disabled="!method.available"
        :class="[
          'relative flex h-16 min-w-0 items-center justify-center rounded-md border px-3 transition-colors',
          !method.available
            ? 'cursor-not-allowed border-line bg-surface-muted opacity-50 dark:border-line dark:bg-canvas/30'
            : selected === method.type
              ? methodSelectedClass(method.type)
              : 'border-line-strong bg-transparent text-ink hover:border-primary-400 dark:border-line-strong dark:text-gray-200 dark:hover:border-primary-400',
        ]"
        @click="method.available && emit('select', method.type)"
      >
        <span v-if="selected === method.type" class="absolute right-1.5 top-1.5 flex h-4 w-4 items-center justify-center rounded-full bg-primary-500 text-white">
          <Icon name="check" size="xs" :stroke-width="2.5" />
        </span>
        <span class="flex w-full min-w-0 items-center justify-center gap-2">
          <img :src="methodIcon(method.type)" :alt="methodLabel(method)" class="h-7 w-7 shrink-0 object-contain" />
          <span class="flex min-w-0 flex-col items-start leading-none">
            <span data-testid="payment-method-label" class="block w-full truncate text-base font-semibold">
              {{ methodLabel(method) }}
            </span>
            <span
              v-if="method.fee_rate > 0"
              class="text-[10px] tracking-wide text-ink-muted dark:text-ink-muted"
            >
              {{ t('payment.fee') }} {{ method.fee_rate }}%
            </span>
          </span>
        </span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { METHOD_ORDER, isBuiltInAlipayMethod, isBuiltInWxpayMethod } from './providerConfig'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'
import stripeIcon from '@/assets/icons/stripe.svg'
import airwallexIcon from '@/assets/icons/airwallex.svg'
import paymentIcon from '@/assets/icons/payment.svg'

export interface PaymentMethodOption {
  type: string
  display_name?: string
  fee_rate: number
  available: boolean
}

const props = defineProps<{
  methods: PaymentMethodOption[]
  selected: string
  hideLabel?: boolean
}>()

const emit = defineEmits<{
  select: [type: string]
}>()

const { t } = useI18n()

const METHOD_ICONS: Record<string, string> = {
  alipay: alipayIcon,
  wxpay: wxpayIcon,
  stripe: stripeIcon,
  airwallex: airwallexIcon,
  credit_card: paymentIcon,
}

const sortedMethods = computed(() => {
  const order: readonly string[] = METHOD_ORDER
  return [...props.methods].sort((a, b) => {
    const ai = order.indexOf(a.type)
    const bi = order.indexOf(b.type)
    return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
  })
})

function methodIcon(type: string): string {
  if (isBuiltInAlipayMethod(type)) return METHOD_ICONS.alipay
  if (isBuiltInWxpayMethod(type)) return METHOD_ICONS.wxpay
  if (type === 'airwallex') return METHOD_ICONS.airwallex
  return METHOD_ICONS[type] || paymentIcon
}

function methodLabel(method: PaymentMethodOption): string {
  return method.display_name || t(`payment.methods.${method.type}`, method.type)
}

function methodSelectedClass(type: string): string {
  if (isBuiltInAlipayMethod(type)) return 'border-[#02A9F1] bg-blue-50/60 text-ink-strong ring-1 ring-[#02A9F1]/20 dark:bg-canvas dark:text-gray-100'
  if (isBuiltInWxpayMethod(type)) return 'border-[#09BB07] bg-green-50/60 text-ink-strong ring-1 ring-[#09BB07]/20 dark:bg-canvas dark:text-gray-100'
  if (type === 'stripe') return 'border-[#676BE5] bg-indigo-50/60 text-ink-strong ring-1 ring-[#676BE5]/20 dark:bg-canvas dark:text-gray-100'
  if (type === 'airwallex') return 'border-[#FF6B3D] bg-orange-50/60 text-ink-strong ring-1 ring-[#FF6B3D]/20 dark:border-[#FF8E3C] dark:bg-canvas dark:text-gray-100'
  return 'border-primary-500 bg-primary-50/60 text-ink-strong ring-1 ring-primary-500/20 dark:bg-canvas dark:text-gray-100'
}
</script>
