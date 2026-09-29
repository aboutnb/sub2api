<template>
  <AppLayout>
    <div class="space-y-6">
      <header class="flex flex-wrap justify-between gap-4">
        <div><h1 class="text-2xl font-semibold">{{ t('admin.helpDocuments.title') }}</h1><p class="mt-2 text-ink-muted">{{ t('admin.helpDocuments.description') }}</p></div>
        <button type="button" class="btn btn-primary" @click="create">{{ t('common.create') }}</button>
      </header>
      <p v-if="error" role="alert" class="text-red-600">{{ error }}</p>
      <p v-if="loading">{{ t('common.loading') }}</p>
      <div v-else class="space-y-3">
        <article v-for="item in items" :key="item.id" class="card flex flex-wrap items-center justify-between gap-4 p-4">
          <div><h2 class="font-semibold">{{ item.title }}</h2><p class="text-xs text-ink-muted">{{ item.slug }} · {{ item.category }} · v{{ item.version }}</p></div>
          <div class="flex flex-wrap items-center gap-2">
            <span class="badge" :class="item.status === 'published' ? 'badge-success' : 'badge-gray'">{{ t('admin.helpDocuments.' + item.status) }}</span>
            <button type="button" class="btn btn-secondary btn-sm" :disabled="busy" @click="edit(item)">{{ t('common.edit') }}</button>
            <button type="button" class="btn btn-primary btn-sm" :disabled="busy" @click="transition(item, 'publish')">{{ t('admin.helpDocuments.publish') }}</button>
            <button v-if="item.status === 'published'" type="button" class="btn btn-secondary btn-sm" :disabled="busy" @click="transition(item, 'unpublish')">{{ t('admin.helpDocuments.unpublish') }}</button>
            <button v-if="item.can_rollback" type="button" class="btn btn-secondary btn-sm" :disabled="busy" @click="transition(item, 'rollback')">{{ t('admin.helpDocuments.rollback') }}</button>
            <button type="button" class="btn btn-secondary btn-sm" :disabled="busy" @click="deleting = item">{{ t('common.delete') }}</button>
          </div>
        </article>
        <p v-if="!items.length" class="py-12 text-center text-ink-muted">{{ t('helpCenter.empty') }}</p>
      </div>
      <div v-if="pages > 1" class="flex items-center justify-end gap-3">
        <button class="btn btn-secondary" :disabled="page === 1 || loading" @click="page--; load()">{{ t('admin.helpDocuments.previous') }}</button>
        <span>{{ page }} / {{ pages }}</span>
        <button class="btn btn-secondary" :disabled="page === pages || loading" @click="page++; load()">{{ t('admin.helpDocuments.next') }}</button>
      </div>
    </div>
    <BaseDialog :show="editing !== null" :title="t('admin.helpDocuments.editor')" width="extra-wide" :pending="busy" @close="editing = null">
      <form class="space-y-4" @submit.prevent="save">
        <p v-if="editorError" role="alert" class="text-red-600">{{ editorError }}</p>
        <div class="grid gap-4 sm:grid-cols-2">
          <label class="block">Slug<input v-model="form.slug" :readonly="Boolean(editing?.id)" class="input mt-1 w-full" required maxlength="160" pattern="[a-z0-9]+([-/][a-z0-9]+)*"></label>
          <label class="block">{{ t('admin.helpDocuments.articleTitle') }}<input v-model="form.title" class="input mt-1 w-full" required maxlength="240"></label>
        </div>
        <label class="block">{{ t('admin.helpDocuments.category') }}<select v-model="form.category" class="input mt-1 w-full"><option v-for="category in categories" :key="category" :value="category">{{ category }}</option></select></label>
        <label class="block">{{ t('admin.helpDocuments.summary') }}<textarea v-model="form.summary" class="input mt-1 w-full" maxlength="1000" /></label>
        <label class="block">{{ t('admin.helpDocuments.selectors') }}<textarea v-model="schemaText" class="input mt-1 w-full font-mono" rows="3" /><small class="text-ink-muted">{{ t('admin.helpDocuments.selectorHint') }}</small></label>
        <label class="block">Markdown<textarea v-model="form.content_markdown" class="input mt-1 min-h-72 w-full font-mono" maxlength="200000" required /></label>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" :disabled="busy" @click="preview">{{ t('admin.helpDocuments.preview') }}</button>
          <button type="submit" class="btn btn-primary" :disabled="busy">{{ t('admin.helpDocuments.saveDraft') }}</button>
        </div>
      </form>
      <section v-if="previewContent !== null" class="mt-6 border-t border-line pt-4"><h2 class="mb-3 font-semibold">{{ t('admin.helpDocuments.preview') }}</h2><HelpMarkdown :content="previewContent" /></section>
    </BaseDialog>
    <BaseDialog :show="deleting !== null" :title="t('common.delete')" :pending="busy" @close="deleting = null">
      <p>{{ t('admin.helpDocuments.deleteConfirm') }}</p>
      <template #footer><button class="btn btn-secondary" @click="deleting = null">{{ t('common.cancel') }}</button><button class="btn btn-danger" :disabled="busy" @click="remove">{{ t('common.delete') }}</button></template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import HelpMarkdown from '@/components/help/HelpMarkdown.vue'
import { apiClient } from '@/api/client'
import type { HelpDocument } from '@/api/helpDocs'
const { t } = useI18n()
const categories = ['quick-start', 'claude-code', 'codex', 'opencode', 'gemini-cli', 'grok-cli', 'cherry-studio', 'cursor', 'cline', 'roo-code', 'dsh', 'pi', 'openclaw', 'hermes', 'workbuddy', 'zcode', 'trae', 'read-frog', 'kiss-translator', 'immersive-translate', 'sillytavern', 'tavernai', 'api', 'faq']
const items = ref<HelpDocument[]>([])
const loading = ref(false)
const busy = ref(false)
const error = ref('')
const editorError = ref('')
const page = ref(1)
const pages = ref(1)
const editing = ref<HelpDocument | null>(null)
const deleting = ref<HelpDocument | null>(null)
const previewContent = ref<string | null>(null)
const schemaText = ref('{}')
const form = reactive({ slug: '', title: '', category: 'quick-start', summary: '', content_markdown: '', version: 1 })
function message(e: unknown) { return e instanceof Error ? e.message : (e as { message?: string })?.message || t('admin.helpDocuments.failed') }
async function load() {
  loading.value = true
  error.value = ''
  try {
    const { data } = await apiClient.get<{ items: HelpDocument[]; pages: number }>('/admin/help-docs', { params: { page: page.value, page_size: 20 } })
    items.value = data.items
    pages.value = data.pages
  } catch (e) { error.value = message(e) } finally { loading.value = false }
}
function edit(item: HelpDocument) {
  editing.value = item
  Object.assign(form, { slug: item.slug, title: item.title, category: item.category, summary: item.summary, content_markdown: item.content_markdown, version: item.version })
  schemaText.value = JSON.stringify(item.selector_schema || {}, null, 2)
  previewContent.value = null
  editorError.value = ''
}
function create() {
  edit({ id: 0, slug: '', title: '', category: 'quick-start', summary: '', content_markdown: '', selector_schema: {}, status: 'draft', version: 1, updated_at: '' })
}
function payload() {
  const schema = JSON.parse(schemaText.value)
  const allowed = ['platforms', 'clients', 'systems', 'shells', 'protocols', 'versions']
  if (!schema || Array.isArray(schema) || typeof schema !== 'object' || Object.entries(schema).some(([key, values]) => !allowed.includes(key) || !Array.isArray(values) || values.some(v => typeof v !== 'string'))) throw new Error(t('admin.helpDocuments.invalidSchema'))
  return { ...form, selector_schema: schema }
}
async function save() {
  busy.value = true
  editorError.value = ''
  try {
    const body = payload()
    if (editing.value?.id) await apiClient.put('/admin/help-docs/' + editing.value.id, body)
    else await apiClient.post('/admin/help-docs', body)
    editing.value = null
    await load()
  } catch (e) { editorError.value = message(e) } finally { busy.value = false }
}
async function preview() {
  busy.value = true
  editorError.value = ''
  try {
    const body = payload()
    if (editing.value?.id) {
      const { data } = await apiClient.post<HelpDocument>('/admin/help-docs/' + editing.value.id + '/preview', body)
      previewContent.value = data.content_markdown
    } else previewContent.value = body.content_markdown
  } catch (e) { editorError.value = message(e) } finally { busy.value = false }
}
async function transition(item: HelpDocument, action: 'publish' | 'unpublish' | 'rollback') {
  busy.value = true
  error.value = ''
  try {
    await apiClient.post('/admin/help-docs/' + item.id + '/' + action, { version: item.version })
    await load()
  } catch (e) { error.value = message(e) } finally { busy.value = false }
}
async function remove() {
  if (!deleting.value) return
  busy.value = true
  try {
    await apiClient.delete('/admin/help-docs/' + deleting.value.id)
    deleting.value = null
    await load()
  } catch (e) { error.value = message(e) } finally { busy.value = false }
}
watch([() => form.content_markdown, schemaText], () => { previewContent.value = null })
onMounted(load)
</script>
