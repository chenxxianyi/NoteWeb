import assert from 'node:assert/strict'
import test from 'node:test'

import {
  drawingsEqual,
  drawingReplacementChangesDrawing,
  sameDrawingIdentity,
} from './pdfDrawingTypes.ts'
import type { Drawing, FreehandStroke, ShapeDrawing, TextDrawing } from './pdfDrawingTypes.ts'

test('unchanged freehand clone is not treated as an eraser replacement', () => {
  const original: FreehandStroke = {
    id: 1,
    tool: 'pen',
    color: '#ff0000',
    width: 4,
    points: [{ x: 1, y: 1 }, { x: 20, y: 20 }],
  }
  const cloned: FreehandStroke = {
    ...original,
    points: original.points.map((point) => ({ ...point })),
  }

  assert.equal(drawingReplacementChangesDrawing(original, [cloned]), false)
})

test('changed freehand fragments are treated as an eraser replacement', () => {
  const original: FreehandStroke = {
    id: 1,
    tool: 'pen',
    color: '#ff0000',
    width: 4,
    points: [{ x: 1, y: 1 }, { x: 20, y: 20 }],
  }
  const fragment: FreehandStroke = {
    ...original,
    points: [{ x: 1, y: 1 }, { x: 8, y: 8 }],
  }

  assert.equal(drawingReplacementChangesDrawing(original, [fragment]), true)
})

test('shape converted to ordinary pen fragments is treated as a replacement', () => {
  const original: ShapeDrawing = {
    id: 2,
    tool: 'shape',
    shapeType: 'rectangle',
    color: '#ff0000',
    width: 4,
    start: { x: 0, y: 0 },
    end: { x: 100, y: 50 },
  }
  const fragment: Drawing = {
    tool: 'pen',
    color: '#ff0000',
    width: 4,
    points: [{ x: 0, y: 0 }, { x: 40, y: 0 }],
  }

  assert.equal(drawingReplacementChangesDrawing(original, [fragment]), true)
})

test('drawing identity matches by object reference or persisted id', () => {
  const original: FreehandStroke = {
    id: 3,
    tool: 'pen',
    color: '#ff0000',
    width: 4,
    points: [{ x: 0, y: 0 }, { x: 10, y: 10 }],
  }
  const sameId: FreehandStroke = {
    ...original,
    points: original.points.map((point) => ({ ...point })),
  }

  assert.equal(sameDrawingIdentity(original, original), true)
  assert.equal(sameDrawingIdentity(sameId, original), true)
  assert.equal(sameDrawingIdentity(undefined, original), false)
})

test('text drawing equality includes content, position, width, and font size', () => {
  const original: TextDrawing = {
    id: 4,
    tool: 'text',
    color: '#ff0000',
    fontSize: 24,
    text: 'hello',
    x: 10,
    y: 20,
    width: 180,
    height: 48,
  }
  const same: TextDrawing = { ...original }
  const moved: TextDrawing = { ...original, x: 12 }
  const resized: TextDrawing = { ...original, width: 220 }
  const taller: TextDrawing = { ...original, height: 96 }

  assert.equal(drawingsEqual(original, same), true)
  assert.equal(drawingsEqual(original, moved), false)
  assert.equal(drawingsEqual(original, resized), false)
  assert.equal(drawingsEqual(original, taller), false)
  assert.equal(drawingReplacementChangesDrawing(original, [same]), false)
  assert.equal(drawingReplacementChangesDrawing(original, [resized]), true)
})
