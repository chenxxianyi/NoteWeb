<script setup lang="ts">
import AppLayout from '../components/layout/AppLayout.vue'

// Mock data matching the demo
const userName = '小明'
const now = new Date()
const hour = now.getHours()
const greeting = hour < 12 ? '早上' : hour < 18 ? '下午' : '晚上'

const weeklyStats = [
  { value: '12', label: '文档总数', trend: '3 本周新增', icon: 'book' },
  { value: '47', label: '批注总数', trend: '8 本周新增', icon: 'bookmark' },
  { value: '8', label: '笔记总数', trend: '2 本周新增', icon: 'pen-tool' },
  { value: '6.5h', label: '本周阅读时长', trend: '12% 较上周', icon: 'clock' },
]

const readingList = [
  { title: '深度学习入门：基于Python的理论与实现', meta: 'PDF · 第3章 · 32%', progress: '32%' },
  { title: 'Vue.js 设计与实现', meta: 'Markdown · 第12章 · 67%', progress: '67%' },
  { title: 'Rust 程序设计语言 中文版', meta: 'PDF · 第7章 · 45%', progress: '45%' },
]

const recentActivity = [
  { type: 'highlight', text: '"注意力机制的本质是让模型学会关注输入序列中重要的部分…"', meta: '高亮', doc: '深度学习入门 · 2 小时前', page: '第3页' },
  { type: 'note', text: '"虚拟 DOM 的性能优势不在于比真实 DOM 快…"', meta: '笔记', doc: 'Vue.js 设计与实现 · 昨天', page: '第12章' },
  { type: 'upload', title: 'Designing Data-Intensive Applications', meta: '上传', doc: 'PDF · 今天 14:30', info: '12.5 MB' },
  { type: 'upload', title: '系统设计面试笔记.md', meta: '上传', doc: 'Markdown · 昨天 20:15', info: '2.3 MB' },
]
</script>

<template>
  <AppLayout>
    <div class="dashboard">
      <div class="main-inner">
        <!-- Topbar -->
        <div class="topbar">
          <div class="topbar__greeting">
            <h1>{{ greeting }}好，{{ userName }}</h1>
            <p>你有 3 篇文档等待阅读 · 今日阅读 23 分钟</p>
          </div>
          <div class="topbar__actions">
            <button class="upload-btn">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
              <span>上传</span>
            </button>
            <div class="topbar__search">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16" class="search-icon"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
              <input type="text" placeholder="搜索文档..." />
            </div>
          </div>
        </div>

        <!-- Weekly Overview -->
        <div class="weekly-overview">
          <div class="weekly-overview__header">
            <h2>本周概览</h2>
            <a href="#">查看详细统计 →</a>
          </div>
          <div class="weekly-overview__grid">
            <div v-for="stat in weeklyStats" :key="stat.label" class="wo-item">
              <div class="wo-item__icon">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="22" height="22">
                  <rect x="4" y="2" width="16" height="20" rx="2" ry="2" v-if="stat.icon==='book'"/>
                  <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z" v-if="stat.icon==='bookmark'"/>
                  <path d="M12 20h9" v-if="stat.icon==='pen-tool'"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" v-if="stat.icon==='pen-tool'"/>
                  <circle cx="12" cy="12" r="10" v-if="stat.icon==='clock'"/><polyline points="12 6 12 12 16 14" v-if="stat.icon==='clock'"/>
                </svg>
              </div>
              <div class="wo-item__info">
                <div class="wo-item__value">{{ stat.value }}</div>
                <div class="wo-item__label">{{ stat.label }}</div>
              </div>
              <span class="wo-item__trend up">{{ stat.trend }}</span>
            </div>
          </div>
        </div>

        <!-- Timeline -->
        <div class="timeline-wrap">
          <!-- Continue Reading -->
          <div class="timeline">
            <div class="timeline__header">
              <h2>继续阅读</h2>
              <a href="#">查看全部 →</a>
            </div>
            <div class="tl-list">
              <div v-for="item in readingList" :key="item.title" class="tl-item">
                <div class="tl-item__icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="17" height="17"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"/><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"/></svg>
                </div>
                <div class="tl-item__body">
                  <div class="tl-item__title">{{ item.title }}</div>
                  <div class="tl-item__meta">{{ item.meta }}</div>
                </div>
                <span class="tl-item__progress">{{ item.progress }}</span>
              </div>
            </div>
          </div>

          <div class="timeline-divider"></div>

          <!-- Recent Activity -->
          <div class="timeline">
            <div class="timeline__header">
              <h2>最新动态</h2>
              <a href="#">查看全部 →</a>
            </div>
            <div class="tl-list">
              <div v-for="(item, idx) in recentActivity" :key="idx" class="tl-item">
                <div class="tl-item__icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="17" height="17">
                    <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z" v-if="item.type==='highlight'"/>
                    <path d="M12 20h9" v-if="item.type==='note'"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" v-if="item.type==='note'"/>
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" v-if="item.type==='upload'"/><polyline points="17 8 12 3 7 8" v-if="item.type==='upload'"/><line x1="12" y1="3" x2="12" y2="15" v-if="item.type==='upload'"/>
                  </svg>
                </div>
                <div class="tl-item__body">
                  <div v-if="item.text" class="tl-quote">{{ item.text }}</div>
                  <div v-if="item.title" class="tl-item__title">{{ item.title }}</div>
                  <div class="tl-item__meta">
                    <span :class="['tl-tag', item.type==='highlight' ? 'tag-hl' : item.type==='note' ? 'tag-note' : 'tag-up']">{{ item.type==='highlight' ? '高亮' : item.type==='note' ? '笔记' : '上传' }}</span>
                    {{ item.doc }}
                  </div>
                </div>
                <span class="tl-item__right">{{ item.page || item.info }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.dashboard {
  padding: 2rem 2.5rem;
  display: flex;
  justify-content: center;
  min-height: 100vh;
}
.main-inner {
  width: 100%;
  max-width: 1600px;
}

/* Topbar */
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2rem;
}
.topbar__greeting h1 {
  font-family: var(--font-display);
  font-size: 1.6rem;
  font-weight: 500;
  color: var(--text-primary);
}
.topbar__greeting p {
  font-family: var(--font-ui);
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-top: 0.2rem;
}
.topbar__actions {
  display: flex;
  gap: 0.6rem;
  align-items: center;
}
.topbar__search {
  position: relative;
}
.topbar__search input {
  font-family: var(--font-ui);
  font-size: 0.85rem;
  padding: 0.5rem 0.8rem 0.5rem 2.2rem;
  border: 1px solid var(--border-color);
  border-radius: 20px;
  background: var(--bg-card);
  color: var(--text-primary);
  outline: none;
  width: 200px;
  transition: border-color 0.15s;
}
.topbar__search input:focus { border-color: var(--accent); }
.topbar__search input::placeholder { color: var(--text-muted); }
.search-icon {
  position: absolute;
  left: 0.7rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
}

.upload-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.9rem 0.4rem 0.6rem;
  border: 1px solid var(--border-color);
  border-radius: 20px;
  background: var(--bg-card);
  font-family: var(--font-ui);
  font-size: 0.8rem;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s;
}
.upload-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--accent-light);
}
.upload-btn svg { width: 16px; height: 16px; }

/* Weekly Overview */
.weekly-overview {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  padding: 1.5rem 2rem;
  margin-bottom: 1.5rem;
}
.weekly-overview__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.25rem;
}
.weekly-overview__header h2 {
  font-family: var(--font-display);
  font-size: 1.1rem;
  font-weight: 500;
  color: var(--text-primary);
}
.weekly-overview__header a {
  font-family: var(--font-ui);
  font-size: 0.75rem;
  color: var(--text-muted);
  text-decoration: none;
}
.weekly-overview__header a:hover { color: var(--accent); }
.weekly-overview__grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}
.wo-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.5rem 0.75rem;
  border-radius: 8px;
  transition: background 0.15s;
}
.wo-item:hover { background: var(--accent-light); }
.wo-item__icon {
  width: 44px; height: 44px;
  border-radius: 10px;
  background: var(--accent-light);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent);
  flex-shrink: 0;
}
.wo-item__icon svg { width: 22px; height: 22px; }
.wo-item__info { flex: 1; min-width: 0; }
.wo-item__value {
  font-family: var(--font-display);
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.2;
}
.wo-item__label {
  font-family: var(--font-ui);
  font-size: 0.7rem;
  color: var(--text-muted);
  margin-top: 0.1rem;
}
.wo-item__trend {
  font-family: var(--font-ui);
  font-size: 0.65rem;
  padding: 0.15rem 0.4rem;
  border-radius: 10px;
  white-space: nowrap;
}
.wo-item__trend.up { background: #D1FAE5; color: #065F46; }
.wo-item__trend.up::before { content: "↑ "; }

/* Timeline */
.timeline-wrap {
  display: flex;
  gap: 0;
  margin-bottom: 2rem;
}
.timeline-wrap .timeline {
  flex: 1;
  margin-bottom: 0;
  min-width: 0;
  max-height: 420px;
  display: flex;
  flex-direction: column;
}
.timeline-wrap .timeline .tl-list {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}
.timeline-wrap .timeline:first-child { padding-right: 1.5rem; }
.timeline-wrap .timeline:last-child { padding-left: 1.5rem; }
.timeline-divider {
  width: 1px;
  flex-shrink: 0;
  background: var(--border-color);
  margin: 0.5rem 0;
}
.timeline {
  margin-bottom: 2rem;
}
.timeline__header {
  display: flex;
  align-items: baseline;
  gap: 0.6rem;
  margin-bottom: 0.6rem;
  padding-bottom: 0.4rem;
  border-bottom: 1px solid var(--border-color);
}
.timeline__header h2 {
  font-family: var(--font-display);
  font-size: 1rem;
  font-weight: 500;
  color: var(--text-primary);
}
.timeline__header a {
  font-family: var(--font-ui);
  font-size: 0.75rem;
  color: var(--text-muted);
  text-decoration: none;
  margin-left: auto;
}
.timeline__header a:hover { color: var(--accent); }
.tl-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 0.6rem;
  border-radius: 6px;
  transition: background 0.12s;
  cursor: pointer;
}
.tl-item:hover { background: var(--accent-light); }
.tl-item + .tl-item { border-top: 1px solid var(--border-color); }
.tl-item__icon {
  width: 34px; height: 34px;
  border-radius: 8px;
  background: var(--accent-light);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent);
  flex-shrink: 0;
}
.tl-item__icon svg { width: 17px; height: 17px; }
.tl-item__body { flex: 1; min-width: 0; }
.tl-item__title {
  font-size: 0.88rem;
  color: var(--text-primary);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.tl-item__meta {
  font-family: var(--font-ui);
  font-size: 0.7rem;
  color: var(--text-muted);
  margin-top: 0.1rem;
}
.tl-item__right {
  font-family: var(--font-ui);
  font-size: 0.7rem;
  color: var(--text-muted);
  flex-shrink: 0;
}
.tl-item__progress {
  font-family: var(--font-ui);
  font-size: 0.7rem;
  color: var(--accent);
  flex-shrink: 0;
}
.tl-tag {
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  font-size: 0.65rem;
  font-weight: 600;
}
.tag-hl { background: #FEF3C7; color: #92400E; }
.tag-note { background: #DBEAFE; color: #1E40AF; }
.tag-up { background: #D1FAE5; color: #065F46; }
.tl-quote {
  color: var(--text-secondary);
  font-style: italic;
  font-size: 0.82rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 1024px) {
  .main-inner { max-width: 100%; }
  .weekly-overview__grid { gap: 0.5rem; }
  .wo-item { padding: 0.4rem 0.5rem; gap: 0.6rem; }
  .wo-item__value { font-size: 1.2rem; }
  .wo-item__icon { width: 36px; height: 36px; }
  .wo-item__trend { font-size: 0.6rem; padding: 0.1rem 0.3rem; }
}
@media (max-width: 820px) {
  .timeline-wrap { flex-direction: column; }
  .timeline-divider { width: auto; height: 1px; margin: 0 0 0.5rem; }
  .timeline-wrap .timeline:first-child { padding-right: 0; }
  .timeline-wrap .timeline:last-child { padding-left: 0; }
  .timeline-wrap .timeline { max-height: 320px; }
  .weekly-overview__grid { grid-template-columns: repeat(2, 1fr); }
  .weekly-overview { padding: 1rem; }
  .dashboard { padding: 1.25rem; }
  .topbar__search input { width: 140px; }
}
@media (max-width: 520px) {
  .dashboard { padding: 0.8rem; }
  .topbar { flex-direction: column; align-items: flex-start; gap: 0.4rem; width: 100%; }
  .topbar__greeting { width: 100%; }
  .topbar__greeting h1 { font-size: 1.1rem; }
  .topbar__greeting p { font-size: 0.7rem; }
  .topbar__actions { width: 100%; display: flex; gap: 0.4rem; }
  .topbar__search { flex: 1; min-width: 0; }
  .topbar__search input { width: 100%; }
  .upload-btn span { display: none; }
  .upload-btn { padding: 0.3rem 0.6rem; flex-shrink: 0; }
  .weekly-overview { padding: 0.7rem; margin-bottom: 1rem; }
  .wo-item__icon { width: 26px; height: 26px; }
  .wo-item__value { font-size: 0.9rem; }
  .timeline-wrap .timeline { max-height: 220px; }
}
</style>
