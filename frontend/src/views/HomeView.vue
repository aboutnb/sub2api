<template>
  <!-- The homepage is a reviewed build-time asset. Legacy database HTML is intentionally ignored. -->
  <div data-testid="fixed-home" v-html="renderedHomeContent"></div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import fixedHomeTemplate from '../../../aivoza-home-pixel.html?raw'

import { useThemeMode } from '@/composables/useThemeMode'
import { useAppStore } from '@/stores/app'
import { resolveBrandLogo } from '@/utils/branding'

const appStore = useAppStore()
const { isDark } = useThemeMode()

const siteLogo = computed(() =>
  resolveBrandLogo(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo, isDark.value)
)

function escapeHtmlAttribute(value: string): string {
  const entities: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  }
  return value.replace(/[&<>"']/g, (character) => entities[character])
}

const renderedHomeContent = computed(() =>
  fixedHomeTemplate.split('{{SITE_LOGO}}').join(escapeHtmlAttribute(siteLogo.value))
)
</script>
