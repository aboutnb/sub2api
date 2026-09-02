<template>
  <BaseDialog
    :show="show"
    :title="mode === 'normal' ? t('checkin.normalConfirmTitle') : t('checkin.luckyConfirmTitle')"
    width="narrow"
    :close-on-escape="!submitting"
    :show-close-button="!submitting"
    @close="cancel"
  >
    <div class="space-y-4">
      <div class="flex items-start gap-3">
        <span
          class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg"
          :class="mode === 'normal'
            ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
            : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'"
        >
          <Icon :name="mode === 'normal' ? 'gift' : 'sparkles'" size="md" />
        </span>
        <p class="pt-0.5 text-sm leading-6 text-gray-600 dark:text-dark-300">
          {{ mode === 'normal'
            ? t('checkin.normalHint')
            : rewardType === 'multiplier' ? t('checkin.luckyMultiplierRisk') : t('checkin.luckyAmountRisk') }}
        </p>
      </div>

      <div
        v-if="mode === 'lucky' && rewardType === 'multiplier'"
        data-testid="lucky-multiplier-range"
        class="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 dark:border-amber-800/60 dark:bg-amber-900/20"
      >
        <p class="text-xs font-medium text-amber-700 dark:text-amber-300">{{ t('checkin.luckyPossibleMultiplier') }}</p>
        <p class="mt-1 font-mono text-xl font-bold text-amber-900 dark:text-amber-100">
          {{ multiplierRange }}
        </p>
      </div>

      <p class="text-xs leading-5 text-gray-500 dark:text-dark-400">
        {{ mode === 'normal' ? t('checkin.normalConfirmOnce') : t('checkin.luckyConfirmOnce') }}
      </p>

      <div
        v-if="verificationRequired"
        :data-testid="mode === 'normal' ? 'normal-checkin-verification' : 'lucky-checkin-verification'"
        class="border-t border-gray-100 pt-4 dark:border-dark-700"
      >
        <div class="mb-3 flex items-center justify-between gap-3">
          <p class="text-xs font-semibold text-gray-700 dark:text-dark-200">
            {{ mode === 'normal' ? t('checkin.normalVerification') : t('checkin.luckyVerification') }}
          </p>
          <span v-if="verificationComplete" class="inline-flex items-center gap-1 text-xs font-medium text-emerald-600 dark:text-emerald-400">
            <Icon name="checkCircle" size="xs" />
            {{ t('checkin.verificationComplete') }}
          </span>
        </div>
        <div class="mx-auto w-full max-w-[300px]">
          <slot name="verification" />
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" :disabled="submitting" @click="cancel">
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          :data-testid="mode === 'normal' ? 'confirm-normal-checkin' : 'confirm-lucky-checkin'"
          class="btn text-white disabled:cursor-not-allowed disabled:opacity-60"
          :class="mode === 'normal'
            ? 'bg-emerald-600 hover:bg-emerald-700 focus:ring-emerald-500'
            : 'bg-amber-600 hover:bg-amber-700 focus:ring-amber-500'"
          :disabled="submitting || (verificationRequired && !verificationComplete)"
          @click="emit('confirm')"
        >
          <Icon :name="submitting ? 'refresh' : mode === 'normal' ? 'gift' : 'sparkles'" size="sm" class="mr-2" :class="{ 'animate-spin': submitting }" />
          {{ submitting ? t('checkin.submitting') : mode === 'normal' ? t('checkin.normalConfirmAction') : t('checkin.luckyConfirmAction') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(defineProps<{
  show: boolean
  mode?: 'normal' | 'lucky'
  rewardType: 'multiplier' | 'amount'
  minMultiplier: number
  maxMultiplier: number
  submitting: boolean
  verificationRequired?: boolean
  verificationComplete?: boolean
}>(), {
  mode: 'lucky',
  verificationRequired: false,
  verificationComplete: false,
})

const emit = defineEmits<{
  (event: 'confirm'): void
  (event: 'cancel'): void
}>()

const { t } = useI18n()

function formatMultiplier(value: number) {
  const amount = Number(value || 0)
  const compact = Math.abs(amount).toFixed(2).replace(/0+$/, '').replace(/\.$/, '') || '0'
  const [integer, decimal = ''] = compact.split('.')
  const formatted = `${integer}.${decimal.padEnd(2, '0')}`
  return `${amount > 0 ? '+' : amount < 0 ? '-' : ''}${formatted}x`
}

const multiplierRange = computed(() => `${formatMultiplier(props.minMultiplier)} ~ ${formatMultiplier(props.maxMultiplier)}`)

function cancel() {
  if (!props.submitting) emit('cancel')
}
</script>
