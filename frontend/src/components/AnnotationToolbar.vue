<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits<{
  highlight: []
  note: []
}>()

const visible = ref(false)
const pos = ref({ x: 0, y: 0 })

function show(x: number, y: number) {
  pos.value = { x, y }
  visible.value = true
}

function hide() {
  visible.value = false
}

function onHighlight() {
  emit('highlight')
  hide()
}

function onNote() {
  emit('note')
  hide()
}

// Expose show/hide for parent to call
defineExpose({ show, hide })
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="anno-toolbar"
      :style="{ left: pos.x + 'px', top: pos.y + 'px' }"
      @click.stop
    >
      <button class="anno-toolbar__btn" title="高亮" @click="onHighlight">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="15" height="15"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>
        <span>高亮</span>
      </button>
      <button class="anno-toolbar__btn" title="批注" @click="onNote">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="15" height="15"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
        <span>批注</span>
      </button>
    </div>
  </Teleport>
</template>

<style scoped>
.anno-toolbar {
  position: fixed;
  z-index: 1000;
  display: flex;
  gap: 2px;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #e5e2dc);
  border-radius: 10px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  padding: 3px;
  transform: translate(-50%, calc(-100% - 8px));
  animation: fadeIn 0.1s ease;
}

.anno-toolbar__btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px 9px;
  border: none;
  border-radius: 7px;
  background: transparent;
  font-family: var(--font-ui, sans-serif);
  font-size: 0.75rem;
  color: var(--text-secondary, #666);
  cursor: pointer;
  transition: all 0.1s;
  white-space: nowrap;
}

.anno-toolbar__btn:hover {
  background: var(--accent-light, #f0eee8);
  color: var(--accent, #2563eb);
}

@keyframes fadeIn {
  from { opacity: 0; transform: translate(-50%, calc(-100% - 4px)); }
  to   { opacity: 1; transform: translate(-50%, calc(-100% - 8px)); }
}
</style>
