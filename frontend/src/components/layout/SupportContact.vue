<template>
  <button
    type="button"
    class="btn btn-ghost gap-1.5 px-2 py-1.5 text-brand leading-tight"
    data-testid="header-contact"
    aria-haspopup="dialog"
    :aria-expanded="open"
    @click="open = true"
  >
    <Icon name="chat" size="sm" aria-hidden="true" />
    <span class="whitespace-nowrap text-xs leading-tight">{{ t('common.contactSupport') }}</span>
  </button>
  <BaseDialog :show="open" :title="t('common.contactSupport')" width="narrow" @close="open = false">
    <div class="space-y-4">
      <p class="whitespace-pre-wrap break-words rounded-xl border border-line bg-surface-muted p-4 text-sm leading-6 text-ink-strong [overflow-wrap:anywhere]">{{ contactInfo }}</p>
      <button type="button" class="btn btn-secondary w-full gap-2" @click="copyToClipboard(contactInfo)">
        <Icon :name="copied ? 'check' : 'copy'" size="sm" aria-hidden="true" />
        {{ copied ? t('common.copied') : t('common.copy') }}
      </button>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'

defineProps<{ contactInfo: string }>()
const { t } = useI18n()
const { copied, copyToClipboard } = useClipboard()
const open = ref(false)
</script>
