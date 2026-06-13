import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/login',
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/LoginView.vue'),
    },
    {
      path: '/register',
      redirect: '/login',
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('../views/DashboardView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/documents',
      name: 'Documents',
      component: () => import('../views/DocumentsView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/reader/:documentId',
      name: 'Reader',
      component: () => import('../views/ReaderView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/notes',
      name: 'Notes',
      component: () => import('../views/NotesView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/settings',
      name: 'Settings',
      component: () => import('../views/SettingsView.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

// Route guard for auth
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
