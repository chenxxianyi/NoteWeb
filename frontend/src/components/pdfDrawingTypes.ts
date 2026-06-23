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

export type Drawing = FreehandStroke | ShapeDrawing

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

export function isShapeDrawing(drawing: Drawing): drawing is ShapeDrawing {
  return drawing.tool === 'shape'
}

export function isFreehandStroke(drawing: Drawing): drawing is FreehandStroke {
  return drawing.tool === 'pen' || drawing.tool === 'highlighter'
}

