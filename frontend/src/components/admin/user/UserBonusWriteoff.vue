<template>
  <section class="space-y-3 rounded-xl border border-default p-4">
    <h3 class="font-medium">{{ t('admin.users.bonusWriteoff.title') }}</h3>
    <p class="input-hint">{{ t('admin.users.bonusWriteoff.hint') }}</p>
    <p v-if="error" role="alert" class="text-red-600">{{ error }}</p>
    <fieldset :disabled="loading || saving" class="space-y-3">
      <label class="flex items-center gap-3">
        <input v-model="form.enabled" type="checkbox" role="switch" />
        <span>{{ t('admin.users.bonusWriteoff.enabled') }}</span>
      </label>
      <label class="block">
        <span class="input-label">{{ t('admin.users.bonusWriteoff.effectiveAt') }}</span>
        <input v-model="form.effectiveAt" class="input" type="datetime-local" />
      </label>
      <label class="block">
        <span class="input-label">{{ t('admin.users.bonusWriteoff.limit') }}</span>
        <input v-model.number="form.amount_limit" class="input" type="number" min="0" max="1000000000" step="0.00000001" />
      </label>
      <label class="block">
        <span class="input-label">{{ t('admin.users.bonusWriteoff.consent') }}</span>
        <textarea v-model="form.consent_note" class="input" maxlength="2000" rows="2" />
      </label>
      <p class="input-hint">{{ t('admin.users.bonusWriteoff.revisionHint') }}</p>
      <button type="button" class="btn btn-secondary" :disabled="!loaded || saving" @click="save">
        {{ t('admin.users.bonusWriteoff.save') }}
      </button>
    </fieldset>
    <details>
      <summary class="cursor-pointer">{{ t('admin.users.bonusWriteoff.history') }} ({{ total }})</summary>
      <div class="mt-3 max-h-64 overflow-auto">
        <table class="w-full text-sm">
          <thead><tr><th>{{ t('admin.users.bonusWriteoff.time') }}</th><th>{{ t('admin.users.bonusWriteoff.action') }}</th><th>{{ t('admin.users.bonusWriteoff.amount') }}</th><th>{{ t('admin.users.bonusWriteoff.details') }}</th></tr></thead>
          <tbody><tr v-for="item in items" :key="item.id">
            <td>{{ new Date(item.created_at).toLocaleString() }}</td>
            <td>{{ t(`admin.users.bonusWriteoff.actions.${item.action}`) }}</td>
            <td>{{ item.amount }}</td>
            <td class="min-w-48 break-all p-2">
              <details><summary>{{ t('admin.users.bonusWriteoff.details') }}</summary>
                <div>{{ t('admin.users.bonusWriteoff.batch') }}: {{ item.batch_id ?? '-' }}</div>
                <div>{{ t('admin.users.bonusWriteoff.request') }}: {{ item.request_id || '-' }}</div>
                <div>{{ t('admin.users.bonusWriteoff.actor') }}: {{ item.actor_id ?? '-' }}</div>
                <div>{{ t('admin.users.bonusWriteoff.balance') }}: {{ item.balance_after ?? '-' }}</div>
                <pre class="whitespace-pre-wrap text-xs">{{ JSON.stringify(item.detail, null, 2) }}</pre>
              </details>
            </td>
          </tr></tbody>
        </table>
      </div>
      <div class="mt-2 flex justify-between">
        <button type="button" class="btn btn-secondary" :disabled="page === 1 || loading" @click="load(page - 1)">{{ t('admin.users.bonusWriteoff.previous') }}</button>
        <button type="button" class="btn btn-secondary" :disabled="page * 50 >= total || loading" @click="load(page + 1)">{{ t('admin.users.bonusWriteoff.next') }}</button>
      </div>
    </details>
  </section>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '@/api/client'
import { useAppStore } from '@/stores/app'

const props = defineProps<{ userId: number }>()
const { t } = useI18n()
const app = useAppStore()
const form = reactive({ enabled: false, effectiveAt: '', amount_limit: 0, consent_note: '' })
const loading = ref(false), saving = ref(false), loaded = ref(false), error = ref('')
const page = ref(1), total = ref(0)
type Audit = { id: number; action: string; amount: number; created_at: string; batch_id: number | null; request_id: string; actor_id: number | null; balance_after: number | null; detail: Record<string, unknown> }
const items = ref<Audit[]>([])
let generation = 0
async function load(targetPage = 1, resetForm = false) {
  const current = ++generation, userId = props.userId
  loading.value = true
  error.value = ''
  try {
    const { data } = await apiClient.get(`/admin/users/${userId}/bonus-writeoff`, { params: { page: targetPage } })
    if (current !== generation || userId !== props.userId) return
    items.value = data.items; total.value = data.total; page.value = targetPage
    if (resetForm) {
      const policy = data.policy
      form.enabled = policy?.enabled ?? false
      form.amount_limit = policy?.amount_limit ?? 0
      form.consent_note = policy?.consent_note ?? ''
      const date = policy?.effective_at ? new Date(policy.effective_at) : new Date()
      form.effectiveAt = new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16)
    }
    loaded.value = true
  } catch {
    if (current === generation) { error.value = t('admin.users.bonusWriteoff.loadError'); loaded.value = false }
  } finally { if (current === generation) loading.value = false }
}
async function save() {
  if (!loaded.value || saving.value) return
  const userId = props.userId
  if ((form.enabled && !form.consent_note.trim()) || !Number.isFinite(form.amount_limit) || form.amount_limit < 0 || form.amount_limit > 1e9 || !Number.isFinite(Date.parse(form.effectiveAt))) {
    error.value = t('admin.users.bonusWriteoff.invalid'); return
  }
  saving.value = true; error.value = ''
  try {
    await apiClient.put(`/admin/users/${userId}/bonus-writeoff`, {
      enabled: form.enabled, effective_at: new Date(form.effectiveAt).toISOString(),
      amount_limit: form.amount_limit, consent_note: form.consent_note.trim()
    })
    if (userId !== props.userId) return
    app.showSuccess(t('admin.users.bonusWriteoff.saved'))
    await load(1, true)
  } catch { if (userId === props.userId) error.value = t('admin.users.bonusWriteoff.saveError') }
  finally { saving.value = false }
}
watch(() => props.userId, () => { loaded.value = false; items.value = []; void load(1, true) }, { immediate: true })
</script>
