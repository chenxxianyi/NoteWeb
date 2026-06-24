import type { Bounds, PointLike } from './pdfAnnotationGeometry.ts'
import type { TextDrawing } from './pdfDrawingTypes.ts'

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
