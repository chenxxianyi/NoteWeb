<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import * as pdfjs from 'pdfjs-dist'
import request from '../utils/request'
import { useAnnotationStore } from '../stores/annotationStore'
import PDFPageWrapper from './PDFPageWrapper.vue'
import type { Stroke, StrokeReplacement } from './PDFPageWrapper.vue'
import { remapStrokeInHistory } from './pdfAnnotationHistory'

pdfjs.GlobalWorkerOptions.workerSrc = '/pdf.worker.min.mjs'

const props = defineProps<{
  documentId: number
  activeTool: 'none' | 'pen' | 'highlighter' | 'eraser'
  eraseMode: 'freehand' | 'area'
  penColor: string
  penWidth: number
}>()

const emit = defineEmits<{
  progress: [value: number]
  currentPageChange: [page: number]
  pageCountChange: [count: number]
}>()

type EraseHistoryReplacement = {
  index: number
  original: Stroke
  fragments: Stroke[]
}

type HistoryEntry =
  | { type: 'draw'; pageNum: number; stroke: Stroke }
  | { type: 'erase'; pageNum: number; replacements: EraseHistoryReplacement[] }

const annotationStore = useAnnotationStore()

const container = ref<HTMLElement | null>(null)
const pageRefs = ref<InstanceType<typeof PDFPageWrapper>[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const pageCount = ref(0)
const currentPage = ref(1)
const zoom = ref(100)

const strokesByPage = ref<Map<number, Stroke[]>>(new Map())
const history: HistoryEntry[] = []
let pdfDoc: pdfjs.PDFDocumentProxy | null = null

async function saveStroke(pageNum: number, stroke: Stroke): Promise<Stroke> {
  const saved = await annotationStore.create({
    document_id: props.documentId,
    page: pageNum,
    selected_text: '',
    color: stroke.color,
    type: 'drawing',
    position_data: { tool: stroke.tool, width: stroke.width, points: stroke.points },
  })
  return { ...stroke, id: saved.id }
}

function setPageStrokes(pageNum: number, strokes: Stroke[]) {
  strokesByPage.value.set(pageNum, [...strokes])
}

function redrawPage(pageNum: number) {
  const wrapper = pageRefs.value[pageNum - 1]
  if (wrapper) {
    wrapper.clearDrawLayer()
    wrapper.redrawStrokes()
  }
}

async function addStroke(pageNum: number, stroke: Stroke) {
  const strokes = strokesByPage.value.get(pageNum) || []
  const optimisticStroke = { ...stroke }
  setPageStrokes(pageNum, [...strokes, optimisticStroke])

  try {
    const savedStroke = await saveStroke(pageNum, stroke)
    const current = strokesByPage.value.get(pageNum) || []
    setPageStrokes(pageNum, current.map((item) => (item === optimisticStroke ? savedStroke : item)))
    history.push({ type: 'draw', pageNum, stroke: savedStroke })
  } catch (e: any) {
    console.warn('保存批注失败:', e?.message || e)
    setPageStrokes(
      pageNum,
      (strokesByPage.value.get(pageNum) || []).filter((item) => item !== optimisticStroke),
    )
  }
}

async function deleteStroke(stroke: Stroke) {
  if (stroke.id) await annotationStore.remove(stroke.id)
}

function loadExistingStrokes() {
  const map = new Map<number, Stroke[]>()
  for (const annotation of annotationStore.annotations) {
    if (annotation.type !== 'drawing') continue
    const position = (annotation.position_data || {}) as Record<string, unknown>
    if (!Array.isArray(position.points) || position.points.length < 2) continue

    const stroke: Stroke = {
      id: annotation.id,
      tool: position.tool === 'highlighter' ? 'highlighter' : 'pen',
      color: annotation.color || '#FF0000',
      width: typeof position.width === 'number' ? position.width : 3,
      points: position.points as Stroke['points'],
    }
    const existing = map.get(annotation.page) || []
    existing.push(stroke)
    map.set(annotation.page, existing)
  }
  strokesByPage.value = map
}

async function reloadStrokes(pageNum: number) {
  await annotationStore.fetchAnnotations(props.documentId)
  loadExistingStrokes()
  redrawPage(pageNum)
}

async function rollbackReplacement(
  pageNum: number,
  createdStrokes: Stroke[],
  deletedStrokes: Stroke[],
) {
  for (const stroke of createdStrokes) {
    try {
      await deleteStroke(stroke)
    } catch {
      // Best effort. The final reload reconciles local state with the server.
    }
  }
  for (const stroke of deletedStrokes) {
    try {
      await saveStroke(pageNum, stroke)
    } catch {
      // Continue restoring the remaining strokes before reloading.
    }
  }
  await reloadStrokes(pageNum)
}

async function replaceStrokes(pageNum: number, replacements: StrokeReplacement[]) {
  const strokes = strokesByPage.value.get(pageNum)
  if (!strokes || strokes.length === 0) return

  const operations = replacements
    .map((replacement) => ({
      ...replacement,
      original: strokes[replacement.index],
    }))
    .filter((operation): operation is StrokeReplacement & { original: Stroke } => Boolean(operation.original))
  if (operations.length === 0) return

  const createdFragments: Stroke[] = []
  const deletedOriginals: Stroke[] = []
  const savedGroups: Stroke[][] = []

  try {
    for (const operation of operations) {
      const savedFragments: Stroke[] = []
      for (const fragment of operation.fragments) {
        const saved = await saveStroke(pageNum, fragment)
        savedFragments.push(saved)
        createdFragments.push(saved)
      }
      savedGroups.push(savedFragments)
    }

    for (const operation of operations) {
      await deleteStroke(operation.original)
      deletedOriginals.push(operation.original)
    }

    const fragmentsByIndex = new Map(
      operations.map((operation, index) => [operation.index, savedGroups[index]]),
    )
    const nextStrokes = strokes.flatMap((stroke, index) => fragmentsByIndex.get(index) || [stroke])
    setPageStrokes(pageNum, nextStrokes)
    history.push({
      type: 'erase',
      pageNum,
      replacements: operations.map((operation, index) => ({
        index: operation.index,
        original: operation.original,
        fragments: savedGroups[index],
      })),
    })
  } catch (e: any) {
    console.warn('局部擦除保存失败:', e?.message || e)
    await rollbackReplacement(pageNum, createdFragments, deletedOriginals)
  }
}

async function undoErase(entryIndex: number, entry: Extract<HistoryEntry, { type: 'erase' }>) {
  const restoredOriginals: Stroke[] = []
  const deletedFragments: Stroke[] = []

  try {
    for (const replacement of entry.replacements) {
      restoredOriginals.push(await saveStroke(entry.pageNum, replacement.original))
    }
    for (const replacement of entry.replacements) {
      for (const fragment of replacement.fragments) {
        await deleteStroke(fragment)
        deletedFragments.push(fragment)
      }
    }
    entry.replacements.forEach((replacement, index) => {
      remapStrokeInHistory(history, replacement.original, restoredOriginals[index], entryIndex)
    })

    const current = strokesByPage.value.get(entry.pageNum) || []
    const fragmentIds = new Set(
      entry.replacements.flatMap((replacement) =>
        replacement.fragments
          .map((fragment) => fragment.id)
          .filter((id): id is number => id !== undefined),
      ),
    )
    const remaining = current.filter((stroke) => !stroke.id || !fragmentIds.has(stroke.id))
    const restoredByIndex = new Map(
      entry.replacements.map((replacement, index) => [replacement.index, restoredOriginals[index]]),
    )

    const result: Stroke[] = []
    let remainingIndex = 0
    const targetLength = remaining.length + restoredOriginals.length
    for (let index = 0; index < targetLength; index++) {
      const restored = restoredByIndex.get(index)
      if (restored) result.push(restored)
      else if (remainingIndex < remaining.length) result.push(remaining[remainingIndex++])
    }
    while (remainingIndex < remaining.length) result.push(remaining[remainingIndex++])

    setPageStrokes(entry.pageNum, result)
    history.splice(entryIndex, 1)
    redrawPage(entry.pageNum)
  } catch (e: any) {
    console.warn('恢复擦除批注失败:', e?.message || e)
    await rollbackReplacement(entry.pageNum, restoredOriginals, deletedFragments)
  }
}

async function undoLastStroke(pageNum: number) {
  let entryIndex = history.length - 1
  while (entryIndex >= 0 && history[entryIndex].pageNum !== pageNum) entryIndex--

  const entry = history[entryIndex]
  if (!entry) return

  if (entry.type === 'erase') {
    await undoErase(entryIndex, entry)
    return
  }

  const strokes = strokesByPage.value.get(entry.pageNum) || []
  setPageStrokes(
    entry.pageNum,
    strokes.filter((stroke) => stroke !== entry.stroke && stroke.id !== entry.stroke.id),
  )
  try {
    await deleteStroke(entry.stroke)
    history.splice(entryIndex, 1)
  } catch (e: any) {
    console.warn('撤销绘制失败:', e?.message || e)
    await reloadStrokes(entry.pageNum)
  }
  redrawPage(entry.pageNum)
}

async function loadPDF() {
  loading.value = true
  try {
    const response = await request.get(`/documents/${props.documentId}/file`, {
      responseType: 'arraybuffer',
    })
    const data = new Uint8Array(response.data)
    pdfDoc = await pdfjs.getDocument({ data }).promise
    pageCount.value = pdfDoc.numPages
    await annotationStore.fetchAnnotations(props.documentId)
    loadExistingStrokes()
    loading.value = false
    await renderAllPages()
  } catch (e: any) {
    error.value = e?.message || '无法加载 PDF 文件'
    loading.value = false
  }
}

async function renderAllPages() {
  if (!pdfDoc) return
  for (let pageNum = 1; pageNum <= pdfDoc.numPages; pageNum++) {
    const wrapper = pageRefs.value[pageNum - 1]
    if (wrapper) await wrapper.renderPDF(pdfDoc)
  }
}

function zoomIn() {
  if (zoom.value < 200) {
    zoom.value = Math.min(200, zoom.value + 10)
    void renderAllPages()
  }
}

function zoomOut() {
  if (zoom.value > 30) {
    zoom.value = Math.max(30, zoom.value - 10)
    void renderAllPages()
  }
}

function onScroll() {
  if (!container.value) return
  const wrappers = container.value.querySelectorAll('.pdf-page-wrapper')
  let closest = 1
  let closestDistance = Infinity
  wrappers.forEach((wrapper, index) => {
    const distance = Math.abs((wrapper as HTMLElement).offsetTop - container.value!.scrollTop)
    if (distance < closestDistance) {
      closestDistance = distance
      closest = index + 1
    }
  })
  currentPage.value = closest
}

watch(currentPage, (page) => {
  if (pageCount.value > 0) {
    emit('progress', Math.min(Math.round((page / pageCount.value) * 100), 100))
    emit('currentPageChange', page)
  }
})
watch(pageCount, (count) => emit('pageCountChange', count))

onMounted(() => {
  void loadPDF()
})

defineExpose({ zoomIn, zoomOut, undoLastStroke, zoom, currentPage, pageCount })
</script>

<template>
  <div class="pdf-viewer">
    <div v-if="loading" class="pdf-status"><p>加载 PDF 中...</p></div>
    <div v-else-if="error" class="pdf-status pdf-error"><p>{{ error }}</p></div>
    <div v-show="!loading && !error" ref="container" class="pdf-pages" @scroll="onScroll">
      <PDFPageWrapper
        v-for="pageNum in pageCount"
        :key="pageNum"
        :ref="(element: any) => { if (element) pageRefs[pageNum - 1] = element }"
        :data-page="pageNum"
        :page-num="pageNum"
        :scale="zoom / 100"
        :strokes="strokesByPage.get(pageNum) || []"
        :active-tool="props.activeTool"
        :erase-mode="props.eraseMode"
        :pen-color="props.penColor"
        :pen-width="props.penWidth"
        :eraser-size="props.penWidth"
        @stroke-created="(stroke: Stroke) => { void addStroke(pageNum, stroke) }"
        @strokes-replaced="(replacements: StrokeReplacement[]) => {
          void replaceStrokes(pageNum, replacements)
        }"
      />
    </div>
  </div>
</template>

<style scoped>
.pdf-viewer { display: flex; flex-direction: column; height: 100%; background: #f5f5f0; }
.pdf-status { flex: 1; display: flex; align-items: center; justify-content: center; color: #888; font-family: var(--font-ui, sans-serif); font-size: 0.9rem; }
.pdf-error { color: #b91c1c; }
.pdf-pages { flex: 1; overflow-y: auto; padding: 1rem; display: flex; flex-direction: column; align-items: center; gap: 1rem; }
</style>
