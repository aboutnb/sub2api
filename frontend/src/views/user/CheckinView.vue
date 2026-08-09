<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-4 sm:space-y-5">
      <div v-if="loadingStatus" class="card flex items-center justify-center p-10">
        <LoadingSpinner />
      </div>

      <template v-else-if="status">
        <section data-testid="checkin-summary" class="card px-5 py-4 sm:px-6 sm:py-5">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex min-w-0 items-center gap-3.5">
              <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
                <CheckinCenterIcon class="h-5 w-5" />
              </div>
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="text-xs font-medium text-gray-500 dark:text-dark-400">{{ t('checkin.currentBalance') }}</p>
                  <span
                    class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-[11px] font-medium"
                    :class="status.checked_in_today
                      ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
                      : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'"
                  >
                    <span class="h-1.5 w-1.5 rounded-full" :class="status.checked_in_today ? 'bg-emerald-500' : 'border border-gray-400 dark:border-dark-400'" />
                    {{ status.checked_in_today ? t('checkin.checkedToday') : t('checkin.notCheckedToday') }}
                    <span v-if="status.checked_in_today" class="text-emerald-600/80 dark:text-emerald-400/80">· {{ modeLabel(status.today_record?.mode) }}</span>
                  </span>
                </div>
                <p class="mt-1 tabular-nums text-2xl font-semibold text-gray-950 dark:text-white sm:text-3xl">{{ formatMoney(user?.balance || 0) }}</p>
              </div>
            </div>

            <div
              v-if="status.enabled && status.eligible && availableModeCount > 0"
              data-testid="checkin-mode-actions"
              class="grid w-full gap-2 border-t border-gray-100 pt-4 sm:w-auto sm:border-l sm:border-t-0 sm:py-1 sm:pl-5 dark:border-dark-700"
              :class="availableModeCount === 1 ? 'grid-cols-1' : 'grid-cols-2'"
            >
              <button
                v-if="status.normal_enabled"
                type="button"
                data-testid="checkin-mode-normal"
                class="group flex min-h-11 min-w-0 items-center justify-center gap-2 rounded-lg bg-emerald-600 px-3.5 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-emerald-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2 dark:bg-emerald-600 dark:hover:bg-emerald-500 sm:min-w-36"
                :disabled="submitting || !status.can_check_in"
                :class="{ 'cursor-not-allowed opacity-50': submitting || !status.can_check_in }"
                @click="requestCheckin('normal')"
              >
                <Icon v-if="submittingMode === 'normal'" name="refresh" size="sm" class="shrink-0 animate-spin" />
                <Icon v-else name="gift" size="sm" class="shrink-0" />
                <span class="truncate">{{ submittingMode === 'normal' ? t('checkin.submitting') : t('checkin.normal') }}</span>
                <Icon v-if="!submitting && status.can_check_in" name="chevronRight" size="xs" class="hidden shrink-0 opacity-70 transition group-hover:translate-x-0.5 sm:block" />
              </button>

              <button
                v-if="status.lucky_enabled"
                type="button"
                data-testid="checkin-mode-lucky"
                class="group flex min-h-11 min-w-0 items-center justify-center gap-2 rounded-lg border border-gray-200 bg-white px-3.5 py-2 text-sm font-semibold text-gray-800 transition hover:border-cyan-300 hover:bg-cyan-50 hover:text-cyan-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-500 focus-visible:ring-offset-2 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-100 dark:hover:border-cyan-700 dark:hover:bg-cyan-900/20 dark:hover:text-cyan-200 sm:min-w-36"
                :disabled="submitting || !status.can_check_in"
                :class="{ 'cursor-not-allowed opacity-50': submitting || !status.can_check_in }"
                @click="requestCheckin('lucky')"
              >
                <Icon v-if="submittingMode === 'lucky'" name="refresh" size="sm" class="shrink-0 animate-spin" />
                <Icon v-else name="sparkles" size="sm" class="shrink-0 text-cyan-600 dark:text-cyan-400" />
                <span class="truncate">{{ submittingMode === 'lucky' ? t('checkin.submitting') : t('checkin.lucky') }}</span>
                <Icon v-if="!submitting && status.can_check_in" name="chevronRight" size="xs" class="hidden shrink-0 text-gray-400 transition group-hover:translate-x-0.5 group-hover:text-cyan-600 sm:block" />
              </button>
            </div>
          </div>
        </section>

        <section v-if="!status.enabled || !status.eligible || availableModeCount === 0" class="card border-amber-200 bg-amber-50 dark:border-amber-800/50 dark:bg-amber-900/20">
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

        <transition name="fade">
          <section v-if="result" class="card border-emerald-200 bg-emerald-50 dark:border-emerald-800/50 dark:bg-emerald-900/20">
            <div class="flex flex-col gap-5 p-6 sm:flex-row sm:items-center sm:justify-between">
              <div class="flex items-center gap-4">
                <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-emerald-100 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400">
                  <Icon name="checkCircle" size="md" />
                </div>
                <div>
                  <h3 class="text-sm font-semibold text-emerald-800 dark:text-emerald-300">{{ t('checkin.success') }}</h3>
                  <p class="mt-1 text-sm text-emerald-700 dark:text-emerald-400">
                    {{ modeLabel(result.mode) }}
                    · {{ formatDate(result.checked_in_at) }}
                  </p>
                </div>
              </div>
              <div class="grid w-full grid-cols-2 gap-x-6 gap-y-4 sm:flex sm:w-auto sm:items-end sm:gap-8 sm:text-right">
                <div v-if="recordMultiplier(result)" data-testid="checkin-result-multiplier">
                  <p class="text-xs text-emerald-700/70 dark:text-emerald-400/70">{{ t('checkin.resultMultiplier') }}</p>
                  <p
                    class="mt-1 text-xl font-bold"
                    :class="result.random_value >= 0 ? 'text-cyan-700 dark:text-cyan-300' : 'text-rose-600 dark:text-rose-400'"
                  >
                    {{ recordMultiplier(result) }}
                  </p>
                </div>
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
          <div class="flex items-start gap-3">
            <div class="flex items-start gap-3">
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-emerald-100 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400">
                <Icon name="calendar" size="md" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('checkin.calendarTitle') }}</h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ monthLabel }} · {{ t('checkin.calendarHint') }}</p>
              </div>
            </div>
          </div>
          <div class="mt-5 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
            <div class="grid grid-cols-7 border-b border-gray-200 bg-gray-50/80 dark:border-dark-700 dark:bg-dark-800/80">
              <span
                v-for="(day, index) in weekDays"
                :key="day"
                class="calendar-weekday border-gray-200 py-2 text-center text-[11px] font-medium dark:border-dark-700"
                :class="index >= 5 ? 'text-gray-500 dark:text-dark-400' : 'text-gray-400 dark:text-dark-500'"
              >
                {{ day }}
              </span>
            </div>
            <div class="grid grid-cols-7">
              <span
                v-for="(cell, index) in calendarCells"
                :key="index"
                :data-testid="cell.day ? `calendar-day-${cell.day}` : undefined"
                :data-mode="cell.record?.mode"
                class="calendar-cell flex min-h-12 min-w-0 flex-col justify-between overflow-hidden border-gray-100 p-1 transition-colors dark:border-dark-700 sm:min-h-16 sm:p-2"
                :class="cellClass(cell)"
                :title="calendarCellTitle(cell)"
                :aria-label="calendarCellTitle(cell) || undefined"
                :aria-hidden="cell.day ? undefined : 'true'"
              >
                <template v-if="cell.day">
                  <span class="flex items-start justify-between gap-1">
                    <span class="text-xs font-medium tabular-nums sm:text-sm">{{ cell.day }}</span>
                    <Icon
                      v-if="cell.record"
                      :name="cell.record.mode === 'lucky' ? 'sparkles' : 'gift'"
                      size="xs"
                      class="shrink-0 opacity-80"
                    />
                  </span>
                  <span v-if="cell.record" class="flex min-w-0 flex-col items-end gap-0.5 overflow-hidden whitespace-nowrap leading-none">
                    <span v-if="recordMultiplier(cell.record)" class="text-[8px] font-medium text-cyan-700 dark:text-cyan-300 sm:text-[10px]">
                      {{ recordMultiplier(cell.record) }}
                    </span>
                    <span
                      class="text-[9px] font-semibold sm:text-[11px]"
                      :class="cell.record.reward_amount >= 0 ? 'text-emerald-700 dark:text-emerald-300' : 'text-rose-600 dark:text-rose-400'"
                    >
                      <span class="sm:hidden">{{ calendarMoney(cell.record.reward_amount, 2) }}</span>
                      <span class="hidden sm:inline">{{ calendarMoney(cell.record.reward_amount) }}</span>
                    </span>
                  </span>
                </template>
              </span>
            </div>
          </div>
          <div data-testid="calendar-month-summary" class="mt-4 flex flex-col gap-3 border-t border-gray-100 pt-4 sm:flex-row sm:items-center sm:justify-between dark:border-dark-700">
            <div class="flex items-center justify-between gap-5 sm:justify-start">
              <div>
                <p class="text-[11px] text-gray-500 dark:text-dark-400">{{ t('checkin.monthSummary') }}</p>
                <p class="mt-0.5 text-sm font-semibold text-gray-900 dark:text-white">{{ t('checkin.daysCount', { count: monthSummary.count }) }}</p>
              </div>
              <div class="border-l border-gray-200 pl-5 dark:border-dark-700">
                <p class="text-[11px] text-gray-500 dark:text-dark-400">{{ t('checkin.netChange') }}</p>
                <p class="mt-0.5 text-sm font-semibold" :class="monthSummary.total >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'" :title="signedMoney(monthSummary.total)">{{ compactSignedMoney(monthSummary.total) }}</p>
              </div>
            </div>
            <div class="grid grid-cols-2 gap-2">
              <div class="flex min-w-0 items-center gap-2 rounded-lg bg-emerald-50 px-3 py-2 dark:bg-emerald-900/20">
                <Icon name="gift" size="xs" class="shrink-0 text-emerald-600 dark:text-emerald-400" />
                <div class="min-w-0">
                  <p class="truncate text-[11px] text-gray-500 dark:text-dark-400">{{ t('checkin.normal') }} · {{ t('checkin.timesCount', { count: monthSummary.normal.count }) }}</p>
                  <p class="text-xs font-semibold" :class="monthSummary.normal.total >= 0 ? 'text-emerald-700 dark:text-emerald-300' : 'text-rose-600 dark:text-rose-400'" :title="signedMoney(monthSummary.normal.total)">{{ compactSignedMoney(monthSummary.normal.total) }}</p>
                </div>
              </div>
              <div class="flex min-w-0 items-center gap-2 rounded-lg bg-cyan-50 px-3 py-2 dark:bg-cyan-900/20">
                <Icon name="sparkles" size="xs" class="shrink-0 text-cyan-600 dark:text-cyan-400" />
                <div class="min-w-0">
                  <p class="truncate text-[11px] text-gray-500 dark:text-dark-400">{{ t('checkin.lucky') }} · {{ t('checkin.timesCount', { count: monthSummary.lucky.count }) }}</p>
                  <p class="text-xs font-semibold" :class="monthSummary.lucky.total >= 0 ? 'text-cyan-700 dark:text-cyan-300' : 'text-rose-600 dark:text-rose-400'" :title="signedMoney(monthSummary.lucky.total)">{{ compactSignedMoney(monthSummary.lucky.total) }}</p>
                </div>
              </div>
            </div>
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
                    <p class="flex flex-wrap items-center gap-1.5 text-sm font-medium text-gray-900 dark:text-white">
                      <span>{{ modeLabel(record.mode) }}</span>
                      <span
                        v-if="recordMultiplier(record)"
                        :data-testid="`history-multiplier-${record.id}`"
                        class="rounded bg-cyan-50 px-1.5 py-0.5 text-[10px] font-semibold text-cyan-700 dark:bg-cyan-900/30 dark:text-cyan-300"
                      >
                        {{ recordMultiplier(record) }}
                      </span>
                    </p>
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
    <LuckyCheckinConfirmDialog
      :show="luckyConfirmOpen"
      :reward-type="status?.lucky_reward_type || 'multiplier'"
      :min-multiplier="status?.lucky_min_multiplier || 0"
      :max-multiplier="status?.lucky_max_multiplier || 0"
      :submitting="submittingMode === 'lucky'"
      @confirm="confirmLuckyCheckin"
      @cancel="luckyConfirmOpen = false"
    />
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
import LuckyCheckinConfirmDialog from '@/components/checkin/LuckyCheckinConfirmDialog.vue'
import CheckinCenterIcon from '@/components/icons/CheckinCenterIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import type { CheckinRecord, CheckinStatus } from '@/types'
import { checkinErrorMessage } from '@/utils/checkinError'

const { t, locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const user = computed(() => authStore.user)
const loadingStatus = ref(true)
const loadingRecords = ref(true)
const submitting = ref(false)
const submittingMode = ref<'normal' | 'lucky' | null>(null)
const luckyConfirmOpen = ref(false)
const status = ref<CheckinStatus | null>(null)
const unavailableMessage = computed(() => {
  if (status.value?.unavailable_reason === 'account_too_new') return t('checkin.accountTooNew')
  if (status.value?.unavailable_reason === 'negative_balance') return t('checkin.negativeBalance')
  if (status.value?.unavailable_reason === 'grant_restricted') return t('checkin.grantRestricted')
  if (status.value?.unavailable_reason === 'no_modes_enabled') return t('checkin.noModesAvailable')
  return t('checkin.unavailable')
})
const availableModeCount = computed(() => Number(Boolean(status.value?.normal_enabled)) + Number(Boolean(status.value?.lucky_enabled)))
const records = ref<CheckinRecord[]>([])
const calendarRecords = ref<CheckinRecord[]>([])
const recordsPage = ref(1)
const recordsPages = ref(1)
const result = ref<CheckinRecord | null>(null)
const statusError = ref('')
const recordsError = ref('')
const recordsPageSize = 10

interface CalendarCell {
  day: number
  date: string
  record: CheckinRecord | null
  today: boolean
  future: boolean
}

const weekDays = computed(() => locale.value === 'zh' ? ['一', '二', '三', '四', '五', '六', '日'] : ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'])
const monthLabel = computed(() => {
  const date = status.value?.business_date ? new Date(`${status.value.business_date}T00:00:00`) : new Date()
  return new Intl.DateTimeFormat(locale.value === 'zh' ? 'zh-CN' : 'en-US', { year: 'numeric', month: 'long' }).format(date)
})
const calendarCells = computed(() => {
  const current = status.value
  if (!current) return []
  const cells: CalendarCell[] = []
  for (let index = 0; index < current.first_weekday; index += 1) cells.push({ day: 0, date: '', record: null, today: false, future: false })
  const recordsByDate = new Map(calendarRecords.value.map(record => [record.checkin_date.slice(0, 10), record]))
  const [year, month] = current.business_date.split('-').map(Number)
  for (let day = 1; day <= current.days_in_month; day += 1) {
    const date = `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    cells.push({ day, date, record: recordsByDate.get(date) || null, today: date === current.business_date, future: date > current.business_date })
  }
  while (cells.length % 7 !== 0) cells.push({ day: 0, date: '', record: null, today: false, future: false })
  return cells
})
const monthSummary = computed(() => {
  const month = status.value?.business_date.slice(0, 7) || ''
  const summary = {
    count: 0,
    total: 0,
    normal: { count: 0, total: 0 },
    lucky: { count: 0, total: 0 },
  }
  for (const record of calendarRecords.value) {
    if (!month || record.checkin_date.slice(0, 7) !== month) continue
    const amount = Number(record.reward_amount || 0)
    const mode = record.mode === 'lucky' ? summary.lucky : summary.normal
    summary.count += 1
    summary.total += amount
    mode.count += 1
    mode.total += amount
  }
  return summary
})

function formatMoney(value: number) {
  return `$${Number(value || 0).toFixed(2)}`
}

function signedMoney(value: number) {
  const amount = Number(value || 0)
  return `${amount >= 0 ? '+' : '-'}$${Math.abs(amount).toFixed(2)}`
}

function calendarMoney(value: number, precision = 2) {
  const amount = Number(value || 0)
  if (Math.abs(amount) >= 10000) return compactSignedMoney(amount)
  const compact = Math.abs(amount).toFixed(precision).replace(/0+$/, '').replace(/\.$/, '') || '0'
  return `${amount >= 0 ? '+' : '-'}$${compact}`
}

function compactSignedMoney(value: number) {
  const amount = Number(value || 0)
  if (Math.abs(amount) < 10000) return signedMoney(amount)
  const compact = new Intl.NumberFormat(locale.value === 'zh' ? 'zh-CN' : 'en-US', {
    notation: 'compact',
    maximumFractionDigits: 2,
  }).format(Math.abs(amount))
  return `${amount >= 0 ? '+' : '-'}$${compact}`
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

function recordMultiplier(record?: CheckinRecord | null) {
  if (record?.mode !== 'lucky' || record.reward_type !== 'multiplier') return ''
  const value = Number(record.random_value || 0)
  return `${value >= 0 ? '+' : ''}${value.toFixed(2)}x`
}

function calendarCellTitle(cell: CalendarCell) {
  if (!cell.day) return ''
  if (!cell.record) return cell.date
  const multiplier = recordMultiplier(cell.record)
  const settlement = multiplier ? ` ${t('checkin.multiplierResult', { value: multiplier })}` : ''
  return `${cell.date}${settlement} ${signedMoney(cell.record.reward_amount)}`
}

function cellClass(cell: CalendarCell) {
  let stateClass = 'bg-white text-gray-600 dark:bg-dark-900 dark:text-dark-300'
  if (!cell.day) stateClass = 'bg-gray-50/60 dark:bg-dark-800/60'
  else if (cell.record?.mode === 'lucky') stateClass = 'bg-cyan-50/70 font-semibold text-cyan-800 dark:bg-cyan-900/20 dark:text-cyan-300'
  else if (cell.record) stateClass = 'bg-emerald-50/70 font-semibold text-emerald-800 dark:bg-emerald-900/20 dark:text-emerald-300'
  else if (cell.today) stateClass = 'bg-cyan-50 font-semibold text-cyan-700 dark:bg-cyan-900/20 dark:text-cyan-300'
  else if (cell.future) stateClass = 'bg-gray-50/70 text-gray-300 dark:bg-dark-800/70 dark:text-dark-600'
  return cell.today ? `${stateClass} ring-1 ring-inset ring-cyan-500` : stateClass
}

async function fetchStatus() {
  loadingStatus.value = true
  statusError.value = ''
  try {
    status.value = await checkinAPI.getStatus()
    result.value = status.value.today_record || null
  } catch (error) {
    status.value = null
    statusError.value = checkinErrorMessage(error, t, t('checkin.failedDescription'))
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
    recordsError.value = checkinErrorMessage(error, t, t('checkin.failedDescription'))
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
  const currentStatus = status.value
  const modeEnabled = mode === 'normal' ? currentStatus?.normal_enabled : currentStatus?.lucky_enabled
  if (!currentStatus?.can_check_in || !modeEnabled || submitting.value) return
  submitting.value = true
  submittingMode.value = mode
  try {
    const response = await checkinAPI.checkIn(mode, currentStatus.business_date)
    result.value = response.record
    await authStore.refreshUser()
    await Promise.all([fetchStatus(), fetchRecords(1), fetchCalendarRecords()])
    appStore.showSuccess(t('checkin.success'))
    window.dispatchEvent(new CustomEvent('checkin:updated'))
  } catch (error) {
    appStore.showError(checkinErrorMessage(error, t, t('checkin.failedDescription')))
  } finally {
    submitting.value = false
    submittingMode.value = null
  }
}

function requestCheckin(mode: 'normal' | 'lucky') {
  if (mode === 'lucky') {
    if (status.value?.can_check_in && status.value.lucky_enabled && !submitting.value) luckyConfirmOpen.value = true
    return
  }
  void submit('normal')
}

async function confirmLuckyCheckin() {
  if (submitting.value) return
  await submit('lucky')
  luckyConfirmOpen.value = false
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

.calendar-weekday,
.calendar-cell {
  border-right-width: 1px;
}

.calendar-cell {
  border-bottom-width: 1px;
}

.calendar-weekday:nth-child(7n),
.calendar-cell:nth-child(7n) {
  border-right-width: 0;
}

.calendar-cell:nth-last-child(-n + 7) {
  border-bottom-width: 0;
}
</style>
