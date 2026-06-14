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

const typeMap: Record<string, string> = {
  pdf: 'PDF',
  md: 'MD',
  docx: 'DOCX',
  txt: 'TXT',
}

const typeClassMap: Record<string, string> = {
  pdf: 'type-pdf',
  md: 'type-md',
  docx: 'type-docx',
  txt: 'type-txt',
}

function triggerUpload() {
  fileInput.value?.click()
}

async function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    await documentStore.upload(file)
  } catch (e: any) {
    console.warn('上传失败:', e?.message || e)
  }
  // 清空 input 以便重复选择同一文件
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

function filterByType(tag: string) {
  activeFilter.value = tag
}

function openReader(id: number) {
  router.push(`/reader/${id}`)
}

onMounted(() => {
  documentStore.fetchDocuments()
})
</script>

<template>
  <AppLayout>
    <div class="documents-page">
      <div class="main-inner">
        <!-- Topbar -->
        <div class="topbar">
          <div class="topbar__left"><h1>文件库</h1></div>
          <div class="topbar__right">
            <div class="topbar__search">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16" class="search-icon"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
              <input v-model="searchQuery" type="text" placeholder="搜索文件..." />
            </div>
            <button class="upload-btn" @click="triggerUpload">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
              上传
            </button>
          </div>
        </div>

        <!-- Hidden file input -->
        <input
          ref="fileInput"
          type="file"
          accept=".pdf,.md,.docx,.txt,application/pdf,text/markdown,application/vnd.openxmlformats-officedocument.wordprocessingml.document,text/plain"
          style="display:none"
          @change="handleFileChange"
        />

        <!-- Upload Zone -->
        <div class="upload-zone" @click="triggerUpload">
          <div class="upload-zone__icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
          </div>
          <div class="upload-zone__text">
            <div class="upload-zone__title">拖拽文件到此处上传</div>
            <div class="upload-zone__hint">支持 PDF / Markdown / DOCX / TXT</div>
          </div>
          <button class="upload-zone__btn" @click.stop="triggerUpload">选择文件</button>
        </div>

        <!-- Filter Tags -->
        <div class="filter-bar">
          <button
            v-for="f in filters"
            :key="f"
            :class="['filter-tag', { active: activeFilter === f }]"
            @click="filterByType(f)"
          >
            {{ f }}
          </button>
        </div>

        <!-- Loading -->
        <div v-if="documentStore.loading" class="loading-state">
          <p>加载中...</p>
        </div>

        <!-- Empty State -->
        <div v-else-if="filteredBooks.length === 0" class="empty-state">
          <p>暂无文件，拖拽或点击上方区域上传</p>
        </div>

        <!-- Book Shelf -->
        <div v-else class="shelf">
          <div
            v-for="book in filteredBooks"
            :key="book.id"
            class="book-card"
            @click="openReader(book.id)"
          >
            <div class="book-card__icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="22" height="22"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
            </div>
            <div class="book-card__title">{{ book.title }}</div>
            <div class="book-card__meta">
              <span :class="['book-card__type', typeClassMap[book.file_type] || 'type-txt']">{{ typeMap[book.file_type] || book.file_type.toUpperCase() }}</span>
              {{ book.read_progress > 0 ? `阅读至${book.read_progress}%` : '未读' }}
            </div>
            <div class="book-card__footer">
              <span>{{ formatSize(book.file_size) }} · {{ formatTime(book.updated_at) }}</span>
              <button class="book-card__action" @click.stop>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="15" height="15"><circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/><circle cx="5" cy="12" r="1"/></svg>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.documents-page { padding: 2rem; display: flex; justify-content: center; min-height: 100vh; }
.main-inner { width: 100%; max-width: 1200px; }

.loading-state, .empty-state { text-align: center; padding: 4rem 0; color: var(--text-muted); font-family: var(--font-ui); font-size: 0.9rem; }

.topbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.5rem; }
.topbar__left h1 { font-family: var(--font-display); font-size: 1.4rem; font-weight: 500; }
.topbar__right { display: flex; gap: 0.6rem; align-items: center; }
.topbar__search { position: relative; }
.search-icon { position: absolute; left: 0.7rem; top: 50%; transform: translateY(-50%); color: var(--text-muted); }
.topbar__search input { font-family: var(--font-ui); font-size: 0.85rem; padding: 0.5rem 0.8rem 0.5rem 2.2rem; border: 1px solid var(--border-color); border-radius: 20px; background: var(--bg-card); color: var(--text-primary); outline: none; width: 220px; transition: border-color 0.15s; }
.topbar__search input:focus { border-color: var(--accent); }
.upload-btn { display: flex; align-items: center; gap: 0.4rem; padding: 0.4rem 0.9rem 0.4rem 0.6rem; border: 1px solid var(--border-color); border-radius: 20px; background: var(--bg-card); font-family: var(--font-ui); font-size: 0.8rem; color: var(--text-secondary); cursor: pointer; transition: all 0.15s; }
.upload-btn:hover { border-color: var(--accent); color: var(--accent); background: var(--accent-light); }

.upload-zone { border: 1px dashed var(--border-color); border-radius: var(--radius-lg); padding: 1rem 1.5rem; margin-bottom: 1.5rem; cursor: pointer; display: flex; align-items: center; gap: 1rem; background: var(--bg-card); transition: all 0.15s; }
.upload-zone:hover { border-color: var(--accent); background: var(--accent-light); }
.upload-zone__icon { width: 40px; height: 40px; border-radius: 8px; background: var(--accent-light); display: flex; align-items: center; justify-content: center; color: var(--accent); flex-shrink: 0; }
.upload-zone__text { flex: 1; }
.upload-zone__title { font-size: 0.9rem; font-weight: 500; color: var(--text-primary); }
.upload-zone__hint { font-family: var(--font-ui); font-size: 0.75rem; color: var(--text-muted); margin-top: 0.1rem; }
.upload-zone__btn { padding: 0.35rem 1rem; border: 1px solid var(--border-color); border-radius: var(--radius); background: var(--bg-card); font-family: var(--font-ui); font-size: 0.8rem; color: var(--text-secondary); cursor: pointer; transition: all 0.12s; flex-shrink: 0; }
.upload-zone__btn:hover { border-color: var(--accent); color: var(--accent); }

.filter-bar { display: flex; gap: 0.4rem; margin-bottom: 1.2rem; flex-wrap: wrap; }
.filter-tag { padding: 0.3rem 0.8rem; border: 1px solid var(--border-color); border-radius: 16px; background: transparent; font-family: var(--font-ui); font-size: 0.78rem; color: var(--text-secondary); cursor: pointer; transition: all 0.12s; }
.filter-tag:hover { border-color: var(--accent); color: var(--accent); }
.filter-tag.active { background: var(--accent); color: #fff; border-color: var(--accent); }

.shelf { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 1rem; }
.book-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-lg); padding: 1.2rem 1rem 1rem; cursor: pointer; transition: all 0.12s; display: flex; flex-direction: column; gap: 0.5rem; }
.book-card:hover { border-color: var(--accent); }
.book-card__icon { width: 44px; height: 44px; border-radius: 8px; background: var(--accent-light); display: flex; align-items: center; justify-content: center; color: var(--accent); }
.book-card__title { font-size: 0.9rem; font-weight: 500; color: var(--text-primary); display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.book-card__meta { font-family: var(--font-ui); font-size: 0.7rem; color: var(--text-muted); display: flex; align-items: center; gap: 0.4rem; }
.book-card__type { padding: 0.05rem 0.3rem; border-radius: 4px; font-size: 0.6rem; font-weight: 600; }
.type-pdf { background: #FEE2E2; color: #B91C1C; }
.type-md  { background: #DBEAFE; color: #1E40AF; }
.type-docx{ background: #D1FAE5; color: #065F46; }
.type-txt { background: #F3E8FF; color: #6B21A8; }
.book-card__footer { display: flex; align-items: center; justify-content: space-between; font-family: var(--font-ui); font-size: 0.7rem; color: var(--text-muted); }
.book-card__action { width: 28px; height: 28px; border-radius: 6px; border: none; background: transparent; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-muted); transition: all 0.1s; }
.book-card__action:hover { background: var(--accent-light); color: var(--accent); }

@media (max-width: 820px) {
  .shelf { grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); }
  .documents-page { padding: 1.2rem; }
  .topbar__search input { width: 160px; }
}
@media (max-width: 520px) {
  .documents-page { padding: 1rem; }
  .topbar { flex-direction: column; align-items: stretch; }
  .topbar__right { flex-wrap: wrap; }
}
</style>
