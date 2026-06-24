<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '../components/layout/AppLayout.vue'
import { useDocumentStore } from '../stores/documentStore'

const router = useRouter()
const documentStore = useDocumentStore()
const searchQuery = ref('')
const activeFilter = ref('全部')
const fileInput = ref<HTMLInputElement | null>(null)
const activeMenuId = ref<number | null>(null)
const menuRef = ref<HTMLElement | null>(null)

// Rename modal state
const renameModalVisible = ref(false)
const renameTarget = ref<{ id: number; title: string } | null>(null)
const renameInput = ref('')
const renameInputRef = ref<HTMLInputElement | null>(null)
const renameLoading = ref(false)
const deleteModalVisible = ref(false)
const deleteTarget = ref<{ id: number; title: string } | null>(null)
const deleteLoading = ref(false)
const deleteError = ref('')
const deleteOverlayRef = ref<HTMLElement | null>(null)
const deleteRequestTimeout = 15000

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

function toggleMenu(id: number) {
  activeMenuId.value = activeMenuId.value === id ? null : id
}

async function handleRename(book: { id: number; title: string }) {
  activeMenuId.value = null
  renameTarget.value = book
  renameInput.value = book.title
  renameModalVisible.value = true
  renameLoading.value = false
  await nextTick()
  renameInputRef.value?.focus()
  renameInputRef.value?.select()
}

async function confirmRename() {
  if (!renameTarget.value) return
  const newTitle = renameInput.value.trim()
  if (!newTitle || newTitle === renameTarget.value.title) {
    closeRenameModal()
    return
  }
  renameLoading.value = true
  try {
    await documentStore.rename(renameTarget.value.id, newTitle)
    closeRenameModal()
  } catch (e: any) {
    console.warn('重命名失败:', e?.message || e)
    renameLoading.value = false
  }
}

function closeRenameModal() {
  renameModalVisible.value = false
  renameTarget.value = null
  renameInput.value = ''
  renameLoading.value = false
}

function onRenameKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') confirmRename()
  else if (e.key === 'Escape') closeRenameModal()
}

function handleDelete(book: { id: number; title: string }) {
  activeMenuId.value = null
  deleteTarget.value = book
  deleteModalVisible.value = true
  deleteLoading.value = false
  deleteError.value = ''
  nextTick(() => deleteOverlayRef.value?.focus())
}

async function confirmDelete() {
  if (!deleteTarget.value || deleteLoading.value) return
  const controller = new AbortController()
  const timeoutId = window.setTimeout(() => controller.abort(), deleteRequestTimeout)

  deleteLoading.value = true
  deleteError.value = ''
  try {
    await documentStore.remove(deleteTarget.value.id, controller.signal)
    forceCloseDeleteModal()
  } catch (e: any) {
    console.warn('删除失败:', e?.message || e)
    if (controller.signal.aborted) {
      deleteError.value = '删除请求超时，请检查后端服务或数据库连接后重试。'
    } else {
      deleteError.value = e?.response?.data?.detail || e?.message || '删除失败，请稍后重试。'
    }
    deleteLoading.value = false
  } finally {
    window.clearTimeout(timeoutId)
  }
}

function closeDeleteModal() {
  if (deleteLoading.value) return
  forceCloseDeleteModal()
}

function forceCloseDeleteModal() {
  deleteModalVisible.value = false
  deleteTarget.value = null
  deleteLoading.value = false
  deleteError.value = ''
}

function onDeleteKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    e.preventDefault()
    confirmDelete()
  }
  else if (e.key === 'Escape') closeDeleteModal()
}

function handleClickOutside(e: MouseEvent) {
  if (activeMenuId.value === null) return
  // menuRef inside v-for is a VNode array; check each element
  const menus = Array.isArray(menuRef.value) ? menuRef.value : (menuRef.value ? [menuRef.value] : [])
  const inside = menus.some((el: Element) => el.contains(e.target as Node))
  if (!inside) {
    activeMenuId.value = null
  }
}

onMounted(() => {
  documentStore.fetchDocuments()
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
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
              <div class="menu-wrapper" ref="menuRef">
                <button class="book-card__action" @click.stop="toggleMenu(book.id)">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="15" height="15"><circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/><circle cx="5" cy="12" r="1"/></svg>
                </button>
                <div v-if="activeMenuId === book.id" class="dropdown-menu" @click.stop>
                  <button class="dropdown-item" @click="handleRename(book)">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="14" height="14"><path d="M17 3a2.85 2.85 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"/></svg>
                    重命名
                  </button>
                  <button class="dropdown-item dropdown-item--danger" @click="handleDelete(book)">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="14" height="14"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                    删除
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>

  <!-- Rename Modal -->
  <Teleport to="body">
    <div
      v-if="renameModalVisible"
      class="rename-overlay"
      @click="closeRenameModal"
      @keydown="onRenameKeydown"
      tabindex="-1"
    >
      <div class="rename-modal" @click.stop>
        <div class="rename-modal__header">
          <h3>重命名</h3>
          <button class="rename-modal__close" @click="closeRenameModal">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <div class="rename-modal__body">
          <input
            ref="renameInputRef"
            v-model="renameInput"
            type="text"
            class="rename-modal__input"
            placeholder="请输入文件名"
            @keydown="onRenameKeydown"
          />
        </div>
        <div class="rename-modal__footer">
          <button class="rename-modal__btn rename-modal__btn--cancel" @click="closeRenameModal">取消</button>
          <button
            class="rename-modal__btn rename-modal__btn--confirm"
            :disabled="renameLoading || !renameInput.trim()"
            @click="confirmRename"
          >
            {{ renameLoading ? '保存中...' : '确认' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- Delete Modal -->
  <Teleport to="body">
    <div
      v-if="deleteModalVisible"
      ref="deleteOverlayRef"
      class="document-modal-overlay"
      tabindex="-1"
      @click="closeDeleteModal"
      @keydown="onDeleteKeydown"
    >
      <div class="document-modal document-modal--danger" role="dialog" aria-modal="true" aria-labelledby="delete-title" @click.stop>
        <div class="document-modal__header">
          <div class="document-modal__heading">
            <span class="document-modal__icon document-modal__icon--danger" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" width="18" height="18"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
            </span>
            <div>
              <h3 id="delete-title">删除文档</h3>
              <p>此操作不可撤销，删除后将无法从文件库恢复。</p>
            </div>
          </div>
          <button class="document-modal__close" :disabled="deleteLoading" aria-label="关闭" @click="closeDeleteModal">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <div class="document-modal__body">
          <div class="delete-summary">
            <span class="delete-summary__label">即将删除</span>
            <strong>{{ deleteTarget?.title || '未命名文档' }}</strong>
          </div>
          <p v-if="deleteError" class="document-modal__error" role="alert">{{ deleteError }}</p>
        </div>
        <div class="document-modal__footer">
          <button class="document-modal__btn document-modal__btn--cancel" :disabled="deleteLoading" @click="closeDeleteModal">取消</button>
          <button class="document-modal__btn document-modal__btn--danger" :disabled="deleteLoading" @click="confirmDelete">
            {{ deleteLoading ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
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

.menu-wrapper { position: relative; }
.dropdown-menu {
  position: absolute;
  right: 0;
  bottom: calc(100% + 4px);
  min-width: 130px;
  background: #fff;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.10);
  z-index: 10;
  overflow: hidden;
}
.dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: none;
  background: transparent;
  font-family: var(--font-ui, sans-serif);
  font-size: 0.8rem;
  color: var(--text-secondary);
  cursor: pointer;
  transition: background 0.1s;
  text-align: left;
}
.dropdown-item:hover { background: var(--accent-light); color: var(--text-primary); }
.dropdown-item--danger:hover { background: #FEE2E2; color: #B91C1C; }

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

/* Rename Modal — matches project aesthetic */
.rename-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.25);
  backdrop-filter: blur(2px);
  animation: fadeIn 0.15s ease;
}
.rename-modal {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #e5e2dc);
  border-radius: var(--radius-lg, 12px);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  width: 360px;
  max-width: 90vw;
  overflow: hidden;
  animation: scaleIn 0.15s ease;
}
.rename-modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem 0;
}
.rename-modal__header h3 {
  font-family: var(--font-display, serif);
  font-size: 1rem;
  font-weight: 500;
  color: var(--text-primary, #2c2c2c);
  margin: 0;
}
.rename-modal__close {
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 6px;
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--text-muted, #999);
  transition: all 0.1s;
}
.rename-modal__close:hover {
  background: var(--accent-light, #f0eee8);
  color: var(--text-primary, #2c2c2c);
}
.rename-modal__body {
  padding: 1rem 1.25rem;
}
.rename-modal__input {
  width: 100%;
  padding: 0.6rem 0.75rem;
  border: 1px solid var(--border-color, #e5e2dc);
  border-radius: var(--radius, 8px);
  background: var(--bg-page, #faf8f5);
  font-family: var(--font-ui, sans-serif);
  font-size: 0.85rem;
  color: var(--text-primary, #2c2c2c);
  outline: none;
  transition: border-color 0.15s;
  box-sizing: border-box;
}
.rename-modal__input:focus {
  border-color: var(--accent, #2563eb);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}
.rename-modal__input::placeholder {
  color: var(--text-muted, #999);
}
.rename-modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0 1.25rem 1rem;
}
.rename-modal__btn {
  padding: 0.4rem 1rem;
  border-radius: var(--radius, 8px);
  font-family: var(--font-ui, sans-serif);
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.12s;
  border: 1px solid var(--border-color, #e5e2dc);
}
.rename-modal__btn--cancel {
  background: transparent;
  color: var(--text-secondary, #666);
}
.rename-modal__btn--cancel:hover {
  background: var(--accent-light, #f0eee8);
  color: var(--text-primary, #2c2c2c);
}
.rename-modal__btn--confirm {
  background: var(--accent, #2563eb);
  color: #fff;
  border-color: var(--accent, #2563eb);
}
.rename-modal__btn--confirm:hover:not(:disabled) {
  opacity: 0.9;
}
.rename-modal__btn--confirm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to   { opacity: 1; }
}
@keyframes scaleIn {
  from { opacity: 0; transform: scale(0.95); }
  to   { opacity: 1; transform: scale(1); }
}

.document-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: rgba(61, 46, 36, 0.34);
  backdrop-filter: blur(3px);
  animation: fadeIn 0.15s ease;
}

.document-modal {
  width: min(420px, 100%);
  overflow: hidden;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  background: rgba(250, 248, 245, 0.98);
  box-shadow: 0 16px 42px rgba(61, 46, 36, 0.18);
  color: var(--text-primary);
  font-family: var(--font-ui);
  animation: scaleIn 0.15s ease;
}

.document-modal__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.05rem 1.15rem 0.85rem;
}

.document-modal__heading {
  display: flex;
  gap: 0.75rem;
  min-width: 0;
}

.document-modal__heading h3 {
  margin: 0;
  color: var(--text-primary);
  font-family: var(--font-display);
  font-size: 1.08rem;
  font-weight: 600;
  line-height: 1.25;
}

.document-modal__heading p {
  margin: 0.3rem 0 0;
  color: var(--text-secondary);
  font-size: 0.8rem;
  line-height: 1.55;
}

.document-modal__icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
}

.document-modal__icon--danger {
  background: #fee2e2;
  color: #b42318;
}

.document-modal__close {
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 7px;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background 0.12s, color 0.12s;
}

.document-modal__close:hover:not(:disabled) {
  background: var(--accent-light);
  color: var(--text-primary);
}

.document-modal__close:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.document-modal__body {
  padding: 0 1.15rem 1rem;
}

.delete-summary {
  display: grid;
  gap: 0.32rem;
  padding: 0.75rem 0.85rem;
  border: 1px solid rgba(180, 35, 24, 0.16);
  border-radius: 8px;
  background: rgba(254, 242, 242, 0.74);
}

.delete-summary__label {
  color: #b42318;
  font-size: 0.72rem;
  font-weight: 700;
}

.delete-summary strong {
  color: var(--text-primary);
  font-family: var(--font-display);
  font-size: 0.98rem;
  line-height: 1.35;
  overflow-wrap: anywhere;
}

.document-modal__error {
  margin: 0.75rem 0 0;
  padding: 0.62rem 0.75rem;
  border: 1px solid rgba(180, 35, 24, 0.18);
  border-radius: 8px;
  background: rgba(255, 247, 237, 0.86);
  color: #9f1f14;
  font-size: 0.78rem;
  line-height: 1.5;
}

.document-modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0 1.15rem 1.05rem;
}

.document-modal__btn {
  min-height: 34px;
  padding: 0 0.95rem;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  font-family: var(--font-ui);
  font-size: 0.8rem;
  transition: background 0.12s, border-color 0.12s, color 0.12s, opacity 0.12s;
}

.document-modal__btn:hover:not(:disabled) {
  background: var(--accent-light);
  color: var(--text-primary);
}

.document-modal__btn--danger {
  border-color: #b42318;
  background: #b42318;
  color: #fff;
}

.document-modal__btn--danger:hover:not(:disabled) {
  border-color: #9f1f14;
  background: #9f1f14;
  color: #fff;
}

.document-modal__btn:disabled {
  cursor: not-allowed;
  opacity: 0.58;
}
</style>
