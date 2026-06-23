# PDF Partial Eraser Design

## Goal

Change PDF drawing erasure from whole-stroke deletion to geometric partial erasure. Keep two modes:

- Freehand eraser: remove only the portions covered by the circular eraser path.
- Rectangle eraser: remove only the portions inside the selected rectangle.

Both modes apply to pen and highlighter strokes.

## Current cause

`PDFPageWrapper.vue` currently hit-tests strokes and emits their array indices. `PDFViewer.vue` then deletes every matched annotation, so touching any part of a stroke removes the complete stroke. Rectangle mode uses the same whole-object deletion model.

## Geometry

Move partial-erasure logic into `pdfAnnotationGeometry.ts` as pure functions.

- Treat a stored stroke as a polyline.
- For freehand erasure, classify sampled points and subdivided segment points by distance to the eraser path. The effective radius includes half the stroke width.
- For rectangle erasure, classify points by whether they are inside the rectangle, with half the stroke width added as padding.
- Split every affected stroke into contiguous outside fragments.
- Insert boundary points using segment subdivision so the visible cut follows the eraser boundary instead of jumping between sparse input samples.
- Discard fragments shorter than a small visible-length threshold.
- Preserve the original tool, color, and width on every fragment.

## Interaction and preview

- The default eraser mode is freehand partial erasure.
- The mode toggle switches between freehand and rectangle erasure; whole-stroke mode is removed.
- While dragging, compute replacements against the original strokes and redraw the surviving fragments as a preview.
- Persist only on pointer-up.
- The eraser cursor and rectangle selection remain visible during interaction.

## Persistence

`PDFPageWrapper` emits replacement operations containing the original stroke index and its surviving fragments. `PDFViewer` owns persistence.

For each gesture:

1. Keep the pre-gesture strokes in memory.
2. Create annotations for all surviving fragments.
3. Delete the replaced original annotations.
4. Replace local strokes only after persistence succeeds.
5. If an API operation fails, remove newly created fragments where possible, restore any deleted originals where possible, refetch annotations, and redraw.

An erase gesture that removes an entire short stroke naturally produces zero fragments.

## Undo

One pointer gesture is one history entry. Undo removes the saved fragments and recreates the original strokes, then redraws the page. Draw undo behavior remains unchanged.

Because recreated annotations receive new backend IDs, undo remaps older history references from the deleted ID to the restored ID. This keeps subsequent undo operations valid across repeated erase/undo chains.

## Verification

- Add unit tests for freehand splitting, rectangle splitting, sparse-segment boundary handling, unaffected strokes, complete removal, fragment filtering, pen width, and highlighter width.
- Run the unit tests, Vue/TypeScript type checking, and the production build.
- Do not run Playwright or E2E tests unless explicitly requested.

## Non-goals

- Editing the source PDF.
- Pixel-mask persistence.
- Whole-stroke erasure.
- Backend schema changes.
