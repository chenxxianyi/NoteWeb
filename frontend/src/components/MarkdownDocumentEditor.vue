<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { Node as TiptapNode } from '@tiptap/core'
import { EditorContent, useEditor } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Link from '@tiptap/extension-link'
import Placeholder from '@tiptap/extension-placeholder'
import TaskItem from '@tiptap/extension-task-item'
import TaskList from '@tiptap/extension-task-list'
import { marked } from 'marked'
import {
  ArrowLeft,
  Bold,
  Check,
  Code2,
  FileText,
  Heading1,
  Heading2,
  Image as ImageIcon,
  Italic,
  Link as LinkIcon,
  List,
  ListOrdered,
  LoaderCircle,
  PanelRightClose,
  PanelRightOpen,
  Quote,
  Redo2,
  Save,
  Search,
  Strikethrough,
  Undo2,
} from 'lucide-vue-next'
import { useDocumentStore } from '../stores/documentStore'

type SaveState = 'clean' | 'dirty' | 'saving' | 'saved' | 'error'
type HeadingItem = {
  id: string
  level: number
  text: string
}

const props = defineProps<{
  documentId: number
  title: string
  content: string
  fileType: string
}>()
const emit = defineEmits<{
  back: []
}>()

const documentStore = useDocumentStore()
const markdownDraft = ref(props.content || '')
const saveState = ref<SaveState>('clean')
const outlineOpen = ref(true)
const searchOpen = ref(false)
const searchQuery = ref('')
const currentMatchIndex = ref(0)
const headings = ref<HeadingItem[]>([])
let saveTimer: ReturnType<typeof setTimeout> | null = null
let ignoreNextUpdate = false

const MarkdownImage = TiptapNode.create({
  name: 'image',
  group: 'block',
  atom: true,
  draggable: true,
  addAttributes() {
    return {
      src: { default: '' },
      alt: { default: '' },
      title: { default: null },
    }
  },
  parseHTML() {
    return [{ tag: 'img[src]' }]
  },
  renderHTML({ HTMLAttributes }) {
    return ['img', HTMLAttributes]
  },
})

const editor = useEditor({
  content: markdownToHtml(props.content || ''),
  extensions: [
    StarterKit.configure({ link: false }),
    Link.configure({
      autolink: true,
      openOnClick: false,
      defaultProtocol: 'https',
    }),
    TaskList,
    TaskItem.configure({ nested: true }),
    MarkdownImage,
    Placeholder.configure({
      placeholder: '开始写作，或粘贴 Markdown 内容...',
    }),
  ],
  editorProps: {
    attributes: {
      class: 'mde-prose',
      spellcheck: 'true',
    },
    handleKeyDown(_view, event) {
      const primary = event.ctrlKey || event.metaKey
      if (primary && event.key.toLowerCase() === 's') {
        event.preventDefault()
        void saveNow()
        return true
      }
      if (primary && event.key.toLowerCase() === 'f') {
        event.preventDefault()
        searchOpen.value = true
        void nextTick(() => document.getElementById('markdown-search-input')?.focus())
        return true
      }
      return false
    },
  },
  onCreate: ({ editor }) => {
    refreshOutline(editor.getJSON())
  },
  onUpdate: ({ editor }) => {
    if (ignoreNextUpdate) return
    markdownDraft.value = htmlToMarkdown(editor.getHTML())
    saveState.value = 'dirty'
    refreshOutline(editor.getJSON())
    scheduleAutoSave()
  },
})

const characterCount = computed(() => markdownDraft.value.length)
const wordCount = computed(() => {
  const latinWords = markdownDraft.value.match(/[A-Za-z0-9_]+/g)?.length || 0
  const cjkChars = markdownDraft.value.match(/[\u4e00-\u9fa5]/g)?.length || 0
  return latinWords + cjkChars
})
const readingMinutes = computed(() => Math.max(1, Math.ceil(wordCount.value / 350)))
const searchMatches = computed(() => {
  const query = searchQuery.value.trim().toLowerCase()
  if (!query) return 0
  const text = editor.value?.state.doc.textContent.toLowerCase() || ''
  let count = 0
  let from = 0
  while (from < text.length) {
    const index = text.indexOf(query, from)
    if (index === -1) break
    count += 1
    from = index + Math.max(query.length, 1)
  }
  return count
})
const saveLabel = computed(() => {
  if (saveState.value === 'saving') return '保存中'
  if (saveState.value === 'dirty') return '未保存'
  if (saveState.value === 'error') return '保存失败'
  return '已保存'
})

watch(
  () => props.content,
  (content) => {
    if (content === markdownDraft.value) return
    markdownDraft.value = content || ''
    ignoreNextUpdate = true
    editor.value?.commands.setContent(markdownToHtml(content || ''), { emitUpdate: false })
    refreshOutline(editor.value?.getJSON())
    void nextTick(() => {
      ignoreNextUpdate = false
      saveState.value = 'clean'
    })
  },
)

watch(searchQuery, () => {
  currentMatchIndex.value = searchMatches.value > 0 ? 1 : 0
})

onBeforeUnmount(() => {
  if (saveTimer) clearTimeout(saveTimer)
  if (saveState.value === 'dirty') void saveNow()
  editor.value?.destroy()
})

function markdownToHtml(markdown: string) {
  return marked.parse(markdown || '', { async: false }) as string
}

function htmlToMarkdown(html: string) {
  const doc = new DOMParser().parseFromString(html, 'text/html')
  return Array.from(doc.body.childNodes).map(nodeToMarkdown).join('\n\n').trimEnd()
}

function nodeToMarkdown(node: ChildNode): string {
  if (node.nodeType === globalThis.Node.TEXT_NODE) return escapeMarkdownText(node.textContent || '')
  if (!(node instanceof HTMLElement)) return ''

  const children = Array.from(node.childNodes).map(nodeToMarkdown).join('')
  const tag = node.tagName.toLowerCase()

  if (/^h[1-6]$/.test(tag)) return `${'#'.repeat(Number(tag[1]))} ${children.trim()}`
  if (tag === 'p') return children.trim()
  if (tag === 'strong' || tag === 'b') return `**${children}**`
  if (tag === 'em' || tag === 'i') return `*${children}*`
  if (tag === 's' || tag === 'strike') return `~~${children}~~`
  if (tag === 'code' && node.parentElement?.tagName.toLowerCase() !== 'pre') return `\`${node.textContent || ''}\``
  if (tag === 'pre') return `\`\`\`\n${node.textContent?.replace(/\n$/, '') || ''}\n\`\`\``
  if (tag === 'blockquote') {
    return children.split('\n').map((line) => `> ${line}`).join('\n')
  }
  if (tag === 'a') return `[${children || node.textContent || ''}](${node.getAttribute('href') || ''})`
  if (tag === 'img') return `![${node.getAttribute('alt') || ''}](${node.getAttribute('src') || ''})`
  if (tag === 'hr') return '---'
  if (tag === 'ul' || tag === 'ol') return listToMarkdown(node, tag === 'ol')
  if (tag === 'li') return children.trim()
  if (tag === 'br') return '\n'
  return children
}

function listToMarkdown(listNode: HTMLElement, ordered: boolean) {
  return Array.from(listNode.children)
    .filter((child): child is HTMLElement => child instanceof HTMLElement && child.tagName.toLowerCase() === 'li')
    .map((item, index) => {
      const checkbox = item.querySelector('input[type="checkbox"]') as HTMLInputElement | null
      const marker = checkbox ? `- [${checkbox.checked ? 'x' : ' '}]` : ordered ? `${index + 1}.` : '-'
      const cloned = item.cloneNode(true) as HTMLElement
      cloned.querySelector('input[type="checkbox"]')?.remove()
      return `${marker} ${Array.from(cloned.childNodes).map(nodeToMarkdown).join('').trim()}`
    })
    .join('\n')
}

function escapeMarkdownText(text: string) {
  return text.replace(/\u00a0/g, ' ')
}

function escapeHtmlAttribute(value: string) {
  return value.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;')
}

function scheduleAutoSave() {
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    void saveNow()
  }, 1200)
}

async function saveNow() {
  if (saveTimer) {
    clearTimeout(saveTimer)
    saveTimer = null
  }
  if (saveState.value === 'saving') return
  saveState.value = 'saving'
  try {
    await documentStore.updateContent(props.documentId, markdownDraft.value)
    saveState.value = 'saved'
  } catch {
    saveState.value = 'error'
  }
}

function runCommand(command: string) {
  const current = editor.value
  const chain = current?.chain().focus()
  if (!current || !chain) return

  if (command === 'bold') chain.toggleBold().run()
  else if (command === 'italic') chain.toggleItalic().run()
  else if (command === 'strike') chain.toggleStrike().run()
  else if (command === 'heading1') chain.toggleHeading({ level: 1 }).run()
  else if (command === 'heading2') chain.toggleHeading({ level: 2 }).run()
  else if (command === 'bulletList') chain.toggleBulletList().run()
  else if (command === 'orderedList') chain.toggleOrderedList().run()
  else if (command === 'taskList') chain.toggleTaskList().run()
  else if (command === 'blockquote') chain.toggleBlockquote().run()
  else if (command === 'code') chain.toggleCode().run()
  else if (command === 'codeBlock') chain.toggleCodeBlock().run()
  else if (command === 'undo') chain.undo().run()
  else if (command === 'redo') chain.redo().run()
  else if (command === 'link') setLink()
  else if (command === 'image') insertImage()
}

function isActive(command: string) {
  const current = editor.value
  if (!current) return false
  if (command === 'bold') return current.isActive('bold')
  if (command === 'italic') return current.isActive('italic')
  if (command === 'strike') return current.isActive('strike')
  if (command === 'heading1') return current.isActive('heading', { level: 1 })
  if (command === 'heading2') return current.isActive('heading', { level: 2 })
  if (command === 'bulletList') return current.isActive('bulletList')
  if (command === 'orderedList') return current.isActive('orderedList')
  if (command === 'taskList') return current.isActive('taskList')
  if (command === 'blockquote') return current.isActive('blockquote')
  if (command === 'code') return current.isActive('code')
  if (command === 'codeBlock') return current.isActive('codeBlock')
  if (command === 'link') return current.isActive('link')
  return false
}

function setLink() {
  const current = editor.value
  if (!current) return
  const previousUrl = current.getAttributes('link').href || ''
  const url = window.prompt('输入链接地址', previousUrl || 'https://')
  if (url === null) return
  if (!url.trim()) {
    current.chain().focus().extendMarkRange('link').unsetLink().run()
    return
  }
  current.chain().focus().extendMarkRange('link').setLink({ href: url.trim() }).run()
}

function insertImage() {
  const src = window.prompt('输入图片 URL', 'https://')
  if (!src?.trim()) return
  const alt = window.prompt('输入图片说明', '') || ''
  editor.value?.chain().focus().insertContent(`<img src="${escapeHtmlAttribute(src.trim())}" alt="${escapeHtmlAttribute(alt)}">`).run()
}

function refreshOutline(json: any) {
  const result: HeadingItem[] = []
  const walk = (node: any) => {
    if (!node) return
    if (node.type === 'heading') {
      const text = extractText(node).trim()
      if (text) {
        result.push({
          id: slugify(`${result.length}-${text}`),
          level: node.attrs?.level || 1,
          text,
        })
      }
    }
    node.content?.forEach(walk)
  }
  walk(json)
  headings.value = result
  void nextTick(applyHeadingIds)
}

function extractText(node: any): string {
  if (!node) return ''
  if (typeof node.text === 'string') return node.text
  return node.content?.map(extractText).join('') || ''
}

function slugify(value: string) {
  return value.toLowerCase().replace(/[^\w\u4e00-\u9fa5]+/g, '-').replace(/^-|-$/g, '')
}

function applyHeadingIds() {
  const root = document.querySelector('.mde-prose')
  if (!root) return
  root.querySelectorAll('h1,h2,h3,h4,h5,h6').forEach((heading, index) => {
    heading.id = headings.value[index]?.id || `heading-${index}`
  })
}

function jumpToHeading(item: HeadingItem) {
  document.getElementById(item.id)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

function findMatch(direction: 1 | -1) {
  const query = searchQuery.value.trim()
  const current = editor.value
  if (!query || !current) return
  const total = searchMatches.value
  if (!total) return

  currentMatchIndex.value += direction
  if (currentMatchIndex.value > total) currentMatchIndex.value = 1
  if (currentMatchIndex.value < 1) currentMatchIndex.value = total

  let seen = 0
  const target = currentMatchIndex.value
  let cursor = 0
  const lowerQuery = query.toLowerCase()
  current.state.doc.descendants((node, pos) => {
    if (!node.isText || !node.text) return true
    let localIndex = node.text.toLowerCase().indexOf(lowerQuery)
    while (localIndex !== -1) {
      seen += 1
      if (seen === target) {
        const from = pos + localIndex
        current.chain().focus().setTextSelection({ from, to: from + query.length }).run()
        return false
      }
      cursor = localIndex + lowerQuery.length
      localIndex = node.text.toLowerCase().indexOf(lowerQuery, cursor)
    }
    return true
  })
}
</script>

<template>
  <section class="mde-shell">
    <header class="mde-topbar">
      <div class="mde-title">
        <button type="button" class="mde-back" title="返回" aria-label="返回" @click="emit('back')">
          <ArrowLeft />
        </button>
        <FileText aria-hidden="true" />
        <div>
          <h1>{{ title || '未命名文档' }}</h1>
          <span>{{ fileType.toUpperCase() }} · Markdown 编辑</span>
        </div>
      </div>

      <div class="mde-actions" role="toolbar" aria-label="Markdown 编辑工具">
        <button type="button" title="撤销" aria-label="撤销" @click="runCommand('undo')"><Undo2 /></button>
        <button type="button" title="重做" aria-label="重做" @click="runCommand('redo')"><Redo2 /></button>
        <span class="mde-divider"></span>
        <button type="button" :class="{ active: isActive('bold') }" title="加粗" aria-label="加粗" @click="runCommand('bold')"><Bold /></button>
        <button type="button" :class="{ active: isActive('italic') }" title="斜体" aria-label="斜体" @click="runCommand('italic')"><Italic /></button>
        <button type="button" :class="{ active: isActive('strike') }" title="删除线" aria-label="删除线" @click="runCommand('strike')"><Strikethrough /></button>
        <button type="button" :class="{ active: isActive('code') }" title="行内代码" aria-label="行内代码" @click="runCommand('code')"><Code2 /></button>
        <span class="mde-divider"></span>
        <button type="button" :class="{ active: isActive('heading1') }" title="一级标题" aria-label="一级标题" @click="runCommand('heading1')"><Heading1 /></button>
        <button type="button" :class="{ active: isActive('heading2') }" title="二级标题" aria-label="二级标题" @click="runCommand('heading2')"><Heading2 /></button>
        <button type="button" :class="{ active: isActive('bulletList') }" title="无序列表" aria-label="无序列表" @click="runCommand('bulletList')"><List /></button>
        <button type="button" :class="{ active: isActive('orderedList') }" title="有序列表" aria-label="有序列表" @click="runCommand('orderedList')"><ListOrdered /></button>
        <button type="button" :class="{ active: isActive('blockquote') }" title="引用" aria-label="引用" @click="runCommand('blockquote')"><Quote /></button>
        <button type="button" :class="{ active: isActive('link') }" title="链接" aria-label="链接" @click="runCommand('link')"><LinkIcon /></button>
        <button type="button" title="图片" aria-label="插入图片" @click="runCommand('image')"><ImageIcon /></button>
      </div>

      <div class="mde-side-actions">
        <button type="button" :class="['mde-save', `mde-save--${saveState}`]" @click="saveNow">
          <LoaderCircle v-if="saveState === 'saving'" class="spin" />
          <Check v-else-if="saveState === 'saved' || saveState === 'clean'" />
          <Save v-else />
          <span>{{ saveLabel }}</span>
        </button>
        <button type="button" title="搜索" aria-label="搜索" @click="searchOpen = !searchOpen"><Search /></button>
        <button type="button" title="大纲" aria-label="切换大纲" @click="outlineOpen = !outlineOpen">
          <PanelRightClose v-if="outlineOpen" />
          <PanelRightOpen v-else />
        </button>
      </div>
    </header>

    <div v-if="searchOpen" class="mde-search" role="search">
      <Search aria-hidden="true" />
      <input
        id="markdown-search-input"
        v-model="searchQuery"
        type="search"
        placeholder="搜索当前文档"
        @keydown.enter.prevent="findMatch(1)"
        @keydown.esc.prevent="searchOpen = false"
      />
      <span>{{ searchMatches ? `${currentMatchIndex || 1}/${searchMatches}` : '0/0' }}</span>
      <button type="button" @click="findMatch(-1)">上一个</button>
      <button type="button" @click="findMatch(1)">下一个</button>
    </div>

    <main :class="['mde-main', { 'outline-closed': !outlineOpen }]">
      <article class="mde-canvas">
        <EditorContent v-if="editor" :editor="editor" />
      </article>

      <aside v-if="outlineOpen" class="mde-outline" aria-label="文档大纲">
        <div class="mde-outline__header">大纲</div>
        <button
          v-for="item in headings"
          :key="item.id"
          type="button"
          class="mde-outline__item"
          :style="{ paddingLeft: `${0.45 + (item.level - 1) * 0.72}rem` }"
          @click="jumpToHeading(item)"
        >
          {{ item.text }}
        </button>
        <p v-if="headings.length === 0" class="mde-outline__empty">暂无标题</p>
      </aside>
    </main>

    <footer class="mde-statusbar">
      <span>{{ wordCount }} 词</span>
      <span>{{ characterCount }} 字符</span>
      <span>约 {{ readingMinutes }} 分钟阅读</span>
      <span>{{ saveLabel }}</span>
      <span>100%</span>
    </footer>
  </section>
</template>

<style scoped>
.mde-shell {
  min-height: 100vh;
  background: var(--bg-page);
  color: var(--text-primary);
  font-family: var(--font-body);
}

.mde-topbar {
  position: sticky;
  top: 0;
  z-index: 20;
  display: grid;
  grid-template-columns: minmax(180px, 1fr) auto minmax(190px, 1fr);
  align-items: center;
  gap: 0.75rem;
  padding: 0.7rem 1rem;
  border-bottom: 1px solid var(--border-color);
  background: rgba(250, 248, 245, 0.94);
  backdrop-filter: blur(12px);
  font-family: var(--font-ui);
}

.mde-title {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  min-width: 0;
}

.mde-title svg {
  width: 22px;
  height: 22px;
  color: var(--accent);
  flex: 0 0 auto;
}

.mde-title h1 {
  overflow: hidden;
  margin: 0;
  color: var(--text-primary);
  font-family: var(--font-display);
  font-size: 1.08rem;
  font-weight: 600;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mde-title span {
  display: block;
  margin-top: 0.1rem;
  color: var(--text-muted);
  font-size: 0.72rem;
}

.mde-actions,
.mde-side-actions {
  display: flex;
  align-items: center;
  gap: 0.18rem;
}

.mde-actions {
  justify-content: center;
}

.mde-side-actions {
  justify-content: flex-end;
}

.mde-back,
.mde-actions button,
.mde-side-actions button,
.mde-search button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  height: 32px;
  border: 1px solid transparent;
  border-radius: 6px;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: background 0.16s, border-color 0.16s, color 0.16s;
}

.mde-back:hover,
.mde-actions button:hover,
.mde-side-actions button:hover,
.mde-search button:hover {
  border-color: var(--border-color);
  background: var(--accent-light);
  color: var(--accent);
}

.mde-back:focus-visible,
.mde-actions button:focus-visible,
.mde-side-actions button:focus-visible,
.mde-search button:focus-visible,
.mde-search input:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.mde-actions button.active {
  border-color: rgba(198, 122, 78, 0.36);
  background: var(--accent-light);
  color: var(--accent);
}

.mde-actions svg,
.mde-back svg,
.mde-side-actions svg,
.mde-search svg {
  width: 17px;
  height: 17px;
}

.mde-divider {
  width: 1px;
  height: 20px;
  margin: 0 0.25rem;
  background: var(--border-color);
}

.mde-save {
  width: auto;
  min-width: 86px;
  gap: 0.35rem;
  padding: 0 0.65rem;
  font-size: 0.76rem;
}

.mde-save--error {
  color: #b42318 !important;
}

.mde-save--dirty {
  color: #8b6f47 !important;
}

.spin {
  animation: mde-spin 0.8s linear infinite;
}

.mde-search {
  position: sticky;
  top: 57px;
  z-index: 18;
  display: grid;
  grid-template-columns: auto minmax(120px, 1fr) auto auto auto;
  align-items: center;
  gap: 0.5rem;
  max-width: 760px;
  margin: 0.75rem auto 0;
  padding: 0.4rem 0.5rem;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: rgba(250, 248, 245, 0.98);
  box-shadow: var(--shadow-float);
  font-family: var(--font-ui);
  font-size: 0.78rem;
}

.mde-search input {
  min-width: 0;
  border: none;
  background: transparent;
  color: var(--text-primary);
  font: inherit;
  outline: none;
}

.mde-search span {
  color: var(--text-muted);
  white-space: nowrap;
}

.mde-search button {
  width: auto;
  padding: 0 0.55rem;
  font-size: 0.74rem;
}

.mde-main {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 240px;
  gap: clamp(1rem, 3vw, 2rem);
  width: min(1160px, calc(100% - 2rem));
  margin: 0 auto;
  padding: 2.25rem 0 4rem;
}

.mde-main.outline-closed {
  grid-template-columns: minmax(0, 1fr);
  width: min(860px, calc(100% - 2rem));
}

.mde-canvas {
  min-width: 0;
}

.mde-outline {
  position: sticky;
  top: 82px;
  align-self: start;
  max-height: calc(100vh - 120px);
  overflow: auto;
  padding: 0.4rem 0;
  border-left: 1px solid var(--border-color);
  font-family: var(--font-ui);
}

.mde-outline__header {
  padding: 0.2rem 0.7rem 0.55rem;
  color: var(--text-secondary);
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
}

.mde-outline__item {
  display: block;
  width: 100%;
  min-height: 30px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  font: inherit;
  font-size: 0.78rem;
  line-height: 1.35;
  overflow-wrap: anywhere;
  padding-right: 0.45rem;
  text-align: left;
  transition: background 0.16s, color 0.16s;
}

.mde-outline__item:hover {
  background: rgba(198, 122, 78, 0.09);
  color: var(--accent);
}

.mde-outline__empty {
  padding: 0.2rem 0.7rem;
  color: var(--text-muted);
  font-size: 0.78rem;
}

.mde-statusbar {
  position: fixed;
  right: 1rem;
  bottom: 0.85rem;
  z-index: 12;
  display: flex;
  align-items: center;
  gap: 0.45rem;
  max-width: calc(100vw - 2rem);
  padding: 0.38rem 0.62rem;
  border: 1px solid var(--border-color);
  border-radius: 18px;
  background: rgba(250, 248, 245, 0.94);
  box-shadow: var(--shadow-float);
  color: var(--text-muted);
  font-family: var(--font-ui);
  font-size: 0.72rem;
  white-space: nowrap;
}

.mde-statusbar span + span::before {
  content: "·";
  margin-right: 0.45rem;
}

.mde-canvas :deep(.mde-prose) {
  max-width: 820px;
  min-height: calc(100vh - 12rem);
  margin: 0 auto;
  outline: none;
  color: var(--text-primary);
  font-size: 1.05rem;
  line-height: 1.82;
}

.mde-canvas :deep(.mde-prose p) {
  margin: 0 0 1em;
}

.mde-canvas :deep(.mde-prose h1),
.mde-canvas :deep(.mde-prose h2),
.mde-canvas :deep(.mde-prose h3),
.mde-canvas :deep(.mde-prose h4),
.mde-canvas :deep(.mde-prose h5),
.mde-canvas :deep(.mde-prose h6) {
  margin: 1.6em 0 0.55em;
  color: var(--text-primary);
  font-family: var(--font-display);
  font-weight: 650;
  line-height: 1.28;
}

.mde-canvas :deep(.mde-prose h1) { font-size: 2rem; }
.mde-canvas :deep(.mde-prose h2) { font-size: 1.55rem; }
.mde-canvas :deep(.mde-prose h3) { font-size: 1.28rem; }

.mde-canvas :deep(.mde-prose ul),
.mde-canvas :deep(.mde-prose ol) {
  margin: 0 0 1em;
  padding-left: 1.45em;
}

.mde-canvas :deep(.mde-prose li) {
  margin: 0.22em 0;
}

.mde-canvas :deep(.mde-prose blockquote) {
  margin: 1.1rem 0;
  padding: 0.45rem 0 0.45rem 1rem;
  border-left: 3px solid var(--accent);
  color: var(--text-secondary);
}

.mde-canvas :deep(.mde-prose code) {
  border-radius: 4px;
  background: #eee8df;
  color: #4f3a2d;
  font-family: var(--font-mono);
  font-size: 0.88em;
  padding: 0.1em 0.35em;
}

.mde-canvas :deep(.mde-prose pre) {
  margin: 1rem 0;
  overflow-x: auto;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: #25211d;
  color: #f8f4ee;
  padding: 1rem;
}

.mde-canvas :deep(.mde-prose pre code) {
  background: transparent;
  color: inherit;
  padding: 0;
}

.mde-canvas :deep(.mde-prose a) {
  color: var(--accent);
  text-decoration: underline;
  text-underline-offset: 3px;
}

.mde-canvas :deep(.mde-prose img) {
  display: block;
  max-width: 100%;
  height: auto;
  margin: 1.2rem auto;
  border-radius: 8px;
}

.mde-canvas :deep(.mde-prose hr) {
  margin: 1.8rem 0;
  border: none;
  border-top: 1px solid var(--border-color);
}

.mde-canvas :deep(.mde-prose ul[data-type="taskList"]) {
  list-style: none;
  padding-left: 0;
}

.mde-canvas :deep(.mde-prose ul[data-type="taskList"] li) {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
}

.mde-canvas :deep(.mde-prose ul[data-type="taskList"] label) {
  padding-top: 0.1rem;
}

.mde-canvas :deep(.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  float: left;
  height: 0;
  color: var(--text-muted);
  pointer-events: none;
}

@keyframes mde-spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@media (max-width: 1040px) {
  .mde-topbar {
    grid-template-columns: 1fr auto;
  }

  .mde-actions {
    order: 3;
    grid-column: 1 / -1;
    justify-content: flex-start;
    overflow-x: auto;
    padding-top: 0.35rem;
  }

  .mde-main {
    grid-template-columns: minmax(0, 1fr);
    width: min(860px, calc(100% - 2rem));
  }

  .mde-outline {
    display: none;
  }
}

@media (max-width: 640px) {
  .mde-topbar {
    padding: 0.6rem 0.7rem;
  }

  .mde-title h1 {
    font-size: 0.96rem;
  }

  .mde-title span,
  .mde-save span {
    display: none;
  }

  .mde-save {
    min-width: 32px;
    padding: 0;
  }

  .mde-main {
    width: min(100% - 1.3rem, 860px);
    padding-top: 1.4rem;
  }

  .mde-search {
    top: 92px;
    grid-template-columns: auto 1fr auto;
    width: calc(100% - 1.2rem);
  }

  .mde-search button {
    display: none;
  }

  .mde-canvas :deep(.mde-prose) {
    font-size: 1rem;
    line-height: 1.78;
  }

  .mde-statusbar {
    left: 0.65rem;
    right: 0.65rem;
    justify-content: center;
    overflow: hidden;
  }
}
</style>
