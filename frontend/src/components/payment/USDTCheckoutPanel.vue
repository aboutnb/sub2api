<template>
  <section class="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800/60">

    <div v-if="!order" class="grid overflow-hidden lg:grid-cols-[minmax(0,1fr)_340px]">
      <section class="min-w-0">
        <div class="border-b border-gray-200 p-5 sm:p-6 dark:border-dark-700">
          <div class="mb-5 flex items-center gap-3">
            <span class="font-mono text-xs font-semibold text-emerald-600 dark:text-emerald-400">01</span>
            <span class="h-px w-6 bg-emerald-500/60" />
            <h2 class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.amountLabel') }}</h2>
          </div>
          <AmountInput v-model="amount" :amounts="[10, 20, 50, 100, 200, 500]" currency-symbol="₮" />
        </div>

        <div v-if="config.checkout_mode !== 'cashier'" class="p-5 sm:p-6">
          <div class="mb-5 flex items-center gap-3">
            <span class="font-mono text-xs font-semibold text-emerald-600 dark:text-emerald-400">02</span>
            <span class="h-px w-6 bg-emerald-500/60" />
            <h2 class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.usdt.network') }}</h2>
          </div>
          <div class="grid grid-cols-2 gap-3 sm:grid-cols-3">
            <button
              v-for="network in readyNetworks"
              :key="network.network"
              type="button"
              :aria-pressed="selectedNetwork === network.network"
              :class="[
                'relative min-h-20 rounded-md border px-3 py-2 text-left transition-colors',
                selectedNetwork === network.network
                  ? 'border-emerald-600 bg-emerald-50 ring-1 ring-emerald-500/30 dark:border-emerald-400 dark:bg-emerald-950/35'
                  : 'border-gray-200 hover:border-emerald-300 dark:border-dark-600 dark:hover:border-emerald-700',
              ]"
              @click="selectedNetwork = network.network"
            >
              <span v-if="selectedNetwork === network.network" class="absolute right-2 top-2 flex h-4 w-4 items-center justify-center rounded-full bg-emerald-500 text-white">
                <Icon name="check" size="xs" :stroke-width="2.5" />
              </span>
              <span class="block break-words text-sm font-semibold leading-5 text-gray-900 dark:text-white">USDT · {{ networkDisplayName(network) }}</span>
              <span class="mt-1 block text-xs font-semibold leading-4 text-emerald-700 dark:text-emerald-300">{{ networkStandardLabel(network.trade_type, network.network) }}</span>
              <span class="block break-all text-[10px] leading-4 text-gray-500 dark:text-gray-400">{{ network.trade_type }}</span>
            </button>
          </div>
          <p v-if="selectedNetworkOption" class="mt-3 border-l-2 border-emerald-500 bg-emerald-50/60 px-3 py-2 text-xs leading-5 text-emerald-900 dark:bg-emerald-950/25 dark:text-emerald-100">
            {{ t('payment.usdt.networkInstruction', { network: networkDisplayName(selectedNetworkOption), standard: networkStandardLabel(selectedNetworkOption.trade_type, selectedNetworkOption.network) }) }}
          </p>
          <p v-if="readyNetworks.length === 0" class="mt-2 text-xs text-amber-700 dark:text-amber-300">{{ t('payment.usdt.unavailable') }}</p>
        </div>
      </section>

      <aside data-testid="usdt-checkout-summary" class="border-t border-gray-200 bg-gray-50/80 lg:border-l lg:border-t-0 dark:border-dark-700 dark:bg-dark-900/45">
        <div class="px-5 pb-3 pt-5">
          <div class="flex items-center gap-2">
            <Icon name="clipboard" size="sm" class="text-emerald-600 dark:text-emerald-400" />
            <h2 class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.checkoutSummary') }}</h2>
          </div>
          <p class="mt-1 truncate text-xs text-gray-500 dark:text-gray-400">{{ t('payment.usdt.title') }}</p>
        </div>
        <div class="px-5 pb-5">
          <div class="space-y-5">
            <div>
              <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('payment.actualPay') }}</p>
              <p class="mt-1 text-3xl font-semibold tabular-nums text-gray-950 dark:text-white">{{ cryptoAmountSummaryLabel }} USDT</p>
            </div>
            <dl class="space-y-3 border-t border-dashed border-gray-300 pt-4 text-sm dark:border-dark-600">
              <div class="flex justify-between gap-4"><dt class="text-gray-500 dark:text-gray-400">{{ t('payment.paymentAmount') }}</dt><dd class="font-medium tabular-nums text-gray-900 dark:text-white">{{ cryptoAmountSummaryLabel }} USDT</dd></div>
              <div class="flex justify-between gap-4"><dt class="text-gray-500 dark:text-gray-400">{{ t('payment.usdt.exchangeRate') }}<span v-if="exchangeRateUpdatedAtLabel" class="mt-0.5 block text-[10px] font-normal text-gray-400 dark:text-gray-500">{{ t('payment.usdt.rateUpdatedAt', { time: exchangeRateUpdatedAtLabel }) }}</span></dt><dd class="font-medium tabular-nums text-gray-900 dark:text-white">¥{{ exchangeRateLabel }} / USDT</dd></div>
              <div class="flex justify-between gap-4"><dt class="text-gray-500 dark:text-gray-400">{{ t('payment.usdt.realtimeConversion') }}</dt><dd class="font-medium tabular-nums text-gray-900 dark:text-white">¥{{ estimatedFiatAmountLabel }}</dd></div>
              <div v-if="bonusPercent > 0" class="flex justify-between gap-4 text-emerald-600 dark:text-emerald-300"><dt>{{ t('payment.usdt.bonusLine', { percent: bonusPercentLabel }) }}</dt><dd class="font-medium tabular-nums">+¥{{ bonusAmountLabel }}</dd></div>
              <div v-if="config.checkout_mode !== 'cashier'" class="flex items-start justify-between gap-4 border-t border-gray-200 pt-3 dark:border-dark-700"><dt class="text-gray-500 dark:text-gray-400">{{ t('payment.usdt.network') }}</dt><dd class="text-right font-medium text-gray-900 dark:text-white">USDT · {{ selectedNetworkLabel }}<span class="block text-xs font-medium text-emerald-700 dark:text-emerald-300">{{ selectedNetworkStandardLabel }}</span></dd></div>
              <div class="flex items-center justify-between gap-4 border-t border-gray-200 pt-3 dark:border-dark-700"><dt class="font-medium text-gray-700 dark:text-gray-300">{{ t('payment.usdt.estimatedBalance') }}</dt><dd class="text-lg font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">¥{{ creditedAmountLabel }}</dd></div>
            </dl>
            <p v-if="exchangeRate <= 0" class="border-l-2 border-rose-400 bg-rose-50 px-3 py-2 text-xs leading-5 text-rose-900 dark:bg-rose-950/30 dark:text-rose-100">{{ t('payment.usdt.rateUnavailable') }}</p>
          </div>
          <button type="button" class="btn btn-primary mt-5 w-full py-3 text-base font-medium" :disabled="!canSubmit || submitting" @click="submit">
            <span v-if="submitting" class="flex items-center justify-center gap-2"><Icon name="refresh" size="sm" class="animate-spin" />{{ t('common.processing') }}</span>
            <span v-else>{{ t('payment.usdt.confirmPayment', { amount: cryptoAmountLabel }) }} <Icon name="arrowRight" size="sm" class="ml-1 inline" /></span>
          </button>
          <p v-if="error" class="mt-3 text-xs text-rose-700 dark:text-rose-300">{{ error }}</p>
        </div>
      </aside>
    </div>

    <div v-else-if="order.payment_mode === 'cashier' && order.payment_url && !isTerminalStatus(order.status)" class="p-5 sm:p-6">
      <div class="border-l-4 border-emerald-500 bg-emerald-50/70 px-4 py-3 dark:bg-emerald-950/25">
        <p class="text-sm font-semibold text-emerald-950 dark:text-emerald-50">{{ t('payment.usdt.cashierTitle') }}</p>
        <p class="mt-1 text-xs leading-5 text-emerald-800 dark:text-emerald-200">{{ t('payment.usdt.cashierHint') }}</p>
      </div>
      <div class="mt-5 grid gap-5 lg:grid-cols-[minmax(0,1fr)_280px]">
        <div class="flex min-h-64 flex-col items-center justify-center rounded-md border border-emerald-200 bg-emerald-50/40 p-6 text-center dark:border-emerald-900/50 dark:bg-emerald-950/15">
          <Icon name="externalLink" size="lg" class="text-emerald-600 dark:text-emerald-300" />
          <p class="mt-3 text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.usdt.cashierPopupTitle') }}</p>
          <p class="mt-1 max-w-md text-xs leading-5 text-gray-600 dark:text-gray-300">{{ t('payment.usdt.cashierPopupHint') }}</p>
          <button type="button" class="btn btn-primary mt-5" @click="openCashier(order.payment_url)">
            <Icon name="externalLink" size="sm" />
            {{ t('payment.usdt.openCashier') }}
          </button>
          <p class="mt-3 text-[11px] leading-4 text-gray-500 dark:text-gray-400">{{ t('payment.usdt.cashierPopupBlockedHint') }}</p>
        </div>
        <aside class="border-t border-gray-100 pt-5 dark:border-dark-700 lg:border-l lg:border-t-0 lg:pl-5 lg:pt-0">
          <div class="flex items-center justify-between"><span class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.status.pending') }}</span><span :class="statusClass">{{ statusLabel }}</span></div>
          <div class="mt-4 flex items-center justify-between gap-4 border-y border-gray-100 py-3 text-sm dark:border-dark-700">
            <span class="text-gray-500 dark:text-gray-400">{{ t('payment.qr.expiresIn') }}</span>
            <strong class="tabular-nums text-gray-950 dark:text-white">{{ countdownDisplay }}</strong>
          </div>
          <p class="mt-4 text-xs leading-5 text-gray-600 dark:text-gray-300">{{ t('payment.usdt.statusHint') }}</p>
          <button v-if="!isTerminalStatus(order.status) && !cancelConfirming" type="button" class="mt-4 w-full rounded-md border border-rose-200 px-3 py-2 text-xs font-medium text-rose-700 hover:bg-rose-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-rose-900/60 dark:text-rose-300 dark:hover:bg-rose-950/30" :disabled="cancelling" @click="beginCancel">
            <Icon v-if="cancelling" name="refresh" size="xs" class="mr-1 inline animate-spin" />
            {{ cancelling ? t('common.processing') : t('payment.usdt.cancel') }}
          </button>
          <div v-if="cancelConfirming" class="mt-4 rounded-md border border-rose-200 bg-rose-50 p-3 dark:border-rose-900/60 dark:bg-rose-950/30" role="alertdialog" aria-live="polite">
            <p class="text-xs leading-5 text-rose-900 dark:text-rose-100">{{ t('payment.usdt.cancelPrompt') }}</p>
            <div class="mt-3 grid grid-cols-2 gap-2">
              <button type="button" class="rounded-md bg-rose-600 px-2 py-2 text-xs font-semibold text-white hover:bg-rose-700 disabled:cursor-not-allowed disabled:opacity-50" :disabled="cancelling" @click="cancelOrder">
                {{ cancelling ? t('common.processing') : t('payment.usdt.confirmCancel') }}
              </button>
              <button type="button" class="rounded-md border border-gray-300 px-2 py-2 text-xs font-medium text-gray-700 hover:bg-white dark:border-dark-600 dark:text-gray-200 dark:hover:bg-dark-800" :disabled="cancelling" @click="cancelConfirming = false">
                {{ t('payment.usdt.keepPaying') }}
              </button>
            </div>
          </div>
          <p v-if="error" class="mt-3 text-xs leading-5 text-rose-700 dark:text-rose-300" role="alert">{{ error }}</p>
          <button v-if="isTerminalStatus(order.status)" type="button" class="mt-4 w-full text-xs text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white" @click="reset">{{ t('payment.usdt.newOrder') }}</button>
        </aside>
      </div>
    </div>
    <div v-else-if="isTerminalStatus(order.status)" class="p-5 sm:p-6">
      <div class="flex min-h-64 flex-col items-center justify-center rounded-md border border-gray-200 bg-gray-50 p-6 text-center dark:border-dark-700 dark:bg-dark-900/40">
        <Icon :name="order.status === 'COMPLETED' ? 'checkCircle' : 'xCircle'" size="lg" :class="order.status === 'COMPLETED' ? 'text-emerald-600 dark:text-emerald-300' : 'text-gray-400 dark:text-gray-500'" />
        <p class="mt-3 text-sm font-semibold text-gray-950 dark:text-white">{{ statusLabel }}</p>
        <p class="mt-1 max-w-md text-xs leading-5 text-gray-600 dark:text-gray-300">{{ t('payment.usdt.terminalHint') }}</p>
        <button type="button" class="btn btn-primary mt-5" @click="reset">{{ t('payment.usdt.newOrder') }}</button>
      </div>
    </div>
    <div v-else class="grid gap-6 p-5 sm:p-6 lg:grid-cols-[minmax(0,1fr)_280px]">
      <div>
        <div class="grid items-start gap-5 sm:grid-cols-[176px_minmax(0,1fr)]">
          <div class="flex h-44 w-44 items-center justify-center border border-gray-200 bg-white p-2 dark:border-dark-600">
            <canvas ref="addressQRCanvas" class="h-40 w-40" aria-label="USDT receiving address QR code"></canvas>
          </div>
          <div>
            <div class="flex items-center gap-3"><span class="flex h-9 w-9 items-center justify-center rounded-full bg-emerald-600 text-lg font-bold text-white">₮</span><div class="min-w-0"><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.usdt.payToAddress') }}</p><p class="break-all font-mono text-sm font-semibold text-gray-950 dark:text-white">{{ order.receiving_address }}</p></div></div>
            <button type="button" class="mt-3 inline-flex items-center gap-1.5 text-xs font-medium text-emerald-700 hover:text-emerald-900 dark:text-emerald-300" @click="copy(order.receiving_address)"><Icon name="copy" size="sm" />{{ t('common.copy') }}</button>
          </div>
        </div>
        <div class="mt-5 border-l-4 border-emerald-500 bg-emerald-50/70 px-4 py-3 dark:bg-emerald-950/25">
          <p class="text-xs font-medium text-emerald-800 dark:text-emerald-200">{{ t('payment.usdt.selectedNetwork') }}</p>
          <p class="mt-1 break-words text-base font-semibold text-emerald-950 dark:text-emerald-50">USDT · {{ orderNetworkName }}</p>
          <p class="mt-1 text-xs font-semibold text-emerald-700 dark:text-emerald-300">{{ orderNetworkStandardLabel }} <span class="font-normal text-emerald-800/70 dark:text-emerald-200/70">({{ order.trade_type }})</span></p>
        </div>
        <div class="mt-6 rounded-md border border-emerald-200 bg-emerald-50/60 p-4 dark:border-emerald-900/50 dark:bg-emerald-950/20">
          <p class="text-xs font-medium text-emerald-800 dark:text-emerald-200">{{ t('payment.usdt.exactAmount') }}</p>
          <p class="mt-1 break-all font-mono text-3xl font-bold tabular-nums text-emerald-900 dark:text-emerald-100">{{ order.crypto_amount }} USDT</p>
        <p class="mt-2 text-xs text-emerald-800/80 dark:text-emerald-200/80">{{ t('payment.usdt.network') }}: USDT · {{ orderNetworkName }} ({{ orderNetworkStandardLabel }}) · {{ t('payment.usdt.fiatAmount') }}: ¥{{ order.fiat_amount }} · 1 USDT = ¥{{ orderExchangeRateLabel }}</p>
        </div>
      </div>
      <aside class="border-t border-gray-100 pt-5 dark:border-dark-700 lg:border-l lg:border-t-0 lg:pl-6 lg:pt-0">
        <div class="flex items-center justify-between"><span class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.status.pending') }}</span><span :class="statusClass">{{ statusLabel }}</span></div>
        <div class="mt-4 flex items-center justify-between gap-4 border-y border-gray-100 py-3 text-sm dark:border-dark-700">
          <span class="text-gray-500 dark:text-gray-400">{{ t('payment.qr.expiresIn') }}</span>
          <strong class="tabular-nums text-gray-950 dark:text-white">{{ countdownDisplay }}</strong>
        </div>
        <p class="mt-4 text-xs leading-5 text-gray-600 dark:text-gray-300">{{ t('payment.usdt.statusHint') }}</p>
        <button v-if="!isTerminalStatus(order.status) && !cancelConfirming" type="button" class="mt-4 w-full rounded-md border border-rose-200 px-3 py-2 text-xs font-medium text-rose-700 hover:bg-rose-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-rose-900/60 dark:text-rose-300 dark:hover:bg-rose-950/30" :disabled="cancelling" @click="beginCancel">
          <Icon v-if="cancelling" name="refresh" size="xs" class="mr-1 inline animate-spin" />
          {{ cancelling ? t('common.processing') : t('payment.usdt.cancel') }}
        </button>
        <div v-if="cancelConfirming" class="mt-4 rounded-md border border-rose-200 bg-rose-50 p-3 dark:border-rose-900/60 dark:bg-rose-950/30" role="alertdialog" aria-live="polite">
          <p class="text-xs leading-5 text-rose-900 dark:text-rose-100">{{ t('payment.usdt.cancelPrompt') }}</p>
          <div class="mt-3 grid grid-cols-2 gap-2">
            <button type="button" class="rounded-md bg-rose-600 px-2 py-2 text-xs font-semibold text-white hover:bg-rose-700 disabled:cursor-not-allowed disabled:opacity-50" :disabled="cancelling" @click="cancelOrder">
              {{ cancelling ? t('common.processing') : t('payment.usdt.confirmCancel') }}
            </button>
            <button type="button" class="rounded-md border border-gray-300 px-2 py-2 text-xs font-medium text-gray-700 hover:bg-white dark:border-dark-600 dark:text-gray-200 dark:hover:bg-dark-800" :disabled="cancelling" @click="cancelConfirming = false">
              {{ t('payment.usdt.keepPaying') }}
            </button>
          </div>
        </div>
        <p v-if="error" class="mt-3 text-xs leading-5 text-rose-700 dark:text-rose-300" role="alert">{{ error }}</p>
        <button v-if="isTerminalStatus(order.status)" type="button" class="mt-4 w-full text-xs text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white" @click="reset">{{ t('payment.usdt.newOrder') }}</button>
      </aside>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import QRCode from 'qrcode'
import { paymentAPI } from '@/api/payment'
import type { USDTConfigResponse, USDTOrder } from '@/types/payment'
import Icon from '@/components/icons/Icon.vue'
import AmountInput from '@/components/payment/AmountInput.vue'

const props = defineProps<{ config: USDTConfigResponse }>()
const { t } = useI18n()
const USDT_RECOVERY_STORAGE_KEY = 'usdt.payment.current'
const amount = ref<number | null>(10)
const selectedNetwork = ref('')
const order = ref<USDTOrder | null>(null)
const submitting = ref(false)
const cancelling = ref(false)
const cancelConfirming = ref(false)
const error = ref('')
let pollTimer: number | undefined
let countdownTimer: number | undefined
const remainingSeconds = ref(0)
const addressQRCanvas = ref<HTMLCanvasElement | null>(null)
const USDT_CREATE_KEY_STORAGE = 'usdt.payment.create-key'
let cashierWindow: Window | null = null

const CASHIER_POPUP_WIDTH = 460
const CASHIER_POPUP_HEIGHT = 760

function cashierPopupFeatures(): string {
  const screenLeft = window.screenLeft ?? window.screenX ?? 0
  const screenTop = window.screenTop ?? window.screenY ?? 0
  const viewportWidth = window.outerWidth || document.documentElement.clientWidth || window.screen.availWidth || CASHIER_POPUP_WIDTH
  const viewportHeight = window.outerHeight || document.documentElement.clientHeight || window.screen.availHeight || CASHIER_POPUP_HEIGHT
  const left = Math.max(Math.round(screenLeft + (viewportWidth - CASHIER_POPUP_WIDTH) / 2), 0)
  const top = Math.max(Math.round(screenTop + (viewportHeight - CASHIER_POPUP_HEIGHT) / 2), 0)
  return `popup,width=${CASHIER_POPUP_WIDTH},height=${CASHIER_POPUP_HEIGHT},left=${left},top=${top},resizable=yes,scrollbars=yes`
}

interface USDTRecoverySnapshot {
  orderId: number
  outTradeNo: string
  savedAt: number
}

interface USDTCreateKeySnapshot {
  fingerprint: string
  key: string
}

function createRequestKey(fingerprint: string): string {
  try {
    const raw = window.sessionStorage.getItem(USDT_CREATE_KEY_STORAGE)
    if (raw) {
      const stored = JSON.parse(raw) as Partial<USDTCreateKeySnapshot>
      if (stored.fingerprint === fingerprint && typeof stored.key === 'string' && stored.key) return stored.key
    }
  } catch { /* fall through to a new key */ }
  const requestID = globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`
  const key = `usdt-order-${requestID}`
  try { window.sessionStorage.setItem(USDT_CREATE_KEY_STORAGE, JSON.stringify({ fingerprint, key })) } catch { /* optional */ }
  return key
}

function clearCreateRequestKey() {
  try { window.sessionStorage.removeItem(USDT_CREATE_KEY_STORAGE) } catch { /* optional */ }
}

const readyNetworks = computed(() => props.config.networks.filter(network => network.accepting_orders))
const selectedNetworkOption = computed(() => readyNetworks.value.find(network => network.network === selectedNetwork.value))
const selectedNetworkLabel = computed(() => selectedNetworkOption.value ? networkDisplayName(selectedNetworkOption.value) : networkDisplayName({ network: selectedNetwork.value }))
const selectedNetworkStandardLabel = computed(() => selectedNetworkOption.value ? networkStandardLabel(selectedNetworkOption.value.trade_type, selectedNetworkOption.value.network) : '-')
const orderNetworkOption = computed(() => readyNetworks.value.find(network => network.network === order.value?.network))
const orderNetworkName = computed(() => orderNetworkOption.value ? networkDisplayName(orderNetworkOption.value) : networkDisplayName({ network: order.value?.network }))
const orderNetworkStandardLabel = computed(() => networkStandardLabel(order.value?.trade_type, order.value?.network))
const cryptoAmount = computed(() => amount.value || 0)
const exchangeRate = computed(() => {
  const value = Number(props.config.rate)
  return Number.isFinite(value) && value > 0 ? value : 0
})
const estimatedFiatAmount = computed(() => cryptoAmount.value * exchangeRate.value)
const estimatedFiatAmountLabel = computed(() => estimatedFiatAmount.value > 0 ? estimatedFiatAmount.value.toFixed(2) : '0.00')
const exchangeRateLabel = computed(() => formatExchangeRate(exchangeRate.value))
const orderExchangeRateLabel = computed(() => formatExchangeRate(order.value?.exchange_rate))
const exchangeRateUpdatedAtLabel = computed(() => {
  const raw = Number(props.config.rate_updated_at)
  if (!Number.isFinite(raw) || raw <= 0) return ''
  const timestamp = raw > 1_000_000_000_000 ? raw : raw * 1000
  const date = new Date(timestamp)
  return Number.isNaN(date.getTime()) ? '' : date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
})
const cryptoAmountLabel = computed(() => cryptoAmount.value > 0 ? cryptoAmount.value.toFixed(8).replace(/0+$/, '').replace(/\.$/, '') : '0')
const cryptoAmountSummaryLabel = computed(() => cryptoAmount.value.toFixed(2))
const bonusPercent = computed(() => {
  const value = Number(props.config.bonus_percent)
  return Number.isFinite(value) && value > 0 ? value : 0
})
const bonusPercentLabel = computed(() => bonusPercent.value.toFixed(2).replace(/0+$/, '').replace(/\.$/, ''))
const roundedFiatAmount = computed(() => Number(estimatedFiatAmountLabel.value))
const bonusAmount = computed(() => bonusPercent.value > 0 ? Number((roundedFiatAmount.value * bonusPercent.value / 100).toFixed(2)) : 0)
const bonusAmountLabel = computed(() => bonusAmount.value.toFixed(2))
const creditedAmountLabel = computed(() => (roundedFiatAmount.value + bonusAmount.value).toFixed(2))
const canSubmit = computed(() => (props.config.checkout_mode === 'cashier' || !!selectedNetwork.value) && cryptoAmount.value > 0 && exchangeRate.value > 0)

function formatExchangeRate(value: string | number | undefined): string {
  const numeric = Number(value)
  return Number.isFinite(numeric) && numeric > 0 ? numeric.toFixed(6) : '-'
}

const statusLabel = computed(() => t(`payment.status.${String(order.value?.status || 'pending').toLowerCase()}`))
const statusClass = computed(() => {
  if (order.value?.status === 'COMPLETED') return 'rounded-full bg-emerald-100 px-2 py-1 text-xs font-semibold text-emerald-800 dark:bg-emerald-900/50 dark:text-emerald-200'
  if (isTerminalStatus(order.value?.status)) return 'rounded-full bg-rose-100 px-2 py-1 text-xs font-semibold text-rose-800 dark:bg-rose-900/50 dark:text-rose-200'
  return 'rounded-full bg-amber-100 px-2 py-1 text-xs font-semibold text-amber-800 dark:bg-amber-900/50 dark:text-amber-200'
})
const countdownDisplay = computed(() => {
  const minutes = Math.floor(remainingSeconds.value / 60)
  const seconds = remainingSeconds.value % 60
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
})

function networkDisplayName(network: { network?: string; network_name?: string }): string {
  return network.network_name || network.network?.toUpperCase() || '-'
}

function networkStandardLabel(tradeType?: string, network?: string): string {
  const normalized = tradeType?.trim().toLowerCase()
  if (normalized === 'usdt.bep20' || normalized === 'bep20') return 'BEP-20'
  if (normalized === 'usdt.trc20' || normalized === 'trc20') return 'TRC-20'
  if (normalized === 'usdt.erc20' || normalized === 'erc20') return 'ERC-20'
  return tradeType?.trim() || network?.toUpperCase() || '-'
}

function isTerminalStatus(status: string | undefined): boolean {
  return status === 'COMPLETED'
    || status === 'EXPIRED'
    || status === 'FAILED'
    || status === 'CANCELLED'
    || status === 'REFUNDED'
}

function persistRecoverySnapshot(current: USDTOrder) {
  if (typeof window === 'undefined' || !current.order_id) return
  try {
    const snapshot: USDTRecoverySnapshot = {
      orderId: current.order_id,
      outTradeNo: current.out_trade_no,
      savedAt: Date.now(),
    }
    window.localStorage.setItem(USDT_RECOVERY_STORAGE_KEY, JSON.stringify(snapshot))
  } catch {
    // Storage can be disabled in private browsing; the server order remains durable.
  }
}

function clearRecoverySnapshot() {
  if (typeof window === 'undefined') return
  try { window.localStorage.removeItem(USDT_RECOVERY_STORAGE_KEY) } catch { /* optional */ }
}

function readRecoverySnapshot(): USDTRecoverySnapshot | null {
  if (typeof window === 'undefined') return null
  try {
    const raw = window.localStorage.getItem(USDT_RECOVERY_STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as Partial<USDTRecoverySnapshot>
    const orderId = parsed.orderId
    if (typeof orderId !== 'number' || !Number.isInteger(orderId) || orderId <= 0 || typeof parsed.outTradeNo !== 'string') return null
    return {
      orderId,
      outTradeNo: parsed.outTradeNo,
      savedAt: typeof parsed.savedAt === 'number' ? parsed.savedAt : 0,
    }
  } catch {
    return null
  }
}

async function submit() {
  if (!canSubmit.value || submitting.value) return
  submitting.value = true
  error.value = ''
  const pendingCashierWindow = props.config.checkout_mode === 'cashier'
    ? window.open('about:blank', 'sub2api-usdt-cashier', cashierPopupFeatures())
    : null
  try {
    const exactAmount = cryptoAmount.value.toFixed(8).replace(/0+$/, '').replace(/\.$/, '')
    const request = { amount: exactAmount, amount_unit: 'USDT' as const, network: selectedNetwork.value, order_type: 'balance', return_url: window.location.href, payment_source: 'usdt-module' }
    const response = await paymentAPI.createUSDTOrder(request, createRequestKey(JSON.stringify(request)))
    order.value = response.data
    if (order.value.payment_mode === 'cashier' && order.value.payment_url) {
      cashierWindow = pendingCashierWindow
      if (cashierWindow && !cashierWindow.closed) cashierWindow.location.href = order.value.payment_url
      else openCashier(order.value.payment_url)
    }
    clearCreateRequestKey()
    persistRecoverySnapshot(order.value)
    startPolling()
  } catch (err: unknown) {
    if (pendingCashierWindow && !pendingCashierWindow.closed) pendingCashierWindow.close()
    error.value = err instanceof Error ? err.message : t('payment.usdt.createFailed')
  } finally {
    submitting.value = false
  }
}

function openCashier(url?: string) {
  if (!url) return
  if (cashierWindow && !cashierWindow.closed) {
    cashierWindow.location.href = url
    cashierWindow.focus()
    return
  }
  cashierWindow = window.open(url, 'sub2api-usdt-cashier', cashierPopupFeatures())
  if (!cashierWindow) error.value = t('payment.usdt.cashierPopupBlocked')
}

function beginCancel() {
  if (!order.value || cancelling.value || isTerminalStatus(order.value.status)) return
  error.value = ''
  cancelConfirming.value = true
}

async function cancelOrder() {
  if (!order.value || cancelling.value || isTerminalStatus(order.value.status)) return
  cancelling.value = true
  error.value = ''
  try {
    await paymentAPI.cancelUSDTOrder(order.value.order_id)
    stopPolling()
    stopCountdown()
    clearRecoverySnapshot()
    if (cashierWindow && !cashierWindow.closed) cashierWindow.close()
    cashierWindow = null
    cancelConfirming.value = false
    order.value = { ...order.value, status: 'CANCELLED' }
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : t('payment.usdt.cancelFailed')
  } finally {
    cancelling.value = false
  }
}

function startPolling() {
  stopPolling()
  pollTimer = window.setInterval(async () => {
    if (!order.value) return stopPolling()
    if (isTerminalStatus(order.value.status)) {
      clearRecoverySnapshot()
      stopCountdown()
      return stopPolling()
    }
    try {
      const response = await paymentAPI.getUSDTOrder(order.value.order_id)
      order.value = response.data
      if (isTerminalStatus(order.value.status)) {
        clearRecoverySnapshot()
        stopCountdown()
      }
      else persistRecoverySnapshot(order.value)
    } catch { /* durable server reconciliation continues after browser errors */ }
  }, 2000)
}

function startCountdown() {
  stopCountdown()
  const expiresAt = order.value ? new Date(order.value.expires_at).getTime() : 0
  if (!Number.isFinite(expiresAt) || expiresAt <= 0) {
    remainingSeconds.value = 0
    return
  }
  const update = () => {
    remainingSeconds.value = Math.max(0, Math.floor((expiresAt - Date.now()) / 1000))
    if (remainingSeconds.value === 0) stopCountdown()
  }
  update()
  if (remainingSeconds.value > 0) countdownTimer = window.setInterval(update, 1000)
}

function stopCountdown() {
  if (countdownTimer !== undefined) window.clearInterval(countdownTimer)
  countdownTimer = undefined
}

async function renderAddressQR() {
  if (!order.value?.receiving_address || !addressQRCanvas.value) return
  try {
    await QRCode.toCanvas(addressQRCanvas.value, order.value.receiving_address, {
      width: 160,
      margin: 1,
      errorCorrectionLevel: 'M',
    })
  } catch { /* the copyable address remains available if canvas rendering fails */ }
}

function stopPolling() {
  if (pollTimer !== undefined) window.clearInterval(pollTimer)
  pollTimer = undefined
}

function reset() {
  stopPolling()
  stopCountdown()
  cancelConfirming.value = false
  cancelling.value = false
  order.value = null
  error.value = ''
  clearRecoverySnapshot()
	clearCreateRequestKey()
}

async function restoreOrder() {
  const snapshot = readRecoverySnapshot()
  if (!snapshot) return
  try {
    const response = await paymentAPI.getUSDTOrder(snapshot.orderId)
    const restored = response.data
    // Never display a snapshot belonging to a different order/account.
    if (snapshot.outTradeNo && restored.out_trade_no !== snapshot.outTradeNo) {
      clearRecoverySnapshot()
      return
    }
    order.value = restored
    if (isTerminalStatus(restored.status)) {
      clearRecoverySnapshot()
      stopCountdown()
    }
    else {
      persistRecoverySnapshot(restored)
      startPolling()
    }
  } catch (err: unknown) {
    const status = typeof err === 'object' && err !== null && 'status' in err
      ? Number((err as { status?: unknown }).status)
      : 0
    if (status === 401 || status === 403 || status === 404) clearRecoverySnapshot()
    // Keep the snapshot for a transient network/API failure; the server order remains durable.
  }
}

async function copy(value: string) {
  try { await navigator.clipboard.writeText(value) } catch { /* clipboard permission is optional */ }
}

watch(readyNetworks, networks => {
  if (!networks.some(network => network.network === selectedNetwork.value)) selectedNetwork.value = networks[0]?.network || ''
}, { immediate: true })
watch(() => order.value?.receiving_address, async address => {
  if (!address) return
  startCountdown()
  await nextTick()
  await renderAddressQR()
})
onMounted(() => { void restoreOrder() })
onBeforeUnmount(() => {
  stopPolling()
  stopCountdown()
})
</script>
