<script setup lang="ts">
import Sidebar from './Sidebar.vue'
import { onMounted } from 'vue'
import { useAuthStore } from '../../stores/authStore'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

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
    <main class="main">
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
}

@media (max-width: 520px) {
  .main {
    margin-left: 0;
  }
}
</style>
