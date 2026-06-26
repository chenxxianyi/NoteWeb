<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import AppLayout from '../components/layout/AppLayout.vue'
import { useAuthStore } from '../stores/authStore'
import { useSettingsStore } from '../stores/settingsStore'
import { useDocumentStore } from '../stores/documentStore'
import * as aiApi from '../api/ai'

const authStore = useAuthStore()
const settingsStore = useSettingsStore()
const documentStore = useDocumentStore()
const { setTheme, setFont } = settingsStore

const username = computed(() => authStore.user?.username || '用户')
const email = computed(() => authStore.user?.email || '—')
const readingMode = computed(() => settingsStore.readingMode)

// Modal states
const showPasswordModal = ref(false)
const showProfileModal = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const editUsername = ref('')
const editEmail = ref('')
const avatarFile = ref<File | null>(null)
const avatarPreview = ref<string>('')

// AI settings are saved to the backend; localStorage keeps the filled key visible across reloads.
const savedAIProvider = localStorage.getItem('ai_provider')
const aiProvider = ref(savedAIProvider && savedAIProvider !== 'Mock API' ? savedAIProvider : 'DeepSeek')
const aiModel = ref(localStorage.getItem('ai_model') || 'deepseek-chat')
const aiKey = ref(localStorage.getItem('ai_key') || '')
const aiBaseUrl = ref(localStorage.getItem('ai_base_url') || '')
const aiSaved = ref(false)
const showNotification = ref(false)
const notificationMessage = ref('')
const loadingAction = ref(false)

// AI Models for each provider
const aiModels: Record<string, string[]> = {
  'OpenAI': ['gpt-4o', 'gpt-4o-mini', 'gpt-4-turbo', 'gpt-4', 'gpt-3.5-turbo'],
  'DeepSeek': ['deepseek-chat', 'deepseek-coder', 'deepseek-reasoner'],
  '自定义': ['custom-model'],
}

// Get available models for current provider
const availableModels = computed(() => {
  return aiModels[aiProvider.value] || ['custom-model']
})

// Reset model when provider changes
watch(aiProvider, (newProvider) => {
  const models = aiModels[newProvider] || ['custom-model']
  if (!models.includes(aiModel.value)) {
    aiModel.value = models[0]
  }
  if (newProvider === 'DeepSeek') {
    aiBaseUrl.value = ''
  }
})

function toProviderAPIValue(provider: string) {
  if (provider === 'DeepSeek') return 'deepseek'
  if (provider === 'OpenAI') return 'openai'
  return 'custom'
}

function toProviderLabel(provider: string) {
  if (provider === 'openai') return 'OpenAI'
  if (provider === 'custom') return '自定义'
  return 'DeepSeek'
}

async function loadAISettings() {
  try {
    const response = await aiApi.getProviderConfig()
    aiProvider.value = toProviderLabel(response.data.provider)
    const models = aiModels[aiProvider.value] || ['custom-model']
    aiModel.value = response.data.model && models.includes(response.data.model)
      ? response.data.model
      : models[0]
    aiBaseUrl.value = aiProvider.value === 'DeepSeek' ? '' : response.data.base_url || ''
  } catch (e: any) {
    const msg = e?.response?.data?.detail || e?.message || 'AI 配置读取失败'
    showNotificationToast(msg)
  }
}

function showNotificationToast(msg: string) {
  notificationMessage.value = msg
  showNotification.value = true
  setTimeout(() => {
    showNotification.value = false
  }, 2000)
}

async function saveAISettings() {
  const provider = toProviderAPIValue(aiProvider.value)
  const apiKey = aiKey.value.trim()
  const baseUrl = provider === 'deepseek' ? '' : aiBaseUrl.value.trim()

  if (provider !== 'deepseek' && !baseUrl) {
    showNotificationToast('请填写 AI Base URL')
    return
  }

  loadingAction.value = true
  try {
    await aiApi.updateProviderConfig({
      provider,
      model: aiModel.value,
      api_key: apiKey || undefined,
      base_url: baseUrl || undefined,
    })

    localStorage.setItem('ai_provider', aiProvider.value)
    localStorage.setItem('ai_model', aiModel.value)
    localStorage.setItem('ai_key', apiKey)
    localStorage.setItem('ai_base_url', baseUrl)
    aiSaved.value = true
    showNotificationToast('AI 配置已保存')
    setTimeout(() => { aiSaved.value = false }, 2000)
  } catch (e: any) {
    const msg = e?.response?.data?.detail || e?.message || 'AI 配置保存失败'
    showNotificationToast(msg)
  } finally {
    loadingAction.value = false
  }
}

function handleReadingModeChange() {
  const newValue = !readingMode.value
  settingsStore.setReadingMode(newValue)
  showNotificationToast(newValue ? '阅读模式已开启' : '阅读模式已关闭')
}

// Password change
function openPasswordModal() {
  oldPassword.value = ''
  newPassword.value = ''
  confirmPassword.value = ''
  showPasswordModal.value = true
}

async function changePassword() {
  if (!oldPassword.value || !newPassword.value) {
    showNotificationToast('请填写所有密码字段')
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    showNotificationToast('新密码两次输入不一致')
    return
  }
  if (newPassword.value.length < 6) {
    showNotificationToast('新密码长度至少6位')
    return
  }

  loadingAction.value = true
  try {
    await authStore.changePassword(oldPassword.value, newPassword.value)
    showPasswordModal.value = false
    showNotificationToast('密码修改成功')
  } catch (e: any) {
    const msg = e?.response?.data?.detail || e?.message || '密码修改失败'
    showNotificationToast(msg)
  } finally {
    loadingAction.value = false
  }
}

// Avatar upload
function handleAvatarSelect(event: Event) {
  const target = event.target as HTMLInputElement
  if (!target.files?.length) return
  avatarFile.value = target.files[0]
  avatarPreview.value = URL.createObjectURL(avatarFile.value)
}

async function uploadAvatar() {
  if (!avatarFile.value) {
    showNotificationToast('请先选择头像图片')
    return
  }

  loadingAction.value = true
  try {
    await authStore.uploadAvatar(avatarFile.value)
    avatarFile.value = null
    avatarPreview.value = ''
    showNotificationToast('头像上传成功')
  } catch (e: any) {
    const msg = e?.response?.data?.detail || e?.message || '头像上传失败'
    showNotificationToast(msg)
  } finally {
    loadingAction.value = false
  }
}

// Profile edit
function openProfileModal() {
  editUsername.value = authStore.user?.username || ''
  editEmail.value = authStore.user?.email || ''
  showProfileModal.value = true
}

async function updateProfile() {
  loadingAction.value = true
  try {
    await authStore.updateProfile(editUsername.value, editEmail.value)
    showProfileModal.value = false
    showNotificationToast('资料更新成功')
  } catch (e: any) {
    const msg = e?.response?.data?.detail || e?.message || '资料更新失败'
    showNotificationToast(msg)
  } finally {
    loadingAction.value = false
  }
}


const storageUsed = computed(() => {
  const bytes = authStore.user?.storage_used || 0
  if (bytes >= 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GB`
  if (bytes >= 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(0)} KB`
  return `${bytes} B`
})

const storageLimit = computed(() => {
  const bytes = authStore.user?.storage_limit || 1073741824
  if (bytes >= 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024 / 1024).toFixed(0)} GB`
  return `${(bytes / 1024 / 1024).toFixed(0)} MB`
})

const storagePercent = computed(() => {
  const used = authStore.user?.storage_used || 0
  const limit = authStore.user?.storage_limit || 1
  return Math.min(100, Math.round((used / limit) * 100))
})

// Get real document count from documentStore
const documentCount = computed(() => documentStore.documents.length)

// Export data function
async function exportData() {
  // Ensure documents are loaded
  if (documentStore.documents.length === 0) {
    await documentStore.fetchDocuments()
  }

  const exportPayload = {
    exportDate: new Date().toISOString(),
    version: '1.0',
    user: {
      username: authStore.user?.username,
      email: authStore.user?.email,
      storageUsed: authStore.user?.storage_used,
      storageLimit: authStore.user?.storage_limit,
    },
    settings: {
      theme: settingsStore.theme,
      font: settingsStore.font,
      readingMode: settingsStore.readingMode,
    },
    statistics: {
      documentCount: documentStore.documents.length,
    },
    // Note: We don't export actual document content, AI keys, or sensitive data
    // Users should backup their documents separately
  }

  const blob = new Blob([JSON.stringify(exportPayload, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  const dateStr = new Date().toISOString().split('T')[0]
  a.href = url
  a.download = `noteweb-export-${dateStr}.json`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)

  showNotificationToast('数据已导出')
}

onMounted(async () => {
  // Ensure documents are loaded for document count
  if (documentStore.documents.length === 0) {
    await documentStore.fetchDocuments()
  }
  // Fetch settings from backend
  await Promise.allSettled([
    settingsStore.fetchSettings(),
    loadAISettings(),
  ])
})
</script>

<template>
  <AppLayout>
    <div class="settings-page">
      <!-- Notification Toast -->
      <div v-if="showNotification" class="notification-toast">
        {{ notificationMessage }}
      </div>

      <!-- Password Modal -->
      <div v-if="showPasswordModal" class="modal-overlay" @click.self="showPasswordModal = false">
        <div class="modal">
          <h3 class="modal__title">修改密码</h3>
          <div class="modal__content">
            <input v-model="oldPassword" class="input modal__input" type="password" placeholder="旧密码" />
            <input v-model="newPassword" class="input modal__input" type="password" placeholder="新密码（至少6位）" />
            <input v-model="confirmPassword" class="input modal__input" type="password" placeholder="确认新密码" />
          </div>
          <div class="modal__actions">
            <button class="btn" @click="showPasswordModal = false">取消</button>
            <button class="btn btn--primary" :disabled="loadingAction" @click="changePassword">
              {{ loadingAction ? '处理中...' : '确认修改' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Profile Modal -->
      <div v-if="showProfileModal" class="modal-overlay" @click.self="showProfileModal = false">
        <div class="modal">
          <h3 class="modal__title">修改资料</h3>
          <div class="modal__content">
            <input v-model="editUsername" class="input modal__input" type="text" placeholder="用户名" />
            <input v-model="editEmail" class="input modal__input" type="email" placeholder="邮箱" />
          </div>
          <div class="modal__actions">
            <button class="btn" @click="showProfileModal = false">取消</button>
            <button class="btn btn--primary" :disabled="loadingAction" @click="updateProfile">
              {{ loadingAction ? '处理中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>


      <div class="main-inner">
        <h1 class="page-title">设置</h1>

        <!-- Profile -->
        <div class="section">
          <h2 class="section__title">用户资料</h2>
          <div class="section__card">
            <div class="setting-item">
              <div class="avatar-upload" @click="($refs.avatarInput as HTMLInputElement).click()">
                <img v-if="authStore.user?.avatar" :src="authStore.user.avatar" class="avatar-img" />
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="28" height="28"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
                <input ref="avatarInput" type="file" accept="image/*" style="display:none" @change="handleAvatarSelect" />
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">头像</div>
                <div class="setting-item__desc">支持 JPG / PNG，建议 256×256</div>
              </div>
              <button class="btn" :disabled="!avatarFile" @click="uploadAvatar">
                {{ avatarFile ? '上传' : '选择图片' }}
              </button>
            </div>
            <div class="setting-item">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">用户名</div>
                <div class="setting-item__desc">{{ username }}</div>
              </div>
              <button class="btn" @click="openProfileModal">修改</button>
            </div>
            <div class="setting-item">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">邮箱</div>
                <div class="setting-item__desc">{{ email }}</div>
              </div>
              <button class="btn" @click="openProfileModal">修改</button>
            </div>
            <div class="setting-item" style="border-bottom:none;">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">密码</div>
                <div class="setting-item__desc">********</div>
              </div>
              <button class="btn" @click="openPasswordModal">修改密码</button>
            </div>
          </div>
        </div>

        <!-- Appearance -->
        <div class="section">
          <h2 class="section__title">外观</h2>
          <div class="section__card">
            <div class="setting-item">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><circle cx="13.5" cy="6.5" r="0.5" fill="currentColor"/><circle cx="17.5" cy="10.5" r="0.5" fill="currentColor"/><circle cx="8.5" cy="7.5" r="0.5" fill="currentColor"/><circle cx="6.5" cy="12.5" r="0.5" fill="currentColor"/><path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.555-2.503 5.555-5.554C21.965 6.012 17.461 2 12 2z"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">主题</div>
                <div class="setting-item__desc">纸书质感 · 适合长时间阅读</div>
              </div>
              <div class="theme-dots">
                <button 
                  :class="['btn', 'theme-dot', { active: settingsStore.theme === 'warm' }]" 
                  style="background:#C67A4E;" 
                  title="暖色主题"
                  @click="setTheme('warm')"
                ></button>
                <button 
                  :class="['btn', 'theme-dot', { active: settingsStore.theme === 'blue' }]" 
                  style="background:#2563EB;" 
                  title="蓝色主题"
                  @click="setTheme('blue')"
                ></button>
                <button 
                  :class="['btn', 'theme-dot', { active: settingsStore.theme === 'dark' }]" 
                  style="background:#1E293B;" 
                  title="深色主题"
                  @click="setTheme('dark')"
                ></button>
              </div>
            </div>
            <div class="setting-item">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">字体</div>
                <div class="setting-item__desc">正文使用衬线字体，更适合阅读</div>
              </div>
              <select :value="settingsStore.font" class="input" style="width:140px;cursor:pointer;" @change="setFont(($event.target as HTMLSelectElement).value)">
                <option value="Noto Serif SC">Noto Serif SC</option>
                <option value="思源宋体">思源宋体</option>
                <option value="Inter">Inter</option>
                <option value="系统默认">系统默认</option>
              </select>
            </div>
            <div class="setting-item" style="border-bottom:none;">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">阅读模式</div>
                <div class="setting-item__desc">沉浸式阅读体验 · 纸书质感 · 暖光护眼 · 自动收起侧边栏</div>
              </div>
              <div
                :class="['toggle', { active: readingMode }]"
                @click="handleReadingModeChange"
              ></div>
            </div>
          </div>
        </div>

        <!-- Storage -->
        <div class="section">
          <h2 class="section__title">存储空间</h2>
          <div class="section__card">
            <div class="setting-item">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><rect x="2" y="2" width="20" height="8" rx="2" ry="2"/><rect x="2" y="14" width="20" height="8" rx="2" ry="2"/><line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">已用空间</div>
                <div class="setting-item__desc">{{ storageUsed }} / {{ storageLimit }}</div>
              </div>
              <div class="progress-bar"><div class="progress-bar__fill" :style="{ width: storagePercent + '%' }"></div></div>
            </div>
            <div class="setting-item" style="border-bottom:none;">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">文档数量</div>
                <div class="setting-item__desc">{{ documentCount }} 篇</div>
              </div>
              <button class="btn" @click="exportData">导出数据</button>
            </div>
          </div>
        </div>

        <!-- AI -->
        <div class="section">
          <h2 class="section__title">AI 配置</h2>
          <div class="section__card">
            <div class="setting-item">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><rect x="4" y="4" width="16" height="16" rx="2" ry="2"/><rect x="9" y="9" width="6" height="6"/><line x1="9" y1="1" x2="9" y2="4"/><line x1="15" y1="20" x2="15" y2="23"/><line x1="20" y1="9" x2="23" y2="9"/><line x1="20" y1="14" x2="23" y2="14"/><line x1="1" y1="9" x2="4" y2="9"/><line x1="1" y1="14" x2="4" y2="14"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">AI 提供商</div>
                <div class="setting-item__desc">选择 AI 阅读助手的后端服务</div>
              </div>
              <select v-model="aiProvider" class="input" style="width:160px;cursor:pointer;">
                <option>DeepSeek</option>
                <option>OpenAI</option>
                <option>自定义</option>
              </select>
            </div>
            <div class="setting-item">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">模型</div>
                <div class="setting-item__desc">选择要使用的 AI 模型</div>
              </div>
              <div class="model-select-wrapper">
                <select v-model="aiModel" class="input" style="width:160px;cursor:pointer;">
                  <option v-for="model in availableModels" :key="model" :value="model">{{ model }}</option>
                </select>
                <input 
                  v-if="aiProvider === '自定义'" 
                  v-model="aiModel" 
                  class="input input--long" 
                  type="text" 
                  placeholder="输入自定义模型名称"
                  style="margin-left:0.5rem;"
                />
              </div>
            </div>
            <div class="setting-item">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">API Key</div>
                <div class="setting-item__desc">用于调用真实模型，调用失败会直接提示错误</div>
              </div>
              <input v-model="aiKey" class="input input--long" type="password" placeholder="sk-xxxxxxxxxxxxxxxx" />
            </div>
            <div v-if="aiProvider !== 'DeepSeek'" class="setting-item">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">Base URL</div>
                <div class="setting-item__desc">API 端点地址，支持 OpenAI 兼容服务或中转站</div>
              </div>
              <input v-model="aiBaseUrl" class="input input--long" type="text" placeholder="https://api.openai.com/v1" />
            </div>
            <div class="setting-item" style="border-bottom:none;">
              <div></div>
              <div class="setting-item__info">
                <div v-if="aiSaved" class="setting-item__desc" style="color:#16A34A;">✅ 已保存</div>
              </div>
              <button class="btn btn--primary" :disabled="loadingAction" @click="saveAISettings">
                {{ loadingAction ? '保存中...' : '保存配置' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.settings-page {
  padding: 2rem;
  display: flex;
  justify-content: center;
  min-height: 100vh;
}
.main-inner { width: 100%; max-width: 800px; }

.notification-toast {
  position: fixed;
  top: 4.5rem;
  left: 50%;
  transform: translateX(-50%);
  z-index: 100;
  padding: 0.75rem 1.5rem;
  background: var(--bg-card);
  border: 1px solid var(--accent);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  font-family: var(--font-ui);
  font-size: 0.875rem;
  color: var(--accent);
  animation: fadeInDown 0.3s ease;
}

@keyframes fadeInDown {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

.page-title { font-family: var(--font-display); font-size: 1.6rem; font-weight: 500; margin-bottom: 2rem; }

.section { margin-bottom: 2rem; }
.section__title { font-family: var(--font-display); font-size: 1.1rem; font-weight: 500; margin-bottom: 1rem; color: var(--text-primary); }
.section__card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-lg); overflow: hidden; }

.setting-item { display: flex; align-items: center; padding: 1rem 1.25rem; border-bottom: 1px solid var(--border-color); gap: 1rem; }
.setting-item:last-child { border-bottom: none; }
.setting-item__icon { width: 40px; height: 40px; border-radius: 8px; background: var(--accent-light); display: flex; align-items: center; justify-content: center; color: var(--accent); flex-shrink: 0; }
.setting-item__icon svg { width: 20px; height: 20px; }
.setting-item__info { flex: 1; min-width: 0; }
.setting-item__label { font-size: 0.9rem; font-weight: 500; color: var(--text-primary); }
.setting-item__desc { font-family: var(--font-ui); font-size: 0.75rem; color: var(--text-muted); margin-top: 0.15rem; }
.setting-item__action { flex-shrink: 0; }

.theme-dots { display: flex; gap: 0.4rem; }
.theme-dot {
  border-radius: 50%;
  width: 32px;
  height: 32px;
  padding: 0;
  min-width: 32px;
  border: 2px solid transparent;
  transition: all 0.2s;
}
.theme-dot:hover {
  transform: scale(1.1);
}
.theme-dot.active {
  border-color: var(--text-primary);
  box-shadow: 0 0 0 2px var(--bg-page);
}

.input { padding: 0.5rem 0.8rem; border: 1px solid var(--border-color); border-radius: var(--radius); background: var(--bg-page); font-family: var(--font-ui); font-size: 0.85rem; color: var(--text-primary); outline: none; width: 200px; }
.input:focus { border-color: var(--accent); }
.input--long { width: 260px; }

.model-select-wrapper { display: flex; align-items: center; gap: 0.5rem; }

.btn { padding: 0.4rem 1rem; border: 1px solid var(--border-color); border-radius: 20px; background: var(--bg-card); font-family: var(--font-ui); font-size: 0.8rem; color: var(--text-secondary); cursor: pointer; transition: all 0.12s; }
.btn:hover { border-color: var(--accent); color: var(--accent); }
.btn--primary { background: var(--accent); color: #fff; border-color: var(--accent); }
.btn--primary:hover { opacity: 0.9; }

.toggle { position: relative; width: 40px; height: 22px; background: var(--border-color); border-radius: 11px; cursor: pointer; transition: background 0.2s; flex-shrink: 0; }
.toggle.active { background: var(--accent); }
.toggle::after { content: ''; position: absolute; top: 2px; left: 2px; width: 18px; height: 18px; border-radius: 50%; background: #fff; transition: transform 0.2s; }
.toggle.active::after { transform: translateX(18px); }

.progress-bar { height: 6px; background: var(--border-color); border-radius: 3px; overflow: hidden; width: 200px; }
.progress-bar__fill { height: 100%; background: var(--accent); border-radius: 3px; }

.avatar-upload { width: 64px; height: 64px; border-radius: 50%; background: var(--accent-light); display: flex; align-items: center; justify-content: center; color: var(--accent); cursor: pointer; position: relative; overflow: hidden; flex-shrink: 0; }
.avatar-upload svg { width: 28px; height: 28px; }
.avatar-img { width: 100%; height: 100%; object-fit: cover; }

/* Modal styles */
.modal-overlay {
  position: fixed;
  z-index: 1000;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.modal {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  max-width: 400px;
  width: 100%;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.modal__title {
  padding: 1.25rem;
  font-size: 1.1rem;
  font-weight: 600;
  border-bottom: 1px solid var(--border-color);
}

.modal__content {
  padding: 1.25rem;
}

.modal__input {
  width: 100%;
  margin-bottom: 1rem;
}

.modal__actions {
  padding: 1rem 1.25rem;
  border-top: 1px solid var(--border-color);
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
}

@media (max-width: 640px) {
  .settings-page { padding: 1.2rem; }
  .setting-item { flex-wrap: wrap; }
  .setting-item__action { width: 100%; }
  .input, .input--long { width: 100%; }
  .progress-bar { width: 100%; }
}
@media (max-width: 520px) {
  .settings-page { padding: 1rem; }
}
</style>
