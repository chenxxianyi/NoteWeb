# PDF Shape Tool Design

## Goal

Add screenshot-style outline shapes to the PDF annotation toolbar:

- Line
- Arrow
- Rectangle
- Ellipse

Shapes use the existing annotation color and width controls.

## Interaction

- Add a Shape tool with a four-item popup.
- Add a Select tool.
- Drag with a shape tool to create a shape preview; pointer-up persists it.
- Select mode can click a shape, drag its body to move it, or drag handles to resize it.
- Line and arrow expose start/end handles.
- Rectangle and ellipse expose corner and edge handles.
- Clicking empty space clears selection.
- Shapes are outline-only. Rotation, fill, multi-select, copy, and paste are out of scope.

## Data model

Keep `type: "drawing"` annotations. `position_data` distinguishes:

- Freehand: `{ tool: "pen" | "highlighter", width, points }`
- Shape: `{ tool: "shape", shapeType, width, start, end }`

Shapes remain editable objects until an eraser first changes them.

## Eraser behavior

Before erasing, convert a shape outline to one or more ordinary pen contours:

- Line: one contour.
- Arrow: shaft and two arrowhead contours.
- Rectangle: one closed contour.
- Ellipse: one sampled closed contour.

If the eraser changes any contour, persist all surviving contours as ordinary pen strokes and delete the original shape. The result is no longer shape-editable.

## Rendering and editing

- Keep persisted drawings on the existing annotation canvas.
- Use a second overlay canvas for shape creation previews, selection bounds, and handles.
- During move/resize, redraw the base layer once without the selected shape, then update only the overlay.
- On pointer-up, replace the original annotation through the existing atomic replace endpoint.

## Undo

Undo supports:

- Shape creation.
- Shape move.
- Shape resize.
- Shape-to-stroke erasure.

History references are remapped when replacement annotations receive new IDs.

## Verification

- Unit tests for shape contours, hit testing, handles, movement, resizing, and history remapping.
- Frontend unit tests and production build.
- No Playwright or E2E tests.

