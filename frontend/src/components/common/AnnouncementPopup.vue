<template>
  <Teleport to="body">
    <Transition name="popup-fade">
      <div
        v-if="displayedAnnouncement"
        class="fixed inset-0 z-[120] flex items-start justify-center overflow-y-auto bg-gray-900/60 p-4 pt-[8vh] backdrop-blur-sm sm:pt-[10vh]"
        :style="popupDialog.zIndexStyle.value"
        :aria-labelledby="popupTitleId"
        :aria-describedby="popupDescriptionId"
        :aria-hidden="popupDialog.isTopmost.value ? undefined : 'true'"
        :inert="popupDialog.isTopmost.value ? undefined : true"
        role="dialog"
        aria-modal="true"
      >
        <div
          ref="popupDialogRef"
          class="announcement-card modal-content max-w-[660px] overflow-hidden"
          tabindex="-1"
          @click.stop
        >
          <div class="relative overflow-hidden border-b-2 border-line bg-primary-50 px-6 py-5 dark:border-line dark:bg-primary-900/15 sm:px-7 sm:py-6">
            <div class="pointer-events-none absolute left-6 top-0 h-1 w-16 bg-accent-500"></div>

            <div class="relative flex items-start gap-4">
              <div class="mt-0.5 flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl border-2 border-primary-700 bg-primary-100 text-primary-700 shadow-pixel-sm dark:border-primary-400 dark:bg-primary-900/30 dark:text-primary-300">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M7 8h10M7 12h6m-9 8l3.5-3.5H18a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v14z" />
                </svg>
              </div>

              <div class="min-w-0 flex-1">
                <div class="mb-2 flex flex-wrap items-center gap-2">
                  <span class="badge badge-gray uppercase tracking-[0.16em]">
                    {{ t('announcements.title') }}
                  </span>
                  <span class="badge badge-primary gap-1.5">
                    <span class="h-1.5 w-1.5 rounded-full bg-primary-600"></span>
                    {{ t('announcements.unread') }}
                  </span>
                </div>

                <h2 :id="popupTitleId" class="text-2xl font-semibold leading-tight tracking-tight text-slate-950 dark:text-white">
                  {{ displayedAnnouncement.title }}
                </h2>

                <div class="mt-3 flex items-center gap-1.5 text-sm text-slate-500 dark:text-slate-400">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <time>{{ formatRelativeWithDateTime(displayedAnnouncement.created_at) }}</time>
                </div>
              </div>
            </div>
          </div>

          <!-- Body -->
          <div class="announcement-scroll max-h-[52vh] overflow-y-auto bg-surface-muted px-6 py-6 dark:bg-surface sm:px-7 sm:py-7">
            <div class="card p-5 sm:p-6">
              <div class="relative border-l-2 border-accent-500 pl-5">
                <div
                  class="markdown-body prose prose-sm max-w-none dark:prose-invert"
                  v-html="renderedContent"
                ></div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="border-t-2 border-line bg-white px-6 py-4 dark:border-line dark:bg-surface sm:px-7">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <p :id="popupDescriptionId" class="text-xs text-slate-500 dark:text-slate-400">
                {{ t('announcements.markReadHint') }}
              </p>
              <button
                ref="dismissButtonRef"
                type="button"
                @click="handleDismiss"
                data-testid="announcement-popup-dismiss"
                class="btn btn-primary"
              >
                <span class="flex items-center gap-2">
                  <svg v-if="preview" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  {{ preview ? t('common.close') : t('announcements.markRead') }}
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeWithDateTime } from '@/utils/format'
import type { Announcement, UserAnnouncement } from '@/types'
import { useManagedDialog } from '@/components/common/dialogStack'
import '@/styles/announcement-markdown.css'

type PreviewAnnouncement = Pick<Announcement | UserAnnouncement, 'title' | 'content' | 'created_at'>

const props = withDefaults(defineProps<{
  announcement?: PreviewAnnouncement | null
  preview?: boolean
}>(), {
  announcement: null,
  preview: false,
})

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
const announcementStore = useAnnouncementStore()
const displayedAnnouncement = computed(() => (
  props.preview ? props.announcement : announcementStore.currentPopup
))
const popupOpen = computed(() => Boolean(displayedAnnouncement.value))
const popupDialogRef = ref<HTMLElement | null>(null)
const dismissButtonRef = ref<HTMLButtonElement | null>(null)
const componentUid = getCurrentInstance()?.uid ?? 0
const popupTitleId = `announcement-popup-title-${componentUid}`
const popupDescriptionId = `announcement-popup-description-${componentUid}`

marked.setOptions({
  breaks: true,
  gfm: true,
})

const renderedContent = computed(() => {
  const content = displayedAnnouncement.value?.content
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})

function handleDismiss() {
  if (props.preview) {
    emit('close')
    return
  }
  announcementStore.dismissPopup()
}

const popupDialog = useManagedDialog({
  open: popupOpen,
  dialogRef: popupDialogRef,
  onClose: handleDismiss,
  initialFocus: () => dismissButtonRef.value,
  name: popupTitleId,
})
</script>

<style scoped>
.popup-fade-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.popup-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 1, 1);
}

.popup-fade-enter-from,
.popup-fade-leave-to {
  opacity: 0;
}

.popup-fade-enter-from > div {
  transform: scale(0.94) translateY(-12px);
  opacity: 0;
}

.popup-fade-leave-to > div {
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
  background: #d8c9bb;
  border-radius: 4px;
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb {
  background: #4a566c;
}

@media (prefers-reduced-motion: reduce) {
  .popup-fade-enter-active,
  .popup-fade-leave-active {
    transition-duration: 1ms;
  }

  .popup-fade-enter-from > div,
  .popup-fade-leave-to > div {
    transform: none;
  }
}
</style>
