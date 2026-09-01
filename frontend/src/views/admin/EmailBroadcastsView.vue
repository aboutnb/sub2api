<template>
  <AppLayout>
    <div class="mx-auto max-w-7xl space-y-5">
      <div class="flex flex-col gap-3 border-b border-gray-200 pb-5 sm:flex-row sm:items-end sm:justify-between dark:border-dark-700">
        <div class="grid grid-cols-3 gap-6">
          <div>
            <div class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.emailBroadcasts.eligibleRecipients') }}</div>
            <div class="mt-1 text-2xl font-semibold text-gray-950 dark:text-white">{{ eligibleRecipients }}</div>
          </div>
          <div>
            <div class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.emailBroadcasts.activeTasks') }}</div>
            <div class="mt-1 text-2xl font-semibold text-gray-950 dark:text-white">{{ activeTaskCount }}</div>
          </div>
          <div>
            <div class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.emailBroadcasts.completedTasks') }}</div>
            <div class="mt-1 text-2xl font-semibold text-gray-950 dark:text-white">{{ completedTaskCount }}</div>
          </div>
        </div>
        <div class="flex flex-wrap gap-2">
          <router-link :to="{ path: '/admin/settings', query: { tab: 'email' } }" class="btn btn-secondary">
            <Icon name="edit" size="sm" class="mr-1.5" />
            {{ t('admin.emailBroadcasts.templateSettings') }}
          </router-link>
          <button class="btn btn-secondary" :disabled="loading" :title="t('admin.emailBroadcasts.refresh')" @click="refreshAll">
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          </button>
          <button class="btn btn-primary" @click="openCreateDialog">
            <Icon name="plus" size="sm" class="mr-1.5" />
            {{ t('admin.emailBroadcasts.newTask') }}
          </button>
        </div>
      </div>

      <div class="overflow-hidden border-y border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800 sm:rounded-lg sm:border">
        <div v-if="loading && tasks.length === 0" class="flex items-center justify-center py-20">
          <span class="h-7 w-7 animate-spin rounded-full border-2 border-gray-200 border-t-primary-600 dark:border-dark-600 dark:border-t-primary-400"></span>
        </div>
        <div v-else-if="tasks.length === 0" class="py-20 text-center text-sm text-gray-500 dark:text-gray-400">
          <Icon name="mail" size="xl" class="mx-auto mb-3 text-gray-300 dark:text-dark-500" />
          {{ t('admin.emailBroadcasts.noTasks') }}
        </div>

        <div v-else>
          <div class="hidden grid-cols-[minmax(240px,1.5fr)_130px_minmax(220px,1fr)_160px_130px] gap-4 border-b border-gray-200 bg-gray-50 px-5 py-3 text-xs font-semibold text-gray-500 lg:grid dark:border-dark-700 dark:bg-dark-850 dark:text-gray-400">
            <span>{{ t('admin.emailBroadcasts.task') }}</span>
            <span>{{ t('common.status') }}</span>
            <span>{{ t('admin.emailBroadcasts.progress') }}</span>
            <span>{{ t('admin.emailBroadcasts.schedule') }}</span>
            <span class="text-right">{{ t('admin.emailBroadcasts.actions') }}</span>
          </div>

          <article
            v-for="task in tasks"
            :key="task.id"
            class="grid gap-4 border-b border-gray-100 px-5 py-4 last:border-b-0 lg:grid-cols-[minmax(240px,1.5fr)_130px_minmax(220px,1fr)_160px_130px] lg:items-center dark:border-dark-700/70"
          >
            <div class="min-w-0">
              <div class="truncate text-sm font-semibold text-gray-950 dark:text-white">{{ task.title }}</div>
              <div class="mt-1 flex flex-wrap items-center gap-x-2 text-xs text-gray-500 dark:text-gray-400">
                <span>#{{ task.id }}</span>
                <span>{{ task.variables.broadcast_heading_zh || task.variables.broadcast_heading_en || task.variables.maintenance_title }}</span>
              </div>
            </div>

            <div>
              <span class="inline-flex rounded px-2 py-1 text-xs font-medium" :class="statusClass(task.status)">
                {{ t(`admin.emailBroadcasts.status.${task.status}`) }}
              </span>
            </div>

            <div>
              <div class="flex items-center justify-between text-xs text-gray-600 dark:text-gray-300">
                <span>{{ task.sent_count + task.failed_count }} / {{ task.total_recipients }}</span>
                <span>{{ progressPercent(task) }}%</span>
              </div>
              <div class="mt-2 h-1.5 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-700">
                <div class="h-full bg-emerald-500 transition-all" :style="{ width: `${progressPercent(task)}%` }"></div>
              </div>
              <div class="mt-2 flex gap-3 text-xs">
                <span class="text-emerald-700 dark:text-emerald-400">{{ t('admin.emailBroadcasts.sent') }} {{ task.sent_count }}</span>
                <span :class="task.failed_count ? 'text-red-600 dark:text-red-400' : 'text-gray-400'">{{ t('admin.emailBroadcasts.failed') }} {{ task.failed_count }}</span>
                <span v-if="pendingCount(task) > 0" class="text-gray-500 dark:text-gray-400">{{ t('admin.emailBroadcasts.pendingCount', { count: pendingCount(task) }) }}</span>
              </div>
            </div>

            <div class="text-xs text-gray-600 dark:text-gray-300">
              <div>{{ formatDateTime(task.scheduled_at) }}</div>
              <div class="mt-1 text-gray-400">{{ audienceLabel(task) }}</div>
            </div>

            <div class="flex items-center justify-start gap-1 lg:justify-end">
              <button
                v-if="task.failed_count > 0"
                class="icon-action text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-950/30"
                :title="t('admin.emailBroadcasts.viewFailures')"
                @click="openFailures(task)"
              >
                <Icon name="exclamationCircle" size="sm" />
              </button>
              <button
                v-if="task.status === 'partially_failed'"
                class="icon-action"
                :title="t('admin.emailBroadcasts.retryFailed')"
                @click="retryTask(task)"
              >
                <Icon name="refresh" size="sm" />
              </button>
              <button
                v-if="isCancelable(task.status)"
                class="icon-action text-gray-500 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950/30"
                :title="t('admin.emailBroadcasts.cancel')"
                @click="cancelTask(task)"
              >
                <Icon name="x" size="sm" />
              </button>
            </div>
          </article>
        </div>
      </div>

      <Pagination
        v-if="pagination.total > pagination.page_size"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="changePage"
        @update:pageSize="changePageSize"
      />
    </div>

    <BaseDialog :show="showCreateDialog" :title="t('admin.emailBroadcasts.newTask')" width="wide" @close="closeCreateDialog">
      <form id="email-broadcast-form" class="space-y-6" @submit.prevent="createTask">
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.emailBroadcasts.preset') }}</label>
            <select v-model="form.preset" class="input" @change="applyPreset(form.preset)">
              <option value="domain_migration">{{ t('admin.emailBroadcasts.presets.domainMigration') }}</option>
              <option value="maintenance">{{ t('admin.emailBroadcasts.presets.maintenance') }}</option>
              <option value="service_notice">{{ t('admin.emailBroadcasts.presets.serviceNotice') }}</option>
              <option value="reactivation">{{ t('admin.emailBroadcasts.presets.reactivation') }}</option>
              <option value="custom">{{ t('admin.emailBroadcasts.presets.custom') }}</option>
            </select>
          </div>
          <div>
            <label class="input-label">{{ t('admin.emailBroadcasts.taskTitle') }}</label>
            <input v-model.trim="form.title" class="input" maxlength="200" required />
          </div>
        </div>

        <div class="border-t border-gray-200 pt-5 dark:border-dark-700">
          <div class="mb-4 flex items-center justify-between gap-3">
            <label class="input-label mb-0">{{ t('admin.emailBroadcasts.content') }}</label>
            <div class="inline-flex rounded-md bg-gray-100 p-1 dark:bg-dark-700">
              <button type="button" class="language-button" :class="contentLocale === 'zh' && 'mode-button-active'" @click="contentLocale = 'zh'">中文</button>
              <button type="button" class="language-button" :class="contentLocale === 'en' && 'mode-button-active'" @click="contentLocale = 'en'">English</button>
            </div>
          </div>

          <div class="grid gap-5 lg:grid-cols-[minmax(0,1.35fr)_minmax(260px,0.65fr)]">
            <div v-if="contentLocale === 'zh'" class="space-y-4">
              <div>
                <label class="input-label">{{ t('admin.emailBroadcasts.emailSubject') }}</label>
                <input v-model.trim="form.subject_zh" class="input" maxlength="200" required />
              </div>
              <div>
                <label class="input-label">{{ t('admin.emailBroadcasts.emailHeading') }}</label>
                <input v-model.trim="form.heading_zh" class="input" maxlength="300" required />
              </div>
              <div>
                <label class="input-label">{{ t('admin.emailBroadcasts.emailBody') }}</label>
                <textarea v-model="form.body_zh" class="input" rows="7" maxlength="20000" required></textarea>
              </div>
              <div>
                <label class="input-label">{{ t('admin.emailBroadcasts.emailAction') }}</label>
                <textarea v-model="form.action_zh" class="input" rows="2" maxlength="2000"></textarea>
              </div>
            </div>
            <div v-else class="space-y-4">
              <div>
                <label class="input-label">{{ t('admin.emailBroadcasts.emailSubject') }}</label>
                <input v-model.trim="form.subject_en" class="input" maxlength="200" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.emailBroadcasts.emailHeading') }}</label>
                <input v-model.trim="form.heading_en" class="input" maxlength="300" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.emailBroadcasts.emailBody') }}</label>
                <textarea v-model="form.body_en" class="input" rows="7" maxlength="20000"></textarea>
              </div>
              <div>
                <label class="input-label">{{ t('admin.emailBroadcasts.emailAction') }}</label>
                <textarea v-model="form.action_en" class="input" rows="2" maxlength="2000"></textarea>
              </div>
            </div>

            <div class="min-w-0 border-l-0 border-gray-200 lg:border-l lg:pl-5 dark:border-dark-700">
              <div class="text-xs font-semibold uppercase text-gray-400">{{ t('admin.emailBroadcasts.preview') }}</div>
              <div class="mt-3 break-words text-base font-semibold text-gray-950 dark:text-white">{{ previewContent.subject }}</div>
              <div class="mt-4 border-t border-gray-200 pt-4 dark:border-dark-700">
                <div class="text-lg font-semibold text-gray-950 dark:text-white">{{ previewContent.heading }}</div>
                <div class="mt-3 whitespace-pre-line break-words text-sm leading-6 text-gray-600 dark:text-gray-300">{{ previewContent.body }}</div>
                <div v-if="previewContent.action" class="mt-4 text-sm font-medium text-gray-900 dark:text-white">{{ previewContent.action }}</div>
              </div>
            </div>
          </div>
        </div>

        <div class="border-t border-gray-200 pt-5 dark:border-dark-700">
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label class="input-label">{{ t('admin.emailBroadcasts.audience') }}</label>
              <select v-model="form.audience_mode" class="input" @change="loadEstimate">
                <option value="all">{{ t('admin.emailBroadcasts.audiences.all') }}</option>
                <option value="role">{{ t('admin.emailBroadcasts.audiences.role') }}</option>
                <option value="groups">{{ t('admin.emailBroadcasts.audiences.groups') }}</option>
                <option value="selected">{{ t('admin.emailBroadcasts.audiences.selected') }}</option>
                <option value="inactive">{{ t('admin.emailBroadcasts.audiences.inactive') }}</option>
              </select>
            </div>
            <div v-if="form.audience_mode === 'role'">
              <label class="input-label">{{ t('admin.emailBroadcasts.role') }}</label>
              <select v-model="form.audience_role" class="input" @change="loadEstimate">
                <option value="user">{{ t('admin.emailBroadcasts.roles.user') }}</option>
                <option value="admin">{{ t('admin.emailBroadcasts.roles.admin') }}</option>
              </select>
            </div>
            <div v-if="form.audience_mode === 'groups'" class="md:col-span-2">
              <label class="input-label">{{ t('admin.emailBroadcasts.groups') }}</label>
              <select v-model="form.audience_group_ids" class="input min-h-28" multiple @change="loadEstimate">
                <option v-for="group in groups" :key="group.id" :value="group.id">{{ group.name }}</option>
              </select>
            </div>
            <div v-if="form.audience_mode === 'selected'" class="md:col-span-2">
              <label class="input-label">{{ t('admin.emailBroadcasts.selectedEmails') }}</label>
              <textarea v-model="form.audience_emails" class="input font-mono" rows="4" :placeholder="t('admin.emailBroadcasts.selectedEmailsPlaceholder')" @blur="loadEstimate"></textarea>
            </div>
            <div v-if="form.audience_mode === 'inactive'">
              <label class="input-label">{{ t('admin.emailBroadcasts.inactiveDays') }}</label>
              <input v-model.number="form.inactive_days" type="number" min="1" max="3650" step="1" class="input" @change="loadEstimate" />
              <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.emailBroadcasts.inactiveDaysHelp') }}</div>
            </div>
          </div>
        </div>

        <div class="border-t border-gray-200 pt-5 dark:border-dark-700">
          <label class="input-label">{{ t('admin.emailBroadcasts.deliveryTime') }}</label>
          <div class="inline-flex rounded-md bg-gray-100 p-1 dark:bg-dark-700">
            <button type="button" class="mode-button" :class="form.delivery_mode === 'immediate' && 'mode-button-active'" @click="form.delivery_mode = 'immediate'">
              {{ t('admin.emailBroadcasts.immediate') }}
            </button>
            <button type="button" class="mode-button" :class="form.delivery_mode === 'scheduled' && 'mode-button-active'" @click="form.delivery_mode = 'scheduled'">
              {{ t('admin.emailBroadcasts.scheduled') }}
            </button>
          </div>
          <div v-if="form.delivery_mode === 'scheduled'" class="mt-3 max-w-sm">
            <label class="input-label">{{ t('admin.emailBroadcasts.scheduledAt') }}</label>
            <input v-model="form.scheduled_at" type="datetime-local" class="input" required />
          </div>
        </div>

        <div class="border-t border-gray-200 pt-5 dark:border-dark-700">
          <div class="grid grid-cols-1 gap-3 sm:grid-cols-[1fr_130px_auto] sm:items-end">
            <div>
              <label class="input-label">{{ t('admin.emailBroadcasts.testRecipient') }}</label>
              <input v-model.trim="form.test_email" type="email" class="input" />
            </div>
            <div>
              <label class="input-label">{{ t('admin.emailBroadcasts.testLocale') }}</label>
              <select v-model="form.test_locale" class="input">
                <option value="zh">中文</option>
                <option value="en">English</option>
              </select>
            </div>
            <button type="button" class="btn btn-secondary" :disabled="testing || !form.test_email" @click="sendTest">
              <Icon name="mail" size="sm" class="mr-1.5" />
              {{ testing ? t('admin.emailBroadcasts.testing') : t('admin.emailBroadcasts.sendTest') }}
            </button>
          </div>
        </div>

        <div class="rounded border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-950/20 dark:text-amber-300">
          <div class="font-semibold">{{ audienceEstimateLoading ? '...' : eligibleRecipients }} {{ t('admin.emailBroadcasts.recipients') }}</div>
          <div class="mt-1 text-xs leading-5">{{ t('admin.emailBroadcasts.confirmAudienceFlexible') }}</div>
          <div class="text-xs leading-5">{{ t('admin.emailBroadcasts.templateSnapshot') }}</div>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="closeCreateDialog">{{ t('common.cancel') }}</button>
          <button type="submit" form="email-broadcast-form" class="btn btn-primary" :disabled="creating || eligibleRecipients === 0">
            {{ creating ? t('admin.emailBroadcasts.creating') : t('admin.emailBroadcasts.create') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="showFailuresDialog" :title="t('admin.emailBroadcasts.failedRecipients')" width="wide" @close="showFailuresDialog = false">
      <div v-if="failuresLoading" class="flex justify-center py-12">
        <span class="h-6 w-6 animate-spin rounded-full border-2 border-gray-200 border-t-primary-600"></span>
      </div>
      <div v-else-if="failedRecipients.length === 0" class="py-12 text-center text-sm text-gray-500">
        {{ t('admin.emailBroadcasts.noFailures') }}
      </div>
      <div v-else class="divide-y divide-gray-100 dark:divide-dark-700">
        <div v-for="recipient in failedRecipients" :key="recipient.id" class="grid gap-2 py-3 text-sm sm:grid-cols-[minmax(180px,1fr)_100px_minmax(220px,1.5fr)]">
          <div class="break-all font-medium text-gray-900 dark:text-white">{{ recipient.email }}</div>
          <div class="text-gray-500">{{ t('admin.emailBroadcasts.attempts') }}: {{ recipient.attempts }}</div>
          <div class="break-words text-red-600 dark:text-red-400">{{ recipient.last_error || '-' }}</div>
        </div>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type {
  EmailBroadcastAudience,
  EmailBroadcastAudienceMode,
  EmailBroadcastEvent,
  EmailBroadcastPayload,
  EmailBroadcastRecipient,
  EmailBroadcastStatus,
  EmailBroadcastTask
} from '@/api/admin'
import type { AdminGroup } from '@/types'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'

const { t, locale } = useI18n()
const appStore = useAppStore()
const tasks = ref<EmailBroadcastTask[]>([])
const eligibleRecipients = ref(0)
const loading = ref(false)
const creating = ref(false)
const testing = ref(false)
const audienceEstimateLoading = ref(false)
const showCreateDialog = ref(false)
const showFailuresDialog = ref(false)
const failuresLoading = ref(false)
const failedRecipients = ref<EmailBroadcastRecipient[]>([])
const groups = ref<AdminGroup[]>([])
const contentLocale = ref<'zh' | 'en'>('zh')
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const activeTaskCount = computed(() => tasks.value.filter(task => ['scheduled', 'pending', 'running'].includes(task.status)).length)
const completedTaskCount = computed(() => tasks.value.filter(task => ['succeeded', 'partially_failed'].includes(task.status)).length)

function localInput(date: Date): string {
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

type BroadcastPreset = 'domain_migration' | 'maintenance' | 'service_notice' | 'reactivation' | 'custom'

const presetContent: Record<Exclude<BroadcastPreset, 'custom'>, Record<string, string>> = {
  domain_migration: {
    title: '服务器升级及域名迁移通知',
    subject_zh: '服务器升级及域名迁移通知',
    heading_zh: '服务器升级与服务迁移',
    body_zh: '为提升服务稳定性，平台将进行服务器升级与服务迁移。\n\n迁移完成后，请优先使用新域名：https://aivoza.com/\n\n原域名 https://flowai.cyou/ 将继续保留并可正常访问，账号、服务配置及使用方式保持不变。',
    action_zh: '升级期间 API、网页及管理后台可能短暂不可用，请提前保存正在进行的工作。如遇访问异常，请稍后重试或切换至新域名。',
    subject_en: 'Server upgrade and domain migration notice',
    heading_en: 'Server upgrade and service migration',
    body_en: 'To improve service stability, we will upgrade and migrate the platform.\n\nAfter the migration, please use our new domain: https://aivoza.com/\n\nThe existing https://flowai.cyou/ domain will remain available. Your account, service configuration, and usage remain unchanged.',
    action_en: 'The API, website, and admin console may be briefly unavailable during the upgrade. Please save ongoing work and retry later if needed.'
  },
  maintenance: {
    title: '服务器维护通知',
    subject_zh: '服务器维护通知',
    heading_zh: '计划服务器维护',
    body_zh: '为提升服务稳定性，平台将进行计划维护。维护期间部分服务可能短暂不可用。',
    action_zh: '请提前保存正在进行的工作，维护完成后无需额外操作。',
    subject_en: 'Planned server maintenance',
    heading_en: 'Planned server maintenance',
    body_en: 'We will perform planned maintenance to improve service stability. Some services may be briefly unavailable during the maintenance window.',
    action_en: 'Please save ongoing work in advance. No additional action is required after maintenance.'
  },
  service_notice: {
    title: '服务通知',
    subject_zh: '服务通知',
    heading_zh: '服务更新',
    body_zh: '我们有一项服务更新需要告知您。请在发送前编辑此处内容。',
    action_zh: '',
    subject_en: 'Service notice',
    heading_en: 'Service update',
    body_en: 'We have a service update to share. Edit this content before sending.',
    action_en: ''
  },
  reactivation: {
    title: '未活跃用户召回',
    subject_zh: '好久不见，欢迎回来',
    heading_zh: '我们期待您的再次使用',
    body_zh: '您好！您已有一段时间未使用我们的服务。平台近期持续优化了稳定性和使用体验，欢迎回来看看最新变化。',
    action_zh: '访问 https://aivoza.com/ 登录原账号即可继续使用，原有账号信息和配置保持不变。',
    subject_en: 'It has been a while — welcome back',
    heading_en: 'We would love to see you again',
    body_en: 'You have not used our service for a while. We have continued improving reliability and the overall experience, and invite you to see what is new.',
    action_en: 'Visit https://aivoza.com/ and sign in with your existing account. Your account information and configuration remain unchanged.'
  }
}

function initialForm() {
  const scheduled = new Date(Date.now() + 10 * 60 * 1000)
  return {
    preset: 'domain_migration' as BroadcastPreset,
    title: presetContent.domain_migration.title,
    subject_zh: presetContent.domain_migration.subject_zh,
    heading_zh: presetContent.domain_migration.heading_zh,
    body_zh: presetContent.domain_migration.body_zh,
    action_zh: presetContent.domain_migration.action_zh,
    subject_en: presetContent.domain_migration.subject_en,
    heading_en: presetContent.domain_migration.heading_en,
    body_en: presetContent.domain_migration.body_en,
    action_en: presetContent.domain_migration.action_en,
    event: 'system.broadcast' as EmailBroadcastEvent,
    audience_mode: 'all' as EmailBroadcastAudienceMode,
    audience_role: 'user' as 'admin' | 'user',
    audience_group_ids: [] as number[],
    audience_emails: '',
    inactive_days: 7,
    delivery_mode: 'immediate' as 'immediate' | 'scheduled',
    scheduled_at: localInput(scheduled),
    test_email: '',
    test_locale: (locale.value.startsWith('zh') ? 'zh' : 'en') as 'zh' | 'en'
  }
}

const form = reactive(initialForm())

function resetForm() {
  Object.assign(form, initialForm())
  contentLocale.value = locale.value.startsWith('zh') ? 'zh' : 'en'
}

function applyPreset(preset: BroadcastPreset) {
  if (preset === 'custom') {
    Object.assign(form, {
      title: '', subject_zh: '', heading_zh: '', body_zh: '', action_zh: '',
      subject_en: '', heading_en: '', body_en: '', action_en: '', event: 'system.broadcast'
    })
    return
  }
  Object.assign(form, presetContent[preset], {
    event: preset === 'reactivation' ? 'system.reactivation' : 'system.broadcast'
  })
  if (preset === 'reactivation') {
    form.audience_mode = 'inactive'
    form.inactive_days = 7
    void loadEstimate()
  }
}

const previewContent = computed(() => contentLocale.value === 'zh'
  ? { subject: form.subject_zh, heading: form.heading_zh, body: form.body_zh, action: form.action_zh }
  : { subject: form.subject_en || form.subject_zh, heading: form.heading_en || form.heading_zh, body: form.body_en || form.body_zh, action: form.action_en || form.action_zh })

function audiencePayload(): EmailBroadcastAudience {
  if (form.audience_mode === 'role') return { mode: 'role', roles: [form.audience_role] }
  if (form.audience_mode === 'groups') return { mode: 'groups', group_ids: form.audience_group_ids.map(Number) }
  if (form.audience_mode === 'selected') {
    return {
      mode: 'selected',
      emails: form.audience_emails.split(/[\n,;]+/).map(email => email.trim()).filter(Boolean)
    }
  }
  if (form.audience_mode === 'inactive') return { mode: 'inactive', inactive_days: Number(form.inactive_days) }
  return { mode: 'all' }
}

function audienceReady() {
  const audience = audiencePayload()
  if (audience.mode === 'groups') return Boolean(audience.group_ids?.length)
  if (audience.mode === 'selected') return Boolean(audience.emails?.length)
  if (audience.mode === 'inactive') return Number(audience.inactive_days) >= 1 && Number(audience.inactive_days) <= 3650
  return true
}

function broadcastPayload(): EmailBroadcastPayload {
  return {
    event: form.event,
    subject_zh: form.subject_zh,
    heading_zh: form.heading_zh,
    body_zh: form.body_zh,
    action_zh: form.action_zh,
    subject_en: form.subject_en,
    heading_en: form.heading_en,
    body_en: form.body_en,
    action_en: form.action_en,
    audience: audiencePayload()
  }
}

function contentComplete() {
  return [form.subject_zh, form.heading_zh, form.body_zh].every(value => value.trim())
}

async function loadTasks() {
  loading.value = true
  try {
    const response = await adminAPI.emailBroadcasts.list(pagination.page, pagination.page_size)
    tasks.value = response.items || []
    pagination.total = response.total
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.emailBroadcasts.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function loadEstimate() {
  if (showCreateDialog.value && !audienceReady()) {
    eligibleRecipients.value = 0
    return
  }
  audienceEstimateLoading.value = true
  try {
    const audience = showCreateDialog.value ? audiencePayload() : undefined
    eligibleRecipients.value = (await adminAPI.emailBroadcasts.estimate(audience)).eligible_recipients
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.emailBroadcasts.loadFailed')))
  } finally {
    audienceEstimateLoading.value = false
  }
}

async function loadGroups() {
  if (groups.value.length > 0) return
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.emailBroadcasts.loadGroupsFailed')))
  }
}

async function refreshAll() {
  await Promise.all([loadTasks(), loadEstimate()])
}

async function openCreateDialog() {
  resetForm()
  showCreateDialog.value = true
  await Promise.all([loadGroups(), loadEstimate()])
}

function closeCreateDialog() {
  if (!creating.value && !testing.value) {
    showCreateDialog.value = false
    loadEstimate()
  }
}

async function sendTest() {
  if (!contentComplete()) {
    appStore.showError(t('admin.emailBroadcasts.placeholdersMissing'))
    return
  }
  testing.value = true
  try {
    await adminAPI.emailBroadcasts.sendTest({ ...broadcastPayload(), email: form.test_email, locale: form.test_locale })
    appStore.showSuccess(t('admin.emailBroadcasts.testSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.emailBroadcasts.testFailed')))
  } finally {
    testing.value = false
  }
}

async function createTask() {
  if (!form.title || !contentComplete() || !audienceReady()) {
    appStore.showError(t('admin.emailBroadcasts.placeholdersMissing'))
    return
  }
  if (!window.confirm(t('admin.emailBroadcasts.createConfirm', { count: eligibleRecipients.value }))) return
  creating.value = true
  try {
    const payload: EmailBroadcastPayload = { ...broadcastPayload(), title: form.title }
    if (form.delivery_mode === 'scheduled') payload.scheduled_at = new Date(form.scheduled_at).toISOString()
    await adminAPI.emailBroadcasts.create(payload)
    appStore.showSuccess(t('admin.emailBroadcasts.createSuccess'))
    showCreateDialog.value = false
    await refreshAll()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.emailBroadcasts.createFailed')))
  } finally {
    creating.value = false
  }
}

function audienceLabel(task: EmailBroadcastTask) {
  const mode = task.audience?.mode || 'all'
  if (mode === 'inactive') return t('admin.emailBroadcasts.audienceDetails.inactive', { days: task.audience.inactive_days })
  return t(`admin.emailBroadcasts.audiences.${mode}`)
}

function progressPercent(task: EmailBroadcastTask) {
  if (!task.total_recipients) return 0
  return Math.min(100, Math.round(((task.sent_count + task.failed_count) / task.total_recipients) * 100))
}

function pendingCount(task: EmailBroadcastTask) {
  return Math.max(0, task.total_recipients - task.sent_count - task.failed_count)
}

function isCancelable(status: EmailBroadcastStatus) {
  return ['scheduled', 'pending', 'running'].includes(status)
}

function statusClass(status: EmailBroadcastStatus) {
  if (status === 'succeeded') return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300'
  if (status === 'partially_failed') return 'bg-red-50 text-red-700 dark:bg-red-950/40 dark:text-red-300'
  if (status === 'canceled') return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
  if (status === 'running') return 'bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300'
  return 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300'
}

async function cancelTask(task: EmailBroadcastTask) {
  if (!window.confirm(t('admin.emailBroadcasts.cancelConfirm'))) return
  try {
    await adminAPI.emailBroadcasts.cancel(task.id)
    appStore.showSuccess(t('admin.emailBroadcasts.cancelSuccess'))
    await loadTasks()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.emailBroadcasts.operationFailed')))
  }
}

async function retryTask(task: EmailBroadcastTask) {
  if (!window.confirm(t('admin.emailBroadcasts.retryConfirm'))) return
  try {
    const result = await adminAPI.emailBroadcasts.retryFailed(task.id)
    appStore.showSuccess(t('admin.emailBroadcasts.retrySuccess', { count: result.retried_recipients }))
    await loadTasks()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.emailBroadcasts.operationFailed')))
  }
}

async function openFailures(task: EmailBroadcastTask) {
  showFailuresDialog.value = true
  failuresLoading.value = true
  failedRecipients.value = []
  try {
    failedRecipients.value = (await adminAPI.emailBroadcasts.listRecipients(task.id, 'failed')).items || []
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.emailBroadcasts.loadFailed')))
  } finally {
    failuresLoading.value = false
  }
}

function changePage(page: number) {
  pagination.page = page
  loadTasks()
}

function changePageSize(pageSize: number) {
  pagination.page = 1
  pagination.page_size = pageSize
  loadTasks()
}

let refreshTimer: number | undefined
onMounted(async () => {
  await refreshAll()
  refreshTimer = window.setInterval(() => {
    if (tasks.value.some(task => isCancelable(task.status))) loadTasks()
  }, 5000)
})
onUnmounted(() => {
  if (refreshTimer) window.clearInterval(refreshTimer)
})
</script>

<style scoped>
.icon-action {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.375rem;
  color: rgb(107 114 128);
  transition: color 150ms, background-color 150ms;
}

.icon-action:hover {
  background: rgb(243 244 246);
  color: rgb(31 41 55);
}

.mode-button {
  min-width: 7rem;
  border-radius: 0.25rem;
  padding: 0.45rem 0.8rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: rgb(107 114 128);
}

.language-button {
  min-width: 5rem;
  border-radius: 0.25rem;
  padding: 0.4rem 0.7rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: rgb(107 114 128);
}

.mode-button-active {
  background: white;
  color: rgb(17 24 39);
  box-shadow: 0 1px 2px rgb(0 0 0 / 0.08);
}

:global(.dark) .mode-button-active {
  background: rgb(30 41 59);
  color: white;
}
</style>
