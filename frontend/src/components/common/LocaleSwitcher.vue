<template>
  <div class="relative" ref="dropdownRef">
    <button
      ref="triggerRef"
      type="button"
      @click="toggleDropdown"
      :disabled="switching"
      class="btn btn-secondary min-h-11 min-w-11 gap-2 px-3 py-2 text-sm"
      :title="currentLocale?.name"
      :aria-label="currentLocale?.name"
      :aria-expanded="isOpen"
      aria-controls="locale-switcher-menu"
      aria-haspopup="menu"
    >
      <span class="text-base" aria-hidden="true">{{ currentLocale?.flag }}</span>
      <span class="hidden sm:inline">{{ currentLocale?.code.toUpperCase() }}</span>
      <Icon
        name="chevronDown"
        size="xs"
        class="text-ink-muted transition-transform duration-200"
        :class="{ 'rotate-180': isOpen }"
      />
    </button>

    <transition name="dropdown">
      <div
        v-if="isOpen"
        id="locale-switcher-menu"
        ref="menuRef"
        role="menu"
        class="dropdown right-0 mt-2 w-40 p-1"
      >
        <button
          v-for="locale in availableLocales"
          :key="locale.code"
          type="button"
          role="menuitemradio"
          :aria-checked="locale.code === currentLocaleCode"
          :disabled="switching"
          @click="selectLocale(locale.code)"
          class="dropdown-item mx-0 min-h-11 w-full"
          :class="{
            'border border-primary-700 bg-primary-100 text-primary-800 shadow-pixel-sm dark:border-primary-500 dark:bg-primary-900/30 dark:text-primary-200':
              locale.code === currentLocaleCode
          }"
        >
          <span class="text-base" aria-hidden="true">{{ locale.flag }}</span>
          <span>{{ locale.name }}</span>
          <Icon v-if="locale.code === currentLocaleCode" name="check" size="sm" class="ml-auto text-primary-500" />
        </button>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { setLocale, availableLocales } from '@/i18n'

const { locale } = useI18n()

const isOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const menuRef = ref<HTMLElement | null>(null)
const switching = ref(false)

const currentLocaleCode = computed(() => locale.value)
const currentLocale = computed(() => availableLocales.find((l) => l.code === locale.value))

async function toggleDropdown() {
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    await nextTick()
    const activeItem = menuRef.value?.querySelector<HTMLElement>('[aria-checked="true"]')
    const firstItem = menuRef.value?.querySelector<HTMLElement>('[role="menuitemradio"]')
    ;(activeItem || firstItem)?.focus()
  }
}

function closeDropdown(restoreFocus = false) {
  isOpen.value = false
  if (restoreFocus) {
    nextTick(() => triggerRef.value?.focus())
  }
}

async function selectLocale(code: string) {
  if (switching.value || code === currentLocaleCode.value) {
    closeDropdown(true)
    return
  }
  switching.value = true
  try {
    await setLocale(code)
    closeDropdown(true)
  } finally {
    switching.value = false
  }
}

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    isOpen.value = false
  }
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && isOpen.value) {
    event.preventDefault()
    closeDropdown(true)
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}
</style>
