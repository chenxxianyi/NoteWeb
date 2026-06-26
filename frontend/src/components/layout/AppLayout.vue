<script setup lang="ts">
import Sidebar from './Sidebar.vue'
import { onMounted, computed } from 'vue'
import { useAuthStore } from '../../stores/authStore'
import { useSettingsStore } from '../../stores/settingsStore'
import { useRouter, useRoute } from 'vue-router'

const authStore = useAuthStore()
const settingsStore = useSettingsStore()
const router = useRouter()
const route = useRoute()

const isReaderView = computed(() => {
  return route.path.startsWith('/reader/')
})

const isReadingModeActive = computed(() => {
  return settingsStore.readingMode && isReaderView.value
})

onMounted(() => {
  if (!authStore.user && authStore.token) {
    authStore.fetchUser().catch(() => {
      router.push('/login')
    })
  }
})
</script>

<template>
  <div class="app-layout">
    <Sidebar />
    <main :class="['main', { 'main--reading-mode': isReadingModeActive, 'main--reader': isReaderView }]">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  background: var(--bg-page);
  background-image: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 1px,
    rgba(0, 0, 0, 0.005) 1px,
    rgba(0, 0, 0, 0.005) 2px
  );
  background-size: 100% 2px;
}

.main {
  margin-left: var(--sidebar-w);
  flex: 1;
  min-width: 0;
  transition: margin-left 0.3s ease;
}

/* Reading mode: main expands when sidebar collapsed */
.main--reading-mode {
  margin-left: 4px;
}

@media (max-width: 640px) {
  .main:not(.main--reader) {
    margin-left: 0;
    padding-bottom: calc(86px + env(safe-area-inset-bottom));
  }

  .main--reader {
    margin-left: 0;
  }
}
</style>
