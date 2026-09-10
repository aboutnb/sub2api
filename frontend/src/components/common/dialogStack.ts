import {
  computed,
  nextTick,
  onMounted,
  onUnmounted,
  ref,
  watch,
  type Ref,
} from 'vue'

export const dialogStack = ref<symbol[]>([])

let savedBodyPaddingRight: string | null = null

function lockBodyScroll() {
  if (typeof document === 'undefined' || savedBodyPaddingRight !== null) return

  savedBodyPaddingRight = document.body.style.paddingRight
  const documentWidth = document.documentElement.clientWidth
  const scrollbarWidth = documentWidth > 0
    ? Math.max(0, window.innerWidth - documentWidth)
    : 0

  if (scrollbarWidth > 0) {
    const currentPadding = Number.parseFloat(window.getComputedStyle(document.body).paddingRight) || 0
    document.body.style.paddingRight = `${currentPadding + scrollbarWidth}px`
  }

  document.body.classList.add('modal-open')
}

function unlockBodyScroll() {
  if (typeof document === 'undefined' || savedBodyPaddingRight === null) return

  document.body.style.paddingRight = savedBodyPaddingRight
  document.body.classList.remove('modal-open')
  savedBodyPaddingRight = null
}

export function registerDialog(token: symbol) {
  if (dialogStack.value.includes(token)) return
  if (dialogStack.value.length === 0) lockBodyScroll()
  dialogStack.value = [...dialogStack.value, token]
}

export function unregisterDialog(token: symbol) {
  if (!dialogStack.value.includes(token)) return
  dialogStack.value = dialogStack.value.filter((entry) => entry !== token)
  if (dialogStack.value.length === 0) unlockBodyScroll()
}

export function isDialogTopmost(token: symbol) {
  return dialogStack.value.at(-1) === token
}

// Existing callers use base layers between 40 and 140. A full layer step keeps
// opening order authoritative even when two dialog types request different bases.
const DIALOG_LAYER_STEP = 200

export function getDialogZIndex(token: symbol, baseZIndex: number) {
  const stackIndex = dialogStack.value.indexOf(token)
  return baseZIndex + Math.max(0, stackIndex) * DIALOG_LAYER_STEP
}

const FOCUSABLE_SELECTOR = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[contenteditable="true"]',
  '[tabindex]:not([tabindex="-1"])',
].join(', ')

interface ManagedDialogOptions {
  open: Readonly<Ref<boolean>>
  dialogRef: Ref<HTMLElement | null>
  onClose: () => void
  closeOnEscape?: boolean
  closeOnClickOutside?: boolean
  initialFocus?: () => HTMLElement | null
  baseZIndex?: number
  name?: string
}

/**
 * Adds the shared dialog-stack, focus, Escape and scroll-lock behavior to
 * purpose-built dialogs whose visual markup cannot be represented by BaseDialog.
 */
export function useManagedDialog({
  open,
  dialogRef,
  onClose,
  closeOnEscape = true,
  closeOnClickOutside = false,
  initialFocus,
  baseZIndex = 50,
  name = 'dialog',
}: ManagedDialogOptions) {
  const token = Symbol(name)
  let previousActiveElement: HTMLElement | null = null
  let mounted = false

  const isTopmost = computed(() => isDialogTopmost(token))
  const zIndexStyle = computed(() => ({
    zIndex: getDialogZIndex(token, baseZIndex),
  }))

  function getFocusableElements(): HTMLElement[] {
    if (!dialogRef.value) return []

    return Array.from(dialogRef.value.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)).filter(
      (element) => {
        if (element.getAttribute('tabindex') === '-1') return false
        if (element.closest('[inert], [aria-hidden="true"]')) return false
        return element.getAttribute('aria-hidden') !== 'true'
      },
    )
  }

  function focusDialog() {
    if (!open.value || !isTopmost.value || !dialogRef.value) return
    const target = initialFocus?.() || getFocusableElements()[0] || dialogRef.value
    target.focus()
  }

  function requestClose() {
    if (!open.value || !isTopmost.value) return
    onClose()
  }

  function handleBackdropClick() {
    if (closeOnClickOutside) requestClose()
  }

  function handleKeydown(event: KeyboardEvent) {
    if (!open.value || !isTopmost.value || !dialogRef.value) return

    if (closeOnEscape && event.key === 'Escape') {
      event.preventDefault()
      event.stopPropagation()
      onClose()
      return
    }

    if (event.key !== 'Tab') return

    const focusableElements = getFocusableElements()
    if (focusableElements.length === 0) {
      event.preventDefault()
      dialogRef.value.focus()
      return
    }

    const firstElement = focusableElements[0]
    const lastElement = focusableElements[focusableElements.length - 1]
    const activeElement = document.activeElement

    if (
      event.shiftKey
      && (activeElement === firstElement || !dialogRef.value.contains(activeElement))
    ) {
      event.preventDefault()
      lastElement.focus()
    } else if (
      !event.shiftKey
      && (activeElement === lastElement || !dialogRef.value.contains(activeElement))
    ) {
      event.preventDefault()
      firstElement.focus()
    }
  }

  watch(
    open,
    async (isOpen) => {
      if (isOpen) {
        previousActiveElement = document.activeElement as HTMLElement | null
        registerDialog(token)
        await nextTick()
        if (mounted) focusDialog()
        return
      }

      const shouldRestoreFocus = isDialogTopmost(token)
      const elementToRestore = previousActiveElement
      previousActiveElement = null
      unregisterDialog(token)

      if (shouldRestoreFocus && elementToRestore?.isConnected) {
        await nextTick()
        if (elementToRestore.isConnected) elementToRestore.focus()
      }
    },
    { immediate: true },
  )

  onMounted(() => {
    mounted = true
    document.addEventListener('keydown', handleKeydown)
    // Immediate watchers run during setup, before template refs exist.
    // The mounted pass guarantees initial focus for dialogs rendered open.
    if (open.value) focusDialog()
  })

  onUnmounted(() => {
    mounted = false
    document.removeEventListener('keydown', handleKeydown)
    const shouldRestoreFocus = isDialogTopmost(token)
    unregisterDialog(token)
    if (shouldRestoreFocus && previousActiveElement?.isConnected) {
      previousActiveElement.focus()
    }
    previousActiveElement = null
  })

  return {
    isTopmost,
    zIndexStyle,
    requestClose,
    handleBackdropClick,
  }
}
