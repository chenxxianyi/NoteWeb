import { boundsIntersect, distanceToSegment, getPathBounds } from './pdfAnnotationGeometry.ts'
import type { Bounds, PointLike } from './pdfAnnotationGeometry.ts'
import type { Point, TextDrawing, TextErasure } from './pdfDrawingTypes.ts'

export const TEXT_PADDING = 4
export const TEXT_LINE_HEIGHT = 1.25
export const MIN_TEXT_WIDTH = 48

type MeasureText = (text: string, fontSize: number) => number

function estimatedTextWidth(text: string, fontSize: number): number {
  let width = 0
  for (const char of text) {
    if (char === ' ') width += fontSize * 0.33
    else if (/[\u3400-\u9fff\uff00-\uffef]/.test(char)) width += fontSize
    else width += fontSize * 0.56
  }
  return width
}

export function wrapTextLines(
  text: string,
  fontSize: number,
  width: number,
  measure: MeasureText = estimatedTextWidth,
): string[] {
  const maxLineWidth = Math.max(fontSize, width - TEXT_PADDING * 2)
  const lines: string[] = []
  let current = ''

  const pushCurrent = () => {
    lines.push(current)
    current = ''
  }

  for (const char of text || '') {
    if (char === '\r') continue
    if (char === '\n') {
      pushCurrent()
      continue
    }

    const next = current + char
    if (current && measure(next, fontSize) > maxLineWidth) {
      pushCurrent()
      current = char
    } else {
      current = next
    }
  }
  lines.push(current)
  return lines.length > 0 ? lines : ['']
}

export function getTextContentHeight(text: TextDrawing): number {
  const lineHeight = text.fontSize * TEXT_LINE_HEIGHT
  return wrapTextLines(text.text, text.fontSize, text.width).length * lineHeight +
    TEXT_PADDING * 2
}

export function getTextHeight(text: TextDrawing): number {
  return Math.max(text.height, getTextContentHeight(text))
}

export function getTextBounds(text: TextDrawing): Bounds {
  const width = Math.max(MIN_TEXT_WIDTH, text.width)
  return {
    left: text.x,
    top: text.y,
    right: text.x + width,
    bottom: text.y + getTextHeight({ ...text, width }),
  }
}

export function hitTestText(
  text: TextDrawing,
  point: PointLike,
  tolerance = 0,
): boolean {
  const bounds = getTextBounds(text)
  return point.x >= bounds.left - tolerance &&
    point.x <= bounds.right + tolerance &&
    point.y >= bounds.top - tolerance &&
    point.y <= bounds.bottom + tolerance
}

function pointDistanceToBounds(point: PointLike, bounds: Bounds): number {
  const dx = Math.max(bounds.left - point.x, 0, point.x - bounds.right)
  const dy = Math.max(bounds.top - point.y, 0, point.y - bounds.bottom)
  return Math.hypot(dx, dy)
}

function cross(a: PointLike, b: PointLike, c: PointLike): number {
  return (b.x - a.x) * (c.y - a.y) - (b.y - a.y) * (c.x - a.x)
}

function pointOnSegment(point: PointLike, a: PointLike, b: PointLike): boolean {
  const epsilon = 1e-9
  return Math.abs(cross(a, b, point)) <= epsilon &&
    point.x >= Math.min(a.x, b.x) - epsilon &&
    point.x <= Math.max(a.x, b.x) + epsilon &&
    point.y >= Math.min(a.y, b.y) - epsilon &&
    point.y <= Math.max(a.y, b.y) + epsilon
}

function segmentsIntersect(a: PointLike, b: PointLike, c: PointLike, d: PointLike): boolean {
  const abC = cross(a, b, c)
  const abD = cross(a, b, d)
  const cdA = cross(c, d, a)
  const cdB = cross(c, d, b)

  if (
    ((abC > 0 && abD < 0) || (abC < 0 && abD > 0)) &&
    ((cdA > 0 && cdB < 0) || (cdA < 0 && cdB > 0))
  ) {
    return true
  }

  return pointOnSegment(c, a, b) ||
    pointOnSegment(d, a, b) ||
    pointOnSegment(a, c, d) ||
    pointOnSegment(b, c, d)
}

function boundsEdges(bounds: Bounds): Array<[PointLike, PointLike]> {
  const topLeft = { x: bounds.left, y: bounds.top }
  const topRight = { x: bounds.right, y: bounds.top }
  const bottomRight = { x: bounds.right, y: bounds.bottom }
  const bottomLeft = { x: bounds.left, y: bounds.bottom }
  return [
    [topLeft, topRight],
    [topRight, bottomRight],
    [bottomRight, bottomLeft],
    [bottomLeft, topLeft],
  ]
}

function segmentDistanceToBounds(a: PointLike, b: PointLike, bounds: Bounds): number {
  if (pointDistanceToBounds(a, bounds) === 0 || pointDistanceToBounds(b, bounds) === 0) return 0

  const edges = boundsEdges(bounds)
  if (edges.some(([start, end]) => segmentsIntersect(a, b, start, end))) return 0

  const corners = [
    { x: bounds.left, y: bounds.top },
    { x: bounds.right, y: bounds.top },
    { x: bounds.right, y: bounds.bottom },
    { x: bounds.left, y: bounds.bottom },
  ]
  return Math.min(
    pointDistanceToBounds(a, bounds),
    pointDistanceToBounds(b, bounds),
    ...corners.map((corner) => distanceToSegment(corner, a, b)),
  )
}

function boundsIntersectsEraserPath(bounds: Bounds, path: PointLike[], radius: number): boolean {
  const effectiveRadius = Math.max(0, radius)
  if (path.length === 0) return false
  if (path.length === 1) return pointDistanceToBounds(path[0], bounds) <= effectiveRadius

  for (let index = 1; index < path.length; index++) {
    if (segmentDistanceToBounds(path[index - 1], path[index], bounds) <= effectiveRadius) {
      return true
    }
  }
  return false
}

function boundsIntersectsRect(bounds: Bounds, start: PointLike, end: PointLike): boolean {
  const rect = getPathBounds([start, end])
  return rect ? boundsIntersect(bounds, rect) : false
}

export function textIntersectsEraserPath(
  text: TextDrawing,
  path: PointLike[],
  radius: number,
): boolean {
  return boundsIntersectsEraserPath(getTextBounds(text), path, radius)
}

export function textIntersectsRect(
  text: TextDrawing,
  start: PointLike,
  end: PointLike,
): boolean {
  return boundsIntersectsRect(getTextBounds(text), start, end)
}

function clonePoint(point: PointLike): Point {
  return { x: point.x, y: point.y }
}

function cloneTextErasure(erasure: TextErasure): TextErasure {
  if (erasure.type === 'path') {
    return {
      type: 'path',
      radius: erasure.radius,
      points: erasure.points.map(clonePoint),
    }
  }
  return {
    type: 'rect',
    start: clonePoint(erasure.start),
    end: clonePoint(erasure.end),
  }
}

function appendTextErasure(text: TextDrawing, erasure: TextErasure): TextDrawing[] {
  const next: TextDrawing = {
    ...text,
    erasures: [
      ...(text.erasures || []).map(cloneTextErasure),
      cloneTextErasure(erasure),
    ],
  }
  delete next.id
  return [next]
}

export function eraseTextByEraserPath(
  text: TextDrawing,
  path: PointLike[],
  radius: number,
): TextDrawing[] {
  if (!textIntersectsEraserPath(text, path, radius)) return [text]
  return appendTextErasure(text, {
    type: 'path',
    radius: Math.max(0, radius),
    points: path.map(clonePoint),
  })
}

export function eraseTextByRect(
  text: TextDrawing,
  start: PointLike,
  end: PointLike,
): TextDrawing[] {
  if (!textIntersectsRect(text, start, end)) return [text]
  return appendTextErasure(text, {
    type: 'rect',
    start: clonePoint(start),
    end: clonePoint(end),
  })
}

export function getTextResizeHandlePoint(text: TextDrawing): PointLike {
  const bounds = getTextBounds(text)
  return { x: bounds.right, y: bounds.bottom }
}

export function hitTestTextResizeHandle(
  text: TextDrawing,
  point: PointLike,
  tolerance: number,
): boolean {
  const handle = getTextResizeHandlePoint(text)
  return Math.abs(point.x - handle.x) <= tolerance &&
    Math.abs(point.y - handle.y) <= tolerance
}

export function moveText(text: TextDrawing, dx: number, dy: number): TextDrawing {
  return {
    ...text,
    x: text.x + dx,
    y: text.y + dy,
  }
}

export function resizeTextWidth(text: TextDrawing, point: PointLike): TextDrawing {
  return resizeTextBox(text, point)
}

export function resizeTextBox(text: TextDrawing, point: PointLike): TextDrawing {
  const nextWidth = Math.max(MIN_TEXT_WIDTH, point.x - text.x)
  const autoHeight = getTextContentHeight({ ...text, width: nextWidth })
  return {
    ...text,
    width: nextWidth,
    height: Math.max(autoHeight, point.y - text.y),
  }
}
