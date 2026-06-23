import {
  distance,
  distanceToSegment,
  getPathBounds,
  splitStrokeByEraserPath,
  splitStrokeByRect,
} from './pdfAnnotationGeometry.ts'
import type { Bounds, PointLike } from './pdfAnnotationGeometry.ts'
import type { FreehandStroke, Point, ShapeDrawing } from './pdfDrawingTypes.ts'

export type ShapeHandle =
  | 'start'
  | 'end'
  | 'nw'
  | 'n'
  | 'ne'
  | 'e'
  | 'se'
  | 's'
  | 'sw'
  | 'w'

export interface ShapeHandlePoint {
  id: ShapeHandle
  point: Point
}

function geometricBounds(shape: ShapeDrawing): Bounds {
  return {
    left: Math.min(shape.start.x, shape.end.x),
    top: Math.min(shape.start.y, shape.end.y),
    right: Math.max(shape.start.x, shape.end.x),
    bottom: Math.max(shape.start.y, shape.end.y),
  }
}

function arrowheadContours(shape: ShapeDrawing): Point[][] {
  const dx = shape.end.x - shape.start.x
  const dy = shape.end.y - shape.start.y
  const length = Math.hypot(dx, dy)
  if (length === 0) return []

  const arrowLength = Math.min(length * 0.45, Math.max(10, shape.width * 4))
  const angle = Math.atan2(dy, dx)
  const spread = Math.PI / 6
  const left = {
    x: shape.end.x - Math.cos(angle - spread) * arrowLength,
    y: shape.end.y - Math.sin(angle - spread) * arrowLength,
  }
  const right = {
    x: shape.end.x - Math.cos(angle + spread) * arrowLength,
    y: shape.end.y - Math.sin(angle + spread) * arrowLength,
  }
  return [
    [{ ...shape.end }, left],
    [{ ...shape.end }, right],
  ]
}

export function shapeToContours(shape: ShapeDrawing, ellipseSegments = 64): Point[][] {
  if (shape.shapeType === 'line') {
    return [[{ ...shape.start }, { ...shape.end }]]
  }
  if (shape.shapeType === 'arrow') {
    return [
      [{ ...shape.start }, { ...shape.end }],
      ...arrowheadContours(shape),
    ]
  }

  const bounds = geometricBounds(shape)
  if (shape.shapeType === 'rectangle') {
    const topLeft = { x: bounds.left, y: bounds.top }
    return [[
      topLeft,
      { x: bounds.right, y: bounds.top },
      { x: bounds.right, y: bounds.bottom },
      { x: bounds.left, y: bounds.bottom },
      { ...topLeft },
    ]]
  }

  const segments = Math.max(12, ellipseSegments)
  const centerX = (bounds.left + bounds.right) / 2
  const centerY = (bounds.top + bounds.bottom) / 2
  const radiusX = (bounds.right - bounds.left) / 2
  const radiusY = (bounds.bottom - bounds.top) / 2
  const contour: Point[] = []
  for (let index = 0; index < segments; index++) {
    const angle = (index / segments) * Math.PI * 2
    contour.push({
      x: centerX + Math.cos(angle) * radiusX,
      y: centerY + Math.sin(angle) * radiusY,
    })
  }
  contour.push({ ...contour[0] })
  return [contour]
}

function contourStroke(shape: ShapeDrawing, points: Point[]): FreehandStroke {
  return {
    tool: 'pen',
    color: shape.color,
    width: shape.width,
    points: points.map((point) => ({ ...point })),
  }
}

function pointsEqual(left: Point[], right: Point[]): boolean {
  return left.length === right.length && left.every((point, index) =>
    point.x === right[index].x && point.y === right[index].y)
}

function splitShapeContours(
  shape: ShapeDrawing,
  split: (stroke: FreehandStroke) => FreehandStroke[],
): FreehandStroke[] | null {
  let changed = false
  const fragments: FreehandStroke[] = []
  for (const contour of shapeToContours(shape)) {
    const result = split(contourStroke(shape, contour))
    if (!(result.length === 1 && pointsEqual(result[0].points, contour))) changed = true
    fragments.push(...result)
  }
  return changed ? fragments : null
}

export function splitShapeByEraserPath(
  shape: ShapeDrawing,
  path: PointLike[],
  radius: number,
): FreehandStroke[] | null {
  return splitShapeContours(
    shape,
    (stroke) => splitStrokeByEraserPath(stroke, path, radius),
  )
}

export function splitShapeByRect(
  shape: ShapeDrawing,
  start: PointLike,
  end: PointLike,
): FreehandStroke[] | null {
  return splitShapeContours(
    shape,
    (stroke) => splitStrokeByRect(stroke, start, end),
  )
}

export function getShapeBounds(shape: ShapeDrawing): Bounds {
  const points = shapeToContours(shape).flat()
  return getPathBounds(points, shape.width / 2) || geometricBounds(shape)
}

function contourHit(contour: PointLike[], point: PointLike, tolerance: number): boolean {
  if (contour.length === 1) return distance(contour[0], point) <= tolerance
  for (let index = 1; index < contour.length; index++) {
    if (distanceToSegment(point, contour[index - 1], contour[index]) <= tolerance) return true
  }
  return false
}

export function hitTestShape(shape: ShapeDrawing, point: PointLike, tolerance: number): boolean {
  const threshold = tolerance + shape.width / 2
  return shapeToContours(shape).some((contour) => contourHit(contour, point, threshold))
}

export function getShapeHandles(shape: ShapeDrawing): ShapeHandlePoint[] {
  if (shape.shapeType === 'line' || shape.shapeType === 'arrow') {
    return [
      { id: 'start', point: { ...shape.start } },
      { id: 'end', point: { ...shape.end } },
    ]
  }

  const bounds = geometricBounds(shape)
  const centerX = (bounds.left + bounds.right) / 2
  const centerY = (bounds.top + bounds.bottom) / 2
  return [
    { id: 'nw', point: { x: bounds.left, y: bounds.top } },
    { id: 'n', point: { x: centerX, y: bounds.top } },
    { id: 'ne', point: { x: bounds.right, y: bounds.top } },
    { id: 'e', point: { x: bounds.right, y: centerY } },
    { id: 'se', point: { x: bounds.right, y: bounds.bottom } },
    { id: 's', point: { x: centerX, y: bounds.bottom } },
    { id: 'sw', point: { x: bounds.left, y: bounds.bottom } },
    { id: 'w', point: { x: bounds.left, y: centerY } },
  ]
}

export function hitTestShapeHandle(
  shape: ShapeDrawing,
  point: PointLike,
  tolerance: number,
): ShapeHandle | null {
  let closest: ShapeHandlePoint | null = null
  let closestDistance = Infinity
  for (const handle of getShapeHandles(shape)) {
    const handleDistance = distance(handle.point, point)
    if (handleDistance <= tolerance && handleDistance < closestDistance) {
      closest = handle
      closestDistance = handleDistance
    }
  }
  return closest?.id || null
}

export function moveShape(shape: ShapeDrawing, dx: number, dy: number): ShapeDrawing {
  return {
    ...shape,
    start: { x: shape.start.x + dx, y: shape.start.y + dy },
    end: { x: shape.end.x + dx, y: shape.end.y + dy },
  }
}

export function resizeShape(
  shape: ShapeDrawing,
  handle: ShapeHandle,
  point: PointLike,
): ShapeDrawing {
  if (shape.shapeType === 'line' || shape.shapeType === 'arrow') {
    if (handle === 'start') return { ...shape, start: { ...point } }
    if (handle === 'end') return { ...shape, end: { ...point } }
    return { ...shape, start: { ...shape.start }, end: { ...shape.end } }
  }

  const bounds = geometricBounds(shape)
  let { left, top, right, bottom } = bounds
  if (handle.includes('w')) left = point.x
  if (handle.includes('e')) right = point.x
  if (handle.includes('n')) top = point.y
  if (handle.includes('s')) bottom = point.y

  return {
    ...shape,
    start: {
      x: Math.min(left, right),
      y: Math.min(top, bottom),
    },
    end: {
      x: Math.max(left, right),
      y: Math.max(top, bottom),
    },
  }
}
