<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '../components/layout/AppLayout.vue'
import { useDocumentStore } from '../stores/documentStore'

const router = useRouter()
const documentStore = useDocumentStore()
const searchQuery = ref('')
const activeFilter = ref('全部')
const fileInput = ref<HTMLInputElement | null>(null)

const filters = ['全部', 'PDF', 'Markdown', 'DOCX', 'TXT']

const typeMap: Record<string, string> = { pdf: 'PDF', md: 'MD', docx: 'DOCX', txt: 'TXT' }
const typeClassMap: Record<string, string> = { pdf: 'type-pdf', md: 'type-md', docx: 'type-docx', txt: 'type-txt' }

function triggerUpload() { fileInput.value?.click() }

async function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try { await documentStore.upload(file) } catch (e: any) { console.warn('上传失败:', e?.message || e) }
  input.value = ''
}

const filteredBooks = computed(() => {
  let list = documentStore.documents
  if (activeFilter.value !== '全部') {
    const ext = activeFilter.value.toLowerCase()
    list = list.filter((d) => d.file_type === ext || (ext === 'markdown' && d.file_type === 'md'))
  }
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter((d) => d.title.toLowerCase().includes(q))
  }
  return list
})

function formatSize(bytes: number): string {
  if (bytes >= 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(0)} KB`
  return `${bytes} B`
}

function formatTime(dateStr: string): string {
  const d = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 60) return `${mins} 分钟前`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days} 天前`
  return dateStr?.substring(0, 10) || '—'
}

function filterByType(tag: string) { activeFilter.value = tag }
function openReader(id: number) { router.push(`/reader/${id}`) }

onMounted(() => { documentStore.fetchDocuments() })
</script>

<template>
  <AppLayout>
    <div class="topbar">
      <div class="topbar__left"><h1>文件库</h1></div>
      <div class="topbar__right">
        <div class="topbar__search">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
          <input v-model="searchQuery" type="text" placeholder="搜索文件..." />
        </div>
        <button class="topbar__btn" @click="triggerUpload">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>上传
        </button>
      </div>
    </div>

    <input ref="fileInput" type="file" accept=".pdf,.md,.docx,.txt" style="display:none" @change="handleFileChange" />

    <div class="upload-zone" @click="triggerUpload">
      <div class="upload-zone__icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
      </div>
      <div class="upload-zone__text"><strong>拖拽文件到此处</strong> 或点击上传 · 支持 PDF / Markdown / DOCX / TXT</div>
      <button class="upload-zone__btn" @click.stop="triggerUpload">选择文件</button>
    </div>

    <div class="filter-bar">
      <button v-for="f in filters" :key="f" :class="['filter-tag', { active: activeFilter === f }]" @click="filterByType(f)">{{ f }}</button>
    </div>

    <div v-if="documentStore.loading" class="loading-state">加载中...</div>
    <div v-else-if="filteredBooks.length === 0" class="empty-state">暂无文件</div>
    <div v-else class="file-list">
      <div v-for="book in filteredBooks" :key="book.id" class="file-card" @click="openReader(book.id)">
        <div class="file-card__icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
        </div>
        <div class="file-card__info">
          <div class="file-card__title">{{ book.title }}</div>
          <div class="file-card__meta">
            <span :class="['file-card__type', typeClassMap[book.file_type] || 'type-txt']">{{ typeMap[book.file_type] || book.file_type.toUpperCase() }}</span>
            {{ book.read_progress > 0 ? `阅读至${book.read_progress}%` : '未读' }}
          </div>
          <div class="file-card__footer">
            <span>{{ formatSize(book.file_size) }} · {{ formatTime(book.updated_at) }}</span>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.topbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.5rem; }
.topbar__left h1 { font-size: 1.35rem; font-weight: 600; }
.topbar__right { display: flex; gap: 0.6rem; align-items: center; }
.topbar__search { position: relative; }
.topbar__search input { font-size: 0.85rem; padding: 0.5rem 0.8rem 0.5rem 2.2rem; border: 1px solid var(--border-color); border-radius: var(--radius); background: var(--bg-card); color: var(--text-primary); outline: none; width: 240px; }
.topbar__search input:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(37,99,235,0.1); }
.topbar__search svg { position: absolute; left: 0.7rem; top: 50%; transform: translateY(-50%); width: 16px; height: 16px; color: var(--text-muted); }
.topbar__btn { height: 36px; padding: 0 1rem; border: 1px solid var(--border-color); border-radius: var(--radius); background: var(--bg-card); color: var(--text-secondary); font-family: var(--font-sans); font-size: 0.85rem; cursor: pointer; display: flex; align-items: center; gap: 0.4rem; }
.topbar__btn:hover { border-color: var(--accent); color: var(--accent); }
.topbar__btn svg { width: 16px; height: 16px; }
.upload-zone { border: 1px dashed var(--border-color); border-radius: var(--radius-lg); padding: 0.8rem 1.25rem; margin-bottom: 1.25rem; cursor: pointer; display: flex; align-items: center; gap: 0.75rem; background: var(--bg-card); }
.upload-zone:hover { border-color: var(--accent); background: var(--accent-light); }
.upload-zone__icon { width: 36px; height: 36px; border-radius: 6px; background: var(--accent-light); display: flex; align-items: center; justify-content: center; color: var(--accent); flex-shrink: 0; }
.upload-zone__icon svg { width: 18px; height: 18px; }
.upload-zone__text { flex: 1; font-size: 0.85rem; color: var(--text-secondary); }
.upload-zone__text strong { color: var(--text-primary); }
.upload-zone__btn { padding: 0.35rem 0.9rem; border: 1px solid var(--border-color); border-radius: var(--radius); background: var(--bg-card); font-family: var(--font-sans); font-size: 0.8rem; color: var(--text-secondary); cursor: pointer; flex-shrink: 0; }
.upload-zone__btn:hover { border-color: var(--accent); color: var(--accent); }
.filter-bar { display: flex; gap: 0.3rem; margin-bottom: 1rem; flex-wrap: wrap; }
.filter-tag { padding: 0.25rem 0.75rem; border: 1px solid var(--border-color); border-radius: var(--radius); background: var(--bg-card); font-family: var(--font-sans); font-size: 0.8rem; color: var(--text-secondary); cursor: pointer; }
.filter-tag:hover { border-color: var(--accent); color: var(--accent); }
.filter-tag.active { background: var(--accent); color: #fff; border-color: var(--accent); }
.loading-state, .empty-state { text-align: center; padding: 3rem 0; color: var(--text-muted); font-size: 0.9rem; }
.file-list { display: flex; flex-direction: column; gap: 0.5rem; }
.file-card { display: flex; align-items: center; gap: 1rem; padding: 1rem; background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-lg); cursor: pointer; transition: all 0.12s; }
.file-card:hover { box-shadow: var(--shadow-md); border-color: var(--accent); }
.file-card__icon { width: 40px; height: 40px; border-radius: var(--radius); background: var(--accent-light); display: flex; align-items: center; justify-content: center; color: var(--accent); flex-shrink: 0; }
.file-card__icon svg { width: 20px; height: 20px; }
.file-card__info { flex: 1; min-width: 0; }
.file-card__title { font-size: 0.9rem; font-weight: 500; }
.file-card__meta { display: flex; gap: 0.5rem; align-items: center; font-size: 0.78rem; color: var(--text-secondary); margin-top: 0.15rem; }
.file-card__type { padding: 0.05rem 0.35rem; border-radius: 4px; font-size: 0.65rem; font-weight: 600; }
.type-pdf { background: #FEE2E2; color: #B91C1C; }
.type-md { background: #DBEAFE; color: #1E40AF; }
.type-docx { background: #E0E7FF; color: #4338CA; }
.type-txt { background: #F3E8FF; color: #7E22CE; }
.file-card__footer { font-size: 0.7rem; color: var(--text-muted); margin-top: 0.2rem; }
</style>
