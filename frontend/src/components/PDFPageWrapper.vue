<script setup lang="ts">
import { ref, watch } from 'vue'
import * as pdfjs from 'pdfjs-dist'
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
  isFreehandStroke,
  isShapeDrawing,
} from './pdfDrawingTypes'
import type {
  Drawing,
  DrawingReplacement,
  FreehandStroke,
  PDFActiveTool,
  Point,
  ShapeDrawing,
  ShapeType,
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

export type {
  Drawing,
  DrawingReplacement,
  FreehandStroke,
  Point,
  ShapeDrawing,
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
  eraserSize: number
}>()

const emit = defineEmits<{
  drawingCreated: [drawing: Drawing]
  drawingsReplaced: [replacements: DrawingReplacement[]]
  shapeEdited: [index: number, shape: ShapeDrawing]
}>()

interface ShapeEditGesture {
  index: number
  mode: 'move' | 'resize'
  handle: ShapeHandle | null
  originPoint: Point
  original: ShapeDrawing
  preview: ShapeDrawing
}

const pdfCanvas = ref<HTMLCanvasElement | null>(null)
const drawCanvas = ref<HTMLCanvasElement | null>(null)
const overlayCanvas = ref<HTMLCanvasElement | null>(null)
const eraserCursor = ref<HTMLElement | null>(null)
const areaSelection = ref<HTMLElement | null>(null)

let pdfPage: pdfjs.PDFPageProxy | null = null
let isDrawing = false
let currentStroke: FreehandStroke | null = null
let currentShape: ShapeDrawing | null = null
let selectedShapeIndex: number | null = null
let shapeEdit: ShapeEditGesture | null = null
let eraserPath: Point[] = []
let queuedPreviewPoints: Point[] = []
let lastPreviewPoint: Point | null = null
let previewFrame: number | null = null
let pendingReplacements: DrawingReplacement[] = []
let awaitingReplacementCommit = false
let gestureRect: DOMRect | null = null
let areaStart: Point | null = null
let areaEnd: Point | null = null

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
  ctx.save()
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.strokeStyle = drawing.color
  ctx.lineWidth = drawing.width
  if (drawing.tool === 'highlighter') ctx.globalAlpha = 0.35
  ctx.stroke(getDrawingPath(drawing))
  ctx.restore()
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

function drawShapeOverlay(shape: ShapeDrawing, includeShape: boolean, includeHandles: boolean) {
  const canvas = overlayCanvas.value
  const ctx = canvas?.getContext('2d')
  if (!canvas || !ctx) return
  ctx.clearRect(0, 0, canvas.width, canvas.height)

  if (includeShape) drawDrawing(ctx, shape)
  if (!includeHandles) return

  const bounds = getShapeBounds(shape)
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

function selectedShape(): ShapeDrawing | null {
  if (selectedShapeIndex === null) return null
  const drawing = props.drawings[selectedShapeIndex]
  return drawing && isShapeDrawing(drawing) ? drawing : null
}

function redrawSelection() {
  if (shapeEdit) {
    drawShapeOverlay(shapeEdit.preview, true, true)
    return
  }
  const shape = selectedShape()
  if (shape && props.activeTool === 'select') {
    drawShapeOverlay(shape, false, true)
  } else {
    clearOverlay()
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
  const bounds = isShapeDrawing(drawing)
    ? getShapeBounds(drawing)
    : getStrokeBounds(drawing)
  drawingBoundsCache.set(drawing, bounds)
  return bounds
}

function splitDrawingByPath(drawing: Drawing, path: Point[], radius: number): Drawing[] {
  if (isFreehandStroke(drawing)) {
    return splitStrokeByEraserPath(drawing, path, radius)
  }
  return splitShapeByEraserPath(drawing, path, radius) || [drawing]
}

function splitDrawingByRect(drawing: Drawing, start: Point, end: Point): Drawing[] {
  if (isFreehandStroke(drawing)) {
    return splitStrokeByRect(drawing, start, end)
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
    const unchanged = fragments.length === 1 && fragments[0] === drawing
    if (!unchanged) replacements.push({ index, fragments })
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

function resetGestureState(redraw = false) {
  if (previewFrame !== null) cancelAnimationFrame(previewFrame)
  previewFrame = null
  queuedPreviewPoints = []
  lastPreviewPoint = null
  eraserPath = []
  areaStart = null
  areaEnd = null
  gestureRect = null
  currentShape = null
  shapeEdit = null
  updateEraserCursor({ clientX: 0, clientY: 0 }, false)
  hideAreaSelection()
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

function topmostShapeAt(point: Point, tolerance: number): number | null {
  for (let index = props.drawings.length - 1; index >= 0; index--) {
    const drawing = props.drawings[index]
    if (isShapeDrawing(drawing) && hitTestShape(drawing, point, tolerance)) return index
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

function onPointerDown(event: PointerEvent) {
  if (props.activeTool === 'none' || awaitingReplacementCommit || !drawCanvas.value) return
  gestureRect = drawCanvas.value.getBoundingClientRect()
  drawCanvas.value.setPointerCapture(event.pointerId)
  isDrawing = true
  const point = pageToCanvas(event)

  if (props.activeTool === 'eraser') {
    selectedShapeIndex = null
    clearOverlay()
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

  if (props.activeTool === 'shape') {
    selectedShapeIndex = null
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
    const selected = selectedShape()
    if (selected && selectedShapeIndex !== null) {
      const handle = hitTestShapeHandle(selected, point, tolerance)
      if (handle || hitTestShape(selected, point, tolerance)) {
        startShapeEdit(selectedShapeIndex, selected, point, handle)
        return
      }
    }

    const hitIndex = topmostShapeAt(point, tolerance)
    if (hitIndex !== null) {
      const shape = props.drawings[hitIndex] as ShapeDrawing
      startShapeEdit(hitIndex, shape, point, null)
    } else {
      selectedShapeIndex = null
      clearOverlay()
      isDrawing = false
      finishPointerCapture(event)
      gestureRect = null
    }
    return
  }

  currentStroke = {
    tool: props.activeTool,
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

function shapesEqual(left: ShapeDrawing, right: ShapeDrawing): boolean {
  return left.start.x === right.start.x &&
    left.start.y === right.start.y &&
    left.end.x === right.end.x &&
    left.end.y === right.end.y
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
    if (!shapesEqual(edit.original, edit.preview)) {
      awaitingReplacementCommit = true
      drawShapeOverlay(edit.preview, true, true)
      emit('shapeEdited', edit.index, edit.preview)
    } else {
      redrawDrawings()
      redrawSelection()
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
  resetGestureState(true)
}

watch(() => props.drawings, () => {
  if (selectedShapeIndex !== null && selectedShapeIndex >= props.drawings.length) {
    selectedShapeIndex = null
  }
  if (awaitingReplacementCommit) {
    awaitingReplacementCommit = false
    pendingReplacements = []
    resetGestureState()
  }
  redrawDrawings()
  redrawSelection()
})

watch(() => [props.activeTool, props.eraseMode, props.shapeType], () => {
  if (isDrawing || awaitingReplacementCommit) return
  pendingReplacements = []
  if (props.activeTool !== 'select') selectedShapeIndex = null
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

defineExpose({ renderPDF, redrawDrawings, clearDrawLayer, cancelPendingReplacement })
</script>

<template>
  <div class="pdf-page-wrapper">
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
    />
    <canvas ref="overlayCanvas" class="pdf-overlay-layer" />
    <div ref="eraserCursor" class="eraser-cursor" />
    <div ref="areaSelection" class="eraser-area-selection" />
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

.pdf-overlay-layer {
  pointer-events: none;
  z-index: 1;
}

.eraser-cursor,
.eraser-area-selection {
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

.eraser-area-selection {
  box-sizing: border-box;
  border: 1.5px dashed #dc2626;
  background: rgba(220, 38, 38, 0.08);
}
</style>
