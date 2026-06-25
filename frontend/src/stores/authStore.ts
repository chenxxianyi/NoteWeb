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
    } catch (e: any) {
      throw e
    } finally {
      loading.value = false
    }
  }

  async function register(username: string, email: string, password: string, confirmPassword: string) {
    loading.value = true
    try {
      const res = await authApi.register({
        username,
        email,
        password,
        confirm_password: confirmPassword,
      })
      return res.data
    } catch (e: any) {
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchUser() {
    if (!token.value) return
    const res = await authApi.getMe()
    user.value = res.data
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
  }

  async function changePassword(oldPassword: string, newPassword: string) {
    loading.value = true
    try {
      await authApi.changePassword({
        old_password: oldPassword,
        new_password: newPassword,
      })
    } catch (e: any) {
      throw e
    } finally {
      loading.value = false
    }
  }

  async function updateProfile(username?: string, email?: string) {
    loading.value = true
    try {
      const res = await authApi.updateProfile({ username, email })
      if (res.data) {
        user.value = res.data
      }
    } catch (e: any) {
      throw e
    } finally {
      loading.value = false
    }
  }

  async function uploadAvatar(file: File) {
    loading.value = true
    try {
      const res = await authApi.uploadAvatar(file)
      if (res.data.user) {
        user.value = res.data.user
      }
      return res.data.url
    } catch (e: any) {
      throw e
    } finally {
      loading.value = false
    }
  }

  async function deleteAccount(password: string) {
    loading.value = true
    try {
      await authApi.deleteAccount({ password })
      logout()
    } catch (e: any) {
      throw e
    } finally {
      loading.value = false
    }
  }

  return {
    user,
    token,
    loading,
    login,
    register,
    fetchUser,
    logout,
    changePassword,
    updateProfile,
    uploadAvatar,
    deleteAccount,
  }
})
