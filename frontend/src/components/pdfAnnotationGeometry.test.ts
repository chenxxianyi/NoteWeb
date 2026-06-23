import assert from 'node:assert/strict'
import test from 'node:test'
import * as geometry from './pdfAnnotationGeometry.ts'

interface TestStroke {
  tool: 'pen' | 'highlighter'
  color: string
  width: number
  points: Array<{ x: number; y: number }>
}

type SplitPath = (
  stroke: TestStroke,
  eraserPath: Array<{ x: number; y: number }>,
  radius: number,
) => TestStroke[]

type SplitRect = (
  stroke: TestStroke,
  start: { x: number; y: number },
  end: { x: number; y: number },
) => TestStroke[]

interface Bounds {
  left: number
  top: number
  right: number
  bottom: number
}

type GetStrokeBounds = (stroke: TestStroke) => Bounds
type GetPathBounds = (
  points: Array<{ x: number; y: number }>,
  padding?: number,
) => Bounds | null
type BoundsIntersect = (a: Bounds, b: Bounds) => boolean
type SimplifyPath = (
  points: Array<{ x: number; y: number }>,
  minDistance: number,
) => Array<{ x: number; y: number }>

const splitStrokeByEraserPath = (geometry as unknown as Record<string, unknown>)
  .splitStrokeByEraserPath as SplitPath
const splitStrokeByRect = (geometry as unknown as Record<string, unknown>)
  .splitStrokeByRect as SplitRect
const getStrokeBounds = (geometry as unknown as Record<string, unknown>)
  .getStrokeBounds as GetStrokeBounds
const getPathBounds = (geometry as unknown as Record<string, unknown>)
  .getPathBounds as GetPathBounds
const boundsIntersect = (geometry as unknown as Record<string, unknown>)
  .boundsIntersect as BoundsIntersect
const simplifyPathByDistance = (geometry as unknown as Record<string, unknown>)
  .simplifyPathByDistance as SimplifyPath

function lineStroke(
  startX = 0,
  endX = 100,
  overrides: Partial<TestStroke> = {},
): TestStroke {
  return {
    tool: 'pen',
    color: '#ff0000',
    width: 2,
    points: [{ x: startX, y: 0 }, { x: endX, y: 0 }],
    ...overrides,
  }
}

test('freehand eraser splits a stroke around the covered path', () => {
  assert.equal(typeof splitStrokeByEraserPath, 'function')

  const fragments = splitStrokeByEraserPath(lineStroke(), [{ x: 50, y: 0 }], 10)

  assert.equal(fragments.length, 2)
  assert.ok(fragments[0].points.at(-1)!.x < 41)
  assert.ok(fragments[1].points[0].x > 59)
})

test('rectangle eraser removes only the section inside the rectangle', () => {
  assert.equal(typeof splitStrokeByRect, 'function')

  const fragments = splitStrokeByRect(lineStroke(), { x: 40, y: -10 }, { x: 60, y: 10 })

  assert.equal(fragments.length, 2)
  assert.ok(fragments[0].points.at(-1)!.x < 40)
  assert.ok(fragments[1].points[0].x > 60)
})

test('an untouched stroke is preserved without extra sampling points', () => {
  assert.equal(typeof splitStrokeByEraserPath, 'function')
  const stroke = lineStroke()

  const fragments = splitStrokeByEraserPath(stroke, [{ x: 50, y: 40 }], 5)

  assert.equal(fragments.length, 1)
  assert.deepEqual(fragments[0], stroke)
  assert.notEqual(fragments[0].points, stroke.points)
})

test('a fully covered stroke produces no fragments', () => {
  assert.equal(typeof splitStrokeByEraserPath, 'function')

  const fragments = splitStrokeByEraserPath(lineStroke(0, 20), [{ x: 10, y: 0 }], 30)

  assert.deepEqual(fragments, [])
})

test('sparse long segments gain sampled boundary points when split', () => {
  assert.equal(typeof splitStrokeByEraserPath, 'function')

  const fragments = splitStrokeByEraserPath(lineStroke(), [{ x: 50, y: 0 }], 5)

  assert.equal(fragments.length, 2)
  assert.ok(fragments[0].points.length > 2)
  assert.ok(fragments[1].points.length > 2)
  assert.ok(fragments[0].points.at(-1)!.x > 40)
  assert.ok(fragments[1].points[0].x < 60)
})

test('stroke width expands the effective freehand erased region', () => {
  assert.equal(typeof splitStrokeByEraserPath, 'function')
  const eraserPath = [{ x: 50, y: 5 }]

  const thin = splitStrokeByEraserPath(lineStroke(0, 100, { width: 2 }), eraserPath, 1)
  const thick = splitStrokeByEraserPath(
    lineStroke(0, 100, { tool: 'highlighter', width: 10 }),
    eraserPath,
    1,
  )

  assert.equal(thin.length, 1)
  assert.equal(thick.length, 2)
  assert.ok(thick.every((fragment) => fragment.tool === 'highlighter'))
})

test('rectangle eraser preserves highlighter properties on both surviving fragments', () => {
  assert.equal(typeof splitStrokeByRect, 'function')
  const stroke = lineStroke(0, 100, {
    tool: 'highlighter',
    color: '#facc15',
    width: 12,
  })

  const fragments = splitStrokeByRect(stroke, { x: 40, y: -2 }, { x: 60, y: 2 })

  assert.equal(fragments.length, 2)
  assert.ok(fragments.every((fragment) => fragment.tool === 'highlighter'))
  assert.ok(fragments.every((fragment) => fragment.color === '#facc15'))
  assert.ok(fragments.every((fragment) => fragment.width === 12))
})

test('tiny fragments at the ends are discarded', () => {
  assert.equal(typeof splitStrokeByEraserPath, 'function')

  const fragments = splitStrokeByEraserPath(lineStroke(0, 20), [{ x: 10, y: 0 }], 8)

  assert.deepEqual(fragments, [])
})

test('stroke bounds include half the stroke width', () => {
  assert.equal(typeof getStrokeBounds, 'function')

  const bounds = getStrokeBounds(lineStroke(0, 100, { width: 10 }))

  assert.deepEqual(bounds, { left: -5, top: -5, right: 105, bottom: 5 })
})

test('path bounds include requested eraser padding', () => {
  assert.equal(typeof getPathBounds, 'function')

  const bounds = getPathBounds([{ x: 10, y: 20 }, { x: 30, y: 5 }], 8)

  assert.deepEqual(bounds, { left: 2, top: -3, right: 38, bottom: 28 })
})

test('bounds intersection rejects distant strokes and keeps touching strokes', () => {
  assert.equal(typeof boundsIntersect, 'function')
  const gesture = { left: 10, top: 10, right: 20, bottom: 20 }

  assert.equal(boundsIntersect(gesture, { left: 20, top: 12, right: 30, bottom: 18 }), true)
  assert.equal(boundsIntersect(gesture, { left: 21, top: 12, right: 30, bottom: 18 }), false)
})

test('distance simplification reduces dense paths while preserving first and last points', () => {
  assert.equal(typeof simplifyPathByDistance, 'function')
  const points = Array.from({ length: 101 }, (_, x) => ({ x, y: 0 }))

  const simplified = simplifyPathByDistance(points, 10)

  assert.deepEqual(simplified[0], points[0])
  assert.deepEqual(simplified.at(-1), points.at(-1))
  assert.ok(simplified.length <= 12)
})
