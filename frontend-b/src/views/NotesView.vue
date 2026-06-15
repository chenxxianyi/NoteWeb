<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import AppLayout from '../components/layout/AppLayout.vue'
import { useNoteStore } from '../stores/noteStore'

const noteStore = useNoteStore()
const activeNoteId = ref<number | null>(null)
const searchQuery = ref('')
const activeTag = ref('全部')
const tags = ['全部', '阅读笔记', '技术', '随笔']
const editting = ref({ title: '', content: '', tags: [] as string[] })

const filteredNotes = computed(() => {
  let list = noteStore.notes
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter((n) => n.title.toLowerCase().includes(q))
  }
  return list
})

const currentNote = computed(() => {
  if (!activeNoteId.value) return filteredNotes.value[0] || null
  return noteStore.notes.find((n) => n.id === activeNoteId.value) || null
})

function selectNote(id: number) {
  activeNoteId.value = id
  noteStore.fetchNote(id)
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 60) return `${mins} 分钟前`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days} 天前`
  return dateStr?.substring(0, 10) || ''
}

async function createNote() {
  try {
    await noteStore.create({ title: '新笔记', content: '', tags: ['阅读笔记'] })
    activeNoteId.value = noteStore.notes[0]?.id || null
  } catch {}
}

onMounted(async () => {
  await noteStore.fetchNotes()
  if (noteStore.notes.length > 0) activeNoteId.value = noteStore.notes[0].id
})
</script>

<template>
  <AppLayout>
    <div class="notes-page">
      <div class="note-list">
        <div class="nl-header">
          <h2>笔记</h2>
          <button class="nl-header__btn" title="新建笔记" @click="createNote">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          </button>
        </div>
        <div class="nl-search"><input v-model="searchQuery" type="text" placeholder="搜索笔记..." /></div>
        <div class="nl-tags">
          <button v-for="tag in tags" :key="tag" :class="['nl-tag', { active: activeTag === tag }]" @click="activeTag = tag">{{ tag }}</button>
        </div>
        <div v-if="noteStore.loading" class="nl-loading">加载中...</div>
        <div v-else-if="filteredNotes.length === 0" class="nl-empty">暂无笔记</div>
        <div v-else class="nl-items">
          <div v-for="note in filteredNotes" :key="note.id" :class="['nl-item', { active: activeNoteId === note.id }]" @click="selectNote(note.id)">
            <div class="nl-item__title">{{ note.title }}</div>
            <div class="nl-item__preview">{{ note.content?.substring(0, 60) || '...' }}</div>
            <div class="nl-item__meta">
              <span class="nl-item__tag tag-reading">{{ note.tags?.[0] || '笔记' }}</span>
              {{ note.document_title || '' }} · {{ formatDate(note.updated_at) }}
            </div>
          </div>
        </div>
      </div>
      <div v-if="currentNote" class="note-editor">
        <div class="ne-toolbar">
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M6 4h8a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z"/><path d="M6 12h9a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z"/></svg></button>
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="17" y1="10" x2="3" y2="10"/><line x1="21" y1="6" x2="3" y2="6"/><line x1="21" y1="14" x2="3" y2="14"/><line x1="17" y1="18" x2="3" y2="18"/></svg></button>
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M6 4h8a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z"/><path d="M6 12h9a4 4 0 0 1 4 4 4 4 0 0 1-4 4H6z"/></svg></button>
          <div class="ne-divider"></div>
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg></button>
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 6H6a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-4"/><path d="M14 2h6v6"/><path d="M10 14L20 4"/></svg></button>
          <div class="ne-divider"></div>
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg></button>
          <button class="ne-btn"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg></button>
        </div>
        <input class="ne-title" type="text" :value="currentNote.title" placeholder="笔记标题..." />
        <textarea class="ne-body" :value="currentNote.content" placeholder="开始书写..."></textarea>
        <div class="ne-footer">
          <span>最后编辑：{{ formatDate(currentNote.updated_at) }}{{ currentNote.document_title ? ` · 来自「${currentNote.document_title}」` : '' }}</span>
          <div class="ne-footer__actions">
            <button class="ne-footer__btn">保存</button>
          </div>
        </div>
      </div>
      <div v-else class="note-editor ne-empty"><p>选择一篇笔记开始编辑</p></div>
    </div>
  </AppLayout>
</template>

<style scoped>
.notes-page { display: flex; height: calc(100vh - 3rem); overflow: hidden; background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-lg); }
.note-list { width: 340px; background: var(--bg-card); border-right: 1px solid var(--border-color); display: flex; flex-direction: column; flex-shrink: 0; }
.nl-header { padding: 1rem 1.2rem; border-bottom: 1px solid var(--border-color); display: flex; align-items: center; justify-content: space-between; }
.nl-header h2 { font-size: 1.1rem; font-weight: 600; }
.nl-header__btn { width: 32px; height: 32px; border: 1px solid var(--border-color); border-radius: 50%; background: transparent; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--accent); }
.nl-header__btn:hover { background: var(--accent-light); }
.nl-search { padding: 0.7rem 1rem; border-bottom: 1px solid var(--border-color); }
.nl-search input { width: 100%; padding: 0.4rem 0.7rem; border: 1px solid var(--border-color); border-radius: 6px; background: var(--bg-page); font-family: var(--font-sans); font-size: 0.78rem; color: var(--text-primary); outline: none; }
.nl-search input:focus { border-color: var(--accent); }
.nl-tags { padding: 0.5rem 1rem; border-bottom: 1px solid var(--border-color); display: flex; gap: 0.3rem; flex-wrap: wrap; }
.nl-tag { padding: 0.15rem 0.5rem; border: 1px solid var(--border-color); border-radius: 4px; background: transparent; font-family: var(--font-sans); font-size: 0.65rem; color: var(--text-secondary); cursor: pointer; }
.nl-tag:hover { border-color: var(--accent); color: var(--accent); }
.nl-tag.active { background: var(--accent); color: #fff; border-color: var(--accent); }
.nl-loading, .nl-empty { padding: 2rem; text-align: center; color: var(--text-muted); font-size: 0.85rem; }
.nl-items { flex: 1; overflow-y: auto; }
.nl-item { padding: 0.8rem 1.2rem; border-bottom: 1px solid var(--border-color); cursor: pointer; }
.nl-item:hover { background: var(--accent-light); }
.nl-item.active { background: var(--accent-light); border-left: 3px solid var(--accent); }
.nl-item__title { font-size: 0.88rem; font-weight: 500; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.nl-item__preview { font-size: 0.78rem; color: var(--text-secondary); line-height: 1.4; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; margin-top: 0.15rem; }
.nl-item__meta { font-family: var(--font-sans); font-size: 0.65rem; color: var(--text-muted); margin-top: 0.25rem; display: flex; gap: 0.3rem; align-items: center; }
.nl-item__tag { padding: 0.05rem 0.3rem; border-radius: 3px; font-size: 0.6rem; font-weight: 600; }
.tag-reading { background: #FDE68A; color: #92400E; }
.note-editor { flex: 1; display: flex; flex-direction: column; background: var(--bg-page); }
.ne-empty { justify-content: center; align-items: center; color: var(--text-muted); }
.ne-toolbar { padding: 0.6rem 1.5rem; border-bottom: 1px solid var(--border-color); display: flex; align-items: center; gap: 0.3rem; flex-wrap: wrap; }
.ne-btn { width: 30px; height: 30px; border: none; border-radius: 4px; background: transparent; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-secondary); }
.ne-btn:hover { background: var(--accent-light); color: var(--accent); }
.ne-divider { width: 1px; height: 20px; background: var(--border-color); margin: 0 0.2rem; }
.ne-title { width: 100%; padding: 0.8rem 1.5rem; border: none; font-size: 1rem; font-weight: 600; font-family: var(--font-sans); color: var(--text-primary); background: transparent; outline: none; }
.ne-body { flex: 1; width: 100%; padding: 0.5rem 1.5rem 1.5rem; border: none; font-size: 0.9rem; line-height: 1.7; font-family: var(--font-sans); color: var(--text-primary); background: transparent; outline: none; resize: none; }
.ne-footer { padding: 0.6rem 1.5rem; border-top: 1px solid var(--border-color); display: flex; justify-content: space-between; align-items: center; font-size: 0.75rem; color: var(--text-muted); }
.ne-footer__btn { padding: 0.3rem 0.8rem; border: 1px solid var(--border-color); border-radius: var(--radius); background: var(--accent); color: #fff; font-family: var(--font-sans); font-size: 0.78rem; cursor: pointer; }
.ne-footer__btn:hover { background: var(--accent-hover); }
</style>
