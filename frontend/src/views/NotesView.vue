<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import AppLayout from '../components/layout/AppLayout.vue'
import { useNoteStore } from '../stores/noteStore'
import { EditorContent, useEditor } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Link from '@tiptap/extension-link'
import Placeholder from '@tiptap/extension-placeholder'
import TaskItem from '@tiptap/extension-task-item'
import TaskList from '@tiptap/extension-task-list'
import {
  Bold,
  CheckCircle2,
  ChevronRight,
  ClipboardPaste,
  Code2,
  Copy,
  Heading1,
  Heading2,
  Italic,
  Link as LinkIcon,
  List as ListIcon,
  ListOrdered,
  Minus,
  Pilcrow,
  Quote,
  Scissors,
  Trash2,
} from 'lucide-vue-next'
import type { Component } from 'vue'

type ContextMenuItem = {
  command?: string
  label: string
  icon?: Component
  disabled?: boolean
  hint?: string
}

type ContextSubmenu = {
  key: string
  label: string
  items: ContextMenuItem[]
}

const noteStore = useNoteStore()

const activeNoteId = ref<number | null>(null)
const searchQuery = ref('')
const activeTag = ref('全部')
const draftTitle = ref('')
const draftContent = ref('')
const saving = ref(false)
const statusToast = ref<{ type: 'success' | 'error'; message: string } | null>(null)
const contextMenu = ref({
  open: false,
  x: 0,
  y: 0,
  activeSubmenu: '',
})
let toastTimer: number | null = null

const tags = ['全部', '阅读笔记', '技术', '随笔']

const editorTools: Array<
  | { type: 'tool'; command: string; label: string; title: string }
  | { type: 'divider' }
> = [
  { type: 'tool', command: 'bold', label: 'B', title: '加粗 Ctrl+B' },
  { type: 'tool', command: 'italic', label: 'I', title: '斜体 Ctrl+I' },
  { type: 'tool', command: 'strike', label: 'S', title: '删除线' },
  { type: 'divider' },
  { type: 'tool', command: 'heading1', label: 'H1', title: '一级标题' },
  { type: 'tool', command: 'heading2', label: 'H2', title: '二级标题' },
  { type: 'divider' },
  { type: 'tool', command: 'unorderedList', label: '•', title: '无序列表' },
  { type: 'tool', command: 'orderedList', label: '1.', title: '有序列表' },
  { type: 'tool', command: 'taskList', label: '☑', title: '任务列表' },
  { type: 'tool', command: 'quote', label: '❝', title: '引用' },
  { type: 'divider' },
  { type: 'tool', command: 'inlineCode', label: '`', title: '行内代码' },
  { type: 'tool', command: 'codeBlock', label: '{}', title: '代码块' },
  { type: 'tool', command: 'link', label: '↗', title: '链接 Ctrl+K' },
]

const contextQuickActions = [
  { command: 'cut', label: '剪切', icon: Scissors, disabled: false },
  { command: 'copy', label: '复制', icon: Copy, disabled: false },
  { command: 'paste', label: '粘贴', icon: ClipboardPaste, disabled: false },
  { command: 'delete', label: '删除', icon: Trash2, disabled: false },
]

const contextFormatActions = [
  { command: 'bold', label: '加粗', icon: Bold },
  { command: 'italic', label: '斜体', icon: Italic },
  { command: 'inlineCode', label: '行内代码', icon: Code2 },
  { command: 'link', label: '插入链接', icon: LinkIcon },
  { command: 'quote', label: '引用', icon: Quote },
  { command: 'orderedList', label: '有序列表', icon: ListOrdered },
  { command: 'unorderedList', label: '无序列表', icon: ListIcon },
  { command: 'taskList', label: '任务列表', icon: CheckCircle2 },
]

const contextSubmenus: ContextSubmenu[] = [
  {
    key: 'paste',
    label: '复制 / 粘贴为...',
    items: [
      { label: '保留格式粘贴', disabled: true, hint: '浏览器限制右键菜单直接读取剪贴板' },
      { label: '粘贴为纯文本', disabled: true, hint: '请使用 Ctrl+Shift+V' },
      { label: '粘贴为代码', disabled: true, hint: '请先粘贴，再点代码按钮' },
    ],
  },
  {
    key: 'paragraph',
    label: '段落',
    items: [
      { command: 'paragraph', label: '正文', icon: Pilcrow },
      { command: 'heading1', label: '一级标题', icon: Heading1 },
      { command: 'heading2', label: '二级标题', icon: Heading2 },
      { command: 'quote', label: '引用', icon: Quote },
    ],
  },
  {
    key: 'insert',
    label: '插入',
    items: [
      { command: 'codeBlock', label: '代码块', icon: Code2 },
      { command: 'horizontalRule', label: '分割线', icon: Minus },
      { command: 'link', label: '链接', icon: LinkIcon },
    ],
  },
]

const editor = useEditor({
  content: '',
  extensions: [
    StarterKit.configure({
      link: false,
    }),
    Link.configure({
      openOnClick: true,
      autolink: true,
      linkOnPaste: true,
    }),
    Placeholder.configure({
      placeholder: '开始书写...',
    }),
    TaskList,
    TaskItem.configure({
      nested: true,
    }),
  ],
  editorProps: {
    attributes: {
      class: 'ne-rich-body',
    },
  },
  onUpdate: ({ editor }) => {
    draftContent.value = editor.getHTML()
  },
})

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
  if (editor.value && editor.value.getHTML() !== (note?.content || '')) {
    editor.value.commands.setContent(note?.content || '', { emitUpdate: false })
  }
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
    const content = editor.value?.getHTML() || draftContent.value
    await noteStore.update(currentNote.value.id, {
      title: draftTitle.value.trim() || '未命名笔记',
      content,
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

function openEditorContextMenu(event: MouseEvent) {
  event.preventDefault()
  editor.value?.commands.focus()

  const menuWidth = 248
  const menuHeight = 206
  const margin = 12
  const x = Math.min(event.clientX, window.innerWidth - menuWidth - margin)
  const y = Math.min(event.clientY, window.innerHeight - menuHeight - margin)

  contextMenu.value = {
    open: true,
    x: Math.max(margin, x),
    y: Math.max(margin, y),
    activeSubmenu: '',
  }
}

function closeContextMenu() {
  if (!contextMenu.value.open) return
  contextMenu.value.open = false
  contextMenu.value.activeSubmenu = ''
}

function handleDocumentPointerDown(event: PointerEvent) {
  const target = event.target as HTMLElement | null
  if (!contextMenu.value.open || target?.closest('.note-context-menu')) return
  closeContextMenu()
}

function handleDocumentKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') closeContextMenu()
}

async function executeContextAction(command?: string, disabled?: boolean) {
  if (!command || disabled) return

  if (command === 'paste') {
    try {
      const text = await navigator.clipboard.readText()
      if (!text) {
        showStatusToast('error', '请使用 Ctrl+V')
        closeContextMenu()
        return
      }

      editor.value?.chain().focus().insertContent(text).run()
      closeContextMenu()
      return
    } catch {
      showStatusToast('error', '请使用 Ctrl+V')
      closeContextMenu()
      return
    }
  }

  if (command === 'copy' || command === 'cut') {
    const ok = document.execCommand(command)
    showStatusToast(ok ? 'success' : 'error', ok ? (command === 'copy' ? '已复制' : '已剪切') : '浏览器阻止了剪贴板操作')
    closeContextMenu()
    return
  }

  if (command === 'delete') {
    editor.value?.chain().focus().deleteSelection().run()
    closeContextMenu()
    return
  }

  applyEditorCommand(command)
  closeContextMenu()
}

function applyEditorCommand(command: string) {
  const current = editor.value
  const chain = current?.chain().focus()
  if (!chain || !current) return

  const hadTextSelection = !current.state.selection.empty

  if (command === 'bold') chain.toggleBold().run()
  else if (command === 'italic') chain.toggleItalic().run()
  else if (command === 'strike') chain.toggleStrike().run()
  else if (command === 'paragraph') chain.setParagraph().run()
  else if (command === 'heading1') chain.toggleHeading({ level: 1 }).run()
  else if (command === 'heading2') chain.toggleHeading({ level: 2 }).run()
  else if (command === 'unorderedList') chain.toggleBulletList().run()
  else if (command === 'orderedList') chain.toggleOrderedList().run()
  else if (command === 'taskList') chain.toggleTaskList().run()
  else if (command === 'quote') chain.toggleBlockquote().run()
  else if (command === 'inlineCode') chain.toggleCode().run()
  else if (command === 'codeBlock') chain.toggleCodeBlock().run()
  else if (command === 'horizontalRule') chain.setHorizontalRule().run()
  else if (command === 'link') {
    const previousUrl = editor.value?.getAttributes('link').href || ''
    const url = window.prompt('输入链接地址', previousUrl || 'https://')
    if (url === null) return
    if (url.trim() === '') {
      editor.value?.chain().focus().extendMarkRange('link').unsetLink().run()
      return
    }
    editor.value?.chain().focus().extendMarkRange('link').setLink({ href: url.trim() }).run()
  }

  if (hadTextSelection && ['bold', 'italic', 'strike', 'inlineCode', 'link'].includes(command)) {
    moveAfterSelectedMark(command)
  }
}

function moveAfterSelectedMark(command: string) {
  const current = editor.value
  if (!current) return

  const to = current.state.selection.to
  const chain = current.chain().focus().setTextSelection(to)

  if (command === 'bold') chain.unsetBold().run()
  else if (command === 'italic') chain.unsetItalic().run()
  else if (command === 'strike') chain.unsetStrike().run()
  else if (command === 'inlineCode') chain.unsetCode().run()
  else if (command === 'link') chain.unsetLink().run()
}

function isToolActive(command: string): boolean {
  const current = editor.value
  if (!current) return false

  if (command === 'bold') return current.isActive('bold')
  if (command === 'italic') return current.isActive('italic')
  if (command === 'strike') return current.isActive('strike')
  if (command === 'heading1') return current.isActive('heading', { level: 1 })
  if (command === 'heading2') return current.isActive('heading', { level: 2 })
  if (command === 'unorderedList') return current.isActive('bulletList')
  if (command === 'orderedList') return current.isActive('orderedList')
  if (command === 'taskList') return current.isActive('taskList')
  if (command === 'quote') return current.isActive('blockquote')
  if (command === 'inlineCode') return current.isActive('code')
  if (command === 'codeBlock') return current.isActive('codeBlock')
  if (command === 'link') return current.isActive('link')
  return false
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
  document.addEventListener('pointerdown', handleDocumentPointerDown)
  document.addEventListener('keydown', handleDocumentKeydown)
  await noteStore.fetchNotes()
  if (noteStore.notes.length > 0 && !activeNoteId.value) {
    activeNoteId.value = noteStore.notes[0].id
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', handleDocumentPointerDown)
  document.removeEventListener('keydown', handleDocumentKeydown)
  if (toastTimer) window.clearTimeout(toastTimer)
  editor.value?.destroy()
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
          <template v-for="(tool, index) in editorTools" :key="tool.type === 'divider' ? `divider-${index}` : tool.command">
            <div v-if="tool.type === 'divider'" class="ne-divider"></div>
            <button
              v-else
              type="button"
              :class="['ne-btn', { active: isToolActive(tool.command) }]"
              :title="tool.title"
              :aria-label="tool.title"
              :aria-pressed="isToolActive(tool.command)"
              @click="applyEditorCommand(tool.command)"
            >
              {{ tool.label }}
            </button>
          </template>
        </div>

        <input v-model="draftTitle" class="ne-title" type="text" placeholder="笔记标题..." />
        <EditorContent
          v-if="editor"
          :editor="editor"
          class="ne-body"
          @contextmenu="openEditorContextMenu"
          @scroll="closeContextMenu"
        />

        <div
          v-if="contextMenu.open"
          class="note-context-menu"
          :style="{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }"
          role="menu"
          @contextmenu.prevent
        >
          <div class="ncm-icon-grid ncm-icon-grid--top">
            <button
              v-for="action in contextQuickActions"
              :key="action.command"
              type="button"
              class="ncm-icon-btn"
              :class="{ disabled: action.disabled }"
              :aria-label="action.label"
              :title="action.disabled ? '浏览器限制此操作，请使用快捷键' : action.label"
              :disabled="action.disabled"
              @click="executeContextAction(action.command, action.disabled)"
            >
              <component :is="action.icon" :size="14" :stroke-width="1.8" />
            </button>
          </div>

          <div
            class="ncm-row"
            role="menuitem"
            tabindex="0"
            @mouseenter="contextMenu.activeSubmenu = 'paste'"
            @focus="contextMenu.activeSubmenu = 'paste'"
          >
            <span>复制 / 粘贴为...</span>
            <ChevronRight :size="12" :stroke-width="1.8" />
          </div>

          <div class="ncm-separator"></div>

          <div class="ncm-icon-grid">
            <button
              v-for="action in contextFormatActions"
              :key="action.command"
              type="button"
              class="ncm-icon-btn"
              :class="{ active: isToolActive(action.command) }"
              :aria-label="action.label"
              :aria-pressed="isToolActive(action.command)"
              :title="action.label"
              @click="executeContextAction(action.command)"
            >
              <component :is="action.icon" :size="14" :stroke-width="1.9" />
            </button>
          </div>

          <div class="ncm-separator"></div>

          <div
            v-for="submenu in contextSubmenus.filter((item) => item.key !== 'paste')"
            :key="submenu.key"
            class="ncm-row"
            role="menuitem"
            tabindex="0"
            @mouseenter="contextMenu.activeSubmenu = submenu.key"
            @focus="contextMenu.activeSubmenu = submenu.key"
          >
            <span>{{ submenu.label }}</span>
            <ChevronRight :size="12" :stroke-width="1.8" />
          </div>

          <div
            v-for="submenu in contextSubmenus"
            v-show="contextMenu.activeSubmenu === submenu.key"
            :key="`flyout-${submenu.key}`"
            class="ncm-flyout"
            role="menu"
          >
            <button
              v-for="item in submenu.items"
              :key="item.label"
              type="button"
              class="ncm-flyout__item"
              :class="{ disabled: item.disabled }"
              :disabled="item.disabled"
              :title="item.hint || item.label"
              @click="executeContextAction(item.command, item.disabled)"
            >
              <component v-if="item.icon" :is="item.icon" :size="13" :stroke-width="1.8" />
              <span>{{ item.label }}</span>
              <small v-if="item.hint">{{ item.hint }}</small>
            </button>
          </div>
        </div>

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
.ne-btn { width: auto; min-width: 30px; height: 30px; padding: 0 0.45rem; border: none; border-radius: 4px; background: transparent; display: inline-flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-secondary); font-family: var(--font-ui); font-size: 0.78rem; font-weight: 600; transition: all 0.1s; }
.ne-btn:hover, .ne-btn.active { background: var(--accent-light); color: var(--accent); }
.ne-divider { width: 1px; height: 18px; background: var(--border-color); margin: 0 0.2rem; }
.ne-title { padding: 1.2rem 1.5rem 0.5rem; border: none; font-family: var(--font-display); font-size: 1.3rem; font-weight: 500; color: var(--text-primary); background: transparent; outline: none; width: 100%; }
.ne-title::placeholder { color: var(--text-muted); }
.ne-body { flex: 1; min-height: 0; overflow-y: auto; padding: 0.5rem 1.5rem 2rem; font-family: var(--font-body); font-size: 0.95rem; line-height: 1.8; color: var(--text-primary); }
.ne-body :deep(.ne-rich-body) { min-height: 100%; outline: none; }
.ne-body :deep(.is-editor-empty:first-child::before) { content: attr(data-placeholder); float: left; height: 0; color: var(--text-muted); pointer-events: none; }
.ne-body :deep(p) { margin: 0 0 0.8rem; }
.ne-body :deep(strong) { font-weight: 700; }
.ne-body :deep(em) { font-style: italic; }
.ne-body :deep(s) { text-decoration: line-through; }
.ne-body :deep(h1) { margin: 0.2rem 0 0.9rem; font-family: var(--font-display); font-size: 1.55rem; font-weight: 600; line-height: 1.35; }
.ne-body :deep(h2) { margin: 0.2rem 0 0.8rem; font-family: var(--font-display); font-size: 1.25rem; font-weight: 600; line-height: 1.4; }
.ne-body :deep(ul), .ne-body :deep(ol) { margin: 0.4rem 0 0.9rem 1.4rem; padding: 0; }
.ne-body :deep(li) { margin: 0.15rem 0; }
.ne-body :deep(ul[data-type="taskList"]) { list-style: none; margin-left: 0; }
.ne-body :deep(ul[data-type="taskList"] li) { display: flex; gap: 0.45rem; align-items: flex-start; }
.ne-body :deep(ul[data-type="taskList"] label) { flex: 0 0 auto; padding-top: 0.15rem; }
.ne-body :deep(blockquote) { margin: 0.6rem 0 0.9rem; padding-left: 0.8rem; border-left: 3px solid var(--accent); color: var(--text-secondary); }
.ne-body :deep(code) { padding: 0.1rem 0.3rem; border-radius: 4px; background: var(--accent-light); font-family: Consolas, 'Courier New', monospace; font-size: 0.88em; }
.ne-body :deep(pre) { margin: 0.6rem 0 1rem; padding: 0.85rem 1rem; border-radius: 8px; background: #2b2520; color: #f7efe7; overflow-x: auto; }
.ne-body :deep(pre code) { padding: 0; background: transparent; color: inherit; }
.ne-body :deep(a) { color: var(--accent); text-decoration: underline; text-underline-offset: 3px; }
.note-context-menu { position: fixed; z-index: 500; width: 248px; padding: 7px 0; border: 1px solid rgba(228, 217, 206, 0.9); border-radius: 8px; background: rgba(255, 255, 255, 0.96); box-shadow: 0 10px 22px rgba(61, 46, 36, 0.12); color: #3f3f3f; font-family: var(--font-ui); }
.ncm-icon-grid { display: grid; grid-template-columns: repeat(4, 42px); gap: 5px; justify-content: center; padding: 0 10px; }
.ncm-icon-grid--top { margin-bottom: 5px; }
.ncm-icon-btn { height: 26px; border: 1px solid #e7ebee; border-radius: 6px; background: #f8fafb; color: #3f3f3f; display: inline-flex; align-items: center; justify-content: center; cursor: pointer; transition: background 0.12s, border-color 0.12s, color 0.12s, transform 0.12s; }
.ncm-icon-btn:hover:not(:disabled), .ncm-icon-btn.active { border-color: var(--accent); background: var(--accent-light); color: var(--accent); }
.ncm-icon-btn:active:not(:disabled) { transform: translateY(1px); }
.ncm-icon-btn.disabled, .ncm-icon-btn:disabled { color: #b7bdc2; cursor: not-allowed; background: #fbfbfb; }
.ncm-row { position: relative; height: 30px; padding: 0 16px 0 24px; display: flex; align-items: center; justify-content: space-between; color: #4a4a4a; font-size: 13px; line-height: 1; cursor: default; outline: none; }
.ncm-row:hover, .ncm-row:focus { background: #f7f3ef; color: var(--accent); }
.ncm-separator { height: 1px; margin: 6px 0; background: #eef0f2; }
.ncm-flyout { position: absolute; top: 38px; left: calc(100% + 6px); width: 172px; padding: 5px; border: 1px solid rgba(228, 217, 206, 0.95); border-radius: 8px; background: rgba(255, 255, 255, 0.98); box-shadow: 0 10px 20px rgba(61, 46, 36, 0.12); }
.ncm-flyout__item { width: 100%; min-height: 26px; padding: 4px 5px; border: none; border-radius: 6px; background: transparent; color: var(--text-primary); display: grid; grid-template-columns: 14px 1fr; column-gap: 5px; align-items: center; text-align: left; font-family: var(--font-ui); font-size: 0.68rem; cursor: pointer; }
.ncm-flyout__item:hover:not(:disabled) { background: var(--accent-light); color: var(--accent); }
.ncm-flyout__item small { grid-column: 2; margin-top: 2px; color: var(--text-muted); font-size: 0.6rem; line-height: 1.2; }
.ncm-flyout__item.disabled, .ncm-flyout__item:disabled { color: #aeb5ba; cursor: not-allowed; }
.ncm-flyout__item.disabled small, .ncm-flyout__item:disabled small { color: #bfc5ca; }
.ne-footer { padding: 0.6rem 1.5rem; border-top: 1px solid var(--border-color); display: flex; align-items: center; justify-content: space-between; font-family: var(--font-ui); font-size: 0.7rem; color: var(--text-muted); }
.ne-footer__actions { display: flex; gap: 0.5rem; }
.ne-footer__btn { padding: 0.25rem 0.7rem; border: 1px solid var(--border-color); border-radius: 14px; background: var(--bg-card); font-family: var(--font-ui); font-size: 0.7rem; color: var(--text-secondary); cursor: pointer; transition: all 0.1s; }
.ne-footer__btn:hover { border-color: var(--accent); color: var(--accent); }

@media (max-width: 820px) { .note-list { width: 280px; } }
@media (max-width: 620px) { .note-list { width: 100%; } .note-editor { display: none; } .note-toast { top: 3.75rem; } }
</style>
