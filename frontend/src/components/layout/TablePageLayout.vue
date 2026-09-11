<template>
  <div class="table-page-layout" :class="{ 'mobile-mode': isMobile }">
    <!-- 固定区域：操作按钮 -->
    <div v-if="$slots.actions" class="layout-section-fixed">
      <slot name="actions" />
    </div>

    <!-- 固定区域：搜索和过滤器 -->
    <div v-if="$slots.filters" class="layout-section-fixed">
      <slot name="filters" />
    </div>

    <!-- 滚动区域：表格 -->
    <div class="layout-section-scrollable">
      <div class="card table-scroll-container">
        <slot name="table" />
      </div>
    </div>

    <!-- 固定区域：分页器 -->
    <div v-if="$slots.pagination" class="layout-section-fixed">
      <slot name="pagination" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const mobileViewportQuery = '(max-width: 1023px)'
let mobileViewportMedia: MediaQueryList | null = null
const isMobile = ref(
  typeof window === 'undefined' ? false : window.matchMedia(mobileViewportQuery).matches
)

const handleViewportChange = (event: MediaQueryListEvent) => {
  isMobile.value = event.matches
}

onMounted(() => {
  mobileViewportMedia = window.matchMedia(mobileViewportQuery)
  isMobile.value = mobileViewportMedia.matches
  if (typeof mobileViewportMedia.addEventListener === 'function') {
    mobileViewportMedia.addEventListener('change', handleViewportChange)
  } else {
    mobileViewportMedia.addListener(handleViewportChange)
  }
})

onUnmounted(() => {
  if (!mobileViewportMedia) return
  if (typeof mobileViewportMedia.removeEventListener === 'function') {
    mobileViewportMedia.removeEventListener('change', handleViewportChange)
  } else {
    mobileViewportMedia.removeListener(handleViewportChange)
  }
  mobileViewportMedia = null
})
</script>

<style scoped>
/* 桌面端：Flexbox 布局 */
.table-page-layout {
  @apply flex min-w-0 flex-col gap-5;
  height: calc(100vh - 64px - 4rem); /* 减去 header + lg:p-8 的上下padding */
  height: calc(100dvh - 64px - 4rem);
}

.layout-section-fixed {
  @apply min-w-0 flex-shrink-0;
}

.layout-section-scrollable {
  @apply flex-1 min-h-64 min-w-0 flex flex-col;
}

/* 表格滚动容器 - 增强版表体滚动方案 */
.table-scroll-container {
  @apply flex h-full flex-col overflow-hidden rounded-2xl border-2 border-line bg-white shadow-card dark:border-line dark:bg-surface;
}

.table-scroll-container :deep(.table-wrapper) {
  @apply flex-1 overflow-x-auto overflow-y-auto;
  /* 确保横向滚动条显示在最底部 */
  scrollbar-gutter: stable;
}

.table-scroll-container :deep(table) {
  @apply w-full;
  min-width: max-content; /* 关键：确保表格宽度根据内容撑开，从而触发横向滚动 */
  display: table; /* 使用标准 table 布局以支持 sticky 列 */
}

.table-scroll-container :deep(thead) {
  @apply bg-primary-50/80 backdrop-blur-sm dark:bg-surface/90;
}

.table-scroll-container :deep(tbody) {
  /* 保持默认 table-row-group 显示，不使用 block */
}

.table-scroll-container :deep(th) {
  @apply border-b-2 border-line px-5 py-3.5 text-left text-xs font-bold uppercase tracking-wider text-ink dark:border-line dark:text-ink;
}

.table-scroll-container :deep(td) {
  @apply px-5 py-4 text-sm text-ink dark:text-ink-muted border-b border-line dark:border-line;
}

/* 移动端：恢复正常滚动 */
.table-page-layout.mobile-mode .table-scroll-container {
  @apply h-auto overflow-visible border-none bg-transparent shadow-none;
}

.table-page-layout.mobile-mode {
  height: auto;
}

.table-page-layout.mobile-mode .layout-section-scrollable {
  @apply flex-none min-h-fit;
}

.table-page-layout.mobile-mode .table-scroll-container :deep(table) {
  @apply flex-none;
  display: table;
  min-width: 100%;
}
</style>
