<template>
  <Teleport to="body">
    <Transition name="popup-fade">
      <div
        v-if="announcementStore.currentPopup"
        class="fixed inset-0 z-[120] flex items-start justify-center overflow-y-auto bg-slate-950/55 p-4 pt-[8vh] backdrop-blur-sm sm:pt-[10vh]"
      >
        <div
          class="announcement-card w-full max-w-[660px] overflow-hidden rounded-[28px] bg-stone-50 shadow-[0_28px_80px_rgba(15,23,42,0.26)] ring-1 ring-slate-900/10 dark:bg-dark-800 dark:ring-white/10"
          @click.stop
        >
          <div class="relative overflow-hidden border-b border-slate-200/80 bg-white px-6 py-5 dark:border-dark-700/70 dark:bg-dark-800 sm:px-7 sm:py-6">
            <div class="pointer-events-none absolute inset-x-0 top-0 h-1 bg-gradient-to-r from-slate-700 via-cyan-500 to-emerald-400 dark:from-slate-200 dark:via-cyan-400 dark:to-emerald-300"></div>
            <div class="pointer-events-none absolute right-0 top-0 h-36 w-36 rounded-full bg-cyan-200/30 blur-3xl dark:bg-cyan-500/10"></div>

            <div class="relative flex items-start gap-4">
              <div class="mt-0.5 flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-2xl border border-slate-200 bg-slate-50 text-slate-700 shadow-sm dark:border-dark-600 dark:bg-dark-700 dark:text-slate-200">
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M7 8h10M7 12h6m-9 8l3.5-3.5H18a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v14z" />
                </svg>
              </div>

              <div class="min-w-0 flex-1">
                <div class="mb-2 flex flex-wrap items-center gap-2">
                  <span class="inline-flex items-center rounded-full border border-slate-200 bg-slate-50 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.16em] text-slate-500 dark:border-dark-600 dark:bg-dark-700/70 dark:text-slate-300">
                    {{ t('announcements.title') }}
                  </span>
                  <span class="inline-flex items-center gap-1.5 rounded-full bg-cyan-50 px-2.5 py-1 text-xs font-medium text-cyan-700 ring-1 ring-cyan-200/80 dark:bg-cyan-500/10 dark:text-cyan-200 dark:ring-cyan-400/20">
                    <span class="h-1.5 w-1.5 rounded-full bg-cyan-500"></span>
                    {{ t('announcements.unread') }}
                  </span>
                </div>

                <h2 class="text-2xl font-semibold leading-tight tracking-tight text-slate-950 dark:text-white">
                  {{ announcementStore.currentPopup.title }}
                </h2>

                <div class="mt-3 flex items-center gap-1.5 text-sm text-slate-500 dark:text-slate-400">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <time>{{ formatRelativeWithDateTime(announcementStore.currentPopup.created_at) }}</time>
                </div>
              </div>
            </div>
          </div>

          <!-- Body -->
          <div class="announcement-scroll max-h-[52vh] overflow-y-auto bg-stone-50 px-6 py-6 dark:bg-dark-800 sm:px-7 sm:py-7">
            <div class="rounded-2xl border border-slate-200/80 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-900/35 sm:p-6">
              <div class="relative border-l border-slate-200 pl-5 dark:border-dark-600">
                <div
                  class="markdown-body prose prose-sm max-w-none dark:prose-invert"
                  v-html="renderedContent"
                ></div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="border-t border-slate-200/80 bg-white px-6 py-4 dark:border-dark-700 dark:bg-dark-800 sm:px-7">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <p class="text-xs text-slate-500 dark:text-slate-400">
                {{ t('announcements.markReadHint') }}
              </p>
              <button
                @click="handleDismiss"
                class="inline-flex items-center justify-center rounded-xl bg-slate-900 px-5 py-2.5 text-sm font-medium text-white shadow-sm transition-all hover:-translate-y-0.5 hover:bg-slate-800 hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-cyan-400 focus:ring-offset-2 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100 dark:focus:ring-offset-dark-800"
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
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAnnouncementStore } from '@/stores/announcements'
import { formatRelativeWithDateTime } from '@/utils/format'

const { t } = useI18n()
const announcementStore = useAnnouncementStore()

marked.setOptions({
  breaks: true,
  gfm: true,
})

const renderedContent = computed(() => {
  const content = announcementStore.currentPopup?.content
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
})

function handleDismiss() {
  announcementStore.dismissPopup()
}

// Manage body overflow — only set, never unset (bell component handles restore)
watch(
  () => announcementStore.currentPopup,
  (popup) => {
    if (popup) {
      document.body.style.overflow = 'hidden'
    }
  }
)
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
  background: linear-gradient(to bottom, #cbd5e1, #94a3b8);
  border-radius: 4px;
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #4b5563, #374151);
}
</style>
