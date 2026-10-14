<template>
  <AppLayout>
    <section ref="workspaceRoot" class="tickets-workspace" :class="{ 'show-conversation': mobileDetail }">
      <div class="ticket-toolbar">
        <nav class="ticket-tabs" :aria-label="t(admin ? 'supportTickets.queue' : 'supportTickets.status')">
          <button v-for="tab in tabs" :key="tab.value" type="button" :aria-pressed="activeTab === tab.value" :class="{ active: activeTab === tab.value }" @click="changeTab(tab.value)">
            {{ tab.label }}
          </button>
        </nav>
        <div class="flex shrink-0 items-center gap-2">
          <button class="btn btn-ghost ticket-icon-button" :disabled="loading || busy || detailLoading" :aria-label="t('common.refresh')" :title="t('common.refresh')" @click="refreshTickets">
            <Icon name="refresh" size="md" :class="{ 'animate-spin': loading }" />
          </button>
          <button v-if="!admin" class="btn btn-primary gap-1.5" :disabled="busy" @click="openCreate">
            <Icon name="plus" size="md" /><span>{{ t('supportTickets.create') }}</span>
          </button>
        </div>
      </div>

      <div class="ticket-panels">
        <aside class="ticket-sidebar" :aria-label="t('supportTickets.ticketList')">
          <div class="ticket-filters">
            <form class="relative" role="search" @submit.prevent="submitSearch">
              <Icon name="search" size="md" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-muted" />
              <input v-model="search" class="input !pl-10 !pr-11" :aria-label="t('common.search')" :placeholder="t(admin ? 'supportTickets.searchPlaceholder' : 'supportTickets.userSearchPlaceholder')" />
              <button type="submit" class="ticket-search-submit" :aria-label="t('common.search')"><Icon name="arrowRight" size="sm" /></button>
            </form>
            <TicketCategoryMenu v-model="category" @change="resetPage" />
          </div>

          <div class="ticket-list" :aria-busy="loading">
            <p v-if="loading && !tickets.length" class="ticket-placeholder" role="status">{{ t('common.loading') }}</p>
            <div v-else-if="listError" class="ticket-empty" role="alert">
              <Icon name="exclamationTriangle" size="xl" class="text-ink-muted" />
              <p>{{ listError }}</p>
              <button class="btn btn-secondary" @click="loadTickets">{{ t('common.refresh') }}</button>
            </div>
            <ul v-else-if="tickets.length" class="divide-y divide-line">
              <li v-for="ticket in tickets" :key="ticket.id">
                <button class="ticket-list-item" :class="{ selected: selectedId === ticket.id && !createOpen }" :aria-pressed="selectedId === ticket.id && !createOpen" :disabled="busy" :data-ticket-id="ticket.id" @click="openTicket(ticket)">
                  <span class="flex min-w-0 items-center justify-between gap-3">
                    <span class="truncate text-sm font-semibold leading-5 text-ink-strong">{{ ticket.subject }}</span>
                    <span :class="['ticket-status', statusColor(ticket.status)]"><span class="status-dot" />{{ statusLabels[ticket.status] }}</span>
                  </span>
                  <span class="mt-1.5 block truncate text-sm leading-5 text-ink-muted">{{ ticket.preview }}</span>
                  <span v-if="admin" class="mt-1.5 block truncate text-xs leading-4 text-ink-muted">{{ ticket.user_email || `#${ticket.user_id}` }}</span>
                  <span class="mt-1.5 flex items-center justify-between gap-2 text-xs leading-4 text-ink-muted">
                    <span class="flex min-w-0 items-center gap-2"><span class="shrink-0">#{{ ticket.id }}</span><span class="truncate">{{ categoryLabels[ticket.category] }}</span></span>
                    <time class="shrink-0" :datetime="ticket.last_message_at">{{ shortDate(ticket.last_message_at) }}</time>
                  </span>
                </button>
              </li>
            </ul>
            <div v-else class="ticket-empty">
              <Icon name="inbox" size="xl" class="text-ink-muted" />
              <p class="font-medium text-ink-strong">{{ t('supportTickets.emptyTitle') }}</p>
              <p class="text-sm">{{ t(hasFilters || admin ? 'supportTickets.adminEmptyDescription' : 'supportTickets.emptyDescription') }}</p>
              <button v-if="hasFilters" class="btn btn-ghost text-primary-600" @click="clearFilters">{{ t('supportTickets.clearFilters') }}</button>
              <button v-else-if="!admin" class="btn btn-secondary" :disabled="busy" @click="openCreate">{{ t('supportTickets.create') }}</button>
            </div>
          </div>

          <nav v-if="pagination.total > 0" class="ticket-pagination" aria-label="Pagination">
            <div class="flex min-w-0 flex-wrap gap-x-2 gap-y-1 text-xs leading-4 text-ink-muted">
              <span>{{ t('pagination.pageOf', { page: pagination.page, total: Math.max(1, pagination.pages) }) }}</span>
              <span>{{ t('supportTickets.ticketCount', { count: pagination.total }) }}</span>
            </div>
            <div class="flex shrink-0 gap-2">
              <button class="btn btn-ghost ticket-icon-button" :disabled="loading || pagination.page === 1" :aria-label="t('pagination.previous')" @click="changePage(pagination.page - 1)"><Icon name="chevronLeft" size="md" /></button>
              <button class="btn btn-ghost ticket-icon-button" :disabled="loading || pagination.page >= pagination.pages" :aria-label="t('pagination.next')" @click="changePage(pagination.page + 1)"><Icon name="chevronRight" size="md" /></button>
            </div>
          </nav>
        </aside>

        <section class="ticket-conversation" :aria-label="t('supportTickets.conversation')">
          <template v-if="createOpen">
            <header class="conversation-header">
              <div>
                <h2 class="text-lg font-semibold text-ink-strong">{{ t('supportTickets.create') }}</h2>
                <p class="mt-1 text-sm text-ink-muted">{{ t('supportTickets.createDescription') }}</p>
              </div>
              <button class="btn btn-ghost ticket-icon-button" :disabled="busy" :aria-label="t('common.cancel')" @click="cancelCreate"><Icon name="x" size="md" /></button>
            </header>
            <form id="ticket-create-form" class="ticket-create-form" @submit.prevent="submitCreate">
              <div class="ticket-create-fields">
                <div>
                  <label for="ticket-subject" class="input-label">{{ t('supportTickets.subject') }}</label>
                  <input id="ticket-subject" v-model="draft.subject" class="input" :placeholder="t('supportTickets.subjectPlaceholder')" :disabled="busy" required />
                  <p class="input-hint text-right">{{ characterCount(draft.subject) }}/80</p>
                </div>
                <div>
                  <p class="input-label">{{ t('supportTickets.category') }}</p>
                  <TicketCategoryMenu v-model="draft.category" :include-all="false" :disabled="busy" />
                </div>
                <div class="flex min-h-0 flex-1 flex-col">
                  <label for="ticket-body" class="input-label">{{ t('supportTickets.body') }}</label>
                  <textarea id="ticket-body" v-model="draft.body" class="input min-h-40 flex-1 resize-y" rows="8" :placeholder="t('supportTickets.bodyPlaceholder')" :disabled="busy" required />
                  <p class="input-hint text-right">{{ characterCount(draft.body) }}/4000</p>
                </div>
                <p class="input-hint">{{ t('supportTickets.limitHint') }}</p>
              </div>
              <footer class="conversation-footer flex justify-end gap-2">
                <button type="button" class="btn btn-secondary" :disabled="busy" @click="cancelCreate">{{ t('common.cancel') }}</button>
                <button type="submit" class="btn btn-primary" :disabled="busy || !canCreate">{{ t('common.submit') }}</button>
              </footer>
            </form>
          </template>

          <template v-else-if="selectedId !== null">
            <header class="conversation-header">
              <div class="min-w-0 flex-1">
                <button class="ticket-back" :disabled="busy" @click="closeDetail"><Icon name="arrowLeft" size="sm" />{{ t('supportTickets.backToList') }}</button>
                <template v-if="selected">
                  <div class="flex flex-wrap items-center gap-x-3 gap-y-1">
                    <h2 ref="conversationHeading" class="min-w-0 max-w-full break-words text-base font-semibold leading-6 text-ink-strong outline-none" tabindex="-1">{{ selected.subject }}</h2>
                    <span class="flex items-center gap-2 text-xs leading-4 text-ink-muted">
                      <span>#{{ selected.id }}</span>
                      <span :class="['ticket-status', statusColor(selected.status)]"><span class="status-dot" />{{ statusLabels[selected.status] }}</span>
                    </span>
                  </div>
                  <div class="mt-1.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs leading-4 text-ink-muted">
                    <span>{{ categoryLabels[selected.category] }}</span>
                    <span v-if="admin" class="break-all">{{ selected.user_email || `#${selected.user_id}` }}</span>
                    <time :datetime="selected.created_at">{{ formatDateTime(selected.created_at) }}</time>
                  </div>
                </template>
              </div>
              <button v-if="!admin && selected && selected.status !== 'closed'" class="btn btn-secondary shrink-0" :disabled="busy || detailLoading" @click="closeTicket">{{ t('supportTickets.closeTicket') }}</button>
            </header>
            <p v-if="detailLoading" class="ticket-placeholder flex-1" role="status">{{ t('common.loading') }}</p>
            <div v-else-if="detailError" class="ticket-empty flex-1" role="alert">
              <p>{{ detailError }}</p>
              <button class="btn btn-secondary" @click="retryDetail">{{ t('common.refresh') }}</button>
            </div>
            <template v-else-if="selected">
              <div v-if="admin" class="ticket-status-editor">
                <span class="text-xs text-ink-muted">{{ t('supportTickets.status') }}</span>
                <Select v-model="editStatus" class="min-w-0 flex-1 sm:max-w-52" :options="statusOptions" :aria-label="t('supportTickets.status')" :disabled="busy" />
                <button class="btn btn-secondary" :disabled="busy || editStatus === selected.status" @click="saveStatus">{{ t('common.save') }}</button>
              </div>
              <div ref="messagesViewport" class="ticket-messages" aria-live="polite" :aria-label="t('supportTickets.conversation')" tabindex="0">
                <article v-for="message in selected.messages" :key="message.id" class="ticket-message" :class="{ 'from-user': message.author_role === 'user' }">
                  <div class="message-avatar" aria-hidden="true"><Icon :name="message.author_role === 'admin' ? 'shield' : 'user'" size="md" /></div>
                  <div class="message-bubble">
                    <div class="mb-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs leading-4 text-ink-muted">
                      <span class="font-semibold text-ink-strong">{{ message.author_role === 'admin' ? t('supportTickets.staff') : t(admin ? 'supportTickets.user' : 'supportTickets.me') }}</span>
                      <time :datetime="message.created_at">{{ formatDateTime(message.created_at) }}</time>
                    </div>
                    <p class="whitespace-pre-wrap break-words text-sm leading-6 text-ink">{{ message.body }}</p>
                  </div>
                </article>
              </div>
              <div class="conversation-footer">
                <p v-if="selected.status === 'closed'" class="mb-3 text-sm text-ink-muted">{{ t(admin ? 'supportTickets.adminClosedHint' : 'supportTickets.closedHint') }}</p>
                <form v-if="admin || selected.status !== 'closed'" id="ticket-reply-form" @submit.prevent="submitReply">
                  <label for="ticket-reply" class="sr-only">{{ t('supportTickets.replyLabel') }}</label>
                  <textarea id="ticket-reply" v-model="replyBody" class="ticket-reply-input" rows="2" :placeholder="t('supportTickets.replyPlaceholder')" :disabled="busy" required />
                  <div class="mt-2 flex items-center justify-between gap-3">
                    <p class="text-xs text-ink-muted">{{ characterCount(replyBody) }}/4000</p>
                    <button type="submit" class="btn btn-primary gap-1.5" :disabled="busy || !canReply"><Icon name="arrowUp" size="sm" />{{ t('supportTickets.reply') }}</button>
                  </div>
                </form>
              </div>
            </template>
          </template>

          <div v-else class="ticket-welcome">
            <span class="welcome-icon"><Icon name="chatBubble" size="xl" /></span>
            <h2 class="mt-5 text-lg font-semibold text-ink-strong">{{ t('supportTickets.selectTicket') }}</h2>
            <p class="mt-2 max-w-xs text-sm leading-6 text-ink-muted">{{ t('supportTickets.selectDescription') }}</p>
            <button v-if="!admin" class="btn btn-secondary mt-5" @click="openCreate">{{ t('supportTickets.create') }}</button>
          </div>
        </section>
      </div>
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ticketsAPI } from '@/api/tickets'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'
import type { CreateTicketRequest, SupportTicket, TicketCategory, TicketStatus } from '@/types/support-ticket'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import TicketCategoryMenu from './TicketCategoryMenu.vue'

const props = withDefaults(defineProps<{ admin?: boolean }>(), { admin: false })
const { t, locale } = useI18n()
const appStore = useAppStore()
const tickets = ref<SupportTicket[]>([])
const loading = ref(false)
const listError = ref('')
const category = ref('')
const activeTab = ref(props.admin ? 'unanswered' : '')
const search = ref('')
const pagination = reactive({ page: 1, page_size: getPersistedPageSize(), total: 0, pages: 1 })
const createOpen = ref(false)
const mobileDetail = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const busy = ref(false)
const selected = ref<SupportTicket | null>(null)
const selectedId = ref<number | null>(null)
const replyDrafts = reactive<Record<number, string>>({})
const replyBody = computed({
  get: () => selectedId.value === null ? '' : replyDrafts[selectedId.value] || '',
  set: (value: string) => { if (selectedId.value !== null) replyDrafts[selectedId.value] = value }
})
const editStatus = ref<TicketStatus>('open')
const draft = reactive<CreateTicketRequest>({ subject: '', body: '', category: 'billing' })
const messagesViewport = ref<HTMLElement | null>(null)
const conversationHeading = ref<HTMLElement | null>(null)
const workspaceRoot = ref<HTMLElement | null>(null)
let workspaceResizeObserver: ResizeObserver | null = null
let listController: AbortController | null = null
let detailController: AbortController | null = null
let searchTimer: ReturnType<typeof setTimeout> | undefined
let disposed = false

const categoryLabels = computed<Record<TicketCategory, string>>(() => ({
  billing: t('supportTickets.categories.billing'), api_models: t('supportTickets.categories.apiModels'),
  account: t('supportTickets.categories.account'), other: t('supportTickets.categories.other')
}))
const statusLabels = computed<Record<TicketStatus, string>>(() => ({
  open: t('supportTickets.statuses.open'), pending_admin: t('supportTickets.statuses.pendingAdmin'),
  pending_user: t('supportTickets.statuses.pendingUser'), closed: t('supportTickets.statuses.closed')
}))
const statusOptions = computed(() => Object.entries(statusLabels.value).map(([value, label]) => ({ value, label })))
const tabs = computed(() => props.admin ? [
  { value: 'unanswered', label: t('supportTickets.queues.unanswered') },
  { value: 'awaiting_user', label: t('supportTickets.queues.awaitingUser') },
  { value: 'closed', label: t('supportTickets.queues.closed') }
] : [
  { value: '', label: t('supportTickets.allTickets') },
  { value: 'unclosed', label: t('supportTickets.unclosed') },
  { value: 'closed', label: t('supportTickets.statuses.closed') }
])
const hasFilters = computed(() => !!category.value || !!search.value.trim() || activeTab.value !== (props.admin ? 'unanswered' : ''))
const characterCount = (value: string) => Array.from(value.trim()).length
const canCreate = computed(() => !!draft.subject.trim() && characterCount(draft.subject) <= 80 && !!draft.body.trim() && characterCount(draft.body) <= 4000)
const canReply = computed(() => !!replyBody.value.trim() && characterCount(replyBody.value) <= 4000)
function shortDate(value: string) {
  return new Intl.DateTimeFormat(locale.value, { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }).format(new Date(value))
}
function statusColor(value: TicketStatus) {
  return value === 'closed' ? 'text-ink-muted' : value === 'pending_user' ? 'text-primary-600 dark:text-primary-400' : 'text-amber-700 dark:text-amber-400'
}
function errorMessage(err: unknown) {
  return extractI18nErrorMessage(err, t, 'supportTickets.errors', t('supportTickets.errors.fallback'))
}
function showError(err: unknown) { if (!disposed) appStore.showError(errorMessage(err)) }
async function loadTickets() {
  listController?.abort()
  const controller = new AbortController()
  listController = controller
  loading.value = true
  listError.value = ''
  try {
    const filters = { category: category.value, search: search.value.trim() }
    const result = props.admin
      ? await adminAPI.tickets.list(pagination.page, pagination.page_size, { ...filters, queue: activeTab.value }, { signal: controller.signal })
      : await ticketsAPI.list(pagination.page, pagination.page_size, { ...filters, status: activeTab.value }, { signal: controller.signal })
    if (controller.signal.aborted) return
    if (pagination.page > Math.max(1, result.pages)) {
      pagination.page = Math.max(1, result.pages)
      return await loadTickets()
    }
    tickets.value = result.items
    pagination.total = result.total
    pagination.pages = result.pages
    if (selectedId.value === null && !createOpen.value && tickets.value.length) void openTicket(tickets.value[0], false)
  } catch (err) {
    if (!controller.signal.aborted) listError.value = errorMessage(err)
  } finally {
    if (listController === controller) loading.value = false
  }
}
function resetPage() { clearTimeout(searchTimer); pagination.page = 1; void loadTickets() }
function submitSearch() { resetPage() }
function changeTab(value: string) { activeTab.value = value; resetPage() }
function changePage(page: number) { pagination.page = page; void loadTickets() }
function clearFilters() { search.value = ''; category.value = ''; activeTab.value = props.admin ? 'unanswered' : ''; resetPage() }
function refreshTickets() {
  void loadTickets()
  if (selectedId.value !== null && !createOpen.value) void openTicket({ id: selectedId.value }, false)
}
function openCreate() {
  if (busy.value) return
  detailController?.abort()
  detailLoading.value = false
  if (!draft.subject && !draft.body) draft.category = (category.value || 'billing') as TicketCategory
  createOpen.value = true
  mobileDetail.value = true
  void nextTick(() => document.getElementById('ticket-subject')?.focus())
}
function cancelCreate() {
  if (busy.value) return
  createOpen.value = false
  mobileDetail.value = false
  if (!selected.value && selectedId.value !== null) void openTicket({ id: selectedId.value }, false)
}
async function scrollMessages() {
  await nextTick()
  if (messagesViewport.value) messagesViewport.value.scrollTop = messagesViewport.value.scrollHeight
}
function showTicket(ticket: SupportTicket) {
  selectedId.value = ticket.id
  selected.value = ticket
  editStatus.value = ticket.status
  void scrollMessages()
}
async function openTicket(ticket: Pick<SupportTicket, 'id'>, reveal = true) {
  if (busy.value) return
  detailController?.abort()
  const controller = new AbortController()
  detailController = controller
  selectedId.value = ticket.id
  selected.value = null
  createOpen.value = false
  if (reveal) mobileDetail.value = true
  detailLoading.value = true
  detailError.value = ''
  try {
    const result = props.admin
      ? await adminAPI.tickets.getById(ticket.id, { signal: controller.signal })
      : await ticketsAPI.getById(ticket.id, { signal: controller.signal })
    if (!controller.signal.aborted) {
      showTicket(result)
      if (reveal) void nextTick(() => conversationHeading.value?.focus({ preventScroll: true }))
    }
  } catch (err) {
    if (!controller.signal.aborted) detailError.value = errorMessage(err)
  } finally {
    if (detailController === controller) detailLoading.value = false
  }
}
function retryDetail() { if (selectedId.value !== null) void openTicket({ id: selectedId.value }) }
function closeDetail() { if (!busy.value) mobileDetail.value = false }
async function submitCreate() {
  if (busy.value || !canCreate.value) return
  busy.value = true
  try {
    const ticket = await ticketsAPI.create({ ...draft })
    if (disposed) return
    clearTimeout(searchTimer)
    search.value = ''
    category.value = ticket.category
    activeTab.value = ''
    createOpen.value = false
    showTicket(ticket)
    Object.assign(draft, { subject: '', body: '', category: 'billing' })
    detailLoading.value = false
    detailError.value = ''
    appStore.showSuccess(t('supportTickets.created'))
    resetPage()
  } catch (err) { showError(err) } finally { busy.value = false }
}
async function submitReply() {
  if (busy.value || !selected.value || !canReply.value) return
  busy.value = true
  try {
    const ticket = props.admin
      ? await adminAPI.tickets.reply(selected.value.id, replyBody.value)
      : await ticketsAPI.reply(selected.value.id, replyBody.value)
    if (disposed) return
    showTicket(ticket)
    delete replyDrafts[ticket.id]
    appStore.showSuccess(t('supportTickets.replied'))
    void loadTickets()
  } catch (err) { showError(err) } finally { busy.value = false }
}
async function closeTicket() {
  if (busy.value || !selected.value) return
  busy.value = true
  try {
    const ticket = await ticketsAPI.close(selected.value.id)
    if (disposed) return
    showTicket(ticket)
    appStore.showSuccess(t('supportTickets.closed'))
    void loadTickets()
  } catch (err) { showError(err) } finally { busy.value = false }
}
async function saveStatus() {
  if (busy.value || !selected.value || editStatus.value === selected.value.status) return
  busy.value = true
  try {
    const ticket = await adminAPI.tickets.updateStatus(selected.value.id, editStatus.value)
    if (disposed) return
    showTicket(ticket)
    appStore.showSuccess(t('supportTickets.statusSaved'))
    void loadTickets()
  } catch (err) { showError(err) } finally { busy.value = false }
}

watch(search, () => { clearTimeout(searchTimer); searchTimer = setTimeout(resetPage, 300) }, { flush: 'sync' })
onMounted(() => {
  void loadTickets()
  const workspace = workspaceRoot.value
  const main = workspace?.closest('main')
  const header = workspace?.closest('.app-workspace')?.querySelector('.app-header')
  if (!workspace || !main || !header) return
  const updateWorkspaceHeight = () => {
    const offset = workspace.getBoundingClientRect().top + window.scrollY + parseFloat(getComputedStyle(main).paddingBottom)
    workspace.style.setProperty('--ticket-workspace-offset', `${offset}px`)
  }
  updateWorkspaceHeight()
  workspaceResizeObserver = new ResizeObserver(updateWorkspaceHeight)
  workspaceResizeObserver.observe(header)
})
onUnmounted(() => { disposed = true; clearTimeout(searchTimer); listController?.abort(); detailController?.abort(); workspaceResizeObserver?.disconnect() })
</script>

<style scoped>
.tickets-workspace { @apply flex min-w-0 flex-col; height: calc(100dvh - var(--ticket-workspace-offset, 10rem)); min-height: 36rem; }
.ticket-toolbar { @apply flex min-w-0 items-center justify-between gap-3 pb-3; }
.ticket-tabs { @apply flex min-w-0 gap-5 overflow-x-auto; }
.ticket-tabs button { @apply min-h-11 shrink-0 border-b-2 border-transparent py-2 text-sm font-medium text-ink-muted transition-colors hover:text-ink-strong; }
.ticket-tabs button.active { @apply border-primary-500 text-primary-600 dark:text-primary-400; }
.ticket-tabs button:focus-visible, .ticket-list-item:focus-visible { @apply outline outline-2 -outline-offset-2 outline-primary-500; }
.ticket-icon-button { @apply min-h-10 min-w-10 !p-2; }
.ticket-panels { @apply grid min-h-0 flex-1 overflow-hidden rounded-2xl border border-line bg-surface; grid-template-columns: 20rem minmax(0, 1fr); }
.ticket-sidebar { @apply flex min-h-0 min-w-0 flex-col border-r border-line; }
.ticket-filters { @apply space-y-2 border-b border-line px-4 py-3; }
.ticket-search-submit { @apply absolute right-1 top-1/2 flex h-9 w-9 -translate-y-1/2 items-center justify-center rounded-lg text-ink-muted hover:bg-surface-muted hover:text-ink-strong; }
.ticket-list { @apply min-h-0 flex-1 overflow-y-auto; }
.ticket-list-item { @apply block w-full border-l-[3px] border-transparent px-4 py-3 text-left transition-colors hover:bg-surface-muted disabled:cursor-wait; }
.ticket-list-item.selected { @apply border-primary-500 bg-primary-50/60 dark:bg-primary-900/15; }
.ticket-status { @apply inline-flex shrink-0 items-center gap-1.5 whitespace-nowrap text-xs font-medium leading-4; }
.status-dot { @apply h-1.5 w-1.5 shrink-0 rounded-full bg-current; }
.ticket-empty { @apply flex flex-col items-center justify-center gap-3 px-5 py-12 text-center text-ink-muted; }
.ticket-pagination { @apply flex shrink-0 items-center justify-between gap-2 border-t border-line px-4 py-2; }
.ticket-conversation { @apply flex min-h-0 min-w-0 flex-col; }
.conversation-header { @apply flex shrink-0 items-center justify-between gap-3 border-b border-line px-5 py-3; }
.ticket-back { @apply mb-2 hidden min-h-9 items-center gap-1 text-xs font-medium text-ink-muted; }
.ticket-status-editor { @apply flex shrink-0 items-center gap-3 border-b border-line px-5 py-2; }
.ticket-status-editor :deep(.select-trigger) { @apply py-2 text-sm; }
.ticket-messages { @apply flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto bg-surface-muted/40 p-5; }
.ticket-message { @apply flex max-w-[88%] items-start gap-2.5 self-start; }
.ticket-message.from-user { @apply flex-row-reverse self-end; }
.message-avatar { @apply mt-1 flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-line bg-surface text-ink-muted; }
.from-user .message-avatar { @apply border-primary-200 text-primary-600 dark:border-primary-800 dark:text-primary-400; }
.message-bubble { @apply min-w-0 rounded-xl border border-line bg-surface px-3 py-2.5; overflow-wrap: anywhere; }
.from-user .message-bubble { @apply border-primary-200 bg-primary-50/60 dark:border-primary-800 dark:bg-primary-900/15; }
.conversation-footer { @apply shrink-0 border-t border-line px-5 py-3; }
.ticket-reply-input { @apply block w-full resize-y rounded-lg border border-line bg-surface-muted/40 px-3 py-2 text-sm leading-6 text-ink placeholder:text-ink-muted focus:border-primary-500 focus:outline focus:outline-2 focus:outline-offset-2 focus:outline-primary-500 focus:ring-0; min-height: 4rem; max-height: 12rem; }
.ticket-create-form { @apply flex min-h-0 flex-1 flex-col; }
.ticket-create-fields { @apply flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto px-5 py-4; }
.ticket-welcome { @apply flex min-h-0 flex-1 flex-col items-center justify-center p-8 text-center; }
.welcome-icon { @apply flex h-16 w-16 items-center justify-center rounded-2xl bg-surface-muted text-ink-muted; }
.ticket-placeholder { @apply flex items-center justify-center p-12 text-sm text-ink-muted; }
@media (max-width: 1023px) {
  .tickets-workspace { min-height: 32rem; }
  .ticket-panels { grid-template-columns: 18rem minmax(0, 1fr); }
}
@media (max-width: 767px) {
  .tickets-workspace { min-height: 30rem; }
  .ticket-toolbar { @apply flex-wrap gap-2; }
  .ticket-tabs { @apply gap-4; }
  .ticket-tabs button { @apply text-xs; }
  .ticket-toolbar .btn-primary { @apply px-3 text-xs; }
  .ticket-panels { @apply flex; }
  .ticket-sidebar { @apply w-full border-r-0; }
  .ticket-conversation { @apply hidden w-full; }
  .show-conversation .ticket-sidebar { @apply hidden; }
  .show-conversation .ticket-conversation { @apply flex; }
  .ticket-back { @apply inline-flex; }
  .conversation-header { @apply items-start; }
  .conversation-header, .ticket-status-editor, .conversation-footer { @apply px-4; }
  .ticket-messages { @apply gap-4 p-4; }
  .ticket-message { @apply max-w-[95%] gap-2; }
  .message-avatar { @apply h-7 w-7; }
  .ticket-create-fields { @apply px-4; }
  .ticket-reply-input { @apply text-base; }
}
</style>
