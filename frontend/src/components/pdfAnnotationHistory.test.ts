import assert from 'node:assert/strict'
import test from 'node:test'

interface TestStroke {
  id?: number
  label: string
}

type TestHistoryEntry =
  | { type: 'draw'; stroke: TestStroke }
  | {
      type: 'erase'
      replacements: Array<{
        original: TestStroke
        fragments: TestStroke[]
      }>
    }
  | {
      type: 'edit'
      original: TestStroke
      replacement: TestStroke
    }

let historyModule: Record<string, unknown> = {}
try {
  historyModule = await import('./pdfAnnotationHistory.ts')
} catch {
  // The RED phase expects the helper module to be absent.
}

const remapStrokeInHistory = historyModule.remapStrokeInHistory as
  | ((
      history: TestHistoryEntry[],
      oldStroke: TestStroke,
      restoredStroke: TestStroke,
      endExclusive: number,
    ) => void)
  | undefined

test('restoring a stroke remaps earlier draw and erase history references to its new id', () => {
  assert.equal(typeof remapStrokeInHistory, 'function')

  const oldStroke = { id: 10, label: 'old' }
  const restoredStroke = { id: 99, label: 'restored' }
  const sibling = { id: 11, label: 'sibling' }
  const history: TestHistoryEntry[] = [
    { type: 'draw', stroke: oldStroke },
    {
      type: 'erase',
      replacements: [{
        original: sibling,
        fragments: [oldStroke, sibling],
      }],
    },
    { type: 'draw', stroke: oldStroke },
  ]

  remapStrokeInHistory!(history, oldStroke, restoredStroke, 2)

  assert.equal(history[0].type, 'draw')
  assert.equal((history[0] as Extract<TestHistoryEntry, { type: 'draw' }>).stroke, restoredStroke)
  const eraseEntry = history[1] as Extract<TestHistoryEntry, { type: 'erase' }>
  assert.equal(eraseEntry.replacements[0].fragments[0], restoredStroke)
  assert.equal(eraseEntry.replacements[0].fragments[1], sibling)
  assert.equal((history[2] as Extract<TestHistoryEntry, { type: 'draw' }>).stroke, oldStroke)
})

test('restoring a drawing remaps earlier edit history references', () => {
  assert.equal(typeof remapStrokeInHistory, 'function')
  const oldShape = { id: 20, label: 'old shape' }
  const restoredShape = { id: 80, label: 'restored shape' }
  const previous = { id: 19, label: 'previous state' }
  const history: TestHistoryEntry[] = [
    { type: 'edit', original: previous, replacement: oldShape },
    { type: 'edit', original: oldShape, replacement: { id: 21, label: 'next state' } },
  ]

  remapStrokeInHistory!(history, oldShape, restoredShape, 2)

  const first = history[0] as Extract<TestHistoryEntry, { type: 'edit' }>
  const second = history[1] as Extract<TestHistoryEntry, { type: 'edit' }>
  assert.equal(first.replacement, restoredShape)
  assert.equal(second.original, restoredShape)
})
