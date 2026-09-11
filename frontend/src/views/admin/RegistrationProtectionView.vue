<template>
  <AppLayout>
    <section class="space-y-5">
      <div class="card p-4 sm:p-6">
        <p class="text-sm text-ink-muted">{{ t('admin.registrationProtection.description') }}</p>
        <nav class="mt-4 flex flex-wrap gap-2" :aria-label="t('admin.registrationProtection.title')">
          <button v-for="item in tabs" :key="item" class="btn" :class="tab === item ? 'btn-primary' : 'btn-secondary'"
            :aria-current="tab === item ? 'page' : undefined" @click="tab = item">{{ t(`admin.registrationProtection.tabs.${item}`) }}</button>
        </nav>
      </div>

      <div v-if="tab === 'settings'" class="card p-4 sm:p-6">
        <p v-if="settingsLoading" role="status">{{ t('admin.registrationProtection.loading') }}</p>
        <form v-else-if="settings" class="space-y-6" @submit.prevent="saveSettings">
          <label class="flex items-center gap-3 font-semibold text-ink-strong">
            <input v-model="settings.enabled" type="checkbox" :disabled="saving" class="h-4 w-4" />
            {{ t('admin.registrationProtection.enabled') }}
          </label>
          <p class="text-sm text-ink-muted">{{ t('admin.registrationProtection.quotaHint') }}</p>
          <div class="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
            <label v-for="field in fields" :key="field.key" class="block">
              <span class="input-label">{{ t(`admin.registrationProtection.fields.${field.key}`) }}</span>
              <input v-model.number="settings[field.key]" type="number" class="input" min="1" :max="field.max" step="1" required :disabled="saving" />
            </label>
          </div>
          <p class="text-sm text-ink-muted">{{ t('admin.registrationProtection.observationHint') }}</p>
          <button class="btn btn-primary" type="submit" :disabled="saving">{{ t('admin.registrationProtection.save') }}</button>
        </form>
        <button v-else class="btn btn-secondary" @click="loadSettings">{{ t('admin.registrationProtection.retry') }}</button>
      </div>

      <div v-else class="card overflow-hidden">
        <form class="flex flex-wrap items-end gap-3 border-b border-line p-4" @submit.prevent="search">
          <label v-if="tab === 'sources'" class="min-w-40">
            <span class="input-label">{{ t('admin.registrationProtection.recordType') }}</span>
            <select v-model="sourceKind" class="input">
              <option v-for="kind in sourceKinds" :key="kind" :value="kind">{{ t(`admin.registrationProtection.kinds.${kind}`) }}</option>
            </select>
          </label>
          <label class="min-w-48 flex-1">
            <span class="input-label">{{ t('admin.registrationProtection.search') }}</span>
            <input v-model.trim="query" class="input" type="search" maxlength="256" :placeholder="t('admin.registrationProtection.searchHint')" />
          </label>
          <label v-if="statusOptions.length" class="min-w-36">
            <span class="input-label">{{ t('admin.registrationProtection.status') }}</span>
            <select v-model="status" class="input" @change="search">
              <option value="">{{ t('admin.registrationProtection.all') }}</option>
              <option v-for="value in statusOptions" :key="value" :value="value">{{ statusLabel(value) }}</option>
            </select>
          </label>
          <button type="submit" class="btn btn-primary" :disabled="loading">{{ t('admin.registrationProtection.search') }}</button>
        </form>
        <p v-if="loading" class="p-6" role="status">{{ t('admin.registrationProtection.loading') }}</p>
        <p v-else-if="!records.length" class="p-8 text-center text-ink-muted">{{ t('admin.registrationProtection.empty') }}</p>
        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-[800px] text-left text-sm">
            <thead class="border-b border-line bg-surface-muted text-ink-muted"><tr>
              <th class="p-4">{{ t('admin.registrationProtection.source') }}</th>
              <th class="p-4">{{ t('admin.registrationProtection.account') }}</th>
              <th class="p-4">{{ t('admin.registrationProtection.reason') }}</th>
              <th class="p-4">{{ t('admin.registrationProtection.created') }}</th>
              <th class="p-4">{{ t('admin.registrationProtection.actions') }}</th>
            </tr></thead>
            <tbody class="divide-y divide-line">
              <tr v-for="row in records" :key="row.id">
                <td class="p-4"><p class="font-mono">{{ row.ip_address || '—' }}</p><p class="mt-1 max-w-56 truncate text-xs text-ink-muted" :title="row.user_agent">{{ row.user_agent || '—' }}</p></td>
                <td class="p-4"><p>{{ row.email || (row.user_id ? `#${row.user_id}` : '—') }}</p><p v-if="row.concurrency !== undefined" class="mt-1 text-xs text-ink-muted">{{ t('admin.registrationProtection.concurrency') }}: {{ row.concurrency }}</p></td>
                <td class="p-4"><p>{{ reasonLabel(row.reason) }}</p><span v-if="row.status" class="mt-1 inline-block rounded border border-line px-2 py-0.5 text-xs">{{ statusLabel(row.status) }}</span></td>
                <td class="whitespace-nowrap p-4 text-ink-muted">{{ formatTime(row.created_at) }}</td>
                <td class="p-4"><div class="flex gap-2">
                  <button type="button" class="btn btn-secondary btn-sm" @click="detail = row">{{ t('admin.registrationProtection.details') }}</button>
                  <button v-if="tab === 'accounts'" type="button" class="btn btn-secondary btn-sm" @click="openReview(row)">{{ t(row.status === 'restricted' ? 'admin.registrationProtection.release' : 'admin.registrationProtection.restrict') }}</button>
                  <button v-if="kind === 'blocks' && row.status === 'active'" type="button" class="btn btn-secondary btn-sm" @click="openReview(row, true)">{{ t('admin.registrationProtection.release') }}</button>
                </div></td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="flex items-center justify-between gap-3 border-t border-line p-4">
          <span class="text-sm text-ink-muted">{{ t('admin.registrationProtection.total', { total, page }) }}</span>
          <div class="flex gap-2"><button class="btn btn-secondary" :disabled="loading || page <= 1" @click="changePage(-1)">{{ t('admin.registrationProtection.previous') }}</button><button class="btn btn-secondary" :disabled="loading || page * 20 >= total" @click="changePage(1)">{{ t('admin.registrationProtection.next') }}</button></div>
        </div>
      </div>
      <p v-if="error" role="alert" class="rounded-lg border border-red-300 bg-red-50 p-4 text-sm text-red-800 dark:bg-red-950 dark:text-red-200">{{ error }}</p>
    </section>
    <BaseDialog :show="!!detail" :title="t('admin.registrationProtection.details')" @close="detail = null">
      <dl v-if="detail" class="space-y-4 break-words">
        <div v-for="field in detailFields" :key="field"><dt class="text-xs text-ink-muted">{{ t(`admin.registrationProtection.detailFields.${field}`) }}</dt><dd class="mt-1 whitespace-pre-wrap">{{ detailValue(field) }}</dd></div>
      </dl>
    </BaseDialog>
    <BaseDialog :show="!!reviewRow" :pending="reviewing" :title="t('admin.registrationProtection.review')" @close="reviewRow = null">
      <p class="mb-4 text-sm text-ink-muted">{{ t(reviewBlock ? 'admin.registrationProtection.releaseSourceHint' : 'admin.registrationProtection.reviewHint') }}</p>
      <p class="mb-3 font-semibold">{{ reviewRow?.email }}</p>
      <label class="block"><span class="input-label">{{ t('admin.registrationProtection.note') }}</span><textarea v-model="reviewNote" class="input min-h-24" maxlength="512" :disabled="reviewing" /></label>
      <p v-if="reviewError" role="alert" class="mt-3 text-sm text-red-600">{{ reviewError }}</p>
      <template #footer><button class="btn btn-secondary" :disabled="reviewing" @click="reviewRow = null">{{ t('admin.registrationProtection.cancel') }}</button><button class="btn btn-primary" :disabled="reviewing || !reviewNote.trim()" @click="submitReview">{{ t('admin.registrationProtection.confirm') }}</button></template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { useAppStore } from '@/stores/app'
import { registrationProtectionAPI as api, type RegistrationProtectionSettings, type RegistrationRecordKind, type RegistrationRiskRecord } from '@/api/admin/registrationProtection'

const { t, te } = useI18n()
const app = useAppStore()
const tabs = ['settings', 'sources', 'accounts'] as const
const tab = ref<typeof tabs[number]>('settings')
const sourceKinds = ['sources', 'events', 'blocks'] as const
const sourceKind = ref<typeof sourceKinds[number]>('sources')
const kind = computed<RegistrationRecordKind>(() => tab.value === 'accounts' ? 'accounts' : sourceKind.value)
const settings = ref<RegistrationProtectionSettings | null>(null)
const settingsLoading = ref(false), saving = ref(false), loading = ref(false), reviewing = ref(false)
const records = ref<RegistrationRiskRecord[]>([]), total = ref(0), page = ref(1)
const query = ref(''), status = ref(''), error = ref(''), reviewError = ref(''), reviewNote = ref('')
const detail = ref<RegistrationRiskRecord | null>(null), reviewRow = ref<RegistrationRiskRecord | null>(null)
const reviewBlock = ref(false)
const statusOptions = computed(() => kind.value === 'accounts' ? ['observed', 'restricted', 'released'] : kind.value === 'blocks' ? ['active', 'expired', 'released'] : [])
type NumericSetting = Exclude<keyof RegistrationProtectionSettings, 'enabled'>
const fields: { key: NumericSetting; max: number }[] = [
  { key: 'identity_success_limit', max: 100000 }, { key: 'ip_success_limit', max: 100000 }, { key: 'success_window_hours', max: 720 },
  { key: 'identity_failure_limit', max: 100000 }, { key: 'ip_failure_limit', max: 100000 }, { key: 'failure_window_minutes', max: 1440 },
  { key: 'block_minutes', max: 10080 }, { key: 'observe_identity_success_limit', max: 100000 }
]
const detailFields = ['ip_address', 'user_agent', 'email', 'user_id', 'reason', 'status', 'concurrency', 'previous_concurrency', 'trigger_path', 'created_at', 'expires_at', 'reviewed_at', 'reviewed_by', 'review_note', 'release_note'] as const
let requestVersion = 0
function message(e: unknown) { return e instanceof Error ? e.message : t('admin.registrationProtection.failed') }
function statusLabel(value: string) { const key = `admin.registrationProtection.statuses.${value}`; return te(key) ? t(key) : value }
function reasonLabel(value?: string) { const key = `admin.registrationProtection.reasons.${value}`; return value ? (te(key) ? t(key) : value) : '—' }
function formatTime(value?: string) { return value ? new Date(value).toLocaleString() : '—' }
function detailValue(field: typeof detailFields[number]) {
  const value = detail.value?.[field]
  if (field === 'reason') return reasonLabel(String(value || ''))
  if (field === 'status') return statusLabel(String(value || ''))
  if (field.endsWith('_at')) return formatTime(value as string | undefined)
  return value ?? '—'
}
async function loadSettings() {
  settingsLoading.value = true; error.value = ''
  try { settings.value = await api.settings() } catch (e) { error.value = message(e) } finally { settingsLoading.value = false }
}
async function saveSettings() {
  if (!settings.value || saving.value) return
  saving.value = true; error.value = ''
  try { settings.value = await api.updateSettings({ ...settings.value }); app.showSuccess(t('admin.registrationProtection.saved')) } catch (e) { error.value = message(e) } finally { saving.value = false }
}
async function loadRecords() {
  const version = ++requestVersion
  loading.value = true; error.value = ''
  try {
    const result = await api.list(kind.value, { page: page.value, page_size: 20, q: query.value, status: status.value })
    if (version !== requestVersion) return
    records.value = result.items; total.value = result.total
  } catch (e) { if (version === requestVersion) { records.value = []; total.value = 0; error.value = message(e) } }
  finally { if (version === requestVersion) loading.value = false }
}
function search() { page.value = 1; void loadRecords() }
function changePage(delta: number) { page.value += delta; void loadRecords() }
function openReview(row: RegistrationRiskRecord, block = false) { reviewBlock.value = block; reviewRow.value = row; reviewNote.value = ''; reviewError.value = '' }
async function submitReview() {
  if (!reviewRow.value || reviewing.value || !reviewNote.value.trim()) return
  reviewing.value = true; reviewError.value = ''
  try {
    if (reviewBlock.value) await api.releaseBlock(reviewRow.value.id, reviewNote.value.trim())
    else await api.review(reviewRow.value.id, reviewRow.value.status === 'restricted' ? 'release' : 'restrict', reviewNote.value.trim())
    reviewRow.value = null; app.showSuccess(t('admin.registrationProtection.reviewed')); await loadRecords()
  } catch (e) { reviewError.value = message(e) } finally { reviewing.value = false }
}
watch([tab, sourceKind], () => { status.value = ''; records.value = []; total.value = 0; ++requestVersion; if (tab.value !== 'settings') search() })
onMounted(loadSettings)
onUnmounted(() => { ++requestVersion })
</script>
