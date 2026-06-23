export interface HistoryStrokeLike {
  id?: number
}

export type StrokeHistoryEntry<T extends HistoryStrokeLike> =
  | { type: 'draw'; stroke: T }
  | {
      type: 'erase'
      replacements: Array<{
        original: T
        fragments: T[]
      }>
    }
  | {
      type: 'edit'
      original: T
      replacement: T
    }

function isSameStroke<T extends HistoryStrokeLike>(left: T, right: T): boolean {
  return left === right || (
    left.id !== undefined &&
    right.id !== undefined &&
    left.id === right.id
  )
}

export function remapStrokeInHistory<T extends HistoryStrokeLike>(
  history: Array<StrokeHistoryEntry<T>>,
  oldStroke: T,
  restoredStroke: T,
  endExclusive = history.length,
): void {
  const limit = Math.min(endExclusive, history.length)
  for (let index = 0; index < limit; index++) {
    const entry = history[index]
    if (entry.type === 'draw') {
      if (isSameStroke(entry.stroke, oldStroke)) entry.stroke = restoredStroke
      continue
    }
    if (entry.type === 'edit') {
      if (isSameStroke(entry.original, oldStroke)) entry.original = restoredStroke
      if (isSameStroke(entry.replacement, oldStroke)) entry.replacement = restoredStroke
      continue
    }

    for (const replacement of entry.replacements) {
      if (isSameStroke(replacement.original, oldStroke)) {
        replacement.original = restoredStroke
      }
      replacement.fragments = replacement.fragments.map((fragment) =>
        isSameStroke(fragment, oldStroke) ? restoredStroke : fragment)
    }
  }
}
