export interface Point {
  x: number
  y: number
}

export type ShapeType = 'line' | 'arrow' | 'rectangle' | 'ellipse'

export interface FreehandStroke {
  id?: number
  tool: 'pen' | 'highlighter'
  color: string
  width: number
  points: Point[]
}

export interface ShapeDrawing {
  id?: number
  tool: 'shape'
  shapeType: ShapeType
  color: string
  width: number
  start: Point
  end: Point
}

export type TextErasure =
  | { type: 'path'; radius: number; points: Point[] }
  | { type: 'rect'; start: Point; end: Point }

export interface TextDrawing {
  id?: number
  tool: 'text'
  color: string
  fontSize: number
  text: string
  x: number
  y: number
  width: number
  height: number
  erasures?: TextErasure[]
}

export type Drawing = FreehandStroke | ShapeDrawing | TextDrawing

export interface DrawingReplacement {
  index: number
  fragments: Drawing[]
}

export type PDFActiveTool =
  | 'none'
  | 'pen'
  | 'highlighter'
  | 'eraser'
  | 'select'
  | 'shape'
  | 'text'

export function isShapeDrawing(drawing: Drawing): drawing is ShapeDrawing {
  return drawing.tool === 'shape'
}

export function isFreehandStroke(drawing: Drawing): drawing is FreehandStroke {
  return drawing.tool === 'pen' || drawing.tool === 'highlighter'
}

export function isTextDrawing(drawing: Drawing): drawing is TextDrawing {
  return drawing.tool === 'text'
}

export function pointsEqual(left: Point[], right: Point[]): boolean {
  return left.length === right.length && left.every((point, index) =>
    point.x === right[index].x && point.y === right[index].y)
}

export function drawingsEqual(left: Drawing, right: Drawing): boolean {
  if (left.tool !== right.tool || left.color !== right.color) {
    return false
  }

  if (isShapeDrawing(left) || isShapeDrawing(right)) {
    return isShapeDrawing(left) &&
      isShapeDrawing(right) &&
      left.shapeType === right.shapeType &&
      left.width === right.width &&
      left.start.x === right.start.x &&
      left.start.y === right.start.y &&
      left.end.x === right.end.x &&
      left.end.y === right.end.y
  }

  if (isTextDrawing(left) || isTextDrawing(right)) {
    return isTextDrawing(left) &&
      isTextDrawing(right) &&
      textDrawingsEqual(left, right)
  }

  if (left.width !== right.width) return false
  return pointsEqual(left.points, right.points)
}

function textErasuresEqual(left: TextErasure[] = [], right: TextErasure[] = []): boolean {
  return left.length === right.length && left.every((erasure, index) => {
    const other = right[index]
    if (!other || erasure.type !== other.type) return false
    if (erasure.type === 'path') {
      return other.type === 'path' &&
        erasure.radius === other.radius &&
        pointsEqual(erasure.points, other.points)
    }
    return other.type === 'rect' &&
      erasure.start.x === other.start.x &&
      erasure.start.y === other.start.y &&
      erasure.end.x === other.end.x &&
      erasure.end.y === other.end.y
  })
}

export function textDrawingsEqual(left: TextDrawing, right: TextDrawing): boolean {
  return left.color === right.color &&
    left.fontSize === right.fontSize &&
    left.text === right.text &&
    left.x === right.x &&
    left.y === right.y &&
    left.width === right.width &&
    left.height === right.height &&
    textErasuresEqual(left.erasures, right.erasures)
}

export function drawingReplacementChangesDrawing(
  original: Drawing,
  fragments: Drawing[],
): boolean {
  return !(fragments.length === 1 && drawingsEqual(original, fragments[0]))
}

export function sameDrawingIdentity(left: Drawing | undefined, right: Drawing): boolean {
  if (!left) return false
  return (
    left === right ||
    (left.id !== undefined && right.id !== undefined && left.id === right.id)
  )
}
