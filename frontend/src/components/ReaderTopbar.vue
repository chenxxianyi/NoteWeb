<script setup lang="ts">
import type { Component } from 'vue'
import { FileText } from 'lucide-vue-next'

withDefaults(defineProps<{
  title: string
  meta: string
  hidden?: boolean
  fixed?: boolean
  icon?: Component | string
  centerLabel?: string
}>(), {
  hidden: false,
  fixed: false,
  centerLabel: '文档工具',
})

const emit = defineEmits<{
  back: []
}>()
</script>

<template>
  <header
    :class="[
      'reader-topbar',
      {
        'reader-topbar--hidden': hidden,
        'reader-topbar--fixed': fixed,
      },
    ]"
  >
    <div class="reader-topbar__title">
      <button type="button" class="rtb-btn reader-topbar__back" title="返回" aria-label="返回" @click="emit('back')">
        <slot name="back-icon"></slot>
      </button>
      <component :is="icon || FileText" class="reader-topbar__icon" aria-hidden="true" />
      <div class="reader-topbar__heading">
        <h1>{{ title || '未命名文档' }}</h1>
        <span>{{ meta }}</span>
      </div>
    </div>

    <div class="reader-topbar__center" role="toolbar" :aria-label="centerLabel">
      <slot name="center"></slot>
    </div>

    <div class="reader-topbar__side">
      <slot name="side"></slot>
    </div>
  </header>
</template>

<style scoped>
.reader-topbar {
  position: sticky;
  top: 0;
  z-index: 30;
  display: grid;
  grid-template-columns: minmax(180px, 1fr) auto minmax(190px, 1fr);
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  padding: 0.7rem 1rem;
  border-bottom: 1px solid var(--border-color);
  background: rgba(250, 248, 245, 0.94);
  backdrop-filter: blur(12px);
  font-family: var(--font-ui);
  transition: opacity 0.3s, transform 0.3s;
}

.reader-topbar--fixed {
  position: fixed;
  left: 0;
  right: 0;
}

.reader-topbar--hidden {
  opacity: 0;
  pointer-events: none;
  transform: translateY(-100%);
}

.reader-topbar__title {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  min-width: 0;
}

.reader-topbar__icon {
  width: 22px;
  height: 22px;
  color: var(--accent);
  flex: 0 0 auto;
}

.reader-topbar__heading {
  min-width: 0;
}

.reader-topbar__heading h1 {
  overflow: hidden;
  margin: 0;
  color: var(--text-primary);
  font-family: var(--font-display);
  font-size: 1.08rem;
  font-weight: 600;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reader-topbar__heading span {
  display: block;
  margin-top: 0.1rem;
  color: var(--text-muted);
  font-size: 0.72rem;
}

.reader-topbar__center,
.reader-topbar__side {
  display: flex;
  align-items: center;
  gap: 0.18rem;
  min-width: 0;
}

.reader-topbar__center {
  justify-content: center;
}

.reader-topbar__side {
  position: relative;
  justify-content: flex-end;
}

.reader-topbar :deep(.rtb-btn),
.reader-topbar :deep(.rtb-status),
.reader-topbar :deep(.rtb-search-button) {
  min-width: 32px;
  height: 32px;
  border: 1px solid transparent;
  border-radius: 6px;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.16s, border-color 0.16s, color 0.16s, opacity 0.16s;
  touch-action: manipulation;
}

.reader-topbar :deep(.rtb-btn) {
  width: 32px;
}

.reader-topbar :deep(.rtb-btn:hover),
.reader-topbar :deep(.rtb-status:hover),
.reader-topbar :deep(.rtb-search-button:hover) {
  border-color: var(--border-color);
  background: var(--accent-light);
  color: var(--accent);
}

.reader-topbar :deep(.rtb-btn.active),
.reader-topbar :deep(.rtb-status.active) {
  border-color: rgba(198, 122, 78, 0.36);
  background: var(--accent-light);
  color: var(--accent);
}

.reader-topbar :deep(.rtb-btn:disabled),
.reader-topbar :deep(.rtb-status:disabled) {
  cursor: not-allowed;
  opacity: 0.38;
}

.reader-topbar :deep(.rtb-btn:disabled:hover),
.reader-topbar :deep(.rtb-status:disabled:hover) {
  background: transparent;
  color: var(--text-secondary);
}

.reader-topbar :deep(.rtb-btn:focus-visible),
.reader-topbar :deep(.rtb-status:focus-visible),
.reader-topbar :deep(.rtb-search-button:focus-visible) {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.reader-topbar :deep(.rtb-btn svg),
.reader-topbar :deep(.rtb-status svg),
.reader-topbar :deep(.rtb-search-button svg),
.reader-topbar__back :deep(svg) {
  width: 17px;
  height: 17px;
}

.reader-topbar :deep(.rtb-divider) {
  width: 1px;
  height: 20px;
  flex: 0 0 auto;
  margin: 0 0.2rem;
  background: var(--border-color);
}

@media (max-width: 1180px) {
  .reader-topbar {
    grid-template-columns: minmax(150px, 0.8fr) auto minmax(160px, 0.8fr);
    gap: 0.45rem;
    padding-inline: 0.75rem;
  }

  .reader-topbar__center,
  .reader-topbar__side {
    gap: 0.12rem;
  }
}

@media (max-width: 1040px) {
  .reader-topbar:not(.reader-topbar--fixed) {
    grid-template-columns: 1fr auto;
  }

  .reader-topbar:not(.reader-topbar--fixed) .reader-topbar__center {
    grid-column: 1 / -1;
    justify-content: flex-start;
    order: 3;
    overflow-x: auto;
    padding-top: 0.35rem;
  }

  .reader-topbar--fixed {
    grid-template-columns: minmax(130px, 0.7fr) minmax(0, auto) auto;
    overflow: visible;
  }
}

@media (max-width: 760px) {
  .reader-topbar--fixed {
    grid-template-columns: auto minmax(0, 1fr) auto;
    gap: 0.35rem;
    padding: 0.55rem 0.6rem;
  }

  .reader-topbar--fixed .reader-topbar__title {
    gap: 0;
  }

  .reader-topbar--fixed .reader-topbar__icon,
  .reader-topbar--fixed .reader-topbar__heading {
    display: none;
  }

  .reader-topbar--fixed .reader-topbar__center {
    justify-content: flex-start;
    overflow: hidden;
  }
}

@media (max-width: 640px) {
  .reader-topbar:not(.reader-topbar--fixed) {
    padding: 0.6rem 0.7rem;
  }

  .reader-topbar__heading h1 {
    font-size: 0.96rem;
  }

  .reader-topbar:not(.reader-topbar--fixed) .reader-topbar__heading span {
    display: none;
  }
}
</style>
