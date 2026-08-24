<template>
  <div class="space-y-3">
    <div>
      <label class="mb-1.5 block text-xs font-medium text-gray-500 dark:text-gray-400">
        {{ t('payment.customAmount') }}
      </label>
      <div class="relative">
        <span class="absolute left-4 top-1/2 -translate-y-1/2 text-lg font-medium text-gray-400 dark:text-dark-500">
          {{ currencySymbol }}
        </span>
        <input
          type="text"
          inputmode="decimal"
          :value="customText"
          :placeholder="placeholderText"
          class="input h-14 w-full rounded-md pl-9 pr-4 text-xl font-semibold tabular-nums"
          @input="handleInput"
        />
      </div>
    </div>

    <div>
      <label class="mb-1.5 block text-xs font-medium text-gray-500 dark:text-gray-400">
        {{ t('payment.quickAmounts') }}
      </label>
      <div data-testid="quick-amounts" class="scrollbar-hide grid grid-flow-col auto-cols-[minmax(4.5rem,1fr)] gap-1.5 overflow-x-auto pb-0.5">
        <button
          v-for="amt in filteredAmounts"
          :key="amt"
          type="button"
          :aria-pressed="modelValue === amt"
          :class="[
            'h-9 rounded-md border px-2 text-center text-sm font-semibold tabular-nums transition-colors',
            modelValue === amt
              ? 'border-primary-500 bg-primary-50/70 text-primary-700 ring-1 ring-primary-500/20 dark:border-primary-400 dark:bg-dark-900 dark:text-primary-300'
              : 'border-gray-300 bg-transparent text-gray-700 hover:border-gray-500 dark:border-dark-600 dark:text-gray-200 dark:hover:border-dark-500',
          ]"
          @click="selectAmount(amt)"
        >
          {{ currencySymbol }}{{ amt }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const props = withDefaults(defineProps<{
  amounts?: number[]
  modelValue: number | null
  min?: number
  max?: number
  currencySymbol?: string
}>(), {
  amounts: () => [10, 20, 50, 100, 200, 500],
  min: 0,
  max: 0,
  currencySymbol: '$',
})

const emit = defineEmits<{
  'update:modelValue': [value: number | null]
}>()

const { t } = useI18n()

const customText = ref('')

// 0 = no limit
const filteredAmounts = computed(() =>
  props.amounts.filter((a) => (props.min <= 0 || a >= props.min) && (props.max <= 0 || a <= props.max))
)

const placeholderText = computed(() => {
  if (props.min > 0 && props.max > 0) return `${props.min} - ${props.max}`
  if (props.min > 0) return `≥ ${props.min}`
  if (props.max > 0) return `≤ ${props.max}`
  return t('payment.enterAmount')
})

const AMOUNT_PATTERN = /^\d*(\.\d{0,2})?$/

function selectAmount(amt: number) {
  customText.value = String(amt)
  emit('update:modelValue', amt)
}

function handleInput(e: Event) {
  const val = (e.target as HTMLInputElement).value
  if (!AMOUNT_PATTERN.test(val)) return
  customText.value = val
  if (val === '') {
    emit('update:modelValue', null)
    return
  }
  const num = parseFloat(val)
  if (!isNaN(num) && num > 0) {
    emit('update:modelValue', num)
  } else {
    emit('update:modelValue', null)
  }
}

watch(() => props.modelValue, (v) => {
  if (v !== null && String(v) !== customText.value) {
    customText.value = String(v)
  }
}, { immediate: true })
</script>
