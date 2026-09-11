<template>
  <section class="space-y-4 rounded-2xl border border-line bg-surface-muted p-4 sm:p-5">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div><h3 class="font-semibold text-ink">{{ t('common.contactSupport') }}</h3><p class="mt-1 max-w-xl text-sm leading-6 text-ink-muted">{{ t('common.support.editorHint') }}</p></div>
      <button type="button" class="btn btn-secondary gap-2" @click="preview = true"><Icon name="eye" size="sm" />{{ t('common.support.preview') }}</button>
    </div>
    <p v-if="!entries.length" class="rounded-xl border border-dashed border-line p-6 text-center text-sm text-ink-muted">{{ t('common.support.empty') }}</p>
    <article v-for="(entry, index) in entries" :key="index" class="space-y-4 rounded-xl border border-line bg-surface p-4">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div class="flex min-w-0 items-center gap-2 font-medium text-ink"><Icon :name="entry.icon" size="sm" /><span class="break-all">{{ String(index + 1).padStart(2, '0') }} · {{ entry.name || t('common.support.newContact') }}</span></div>
        <div class="flex gap-1">
          <button type="button" class="btn btn-ghost btn-icon" :disabled="index === 0" :aria-label="t('common.support.moveUp')" @click="move(index, -1)"><Icon name="arrowUp" size="sm" /></button>
          <button type="button" class="btn btn-ghost btn-icon" :disabled="index === entries.length - 1" :aria-label="t('common.support.moveDown')" @click="move(index, 1)"><Icon name="arrowDown" size="sm" /></button>
          <button type="button" class="btn btn-ghost btn-icon" :aria-label="t('common.delete')" @click="remove(index)"><Icon name="trash" size="sm" /></button>
        </div>
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <label class="block text-sm font-medium text-ink">{{ t('common.support.name') }}<input :value="entry.name" class="input mt-2" maxlength="80" required :placeholder="t('common.support.namePlaceholder')" @input="update(index, 'name', ($event.target as HTMLInputElement).value)" /></label>
        <label class="block text-sm font-medium text-ink">{{ t('common.support.tag') }}<input :value="entry.tag" class="input mt-2" maxlength="40" :placeholder="t('common.support.tagPlaceholder')" @input="update(index, 'tag', ($event.target as HTMLInputElement).value)" /></label>
        <label class="block text-sm font-medium text-ink">{{ t('common.support.icon') }}<select :value="entry.icon" class="input mt-2" @change="update(index, 'icon', ($event.target as HTMLSelectElement).value)"><option v-for="icon in supportIcons" :key="icon" :value="icon">{{ t(`common.support.icons.${icon}`) }}</option></select></label>
        <label class="block text-sm font-medium text-ink">{{ t('common.support.account') }}<input :value="entry.account" class="input mt-2" maxlength="2000" :placeholder="t('common.support.accountPlaceholder')" @input="update(index, 'account', ($event.target as HTMLInputElement).value)" /></label>
        <label class="block text-sm font-medium text-ink sm:col-span-2">{{ t('common.support.url') }}<input :value="entry.url" class="input mt-2" maxlength="2048" placeholder="https:// / mailto: / tel:" :aria-invalid="!!entry.url && !safeContactUrl(entry.url)" @input="updateUrl(index, $event)" /><span v-if="entry.url && !safeContactUrl(entry.url)" class="mt-1 block text-xs text-red-600" role="alert">{{ t('common.support.invalidUrl') }}</span><span v-else class="mt-1 block text-xs font-normal text-ink-muted">{{ t('common.support.urlHint') }}</span></label>
      </div>
    </article>
    <button type="button" class="btn btn-secondary w-full gap-2" @click="add"><Icon name="plus" size="sm" />{{ t('common.support.add') }}</button>
    <BaseDialog :show="preview" :title="t('common.contactSupport')" width="normal" @close="preview = false"><SupportContactList :contacts="entries" /></BaseDialog>
  </section>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import SupportContactList from '@/components/common/SupportContactList.vue'
import Icon from '@/components/icons/Icon.vue'
import { parseSupportContacts, serializeSupportContacts, supportIcons, safeContactUrl, type SupportContactEntry } from '@/utils/supportContacts'
const props = defineProps<{ modelValue: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()
const { t } = useI18n()
const preview = ref(false)
const entries = computed(() => parseSupportContacts(props.modelValue))
function publish(value: SupportContactEntry[]) { emit('update:modelValue', serializeSupportContacts(value)) }
function update(index: number, key: keyof SupportContactEntry, value: string) {
  publish(entries.value.map((entry, i) => i === index ? { ...entry, [key]: value } as SupportContactEntry : entry))
}
function updateUrl(index: number, event: Event) {
  const input = event.target as HTMLInputElement
  input.setCustomValidity(input.value && !safeContactUrl(input.value) ? t('common.support.invalidUrl') : '')
  update(index, 'url', input.value)
}
function add() { publish([...entries.value, { name: '', tag: '', icon: 'chat', account: '', url: '' }]) }
function remove(index: number) { publish(entries.value.filter((_, i) => i !== index)) }
function move(index: number, direction: number) {
  const list = [...entries.value]
  const target = index + direction
  if (target < 0 || target >= list.length) return
  ;[list[index], list[target]] = [list[target], list[index]]
  publish(list)
}
</script>
