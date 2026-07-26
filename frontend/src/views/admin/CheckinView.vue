<template>
  <AppLayout>
    <div class="mx-auto max-w-[1600px] space-y-6">
      <header class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <div class="mb-2 flex items-center gap-3">
            <span class="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-100 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400">
              <CheckinCenterIcon class="h-5 w-5" />
            </span>
            <h1 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">{{ t('admin.checkin.title') }}</h1>
          </div>
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.checkin.description') }}</p>
        </div>
        <button type="button" class="btn btn-secondary self-start sm:self-auto" :disabled="loading" @click="load">
          <Icon name="refresh" size="sm" class="mr-2" :class="{ 'animate-spin': loading }" />
          {{ t('common.refresh') }}
        </button>
      </header>

      <div v-if="loadError" class="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-300">{{ loadError }}</div>

      <section class="grid grid-cols-2 gap-3 lg:grid-cols-5">
        <article v-for="item in overviewCards" :key="item.label" class="card p-5">
          <p class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ item.label }}</p>
          <p class="mt-2 text-2xl font-bold text-gray-900 dark:text-white" :class="item.tone">{{ item.value }}</p>
        </article>
      </section>

      <section class="card overflow-hidden">
        <div class="flex flex-col gap-4 border-b border-gray-100 px-5 py-5 sm:flex-row sm:items-center sm:justify-between sm:px-6 dark:border-dark-700">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.checkin.title') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.checkin.configHint') }}</p>
          </div>
          <div v-if="config" class="text-xs text-gray-400 dark:text-dark-500">{{ t('admin.checkin.version', { version: config.config_version }) }}</div>
        </div>
        <form class="space-y-5 p-5 sm:p-6" @submit.prevent="saveConfig">
          <div class="flex items-center justify-between rounded-xl bg-gray-50 p-4 dark:bg-dark-800">
            <div>
              <p class="font-medium text-gray-900 dark:text-white">{{ t('admin.checkin.enabled') }}</p>
              <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('admin.checkin.enabledHint') }}</p>
            </div>
            <Toggle v-model="form.enabled" />
          </div>
          <div class="grid grid-cols-1 gap-5 lg:grid-cols-2">
            <div class="rounded-xl border border-gray-200 p-4 dark:border-dark-700">
              <div class="flex items-center justify-between gap-4">
                <h3 class="font-medium text-gray-900 dark:text-white">{{ t('admin.checkin.normal') }}</h3>
                <div class="flex items-center gap-2">
                  <span class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.checkin.normalEnabled') }}</span>
                  <Toggle v-model="form.normal_enabled" data-testid="normal-enabled-toggle" :aria-label="t('admin.checkin.normalEnabled')" />
                </div>
              </div>
              <div class="mt-4 grid grid-cols-2 gap-3" :class="{ 'opacity-50': !form.normal_enabled }">
                <label class="text-sm text-gray-600 dark:text-dark-300">{{ t('admin.checkin.normalMin') }}<input v-model="form.normal_min" class="input mt-2" type="number" min="0" max="100" step="0.00000001" :disabled="!form.normal_enabled" /></label>
                <label class="text-sm text-gray-600 dark:text-dark-300">{{ t('admin.checkin.normalMax') }}<input v-model="form.normal_max" class="input mt-2" type="number" min="0" max="100" step="0.00000001" :disabled="!form.normal_enabled" /></label>
              </div>
            </div>
            <div class="rounded-xl border border-gray-200 p-4 dark:border-dark-700">
              <div class="flex items-center justify-between gap-4">
                <div>
                  <h3 class="font-medium text-gray-900 dark:text-white">{{ t('admin.checkin.lucky') }}</h3>
                  <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('admin.checkin.luckyTypeHint') }}</p>
                </div>
                <div class="flex shrink-0 items-center gap-2">
                  <span class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.checkin.luckyEnabled') }}</span>
                  <Toggle v-model="form.lucky_enabled" data-testid="lucky-enabled-toggle" :aria-label="t('admin.checkin.luckyEnabled')" />
                </div>
              </div>
              <div class="mt-4" :class="{ 'opacity-50': !form.lucky_enabled }">
                <div class="grid grid-cols-2 rounded-lg bg-gray-100 p-1 dark:bg-dark-800" role="group" :aria-label="t('admin.checkin.luckyType')">
                  <button type="button" class="min-h-8 px-3 text-xs font-medium" :class="form.lucky_reward_type === 'multiplier' ? 'rounded-md bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white' : 'text-gray-500 dark:text-dark-400'" :aria-pressed="form.lucky_reward_type === 'multiplier'" :disabled="!form.lucky_enabled" @click="form.lucky_reward_type = 'multiplier'">{{ t('admin.checkin.luckyMultiplier') }}</button>
                  <button type="button" class="min-h-8 px-3 text-xs font-medium" :class="form.lucky_reward_type === 'amount' ? 'rounded-md bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white' : 'text-gray-500 dark:text-dark-400'" :aria-pressed="form.lucky_reward_type === 'amount'" :disabled="!form.lucky_enabled" @click="form.lucky_reward_type = 'amount'">{{ t('admin.checkin.luckyAmount') }}</button>
                </div>
                <div v-if="form.lucky_reward_type === 'multiplier'" class="mt-4 grid grid-cols-2 gap-3">
                  <label class="text-sm text-gray-600 dark:text-dark-300">{{ t('admin.checkin.luckyMin') }}<input v-model="form.lucky_min_multiplier" class="input mt-2" type="number" min="-1" max="10" step="0.00000001" :disabled="!form.lucky_enabled" /></label>
                  <label class="text-sm text-gray-600 dark:text-dark-300">{{ t('admin.checkin.luckyMax') }}<input v-model="form.lucky_max_multiplier" class="input mt-2" type="number" min="-1" max="10" step="0.00000001" :disabled="!form.lucky_enabled" /></label>
                </div>
                <div v-else class="mt-4 grid grid-cols-2 gap-3">
                  <label class="text-sm text-gray-600 dark:text-dark-300">{{ t('admin.checkin.luckyAmountMin') }}<input v-model="form.lucky_amount_min" class="input mt-2" type="number" min="-100" max="100" step="0.00000001" :disabled="!form.lucky_enabled" /></label>
                  <label class="text-sm text-gray-600 dark:text-dark-300">{{ t('admin.checkin.luckyAmountMax') }}<input v-model="form.lucky_amount_max" class="input mt-2" type="number" min="-100" max="100" step="0.00000001" :disabled="!form.lucky_enabled" /></label>
                </div>
              </div>
            </div>
          </div>
          <div class="border-y border-gray-100 py-5 dark:border-dark-700">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <h3 class="font-medium text-gray-900 dark:text-white">{{ t('admin.checkin.riskTitle') }}</h3>
                <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('admin.checkin.riskHint') }}</p>
              </div>
              <div class="flex items-center gap-3">
                <span class="text-sm font-medium text-gray-600 dark:text-dark-300">{{ t('admin.checkin.riskEnabled') }}</span>
                <Toggle v-model="form.risk_control_enabled" />
              </div>
            </div>
            <div class="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-3" :class="{ 'opacity-50': !form.risk_control_enabled }">
              <label class="text-sm text-gray-600 dark:text-dark-300">
                {{ t('admin.checkin.minAccountAge') }}
                <input v-model.number="form.min_account_age_hours" class="input mt-2" type="number" min="0" max="720" step="1" :disabled="!form.risk_control_enabled" />
              </label>
              <label class="text-sm text-gray-600 dark:text-dark-300">
                {{ t('admin.checkin.ipWindow') }}
                <input v-model.number="form.ip_window_minutes" class="input mt-2" type="number" min="1" max="1440" step="1" :disabled="!form.risk_control_enabled" />
              </label>
              <label class="text-sm text-gray-600 dark:text-dark-300">
                {{ t('admin.checkin.ipMaxUsers') }}
                <input v-model.number="form.ip_max_users" class="input mt-2" type="number" min="1" max="10000" step="1" :disabled="!form.risk_control_enabled" />
              </label>
            </div>
          </div>
          <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
            <label class="block w-full sm:max-w-xl text-sm text-gray-600 dark:text-dark-300">{{ t('admin.checkin.reason') }}<textarea v-model="form.change_reason" class="input mt-2 min-h-20" maxlength="500" :placeholder="t('admin.checkin.reasonPlaceholder')" required /></label>
            <button type="submit" class="btn btn-primary self-start sm:self-auto" :disabled="saving || !config">
              <Icon name="check" size="sm" class="mr-2" />{{ saving ? t('admin.checkin.saving') : t('admin.checkin.save') }}
            </button>
          </div>
        </form>
      </section>

      <section class="card overflow-hidden">
        <div class="border-b border-gray-100 px-5 py-5 dark:border-dark-700 sm:px-6">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.checkin.records') }}</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('admin.checkin.recordsHint') }}</p>
        </div>
        <div class="flex flex-col gap-3 border-b border-gray-100 p-5 sm:flex-row sm:items-end dark:border-dark-700 sm:px-6">
          <label class="w-full text-sm text-gray-600 dark:text-dark-300 sm:w-48">{{ t('admin.checkin.date') }}<input v-model="filters.date" type="date" class="input mt-2" @change="searchRecords" /></label>
          <label class="w-full text-sm text-gray-600 dark:text-dark-300 sm:w-64">{{ t('admin.checkin.email') }}<input v-model.trim="filters.email" type="search" class="input mt-2" @keyup.enter="searchRecords" /></label>
          <div class="flex gap-2">
            <button type="button" class="btn btn-primary" @click="searchRecords"><Icon name="search" size="sm" class="mr-2" />{{ t('admin.checkin.search') }}</button>
            <button type="button" class="btn btn-secondary" @click="resetFilters">{{ t('admin.checkin.reset') }}</button>
          </div>
        </div>
        <div v-if="recordsError" class="border-b border-red-200 bg-red-50 px-5 py-3 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-300 sm:px-6">{{ recordsError }}</div>
        <div v-if="recordsLoading" class="flex justify-center p-12"><LoadingSpinner /></div>
        <div v-else-if="!records.length" class="p-12 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('admin.checkin.empty') }}</div>
        <div v-else class="overflow-x-auto">
          <table class="min-w-[1050px] w-full text-left text-sm">
            <thead class="bg-gray-50 text-xs uppercase tracking-wide text-gray-500 dark:bg-dark-800 dark:text-dark-400">
              <tr><th class="px-5 py-3">{{ t('admin.checkin.checkedAt') }}</th><th class="px-5 py-3">{{ t('admin.checkin.user') }}</th><th class="px-5 py-3">{{ t('admin.checkin.mode') }}</th><th class="px-5 py-3">{{ t('admin.checkin.randomValue') }}</th><th class="px-5 py-3">{{ t('admin.checkin.before') }}</th><th class="px-5 py-3">{{ t('admin.checkin.change') }}</th><th class="px-5 py-3">{{ t('admin.checkin.after') }}</th></tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="record in records" :key="record.id" class="text-gray-700 dark:text-dark-300">
                <td class="whitespace-nowrap px-5 py-4"><strong class="text-gray-900 dark:text-white">{{ record.checkin_date.slice(0, 10) }}</strong><br /><span class="text-xs text-gray-400">{{ formatDate(record.checked_in_at) }}</span></td>
                <td class="px-5 py-4"><span class="font-medium text-gray-900 dark:text-white">{{ record.user_email || `#${record.user_id}` }}</span><br /><span class="text-xs text-gray-400">ID {{ record.user_id }}</span></td>
                <td class="px-5 py-4"><span class="badge" :class="record.mode === 'lucky' ? 'badge-warning' : 'badge-success'">{{ record.mode === 'lucky' ? t('admin.checkin.luckyLabel') : t('admin.checkin.normalLabel') }}</span></td>
                <td class="px-5 py-4 font-mono">{{ randomValue(record) }}</td>
                <td class="px-5 py-4 font-mono">{{ money(record.balance_before) }}</td>
                <td class="px-5 py-4 font-mono font-semibold" :class="record.reward_amount >= 0 ? 'text-emerald-600' : 'text-rose-600'">{{ signedMoney(record.reward_amount) }}</td>
                <td class="px-5 py-4 font-mono font-semibold text-gray-900 dark:text-white">{{ money(record.balance_after) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <Pagination v-if="pagination.total > 0" :page="pagination.page" :total="pagination.total" :page-size="pagination.page_size" @update:page="loadRecords" @update:pageSize="handlePageSizeChange" />
      </section>
    </div>
    <TotpStepUpDialog :controller="checkinStepUp" />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import CheckinCenterIcon from '@/components/icons/CheckinCenterIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import Toggle from '@/components/common/Toggle.vue'
import TotpStepUpDialog from '@/components/auth/TotpStepUpDialog.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { useStepUp, isStepUpBlocked, isStepUpCancelled, stepUpBlockReason } from '@/composables/useStepUp'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { AdminCheckinConfig, AdminCheckinOverview, AdminCheckinRecord } from '@/api/admin/checkin'

const { t, locale } = useI18n()
const appStore = useAppStore()
const checkinStepUp = useStepUp()
const loading = ref(true)
const recordsLoading = ref(true)
const saving = ref(false)
const loadError = ref('')
const recordsError = ref('')
const config = ref<AdminCheckinConfig | null>(null)
const overview = ref<AdminCheckinOverview>({ business_date: '', total: 0, normal_count: 0, lucky_count: 0, positive_total: 0, negative_total: 0 })
const records = ref<AdminCheckinRecord[]>([])
const filters = reactive({ date: '', email: '' })
const form = reactive({
  enabled: false,
  normal_enabled: true,
  lucky_enabled: true,
  normal_min: '0.01',
  normal_max: '0.05',
  lucky_reward_type: 'multiplier' as 'multiplier' | 'amount',
  lucky_min_multiplier: '-0.05',
  lucky_max_multiplier: '0.10',
  lucky_amount_min: '-0.05',
  lucky_amount_max: '0.10',
  risk_control_enabled: true,
  min_account_age_hours: 24,
  ip_window_minutes: 10,
  ip_max_users: 20,
  change_reason: ''
})
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const overviewCards = computed(() => [
  { label: t('admin.checkin.total'), value: overview.value.total, tone: 'text-gray-900 dark:text-white' },
  { label: t('admin.checkin.normalCount'), value: overview.value.normal_count, tone: 'text-emerald-600 dark:text-emerald-400' },
  { label: t('admin.checkin.luckyCount'), value: overview.value.lucky_count, tone: 'text-amber-600 dark:text-amber-400' },
  { label: t('admin.checkin.positiveTotal'), value: money(overview.value.positive_total), tone: 'text-emerald-600 dark:text-emerald-400' },
  { label: t('admin.checkin.negativeTotal'), value: money(overview.value.negative_total), tone: 'text-rose-600 dark:text-rose-400' }
])

function money(value: number) { return `$${Number(value || 0).toFixed(4)}` }
function signedMoney(value: number) { const amount = Number(value || 0); return `${amount >= 0 ? '+' : '-'}$${Math.abs(amount).toFixed(4)}` }
function randomValue(record: AdminCheckinRecord) { return record.reward_type === 'multiplier' ? `${(record.random_value * 100).toFixed(2)}%` : signedMoney(record.random_value) }
function formatDate(value: string) { return new Intl.DateTimeFormat(locale.value === 'zh' ? 'zh-CN' : 'en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }).format(new Date(value)) }
function applyConfig(value: AdminCheckinConfig) {
  config.value = value
  form.enabled = value.enabled
  form.normal_enabled = value.normal_enabled
  form.lucky_enabled = value.lucky_enabled
  form.normal_min = value.normal_min
  form.normal_max = value.normal_max
  form.lucky_reward_type = value.lucky_reward_type
  form.lucky_min_multiplier = value.lucky_min_multiplier
  form.lucky_max_multiplier = value.lucky_max_multiplier
  form.lucky_amount_min = value.lucky_amount_min
  form.lucky_amount_max = value.lucky_amount_max
  form.risk_control_enabled = value.risk_control_enabled
  form.min_account_age_hours = value.min_account_age_hours
  form.ip_window_minutes = value.ip_window_minutes
  form.ip_max_users = value.ip_max_users
}
function saveErrorMessage(value: unknown) {
  return extractApiErrorMessage(value, t('admin.checkin.saveFailed'), {
    CHECKIN_CHANGE_REASON_REQUIRED: t('admin.checkin.reasonRequired'),
    CHECKIN_CHANGE_REASON_TOO_LONG: t('admin.checkin.reasonTooLong'),
    CHECKIN_CONFIG_INVALID: t('admin.checkin.configInvalid'),
    CHECKIN_CONFIG_VERSION_CONFLICT: t('admin.checkin.configConflict')
  })
}
function decimalString(value: string | number) { return String(value).trim() }

async function loadRecords(page = pagination.page) {
  recordsLoading.value = true; recordsError.value = ''
  try { const response = await adminAPI.checkin.getRecords({ page, page_size: pagination.page_size, date: filters.date || undefined, email: filters.email || undefined }); records.value = response.items || []; pagination.page = response.page; pagination.total = response.total } catch (value) { recordsError.value = extractApiErrorMessage(value, t('admin.checkin.recordsLoadFailed')) } finally { recordsLoading.value = false }
}
async function load() {
  loading.value = true; loadError.value = ''
  try { const [nextConfig, nextOverview] = await Promise.all([adminAPI.checkin.getConfig(), adminAPI.checkin.getOverview(filters.date || undefined)]); applyConfig(nextConfig); overview.value = nextOverview; await loadRecords(1) } catch (value) { loadError.value = extractApiErrorMessage(value, t('admin.checkin.loadFailed')); recordsLoading.value = false } finally { loading.value = false }
}
async function saveConfig() {
  if (!config.value || saving.value) return
  const changeReason = form.change_reason.trim()
  if (!changeReason) {
    appStore.showError(t('admin.checkin.reasonRequired'))
    return
  }
  saving.value = true
  try {
    const next = await checkinStepUp.run(() => adminAPI.checkin.updateConfig({
      ...form,
      normal_min: decimalString(form.normal_min),
      normal_max: decimalString(form.normal_max),
      lucky_min_multiplier: decimalString(form.lucky_min_multiplier),
      lucky_max_multiplier: decimalString(form.lucky_max_multiplier),
      lucky_amount_min: decimalString(form.lucky_amount_min),
      lucky_amount_max: decimalString(form.lucky_amount_max),
      change_reason: changeReason,
      expected_config_version: config.value!.config_version
    }))
    applyConfig(next)
    form.change_reason = ''
    await appStore.fetchPublicSettings(true)
    window.dispatchEvent(new CustomEvent('checkin:config-updated'))
    appStore.showSuccess(t('admin.checkin.saveSuccess'))
  } catch (value) {
    if (isStepUpCancelled(value)) return
    if (isStepUpBlocked(value)) {
      appStore.showError(stepUpBlockReason(value) === 'STEP_UP_ADMIN_API_KEY_FORBIDDEN' ? t('stepUp.adminApiKeyForbidden') : t('stepUp.notEnabled'))
      return
    }
    appStore.showError(saveErrorMessage(value))
  } finally { saving.value = false }
}
function searchRecords() { loadRecords(1) }
function resetFilters() { filters.date = ''; filters.email = ''; searchRecords() }
function handlePageSizeChange(value: number) { pagination.page_size = value; loadRecords(1) }
onMounted(load)
</script>
