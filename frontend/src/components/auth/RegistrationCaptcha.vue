<template>
  <div>
    <CaptchaChallenge
      v-if="settings"
      ref="challenge"
      :turnstile-enabled="settings.turnstile_enabled"
      :turnstile-site-key="settings.turnstile_site_key || ''"
      :tencent-enabled="settings.tencent_captcha_enabled === true"
      :tencent-app-id="settings.tencent_captcha_app_id || ''"
      :tencent-region="settings.tencent_captcha_region"
      :aliyun-enabled="settings.aliyun_captcha_enabled === true"
      :aliyun-scene-id="settings.aliyun_captcha_scene_id || ''"
      :aliyun-prefix="settings.aliyun_captcha_prefix || ''"
      :aliyun-region="settings.aliyun_captcha_region"
      @verify="(value: string) => token = value"
      @expire="token = ''"
      @error="token = ''"
    />
    <button v-if="loadFailed" type="button" class="btn btn-secondary" @click="loadSettings">
      {{ t('common.tryAgain') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getPublicSettings } from '@/api/auth'
import type { ActionCaptchaRequestProof, PublicSettings } from '@/types'
import CaptchaChallenge from '@/components/CaptchaChallenge.vue'

const { t } = useI18n()
const settings = ref<PublicSettings | null>(null)
const loadFailed = ref(false)
const token = ref('')
const challenge = ref<InstanceType<typeof CaptchaChallenge> | null>(null)
async function loadSettings() {
  loadFailed.value = false
  try { settings.value = await getPublicSettings() }
  catch { loadFailed.value = true }
}
async function getProof(): Promise<ActionCaptchaRequestProof> {
  const config = settings.value
  if (!config) throw new Error(t('auth.turnstileFailed'))
  if (config.turnstile_enabled) {
    if (!token.value) throw new Error(t('auth.completeVerification'))
    return { turnstile_token: token.value }
  }
  if (config.tencent_captcha_enabled || config.aliyun_captcha_enabled) {
    const proof = await challenge.value?.verifyAction()
    if (!proof) throw new Error(t('auth.completeVerification'))
    return config.tencent_captcha_enabled
      ? { tencent_captcha_ticket: proof.token, tencent_captcha_randstr: proof.randstr }
      : { turnstile_token: proof.token }
  }
  return {}
}
function reset() { token.value = ''; challenge.value?.reset() }
onMounted(loadSettings)
defineExpose({ getProof, reset })
</script>
