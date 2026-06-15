<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/authStore'

const router = useRouter()
const authStore = useAuthStore()
const isLogin = ref(true)
const email = ref('')
const password = ref('')
const username = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const errorMsg = ref('')

const errorMessages: Record<string, string> = {
  'Invalid email or password': '邮箱或密码错误',
  'Email already registered': '该邮箱已被注册',
  'Username already taken': '该用户名已被使用',
  'Passwords do not match': '两次密码不一致',
}

function translateError(err: any): string {
  const detail = err.response?.data?.detail
  if (detail && typeof detail === 'string') return errorMessages[detail] || detail
  if (err.code === 'ERR_NETWORK') return '无法连接到服务器'
  const status = err.response?.status
  if (status === 401) return '邮箱或密码错误'
  if (status >= 500) return '服务器内部错误'
  return '操作失败，请重试'
}

function toggleMode() {
  isLogin.value = !isLogin.value
  errorMsg.value = ''
}

async function handleSubmit() {
  errorMsg.value = ''
  loading.value = true
  try {
    if (isLogin.value) {
      await authStore.login(email.value, password.value)
      router.push('/dashboard')
    } else {
      if (password.value !== confirmPassword.value) {
        errorMsg.value = '两次密码不一致'
        loading.value = false
        return
      }
      await authStore.register(username.value, email.value, password.value, confirmPassword.value)
      alert('注册成功！请使用你的邮箱和密码登录。')
      password.value = ''
      username.value = ''
      confirmPassword.value = ''
      isLogin.value = true
    }
  } catch (e: any) {
    errorMsg.value = translateError(e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="brand-panel">
      <div class="brand-pattern"></div>
      <div class="brand-grid"></div>
      <div class="brand-content">
        <div class="brand-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/>
          </svg>
        </div>
        <div class="brand-title">NoteWeb</div>
        <div class="brand-desc">知识型阅读工作台，支持多格式文档阅读、划线批注、笔记编辑和 AI 阅读辅助。</div>
        <div class="brand-features">
          <div class="brand-feature"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>多格式文档阅读</div>
          <div class="brand-feature"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>划线批注与摘录</div>
          <div class="brand-feature"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>AI 阅读助手</div>
        </div>
      </div>
    </div>
    <div class="form-panel">
      <div class="form-container">
        <div class="form-header">
          <h2>{{ isLogin ? '欢迎回来' : '创建账号' }}</h2>
          <p>{{ isLogin ? '登录以继续使用 NoteWeb' : '注册一个 NoteWeb 账号' }}</p>
        </div>
        <form @submit.prevent="handleSubmit">
          <div v-if="!isLogin" class="form-group">
            <label for="username">用户名</label>
            <input id="username" v-model="username" type="text" class="form-input" placeholder="给自己取个名字" autocomplete="username" />
          </div>
          <div class="form-group">
            <label for="email">{{ isLogin ? '邮箱 / 用户名' : '邮箱' }}</label>
            <input id="email" v-model="email" type="text" class="form-input" placeholder="your@email.com" autocomplete="email" />
          </div>
          <div class="form-group">
            <label for="password">{{ isLogin ? '密码' : '设置密码' }}</label>
            <input id="password" v-model="password" type="password" class="form-input" placeholder="········" autocomplete="current-password" />
          </div>
          <div v-if="!isLogin" class="form-group">
            <label for="confirm-password">确认密码</label>
            <input id="confirm-password" v-model="confirmPassword" type="password" class="form-input" placeholder="再输一次" autocomplete="new-password" />
          </div>
          <div v-if="errorMsg" class="form-error">{{ errorMsg }}</div>
          <button type="submit" class="btn-submit" :disabled="loading">{{ loading ? '处理中...' : isLogin ? '登 录' : '注 册' }}</button>
        </form>
        <div class="form-footer">
          <span>{{ isLogin ? '没有账号？' : '已有账号？' }}</span>
          <a href="#" @click.prevent="toggleMode">{{ isLogin ? '去注册 →' : '去登录 →' }}</a>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page { display: flex; min-height: 100vh; }
.brand-panel {
  flex: 0 0 480px; background: var(--bg-brand); position: relative;
  overflow: hidden; display: flex; flex-direction: column;
  justify-content: center; align-items: center; padding: 3rem;
}
.brand-pattern {
  position: absolute; inset: 0; opacity: 0.08;
  background: radial-gradient(circle at 20% 30%, #3B82F6 0%, transparent 50%),
              radial-gradient(circle at 80% 70%, #6366F1 0%, transparent 50%),
              radial-gradient(circle at 50% 50%, #0EA5E9 0%, transparent 60%);
}
.brand-grid {
  position: absolute; inset: 0;
  background-image: linear-gradient(rgba(255,255,255,0.04) 1px, transparent 1px),
                    linear-gradient(90deg, rgba(255,255,255,0.04) 1px, transparent 1px);
  background-size: 48px 48px;
}
.brand-content { position: relative; z-index: 1; text-align: center; color: #fff; }
.brand-icon {
  width: 80px; height: 80px; margin: 0 auto 2rem;
  background: linear-gradient(135deg, #3B82F6, #6366F1);
  border-radius: 20px; display: flex; align-items: center; justify-content: center;
  box-shadow: 0 8px 32px rgba(59,130,246,0.3);
}
.brand-icon svg { width: 40px; height: 40px; color: #fff; }
.brand-title { font-size: 2rem; font-weight: 700; letter-spacing: -0.02em; margin-bottom: 0.5rem; }
.brand-desc { font-size: 0.95rem; color: rgba(255,255,255,0.6); line-height: 1.7; max-width: 320px; margin: 0 auto; }
.brand-features { margin-top: 2.5rem; text-align: left; max-width: 300px; }
.brand-feature { display: flex; align-items: center; gap: 0.75rem; padding: 0.6rem 0; color: rgba(255,255,255,0.7); font-size: 0.85rem; }
.brand-feature svg { width: 18px; height: 18px; color: #3B82F6; flex-shrink: 0; }
.form-panel { flex: 1; display: flex; align-items: center; justify-content: center; padding: 2rem; background: var(--bg-page); }
.form-container { width: 100%; max-width: 400px; }
.form-header { margin-bottom: 2rem; }
.form-header h2 { font-size: 1.5rem; font-weight: 700; }
.form-header p { font-size: 0.9rem; color: var(--text-secondary); margin-top: 0.3rem; }
.form-group { margin-bottom: 1.2rem; }
.form-group label { display: block; font-size: 0.85rem; font-weight: 500; margin-bottom: 0.4rem; color: var(--text-primary); }
.form-input {
  width: 100%; padding: 0.7rem 0.85rem; font-size: 0.9rem;
  border: 1px solid var(--border-color); border-radius: var(--radius);
  background: var(--bg-card); color: var(--text-primary);
  font-family: var(--font-sans); outline: none; transition: all 0.15s;
}
.form-input:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(37,99,235,0.1); }
.form-error { color: var(--danger); font-size: 0.85rem; margin-bottom: 1rem; padding: 0.5rem; background: #FEF2F2; border-radius: var(--radius-sm); }
.btn-submit {
  width: 100%; padding: 0.7rem; border: none; border-radius: var(--radius);
  background: var(--accent); color: #fff; font-size: 0.9rem; font-weight: 600;
  font-family: var(--font-sans); cursor: pointer;
  transition: background 0.15s;
}
.btn-submit:hover { background: var(--accent-hover); }
.btn-submit:disabled { opacity: 0.6; cursor: not-allowed; }
.form-footer { margin-top: 1.5rem; text-align: center; font-size: 0.85rem; color: var(--text-secondary); }
.form-footer a { color: var(--accent); font-weight: 500; }
</style>
