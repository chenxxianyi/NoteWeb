<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import AppLayout from '../components/layout/AppLayout.vue'
import { useNoteStore } from '../stores/noteStore'

const noteStore = useNoteStore()

const activeNoteId = ref<number | null>(null)
const searchQuery = ref('')
const activeTag = ref('全部')
const draftTitle = ref('')
const draftContent = ref('')
const saving = ref(false)
const statusToast = ref<{ type: 'success' | 'error'; message: string } | null>(null)
let toastTimer: number | null = null

const tags = ['全部', '阅读笔记', '技术', '随笔']

const filteredNotes = computed(() => {
  let list = noteStore.notes

  if (activeTag.value !== '全部') {
    list = list.filter((note) => note.tags?.includes(activeTag.value))
  }

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase()
    list = list.filter((note) =>
      `${note.title} ${note.content || ''}`.toLowerCase().includes(q),
    )
  }

  return list
})

const currentNote = computed(() => {
  if (!activeNoteId.value) return filteredNotes.value[0] || null
  return noteStore.notes.find((note) => note.id === activeNoteId.value) || null
})

watch(currentNote, (note) => {
  draftTitle.value = note?.title || ''
  draftContent.value = note?.content || ''
}, { immediate: true })

function selectNote(id: number) {
  activeNoteId.value = id
  void noteStore.fetchNote(id)
}

async function createNote() {
  const note = await noteStore.create({
    title: '未命名笔记',
    content: '',
    tags: ['随笔'],
  })
  activeTag.value = '全部'
  activeNoteId.value = note.id
}

async function saveNote() {
  if (!currentNote.value || saving.value) return

  saving.value = true
  try {
    await noteStore.update(currentNote.value.id, {
      title: draftTitle.value.trim() || '未命名笔记',
      content: draftContent.value,
      tags: currentNote.value.tags || [],
    })
    showStatusToast('success', '笔记保存成功')
  } catch (e: any) {
    showStatusToast('error', e?.response?.data?.detail || e?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

function showStatusToast(type: 'success' | 'error', message: string) {
  statusToast.value = { type, message }
  if (toastTimer) window.clearTimeout(toastTimer)
  toastTimer = window.setTimeout(() => {
    statusToast.value = null
    toastTimer = null
  }, 2000)
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return ''

  const d = new Date(dateStr)
  const diff = Date.now() - d.getTime()
  const mins = Math.floor(diff / 60000)

  if (Number.isNaN(mins)) return ''
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins} 分钟前`

  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} 小时前`

  const days = Math.floor(hours / 24)
  if (days < 30) return `${days} 天前`

  return dateStr.substring(0, 10)
}

onMounted(async () => {
  await noteStore.fetchNotes()
  if (noteStore.notes.length > 0 && !activeNoteId.value) {
    activeNoteId.value = noteStore.notes[0].id
  }
})
</script>

<template>
  <AppLayout>
    <div class="notes-page">
      <div v-if="statusToast" :class="['note-toast', `note-toast--${statusToast.type}`]">
        {{ statusToast.message }}
      </div>

      <div class="note-list">
        <div class="nl-header">
          <h2>笔记</h2>
          <button class="nl-header__btn" title="新建笔记" @click="createNote">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          </button>
        </div>

        <div class="nl-search">
          <input v-model="searchQuery" type="text" placeholder="搜索笔记..." />
        </div>

        <div class="nl-tags">
          <button
            v-for="tag in tags"
            :key="tag"
            :class="['nl-tag', { active: activeTag === tag }]"
            @click="activeTag = tag"
          >
            {{ tag }}
          </button>
        </div>

        <div v-if="noteStore.loading" class="nl-loading">加载中...</div>
        <div v-else-if="filteredNotes.length === 0" class="nl-empty">暂无笔记</div>

        <div v-else class="nl-items">
          <div
            v-for="note in filteredNotes"
            :key="note.id"
            :class="['nl-item', { active: activeNoteId === note.id }]"
            @click="selectNote(note.id)"
          >
            <div class="nl-item__title">{{ note.title || '未命名笔记' }}</div>
            <div class="nl-item__preview">{{ note.content?.substring(0, 60) || '...' }}</div>
            <div class="nl-item__meta">
              <span class="nl-item__tag tag-reading">{{ note.tags?.[0] || '笔记' }}</span>
              <span>{{ note.document_title || '独立笔记' }} · {{ formatDate(note.updated_at) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="currentNote" class="note-editor">
        <div class="ne-toolbar">
          <button class="ne-btn" title="加粗"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M6 4h8a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z"/><path d="M6 12h9a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z"/></svg></button>
          <button class="ne-btn" title="列表"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg></button>
          <div class="ne-divider"></div>
          <button class="ne-btn" title="引用"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg></button>
        </div>

        <input v-model="draftTitle" class="ne-title" type="text" placeholder="笔记标题..." />
        <textarea v-model="draftContent" class="ne-body" placeholder="开始书写..."></textarea>

        <div class="ne-footer">
          <span>最后编辑：{{ formatDate(currentNote.updated_at) }}</span>
          <div class="ne-footer__actions">
            <button class="ne-footer__btn" @click="saveNote">
              {{ saving ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>

      <div v-else class="note-editor ne-empty">
        <p>选择一篇笔记开始编辑，或点击左上角新建笔记</p>
      </div>
    </div>
  </AppLayout>
</template>

<style scoped>
.notes-page { display: flex; height: 100vh; overflow: hidden; }
.note-toast { position: fixed; top: 4.55rem; left: 50%; z-index: 300; min-width: 180px; max-width: min(360px, calc(100vw - 2rem)); padding: 0.65rem 1rem; border: 1px solid var(--border-color); border-radius: 10px; background: var(--bg-card); box-shadow: 0 10px 24px rgba(61, 46, 36, 0.14); font-family: var(--font-ui); font-size: 0.85rem; text-align: center; transform: translateX(-50%); }
.note-toast--success { color: var(--accent); }
.note-toast--error { color: #b42318; }

.note-list { width: 340px; background: var(--bg-card); border-right: 1px solid var(--border-color); display: flex; flex-direction: column; flex-shrink: 0; }
.nl-header { padding: 1rem 1.2rem; border-bottom: 1px solid var(--border-color); display: flex; align-items: center; justify-content: space-between; }
.nl-header h2 { font-family: var(--font-display); font-size: 1.1rem; font-weight: 500; }
.nl-header__btn { width: 32px; height: 32px; border: 1px solid var(--border-color); border-radius: 50%; background: transparent; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--accent); transition: all 0.12s; }
.nl-header__btn:hover { background: var(--accent-light); }

.nl-search { padding: 0.7rem 1rem; border-bottom: 1px solid var(--border-color); }
.nl-search input { width: 100%; padding: 0.4rem 0.7rem; border: 1px solid var(--border-color); border-radius: 16px; background: var(--bg-page); font-family: var(--font-ui); font-size: 0.78rem; color: var(--text-primary); outline: none; }
.nl-search input:focus { border-color: var(--accent); }

.nl-tags { padding: 0.5rem 1rem; border-bottom: 1px solid var(--border-color); display: flex; gap: 0.3rem; flex-wrap: wrap; }
.nl-tag { padding: 0.15rem 0.5rem; border: 1px solid var(--border-color); border-radius: 10px; background: transparent; font-family: var(--font-ui); font-size: 0.65rem; color: var(--text-secondary); cursor: pointer; transition: all 0.1s; }
.nl-tag:hover { border-color: var(--accent); color: var(--accent); }
.nl-tag.active { background: var(--accent); color: #fff; border-color: var(--accent); }

.nl-loading, .nl-empty { padding: 2rem; text-align: center; color: var(--text-muted); font-family: var(--font-ui); font-size: 0.85rem; }

.nl-items { flex: 1; overflow-y: auto; }
.nl-item { padding: 0.8rem 1.2rem; border-bottom: 1px solid var(--border-color); cursor: pointer; transition: background 0.1s; }
.nl-item:hover { background: var(--accent-light); }
.nl-item.active { background: var(--accent-light); border-left: 3px solid var(--accent); }
.nl-item__title { font-size: 0.88rem; font-weight: 500; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.nl-item__preview { font-size: 0.78rem; color: var(--text-secondary); line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; margin-top: 0.15rem; }
.nl-item__meta { font-family: var(--font-ui); font-size: 0.65rem; color: var(--text-muted); margin-top: 0.25rem; display: flex; gap: 0.3rem; align-items: center; min-width: 0; }
.nl-item__tag { padding: 0.05rem 0.3rem; border-radius: 3px; font-size: 0.6rem; font-weight: 600; flex-shrink: 0; }
.tag-reading { background: #FDE68A; color: #92400E; }

.note-editor { flex: 1; display: flex; flex-direction: column; background: var(--bg-page); min-width: 0; }
.ne-empty { justify-content: center; align-items: center; color: var(--text-muted); font-family: var(--font-ui); }
.ne-toolbar { padding: 0.6rem 1.5rem; border-bottom: 1px solid var(--border-color); display: flex; align-items: center; gap: 0.3rem; flex-wrap: wrap; }
.ne-btn { width: 30px; height: 30px; border: none; border-radius: 4px; background: transparent; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-secondary); transition: all 0.1s; }
.ne-btn:hover { background: var(--accent-light); color: var(--accent); }
.ne-divider { width: 1px; height: 18px; background: var(--border-color); margin: 0 0.2rem; }
.ne-title { padding: 1.2rem 1.5rem 0.5rem; border: none; font-family: var(--font-display); font-size: 1.3rem; font-weight: 500; color: var(--text-primary); background: transparent; outline: none; width: 100%; }
.ne-title::placeholder { color: var(--text-muted); }
.ne-body { flex: 1; padding: 0.5rem 1.5rem 2rem; border: none; font-family: var(--font-body); font-size: 0.95rem; line-height: 1.8; color: var(--text-primary); background: transparent; outline: none; resize: none; width: 100%; }
.ne-body::placeholder { color: var(--text-muted); }
.ne-footer { padding: 0.6rem 1.5rem; border-top: 1px solid var(--border-color); display: flex; align-items: center; justify-content: space-between; font-family: var(--font-ui); font-size: 0.7rem; color: var(--text-muted); }
.ne-footer__actions { display: flex; gap: 0.5rem; }
.ne-footer__btn { padding: 0.25rem 0.7rem; border: 1px solid var(--border-color); border-radius: 14px; background: var(--bg-card); font-family: var(--font-ui); font-size: 0.7rem; color: var(--text-secondary); cursor: pointer; transition: all 0.1s; }
.ne-footer__btn:hover { border-color: var(--accent); color: var(--accent); }

@media (max-width: 820px) { .note-list { width: 280px; } }
@media (max-width: 620px) { .note-list { width: 100%; } .note-editor { display: none; } .note-toast { top: 3.75rem; } }
</style>
