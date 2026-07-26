<template>
  <BaseDialog
    :show="show"
    :title="t('checkin.luckyConfirmTitle')"
    width="narrow"
    :close-on-escape="!submitting"
    :show-close-button="!submitting"
    @close="cancel"
  >
    <div class="space-y-4">
      <div class="flex items-start gap-3">
        <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300">
          <Icon name="sparkles" size="md" />
        </span>
        <p class="pt-0.5 text-sm leading-6 text-gray-600 dark:text-dark-300">
          {{ rewardType === 'multiplier' ? t('checkin.luckyMultiplierRisk') : t('checkin.luckyAmountRisk') }}
        </p>
      </div>

      <div
        v-if="rewardType === 'multiplier'"
        data-testid="lucky-multiplier-range"
        class="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 dark:border-amber-800/60 dark:bg-amber-900/20"
      >
        <p class="text-xs font-medium text-amber-700 dark:text-amber-300">{{ t('checkin.luckyPossibleMultiplier') }}</p>
        <p class="mt-1 font-mono text-xl font-bold text-amber-900 dark:text-amber-100">
          {{ multiplierRange }}
        </p>
      </div>

      <p class="text-xs leading-5 text-gray-500 dark:text-dark-400">{{ t('checkin.luckyConfirmOnce') }}</p>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" :disabled="submitting" @click="cancel">
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          data-testid="confirm-lucky-checkin"
          class="btn bg-amber-600 text-white hover:bg-amber-700 focus:ring-amber-500 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="submitting"
          @click="emit('confirm')"
        >
          <Icon :name="submitting ? 'refresh' : 'sparkles'" size="sm" class="mr-2" :class="{ 'animate-spin': submitting }" />
          {{ submitting ? t('checkin.submitting') : t('checkin.luckyConfirmAction') }}
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

const props = defineProps<{
  show: boolean
  rewardType: 'multiplier' | 'amount'
  minMultiplier: number
  maxMultiplier: number
  submitting: boolean
}>()

const emit = defineEmits<{
  (event: 'confirm'): void
  (event: 'cancel'): void
}>()

const { t } = useI18n()

function formatMultiplier(value: number) {
  const amount = Number(value || 0)
  const compact = Math.abs(amount).toFixed(8).replace(/0+$/, '').replace(/\.$/, '') || '0'
  const [integer, decimal = ''] = compact.split('.')
  const formatted = `${integer}.${decimal.padEnd(2, '0')}`
  return `${amount > 0 ? '+' : amount < 0 ? '-' : ''}${formatted}x`
}

const multiplierRange = computed(() => `${formatMultiplier(props.minMultiplier)} ~ ${formatMultiplier(props.maxMultiplier)}`)

function cancel() {
  if (!props.submitting) emit('cancel')
}
</script>
