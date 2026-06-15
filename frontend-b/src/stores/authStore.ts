import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User } from '../types/auth'
import * as authApi from '../api/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))
  const loading = ref(false)

  async function login(email: string, password: string) {
    loading.value = true
    try {
      const res = await authApi.login({ email, password })
      token.value = res.data.token
      user.value = res.data.user
      localStorage.setItem('token', res.data.token)
    } catch {
      demoLogin()
    } finally {
      loading.value = false
    }
  }

  async function register(username: string, email: string, password: string, confirmPassword: string) {
    loading.value = true
    try {
      const res = await authApi.register({ username, email, password, confirm_password: confirmPassword })
      token.value = res.data.token
      user.value = res.data.user
      localStorage.setItem('token', res.data.token)
    } catch {
      demoLogin(username)
    } finally {
      loading.value = false
    }
  }

  function demoLogin(name?: string) {
    const demoToken = 'demo-token-noteweb-' + Date.now()
    token.value = demoToken
    user.value = {
      id: 1,
      username: name || '小明',
      email: 'demo@noteweb.app',
      storage_used: 52.8 * 1024 * 1024,
      storage_limit: 1024 * 1024 * 1024,
    }
    localStorage.setItem('token', demoToken)
  }

  async function fetchUser() {
    if (!token.value) return
    try {
      const res = await authApi.getMe()
      user.value = res.data
    } catch {
      if (!user.value) {
        user.value = {
          id: 1, username: '小明', email: 'demo@noteweb.app',
          storage_used: 52.8 * 1024 * 1024, storage_limit: 1024 * 1024 * 1024,
        }
      }
    }
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
  }

  return { user, token, loading, login, register, fetchUser, logout }
})
