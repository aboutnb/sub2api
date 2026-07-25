<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-4 sm:space-y-5">
      <div v-if="loadingStatus" class="card flex items-center justify-center p-10">
        <LoadingSpinner />
      </div>

      <template v-else-if="status">
        <section class="card overflow-hidden">
          <div class="bg-gradient-to-br from-emerald-500 to-cyan-600 px-5 py-6 text-white sm:px-6">
            <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div class="flex items-center gap-4">
                <div class="inline-flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-white/20 backdrop-blur-sm">
                  <Icon name="calendar" size="lg" />
                </div>
                <div>
                  <p class="text-xs font-medium text-emerald-50">{{ t('checkin.currentBalance') }}</p>
                  <p class="mt-1 text-3xl font-bold">{{ formatMoney(user?.balance || 0) }}</p>
                </div>
              </div>

              <div class="min-w-40 rounded-lg border border-white/15 bg-white/15 px-4 py-3 backdrop-blur-sm">
                <p class="text-base font-semibold">
                  {{ status.checked_in_today ? t('checkin.checkedToday') : t('checkin.notCheckedToday') }}
                </p>
                <p v-if="status.checked_in_today" class="mt-1 text-xs text-emerald-50">
                  {{ modeLabel(status.today_record?.mode) }}
                </p>
              </div>
            </div>
          </div>
        </section>

        <section v-if="!status.enabled || !status.eligible" class="card border-amber-200 bg-amber-50 dark:border-amber-800/50 dark:bg-amber-900/20">
          <div class="flex items-start gap-4 p-6">
            <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400">
              <Icon name="infoCircle" size="md" />
            </div>
            <div>
              <h2 class="text-sm font-semibold text-amber-800 dark:text-amber-300">{{ t('checkin.unavailableTitle') }}</h2>
              <p class="mt-2 text-sm text-amber-700 dark:text-amber-400">{{ unavailableMessage }}</p>
            </div>
          </div>
        </section>

        <template v-else>
          <section class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <button
              type="button"
              class="card group flex h-full flex-col border-gray-200 p-5 text-left transition hover:border-emerald-300 hover:shadow-lg dark:border-dark-700 dark:hover:border-emerald-700"
              :disabled="submitting || !status.can_check_in"
              :class="{ 'cursor-not-allowed opacity-60': submitting || !status.can_check_in }"
              @click="submit('normal')"
            >
              <div class="flex items-start justify-between gap-4">
                <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-100 dark:bg-emerald-900/30">
                  <Icon name="gift" size="md" class="text-emerald-600 dark:text-emerald-400" />
                </div>
                <span class="rounded-md bg-emerald-50 px-2 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
                  {{ status.checked_in_today ? t('checkin.checkedToday') : t('checkin.notCheckedToday') }}
                </span>
              </div>
              <h2 class="mt-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('checkin.normal') }}</h2>
              <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-dark-400">{{ t('checkin.normalHint') }}</p>
              <span class="mt-4 inline-flex items-center text-sm font-medium text-emerald-600 dark:text-emerald-400">
                <Icon v-if="submitting" name="refresh" size="sm" class="mr-1 animate-spin" />
                {{ status.can_check_in ? t('checkin.normalAction') : t('checkin.checkedIn') }}
                <Icon v-if="!submitting && status.can_check_in" name="arrowRight" size="sm" class="ml-1 transition group-hover:translate-x-0.5" />
              </span>
            </button>

            <button
              type="button"
              class="card group flex h-full flex-col border-gray-200 p-5 text-left transition hover:border-cyan-300 hover:shadow-lg dark:border-dark-700 dark:hover:border-cyan-700"
              :disabled="submitting || !status.can_check_in"
              :class="{ 'cursor-not-allowed opacity-60': submitting || !status.can_check_in }"
              @click="submit('lucky')"
            >
              <div class="flex items-start justify-between gap-4">
                <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-cyan-100 dark:bg-cyan-900/30">
                  <Icon name="sparkles" size="md" class="text-cyan-600 dark:text-cyan-400" />
                </div>
                <span class="rounded-md bg-cyan-50 px-2 py-1 text-xs font-medium text-cyan-700 dark:bg-cyan-900/30 dark:text-cyan-300">
                  {{ status.checked_in_today ? t('checkin.checkedToday') : t('checkin.notCheckedToday') }}
                </span>
              </div>
              <h2 class="mt-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('checkin.lucky') }}</h2>
              <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-dark-400">{{ t('checkin.luckyHint') }}</p>
              <span class="mt-4 inline-flex items-center text-sm font-medium text-cyan-600 dark:text-cyan-400">
                <Icon v-if="submitting" name="refresh" size="sm" class="mr-1 animate-spin" />
                {{ status.can_check_in ? t('checkin.normalAction') : t('checkin.checkedIn') }}
                <Icon v-if="!submitting && status.can_check_in" name="arrowRight" size="sm" class="ml-1 transition group-hover:translate-x-0.5" />
              </span>
            </button>
          </section>
        </template>

        <transition name="fade">
          <section v-if="result" class="card border-emerald-200 bg-emerald-50 dark:border-emerald-800/50 dark:bg-emerald-900/20">
            <div class="flex flex-col gap-5 p-6 sm:flex-row sm:items-center sm:justify-between">
              <div class="flex items-center gap-4">
                <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-emerald-100 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400">
                  <Icon name="checkCircle" size="md" />
                </div>
                <div>
                  <h3 class="text-sm font-semibold text-emerald-800 dark:text-emerald-300">{{ t('checkin.success') }}</h3>
                  <p class="mt-1 text-sm text-emerald-700 dark:text-emerald-400">{{ modeLabel(result.mode) }} · {{ formatDate(result.checked_in_at) }}</p>
                </div>
              </div>
              <div class="flex items-end gap-8 sm:text-right">
                <div>
                  <p class="text-xs text-emerald-700/70 dark:text-emerald-400/70">{{ t('checkin.change') }}</p>
                  <p class="mt-1 text-xl font-bold" :class="result.reward_amount >= 0 ? 'text-emerald-700 dark:text-emerald-300' : 'text-rose-600 dark:text-rose-400'">{{ signedMoney(result.reward_amount) }}</p>
                </div>
                <div>
                  <p class="text-xs text-emerald-700/70 dark:text-emerald-400/70">{{ t('checkin.balanceAfter') }}</p>
                  <p class="mt-1 text-base font-semibold text-emerald-800 dark:text-emerald-200">{{ formatMoney(result.balance_after) }}</p>
                </div>
              </div>
            </div>
          </section>
        </transition>

        <section class="card p-5 sm:p-6">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div class="flex items-start gap-3">
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-emerald-100 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400">
                <Icon name="calendar" size="md" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('checkin.calendarTitle') }}</h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ monthLabel }} · {{ t('checkin.calendarHint') }}</p>
              </div>
            </div>
            <div class="flex items-center gap-3 text-xs text-gray-500 dark:text-dark-400">
              <span class="inline-flex items-center gap-1.5"><i class="h-2.5 w-2.5 rounded-full bg-emerald-500" />{{ t('checkin.calendarChecked') }}</span>
              <span class="inline-flex items-center gap-1.5"><i class="h-2.5 w-2.5 rounded-full border border-cyan-500" />{{ t('checkin.calendarToday') }}</span>
              <span class="inline-flex items-center gap-1.5"><i class="h-2.5 w-2.5 rounded-full bg-gray-100 dark:bg-dark-700" />{{ t('checkin.calendarUnavailable') }}</span>
            </div>
          </div>
          <div class="mt-6 grid grid-cols-7 gap-1.5 text-center text-xs font-medium text-gray-400 dark:text-dark-500 sm:gap-2">
            <span v-for="day in weekDays" :key="day">{{ day }}</span>
          </div>
          <div class="mt-2 grid grid-cols-7 gap-1.5 sm:gap-2">
            <span v-for="(cell, index) in calendarCells" :key="index" class="flex min-h-10 items-center justify-center rounded-lg border text-sm transition sm:min-h-12" :class="cellClass(cell)">
              <span v-if="cell.day" class="flex flex-col items-center gap-0.5">
                <span>{{ cell.day }}</span>
                <Icon v-if="cell.checked" name="check" size="xs" />
              </span>
            </span>
          </div>
        </section>

        <section class="card overflow-hidden">
          <div class="flex flex-col gap-3 border-b border-gray-100 px-5 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-6 dark:border-dark-700">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('checkin.history') }}</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('checkin.historyHint') }}</p>
            </div>
            <button type="button" class="btn btn-secondary btn-sm self-start sm:self-auto" :disabled="loadingRecords" @click="fetchRecords(recordsPage)">
              <Icon name="refresh" size="sm" class="mr-1" :class="{ 'animate-spin': loadingRecords }" />
              {{ t('common.refresh') }}
            </button>
          </div>
          <div class="p-5 sm:p-6">
            <div v-if="loadingRecords" class="flex items-center justify-center py-10"><LoadingSpinner /></div>
            <div v-else-if="recordsError" class="rounded-xl border border-red-200 bg-red-50 p-5 text-sm text-red-700 dark:border-red-800/50 dark:bg-red-900/20 dark:text-red-300">
              <div class="flex items-center justify-between gap-4">
                <span>{{ recordsError }}</span>
                <button type="button" class="btn btn-secondary btn-sm shrink-0" @click="fetchRecords(recordsPage)">{{ t('checkin.retry') }}</button>
              </div>
            </div>
            <div v-else-if="records.length" class="space-y-3">
              <div v-for="record in records" :key="record.id" class="flex flex-col gap-3 rounded-xl bg-gray-50 p-4 md:flex-row md:items-center md:justify-between dark:bg-dark-800">
                <div class="flex items-center gap-4">
                  <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl" :class="record.mode === 'lucky' ? 'bg-cyan-100 text-cyan-600 dark:bg-cyan-900/30 dark:text-cyan-400' : 'bg-emerald-100 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400'">
                    <Icon :name="record.mode === 'lucky' ? 'sparkles' : 'gift'" size="md" />
                  </div>
                  <div>
                    <p class="text-sm font-medium text-gray-900 dark:text-white">{{ modeLabel(record.mode) }}</p>
                    <p class="text-xs text-gray-500 dark:text-dark-400">{{ formatRecordDate(record) }}</p>
                  </div>
                </div>
                <div class="text-left md:text-right">
                  <p class="text-sm font-semibold" :class="record.reward_amount >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'">{{ signedMoney(record.reward_amount) }}</p>
                  <p class="text-xs text-gray-500 dark:text-dark-400">{{ formatMoney(record.balance_before) }} → {{ formatMoney(record.balance_after) }}</p>
                </div>
              </div>
              <div v-if="recordsPages > 1" class="flex items-center justify-between gap-3 pt-2">
                <button type="button" class="btn btn-secondary btn-sm" :disabled="recordsPage <= 1 || loadingRecords" @click="fetchRecords(recordsPage - 1)"><Icon name="chevronLeft" size="sm" class="mr-1" />{{ t('checkin.previous') }}</button>
                <span class="text-xs text-gray-500 dark:text-dark-400">{{ t('checkin.pageInfo', { page: recordsPage, pages: recordsPages }) }}</span>
                <button type="button" class="btn btn-secondary btn-sm" :disabled="recordsPage >= recordsPages || loadingRecords" @click="fetchRecords(recordsPage + 1)">{{ t('checkin.next') }}<Icon name="chevronRight" size="sm" class="ml-1" /></button>
              </div>
            </div>
            <div v-else class="empty-state py-10">
              <div class="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-gray-100 dark:bg-dark-800"><Icon name="clock" size="lg" class="text-gray-400 dark:text-dark-500" /></div>
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('checkin.noHistory') }}</p>
            </div>
          </div>
        </section>
      </template>

      <section v-else class="card border-red-200 bg-red-50 p-6 text-sm text-red-700 dark:border-red-800/50 dark:bg-red-900/20 dark:text-red-300">
        <div class="flex items-center justify-between gap-4">
          <span>{{ statusError }}</span>
          <button type="button" class="btn btn-secondary btn-sm shrink-0" @click="load">{{ t('checkin.retry') }}</button>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { checkinAPI } from '@/api/checkin'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import type { CheckinRecord, CheckinStatus } from '@/types'

const { t, locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const user = computed(() => authStore.user)
const loadingStatus = ref(true)
const loadingRecords = ref(true)
const submitting = ref(false)
const status = ref<CheckinStatus | null>(null)
const unavailableMessage = computed(() => {
  if (status.value?.unavailable_reason === 'account_too_new') return t('checkin.accountTooNew')
  if (status.value?.unavailable_reason === 'negative_balance') return t('checkin.negativeBalance')
  return t('checkin.unavailable')
})
const records = ref<CheckinRecord[]>([])
const calendarRecords = ref<CheckinRecord[]>([])
const recordsPage = ref(1)
const recordsPages = ref(1)
const result = ref<CheckinRecord | null>(null)
const statusError = ref('')
const recordsError = ref('')
const recordsPageSize = 10

const weekDays = computed(() => locale.value === 'zh' ? ['一', '二', '三', '四', '五', '六', '日'] : ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'])
const monthLabel = computed(() => {
  const date = status.value?.business_date ? new Date(`${status.value.business_date}T00:00:00`) : new Date()
  return new Intl.DateTimeFormat(locale.value === 'zh' ? 'zh-CN' : 'en-US', { year: 'numeric', month: 'long' }).format(date)
})
const calendarCells = computed(() => {
  const current = status.value
  if (!current) return []
  const cells: Array<{ day: number; checked: boolean; today: boolean; future: boolean }> = []
  for (let index = 0; index < current.first_weekday; index += 1) cells.push({ day: 0, checked: false, today: false, future: false })
  const checkedDates = new Set(calendarRecords.value.map(record => record.checkin_date.slice(0, 10)))
  const [year, month] = current.business_date.split('-').map(Number)
  for (let day = 1; day <= current.days_in_month; day += 1) {
    const date = `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    cells.push({ day, checked: checkedDates.has(date), today: date === current.business_date, future: date > current.business_date })
  }
  return cells
})

function formatMoney(value: number) {
  return `$${Number(value || 0).toFixed(4)}`
}

function signedMoney(value: number) {
  const amount = Number(value || 0)
  return `${amount >= 0 ? '+' : '-'}$${Math.abs(amount).toFixed(4)}`
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(locale.value === 'zh' ? 'zh-CN' : 'en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }).format(new Date(value))
}

function formatRecordDate(record: CheckinRecord) {
  return `${record.checkin_date.slice(0, 10)} · ${formatDate(record.checked_in_at)}`
}

function modeLabel(mode?: CheckinRecord['mode']) {
  return mode === 'lucky' ? t('checkin.lucky') : t('checkin.normal')
}

function cellClass(cell: { day: number; checked: boolean; today: boolean; future: boolean }) {
  if (!cell.day) return 'invisible border-transparent'
  if (cell.checked) return 'border-emerald-200 bg-emerald-50 font-semibold text-emerald-700 dark:border-emerald-800/60 dark:bg-emerald-900/20 dark:text-emerald-300'
  if (cell.today) return 'border-cyan-500 bg-cyan-50 font-semibold text-cyan-700 dark:bg-cyan-900/20 dark:text-cyan-300'
  if (cell.future) return 'border-gray-100 bg-gray-50 text-gray-400 opacity-60 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-500'
  return 'border-gray-100 bg-gray-50 text-gray-600 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-300'
}

function errorMessage(error: unknown, fallback: string) {
  if (!error || typeof error !== 'object') return fallback
  const response = (error as { response?: { data?: { message?: string; detail?: string; reason?: string } } }).response
  if (response?.data?.reason === 'CHECKIN_SOURCE_LIMITED') return t('checkin.sourceLimited')
  if (response?.data?.reason === 'CHECKIN_RATE_LIMITED') return t('checkin.rateLimited')
  if (response?.data?.reason === 'CHECKIN_RISK_UNAVAILABLE') return t('checkin.riskUnavailable')
  if (response?.data?.reason === 'CHECKIN_SECURITY_UNAVAILABLE') return t('checkin.securityUnavailable')
  if (response?.data?.reason === 'CHECKIN_ENTROPY_UNAVAILABLE') return t('checkin.entropyUnavailable')
  if (response?.data?.reason === 'CHECKIN_ACCOUNT_TOO_NEW') return t('checkin.accountTooNew')
  if (response?.data?.reason === 'CHECKIN_NEGATIVE_BALANCE') return t('checkin.negativeBalance')
  if (response?.data?.reason === 'CHECKIN_INVALID_REQUEST' || response?.data?.reason?.startsWith('IDEMPOTENCY_')) return t('checkin.invalidRequest')
  return response?.data?.message || response?.data?.detail || (error as { message?: string }).message || fallback
}

async function fetchStatus() {
  loadingStatus.value = true
  statusError.value = ''
  try {
    status.value = await checkinAPI.getStatus()
    result.value = status.value.today_record || null
  } catch (error) {
    status.value = null
    statusError.value = errorMessage(error, t('checkin.failedDescription'))
  } finally {
    loadingStatus.value = false
  }
}

async function fetchRecords(page = 1) {
  loadingRecords.value = true
  recordsError.value = ''
  try {
    const response = await checkinAPI.getRecords(page, recordsPageSize)
    records.value = response.items || []
    recordsPage.value = response.page || page
    recordsPages.value = Math.max(1, response.pages || 1)
  } catch (error) {
    recordsError.value = errorMessage(error, t('checkin.failedDescription'))
  } finally {
    loadingRecords.value = false
  }
}

async function fetchCalendarRecords() {
  try {
    const response = await checkinAPI.getRecords(1, 100)
    calendarRecords.value = response.items || []
  } catch {
    // The history list has its own retry state; a calendar refresh should not block it.
  }
}

async function load() {
  await Promise.all([fetchStatus(), fetchRecords(1), fetchCalendarRecords()])
}

async function submit(mode: 'normal' | 'lucky') {
  if (!status.value?.can_check_in || submitting.value) return
  submitting.value = true
  try {
    const response = await checkinAPI.checkIn(mode, status.value.business_date)
    result.value = response.record
    await authStore.refreshUser()
    await Promise.all([fetchStatus(), fetchRecords(1), fetchCalendarRecords()])
    appStore.showSuccess(t('checkin.success'))
    window.dispatchEvent(new CustomEvent('checkin:updated'))
  } catch (error) {
    appStore.showError(errorMessage(error, t('checkin.failedDescription')))
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
