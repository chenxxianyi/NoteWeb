import assert from 'node:assert/strict'
import test from 'node:test'

interface Point {
  x: number
  y: number
}

interface Shape {
  tool: 'shape'
  shapeType: 'line' | 'arrow' | 'rectangle' | 'ellipse'
  color: string
  width: number
  start: Point
  end: Point
}

interface Stroke {
  tool: 'pen' | 'highlighter'
  color: string
  width: number
  points: Point[]
}

let shapeModule: Record<string, unknown> = {}
try {
  shapeModule = await import('./pdfShapeGeometry.ts')
} catch {
  // RED expects the module to be absent.
}

const shapeToContours = shapeModule.shapeToContours as
  | ((shape: Shape, ellipseSegments?: number) => Point[][])
  | undefined
const getShapeBounds = shapeModule.getShapeBounds as
  | ((shape: Shape) => { left: number; top: number; right: number; bottom: number })
  | undefined
const hitTestShape = shapeModule.hitTestShape as
  | ((shape: Shape, point: Point, tolerance: number) => boolean)
  | undefined
const getShapeHandles = shapeModule.getShapeHandles as
  | ((shape: Shape) => Array<{ id: string; point: Point }>)
  | undefined
const hitTestShapeHandle = shapeModule.hitTestShapeHandle as
  | ((shape: Shape, point: Point, tolerance: number) => string | null)
  | undefined
const moveShape = shapeModule.moveShape as
  | ((shape: Shape, dx: number, dy: number) => Shape)
  | undefined
const resizeShape = shapeModule.resizeShape as
  | ((shape: Shape, handle: string, point: Point) => Shape)
  | undefined
const splitShapeByEraserPath = shapeModule.splitShapeByEraserPath as
  | ((shape: Shape, path: Point[], radius: number) => Stroke[] | null)
  | undefined
const splitShapeByRect = shapeModule.splitShapeByRect as
  | ((shape: Shape, start: Point, end: Point) => Stroke[] | null)
  | undefined

function shape(shapeType: Shape['shapeType'], start = { x: 10, y: 20 }, end = { x: 110, y: 80 }): Shape {
  return {
    tool: 'shape',
    shapeType,
    color: '#ff0000',
    width: 4,
    start,
    end,
  }
}

test('line has one contour from start to end', () => {
  assert.equal(typeof shapeToContours, 'function')

  assert.deepEqual(shapeToContours!(shape('line')), [[{ x: 10, y: 20 }, { x: 110, y: 80 }]])
})

test('arrow has a shaft and two arrowhead contours ending at the tip', () => {
  assert.equal(typeof shapeToContours, 'function')

  const contours = shapeToContours!(shape('arrow'))

  assert.equal(contours.length, 3)
  assert.deepEqual(contours[0], [{ x: 10, y: 20 }, { x: 110, y: 80 }])
  assert.deepEqual(contours[1][0], { x: 110, y: 80 })
  assert.deepEqual(contours[2][0], { x: 110, y: 80 })
})

test('rectangle is one closed contour', () => {
  assert.equal(typeof shapeToContours, 'function')

  const contour = shapeToContours!(shape('rectangle'))[0]

  assert.equal(contour.length, 5)
  assert.deepEqual(contour[0], contour.at(-1))
  assert.deepEqual(contour.slice(0, 4), [
    { x: 10, y: 20 },
    { x: 110, y: 20 },
    { x: 110, y: 80 },
    { x: 10, y: 80 },
  ])
})

test('ellipse is sampled as a closed contour inside its bounds', () => {
  assert.equal(typeof shapeToContours, 'function')

  const contour = shapeToContours!(shape('ellipse'), 16)[0]

  assert.equal(contour.length, 17)
  assert.deepEqual(contour[0], contour.at(-1))
  assert.deepEqual(contour[0], { x: 110, y: 50 })
})

test('shape bounds include half the outline width', () => {
  assert.equal(typeof getShapeBounds, 'function')

  assert.deepEqual(getShapeBounds!(shape('rectangle')), {
    left: 8,
    top: 18,
    right: 112,
    bottom: 82,
  })
})

test('hit testing detects an outline and rejects the empty center', () => {
  assert.equal(typeof hitTestShape, 'function')
  const rectangle = shape('rectangle')

  assert.equal(hitTestShape!(rectangle, { x: 60, y: 20 }, 3), true)
  assert.equal(hitTestShape!(rectangle, { x: 60, y: 50 }, 3), false)
})

test('line and arrow expose two endpoint handles', () => {
  assert.equal(typeof getShapeHandles, 'function')

  assert.deepEqual(getShapeHandles!(shape('line')).map((handle) => handle.id), ['start', 'end'])
  assert.deepEqual(getShapeHandles!(shape('arrow')).map((handle) => handle.id), ['start', 'end'])
})

test('rectangle and ellipse expose eight resize handles', () => {
  assert.equal(typeof getShapeHandles, 'function')

  assert.deepEqual(getShapeHandles!(shape('ellipse')).map((handle) => handle.id), [
    'nw', 'n', 'ne', 'e', 'se', 's', 'sw', 'w',
  ])
})

test('handle hit testing returns the nearest handle', () => {
  assert.equal(typeof hitTestShapeHandle, 'function')

  assert.equal(hitTestShapeHandle!(shape('rectangle'), { x: 11, y: 21 }, 5), 'nw')
  assert.equal(hitTestShapeHandle!(shape('rectangle'), { x: 60, y: 50 }, 5), null)
})

test('moving a shape translates both endpoints without mutating the original', () => {
  assert.equal(typeof moveShape, 'function')
  const original = shape('arrow')

  const moved = moveShape!(original, 5, -10)

  assert.deepEqual(moved.start, { x: 15, y: 10 })
  assert.deepEqual(moved.end, { x: 115, y: 70 })
  assert.deepEqual(original.start, { x: 10, y: 20 })
})

test('resizing a line updates only the selected endpoint', () => {
  assert.equal(typeof resizeShape, 'function')

  const resized = resizeShape!(shape('line'), 'end', { x: 150, y: 90 })

  assert.deepEqual(resized.start, { x: 10, y: 20 })
  assert.deepEqual(resized.end, { x: 150, y: 90 })
})

test('resizing a rectangle edge preserves the opposite sides', () => {
  assert.equal(typeof resizeShape, 'function')

  const resized = resizeShape!(shape('rectangle'), 'w', { x: 0, y: 50 })

  assert.deepEqual(resized.start, { x: 0, y: 20 })
  assert.deepEqual(resized.end, { x: 110, y: 80 })
})

test('freehand erasing a shape converts all surviving contours to ordinary pen strokes', () => {
  assert.equal(typeof splitShapeByEraserPath, 'function')
  const arrow = shape('arrow')

  const fragments = splitShapeByEraserPath!(
    arrow,
    [{ x: 60, y: 35 }, { x: 60, y: 65 }],
    4,
  )

  assert.ok(fragments)
  assert.ok(fragments.length >= 4)
  assert.ok(fragments.every((fragment) => fragment.tool === 'pen'))
  assert.ok(fragments.every((fragment) => fragment.color === arrow.color))
  assert.ok(fragments.every((fragment) => fragment.width === arrow.width))
})

test('an eraser that misses a shape leaves it editable', () => {
  assert.equal(typeof splitShapeByEraserPath, 'function')

  const fragments = splitShapeByEraserPath!(
    shape('rectangle'),
    [{ x: 200, y: 200 }, { x: 220, y: 220 }],
    3,
  )

  assert.equal(fragments, null)
})

test('area erasing a shape removes only the outline section inside the box', () => {
  assert.equal(typeof splitShapeByRect, 'function')

  const fragments = splitShapeByRect!(
    shape('rectangle'),
    { x: 50, y: 15 },
    { x: 70, y: 25 },
  )

  assert.ok(fragments)
  assert.ok(fragments.every((fragment) => fragment.tool === 'pen'))
  assert.ok(fragments.some((fragment) =>
    fragment.points.some((point) => point.x < 50)))
  assert.ok(fragments.some((fragment) =>
    fragment.points.some((point) => point.x > 70)))
})
