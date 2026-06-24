<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import * as pdfjs from 'pdfjs-dist'
import { Trash2 } from 'lucide-vue-next'
import {
  boundsIntersect,
  distance,
  getPathBounds,
  getStrokeBounds,
  simplifyPathByDistance,
  splitStrokeByEraserPath,
  splitStrokeByRect,
} from './pdfAnnotationGeometry'
import type { Bounds } from './pdfAnnotationGeometry'
import {
  drawingsEqual,
  drawingReplacementChangesDrawing,
  isFreehandStroke,
  isShapeDrawing,
  isTextDrawing,
} from './pdfDrawingTypes'
import type {
  Drawing,
  DrawingReplacement,
  FreehandStroke,
  PDFActiveTool,
  Point,
  ShapeDrawing,
  ShapeType,
  TextDrawing,
  TextErasure,
} from './pdfDrawingTypes'
import {
  getShapeBounds,
  getShapeHandles,
  hitTestShape,
  hitTestShapeHandle,
  moveShape,
  resizeShape,
  shapeToContours,
  splitShapeByEraserPath,
  splitShapeByRect,
} from './pdfShapeGeometry'
import type { ShapeHandle } from './pdfShapeGeometry'
import {
  eraseTextByEraserPath,
  eraseTextByRect,
  getTextBounds,
  getTextResizeHandlePoint,
  hitTestText,
  hitTestTextResizeHandle,
  MIN_TEXT_WIDTH,
  moveText,
  resizeTextBox,
  TEXT_LINE_HEIGHT,
  TEXT_PADDING,
  wrapTextLines,
} from './pdfTextGeometry'

export type {
  Drawing,
  DrawingReplacement,
  FreehandStroke,
  Point,
  ShapeDrawing,
  TextDrawing,
}

const props = defineProps<{
  pageNum: number
  scale: number
  drawings: Drawing[]
  activeTool: PDFActiveTool
  shapeType: ShapeType
  eraseMode: 'freehand' | 'area'
  penColor: string
  penWidth: number
  textSize: number
  eraserSize: number
}>()

const emit = defineEmits<{
  drawingCreated: [drawing: Drawing]
  drawingsReplaced: [replacements: DrawingReplacement[]]
  drawingEdited: [index: number, drawing: Drawing]
  textSelectionChanged: [drawing: TextDrawing | null]
}>()

interface ShapeEditGesture {
  index: number
  mode: 'move' | 'resize'
  handle: ShapeHandle | null
  originPoint: Point
  original: ShapeDrawing
  preview: ShapeDrawing
}

interface TextEditGesture {
  index: number
  mode: 'move' | 'resize'
  originPoint: Point
  original: TextDrawing
  preview: TextDrawing
}

interface TextEditorState {
  mode: 'create' | 'edit'
  index: number | null
  original: TextDrawing | null
  draft: TextDrawing
}

interface SelectionDragGesture {
  indices: number[]
  originPoint: Point
  originals: Drawing[]
  previews: Drawing[]
}

const pdfCanvas = ref<HTMLCanvasElement | null>(null)
const drawCanvas = ref<HTMLCanvasElement | null>(null)
const overlayCanvas = ref<HTMLCanvasElement | null>(null)
const eraserCursor = ref<HTMLElement | null>(null)
const areaSelection = ref<HTMLElement | null>(null)
const selectionDeleteButton = ref<HTMLElement | null>(null)
const textArea = ref<HTMLTextAreaElement | null>(null)
const textEditor = ref<TextEditorState | null>(null)

let pdfPage: pdfjs.PDFPageProxy | null = null
let isDrawing = false
let currentStroke: FreehandStroke | null = null
let currentShape: ShapeDrawing | null = null
let selectedShapeIndex: number | null = null
let selectedTextIndex: number | null = null
let selectedDrawingIndices: number[] = []
let shapeEdit: ShapeEditGesture | null = null
let textEdit: TextEditGesture | null = null
let selectionDrag: SelectionDragGesture | null = null
let ignoreNextTextBlur = false
let focusFrameTimer: number | null = null
let eraserPath: Point[] = []
let queuedPreviewPoints: Point[] = []
let lastPreviewPoint: Point | null = null
let previewFrame: number | null = null
let pendingReplacements: DrawingReplacement[] = []
let awaitingReplacementCommit = false
let gestureRect: DOMRect | null = null
let areaStart: Point | null = null
let areaEnd: Point | null = null
let selectionStart: Point | null = null
let selectionEnd: Point | null = null

const drawingBoundsCache = new WeakMap<Drawing, Bounds | null>()
const drawingPathCache = new WeakMap<Drawing, Path2D>()

async function renderPDF(pdfDoc: pdfjs.PDFDocumentProxy) {
  const canvas = pdfCanvas.value
  if (!canvas) return
  const page = await pdfDoc.getPage(props.pageNum)
  pdfPage = page
  const viewport = page.getViewport({ scale: props.scale })
  canvas.width = viewport.width
  canvas.height = viewport.height
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  await page.render({ canvas, canvasContext: ctx, viewport }).promise

  for (const layer of [drawCanvas.value, overlayCanvas.value]) {
    if (layer) {
      layer.width = viewport.width
      layer.height = viewport.height
    }
  }
  redrawDrawings()
  redrawSelection()
}

function drawingContours(drawing: Drawing): Point[][] {
  if (isTextDrawing(drawing)) return []
  return isShapeDrawing(drawing)
    ? shapeToContours(drawing)
    : [drawing.points]
}

function getDrawingPath(drawing: Drawing): Path2D {
  const cached = drawingPathCache.get(drawing)
  if (cached) return cached

  const path = new Path2D()
  for (const contour of drawingContours(drawing)) {
    if (contour.length === 0) continue
    path.moveTo(contour[0].x, contour[0].y)
    for (let index = 1; index < contour.length; index++) {
      path.lineTo(contour[index].x, contour[index].y)
    }
  }
  drawingPathCache.set(drawing, path)
  return path
}

function drawDrawing(ctx: CanvasRenderingContext2D, drawing: Drawing) {
  if (isTextDrawing(drawing)) {
    drawTextDrawing(ctx, drawing)
    return
  }

  ctx.save()
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.strokeStyle = drawing.color
  ctx.lineWidth = drawing.width
  if (drawing.tool === 'highlighter') ctx.globalAlpha = 0.35
  ctx.stroke(getDrawingPath(drawing))
  ctx.restore()
}

function textFont(text: TextDrawing): string {
  return `${text.fontSize}px -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`
}

function drawTextContent(ctx: CanvasRenderingContext2D, text: TextDrawing) {
  ctx.save()
  ctx.fillStyle = text.color
  ctx.font = textFont(text)
  ctx.textBaseline = 'top'
  const lineHeight = text.fontSize * TEXT_LINE_HEIGHT
  const lines = wrapTextLines(
    text.text,
    text.fontSize,
    text.width,
    (value) => ctx.measureText(value).width,
  )
  lines.forEach((line, index) => {
    ctx.fillText(
      line,
      text.x + TEXT_PADDING,
      text.y + TEXT_PADDING + index * lineHeight,
    )
  })
  ctx.restore()
}

function drawTextErasure(ctx: CanvasRenderingContext2D, erasure: TextErasure) {
  ctx.save()
  ctx.globalCompositeOperation = 'destination-out'
  if (erasure.type === 'path') {
    ctx.lineCap = 'round'
    ctx.lineJoin = 'round'
    ctx.lineWidth = erasure.radius * 2
    const points = erasure.points
    if (points.length === 1) {
      ctx.beginPath()
      ctx.arc(points[0].x, points[0].y, erasure.radius, 0, Math.PI * 2)
      ctx.fill()
    } else if (points.length > 1) {
      ctx.beginPath()
      ctx.moveTo(points[0].x, points[0].y)
      for (let index = 1; index < points.length; index++) {
        ctx.lineTo(points[index].x, points[index].y)
      }
      ctx.stroke()
    }
  } else {
    const left = Math.min(erasure.start.x, erasure.end.x)
    const top = Math.min(erasure.start.y, erasure.end.y)
    ctx.fillRect(
      left,
      top,
      Math.abs(erasure.start.x - erasure.end.x),
      Math.abs(erasure.start.y - erasure.end.y),
    )
  }
  ctx.restore()
}

function drawTextDrawing(ctx: CanvasRenderingContext2D, text: TextDrawing) {
  if (!text.erasures || text.erasures.length === 0) {
    drawTextContent(ctx, text)
    return
  }

  const bounds = getTextBounds(text)
  const width = Math.ceil(Math.max(1, bounds.right - bounds.left))
  const height = Math.ceil(Math.max(1, bounds.bottom - bounds.top))
  const layer = document.createElement('canvas')
  layer.width = width
  layer.height = height
  const layerCtx = layer.getContext('2d')
  if (!layerCtx) {
    drawTextContent(ctx, text)
    return
  }

  layerCtx.translate(-bounds.left, -bounds.top)
  drawTextContent(layerCtx, text)
  text.erasures.forEach((erasure) => drawTextErasure(layerCtx, erasure))
  ctx.drawImage(layer, bounds.left, bounds.top)
}

function redrawDrawings(
  previewReplacements: DrawingReplacement[] = [],
  excludeIndex: number | null = null,
) {
  const canvas = drawCanvas.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  ctx.clearRect(0, 0, canvas.width, canvas.height)
  const replacementMap = new Map(
    previewReplacements.map((replacement) => [replacement.index, replacement.fragments]),
  )
  for (let index = 0; index < props.drawings.length; index++) {
    if (index === excludeIndex) continue
    const fragments = replacementMap.get(index)
    if (fragments) {
      for (const fragment of fragments) drawDrawing(ctx, fragment)
    } else {
      drawDrawing(ctx, props.drawings[index])
    }
  }
}

function clearOverlay() {
  const canvas = overlayCanvas.value
  const ctx = canvas?.getContext('2d')
  if (canvas && ctx) ctx.clearRect(0, 0, canvas.width, canvas.height)
}

function canvasUnits(screenPixels: number): number {
  const canvas = drawCanvas.value
  if (!canvas) return screenPixels
  const rect = currentCanvasRect()
  return screenPixels * (canvas.width / rect.width)
}

function hideSelectionDeleteButton() {
  if (selectionDeleteButton.value) selectionDeleteButton.value.style.display = 'none'
}

function updateSelectionDeleteButton(bounds: Bounds | null) {
  const button = selectionDeleteButton.value
  const canvas = drawCanvas.value
  if (!button || !canvas || !bounds || props.activeTool !== 'select') {
    hideSelectionDeleteButton()
    return
  }

  const rect = currentCanvasRect()
  const scaleX = rect.width / canvas.width
  const scaleY = rect.height / canvas.height
  const buttonSize = 34
  const gap = 8
  const pageWidth = rect.width
  const pageHeight = rect.height
  const preferredLeft = bounds.right * scaleX - buttonSize
  const preferredTop = bounds.top * scaleY - buttonSize - gap
  const fallbackTop = bounds.top * scaleY + gap
  const left = Math.max(gap, Math.min(preferredLeft, pageWidth - buttonSize - gap))
  const top = Math.max(gap, Math.min(
    preferredTop >= gap ? preferredTop : fallbackTop,
    pageHeight - buttonSize - gap,
  ))

  button.style.display = 'inline-flex'
  button.style.transform = `translate3d(${left}px, ${top}px, 0)`
}

const textEditorStyle = computed(() => {
  const editor = textEditor.value
  const canvas = drawCanvas.value
  if (!editor || !canvas) return {}
  const rect = canvas.getBoundingClientRect()
  const scaleX = rect.width / canvas.width
  const scaleY = rect.height / canvas.height
  const bounds = getTextBounds(editor.draft)
  return {
    transform: `translate3d(${bounds.left * scaleX}px, ${bounds.top * scaleY}px, 0)`,
    width: `${(bounds.right - bounds.left) * scaleX}px`,
    height: `${(bounds.bottom - bounds.top) * scaleY}px`,
    minHeight: `${(bounds.bottom - bounds.top) * scaleY}px`,
    color: editor.draft.color,
    fontSize: `${editor.draft.fontSize * scaleY}px`,
    lineHeight: String(TEXT_LINE_HEIGHT),
  }
})

function drawShapeOverlay(shape: ShapeDrawing, includeShape: boolean, includeHandles: boolean) {
  const canvas = overlayCanvas.value
  const ctx = canvas?.getContext('2d')
  if (!canvas || !ctx) return
  ctx.clearRect(0, 0, canvas.width, canvas.height)

  if (includeShape) drawDrawing(ctx, shape)
  if (!includeHandles) return

  const bounds = getShapeBounds(shape)
  if (props.activeTool === 'select' && !includeShape) updateSelectionDeleteButton(bounds)
  const padding = shape.width / 2
  ctx.save()
  ctx.strokeStyle = '#2563eb'
  ctx.lineWidth = canvasUnits(1)
  ctx.setLineDash([canvasUnits(5), canvasUnits(3)])
  ctx.strokeRect(
    bounds.left + padding,
    bounds.top + padding,
    Math.max(0, bounds.right - bounds.left - shape.width),
    Math.max(0, bounds.bottom - bounds.top - shape.width),
  )
  ctx.setLineDash([])
  const radius = canvasUnits(4.5)
  for (const handle of getShapeHandles(shape)) {
    ctx.beginPath()
    ctx.fillStyle = '#ffffff'
    ctx.strokeStyle = '#2563eb'
    ctx.lineWidth = canvasUnits(1.5)
    ctx.arc(handle.point.x, handle.point.y, radius, 0, Math.PI * 2)
    ctx.fill()
    ctx.stroke()
  }
  ctx.restore()
}

function normalizedBounds(start: Point, end: Point): Bounds {
  return {
    left: Math.min(start.x, end.x),
    top: Math.min(start.y, end.y),
    right: Math.max(start.x, end.x),
    bottom: Math.max(start.y, end.y),
  }
}

function boundsContainBounds(outer: Bounds, inner: Bounds): boolean {
  return inner.left >= outer.left &&
    inner.right <= outer.right &&
    inner.top >= outer.top &&
    inner.bottom <= outer.bottom
}

function mergeBounds(boundsList: Bounds[]): Bounds | null {
  if (boundsList.length === 0) return null
  return boundsList.slice(1).reduce((merged, bounds) => ({
    left: Math.min(merged.left, bounds.left),
    top: Math.min(merged.top, bounds.top),
    right: Math.max(merged.right, bounds.right),
    bottom: Math.max(merged.bottom, bounds.bottom),
  }), { ...boundsList[0] })
}

function drawBoundsFrame(
  ctx: CanvasRenderingContext2D,
  bounds: Bounds,
  style: 'item' | 'group',
) {
  const padding = style === 'group' ? canvasUnits(5) : canvasUnits(2)
  ctx.save()
  ctx.strokeStyle = style === 'group' ? '#2563eb' : 'rgba(37, 99, 235, 0.72)'
  ctx.lineWidth = canvasUnits(style === 'group' ? 1.4 : 1)
  ctx.setLineDash(style === 'group'
    ? [canvasUnits(6), canvasUnits(4)]
    : [canvasUnits(3), canvasUnits(3)])
  ctx.strokeRect(
    bounds.left - padding,
    bounds.top - padding,
    Math.max(0, bounds.right - bounds.left + padding * 2),
    Math.max(0, bounds.bottom - bounds.top + padding * 2),
  )
  ctx.restore()
}

function drawMultiSelectionOverlay(drawings: Drawing[]) {
  const canvas = overlayCanvas.value
  const ctx = canvas?.getContext('2d')
  if (!canvas || !ctx) return
  ctx.clearRect(0, 0, canvas.width, canvas.height)

  const boundsList = drawings
    .map((drawing) => getDrawingBounds(drawing))
    .filter((bounds): bounds is Bounds => Boolean(bounds))
  if (boundsList.length === 0) return

  if (drawings.length === 1) {
    drawBoundsFrame(ctx, boundsList[0], 'group')
    updateSelectionDeleteButton(boundsList[0])
    return
  }

  boundsList.forEach((bounds) => drawBoundsFrame(ctx, bounds, 'item'))
  const merged = mergeBounds(boundsList)
  if (merged) {
    drawBoundsFrame(ctx, merged, 'group')
    updateSelectionDeleteButton(merged)
  }
}

function drawTextOverlay(text: TextDrawing, includeText: boolean, includeHandles: boolean) {
  const canvas = overlayCanvas.value
  const ctx = canvas?.getContext('2d')
  if (!canvas || !ctx) return
  ctx.clearRect(0, 0, canvas.width, canvas.height)

  if (includeText) drawDrawing(ctx, text)
  if (!includeHandles) return

  const bounds = getTextBounds(text)
  if (props.activeTool === 'select' && !includeText) updateSelectionDeleteButton(bounds)
  ctx.save()
  ctx.strokeStyle = '#2563eb'
  ctx.lineWidth = canvasUnits(1)
  ctx.setLineDash([canvasUnits(5), canvasUnits(3)])
  ctx.strokeRect(
    bounds.left,
    bounds.top,
    Math.max(0, bounds.right - bounds.left),
    Math.max(0, bounds.bottom - bounds.top),
  )
  ctx.setLineDash([])

  const handle = getTextResizeHandlePoint(text)
  const size = canvasUnits(8)
  ctx.fillStyle = '#ffffff'
  ctx.strokeStyle = '#2563eb'
  ctx.lineWidth = canvasUnits(1.5)
  ctx.fillRect(handle.x - size / 2, handle.y - size / 2, size, size)
  ctx.strokeRect(handle.x - size / 2, handle.y - size / 2, size, size)
  ctx.restore()
}

function selectedShape(): ShapeDrawing | null {
  if (selectedShapeIndex === null) return null
  const drawing = props.drawings[selectedShapeIndex]
  return drawing && isShapeDrawing(drawing) ? drawing : null
}

function selectedText(): TextDrawing | null {
  if (selectedTextIndex === null) return null
  const drawing = props.drawings[selectedTextIndex]
  return drawing && isTextDrawing(drawing) ? drawing : null
}

function selectedDrawings(): Drawing[] {
  return selectedDrawingIndices
    .map((index) => props.drawings[index])
    .filter((drawing): drawing is Drawing => Boolean(drawing))
}

function notifyTextSelection() {
  emit('textSelectionChanged', selectedText())
}

function clearSelectedText(notify = true) {
  selectedTextIndex = null
  if (notify) emit('textSelectionChanged', null)
}

function clearMultiSelection() {
  selectedDrawingIndices = []
}

function setSingleSelection(index: number) {
  selectedDrawingIndices = [index]
}

function setAreaSelection(indices: number[]) {
  selectedShapeIndex = null
  clearSelectedText()
  selectedDrawingIndices = [...indices]
}

function redrawSelection() {
  hideSelectionDeleteButton()
  if (selectionDrag) {
    drawMultiSelectionOverlay(selectionDrag.previews)
    return
  }
  if (shapeEdit) {
    drawShapeOverlay(shapeEdit.preview, true, true)
    return
  }
  if (textEdit) {
    drawTextOverlay(textEdit.preview, true, true)
    return
  }
  const shape = selectedShape()
  if (shape && props.activeTool === 'select') {
    drawShapeOverlay(shape, false, true)
    return
  }
  const text = selectedText()
  if (text && props.activeTool === 'select') {
    drawTextOverlay(text, false, true)
    return
  }
  const multiSelected = selectedDrawings()
  if (multiSelected.length > 0 && props.activeTool === 'select') {
    drawMultiSelectionOverlay(multiSelected)
    return
  }
  clearOverlay()
}

function clearFocusFrameTimer() {
  if (focusFrameTimer !== null) {
    window.clearTimeout(focusFrameTimer)
    focusFrameTimer = null
  }
}

function currentCanvasRect(): DOMRect {
  if (gestureRect) return gestureRect
  return drawCanvas.value!.getBoundingClientRect()
}

function pageToCanvas(event: Pick<PointerEvent, 'clientX' | 'clientY'>): Point {
  const canvas = drawCanvas.value!
  const rect = currentCanvasRect()
  return {
    x: (event.clientX - rect.left) * (canvas.width / rect.width),
    y: (event.clientY - rect.top) * (canvas.height / rect.height),
  }
}

function eraserRadius() {
  return Math.max(4, props.eraserSize / 2)
}

function updateEraserCursor(event: Pick<PointerEvent, 'clientX' | 'clientY'>, visible = true) {
  const element = eraserCursor.value
  const canvas = drawCanvas.value
  if (!element || !canvas) return
  if (!visible) {
    element.style.display = 'none'
    return
  }

  const rect = currentCanvasRect()
  const diameter = props.eraserSize * (rect.width / canvas.width)
  element.style.display = 'block'
  element.style.width = `${diameter}px`
  element.style.height = `${diameter}px`
  element.style.transform = `translate3d(${event.clientX - rect.left - diameter / 2}px, ${event.clientY - rect.top - diameter / 2}px, 0)`
}

function hideAreaSelection() {
  if (areaSelection.value) areaSelection.value.style.display = 'none'
}

function updateAreaSelection(start: Point, end: Point) {
  const element = areaSelection.value
  const canvas = drawCanvas.value
  if (!element || !canvas) return
  const rect = currentCanvasRect()
  const scaleX = rect.width / canvas.width
  const scaleY = rect.height / canvas.height
  const left = Math.min(start.x, end.x) * scaleX
  const top = Math.min(start.y, end.y) * scaleY
  const width = Math.abs(start.x - end.x) * scaleX
  const height = Math.abs(start.y - end.y) * scaleY

  element.style.display = 'block'
  element.style.transform = `translate3d(${left}px, ${top}px, 0)`
  element.style.width = `${width}px`
  element.style.height = `${height}px`
}

function appendEraserPoint(point: Point) {
  const last = eraserPath[eraserPath.length - 1]
  if (last && distance(last, point) < 0.75) return
  eraserPath.push(point)
  queuedPreviewPoints.push(point)
}

function flushFreehandPreview() {
  previewFrame = null
  if (queuedPreviewPoints.length === 0) return
  const canvas = drawCanvas.value
  const ctx = canvas?.getContext('2d')
  if (!canvas || !ctx) return

  const points = queuedPreviewPoints
  queuedPreviewPoints = []
  ctx.save()
  ctx.globalCompositeOperation = 'destination-out'
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.lineWidth = props.eraserSize

  if (!lastPreviewPoint) {
    const first = points[0]
    ctx.beginPath()
    ctx.arc(first.x, first.y, eraserRadius(), 0, Math.PI * 2)
    ctx.fill()
    lastPreviewPoint = first
  }

  ctx.beginPath()
  ctx.moveTo(lastPreviewPoint.x, lastPreviewPoint.y)
  for (const point of points) ctx.lineTo(point.x, point.y)
  ctx.stroke()
  lastPreviewPoint = points[points.length - 1]
  ctx.restore()
}

function scheduleFreehandPreview() {
  if (previewFrame !== null) return
  previewFrame = requestAnimationFrame(flushFreehandPreview)
}

function finishPreviewFrame() {
  if (previewFrame !== null) {
    cancelAnimationFrame(previewFrame)
    previewFrame = null
  }
  flushFreehandPreview()
}

function getDrawingBounds(drawing: Drawing): Bounds | null {
  if (drawingBoundsCache.has(drawing)) return drawingBoundsCache.get(drawing) || null
  const bounds = isTextDrawing(drawing)
    ? getTextBounds(drawing)
    : isShapeDrawing(drawing)
      ? getShapeBounds(drawing)
      : getStrokeBounds(drawing)
  drawingBoundsCache.set(drawing, bounds)
  return bounds
}

function moveDrawing(drawing: Drawing, dx: number, dy: number): Drawing {
  if (isShapeDrawing(drawing)) return moveShape(drawing, dx, dy)
  if (isTextDrawing(drawing)) return moveText(drawing, dx, dy)
  return {
    ...drawing,
    points: drawing.points.map((point) => ({ x: point.x + dx, y: point.y + dy })),
  }
}

function areaSelectedIndices(start: Point, end: Point): number[] {
  const rect = normalizedBounds(start, end)
  const indices: number[] = []
  for (let index = 0; index < props.drawings.length; index++) {
    const bounds = getDrawingBounds(props.drawings[index])
    if (!bounds) continue
    if (boundsContainBounds(rect, bounds) || boundsIntersect(rect, bounds)) {
      indices.push(index)
    }
  }
  return indices
}

function selectedIndexAt(point: Point): number | null {
  for (const index of [...selectedDrawingIndices].reverse()) {
    const drawing = props.drawings[index]
    const bounds = drawing ? getDrawingBounds(drawing) : null
    if (bounds && boundsContainBounds(bounds, {
      left: point.x,
      right: point.x,
      top: point.y,
      bottom: point.y,
    })) {
      return index
    }
  }
  return null
}

function commitSelectionDrag(edit: SelectionDragGesture) {
  const replacements = edit.indices.map((index, offset) => ({
    index,
    fragments: [edit.previews[offset]],
  }))
  awaitingReplacementCommit = true
  redrawDrawings(replacements)
  drawMultiSelectionOverlay(edit.previews)
  emit('drawingsReplaced', replacements)
}

function splitDrawingByPath(drawing: Drawing, path: Point[], radius: number): Drawing[] {
  if (isFreehandStroke(drawing)) {
    return splitStrokeByEraserPath(drawing, path, radius)
  }
  if (isTextDrawing(drawing)) {
    return eraseTextByEraserPath(drawing, path, radius)
  }
  return splitShapeByEraserPath(drawing, path, radius) || [drawing]
}

function splitDrawingByRect(drawing: Drawing, start: Point, end: Point): Drawing[] {
  if (isFreehandStroke(drawing)) {
    return splitStrokeByRect(drawing, start, end)
  }
  if (isTextDrawing(drawing)) {
    return eraseTextByRect(drawing, start, end)
  }
  return splitShapeByRect(drawing, start, end) || [drawing]
}

function collectReplacements(
  gestureBounds: Bounds | null,
  split: (drawing: Drawing) => Drawing[],
): DrawingReplacement[] {
  if (!gestureBounds) return []
  const replacements: DrawingReplacement[] = []
  for (let index = 0; index < props.drawings.length; index++) {
    const drawing = props.drawings[index]
    const bounds = getDrawingBounds(drawing)
    if (!bounds || !boundsIntersect(bounds, gestureBounds)) continue
    const fragments = split(drawing)
    if (drawingReplacementChangesDrawing(drawing, fragments)) {
      replacements.push({ index, fragments })
    }
  }
  return replacements
}

function findFreehandReplacements(path: Point[]): DrawingReplacement[] {
  const radius = eraserRadius()
  const simplifiedPath = simplifyPathByDistance(path, Math.max(2, radius * 0.35))
  return collectReplacements(
    getPathBounds(simplifiedPath, radius),
    (drawing) => splitDrawingByPath(drawing, simplifiedPath, radius),
  )
}

function findAreaReplacements(start: Point, end: Point): DrawingReplacement[] {
  return collectReplacements(
    getPathBounds([start, end]),
    (drawing) => splitDrawingByRect(drawing, start, end),
  )
}

function focusTextEditor() {
  void nextTick(() => {
    const element = textArea.value
    if (!element) return
    requestAnimationFrame(() => {
      if (textArea.value !== element) return
      ignoreNextTextBlur = true
      element.focus({ preventScroll: true })
      element.select()
      requestAnimationFrame(() => {
        ignoreNextTextBlur = false
      })
    })
  })
}

function defaultTextWidth(point: Point): number {
  const canvas = drawCanvas.value
  if (!canvas) return 220
  return Math.max(MIN_TEXT_WIDTH, Math.min(canvas.width - point.x - TEXT_PADDING, canvasUnits(220)))
}

function startNewTextEditor(point: Point) {
  selectedShapeIndex = null
  clearSelectedText()
  clearOverlay()
  textEditor.value = {
    mode: 'create',
    index: null,
    original: null,
    draft: {
      tool: 'text',
      color: props.penColor,
      fontSize: props.textSize,
      text: '',
      x: point.x,
      y: point.y,
      width: defaultTextWidth(point),
      height: props.textSize * TEXT_LINE_HEIGHT + TEXT_PADDING * 2,
    },
  }
  focusTextEditor()
}

function startExistingTextEditor(index: number, text: TextDrawing) {
  selectedShapeIndex = null
  selectedTextIndex = index
  notifyTextSelection()
  textEditor.value = {
    mode: 'edit',
    index,
    original: text,
    draft: {
      ...text,
      text: text.text,
    },
  }
  redrawDrawings([], index)
  clearOverlay()
  focusTextEditor()
}

function updateTextEditorValue(event: Event) {
  if (!textEditor.value) return
  textEditor.value.draft = {
    ...textEditor.value.draft,
    text: (event.target as HTMLTextAreaElement).value,
  }
}

function cancelTextEditor(redraw = true) {
  textEditor.value = null
  if (redraw) {
    redrawDrawings()
    redrawSelection()
  }
}

function commitTextEditor() {
  const editor = textEditor.value
  if (!editor) return
  const text = editor.draft.text.replace(/\s+$/g, '')
  if (!text.trim()) {
    cancelTextEditor(editor.mode === 'edit')
    return
  }

  const drawing: TextDrawing = {
    ...editor.draft,
    text,
    width: Math.max(MIN_TEXT_WIDTH, editor.draft.width),
    height: editor.draft.height,
  }
  textEditor.value = null

  if (editor.mode === 'create') {
    emit('drawingCreated', drawing)
    return
  }

  if (
    editor.index !== null &&
    editor.original &&
    !drawingsEqual(editor.original, drawing)
  ) {
    awaitingReplacementCommit = true
    selectedTextIndex = editor.index
    redrawDrawings([{ index: editor.index, fragments: [drawing] }])
    drawTextOverlay(drawing, false, true)
    emit('drawingEdited', editor.index, drawing)
    return
  }

  redrawDrawings()
  redrawSelection()
}

function onTextEditorKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    event.preventDefault()
    cancelTextEditor()
    return
  }
  if (event.key === 'Enter' && (event.ctrlKey || event.metaKey)) {
    event.preventDefault()
    commitTextEditor()
  }
}

function onTextEditorBlur() {
  if (ignoreNextTextBlur) {
    focusTextEditor()
    return
  }
  commitTextEditor()
}

function resetGestureState(redraw = false) {
  clearFocusFrameTimer()
  if (previewFrame !== null) cancelAnimationFrame(previewFrame)
  previewFrame = null
  queuedPreviewPoints = []
  lastPreviewPoint = null
  eraserPath = []
  areaStart = null
  areaEnd = null
  selectionStart = null
  selectionEnd = null
  gestureRect = null
  currentShape = null
  shapeEdit = null
  textEdit = null
  selectionDrag = null
  updateEraserCursor({ clientX: 0, clientY: 0 }, false)
  hideAreaSelection()
  hideSelectionDeleteButton()
  if (redraw) {
    redrawDrawings()
    redrawSelection()
  }
}

function finishPointerCapture(event: PointerEvent) {
  if (drawCanvas.value?.hasPointerCapture(event.pointerId)) {
    drawCanvas.value.releasePointerCapture(event.pointerId)
  }
}

function topmostTextAt(point: Point, tolerance: number): number | null {
  for (let index = props.drawings.length - 1; index >= 0; index--) {
    const drawing = props.drawings[index]
    if (isTextDrawing(drawing) && hitTestText(drawing, point, tolerance)) return index
  }
  return null
}

function startShapeEdit(
  index: number,
  shape: ShapeDrawing,
  point: Point,
  handle: ShapeHandle | null,
) {
  selectedShapeIndex = index
  clearSelectedText()
  clearMultiSelection()
  shapeEdit = {
    index,
    mode: handle ? 'resize' : 'move',
    handle,
    originPoint: point,
    original: shape,
    preview: shape,
  }
  redrawDrawings([], index)
  drawShapeOverlay(shape, true, true)
}

function startTextEdit(
  index: number,
  text: TextDrawing,
  point: Point,
  mode: 'move' | 'resize',
) {
  selectedTextIndex = index
  notifyTextSelection()
  selectedShapeIndex = null
  clearMultiSelection()
  textEdit = {
    index,
    mode,
    originPoint: point,
    original: text,
    preview: text,
  }
  redrawDrawings([], index)
  drawTextOverlay(text, true, true)
}

function deleteSelectedDrawing() {
  if (awaitingReplacementCommit || textEditor.value) return
  const indices = selectedDrawingIndices.length > 0
    ? [...selectedDrawingIndices]
    : selectedTextIndex !== null
      ? [selectedTextIndex]
      : selectedShapeIndex !== null
        ? [selectedShapeIndex]
        : []
  if (indices.length === 0) return
  awaitingReplacementCommit = true
  clearSelectedText()
  selectedShapeIndex = null
  clearMultiSelection()
  clearOverlay()
  hideSelectionDeleteButton()
  const replacements = indices
    .filter((index) => props.drawings[index])
    .map((index) => ({ index, fragments: [] }))
  redrawDrawings(replacements)
  emit('drawingsReplaced', replacements)
}

function applySelectedTextStyle(style: { color: string; fontSize: number }): boolean {
  if (awaitingReplacementCommit || textEditor.value || selectedTextIndex === null) return false
  const text = selectedText()
  if (!text) return false

  const next: TextDrawing = {
    ...text,
    color: style.color,
    fontSize: style.fontSize,
    height: Math.max(text.height, style.fontSize * TEXT_LINE_HEIGHT + TEXT_PADDING * 2),
  }
  if (drawingsEqual(text, next)) return true

  awaitingReplacementCommit = true
  redrawDrawings([{ index: selectedTextIndex, fragments: [next] }])
  drawTextOverlay(next, false, true)
  emit('drawingEdited', selectedTextIndex, next)
  return true
}

function onPointerDown(event: PointerEvent) {
  if (props.activeTool === 'none' || awaitingReplacementCommit || !drawCanvas.value) return
  gestureRect = drawCanvas.value.getBoundingClientRect()
  drawCanvas.value.setPointerCapture(event.pointerId)
  isDrawing = true
  const point = pageToCanvas(event)

  if (props.activeTool === 'eraser') {
    selectedShapeIndex = null
    clearSelectedText()
    clearMultiSelection()
    clearOverlay()
    hideSelectionDeleteButton()
    pendingReplacements = []
    if (props.eraseMode === 'area') {
      areaStart = point
      areaEnd = point
      updateAreaSelection(point, point)
    } else {
      appendEraserPoint(point)
      updateEraserCursor(event)
      scheduleFreehandPreview()
    }
    return
  }

  if (props.activeTool === 'text') {
    event.preventDefault()
    if (textEditor.value) commitTextEditor()
    isDrawing = false
    finishPointerCapture(event)
    gestureRect = null
    if (!awaitingReplacementCommit) startNewTextEditor(point)
    return
  }

  if (props.activeTool === 'shape') {
    selectedShapeIndex = null
    clearSelectedText()
    clearMultiSelection()
    hideSelectionDeleteButton()
    currentShape = {
      tool: 'shape',
      shapeType: props.shapeType,
      color: props.penColor,
      width: props.penWidth,
      start: point,
      end: point,
    }
    drawShapeOverlay(currentShape, true, false)
    return
  }

  if (props.activeTool === 'select') {
    const tolerance = canvasUnits(8)
    const selectedMultiIndex = selectedIndexAt(point)
    if (selectedMultiIndex !== null && selectedDrawingIndices.length > 0) {
      selectionDrag = {
        indices: [...selectedDrawingIndices],
        originPoint: point,
        originals: selectedDrawingIndices.map((index) => props.drawings[index]),
        previews: selectedDrawingIndices.map((index) => props.drawings[index]),
      }
      redrawDrawings([], null)
      drawMultiSelectionOverlay(selectionDrag.previews)
      return
    }

    const selectedTextDrawing = selectedText()
    if (selectedTextDrawing && selectedTextIndex !== null) {
      if (hitTestTextResizeHandle(selectedTextDrawing, point, tolerance)) {
        startTextEdit(selectedTextIndex, selectedTextDrawing, point, 'resize')
        return
      }
      if (hitTestText(selectedTextDrawing, point, tolerance)) {
        startTextEdit(selectedTextIndex, selectedTextDrawing, point, 'move')
        return
      }
    }

    const selected = selectedShape()
    if (selected && selectedShapeIndex !== null) {
      const handle = hitTestShapeHandle(selected, point, tolerance)
      if (handle || hitTestShape(selected, point, tolerance)) {
        startShapeEdit(selectedShapeIndex, selected, point, handle)
        return
      }
    }

    for (let index = props.drawings.length - 1; index >= 0; index--) {
      const drawing = props.drawings[index]
      if (isTextDrawing(drawing) && hitTestText(drawing, point, tolerance)) {
        const mode = hitTestTextResizeHandle(drawing, point, tolerance) ? 'resize' : 'move'
        setSingleSelection(index)
        startTextEdit(index, drawing, point, mode)
        return
      }
      if (isShapeDrawing(drawing) && hitTestShape(drawing, point, tolerance)) {
        setSingleSelection(index)
        startShapeEdit(index, drawing, point, null)
        return
      }
    }

    selectedShapeIndex = null
    clearSelectedText()
    clearMultiSelection()
    hideSelectionDeleteButton()
    selectionStart = point
    selectionEnd = point
    updateAreaSelection(point, point)
    return
  }

  currentStroke = {
    tool: props.activeTool === 'highlighter' ? 'highlighter' : 'pen',
    color: props.penColor,
    width: props.penWidth,
    points: [point],
  }
}

function onPointerMove(event: PointerEvent) {
  if (!isDrawing || !drawCanvas.value) return

  if (props.activeTool === 'eraser') {
    if (props.eraseMode === 'area') {
      areaEnd = pageToCanvas(event)
      if (areaStart) updateAreaSelection(areaStart, areaEnd)
      return
    }
    const events = event.getCoalescedEvents?.() || []
    const samples = events.length > 0 ? events : [event]
    for (const sample of samples) appendEraserPoint(pageToCanvas(sample))
    updateEraserCursor(event)
    scheduleFreehandPreview()
    return
  }

  const point = pageToCanvas(event)
  if (props.activeTool === 'shape' && currentShape) {
    currentShape = { ...currentShape, end: point }
    drawShapeOverlay(currentShape, true, false)
    return
  }

  if (props.activeTool === 'select' && shapeEdit) {
    shapeEdit.preview = shapeEdit.mode === 'move'
      ? moveShape(
          shapeEdit.original,
          point.x - shapeEdit.originPoint.x,
          point.y - shapeEdit.originPoint.y,
        )
      : resizeShape(shapeEdit.original, shapeEdit.handle!, point)
    drawShapeOverlay(shapeEdit.preview, true, true)
    return
  }

  if (props.activeTool === 'select' && textEdit) {
    textEdit.preview = textEdit.mode === 'move'
      ? moveText(
          textEdit.original,
          point.x - textEdit.originPoint.x,
          point.y - textEdit.originPoint.y,
        )
      : resizeTextBox(textEdit.original, point)
    drawTextOverlay(textEdit.preview, true, true)
    return
  }

  if (props.activeTool === 'select' && selectionDrag) {
    const dx = point.x - selectionDrag.originPoint.x
    const dy = point.y - selectionDrag.originPoint.y
    selectionDrag.previews = selectionDrag.originals.map((drawing) => moveDrawing(drawing, dx, dy))
    redrawDrawings(selectionDrag.indices.map((index, offset) => ({
      index,
      fragments: [selectionDrag!.previews[offset]],
    })))
    drawMultiSelectionOverlay(selectionDrag.previews)
    return
  }

  if (props.activeTool === 'select' && selectionStart) {
    selectionEnd = point
    updateAreaSelection(selectionStart, selectionEnd)
    return
  }

  if (!currentStroke) return
  currentStroke.points.push(point)
  const ctx = drawCanvas.value.getContext('2d')
  if (!ctx) return
  ctx.save()
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.strokeStyle = currentStroke.color
  ctx.lineWidth = currentStroke.width
  if (currentStroke.tool === 'highlighter') ctx.globalAlpha = 0.35
  const points = currentStroke.points
  ctx.beginPath()
  ctx.moveTo(points[points.length - 2].x, points[points.length - 2].y)
  ctx.lineTo(points[points.length - 1].x, points[points.length - 1].y)
  ctx.stroke()
  ctx.restore()
}

function onPointerUp(event: PointerEvent) {
  if (!isDrawing) return
  isDrawing = false
  finishPointerCapture(event)

  if (props.activeTool === 'eraser') {
    if (props.eraseMode === 'area') {
      if (areaStart && areaEnd) pendingReplacements = findAreaReplacements(areaStart, areaEnd)
      hideAreaSelection()
      if (pendingReplacements.length > 0) redrawDrawings(pendingReplacements)
    } else {
      finishPreviewFrame()
      pendingReplacements = findFreehandReplacements(eraserPath)
      updateEraserCursor(event, false)
      if (pendingReplacements.length > 0) redrawDrawings(pendingReplacements)
    }
    gestureRect = null
    if (pendingReplacements.length > 0) {
      awaitingReplacementCommit = true
      emit('drawingsReplaced', pendingReplacements)
    } else {
      pendingReplacements = []
      resetGestureState(true)
    }
    return
  }

  if (props.activeTool === 'shape' && currentShape) {
    const created = currentShape
    currentShape = null
    gestureRect = null
    clearOverlay()
    if (distance(created.start, created.end) >= canvasUnits(3)) {
      emit('drawingCreated', created)
    }
    return
  }

  if (props.activeTool === 'select' && shapeEdit) {
    const edit = shapeEdit
    shapeEdit = null
    gestureRect = null
    if (!drawingsEqual(edit.original, edit.preview)) {
      awaitingReplacementCommit = true
      drawShapeOverlay(edit.preview, true, true)
      emit('drawingEdited', edit.index, edit.preview)
    } else {
      redrawDrawings()
      redrawSelection()
    }
    return
  }

  if (props.activeTool === 'select' && textEdit) {
    const edit = textEdit
    textEdit = null
    gestureRect = null
    if (!drawingsEqual(edit.original, edit.preview)) {
      awaitingReplacementCommit = true
      selectedTextIndex = edit.index
      drawTextOverlay(edit.preview, true, true)
      emit('drawingEdited', edit.index, edit.preview)
    } else {
      redrawDrawings()
      redrawSelection()
    }
    return
  }

  if (props.activeTool === 'select' && selectionDrag) {
    const edit = selectionDrag
    selectionDrag = null
    gestureRect = null
    const changed = edit.originals.some((drawing, index) =>
      !drawingsEqual(drawing, edit.previews[index]))
    if (changed) {
      commitSelectionDrag(edit)
    } else {
      redrawDrawings()
      redrawSelection()
    }
    return
  }

  if (props.activeTool === 'select' && selectionStart) {
    const start = selectionStart
    const end = selectionEnd || pageToCanvas(event)
    const dragDistance = distance(start, end)
    selectionStart = null
    selectionEnd = null
    gestureRect = null
    hideAreaSelection()
    if (dragDistance >= canvasUnits(4)) {
      setAreaSelection(areaSelectedIndices(start, end))
      redrawDrawings()
      redrawSelection()
    } else {
      clearMultiSelection()
      clearOverlay()
    }
    return
  }

  if (currentStroke && currentStroke.points.length > 1) {
    emit('drawingCreated', { ...currentStroke, points: [...currentStroke.points] })
  }
  currentStroke = null
  gestureRect = null
}

function onPointerCancel(event: PointerEvent) {
  if (!isDrawing) return
  isDrawing = false
  finishPointerCapture(event)
  currentStroke = null
  pendingReplacements = []
  clearMultiSelection()
  hideSelectionDeleteButton()
  resetGestureState(true)
}

function onDoubleClick(event: MouseEvent) {
  if (props.activeTool !== 'select' || awaitingReplacementCommit || !drawCanvas.value) return
  gestureRect = drawCanvas.value.getBoundingClientRect()
  const point = pageToCanvas(event)
  const hitIndex = topmostTextAt(point, canvasUnits(8))
  gestureRect = null
  if (hitIndex === null) return
  const drawing = props.drawings[hitIndex]
  if (drawing && isTextDrawing(drawing)) startExistingTextEditor(hitIndex, drawing)
}

function onWindowKeyDown(event: KeyboardEvent) {
  const target = event.target as HTMLElement | null
  const tagName = target?.tagName
  if (tagName === 'INPUT' || tagName === 'TEXTAREA' || target?.isContentEditable) return
  if (props.activeTool !== 'select') return
  if (event.key !== 'Delete' && event.key !== 'Backspace') return
  if (
    selectedTextIndex === null &&
    selectedShapeIndex === null &&
    selectedDrawingIndices.length === 0
  ) return
  event.preventDefault()
  deleteSelectedDrawing()
}

watch(() => props.drawings, () => {
  if (selectedShapeIndex !== null && selectedShapeIndex >= props.drawings.length) {
    selectedShapeIndex = null
  }
  if (selectedTextIndex !== null && selectedTextIndex >= props.drawings.length) {
    clearSelectedText()
  } else if (selectedTextIndex !== null) {
    notifyTextSelection()
  }
  selectedDrawingIndices = selectedDrawingIndices.filter((index) => index < props.drawings.length)
  if (awaitingReplacementCommit) {
    awaitingReplacementCommit = false
    pendingReplacements = []
    clearMultiSelection()
    hideSelectionDeleteButton()
    resetGestureState()
  }
  redrawDrawings()
  redrawSelection()
})

watch(() => [props.activeTool, props.eraseMode, props.shapeType, props.textSize], () => {
  if (textEditor.value && props.activeTool !== 'text') commitTextEditor()
  if (isDrawing || awaitingReplacementCommit) return
  pendingReplacements = []
  if (props.activeTool !== 'select') {
    selectedShapeIndex = null
    clearSelectedText()
    clearMultiSelection()
    hideSelectionDeleteButton()
  }
  resetGestureState(true)
})

watch(() => props.scale, () => {
  if (pdfPage) {
    const viewport = pdfPage.getViewport({ scale: props.scale })
    for (const canvas of [drawCanvas.value, overlayCanvas.value]) {
      if (canvas) {
        canvas.width = viewport.width
        canvas.height = viewport.height
      }
    }
  }
})

function clearDrawLayer() {
  for (const canvas of [drawCanvas.value, overlayCanvas.value]) {
    const ctx = canvas?.getContext('2d')
    if (canvas && ctx) ctx.clearRect(0, 0, canvas.width, canvas.height)
  }
}

function cancelPendingReplacement() {
  awaitingReplacementCommit = false
  pendingReplacements = []
  resetGestureState(true)
}

function focusAnnotation(annotationId: number): boolean {
  const index = props.drawings.findIndex((drawing) => drawing.id === annotationId)
  const drawing = props.drawings[index]
  if (!drawing) return false

  clearFocusFrameTimer()
  selectedShapeIndex = null
  clearSelectedText()
  selectedDrawingIndices = [index]
  redrawDrawings()
  drawMultiSelectionOverlay([drawing])
  focusFrameTimer = window.setTimeout(() => {
    focusFrameTimer = null
    if (props.activeTool !== 'select') {
      clearMultiSelection()
      clearOverlay()
      hideSelectionDeleteButton()
    }
  }, 2200)
  return true
}

function exportCompositeCanvas(): HTMLCanvasElement | null {
  const base = pdfCanvas.value
  const drawings = drawCanvas.value
  if (!base || !drawings) return null

  const output = document.createElement('canvas')
  output.width = base.width
  output.height = base.height
  const ctx = output.getContext('2d')
  if (!ctx) return null

  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, output.width, output.height)
  ctx.drawImage(base, 0, 0)
  ctx.drawImage(drawings, 0, 0)
  return output
}

onMounted(() => {
  window.addEventListener('keydown', onWindowKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onWindowKeyDown)
  clearFocusFrameTimer()
})

defineExpose({
  renderPDF,
  redrawDrawings,
  clearDrawLayer,
  cancelPendingReplacement,
  applySelectedTextStyle,
  exportCompositeCanvas,
  focusAnnotation,
})
</script>

<template>
  <div class="pdf-page-wrapper" :data-page="pageNum">
    <canvas ref="pdfCanvas" class="pdf-page-layer" />
    <canvas
      ref="drawCanvas"
      :class="[
        'pdf-draw-layer',
        `pdf-draw-layer--${activeTool}`,
      ]"
      @pointerdown="onPointerDown"
      @pointermove="onPointerMove"
      @pointerup="onPointerUp"
      @pointercancel="onPointerCancel"
      @dblclick="onDoubleClick"
    />
    <canvas ref="overlayCanvas" class="pdf-overlay-layer" />
    <textarea
      v-if="textEditor"
      ref="textArea"
      class="pdf-text-editor"
      :style="textEditorStyle"
      :value="textEditor.draft.text"
      placeholder="输入文本"
      spellcheck="false"
      @input="updateTextEditorValue"
      @pointerdown.stop
      @keydown.stop="onTextEditorKeydown"
      @blur="onTextEditorBlur"
    />
    <div ref="eraserCursor" class="eraser-cursor" />
    <div
      ref="areaSelection"
      :class="[
        'area-selection',
        activeTool === 'select' ? 'area-selection--select' : 'area-selection--eraser',
      ]"
    />
    <button
      ref="selectionDeleteButton"
      class="selection-delete-button"
      type="button"
      title="删除选中批注"
      aria-label="删除选中批注"
      @pointerdown.stop
      @click.stop="deleteSelectedDrawing"
    >
      <Trash2 />
    </button>
  </div>
</template>

<style scoped>
.pdf-page-wrapper {
  position: relative;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.12);
  border-radius: 2px;
  max-width: 100%;
  touch-action: none;
  user-select: none;
}

.pdf-page-layer,
.pdf-draw-layer,
.pdf-overlay-layer {
  display: block;
  max-width: 100%;
  height: auto;
}

.pdf-draw-layer,
.pdf-overlay-layer {
  position: absolute;
  top: 0;
  left: 0;
}

.pdf-draw-layer {
  cursor: crosshair;
}

.pdf-draw-layer--eraser {
  cursor: none;
}

.pdf-draw-layer--select {
  cursor: default;
}

.pdf-draw-layer--text {
  cursor: text;
}

.pdf-overlay-layer {
  pointer-events: none;
  z-index: 1;
}

.pdf-text-editor {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 3;
  box-sizing: border-box;
  padding: 4px;
  border: 1.5px solid #2563eb;
  border-radius: 3px;
  outline: none;
  resize: none;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.86);
  box-shadow: 0 2px 10px rgba(37, 99, 235, 0.16);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  white-space: pre-wrap;
  user-select: text;
}

.pdf-text-editor::placeholder {
  color: rgba(37, 99, 235, 0.55);
}

.eraser-cursor,
.area-selection {
  position: absolute;
  top: 0;
  left: 0;
  display: none;
  pointer-events: none;
  z-index: 2;
  will-change: transform, width, height;
}

.eraser-cursor {
  box-sizing: border-box;
  border: 1.5px solid #dc2626;
  border-radius: 50%;
  background: rgba(220, 38, 38, 0.08);
}

.area-selection {
  box-sizing: border-box;
}

.area-selection--eraser {
  border: 1.5px dashed #dc2626;
  background: rgba(220, 38, 38, 0.08);
}

.area-selection--select {
  border: 1.5px dashed #2563eb;
  background: rgba(37, 99, 235, 0.08);
}

.selection-delete-button {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 4;
  display: none;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  min-width: 34px;
  padding: 0;
  border: 1px solid rgba(185, 28, 28, 0.22);
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.94);
  color: #dc2626;
  box-shadow: 0 6px 18px rgba(61, 46, 36, 0.16);
  cursor: pointer;
  pointer-events: auto;
  transition: background 0.16s, border-color 0.16s, color 0.16s, box-shadow 0.16s;
  will-change: transform;
}

.selection-delete-button:hover {
  border-color: rgba(220, 38, 38, 0.36);
  background: #fee2e2;
  color: #b91c1c;
  box-shadow: 0 8px 22px rgba(185, 28, 28, 0.18);
}

.selection-delete-button:focus-visible {
  outline: 2px solid #dc2626;
  outline-offset: 2px;
}

.selection-delete-button svg {
  width: 16px;
  height: 16px;
  stroke-width: 2;
}
</style>
