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
const showRegisterPassword = ref(false)
const loading = ref(false)
const errorMsg = ref('')

// 后端英文错误 → 中文提示映射
const errorMessages: Record<string, string> = {
  'Invalid email or password': '邮箱或密码错误',
  'Email already registered': '该邮箱已被注册',
  'Username already taken': '该用户名已被使用',
  'Passwords do not match': '两次密码不一致',
  'Invalid or expired token': '登录已过期，请重新登录',
  'Not authenticated': '未登录，请先登录',
  'Not Found': '接口地址不存在',
  'Internal Server Error': '服务器内部错误，请稍后重试',
}

function translateError(err: any): string {
  // 1. 优先取后端 detail 字段
  const detail = err.response?.data?.detail
  if (detail && typeof detail === 'string') {
    const cn = errorMessages[detail]
    if (cn) return cn
    // 如果后端返回了不在映射表里的英文，简短转中文
    if (/^[A-Z]/.test(detail)) return detail
    return detail
  }

  // 2. 网络错误（后端没响应）
  if (err.code === 'ERR_NETWORK') return '无法连接到服务器，请检查后端是否启动'
  if (err.code === 'ECONNABORTED') return '请求超时，请重试'

  // 3. HTTP 状态码兜底
  const status = err.response?.status
  if (status === 401) return '邮箱或密码错误'
  if (status === 403) return '没有权限执行此操作'
  if (status === 404) return '请求的资源不存在'
  if (status >= 500) return '服务器内部错误，请稍后重试'

  // 4. 最后兜底
  return '操作失败，请重试'
}

function toggleMode() {
  isLogin.value = !isLogin.value
  errorMsg.value = ''
  showRegisterPassword.value = false
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
      // 注册成功后：显示成功消息 → 清空表单 → 切换到登录模式
      errorMsg.value = ''
      alert('注册成功！请使用你的邮箱和密码登录。')
      email.value = email.value
      password.value = ''
      username.value = ''
      confirmPassword.value = ''
      isLogin.value = true
    }
  } catch (e: any) {
    console.error('[LoginView] 操作失败:', e)
    errorMsg.value = translateError(e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <!-- Toggle -->
    <div class="mode-toggle">
      <button
        :class="['mode-btn', { active: isLogin }]"
        @click="isLogin = true"
      >
        登录
      </button>
      <button
        :class="['mode-btn', { active: !isLogin }]"
        @click="isLogin = false"
      >
        注册
      </button>
    </div>

    <div class="login-container">
      <!-- Logo -->
      <div class="logo">
        <div class="logo__icon">N<span>ote</span>Web</div>
        <div class="logo__sub">知 识 型 阅 读 工 作 台</div>
      </div>

      <!-- Form -->
      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label for="email">{{ isLogin ? '邮箱 / 用户名' : '邮箱' }}</label>
          <input
            id="email"
            v-model="email"
            type="text"
            class="form-input"
            placeholder="your@email.com"
            autocomplete="email"
          />
        </div>

        <!-- 注册模式下的用户名（邮箱之后、密码之前） -->
        <div v-if="!isLogin" class="form-group">
          <label for="username">用户名</label>
          <input
            id="username"
            v-model="username"
            type="text"
            class="form-input"
            placeholder="给自己取个名字"
            autocomplete="username"
          />
        </div>

        <div class="form-group">
          <label for="password">{{ isLogin ? '密码' : '设置密码' }}</label>
          <div :class="['password-field', { 'password-field--with-toggle': !isLogin }]">
            <input
              id="password"
              v-model="password"
              :type="!isLogin && showRegisterPassword ? 'text' : 'password'"
              class="form-input"
              placeholder="········"
              :autocomplete="isLogin ? 'current-password' : 'new-password'"
            />
            <button
              v-if="!isLogin"
              type="button"
              class="password-toggle"
              :aria-label="showRegisterPassword ? '隐藏密码' : '显示密码'"
              :aria-pressed="showRegisterPassword"
              @click="showRegisterPassword = !showRegisterPassword"
            >
              <svg v-if="!showRegisterPassword" viewBox="0 0 24 24" aria-hidden="true">
                <path d="M2.25 12s3.5-6.5 9.75-6.5S21.75 12 21.75 12s-3.5 6.5-9.75 6.5S2.25 12 2.25 12Z" />
                <circle cx="12" cy="12" r="2.75" />
              </svg>
              <svg v-else viewBox="0 0 24 24" aria-hidden="true">
                <path d="m3.75 4.25 16.5 15.5" />
                <path d="M9.2 5.85A9.3 9.3 0 0 1 12 5.5c6.25 0 9.75 6.5 9.75 6.5a16.4 16.4 0 0 1-3.05 3.75" />
                <path d="M14.1 14.3A2.75 2.75 0 0 1 9.7 9.9" />
                <path d="M6.65 7.35A16.1 16.1 0 0 0 2.25 12s3.5 6.5 9.75 6.5c1.4 0 2.67-.32 3.78-.82" />
              </svg>
            </button>
          </div>
        </div>

        <!-- 注册模式下的确认密码 -->
        <div v-if="!isLogin" class="form-group">
          <label for="confirm-password">确认密码</label>
          <input
            id="confirm-password"
            v-model="confirmPassword"
            type="password"
            class="form-input"
            placeholder="再输一次"
            autocomplete="new-password"
          />
        </div>

        <div v-if="errorMsg" class="form-error">{{ errorMsg }}</div>

        <button
          type="submit"
          class="btn-submit"
          :disabled="loading"
        >
          {{ loading ? '处理中...' : isLogin ? '登 录' : '注 册' }}
        </button>
      </form>

      <div class="form-footer">
        <span>{{ isLogin ? '没有账号？' : '已有账号？' }}</span>
        <a href="#" @click.prevent="toggleMode">
          {{ isLogin ? '去注册 →' : '去登录 →' }}
        </a>
      </div>

      <!-- Social login -->
      <div class="divider">或</div>
      <div class="social-row">
        <button class="social-btn" title="微信登录">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M17 14.5c.6-.4 1-.9 1.2-1.5a4 4 0 0 0-1.7-5.2 4 4 0 0 0-5.4 1.5"/>
            <path d="M7 14.5c-.6-.4-1-.9-1.2-1.5a4 4 0 0 1 1.7-5.2 4 4 0 0 1 5.4 1.5"/>
            <path d="M12 20.5c2.5 0 4.7-1 6-2.5a4.5 4.5 0 0 0-4.5-5.5c-2.7 0-5 2-5 4.5 0 .7.2 1.4.5 2"/>
            <path d="M9.5 13a.5.5 0 1 0 0-1 .5.5 0 0 0 0 1Z"/>
            <path d="M14.5 13a.5.5 0 1 0 0-1 .5.5 0 0 0 0 1Z"/>
          </svg>
        </button>
        <button class="social-btn" title="Apple 登录">
          <svg viewBox="0 0 24 24" fill="currentColor">
            <path d="M17.05 20.28c-.98.95-2.05.8-3.08.35-1.09-.46-2.09-.48-3.24 0-1.44.62-2.2.44-3.06-.35C4.79 17.2 4.43 12.2 6.98 9.68c1.08-1.1 2.49-1.7 3.91-1.73 1.22.03 2.37.46 3.22 1.2.96-.49 2.01-.74 3.09-.71 1.84.04 3.24.72 4.25 2.04-1.66 1.02-2.49 2.45-2.75 4.26-.27 2.01.82 3.74 2.5 4.54-.47 1.23-1.07 2.44-1.95 3.5l-.2-.02zM12.03 7.25c-.14-2.02 1.65-3.75 3.63-3.88.28 2.22-1.66 4.01-3.63 3.88z"/>
          </svg>
        </button>
        <button class="social-btn" title="Google 登录">
          <svg viewBox="0 0 24 24" fill="currentColor">
            <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"/>
            <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
            <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
            <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
          </svg>
        </button>
      </div>

      <div class="page-footer">© 2025 NoteWeb · 知识型阅读工作台</div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--bg-page);
  background-image: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 1px,
    rgba(0, 0, 0, 0.008) 1px,
    rgba(0, 0, 0, 0.008) 2px
  );
  background-size: 100% 2px;
  overflow-x: hidden;
  font-family: var(--font-body);
}

.mode-toggle {
  position: fixed;
  top: 1.5rem;
  right: 1.5rem;
  display: flex;
  gap: 0.3rem;
  padding: 0.3rem;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}
.mode-btn {
  padding: 0.35rem 0.7rem;
  border: none;
  background: transparent;
  border-radius: 6px;
  font-family: var(--font-ui);
  font-size: 0.7rem;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s;
}
.mode-btn.active {
  background: var(--accent);
  color: #fff;
}
.mode-btn:not(.active):hover {
  color: var(--text-secondary);
}

.login-container {
  width: 100%;
  max-width: 420px;
  padding: 2rem;
}

.logo {
  text-align: center;
  margin-bottom: 3rem;
}
.logo__icon {
  font-family: var(--font-display);
  font-size: 2.5rem;
  font-weight: 600;
  font-style: italic;
  color: var(--accent);
  letter-spacing: -0.02em;
  line-height: 1;
}
.logo__icon span {
  font-style: normal;
  font-weight: 400;
  font-size: 2rem;
  color: var(--text-muted);
}
.logo__sub {
  font-family: var(--font-ui);
  font-size: 0.8rem;
  font-weight: 300;
  color: var(--text-muted);
  margin-top: 0.5rem;
  letter-spacing: 0.15em;
}

.form-group {
  margin-bottom: 1.75rem;
}
.form-group label {
  display: block;
  font-family: var(--font-ui);
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-secondary);
  letter-spacing: 0.05em;
  text-transform: uppercase;
  margin-bottom: 0.4rem;
}
.form-input {
  width: 100%;
  border: none;
  border-bottom: 1px solid var(--border-input);
  background: transparent;
  padding: 0.6rem 0 0.4rem;
  font-family: var(--font-body);
  font-size: 1rem;
  color: var(--text-primary);
  outline: none;
  transition: border-color 0.2s;
}
.form-input::placeholder {
  color: var(--text-muted);
  font-weight: 300;
  font-size: 0.9rem;
}
.form-input:focus {
  border-bottom-color: var(--accent);
}

.password-field {
  position: relative;
}

.password-field--with-toggle .form-input {
  padding-right: 2.4rem;
}

.password-toggle {
  position: absolute;
  right: 0;
  bottom: 0.25rem;
  width: 2rem;
  height: 2rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  transition: color 0.2s, background 0.2s;
}

.password-toggle:hover,
.password-toggle:focus-visible {
  color: var(--accent);
  background: var(--accent-light);
  outline: none;
}

.password-toggle svg {
  width: 1.05rem;
  height: 1.05rem;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.form-error {
  color: #DC2626;
  font-family: var(--font-ui);
  font-size: 0.8rem;
  margin-top: -1rem;
  margin-bottom: 1rem;
}

.btn-submit {
  display: block;
  width: 100%;
  padding: 0.8rem 0;
  margin-top: 2.5rem;
  background: var(--accent);
  color: #fff;
  border: none;
  border-radius: var(--radius-md);
  font-family: var(--font-ui);
  font-size: 0.95rem;
  font-weight: 500;
  letter-spacing: 0.06em;
  cursor: pointer;
  transition: background 0.2s, transform 0.15s;
}
.btn-submit:hover {
  background: var(--accent-hover);
}
.btn-submit:active {
  transform: scale(0.98);
}
.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.form-footer {
  text-align: center;
  margin-top: 1.8rem;
  font-family: var(--font-ui);
  font-size: 0.85rem;
  color: var(--text-muted);
}
.form-footer a {
  color: var(--accent);
  text-decoration: none;
  font-weight: 500;
  margin-left: 0.3rem;
}
.form-footer a:hover {
  text-decoration: underline;
}

.divider {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin: 2rem 0 1.5rem;
  color: var(--border-line, var(--border-color));
  font-family: var(--font-ui);
  font-size: 0.7rem;
  letter-spacing: 0.05em;
}
.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--border-line, var(--border-color));
}

.social-row {
  display: flex;
  justify-content: center;
  gap: 1rem;
}
.social-btn {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: 1px solid var(--border-color);
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: border-color 0.2s, background 0.2s;
  color: var(--text-secondary);
}
.social-btn:hover {
  border-color: var(--accent);
  background: var(--accent-light);
  color: var(--accent);
}
.social-btn svg {
  width: 20px;
  height: 20px;
}

.page-footer {
  text-align: center;
  margin-top: 3rem;
  font-family: var(--font-ui);
  font-size: 0.7rem;
  color: var(--text-muted);
  letter-spacing: 0.03em;
}

@media (max-width: 768px) {
  .login-container { max-width: 380px; padding: 1.5rem; }
  .logo { margin-bottom: 2.5rem; }
  .logo__icon { font-size: 2.2rem; }
  .social-row { gap: 0.75rem; }
}
@media (max-width: 480px) {
  .login-page { align-items: flex-start; padding-top: 3rem; }
  .login-container { max-width: 100%; padding: 1.2rem; }
  .logo { margin-bottom: 2rem; }
  .logo__icon { font-size: 1.8rem; }
  .logo__sub { font-size: 0.7rem; }
  .mode-toggle { top: 0.6rem; right: 0.6rem; padding: 0.2rem; }
  .mode-btn { font-size: 0.65rem; padding: 0.25rem 0.5rem; }
  .form-group { margin-bottom: 1.25rem; }
  .form-input { font-size: 0.9rem; padding: 0.5rem 0; }
  .btn-submit { padding: 0.7rem 0; font-size: 0.9rem; }
  .form-footer { font-size: 0.8rem; }
  .social-btn { width: 40px; height: 40px; }
  .page-footer { font-size: 0.6rem; margin-top: 2rem; }
}
@media (max-width: 360px) {
  .login-container { padding: 1rem; }
  .logo__icon { font-size: 1.5rem; }
  .social-row { gap: 0.5rem; }
  .social-btn { width: 36px; height: 36px; }
  .social-btn svg { width: 16px; height: 16px; }
}
</style>
