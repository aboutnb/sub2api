<template>
  <div>
    <!-- Tags display -->
    <div class="flex min-h-11 flex-wrap items-center gap-1.5 rounded-xl border border-line-control bg-surface px-2 py-[7px] focus-within:border-brand">
      <span
        v-for="(model, idx) in models"
        :key="idx"
        class="inline-flex max-w-full items-center gap-1 rounded-md px-2 py-0.5 text-xs leading-5"
        :class="getPlatformTagClass(props.platform || '')"
      >
        <span class="min-w-0 [overflow-wrap:anywhere]">{{ model }}</span>
        <button
          type="button"
          @click="removeModel(idx)"
          :aria-label="`${t('common.remove')} ${model}`"
          class="model-tag-remove ml-0.5 flex h-6 min-h-0 w-6 min-w-0 shrink-0 items-center justify-center rounded hover:bg-primary-200 dark:hover:bg-primary-800"
        >
          <Icon name="x" size="xs" />
        </button>
      </span>
      <input
        ref="inputRef"
        v-model="inputValue"
        type="text"
        class="h-7 min-h-0 min-w-[120px] max-w-full flex-1 border-none bg-transparent p-0 text-sm leading-5 outline-none placeholder:text-ink-muted dark:text-white"
        :placeholder="models.length === 0 ? placeholder : ''"
        @keydown.enter.prevent="addModel"
        @keydown.tab.prevent="addModel"
        @keydown.delete="handleBackspace"
        @paste="handlePaste"
        @blur="addModel"
      />
    </div>
    <p class="mt-1 text-xs text-ink-muted">
      {{ t('admin.channels.form.modelInputHint', 'Press Enter to add, supports paste for batch import.') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { getPlatformTagClass } from './types'

const { t } = useI18n()

const props = defineProps<{
  models: string[]
  placeholder?: string
  platform?: string
}>()

const emit = defineEmits<{
  'update:models': [models: string[]]
}>()

const inputValue = ref('')
const inputRef = ref<HTMLInputElement>()

function addModel() {
  const val = inputValue.value.trim()
  if (!val) return
  if (!props.models.includes(val)) {
    emit('update:models', [...props.models, val])
  }
  inputValue.value = ''
}

function removeModel(idx: number) {
  const newModels = [...props.models]
  newModels.splice(idx, 1)
  emit('update:models', newModels)
}

function handleBackspace() {
  if (inputValue.value === '' && props.models.length > 0) {
    removeModel(props.models.length - 1)
  }
}

function handlePaste(e: ClipboardEvent) {
  e.preventDefault()
  const text = e.clipboardData?.getData('text') || ''
  const items = text.split(/[,\n;]+/).map(s => s.trim()).filter(Boolean)
  if (items.length === 0) return
  const unique = [...new Set([...props.models, ...items])]
  emit('update:models', unique)
  inputValue.value = ''
}
</script>

<style scoped>
@media (pointer: coarse) {
  .model-tag-remove {
    min-width: 2.75rem;
    min-height: 2.75rem;
  }
}
</style>
