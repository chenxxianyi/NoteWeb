export interface PointLike {
  x: number
  y: number
}

export interface StrokeLike {
  points: PointLike[]
}

export interface WidthStrokeLike extends StrokeLike {
  width: number
}

export interface SplitStrokeOptions {
  sampleStep?: number
  minFragmentLength?: number
}

export interface Bounds {
  left: number
  top: number
  right: number
  bottom: number
}

export function distance(a: PointLike, b: PointLike): number {
  return Math.hypot(a.x - b.x, a.y - b.y)
}

export function getPathBounds(points: PointLike[], padding = 0): Bounds | null {
  if (points.length === 0) return null

  let left = points[0].x
  let right = points[0].x
  let top = points[0].y
  let bottom = points[0].y
  for (let index = 1; index < points.length; index++) {
    const point = points[index]
    left = Math.min(left, point.x)
    right = Math.max(right, point.x)
    top = Math.min(top, point.y)
    bottom = Math.max(bottom, point.y)
  }

  return {
    left: left - padding,
    top: top - padding,
    right: right + padding,
    bottom: bottom + padding,
  }
}

export function getStrokeBounds(stroke: WidthStrokeLike): Bounds | null {
  return getPathBounds(stroke.points, stroke.width / 2)
}

export function boundsIntersect(a: Bounds, b: Bounds): boolean {
  return a.left <= b.right &&
    a.right >= b.left &&
    a.top <= b.bottom &&
    a.bottom >= b.top
}

export function simplifyPathByDistance(points: PointLike[], minDistance: number): PointLike[] {
  if (points.length <= 2 || minDistance <= 0) {
    return points.map((point) => ({ ...point }))
  }

  const minDistanceSquared = minDistance * minDistance
  const simplified: PointLike[] = [{ ...points[0] }]
  let lastKept = points[0]

  for (let index = 1; index < points.length - 1; index++) {
    const point = points[index]
    const dx = point.x - lastKept.x
    const dy = point.y - lastKept.y
    if (dx * dx + dy * dy >= minDistanceSquared) {
      simplified.push({ ...point })
      lastKept = point
    }
  }

  const last = points[points.length - 1]
  const tail = simplified[simplified.length - 1]
  if (tail.x !== last.x || tail.y !== last.y) simplified.push({ ...last })
  return simplified
}

export function distanceToSegment(point: PointLike, a: PointLike, b: PointLike): number {
  const dx = b.x - a.x
  const dy = b.y - a.y
  const lenSq = dx * dx + dy * dy

  if (lenSq === 0) return distance(point, a)

  const t = Math.max(0, Math.min(1, ((point.x - a.x) * dx + (point.y - a.y) * dy) / lenSq))
  return distance(point, { x: a.x + t * dx, y: a.y + t * dy })
}

function distanceToPath(point: PointLike, path: PointLike[]): number {
  if (path.length === 0) return Infinity
  if (path.length === 1) return distance(point, path[0])

  let closest = Infinity
  for (let i = 1; i < path.length; i++) {
    closest = Math.min(closest, distanceToSegment(point, path[i - 1], path[i]))
  }
  return closest
}

function samplePolyline(points: PointLike[], maxStep: number): PointLike[] {
  if (points.length < 2) return points.map((point) => ({ ...point }))

  const sampled: PointLike[] = [{ ...points[0] }]
  for (let i = 1; i < points.length; i++) {
    const start = points[i - 1]
    const end = points[i]
    const steps = Math.max(1, Math.ceil(distance(start, end) / maxStep))
    for (let step = 1; step <= steps; step++) {
      const ratio = step / steps
      sampled.push({
        x: start.x + (end.x - start.x) * ratio,
        y: start.y + (end.y - start.y) * ratio,
      })
    }
  }
  return sampled
}

function polylineLength(points: PointLike[]): number {
  let length = 0
  for (let i = 1; i < points.length; i++) {
    length += distance(points[i - 1], points[i])
  }
  return length
}

function cloneWithPoints<T extends StrokeLike>(stroke: T, points: PointLike[]): T {
  return { ...stroke, points: points.map((point) => ({ ...point })) }
}

function splitStroke<T extends WidthStrokeLike>(
  stroke: T,
  isErased: (point: PointLike) => boolean,
  options: SplitStrokeOptions,
): T[] {
  if (stroke.points.length === 0) return []

  const sampleStep = Math.max(0.5, options.sampleStep ?? Math.min(4, Math.max(1, stroke.width / 2)))
  const sampled = samplePolyline(stroke.points, sampleStep)
  const erased = sampled.map(isErased)

  if (!erased.some(Boolean)) {
    return [cloneWithPoints(stroke, stroke.points)]
  }

  const minFragmentLength = options.minFragmentLength ?? Math.max(2, stroke.width * 0.75)
  const fragments: T[] = []
  let current: PointLike[] = []

  const finishCurrent = () => {
    if (current.length >= 2 && polylineLength(current) >= minFragmentLength) {
      fragments.push(cloneWithPoints(stroke, current))
    }
    current = []
  }

  for (let i = 0; i < sampled.length; i++) {
    if (erased[i]) {
      finishCurrent()
    } else {
      current.push(sampled[i])
    }
  }
  finishCurrent()

  return fragments
}

export function splitStrokeByEraserPath<T extends WidthStrokeLike>(
  stroke: T,
  eraserPath: PointLike[],
  radius: number,
  options: SplitStrokeOptions = {},
): T[] {
  if (eraserPath.length === 0) return [cloneWithPoints(stroke, stroke.points)]

  const effectiveRadius = Math.max(0, radius) + stroke.width / 2
  const sampleStep = options.sampleStep ?? Math.min(4, Math.max(1, effectiveRadius / 2))
  return splitStroke(
    stroke,
    (point) => distanceToPath(point, eraserPath) <= effectiveRadius,
    { ...options, sampleStep },
  )
}

export function splitStrokeByRect<T extends WidthStrokeLike>(
  stroke: T,
  start: PointLike,
  end: PointLike,
  options: SplitStrokeOptions = {},
): T[] {
  const padding = stroke.width / 2
  const left = Math.min(start.x, end.x) - padding
  const right = Math.max(start.x, end.x) + padding
  const top = Math.min(start.y, end.y) - padding
  const bottom = Math.max(start.y, end.y) + padding

  return splitStroke(
    stroke,
    (point) => point.x >= left && point.x <= right && point.y >= top && point.y <= bottom,
    options,
  )
}
