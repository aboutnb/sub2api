<template>
  <button v-if="!hideTrigger && entries.length" type="button" class="btn btn-ghost gap-1.5 px-2 py-1.5 text-brand leading-tight"
    data-testid="header-contact" aria-haspopup="dialog" :aria-expanded="open" @click="open = true">
    <Icon name="chat" size="sm" aria-hidden="true" />
    <span class="whitespace-nowrap text-xs leading-tight">{{ t('common.contactSupport') }}</span>
  </button>
  <BaseDialog :show="open" :title="t('common.contactSupport')" width="normal" @close="open = false">
    <SupportContactList :contacts="entries" />
  </BaseDialog>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import SupportContactList from '@/components/common/SupportContactList.vue'
import { parseSupportContacts } from '@/utils/supportContacts'
const props = withDefaults(defineProps<{ contactInfo: string; show?: boolean; hideTrigger?: boolean }>(), { show: undefined, hideTrigger: false })
const emit = defineEmits<{ 'update:show': [value: boolean] }>()
const { t } = useI18n()
const localOpen = ref(false)
const open = computed({
  get: () => props.show ?? localOpen.value,
  set: (value: boolean) => {
    localOpen.value = value
    emit('update:show', value)
  },
})
const entries = computed(() => parseSupportContacts(props.contactInfo).filter(entry => entry.name.trim() || entry.account.trim() || entry.url.trim()))
</script>
