<template>
  <button
    type="button"
    @click="toggle"
    :id="id"
    :disabled="disabled"
    class="relative inline-flex h-11 w-12 flex-shrink-0 cursor-pointer items-center rounded-xl transition-colors duration-200 ease-in-out"
    :class="[
      disabled && 'cursor-not-allowed opacity-50'
    ]"
    role="switch"
    :aria-checked="modelValue"
    :aria-label="ariaLabel"
    :aria-labelledby="ariaLabelledby"
    :aria-describedby="ariaDescribedby"
  >
    <span
      aria-hidden="true"
      class="pointer-events-none absolute inset-x-0 top-1/2 h-7 -translate-y-1/2 rounded-lg border border-line-control transition-colors duration-200 ease-in-out"
      :class="modelValue ? 'bg-primary-500' : 'bg-line dark:bg-line-strong'"
    />
    <span
      class="pointer-events-none absolute left-0.5 top-1/2 inline-block h-5 w-5 -translate-y-1/2 transform rounded-md border border-line-control bg-white shadow-sm ring-0 transition duration-200 ease-in-out"
      :class="[modelValue ? 'translate-x-5' : 'translate-x-0']"
    />
  </button>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean
  id?: string
  disabled?: boolean
  ariaLabel?: string
  ariaLabelledby?: string
  ariaDescribedby?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

function toggle() {
  if (props.disabled) return
  emit('update:modelValue', !props.modelValue)
}
</script>
