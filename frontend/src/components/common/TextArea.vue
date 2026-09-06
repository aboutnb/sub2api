<template>
  <div class="w-full">
    <label v-if="label" :for="textAreaId" class="input-label mb-1.5 block">
      {{ label }}
      <span v-if="required" class="text-red-500" aria-hidden="true">*</span>
    </label>
    <div class="relative">
      <textarea
        :id="textAreaId"
        ref="textAreaRef"
        :name="name"
        :value="modelValue"
        :disabled="disabled"
        :required="required"
        :placeholder="placeholderText"
        :autocomplete="autocomplete"
        :readonly="readonly"
        :rows="rows"
        :aria-label="ariaLabel || (!label ? placeholderText : undefined)"
        :aria-required="required || undefined"
        :aria-invalid="error ? 'true' : undefined"
        :aria-describedby="describedBy"
        :class="[
          'input w-full min-h-[80px] transition-all duration-200 resize-y',
          error ? 'input-error ring-2 ring-red-500/20' : '',
          disabled ? 'cursor-not-allowed bg-surface-muted opacity-60 dark:bg-canvas' : ''
        ]"
        @input="onInput"
        @change="$emit('change', ($event.target as HTMLTextAreaElement).value)"
        @blur="$emit('blur', $event)"
        @focus="$emit('focus', $event)"
      ></textarea>
    </div>
    <!-- Hint / Error Text -->
    <p v-if="error" :id="supportTextId" class="input-error-text mt-1.5" role="alert">
      {{ error }}
    </p>
    <p v-else-if="hint" :id="supportTextId" class="input-hint mt-1.5">
      {{ hint }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, ref } from 'vue'

interface Props {
  modelValue: string | null | undefined
  label?: string
  placeholder?: string
  disabled?: boolean
  required?: boolean
  readonly?: boolean
  error?: string
  hint?: string
  id?: string
  name?: string
  ariaLabel?: string
  ariaDescribedby?: string
  autocomplete?: string
  rows?: number | string
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  required: false,
  readonly: false,
  rows: 3
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
  (e: 'blur', event: FocusEvent): void
  (e: 'focus', event: FocusEvent): void
}>()

const textAreaRef = ref<HTMLTextAreaElement | null>(null)
const instanceUid = getCurrentInstance()?.uid ?? 0
const textAreaId = computed(() => props.id || `textarea-${instanceUid}`)
const placeholderText = computed(() => props.placeholder || '')
const supportTextId = computed(() => {
  if (!props.error && !props.hint) return undefined
  return `${textAreaId.value}-${props.error ? 'error' : 'hint'}`
})
const describedBy = computed(() =>
  [props.ariaDescribedby, supportTextId.value].filter(Boolean).join(' ') || undefined
)

const onInput = (event: Event) => {
  const value = (event.target as HTMLTextAreaElement).value
  emit('update:modelValue', value)
}

// Expose focus method
defineExpose({
  focus: () => textAreaRef.value?.focus(),
  select: () => textAreaRef.value?.select()
})
</script>
