<template>
  <div
    class="group relative flex flex-col overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm transition-[border-color,box-shadow] hover:border-gray-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-800 dark:hover:border-dark-600"
  >
    <div :class="['h-0.5', accentClass]" />

    <div class="flex flex-1 flex-col p-5">
      <!-- Header: name + badge + price -->
      <div class="mb-3 flex items-start justify-between gap-2">
        <div class="min-w-0 flex-1">
          <h3
            :title="plan.name"
            class="h-12 min-w-0 break-words [overflow-wrap:anywhere] text-base font-bold leading-6 text-gray-900 dark:text-white line-clamp-2"
          >
            {{ plan.name }}
          </h3>
          <p v-if="plan.description" class="mt-0.5 text-xs leading-relaxed text-gray-500 dark:text-dark-400 line-clamp-2">
            {{ plan.description }}
          </p>
        </div>
        <div class="shrink-0 text-right">
          <div class="flex items-baseline gap-1">
            <span class="text-xs text-gray-400 dark:text-dark-500">{{ planCurrencySymbol }}</span>
            <span class="text-2xl font-bold tabular-nums text-gray-950 dark:text-white">{{ plan.price }}</span>
            <span v-if="plan.currency" class="text-xs font-medium text-gray-400 dark:text-dark-500">{{ plan.currency }}</span>
          </div>
          <div class="flex items-center justify-end gap-1">
            <span :class="['inline-flex shrink-0 rounded-full px-2 py-0.5 text-[11px] font-medium', badgeLightClass]">
              {{ pLabel }}
            </span>
            <span class="text-[11px] text-gray-400 dark:text-dark-500">/ {{ validitySuffix }}</span>
          </div>
          <div v-if="plan.original_price" class="mt-0.5 flex items-center justify-end gap-1.5">
            <span class="text-xs text-gray-400 line-through dark:text-dark-500">{{ planCurrencySymbol }}{{ plan.original_price }}<template v-if="plan.currency"> {{ plan.currency }}</template></span>
            <span :class="['rounded px-1 py-0.5 text-[10px] font-semibold', discountClass]">{{ discountText }}</span>
          </div>
        </div>
      </div>

      <!-- Group quota info (compact) -->
      <div class="mb-3 grid grid-cols-2 gap-x-3 gap-y-1 border-y border-gray-100 bg-gray-50/60 px-3 py-3 text-xs dark:border-dark-700 dark:bg-dark-900/30">
        <div class="flex items-center justify-between">
          <span class="text-gray-400 dark:text-dark-500">{{ t('payment.planCard.rate') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">{{ rateDisplay }}</span>
        </div>
        <div v-if="hasPeakRate" class="col-span-2 flex items-center justify-between gap-2">
          <span class="text-gray-400 dark:text-dark-500">{{ t('payment.planCard.peakRate') }}</span>
          <span class="text-right font-medium text-amber-700 dark:text-amber-300">{{ peakRateDisplay }}</span>
        </div>
        <div v-if="plan.daily_limit_usd != null" class="flex items-center justify-between">
          <span class="text-gray-400 dark:text-dark-500">{{ t('payment.planCard.dailyLimit') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">${{ plan.daily_limit_usd }}</span>
        </div>
        <div v-if="plan.weekly_limit_usd != null" class="flex items-center justify-between">
          <span class="text-gray-400 dark:text-dark-500">{{ t('payment.planCard.weeklyLimit') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">${{ plan.weekly_limit_usd }}</span>
        </div>
        <div v-if="plan.monthly_limit_usd != null" class="flex items-center justify-between">
          <span class="text-gray-400 dark:text-dark-500">{{ t('payment.planCard.monthlyLimit') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">${{ plan.monthly_limit_usd }}</span>
        </div>
        <div v-if="plan.daily_limit_usd == null && plan.weekly_limit_usd == null && plan.monthly_limit_usd == null" class="flex items-center justify-between">
          <span class="text-gray-400 dark:text-dark-500">{{ t('payment.planCard.quota') }}</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">{{ t('payment.planCard.unlimited') }}</span>
        </div>
        <div v-if="modelScopeLabels.length > 0" class="col-span-2 flex items-center justify-between">
          <span class="text-gray-400 dark:text-dark-500">{{ t('payment.planCard.models') }}</span>
          <div class="flex flex-wrap justify-end gap-1">
            <span v-for="scope in modelScopeLabels" :key="scope"
              class="rounded bg-gray-200/80 px-1.5 py-0.5 text-[10px] font-medium text-gray-600 dark:bg-dark-600 dark:text-gray-300">
              {{ scope }}
            </span>
          </div>
        </div>
      </div>

      <!-- Features list (compact) -->
      <div v-if="plan.features.length > 0" class="mb-3 space-y-1">
        <div v-for="feature in plan.features" :key="feature" class="flex items-start gap-1.5">
          <Icon name="check" size="sm" class="mt-0.5 flex-shrink-0 text-primary-600 dark:text-primary-400" :stroke-width="2.5" />
          <span class="text-xs text-gray-600 dark:text-gray-300">{{ feature }}</span>
        </div>
      </div>

      <div class="flex-1" />

      <!-- Subscribe Button -->
      <button
        type="button"
        class="flex w-full items-center justify-center gap-2 rounded-md bg-gray-950 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-gray-800 dark:bg-white dark:text-gray-950 dark:hover:bg-gray-200"
        @click="emit('select', plan)"
      >
        <span>{{ isRenewal ? t('payment.renewNow') : t('payment.subscribeNow') }}</span>
        <Icon name="arrowRight" size="sm" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import { useAppStore } from '@/stores/app'
import { hasPeakRate as groupHasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import { planValiditySuffix } from './validity'
import { currencySymbol } from '@/components/payment/currency'
import {
  platformAccentBarClass,
  platformBadgeLightClass,
  platformDiscountClass,
  platformLabel,
} from '@/utils/platformColors'

const props = defineProps<{ plan: SubscriptionPlan; activeSubscriptions?: UserSubscription[] }>()
const emit = defineEmits<{ select: [plan: SubscriptionPlan] }>()
const { t } = useI18n()

const platform = computed(() => props.plan.group_platform || '')
const isRenewal = computed(() =>
  props.activeSubscriptions?.some(s => s.group_id === props.plan.group_id && s.status === 'active') ?? false
)

// Derived color classes from central config
const accentClass = computed(() => platformAccentBarClass(platform.value))
const badgeLightClass = computed(() => platformBadgeLightClass(platform.value))
const discountClass = computed(() => platformDiscountClass(platform.value))
const pLabel = computed(() => platformLabel(platform.value))

const discountText = computed(() => {
  if (!props.plan.original_price || props.plan.original_price <= 0) return ''
  const pct = Math.round((1 - props.plan.price / props.plan.original_price) * 100)
  return pct > 0 ? `-${pct}%` : ''
})

const rateDisplay = computed(() => {
  const rate = props.plan.rate_multiplier ?? 1
  return `×${Number(rate.toPrecision(10))}`
})

const appStore = useAppStore()
const planCurrencySymbol = computed(() => currencySymbol(props.plan.currency || 'USD'))

const hasPeakRate = computed(() => groupHasPeakRate(props.plan))

const peakRateDisplay = computed(() => {
  return formatPeakRateWindow(props.plan, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
})

const MODEL_SCOPE_LABELS: Record<string, string> = {
  claude: 'Claude',
  gemini_text: 'Gemini',
  gemini_image: 'Imagen',
}

const modelScopeLabels = computed(() => {
  if (platform.value !== 'antigravity') return []
  const scopes = props.plan.supported_model_scopes
  if (!scopes || scopes.length === 0) return []
  return scopes.map(s => MODEL_SCOPE_LABELS[s] || s)
})

const validitySuffix = computed(() => planValiditySuffix(props.plan, t))
</script>
