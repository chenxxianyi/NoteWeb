<script setup lang="ts">
import { ref, watch } from 'vue'
import * as pdfjs from 'pdfjs-dist'
import { splitStrokeByEraserPath, splitStrokeByRect } from './pdfAnnotationGeometry'

/* ── Types ── */
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
let pdfPage: pdfjs.PDFPageProxy | null = null
let isDrawing = false
let currentStroke: Stroke | null = null
let eraserPath: Point[] = []
let pendingReplacements: StrokeReplacement[] = []
let awaitingReplacementCommit = false
// Area select
let areaStart: Point | null = null
let areaEnd: Point | null = null

/* ── Render PDF page ── */
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

  // Sync draw layer size
  const dc = drawCanvas.value
  if (dc) {
    dc.width = viewport.width
    dc.height = viewport.height
  }
  // Draw existing strokes
  redrawStrokes()
}

/* ── Draw stored strokes ── */
function redrawStrokes(previewReplacements: StrokeReplacement[] = []) {
  const dc = drawCanvas.value
  if (!dc) return
  const ctx = dc.getContext('2d')
  if (!ctx) return
  ctx.clearRect(0, 0, dc.width, dc.height)
  const replacementMap = new Map(previewReplacements.map((replacement) => [replacement.index, replacement.fragments]))

  for (let index = 0; index < props.strokes.length; index++) {
    const fragments = replacementMap.get(index)
    if (fragments) {
      for (const fragment of fragments) drawStroke(ctx, fragment)
    } else {
      drawStroke(ctx, props.strokes[index])
    }
  }
}

function drawStroke(
  ctx: CanvasRenderingContext2D,
  s: Stroke,
  options: { color?: string; alpha?: number; width?: number } = {},
) {
  if (s.points.length < 2) return
  ctx.save()
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.strokeStyle = options.color || s.color
  ctx.lineWidth = options.width || s.width

  if (s.tool === 'highlighter') {
    ctx.globalAlpha = 0.35
  }
  if (options.alpha !== undefined) {
    ctx.globalAlpha = options.alpha
  }

  ctx.beginPath()
  ctx.moveTo(s.points[0].x, s.points[0].y)
  for (let i = 1; i < s.points.length; i++) {
    ctx.lineTo(s.points[i].x, s.points[i].y)
  }
  ctx.stroke()
  ctx.restore()
}

/* ── Drawing interaction ── */
function pageToCanvas(e: MouseEvent | Touch): Point {
  const rect = drawCanvas.value!.getBoundingClientRect()
  return {
    x: (e.clientX - rect.left) * (drawCanvas.value!.width / rect.width),
    y: (e.clientY - rect.top) * (drawCanvas.value!.height / rect.height),
  }
}

function eraserRadius() {
  return Math.max(4, props.eraserSize / 2)
}

function drawEraserCursor(pt: Point) {
  const ctx = drawCanvas.value?.getContext('2d')
  if (!ctx) return
  const r = eraserRadius()
  ctx.save()
  ctx.fillStyle = 'rgba(220, 38, 38, 0.08)'
  ctx.strokeStyle = '#DC2626'
  ctx.lineWidth = 1.5
  ctx.beginPath()
  ctx.arc(pt.x, pt.y, r, 0, Math.PI * 2)
  ctx.fill()
  ctx.stroke()
  ctx.restore()
}

function drawAreaSelection(start: Point, end: Point) {
  const ctx = drawCanvas.value?.getContext('2d')
  if (!ctx) return
  const x = Math.min(start.x, end.x)
  const y = Math.min(start.y, end.y)
  const w = Math.abs(start.x - end.x)
  const h = Math.abs(start.y - end.y)
  ctx.save()
  ctx.strokeStyle = '#DC2626'
  ctx.lineWidth = 1.5
  ctx.setLineDash([6, 3])
  ctx.strokeRect(x, y, w, h)
  ctx.setLineDash([])
  ctx.fillStyle = 'rgba(220,38,38,0.08)'
  ctx.fillRect(x, y, w, h)
  ctx.restore()
}

function pointsEqual(a: Point[], b: Point[]): boolean {
  return a.length === b.length && a.every((point, index) =>
    point.x === b[index].x && point.y === b[index].y)
}

function collectReplacements(split: (stroke: Stroke) => Stroke[]): StrokeReplacement[] {
  return props.strokes.flatMap((stroke, index) => {
    const fragments = split(stroke)
    const unchanged = fragments.length === 1 && pointsEqual(fragments[0].points, stroke.points)
    return unchanged ? [] : [{ index, fragments }]
  })
}

function findFreehandReplacements(path: Point[]): StrokeReplacement[] {
  const radius = eraserRadius()
  return collectReplacements((stroke) => splitStrokeByEraserPath(stroke, path, radius))
}

function findAreaReplacements(start: Point, end: Point): StrokeReplacement[] {
  return collectReplacements((stroke) => splitStrokeByRect(stroke, start, end))
}

function redrawEraserPreview() {
  redrawStrokes(pendingReplacements)
  if (props.eraseMode === 'area') {
    if (areaStart && areaEnd) drawAreaSelection(areaStart, areaEnd)
  } else if (eraserPath.length > 0) {
    drawEraserCursor(eraserPath[eraserPath.length - 1])
  }
}

function onPointerDown(e: PointerEvent) {
  if (props.activeTool === 'none') return
  if (awaitingReplacementCommit) return
  if (!drawCanvas.value) return
  drawCanvas.value.setPointerCapture(e.pointerId)
  isDrawing = true
  const pt = pageToCanvas(e)

  if (props.activeTool === 'eraser') {
    awaitingReplacementCommit = false
    pendingReplacements = []
    if (props.eraseMode === 'area') {
      areaStart = pt
      areaEnd = pt
      pendingReplacements = findAreaReplacements(pt, pt)
    } else {
      eraserPath = [pt]
      pendingReplacements = findFreehandReplacements(eraserPath)
    }
    redrawEraserPreview()
  } else {
    currentStroke = {
      tool: props.activeTool,
      color: props.penColor,
      width: props.penWidth,
      points: [pt],
    }
  }
}

function onPointerMove(e: PointerEvent) {
  if (!isDrawing || !drawCanvas.value) return
  const pt = pageToCanvas(e)

  if (props.activeTool === 'eraser') {
    if (props.eraseMode === 'area') {
      areaEnd = pt
      if (areaStart) {
        pendingReplacements = findAreaReplacements(areaStart, pt)
        redrawEraserPreview()
      }
      return
    }
    eraserPath.push(pt)
    pendingReplacements = findFreehandReplacements(eraserPath)
    redrawEraserPreview()
    return
  }

  if (!currentStroke) return
  currentStroke.points.push(pt)

  // Draw preview
  const ctx = drawCanvas.value.getContext('2d')
  if (!ctx) return
  ctx.save()
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.strokeStyle = currentStroke.color
  ctx.lineWidth = currentStroke.width
  if (currentStroke.tool === 'highlighter') ctx.globalAlpha = 0.35
  const pts = currentStroke.points
  ctx.beginPath()
  ctx.moveTo(pts[pts.length - 2].x, pts[pts.length - 2].y)
  ctx.lineTo(pts[pts.length - 1].x, pts[pts.length - 1].y)
  ctx.stroke()
  ctx.restore()
}

function onPointerUp(_e: PointerEvent) {
  if (!isDrawing) return
  isDrawing = false

  if (props.activeTool === 'eraser') {
    if (props.eraseMode === 'area') {
      if (areaStart && areaEnd) {
        pendingReplacements = findAreaReplacements(areaStart, areaEnd)
      }
      areaStart = null
      areaEnd = null
    } else {
      pendingReplacements = findFreehandReplacements(eraserPath)
      eraserPath = []
    }

    if (pendingReplacements.length > 0) {
      awaitingReplacementCommit = true
      redrawStrokes(pendingReplacements)
      emit('strokesReplaced', pendingReplacements)
    } else {
      pendingReplacements = []
      redrawStrokes()
    }
    return
  }

  if (!currentStroke) return
  if (currentStroke.points.length > 1) {
    emit('strokeCreated', { ...currentStroke, points: [...currentStroke.points] })
  }
  currentStroke = null
}

// Redraw when strokes change (from parent)
watch(() => props.strokes, () => {
  if (awaitingReplacementCommit) {
    awaitingReplacementCommit = false
    pendingReplacements = []
  }
  redrawStrokes(pendingReplacements)
}, { deep: true })

watch(() => [props.activeTool, props.eraseMode], () => {
  if (isDrawing || awaitingReplacementCommit) return
  awaitingReplacementCommit = false
  pendingReplacements = []
  eraserPath = []
  areaStart = null
  areaEnd = null
  redrawStrokes()
})

// Re-render PDF when scale changes
watch(() => props.scale, () => {
  if (pdfPage) {
    // Parent will re-render; we just resize draw canvas
    const vp = pdfPage.getViewport({ scale: props.scale })
    const dc = drawCanvas.value
    if (dc) { dc.width = vp.width; dc.height = vp.height }
  }
})

// Clear draw layer for eraser "delete" effect that parent handles
function clearDrawLayer() {
  const dc = drawCanvas.value
  if (!dc) return
  const ctx = dc.getContext('2d')
  if (ctx) ctx.clearRect(0, 0, dc.width, dc.height)
}

defineExpose({ renderPDF, redrawStrokes, clearDrawLayer })
</script>

<template>
  <div class="pdf-page-wrapper">
    <canvas ref="pdfCanvas" class="pdf-page-layer" />
    <canvas
      ref="drawCanvas"
      class="pdf-draw-layer"
      @pointerdown="onPointerDown"
      @pointermove="onPointerMove"
      @pointerup="onPointerUp"
      @pointerleave="onPointerUp"
      @pointercancel="onPointerUp"
    />
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
</style>
