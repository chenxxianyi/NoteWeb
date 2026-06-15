<script setup lang="ts">
import AppLayout from '../components/layout/AppLayout.vue'
import { useAuthStore } from '../stores/authStore'

const authStore = useAuthStore()
const user = authStore.user
</script>

<template>
  <AppLayout>
    <div class="topbar">
      <div class="topbar__left"><h1>设置</h1><p>管理你的账户和应用偏好</p></div>
    </div>

    <div class="settings-grid">
      <!-- Profile -->
      <div class="settings-section">
        <div class="settings-section__header">用户资料</div>
        <div class="settings-card">
          <div class="settings-row">
            <div class="settings-row__label">用户名</div>
            <div class="settings-row__value">{{ user?.username || '—' }}</div>
          </div>
          <div class="settings-row">
            <div class="settings-row__label">邮箱</div>
            <div class="settings-row__value">{{ user?.email || '—' }}</div>
          </div>
          <div class="settings-row">
            <div class="settings-row__label">头像</div>
            <div class="settings-row__value">
              <div class="avatar">{{ user?.username?.charAt(0).toUpperCase() || 'U' }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Storage -->
      <div class="settings-section">
        <div class="settings-section__header">存储空间</div>
        <div class="settings-card">
          <div class="settings-row">
            <div class="settings-row__label">已使用</div>
            <div class="settings-row__value">{{ user?.storage_used ? `${(user.storage_used / 1024 / 1024).toFixed(1)} MB` : '0 MB' }}</div>
          </div>
          <div class="settings-row">
            <div class="settings-row__label">总容量</div>
            <div class="settings-row__value">{{ user?.storage_limit ? `${(user.storage_limit / 1024 / 1024).toFixed(0)} MB` : '1 GB' }}</div>
          </div>
          <div class="progress-bar"><div class="progress-fill" :style="{ width: `${Math.min((user?.storage_used || 0) / (user?.storage_limit || 1) * 100, 100)}%` }"></div></div>
        </div>
      </div>

      <!-- Theme -->
      <div class="settings-section">
        <div class="settings-section__header">主题设置</div>
        <div class="settings-card">
          <div class="settings-row">
            <div class="settings-row__label">当前主题</div>
            <div class="settings-row__value"><span class="theme-badge">数字清爽</span></div>
          </div>
          <div class="settings-row" style="flex-direction:column;align-items:flex-start;gap:0.5rem;">
            <div class="settings-row__label">预览</div>
            <div style="display:flex;gap:0.5rem;">
              <div style="width:80px;height:48px;border-radius:var(--radius);background:#F1F5F9;border:2px solid var(--accent);display:flex;align-items:center;justify-content:center;font-size:0.6rem;color:var(--accent);font-weight:600;">B-数字</div>
              <div style="width:80px;height:48px;border-radius:var(--radius);background:#F5F0EB;border:1px solid var(--border-color);display:flex;align-items:center;justify-content:center;font-size:0.6rem;color:var(--text-muted);">A-纸书</div>
              <div style="width:80px;height:48px;border-radius:var(--radius);background:linear-gradient(135deg,#F5F0EB 50%,#F1F5F9 50%);border:1px solid var(--border-color);display:flex;align-items:center;justify-content:center;font-size:0.6rem;color:var(--text-muted);">C-双主题</div>
            </div>
          </div>
        </div>
      </div>

      <!-- AI -->
      <div class="settings-section">
        <div class="settings-section__header">AI 配置</div>
        <div class="settings-card">
          <div class="settings-row">
            <div class="settings-row__label">AI 服务地址</div>
            <div class="settings-row__value"><span style="color:var(--text-muted);">Mock 模式（开发中）</span></div>
          </div>
          <div class="settings-row" style="flex-direction:column;align-items:flex-start;gap:0.5rem;">
            <div class="settings-row__label">API Key</div>
            <input type="password" class="settings-input" placeholder="预留 API Key" disabled />
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.topbar { margin-bottom: 1.5rem; }
.topbar__left h1 { font-size: 1.35rem; font-weight: 600; }
.topbar__left p { font-size: 0.85rem; color: var(--text-secondary); margin-top: 0.15rem; }
.settings-grid { display: flex; flex-direction: column; gap: 1.5rem; max-width: 720px; }
.settings-section__header { font-size: 0.85rem; font-weight: 600; color: var(--text-secondary); margin-bottom: 0.6rem; text-transform: uppercase; letter-spacing: 0.03em; }
.settings-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-lg); padding: 1.25rem; }
.settings-row { display: flex; align-items: center; justify-content: space-between; padding: 0.6rem 0; }
.settings-row + .settings-row { border-top: 1px solid var(--border-color); }
.settings-row__label { font-size: 0.85rem; color: var(--text-secondary); }
.settings-row__value { font-size: 0.85rem; color: var(--text-primary); display: flex; align-items: center; gap: 0.5rem; }
.avatar { width: 36px; height: 36px; border-radius: 50%; background: var(--accent); color: #fff; display: flex; align-items: center; justify-content: center; font-weight: 600; font-size: 0.85rem; }
.progress-bar { height: 6px; background: var(--border-color); border-radius: 3px; margin-top: 0.5rem; overflow: hidden; }
.progress-fill { height: 100%; background: var(--accent); border-radius: 3px; transition: width 0.3s; }
.theme-badge { padding: 0.15rem 0.5rem; background: var(--accent-light); color: var(--accent); border-radius: 4px; font-size: 0.78rem; font-weight: 500; }
.settings-input { width: 100%; padding: 0.5rem 0.7rem; border: 1px solid var(--border-color); border-radius: var(--radius); font-family: var(--font-sans); font-size: 0.85rem; color: var(--text-primary); background: var(--bg-page); outline: none; }
.settings-input:disabled { opacity: 0.5; }
</style>
