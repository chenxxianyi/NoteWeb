<script setup lang="ts">
import { ref, computed } from 'vue'
import AppLayout from '../components/layout/AppLayout.vue'
import { useAuthStore } from '../stores/authStore'

const authStore = useAuthStore()
const username = computed(() => authStore.user?.username || '用户')
const email = computed(() => authStore.user?.email || '—')
const readingMode = ref(true)

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
</script>

<template>
  <AppLayout>
    <div class="settings-page">
      <div class="main-inner">
        <h1 class="page-title">设置</h1>

        <!-- Profile -->
        <div class="section">
          <h2 class="section__title">用户资料</h2>
          <div class="section__card">
            <div class="setting-item">
              <div class="avatar-upload">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="28" height="28"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">头像</div>
                <div class="setting-item__desc">支持 JPG / PNG，建议 256×256</div>
              </div>
              <button class="btn">上传头像</button>
            </div>
            <div class="setting-item">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">用户名</div>
                <div class="setting-item__desc">你的显示名称</div>
              </div>
              <input :value="username" class="input" type="text" readonly />
            </div>
            <div class="setting-item">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">邮箱</div>
                <div class="setting-item__desc">用于登录和通知</div>
              </div>
              <input :value="email" class="input input--long" type="email" readonly />
            </div>
            <div class="setting-item" style="border-bottom:none;">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">密码</div>
                <div class="setting-item__desc">********</div>
              </div>
              <button class="btn">修改密码</button>
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
                <button class="btn" style="border-radius:50%;width:32px;height:32px;padding:0;background:#C67A4E;border-color:#C67A4E;"></button>
                <button class="btn" style="border-radius:50%;width:32px;height:32px;padding:0;background:#2563EB;border-color:#2563EB;"></button>
                <button class="btn" style="border-radius:50%;width:32px;height:32px;padding:0;background:#1E293B;border-color:#1E293B;"></button>
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
              <select class="input" style="width:140px;cursor:pointer;">
                <option>Noto Serif SC</option>
                <option>思源宋体</option>
                <option>Inter</option>
              </select>
            </div>
            <div class="setting-item" style="border-bottom:none;">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">阅读模式</div>
                <div class="setting-item__desc">开启后阅读器背景模拟纸质书质感</div>
              </div>
              <div
                :class="['toggle', { active: readingMode }]"
                @click="readingMode = !readingMode"
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
                <div class="setting-item__desc">—</div>
              </div>
              <button class="btn">导出数据</button>
            </div>
          </div>
        </div>

        <!-- AI -->
        <div class="section">
          <h2 class="section__title">AI 配置</h2>
          <div class="section__card">
            <div class="setting-item">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><rect x="4" y="4" width="16" height="16" rx="2" ry="2"/><rect x="9" y="9" width="6" height="6"/><line x1="9" y1="1" x2="9" y2="4"/><line x1="15" y1="1" x2="15" y2="4"/><line x1="9" y1="20" x2="9" y2="23"/><line x1="15" y1="20" x2="15" y2="23"/><line x1="20" y1="9" x2="23" y2="9"/><line x1="20" y1="14" x2="23" y2="14"/><line x1="1" y1="9" x2="4" y2="9"/><line x1="1" y1="14" x2="4" y2="14"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">AI 提供商</div>
                <div class="setting-item__desc">选择 AI 阅读助手的后端服务</div>
              </div>
              <select class="input" style="width:160px;cursor:pointer;">
                <option>Mock API</option>
                <option>OpenAI</option>
                <option>DeepSeek</option>
                <option>自定义</option>
              </select>
            </div>
            <div class="setting-item" style="border-bottom:none;">
              <div class="setting-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label">API Key</div>
                <div class="setting-item__desc">暂不填则使用 Mock 服务</div>
              </div>
              <input class="input input--long" type="password" value="sk-xxxxxxxxxxxxxxxx" />
            </div>
          </div>
        </div>

        <!-- Danger Zone -->
        <div class="section">
          <h2 class="section__title" style="color:#DC2626;">危险操作</h2>
          <div class="section__card" style="border-color:#FECACA;">
            <div class="setting-item" style="border-bottom:none;">
              <div class="setting-item__icon" style="background:#FEE2E2;color:#DC2626;">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              </div>
              <div class="setting-item__info">
                <div class="setting-item__label" style="color:#DC2626;">删除账号</div>
                <div class="setting-item__desc">删除所有数据，此操作不可撤销</div>
              </div>
              <button class="btn" style="border-color:#FECACA;color:#DC2626;">删除账号</button>
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

.input { padding: 0.5rem 0.8rem; border: 1px solid var(--border-color); border-radius: var(--radius); background: var(--bg-page); font-family: var(--font-ui); font-size: 0.85rem; color: var(--text-primary); outline: none; width: 200px; }
.input:focus { border-color: var(--accent); }
.input--long { width: 260px; }

.btn { padding: 0.4rem 1rem; border: 1px solid var(--border-color); border-radius: 20px; background: var(--bg-card); font-family: var(--font-ui); font-size: 0.8rem; color: var(--text-secondary); cursor: pointer; transition: all 0.12s; }
.btn:hover { border-color: var(--accent); color: var(--accent); }
.btn--primary { background: var(--accent); color: #fff; border-color: var(--accent); }
.btn--primary:hover { background: var(--accent-hover); }

.toggle { position: relative; width: 40px; height: 22px; background: var(--border-color); border-radius: 11px; cursor: pointer; transition: background 0.2s; flex-shrink: 0; }
.toggle.active { background: var(--accent); }
.toggle::after { content: ''; position: absolute; top: 2px; left: 2px; width: 18px; height: 18px; border-radius: 50%; background: #fff; transition: transform 0.2s; }
.toggle.active::after { transform: translateX(18px); }

.progress-bar { height: 6px; background: var(--border-color); border-radius: 3px; overflow: hidden; width: 200px; }
.progress-bar__fill { height: 100%; background: var(--accent); border-radius: 3px; }

.avatar-upload { width: 64px; height: 64px; border-radius: 50%; background: var(--accent-light); display: flex; align-items: center; justify-content: center; color: var(--accent); cursor: pointer; position: relative; overflow: hidden; flex-shrink: 0; }
.avatar-upload svg { width: 28px; height: 28px; }

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
