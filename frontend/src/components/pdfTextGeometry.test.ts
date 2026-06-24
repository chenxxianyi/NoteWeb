import assert from 'node:assert/strict'
import test from 'node:test'

import {
  eraseTextByEraserPath,
  eraseTextByRect,
  getTextBounds,
  hitTestText,
  hitTestTextResizeHandle,
  moveText,
  resizeTextWidth,
  textIntersectsEraserPath,
  textIntersectsRect,
  wrapTextLines,
} from './pdfTextGeometry.ts'
import type { TextDrawing } from './pdfDrawingTypes.ts'

function textDrawing(overrides: Partial<TextDrawing> = {}): TextDrawing {
  return {
    id: 1,
    tool: 'text',
    color: '#ff0000',
    fontSize: 20,
    text: '太阳光电系统',
    x: 100,
    y: 80,
    width: 120,
    height: 40,
    ...overrides,
  }
}

test('text wrapping respects explicit new lines and available width', () => {
  const lines = wrapTextLines('abc\ndefgh', 10, 28, (value) => value.length * 10)

  assert.deepEqual(lines, ['ab', 'c', 'de', 'fg', 'h'])
})

test('text bounds include padding and wrapped line height', () => {
  const bounds = getTextBounds(textDrawing({
    text: 'abcdefghij',
    width: 48,
    height: 10,
    fontSize: 10,
  }))

  assert.equal(bounds.left, 100)
  assert.equal(bounds.top, 80)
  assert.equal(bounds.right, 148)
  assert.equal(bounds.bottom, 113)
})

test('text hit testing checks the editable text box', () => {
  const drawing = textDrawing()

  assert.equal(hitTestText(drawing, { x: 110, y: 90 }), true)
  assert.equal(hitTestText(drawing, { x: 10, y: 10 }), false)
})

test('freehand eraser path detects text box contact', () => {
  const drawing = textDrawing({ text: 'hello', x: 100, y: 80, width: 120, height: 40 })

  assert.equal(textIntersectsEraserPath(drawing, [{ x: 90, y: 90 }, { x: 105, y: 90 }], 6), true)
  assert.equal(textIntersectsEraserPath(drawing, [{ x: 20, y: 20 }, { x: 40, y: 20 }], 6), false)
})

test('area eraser detects text box overlap', () => {
  const drawing = textDrawing({ text: 'hello', x: 100, y: 80, width: 120, height: 40 })

  assert.equal(textIntersectsRect(drawing, { x: 90, y: 70 }, { x: 130, y: 95 }), true)
  assert.equal(textIntersectsRect(drawing, { x: 10, y: 10 }, { x: 30, y: 30 }), false)
})

test('freehand eraser records a text-local erasure mask', () => {
  const drawing = textDrawing({ text: 'hello', x: 100, y: 80, width: 200, height: 40, fontSize: 10 })
  const fragments = eraseTextByEraserPath(
    drawing,
    [{ x: 119, y: 90 }],
    2,
  )

  assert.equal(fragments.length, 1)
  assert.equal(fragments[0].text, 'hello')
  assert.deepEqual(fragments[0].erasures, [{
    type: 'path',
    radius: 2,
    points: [{ x: 119, y: 90 }],
  }])
})

test('area eraser records a text-local rectangle erasure', () => {
  const drawing = textDrawing({ text: 'hello', x: 100, y: 80, width: 200, height: 40, fontSize: 10 })
  const fragments = eraseTextByRect(
    drawing,
    { x: 115, y: 83 },
    { x: 125, y: 97 },
  )

  assert.equal(fragments.length, 1)
  assert.equal(fragments[0].text, 'hello')
  assert.deepEqual(fragments[0].erasures, [{
    type: 'rect',
    start: { x: 115, y: 83 },
    end: { x: 125, y: 97 },
  }])
})

test('text resize handle is detected near the lower right corner', () => {
  const drawing = textDrawing({ text: 'abc', width: 140 })
  const bounds = getTextBounds(drawing)

  assert.equal(hitTestTextResizeHandle(drawing, { x: bounds.right + 2, y: bounds.bottom - 2 }, 5), true)
  assert.equal(hitTestTextResizeHandle(drawing, { x: drawing.x, y: drawing.y }, 5), false)
})

test('moving and resizing text returns updated copies without mutation', () => {
  const drawing = textDrawing()
  const moved = moveText(drawing, 10, -5)
  const resized = resizeTextWidth(drawing, { x: 180, y: 90 })

  assert.deepEqual({ x: moved.x, y: moved.y }, { x: 110, y: 75 })
  assert.equal(resized.width, 80)
  assert.deepEqual(
    { x: drawing.x, y: drawing.y, width: drawing.width, height: drawing.height },
    { x: 100, y: 80, width: 120, height: 40 },
  )
})
