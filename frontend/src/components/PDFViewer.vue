<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import * as pdfjs from 'pdfjs-dist'
import request from '../utils/request'
import type { AnnotationReplacementCreate } from '../api/annotation'
import { useAnnotationStore } from '../stores/annotationStore'
import PDFPageWrapper from './PDFPageWrapper.vue'
import type {
  Drawing,
  DrawingReplacement,
  PDFActiveTool,
  Point,
  ShapeDrawing,
  ShapeType,
} from './pdfDrawingTypes'
import { isShapeDrawing } from './pdfDrawingTypes'
import { remapStrokeInHistory } from './pdfAnnotationHistory'

pdfjs.GlobalWorkerOptions.workerSrc = '/pdf.worker.min.mjs'

const props = defineProps<{
  documentId: number
  activeTool: PDFActiveTool
  shapeType: ShapeType
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
  original: Drawing
  fragments: Drawing[]
}

type HistoryEntry =
  | { type: 'draw'; pageNum: number; stroke: Drawing }
  | { type: 'erase'; pageNum: number; replacements: EraseHistoryReplacement[] }
  | {
      type: 'edit'
      pageNum: number
      index: number
      original: Drawing
      replacement: Drawing
    }

const annotationStore = useAnnotationStore()

const container = ref<HTMLElement | null>(null)
const pageRefs = ref<InstanceType<typeof PDFPageWrapper>[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const pageCount = ref(0)
const currentPage = ref(1)
const zoom = ref(100)

const drawingsByPage = ref<Map<number, Drawing[]>>(new Map())
const history: HistoryEntry[] = []
let pdfDoc: pdfjs.PDFDocumentProxy | null = null

function drawingCreateData(pageNum: number, drawing: Drawing): AnnotationReplacementCreate {
  const positionData = isShapeDrawing(drawing)
    ? {
        tool: 'shape',
        shapeType: drawing.shapeType,
        width: drawing.width,
        start: drawing.start,
        end: drawing.end,
      }
    : {
        tool: drawing.tool,
        width: drawing.width,
        points: drawing.points,
      }

  return {
    page: pageNum,
    selected_text: '',
    color: drawing.color,
    type: 'drawing',
    position_data: positionData,
  }
}

async function saveDrawing(pageNum: number, drawing: Drawing): Promise<Drawing> {
  const saved = await annotationStore.create({
    document_id: props.documentId,
    ...drawingCreateData(pageNum, drawing),
  })
  return { ...drawing, id: saved.id }
}

async function replacePersistedDrawings(
  pageNum: number,
  deletedDrawings: Drawing[],
  createdDrawings: Drawing[],
): Promise<Drawing[]> {
  const deleteIDs = deletedDrawings
    .map((drawing) => drawing.id)
    .filter((id): id is number => id !== undefined)
  if (deleteIDs.length !== deletedDrawings.length) {
    throw new Error('存在尚未保存的批注，无法执行批量替换')
  }

  const created = await annotationStore.replace({
    document_id: props.documentId,
    delete_ids: deleteIDs,
    creates: createdDrawings.map((drawing) => drawingCreateData(pageNum, drawing)),
  })
  return created.map((annotation, index) => ({
    ...createdDrawings[index],
    id: annotation.id,
  }))
}

function setPageDrawings(pageNum: number, drawings: Drawing[]) {
  drawingsByPage.value.set(pageNum, [...drawings])
}

function redrawPage(pageNum: number) {
  const wrapper = pageRefs.value[pageNum - 1]
  if (wrapper) {
    wrapper.clearDrawLayer()
    wrapper.redrawDrawings()
  }
}

function cancelReplacementPreview(pageNum: number) {
  pageRefs.value[pageNum - 1]?.cancelPendingReplacement()
}

async function addDrawing(pageNum: number, drawing: Drawing) {
  const drawings = drawingsByPage.value.get(pageNum) || []
  const optimistic = { ...drawing }
  setPageDrawings(pageNum, [...drawings, optimistic])

  try {
    const saved = await saveDrawing(pageNum, drawing)
    const current = drawingsByPage.value.get(pageNum) || []
    setPageDrawings(pageNum, current.map((item) => (item === optimistic ? saved : item)))
    history.push({ type: 'draw', pageNum, stroke: saved })
  } catch (e: any) {
    console.warn('保存批注失败:', e?.message || e)
    setPageDrawings(
      pageNum,
      (drawingsByPage.value.get(pageNum) || []).filter((item) => item !== optimistic),
    )
  }
}

async function deleteDrawing(drawing: Drawing) {
  if (drawing.id) await annotationStore.remove(drawing.id)
}

function validPoint(value: unknown): value is { x: number; y: number } {
  if (!value || typeof value !== 'object') return false
  const point = value as Record<string, unknown>
  return typeof point.x === 'number' && typeof point.y === 'number'
}

function loadExistingDrawings() {
  const map = new Map<number, Drawing[]>()
  for (const annotation of annotationStore.annotations) {
    if (annotation.type !== 'drawing') continue
    const position = (annotation.position_data || {}) as Record<string, unknown>
    let drawing: Drawing | null = null

    if (
      position.tool === 'shape' &&
      ['line', 'arrow', 'rectangle', 'ellipse'].includes(String(position.shapeType)) &&
      validPoint(position.start) &&
      validPoint(position.end)
    ) {
      drawing = {
        id: annotation.id,
        tool: 'shape',
        shapeType: position.shapeType as ShapeType,
        color: annotation.color || '#FF0000',
        width: typeof position.width === 'number' ? position.width : 3,
        start: position.start,
        end: position.end,
      }
    } else if (Array.isArray(position.points) && position.points.length >= 2) {
      drawing = {
        id: annotation.id,
        tool: position.tool === 'highlighter' ? 'highlighter' : 'pen',
        color: annotation.color || '#FF0000',
        width: typeof position.width === 'number' ? position.width : 3,
        points: position.points as Point[],
      }
    }

    if (!drawing) continue
    const existing = map.get(annotation.page) || []
    existing.push(drawing)
    map.set(annotation.page, existing)
  }
  drawingsByPage.value = map
}

async function reloadDrawings(pageNum: number) {
  await annotationStore.fetchAnnotations(props.documentId)
  loadExistingDrawings()
  redrawPage(pageNum)
}

async function replaceDrawings(pageNum: number, replacements: DrawingReplacement[]) {
  const drawings = drawingsByPage.value.get(pageNum)
  if (!drawings || drawings.length === 0) return

  const operations = replacements
    .map((replacement) => ({
      ...replacement,
      original: drawings[replacement.index],
    }))
    .filter((operation): operation is DrawingReplacement & { original: Drawing } =>
      Boolean(operation.original))
  if (operations.length === 0) {
    cancelReplacementPreview(pageNum)
    return
  }

  try {
    const fragmentCounts = operations.map((operation) => operation.fragments.length)
    const savedFragments = await replacePersistedDrawings(
      pageNum,
      operations.map((operation) => operation.original),
      operations.flatMap((operation) => operation.fragments),
    )

    let fragmentOffset = 0
    const savedGroups = fragmentCounts.map((count) => {
      const group = savedFragments.slice(fragmentOffset, fragmentOffset + count)
      fragmentOffset += count
      return group
    })
    const fragmentsByIndex = new Map(
      operations.map((operation, index) => [operation.index, savedGroups[index]]),
    )
    const nextDrawings = drawings.flatMap((drawing, index) =>
      fragmentsByIndex.get(index) || [drawing])

    setPageDrawings(pageNum, nextDrawings)
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
    cancelReplacementPreview(pageNum)
  }
}

async function editShape(pageNum: number, index: number, shape: ShapeDrawing) {
  const drawings = drawingsByPage.value.get(pageNum) || []
  const original = drawings[index]
  if (!original || !isShapeDrawing(original)) {
    cancelReplacementPreview(pageNum)
    return
  }

  try {
    const replacement = (await replacePersistedDrawings(pageNum, [original], [shape]))[0]
    const next = [...drawings]
    next[index] = replacement
    setPageDrawings(pageNum, next)
    history.push({
      type: 'edit',
      pageNum,
      index,
      original,
      replacement,
    })
  } catch (e: any) {
    console.warn('保存形状编辑失败:', e?.message || e)
    cancelReplacementPreview(pageNum)
  }
}

async function undoErase(entryIndex: number, entry: Extract<HistoryEntry, { type: 'erase' }>) {
  try {
    const restoredOriginals = await replacePersistedDrawings(
      entry.pageNum,
      entry.replacements.flatMap((replacement) => replacement.fragments),
      entry.replacements.map((replacement) => replacement.original),
    )
    entry.replacements.forEach((replacement, index) => {
      remapStrokeInHistory(history, replacement.original, restoredOriginals[index], entryIndex)
    })

    const current = drawingsByPage.value.get(entry.pageNum) || []
    const fragmentIDs = new Set(
      entry.replacements.flatMap((replacement) =>
        replacement.fragments
          .map((fragment) => fragment.id)
          .filter((id): id is number => id !== undefined),
      ),
    )
    const remaining = current.filter((drawing) => !drawing.id || !fragmentIDs.has(drawing.id))
    const restoredByIndex = new Map(
      entry.replacements.map((replacement, index) => [replacement.index, restoredOriginals[index]]),
    )

    const result: Drawing[] = []
    let remainingIndex = 0
    const targetLength = remaining.length + restoredOriginals.length
    for (let index = 0; index < targetLength; index++) {
      const restored = restoredByIndex.get(index)
      if (restored) result.push(restored)
      else if (remainingIndex < remaining.length) result.push(remaining[remainingIndex++])
    }
    while (remainingIndex < remaining.length) result.push(remaining[remainingIndex++])

    setPageDrawings(entry.pageNum, result)
    history.splice(entryIndex, 1)
    redrawPage(entry.pageNum)
  } catch (e: any) {
    console.warn('恢复擦除批注失败:', e?.message || e)
    redrawPage(entry.pageNum)
  }
}

async function undoEdit(entryIndex: number, entry: Extract<HistoryEntry, { type: 'edit' }>) {
  try {
    const restored = (await replacePersistedDrawings(
      entry.pageNum,
      [entry.replacement],
      [entry.original],
    ))[0]
    remapStrokeInHistory(history, entry.original, restored, entryIndex)

    const current = drawingsByPage.value.get(entry.pageNum) || []
    const next = [...current]
    next[entry.index] = restored
    setPageDrawings(entry.pageNum, next)
    history.splice(entryIndex, 1)
    redrawPage(entry.pageNum)
  } catch (e: any) {
    console.warn('撤销形状编辑失败:', e?.message || e)
    redrawPage(entry.pageNum)
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
  if (entry.type === 'edit') {
    await undoEdit(entryIndex, entry)
    return
  }

  const drawings = drawingsByPage.value.get(entry.pageNum) || []
  setPageDrawings(
    entry.pageNum,
    drawings.filter((drawing) => drawing !== entry.stroke && drawing.id !== entry.stroke.id),
  )
  try {
    await deleteDrawing(entry.stroke)
    history.splice(entryIndex, 1)
  } catch (e: any) {
    console.warn('撤销绘制失败:', e?.message || e)
    await reloadDrawings(entry.pageNum)
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
    loadExistingDrawings()
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
        :drawings="drawingsByPage.get(pageNum) || []"
        :active-tool="props.activeTool"
        :shape-type="props.shapeType"
        :erase-mode="props.eraseMode"
        :pen-color="props.penColor"
        :pen-width="props.penWidth"
        :eraser-size="props.penWidth"
        @drawing-created="(drawing: Drawing) => { void addDrawing(pageNum, drawing) }"
        @drawings-replaced="(replacements: DrawingReplacement[]) => {
          void replaceDrawings(pageNum, replacements)
        }"
        @shape-edited="(index: number, shape: ShapeDrawing) => {
          void editShape(pageNum, index, shape)
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
