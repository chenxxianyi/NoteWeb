# PDF Eraser Performance Design

## Goal

Keep freehand and rectangle partial erasure visually accurate while making pointer movement remain responsive on annotation-heavy PDF pages.

## Confirmed bottlenecks

- Every `pointermove` runs geometric splitting against every stroke on the page.
- Freehand hit testing compares every sampled stroke point with the complete accumulated eraser path, so work grows as the gesture gets longer.
- Every move clears and redraws the complete annotation canvas.
- Pointer-up persists fragments and deletions through sequential HTTP requests.

## Interaction architecture

### Drag preview

- Freehand erasure uses Canvas `destination-out` incrementally against only the newest pointer segment.
- Pointer events are coalesced and flushed at most once per animation frame.
- The visible eraser cursor is a DOM overlay, so moving it does not redraw annotations.
- Rectangle mode updates only a DOM selection rectangle while dragging.
- Neither mode runs geometric splitting during pointer movement.

### Pointer-up geometry

- Cache each stroke's width-aware bounding box.
- Simplify the recorded eraser path with a small distance threshold.
- Build a gesture bounding box and reject non-intersecting strokes before splitting.
- Run the existing exact split algorithm once, only for candidate strokes.
- Keep the current width-aware pen and highlighter behavior.

### Canvas redraw

- Cache `Path2D` objects for stable stroke objects.
- Remove the deep Vue watcher; parent code already replaces stroke arrays.
- Cache the canvas client rectangle for the duration of a gesture.
- On cancellation or persistence failure, redraw stored strokes to restore the preview.

## Persistence

Add `POST /api/v1/annotations/replace` to the Go backend.

The request contains:

- `document_id`
- annotation IDs to delete
- replacement drawing annotations to create

The repository validates ownership and document membership, creates replacement annotations, and soft-deletes originals inside one database transaction. The endpoint returns created annotations in request order.

The same endpoint is used for erase and erase-undo, reducing each operation to one request and removing client-side partial rollback.

## Verification

- Unit tests for path simplification, bounds filtering, and existing split behavior.
- Go service tests for batch request conversion and error propagation.
- Synthetic eraser benchmark before and after optimization.
- Frontend unit tests and production build.
- `go test ./...` and `go build ./cmd/server`.
- No Playwright or E2E tests.

## Success criteria

- No geometric splitting or full annotation redraw occurs in `pointermove`.
- Synthetic 100-stroke/200-point candidate filtering stays below the 16.7 ms frame budget because drag frames do only incremental preview work.
- One erase gesture produces one persistence request.

## Verification result

With 100 strokes, 200 points per stroke, and a 600-point freehand gesture:

- Previous full-page/full-path calculation: 284.03 ms median per move.
- Optimized drag path: zero geometry calls per move.
- Pointer-up exact calculation: 16.19 ms median after reducing the path to 167 points and filtering to 20 candidate strokes.
