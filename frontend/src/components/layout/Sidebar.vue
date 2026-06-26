<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/authStore'
import { useSettingsStore } from '../../stores/settingsStore'
import { computed, ref, onMounted, onUnmounted } from 'vue'
import {
  LayoutDashboard,
  FileText,
  PenSquare,
  Settings,
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const settingsStore = useSettingsStore()

const iconMap: Record<string, any> = {
  'layout-dashboard': LayoutDashboard,
  'file-text': FileText,
  'pen-square': PenSquare,
  'settings': Settings,
}

const navItems = [
  { name: '仪表盘', icon: 'layout-dashboard', route: '/dashboard' },
  { name: '文件库', icon: 'file-text', route: '/documents' },
  { name: '笔记', icon: 'pen-square', route: '/notes' },
  { name: '设置', icon: 'settings', route: '/settings' },
]

function isActive(itemRoute: string) {
  if (itemRoute === '/documents') return route.path === '/documents'
  return route.path.startsWith(itemRoute)
}

function navigateTo(path: string) {
  router.push(path)
}

const userInitial = computed(() => {
  return authStore.user?.username?.charAt(0).toUpperCase() || 'M'
})

// Reading mode: auto-collapse sidebar in reader view
const isCollapsed = ref(false)
const isHovered = ref(false)

const isReaderView = computed(() => {
  return route.path.startsWith('/reader/')
})

const shouldCollapse = computed(() => {
  return settingsStore.readingMode && isReaderView.value && !isHovered.value
})

function handleMouseEnter() {
  isHovered.value = true
}

function handleMouseLeave() {
  isHovered.value = false
}

// Check reading mode on route change
onMounted(() => {
  if (settingsStore.readingMode && isReaderView.value) {
    isCollapsed.value = true
  }
})

onUnmounted(() => {
  isCollapsed.value = false
})
</script>

<template>
  <nav
    :class="['sidebar', { 'sidebar--collapsed': shouldCollapse }]"
    aria-label="主功能导航"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
  >
    <a href="#/" class="sidebar__logo" @click.prevent="router.push('/dashboard')">
      N
    </a>
    <div class="sidebar__nav">
      <a
        v-for="item in navItems"
        :key="item.route"
        href="#"
        :class="['sidebar__item', { active: isActive(item.route) }]"
        @click.prevent="navigateTo(item.route)"
      >
        <component :is="iconMap[item.icon]" :size="22" :stroke-width="1.5" />
        <span class="tooltip">{{ item.name }}</span>
      </a>
    </div>
    <div class="sidebar__avatar" @click="router.push('/settings')">
      {{ userInitial }}
    </div>
  </nav>

  <nav v-if="!isReaderView" class="mobile-nav" aria-label="移动端功能导航">
    <a
      v-for="item in navItems"
      :key="`mobile-${item.route}`"
      href="#"
      :class="['mobile-nav__item', { active: isActive(item.route) }]"
      @click.prevent="navigateTo(item.route)"
    >
      <component :is="iconMap[item.icon]" :size="20" :stroke-width="1.7" />
      <span>{{ item.name }}</span>
    </a>
  </nav>
</template>

<style scoped>
.sidebar {
  width: var(--sidebar-w);
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1.2rem 0;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 100;
  transition: width 0.3s ease, transform 0.3s ease;
  overflow: hidden;
}

.sidebar__logo {
  font-family: var(--font-display);
  font-size: 1.3rem;
  color: var(--accent);
  margin-bottom: 2rem;
  text-decoration: none;
  transition: opacity 0.3s ease;
  white-space: nowrap;
  overflow: hidden;
}

.sidebar__nav {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  flex: 1;
}

.sidebar__item {
  width: 44px;
  height: 44px;
  border-radius: var(--radius);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  text-decoration: none;
  transition: all 0.15s;
  position: relative;
}

.sidebar__item:hover,
.sidebar__item.active {
  color: var(--accent);
  background: var(--accent-light);
}

.sidebar__item .tooltip {
  position: absolute;
  left: calc(100% + 10px);
  top: 50%;
  transform: translateY(-50%);
  background: var(--text-primary);
  color: #fff;
  font-family: var(--font-ui);
  font-size: 0.75rem;
  padding: 0.25rem 0.6rem;
  border-radius: 4px;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.15s;
}

.sidebar__item:hover .tooltip {
  opacity: 1;
}

.sidebar__avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--accent-light);
  border: 2px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent);
  font-family: var(--font-ui);
  font-size: 0.8rem;
  cursor: pointer;
  transition: opacity 0.3s ease;
}

.mobile-nav {
  display: none;
}

/* Reading mode: collapsed sidebar */
.sidebar--collapsed {
  width: 4px;
  min-width: 4px;
}

.sidebar--collapsed .sidebar__logo,
.sidebar--collapsed .sidebar__nav,
.sidebar--collapsed .sidebar__avatar {
  opacity: 0;
  pointer-events: none;
}

@media (max-width: 640px) {
  .sidebar {
    display: none;
  }

  .mobile-nav {
    position: fixed;
    left: max(10px, env(safe-area-inset-left));
    right: max(10px, env(safe-area-inset-right));
    bottom: max(10px, env(safe-area-inset-bottom));
    z-index: 120;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 0.25rem;
    min-height: 62px;
    padding: 0.38rem;
    background: rgba(250, 248, 245, 0.96);
    border: 1px solid rgba(228, 217, 206, 0.92);
    border-radius: 8px;
    box-shadow: 0 10px 28px rgba(61, 46, 36, 0.12);
    backdrop-filter: blur(14px);
  }

  .mobile-nav__item {
    min-width: 0;
    min-height: 50px;
    border-radius: 6px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.18rem;
    color: var(--text-muted);
    font-family: var(--font-ui);
    font-size: 0.68rem;
    line-height: 1.1;
    text-decoration: none;
    touch-action: manipulation;
    transition: background 0.15s, color 0.15s;
  }

  .mobile-nav__item svg {
    flex: 0 0 auto;
  }

  .mobile-nav__item span {
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .mobile-nav__item.active {
    color: var(--accent);
    background: var(--accent-light);
  }

  .mobile-nav__item:hover {
    color: var(--accent);
  }
}
</style>
