<template>
  <Select
    :model-value="modelValue === 'all' && includeAll ? '' : modelValue"
    :options="options"
    :aria-label="t('supportTickets.category')"
    :disabled="disabled"
    @update:model-value="updateValue"
    @change="emit('change')"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'

const props = withDefaults(defineProps<{ modelValue: string; includeAll?: boolean; disabled?: boolean }>(), {
  includeAll: true,
  disabled: false
})
const emit = defineEmits<{ 'update:modelValue': [value: string]; change: [] }>()
const { t } = useI18n()
const options = computed(() => [
  ...(props.includeAll ? [{ value: '', label: t('supportTickets.allCategories') }] : []),
  { value: 'billing', label: t('supportTickets.categories.billing') },
  { value: 'api_models', label: t('supportTickets.categories.apiModels') },
  { value: 'account', label: t('supportTickets.categories.account') },
  { value: 'other', label: t('supportTickets.categories.other') }
])
function updateValue(value: string | number | boolean | null) {
  emit('update:modelValue', String(value ?? ''))
}
</script>
