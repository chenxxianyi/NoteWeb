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

export interface Point { x: number; y: number }
export interface Stroke {
  id?: number
  tool: 'pen' | 'highlighter'
  color: string
  width: number
  points: Point[]
}

export interface StrokeReplacement {
  index: number
  fragments: Stroke[]
}

const props = defineProps<{
  pageNum: number
  scale: number
  strokes: Stroke[]
  activeTool: 'none' | 'pen' | 'highlighter' | 'eraser'
  eraseMode: 'freehand' | 'area'
  penColor: string
  penWidth: number
  eraserSize: number
}>()

const emit = defineEmits<{
  strokeCreated: [stroke: Stroke]
  strokesReplaced: [replacements: StrokeReplacement[]]
}>()

const pdfCanvas = ref<HTMLCanvasElement | null>(null)
const drawCanvas = ref<HTMLCanvasElement | null>(null)
const eraserCursor = ref<HTMLElement | null>(null)
const areaSelection = ref<HTMLElement | null>(null)

let pdfPage: pdfjs.PDFPageProxy | null = null
let isDrawing = false
let currentStroke: Stroke | null = null
let eraserPath: Point[] = []
let queuedPreviewPoints: Point[] = []
let lastPreviewPoint: Point | null = null
let previewFrame: number | null = null
let pendingReplacements: StrokeReplacement[] = []
let awaitingReplacementCommit = false
let gestureRect: DOMRect | null = null
let areaStart: Point | null = null
let areaEnd: Point | null = null

const strokeBoundsCache = new WeakMap<Stroke, Bounds | null>()
const strokePathCache = new WeakMap<Stroke, Path2D>()

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

  const dc = drawCanvas.value
  if (dc) {
    dc.width = viewport.width
    dc.height = viewport.height
  }
  redrawStrokes()
}

function getStrokePath(stroke: Stroke): Path2D {
  const cached = strokePathCache.get(stroke)
  if (cached) return cached

  const path = new Path2D()
  if (stroke.points.length > 0) {
    path.moveTo(stroke.points[0].x, stroke.points[0].y)
    for (let index = 1; index < stroke.points.length; index++) {
      path.lineTo(stroke.points[index].x, stroke.points[index].y)
    }
  }
  strokePathCache.set(stroke, path)
  return path
}

function redrawStrokes(previewReplacements: StrokeReplacement[] = []) {
  const canvas = drawCanvas.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  ctx.clearRect(0, 0, canvas.width, canvas.height)
  const replacementMap = new Map(
    previewReplacements.map((replacement) => [replacement.index, replacement.fragments]),
  )
  for (let index = 0; index < props.strokes.length; index++) {
    const fragments = replacementMap.get(index)
    if (fragments) {
      for (const fragment of fragments) drawStroke(ctx, fragment)
    } else {
      drawStroke(ctx, props.strokes[index])
    }
  }
}

function drawStroke(ctx: CanvasRenderingContext2D, stroke: Stroke) {
  if (stroke.points.length < 2) return
  ctx.save()
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.strokeStyle = stroke.color
  ctx.lineWidth = stroke.width
  if (stroke.tool === 'highlighter') ctx.globalAlpha = 0.35
  ctx.stroke(getStrokePath(stroke))
  ctx.restore()
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

function getCachedStrokeBounds(stroke: Stroke): Bounds | null {
  if (strokeBoundsCache.has(stroke)) return strokeBoundsCache.get(stroke) || null
  const bounds = getStrokeBounds(stroke)
  strokeBoundsCache.set(stroke, bounds)
  return bounds
}

function pointsEqual(a: Point[], b: Point[]): boolean {
  return a.length === b.length && a.every((point, index) =>
    point.x === b[index].x && point.y === b[index].y)
}

function collectReplacements(
  gestureBounds: Bounds | null,
  split: (stroke: Stroke) => Stroke[],
): StrokeReplacement[] {
  if (!gestureBounds) return []
  const replacements: StrokeReplacement[] = []
  for (let index = 0; index < props.strokes.length; index++) {
    const stroke = props.strokes[index]
    const bounds = getCachedStrokeBounds(stroke)
    if (!bounds || !boundsIntersect(bounds, gestureBounds)) continue

    const fragments = split(stroke)
    const unchanged = fragments.length === 1 && pointsEqual(fragments[0].points, stroke.points)
    if (!unchanged) replacements.push({ index, fragments })
  }
  return replacements
}

function findFreehandReplacements(path: Point[]): StrokeReplacement[] {
  const radius = eraserRadius()
  const simplifiedPath = simplifyPathByDistance(path, Math.max(2, radius * 0.35))
  return collectReplacements(
    getPathBounds(simplifiedPath, radius),
    (stroke) => splitStrokeByEraserPath(stroke, simplifiedPath, radius),
  )
}

function findAreaReplacements(start: Point, end: Point): StrokeReplacement[] {
  return collectReplacements(
    getPathBounds([start, end]),
    (stroke) => splitStrokeByRect(stroke, start, end),
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
  updateEraserCursor({ clientX: 0, clientY: 0 }, false)
  hideAreaSelection()
  if (redraw) redrawStrokes()
}

function onPointerDown(event: PointerEvent) {
  if (props.activeTool === 'none' || awaitingReplacementCommit || !drawCanvas.value) return
  gestureRect = drawCanvas.value.getBoundingClientRect()
  drawCanvas.value.setPointerCapture(event.pointerId)
  isDrawing = true
  const point = pageToCanvas(event)

  if (props.activeTool === 'eraser') {
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
  if (drawCanvas.value?.hasPointerCapture(event.pointerId)) {
    drawCanvas.value.releasePointerCapture(event.pointerId)
  }

  if (props.activeTool === 'eraser') {
    if (props.eraseMode === 'area') {
      if (areaStart && areaEnd) pendingReplacements = findAreaReplacements(areaStart, areaEnd)
      hideAreaSelection()
      if (pendingReplacements.length > 0) redrawStrokes(pendingReplacements)
    } else {
      finishPreviewFrame()
      pendingReplacements = findFreehandReplacements(eraserPath)
      updateEraserCursor(event, false)
    }

    gestureRect = null
    if (pendingReplacements.length > 0) {
      awaitingReplacementCommit = true
      emit('strokesReplaced', pendingReplacements)
    } else {
      pendingReplacements = []
      resetGestureState(true)
    }
    return
  }

  if (currentStroke && currentStroke.points.length > 1) {
    emit('strokeCreated', { ...currentStroke, points: [...currentStroke.points] })
  }
  currentStroke = null
  gestureRect = null
}

function onPointerCancel(event: PointerEvent) {
  if (!isDrawing) return
  isDrawing = false
  if (drawCanvas.value?.hasPointerCapture(event.pointerId)) {
    drawCanvas.value.releasePointerCapture(event.pointerId)
  }
  currentStroke = null
  pendingReplacements = []
  resetGestureState(true)
}

watch(() => props.strokes, () => {
  if (awaitingReplacementCommit) {
    awaitingReplacementCommit = false
    pendingReplacements = []
    resetGestureState()
  }
  redrawStrokes()
})

watch(() => [props.activeTool, props.eraseMode], () => {
  if (isDrawing || awaitingReplacementCommit) return
  pendingReplacements = []
  resetGestureState(true)
})

watch(() => props.scale, () => {
  if (pdfPage) {
    const viewport = pdfPage.getViewport({ scale: props.scale })
    const canvas = drawCanvas.value
    if (canvas) {
      canvas.width = viewport.width
      canvas.height = viewport.height
    }
  }
})

function clearDrawLayer() {
  const canvas = drawCanvas.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (ctx) ctx.clearRect(0, 0, canvas.width, canvas.height)
}

function cancelPendingReplacement() {
  awaitingReplacementCommit = false
  pendingReplacements = []
  resetGestureState(true)
}

defineExpose({ renderPDF, redrawStrokes, clearDrawLayer, cancelPendingReplacement })
</script>

<template>
  <div class="pdf-page-wrapper">
    <canvas ref="pdfCanvas" class="pdf-page-layer" />
    <canvas
      ref="drawCanvas"
      :class="['pdf-draw-layer', { 'pdf-draw-layer--eraser': activeTool === 'eraser' }]"
      @pointerdown="onPointerDown"
      @pointermove="onPointerMove"
      @pointerup="onPointerUp"
      @pointercancel="onPointerCancel"
    />
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
.pdf-draw-layer {
  display: block;
  max-width: 100%;
  height: auto;
}

.pdf-draw-layer {
  position: absolute;
  top: 0;
  left: 0;
  cursor: crosshair;
}

.pdf-draw-layer--eraser {
  cursor: none;
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
