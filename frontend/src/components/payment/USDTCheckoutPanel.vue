<template>
  <section class="overflow-hidden rounded-lg border border-emerald-200 bg-white shadow-sm dark:border-emerald-900/60 dark:bg-dark-800/70">
    <div class="border-b border-emerald-100 bg-emerald-50/70 px-5 py-4 dark:border-emerald-900/50 dark:bg-emerald-950/20 sm:px-6">
      <div class="flex items-start justify-between gap-4">
        <div>
          <div class="flex items-center gap-2">
            <span class="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-600 text-sm font-bold text-white">₮</span>
            <h2 class="text-base font-semibold text-gray-950 dark:text-white">{{ t('payment.usdt.title') }}</h2>
          </div>
        </div>
        <span class="shrink-0 rounded-full bg-emerald-100 px-2.5 py-1 text-[11px] font-semibold text-emerald-800 dark:bg-emerald-900/50 dark:text-emerald-200">USDT</span>
      </div>
    </div>

    <div v-if="!order" class="grid gap-6 p-5 sm:p-6 lg:grid-cols-[minmax(0,1fr)_280px]">
      <div class="space-y-5">
        <div>
          <label v-if="config.checkout_mode !== 'cashier'" class="mb-2 block text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('payment.usdt.network') }}</label>
          <div v-if="config.checkout_mode !== 'cashier'" class="grid grid-cols-2 gap-2 sm:grid-cols-3">
            <button
              v-for="network in readyNetworks"
              :key="network.network"
              type="button"
              :class="[
                'min-h-16 rounded-md border px-3 py-2 text-left transition-colors',
                selectedNetwork === network.network
                  ? 'border-emerald-600 bg-emerald-50 ring-1 ring-emerald-500/30 dark:border-emerald-400 dark:bg-emerald-950/35'
                  : 'border-gray-200 hover:border-emerald-300 dark:border-dark-600 dark:hover:border-emerald-700',
              ]"
              @click="selectedNetwork = network.network"
            >
              <span class="block text-sm font-semibold text-gray-900 dark:text-white">{{ network.network_name || network.network }}</span>
              <span class="mt-1 block text-[11px] text-gray-500 dark:text-gray-400">{{ network.trade_type }}</span>
            </button>
          </div>
          <p v-if="config.checkout_mode !== 'cashier' && readyNetworks.length === 0" class="mt-2 text-xs text-amber-700 dark:text-amber-300">{{ t('payment.usdt.unavailable') }}</p>
          <p v-if="config.checkout_mode === 'cashier'" class="text-xs leading-5 text-gray-600 dark:text-gray-300">{{ t('payment.usdt.cashierMode') }}</p>
        </div>

        <div>
          <label class="mb-2 block text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('payment.amountLabel') }}</label>
          <AmountInput
            v-model="amount"
            :amounts="[10, 20, 50, 100, 200, 500]"
            currency-symbol="₮"
          />
        </div>
      </div>

      <aside class="flex flex-col justify-between border-t border-gray-100 pt-5 dark:border-dark-700 lg:border-l lg:border-t-0 lg:pl-6 lg:pt-0">
        <div class="space-y-4 text-sm">
          <div class="flex items-center justify-between gap-4"><span class="text-gray-500 dark:text-gray-400">{{ t('payment.usdt.fiatAmount') }}</span><strong class="tabular-nums text-gray-950 dark:text-white">{{ fiatCurrencyLabel }} {{ estimatedFiatAmountLabel }}</strong></div>
          <div class="flex items-center justify-between gap-4"><span class="text-gray-500 dark:text-gray-400">{{ t('payment.usdt.exchangeRate') }}</span><strong class="tabular-nums text-gray-950 dark:text-white">1 USDT = ¥{{ exchangeRateLabel }}</strong></div>
          <p v-if="exchangeRate <= 0" class="border-l-2 border-rose-400 bg-rose-50 px-3 py-2 text-xs leading-5 text-rose-900 dark:bg-rose-950/30 dark:text-rose-100">{{ t('payment.usdt.rateUnavailable') }}</p>
          <div v-if="config.checkout_mode !== 'cashier'" class="flex items-center justify-between gap-4"><span class="text-gray-500 dark:text-gray-400">{{ t('payment.usdt.network') }}</span><strong class="text-gray-950 dark:text-white">{{ selectedNetworkLabel }}</strong></div>
          <p class="border-l-2 border-amber-400 bg-amber-50 px-3 py-2 text-xs leading-5 text-amber-900 dark:bg-amber-950/30 dark:text-amber-100">{{ t('payment.usdt.notice') }}</p>
        </div>
        <button type="button" class="btn mt-6 w-full bg-emerald-600 py-3 text-base font-semibold text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-50" :disabled="!canSubmit || submitting" @click="submit">
          <span v-if="submitting" class="flex items-center justify-center gap-2"><Icon name="refresh" size="sm" class="animate-spin" />{{ t('common.processing') }}</span>
          <span v-else>{{ t('payment.usdt.create') }} <Icon name="arrowRight" size="sm" class="ml-1 inline" /></span>
        </button>
        <p v-if="error" class="mt-3 text-xs text-rose-700 dark:text-rose-300">{{ error }}</p>
      </aside>
    </div>

    <div v-else class="grid gap-6 p-5 sm:p-6 lg:grid-cols-[minmax(0,1fr)_280px]">
      <div>
        <iframe v-if="order.payment_mode === 'cashier'" :src="order.payment_url" class="h-[620px] w-full border-0 bg-white" title="BEpusdt checkout" />
        <template v-else>
          <div class="flex items-center gap-3"><span class="flex h-9 w-9 items-center justify-center rounded-full bg-emerald-600 text-lg font-bold text-white">₮</span><div><p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.usdt.payToAddress') }}</p><p class="break-all font-mono text-sm font-semibold text-gray-950 dark:text-white">{{ order.receiving_address }}</p></div></div>
          <button type="button" class="mt-3 inline-flex items-center gap-1.5 text-xs font-medium text-emerald-700 hover:text-emerald-900 dark:text-emerald-300" @click="copy(order.receiving_address)"><Icon name="copy" size="sm" />{{ t('common.copy') }}</button>
        </template>
        <div v-if="order.payment_mode !== 'cashier'" class="mt-6 rounded-md border border-emerald-200 bg-emerald-50/60 p-4 dark:border-emerald-900/50 dark:bg-emerald-950/20">
          <p class="text-xs font-medium text-emerald-800 dark:text-emerald-200">{{ t('payment.usdt.exactAmount') }}</p>
          <p class="mt-1 break-all font-mono text-3xl font-bold tabular-nums text-emerald-900 dark:text-emerald-100">{{ order.crypto_amount }} USDT</p>
        <p class="mt-2 text-xs text-emerald-800/80 dark:text-emerald-200/80">{{ t('payment.usdt.network') }}: {{ order.network }} · {{ t('payment.usdt.fiatAmount') }}: ¥{{ order.fiat_amount }} · 1 USDT = ¥{{ order.exchange_rate }}</p>
        </div>
      </div>
      <aside class="border-t border-gray-100 pt-5 dark:border-dark-700 lg:border-l lg:border-t-0 lg:pl-6 lg:pt-0">
        <div class="flex items-center justify-between"><span class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.status.pending') }}</span><span :class="statusClass">{{ statusLabel }}</span></div>
        <p class="mt-4 text-xs leading-5 text-gray-600 dark:text-gray-300">{{ t('payment.usdt.statusHint') }}</p>
        <a v-if="order.payment_mode !== 'cashier'" :href="order.payment_url" target="_blank" rel="noopener" class="btn mt-5 flex w-full items-center justify-center gap-2 border border-emerald-600 bg-transparent py-2.5 text-sm font-semibold text-emerald-700 hover:bg-emerald-50 dark:text-emerald-300 dark:hover:bg-emerald-950/30"><Icon name="link" size="sm" />{{ t('payment.usdt.openCheckout') }}</a>
        <button type="button" class="mt-3 w-full text-xs text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white" @click="reset">{{ t('payment.usdt.newOrder') }}</button>
      </aside>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
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
const error = ref('')
let pollTimer: number | undefined

interface USDTRecoverySnapshot {
  orderId: number
  outTradeNo: string
  savedAt: number
}

const readyNetworks = computed(() => props.config.networks.filter(network => network.accepting_orders))
const selectedNetworkLabel = computed(() => readyNetworks.value.find(network => network.network === selectedNetwork.value)?.network_name || selectedNetwork.value || '-')
const cryptoAmount = computed(() => amount.value || 0)
const exchangeRate = computed(() => {
  const value = Number(props.config.rate)
  return Number.isFinite(value) && value > 0 ? value : 0
})
const estimatedFiatAmount = computed(() => cryptoAmount.value * exchangeRate.value)
const fiatCurrencyLabel = 'CNY'
const estimatedFiatAmountLabel = computed(() => estimatedFiatAmount.value > 0 ? estimatedFiatAmount.value.toFixed(8).replace(/0+$/, '').replace(/\.$/, '') : '0')
const exchangeRateLabel = computed(() => exchangeRate.value > 0 ? exchangeRate.value.toFixed(4).replace(/0+$/, '').replace(/\.$/, '') : '-')
const canSubmit = computed(() => (props.config.checkout_mode === 'cashier' || !!selectedNetwork.value) && cryptoAmount.value > 0 && exchangeRate.value > 0)
const statusLabel = computed(() => t(`payment.status.${String(order.value?.status || 'pending').toLowerCase()}`))
const statusClass = computed(() => order.value?.status === 'COMPLETED' ? 'rounded-full bg-emerald-100 px-2 py-1 text-xs font-semibold text-emerald-800 dark:bg-emerald-900/50 dark:text-emerald-200' : 'rounded-full bg-amber-100 px-2 py-1 text-xs font-semibold text-amber-800 dark:bg-amber-900/50 dark:text-amber-200')

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
  try {
    const response = await paymentAPI.createUSDTOrder({ amount: cryptoAmount.value, amount_unit: 'USDT', network: selectedNetwork.value, order_type: 'balance', return_url: window.location.href, payment_source: 'usdt-module' })
    order.value = response.data
    persistRecoverySnapshot(order.value)
    // Cashier mode is rendered inline; fixed mode keeps the explicit external checkout link.
    startPolling()
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : t('payment.usdt.createFailed')
  } finally {
    submitting.value = false
  }
}

function startPolling() {
  stopPolling()
  pollTimer = window.setInterval(async () => {
    if (!order.value) return stopPolling()
    if (isTerminalStatus(order.value.status)) {
      clearRecoverySnapshot()
      return stopPolling()
    }
    try {
      const response = await paymentAPI.getUSDTOrder(order.value.order_id)
      order.value = response.data
      if (isTerminalStatus(order.value.status)) clearRecoverySnapshot()
      else persistRecoverySnapshot(order.value)
    } catch { /* durable server reconciliation continues after browser errors */ }
  }, 2000)
}

function stopPolling() {
  if (pollTimer !== undefined) window.clearInterval(pollTimer)
  pollTimer = undefined
}

function reset() {
  stopPolling()
  order.value = null
  error.value = ''
  clearRecoverySnapshot()
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
    if (isTerminalStatus(restored.status)) clearRecoverySnapshot()
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
onMounted(() => { void restoreOrder() })
onBeforeUnmount(stopPolling)
</script>
