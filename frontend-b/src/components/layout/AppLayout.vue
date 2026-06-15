<script setup lang="ts">
import Sidebar from './Sidebar.vue'
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
}
.main {
  margin-left: var(--sidebar-w);
  flex: 1;
  padding: 1.5rem 2rem;
  max-width: 1300px;
}
</style>
