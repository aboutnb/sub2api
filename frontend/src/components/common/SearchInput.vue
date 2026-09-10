<template>
  <div class="w-full">
    <label v-if="label" :for="inputId" class="input-label mb-1.5 block">{{ label }}</label>
    <div class="relative w-full">
      <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
        <Icon name="search" size="md" class="text-ink-muted" aria-hidden="true" />
      </div>
      <input
        :id="inputId"
        :name="name"
        :value="modelValue"
        type="search"
        class="input w-full pl-10"
        :placeholder="placeholder"
        :disabled="disabled"
        :autocomplete="autocomplete"
        :aria-label="ariaLabel || (!label ? placeholder : undefined)"
        :aria-describedby="hint ? hintId : undefined"
        @input="handleInput"
      />
    </div>
    <p v-if="hint" :id="hintId" class="input-hint mt-1.5">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
import { useDebounceFn } from '@vueuse/core'
import { computed, getCurrentInstance } from 'vue'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  debounceMs?: number
  id?: string
  name?: string
  label?: string
  hint?: string
  ariaLabel?: string
  autocomplete?: string
  disabled?: boolean
}>(), {
  placeholder: 'Search...',
  debounceMs: 300,
  autocomplete: 'off',
  disabled: false
})

const instanceUid = getCurrentInstance()?.uid ?? 0
const inputId = computed(() => props.id || `search-input-${instanceUid}`)
const hintId = computed(() => `${inputId.value}-hint`)

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'search', value: string): void
}>()

const debouncedEmitSearch = useDebounceFn((value: string) => {
  emit('search', value)
}, props.debounceMs)

const handleInput = (event: Event) => {
  if (props.disabled) return
  const value = (event.target as HTMLInputElement).value
  emit('update:modelValue', value)
  debouncedEmitSearch(value)
}
</script>
