<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.dataImportTitle')"
    width="normal"
    close-on-click-outside
    @close="handleClose"
  >
    <form id="import-data-form" class="space-y-4" @submit.prevent="handleImport">
      <div class="text-sm text-gray-600 dark:text-dark-300">
        {{ t('admin.accounts.dataImportHint') }}
      </div>
      <div
        class="rounded-lg border border-amber-200 bg-amber-50 p-3 text-xs text-amber-600 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-400"
      >
        {{ t('admin.accounts.dataImportWarning') }}
      </div>

      <div
        v-if="groupsLoading"
        class="rounded-lg border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-400"
      >
        {{ t('admin.accounts.dataImportGroupsLoading') }}
      </div>
      <GroupSelector v-else v-model="groupIds" :groups="groups" searchable />

      <div>
        <label class="input-label">{{ t('admin.accounts.dataImportFile') }}</label>
        <div
          class="flex items-center justify-between gap-3 rounded-lg border border-dashed border-gray-300 bg-gray-50 px-4 py-3 dark:border-dark-600 dark:bg-dark-800"
        >
          <div class="min-w-0">
            <div class="truncate text-sm text-gray-700 dark:text-dark-200">
              {{ fileName || t('admin.accounts.dataImportSelectFile') }}
            </div>
            <div class="text-xs text-gray-500 dark:text-dark-400">JSON (.json)</div>
          </div>
          <button type="button" class="btn btn-secondary shrink-0" @click="openFilePicker">
            {{ t('common.chooseFile') }}
          </button>
        </div>
        <input
          ref="fileInput"
          type="file"
          class="hidden"
          accept="application/json,.json"
          @change="handleFileChange"
        />
      </div>

      <label class="flex items-start gap-3 rounded-lg border border-gray-200 bg-gray-50 px-4 py-3 text-sm text-gray-700 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-200">
        <input
          v-model="useProjectMihomoPool"
          type="checkbox"
          class="mt-0.5 h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
        />
        <span>
          <span class="block font-medium">{{ t('admin.accounts.dataImportUseProjectMihomo') }}</span>
          <span class="mt-1 block text-xs text-gray-500 dark:text-dark-400">
            {{ t('admin.accounts.dataImportUseProjectMihomoHint') }}
          </span>
        </span>
      </label>

      <div
        v-if="result"
        class="space-y-2 rounded-xl border border-gray-200 p-4 dark:border-dark-700"
      >
        <div class="text-sm font-medium text-gray-900 dark:text-white">
          {{ t('admin.accounts.dataImportResult') }}
        </div>
        <div class="text-sm text-gray-700 dark:text-dark-300">
          {{ t('admin.accounts.dataImportResultSummary', result) }}
        </div>

        <div v-if="errorItems.length" class="mt-2">
          <div class="text-sm font-medium text-red-600 dark:text-red-400">
            {{ t('admin.accounts.dataImportErrors') }}
          </div>
          <div
            class="mt-2 max-h-48 overflow-auto rounded-lg bg-gray-50 p-3 font-mono text-xs dark:bg-dark-800"
          >
            <div v-for="(item, idx) in errorItems" :key="idx" class="whitespace-pre-wrap">
              {{ item.kind }} {{ item.name || item.proxy_key || '-' }} — {{ item.message }}
            </div>
          </div>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button class="btn btn-secondary" type="button" :disabled="importing" @click="handleClose">
          {{ t('common.cancel') }}
        </button>
        <button
          class="btn btn-primary"
          type="submit"
          form="import-data-form"
          :disabled="importing"
        >
          {{ importing ? t('admin.accounts.dataImporting') : t('admin.accounts.dataImportButton') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import GroupSelector from '@/components/common/GroupSelector.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { AdminDataImportResult, AdminGroup } from '@/types'

interface Props {
  show: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'imported'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

const importing = ref(false)
const file = ref<File | null>(null)
const result = ref<AdminDataImportResult | null>(null)
const useProjectMihomoPool = ref(false)
const groups = ref<AdminGroup[]>([])
const groupIds = ref<number[]>([])
const groupsLoading = ref(false)

const fileInput = ref<HTMLInputElement | null>(null)
const fileName = computed(() => file.value?.name || '')

const errorItems = computed(() => result.value?.errors || [])

const loadGroups = async () => {
  groupsLoading.value = true
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch (error) {
    groups.value = []
    appStore.showError(t('admin.accounts.dataImportGroupsLoadFailed'))
  } finally {
    groupsLoading.value = false
  }
}

watch(
  () => props.show,
  (open) => {
    if (open) {
      file.value = null
      result.value = null
      useProjectMihomoPool.value = false
      groupIds.value = []
      loadGroups()
      if (fileInput.value) {
        fileInput.value.value = ''
      }
    }
  }
)

onMounted(() => {
  if (props.show) {
    loadGroups()
  }
})

const openFilePicker = () => {
  fileInput.value?.click()
}

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  file.value = target.files?.[0] || null
}

const handleClose = () => {
  if (importing.value) return
  emit('close')
}

const readFileAsText = async (sourceFile: File): Promise<string> => {
  if (typeof sourceFile.text === 'function') {
    return sourceFile.text()
  }

  if (typeof sourceFile.arrayBuffer === 'function') {
    const buffer = await sourceFile.arrayBuffer()
    return new TextDecoder().decode(buffer)
  }

  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(reader.error || new Error('Failed to read file'))
    reader.readAsText(sourceFile)
  })
}

const handleImport = async () => {
  if (!file.value) {
    appStore.showError(t('admin.accounts.dataImportSelectFile'))
    return
  }

  importing.value = true
  try {
    const text = await readFileAsText(file.value)
    const dataPayload = JSON.parse(text)

    const res = await adminAPI.accounts.importData({
      data: dataPayload,
      skip_default_group_bind: true,
      group_ids: groupIds.value,
      proxy_provider: useProjectMihomoPool.value ? 'project_mihomo' : undefined
    })

    result.value = res

    const msgParams: Record<string, unknown> = {
      account_created: res.account_created,
      account_failed: res.account_failed,
      proxy_created: res.proxy_created,
      proxy_reused: res.proxy_reused,
      proxy_failed: res.proxy_failed,
    }
    if (res.account_failed > 0 || res.proxy_failed > 0) {
      appStore.showError(t('admin.accounts.dataImportCompletedWithErrors', msgParams))
    } else {
      appStore.showSuccess(t('admin.accounts.dataImportSuccess', msgParams))
      emit('imported')
    }
  } catch (error: any) {
    if (error instanceof SyntaxError) {
      appStore.showError(t('admin.accounts.dataImportParseFailed'))
    } else {
      appStore.showError(error?.message || t('admin.accounts.dataImportFailed'))
    }
  } finally {
    importing.value = false
  }
}
</script>
