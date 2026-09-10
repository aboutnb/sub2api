<template>
  <ChannelStatusV1View v-if="isV1" />
  <ChannelStatusV2View v-else-if="isLegacyV2" />
  <ChannelStatusV3View v-else />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { isChannelMonitorV1Mode, isChannelMonitorV2Mode } from '@/utils/featureFlags'
import ChannelStatusV1View from './ChannelStatusV1View.vue'
import ChannelStatusV2View from './ChannelStatusV2View.vue'
import ChannelStatusV3View from './ChannelStatusV3View.vue'

const isV1 = computed(() => isChannelMonitorV1Mode())
// Keep the full V2 Ops surface available as a diagnostic/rollback view.
const isLegacyV2 = computed(() => isChannelMonitorV2Mode() && new URLSearchParams(window.location.search).get('monitor_view') === 'v2')
</script>
