<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../../stores/authStore'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// Import document store for reader navigation
import { useDocumentStore } from '../../stores/documentStore'
const documentStore = useDocumentStore()

const userName = computed(() => authStore.user?.username || '用户')
const initial = computed(() => userName.value.charAt(0).toUpperCase())

const navItems = [
  { label: '工作台', path: '/dashboard', icon: 'grid' },
  { label: '文件库', path: '/documents', icon: 'book' },
  { label: '阅读器', path: '/documents', icon: 'book-open', matchPrefix: '/reader', openReader: true },
  { label: '笔记', path: '/notes', icon: 'edit' },
]

const bottomNav = [
  { label: '设置', path: '/settings', icon: 'settings' },
]

function isActive(item: typeof navItems[0]): boolean {
  if (item.matchPrefix) return route.path.startsWith(item.matchPrefix)
  return route.path === item.path
}

function navigate(path: string, openReader?: boolean) {
  if (openReader) {
    // Try to open the most recent document in reader
    if (documentStore.documents.length > 0) {
      const lastDoc = documentStore.documents[0]
      router.push(`/reader/${lastDoc.id}`)
    } else {
      // Fetch documents first, then try again
      documentStore.fetchDocuments().then(() => {
        if (documentStore.documents.length > 0) {
          router.push(`/reader/${documentStore.documents[0].id}`)
        } else {
          router.push('/documents')
        }
      })
    }
    return
  }
  router.push(path)
}

function logout() {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar__brand">
      <h1>NoteWeb</h1>
      <span>知识型阅读工作台</span>
    </div>

    <nav class="sidebar__nav">
      <div class="sidebar__label">导航</div>
      <a
        v-for="item in navItems"
        :key="item.label"
        :class="['sidebar__item', { active: isActive(item) }]"
        :href="item.path || '#'"
        @click.prevent="navigate(item.path, item.openReader)"
      >
        <svg v-if="item.icon === 'grid'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
        <svg v-if="item.icon === 'book'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
        <svg v-if="item.icon === 'book-open'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>
        <svg v-if="item.icon === 'edit'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
        <svg v-if="item.icon === 'settings'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
        {{ item.label }}
      </a>
    </nav>

    <div class="sidebar__user" @click="logout" style="cursor:pointer">
      <div class="sidebar__user-av">{{ initial }}</div>
      <div>
        <div>{{ userName }}</div>
        <div style="font-size: 0.7rem; color: rgba(255,255,255,0.35);">{{ authStore.user?.email || '' }}</div>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: var(--sidebar-w);
  background: var(--bg-sidebar);
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0; left: 0; bottom: 0;
  z-index: 10;
}
.sidebar__brand {
  padding: 1.5rem 1.25rem;
  border-bottom: 1px solid rgba(255,255,255,0.06);
}
.sidebar__brand h1 {
  font-size: 1.25rem;
  font-weight: 700;
  color: #fff;
  letter-spacing: -0.02em;
}
.sidebar__brand span {
  font-size: 0.75rem;
  color: rgba(255,255,255,0.35);
  margin-top: 0.2rem;
  display: block;
}
.sidebar__nav {
  flex: 1;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}
.sidebar__label {
  font-size: 0.65rem;
  font-weight: 600;
  color: rgba(255,255,255,0.3);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 1rem 0.75rem 0.4rem;
}
.sidebar__item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 0.75rem;
  border-radius: var(--radius);
  color: rgba(255,255,255,0.6);
  text-decoration: none;
  font-size: 0.85rem;
  font-weight: 500;
  transition: all 0.12s;
}
.sidebar__item:hover { background: rgba(255,255,255,0.06); color: #fff; }
.sidebar__item.active { background: var(--accent); color: #fff; }
.sidebar__item svg { width: 18px; height: 18px; flex-shrink: 0; opacity: 0.7; }
.sidebar__item.active svg { opacity: 1; }
.sidebar__user {
  padding: 0.75rem;
  border-top: 1px solid rgba(255,255,255,0.06);
  display: flex;
  align-items: center;
  gap: 0.6rem;
  color: rgba(255,255,255,0.6);
  font-size: 0.8rem;
}
.sidebar__user:hover { background: rgba(255,255,255,0.06); }
.sidebar__user-av {
  width: 32px; height: 32px;
  border-radius: 50%;
  background: var(--accent);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 0.75rem;
  font-weight: 600;
  flex-shrink: 0;
}
</style>
