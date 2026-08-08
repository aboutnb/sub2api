<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-5">
      <header class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-950 dark:text-white">{{ t('payment.orders.title') }}</h1>
          <p v-if="invoiceConfig.enabled" class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('payment.invoice.eligibleHint') }}</p>
        </div>
        <div class="flex items-center gap-2">
          <button class="btn btn-secondary" @click="router.push('/purchase')">
            <Icon name="creditCard" size="md" />
            <span>{{ t('payment.result.backToRecharge') }}</span>
          </button>
          <button class="btn btn-secondary px-3" :disabled="loading" :title="t('common.refresh')" @click="fetchOrders">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </header>

      <div class="flex flex-col gap-3 border-b border-gray-200 pb-4 sm:flex-row sm:items-center dark:border-dark-700">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
          <Select v-model="currentFilter" :options="statusFilters" class="w-full sm:w-40" @change="fetchOrders" />
          <div v-if="invoiceConfig.enabled" class="flex flex-1 items-center gap-2 sm:justify-end">
            <button class="btn btn-secondary flex-1 sm:flex-none" @click="openInvoiceRecords">
              <Icon name="document" size="md" />
              <span>{{ t('payment.invoice.records') }}</span>
            </button>
            <button class="btn btn-primary flex-1 sm:flex-none" @click="startInvoiceSelection()">
              <Icon name="plus" size="md" />
              <span>{{ t('payment.invoice.apply') }}</span>
            </button>
          </div>
        </div>
      </div>

      <section
        v-if="invoiceDraft && !invoiceDialogOpen"
        data-test="invoice-draft-resume"
        class="flex flex-col gap-3 border-l-2 border-amber-500 bg-amber-50/60 px-4 py-3 sm:flex-row sm:items-center sm:justify-between dark:bg-amber-950/15"
      >
        <div class="min-w-0">
          <p class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.invoice.draftPending') }}</p>
          <p class="mt-0.5 text-xs text-gray-600 dark:text-gray-300">
            {{ t('payment.invoice.draftSummary', { count: invoiceDraft.order_ids.length }) }}
          </p>
        </div>
        <div class="flex shrink-0 gap-2">
          <button v-if="!invoiceDraft.need_pay_tax" class="btn btn-secondary btn-sm" :disabled="invoiceBusy" @click="abandonInvoiceDraft">
            {{ t('payment.invoice.abandonDraft') }}
          </button>
          <button class="btn btn-primary btn-sm" :disabled="invoiceBusy" @click="resumeInvoiceDraft">
            <Icon name="arrowRight" size="sm" />
            <span>{{ t('payment.invoice.resumeDraft') }}</span>
          </button>
        </div>
      </section>

      <div v-if="invoiceSelectionMode && !invoiceDraft" data-test="invoice-selection-bar" class="sticky top-3 z-20 border-l-2 border-primary-500 bg-white p-4 shadow-md dark:bg-dark-800">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
          <div class="min-w-0 sm:w-64">
            <p class="mb-2 text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t(invoiceAllowsFeePayerChoice ? 'payment.invoice.taxMode' : 'payment.invoice.feePolicy') }}
            </p>
            <div
              v-if="invoiceAllowsFeePayerChoice"
              data-test="invoice-tax-mode"
              class="grid grid-cols-2 rounded-md bg-gray-100 p-1 dark:bg-dark-700"
            >
              <button
                type="button"
                class="rounded px-2 py-1.5 text-xs font-medium transition-colors"
                :class="!invoiceNeedPayTax ? 'bg-white text-gray-950 shadow-sm dark:bg-dark-800 dark:text-white' : 'text-gray-600 dark:text-gray-300'"
                :aria-pressed="!invoiceNeedPayTax"
                @click="invoiceNeedPayTax = false"
              >
                {{ t('payment.invoice.taxNotRequired') }}
              </button>
              <button
                type="button"
                class="rounded px-2 py-1.5 text-xs font-medium transition-colors"
                :class="invoiceNeedPayTax ? 'bg-white text-gray-950 shadow-sm dark:bg-dark-800 dark:text-white' : 'text-gray-600 dark:text-gray-300'"
                :aria-pressed="invoiceNeedPayTax"
                @click="invoiceNeedPayTax = true"
              >
                {{ t('payment.invoice.taxRequired') }}
              </button>
            </div>
            <div
              v-else
              data-test="invoice-fixed-fee-payer"
              class="inline-flex items-center gap-1.5 border-l-2 border-primary-500 pl-2 text-sm font-semibold text-gray-950 dark:text-white"
            >
              <Icon name="lock" size="xs" class="text-primary-600 dark:text-primary-400" />
              {{ t(invoiceNeedPayTax ? 'payment.invoice.taxRequired' : 'payment.invoice.taxNotRequired') }}
            </div>
            <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-gray-400">
              {{ invoiceNeedPayTax ? t('payment.invoice.userPaysTaxNotice') : t('payment.invoice.platformPaysTaxNotice') }}
            </p>
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center justify-between gap-3 text-sm font-semibold text-gray-950 dark:text-white">
              {{ t('payment.invoice.selectedCount', { count: selectedInvoiceOrderIds.size, max: invoiceConfig.max_orders }) }}
              <span class="text-xs tabular-nums text-primary-700 dark:text-primary-300">{{ Math.round((selectedInvoiceOrderIds.size / invoiceConfig.max_orders) * 100) }}%</span>
            </div>
            <div class="mt-2 h-1 overflow-hidden bg-primary-100 dark:bg-primary-900">
              <div class="h-full bg-primary-600 transition-[width]" :style="{ width: `${(selectedInvoiceOrderIds.size / invoiceConfig.max_orders) * 100}%` }" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-2 sm:flex">
            <button class="btn btn-secondary" :disabled="invoiceBusy" @click="cancelInvoiceSelection">{{ t('common.cancel') }}</button>
            <button class="btn btn-primary" :disabled="invoiceBusy || selectedInvoiceOrderIds.size === 0" @click="validateInvoiceSelection">
              <Icon v-if="invoiceBusy" name="refresh" size="md" class="animate-spin" />
              <span>{{ t('common.next') }}</span>
              <Icon v-if="!invoiceBusy" name="arrowRight" size="sm" />
            </button>
          </div>
        </div>
      </div>

      <!-- Table -->
      <OrderTable :orders="orders" :loading="loading">
        <template #actions="{ row }">
          <div class="flex flex-wrap items-center gap-2">
            <label
              v-if="invoiceConfig.enabled && invoiceSelectionMode && row.status === 'COMPLETED'"
              class="inline-flex cursor-pointer items-center gap-1.5 rounded-md px-2 py-1 text-xs font-medium text-primary-700 hover:bg-primary-50 dark:text-primary-300 dark:hover:bg-primary-950/30"
            >
              <input
                type="checkbox"
                class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                :checked="selectedInvoiceOrderIds.has(row.id)"
                :disabled="!selectedInvoiceOrderIds.has(row.id) && selectedInvoiceOrderIds.size >= invoiceConfig.max_orders"
                @change="toggleInvoiceOrder(row.id)"
              />
              <span>{{ t('payment.invoice.selectOrder') }}</span>
            </label>
            <button
              v-else-if="invoiceConfig.enabled && row.status === 'COMPLETED'"
              class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-primary-700 hover:bg-primary-50 dark:text-primary-300 dark:hover:bg-primary-950/30"
              @click="startInvoiceSelection(row.id)"
            >
              <Icon name="plus" size="sm" />
              <span>{{ t('payment.invoice.apply') }}</span>
            </button>
            <button v-if="row.status === 'PENDING'" @click="handleCancel(row.id)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-yellow-600 hover:bg-yellow-50 dark:text-yellow-400 dark:hover:bg-yellow-900/20">
              <Icon name="x" size="sm" />
              <span>{{ t('payment.orders.cancel') }}</span>
            </button>
            <button v-if="canRequestRefund(row)" @click="openRefundDialog(row)" class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700">
              <Icon name="dollar" size="sm" />
              <span>{{ t('payment.orders.requestRefund') }}</span>
            </button>
          </div>
        </template>
      </OrderTable>

      <!-- Pagination -->
      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>

    <!-- Cancel Confirm Dialog -->
    <BaseDialog :show="!!cancelTargetId" :title="t('payment.orders.cancel')" width="narrow" @close="cancelTargetId = null">
      <p class="text-sm text-gray-600 dark:text-gray-300">{{ t('payment.confirmCancel') }}</p>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="cancelTargetId = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="actionLoading" @click="confirmCancel">{{ actionLoading ? t('common.processing') : t('payment.orders.cancel') }}</button>
        </div>
      </template>
    </BaseDialog>

    <!-- Refund Dialog -->
    <BaseDialog :show="!!refundTarget" :title="t('payment.orders.requestRefund')" @close="refundTarget = null">
      <div v-if="refundTarget" class="space-y-4">
        <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-800">
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.orderId') }}</span>
            <span class="font-mono text-gray-900 dark:text-white">#{{ refundTarget.id }}</span>
          </div>
          <div class="mt-2 flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400">{{ t('payment.orders.amount') }}</span>
            <span class="text-gray-900 dark:text-white">${{ refundTarget.amount.toFixed(2) }}</span>
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('payment.refundReason') }}</label>
          <textarea v-model="refundReason" rows="3" class="input mt-1 w-full" :placeholder="t('payment.refundReasonPlaceholder')" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="refundTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="actionLoading || !refundReason.trim()" @click="confirmRefund">{{ actionLoading ? t('common.processing') : t('payment.orders.requestRefund') }}</button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="invoiceDialogOpen"
      :title="t('payment.invoice.applicationTitle')"
      width="wide"
      :close-on-escape="!invoiceBusy"
      @close="closeInvoiceDialog"
    >
      <div v-if="invoiceDraft" class="space-y-5">
        <ol data-test="invoice-progress" class="relative grid grid-cols-3 gap-2 pb-2 before:absolute before:left-[16.66%] before:right-[16.66%] before:top-3 before:h-px before:bg-gray-200 dark:before:bg-dark-600">
          <li class="relative z-10 flex min-w-0 flex-col items-center gap-2 text-center text-emerald-700 dark:text-emerald-300">
            <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-emerald-100 ring-4 ring-white dark:bg-emerald-900/40 dark:ring-dark-800">
              <Icon name="check" size="xs" :stroke-width="2.5" />
            </span>
            <span class="text-xs font-medium leading-4">{{ t('payment.invoice.selectOrdersStep') }}</span>
          </li>
          <li class="relative z-10 flex min-w-0 flex-col items-center gap-2 text-center" :class="invoiceReady ? 'text-emerald-700 dark:text-emerald-300' : 'text-amber-700 dark:text-amber-300'">
            <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full ring-4 ring-white dark:ring-dark-800" :class="invoiceReady ? 'bg-emerald-100 dark:bg-emerald-900/40' : 'bg-amber-100 dark:bg-amber-900/40'">
              <Icon :name="invoiceReady ? 'check' : 'creditCard'" size="xs" :stroke-width="2.5" />
            </span>
            <span class="text-xs font-medium leading-4">{{ t('payment.invoice.taxReviewStep') }}</span>
          </li>
          <li class="relative z-10 flex min-w-0 flex-col items-center gap-2 text-center" :class="invoiceReady ? 'text-primary-700 dark:text-primary-300' : 'text-gray-400 dark:text-gray-500'">
            <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full ring-4 ring-white dark:ring-dark-800" :class="invoiceReady ? 'bg-primary-100 dark:bg-primary-900/40' : 'bg-gray-100 dark:bg-dark-700'">
              <Icon name="document" size="xs" :stroke-width="2" />
            </span>
            <span class="text-xs font-medium leading-4">{{ t('payment.invoice.buyerInfoStep') }}</span>
          </li>
        </ol>

        <div data-test="invoice-amount-summary" class="grid grid-cols-2 gap-x-4 gap-y-3 border-y border-gray-200 py-4 sm:grid-cols-4 dark:border-dark-700">
          <div>
            <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.invoice.orderCount') }}</div>
            <div class="mt-1 text-lg font-semibold tabular-nums text-gray-950 dark:text-white">{{ invoiceDraft.order_ids.length }}</div>
          </div>
          <div>
            <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.invoice.orderAmount') }}</div>
            <div class="mt-1 text-lg font-semibold tabular-nums text-gray-950 dark:text-white">{{ invoiceCurrency }} {{ invoiceValidation.totalAmount || '--' }}</div>
          </div>
          <div v-if="invoiceDraft.need_pay_tax">
            <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.invoice.taxAmount') }}</div>
            <div class="mt-1 text-lg font-semibold tabular-nums text-gray-950 dark:text-white">{{ invoiceCurrency }} {{ invoiceValidation.taxAmount || '--' }}</div>
          </div>
          <div>
            <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.invoice.invoiceAmount') }}</div>
            <div data-test="invoice-final-amount" class="mt-1 text-lg font-semibold tabular-nums text-primary-700 dark:text-primary-300">
              {{ invoiceCurrency }} {{ invoiceValidation.invoiceAmount || invoiceValidation.totalAmount || '--' }}
            </div>
          </div>
        </div>

        <section
          v-if="invoiceDraft.need_pay_tax && !invoiceReady"
          data-test="invoice-tax-step"
          class="overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-850"
        >
          <div class="grid md:grid-cols-[minmax(0,1fr)_220px]">
            <div class="p-4 sm:p-5">
              <div class="flex items-start gap-3">
                <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                  <Icon name="creditCard" size="md" :stroke-width="1.8" />
                </span>
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <h4 class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.invoice.payTaxTitle') }}</h4>
                    <span class="rounded-full bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-950/40 dark:text-amber-300">
                      {{ t('payment.invoice.taxPending') }}
                    </span>
                  </div>
                  <p class="mt-1 text-sm leading-6 text-gray-500 dark:text-gray-400">
                    {{ t('payment.invoice.payTaxNotice', {
                      fee: `${invoiceCurrency} ${invoiceValidation.taxAmount || '--'}`,
                    }) }}
                  </p>
                </div>
              </div>

              <div class="mt-5">
                <p class="mb-2 text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('payment.paymentMethod') }}</p>
                <div class="grid gap-2 sm:grid-cols-2">
                  <button
                    v-for="paymentOption in invoiceTaxPayments"
                    :key="paymentOption.taxOrderNo"
                    type="button"
                    data-test="invoice-tax-payment-option"
                    :aria-pressed="selectedTaxOrderNo === paymentOption.taxOrderNo"
                    class="group flex h-14 min-w-0 items-center gap-3 rounded-md border px-3 text-left transition-colors"
                    :class="selectedTaxOrderNo === paymentOption.taxOrderNo
                      ? 'border-primary-500 bg-primary-50/50 ring-1 ring-primary-500/20 dark:bg-primary-950/20'
                      : 'border-gray-200 bg-transparent hover:border-gray-400 dark:border-dark-600 dark:hover:border-dark-500'"
                    @click="openInvoiceTaxPayment(paymentOption)"
                  >
                    <img :src="invoicePaymentMethodIcon(paymentOption.channel)" alt="" class="h-7 w-7 shrink-0 object-contain" />
                    <span class="min-w-0 flex-1">
                      <span class="block truncate text-sm font-semibold text-gray-900 dark:text-white">{{ taxChannelLabel(paymentOption.channel) }}</span>
                      <span class="mt-0.5 block text-xs text-gray-500 dark:text-gray-400">{{ t('payment.invoice.openCashier') }}</span>
                    </span>
                    <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-gray-400 transition-colors group-hover:bg-gray-100 group-hover:text-gray-700 dark:group-hover:bg-dark-700 dark:group-hover:text-gray-200">
                      <Icon name="externalLink" size="sm" />
                    </span>
                  </button>
                </div>
              </div>
            </div>

            <div class="flex flex-col justify-between border-t border-gray-200 bg-gray-50/70 p-4 md:border-l md:border-t-0 sm:p-5 dark:border-dark-700 dark:bg-dark-900/40">
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('payment.invoice.taxDue') }}</p>
                <p class="mt-1 flex items-baseline gap-1.5 text-gray-950 dark:text-white">
                  <span class="text-xs font-semibold">{{ invoiceCurrency }}</span>
                  <span class="text-3xl font-semibold tabular-nums">{{ invoiceValidation.taxDueAmount || '--' }}</span>
                </p>
                <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-gray-400">{{ t('payment.invoice.returnToConfirm') }}</p>
              </div>
              <button
                type="button"
                class="btn mt-5 w-full justify-center"
                :class="selectedTaxOrderNo
                  ? 'btn-primary'
                  : 'cursor-not-allowed border-gray-200 bg-gray-200 text-gray-500 shadow-none dark:border-dark-700 dark:bg-dark-700 dark:text-gray-400'"
                :disabled="invoiceBusy || !selectedTaxOrderNo"
                @click="checkInvoiceTaxPayment"
              >
                <Icon v-if="invoiceBusy" name="refresh" size="md" class="animate-spin" />
                <Icon v-else name="checkCircle" size="md" />
                <span>{{ t('payment.invoice.checkTaxPayment') }}</span>
              </button>
            </div>
          </div>
        </section>

        <form v-else data-test="invoice-buyer-form" class="space-y-5" @submit.prevent="submitInvoiceApplication">
          <section class="grid gap-5 md:grid-cols-[180px_minmax(0,1fr)]">
            <div>
              <h4 class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.invoice.buyerType') }}</h4>
            </div>
            <div class="grid grid-cols-2 rounded-lg bg-gray-100 p-1 dark:bg-dark-800">
              <button
                v-for="type in invoiceBuyerTypes"
                :key="type.value"
                type="button"
                :aria-pressed="invoiceForm.buyer_type === type.value"
                class="rounded-md px-3 py-2 text-sm font-medium transition-colors"
                :class="invoiceForm.buyer_type === type.value ? 'bg-white text-gray-950 shadow-sm dark:bg-dark-700 dark:text-white' : 'text-gray-600 dark:text-gray-300'"
                @click="invoiceForm.buyer_type = type.value"
              >
                {{ type.label }}
              </button>
            </div>
          </section>

          <section class="grid gap-5 border-t border-gray-200 pt-5 md:grid-cols-[180px_minmax(0,1fr)] dark:border-dark-700">
            <div>
              <h4 class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.invoice.requiredInfo') }}</h4>
            </div>
            <div class="grid gap-4 sm:grid-cols-2">
              <div>
                <label class="input-label" for="invoice-title">{{ t('payment.invoice.invoiceTitle') }}</label>
                <input id="invoice-title" v-model.trim="invoiceForm.title" class="input mt-1 w-full" maxlength="255" required />
              </div>
              <div>
                <label class="input-label" for="invoice-taxpayer-id">{{ t('payment.invoice.taxpayerId') }}</label>
                <input id="invoice-taxpayer-id" v-model.trim="invoiceForm.taxpayer_id" class="input mt-1 w-full" maxlength="32" required />
              </div>
              <div class="sm:col-span-2">
                <label class="input-label" for="invoice-email">{{ t('payment.invoice.recipientEmail') }}</label>
                <input id="invoice-email" v-model.trim="invoiceForm.recipient_email" type="email" class="input mt-1 w-full" maxlength="255" required />
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('payment.invoice.emailNotice') }}</p>
              </div>
            </div>
          </section>

          <section class="grid gap-5 border-t border-gray-200 pt-5 md:grid-cols-[180px_minmax(0,1fr)] dark:border-dark-700">
            <div>
              <h4 class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('payment.invoice.optionalInfo') }}</h4>
            </div>
            <div class="grid gap-4 sm:grid-cols-2">
              <div>
                <label class="input-label" for="invoice-address">{{ t('payment.invoice.buyerAddress') }}</label>
                <input id="invoice-address" v-model.trim="invoiceForm.buyer_address" class="input mt-1 w-full" maxlength="255" />
              </div>
              <div>
                <label class="input-label" for="invoice-phone">{{ t('payment.invoice.buyerPhone') }}</label>
                <input id="invoice-phone" v-model.trim="invoiceForm.buyer_phone" class="input mt-1 w-full" />
              </div>
              <div>
                <label class="input-label" for="invoice-bank">{{ t('payment.invoice.buyerBank') }}</label>
                <input id="invoice-bank" v-model.trim="invoiceForm.buyer_bank" class="input mt-1 w-full" maxlength="255" />
              </div>
              <div>
                <label class="input-label" for="invoice-bank-account">{{ t('payment.invoice.buyerBankAccount') }}</label>
                <input id="invoice-bank-account" v-model.trim="invoiceForm.buyer_bank_account" inputmode="numeric" class="input mt-1 w-full" maxlength="40" />
              </div>
            </div>
          </section>

          <div class="flex flex-col-reverse gap-2 border-t border-gray-200 pt-4 sm:flex-row sm:justify-end dark:border-dark-700">
            <button type="button" class="btn btn-secondary" :disabled="invoiceBusy" @click="closeInvoiceDialog">{{ t('payment.invoice.saveAndClose') }}</button>
            <button type="submit" class="btn btn-primary" :disabled="invoiceBusy || !invoiceFormValid">
              <Icon v-if="invoiceBusy" name="refresh" size="md" class="animate-spin" />
              <Icon v-else name="document" size="md" />
              <span>{{ t('payment.invoice.submitApplication') }}</span>
            </button>
          </div>
        </form>
      </div>
    </BaseDialog>

    <BaseDialog :show="invoiceRecordsOpen" :title="t('payment.invoice.recordsTitle')" width="extra-wide" @close="invoiceRecordsOpen = false">
      <div class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('payment.invoice.recordCount', { count: invoiceRecordTotal }) }}</span>
          <button class="btn btn-secondary px-3" :disabled="invoiceRecordsLoading" :title="t('common.refresh')" @click="fetchInvoiceRecords">
            <Icon name="refresh" size="md" :class="invoiceRecordsLoading ? 'animate-spin' : ''" />
          </button>
        </div>
        <div v-if="invoiceRecordsLoading && invoiceRecords.length === 0" class="rounded-lg border border-dashed border-gray-300 py-12 text-center text-sm text-gray-500 dark:border-dark-600">{{ t('common.loading') }}</div>
        <div v-else-if="invoiceRecords.length === 0" class="rounded-lg border border-dashed border-gray-300 py-12 text-center dark:border-dark-600">
          <Icon name="document" size="xl" class="mx-auto mb-3 text-gray-300 dark:text-dark-600" />
          <p class="text-sm text-gray-500">{{ t('payment.invoice.noRecords') }}</p>
        </div>
        <template v-else>
          <div data-test="invoice-record-cards" class="divide-y divide-gray-100 overflow-hidden rounded-lg border border-gray-200 sm:hidden dark:divide-dark-700 dark:border-dark-700">
            <div
              v-for="application in invoiceRecords"
              :key="application.id"
              class="bg-white p-4 dark:bg-dark-800"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.invoice.applicationNo') }}</div>
                  <div class="mt-1 break-all font-mono text-xs text-gray-700 dark:text-gray-300">
                    {{ application.external_id || `#${application.id}` }}
                  </div>
                </div>
                <span class="inline-flex shrink-0 rounded-full px-2 py-1 text-xs font-medium" :class="invoiceStatusClass(application.status)">
                  {{ t(`payment.invoice.status.${application.status}`, application.status) }}
                </span>
              </div>
              <div class="mt-4 grid grid-cols-2 gap-3 text-sm">
                <div class="min-w-0">
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.invoice.invoiceTitle') }}</div>
                  <div class="mt-1 break-words font-medium text-gray-950 dark:text-white">{{ application.title || '--' }}</div>
                </div>
                <div class="text-right">
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.invoice.invoiceAmount') }}</div>
                  <div class="mt-1 text-lg font-semibold tabular-nums text-gray-950 dark:text-white">{{ application.currency }} {{ application.total_amount || '--' }}</div>
                </div>
                <div class="col-span-2">
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.orders.createdAt') }}</div>
                  <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ formatDate(application.created_at) }}</div>
                </div>
              </div>
              <div v-if="application.status === 'completed' || application.status === 'pending' || application.status === 'approved'" class="mt-4 flex justify-end border-t border-gray-100 pt-3 dark:border-dark-700">
                <button
                  v-if="application.status === 'completed'"
                  class="btn btn-secondary btn-sm"
                  :disabled="invoiceActionId === application.id"
                  @click="downloadInvoice(application)"
                >
                  <Icon name="download" size="sm" />
                  <span>{{ t('payment.invoice.downloadPdf') }}</span>
                </button>
                <button
                  v-else
                  class="btn btn-ghost btn-sm text-red-600 dark:text-red-400"
                  :disabled="invoiceActionId === application.id"
                  @click="cancelInvoiceApplication(application)"
                >
                  <Icon name="x" size="sm" />
                  <span>{{ t('payment.invoice.cancelApplication') }}</span>
                </button>
              </div>
            </div>
          </div>
          <div data-test="invoice-record-table" class="hidden overflow-x-auto rounded-lg border border-gray-200 sm:block dark:border-dark-700">
          <table class="w-full min-w-[760px] text-left text-sm">
            <thead class="border-b border-gray-200 bg-gray-50 text-xs text-gray-500 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-400">
              <tr>
                <th class="px-3 py-3 font-medium">{{ t('payment.invoice.applicationNo') }}</th>
                <th class="px-3 py-3 font-medium">{{ t('payment.invoice.invoiceTitle') }}</th>
                <th class="px-3 py-3 font-medium">{{ t('payment.invoice.invoiceAmount') }}</th>
                <th class="px-3 py-3 font-medium">{{ t('payment.orders.status') }}</th>
                <th class="px-3 py-3 font-medium">{{ t('payment.orders.createdAt') }}</th>
                <th class="px-3 py-3 text-right font-medium">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-800">
              <tr v-for="application in invoiceRecords" :key="application.id">
                <td class="px-3 py-3 font-mono text-xs text-gray-700 dark:text-gray-300">{{ application.external_id || `#${application.id}` }}</td>
                <td class="px-3 py-3 text-gray-900 dark:text-white">{{ application.title || '--' }}</td>
                <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ application.currency }} {{ application.total_amount || '--' }}</td>
                <td class="px-3 py-3">
                  <span class="inline-flex rounded-full px-2 py-1 text-xs font-medium" :class="invoiceStatusClass(application.status)">
                    {{ t(`payment.invoice.status.${application.status}`, application.status) }}
                  </span>
                </td>
                <td class="px-3 py-3 text-xs text-gray-500 dark:text-gray-400">{{ formatDate(application.created_at) }}</td>
                <td class="px-3 py-3">
                  <div class="flex justify-end gap-2">
                    <button
                      v-if="application.status === 'completed'"
                      class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-primary-700 hover:bg-primary-50 dark:text-primary-300 dark:hover:bg-primary-950/30"
                      :disabled="invoiceActionId === application.id"
                      @click="downloadInvoice(application)"
                    >
                      <Icon name="download" size="sm" />
                      <span>{{ t('payment.invoice.downloadPdf') }}</span>
                    </button>
                    <button
                      v-if="application.status === 'pending' || application.status === 'approved'"
                      class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-950/30"
                      :disabled="invoiceActionId === application.id"
                      @click="cancelInvoiceApplication(application)"
                    >
                      <Icon name="x" size="sm" />
                      <span>{{ t('payment.invoice.cancelApplication') }}</span>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        </template>
        <div v-if="invoiceRecordTotal > invoiceRecordPageSize" class="flex items-center justify-between border-t border-gray-200 pt-4 dark:border-dark-700">
          <button class="btn btn-secondary" :disabled="invoiceRecordPage <= 1 || invoiceRecordsLoading" @click="changeInvoiceRecordPage(-1)">{{ t('common.back') }}</button>
          <span class="text-sm text-gray-500">{{ invoiceRecordPage }} / {{ invoiceRecordPages }}</span>
          <button class="btn btn-secondary" :disabled="invoiceRecordPage >= invoiceRecordPages || invoiceRecordsLoading" @click="changeInvoiceRecordPage(1)">{{ t('common.next') }}</button>
        </div>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type {
  PaymentOrder,
  InvoiceConfig,
  InvoiceDraft,
  InvoiceValidation,
  InvoiceApplyRequest,
  InvoiceApplication,
  InvoiceStatus,
} from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderTable from '@/components/payment/OrderTable.vue'
import alipayIcon from '@/assets/icons/alipay.svg'
import wxpayIcon from '@/assets/icons/wxpay.svg'
import paymentIcon from '@/assets/icons/payment.svg'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const loading = ref(false)
const actionLoading = ref(false)
const orders = ref<PaymentOrder[]>([])
const refundEligibleProviders = ref<Set<string>>(new Set())
const currentFilter = ref('')
const cancelTargetId = ref<number | null>(null)
const refundTarget = ref<PaymentOrder | null>(null)
const refundReason = ref('')
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
const invoiceConfig = reactive<InvoiceConfig>({
  enabled: false,
  supports_tax_payment: false,
  max_orders: 20,
  fee_payer: 'customer',
})
const invoiceSelectionMode = ref(false)
const selectedInvoiceOrderIds = ref<Set<number>>(new Set())
const invoiceNeedPayTax = ref(false)
const invoiceBusy = ref(false)
const invoiceDialogOpen = ref(false)
const invoiceDraft = ref<InvoiceDraft | null>(null)
const selectedTaxOrderNo = ref('')
const invoiceRecordsOpen = ref(false)
const invoiceRecordsLoading = ref(false)
const invoiceRecords = ref<InvoiceApplication[]>([])
const invoiceRecordTotal = ref(0)
const invoiceRecordPage = ref(1)
const invoiceRecordPageSize = 20
const invoiceActionId = ref<number | null>(null)
const invoiceForm = reactive<InvoiceApplyRequest>({
  buyer_type: 'company',
  title: '',
  taxpayer_id: '',
  buyer_address: '',
  buyer_phone: '',
  buyer_bank: '',
  buyer_bank_account: '',
  recipient_email: '',
})

interface InvoiceTaxPaymentOption {
  channel: string
  taxOrderNo: string
  payUrl: string
}

const statusFilters = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
])

const invoiceValidation = computed<InvoiceValidation>(() => invoiceDraft.value?.validation || {})
const invoiceCurrency = computed(() => invoiceValidation.value.currency || 'CNY')
const invoiceAllowsFeePayerChoice = computed(() => invoiceConfig.fee_payer === 'user_choice')
const configuredInvoiceNeedPayTax = computed(() => invoiceConfig.fee_payer !== 'platform')
const invoiceReady = computed(() => {
  if (!invoiceDraft.value?.need_pay_tax) return true
  return invoiceValidation.value.taxDueAmount === '0.00'
})
const invoiceTaxPayments = computed<InvoiceTaxPaymentOption[]>(() => {
  const options = Object.entries(invoiceValidation.value.taxPayments || {})
    .map(([channel, payment]) => ({ channel, taxOrderNo: payment.taxOrderNo, payUrl: payment.payUrl }))
    .filter(payment => payment.taxOrderNo && payment.payUrl)
  if (options.length > 0) return options
  if (invoiceValidation.value.taxOrderNo && invoiceValidation.value.payUrl) {
    return [{ channel: 'alipay', taxOrderNo: invoiceValidation.value.taxOrderNo, payUrl: invoiceValidation.value.payUrl }]
  }
  return []
})
const invoiceBuyerTypes = computed<Array<{ value: InvoiceApplyRequest['buyer_type']; label: string }>>(() => [
  { value: 'company', label: t('payment.invoice.company') },
  { value: 'individual', label: t('payment.invoice.individual') },
])
const invoiceFormValid = computed(() => Boolean(
  invoiceForm.title.trim() &&
  invoiceForm.taxpayer_id.trim() &&
  /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(invoiceForm.recipient_email.trim())
))
const invoiceRecordPages = computed(() => Math.max(1, Math.ceil(invoiceRecordTotal.value / invoiceRecordPageSize)))

async function fetchOrders() {
  loading.value = true
  try {
    const res = await paymentAPI.getMyOrders({
      page: pagination.page,
      page_size: pagination.page_size,
      status: currentFilter.value || undefined,
    })
    orders.value = res.data.items || []
    pagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) { pagination.page = page; fetchOrders() }
function handlePageSizeChange(size: number) { pagination.page_size = size; pagination.page = 1; fetchOrders() }

function handleCancel(orderId: number) { cancelTargetId.value = orderId }

async function confirmCancel() {
  if (!cancelTargetId.value) return
  actionLoading.value = true
  try {
    await paymentAPI.cancelOrder(cancelTargetId.value)
    appStore.showSuccess(t('common.success'))
    cancelTargetId.value = null
    await fetchOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function openRefundDialog(order: PaymentOrder) { refundTarget.value = order; refundReason.value = '' }

async function confirmRefund() {
  if (!refundTarget.value || !refundReason.value.trim()) return
  actionLoading.value = true
  try {
    await paymentAPI.requestRefund(refundTarget.value.id, { reason: refundReason.value.trim() })
    appStore.showSuccess(t('common.success'))
    refundTarget.value = null
    refundReason.value = ''
    await fetchOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function canRequestRefund(order: PaymentOrder): boolean {
  if (order.status !== 'COMPLETED') return false
  if (!order.provider_instance_id) return false
  return refundEligibleProviders.value.has(order.provider_instance_id)
}

async function loadRefundEligibility() {
  try {
    const res = await paymentAPI.getRefundEligibleProviders()
    refundEligibleProviders.value = new Set(res.data.provider_instance_ids || [])
  } catch { /* ignore — default to hiding refund button */ }
}

async function loadInvoiceConfig() {
  try {
    const res = await paymentAPI.getInvoiceConfig()
    Object.assign(invoiceConfig, res.data)
    if (!['customer', 'platform', 'user_choice'].includes(invoiceConfig.fee_payer)) {
      invoiceConfig.fee_payer = 'customer'
    }
    invoiceNeedPayTax.value = configuredInvoiceNeedPayTax.value
  } catch {
    invoiceConfig.enabled = false
  }
}

async function loadCurrentInvoiceDraft() {
  if (!invoiceConfig.enabled) return
  try {
    const res = await paymentAPI.getCurrentInvoiceDraft()
    if (!res.data) return
    invoiceDraft.value = res.data
    invoiceNeedPayTax.value = res.data.need_pay_tax
    selectedInvoiceOrderIds.value = new Set(res.data.order_ids)
    invoiceSelectionMode.value = true
  } catch {
    // A failed draft lookup must not prevent ordinary order management.
  }
}

function startInvoiceSelection(orderId?: number) {
  if (invoiceDraft.value) {
    resumeInvoiceDraft()
    return
  }
  invoiceNeedPayTax.value = configuredInvoiceNeedPayTax.value
  invoiceSelectionMode.value = true
  if (orderId) {
    selectedInvoiceOrderIds.value = new Set([orderId])
  } else if (currentFilter.value !== 'COMPLETED') {
    currentFilter.value = 'COMPLETED'
    pagination.page = 1
    void fetchOrders()
  }
}

function toggleInvoiceOrder(orderId: number) {
  const selected = new Set(selectedInvoiceOrderIds.value)
  if (selected.has(orderId)) {
    selected.delete(orderId)
  } else if (selected.size < invoiceConfig.max_orders) {
    selected.add(orderId)
  }
  selectedInvoiceOrderIds.value = selected
}

function cancelInvoiceSelection() {
  invoiceSelectionMode.value = false
  selectedInvoiceOrderIds.value = new Set()
  invoiceNeedPayTax.value = configuredInvoiceNeedPayTax.value
  invoiceDraft.value = null
  selectedTaxOrderNo.value = ''
}

async function validateInvoiceSelection() {
  if (selectedInvoiceOrderIds.value.size === 0) return
  invoiceBusy.value = true
  try {
    const res = await paymentAPI.validateInvoiceOrders([...selectedInvoiceOrderIds.value], invoiceNeedPayTax.value)
    invoiceDraft.value = res.data
    selectedTaxOrderNo.value = ''
    invoiceDialogOpen.value = true
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('payment.invoice.validationFailed')))
  } finally {
    invoiceBusy.value = false
  }
}

function closeInvoiceDialog() {
  if (invoiceBusy.value) return
  invoiceDialogOpen.value = false
}

function resumeInvoiceDraft() {
  if (!invoiceDraft.value) return
  selectedInvoiceOrderIds.value = new Set(invoiceDraft.value.order_ids)
  invoiceNeedPayTax.value = invoiceDraft.value.need_pay_tax
  invoiceSelectionMode.value = true
  invoiceDialogOpen.value = true
}

async function abandonInvoiceDraft() {
  if (!invoiceDraft.value || invoiceDraft.value.need_pay_tax) return
  invoiceBusy.value = true
  try {
    await paymentAPI.abandonInvoiceDraft(invoiceDraft.value.draft_id)
    cancelInvoiceSelection()
    resetInvoiceForm()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    invoiceBusy.value = false
  }
}

function taxChannelLabel(channel: string): string {
  if (channel === 'wxpay') return t('payment.methods.wxpay')
  if (channel === 'alipay') return t('payment.methods.alipay')
  return channel
}

function invoicePaymentMethodIcon(channel: string): string {
  const normalized = channel.toLowerCase()
  if (normalized.includes('alipay')) return alipayIcon
  if (normalized.includes('wxpay') || normalized.includes('wechat')) return wxpayIcon
  return paymentIcon
}

function openInvoiceTaxPayment(paymentOption: InvoiceTaxPaymentOption) {
  selectedTaxOrderNo.value = paymentOption.taxOrderNo
  window.open(paymentOption.payUrl, '_blank', 'noopener,noreferrer')
}

async function checkInvoiceTaxPayment() {
  if (!invoiceDraft.value || !selectedTaxOrderNo.value) return
  invoiceBusy.value = true
  try {
    const res = await paymentAPI.checkInvoiceTaxPayment(invoiceDraft.value.draft_id, selectedTaxOrderNo.value)
    invoiceDraft.value = {
      ...invoiceDraft.value,
      validation: res.data.validation || invoiceDraft.value.validation,
      tax_order_nos: res.data.tax_order_nos,
    }
    if (!res.data.paid) {
      appStore.showError(t('payment.invoice.taxPaymentNotConfirmed'))
    } else if (!res.data.ready) {
      selectedTaxOrderNo.value = ''
      appStore.showError(t('payment.invoice.taxReconciliationPending'))
    } else {
      appStore.showSuccess(t('payment.invoice.taxPaymentConfirmed'))
    }
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('payment.invoice.taxCheckFailed')))
  } finally {
    invoiceBusy.value = false
  }
}

async function submitInvoiceApplication() {
  if (!invoiceDraft.value || !invoiceReady.value || !invoiceFormValid.value) return
  invoiceBusy.value = true
  try {
    await paymentAPI.applyInvoice(invoiceDraft.value.draft_id, {
      ...invoiceForm,
      title: invoiceForm.title.trim(),
      taxpayer_id: invoiceForm.taxpayer_id.trim(),
      recipient_email: invoiceForm.recipient_email.trim(),
    })
    appStore.showSuccess(t('payment.invoice.applicationSubmitted'))
    invoiceDialogOpen.value = false
    cancelInvoiceSelection()
    resetInvoiceForm()
    if (invoiceRecordsOpen.value) await fetchInvoiceRecords()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('payment.invoice.submitFailed')))
  } finally {
    invoiceBusy.value = false
  }
}

function resetInvoiceForm() {
  Object.assign(invoiceForm, {
    buyer_type: 'company',
    title: '',
    taxpayer_id: '',
    buyer_address: '',
    buyer_phone: '',
    buyer_bank: '',
    buyer_bank_account: '',
    recipient_email: '',
  })
}

async function openInvoiceRecords() {
  invoiceRecordsOpen.value = true
  invoiceRecordPage.value = 1
  await fetchInvoiceRecords()
}

async function fetchInvoiceRecords() {
  invoiceRecordsLoading.value = true
  try {
    const res = await paymentAPI.getInvoices({ page: invoiceRecordPage.value, page_size: invoiceRecordPageSize })
    invoiceRecords.value = res.data.items || []
    invoiceRecordTotal.value = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('payment.invoice.recordsFailed')))
  } finally {
    invoiceRecordsLoading.value = false
  }
}

async function changeInvoiceRecordPage(delta: number) {
  const nextPage = Math.min(invoiceRecordPages.value, Math.max(1, invoiceRecordPage.value + delta))
  if (nextPage === invoiceRecordPage.value) return
  invoiceRecordPage.value = nextPage
  await fetchInvoiceRecords()
}

async function cancelInvoiceApplication(application: InvoiceApplication) {
  if (!window.confirm(t('payment.invoice.confirmCancelApplication'))) return
  invoiceActionId.value = application.id
  try {
    await paymentAPI.cancelInvoice(application.id)
    appStore.showSuccess(t('payment.invoice.applicationCanceled'))
    await fetchInvoiceRecords()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('payment.invoice.cancelFailed')))
  } finally {
    invoiceActionId.value = null
  }
}

async function downloadInvoice(application: InvoiceApplication) {
  invoiceActionId.value = application.id
  try {
    const res = await paymentAPI.downloadInvoicePDF(application.id)
    const objectURL = URL.createObjectURL(res.data)
    const link = document.createElement('a')
    link.href = objectURL
    link.download = `invoice-${application.id}.pdf`
    document.body.appendChild(link)
    link.click()
    link.remove()
    setTimeout(() => URL.revokeObjectURL(objectURL), 0)
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('payment.invoice.downloadFailed')))
  } finally {
    invoiceActionId.value = null
  }
}

function invoiceStatusClass(status: InvoiceStatus): string {
  if (status === 'completed') return 'bg-green-100 text-green-700 dark:bg-green-950/40 dark:text-green-300'
  if (status === 'rejected' || status === 'failed') return 'bg-red-100 text-red-700 dark:bg-red-950/40 dark:text-red-300'
  if (status === 'canceled') return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
  if (status === 'approved') return 'bg-blue-100 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300'
  if (status === 'submission_unknown') return 'bg-amber-100 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300'
  return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-950/40 dark:text-yellow-300'
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString()
}

onMounted(() => {
  fetchOrders()
  loadRefundEligibility()
  void (async () => {
    await loadInvoiceConfig()
    await loadCurrentInvoiceDraft()
  })()
})
</script>
