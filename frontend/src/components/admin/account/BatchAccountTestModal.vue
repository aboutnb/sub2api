<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.batchTest.title')"
    width="full"
    @close="handleClose"
  >
    <div class="space-y-5">
      <div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <div class="batch-test-stat">
          <span>{{ t('admin.accounts.batchTest.total') }}</span>
          <strong data-testid="batch-test-total-count">{{ totalCount }}</strong>
        </div>
        <div class="batch-test-stat">
          <span class="inline-flex items-center gap-1.5">
            <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
            {{ t('admin.accounts.batchTest.success') }}
          </span>
          <strong class="text-emerald-600 dark:text-emerald-300" data-testid="batch-test-success-count">
            {{ successCount }}
          </strong>
        </div>
        <div class="batch-test-stat">
          <span>{{ t('admin.accounts.batchTest.failed') }}</span>
          <strong class="text-rose-600 dark:text-rose-300" data-testid="batch-test-failed-count">
            {{ failedCount }}
          </strong>
        </div>
        <div class="batch-test-stat">
          <span class="inline-flex items-center gap-1.5">
            <span class="h-1.5 w-1.5 rounded-full bg-sky-500"></span>
            {{ t('admin.accounts.batchTest.progress') }}
          </span>
          <strong class="text-sky-600 dark:text-sky-300" data-testid="batch-test-progress-count">
            {{ t('admin.accounts.batchTest.progressValue', { done: doneCount, total: totalCount }) }}
          </strong>
        </div>
      </div>

      <div class="h-2 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-700">
        <div
          class="h-full rounded-full bg-sky-500 transition-all duration-300"
          :style="{ width: `${progressPercent}%` }"
        ></div>
      </div>

      <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_16rem]">
        <label class="space-y-1.5">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('admin.accounts.batchTest.testModel') }}
          </span>
          <select
            v-model="modelMode"
            :disabled="isRunning"
            class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-400 dark:border-dark-500 dark:bg-dark-700 dark:text-gray-100 dark:disabled:bg-dark-800"
          >
            <option value="default">{{ t('admin.accounts.batchTest.defaultModel') }}</option>
            <option value="custom">{{ t('admin.accounts.batchTest.customModel') }}</option>
          </select>
          <input
            v-if="modelMode === 'custom'"
            v-model.trim="customModelId"
            :disabled="isRunning"
            type="text"
            class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-400 dark:border-dark-500 dark:bg-dark-700 dark:text-gray-100 dark:disabled:bg-dark-800"
            :placeholder="t('admin.accounts.batchTest.customModelPlaceholder')"
          />
        </label>

        <label class="space-y-1.5">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('admin.accounts.batchTest.concurrency') }}
          </span>
          <input
            v-model.number="concurrencyInput"
            :disabled="isRunning"
            type="number"
            min="1"
            max="50"
            data-testid="batch-test-concurrency"
            class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-400 dark:border-dark-500 dark:bg-dark-700 dark:text-gray-100 dark:disabled:bg-dark-800"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400">
            {{ t('admin.accounts.batchTest.concurrencyHint') }}
          </p>
        </label>
      </div>

      <div class="space-y-3">
        <div class="grid gap-3 lg:grid-cols-[17rem_minmax(0,1fr)] lg:items-end">
          <label class="space-y-1.5">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.accounts.batchTest.resultFilter') }}
            </span>
            <select
              v-model="resultFilter"
              class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-dark-500 dark:bg-dark-700 dark:text-gray-100"
            >
              <option v-for="filter in resultFilters" :key="filter.value" :value="filter.value">
                {{ filter.label }}
              </option>
            </select>
          </label>

          <div class="flex flex-wrap gap-2">
            <button
              v-for="filter in resultFilters"
              :key="filter.value"
              type="button"
              :class="[
                'rounded-full border px-3 py-1.5 text-sm transition-colors',
                resultFilter === filter.value
                  ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-500/15 dark:text-primary-200'
                  : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300 hover:bg-gray-50 dark:border-dark-500 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'
              ]"
              @click="resultFilter = filter.value"
            >
              {{ filter.label }}
            </button>
          </div>
        </div>

        <div class="overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
          <div class="max-h-[52vh] overflow-auto">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-600">
              <thead class="sticky top-0 z-10 bg-gray-50 dark:bg-dark-700">
                <tr>
                  <th class="batch-test-th w-[34%]">{{ t('admin.accounts.batchTest.columns.account') }}</th>
                  <th class="batch-test-th w-[12rem]">{{ t('admin.accounts.batchTest.columns.platform') }}</th>
                  <th class="batch-test-th w-[8rem]">{{ t('admin.accounts.batchTest.columns.status') }}</th>
                  <th class="batch-test-th">{{ t('admin.accounts.batchTest.columns.result') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr
                  v-for="row in filteredRows"
                  :key="row.account.id"
                  :data-testid="`batch-test-row-${row.account.id}`"
                  class="hover:bg-gray-50 dark:hover:bg-dark-700/60"
                >
                  <td class="batch-test-td">
                    <div class="min-w-0">
                      <div class="truncate font-medium text-gray-900 dark:text-gray-100" :title="row.account.name">
                        {{ row.account.name }}
                      </div>
                      <div class="text-xs text-gray-500 dark:text-gray-400">ID {{ row.account.id }}</div>
                    </div>
                  </td>
                  <td class="batch-test-td">
                    <div class="flex flex-wrap gap-1.5">
                      <span class="rounded bg-gray-100 px-2 py-0.5 text-xs font-medium uppercase text-gray-700 dark:bg-dark-600 dark:text-gray-200">
                        {{ row.account.platform }}
                      </span>
                      <span class="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-600 dark:bg-dark-600 dark:text-gray-300">
                        {{ row.account.type }}
                      </span>
                    </div>
                  </td>
                  <td class="batch-test-td">
                    <span :class="statusBadgeClass(row.status)">
                      <Icon
                        v-if="row.status === 'running'"
                        name="refresh"
                        size="xs"
                        class="animate-spin"
                      />
                      <span v-else-if="row.status === 'failed'" aria-hidden="true">x</span>
                      <span>{{ t(`admin.accounts.batchTest.status.${row.status}`) }}</span>
                    </span>
                  </td>
                  <td class="batch-test-td">
                    <div class="max-w-[34rem] truncate text-sm text-gray-600 dark:text-gray-300" :title="row.result || ''">
                      {{ row.result || t('admin.accounts.batchTest.noResult') }}
                    </div>
                  </td>
                </tr>
                <tr v-if="filteredRows.length === 0">
                  <td colspan="4" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
                    {{ totalCount === 0 ? t('admin.accounts.batchTest.empty') : t('admin.accounts.batchTest.noFilteredResults') }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          type="button"
          class="btn btn-secondary"
          @click="handleClose"
        >
          {{ t('common.close') }}
        </button>
        <button
          v-if="isRunning"
          type="button"
          class="btn btn-warning"
          data-testid="batch-test-stop"
          :disabled="isStopping"
          @click="stopBatch"
        >
          <Icon name="x" size="sm" />
          {{ isStopping ? t('admin.accounts.batchTest.stopping') : t('admin.accounts.batchTest.stop') }}
        </button>
        <button
          v-else
          type="button"
          class="btn btn-primary"
          data-testid="batch-test-start"
          :disabled="!canStart"
          @click="runBatch"
        >
          <Icon :name="doneCount > 0 ? 'refresh' : 'play'" size="sm" />
          {{ doneCount > 0 ? t('admin.accounts.batchTest.retry') : t('admin.accounts.batchTest.start') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { Account } from '@/types'

type BatchTestStatus = 'pending' | 'running' | 'success' | 'failed' | 'stopped'
type FailureCode = '401' | '429' | 'other'
type ResultFilter = 'all' | 'success' | 'failed' | FailureCode

interface BatchTestRow {
  account: Account
  status: BatchTestStatus
  result: string
  failureCode: FailureCode | null
}

interface TestEvent {
  type?: string
  text?: string
  model?: string
  success?: boolean
  error?: string
  image_url?: string
}

interface SingleAccountResult {
  status: Exclude<BatchTestStatus, 'pending' | 'running'>
  message: string
}

const props = defineProps<{
  show: boolean
  accounts: Account[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()

const rowResults = ref<BatchTestRow[]>([])
const modelMode = ref<'default' | 'custom'>('default')
const customModelId = ref('')
const concurrencyInput = ref(50)
const resultFilter = ref<ResultFilter>('all')
const isRunning = ref(false)
const isStopping = ref(false)
const stopRequested = ref(false)
const activeControllers = new Map<number, AbortController>()

const abortActiveTests = () => {
  stopRequested.value = true
  activeControllers.forEach(controller => controller.abort())
}

const totalCount = computed(() => rowResults.value.length)
const successCount = computed(() => rowResults.value.filter(row => row.status === 'success').length)
const failedCount = computed(() => rowResults.value.filter(row => row.status === 'failed').length)
const doneCount = computed(() =>
  rowResults.value.filter(row => row.status === 'success' || row.status === 'failed' || row.status === 'stopped').length
)
const progressPercent = computed(() => (totalCount.value === 0 ? 0 : Math.round((doneCount.value / totalCount.value) * 100)))
const selectedModelId = computed(() => (modelMode.value === 'custom' ? customModelId.value.trim() : ''))
const effectiveConcurrency = computed(() => {
  const value = Number(concurrencyInput.value)
  if (!Number.isFinite(value)) return 1
  return Math.max(1, Math.min(50, Math.floor(value)))
})
const canStart = computed(() => {
  if (isRunning.value || totalCount.value === 0) return false
  if (modelMode.value === 'custom' && !customModelId.value.trim()) return false
  return true
})

const filterCount = (filter: ResultFilter) => {
  switch (filter) {
    case 'all':
      return totalCount.value
    case 'success':
      return successCount.value
    case 'failed':
      return failedCount.value
    case '401':
    case '429':
    case 'other':
      return rowResults.value.filter(row => row.status === 'failed' && row.failureCode === filter).length
  }
}

const resultFilters = computed<Array<{ value: ResultFilter; label: string }>>(() => [
  { value: 'all', label: t('admin.accounts.batchTest.filters.all', { count: filterCount('all') }) },
  { value: 'success', label: t('admin.accounts.batchTest.filters.success', { count: filterCount('success') }) },
  { value: 'failed', label: t('admin.accounts.batchTest.filters.failed', { count: filterCount('failed') }) },
  { value: '401', label: t('admin.accounts.batchTest.filters.unauthorized', { count: filterCount('401') }) },
  { value: '429', label: t('admin.accounts.batchTest.filters.rateLimited', { count: filterCount('429') }) },
  { value: 'other', label: t('admin.accounts.batchTest.filters.otherFailed', { count: filterCount('other') }) }
])

const filteredRows = computed(() => {
  switch (resultFilter.value) {
    case 'all':
      return rowResults.value
    case 'success':
      return rowResults.value.filter(row => row.status === 'success')
    case 'failed':
      return rowResults.value.filter(row => row.status === 'failed')
    case '401':
    case '429':
    case 'other':
      return rowResults.value.filter(row => row.status === 'failed' && row.failureCode === resultFilter.value)
  }
  return rowResults.value
})

const resetRows = () => {
  rowResults.value = props.accounts.map(account => ({
    account,
    status: 'pending',
    result: '',
    failureCode: null
  }))
  resultFilter.value = 'all'
  modelMode.value = 'default'
  customModelId.value = ''
  concurrencyInput.value = 50
  stopRequested.value = false
  isRunning.value = false
  isStopping.value = false
  activeControllers.clear()
}

watch(
  () => props.show,
  (show) => {
    if (show) {
      resetRows()
    } else {
      abortActiveTests()
    }
  },
  { immediate: true }
)

watch(
  () => props.accounts,
  () => {
    if (props.show && !isRunning.value) {
      resetRows()
    }
  }
)

const updateRow = (accountId: number, patch: Partial<Omit<BatchTestRow, 'account'>>) => {
  const row = rowResults.value.find(item => item.account.id === accountId)
  if (row) Object.assign(row, patch)
}

const markPendingAsStopped = () => {
  rowResults.value.forEach(row => {
    if (row.status === 'pending') {
      row.status = 'stopped'
      row.result = t('admin.accounts.batchTest.stoppedResult')
      row.failureCode = null
    }
  })
}

const stopBatch = () => {
  if (!isRunning.value) return
  isStopping.value = true
  abortActiveTests()
  markPendingAsStopped()
}

const handleClose = () => {
  if (isRunning.value) {
    stopBatch()
  } else {
    abortActiveTests()
  }
  emit('close')
}

const runBatch = async () => {
  if (!canStart.value) return

  rowResults.value.forEach(row => {
    row.status = 'pending'
    row.result = ''
    row.failureCode = null
  })
  resultFilter.value = 'all'
  stopRequested.value = false
  isRunning.value = true
  isStopping.value = false

  let nextIndex = 0
  const accountsToTest = rowResults.value.map(row => row.account)
  const workerCount = Math.min(effectiveConcurrency.value, accountsToTest.length)

  const runWorker = async () => {
    while (!stopRequested.value && nextIndex < accountsToTest.length) {
      const account = accountsToTest[nextIndex]
      nextIndex += 1
      updateRow(account.id, {
        status: 'running',
        result: t('admin.accounts.batchTest.runningResult'),
        failureCode: null
      })

      const result = await testSingleAccount(account)
      updateRow(account.id, {
        status: result.status,
        result: result.message,
        failureCode: result.status === 'failed' ? extractFailureCode(result.message) : null
      })
    }
  }

  try {
    await Promise.all(Array.from({ length: workerCount }, () => runWorker()))
  } finally {
    if (stopRequested.value) markPendingAsStopped()
    activeControllers.clear()
    isRunning.value = false
    isStopping.value = false
  }
}

const testSingleAccount = async (account: Account): Promise<SingleAccountResult> => {
  const controller = new AbortController()
  activeControllers.set(account.id, controller)
  let streamedText = ''
  let lastModel = ''

  try {
    const response = await fetch(`/api/v1/admin/accounts/${account.id}/test`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${localStorage.getItem('auth_token')}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        model_id: selectedModelId.value,
        prompt: '',
        mode: 'default'
      }),
      signal: controller.signal
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error(t('admin.accounts.batchTest.noResponseBody'))
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        const parsed = parseSseLine(line)
        if (!parsed) continue

        const result = handleTestEvent(parsed, {
          appendText: (text) => { streamedText += text },
          setModel: (model) => { lastModel = model },
          getText: () => streamedText
        })
        if (result) {
          await reader.cancel().catch(() => undefined)
          return result
        }
      }
    }

    if (streamedText.trim()) {
      return {
        status: 'success',
        message: streamedText.trim()
      }
    }

    return {
      status: 'failed',
      message: lastModel
        ? t('admin.accounts.batchTest.streamEndedWithModel', { model: lastModel })
        : t('admin.accounts.batchTest.streamEnded')
    }
  } catch (error: unknown) {
    if (isAbortError(error)) {
      return {
        status: 'stopped',
        message: t('admin.accounts.batchTest.stoppedResult')
      }
    }
    return {
      status: 'failed',
      message: error instanceof Error ? error.message : String(error)
    }
  } finally {
    activeControllers.delete(account.id)
  }
}

const parseSseLine = (line: string): TestEvent | null => {
  if (!line.startsWith('data: ')) return null
  const jsonStr = line.slice(6).trim()
  if (!jsonStr) return null
  try {
    return JSON.parse(jsonStr) as TestEvent
  } catch {
    return null
  }
}

const handleTestEvent = (
  event: TestEvent,
  state: {
    appendText: (text: string) => void
    setModel: (model: string) => void
    getText: () => string
  }
): SingleAccountResult | null => {
  switch (event.type) {
    case 'test_start':
      if (event.model) state.setModel(event.model)
      return null
    case 'content':
      if (event.text) state.appendText(event.text)
      return null
    case 'image':
      return null
    case 'test_complete':
      if (event.success) {
        const text = state.getText().trim()
        return {
          status: 'success',
          message: text || t('admin.accounts.batchTest.successResult')
        }
      }
      return {
        status: 'failed',
        message: event.error || t('admin.accounts.testFailed')
      }
    case 'error':
      return {
        status: 'failed',
        message: event.error || t('admin.accounts.testFailed')
      }
    default:
      return null
  }
}

const extractFailureCode = (message: string): FailureCode => {
  if (/\b401\b/.test(message)) return '401'
  if (/\b429\b/.test(message)) return '429'
  return 'other'
}

const isAbortError = (error: unknown) => {
  return (
    (error instanceof DOMException && error.name === 'AbortError') ||
    (typeof error === 'object' && error !== null && (error as { name?: string }).name === 'AbortError')
  )
}

const statusBadgeClass = (status: BatchTestStatus) => {
  const base = 'inline-flex items-center gap-1 rounded-full px-2 py-1 text-xs font-medium'
  switch (status) {
    case 'pending':
      return `${base} bg-gray-100 text-gray-600 dark:bg-dark-600 dark:text-gray-300`
    case 'running':
      return `${base} bg-sky-100 text-sky-700 dark:bg-sky-500/15 dark:text-sky-300`
    case 'success':
      return `${base} bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300`
    case 'failed':
      return `${base} bg-rose-100 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300`
    case 'stopped':
      return `${base} bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300`
  }
}
</script>

<style scoped>
.batch-test-stat {
  @apply rounded-lg border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800;
}

.batch-test-stat span {
  @apply text-sm text-gray-500 dark:text-gray-400;
}

.batch-test-stat strong {
  @apply mt-2 block text-3xl font-semibold leading-none text-gray-900 dark:text-gray-100;
}

.batch-test-th {
  @apply px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-300;
}

.batch-test-td {
  @apply px-4 py-3 align-middle text-sm;
}
</style>
