<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="show"
        class="modal-overlay"
        :style="zIndexStyle"
        :aria-labelledby="dialogId"
        :aria-describedby="ariaDescribedby"
        :aria-busy="pending || undefined"
        :aria-hidden="isTopmost ? undefined : 'true'"
        :inert="isTopmost ? undefined : true"
        role="dialog"
        aria-modal="true"
        @click.self="handleClose"
      >
        <!-- Modal panel -->
        <div
          ref="dialogRef"
          :class="['modal-content', widthClasses]"
          tabindex="-1"
          @click.stop
        >
          <!-- Header -->
          <div class="modal-header">
            <h3 :id="dialogId" class="modal-title">
              {{ title }}
            </h3>
            <button
              v-if="showCloseButton"
              type="button"
              class="btn btn-ghost btn-icon -mr-1 shrink-0"
              :disabled="pending"
              :aria-label="t('common.close')"
              @click="requestClose"
            >
              <Icon name="x" size="md" />
            </button>
          </div>

          <!-- Body -->
          <div ref="modalBodyRef" class="modal-body">
            <slot></slot>
          </div>

          <!-- Footer -->
          <div v-if="$slots.footer" class="modal-footer">
            <slot name="footer"></slot>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, watch, onMounted, onUnmounted, ref, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { getDialogZIndex, isDialogTopmost, registerDialog, unregisterDialog } from './dialogStack'

const { t } = useI18n()

// Vue instance UIDs keep multiple simultaneously mounted dialogs uniquely labelled.
const dialogId = `modal-title-${getCurrentInstance()?.uid ?? 0}`

// 焦点管理
const dialogRef = ref<HTMLElement | null>(null)
const modalBodyRef = ref<HTMLElement | null>(null)
let previousActiveElement: HTMLElement | null = null
const dialogToken = Symbol(dialogId)

type DialogWidth = 'narrow' | 'normal' | 'wide' | 'extra-wide' | 'full'

interface Props {
  show: boolean
  title: string
  width?: DialogWidth
  closeOnEscape?: boolean
  closeOnClickOutside?: boolean
  showCloseButton?: boolean
  zIndex?: number
  pending?: boolean
  ariaDescribedby?: string
}

interface Emits {
  (e: 'close'): void
}

const props = withDefaults(defineProps<Props>(), {
  width: 'normal',
  closeOnEscape: true,
  closeOnClickOutside: false,
  showCloseButton: true,
  zIndex: 50,
  pending: false
})

const emit = defineEmits<Emits>()

const isTopmost = computed(() => isDialogTopmost(dialogToken))

// Keep stacked dialogs in a deterministic layer order, even when callers use the default z-index.
const zIndexStyle = computed(() => {
  return { zIndex: getDialogZIndex(dialogToken, props.zIndex) }
})

const widthClasses = computed(() => {
  // Width guidance: narrow=confirm/short prompts, normal=standard forms,
  // wide=multi-section forms or rich content, extra-wide=analytics/tables,
  // full=full-screen or very dense layouts.
  const widths: Record<DialogWidth, string> = {
    narrow: 'max-w-md',
    normal: 'max-w-lg',
    wide: 'w-full sm:max-w-2xl md:max-w-3xl lg:max-w-4xl',
    'extra-wide': 'w-full sm:max-w-3xl md:max-w-4xl lg:max-w-5xl xl:max-w-6xl',
    full: 'w-full sm:max-w-4xl md:max-w-5xl lg:max-w-6xl xl:max-w-7xl'
  }
  return widths[props.width]
})

const handleClose = () => {
  if (props.closeOnClickOutside && !props.pending && isTopmost.value) {
    emit('close')
  }
}

const requestClose = () => {
  if (props.pending || !isTopmost.value) return
  emit('close')
}

const getFocusableElements = () => {
  if (!dialogRef.value) return []

  return Array.from(
    dialogRef.value.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
    )
  ).filter((element) => {
    if (element.getAttribute('tabindex') === '-1') return false
    if (element.closest('[inert], [aria-hidden="true"]')) return false
    return element.getAttribute('aria-hidden') !== 'true'
  })
}

const handleKeydown = (event: KeyboardEvent) => {
  if (!props.show || !isTopmost.value) return

  if (props.closeOnEscape && !props.pending && event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    emit('close')
    return
  }

  if (event.key !== 'Tab' || !dialogRef.value) return

  const focusableElements = getFocusableElements()
  if (focusableElements.length === 0) {
    event.preventDefault()
    dialogRef.value.focus()
    return
  }

  const firstElement = focusableElements[0]
  const lastElement = focusableElements[focusableElements.length - 1]
  const activeElement = document.activeElement

  if (event.shiftKey && (activeElement === firstElement || !dialogRef.value.contains(activeElement))) {
    event.preventDefault()
    lastElement.focus()
  } else if (!event.shiftKey && activeElement === lastElement) {
    event.preventDefault()
    firstElement.focus()
  }
}

// Prevent body scroll when modal is open and manage focus
watch(
  () => props.show,
  async (isOpen) => {
    if (isOpen) {
      // 保存当前焦点元素
      previousActiveElement = document.activeElement as HTMLElement
      registerDialog(dialogToken)

      // 等待DOM更新后设置焦点到对话框
      await nextTick()
      if (modalBodyRef.value) {
        modalBodyRef.value.scrollTop = 0
      }
      if (props.show && isTopmost.value && dialogRef.value) {
        const firstFocusable = getFocusableElements()[0]
        ;(firstFocusable || dialogRef.value).focus()
      }
    } else {
      const shouldRestoreFocus = isDialogTopmost(dialogToken)
      const elementToRestore = previousActiveElement
      previousActiveElement = null
      unregisterDialog(dialogToken)
      // Only the topmost dialog owns focus restoration. Closing a covered dialog
      // must not move focus out from under the active one.
      if (
        shouldRestoreFocus
        && elementToRestore?.isConnected
        && typeof elementToRestore.focus === 'function'
      ) {
        await nextTick()
        if (elementToRestore.isConnected) elementToRestore.focus()
      }
    }
  },
  { immediate: true }
)

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
  const shouldRestoreFocus = isDialogTopmost(dialogToken)
  unregisterDialog(dialogToken)
  if (shouldRestoreFocus && previousActiveElement?.isConnected) {
    previousActiveElement.focus()
  }
  previousActiveElement = null
})
</script>
