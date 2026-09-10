<template>
  <BaseDialog
    :show="show"
    :title="title"
    width="narrow"
    :pending="pending"
    @close="handleCancel"
  >
    <div class="space-y-4">
      <p class="text-sm text-ink dark:text-ink-muted">{{ message }}</p>
      <slot></slot>
    </div>

    <template #footer>
      <div class="flex justify-end space-x-3">
        <button
          @click="handleCancel"
          type="button"
          class="btn btn-secondary"
          :disabled="pending || cancelDisabled"
        >
          {{ cancelText }}
        </button>
        <button
          @click="handleConfirm"
          type="button"
          :disabled="pending || confirmDisabled"
          :aria-busy="pending || undefined"
          :class="[
            'btn',
            danger
              ? 'btn-danger'
              : 'btn-primary'
          ]"
        >
          <span
            v-if="pending"
            class="h-4 w-4 animate-spin rounded-full border-2 border-current border-r-transparent"
            aria-hidden="true"
          ></span>
          {{ confirmText }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from './BaseDialog.vue'

const { t } = useI18n()

interface Props {
  show: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
  pending?: boolean
  confirmDisabled?: boolean
  cancelDisabled?: boolean
}

interface Emits {
  (e: 'confirm'): void
  (e: 'cancel'): void
}

const props = withDefaults(defineProps<Props>(), {
  danger: false,
  pending: false,
  confirmDisabled: false,
  cancelDisabled: false
})

const confirmText = computed(() => props.confirmText || t('common.confirm'))
const cancelText = computed(() => props.cancelText || t('common.cancel'))

const emit = defineEmits<Emits>()

const handleConfirm = () => {
  if (props.pending || props.confirmDisabled) return
  emit('confirm')
}

const handleCancel = () => {
  if (props.pending || props.cancelDisabled) return
  emit('cancel')
}
</script>
