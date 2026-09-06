<template>
  <div
    class="water-ripple-backdrop"
    data-testid="water-ripple-backdrop"
    aria-hidden="true"
  >
    <span
      v-for="ripple in ripples"
      :key="ripple.id"
      class="water-ripple"
      :style="{ left: `${ripple.x}px`, top: `${ripple.y}px` }"
    >
      <span class="water-ripple__ring water-ripple__ring--inner"></span>
      <span class="water-ripple__ring water-ripple__ring--middle"></span>
      <span class="water-ripple__ring water-ripple__ring--outer"></span>
    </span>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

interface Ripple {
  id: number
  x: number
  y: number
}

const MAX_RIPPLES = 7
const MIN_INTERVAL_MS = 72
const MIN_DISTANCE_PX = 28
const RIPPLE_LIFETIME_MS = 1400

const ripples = ref<Ripple[]>([])
const removalTimers = new Map<number, number>()

let nextRippleId = 0
let lastRippleAt = Number.NEGATIVE_INFINITY
let lastPoint: { x: number; y: number } | null = null
let reducedMotionMedia: MediaQueryList | null = null
let finePointerMedia: MediaQueryList | null = null
let effectEnabled = false

function removeRipple(id: number) {
  const timer = removalTimers.get(id)
  if (timer !== undefined) {
    window.clearTimeout(timer)
    removalTimers.delete(id)
  }
  ripples.value = ripples.value.filter((ripple) => ripple.id !== id)
}

function clearRipples() {
  for (const timer of removalTimers.values()) {
    window.clearTimeout(timer)
  }
  removalTimers.clear()
  ripples.value = []
  lastPoint = null
  lastRippleAt = Number.NEGATIVE_INFINITY
}

function addRipple(x: number, y: number) {
  if (ripples.value.length >= MAX_RIPPLES) {
    const oldest = ripples.value[0]
    if (oldest) removeRipple(oldest.id)
  }

  const ripple = { id: nextRippleId++, x, y }
  ripples.value = [...ripples.value, ripple]
  removalTimers.set(
    ripple.id,
    window.setTimeout(() => removeRipple(ripple.id), RIPPLE_LIFETIME_MS),
  )
}

function isWorkspaceBackgroundEvent(event: PointerEvent) {
  const target = event.target
  if (!(target instanceof Node)) return false

  const workspace = document.querySelector('.app-workspace')
  if (!workspace?.contains(target)) return false

  return !(target instanceof Element && target.closest('.app-header'))
}

function handlePointerMove(event: PointerEvent) {
  if (!effectEnabled || (event.pointerType && event.pointerType !== 'mouse')) return
  if (!isWorkspaceBackgroundEvent(event)) return

  const now = Date.now()
  if (now - lastRippleAt < MIN_INTERVAL_MS) return

  if (lastPoint) {
    const distance = Math.hypot(event.clientX - lastPoint.x, event.clientY - lastPoint.y)
    if (distance < MIN_DISTANCE_PX) return
  }

  lastRippleAt = now
  lastPoint = { x: event.clientX, y: event.clientY }
  addRipple(event.clientX, event.clientY)
}

function syncMotionPreference() {
  effectEnabled = Boolean(finePointerMedia?.matches && !reducedMotionMedia?.matches)
  if (!effectEnabled) clearRipples()
}

function addMediaListener(media: MediaQueryList | null) {
  if (!media) return
  if (typeof media.addEventListener === 'function') {
    media.addEventListener('change', syncMotionPreference)
  } else {
    media.addListener(syncMotionPreference)
  }
}

function removeMediaListener(media: MediaQueryList | null) {
  if (!media) return
  if (typeof media.removeEventListener === 'function') {
    media.removeEventListener('change', syncMotionPreference)
  } else {
    media.removeListener(syncMotionPreference)
  }
}

function handleVisibilityChange() {
  if (document.hidden) clearRipples()
}

onMounted(() => {
  reducedMotionMedia = window.matchMedia('(prefers-reduced-motion: reduce)')
  finePointerMedia = window.matchMedia('(hover: hover) and (pointer: fine)')
  syncMotionPreference()
  addMediaListener(reducedMotionMedia)
  addMediaListener(finePointerMedia)
  window.addEventListener('pointermove', handlePointerMove, { passive: true })
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onBeforeUnmount(() => {
  window.removeEventListener('pointermove', handlePointerMove)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  removeMediaListener(reducedMotionMedia)
  removeMediaListener(finePointerMedia)
  clearRipples()
})
</script>

<style scoped>
.water-ripple-backdrop {
  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
  contain: strict;
}

.water-ripple {
  position: absolute;
  width: 10.5rem;
  height: 10.5rem;
  opacity: 0;
  transform: translate3d(-50%, -50%, 0);
  animation: water-ripple-life 1400ms linear forwards;
  will-change: opacity;
}

.water-ripple__ring {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 100%;
  height: 100%;
  border: 1px solid rgb(var(--av-rgb-brand-primary) / 0.2);
  border-radius: 50%;
  box-shadow: 0 0 12px rgb(var(--av-rgb-brand-primary) / 0.06);
  opacity: 0;
  transform: translate(-50%, -50%) scale(0.08);
  animation: water-ripple-expand 1120ms cubic-bezier(0.16, 0.72, 0.24, 1) forwards;
  will-change: transform, opacity;
}

.water-ripple__ring--inner {
  width: 58%;
  height: 58%;
}

.water-ripple__ring--middle {
  width: 80%;
  height: 80%;
  border-color: rgb(var(--av-rgb-brand-primary) / 0.16);
  animation-delay: 90ms;
}

.water-ripple__ring--outer {
  border-color: rgb(var(--av-rgb-brand-accent) / 0.11);
  animation-delay: 180ms;
}

:global(.dark) .water-ripple__ring {
  border-color: rgb(var(--av-rgb-brand-primary) / 0.24);
  box-shadow: 0 0 14px rgb(var(--av-rgb-brand-primary) / 0.08);
}

:global(.dark) .water-ripple__ring--middle {
  border-color: rgb(var(--av-rgb-brand-primary) / 0.19);
}

:global(.dark) .water-ripple__ring--outer {
  border-color: rgb(var(--av-rgb-brand-accent) / 0.13);
}

@keyframes water-ripple-life {
  0% { opacity: 0; }
  8% { opacity: 1; }
  76% { opacity: 0.82; }
  100% { opacity: 0; }
}

@keyframes water-ripple-expand {
  0% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.08);
  }
  18% { opacity: 0.78; }
  66% { opacity: 0.3; }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(1);
  }
}

@media (hover: none), (pointer: coarse), (prefers-reduced-motion: reduce) {
  .water-ripple-backdrop {
    display: none;
  }
}
</style>
