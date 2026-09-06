<template>
  <div ref="dropdownRef" class="relative">
    <button
      ref="triggerRef"
      type="button"
      class="btn btn-secondary btn-sm gap-1.5"
      :title="t('common.autoRefresh.title')"
      aria-haspopup="menu"
      :aria-expanded="showDropdown"
      :aria-controls="menuId"
      @click="toggleDropdown"
      @keydown.down.prevent="openDropdown"
    >
      <svg
        class="h-3.5 w-3.5"
        :class="enabled ? 'animate-spin' : ''"
        xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
      >
        <path fill-rule="evenodd" d="M15.312 11.424a5.5 5.5 0 01-9.201 2.466l-.312-.311h2.433a.75.75 0 000-1.5H4.598a.75.75 0 00-.75.75v3.634a.75.75 0 001.5 0v-2.033l.312.312a7 7 0 0011.712-3.138.75.75 0 00-1.449-.39zm-10.624-2.848a5.5 5.5 0 019.201-2.466l.312.311H11.768a.75.75 0 000 1.5h3.634a.75.75 0 00.75-.75V3.537a.75.75 0 00-1.5 0v2.034l-.312-.312A7 7 0 002.628 8.397a.75.75 0 001.449.39z" clip-rule="evenodd" />
      </svg>
      <span>
        {{ enabled
          ? t('common.autoRefresh.countdown', { seconds: countdown })
          : t('common.autoRefresh.title')
        }}
      </span>
    </button>

    <div
      v-if="showDropdown"
      :id="menuId"
      ref="menuRef"
      role="menu"
      :aria-label="t('common.autoRefresh.title')"
      class="dropdown absolute right-0 z-20 mt-2 w-48 p-1.5"
    >
      <div>
        <button
          type="button"
          role="menuitemcheckbox"
          :aria-checked="enabled"
          class="flex min-h-11 w-full items-center justify-between rounded-lg px-3 py-2 text-left text-sm font-medium text-ink hover:bg-primary-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500/40 dark:text-gray-200 dark:hover:bg-primary-900/20"
          @click="handleEnabledChange"
        >
          <span>{{ t('common.autoRefresh.enable') }}</span>
          <svg v-if="enabled" class="h-4 w-4 text-primary-500" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z" clip-rule="evenodd" />
          </svg>
        </button>
        <div role="separator" class="my-1 border-t-2 border-line dark:border-line"></div>
        <button
          v-for="sec in intervals"
          :key="sec"
          type="button"
          role="menuitemradio"
          :aria-checked="intervalSeconds === sec"
          class="flex min-h-11 w-full items-center justify-between rounded-lg px-3 py-2 text-left text-sm text-ink hover:bg-primary-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500/40 dark:text-gray-200 dark:hover:bg-primary-900/20"
          @click="handleIntervalChange(sec)"
        >
          <span>{{ t('common.autoRefresh.seconds', { n: sec }) }}</span>
          <svg v-if="intervalSeconds === sec" class="h-4 w-4 text-primary-500" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z" clip-rule="evenodd" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, getCurrentInstance, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  enabled: boolean
  intervalSeconds: number
  countdown: number
  intervals: readonly number[]
}>()

const emit = defineEmits<{
  (e: 'update:enabled', value: boolean): void
  (e: 'update:interval', value: number): void
}>()

const { t } = useI18n()
const showDropdown = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const menuRef = ref<HTMLElement | null>(null)

const menuId = `auto-refresh-menu-${getCurrentInstance()?.uid ?? 0}`

function focusFirstItem() {
  nextTick(() => menuRef.value?.querySelector<HTMLElement>('[role^="menuitem"]')?.focus())
}

function openDropdown() {
  if (!showDropdown.value) {
    showDropdown.value = true
  }
  focusFirstItem()
}

function closeDropdown(restoreFocus = false) {
  showDropdown.value = false
  if (restoreFocus) nextTick(() => triggerRef.value?.focus())
}

function toggleDropdown() {
  showDropdown.value ? closeDropdown() : openDropdown()
}

function handleEnabledChange() {
  emit('update:enabled', !props.enabled)
  closeDropdown(true)
}

function handleIntervalChange(seconds: number) {
  emit('update:interval', seconds)
  closeDropdown(true)
}

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    closeDropdown()
  }
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && showDropdown.value) {
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
