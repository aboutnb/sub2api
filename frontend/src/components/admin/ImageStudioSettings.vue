<template>
  <div class="flex items-center justify-between gap-4 border-b border-line px-6 py-4">
    <label class="text-sm font-medium text-ink" for="image-studio-enabled">{{ t('nav.imageStudio') }}</label>
    <span v-if="error" role="alert" class="text-sm text-red-600">{{ error }}</span>
    <Toggle id="image-studio-enabled" :model-value="enabled" :disabled="busy" :aria-label="t('nav.imageStudio')" @update:model-value="save" />
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import apiClient from '@/api/client'
const { t } = useI18n()
const enabled = ref(false)
const busy = ref(true)
const error = ref('')
onMounted(async () => {
  try { enabled.value = (await apiClient.get('/admin/image-studio/settings')).data.enabled }
  catch (e) { error.value = (e as Error).message }
  finally { busy.value = false }
})
async function save(value: boolean) {
  busy.value = true
  error.value = ''
  try { enabled.value = (await apiClient.put('/admin/image-studio/settings', { enabled: value })).data.enabled }
  catch (e) { error.value = (e as Error).message }
  finally { busy.value = false }
}
</script>
