<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import * as pdfjs from 'pdfjs-dist'
import request from '../utils/request'
import { createRasterPdf } from '../utils/pdfRasterExport'
import type { AnnotationReplacementCreate } from '../api/annotation'
import { useAnnotationStore } from '../stores/annotationStore'
import PDFPageWrapper from './PDFPageWrapper.vue'
import type {
  Drawing,
  DrawingReplacement,
  PDFActiveTool,
  Point,
  ShapeType,
  TextErasure,
  TextDrawing,
} from './pdfDrawingTypes'
import {
  isShapeDrawing,
  isTextDrawing,
  sameDrawingIdentity,
} from './pdfDrawingTypes'
import { remapStrokeInHistory } from './pdfAnnotationHistory'

pdfjs.GlobalWorkerOptions.workerSrc = '/pdf.worker.min.mjs'

const props = defineProps<{
  documentId: number
  documentTitle?: string
  activeTool: PDFActiveTool
  shapeType: ShapeType
  eraseMode: 'freehand' | 'area'
  penColor: string
  penWidth: number
  textSize: number
}>()

const emit = defineEmits<{
  progress: [value: number]
  currentPageChange: [page: number]
  pageCountChange: [count: number]
  zoomChange: [zoom: number]
  historyStateChange: [state: { canUndo: boolean; canRedo: boolean }]
  savingStateChange: [saving: boolean]
  textSelectionChange: [drawing: TextDrawing | null]
  exportStateChange: [exporting: boolean]
}>()

type EraseHistoryReplacement = {
  index: number
  original: Drawing
  fragments: Drawing[]
}

type ReplacementOperation = DrawingReplacement & {
  original: Drawing
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
const canUndo = ref(false)
const canRedo = ref(false)
const pendingSaveCount = ref(0)
const exporting = ref(false)
const resolvedDocumentTitle = ref('')

const drawingsByPage = ref<Map<number, Drawing[]>>(new Map())
const history: HistoryEntry[] = []
const redoStack: HistoryEntry[] = []
let pdfDoc: pdfjs.PDFDocumentProxy | null = null
const pageTextCache = new Map<number, string>()
const pendingDrawingSaves = new WeakMap<Drawing, Promise<Drawing>>()

function updateHistoryState() {
  canUndo.value = history.some((entry) => entry.pageNum === currentPage.value)
  canRedo.value = redoStack.length > 0
  emit('historyStateChange', { canUndo: canUndo.value, canRedo: canRedo.value })
}

function recordHistory(entry: HistoryEntry) {
  history.push(entry)
  redoStack.length = 0
  updateHistoryState()
}

function moveHistoryEntryToRedo(entryIndex: number) {
  const [entry] = history.splice(entryIndex, 1)
  if (entry) redoStack.push(entry)
  updateHistoryState()
  return entry
}

function moveRedoEntryToHistory(entry: HistoryEntry) {
  history.push(entry)
  updateHistoryState()
}

function stripDrawingId(drawing: Drawing): Drawing {
  const next = { ...drawing }
  delete next.id
  return next
}

function applyDrawingReplacements(
  drawings: Drawing[],
  operations: Array<{ index: number; fragments: Drawing[] }>,
): Drawing[] {
  const fragmentsByIndex = new Map(
    operations.map((operation) => [operation.index, operation.fragments]),
  )
  return drawings.flatMap((drawing, index) => fragmentsByIndex.get(index) || [drawing])
}

function groupSavedFragments(savedFragments: Drawing[], fragmentCounts: number[]): Drawing[][] {
  let fragmentOffset = 0
  return fragmentCounts.map((count) => {
    const group = savedFragments.slice(fragmentOffset, fragmentOffset + count)
    fragmentOffset += count
    return group
  })
}

function sameDrawingObjects(left: Drawing[], right: Drawing[]) {
  return left.length === right.length && left.every((drawing, index) => drawing === right[index])
}

function beginSave() {
  pendingSaveCount.value += 1
  emit('savingStateChange', true)
}

function endSave() {
  pendingSaveCount.value = Math.max(0, pendingSaveCount.value - 1)
  emit('savingStateChange', pendingSaveCount.value > 0)
}

function drawingCreateData(pageNum: number, drawing: Drawing): AnnotationReplacementCreate {
  const positionData = isShapeDrawing(drawing)
    ? {
        tool: 'shape',
        shapeType: drawing.shapeType,
        width: drawing.width,
        start: drawing.start,
        end: drawing.end,
      }
    : isTextDrawing(drawing)
      ? {
          tool: 'text',
          text: drawing.text,
          fontSize: drawing.fontSize,
          x: drawing.x,
          y: drawing.y,
          width: drawing.width,
          height: drawing.height,
          erasures: drawing.erasures || [],
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

async function resolvePersistedDrawing(drawing: Drawing): Promise<Drawing> {
  if (drawing.id !== undefined) return drawing

  const pendingSave = pendingDrawingSaves.get(drawing)
  if (!pendingSave) {
    throw new Error('存在尚未保存的批注，无法执行批量替换')
  }
  return pendingSave
}

function findCurrentDrawingIndex(drawings: Drawing[], preferredIndex: number, target: Drawing) {
  if (sameDrawingIdentity(drawings[preferredIndex], target)) return preferredIndex
  return drawings.findIndex((drawing) => sameDrawingIdentity(drawing, target))
}

function setPageDrawings(pageNum: number, drawings: Drawing[]) {
  drawingsByPage.value = new Map(drawingsByPage.value).set(pageNum, [...drawings])
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

function applySelectedTextStyle(style: { color: string; fontSize: number }) {
  for (const wrapper of pageRefs.value) {
    if (wrapper?.applySelectedTextStyle?.(style)) return true
  }
  return false
}

async function addDrawing(pageNum: number, drawing: Drawing) {
  const drawings = drawingsByPage.value.get(pageNum) || []
  const optimistic = { ...drawing }
  setPageDrawings(pageNum, [...drawings, optimistic])
  beginSave()
  const savePromise = saveDrawing(pageNum, drawing)
  pendingDrawingSaves.set(optimistic, savePromise)

  try {
    const saved = await savePromise
    pendingDrawingSaves.delete(optimistic)
    const current = drawingsByPage.value.get(pageNum) || []
    let stillPresent = false
    const next = current.map((item) => {
      if (item !== optimistic) return item
      stillPresent = true
      return saved
    })
    if (stillPresent) {
      setPageDrawings(pageNum, next)
      recordHistory({ type: 'draw', pageNum, stroke: saved })
    }
  } catch (e: any) {
    pendingDrawingSaves.delete(optimistic)
    console.warn('保存批注失败:', e?.message || e)
    setPageDrawings(
      pageNum,
      (drawingsByPage.value.get(pageNum) || []).filter((item) => item !== optimistic),
    )
  } finally {
    endSave()
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

function validNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value)
}

function loadTextErasures(value: unknown): TextErasure[] {
  if (!Array.isArray(value)) return []
  const erasures: TextErasure[] = []

  for (const item of value) {
    if (!item || typeof item !== 'object') continue
    const erasure = item as Record<string, unknown>
    if (
      erasure.type === 'path' &&
      validNumber(erasure.radius) &&
      Array.isArray(erasure.points) &&
      erasure.points.every(validPoint)
    ) {
      erasures.push({
        type: 'path',
        radius: erasure.radius,
        points: erasure.points.map((point) => ({ x: point.x, y: point.y })),
      })
    } else if (
      erasure.type === 'rect' &&
      validPoint(erasure.start) &&
      validPoint(erasure.end)
    ) {
      erasures.push({
        type: 'rect',
        start: { x: erasure.start.x, y: erasure.start.y },
        end: { x: erasure.end.x, y: erasure.end.y },
      })
    }
  }

  return erasures
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
    } else if (
      position.tool === 'text' &&
      typeof position.text === 'string' &&
      validNumber(position.x) &&
      validNumber(position.y) &&
      validNumber(position.width)
    ) {
      drawing = {
        id: annotation.id,
        tool: 'text',
        color: annotation.color || '#FF0000',
        fontSize: validNumber(position.fontSize) ? position.fontSize : 24,
        text: position.text,
        x: position.x,
        y: position.y,
        width: position.width,
        height: validNumber(position.height)
          ? position.height
          : (validNumber(position.fontSize) ? position.fontSize : 24) * 1.25 + 8,
        erasures: loadTextErasures(position.erasures),
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

  const optimisticOperations: ReplacementOperation[] = operations.map((operation) => ({
    ...operation,
    fragments: operation.fragments.map(stripDrawingId),
  }))
  const optimisticDrawings = applyDrawingReplacements(drawings, optimisticOperations)
  const optimisticFragments = optimisticOperations.flatMap((operation) => operation.fragments)
  setPageDrawings(pageNum, optimisticDrawings)

  beginSave()
  const replacementSavePromise = (async () => {
    const resolvedOperations = await Promise.all(optimisticOperations.map(async (operation) => ({
      ...operation,
      original: await resolvePersistedDrawing(operation.original),
    })))
    const fragmentCounts = resolvedOperations.map((operation) => operation.fragments.length)
    const savedFragments = await replacePersistedDrawings(
      pageNum,
      resolvedOperations.map((operation) => operation.original),
      resolvedOperations.flatMap((operation) => operation.fragments),
    )

    return {
      resolvedOperations,
      savedGroups: groupSavedFragments(savedFragments, fragmentCounts),
    }
  })()
  optimisticOperations.forEach((operation, operationIndex) => {
    operation.fragments.forEach((fragment, fragmentIndex) => {
      const pendingSave = replacementSavePromise.then(({ savedGroups }) => {
        const saved = savedGroups[operationIndex]?.[fragmentIndex]
        if (!saved) throw new Error('Saved eraser fragments do not match local preview')
        return saved
      })
      void pendingSave.catch(() => {})
      pendingDrawingSaves.set(fragment, pendingSave)
    })
  })

  try {
    const { resolvedOperations, savedGroups } = await replacementSavePromise
    const savedByOptimistic = new Map<Drawing, Drawing>()
    optimisticOperations.forEach((operation, operationIndex) => {
      operation.fragments.forEach((fragment, fragmentIndex) => {
        const saved = savedGroups[operationIndex]?.[fragmentIndex]
        if (saved) savedByOptimistic.set(fragment, saved)
      })
    })
    const currentDrawings = drawingsByPage.value.get(pageNum) || []
    setPageDrawings(
      pageNum,
      currentDrawings.map((drawing) => savedByOptimistic.get(drawing) || drawing),
    )
    recordHistory({
      type: 'erase',
      pageNum,
      replacements: resolvedOperations.map((operation, index) => ({
        index: operation.index,
        original: operation.original,
        fragments: savedGroups[index],
      })),
    })
  } catch (e: any) {
    console.warn('局部擦除保存失败:', e?.message || e)
    const currentDrawings = drawingsByPage.value.get(pageNum) || []
    if (sameDrawingObjects(currentDrawings, optimisticDrawings)) {
      setPageDrawings(pageNum, drawings)
    } else if (currentDrawings.some((drawing) => optimisticFragments.includes(drawing))) {
      try {
        await reloadDrawings(pageNum)
      } catch {
        setPageDrawings(pageNum, drawings)
      }
    } else {
      cancelReplacementPreview(pageNum)
    }
  } finally {
    optimisticFragments.forEach((fragment) => pendingDrawingSaves.delete(fragment))
    endSave()
  }
}

async function editDrawing(pageNum: number, index: number, drawing: Drawing) {
  const drawings = drawingsByPage.value.get(pageNum) || []
  const original = drawings[index]
  if (!original) {
    cancelReplacementPreview(pageNum)
    return
  }

  const optimistic = { ...drawing }
  const optimisticDrawings = [...drawings]
  optimisticDrawings[index] = optimistic
  setPageDrawings(pageNum, optimisticDrawings)

  beginSave()
  const replacePromise = (async () => {
    const persistedOriginal = await resolvePersistedDrawing(original)
    return {
      persistedOriginal,
      replacement: (await replacePersistedDrawings(pageNum, [persistedOriginal], [drawing]))[0],
    }
  })()
  const replacementSavePromise = replacePromise.then(({ replacement }) => replacement)
  void replacementSavePromise.catch(() => {})
  pendingDrawingSaves.set(optimistic, replacementSavePromise)

  try {
    const { persistedOriginal, replacement } = await replacePromise
    const current = drawingsByPage.value.get(pageNum) || []
    const currentIndex = findCurrentDrawingIndex(current, index, optimistic)
    if (currentIndex < 0) {
      cancelReplacementPreview(pageNum)
      return
    }
    const committedDrawings = [...current]
    committedDrawings[currentIndex] = replacement
    setPageDrawings(pageNum, committedDrawings)
    recordHistory({
      type: 'edit',
      pageNum,
      index: currentIndex,
      original: persistedOriginal,
      replacement,
    })
  } catch (e: any) {
    console.warn('保存形状编辑失败:', e?.message || e)
    const current = drawingsByPage.value.get(pageNum) || []
    const currentIndex = findCurrentDrawingIndex(current, index, optimistic)
    if (currentIndex >= 0) {
      const rolledBackDrawings = [...current]
      rolledBackDrawings[currentIndex] = original
      setPageDrawings(pageNum, rolledBackDrawings)
    }
    cancelReplacementPreview(pageNum)
  } finally {
    pendingDrawingSaves.delete(optimistic)
    endSave()
  }
}

async function undoErase(entryIndex: number, entry: Extract<HistoryEntry, { type: 'erase' }>) {
  beginSave()
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
    const undone = moveHistoryEntryToRedo(entryIndex)
    if (undone?.type === 'erase') {
      undone.replacements = undone.replacements.map((replacement, index) => ({
        ...replacement,
        original: restoredOriginals[index],
      }))
    }
    redrawPage(entry.pageNum)
  } catch (e: any) {
    console.warn('恢复擦除批注失败:', e?.message || e)
    redrawPage(entry.pageNum)
  } finally {
    endSave()
  }
}

async function undoEdit(entryIndex: number, entry: Extract<HistoryEntry, { type: 'edit' }>) {
  beginSave()
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
    const undone = moveHistoryEntryToRedo(entryIndex)
    if (undone?.type === 'edit') undone.original = restored
    redrawPage(entry.pageNum)
  } catch (e: any) {
    console.warn('撤销形状编辑失败:', e?.message || e)
    redrawPage(entry.pageNum)
  } finally {
    endSave()
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
  beginSave()
  try {
    await deleteDrawing(entry.stroke)
    moveHistoryEntryToRedo(entryIndex)
  } catch (e: any) {
    console.warn('撤销绘制失败:', e?.message || e)
    await reloadDrawings(entry.pageNum)
  } finally {
    endSave()
  }
  redrawPage(entry.pageNum)
}

async function redoDraw(entry: Extract<HistoryEntry, { type: 'draw' }>) {
  beginSave()
  try {
    const saved = await saveDrawing(entry.pageNum, stripDrawingId(entry.stroke))
    entry.stroke = saved
    const drawings = drawingsByPage.value.get(entry.pageNum) || []
    setPageDrawings(entry.pageNum, [...drawings, saved])
    moveRedoEntryToHistory(entry)
    redrawPage(entry.pageNum)
  } catch (e: any) {
    console.warn('重做绘制失败:', e?.message || e)
    await reloadDrawings(entry.pageNum)
  } finally {
    endSave()
  }
}

async function redoErase(entry: Extract<HistoryEntry, { type: 'erase' }>) {
  beginSave()
  try {
    const savedFragments = await replacePersistedDrawings(
      entry.pageNum,
      entry.replacements.map((replacement) => replacement.original),
      entry.replacements.flatMap((replacement) =>
        replacement.fragments.map((fragment) => stripDrawingId(fragment))),
    )

    let fragmentOffset = 0
    const savedGroups = entry.replacements.map((replacement) => {
      const group = savedFragments.slice(fragmentOffset, fragmentOffset + replacement.fragments.length)
      fragmentOffset += replacement.fragments.length
      return group
    })
    entry.replacements = entry.replacements.map((replacement, index) => ({
      ...replacement,
      fragments: savedGroups[index],
    }))

    const current = drawingsByPage.value.get(entry.pageNum) || []
    const originalIDs = new Set(
      entry.replacements
        .map((replacement) => replacement.original.id)
        .filter((id): id is number => id !== undefined),
    )
    const fragmentsByIndex = new Map(
      entry.replacements.map((replacement, index) => [replacement.index, savedGroups[index]]),
    )
    const nextDrawings = current.flatMap((drawing, index) => {
      if (drawing.id && originalIDs.has(drawing.id)) {
        return fragmentsByIndex.get(index) || []
      }
      return [drawing]
    })
    setPageDrawings(entry.pageNum, nextDrawings)
    moveRedoEntryToHistory(entry)
    redrawPage(entry.pageNum)
  } catch (e: any) {
    console.warn('重做擦除失败:', e?.message || e)
    await reloadDrawings(entry.pageNum)
  } finally {
    endSave()
  }
}

async function redoEdit(entry: Extract<HistoryEntry, { type: 'edit' }>) {
  beginSave()
  try {
    const replacement = (await replacePersistedDrawings(
      entry.pageNum,
      [entry.original],
      [stripDrawingId(entry.replacement)],
    ))[0]
    entry.replacement = replacement
    const current = drawingsByPage.value.get(entry.pageNum) || []
    const currentIndex = findCurrentDrawingIndex(current, entry.index, entry.original)
    if (currentIndex >= 0) {
      const next = [...current]
      next[currentIndex] = replacement
      setPageDrawings(entry.pageNum, next)
      entry.index = currentIndex
    }
    moveRedoEntryToHistory(entry)
    redrawPage(entry.pageNum)
  } catch (e: any) {
    console.warn('重做形状编辑失败:', e?.message || e)
    await reloadDrawings(entry.pageNum)
  } finally {
    endSave()
  }
}

async function redoLastStroke() {
  const entry = redoStack.pop()
  if (!entry) return
  updateHistoryState()
  if (entry.type === 'draw') {
    await redoDraw(entry)
    return
  }
  if (entry.type === 'erase') {
    await redoErase(entry)
    return
  }
  await redoEdit(entry)
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

async function getPageText(pageNum: number): Promise<string> {
  if (!pdfDoc) return ''
  const cached = pageTextCache.get(pageNum)
  if (cached !== undefined) return cached

  const page = await pdfDoc.getPage(pageNum)
  const content = await page.getTextContent()
  const text = content.items
    .map((item) => ('str' in item ? item.str : ''))
    .join('')
  pageTextCache.set(pageNum, text)
  return text
}

async function searchDocument(query: string): Promise<Array<{ page: number; excerpt: string }>> {
  const keyword = query.trim()
  if (!keyword || !pdfDoc) return []

  const normalizedKeyword = keyword.toLowerCase()
  const results: Array<{ page: number; excerpt: string }> = []
  for (let pageNum = 1; pageNum <= pdfDoc.numPages; pageNum++) {
    const text = await getPageText(pageNum)
    const lowerText = text.toLowerCase()
    const index = lowerText.indexOf(normalizedKeyword)
    if (index < 0) continue
    const start = Math.max(0, index - 24)
    const end = Math.min(text.length, index + keyword.length + 36)
    results.push({
      page: pageNum,
      excerpt: `${start > 0 ? '...' : ''}${text.slice(start, end)}${end < text.length ? '...' : ''}`,
    })
  }
  return results
}

async function setZoom(value: number) {
  const next = Math.max(30, Math.min(240, Math.round(value)))
  if (next === zoom.value) return
  zoom.value = next
  await renderAllPages()
}

function zoomIn() {
  void setZoom(zoom.value + 10)
}

function zoomOut() {
  void setZoom(zoom.value - 10)
}

async function fitWidth() {
  if (!pdfDoc || !container.value) return
  const page = await pdfDoc.getPage(Math.max(1, Math.min(currentPage.value, pageCount.value)))
  const viewport = page.getViewport({ scale: 1 })
  const availableWidth = Math.max(240, container.value.clientWidth - 32)
  await setZoom((availableWidth / viewport.width) * 100)
}

async function fitPage() {
  if (!pdfDoc || !container.value) return
  const page = await pdfDoc.getPage(Math.max(1, Math.min(currentPage.value, pageCount.value)))
  const viewport = page.getViewport({ scale: 1 })
  const availableWidth = Math.max(240, container.value.clientWidth - 32)
  const availableHeight = Math.max(240, container.value.clientHeight - 32)
  await setZoom(Math.min(
    (availableWidth / viewport.width) * 100,
    (availableHeight / viewport.height) * 100,
  ))
}

function canvasToJpegBytes(canvas: HTMLCanvasElement, quality = 0.92): Promise<Uint8Array> {
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (!blob) {
        reject(new Error('无法生成 PDF 页面图像'))
        return
      }
      blob.arrayBuffer()
        .then((buffer) => resolve(new Uint8Array(buffer)))
        .catch(reject)
    }, 'image/jpeg', quality)
  })
}

function downloadBlob(blob: Blob, fileName: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  link.remove()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}

function sanitizeDownloadFileName(value: string) {
  const normalized = value
    .trim()
    .replace(/[\\/]+/g, '-')
    .replace(/[<>:"|?*\x00-\x1F]/g, '-')
    .replace(/\s+/g, ' ')
    .replace(/[. ]+$/g, '')
    .slice(0, 120)

  return normalized || `document-${props.documentId}`
}

async function resolveDocumentTitle() {
  const propTitle = props.documentTitle?.trim()
  if (propTitle) {
    resolvedDocumentTitle.value = propTitle
    return propTitle
  }
  if (resolvedDocumentTitle.value) return resolvedDocumentTitle.value

  try {
    const response = await request.get(`/documents/${props.documentId}`)
    const title = String(response.data?.title || '').trim()
    if (title) {
      resolvedDocumentTitle.value = title
      return title
    }
  } catch (e: any) {
    console.warn('获取导出文件名失败:', e?.message || e)
  }

  return ''
}

async function exportedFileName() {
  const title = sanitizeDownloadFileName(await resolveDocumentTitle())
  return /\.pdf$/i.test(title) ? title : `${title}.pdf`
}

async function exportAnnotatedPDF() {
  if (exporting.value || loading.value || error.value) return false
  if (pendingSaveCount.value > 0) {
    throw new Error('批注仍在保存中，请稍后再导出')
  }

  exporting.value = true
  emit('exportStateChange', true)
  const previousZoom = zoom.value

  try {
    const exportScale = 1.5
    if (zoom.value !== exportScale * 100) {
      zoom.value = exportScale * 100
      await renderAllPages()
    }

    const pages = []
    for (let pageNum = 1; pageNum <= pageCount.value; pageNum++) {
      const canvas = pageRefs.value[pageNum - 1]?.exportCompositeCanvas?.()
      if (!canvas) throw new Error(`第 ${pageNum} 页尚未渲染完成`)
      pages.push({
        width: canvas.width,
        height: canvas.height,
        jpegData: await canvasToJpegBytes(canvas),
      })
    }

    downloadBlob(createRasterPdf(pages), await exportedFileName())
    return true
  } finally {
    if (zoom.value !== previousZoom) {
      zoom.value = previousZoom
      await renderAllPages()
    }
    exporting.value = false
    emit('exportStateChange', false)
  }
}

function jumpToPage(page: number) {
  if (!container.value || pageCount.value < 1) return
  const targetPage = Math.max(1, Math.min(pageCount.value, Math.round(page)))
  const wrapper = container.value.querySelector<HTMLElement>(
    `.pdf-page-wrapper[data-page="${targetPage}"]`,
  )
  if (wrapper) {
    container.value.scrollTo({ top: wrapper.offsetTop, behavior: 'smooth' })
  }
  currentPage.value = targetPage
}

async function focusAnnotation(annotationId: number, page: number) {
  const targetPage = Math.max(1, Math.min(pageCount.value || 1, Math.round(page || 1)))
  jumpToPage(targetPage)
  await new Promise((resolve) => requestAnimationFrame(resolve))
  pageRefs.value[targetPage - 1]?.focusAnnotation?.(annotationId)
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
    updateHistoryState()
  }
})
watch(pageCount, (count) => emit('pageCountChange', count))
watch(zoom, (value) => emit('zoomChange', value))

onMounted(() => {
  updateHistoryState()
  emit('zoomChange', zoom.value)
  emit('savingStateChange', false)
  emit('exportStateChange', false)
  void loadPDF()
})

defineExpose({
  zoomIn,
  zoomOut,
  setZoom,
  fitWidth,
  fitPage,
  searchDocument,
  jumpToPage,
  focusAnnotation,
  undoLastStroke,
  redoLastStroke,
  applySelectedTextStyle,
  exportAnnotatedPDF,
  zoom,
  currentPage,
  pageCount,
  canUndo,
  canRedo,
})
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
        :page-num="pageNum"
        :scale="zoom / 100"
        :drawings="drawingsByPage.get(pageNum) || []"
        :active-tool="props.activeTool"
        :shape-type="props.shapeType"
        :erase-mode="props.eraseMode"
        :pen-color="props.penColor"
        :pen-width="props.penWidth"
        :text-size="props.textSize"
        :eraser-size="props.penWidth"
        @drawing-created="(drawing: Drawing) => { void addDrawing(pageNum, drawing) }"
        @drawings-replaced="(replacements: DrawingReplacement[]) => {
          void replaceDrawings(pageNum, replacements)
        }"
        @drawing-edited="(index: number, drawing: Drawing) => {
          void editDrawing(pageNum, index, drawing)
        }"
        @text-selection-changed="(drawing: TextDrawing | null) => {
          emit('textSelectionChange', drawing)
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
