<script setup lang="ts">
import { ref, computed, nextTick, onBeforeUnmount, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import { useAnnotationStore } from '../stores/annotationStore'
import { useAIConversationStore } from '../stores/aiConversationStore'
import {
  ArrowLeft,
  ArrowRight,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  Circle,
  Download,
  Eraser,
  FileText,
  Highlighter,
  LoaderCircle,
  Maximize,
  MessageSquare,
  Minus,
  MoreHorizontal,
  MousePointer2,
  Palette,
  PenLine,
  Redo2,
  Search,
  Send,
  Settings,
  Share2,
  Square,
  Sparkles,
  Type,
  Undo2,
  ZoomIn,
  ZoomOut,
} from 'lucide-vue-next'
import PDFViewer from '../components/PDFViewer.vue'
import AnnotationToolbar from '../components/AnnotationToolbar.vue'
import MarkdownDocumentEditor from '../components/MarkdownDocumentEditor.vue'
import type { Annotation } from '../types/annotation'
import type { PDFActiveTool, ShapeType, TextDrawing } from '../components/pdfDrawingTypes'

const route = useRoute()
const router = useRouter()
const documentStore = useDocumentStore()
const annotationStore = useAnnotationStore()
const aiStore = useAIConversationStore()

const docId = computed(() => Number(route.params.documentId))
// Capture docId at mount time so unmount still has a valid id (route param may be NaN after navigation)
let capturedDocId = 0
const panelLeftOpen = ref(false)
const panelRightOpen = ref(false)
const topbarHidden = ref(false)
const loading = ref(true)
let lastScroll = 0

// AI Assistant state
const aiInput = ref('')
const aiMode = ref<'chat' | 'search'>('chat')
const aiInputPlaceholder = computed(() => {
  if (aiMode.value === 'search') return '联网搜索...'
  return '向 AI 提问...'
})
const messages = computed(() => aiStore.messages)
const aiMessagesRef = ref<HTMLElement | null>(null)

// PDF drawing state
const pdfRef = ref<InstanceType<typeof PDFViewer> | null>(null)
const pdfActiveTool = ref<PDFActiveTool>('none')
const pdfShapeType = ref<ShapeType>('rectangle')
const pdfShapeMenuOpen = ref(false)
const pdfStylePanelOpen = ref(false)
const pdfStyleButtonRef = ref<HTMLElement | null>(null)
const pdfStylePanelRef = ref<HTMLElement | null>(null)
const pdfZoomMenuOpen = ref(false)
const pdfMoreMenuOpen = ref(false)
const pdfSearchOpen = ref(false)
const pdfSettingsOpen = ref(false)
const aiPanelOpen = ref(false)
const pdfSearchQuery = ref('')
const pdfSearchLoading = ref(false)
const pdfSearchResults = ref<Array<{ page: number; excerpt: string }>>([])
const pdfPageEditing = ref(false)
const pdfPageInput = ref('1')
const pdfPenColor = ref('#FF0000')
const pdfPenWidth = ref(3)
const pdfTextSize = ref(24)
const pdfSelectedText = ref<TextDrawing | null>(null)
const pdfEraserSize = ref(24)
const pdfEraseMode = ref<'freehand' | 'area'>('freehand')
const pdfZoom = ref(100)
const pdfCurPage = ref(1)
const pdfPageCount = ref(1)
const pdfCanUndo = ref(false)
const pdfCanRedo = ref(false)
const pdfSaving = ref(false)
const pdfExporting = ref(false)
const pdfSharing = ref(false)
const readerToolbarAutoHide = ref(true)
const pdfCustomColorInput = ref(pdfPenColor.value)
const pdfRecentColors = ref<string[]>(['#FF0000', '#00BFFF', '#DCC00D'])

const pdfThemeColors = [
  '#1F1B16',
  '#7A6A5C',
  '#C67A4E',
  '#E25D5D',
  '#FF8A3D',
  '#DCC00D',
  '#62B255',
  '#22A7A7',
  '#2F80ED',
  '#6C5CE7',
  '#C43ACF',
  '#F05A93',
]

const pdfHighlightColors = [
  '#FFF2A8',
  '#D8FF9F',
  '#BFE3FF',
  '#FFD2E5',
  '#E5D4FF',
  '#FFE1B2',
]

const pdfUsesStyle = computed(() =>
  pdfActiveTool.value === 'pen' ||
  pdfActiveTool.value === 'highlighter' ||
  pdfActiveTool.value === 'shape' ||
  pdfActiveTool.value === 'text' ||
  pdfActiveTool.value === 'eraser' ||
  pdfSelectedText.value !== null)

const pdfShowsTextControls = computed(() =>
  pdfActiveTool.value === 'text' || pdfSelectedText.value !== null)

const pdfShowsColorControls = computed(() =>
  pdfActiveTool.value !== 'eraser')

const pdfActiveToolLabel = computed(() => {
  const labels: Record<PDFActiveTool, string> = {
    none: '浏览',
    pen: '画笔',
    highlighter: '高亮',
    eraser: '橡皮',
    select: '选择',
    shape: '形状',
    text: '文本',
  }
  return labels[pdfActiveTool.value]
})

const pdfSaveStatusLabel = computed(() => {
  if (pdfSaving.value) return '保存中'
  return '已保存'
})

function pdfSwitchTool(tool: PDFActiveTool) {
  pdfActiveTool.value = pdfActiveTool.value === tool ? 'none' : tool
  if (tool !== 'shape') pdfShapeMenuOpen.value = false
  if (tool !== 'select') pdfSelectedText.value = null
  pdfZoomMenuOpen.value = false
  pdfMoreMenuOpen.value = false
  pdfStylePanelOpen.value = pdfUsesStyle.value
}
function pdfToggleShapeMenu() {
  pdfShapeMenuOpen.value = !pdfShapeMenuOpen.value
  pdfStylePanelOpen.value = false
  pdfZoomMenuOpen.value = false
  pdfMoreMenuOpen.value = false
}
function pdfChooseShape(shapeType: ShapeType) {
  pdfShapeType.value = shapeType
  pdfActiveTool.value = 'shape'
  pdfShapeMenuOpen.value = false
  pdfStylePanelOpen.value = true
}
function pdfToggleEraseMode() {
  pdfEraseMode.value = pdfEraseMode.value === 'freehand' ? 'area' : 'freehand'
}
function pdfToggleStylePanel() {
  if (!pdfUsesStyle.value) return
  pdfStylePanelOpen.value = !pdfStylePanelOpen.value
  pdfShapeMenuOpen.value = false
  pdfZoomMenuOpen.value = false
  pdfMoreMenuOpen.value = false
}
function onDocumentPointerDown(event: PointerEvent) {
  if (!pdfStylePanelOpen.value) return
  const target = event.target
  if (!(target instanceof Node)) return
  if (pdfStylePanelRef.value?.contains(target) || pdfStyleButtonRef.value?.contains(target)) return
  pdfStylePanelOpen.value = false
}
function pdfToggleZoomMenu() {
  pdfZoomMenuOpen.value = !pdfZoomMenuOpen.value
  pdfStylePanelOpen.value = false
  pdfShapeMenuOpen.value = false
  pdfMoreMenuOpen.value = false
}
function pdfToggleMoreMenu() {
  pdfMoreMenuOpen.value = !pdfMoreMenuOpen.value
  pdfStylePanelOpen.value = false
  pdfShapeMenuOpen.value = false
  pdfZoomMenuOpen.value = false
  pdfSearchOpen.value = false
  pdfSettingsOpen.value = false
}
function pdfZoomIn() { pdfRef.value?.zoomIn() }
function pdfZoomOut() { pdfRef.value?.zoomOut() }
function pdfSetZoom(value: number) {
  pdfZoomMenuOpen.value = false
  void pdfRef.value?.setZoom?.(value)
}
function pdfFitWidth() {
  pdfZoomMenuOpen.value = false
  void pdfRef.value?.fitWidth?.()
}
function pdfFitPage() {
  pdfZoomMenuOpen.value = false
  void pdfRef.value?.fitPage?.()
}
function pdfUndo() { pdfRef.value?.undoLastStroke(pdfCurPage.value) }
function pdfRedo() { pdfRef.value?.redoLastStroke?.() }
function pdfBeginPageEdit() {
  pdfPageEditing.value = true
  pdfPageInput.value = String(pdfCurPage.value)
}
function pdfCommitPageEdit() {
  const target = Number(pdfPageInput.value)
  if (!Number.isFinite(target)) {
    pdfPageInput.value = String(pdfCurPage.value)
    pdfPageEditing.value = false
    return
  }
  pdfRef.value?.jumpToPage?.(target)
  pdfPageEditing.value = false
}
function pdfGoPage(delta: number) {
  pdfRef.value?.jumpToPage?.(pdfCurPage.value + delta)
}
function pdfCancelPageEdit() {
  pdfPageInput.value = String(pdfCurPage.value)
  pdfPageEditing.value = false
}
function onPDFZoomChange(value: number) { pdfZoom.value = value }
function onPDFPageChange(page: number) {
  pdfCurPage.value = page
  if (!pdfPageEditing.value) pdfPageInput.value = String(page)
}
function onPDFPageCountChange(count: number) { pdfPageCount.value = count }
function onPDFHistoryStateChange(state: { canUndo: boolean; canRedo: boolean }) {
  pdfCanUndo.value = state.canUndo
  pdfCanRedo.value = state.canRedo
}
function onPDFSavingStateChange(saving: boolean) {
  pdfSaving.value = saving
}
function onPDFExportStateChange(exporting: boolean) {
  pdfExporting.value = exporting
}
function onPDFTextSelectionChange(drawing: TextDrawing | null) {
  pdfSelectedText.value = drawing
  if (!drawing) return
  pdfPenColor.value = drawing.color
  pdfCustomColorInput.value = drawing.color
  pdfTextSize.value = drawing.fontSize
}
function pdfApplySelectedTextStyle() {
  if (!pdfSelectedText.value) return
  pdfRef.value?.applySelectedTextStyle?.({
    color: pdfPenColor.value,
    fontSize: pdfTextSize.value,
  })
}
function normalizeHexColor(value: string) {
  const trimmed = value.trim().replace(/^#?/, '#').toUpperCase()
  const shortHex = /^#([0-9A-F]{3})$/.exec(trimmed)
  if (shortHex) {
    const [, hex] = shortHex
    return `#${hex[0]}${hex[0]}${hex[1]}${hex[1]}${hex[2]}${hex[2]}`
  }
  return /^#[0-9A-F]{6}$/.test(trimmed) ? trimmed : null
}
function pdfIsActiveColor(color: string) {
  return normalizeHexColor(color) === normalizeHexColor(pdfPenColor.value)
}
function pdfRememberColor(color: string) {
  const normalized = normalizeHexColor(color)
  if (!normalized) return
  pdfRecentColors.value = [
    normalized,
    ...pdfRecentColors.value.filter((item) => normalizeHexColor(item) !== normalized),
  ].slice(0, 6)
}
function pdfSetColor(color: string) {
  const normalized = normalizeHexColor(color)
  if (!normalized) return
  pdfPenColor.value = normalized
  pdfCustomColorInput.value = normalized
  pdfRememberColor(normalized)
  pdfApplySelectedTextStyle()
}
function pdfCommitCustomColor() {
  const normalized = normalizeHexColor(pdfCustomColorInput.value)
  if (!normalized) {
    pdfCustomColorInput.value = pdfPenColor.value
    return
  }
  pdfSetColor(normalized)
}
async function pdfExportAnnotated() {
  pdfMoreMenuOpen.value = false
  try {
    await pdfRef.value?.exportAnnotatedPDF?.()
  } catch (e: any) {
    window.alert(e?.message || '导出失败，请稍后再试')
  }
}
function pdfOpenSearch() {
  pdfMoreMenuOpen.value = false
  pdfStylePanelOpen.value = false
  pdfShapeMenuOpen.value = false
  pdfZoomMenuOpen.value = false
  pdfSettingsOpen.value = false
  pdfSearchOpen.value = true
}
async function pdfRunSearch() {
  const query = pdfSearchQuery.value.trim()
  if (!query) {
    pdfSearchResults.value = []
    return
  }
  pdfSearchLoading.value = true
  try {
    pdfSearchResults.value = await pdfRef.value?.searchDocument?.(query) || []
  } catch (e: any) {
    console.warn('搜索 PDF 失败:', e?.message || e)
    pdfSearchResults.value = []
  } finally {
    pdfSearchLoading.value = false
  }
}
function pdfJumpToSearchResult(page: number) {
  pdfRef.value?.jumpToPage?.(page)
  pdfSearchOpen.value = false
}
function pdfOpenSettings() {
  pdfMoreMenuOpen.value = false
  pdfStylePanelOpen.value = false
  pdfShapeMenuOpen.value = false
  pdfZoomMenuOpen.value = false
  pdfSearchOpen.value = false
  pdfSettingsOpen.value = true
}
function setReaderToolbarAutoHide(value: boolean) {
  readerToolbarAutoHide.value = value
  if (!value) topbarHidden.value = false
}

function toggleAiPanel() {
  aiPanelOpen.value = !aiPanelOpen.value
  // Close other panels
  pdfStylePanelOpen.value = false
  pdfShapeMenuOpen.value = false
  pdfZoomMenuOpen.value = false
  pdfMoreMenuOpen.value = false
  pdfSearchOpen.value = false
  pdfSettingsOpen.value = false
  panelRightOpen.value = false
}
async function shareDocument() {
  pdfMoreMenuOpen.value = false
  pdfSharing.value = true
  try {
    const url = window.location.href
    const title = doc.value?.title || 'NoteWeb 文档'
    if (navigator.share) {
      await navigator.share({ title, url })
      return
    }
    await navigator.clipboard.writeText(url)
    window.alert('分享链接已复制')
  } catch (e: any) {
    if (e?.name !== 'AbortError') {
      window.alert(e?.message || '分享失败，请稍后再试')
    }
  } finally {
    pdfSharing.value = false
  }
}

function onReaderKeydown(event: KeyboardEvent) {
  if (doc.value?.file_type !== 'pdf') return
  const target = event.target as HTMLElement | null
  if (target && ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName)) return
  if (event.ctrlKey && !event.shiftKey && event.key.toLowerCase() === 'z') {
    event.preventDefault()
    pdfUndo()
    return
  }
  if (
    (event.ctrlKey && event.key.toLowerCase() === 'y') ||
    (event.ctrlKey && event.shiftKey && event.key.toLowerCase() === 'z')
  ) {
    event.preventDefault()
    pdfRedo()
    return
  }
  if (event.ctrlKey && ['=', '+'].includes(event.key)) {
    event.preventDefault()
    pdfZoomIn()
    return
  }
  if (event.ctrlKey && event.key === '-') {
    event.preventDefault()
    pdfZoomOut()
    return
  }
  if (event.ctrlKey && event.key === '0') {
    event.preventDefault()
    pdfSetZoom(100)
    return
  }
  if (event.key === 'Escape') {
    pdfStylePanelOpen.value = false
    pdfShapeMenuOpen.value = false
    pdfZoomMenuOpen.value = false
    pdfMoreMenuOpen.value = false
    pdfSearchOpen.value = false
    pdfSettingsOpen.value = false
    aiPanelOpen.value = false
    pdfCancelPageEdit()
  }
}

// Progress tracking
let progressTimer: ReturnType<typeof setInterval> | null = null
const CONTENT_SAVE_INTERVAL = 3000 // ms
const MIN_OPENED_PROGRESS = 1

const doc = computed(() => documentStore.currentDocument)
const content = computed(() => documentStore.documentContent)
const annotations = computed(() => annotationStore.annotations)

function toggleRight() { panelRightOpen.value = !panelRightOpen.value }
function closePanels() { panelRightOpen.value = false }

// ── AI Assistant: methods ──

/**
 * 处理AI输入
 */
async function handleAiInput() {
  const text = aiInput.value.trim()
  if (!text) return

  aiInput.value = ''

  if (aiMode.value === 'search') {
    await aiStore.search(text)
  } else {
    const docId = capturedDocId > 0 ? capturedDocId : null
    await aiStore.chat(docId, text, 'chat')
  }
}

/**
 * 生成文档总结
 */
async function handleSummary() {
  if (!capturedDocId) return
  await aiStore.getSummary(capturedDocId)
}

/**
 * 联网搜索
 */
async function handleSearch() {
  aiMode.value = 'search'
  aiInput.value = ''
  await nextTick()
  document.querySelector<HTMLInputElement>('.ai-panel-input input')?.focus()
}

watch(
  () => aiStore.messages.map((message) => message.content).join('\u0000'),
  async () => {
    await nextTick()
    if (aiMessagesRef.value) {
      aiMessagesRef.value.scrollTop = aiMessagesRef.value.scrollHeight
    }
  },
  { flush: 'post' },
)

// ── Annotation: selection & toolbar ──

async function onAnnoHighlight() {
  const sel = window.getSelection()
  if (!sel) return
  const text = sel.toString().trim()
  if (!text) return
  try {
    await annotationStore.create({
      document_id: capturedDocId,
      page: 1,
      selected_text: text,
      color: '#FFD700',
      type: 'highlight',
      position_data: {},
    })
    sel.removeAllRanges()
  } catch (e: any) {
    console.warn('创建批注失败:', e?.message || e)
  }
}

async function onAnnoNote() {
  const sel = window.getSelection()
  if (!sel) return
  const text = sel.toString().trim()
  if (!text) return
  const note = window.prompt('请输入批注内容：', '')
  if (note === null) return
  try {
    await annotationStore.create({
      document_id: capturedDocId,
      page: 1,
      selected_text: text,
      color: '#93C5FD',
      type: 'highlight',
      note: note || '',
      position_data: {},
    })
    sel.removeAllRanges()
  } catch (e: any) {
    console.warn('创建批注失败:', e?.message || e)
  }
}

async function deleteAnnotation(id: number) {
  try {
    await annotationStore.remove(id)
  } catch (e: any) {
    console.warn('删除批注失败:', e?.message || e)
  }
}

function scrollToAnnotation(text: string) {
  const marks = document.querySelectorAll('mark')
  marks.forEach((m) => {
    if (m.textContent === text) {
      m.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  })
}

function annotationLabel(annotation: Annotation): string {
  if (annotation.selected_text) return annotation.selected_text
  if (annotation.type !== 'drawing') return annotation.note || '文本批注'
  const position = annotation.position_data || {}
  if (position.tool === 'text' && typeof position.text === 'string') return position.text
  if (position.tool === 'shape') {
    const shapeNames: Record<string, string> = {
      line: '直线批注',
      arrow: '箭头批注',
      rectangle: '矩形批注',
      ellipse: '椭圆批注',
    }
    return shapeNames[String(position.shapeType)] || '形状批注'
  }
  if (position.tool === 'highlighter') return '手写高亮'
  return '手写批注'
}

function annotationTypeLabel(annotation: Annotation): string {
  if (annotation.type === 'highlight') return '高亮'
  if (annotation.type === 'comment') return '批注'
  if (annotation.type === 'underline') return '下划线'
  if (annotation.type !== 'drawing') return annotation.type
  const position = annotation.position_data || {}
  if (position.tool === 'text') return '文本'
  if (position.tool === 'shape') return '形状'
  if (position.tool === 'highlighter') return '高亮笔'
  return '画笔'
}

function goToAnnotation(annotation: Annotation) {
  if (doc.value?.file_type === 'pdf') {
    panelRightOpen.value = false
    void pdfRef.value?.focusAnnotation?.(annotation.id, annotation.page || 1)
    return
  }
  scrollToAnnotation(annotation.selected_text)
}

function clampProgress(progress: number) {
  return Math.max(MIN_OPENED_PROGRESS, Math.min(100, progress))
}

function getTextReaderElement() {
  return document.querySelector<HTMLElement>('.mde-canvas .mde-prose') ||
    document.querySelector<HTMLElement>('.mde-main')
}

function getReaderViewportHeight() {
  const topbar = document.querySelector<HTMLElement>('.mde-topbar')
  const search = document.querySelector<HTMLElement>('.mde-search')
  const chromeHeight = (topbar?.offsetHeight || 0) + (search?.offsetHeight || 0)
  return Math.max(1, window.innerHeight - chromeHeight)
}

/** Calculate reading progress based on the visible text document, not the whole page chrome. */
function calcScrollProgress(): number {
  const reader = getTextReaderElement()
  if (!reader) return MIN_OPENED_PROGRESS

  const rect = reader.getBoundingClientRect()
  const readerTop = rect.top + window.scrollY
  const readerHeight = Math.max(reader.scrollHeight, rect.height)
  const viewportHeight = getReaderViewportHeight()

  if (readerHeight <= viewportHeight + 1) return 100

  const scrollableDistance = readerHeight - viewportHeight
  const topbar = document.querySelector<HTMLElement>('.mde-topbar')
  const search = document.querySelector<HTMLElement>('.mde-search')
  const chromeHeight = (topbar?.offsetHeight || 0) + (search?.offsetHeight || 0)
  const progress = ((window.scrollY - Math.max(0, readerTop - chromeHeight)) / scrollableDistance) * 100
  return clampProgress(Math.round(progress))
}

/** Save current scroll-based progress to backend (only if higher than last saved) */
let lastSavedTextProgress = 0

function saveTextProgress() {
  const nextProgress = calcScrollProgress()
  // Only update if progress has increased (prevents opening a doc from resetting progress)
  if (nextProgress > lastSavedTextProgress) {
    lastSavedTextProgress = nextProgress
    console.log(`[Reader] 💾 Saving progress: ${nextProgress}% (scroll ${window.scrollY} / ${document.documentElement.scrollHeight - window.innerHeight})`)
    void documentStore.updateProgress(capturedDocId, nextProgress)
  }
}

// For PDF: debounce progress saves so we don't hammer the API
let pdfSaveTimer: ReturnType<typeof setTimeout> | null = null

function onPDFProgress(progress: number) {
  if (pdfSaveTimer) return // debounce: only save the latest value on unmount
  pdfSaveTimer = setTimeout(() => {
    pdfSaveTimer = null
  }, 200)
  // Store the latest progress for final save on unmount
  lastPDFProgress = progress
}

let lastPDFProgress = 0

function handleScroll() {
  const cur = window.scrollY
  topbarHidden.value = readerToolbarAutoHide.value && cur > 200 && cur > lastScroll
  lastScroll = cur
}

onMounted(async () => {
  window.addEventListener('scroll', handleScroll)
  window.addEventListener('keydown', onReaderKeydown)
  document.addEventListener('pointerdown', onDocumentPointerDown, true)
  capturedDocId = docId.value // lock in the id before any redirect may happen
  try {
    await Promise.all([
      documentStore.fetchDocument(capturedDocId),
      documentStore.fetchDocumentContent(capturedDocId),
      annotationStore.fetchAnnotations(capturedDocId),
    ])
  } catch {
    // document not found — redirect
    router.push('/documents')
  }
  loading.value = false

  // Start periodic progress saving for non-PDF files
  if (doc.value?.file_type !== 'pdf') {
    // Seed lastSavedTextProgress with the server's stored progress so we don't regress
    lastSavedTextProgress = doc.value?.read_progress || 0
    console.log(`[Reader] 📖 Opened doc #${capturedDocId} (type=${doc.value?.file_type}), server progress=${lastSavedTextProgress}%`)
    
    // Mark as opened (1%) immediately if no progress yet
    if (lastSavedTextProgress < 1) {
      await documentStore.markAsRead(capturedDocId)
      lastSavedTextProgress = 1
    }
    
    // Check if document fits in viewport (short document) - auto mark as 100%
    setTimeout(() => {
      const progress = calcScrollProgress()
      if (progress >= 100) {
        console.log(`[Reader] 📄 Short document detected, auto-marking as 100%`)
        lastSavedTextProgress = 100
        void documentStore.updateProgress(capturedDocId, 100)
      }
    }, 100)
    
    // Save periodically while reading
    progressTimer = setInterval(saveTextProgress, CONTENT_SAVE_INTERVAL)
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', handleScroll)
  window.removeEventListener('keydown', onReaderKeydown)
  document.removeEventListener('pointerdown', onDocumentPointerDown, true)
  // Final progress save
  if (progressTimer) {
    clearInterval(progressTimer)
    progressTimer = null
  }
  if (doc.value?.file_type === 'pdf') {
    if (lastPDFProgress > 0) {
      void documentStore.updateProgress(capturedDocId, lastPDFProgress)
    }
  } else {
    saveTextProgress()
  }
})
</script>

<template>
  <div class="reader-page">
    <!-- Overlay -->
    <div :class="['panel-overlay', { show: panelRightOpen }]" @click="closePanels"></div>

    <!-- Floating Top Bar (PDF only; text documents use their editor toolbar) -->
    <div
      v-if="doc?.file_type === 'pdf'"
      :class="['reader-topbar', { hidden: topbarHidden, 'eraser-active': pdfActiveTool === 'eraser' }]"
    >
      <button class="tb-btn" title="返回" aria-label="返回" @click="router.push('/documents')">
        <ArrowLeft />
      </button>
      <div class="tb-divider"></div>
      <!-- AI Assistant button (all document types) -->
      <button :class="['tb-btn', { active: aiPanelOpen }]" title="AI助手" aria-label="AI助手" @click="toggleAiPanel">
        <Sparkles />
      </button>
      <!-- Annotations button (PDF only) -->
      <button v-if="doc?.file_type === 'pdf'" class="tb-btn" title="批注列表" aria-label="批注列表" @click="toggleRight">
        <MessageSquare />
      </button>

      <!-- PDF drawing tools -->
      <template v-if="doc?.file_type === 'pdf'">
        <div class="tb-divider"></div>
        <button :class="['tb-btn', { active: pdfActiveTool === 'pen' }]" title="画笔" aria-label="画笔" @click="pdfSwitchTool('pen')">
          <PenLine />
        </button>
        <button :class="['tb-btn', { active: pdfActiveTool === 'highlighter' }]" title="高亮" aria-label="高亮" @click="pdfSwitchTool('highlighter')">
          <Highlighter />
        </button>
        <button :class="['tb-btn', { active: pdfActiveTool === 'text' }]" title="添加文本" aria-label="添加文本" @click="pdfSwitchTool('text')">
          <Type />
        </button>
        <button :class="['tb-btn', { active: pdfActiveTool === 'select' }]" title="选择批注" aria-label="选择批注" @click="pdfSwitchTool('select')">
          <MousePointer2 />
        </button>
        <div class="shape-tool-wrap">
          <button :class="['tb-btn', { active: pdfActiveTool === 'shape' }]" title="形状" aria-label="形状" @click="pdfToggleShapeMenu">
            <Square />
          </button>
          <div v-if="pdfShapeMenuOpen" class="shape-menu">
            <button :class="{ active: pdfShapeType === 'line' }" @click="pdfChooseShape('line')">
              <Minus /><span>直线</span>
            </button>
            <button :class="{ active: pdfShapeType === 'arrow' }" @click="pdfChooseShape('arrow')">
              <ArrowRight /><span>箭头</span>
            </button>
            <button :class="{ active: pdfShapeType === 'rectangle' }" @click="pdfChooseShape('rectangle')">
              <Square /><span>矩形</span>
            </button>
            <button :class="{ active: pdfShapeType === 'ellipse' }" @click="pdfChooseShape('ellipse')">
              <Circle /><span>椭圆</span>
            </button>
          </div>
        </div>
        <button :class="['tb-btn', { active: pdfActiveTool === 'eraser' }]" title="橡皮擦" aria-label="橡皮擦" @click="pdfSwitchTool('eraser')">
          <Eraser />
        </button>
        <button
          v-if="pdfUsesStyle"
          ref="pdfStyleButtonRef"
          :class="['tb-btn', { active: pdfStylePanelOpen }]"
          title="样式设置"
          aria-label="样式设置"
          @click="pdfToggleStylePanel"
        >
          <Palette />
        </button>
        <div class="tb-popover-wrap">
          <div v-if="pdfStylePanelOpen" ref="pdfStylePanelRef" class="tb-popover tb-style-panel">
            <div class="tb-popover__header">
              <strong>{{ pdfActiveToolLabel }}样式</strong>
              <span v-if="pdfShowsColorControls">{{ pdfPenColor }}</span>
            </div>
            <div v-if="pdfShowsColorControls" class="color-picker">
              <div class="color-picker__current">
                <span class="color-picker__preview" :style="{ background: pdfPenColor }"></span>
                <span class="color-picker__value">{{ pdfPenColor }}</span>
              </div>

              <div class="color-section">
                <div class="color-section__title">主题颜色</div>
                <div class="color-grid">
                  <button
                    v-for="color in pdfThemeColors"
                    :key="color"
                    type="button"
                    :class="['color-chip', { active: pdfIsActiveColor(color) }]"
                    :style="{ background: color }"
                    :aria-label="`选择颜色 ${color}`"
                    @click="pdfSetColor(color)"
                  ></button>
                </div>
              </div>

              <div class="color-section">
                <div class="color-section__title">高亮颜色</div>
                <div class="color-grid color-grid--soft">
                  <button
                    v-for="color in pdfHighlightColors"
                    :key="color"
                    type="button"
                    :class="['color-chip color-chip--soft', { active: pdfIsActiveColor(color) }]"
                    :style="{ background: color }"
                    :aria-label="`选择高亮颜色 ${color}`"
                    @click="pdfSetColor(color)"
                  ></button>
                </div>
              </div>

              <div class="color-section">
                <div class="color-section__title">最近使用</div>
                <div class="color-grid color-grid--recent">
                  <button
                    v-for="color in pdfRecentColors"
                    :key="color"
                    type="button"
                    :class="['color-chip', { active: pdfIsActiveColor(color) }]"
                    :style="{ background: color }"
                    :aria-label="`选择最近颜色 ${color}`"
                    @click="pdfSetColor(color)"
                  ></button>
                </div>
              </div>

              <label class="custom-color">
                <span>自定义</span>
                <input
                  v-model="pdfCustomColorInput"
                  type="text"
                  inputmode="text"
                  maxlength="7"
                  spellcheck="false"
                  placeholder="#FF0000"
                  @change="pdfCommitCustomColor"
                  @keydown.enter.prevent="pdfCommitCustomColor"
                />
                <button type="button" @click="pdfCommitCustomColor">应用</button>
              </label>
            </div>
            <label v-if="pdfShowsTextControls" class="tool-field">
              <span>字号 {{ pdfTextSize }}px</span>
              <input type="range" v-model.number="pdfTextSize" min="10" max="72" @change="pdfApplySelectedTextStyle" />
            </label>
            <label v-else-if="pdfActiveTool !== 'eraser'" class="tool-field">
              <span>线宽 {{ pdfPenWidth }}px</span>
              <input type="range" v-model.number="pdfPenWidth" min="1" max="20" />
            </label>
            <label v-if="pdfActiveTool === 'eraser'" class="tool-field">
              <span>橡皮 {{ pdfEraserSize }}px</span>
              <input type="range" v-model.number="pdfEraserSize" min="8" max="48" />
            </label>
            <button v-if="pdfActiveTool === 'eraser'" class="panel-action" @click="pdfToggleEraseMode">
              {{ pdfEraseMode === 'freehand' ? '切换为框选擦除' : '切换为自由擦除' }}
            </button>
          </div>
        </div>
        <div class="tb-divider"></div>
        <button class="tb-btn" title="撤销" aria-label="撤销" :disabled="!pdfCanUndo" @click="pdfUndo">
          <Undo2 />
        </button>
        <button class="tb-btn" title="重做" aria-label="重做" :disabled="!pdfCanRedo" @click="pdfRedo">
          <Redo2 />
        </button>
        <div class="tb-divider"></div>
        <button class="tb-btn tb-btn--compact" title="上一页" aria-label="上一页" :disabled="pdfCurPage <= 1" @click="pdfGoPage(-1)">
          <ChevronLeft />
        </button>
        <form v-if="pdfPageEditing" class="page-jump" @submit.prevent="pdfCommitPageEdit">
          <input
            v-model="pdfPageInput"
            type="number"
            min="1"
            :max="pdfPageCount"
            aria-label="跳转页码"
            @blur="pdfCommitPageEdit"
            @keydown.esc.prevent="pdfCancelPageEdit"
          />
          <span>/{{ pdfPageCount }}</span>
        </form>
        <button v-else class="tb-page-label" title="跳转页码" aria-label="跳转页码" @click="pdfBeginPageEdit">
          {{ pdfCurPage }}/{{ pdfPageCount }}
        </button>
        <button class="tb-btn tb-btn--compact" title="下一页" aria-label="下一页" :disabled="pdfCurPage >= pdfPageCount" @click="pdfGoPage(1)">
          <ChevronRight />
        </button>
        <div class="tb-divider"></div>
        <button class="tb-btn" title="缩小" aria-label="缩小" @click="pdfZoomOut">
          <ZoomOut />
        </button>
        <div class="tb-popover-wrap">
          <button class="tb-zoom-label" title="缩放菜单" aria-label="缩放菜单" @click="pdfToggleZoomMenu">
            {{ pdfZoom }}%
          </button>
          <div v-if="pdfZoomMenuOpen" class="tb-popover tb-zoom-menu">
            <button @click="pdfFitWidth"><Maximize />适配宽度</button>
            <button @click="pdfFitPage"><FileText />适配页面</button>
            <button
              v-for="value in [50, 75, 100, 125, 150, 200]"
              :key="value"
              :class="{ active: pdfZoom === value }"
              @click="pdfSetZoom(value)"
            >
              {{ value }}%
            </button>
          </div>
        </div>
        <button class="tb-btn" title="放大" aria-label="放大" @click="pdfZoomIn">
          <ZoomIn />
        </button>
        <div class="tb-divider"></div>
        <button class="tb-status" :title="pdfSaveStatusLabel" aria-label="保存状态">
          <LoaderCircle v-if="pdfSaving" class="spin" />
          <CheckCircle2 v-else />
          <span>{{ pdfSaveStatusLabel }}</span>
        </button>
        <div class="tb-popover-wrap">
          <button class="tb-btn" title="更多" aria-label="更多" @click="pdfToggleMoreMenu">
            <MoreHorizontal />
          </button>
          <div v-if="pdfMoreMenuOpen" class="tb-popover tb-more-menu">
            <button @click="toggleRight"><MessageSquare />批注列表</button>
            <button @click="pdfOpenSearch"><Search />搜索文档</button>
            <button :disabled="pdfExporting || pdfSaving" @click="pdfExportAnnotated">
              <LoaderCircle v-if="pdfExporting" class="spin" />
              <Download v-else />
              {{ pdfExporting ? '导出中...' : '导出带批注 PDF' }}
            </button>
            <button :disabled="pdfSharing" @click="shareDocument">
              <LoaderCircle v-if="pdfSharing" class="spin" />
              <Share2 v-else />
              {{ pdfSharing ? '分享中...' : '分享' }}
            </button>
            <button @click="pdfOpenSettings"><Settings />阅读设置</button>
          </div>
        </div>

        <div v-if="pdfSearchOpen" class="tb-search-panel">
          <div class="search-box">
            <Search />
            <input
              v-model="pdfSearchQuery"
              type="search"
              placeholder="搜索 PDF 文本"
              @keydown.enter.prevent="pdfRunSearch"
            />
            <button type="button" :disabled="pdfSearchLoading" @click="pdfRunSearch">
              {{ pdfSearchLoading ? '搜索中' : '搜索' }}
            </button>
          </div>
          <div class="search-results">
            <div v-if="!pdfSearchQuery.trim()" class="search-empty">输入关键词后搜索文档内容</div>
            <div v-else-if="!pdfSearchLoading && pdfSearchResults.length === 0" class="search-empty">没有找到匹配内容</div>
            <button
              v-for="result in pdfSearchResults"
              :key="`${result.page}-${result.excerpt}`"
              type="button"
              class="search-result"
              @click="pdfJumpToSearchResult(result.page)"
            >
              <span>第 {{ result.page }} 页</span>
              <strong>{{ result.excerpt }}</strong>
            </button>
          </div>
        </div>

        <div v-if="pdfSettingsOpen" class="tb-settings-panel">
          <div class="settings-panel__header">
            <strong>阅读设置</strong>
            <span>PDF</span>
          </div>
          <label class="settings-toggle">
            <span>
              <strong>滚动时隐藏工具栏</strong>
              <small>阅读时自动收起顶部工具栏</small>
            </span>
            <input
              type="checkbox"
              :checked="readerToolbarAutoHide"
              @change="setReaderToolbarAutoHide(($event.target as HTMLInputElement).checked)"
            />
          </label>
          <div class="settings-grid">
            <button type="button" @click="pdfFitWidth">适配宽度</button>
            <button type="button" @click="pdfFitPage">适配页面</button>
            <button type="button" @click="pdfSetZoom(100)">100%</button>
            <button type="button" @click="toggleRight">批注侧栏</button>
          </div>
        </div>
      </template>
    </div>

    <!-- AI Panel Popover (positioned independently) -->
    <div v-if="aiPanelOpen" class="ai-panel-overlay" @click="aiPanelOpen = false"></div>
    <div v-if="aiPanelOpen" :class="['tb-ai-panel', { 'tb-ai-panel--editor': doc?.file_type !== 'pdf' }]">
      <div class="ai-panel-header">
        <Sparkles :size="16" />
        <span>AI 助手</span>
      </div>
      <!-- Chat messages -->
      <div ref="aiMessagesRef" class="ai-panel-messages">
        <div v-if="messages.length === 0" class="ai-panel-empty">
          <p>AI 助手已就绪</p>
          <p class="hint">询问文档问题、搜索互联网或生成总结</p>
        </div>
        <div v-for="(msg, index) in messages" :key="index" :class="['ai-panel-msg', `ai-panel-msg--${msg.role}`]">
          <div class="ai-panel-msg__content">{{ msg.content }}</div>
        </div>
        <div v-if="aiStore.loading" class="ai-panel-loading">
          <LoaderCircle class="spin" :size="14" />
          <span>思考中...</span>
        </div>
        <div v-if="aiStore.error" class="ai-panel-error">
          {{ aiStore.error }}
        </div>
      </div>
      <!-- Quick actions -->
      <div class="ai-panel-actions">
        <button v-if="doc" @click="handleSummary" :disabled="aiStore.loading" class="ai-panel-action">
          <Sparkles :size="14" />
          <span>生成总结</span>
        </button>
        <button @click="handleSearch" :disabled="aiStore.loading" class="ai-panel-action">
          <Search :size="14" />
          <span>联网搜索</span>
        </button>
      </div>
      <!-- Input area -->
      <div class="ai-panel-input">
        <input
          v-model="aiInput"
          type="text"
          :placeholder="aiInputPlaceholder"
          @keydown.enter.exact.prevent="handleAiInput"
        />
        <button :disabled="aiStore.loading || !aiInput.trim()" @click="handleAiInput">
          <Send v-if="!aiStore.loading" :size="14" />
          <LoaderCircle v-else class="spin" :size="14" />
        </button>
      </div>
    </div>

    <!-- Left Panel: TOC (placeholder — real TOC needs parsed content) -->
    <div v-if="doc?.file_type === 'pdf'" :class="['panel-left', { open: panelLeftOpen }]">
      <div class="panel__header">
        <h3>目录</h3>
        <button @click="panelLeftOpen = false">✕</button>
      </div>
      <div class="panel__body">
        <div class="toc-item level-1 active">{{ doc?.title || '文档' }}</div>
        <div class="toc-placeholder">目录解析暂未实现</div>
      </div>
    </div>

    <!-- Right Panel: Annotations (PDF only) -->
    <div v-if="doc?.file_type === 'pdf'" :class="['panel-right', { open: panelRightOpen }]">
      <div class="panel__header">
        <h3>批注列表</h3>
        <button @click="panelRightOpen = false">✕</button>
      </div>
      <div class="panel__body">
        <div v-if="annotations.length === 0" class="anno-empty">选中文本即可添加批注</div>
        <div
          v-for="anno in annotations"
          :key="anno.id"
          class="anno-card"
          @click="goToAnnotation(anno)"
        >
          <div class="anno-card__text">{{ annotationLabel(anno) }}</div>
          <div class="anno-card__meta">
            <span class="anno-highlight" :style="{ background: anno.color || '#FDE68A' }">{{ annotationTypeLabel(anno) }}</span>
            <span>第 {{ anno.page || 1 }} 页</span>
            {{ anno.note ? `· ${anno.note}` : '' }}
          </div>
          <button class="anno-card__del" title="删除" @click.stop="deleteAnnotation(anno.id)">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Right page edge decoration -->
    <div v-if="doc?.file_type === 'pdf'" class="page-edge"></div>

    <!-- Annotation Toolbar -->
    <AnnotationToolbar v-if="doc?.file_type === 'pdf'" @highlight="onAnnoHighlight" @note="onAnnoNote" />

    <!-- Reading Content -->
    <div v-if="loading" class="reader-content">
      <div class="reader-inner"><p style="text-align:center;padding:4rem 0;color:var(--text-muted)">加载中...</p></div>
    </div>

    <div v-else-if="doc?.file_type === 'pdf'" class="reader-pdf-wrapper">
      <PDFViewer
        ref="pdfRef"
        :document-id="capturedDocId"
        :document-title="doc?.title"
        :active-tool="pdfActiveTool"
        :shape-type="pdfShapeType"
        :erase-mode="pdfEraseMode"
        :pen-color="pdfPenColor"
        :pen-width="pdfActiveTool === 'eraser' ? pdfEraserSize : pdfPenWidth"
        :text-size="pdfTextSize"
        @progress="onPDFProgress"
        @zoom-change="onPDFZoomChange"
        @current-page-change="onPDFPageChange"
        @page-count-change="onPDFPageCountChange"
        @history-state-change="onPDFHistoryStateChange"
        @saving-state-change="onPDFSavingStateChange"
        @export-state-change="onPDFExportStateChange"
        @text-selection-change="onPDFTextSelectionChange"
      />
    </div>

    <MarkdownDocumentEditor
      v-else
      :document-id="capturedDocId"
      :title="doc?.title || '未命名文档'"
      :content="content"
      :file-type="doc?.file_type || 'md'"
      :ai-active="aiPanelOpen"
      @back="router.push('/documents')"
      @toggle-ai="toggleAiPanel"
    />
  </div>
</template>

<style scoped>
.reader-page { position: relative; min-height: 100vh; background: var(--bg-page); background-image: repeating-linear-gradient(0deg, transparent, transparent 1px, rgba(0,0,0,0.005) 1px, rgba(0,0,0,0.005) 2px); background-size: 100% 2px; font-family: var(--font-body); }

.panel-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.15); z-index: 15; opacity: 0; pointer-events: none; transition: opacity 0.3s; }
.panel-overlay.show { opacity: 1; pointer-events: auto; }

.reader-topbar { position: fixed; top: 0.75rem; left: 50%; transform: translateX(-50%); z-index: 30; display: flex; align-items: center; gap: 0.28rem; max-width: calc(100vw - 1.5rem); padding: 0.35rem 0.6rem; background: rgba(250,248,245,0.94); backdrop-filter: blur(10px); border: 1px solid var(--border-color); border-radius: 24px; box-shadow: 0 2px 12px rgba(61,46,36,0.08); transition: opacity 0.3s, transform 0.3s; font-family: var(--font-ui); }
.reader-topbar.hidden { opacity: 0; transform: translateX(-50%) translateY(-10px); pointer-events: none; }
.tb-btn { width: 34px; height: 34px; min-width: 34px; border: none; border-radius: 50%; background: transparent; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-secondary); transition: background 0.16s, color 0.16s, opacity 0.16s; touch-action: manipulation; }
.tb-btn:hover { background: var(--accent-light); color: var(--accent); }
.tb-btn.active { background: var(--accent); color: #fff; }
.tb-btn.active:hover { background: var(--accent); opacity: 0.85; }
.tb-btn:disabled { opacity: 0.38; cursor: not-allowed; }
.tb-btn:disabled:hover { background: transparent; color: var(--text-secondary); }
.tb-btn:focus-visible,
.tb-page-label:focus-visible,
.tb-zoom-label:focus-visible,
.tb-status:focus-visible,
.shape-menu button:focus-visible,
.tb-popover button:focus-visible,
.page-jump input:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
.tb-btn svg { width: 18px; height: 18px; }
.tb-btn--compact { width: 28px; min-width: 28px; }
.tb-btn--compact svg { width: 16px; height: 16px; }
.tb-divider { width: 1px; height: 20px; background: var(--border-color); margin: 0 0.2rem; }
.tb-label { font-size: 0.7rem; color: var(--text-muted); padding: 0 0.4rem; }
.tb-page-label,
.tb-zoom-label,
.tb-status { height: 34px; border: none; border-radius: 17px; background: transparent; color: var(--text-secondary); font-family: var(--font-ui); font-size: 0.76rem; display: inline-flex; align-items: center; justify-content: center; gap: 0.35rem; cursor: pointer; transition: background 0.16s, color 0.16s; white-space: nowrap; touch-action: manipulation; }
.tb-page-label { min-width: 58px; padding: 0 0.42rem; }
.tb-zoom-label { min-width: 54px; padding: 0 0.46rem; color: var(--text-muted); }
.tb-status { min-width: 78px; padding: 0 0.55rem; color: var(--text-muted); cursor: default; }
.tb-status svg { width: 15px; height: 15px; }
.tb-page-label:hover,
.tb-zoom-label:hover { background: var(--accent-light); color: var(--accent); }
.page-jump { height: 34px; display: inline-flex; align-items: center; gap: 0.25rem; padding: 0 0.45rem; border-radius: 17px; background: rgba(255,255,255,0.72); border: 1px solid var(--border-color); color: var(--text-muted); font-family: var(--font-ui); font-size: 0.75rem; }
.page-jump input { width: 42px; border: none; background: transparent; color: var(--text-primary); font: inherit; text-align: center; outline: none; }
.tb-popover-wrap { position: relative; display: inline-flex; align-items: center; }
.tb-popover { position: absolute; top: calc(100% + 0.65rem); left: 50%; transform: translateX(-50%); z-index: 35; padding: 0.55rem; background: rgba(250,248,245,0.98); border: 1px solid var(--border-color); border-radius: 10px; box-shadow: 0 8px 24px rgba(61,46,36,0.16); font-family: var(--font-ui); }
.tb-popover__header { display: flex; align-items: center; justify-content: space-between; gap: 1rem; margin-bottom: 0.55rem; color: var(--text-primary); font-size: 0.76rem; }
.tb-popover__header span { color: var(--text-muted); font-size: 0.68rem; }
.tb-style-panel { width: 274px; }
.tool-field { display: grid; gap: 0.35rem; margin-bottom: 0.55rem; color: var(--text-muted); font-size: 0.72rem; }
.tool-field input[type="range"] { width: 100%; accent-color: var(--accent); cursor: pointer; }
.panel-action,
.tb-popover button { min-height: 32px; border: none; border-radius: 7px; background: transparent; color: var(--text-secondary); font-family: var(--font-ui); font-size: 0.75rem; cursor: pointer; }
.panel-action { width: 100%; background: var(--accent-light); color: var(--accent); }
.tb-popover button:hover,
.tb-popover button.active { background: var(--accent-light); color: var(--accent); }
.tb-popover button:disabled { opacity: 0.45; cursor: not-allowed; }
.color-picker { display: grid; gap: 0.55rem; margin-bottom: 0.6rem; }
.color-picker__current { display: flex; align-items: center; gap: 0.45rem; padding: 0.35rem 0.45rem; border: 1px solid var(--border-color); border-radius: 8px; background: rgba(255,255,255,0.55); }
.color-picker__preview { width: 28px; height: 18px; border-radius: 4px; border: 1px solid rgba(61,46,36,0.18); box-shadow: inset 0 0 0 1px rgba(255,255,255,0.38); }
.color-picker__value { color: var(--text-secondary); font-size: 0.72rem; font-variant-numeric: tabular-nums; }
.color-section { display: grid; gap: 0.32rem; }
.color-section__title { color: var(--text-muted); font-size: 0.68rem; }
.color-grid { display: grid; grid-template-columns: repeat(6, 1fr); gap: 0.35rem; }
.color-grid--soft,
.color-grid--recent { grid-template-columns: repeat(6, 1fr); }
.tb-popover .color-chip {
  position: relative;
  width: 100%;
  height: 24px;
  min-height: 24px;
  padding: 0;
  border: 1px solid rgba(61,46,36,0.18);
  border-radius: 6px;
  box-shadow: inset 0 0 0 1px rgba(255,255,255,0.32);
  transition: transform 0.12s, border-color 0.12s, box-shadow 0.12s;
}
.tb-popover .color-chip:hover { transform: translateY(-1px); border-color: rgba(198,122,78,0.7); box-shadow: 0 3px 8px rgba(61,46,36,0.12), inset 0 0 0 1px rgba(255,255,255,0.38); }
.tb-popover .color-chip.active { border-color: var(--accent); box-shadow: 0 0 0 2px rgba(198,122,78,0.22), inset 0 0 0 2px rgba(255,255,255,0.86); }
.tb-popover .color-chip.active::after {
  content: "";
  position: absolute;
  left: 50%;
  top: 50%;
  width: 7px;
  height: 4px;
  border-left: 2px solid #fff;
  border-bottom: 2px solid #fff;
  transform: translate(-50%, -60%) rotate(-45deg);
  filter: drop-shadow(0 1px 1px rgba(0,0,0,0.45));
}
.tb-popover .color-chip--soft.active::after { border-color: #3d2e24; filter: none; }
.custom-color { display: grid; grid-template-columns: auto 1fr auto; align-items: center; gap: 0.45rem; color: var(--text-muted); font-size: 0.68rem; }
.custom-color input { min-width: 0; height: 30px; border: 1px solid var(--border-color); border-radius: 7px; background: rgba(255,255,255,0.66); color: var(--text-primary); font-family: var(--font-ui); font-size: 0.74rem; padding: 0 0.55rem; outline: none; text-transform: uppercase; }
.custom-color input:focus { border-color: var(--accent); box-shadow: 0 0 0 2px rgba(198,122,78,0.16); }
.custom-color button { min-height: 30px; padding: 0 0.65rem; background: var(--accent-light); color: var(--accent); }
.tb-zoom-menu,
.tb-more-menu { display: grid; gap: 0.25rem; min-width: 150px; }
.tb-zoom-menu button,
.tb-more-menu button { display: flex; align-items: center; gap: 0.45rem; padding: 0 0.55rem; text-align: left; }
.tb-zoom-menu svg,
.tb-more-menu svg { width: 15px; height: 15px; flex-shrink: 0; }
.tb-search-panel {
  position: absolute;
  top: calc(100% + 0.65rem);
  right: 0;
  z-index: 36;
  width: min(420px, calc(100vw - 1.5rem));
  padding: 0.6rem;
  background: rgba(250,248,245,0.98);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(61,46,36,0.16);
  font-family: var(--font-ui);
}
.search-box {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 0.45rem;
  padding: 0.35rem 0.45rem;
  border: 1px solid var(--border-color);
  border-radius: 9px;
  background: rgba(255,255,255,0.7);
}
.search-box svg { width: 16px; height: 16px; color: var(--text-muted); }
.search-box input {
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  color: var(--text-primary);
  font: inherit;
  font-size: 0.78rem;
}
.search-box button {
  min-height: 28px;
  border: none;
  border-radius: 7px;
  padding: 0 0.65rem;
  background: var(--accent-light);
  color: var(--accent);
  font: inherit;
  font-size: 0.74rem;
  cursor: pointer;
}
.search-box button:disabled { opacity: 0.45; cursor: wait; }
.search-results {
  display: grid;
  gap: 0.35rem;
  max-height: min(320px, 46vh);
  overflow-y: auto;
  margin-top: 0.55rem;
}
.search-empty {
  padding: 1.1rem 0.5rem;
  color: var(--text-muted);
  text-align: center;
  font-size: 0.78rem;
}
.search-result {
  display: grid;
  gap: 0.2rem;
  width: 100%;
  min-height: 0;
  padding: 0.55rem 0.6rem;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--text-secondary);
  text-align: left;
  cursor: pointer;
}
.search-result:hover { background: var(--accent-light); color: var(--accent); }
.search-result span { color: var(--text-muted); font-size: 0.68rem; }
.search-result strong {
  color: inherit;
  font-size: 0.78rem;
  font-weight: 500;
  line-height: 1.45;
}
.tb-settings-panel {
  position: absolute;
  top: calc(100% + 0.65rem);
  right: 0;
  z-index: 36;
  width: min(320px, calc(100vw - 1.5rem));
  padding: 0.65rem;
  background: rgba(250,248,245,0.98);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(61,46,36,0.16);
  font-family: var(--font-ui);
}

/* AI Panel Overlay */
.ai-panel-overlay {
  position: fixed;
  inset: 0;
  z-index: 29;
}

/* AI Panel Styles */
.tb-ai-panel {
  position: fixed;
  top: 3.5rem;
  left: 50%;
  transform: translateX(-50%);
  z-index: 36;
  width: min(380px, calc(100vw - 1.5rem));
  padding: 0.65rem;
  background: rgba(250,248,245,0.98);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(61,46,36,0.16);
  font-family: var(--font-ui);
}
.tb-ai-panel--editor {
  top: 4.25rem;
  left: auto;
  right: 1rem;
  transform: none;
}
.ai-panel-header {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin-bottom: 0.55rem;
  color: var(--accent);
  font-size: 0.78rem;
  font-weight: 500;
}
.ai-panel-messages {
  max-height: min(280px, 40vh);
  overflow-y: auto;
  margin-bottom: 0.55rem;
}
.ai-panel-empty {
  padding: 1rem 0.5rem;
  color: var(--text-muted);
  text-align: center;
  font-size: 0.78rem;
}
.ai-panel-empty .hint {
  font-size: 0.7rem;
  opacity: 0.7;
  margin-top: 0.25rem;
}
.ai-panel-msg {
  padding: 0.5rem 0.6rem;
  border-radius: 8px;
  margin-bottom: 0.4rem;
  font-size: 0.78rem;
  line-height: 1.5;
  max-width: 90%;
  word-wrap: break-word;
}
.ai-panel-msg--user {
  background: var(--accent-light);
  color: var(--text-primary);
  margin-left: auto;
}
.ai-panel-msg--assistant {
  background: var(--bg-page);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}
.ai-panel-msg__content {
  white-space: pre-wrap;
}
.ai-panel-loading {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem;
  color: var(--text-muted);
  font-size: 0.75rem;
}
.ai-panel-error {
  padding: 0.55rem 0.65rem;
  border: 1px solid rgba(180, 35, 24, 0.28);
  border-radius: 8px;
  background: rgba(180, 35, 24, 0.06);
  color: #b42318;
  font-size: 0.76rem;
  line-height: 1.5;
}
.ai-panel-actions {
  display: flex;
  gap: 0.35rem;
  margin-bottom: 0.55rem;
}
.ai-panel-action {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
  padding: 0.4rem 0.6rem;
  border: 1px solid var(--border-color);
  border-radius: 7px;
  background: var(--bg-page);
  color: var(--text-secondary);
  font-size: 0.72rem;
  cursor: pointer;
  transition: all 0.15s;
}
.ai-panel-action:hover:not(:disabled) {
  background: var(--accent-light);
  color: var(--accent);
  border-color: var(--accent);
}
.ai-panel-action:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.ai-panel-input {
  display: flex;
  gap: 0.35rem;
}
.ai-panel-input input {
  flex: 1;
  padding: 0.45rem 0.7rem;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-page);
  color: var(--text-primary);
  font-size: 0.78rem;
  outline: none;
}
.ai-panel-input input:focus {
  border-color: var(--accent);
}
.ai-panel-input button {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: var(--accent);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: opacity 0.15s;
}
.ai-panel-input button:hover:not(:disabled) {
  opacity: 0.85;
}
.ai-panel-input button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.settings-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.6rem;
  color: var(--text-primary);
  font-size: 0.78rem;
}
.settings-panel__header span {
  color: var(--text-muted);
  font-size: 0.68rem;
}
.settings-toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.8rem;
  padding: 0.55rem 0.6rem;
  border: 1px solid var(--border-color);
  border-radius: 9px;
  background: rgba(255,255,255,0.58);
  color: var(--text-primary);
}
.settings-toggle span {
  display: grid;
  gap: 0.16rem;
}
.settings-toggle strong {
  font-size: 0.76rem;
  font-weight: 500;
}
.settings-toggle small {
  color: var(--text-muted);
  font-size: 0.68rem;
}
.settings-toggle input {
  width: 34px;
  height: 20px;
  accent-color: var(--accent);
  cursor: pointer;
}
.settings-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.4rem;
  margin-top: 0.6rem;
}
.settings-grid button {
  min-height: 32px;
  border: none;
  border-radius: 8px;
  background: var(--accent-light);
  color: var(--accent);
  font: inherit;
  font-size: 0.74rem;
  cursor: pointer;
}
.settings-grid button:hover {
  filter: brightness(0.98);
}
.spin { animation: spin 0.9s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.shape-tool-wrap { position: relative; display: flex; }
.shape-menu {
  position: absolute;
  top: calc(100% + 0.65rem);
  left: 50%;
  transform: translateX(-50%);
  display: grid;
  grid-template-columns: repeat(2, minmax(72px, 1fr));
  gap: 0.35rem;
  padding: 0.45rem;
  min-width: 160px;
  background: rgba(250,248,245,0.98);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(61,46,36,0.16);
}
.shape-menu button {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.45rem 0.55rem;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--text-secondary);
  font-family: var(--font-ui);
  font-size: 0.75rem;
  cursor: pointer;
}
.shape-menu button:hover,
.shape-menu button.active { background: var(--accent-light); color: var(--accent); }
.shape-menu svg { width: 18px; height: 18px; fill: none; stroke: currentColor; stroke-width: 1.8; }

.panel-left, .panel-right { position: fixed; top: 0; bottom: 0; width: 300px; background: var(--bg-card); z-index: 20; transition: transform 0.3s; display: flex; flex-direction: column; }
.panel-left { left: 0; border-right: 1px solid var(--border-color); transform: translateX(-100%); }
.panel-left.open { transform: translateX(0); }
.panel-right { right: 0; border-left: 1px solid var(--border-color); transform: translateX(100%); }
.panel-right.open { transform: translateX(0); }
.panel__header { display: flex; align-items: center; justify-content: space-between; padding: 1rem 1.25rem; border-bottom: 1px solid var(--border-color); font-family: var(--font-ui); flex-shrink: 0; }
.panel__header h3 { font-size: 0.9rem; font-weight: 500; color: var(--text-primary); }
.panel__header button { background: none; border: none; cursor: pointer; color: var(--text-muted); font-size: 1.2rem; padding: 0.2rem; }
.panel__header button:hover { color: var(--text-primary); }
.panel__body { flex: 1; overflow-y: auto; padding: 1rem 1.25rem; }

.panel-tabs { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
.pt-btn { padding: 0.3rem 1rem; border-radius: 20px; border: none; font-family: var(--font-ui); font-size: 0.8rem; cursor: pointer; transition: all 0.12s; background: transparent; color: var(--text-muted); }
.pt-btn.active { background: var(--accent-light); color: var(--accent); font-weight: 500; }

.toc-item { padding: 0.5rem 0; cursor: pointer; color: var(--text-secondary); font-family: var(--font-ui); font-size: 0.85rem; border-bottom: 1px solid var(--border-color); }
.toc-item.level-1 { font-weight: 500; }
.toc-placeholder { padding: 1rem 0; color: var(--text-muted); font-family: var(--font-ui); font-size: 0.8rem; text-align: center; }

.anno-card { padding: 0.8rem 0; border-bottom: 1px solid var(--border-color); cursor: pointer; display: flex; align-items: flex-start; gap: 0.5rem; transition: background 0.1s; }
.anno-card:hover { background: var(--accent-light); margin: 0 -0.5rem; padding-left: 0.5rem; padding-right: 0.5rem; border-radius: 6px; }
.anno-card__text { font-size: 0.85rem; line-height: 1.6; color: var(--text-primary); flex: 1; overflow: hidden; text-overflow: ellipsis; display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; }
.anno-card__meta { font-family: var(--font-ui); font-size: 0.7rem; color: var(--text-muted); margin-top: 0.3rem; display: flex; gap: 0.4rem; align-items: center; }
.anno-highlight { display: inline-block; background: #FDE68A; padding: 0 0.3rem; border-radius: 3px; font-size: 0.65rem; color: #92400E; }
.anno-card__del { width: 26px; height: 26px; border: none; border-radius: 6px; background: transparent; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-muted); flex-shrink: 0; margin-top: 2px; transition: all 0.1s; }
.anno-card__del:hover { background: #FEE2E2; color: #DC2626; }
.anno-empty { text-align: center; padding: 2rem 0; color: var(--text-muted); font-family: var(--font-ui); font-size: 0.8rem; }

.ai-message { padding: 0.8rem; background: var(--accent-light); border-radius: 8px; margin-bottom: 0.8rem; font-size: 0.85rem; line-height: 1.7; }
.ai-message__label { font-family: var(--font-ui); font-size: 0.7rem; color: var(--accent); font-weight: 600; margin-bottom: 0.3rem; }
.ai-input { display: flex; gap: 0.5rem; margin-top: 0.5rem; }
.ai-input input { flex: 1; padding: 0.5rem 0.8rem; border: 1px solid var(--border-color); border-radius: 20px; background: var(--bg-page); font-family: var(--font-ui); font-size: 0.8rem; color: var(--text-primary); outline: none; }
.ai-input input:focus { border-color: var(--accent); }

/* AI Panel Styles */
.ai-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 400px;
}

.ai-chat {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.ai-messages {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem 0;
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

.ai-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem 1rem;
  color: var(--text-muted);
  text-align: center;
}

.ai-empty svg {
  width: 32px;
  height: 32px;
  margin-bottom: 0.5rem;
  opacity: 0.5;
}

.ai-empty p {
  margin: 0.2rem 0;
  font-size: 0.85rem;
}

.ai-empty__hint {
  font-size: 0.75rem !important;
  opacity: 0.7;
}

.ai-message {
  padding: 0.6rem 0.8rem;
  border-radius: 8px;
  font-size: 0.85rem;
  line-height: 1.6;
  max-width: 85%;
  word-wrap: break-word;
}

.ai-message--user {
  background: var(--accent-light);
  color: var(--text-primary);
  align-self: flex-end;
  margin-left: 1rem;
}

.ai-message--assistant {
  background: var(--bg-page);
  color: var(--text-primary);
  align-self: flex-start;
  margin-right: 1rem;
  border: 1px solid var(--border-color);
}

.ai-message__content {
  white-space: pre-wrap;
}

.ai-loading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem;
  color: var(--text-muted);
  font-size: 0.85rem;
}

.ai-input-area {
  display: flex;
  gap: 0.5rem;
  padding: 0.8rem 0;
  border-top: 1px solid var(--border-color);
  background: var(--bg-card);
}

.ai-input-wrapper {
  flex: 1;
}

.ai-input-wrapper input {
  width: 100%;
  padding: 0.5rem 0.8rem;
  border: 1px solid var(--border-color);
  border-radius: 20px;
  background: var(--bg-page);
  font-family: var(--font-ui);
  font-size: 0.8rem;
  color: var(--text-primary);
  outline: none;
  transition: border-color 0.2s;
}

.ai-input-wrapper input:focus {
  border-color: var(--accent);
}

.ai-send-btn {
  width: 36px;
  height: 36px;
  min-width: 36px;
  border: none;
  border-radius: 50%;
  background: var(--accent);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: opacity 0.2s;
}

.ai-send-btn:hover:not(.disabled) {
  opacity: 0.85;
}

.ai-send-btn.disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.ai-actions {
  display: flex;
  gap: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid var(--border-color);
}

.ai-action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 0.5rem 0.8rem;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-page);
  color: var(--text-secondary);
  font-family: var(--font-ui);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.ai-action-btn:hover:not(:disabled) {
  background: var(--accent-light);
  color: var(--accent);
  border-color: var(--accent);
}

.ai-action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-edge { position: fixed; top: 0; right: 0; bottom: 0; width: clamp(60px, 8vw, 140px); background: linear-gradient(to right, transparent, rgba(0,0,0,0.015) 40%, rgba(0,0,0,0.025)); pointer-events: none; z-index: 1; }

.reader-content { max-width: 920px; margin: 0 auto; padding: 4.5rem 3rem 6rem; min-height: 100vh; }
.reader-inner { max-width: 680px; margin: 0 auto; }

/* PDF viewer takes full width / height of the reader area */
.reader-pdf-wrapper {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1;
  padding-top: 4rem; /* space for the floating topbar */
}
.doc-title { font-family: var(--font-display); font-size: 2rem; font-weight: 600; line-height: 1.3; margin-bottom: 0.5rem; color: var(--text-primary); }
.doc-meta { font-family: var(--font-ui); font-size: 0.8rem; color: var(--text-muted); margin-bottom: 2rem; padding-bottom: 1rem; border-bottom: 1px solid var(--border-color); }
.doc-body { font-size: 1rem; line-height: 1.9; color: var(--text-primary); }
.doc-body p { margin-bottom: 1rem; }
/* Markdown rendering styles */
.doc-body h1, .doc-body h2, .doc-body h3, .doc-body h4 { font-weight: 600; margin-top: 1.8em; margin-bottom: 0.6em; line-height: 1.3; color: var(--text-primary); }
.doc-body h1 { font-size: 1.6rem; }
.doc-body h2 { font-size: 1.3rem; }
.doc-body h3 { font-size: 1.1rem; }
.doc-body ul, .doc-body ol { padding-left: 1.5em; margin-bottom: 1rem; }
.doc-body li { margin-bottom: 0.25rem; }
.doc-body blockquote { border-left: 3px solid var(--accent); margin: 1rem 0; padding: 0.5rem 1rem; background: var(--accent-light); color: var(--text-secondary); border-radius: 0 6px 6px 0; }
.doc-body code { font-family: 'JetBrains Mono', 'Fira Code', monospace; font-size: 0.85em; background: #f0eee8; padding: 0.15em 0.4em; border-radius: 4px; }
.doc-body pre { background: #f0eee8; padding: 1rem; border-radius: 8px; overflow-x: auto; margin-bottom: 1rem; }
.doc-body pre code { background: none; padding: 0; }
.doc-body a { color: var(--accent); text-decoration: underline; }
.doc-body table { width: 100%; border-collapse: collapse; margin-bottom: 1rem; }
.doc-body th, .doc-body td { padding: 0.5rem 0.75rem; border: 1px solid var(--border-color); text-align: left; }
.doc-body th { background: var(--accent-light); font-weight: 500; }
.doc-body hr { border: none; border-top: 1px solid var(--border-color); margin: 1.5rem 0; }

@media (max-width: 1180px) {
  .reader-topbar { gap: 0.18rem; padding-inline: 0.45rem; }
  .tb-status span { display: none; }
  .tb-status { min-width: 34px; padding: 0; }
}

@media (max-width: 1024px) {
  .reader-content { max-width: 100%; padding: 4.5rem 2rem 5rem; }
  .reader-inner { max-width: 100%; }
  .reader-topbar { top: 0.5rem; max-width: calc(100vw - 1rem); overflow: visible; }
  .tb-divider:nth-of-type(2),
  .tb-btn[aria-label="高亮"],
  .tb-btn[aria-label="添加文本"],
  .tb-btn[aria-label="形状"],
  .tb-status { display: none; }
  .tb-ai-panel { width: calc(100vw - 1rem); max-width: 380px; }
}

@media (max-width: 760px) {
  .reader-topbar { left: 0.5rem; right: 0.5rem; transform: none; justify-content: space-between; border-radius: 22px; }
  .reader-topbar.hidden { transform: translateY(-10px); }
  .tb-btn { width: 38px; height: 38px; min-width: 38px; }
  .tb-btn svg { width: 19px; height: 19px; }
  .tb-btn[aria-label="目录"],
  .tb-btn[aria-label="选择批注"],
  .tb-btn[aria-label="橡皮擦"],
  .tb-btn[aria-label="重做"],
  .tb-btn[aria-label="缩小"],
  .tb-btn[aria-label="放大"],
  .tb-zoom-label { display: none; }
  .tb-page-label { min-width: 64px; }
  .tb-popover { position: fixed; top: 3.8rem; left: 0.75rem; right: 0.75rem; transform: none; width: auto; }
  .tb-ai-panel,
  .tb-ai-panel--editor { left: 0.75rem; right: 0.75rem; transform: none; width: auto; }
  .shape-menu { position: fixed; top: 3.8rem; left: 0.75rem; right: 0.75rem; transform: none; grid-template-columns: repeat(4, minmax(0, 1fr)); min-width: 0; }
  .shape-menu button { justify-content: center; }
  .shape-menu button span { display: none; }
}

@media (max-width: 600px) {
  .page-edge { display: none; }
  .reader-content { padding: 3.5rem 1rem 4rem; }
  .panel-left, .panel-right { width: 100%; }
}

@media (max-width: 440px) {
  .tb-btn[aria-label="批注列表"],
  .tb-btn[aria-label="上一页"],
  .tb-btn[aria-label="下一页"] { display: none; }
  .tb-page-label { min-width: 58px; padding-inline: 0.3rem; }
}
</style>

<!-- Reading Mode Styles (non-scoped, applied globally for this view) -->
<style>
/* Reading mode enhanced styles for ReaderView */
html.reading-mode .reader-page {
  background-color: var(--reading-bg);
}

html.reading-mode .reader-content {
  background: var(--reading-bg);
  box-shadow: var(--reading-shadow);
  position: relative;
  border-radius: 4px;
}

html.reading-mode .reader-inner {
  max-width: var(--reading-max-width);
}

html.reading-mode .doc-body {
  color: var(--reading-text);
  letter-spacing: var(--reading-letter-spacing);
  line-height: var(--reading-line-height);
}

html.reading-mode .doc-body p {
  margin-bottom: var(--reading-para-margin);
}

/* Enhanced toolbar behavior in reading mode */
html.reading-mode .reader-topbar {
  opacity: var(--reading-toolbar-opacity);
  transition: opacity 0.3s ease, transform 0.3s ease;
}

html.reading-mode .reader-topbar:hover {
  opacity: 1;
}

/* Reading mode: softer text colors */
html.reading-mode .doc-meta {
  color: var(--reading-text-muted);
}

html.reading-mode .doc-body h1,
html.reading-mode .doc-body h2,
html.reading-mode .doc-body h3,
html.reading-mode .doc-body h4 {
  color: var(--reading-text);
}

/* Reading mode: enhanced code blocks and quotes */
html.reading-mode .doc-body blockquote {
  background: rgba(198, 122, 78, 0.06);
  border-left-color: rgba(198, 122, 78, 0.5);
}

html.reading-mode .doc-body code {
  background: rgba(240, 230, 220, 0.6);
}

html.reading-mode .doc-body pre {
  background: rgba(240, 230, 220, 0.4);
}

/* Reading mode: PDF viewer enhancements */
html.reading-mode .reader-pdf-wrapper {
  background: var(--reading-bg);
}

/* Reading mode: page edge effect (simulate book pages) */
html.reading-mode .page-edge {
  background: linear-gradient(to right, transparent, rgba(0, 0, 0, 0.03) 40%, rgba(0, 0, 0, 0.06));
}

/* Reading mode: panel overlays with softer appearance */
html.reading-mode .panel-overlay {
  background: rgba(0, 0, 0, 0.08);
}

html.reading-mode .panel-left,
html.reading-mode .panel-right {
  background: rgba(250, 248, 245, 0.95);
}
</style>
