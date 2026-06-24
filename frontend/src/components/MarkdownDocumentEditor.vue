<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { Node as TiptapNode, type Editor } from '@tiptap/core'
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
  FolderOpen,
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
  UploadCloud,
  Undo2,
  X,
} from 'lucide-vue-next'
import { useDocumentStore } from '../stores/documentStore'
import { uploadDocumentAsset } from '../api/document'

type SaveState = 'clean' | 'dirty' | 'saving' | 'saved' | 'error'
type HeadingItem = {
  id: string
  level: number
  pos: number
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
const imagePanelOpen = ref(false)
const imagePanelMode = ref<'upload' | 'url'>('upload')
const imageUrl = ref('')
const imageAlt = ref('')
const imageUploading = ref(false)
const imageError = ref('')
const imageDragging = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)
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
    handlePaste(_view, event) {
      const files = Array.from(event.clipboardData?.files || []).filter(isImageFile)
      if (files.length === 0) return false
      event.preventDefault()
      void uploadAndInsertImages(files)
      return true
    },
    handleDrop(_view, event) {
      const files = Array.from(event.dataTransfer?.files || []).filter(isImageFile)
      if (files.length === 0) return false
      event.preventDefault()
      imageDragging.value = false
      void uploadAndInsertImages(files)
      return true
    },
  },
  onCreate: ({ editor }) => {
    refreshOutline(editor)
  },
  onUpdate: ({ editor }) => {
    if (ignoreNextUpdate) return
    markdownDraft.value = htmlToMarkdown(editor.getHTML())
    saveState.value = 'dirty'
    refreshOutline(editor)
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
    refreshOutline(editor.value)
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
  else if (command === 'image') openImagePanel()
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

function openImagePanel() {
  imagePanelOpen.value = true
  imageError.value = ''
  void nextTick(() => {
    if (imagePanelMode.value === 'url') document.getElementById('markdown-image-url')?.focus()
  })
}

function closeImagePanel() {
  imagePanelOpen.value = false
  imageDragging.value = false
  imageError.value = ''
}

function chooseImageFile() {
  fileInputRef.value?.click()
}

function onImageFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || []).filter(isImageFile)
  input.value = ''
  if (files.length === 0) return
  void uploadAndInsertImages(files)
}

function onImageDrop(event: DragEvent) {
  imageDragging.value = false
  const files = Array.from(event.dataTransfer?.files || []).filter(isImageFile)
  if (files.length === 0) {
    imageError.value = '请拖入 JPG、PNG、GIF 或 WEBP 图片'
    return
  }
  void uploadAndInsertImages(files)
}

function isImageFile(file: File) {
  return file.type.startsWith('image/') && /\.(jpe?g|png|gif|webp)$/i.test(file.name)
}

async function uploadAndInsertImages(files: File[]) {
  imageUploading.value = true
  imageError.value = ''
  try {
    for (const file of files) {
      const res = await uploadDocumentAsset(props.documentId, file)
      insertImageNode(res.data.url, imageAlt.value.trim() || file.name.replace(/\.[^.]+$/, ''))
    }
    imageAlt.value = ''
    closeImagePanel()
  } catch (error: any) {
    imageError.value = error?.response?.data?.detail || error?.message || '图片上传失败'
  } finally {
    imageUploading.value = false
  }
}

function insertImageFromUrl() {
  const url = imageUrl.value.trim()
  if (!url) {
    imageError.value = '请输入图片 URL'
    return
  }
  insertImageNode(url, imageAlt.value.trim())
  imageUrl.value = ''
  imageAlt.value = ''
  closeImagePanel()
}

function insertImageNode(src: string, alt: string) {
  editor.value?.chain().focus().insertContent(`<img src="${escapeHtmlAttribute(src)}" alt="${escapeHtmlAttribute(alt)}">`).run()
}

function refreshOutline(current?: Editor | null) {
  const result: HeadingItem[] = []
  if (!current) {
    headings.value = result
    return
  }

  current.state.doc.descendants((node, pos) => {
    if (node.type.name === 'heading') {
      const text = node.textContent.trim()
      if (text) {
        result.push({
          id: slugify(`${result.length}-${text}`),
          level: node.attrs.level || 1,
          pos,
          text,
        })
      }
    }
    return true
  })
  headings.value = result
}

function slugify(value: string) {
  return value.toLowerCase().replace(/[^\w\u4e00-\u9fa5]+/g, '-').replace(/^-|-$/g, '')
}

function jumpToHeading(item: HeadingItem) {
  const current = editor.value
  if (!current) return

  const target = current.view.nodeDOM(item.pos)
  if (target instanceof HTMLElement) {
    scrollHeadingIntoView(target)
    return
  }

  const targetPos = Math.min(item.pos + 1, current.state.doc.content.size)
  const coords = current.view.coordsAtPos(targetPos)
  const top = Math.max(0, window.scrollY + coords.top - getEditorChromeOffset())
  window.scrollTo({ top, behavior: 'smooth' })
}

function scrollHeadingIntoView(target: HTMLElement) {
  const top = Math.max(0, window.scrollY + target.getBoundingClientRect().top - getEditorChromeOffset())
  window.scrollTo({ top, behavior: 'smooth' })
}

function getEditorChromeOffset() {
  const topbar = document.querySelector<HTMLElement>('.mde-topbar')
  const search = searchOpen.value ? document.querySelector<HTMLElement>('.mde-search') : null
  return (topbar?.offsetHeight || 0) + (search?.offsetHeight || 0) + 28
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
        <div class="mde-image-tool">
          <button type="button" :class="{ active: imagePanelOpen }" title="图片" aria-label="插入图片" @click="runCommand('image')"><ImageIcon /></button>
          <div v-if="imagePanelOpen" class="mde-image-panel" @keydown.esc.prevent="closeImagePanel">
            <div class="mde-image-panel__header">
              <strong>插入图片</strong>
              <button type="button" title="关闭" aria-label="关闭图片面板" @click="closeImagePanel"><X /></button>
            </div>
            <div class="mde-image-tabs" role="tablist" aria-label="图片来源">
              <button type="button" :class="{ active: imagePanelMode === 'upload' }" @click="imagePanelMode = 'upload'">本地图片</button>
              <button type="button" :class="{ active: imagePanelMode === 'url' }" @click="imagePanelMode = 'url'">图片 URL</button>
            </div>

            <label class="mde-image-field">
              <span>Alt 文本</span>
              <input v-model="imageAlt" type="text" placeholder="用于无障碍和图片说明" />
            </label>

            <div v-if="imagePanelMode === 'upload'" class="mde-image-upload">
              <input ref="fileInputRef" type="file" accept="image/png,image/jpeg,image/gif,image/webp" multiple hidden @change="onImageFileChange" />
              <div
                :class="['mde-image-dropzone', { dragging: imageDragging }]"
                @dragenter.prevent="imageDragging = true"
                @dragover.prevent="imageDragging = true"
                @dragleave.prevent="imageDragging = false"
                @drop.prevent="onImageDrop"
              >
                <UploadCloud />
                <strong>拖拽图片到这里</strong>
                <span>支持 JPG / PNG / GIF / WEBP，也可以直接粘贴截图</span>
              </div>
              <button type="button" class="mde-image-primary" :disabled="imageUploading" @click="chooseImageFile">
                <LoaderCircle v-if="imageUploading" class="spin" />
                <FolderOpen v-else />
                {{ imageUploading ? '上传中...' : '选择本地图片' }}
              </button>
            </div>

            <div v-else class="mde-image-url">
              <label class="mde-image-field">
                <span>图片 URL</span>
                <input
                  id="markdown-image-url"
                  v-model="imageUrl"
                  type="url"
                  placeholder="https://example.com/image.png"
                  @keydown.enter.prevent="insertImageFromUrl"
                />
              </label>
              <button type="button" class="mde-image-primary" @click="insertImageFromUrl">
                <ImageIcon />
                插入 URL 图片
              </button>
            </div>

            <p v-if="imageError" class="mde-image-error">{{ imageError }}</p>
          </div>
        </div>
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

.mde-image-tool {
  position: relative;
  display: inline-flex;
}

.mde-image-panel {
  position: absolute;
  top: calc(100% + 0.65rem);
  left: 50%;
  z-index: 42;
  display: grid;
  width: min(360px, calc(100vw - 1.5rem));
  gap: 0.7rem;
  padding: 0.8rem;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: rgba(250, 248, 245, 0.98);
  box-shadow: 0 12px 28px rgba(61, 46, 36, 0.16);
  color: var(--text-primary);
  font-family: var(--font-ui);
  transform: translateX(-50%);
}

.mde-image-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.mde-image-panel__header strong {
  font-size: 0.84rem;
}

.mde-image-panel__header button {
  width: 28px;
  min-width: 28px;
  height: 28px;
}

.mde-image-tabs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.35rem;
  padding: 0.2rem;
  border: 1px solid var(--border-color);
  border-radius: 7px;
  background: rgba(255, 255, 255, 0.45);
}

.mde-image-tabs button {
  width: auto;
  min-width: 0;
  border-radius: 5px;
  font-size: 0.76rem;
}

.mde-image-tabs button.active {
  border-color: rgba(198, 122, 78, 0.34);
  background: var(--accent-light);
  color: var(--accent);
}

.mde-image-field {
  display: grid;
  gap: 0.32rem;
  color: var(--text-secondary);
  font-size: 0.74rem;
  text-align: left;
}

.mde-image-field input {
  width: 100%;
  min-width: 0;
  height: 34px;
  border: 1px solid var(--border-color);
  border-radius: 7px;
  background: rgba(255, 255, 255, 0.68);
  color: var(--text-primary);
  font: inherit;
  outline: none;
  padding: 0 0.6rem;
}

.mde-image-field input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px rgba(198, 122, 78, 0.16);
}

.mde-image-upload,
.mde-image-url {
  display: grid;
  gap: 0.62rem;
}

.mde-image-dropzone {
  display: grid;
  justify-items: center;
  gap: 0.35rem;
  padding: 1rem 0.85rem;
  border: 1px dashed var(--border-input);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.44);
  color: var(--text-secondary);
  text-align: center;
  transition: border-color 0.16s, background 0.16s, color 0.16s;
}

.mde-image-dropzone.dragging {
  border-color: var(--accent);
  background: var(--accent-light);
  color: var(--accent);
}

.mde-image-dropzone svg {
  width: 24px;
  height: 24px;
}

.mde-image-dropzone strong {
  color: var(--text-primary);
  font-size: 0.82rem;
}

.mde-image-dropzone span {
  color: var(--text-muted);
  font-size: 0.72rem;
  line-height: 1.45;
}

.mde-image-primary {
  width: 100% !important;
  gap: 0.42rem;
  border-color: rgba(198, 122, 78, 0.34) !important;
  background: var(--accent-light) !important;
  color: var(--accent) !important;
  font-size: 0.78rem;
}

.mde-image-primary:disabled {
  cursor: wait;
  opacity: 0.68;
}

.mde-image-error {
  margin: 0;
  color: #b42318;
  font-size: 0.74rem;
  line-height: 1.45;
  text-align: left;
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
