<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="card p-4 sm:p-5">
          <div class="flex flex-wrap items-end justify-between gap-4">
            <div class="flex flex-1 flex-wrap items-end gap-3">
              <div class="w-full sm:min-w-[280px] sm:flex-1">
                <label class="input-label">{{ t('common.search') }}</label>
                <div class="relative">
                  <Icon name="search" size="md" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                  <input
                    v-model.trim="filters.q"
                    type="search"
                    class="input pl-10"
                    :placeholder="t('admin.authIPBan.searchPlaceholder')"
                    @keyup.enter="search"
                  />
                </div>
              </div>
              <div class="w-full sm:w-44">
                <label class="input-label">{{ t('admin.authIPBan.filters.status') }}</label>
                <Select v-model="filters.status" :options="statusOptions" @change="search" />
              </div>
            </div>
            <div class="flex w-full justify-end gap-2 sm:w-auto">
              <button type="button" class="btn btn-secondary" :disabled="loading" @click="resetFilters">
                {{ t('common.reset') }}
              </button>
              <button type="button" class="btn btn-primary inline-flex items-center gap-2" :disabled="loading" @click="loadRecords">
                <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
                {{ t('admin.authIPBan.refresh') }}
              </button>
            </div>
          </div>

          <div v-if="policies.length" class="mt-4 border-t border-gray-100 pt-4 dark:border-dark-700">
            <div class="mb-2 text-xs font-semibold text-gray-500 dark:text-gray-400">
              {{ t('admin.authIPBan.policy.title') }}
            </div>
            <div class="flex flex-wrap gap-2">
              <div
                v-for="policy in policies"
                :key="policy.ua_category"
                class="rounded-md border border-gray-200 bg-gray-50 px-3 py-2 text-xs text-gray-600 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300"
              >
                <span class="font-semibold text-gray-900 dark:text-white">{{ categoryLabel(policy.ua_category) }}</span>
                <span class="mx-1 text-gray-300 dark:text-dark-500">·</span>
                {{ policySummary(policy) }}
                <span class="ml-1 text-gray-400">({{ scopeLabel(policy.ban_scope) }})</span>
              </div>
            </div>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="records" :loading="loading" row-key="id">
          <template #cell-source="{ row }">
            <div class="min-w-0 max-w-[220px]">
              <div class="whitespace-nowrap font-mono text-sm font-semibold text-gray-900 dark:text-white">
                {{ row.ip_address }}
              </div>
              <div class="mt-1 truncate text-xs text-gray-400" :title="row.trigger_path">
                {{ row.trigger_path }}
              </div>
            </div>
          </template>

          <template #cell-fingerprint="{ row }">
            <div class="min-w-0 max-w-[280px]">
              <div class="flex flex-wrap items-center gap-1.5">
                <span :class="categoryBadgeClass(row.ua_category)">{{ categoryLabel(row.ua_category) }}</span>
                <span class="rounded-md bg-gray-100 px-2 py-0.5 text-[11px] font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                  {{ scopeLabel(row.ban_scope) }}
                </span>
              </div>
              <div class="mt-1.5 truncate font-mono text-xs text-gray-500 dark:text-gray-400" :title="row.user_agent || '—'">
                {{ row.user_agent || '—' }}
              </div>
            </div>
          </template>

          <template #cell-target="{ row }">
            <div class="min-w-0 max-w-[240px]">
              <div class="truncate text-sm font-medium text-gray-800 dark:text-gray-200" :title="row.target_identifier">
                {{ row.target_identifier || '—' }}
              </div>
              <div class="mt-1 truncate text-xs text-gray-400" :title="reasonLabel(row.reason)">
                {{ reasonLabel(row.reason) }}
              </div>
            </div>
          </template>

          <template #cell-failure_count="{ row }">
            <div class="whitespace-nowrap text-center">
              <div class="text-base font-bold text-gray-900 dark:text-white">{{ row.failure_count }}</div>
              <div v-if="row.ban_count > 1" class="text-[11px] text-gray-400">× {{ row.ban_count }}</div>
            </div>
          </template>

          <template #cell-timeline="{ row }">
            <div class="whitespace-nowrap text-xs text-gray-500 dark:text-gray-400">
              <div>{{ formatDateTime(row.banned_at) }}</div>
              <div class="mt-1 text-gray-400">{{ formatDateTime(row.expires_at) }}</div>
            </div>
          </template>

          <template #cell-status="{ row }">
            <span :class="statusBadgeClass(row.status)">{{ statusLabel(row.status) }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center justify-end gap-3">
              <button type="button" class="inline-flex items-center gap-1 text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400" @click="openDetail(row)">
                <Icon name="eye" size="sm" />
                {{ t('admin.authIPBan.actions.detail') }}
              </button>
              <button
                v-if="row.status === 'active'"
                type="button"
                class="inline-flex items-center gap-1 text-sm font-medium text-red-600 hover:text-red-700 dark:text-red-400"
                @click="openRelease(row)"
              >
                <Icon name="xCircle" size="sm" />
                {{ t('admin.authIPBan.actions.release') }}
              </button>
            </div>
          </template>

          <template #empty>
            <div class="flex flex-col items-center py-10">
              <Icon name="shield" size="xl" class="mb-3 h-12 w-12 text-gray-300 dark:text-dark-600" />
              <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('admin.authIPBan.empty') }}</p>
            </div>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="total > 0"
          :total="total"
          :page="page"
          :page-size="pageSize"
          @update:page="changePage"
          @update:pageSize="changePageSize"
        />
      </template>
    </TablePageLayout>

    <BaseDialog :show="detailVisible" :title="t('admin.authIPBan.detail.title')" width="wide" @close="detailVisible = false">
      <div v-if="detail" class="space-y-5 py-1">
        <div class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-100 pb-4 dark:border-dark-700">
          <div>
            <div class="font-mono text-lg font-bold text-gray-900 dark:text-white">{{ detail.ip_address }}</div>
            <div class="mt-1 flex flex-wrap items-center gap-2">
              <span :class="statusBadgeClass(detail.status)">{{ statusLabel(detail.status) }}</span>
              <span :class="categoryBadgeClass(detail.ua_category)">{{ categoryLabel(detail.ua_category) }}</span>
              <span class="text-xs text-gray-400">{{ scopeLabel(detail.ban_scope) }}</span>
            </div>
          </div>
          <button v-if="detail.status === 'active'" type="button" class="btn btn-danger btn-sm" @click="openRelease(detail)">
            {{ t('admin.authIPBan.actions.release') }}
          </button>
        </div>

        <dl class="grid grid-cols-1 gap-x-6 gap-y-4 sm:grid-cols-2">
          <div v-for="field in detailFields" :key="field.label" class="min-w-0 border-b border-gray-100 pb-3 dark:border-dark-700">
            <dt class="text-xs font-semibold text-gray-400">{{ field.label }}</dt>
            <dd class="mt-1 break-words text-sm text-gray-800 dark:text-gray-200" :class="field.mono ? 'font-mono' : ''">
              {{ field.value || '—' }}
            </dd>
          </div>
        </dl>

        <div>
          <div class="text-xs font-semibold text-gray-400">{{ t('admin.authIPBan.detail.userAgent') }}</div>
          <div class="mt-2 break-all rounded-md bg-gray-50 p-3 font-mono text-xs leading-5 text-gray-600 dark:bg-dark-800 dark:text-gray-300">
            {{ detail.user_agent || '—' }}
          </div>
        </div>
      </div>
    </BaseDialog>

    <ConfirmDialog
      :show="releaseVisible"
      :title="t('admin.authIPBan.actions.releaseTitle')"
      :message="t('admin.authIPBan.actions.releaseMessage', { ip: releasing?.ip_address || '' })"
      :confirm-text="t('admin.authIPBan.actions.release')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmRelease"
      @cancel="releaseVisible = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI, type AuthIPBan, type AuthIPBanPolicy, type AuthUserAgentCategory } from '@/api/admin'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { Column } from '@/components/common/types'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'

const { t, te } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const records = ref<AuthIPBan[]>([])
const policies = ref<AuthIPBanPolicy[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const filters = reactive({ q: '', status: 'active' })

const columns = computed<Column[]>(() => [
  { key: 'source', label: t('admin.authIPBan.columns.source') },
  { key: 'fingerprint', label: t('admin.authIPBan.columns.fingerprint') },
  { key: 'target', label: t('admin.authIPBan.columns.target') },
  { key: 'failure_count', label: t('admin.authIPBan.columns.failures') },
  { key: 'timeline', label: t('admin.authIPBan.columns.timeline') },
  { key: 'status', label: t('admin.authIPBan.columns.status') },
  { key: 'actions', label: t('common.actions') }
])

const statusOptions = computed(() => [
  { value: 'all', label: t('admin.authIPBan.filters.all') },
  { value: 'active', label: t('admin.authIPBan.filters.active') },
  { value: 'expired', label: t('admin.authIPBan.filters.expired') },
  { value: 'released', label: t('admin.authIPBan.filters.released') }
])

async function loadRecords() {
  loading.value = true
  try {
    const [listResult, policyResult] = await Promise.all([
      adminAPI.authIPBans.list({
        page: page.value,
        page_size: pageSize.value,
        status: filters.status as 'all' | 'active' | 'expired' | 'released',
        q: filters.q || undefined
      }),
      policies.value.length ? Promise.resolve(policies.value) : adminAPI.authIPBans.getPolicy()
    ])
    records.value = listResult.items
    total.value = listResult.total
    policies.value = policyResult
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.authIPBan.loadFailed'))
  } finally {
    loading.value = false
  }
}

function search() {
  page.value = 1
  void loadRecords()
}

function resetFilters() {
  filters.q = ''
  filters.status = 'active'
  search()
}

function changePage(value: number) {
  page.value = value
  void loadRecords()
}

function changePageSize(value: number) {
  pageSize.value = value
  page.value = 1
  void loadRecords()
}

function categoryLabel(category: AuthUserAgentCategory): string {
  return t(`admin.authIPBan.category.${category}`)
}

function scopeLabel(scope: string): string {
  return t(`admin.authIPBan.scope.${scope}`)
}

function statusLabel(status: string): string {
  return t(`admin.authIPBan.filters.${status}`)
}

function reasonLabel(reason: string): string {
  const key = `admin.authIPBan.reason.${reason}`
  return reason && te(key) ? t(key) : reason || '—'
}

function durationLabel(minutes: number): string {
  if (minutes >= 60 && minutes % 60 === 0) {
    return t('admin.authIPBan.policy.hours', { count: minutes / 60 })
  }
  return t('admin.authIPBan.policy.minutes', { count: minutes })
}

function policySummary(policy: AuthIPBanPolicy): string {
  return t('admin.authIPBan.policy.summary', {
    threshold: policy.threshold,
    window: policy.window_minutes,
    duration: durationLabel(policy.ban_minutes)
  })
}

function statusBadgeClass(status: string): string {
  const base = 'inline-flex rounded-md px-2 py-1 text-xs font-semibold '
  if (status === 'active') return base + 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  if (status === 'expired') return base + 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
  return base + 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
}

function categoryBadgeClass(category: string): string {
  const base = 'inline-flex rounded-md px-2 py-1 text-xs font-semibold '
  if (category === 'automation' || category === 'empty') {
    return base + 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  }
  if (category === 'browser') {
    return base + 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
  }
  return base + 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
}

const detailVisible = ref(false)
const detail = ref<AuthIPBan | null>(null)

function openDetail(record: AuthIPBan) {
  detail.value = record
  detailVisible.value = true
}

const detailFields = computed(() => {
  if (!detail.value) return []
  const row = detail.value
  return [
    { label: t('admin.authIPBan.detail.target'), value: row.target_identifier },
    { label: t('admin.authIPBan.detail.reason'), value: reasonLabel(row.reason) },
    { label: t('admin.authIPBan.detail.path'), value: row.trigger_path, mono: true },
    { label: t('admin.authIPBan.detail.banCount'), value: String(row.ban_count) },
    { label: t('admin.authIPBan.detail.firstSeen'), value: formatDateTime(row.first_seen_at) },
    { label: t('admin.authIPBan.detail.lastSeen'), value: formatDateTime(row.last_seen_at) },
    { label: t('admin.authIPBan.detail.bannedAt'), value: formatDateTime(row.banned_at) },
    { label: t('admin.authIPBan.detail.expiresAt'), value: formatDateTime(row.expires_at) },
    { label: t('admin.authIPBan.detail.releasedAt'), value: row.released_at ? formatDateTime(row.released_at) : '' },
    { label: t('admin.authIPBan.detail.releasedBy'), value: row.released_by_email },
    { label: t('admin.authIPBan.detail.releaseNote'), value: row.release_note }
  ]
})

const releaseVisible = ref(false)
const releasing = ref<AuthIPBan | null>(null)

function openRelease(record: AuthIPBan) {
  detailVisible.value = false
  releasing.value = record
  releaseVisible.value = true
}

async function confirmRelease() {
  if (!releasing.value) return
  try {
    await adminAPI.authIPBans.release(releasing.value.id)
    releaseVisible.value = false
    appStore.showSuccess(t('admin.authIPBan.actions.releaseSuccess'))
    await loadRecords()
  } catch (error: any) {
    appStore.showError(error?.message || t('admin.authIPBan.actions.releaseFailed'))
  }
}

onMounted(loadRecords)
</script>
