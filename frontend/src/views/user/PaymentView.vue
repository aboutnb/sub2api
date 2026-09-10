<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-5">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-primary-500 border-t-transparent"></div>
      </div>
      <template v-else>
        <template v-if="paymentPhase === 'paying'">
          <div class="mx-auto max-w-4xl">
            <PaymentStatusPanel
              :order-id="paymentState.orderId"
              :amount="paymentState.amount"
              :pay-amount="paymentState.payAmount"
              :qr-code="paymentState.qrCode"
              :expires-at="paymentState.expiresAt"
              :payment-type="paymentState.paymentType"
              :pay-url="paymentState.payUrl"
              :order-type="paymentState.orderType"
              :currency="paymentState.currency || selectedCurrency"
              :out-trade-no="paymentState.outTradeNo"
              :mobile-alipay-deep-link="paymentState.alipayMobilePrecreateDeepLink"
              @done="onPaymentDone"
              @success="onPaymentSuccess"
              @settled="onPaymentSettled"
            />
          </div>
        </template>
        <template v-else>
          <header class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
            <div class="min-w-0">
              <h1 class="text-2xl font-semibold text-gray-950 dark:text-white">{{ t('payment.title') }}</h1>
              <div v-if="activeTab === 'recharge'" class="mt-1 flex min-w-0 items-center gap-2 text-sm text-ink-muted dark:text-ink-muted">
                <Icon name="userCircle" size="sm" class="shrink-0" />
                <span class="truncate">{{ user?.username || '' }}</span>
              </div>
            </div>
            <div v-if="activeTab === 'recharge'" class="flex items-end justify-between gap-8 sm:block sm:text-right">
              <span class="block text-xs font-medium text-ink-muted dark:text-ink-muted">{{ t('payment.currentBalance') }}</span>
              <span class="mt-1 block text-xl font-semibold tabular-nums text-gray-950 dark:text-white">${{ user?.balance?.toFixed(2) || '0.00' }}</span>
            </div>
          </header>

          <div
            v-if="tabs.length > 1 && !selectedPlan"
            role="tablist"
            class="grid grid-cols-2 border-b border-line dark:border-line"
          >
            <button
              v-for="tab in tabs"
              :key="tab.key"
              type="button"
              role="tab"
              :aria-selected="activeTab === tab.key"
              class="relative flex min-h-11 items-center justify-center gap-2 px-4 py-3 text-sm font-medium transition-colors after:absolute after:inset-x-5 after:-bottom-px after:h-0.5 after:transition-colors"
              :class="activeTab === tab.key ? 'text-primary-700 after:bg-primary-500 dark:text-primary-300' : 'text-ink-muted after:bg-transparent hover:text-ink-strong dark:text-ink-muted dark:hover:text-gray-200'"
              @click="activeTab = tab.key"
            >
              <Icon :name="tab.key === 'recharge' ? 'creditCard' : 'gift'" size="sm" />
              <span>{{ tab.label }}</span>
            </button>
          </div>

          <!-- Top-up Tab -->
          <template v-if="activeTab === 'recharge'">
            <div v-if="enabledMethods.length === 0" class="card py-16 text-center">
              <Icon name="creditCard" size="xl" class="mx-auto mb-3 text-ink-muted dark:text-dark-600" />
              <p class="text-ink-muted dark:text-ink-muted">{{ t('payment.notAvailable') }}</p>
            </div>
            <div v-else data-test="recharge-checkout-layout" class="grid overflow-hidden rounded-lg border border-line bg-white shadow-sm lg:grid-cols-[minmax(0,1fr)_340px] dark:border-line dark:bg-surface/60">
              <section class="min-w-0">
                <div class="border-b border-line p-5 sm:p-6 dark:border-line">
                  <div class="mb-5 flex items-center gap-3">
                    <span class="font-mono text-xs font-semibold text-primary-600 dark:text-primary-400">01</span>
                    <span class="h-px w-6 bg-primary-500/60" />
                    <h2 class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.amountLabel') }}</h2>
                  </div>
                  <div
                    v-if="validAmount > 0 && rechargeFeeCredited && feeRate > 0"
                    role="status"
                    class="mb-5 flex gap-2.5 border-l-2 border-emerald-500 bg-emerald-50/70 px-3 py-2.5 text-sm font-medium leading-5 text-emerald-950 dark:bg-emerald-950/25 dark:text-emerald-100"
                  >
                    <Icon name="checkCircle" size="md" class="mt-0.5 shrink-0 text-emerald-600 dark:text-emerald-300" />
                    <span>{{ t('payment.feeCreditedNotice', {
                      pay: formatSelectedPaymentAmount(totalAmount),
                      credited: `$${creditedAmount.toFixed(2)}`,
                    }) }}</span>
                  </div>
                  <AmountInput
                    v-model="amount"
                    :amounts="rechargeQuickAmounts"
                    :amount-badges="quickAmountBadges"
                    :min="globalMinAmount"
                    :max="globalMaxAmount"
                  />
                  <p v-if="amountError" class="mt-3 flex items-start gap-2 text-xs text-amber-700 dark:text-amber-300">
                    <Icon name="exclamationTriangle" size="sm" class="mt-0.5 shrink-0" />
                    <span>{{ amountError }}</span>
                  </p>
                </div>
                <div class="p-5 sm:p-6">
                  <div class="mb-5 flex items-center gap-3">
                    <span class="font-mono text-xs font-semibold text-primary-600 dark:text-primary-400">02</span>
                    <span class="h-px w-6 bg-primary-500/60" />
                    <h2 class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.paymentMethod') }}</h2>
                  </div>
                  <PaymentMethodSelector
                    :methods="methodOptions"
                    :selected="selectedMethod"
                    :hide-label="true"
                    @select="selectedMethod = $event"
                  />
                </div>
              </section>

              <aside data-test="recharge-summary" class="border-t border-line bg-surface-muted/80 lg:border-l lg:border-t-0 dark:border-line dark:bg-canvas/45">
                <div class="px-5 pb-3 pt-5">
                  <div class="flex items-center gap-2">
                    <Icon name="clipboard" size="sm" class="text-primary-600 dark:text-primary-400" />
                    <h2 class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.checkoutSummary') }}</h2>
                  </div>
                  <p class="mt-1 truncate text-xs text-ink-muted dark:text-ink-muted">{{ t('payment.rechargeAccount') }}: {{ user?.username || '' }}</p>
                </div>
                <div class="px-5 pb-5">
                  <div v-if="validAmount > 0" class="space-y-5">
                    <div>
                      <p class="text-xs font-medium text-ink-muted dark:text-ink-muted">{{ t('payment.actualPay') }}</p>
                      <p class="mt-1 text-3xl font-semibold tabular-nums text-gray-950 dark:text-white">{{ formatSelectedPaymentAmount(totalAmount) }}</p>
                    </div>

                    <dl class="space-y-3 border-t border-dashed border-line-strong pt-4 text-sm dark:border-line-strong">
                      <div class="flex justify-between gap-4">
                        <dt class="text-ink-muted dark:text-ink-muted">{{ t('payment.paymentAmount') }}</dt>
                        <dd class="font-medium tabular-nums text-ink-strong dark:text-white">{{ formatSelectedPaymentAmount(validAmount) }}</dd>
                      </div>
                      <div v-if="feeRate > 0" class="flex justify-between gap-4">
                        <dt class="text-ink-muted dark:text-ink-muted">{{ t('payment.fee') }} ({{ feeRate }}%)</dt>
                        <dd class="font-medium tabular-nums text-ink-strong dark:text-white">{{ formatSelectedPaymentAmount(feeAmount) }}</dd>
                      </div>
                      <div v-if="activeRechargeBonusPercent > 0" class="flex justify-between gap-4">
                        <dt class="text-ink-muted dark:text-ink-muted">{{ t('payment.rechargeBonus', { percent: formatRechargeBonusPercent(activeRechargeBonusPercent) }) }}</dt>
                        <dd class="font-medium tabular-nums text-emerald-600 dark:text-emerald-400">+${{ bonusCreditedAmount.toFixed(2) }}</dd>
                      </div>
                      <div v-if="showCreditedAmount" class="flex items-center justify-between gap-4 border-t border-line pt-3 dark:border-line">
                        <dt class="font-medium text-ink dark:text-ink-muted">{{ t('payment.creditedBalance') }}</dt>
                        <dd class="text-lg font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">${{ creditedAmount.toFixed(2) }}</dd>
                      </div>
                    </dl>

                    <p v-if="balanceRechargeMultiplier !== 1" class="text-xs leading-5 text-ink-muted dark:text-ink-muted">
                      {{ t('payment.rechargeRatePreview', { currency: selectedCurrency, usd: balanceRechargeMultiplier.toFixed(2) }) }}
                    </p>
                  </div>
                  <div v-else class="flex min-h-40 flex-col items-center justify-center text-center text-ink-muted dark:text-ink-muted">
                    <Icon name="calculator" size="xl" />
                    <p class="mt-3 text-sm">{{ t('payment.enterAmount') }}</p>
                  </div>

                  <button class="btn btn-primary mt-5 w-full py-3 text-base font-medium" :disabled="!canSubmit || submitting" @click="handleSubmitRecharge">
                    <span v-if="submitting" class="flex items-center justify-center gap-2">
                      <span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
                      {{ t('common.processing') }}
                    </span>
                    <span v-else class="flex items-center gap-2">
                      {{ t('payment.createOrder') }} {{ formatSelectedPaymentAmount(totalAmount) }}
                      <Icon name="arrowRight" size="sm" />
                    </span>
                  </button>
                </div>
              </aside>
            </div>
          </template>
          <!-- Subscribe Tab -->
          <template v-else-if="activeTab === 'subscription'">
            <template v-if="selectedPlan">
              <div data-test="subscription-checkout-layout" class="grid overflow-hidden rounded-lg border border-line bg-white shadow-sm lg:grid-cols-[minmax(0,1fr)_340px] dark:border-line dark:bg-surface/60">
                <section class="min-w-0">
                  <div class="p-5 sm:p-6">
                    <button type="button" class="mb-5 inline-flex items-center gap-1.5 text-sm font-medium text-ink-muted hover:text-ink-strong dark:text-ink-muted dark:hover:text-white" @click="selectedPlan = null">
                      <Icon name="arrowLeft" size="sm" />
                      <span>{{ t('common.back') }}</span>
                    </button>
                    <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                      <div class="min-w-0">
                        <span :class="['inline-flex rounded-md border px-2 py-0.5 text-xs font-medium', planBadgeClass]">
                          {{ platformLabel(selectedPlan.group_platform || '') }}
                        </span>
                        <h2 class="mt-2 break-words text-xl font-semibold text-gray-950 dark:text-white">{{ selectedPlan.name }}</h2>
                        <p v-if="selectedPlan.description" class="mt-2 max-w-2xl text-sm leading-6 text-ink-muted dark:text-ink-muted">
                          {{ selectedPlan.description }}
                        </p>
                      </div>
                      <div class="shrink-0 sm:text-right">
                        <span v-if="selectedPlan.original_price" class="block text-sm text-ink-muted line-through dark:text-ink-muted">
                          {{ formatSelectedSubscriptionPaymentAmount(selectedPlan.original_price) }}
                        </span>
                        <span class="text-3xl font-semibold tabular-nums text-gray-950 dark:text-white">{{ formatSelectedSubscriptionPaymentAmount(selectedPlan.price) }}</span>
                        <span class="ml-1 text-sm text-ink-muted dark:text-ink-muted">/ {{ planValiditySuffix }}</span>
                      </div>
                    </div>

                    <dl class="mt-6 grid grid-cols-2 gap-x-6 gap-y-4 border-t border-line pt-5 sm:grid-cols-3 dark:border-line">
                      <div>
                        <dt class="text-xs text-ink-muted dark:text-ink-muted">{{ t('payment.planCard.rate') }}</dt>
                        <dd class="mt-1 text-lg font-semibold text-primary-700 dark:text-primary-300">×{{ selectedPlan.rate_multiplier ?? 1 }}</dd>
                      </div>
                      <div v-if="planHasPeakRate(selectedPlan)" class="col-span-2 sm:col-span-1">
                        <dt class="text-xs text-ink-muted dark:text-ink-muted">{{ t('payment.planCard.peakRate') }}</dt>
                        <dd class="mt-1 text-sm font-semibold text-amber-700 dark:text-amber-300">{{ planPeakRateLabel(selectedPlan) }}</dd>
                      </div>
                      <div v-if="selectedPlan.daily_limit_usd != null">
                        <dt class="text-xs text-ink-muted dark:text-ink-muted">{{ t('payment.planCard.dailyLimit') }}</dt>
                        <dd class="mt-1 text-lg font-semibold text-ink-strong dark:text-white">${{ selectedPlan.daily_limit_usd }}</dd>
                      </div>
                      <div v-if="selectedPlan.weekly_limit_usd != null">
                        <dt class="text-xs text-ink-muted dark:text-ink-muted">{{ t('payment.planCard.weeklyLimit') }}</dt>
                        <dd class="mt-1 text-lg font-semibold text-ink-strong dark:text-white">${{ selectedPlan.weekly_limit_usd }}</dd>
                      </div>
                      <div v-if="selectedPlan.monthly_limit_usd != null">
                        <dt class="text-xs text-ink-muted dark:text-ink-muted">{{ t('payment.planCard.monthlyLimit') }}</dt>
                        <dd class="mt-1 text-lg font-semibold text-ink-strong dark:text-white">${{ selectedPlan.monthly_limit_usd }}</dd>
                      </div>
                      <div v-if="selectedPlan.daily_limit_usd == null && selectedPlan.weekly_limit_usd == null && selectedPlan.monthly_limit_usd == null">
                        <dt class="text-xs text-ink-muted dark:text-ink-muted">{{ t('payment.planCard.quota') }}</dt>
                        <dd class="mt-1 text-lg font-semibold text-ink-strong dark:text-white">{{ t('payment.planCard.unlimited') }}</dd>
                      </div>
                    </dl>
                  </div>
                  <div v-if="enabledMethods.length >= 1" class="border-t border-line p-5 sm:p-6 dark:border-line">
                    <PaymentMethodSelector
                      :methods="subMethodOptions"
                      :selected="selectedMethod"
                      @select="selectedMethod = $event"
                    />
                  </div>
                </section>

                <aside data-test="subscription-summary" class="border-t border-line bg-surface-muted/80 lg:border-l lg:border-t-0 dark:border-line dark:bg-canvas/45">
                  <div class="px-5 pb-3 pt-5">
                    <div class="flex items-center gap-2">
                      <Icon name="clipboard" size="sm" class="text-primary-600 dark:text-primary-400" />
                      <h2 class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.checkoutSummary') }}</h2>
                    </div>
                  </div>
                  <div class="px-5 pb-5">
                    <p class="truncate text-sm font-medium text-ink-strong dark:text-white">{{ selectedPlan.name }}</p>
                    <div class="mt-5">
                      <p class="text-xs font-medium text-ink-muted dark:text-ink-muted">{{ t('payment.actualPay') }}</p>
                      <p class="mt-1 text-3xl font-semibold tabular-nums text-gray-950 dark:text-white">{{ formatSelectedPaymentAmount(subTotalAmount) }}</p>
                    </div>
                    <dl class="mt-5 space-y-3 border-t border-dashed border-line-strong pt-4 text-sm dark:border-line-strong">
                      <div class="flex justify-between gap-4">
                        <dt class="text-ink-muted dark:text-ink-muted">{{ t('payment.amountLabel') }}</dt>
                        <dd class="font-medium tabular-nums text-ink-strong dark:text-white">{{ formatSelectedPaymentAmount(subPaymentAmount) }}</dd>
                      </div>
                      <div v-if="subscriptionFeeRate > 0 && selectedPlan.price > 0" class="flex justify-between gap-4">
                        <dt class="text-ink-muted dark:text-ink-muted">{{ t('payment.fee') }} ({{ subscriptionFeeRate }}%)</dt>
                        <dd class="font-medium tabular-nums text-ink-strong dark:text-white">{{ formatSelectedPaymentAmount(subFeeAmount) }}</dd>
                      </div>
                    </dl>
                    <button class="btn btn-primary mt-6 w-full py-3 text-base font-medium" :disabled="!canSubmitSubscription || submitting" @click="confirmSubscribe">
                      <span v-if="submitting" class="flex items-center justify-center gap-2">
                        <span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
                        {{ t('common.processing') }}
                      </span>
                      <span v-else class="flex items-center gap-2">
                        {{ t('payment.createOrder') }} {{ formatSelectedPaymentAmount(subTotalAmount) }}
                        <Icon name="arrowRight" size="sm" />
                      </span>
                    </button>
                  </div>
                </aside>
              </div>
            </template>
            <!-- Plan list -->
            <template v-else>
              <div class="flex items-center justify-between gap-4">
                <h2 class="text-base font-semibold text-gray-950 dark:text-white">{{ t('payment.selectPlan') }}</h2>
                <span v-if="activeSubscriptions.length > 0" class="text-xs text-ink-muted dark:text-ink-muted">{{ t('payment.activeSubscription') }} · {{ activeSubscriptions.length }}</span>
              </div>
              <div v-if="checkout.plans.length === 0" class="rounded-lg border border-dashed border-line-strong py-16 text-center dark:border-line-strong">
                <Icon name="gift" size="xl" class="mx-auto mb-3 text-ink-muted dark:text-dark-600" />
                <p class="text-ink-muted dark:text-ink-muted">{{ t('payment.noPlans') }}</p>
              </div>
              <div v-else :class="planGridClass">
                <SubscriptionPlanCard v-for="plan in checkout.plans" :key="plan.id" :plan="plan" :active-subscriptions="activeSubscriptions" @select="selectPlan" />
              </div>
              <section v-if="activeSubscriptions.length > 0" class="border-t border-line pt-5 dark:border-line">
                <h2 class="mb-3 text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.activeSubscription') }}</h2>
                <div class="divide-y divide-line rounded-lg border border-line bg-white px-4 dark:divide-line dark:border-line dark:bg-surface/60">
                  <div v-for="sub in activeSubscriptions" :key="sub.id"
                    class="flex items-center gap-3 py-3">
                    <div :class="['h-8 w-1 shrink-0 rounded-full', platformAccentBarClass(sub.group?.platform || '')]" />
                    <div class="min-w-0 flex-1">
                      <div class="flex items-center gap-1.5">
                        <span class="truncate text-xs font-semibold text-ink-strong dark:text-white">{{ sub.group?.name || t('payment.groupFallback', { id: sub.group_id }) }}</span>
                        <span :class="['shrink-0 rounded-full px-1.5 py-0.5 text-[9px] font-medium', platformBadgeLightClass(sub.group?.platform || '')]">{{ platformLabel(sub.group?.platform || '') }}</span>
                      </div>
                      <div class="flex flex-wrap gap-x-3 text-[11px] text-ink-muted dark:text-ink-muted">
                        <span>{{ t('payment.planCard.rate') }}: ×{{ sub.group?.rate_multiplier ?? 1 }}</span>
                        <span v-if="subscriptionHasPeakRate(sub)">{{ t('payment.planCard.peakRate') }}: {{ subscriptionPeakRateLabel(sub) }}</span>
                        <span v-if="sub.group?.daily_limit_usd == null && sub.group?.weekly_limit_usd == null && sub.group?.monthly_limit_usd == null">{{ t('payment.planCard.quota') }}: {{ t('payment.planCard.unlimited') }}</span>
                        <span v-if="subscriptionExpirationEnabled && sub.expires_at">{{ t('userSubscriptions.daysRemaining', { days: getDaysRemaining(sub.expires_at) }) }}</span>
                        <span v-else>{{ t('userSubscriptions.noExpiration') }}</span>
                      </div>
                    </div>
                    <span class="badge badge-success shrink-0 text-[10px]">{{ t('userSubscriptions.status.active') }}</span>
                  </div>
                </div>
              </section>
            </template>
          </template>
        </template>
        <div v-if="(checkout.help_text || checkout.help_image_url) && paymentPhase === 'select' && !selectedPlan" class="border-t border-line pt-5 dark:border-line">
          <div class="flex flex-col items-center gap-3">
            <img v-if="checkout.help_image_url" :src="checkout.help_image_url" alt=""
              class="h-40 max-w-full cursor-pointer rounded-lg object-contain transition-opacity hover:opacity-80"
              @click="previewImage = checkout.help_image_url" />
            <div v-if="checkout.help_text" class="markdown-body w-full overflow-x-auto break-words" v-html="renderedHelpText"></div>
          </div>
        </div>
      </template>
    </div>
    <!-- Renewal Plan Selection Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showRenewalModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4" @click.self="closeRenewalModal">
          <div class="relative w-full max-w-lg rounded-lg border border-line bg-white p-6 shadow-2xl dark:border-line dark:bg-canvas">
            <button class="absolute right-4 top-4 rounded-md p-1.5 text-ink-muted transition-colors hover:bg-surface-muted hover:text-ink dark:hover:bg-dark-700 dark:hover:text-gray-200" :title="t('common.close')" @click="closeRenewalModal">
              <Icon name="x" size="md" />
            </button>
            <h3 class="mb-4 text-lg font-semibold text-ink-strong dark:text-white">{{ t('payment.selectPlan') }}</h3>
            <div class="space-y-4">
              <SubscriptionPlanCard v-for="plan in renewalPlans" :key="plan.id" :plan="plan" :active-subscriptions="activeSubscriptions" @select="selectPlanFromModal" />
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
    <!-- Image Preview Overlay -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="previewImage" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 backdrop-blur-sm" @click="previewImage = ''">
          <img :src="previewImage" alt="" class="max-h-[85vh] max-w-[90vw] rounded-lg object-contain shadow-2xl" />
        </div>
      </Transition>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onBeforeUnmount, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import '@/styles/announcement-markdown.css'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePaymentStore } from '@/stores/payment'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractApiErrorMessage, extractI18nErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import { hasPeakRate, formatPeakRateWindow, serverTimezoneLabel, type PeakRateFields } from '@/utils/peak-rate'
import type { SubscriptionPlan, CheckoutInfoResponse, CreateOrderResult, OrderType } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import AmountInput from '@/components/payment/AmountInput.vue'
import PaymentMethodSelector from '@/components/payment/PaymentMethodSelector.vue'
import { METHOD_ORDER, getPaymentPopupFeaturesForMethod } from '@/components/payment/providerConfig'
import {
  calculateRechargeBonusCredit,
  calculateRechargeCreditedAmount,
  formatRechargeBonusPercent,
  normalizeRechargeBonusTiers,
  resolveRechargeBonusPercent,
} from '@/components/payment/rechargeBonus'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  buildCreateOrderPayload,
  clearPaymentRecoverySnapshot,
  decidePaymentLaunch,
  getVisibleMethods,
  normalizeVisibleMethod,
  readPaymentRecoverySnapshot,
  shouldPreopenPaymentPopup,
  type PaymentRecoverySnapshot,
  writePaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import { platformAccentBarClass, platformBadgeLightClass, platformBadgeClass, platformLabel } from '@/utils/platformColors'
import SubscriptionPlanCard from '@/components/payment/SubscriptionPlanCard.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import Icon from '@/components/icons/Icon.vue'
import { DEFAULT_PAYMENT_CURRENCY, formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'
import { planValiditySuffix as validitySuffixOf } from '@/components/payment/validity'
import type { PaymentMethodOption } from '@/components/payment/PaymentMethodSelector.vue'
import { buildPaymentErrorToastMessage, describePaymentScenarioError } from './paymentUx'
import { hasWechatResumeQuery, parseWechatResumeRoute, stripWechatResumeQuery } from './paymentWechatResume'

const i18n = useI18n()
const { t } = i18n
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const paymentStore = usePaymentStore()
const subscriptionStore = useSubscriptionStore()
const appStore = useAppStore()

const user = computed(() => authStore.user)
const activeSubscriptions = computed(() => subscriptionStore.activeSubscriptions)
const subscriptionExpirationEnabled = computed(
  () => appStore.cachedPublicSettings?.subscription_expiration_enabled !== false
)

function getDaysRemaining(expiresAt: string): number {
  const diff = new Date(expiresAt).getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
}

function subscriptionHasPeakRate(sub: { group?: PeakRateFields | null }): boolean {
  return hasPeakRate(sub.group)
}

function subscriptionPeakRateLabel(sub: { group?: PeakRateFields | null }): string {
  return formatPeakRateWindow(sub.group, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
}

const loading = ref(true)
const submitting = ref(false)
const errorMessage = ref('')
const errorHintMessage = ref('')
const activeTab = ref<'recharge' | 'subscription'>('recharge')
const amount = ref<number | null>(null)
const selectedMethod = ref('')
const selectedPlan = ref<SubscriptionPlan | null>(null)
const previewImage = ref('')

const paymentPhase = ref<'select' | 'paying'>('select')

interface CreateOrderOptions {
  openid?: string
  wechatResumeToken?: string
  paymentType?: string
  isResume?: boolean
  mobileQrFallbackAttempted?: boolean
}

interface WeixinJSBridgeLike {
  invoke(
    action: string,
    payload: Record<string, unknown>,
    callback: (result: Record<string, unknown>) => void,
  ): void
}

function emptyPaymentState(): PaymentRecoverySnapshot {
  return {
    orderId: 0,
    amount: 0,
    qrCode: '',
    expiresAt: '',
    paymentType: '',
    payUrl: '',
    outTradeNo: '',
    clientSecret: '',
    intentId: '',
    currency: '',
    countryCode: '',
    paymentEnv: '',
    payAmount: 0,
    orderType: '',
    paymentMode: '',
    resumeToken: '',
    alipayMobilePrecreateDeepLink: false,
    createdAt: 0,
  }
}

function getWeixinJSBridge(): WeixinJSBridgeLike | undefined {
  return (window as Window & { WeixinJSBridge?: WeixinJSBridgeLike }).WeixinJSBridge
}

function waitForWeixinJSBridge(timeoutMs = 4000): Promise<WeixinJSBridgeLike | null> {
  const existing = getWeixinJSBridge()
  if (existing) return Promise.resolve(existing)

  return new Promise((resolve) => {
    let settled = false
    const finish = (bridge: WeixinJSBridgeLike | null) => {
      if (settled) return
      settled = true
      document.removeEventListener('WeixinJSBridgeReady', handleReady)
      document.removeEventListener('onWeixinJSBridgeReady', handleReady)
      window.clearTimeout(timer)
      resolve(bridge)
    }
    const handleReady = () => finish(getWeixinJSBridge() ?? null)
    const timer = window.setTimeout(() => finish(getWeixinJSBridge() ?? null), timeoutMs)
    document.addEventListener('WeixinJSBridgeReady', handleReady, false)
    document.addEventListener('onWeixinJSBridgeReady', handleReady, false)
  })
}

async function invokeWechatJsapiPayment(payload: Record<string, unknown>): Promise<Record<string, unknown>> {
  const bridge = await waitForWeixinJSBridge()
  if (!bridge) {
    throw new Error('WECHAT_JSAPI_UNAVAILABLE')
  }
  return new Promise((resolve) => {
    bridge.invoke('getBrandWCPayRequest', payload, (result) => resolve(result || {}))
  })
}

const paymentState = ref<PaymentRecoverySnapshot>(emptyPaymentState())
let disposePendingPaymentPopup: (() => void) | null = null

onBeforeUnmount(() => {
  disposePendingPaymentPopup?.()
  disposePendingPaymentPopup = null
})

function persistRecoverySnapshot(snapshot: PaymentRecoverySnapshot) {
  if (typeof window === 'undefined' || !snapshot.orderId) return
  writePaymentRecoverySnapshot(window.localStorage, snapshot, PAYMENT_RECOVERY_STORAGE_KEY)
}

function removeRecoverySnapshot() {
  if (typeof window === 'undefined') return
  clearPaymentRecoverySnapshot(window.localStorage, PAYMENT_RECOVERY_STORAGE_KEY)
}

function resetPayment() {
  paymentPhase.value = 'select'
  paymentState.value = emptyPaymentState()
  removeRecoverySnapshot()
}

async function redirectToPaymentResult(state: PaymentRecoverySnapshot): Promise<void> {
  const query: Record<string, string | undefined> = {}
  if (state.orderId > 0) {
    query.order_id = String(state.orderId)
  }
  if (state.outTradeNo) {
    query.out_trade_no = state.outTradeNo
  }
  if (state.resumeToken) {
    query.resume_token = state.resumeToken
  }
  await router.push({
    path: '/payment/result',
    query,
  })
}

function buildWechatOAuthAuthorizeUrl(
  authorizeUrl: string,
  context: { paymentType: string; orderType: OrderType; planId?: number; orderAmount: number },
): string {
  const normalizedUrl = authorizeUrl.trim()
  if (!normalizedUrl || typeof window === 'undefined') {
    return normalizedUrl
  }

  try {
    const targetUrl = new URL(normalizedUrl, window.location.origin)
    const redirectPath = targetUrl.searchParams.get('redirect') || '/purchase'
    const redirectUrl = new URL(redirectPath, window.location.origin)
    const paymentType = normalizeVisibleMethod(context.paymentType) || context.paymentType.trim() || 'wxpay'

    redirectUrl.searchParams.set('payment_type', paymentType)
    redirectUrl.searchParams.set('order_type', context.orderType)

    if (context.planId) {
      redirectUrl.searchParams.set('plan_id', String(context.planId))
    } else {
      redirectUrl.searchParams.delete('plan_id')
    }

    if (context.orderAmount > 0) {
      redirectUrl.searchParams.set('amount', String(context.orderAmount))
    } else {
      redirectUrl.searchParams.delete('amount')
    }

    targetUrl.searchParams.set('redirect', `${redirectUrl.pathname}${redirectUrl.search}`)
    return targetUrl.toString()
  } catch {
    return normalizedUrl
  }
}

function onPaymentDone() {
  const wasSubscription = paymentState.value.orderType === 'subscription'
  resetPayment()
  selectedPlan.value = null
  if (wasSubscription) {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

async function onPaymentSuccess() {
  const completedPayment = { ...paymentState.value }
  removeRecoverySnapshot()
  authStore.refreshUser()
  if (paymentState.value.orderType === 'subscription') {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
  await redirectToPaymentResult(completedPayment)
}

function onPaymentSettled() {
  removeRecoverySnapshot()
}

// All checkout data from single API call
const checkout = ref<CheckoutInfoResponse>({
  methods: {}, global_min: 0, global_max: 0,
  plans: [], balance_disabled: false, balance_recharge_multiplier: 1, recharge_bonus_tiers: [], subscription_usd_to_cny_rate: 0, subscription_fee_enabled: true, recharge_fee_rate: 0, recharge_fee_credited: false, help_text: '', help_image_url: '', stripe_publishable_key: '',
})

const rechargeQuickAmounts = [10, 20, 50, 100, 200, 500]
const renderedHelpText = computed(() => DOMPurify.sanitize(
  marked.parse(checkout.value.help_text || '', { async: false, gfm: true, breaks: false }),
))

const tabs = computed(() => {
  const result: { key: 'recharge' | 'subscription'; label: string }[] = []
  if (!checkout.value.balance_disabled) result.push({ key: 'recharge', label: t('payment.tabTopUp') })
  result.push({ key: 'subscription', label: t('payment.tabSubscribe') })
  return result
})

const visibleMethods = computed(() => getVisibleMethods(checkout.value.methods))
const enabledMethods = computed(() => Object.keys(visibleMethods.value))
const validAmount = computed(() => amount.value ?? 0)
const balanceRechargeMultiplier = computed(() => {
  const multiplier = checkout.value.balance_recharge_multiplier
  return Number.isFinite(multiplier) && multiplier > 0 ? multiplier : 1
})
const rechargeFeeCredited = computed(() => checkout.value.recharge_fee_credited === true)
const rechargeBonusTiers = computed(() => normalizeRechargeBonusTiers(checkout.value.recharge_bonus_tiers))
const activeRechargeBonusPercent = computed(() => resolveRechargeBonusPercent(validAmount.value, rechargeBonusTiers.value))
const quickAmountBadges = computed<Record<number, string>>(() => {
  const badges: Record<number, string> = {}
  for (const quickAmount of rechargeQuickAmounts) {
    const percent = resolveRechargeBonusPercent(quickAmount, rechargeBonusTiers.value)
    if (percent > 0) {
      badges[quickAmount] = t('payment.bonusBadge', { percent: formatRechargeBonusPercent(percent) })
    }
  }
  return badges
})
// 订阅 CNY 换算汇率（1 USD = X CNY）。0 = 未配置，订阅保持 price 直付（与后端 opt-in 条件严格镜像）。
const subscriptionUsdToCnyRate = computed(() => {
  const rate = checkout.value.subscription_usd_to_cny_rate
  return Number.isFinite(rate) && rate > 0 ? rate : 0
})
// Adaptive grid: center single card, 2-col for 2 plans, 3-col for 3+
const planGridClass = computed(() => {
  const n = checkout.value.plans.length
  if (n <= 2) return 'grid grid-cols-1 gap-5 sm:grid-cols-2'
  return 'grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3'
})

// Check if an amount fits a method's [min, max]. 0 = no limit.
function amountFitsMethod(amt: number, methodType: string): boolean {
  if (amt <= 0) return true
  const ml = visibleMethods.value[methodType]
  if (!ml) return false
  if (ml.single_min > 0 && amt < ml.single_min) return false
  if (ml.single_max > 0 && amt > ml.single_max) return false
  return true
}

// Visible methods decide the amount range shown to users.
const globalMinAmount = computed(() => {
  const limits = Object.values(visibleMethods.value)
  if (limits.length === 0) return 0
  if (limits.some(limit => limit.single_min <= 0)) return 0
  return Math.min(...limits.map(limit => limit.single_min))
})
const globalMaxAmount = computed(() => {
  const limits = Object.values(visibleMethods.value)
  if (limits.length === 0) return 0
  if (limits.some(limit => limit.single_max <= 0)) return 0
  return Math.max(...limits.map(limit => limit.single_max))
})

// Selected method's limits (for validation and error messages)
const selectedLimit = computed(() => visibleMethods.value[selectedMethod.value])
const selectedCurrency = computed(() => normalizePaymentCurrency(selectedLimit.value?.currency))
const localeCode = computed(() => {
  const raw = i18n.locale as unknown
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object' && 'value' in raw) {
    return String((raw as { value?: string }).value || '')
  }
  return undefined
})

function currencyFractionDigits(currency: string): number {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
    }).resolvedOptions().maximumFractionDigits ?? 2
  } catch {
    return 2
  }
}

function roundPaymentAmount(value: number, currency: string): number {
  if (!Number.isFinite(value)) return 0
  const factor = 10 ** currencyFractionDigits(currency)
  return Math.round(value * factor) / factor
}

function ceilPaymentAmount(value: number, currency: string): number {
  if (!Number.isFinite(value)) return 0
  const factor = 10 ** currencyFractionDigits(currency)
  return Math.ceil(value * factor) / factor
}

function subscriptionPaymentAmountForCurrency(value: number, currency: string): number {
  const rate = subscriptionUsdToCnyRate.value
  if (rate <= 0 || currency !== DEFAULT_PAYMENT_CURRENCY) return roundPaymentAmount(value, currency)
  return roundPaymentAmount(value * rate, currency)
}

function formatSelectedPaymentAmount(value: number): string {
  return formatPaymentAmount(value, selectedCurrency.value, localeCode.value)
}

function formatSelectedSubscriptionPaymentAmount(value: number): string {
  return formatSelectedPaymentAmount(subscriptionPaymentAmountForCurrency(value, selectedCurrency.value))
}

const methodOptions = computed<PaymentMethodOption[]>(() =>
  enabledMethods.value.map((type) => {
    const ml = visibleMethods.value[type]
    return {
      type,
      display_name: ml?.display_name,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(validAmount.value, type),
    }
  })
)

const feeRate = computed(() => checkout.value?.recharge_fee_rate ?? 0)
const subscriptionFeeRate = computed(() =>
  checkout.value.subscription_fee_enabled !== false ? feeRate.value : 0
)
const feeAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.ceil(((validAmount.value * feeRate.value) / 100) * 100) / 100
    : 0
)
const totalAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.round((validAmount.value + feeAmount.value) * 100) / 100
    : validAmount.value
)
const creditedAmount = computed(() => {
  return calculateRechargeCreditedAmount(
    validAmount.value,
    totalAmount.value,
    balanceRechargeMultiplier.value,
    rechargeFeeCredited.value,
    rechargeBonusTiers.value,
  )
})
const bonusCreditedAmount = computed(() => calculateRechargeBonusCredit(
  validAmount.value,
  balanceRechargeMultiplier.value,
  rechargeBonusTiers.value,
))
const showCreditedAmount = computed(() =>
  balanceRechargeMultiplier.value !== 1
    || (rechargeFeeCredited.value && feeRate.value > 0)
    || activeRechargeBonusPercent.value > 0
)

const amountError = computed(() => {
  if (validAmount.value <= 0) return ''
  // No method can handle this amount
  if (!enabledMethods.value.some((m) => amountFitsMethod(validAmount.value, m))) {
    return t('payment.amountNoMethod')
  }
  // Selected method can't handle this amount (but others can)
  const ml = selectedLimit.value
  if (ml) {
    if (ml.single_min > 0 && validAmount.value < ml.single_min) return t('payment.amountTooLow', { min: formatSelectedPaymentAmount(ml.single_min) })
    if (ml.single_max > 0 && validAmount.value > ml.single_max) return t('payment.amountTooHigh', { max: formatSelectedPaymentAmount(ml.single_max) })
  }
  return ''
})

const canSubmit = computed(() =>
  validAmount.value > 0
    && amountFitsMethod(validAmount.value, selectedMethod.value)
    && selectedLimit.value?.available !== false
)

const subPaymentAmount = computed(() => {
  const price = selectedPlan.value?.price ?? 0
  return subscriptionPaymentAmountForCurrency(price, selectedCurrency.value)
})

const subFeeAmount = computed(() => {
  if (subscriptionFeeRate.value <= 0 || subPaymentAmount.value <= 0) return 0
  return ceilPaymentAmount((subPaymentAmount.value * subscriptionFeeRate.value) / 100, selectedCurrency.value)
})

const subTotalAmount = computed(() => {
  if (subscriptionFeeRate.value <= 0 || subPaymentAmount.value <= 0) return subPaymentAmount.value
  return roundPaymentAmount(subPaymentAmount.value + subFeeAmount.value, selectedCurrency.value)
})

function subscriptionTotalAmountForCurrency(value: number, currency: string): number {
  const paymentAmount = subscriptionPaymentAmountForCurrency(value, currency)
  if (subscriptionFeeRate.value <= 0 || paymentAmount <= 0) return paymentAmount
  const fee = ceilPaymentAmount((paymentAmount * subscriptionFeeRate.value) / 100, currency)
  return roundPaymentAmount(paymentAmount + fee, currency)
}

// Subscription-specific: method options based on gateway pay amount
const subMethodOptions = computed<PaymentMethodOption[]>(() => {
  const price = selectedPlan.value?.price ?? 0
  return enabledMethods.value.map((type) => {
    const ml = visibleMethods.value[type]
    const currency = normalizePaymentCurrency(ml?.currency)
    return {
      type,
      display_name: ml?.display_name,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(subscriptionTotalAmountForCurrency(price, currency), type),
    }
  })
})

const canSubmitSubscription = computed(() =>
  selectedPlan.value !== null
    && amountFitsMethod(subTotalAmount.value, selectedMethod.value)
    && selectedLimit.value?.available !== false
)

// Auto-switch to first available method when current selection can't handle the amount
watch(() => [validAmount.value, selectedMethod.value] as const, ([amt, method]) => {
  if (amt <= 0 || amountFitsMethod(amt, method)) return
  const available = enabledMethods.value.find((m) => amountFitsMethod(amt, m))
  if (available) selectedMethod.value = available
})

// Subscription confirm: keep platform color scoped to its identifying badge.
const planBadgeClass = computed(() => platformBadgeClass(selectedPlan.value?.group_platform || ''))

// Renewal modal state
const showRenewalModal = ref(false)
const renewGroupId = ref<number | null>(null)
const renewalPlans = computed(() => {
  if (renewGroupId.value == null) return []
  return checkout.value.plans.filter(p => p.group_id === renewGroupId.value)
})

const planValiditySuffix = computed(() => {
  if (!selectedPlan.value) return ''
  return validitySuffixOf(selectedPlan.value, t)
})

function planHasPeakRate(plan: SubscriptionPlan): boolean {
  return hasPeakRate(plan)
}

function planPeakRateLabel(plan: SubscriptionPlan): string {
  return formatPeakRateWindow(plan, serverTimezoneLabel(appStore.cachedPublicSettings?.server_utc_offset))
}

function selectPlan(plan: SubscriptionPlan) {
  selectedPlan.value = plan
  errorMessage.value = ''
}

function selectPlanFromModal(plan: SubscriptionPlan) {
  showRenewalModal.value = false
  renewGroupId.value = null
  selectedPlan.value = plan
  errorMessage.value = ''
}

function closeRenewalModal() {
  showRenewalModal.value = false
  renewGroupId.value = null
}

async function handleSubmitRecharge() {
  if (!canSubmit.value || submitting.value) return
  await createOrder(validAmount.value, 'balance')
}

async function confirmSubscribe() {
  if (!selectedPlan.value || submitting.value) return
  await createOrder(selectedPlan.value.price, 'subscription', selectedPlan.value.id)
}

async function createOrder(orderAmount: number, orderType: OrderType, planId?: number, options: CreateOrderOptions = {}) {
  submitting.value = true
  errorMessage.value = ''
  errorHintMessage.value = ''
  const requestType = normalizeVisibleMethod(options.paymentType || selectedMethod.value) || options.paymentType || selectedMethod.value
  const mobile = isMobileDevice()
  const wechatBrowser = typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent)
  let preopenedPaymentPopup: Window | null = null
  let paymentPopupNavigated = false

  const closePreopenedPaymentPopup = () => {
    if (!preopenedPaymentPopup || paymentPopupNavigated) return
    try {
      if (!preopenedPaymentPopup.closed) preopenedPaymentPopup.close()
    } catch {
      // A browser can revoke access to a popup while the checkout request is in flight.
    }
    preopenedPaymentPopup = null
  }
  disposePendingPaymentPopup = closePreopenedPaymentPopup

  const usePreopenedPaymentPopup = (url: string): boolean => {
    const popup = preopenedPaymentPopup
    if (!popup) return false
    try {
      if (popup.closed) return false
      popup.location.href = url
      popup.focus()
      paymentPopupNavigated = true
      preopenedPaymentPopup = null
      return true
    } catch {
      return false
    }
  }

  const openWindow = (url: string) => {
    if (usePreopenedPaymentPopup(url)) return
    const win = window.open(url, 'paymentPopup', getPaymentPopupFeaturesForMethod(requestType))
    if (!win || win.closed) {
      window.location.href = url
    }
  }

  // Open the placeholder during the click event so browsers do not block the
  // later GM checkout navigation after the asynchronous order request.
  const configuredMethod = normalizeVisibleMethod(requestType) || requestType
  if (
    typeof window !== 'undefined'
    && shouldPreopenPaymentPopup(visibleMethods.value[configuredMethod]?.payment_mode, mobile, options.isResume === true)
  ) {
    try {
      preopenedPaymentPopup = window.open(
        'about:blank',
        `paymentPopup-${Date.now()}`,
        getPaymentPopupFeaturesForMethod(requestType),
      )
    } catch {
      preopenedPaymentPopup = null
    }
  }

  try {
    const payload = buildCreateOrderPayload({
      amount: orderAmount,
      paymentType: requestType,
      orderType,
      planId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: mobile,
      isWechatBrowser: wechatBrowser,
      forceQRCode: !!(checkout.value.alipay_force_qrcode && normalizeVisibleMethod(requestType) === 'alipay'),
      mobilePrecreateDeepLink: checkout.value.alipay_mobile_precreate_deep_link === true,
    })
    if (options.openid) {
      payload.openid = options.openid
    }
    if (options.wechatResumeToken) {
      payload.wechat_resume_token = options.wechatResumeToken
    }

    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
    const visibleMethod = normalizeVisibleMethod(requestType) || requestType
    // When user clicks the dedicated Stripe button, leave method blank so the
    // landing page renders Stripe's full Payment Element (card/link/alipay/wxpay).
    const stripeMethod = visibleMethod === 'stripe'
      ? ''
      : visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret && visibleMethod !== 'airwallex'
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const airwallexRouteUrl = result.client_secret && result.intent_id
      ? router.resolve({
        path: '/payment/airwallex',
        query: {
          order_id: String(result.order_id),
          out_trade_no: result.out_trade_no || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType,
      isMobile: mobile,
      isWechatBrowser: wechatBrowser,
      forceQRCode: !!(checkout.value.alipay_force_qrcode && visibleMethod === 'alipay'),
      mobilePrecreateDeepLink: checkout.value.alipay_mobile_precreate_deep_link === true,
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
      airwallexRouteUrl,
    })

    if (decision.kind === 'wechat_oauth' && decision.oauth?.authorize_url) {
      closePreopenedPaymentPopup()
      window.location.href = buildWechatOAuthAuthorizeUrl(decision.oauth.authorize_url, {
        paymentType: visibleMethod,
        orderType,
        planId,
        orderAmount,
      })
      return
    }

    if (decision.kind === 'unhandled') {
      closePreopenedPaymentPopup()
      applyScenarioError({ reason: 'UNHANDLED_PAYMENT_SCENARIO' }, visibleMethod)
      return
    }

    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)

    if (decision.kind === 'stripe_popup') {
      openWindow(decision.paymentState.payUrl)
      return
    }
    if (decision.kind === 'stripe_route') {
      closePreopenedPaymentPopup()
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'airwallex_route') {
      closePreopenedPaymentPopup()
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'wechat_jsapi' && decision.jsapi) {
      closePreopenedPaymentPopup()
      try {
        const jsapiResult = await invokeWechatJsapiPayment(decision.jsapi as Record<string, unknown>)
        const errMsg = String(jsapiResult.err_msg || '').toLowerCase()
        if (errMsg.includes('cancel')) {
          appStore.showInfo(t('payment.qr.cancelled'))
          resetPayment()
        } else if (errMsg && !errMsg.includes('ok')) {
          resetPayment()
          const fallbackApplied = await attemptMobileQrFallback(
            { reason: 'WECHAT_JSAPI_FAILED', message: errMsg },
            {
              orderAmount,
              orderType,
              planId,
              paymentType: visibleMethod,
              attempted: options.mobileQrFallbackAttempted === true,
            },
          )
          if (!fallbackApplied) {
            applyScenarioError({ reason: 'WECHAT_JSAPI_FAILED', message: errMsg }, visibleMethod)
          }
        } else {
          const resultState = { ...decision.paymentState }
          resetPayment()
          await redirectToPaymentResult(resultState)
        }
      } catch (err: unknown) {
        resetPayment()
        const fallbackApplied = await attemptMobileQrFallback(err, {
          orderAmount,
          orderType,
          planId,
          paymentType: visibleMethod,
          attempted: options.mobileQrFallbackAttempted === true,
        })
        if (!fallbackApplied) {
          throw err
        }
      }
      return
    }
    if (decision.kind === 'redirect_waiting' && decision.paymentState.payUrl) {
      if (mobile) {
        closePreopenedPaymentPopup()
        window.location.href = decision.paymentState.payUrl
        return
      }
      openWindow(decision.paymentState.payUrl)
    }
  } catch (err: unknown) {
    const apiErr = err as Record<string, unknown>
    if (apiErr.reason === 'TOO_MANY_PENDING') {
      const metadata = apiErr.metadata as Record<string, unknown> | undefined
      errorMessage.value = t('payment.errors.tooManyPending', { max: metadata?.max || '' })
      errorHintMessage.value = ''
    } else if (apiErr.reason === 'CANCEL_RATE_LIMITED') {
      errorMessage.value = t('payment.errors.cancelRateLimited')
      errorHintMessage.value = ''
    } else if (await attemptMobileQrFallback(err, {
      orderAmount,
      orderType,
      planId,
      paymentType: requestType,
      attempted: options.mobileQrFallbackAttempted === true,
    })) {
      return
    } else {
      const handled = applyScenarioError(
        err,
        normalizeVisibleMethod(options.paymentType || selectedMethod.value) || selectedMethod.value,
      )
      if (!handled) {
        errorMessage.value = extractI18nErrorMessage(err, t, 'payment.errors', extractApiErrorMessage(err, t('payment.result.failed')))
        errorHintMessage.value = ''
      }
      if (handled) {
        return
      }
    }
    appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  } finally {
    closePreopenedPaymentPopup()
    if (disposePendingPaymentPopup === closePreopenedPaymentPopup) {
      disposePendingPaymentPopup = null
    }
    submitting.value = false
  }
}

interface MobileQrFallbackContext {
  orderAmount: number
  orderType: OrderType
  planId?: number
  paymentType: string
  attempted: boolean
}

function shouldFallbackToDesktopQr(err: unknown, paymentMethod: string, attempted: boolean): boolean {
  if (attempted || !isMobileDevice()) {
    return false
  }

  const normalizedMethod = normalizeVisibleMethod(paymentMethod) || paymentMethod
  const reason = typeof err === 'object' && err && 'reason' in err && typeof err.reason === 'string'
    ? err.reason
    : ''
  const message = err instanceof Error
    ? err.message
    : (typeof err === 'object' && err && 'message' in err && typeof err.message === 'string'
      ? err.message
      : '')
  const normalizedMessage = message.toLowerCase()

  if (normalizedMethod === 'wxpay') {
    return reason === 'WECHAT_H5_NOT_AUTHORIZED'
      || reason === 'WECHAT_PAYMENT_MP_NOT_CONFIGURED'
      || reason === 'WECHAT_JSAPI_FAILED'
      || reason === 'PAYMENT_GATEWAY_ERROR'
      || reason === 'UNHANDLED_PAYMENT_SCENARIO'
      || normalizedMessage.includes('weixinjsbridge is unavailable')
      || normalizedMessage.includes('wechat_jsapi_unavailable')
  }

  if (normalizedMethod === 'alipay') {
    return reason === 'PAYMENT_GATEWAY_ERROR' || reason === 'UNHANDLED_PAYMENT_SCENARIO'
  }

  return false
}

async function attemptMobileQrFallback(err: unknown, context: MobileQrFallbackContext): Promise<boolean> {
  if (!shouldFallbackToDesktopQr(err, context.paymentType, context.attempted)) {
    return false
  }

  try {
    const visibleMethod = normalizeVisibleMethod(context.paymentType) || context.paymentType
    const payload = buildCreateOrderPayload({
      amount: context.orderAmount,
      paymentType: visibleMethod,
      orderType: context.orderType,
      planId: context.planId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: false,
      isWechatBrowser: false,
    })
    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
    const stripeMethod = visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType: context.orderType,
      isMobile: false,
      isWechatBrowser: false,
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
    })

    if (decision.kind !== 'qr_waiting' || !decision.paymentState.qrCode) {
      return false
    }

    errorMessage.value = ''
    errorHintMessage.value = ''
    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)
    appStore.showWarning(t('payment.errors.mobilePaymentFallbackToQr'))
    return true
  } catch {
    return false
  }
}

function applyScenarioError(err: unknown, paymentMethod: string): boolean {
  const descriptor = describePaymentScenarioError(err, {
    paymentMethod,
    isMobile: isMobileDevice(),
    isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
  })
  if (!descriptor) {
    errorMessage.value = ''
    errorHintMessage.value = ''
    return false
  }
  errorMessage.value = t(descriptor.messageKey)
  errorHintMessage.value = descriptor.hintKey ? t(descriptor.hintKey) : ''
  appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  return true
}

async function resumeWechatPaymentFromQuery() {
  const resume = parseWechatResumeRoute(route.query, checkout.value.plans, validAmount.value)
  if (!resume) {
    return
  }

  selectedMethod.value = resume.paymentType
  if (resume.orderType === 'balance' && resume.orderAmount > 0) {
    amount.value = resume.orderAmount
  }
  if (resume.orderType === 'subscription' && resume.planId) {
    selectedPlan.value = checkout.value.plans.find(plan => plan.id === resume.planId) ?? null
  }

  await router.replace({ path: route.path, query: stripWechatResumeQuery(route.query) })

  if (resume.wechatResumeToken) {
    await createOrder(0, resume.orderType, resume.planId, {
      wechatResumeToken: resume.wechatResumeToken,
      paymentType: resume.paymentType,
      isResume: true,
    })
    return
  }

  if (resume.orderAmount > 0 && resume.openid) {
    await createOrder(resume.orderAmount, resume.orderType, resume.planId, {
      openid: resume.openid,
      paymentType: resume.paymentType,
      isResume: true,
    })
  }
}

onMounted(async () => {
  try {
    checkout.value = (await paymentAPI.getCheckoutInfo()).data

    if (enabledMethods.value.length) {
      const order: readonly string[] = METHOD_ORDER
      const sorted = [...enabledMethods.value].sort((a, b) => {
        const ai = order.indexOf(a)
        const bi = order.indexOf(b)
        return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
      })
      selectedMethod.value = sorted[0]
    }
    if (typeof window !== 'undefined') {
      if (hasWechatResumeQuery(route.query)) {
        removeRecoverySnapshot()
      }
      const routeResumeToken = typeof route.query.resume_token === 'string'
        ? route.query.resume_token
        : typeof route.query.wechat_resume_token === 'string'
          ? route.query.wechat_resume_token
          : undefined
      const restored = readPaymentRecoverySnapshot(
        window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY),
        { resumeToken: routeResumeToken },
      )
      if (restored) {
        paymentState.value = restored
        paymentPhase.value = 'paying'
        const restoredMethod = normalizeVisibleMethod(restored.paymentType)
          || (visibleMethods.value[restored.paymentType] ? restored.paymentType : '')
        if (restoredMethod) {
          selectedMethod.value = restoredMethod
        }
      } else {
        removeRecoverySnapshot()
      }
    }
    await resumeWechatPaymentFromQuery()
    if (checkout.value.balance_disabled) {
      activeTab.value = 'subscription'
    }
    // Handle renewal navigation: ?tab=subscription&group=123
    if (route.query.tab === 'subscription') {
      activeTab.value = 'subscription'
      if (route.query.group) {
        const groupId = Number(route.query.group)
        const groupPlans = checkout.value.plans.filter(p => p.group_id === groupId)
        if (groupPlans.length === 1) {
          selectedPlan.value = groupPlans[0]
        } else if (groupPlans.length > 1) {
          renewGroupId.value = groupId
          showRenewalModal.value = true
        }
      }
    }
  } catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { loading.value = false }
  // Fetch active subscriptions (uses cache, non-blocking)
  subscriptionStore.fetchActiveSubscriptions().catch(() => {})
})

</script>
