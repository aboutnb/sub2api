<template>
  <div>
    <!-- 铃铛按钮 -->
    <button
      v-if="!hideTrigger"
      type="button"
      @click="openModal"
      class="btn btn-ghost btn-icon relative text-ink dark:text-ink-muted"
      :class="{ 'text-primary-700 dark:text-primary-300': unreadCount > 0 }"
      :aria-label="t('announcements.title')"
    >
      <Icon name="bell" size="md" />
      <!-- 未读红点 -->
      <span
        v-if="unreadCount > 0"
        class="absolute right-1 top-1 flex h-2 w-2"
      >
        <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-red-500 opacity-75"></span>
        <span class="relative inline-flex h-2 w-2 rounded-full bg-red-500"></span>
      </span>
    </button>

    <!-- 公告列表 Modal -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div
          v-if="isModalOpen"
          class="fixed inset-0 z-[100] flex items-start justify-center overflow-y-auto bg-gray-900/60 p-4 pt-[8vh] backdrop-blur-sm"
          :style="listDialog.zIndexStyle.value"
          :aria-labelledby="listTitleId"
          :aria-hidden="listDialog.isTopmost.value ? undefined : 'true'"
          :inert="listDialog.isTopmost.value ? undefined : true"
          role="dialog"
          aria-modal="true"
          @click.self="listDialog.handleBackdropClick"
        >
          <div
            ref="listDialogRef"
            class="modal-content max-w-[620px] overflow-hidden"
            tabindex="-1"
            @click.stop
          >
            <div class="relative overflow-hidden border-b-2 border-line bg-primary-50 px-6 py-5 dark:border-line dark:bg-primary-900/15">
              <div class="relative z-10 flex items-start justify-between">
                <div>
                  <div class="flex items-center gap-2">
                    <div class="flex h-9 w-9 items-center justify-center rounded-lg border-2 border-primary-700 bg-primary-100 text-primary-700 shadow-pixel-sm dark:border-primary-400 dark:bg-primary-900/30 dark:text-primary-300">
                      <Icon name="bell" size="sm" />
                    </div>
                    <h2 :id="listTitleId" class="text-lg font-semibold text-ink-strong dark:text-white">
                      {{ t('announcements.title') }}
                    </h2>
                  </div>
                  <p v-if="unreadCount > 0" class="mt-2 text-sm text-ink dark:text-ink-muted">
                    <span class="font-bold text-primary-700 dark:text-primary-300">{{ unreadCount }}</span>
                    {{ t('announcements.unread') }}
                  </p>
                </div>
                <div class="flex items-center gap-2">
                  <button
                    type="button"
                    v-if="unreadCount > 0"
                    @click="markAllAsRead"
                    :disabled="loading"
                    class="btn btn-primary btn-sm"
                  >
                    {{ t('announcements.markAllRead') }}
                  </button>
                  <button
                    type="button"
                    @click="closeModal"
                    class="btn btn-ghost btn-icon text-ink-muted dark:text-ink-muted"
                    :aria-label="t('common.close')"
                  >
                    <Icon name="x" size="sm" />
                  </button>
                </div>
              </div>
              <div class="absolute bottom-0 right-6 h-1 w-14 bg-accent-500" aria-hidden="true"></div>
            </div>

            <!-- Body -->
            <div class="max-h-[65vh] overflow-y-auto">
              <!-- Loading -->
              <div v-if="loading" class="flex items-center justify-center py-16">
                <div class="relative">
                  <div class="h-12 w-12 animate-spin rounded-xl border-4 border-line border-t-primary-600 dark:border-line-strong dark:border-t-primary-400"></div>
                </div>
              </div>

              <!-- Announcements List -->
              <div v-else-if="announcements.length > 0">
                <button
                  v-for="item in announcements"
                  :key="item.id"
                  type="button"
                  class="group relative flex w-full items-center gap-4 border-b border-line px-6 py-4 text-left transition-all hover:bg-surface-muted focus-visible:z-10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-primary-500 dark:border-line dark:hover:bg-dark-700/30"
                  :class="{ 'bg-primary-50/70 dark:bg-primary-900/10': !item.read_at }"
                  style="min-height: 72px"
                  @click="openDetail(item)"
                >
                  <!-- Status Indicator -->
                  <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center">
                    <div
                      v-if="!item.read_at"
                      class="relative flex h-10 w-10 items-center justify-center rounded-xl border-2 border-primary-700 bg-primary-100 text-primary-700 shadow-pixel-sm dark:border-primary-400 dark:bg-primary-900/30 dark:text-primary-300"
                    >
                      <!-- Pulse ring -->
                      <!-- Icon -->
                      <svg class="relative z-10 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <div
                      v-else
                      class="flex h-10 w-10 items-center justify-center rounded-xl bg-surface-muted text-ink-muted dark:bg-surface-muted dark:text-ink"
                    >
                      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                  </div>

                  <!-- Content -->
                  <div class="flex min-w-0 flex-1 items-center justify-between gap-4">
                    <div class="min-w-0 flex-1">
                      <h3 class="truncate text-sm font-medium text-ink-strong dark:text-white">
                        {{ item.title }}
                      </h3>
                      <div class="mt-1 flex items-center gap-2">
                        <time class="text-xs text-ink-muted dark:text-ink-muted">
                          {{ formatRelativeTime(item.created_at) }}
                        </time>
                        <span
                          v-if="!item.read_at"
                          class="badge badge-primary"
                        >
                          <span class="relative flex h-1.5 w-1.5">
                            <span class="relative inline-flex h-1.5 w-1.5 rounded-full bg-primary-600"></span>
                          </span>
                          {{ t('announcements.unread') }}
                        </span>
                      </div>
                    </div>

                    <!-- Arrow -->
                    <div class="flex-shrink-0">
                      <svg
                        class="h-5 w-5 text-ink-muted transition-transform group-hover:translate-x-1 dark:text-ink"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                        stroke-width="2"
                      >
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                      </svg>
                    </div>
                  </div>

                  <!-- Unread indicator bar -->
                  <div
                    v-if="!item.read_at"
                    class="absolute left-0 top-0 h-full w-1 bg-accent-500"
                  ></div>
                </button>
              </div>

              <!-- Empty State -->
              <div v-else class="flex flex-col items-center justify-center py-16">
                <div class="relative mb-4">
                  <div class="flex h-16 w-16 items-center justify-center rounded-2xl border-2 border-line-strong bg-surface-muted shadow-pixel-sm dark:border-line-strong dark:bg-surface-muted">
                    <Icon name="inbox" size="xl" class="text-ink-muted dark:text-ink-muted" />
                  </div>
                  <div class="absolute -right-1 -top-1 flex h-6 w-6 items-center justify-center rounded-full bg-green-500 text-white">
                    <svg class="h-3.5 w-3.5" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                    </svg>
                  </div>
                </div>
                <p class="text-sm font-medium text-ink-strong dark:text-white">{{ t('announcements.empty') }}</p>
                <p class="mt-1 text-xs text-ink-muted dark:text-ink-muted">{{ t('announcements.emptyDescription') }}</p>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 公告详情 Modal -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div
          v-if="detailModalOpen && selectedAnnouncement"
          class="fixed inset-0 z-[110] flex items-start justify-center overflow-y-auto bg-gray-900/60 p-4 pt-[6vh] backdrop-blur-sm"
          :style="detailDialog.zIndexStyle.value"
          :aria-labelledby="detailTitleId"
          :aria-hidden="detailDialog.isTopmost.value ? undefined : 'true'"
          :inert="detailDialog.isTopmost.value ? undefined : true"
          role="dialog"
          aria-modal="true"
          @click.self="detailDialog.handleBackdropClick"
        >
          <div
            ref="detailDialogRef"
            class="modal-content max-w-[780px] overflow-hidden"
            tabindex="-1"
            @click.stop
          >
            <div class="relative overflow-hidden border-b-2 border-line bg-primary-50 px-5 py-6 dark:border-line dark:bg-primary-900/15 sm:px-8">
              <div class="absolute right-6 top-0 h-1 w-16 bg-accent-500" aria-hidden="true"></div>

              <div class="relative z-10 flex items-start justify-between gap-4">
                <div class="flex-1 min-w-0">
                  <!-- Icon and Category -->
                  <div class="mb-3 flex items-center gap-2">
                    <div class="flex h-10 w-10 items-center justify-center rounded-xl border-2 border-primary-700 bg-primary-100 text-primary-700 shadow-pixel-sm dark:border-primary-400 dark:bg-primary-900/30 dark:text-primary-300">
                      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <div class="flex items-center gap-2">
                      <span class="badge badge-primary">
                        {{ t('announcements.title') }}
                      </span>
                      <span
                        v-if="!selectedAnnouncement.read_at"
                        class="badge border-accent-700 bg-accent-100 text-accent-800 dark:border-accent-500 dark:bg-accent-900/30 dark:text-accent-200"
                      >
                        <span class="relative flex h-2 w-2">
                          <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-white opacity-75"></span>
                          <span class="relative inline-flex h-2 w-2 rounded-full bg-white"></span>
                        </span>
                        {{ t('announcements.unread') }}
                      </span>
                    </div>
                  </div>

                  <!-- Title -->
                  <h2 :id="detailTitleId" class="mb-3 text-2xl font-bold leading-tight text-ink-strong dark:text-white">
                    {{ selectedAnnouncement.title }}
                  </h2>

                  <!-- Meta Info -->
                  <div class="flex items-center gap-4 text-sm text-ink dark:text-ink-muted">
                    <div class="flex items-center gap-1.5">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      <time>{{ formatRelativeWithDateTime(selectedAnnouncement.created_at) }}</time>
                    </div>
                    <div class="flex items-center gap-1.5">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                      <span>{{ selectedAnnouncement.read_at ? t('announcements.read') : t('announcements.unread') }}</span>
                    </div>
                  </div>
                </div>

                <!-- Close button -->
                <button
                  type="button"
                  @click="closeDetail"
                  class="btn btn-ghost btn-icon flex-shrink-0 text-ink-muted dark:text-ink-muted"
                  :aria-label="t('common.close')"
                >
                  <Icon name="x" size="md" />
                </button>
              </div>
            </div>

            <!-- Body with Enhanced Markdown -->
            <div class="max-h-[60vh] overflow-y-auto bg-white px-8 py-8 dark:bg-surface">
              <!-- Content with decorative border -->
              <div class="relative">
                <!-- Decorative left border -->
                <div class="absolute bottom-0 left-0 top-0 w-1 rounded-full bg-accent-500"></div>

                <div class="pl-6">
                  <div
                    class="markdown-body prose prose-sm max-w-none dark:prose-invert"
                    v-html="renderMarkdown(selectedAnnouncement.content)"
                  ></div>
                </div>
              </div>
            </div>

            <!-- Footer with Actions -->
            <div class="border-t-2 border-line bg-surface-muted/60 px-5 py-5 dark:border-line dark:bg-canvas/30 sm:px-8">
              <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div class="flex items-center gap-2 text-xs text-ink-muted dark:text-ink-muted">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span>{{ selectedAnnouncement.read_at ? t('announcements.readStatus') : t('announcements.markReadHint') }}</span>
                </div>
                <div class="flex items-center gap-3">
                  <button
                    type="button"
                    @click="closeDetail"
                    class="btn btn-secondary"
                  >
                    {{ t('common.close') }}
                  </button>
                  <button
                    type="button"
                    v-if="!selectedAnnouncement.read_at"
                    @click="markAsReadAndClose(selectedAnnouncement.id)"
                    class="btn btn-primary"
                  >
                    <span class="flex items-center gap-2">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                      </svg>
                      {{ t('announcements.markRead') }}
                    </span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, getCurrentInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import { storeToRefs } from 'pinia'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAppStore } from '@/stores/app'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeTime, formatRelativeWithDateTime } from '@/utils/format'
import type { UserAnnouncement } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import { useManagedDialog } from '@/components/common/dialogStack'
import '@/styles/announcement-markdown.css'

const { t } = useI18n()
const { hideTrigger = false } = defineProps<{ hideTrigger?: boolean }>()
const appStore = useAppStore()
const announcementStore = useAnnouncementStore()

// Configure marked
marked.setOptions({
  breaks: true,
  gfm: true,
})

// Use store state (storeToRefs for reactivity)
const { announcements, loading } = storeToRefs(announcementStore)
const unreadCount = computed(() => announcementStore.unreadCount)

// Local modal state
const isModalOpen = ref(false)
const detailModalOpen = ref(false)
const selectedAnnouncement = ref<UserAnnouncement | null>(null)
const listDialogRef = ref<HTMLElement | null>(null)
const detailDialogRef = ref<HTMLElement | null>(null)
const componentUid = getCurrentInstance()?.uid ?? 0
const listTitleId = `announcement-list-title-${componentUid}`
const detailTitleId = `announcement-detail-title-${componentUid}`

// Methods
function renderMarkdown(content: string): string {
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
}

function openModal() {
  isModalOpen.value = true
}

defineExpose({ openModal })

function closeModal() {
  isModalOpen.value = false
}

function openDetail(announcement: UserAnnouncement) {
  selectedAnnouncement.value = announcement
  detailModalOpen.value = true
  if (!announcement.read_at) {
    markAsRead(announcement.id)
  }
}

function closeDetail() {
  detailModalOpen.value = false
  selectedAnnouncement.value = null
}

async function markAsRead(id: number) {
  try {
    await announcementStore.markAsRead(id)
  } catch (err: any) {
    appStore.showError(err?.message || t('common.unknownError'))
  }
}

async function markAsReadAndClose(id: number) {
  await markAsRead(id)
  appStore.showSuccess(t('announcements.markedAsRead'))
  closeDetail()
}

async function markAllAsRead() {
  try {
    await announcementStore.markAllAsRead()
    appStore.showSuccess(t('announcements.allMarkedAsRead'))
  } catch (err: any) {
    appStore.showError(err?.message || t('common.unknownError'))
  }
}

const listDialog = useManagedDialog({
  open: isModalOpen,
  dialogRef: listDialogRef,
  onClose: closeModal,
  closeOnClickOutside: true,
  name: listTitleId,
})

const detailDialog = useManagedDialog({
  open: detailModalOpen,
  dialogRef: detailDialogRef,
  onClose: closeDetail,
  closeOnClickOutside: true,
  name: detailTitleId,
})
</script>

<style scoped>
/* Modal Animations */
.modal-fade-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 1, 1);
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-from > div {
  transform: scale(0.94) translateY(-12px);
  opacity: 0;
}

.modal-fade-leave-to > div {
  transform: scale(0.96) translateY(-8px);
  opacity: 0;
}

/* Scrollbar Styling */
.overflow-y-auto::-webkit-scrollbar {
  width: 8px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: transparent;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #cbd5e1, #94a3b8);
  border-radius: 4px;
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #4b5563, #374151);
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(to bottom, #94a3b8, #64748b);
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(to bottom, #6b7280, #4b5563);
}

@media (prefers-reduced-motion: reduce) {
  .modal-fade-enter-active,
  .modal-fade-leave-active {
    transition-duration: 1ms;
  }

  .modal-fade-enter-from > div,
  .modal-fade-leave-to > div {
    transform: none;
  }
}
</style>
