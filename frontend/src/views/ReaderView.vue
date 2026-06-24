<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import { useAnnotationStore } from '../stores/annotationStore'
import { marked } from 'marked'
import PDFViewer from '../components/PDFViewer.vue'
import AnnotationToolbar from '../components/AnnotationToolbar.vue'
import type { PDFActiveTool, ShapeType, TextDrawing } from '../components/pdfDrawingTypes'

const route = useRoute()
const router = useRouter()
const documentStore = useDocumentStore()
const annotationStore = useAnnotationStore()

const docId = computed(() => Number(route.params.documentId))
// Capture docId at mount time so unmount still has a valid id (route param may be NaN after navigation)
let capturedDocId = 0
const panelLeftOpen = ref(false)
const panelRightOpen = ref(false)
const topbarHidden = ref(false)
const showAnnoTab = ref(true)
const loading = ref(true)
let lastScroll = 0

// Annotation toolbar
const annoToolbar = ref<InstanceType<typeof AnnotationToolbar> | null>(null)
const docBodyRef = ref<HTMLElement | null>(null)

// PDF drawing state
const pdfRef = ref<InstanceType<typeof PDFViewer> | null>(null)
const pdfActiveTool = ref<PDFActiveTool>('none')
const pdfShapeType = ref<ShapeType>('rectangle')
const pdfShapeMenuOpen = ref(false)
const pdfPenColor = ref('#FF0000')
const pdfPenWidth = ref(3)
const pdfTextSize = ref(24)
const pdfSelectedText = ref<TextDrawing | null>(null)
const pdfEraserSize = ref(24)
const pdfEraseMode = ref<'freehand' | 'area'>('freehand')
const pdfZoom = ref(100)
const pdfCurPage = ref(1)
const pdfPageCount = ref(1)

const pdfUsesStyle = computed(() =>
  pdfActiveTool.value === 'pen' ||
  pdfActiveTool.value === 'highlighter' ||
  pdfActiveTool.value === 'shape' ||
  pdfActiveTool.value === 'text' ||
  pdfSelectedText.value !== null)

const pdfShowsTextControls = computed(() =>
  pdfActiveTool.value === 'text' || pdfSelectedText.value !== null)

function pdfSwitchTool(tool: PDFActiveTool) {
  pdfActiveTool.value = pdfActiveTool.value === tool ? 'none' : tool
  if (tool !== 'shape') pdfShapeMenuOpen.value = false
  if (tool !== 'select') pdfSelectedText.value = null
}
function pdfToggleShapeMenu() {
  pdfShapeMenuOpen.value = !pdfShapeMenuOpen.value
}
function pdfChooseShape(shapeType: ShapeType) {
  pdfShapeType.value = shapeType
  pdfActiveTool.value = 'shape'
  pdfShapeMenuOpen.value = false
}
function pdfToggleEraseMode() {
  pdfEraseMode.value = pdfEraseMode.value === 'freehand' ? 'area' : 'freehand'
}
function pdfZoomIn() { pdfRef.value?.zoomIn() }
function pdfZoomOut() { pdfRef.value?.zoomOut() }
function pdfUndo() { pdfRef.value?.undoLastStroke(pdfCurPage.value) }
function onPDFPageChange(page: number) { pdfCurPage.value = page }
function onPDFPageCountChange(count: number) { pdfPageCount.value = count }
function onPDFTextSelectionChange(drawing: TextDrawing | null) {
  pdfSelectedText.value = drawing
  if (!drawing) return
  pdfPenColor.value = drawing.color
  pdfTextSize.value = drawing.fontSize
}
function pdfApplySelectedTextStyle() {
  if (!pdfSelectedText.value) return
  pdfRef.value?.applySelectedTextStyle?.({
    color: pdfPenColor.value,
    fontSize: pdfTextSize.value,
  })
}

// Progress tracking
let progressTimer: ReturnType<typeof setInterval> | null = null
const CONTENT_SAVE_INTERVAL = 3000 // ms

const doc = computed(() => documentStore.currentDocument)
const content = computed(() => documentStore.documentContent)
const annotations = computed(() => annotationStore.annotations)

/** Rendered markdown HTML or escaped plain text, depending on file type */
const renderedContent = computed(() => {
  const raw = content.value
  if (!raw) return ''
  const ft = doc.value?.file_type
  if (ft === 'md') {
    return marked.parse(raw, { async: false }) as string
  }
  // TXT / DOCX: escape HTML to show as plain text
  return raw
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
})

/** Content with annotation highlights applied */
const highlightedContent = computed(() => {
  let html = renderedContent.value
  if (!html) return ''
  const anns = annotations.value
  if (anns.length === 0) return html

  anns.forEach((a) => {
    if (!a.selected_text) return
    const escaped = a.selected_text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
    const regexStr = escaped.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    const regex = new RegExp(regexStr, 'g')
    html = html.replace(regex, `<mark data-anno-id="${a.id}" style="background:${a.color || '#FFD700'};cursor:pointer;border-radius:2px;padding:0 1px;" title="${escaped.replace(/"/g, '&quot;')}">${escaped}</mark>`)
  })
  return html
})

function toggleLeft() { panelLeftOpen.value = !panelLeftOpen.value }
function toggleRight() { panelRightOpen.value = !panelRightOpen.value }
function closePanels() { panelLeftOpen.value = false; panelRightOpen.value = false }

// ── Annotation: selection & toolbar ──

function onContentMouseUp(e: MouseEvent) {
  const sel = window.getSelection()
  if (!sel || sel.isCollapsed || !sel.toString().trim()) {
    annoToolbar.value?.hide()
    return
  }
  const range = sel.getRangeAt(0)
  if (!docBodyRef.value?.contains(range.commonAncestorContainer)) {
    annoToolbar.value?.hide()
    return
  }
  annoToolbar.value?.show(e.clientX, e.clientY)
}

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
  if (!docBodyRef.value) return
  const marks = docBodyRef.value.querySelectorAll('mark')
  marks.forEach((m) => {
    if (m.textContent === text) {
      m.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  })
}

/** Calculate reading progress based on scroll position (for non-PDF files) */
function calcScrollProgress(): number {
  const scrollTop = window.scrollY
  const docHeight = document.documentElement.scrollHeight - window.innerHeight
  if (docHeight <= 0) return 0
  return Math.round((scrollTop / docHeight) * 100)
}

/** Save current scroll-based progress to backend (only if higher than last saved) */
let lastSavedTextProgress = 0

function saveTextProgress() {
  const p = calcScrollProgress()
  // Only update if progress has increased (prevents opening a doc from resetting progress)
  if (p > lastSavedTextProgress) {
    lastSavedTextProgress = p
    console.log(`[Reader] 💾 Saving progress: ${p}% (scroll ${window.scrollY} / ${document.documentElement.scrollHeight - window.innerHeight})`)
    documentStore.updateProgress(capturedDocId, Math.max(1, p))
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
  topbarHidden.value = cur > 200 && cur > lastScroll
  lastScroll = cur
}

onMounted(async () => {
  window.addEventListener('scroll', handleScroll)
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
    // Save initial "opened" status immediately (1% floor), won't regress due to > guard
    saveTextProgress()
    // Save periodically while reading
    progressTimer = setInterval(saveTextProgress, CONTENT_SAVE_INTERVAL)
  }
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
  // Final progress save
  if (progressTimer) {
    clearInterval(progressTimer)
    progressTimer = null
  }
  if (doc.value?.file_type === 'pdf') {
    if (lastPDFProgress > 0) {
      documentStore.updateProgress(capturedDocId, lastPDFProgress)
    }
  } else {
    const finalP = calcScrollProgress()
    // Only save if it's higher than what was last saved
    if (finalP > lastSavedTextProgress) {
      documentStore.updateProgress(capturedDocId, Math.max(1, finalP))
    }
  }
})
</script>

<template>
  <div class="reader-page">
    <!-- Overlay -->
    <div :class="['panel-overlay', { show: panelLeftOpen || panelRightOpen }]" @click="closePanels"></div>

    <!-- Floating Top Bar -->
    <div :class="['reader-topbar', { hidden: topbarHidden, 'eraser-active': pdfActiveTool === 'eraser' }]">
      <button class="tb-btn" title="返回" @click="router.push('/documents')">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="19" y1="12" x2="5" y2="12"/><polyline points="12 19 5 12 12 5"/></svg>
      </button>
      <div class="tb-divider"></div>
      <button class="tb-btn" title="目录" @click="toggleLeft">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
      </button>
      <button class="tb-btn" title="批注和AI" @click="toggleRight">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
      </button>

      <!-- PDF drawing tools -->
      <template v-if="doc?.file_type === 'pdf'">
        <input v-if="pdfActiveTool === 'eraser'" type="range" v-model.number="pdfEraserSize" class="tb-size-slider tb-size-slider--eraser" min="8" max="48" title="橡皮大小" />
        <span v-if="pdfActiveTool === 'eraser'" class="tb-label">橡皮 {{ pdfEraserSize }}px</span>
        <div class="tb-divider"></div>
        <button :class="['tb-btn', { active: pdfActiveTool === 'pen' }]" title="画笔" @click="pdfSwitchTool('pen')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>
        </button>
        <button :class="['tb-btn', { active: pdfActiveTool === 'highlighter' }]" title="荧光笔" @click="pdfSwitchTool('highlighter')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M9 11l-6 6v3h9l3-23"/><path d="M9 3h-2l-7 14h2l7-14z"/></svg>
        </button>
        <button :class="['tb-btn', { active: pdfActiveTool === 'text' }]" title="添加文本" @click="pdfSwitchTool('text')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M4 6V4h16v2"/><path d="M12 4v16"/><path d="M8 20h8"/></svg>
        </button>
        <button :class="['tb-btn', { active: pdfActiveTool === 'select' }]" title="选择形状" @click="pdfSwitchTool('select')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M5 3l14 8-6 2-3 6z"/></svg>
        </button>
        <div class="shape-tool-wrap">
          <button :class="['tb-btn', { active: pdfActiveTool === 'shape' }]" title="形状" @click="pdfToggleShapeMenu">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><rect x="3" y="5" width="11" height="11" rx="1"/><circle cx="16.5" cy="14.5" r="4.5"/></svg>
          </button>
          <div v-if="pdfShapeMenuOpen" class="shape-menu">
            <button :class="{ active: pdfShapeType === 'line' }" @click="pdfChooseShape('line')">
              <svg viewBox="0 0 24 24"><line x1="4" y1="20" x2="20" y2="4"/></svg><span>直线</span>
            </button>
            <button :class="{ active: pdfShapeType === 'arrow' }" @click="pdfChooseShape('arrow')">
              <svg viewBox="0 0 24 24"><line x1="4" y1="20" x2="19" y2="5"/><polyline points="11,5 19,5 19,13"/></svg><span>箭头</span>
            </button>
            <button :class="{ active: pdfShapeType === 'rectangle' }" @click="pdfChooseShape('rectangle')">
              <svg viewBox="0 0 24 24"><rect x="4" y="5" width="16" height="14" rx="1"/></svg><span>矩形</span>
            </button>
            <button :class="{ active: pdfShapeType === 'ellipse' }" @click="pdfChooseShape('ellipse')">
              <svg viewBox="0 0 24 24"><ellipse cx="12" cy="12" rx="8" ry="6"/></svg><span>椭圆</span>
            </button>
          </div>
        </div>
        <button :class="['tb-btn', { active: pdfActiveTool === 'eraser' }]" title="橡皮擦" @click="pdfSwitchTool('eraser')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M20 20H7L3 16c-.8-.8-.8-2 0-2.8L14.5 1.7c.8-.8 2-.8 2.8 0L20 4.3c.8.8.8 2 0 2.8L8.5 18.7"/></svg>
        </button>
        <button v-if="pdfActiveTool === 'eraser'" class="tb-btn" :title="pdfEraseMode === 'freehand' ? '局部擦除（点击切换框选擦除）' : '框选局部擦除（点击切换自由擦除）'" @click="pdfToggleEraseMode">
          <svg v-if="pdfEraseMode === 'freehand'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M17 3a2.85 2.85 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"/></svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><rect x="3" y="3" width="18" height="18" rx="2"/></svg>
        </button>
        <button class="tb-btn" title="撤销" @click="pdfUndo">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
        </button>
        <input v-if="pdfUsesStyle" type="color" v-model="pdfPenColor" class="tb-color-picker" title="颜色" @change="pdfApplySelectedTextStyle" />
        <input v-if="pdfShowsTextControls" type="range" v-model.number="pdfTextSize" class="tb-size-slider tb-size-slider--text" min="10" max="72" title="字号" @change="pdfApplySelectedTextStyle" />
        <span v-if="pdfShowsTextControls" class="tb-label">字号 {{ pdfTextSize }}px</span>
        <input v-else-if="pdfUsesStyle" type="range" v-model.number="pdfPenWidth" class="tb-size-slider" min="1" max="20" title="粗细" />
        <div class="tb-divider"></div>
        <span class="tb-label">{{ pdfCurPage }}/{{ pdfPageCount }}</span>
        <button class="tb-btn" title="缩小" @click="pdfZoomOut">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="8" y1="11" x2="14" y2="11"/></svg>
        </button>
        <span class="tb-label">{{ pdfZoom }}%</span>
        <button class="tb-btn" title="放大" @click="pdfZoomIn">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="11" y1="8" x2="11" y2="14"/><line x1="8" y1="11" x2="14" y2="11"/></svg>
        </button>
      </template>

      <!-- Text reader tools -->
      <template v-if="doc?.file_type !== 'pdf'">
        <div class="tb-divider"></div>
        <button class="tb-btn" title="搜索">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
        </button>
        <button class="tb-btn" title="缩小">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
        </button>
        <span class="tb-label">100%</span>
        <button class="tb-btn" title="放大">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="16"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
        </button>
        <div class="tb-divider"></div>
        <button class="tb-btn" title="AI总结">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>
        </button>
      </template>
    </div>

    <!-- Left Panel: TOC (placeholder — real TOC needs parsed content) -->
    <div :class="['panel-left', { open: panelLeftOpen }]">
      <div class="panel__header">
        <h3>目录</h3>
        <button @click="panelLeftOpen = false">✕</button>
      </div>
      <div class="panel__body">
        <div class="toc-item level-1 active">{{ doc?.title || '文档' }}</div>
        <div class="toc-placeholder">目录解析暂未实现</div>
      </div>
    </div>

    <!-- Right Panel: Annotations + AI -->
    <div :class="['panel-right', { open: panelRightOpen }]">
      <div class="panel__header">
        <h3>批注与 AI</h3>
        <button @click="panelRightOpen = false">✕</button>
      </div>
      <div class="panel__body">
        <div class="panel-tabs">
          <button :class="['pt-btn', { active: showAnnoTab }]" @click="showAnnoTab = true">
            批注 ({{ annotations.length }})
          </button>
          <button :class="['pt-btn', { active: !showAnnoTab }]" @click="showAnnoTab = false">
            AI 助手
          </button>
        </div>

        <!-- Annotations tab -->
        <div v-if="showAnnoTab">
          <div v-if="annotations.length === 0" class="anno-empty">选中文本即可添加批注</div>
          <div
            v-for="anno in annotations"
            :key="anno.id"
            class="anno-card"
            @click="scrollToAnnotation(anno.selected_text)"
          >
            <div class="anno-card__text">{{ anno.selected_text }}</div>
            <div class="anno-card__meta">
              <span class="anno-highlight" :style="{ background: anno.color || '#FDE68A' }">{{ anno.type === 'highlight' ? '高亮' : anno.type }}</span>
              {{ anno.note ? `· ${anno.note}` : '' }}
            </div>
            <button class="anno-card__del" title="删除" @click.stop="deleteAnnotation(anno.id)">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
            </button>
          </div>
        </div>

        <!-- AI tab -->
        <div v-else>
          <div class="ai-message">
            <div class="ai-message__label">AI 总结</div>
            <p>你的 AI 阅读助手已就绪。选中文本后可进行解释、翻译或提问。</p>
          </div>
          <div class="ai-input">
            <input type="text" placeholder="向 AI 提问..." />
          </div>
        </div>
      </div>
    </div>

    <!-- Right page edge decoration -->
    <div class="page-edge"></div>

    <!-- Annotation Toolbar -->
    <AnnotationToolbar ref="annoToolbar" @highlight="onAnnoHighlight" @note="onAnnoNote" />

    <!-- Reading Content -->
    <div v-if="loading" class="reader-content">
      <div class="reader-inner"><p style="text-align:center;padding:4rem 0;color:var(--text-muted)">加载中...</p></div>
    </div>

    <div v-else-if="doc?.file_type === 'pdf'" class="reader-pdf-wrapper">
      <PDFViewer
        ref="pdfRef"
        :document-id="capturedDocId"
        :active-tool="pdfActiveTool"
        :shape-type="pdfShapeType"
        :erase-mode="pdfEraseMode"
        :pen-color="pdfPenColor"
        :pen-width="pdfActiveTool === 'eraser' ? pdfEraserSize : pdfPenWidth"
        :text-size="pdfTextSize"
        @progress="onPDFProgress"
        @current-page-change="onPDFPageChange"
        @page-count-change="onPDFPageCountChange"
        @text-selection-change="onPDFTextSelectionChange"
      />
    </div>

    <div v-else class="reader-content">
      <div class="reader-inner">
        <h1 class="doc-title">{{ doc?.title || '未命名文档' }}</h1>
        <div class="doc-meta">
          {{ doc?.file_type?.toUpperCase() || '' }}
          {{ doc?.read_progress ? `· 已读至${doc.read_progress}%` : '' }}
        </div>

        <div v-if="highlightedContent" ref="docBodyRef" class="doc-body" v-html="highlightedContent" @mouseup="onContentMouseUp"></div>
        <div v-else class="doc-body">
          <p style="color:var(--text-muted)">文档内容暂不可用</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.reader-page { position: relative; min-height: 100vh; background: var(--bg-page); background-image: repeating-linear-gradient(0deg, transparent, transparent 1px, rgba(0,0,0,0.005) 1px, rgba(0,0,0,0.005) 2px); background-size: 100% 2px; font-family: var(--font-body); }

.panel-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.15); z-index: 15; opacity: 0; pointer-events: none; transition: opacity 0.3s; }
.panel-overlay.show { opacity: 1; pointer-events: auto; }

.reader-topbar { position: fixed; top: 0.75rem; left: 50%; transform: translateX(-50%); z-index: 30; display: flex; align-items: center; gap: 0.3rem; padding: 0.35rem 0.6rem; background: rgba(250,248,245,0.92); backdrop-filter: blur(8px); border: 1px solid var(--border-color); border-radius: 24px; box-shadow: 0 2px 12px rgba(61,46,36,0.08); transition: opacity 0.3s, transform 0.3s; font-family: var(--font-ui); }
.reader-topbar.hidden { opacity: 0; transform: translateX(-50%) translateY(-10px); pointer-events: none; }
.tb-btn { width: 34px; height: 34px; border: none; border-radius: 50%; background: transparent; display: flex; align-items: center; justify-content: center; cursor: pointer; color: var(--text-secondary); transition: all 0.12s; }
.tb-btn:hover { background: var(--accent-light); color: var(--accent); }
.tb-btn.active { background: var(--accent); color: #fff; }
.tb-btn.active:hover { background: var(--accent); opacity: 0.85; }
.tb-btn svg { width: 18px; height: 18px; }
.tb-divider { width: 1px; height: 20px; background: var(--border-color); margin: 0 0.2rem; }
.tb-label { font-size: 0.7rem; color: var(--text-muted); padding: 0 0.4rem; }
.tb-color-picker { width: 24px; height: 24px; border: 1px solid var(--border-color); border-radius: 4px; padding: 0; cursor: pointer; background: transparent; }
.tb-size-slider { width: 60px; height: 20px; cursor: pointer; accent-color: var(--accent); }
.tb-size-slider--eraser { width: 88px; }
.tb-size-slider--text { width: 96px; }
.reader-topbar.eraser-active .tb-color-picker,
.reader-topbar.eraser-active .tb-size-slider:not(.tb-size-slider--eraser) {
  display: none;
}
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

@media (max-width: 1024px) { .reader-content { max-width: 100%; padding: 4.5rem 2rem 5rem; } .reader-inner { max-width: 100%; } }
@media (max-width: 600px) { .page-edge { display: none; } .reader-content { padding: 3.5rem 1rem 4rem; } .panel-left, .panel-right { width: 100%; } }
</style>
