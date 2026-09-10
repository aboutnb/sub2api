<template>
  <div v-if="visible" ref="rootRef" class="relative shrink-0">
    <div
      v-if="!status?.checked_in_today"
      data-testid="checkin-shortcut-actions"
      class="hidden items-center gap-1.5 xl:flex"
    >
      <button
        v-if="status?.normal_enabled"
        type="button"
        data-testid="quick-checkin-normal"
        class="flex h-8 items-center gap-1.5 rounded-xl bg-amber-50 px-3 text-sm font-semibold text-amber-700 transition-colors hover:bg-amber-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-amber-400/50 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-amber-900/20 dark:text-amber-300 dark:hover:bg-amber-900/30"
        :disabled="submitting"
        :aria-label="t('checkin.normal')"
        :title="t('checkin.normal')"
        @click.stop="requestCheckin('normal')"
      >
        <Icon :name="submittingMode === 'normal' ? 'refresh' : 'checkCircle'" size="sm" :class="{ 'animate-spin': submittingMode === 'normal' }" />
        <span>{{ submittingMode === 'normal' ? t('checkin.submitting') : t('checkin.normal') }}</span>
      </button>

      <button
        v-if="status?.lucky_enabled"
        type="button"
        data-testid="quick-checkin-lucky"
        class="flex h-8 items-center gap-1.5 rounded-xl bg-violet-50 px-3 text-sm font-semibold text-violet-700 transition-colors hover:bg-violet-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-violet-400/50 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-violet-900/20 dark:text-violet-300 dark:hover:bg-violet-900/30"
        :disabled="submitting"
        :aria-label="t('checkin.lucky')"
        :title="t('checkin.lucky')"
        @click.stop="requestCheckin('lucky')"
      >
        <Icon :name="submittingMode === 'lucky' ? 'refresh' : 'sparkles'" size="sm" :class="{ 'animate-spin': submittingMode === 'lucky' }" />
        <span>{{ submittingMode === 'lucky' ? t('checkin.submitting') : t('checkin.lucky') }}</span>
      </button>
    </div>

    <button
      type="button"
      data-testid="checkin-shortcut"
      class="h-8 items-center gap-1.5 rounded-xl px-3 text-xs font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 sm:text-sm"
      :class="status?.checked_in_today
        ? 'flex bg-emerald-50 text-emerald-700 hover:bg-emerald-100 focus-visible:ring-emerald-400/50 dark:bg-emerald-900/20 dark:text-emerald-300 dark:hover:bg-emerald-900/30'
        : 'flex bg-amber-50 text-amber-700 hover:bg-amber-100 focus-visible:ring-amber-400/50 dark:bg-amber-900/20 dark:text-amber-300 dark:hover:bg-amber-900/30 xl:hidden'"
      :disabled="submitting"
      :aria-label="shortcutLabel"
      :title="shortcutLabel"
      :aria-expanded="menuOpen"
      :aria-haspopup="status?.checked_in_today ? undefined : 'menu'"
      aria-controls="checkin-shortcut-menu"
      @click.stop="handleShortcutClick"
    >
      <Icon
        :name="submitting ? 'refresh' : status?.checked_in_today ? 'checkCircle' : 'gift'"
        size="sm"
        :class="{ 'animate-spin': submitting }"
      />
      <span class="hidden sm:inline">{{ shortcutLabel }}</span>
      <Icon
        v-if="!submitting && !status?.checked_in_today"
        name="chevronDown"
        size="xs"
        class="hidden opacity-60 sm:block"
      />
    </button>

    <transition name="checkin-menu">
      <div
        v-if="menuOpen"
        id="checkin-shortcut-menu"
        data-testid="checkin-shortcut-menu"
        role="menu"
        class="absolute right-0 top-full z-50 mt-2 w-56 overflow-hidden rounded-lg border border-gray-200 bg-white py-1 shadow-xl dark:border-dark-700 dark:bg-dark-800"
        @click.stop
        @keydown.esc="closeMenu"
      >
        <div v-if="status?.normal_enabled">
          <button
            type="button"
            role="menuitem"
            data-testid="quick-checkin-menu-normal"
            class="flex h-11 w-full items-center gap-3 px-3 text-left text-sm font-medium text-gray-700 transition-colors hover:bg-amber-50 hover:text-amber-700 focus:bg-amber-50 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50 dark:text-dark-200 dark:hover:bg-amber-900/20 dark:hover:text-amber-300"
            :disabled="submitting"
            @click="requestCheckin('normal')"
          >
            <span class="flex h-7 w-7 items-center justify-center rounded-md bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400">
              <Icon name="checkCircle" size="sm" />
            </span>
            <span>{{ t('checkin.normal') }}</span>
          </button>
        </div>

        <button
          v-if="status?.lucky_enabled"
          type="button"
          role="menuitem"
          data-testid="quick-checkin-menu-lucky"
          class="flex h-11 w-full items-center gap-3 px-3 text-left text-sm font-medium text-gray-700 transition-colors hover:bg-violet-50 hover:text-violet-700 focus:bg-violet-50 focus:outline-none dark:text-dark-200 dark:hover:bg-violet-900/20 dark:hover:text-violet-300"
          :disabled="submitting"
          @click="requestCheckin('lucky')"
        >
          <span class="flex h-7 w-7 items-center justify-center rounded-md bg-violet-100 text-violet-600 dark:bg-violet-900/30 dark:text-violet-400">
            <Icon name="sparkles" size="sm" />
          </span>
          <span>{{ t('checkin.lucky') }}</span>
        </button>

        <router-link
          to="/checkin"
          role="menuitem"
          data-testid="open-checkin-page"
          class="mt-1 flex h-10 items-center gap-3 border-t border-gray-100 px-3 pt-1 text-sm text-gray-500 transition-colors hover:bg-gray-50 hover:text-gray-800 focus:bg-gray-50 focus:outline-none dark:border-dark-700 dark:text-dark-400 dark:hover:bg-dark-700 dark:hover:text-white"
          @click="closeMenu"
        >
          <span class="flex h-7 w-7 items-center justify-center">
            <CheckinCenterIcon class="h-4 w-4" />
          </span>
          <span>{{ t('checkin.openPage') }}</span>
          <Icon name="chevronRight" size="xs" class="ml-auto" />
        </router-link>
      </div>
    </transition>
  </div>
  <LuckyCheckinConfirmDialog
    :show="confirmOpen !== null"
    :mode="confirmOpen || 'lucky'"
    :reward-type="status?.lucky_reward_type || 'multiplier'"
    :min-multiplier="status?.lucky_min_multiplier || 0"
    :max-multiplier="status?.lucky_max_multiplier || 0"
    :submitting="Boolean(confirmOpen && submittingMode === confirmOpen)"
    :verification-required="Boolean(status?.turnstile_enabled && status.turnstile_site_key)"
    :verification-complete="Boolean(activeTurnstileToken)"
    @confirm="confirmCheckin"
    @cancel="closeConfirm"
  >
    <template #verification>
      <TurnstileWidget
        v-if="confirmOpen === 'normal' && status?.turnstile_enabled && status.turnstile_site_key"
        ref="normalTurnstileRef"
        :site-key="status.turnstile_site_key"
        size="compact"
        @verify="handleTurnstileVerify('normal', $event)"
        @expire="handleTurnstileExpire('normal')"
        @error="handleTurnstileError('normal')"
      />
      <TurnstileWidget
        v-else-if="confirmOpen === 'lucky' && status?.turnstile_enabled && status.turnstile_site_key"
        ref="luckyTurnstileRef"
        :site-key="status.turnstile_site_key"
        size="flexible"
        @verify="handleTurnstileVerify('lucky', $event)"
        @expire="handleTurnstileExpire('lucky')"
        @error="handleTurnstileError('lucky')"
      />
    </template>
  </LuckyCheckinConfirmDialog>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { checkinAPI } from '@/api/checkin'
import LuckyCheckinConfirmDialog from '@/components/checkin/LuckyCheckinConfirmDialog.vue'
import CheckinCenterIcon from '@/components/icons/CheckinCenterIcon.vue'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import type { CheckinStatus } from '@/types'
import { checkinErrorMessage } from '@/utils/checkinError'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()

const rootRef = ref<HTMLElement | null>(null)
const status = ref<CheckinStatus | null>(null)
const menuOpen = ref(false)
const submitting = ref(false)
const submittingMode = ref<'normal' | 'lucky' | null>(null)
type CheckinMode = 'normal' | 'lucky'
const confirmOpen = ref<CheckinMode | null>(null)
const normalTurnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const luckyTurnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const normalTurnstileToken = ref('')
const luckyTurnstileToken = ref('')
const activeTurnstileToken = computed(() => confirmOpen.value ? turnstileTokenFor(confirmOpen.value).value : '')
let statusRequest = 0

const hasAvailableMode = computed(() => Boolean(status.value?.normal_enabled || status.value?.lucky_enabled))
const visible = computed(() => Boolean(
  authStore.user
  && !authStore.isSimpleMode
  && status.value?.enabled
  && status.value?.eligible
  && hasAvailableMode.value,
))
const shortcutLabel = computed(() => {
  if (submitting.value) return t('checkin.submitting')
  return status.value?.checked_in_today ? t('checkin.checkedToday') : t('checkin.quickAction')
})

function closeMenu() {
  menuOpen.value = false
}

function turnstileTokenFor(mode: CheckinMode) {
  return mode === 'normal' ? normalTurnstileToken : luckyTurnstileToken
}

function turnstileRefFor(mode: CheckinMode) {
  return mode === 'normal' ? normalTurnstileRef : luckyTurnstileRef
}

function handleTurnstileVerify(mode: CheckinMode, token: string) {
  turnstileTokenFor(mode).value = token
}

function handleTurnstileExpire(mode: CheckinMode) {
  turnstileTokenFor(mode).value = ''
}

function handleTurnstileError(mode: CheckinMode) {
  turnstileTokenFor(mode).value = ''
  appStore.showError(t('checkin.turnstileFailed'))
}

function resetTurnstile(mode: CheckinMode) {
  turnstileRefFor(mode).value?.reset()
  turnstileTokenFor(mode).value = ''
}

function handleShortcutClick() {
  if (status.value?.checked_in_today || !status.value?.can_check_in) {
    closeMenu()
    void router.push('/checkin')
    return
  }
  menuOpen.value = !menuOpen.value
}

function handleClickOutside(event: MouseEvent) {
  if (rootRef.value && !rootRef.value.contains(event.target as Node)) closeMenu()
}

async function loadStatus() {
  const request = ++statusRequest
  if (!authStore.user || authStore.isSimpleMode) {
    status.value = null
    closeMenu()
    return
  }
  try {
    const nextStatus = await checkinAPI.getStatus()
    if (request !== statusRequest) return
    status.value = nextStatus
    if (!nextStatus.can_check_in) closeMenu()
    if (!nextStatus.lucky_enabled && confirmOpen.value === 'lucky') closeConfirm()
    if (!nextStatus.normal_enabled && confirmOpen.value === 'normal') closeConfirm()
  } catch {
    if (request === statusRequest) {
      status.value = null
      closeMenu()
    }
  }
}

async function refreshAvailability() {
  try {
    await appStore.fetchPublicSettings(true)
  } catch {
    // The authenticated status endpoint remains authoritative for the shortcut.
  }
  await loadStatus()
}

async function submit(mode: 'normal' | 'lucky') {
  const currentStatus = status.value
  const modeEnabled = mode === 'normal' ? currentStatus?.normal_enabled : currentStatus?.lucky_enabled
  if (!currentStatus?.can_check_in || !modeEnabled || submitting.value) return
  const turnstileToken = turnstileTokenFor(mode).value
  if (currentStatus.turnstile_enabled && !turnstileToken) {
    appStore.showError(t('checkin.turnstileRequired'))
    return
  }

  submitting.value = true
  submittingMode.value = mode
  closeMenu()
  try {
    const response = currentStatus.turnstile_enabled
      ? await checkinAPI.checkIn(mode, currentStatus.business_date, turnstileToken)
      : await checkinAPI.checkIn(mode, currentStatus.business_date)
    status.value = {
      ...currentStatus,
      can_check_in: false,
      checked_in_today: true,
      unavailable_reason: 'already_checked_in',
      today_record: response.record,
    }
    appStore.showSuccess(t('checkin.success'))
    resetTurnstile(mode)
    await Promise.allSettled([authStore.refreshUser()])
    window.dispatchEvent(new CustomEvent('checkin:updated'))
  } catch (error) {
    if (currentStatus.turnstile_enabled) resetTurnstile(mode)
    appStore.showError(checkinErrorMessage(error, t, t('checkin.failedDescription')))
    void loadStatus()
  } finally {
    submitting.value = false
    submittingMode.value = null
  }
}

function requestCheckin(mode: 'normal' | 'lucky') {
  closeMenu()
  const modeEnabled = mode === 'normal' ? status.value?.normal_enabled : status.value?.lucky_enabled
  if (status.value?.can_check_in && modeEnabled && !submitting.value) {
    resetTurnstile(mode)
    confirmOpen.value = mode
  }
}

async function confirmCheckin() {
  const mode = confirmOpen.value
  if (!mode || submitting.value) return
  await submit(mode)
  confirmOpen.value = null
}

function closeConfirm() {
  if (submitting.value) return
  const mode = confirmOpen.value
  confirmOpen.value = null
  if (mode) resetTurnstile(mode)
}

function handleCheckinUpdated() {
  void loadStatus()
}

watch(() => [authStore.user?.id, authStore.isSimpleMode], () => {
  void loadStatus()
}, { immediate: true })

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  window.addEventListener('checkin:updated', handleCheckinUpdated)
  window.addEventListener('checkin:config-updated', refreshAvailability)
  window.addEventListener('focus', refreshAvailability)
})

onBeforeUnmount(() => {
  statusRequest += 1
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('checkin:updated', handleCheckinUpdated)
  window.removeEventListener('checkin:config-updated', refreshAvailability)
  window.removeEventListener('focus', refreshAvailability)
})
</script>

<style scoped>
.checkin-menu-enter-active,
.checkin-menu-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.checkin-menu-enter-from,
.checkin-menu-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
