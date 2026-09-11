<template>
  <div class="relative shrink-0" ref="containerRef" @focusout="handleFocusout">
    <!-- Mini Progress Display -->
    <button
      v-if="hasActiveSubscriptions"
      id="subscription-progress-trigger"
      ref="triggerRef"
      type="button"
      @click="toggleTooltip"
      class="flex min-h-11 min-w-11 cursor-pointer items-center justify-center gap-2 rounded-xl bg-purple-50 px-3 py-1.5 leading-tight transition-colors hover:bg-purple-100 dark:bg-purple-900/20 dark:hover:bg-purple-900/30"
      :title="t('subscriptionProgress.viewDetails')"
      :aria-label="t('subscriptionProgress.viewDetails')"
      :aria-expanded="tooltipOpen"
      aria-controls="subscription-progress-popover"
    >
      <Icon name="creditCard" size="sm" class="text-purple-600 dark:text-purple-400" />
      <div class="flex items-center gap-1.5">
        <!-- Combined progress indicator -->
        <div class="flex items-center gap-0.5">
          <div
            v-for="(sub, index) in displaySubscriptions.slice(0, 3)"
            :key="index"
            class="h-2 w-2 rounded-full"
            :class="getProgressDotClass(sub)"
          ></div>
        </div>
        <span class="text-xs font-medium leading-tight text-purple-700 dark:text-purple-300">
          {{ activeSubscriptions.length }}
        </span>
      </div>
    </button>

    <!-- Keep the common destination visible even before the user owns a plan. -->
    <router-link
      v-else
      to="/subscriptions"
      data-testid="subscription-shortcut-empty"
      class="flex min-h-11 min-w-11 items-center justify-center gap-1.5 rounded-xl bg-purple-50 px-3 text-sm font-semibold leading-tight text-purple-700 transition-colors hover:bg-purple-100 dark:bg-purple-900/20 dark:text-purple-300 dark:hover:bg-purple-900/30"
      :title="t('nav.mySubscriptions')"
      :aria-label="t('nav.mySubscriptions')"
    >
      <Icon name="creditCard" size="sm" class="text-purple-600 dark:text-purple-400" />
      <span class="hidden sm:inline">{{ t('nav.mySubscriptions') }}</span>
    </router-link>

    <!-- Hover/Click Tooltip -->
    <transition name="dropdown">
      <div
        v-if="tooltipOpen"
        id="subscription-progress-popover"
        class="subscription-progress-popover absolute right-0 z-50 mt-2 w-[340px] max-w-[calc(100vw-1rem)] overflow-hidden rounded-xl border border-line bg-white shadow-xl dark:border-line dark:bg-surface"
        role="region"
        aria-labelledby="subscription-progress-title"
      >
        <div class="border-b border-line p-3 dark:border-line">
          <h3 id="subscription-progress-title" class="text-sm font-semibold text-ink-strong dark:text-white">
            {{ t('subscriptionProgress.title') }}
          </h3>
          <p class="mt-0.5 text-xs text-ink-muted dark:text-ink-muted">
            {{ t('subscriptionProgress.activeCount', { count: activeSubscriptions.length }) }}
          </p>
        </div>

        <div class="max-h-64 overflow-y-auto">
          <div
            v-for="subscription in displaySubscriptions"
            :key="subscription.id"
            class="border-b border-line p-3 last:border-b-0 dark:border-line/50"
          >
            <div class="mb-2 flex items-center justify-between">
              <span class="text-sm font-medium text-ink-strong dark:text-white">
                {{ subscription.group?.name || `Group #${subscription.group_id}` }}
              </span>
              <span
                v-if="subscriptionExpirationEnabled && subscription.expires_at"
                class="text-xs"
                :class="getDaysRemainingClass(subscription.expires_at)"
              >
                {{ formatDaysRemaining(subscription.expires_at) }}
              </span>
              <span v-else class="text-xs text-ink-muted dark:text-ink-muted">
                {{ t('subscriptionProgress.noExpiration') }}
              </span>
            </div>

            <!-- Progress bars or Unlimited badge -->
            <div class="space-y-1.5">
              <!-- Unlimited subscription badge -->
              <div
                v-if="isUnlimited(subscription)"
                class="flex items-center gap-2 rounded-lg border border-emerald-200 bg-emerald-50 px-2.5 py-1.5 dark:border-emerald-800 dark:bg-emerald-900/20"
              >
                <span class="text-lg text-emerald-600 dark:text-emerald-400">∞</span>
                <span class="text-xs font-medium text-emerald-700 dark:text-emerald-300">
                  {{ t('subscriptionProgress.unlimited') }}
                </span>
              </div>

              <!-- Progress bars for limited subscriptions -->
              <template v-else>
                <div v-if="subscription.group?.daily_limit_usd" class="flex items-center gap-2">
                  <span class="w-8 flex-shrink-0 text-[10px] text-ink-muted">{{
                    t('subscriptionProgress.daily')
                  }}</span>
                  <div class="h-1.5 min-w-0 flex-1 rounded-full bg-line dark:bg-line-strong">
                    <div
                      class="h-1.5 rounded-full transition-all"
                      :class="
                        getProgressBarClass(
                          subscription.daily_usage_usd,
                          subscription.group?.daily_limit_usd
                        )
                      "
                      :style="{
                        width: getProgressWidth(
                          subscription.daily_usage_usd,
                          subscription.group?.daily_limit_usd
                        )
                      }"
                    ></div>
                  </div>
                  <span class="w-24 flex-shrink-0 text-right text-[10px] text-ink-muted">
                    {{
                      formatUsage(subscription.daily_usage_usd, subscription.group?.daily_limit_usd)
                    }}
                  </span>
                </div>

                <div v-if="subscription.group?.weekly_limit_usd" class="flex items-center gap-2">
                  <span class="w-8 flex-shrink-0 text-[10px] text-ink-muted">{{
                    t('subscriptionProgress.weekly')
                  }}</span>
                  <div class="h-1.5 min-w-0 flex-1 rounded-full bg-line dark:bg-line-strong">
                    <div
                      class="h-1.5 rounded-full transition-all"
                      :class="
                        getProgressBarClass(
                          subscription.weekly_usage_usd,
                          subscription.group?.weekly_limit_usd
                        )
                      "
                      :style="{
                        width: getProgressWidth(
                          subscription.weekly_usage_usd,
                          subscription.group?.weekly_limit_usd
                        )
                      }"
                    ></div>
                  </div>
                  <span class="w-24 flex-shrink-0 text-right text-[10px] text-ink-muted">
                    {{
                      formatUsage(subscription.weekly_usage_usd, subscription.group?.weekly_limit_usd)
                    }}
                  </span>
                </div>

                <div v-if="subscription.group?.monthly_limit_usd" class="flex items-center gap-2">
                  <span class="w-8 flex-shrink-0 text-[10px] text-ink-muted">{{
                    t('subscriptionProgress.monthly')
                  }}</span>
                  <div class="h-1.5 min-w-0 flex-1 rounded-full bg-line dark:bg-line-strong">
                    <div
                      class="h-1.5 rounded-full transition-all"
                      :class="
                        getProgressBarClass(
                          subscription.monthly_usage_usd,
                          subscription.group?.monthly_limit_usd
                        )
                      "
                      :style="{
                        width: getProgressWidth(
                          subscription.monthly_usage_usd,
                          subscription.group?.monthly_limit_usd
                        )
                      }"
                    ></div>
                  </div>
                  <span class="w-24 flex-shrink-0 text-right text-[10px] text-ink-muted">
                    {{
                      formatUsage(
                        subscription.monthly_usage_usd,
                        subscription.group?.monthly_limit_usd
                      )
                    }}
                  </span>
                </div>
              </template>
            </div>
          </div>
        </div>

        <div class="border-t border-line p-2 dark:border-line">
          <router-link
            to="/subscriptions"
            @click="closeTooltip"
            class="block w-full py-1 text-center text-xs text-primary-600 hover:underline dark:text-primary-400"
          >
            {{ t('subscriptionProgress.viewAll') }}
          </router-link>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore, useSubscriptionStore } from '@/stores'
import type { UserSubscription } from '@/types'

const { t } = useI18n()

const subscriptionStore = useSubscriptionStore()
const appStore = useAppStore()
const subscriptionExpirationEnabled = computed(
  () => appStore.cachedPublicSettings?.subscription_expiration_enabled !== false
)

const containerRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const tooltipOpen = ref(false)

// Use store data instead of local state
const activeSubscriptions = computed(() => subscriptionStore.activeSubscriptions)
const hasActiveSubscriptions = computed(() => subscriptionStore.hasActiveSubscriptions)

const displaySubscriptions = computed(() => {
  // Sort by most usage (highest percentage first)
  return [...activeSubscriptions.value].sort((a, b) => {
    const aMax = getMaxUsagePercentage(a)
    const bMax = getMaxUsagePercentage(b)
    return bMax - aMax
  })
})

function getMaxUsagePercentage(sub: UserSubscription): number {
  const percentages: number[] = []
  if (sub.group?.daily_limit_usd) {
    percentages.push(((sub.daily_usage_usd || 0) / sub.group.daily_limit_usd) * 100)
  }
  if (sub.group?.weekly_limit_usd) {
    percentages.push(((sub.weekly_usage_usd || 0) / sub.group.weekly_limit_usd) * 100)
  }
  if (sub.group?.monthly_limit_usd) {
    percentages.push(((sub.monthly_usage_usd || 0) / sub.group.monthly_limit_usd) * 100)
  }
  return percentages.length > 0 ? Math.max(...percentages) : 0
}

function isUnlimited(sub: UserSubscription): boolean {
  return (
    !sub.group?.daily_limit_usd &&
    !sub.group?.weekly_limit_usd &&
    !sub.group?.monthly_limit_usd
  )
}

function getProgressDotClass(sub: UserSubscription): string {
  // Unlimited subscriptions get a special color
  if (isUnlimited(sub)) {
    return 'bg-emerald-500'
  }
  const maxPercentage = getMaxUsagePercentage(sub)
  if (maxPercentage >= 90) return 'bg-red-500'
  if (maxPercentage >= 70) return 'bg-orange-500'
  return 'bg-green-500'
}

function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return 'bg-gray-400'
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) return 'bg-red-500'
  if (percentage >= 70) return 'bg-orange-500'
  return 'bg-green-500'
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

function formatUsage(used: number | undefined, limit: number | null | undefined): string {
  const usedValue = (used || 0).toFixed(2)
  const limitValue = limit?.toFixed(2) || '∞'
  return `$${usedValue}/$${limitValue}`
}

function formatDaysRemaining(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  if (diff < 0) return t('subscriptionProgress.expired')
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  if (days === 0) return t('subscriptionProgress.expiresToday')
  if (days === 1) return t('subscriptionProgress.expiresTomorrow')
  return t('subscriptionProgress.daysRemaining', { days })
}

function getDaysRemainingClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  if (days <= 3) return 'text-red-600 dark:text-red-400'
  if (days <= 7) return 'text-orange-600 dark:text-orange-400'
  return 'text-ink-muted dark:text-ink-muted'
}

function toggleTooltip() {
  tooltipOpen.value = !tooltipOpen.value
}

function closeTooltip(restoreFocus: boolean | Event = false) {
  tooltipOpen.value = false
  if (restoreFocus === true) {
    requestAnimationFrame(() => triggerRef.value?.focus())
  }
}

function handleClickOutside(event: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(event.target as Node)) {
    closeTooltip()
  }
}

function handleFocusout(event: FocusEvent) {
  const nextTarget = event.relatedTarget as Node | null
  if (nextTarget && containerRef.value?.contains(nextTarget)) return
  closeTooltip()
}

function handleDocumentKeydown(event: KeyboardEvent) {
  if (event.key !== 'Escape' || !tooltipOpen.value) return
  event.preventDefault()
  closeTooltip(true)
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleDocumentKeydown)
  // Trigger initial fetch if not already loaded
  // The actual data loading is handled by App.vue globally
  subscriptionStore.fetchActiveSubscriptions().catch((error) => {
    console.error('Failed to load subscriptions in SubscriptionProgressMini:', error)
  })
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleDocumentKeydown)
})
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}

@media (max-width: 639.98px) {
  .subscription-progress-popover {
    position: fixed;
    top: 7.75rem;
    right: 0.5rem;
    left: 0.5rem;
    width: auto;
    max-width: none;
    margin-top: 0;
  }
}
</style>
