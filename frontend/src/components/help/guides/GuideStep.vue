<template><section class="client-guide-step" :data-step="id"><h4><span>{{ number }}</span>{{ title }}</h4><p v-if="description" class="guide-description">{{ description }}</p><slot /></section></template>
<script setup lang="ts">
defineProps<{ number: number; title: string; description?: string; id: string }>()
</script>
<style scoped>
.client-guide-step { margin: 1.75rem 0; min-width: 0; }
h4 { display: flex; align-items: center; gap: .6rem; font-size: .95rem; font-weight: 600; }
h4 span { display: grid; place-items: center; width: 24px; height: 24px; flex-shrink: 0; border-radius: 6px; color: var(--client-on-accent, #fff); background: color-mix(in srgb, var(--client-accent, var(--av-color-brand-primary)) 62%, #111827); font-size: .75rem; box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--scene-fg) 16%, transparent); }
.guide-description { margin: .7rem 0 1rem; font-size: .8rem; line-height: 1.85; }
</style>
