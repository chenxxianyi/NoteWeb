<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '../components/layout/AppLayout.vue'
import { useAuthStore } from '../stores/authStore'
import { useDocumentStore } from '../stores/documentStore'
import { useNoteStore } from '../stores/noteStore'
import { useAnnotationStore } from '../stores/annotationStore'

const router = useRouter()
const authStore = useAuthStore()
const documentStore = useDocumentStore()
const noteStore = useNoteStore()
const annotationStore = useAnnotationStore()

const loading = ref(true)
import { ref } from 'vue'

const now = new Date()
const hour = now.getHours()
const greeting = hour < 12 ? '早上' : hour < 18 ? '下午' : '晚上'
const userName = computed(() => authStore.user?.username || '用户')

const weeklyStats = computed(() => [
  { value: String(documentStore.documents.length), label: '文档总数', icon: 'book' },
  { value: String(annotationStore.annotations.length), label: '批注总数', icon: 'bookmark' },
  { value: String(noteStore.notes.length), label: '笔记总数', icon: 'edit' },
  { value: '—', label: '本周阅读', icon: 'clock' },
])

const readingList = computed(() =>
  documentStore.documents
    .filter((d) => d.read_progress > 0)
    .map((d) => ({ title: d.title, meta: d.file_type.toUpperCase(), progress: d.read_progress, id: d.id }))
)

const recentActivity = computed(() => {
  const items: Array<{ type: string; text: string; doc: string; info: string }> = []
  annotationStore.annotations.slice(0, 3).forEach((a) => {
    items.push({ type: 'highlight', text: `"${a.selected_text?.substring(0, 40)}…"`, doc: `文档 #${a.document_id}`, info: `第${a.page}页` })
  })
  noteStore.notes.slice(0, 3).forEach((n) => {
    items.push({ type: 'note', text: n.title, doc: n.document_title || '—', info: '' })
  })
  return items.slice(0, 6)
})

onMounted(async () => {
  try {
    await Promise.all([
      documentStore.fetchDocuments(),
      noteStore.fetchNotes(),
      annotationStore.fetchAnnotations(0).catch(() => {}),
    ])
  } catch {}
  loading.value = false
})
</script>

<template>
  <AppLayout>
    <div class="topbar">
      <div class="topbar__left">
        <h1>{{ greeting }}好，{{ userName }}</h1>
        <p>今天也是阅读的好日子</p>
      </div>
      <div class="topbar__right">
        <div class="topbar__search">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input type="text" placeholder="搜索文件..." />
        </div>
        <button class="topbar__btn" @click="router.push('/documents')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          上传
        </button>
      </div>
    </div>

    <!-- KPI Row -->
    <div class="kpi-row">
      <div v-for="stat in weeklyStats" :key="stat.label" class="kpi-card">
        <div class="kpi-icon">
          <svg v-if="stat.icon === 'book'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
          <svg v-else-if="stat.icon === 'bookmark'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"/></svg>
          <svg v-else-if="stat.icon === 'edit'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
        </div>
        <div class="kpi-value">{{ stat.value }}</div>
        <div class="kpi-label">{{ stat.label }}</div>
      </div>
    </div>

    <div class="dashboard-grid">
      <!-- Reading List -->
      <div class="section-card">
        <div class="section-header"><h3>最近阅读</h3><a href="/documents" @click.prevent="router.push('/documents')">全部</a></div>
        <div v-if="readingList.length === 0" class="empty-hint">暂无阅读记录</div>
        <div v-else v-for="book in readingList" :key="book.id" class="reading-item" @click="router.push(`/reader/${book.id}`)">
          <div class="reading-item__icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg></div>
          <div class="reading-item__info">
            <div class="reading-item__title">{{ book.title }}</div>
            <div class="reading-item__meta">{{ book.meta }}</div>
          </div>
          <div class="reading-item__progress">{{ book.progress }}%</div>
        </div>
      </div>

      <!-- Recent Activity -->
      <div class="section-card">
        <div class="section-header"><h3>最近动态</h3><span class="text-muted">批注 & 笔记</span></div>
        <div v-if="recentActivity.length === 0" class="empty-hint">暂无活动</div>
        <div v-else v-for="act in recentActivity" :key="act.text" class="activity-item">
          <div :class="['activity-icon', act.type === 'highlight' ? 'icon-highlight' : 'icon-note']">
            <svg v-if="act.type === 'highlight'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"/></svg>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
          </div>
          <div class="activity-text">{{ act.text }}</div>
          <div class="activity-meta">{{ act.doc }} {{ act.info }}</div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.topbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.5rem; }
.topbar__left h1 { font-size: 1.35rem; font-weight: 600; }
.topbar__left p { font-size: 0.85rem; color: var(--text-secondary); margin-top: 0.15rem; }
.topbar__right { display: flex; gap: 0.75rem; align-items: center; }
.topbar__search { position: relative; }
.topbar__search input { font-size: 0.85rem; padding: 0.5rem 0.8rem 0.5rem 2.2rem; border: 1px solid var(--border-color); border-radius: var(--radius); background: var(--bg-card); color: var(--text-primary); outline: none; width: 240px; }
.topbar__search input:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(37,99,235,0.1); }
.topbar__search svg { position: absolute; left: 0.7rem; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.topbar__btn { height: 36px; padding: 0 1rem; border: 1px solid var(--border-color); border-radius: var(--radius); background: var(--bg-card); color: var(--text-secondary); font-family: var(--font-sans); font-size: 0.85rem; cursor: pointer; display: flex; align-items: center; gap: 0.4rem; transition: all 0.12s; }
.topbar__btn:hover { border-color: var(--accent); color: var(--accent); }
.topbar__btn svg { width: 16px; height: 16px; }
.kpi-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; margin-bottom: 1.5rem; }
.kpi-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-lg); padding: 1.2rem; transition: all 0.12s; }
.kpi-card:hover { box-shadow: var(--shadow-md); border-color: var(--accent); }
.kpi-icon { width: 36px; height: 36px; border-radius: var(--radius); background: var(--accent-light); display: flex; align-items: center; justify-content: center; color: var(--accent); margin-bottom: 0.75rem; }
.kpi-icon svg { width: 18px; height: 18px; }
.kpi-value { font-size: 1.5rem; font-weight: 700; }
.kpi-label { font-size: 0.78rem; color: var(--text-secondary); margin-top: 0.15rem; }
.dashboard-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.section-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-lg); padding: 1.2rem; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
.section-header h3 { font-size: 0.95rem; font-weight: 600; }
.section-header a { font-size: 0.8rem; }
.text-muted { font-size: 0.78rem; color: var(--text-muted); }
.empty-hint { text-align: center; padding: 2rem 0; color: var(--text-muted); font-size: 0.85rem; }
.reading-item, .activity-item { display: flex; align-items: center; gap: 0.75rem; padding: 0.6rem 0; border-bottom: 1px solid var(--border-color); cursor: pointer; }
.reading-item:last-child, .activity-item:last-child { border-bottom: none; }
.reading-item:hover { background: var(--accent-light); margin: 0 -0.75rem; padding: 0.6rem 0.75rem; border-radius: var(--radius); }
.reading-item__icon { width: 32px; height: 32px; border-radius: var(--radius); background: var(--accent-light); display: flex; align-items: center; justify-content: center; color: var(--accent); flex-shrink: 0; }
.reading-item__icon svg { width: 16px; height: 16px; }
.reading-item__info { flex: 1; }
.reading-item__title { font-size: 0.85rem; font-weight: 500; }
.reading-item__meta { font-size: 0.7rem; color: var(--text-muted); }
.reading-item__progress { font-size: 0.8rem; font-weight: 600; color: var(--accent); }
.activity-icon { width: 28px; height: 28px; border-radius: 50%; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.activity-icon svg { width: 14px; height: 14px; }
.icon-highlight { background: #FEF3C7; color: #D97706; }
.icon-note { background: #DBEAFE; color: #2563EB; }
.activity-text { flex: 1; font-size: 0.8rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.activity-meta { font-size: 0.7rem; color: var(--text-muted); white-space: nowrap; }
</style>
