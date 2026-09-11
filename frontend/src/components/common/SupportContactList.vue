<template>
  <div class="space-y-4">
    <p v-if="!contacts.length" class="py-6 text-center text-sm text-ink-muted">{{ t('common.support.empty') }}</p>
    <ul v-else class="space-y-3">
      <li v-for="(contact, index) in contacts" :key="index" class="rounded-xl border border-line bg-surface p-4 sm:p-5">
        <div class="flex items-start gap-3">
          <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-line bg-surface-muted text-brand"><Icon :name="contact.icon" size="md" aria-hidden="true" /></span>
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2"><h4 class="break-words font-semibold text-ink [overflow-wrap:anywhere]">{{ contact.name || t('common.contactSupport') }}</h4><span v-if="contact.tag" class="max-w-full rounded-md bg-brand-soft px-2 py-1 text-xs font-medium text-brand [overflow-wrap:anywhere]">{{ contact.tag }}</span></div>
            <p v-if="contact.account || contact.url" class="mt-2 whitespace-pre-wrap break-words text-sm leading-6 text-ink-muted [overflow-wrap:anywhere]">{{ contact.account || contact.url }}</p>
            <div v-if="contact.account || safeContactUrl(contact.url)" class="mt-4 flex flex-wrap gap-2">
              <button v-if="contact.account" type="button" class="btn btn-secondary gap-2" @click="copy(contact.account, index)"><Icon :name="copied && copiedIndex === index ? 'check' : 'copy'" size="sm" /><span aria-live="polite">{{ copied && copiedIndex === index ? t('common.copied') : t('common.support.copyAccount') }}</span></button>
              <a v-if="safeContactUrl(contact.url)" :href="safeContactUrl(contact.url)" target="_blank" rel="noopener noreferrer" class="btn btn-primary gap-2">{{ t('common.support.openLink') }}<Icon name="externalLink" size="sm" /></a>
            </div>
          </div>
        </div>
      </li>
    </ul>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { safeContactUrl, type SupportContactEntry } from '@/utils/supportContacts'
defineProps<{ contacts: SupportContactEntry[] }>()
const { t } = useI18n()
const { copied, copyToClipboard } = useClipboard()
const copiedIndex = ref(-1)
async function copy(account: string, index: number) {
  if (await copyToClipboard(account)) copiedIndex.value = index
}
</script>
