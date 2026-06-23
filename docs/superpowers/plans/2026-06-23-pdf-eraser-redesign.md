# PDF Partial Eraser Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make freehand and rectangle erasers remove only covered portions of pen and highlighter strokes.

**Architecture:** Pure geometry helpers split polylines into surviving fragments. `PDFPageWrapper` owns pointer interaction and preview generation, while `PDFViewer` applies each replacement gesture to persisted annotation objects and undo history.

**Tech Stack:** Vue 3, TypeScript, Canvas 2D, Node test runner, Pinia, PDF.js.

---

### Task 1: Partial-erasure geometry

**Files:**
- Modify: `frontend/src/components/pdfAnnotationGeometry.ts`
- Create: `frontend/src/components/pdfAnnotationGeometry.test.ts`
- Modify: `frontend/package.json`

- [ ] Write Node unit tests that specify:
  - freehand erasure splits a line into left and right fragments;
  - rectangle erasure removes only the inside portion;
  - an untouched stroke is returned unchanged;
  - a fully covered stroke returns no fragments;
  - sparse long segments receive boundary points;
  - stroke width expands the effective erased region;
  - tiny fragments are discarded.
- [ ] Run `npm.cmd run test:unit` and verify failure because the split functions do not exist.
- [ ] Add `splitStrokeByEraserPath` and `splitStrokeByRect`, backed by segment subdivision, inside/outside classification, fragment assembly, and visible-length filtering.
- [ ] Run `npm.cmd run test:unit` and verify all geometry tests pass.

### Task 2: Preview replacement operations

**Files:**
- Modify: `frontend/src/components/PDFPageWrapper.vue`

- [ ] Replace index-based `strokesErased` events with a typed `strokesReplaced` event containing original indices and surviving stroke fragments.
- [ ] For freehand mode, continuously recompute fragment replacements from the gesture path and current eraser radius.
- [ ] For rectangle mode, continuously recompute fragment replacements from the selection rectangle.
- [ ] Redraw unaffected strokes plus replacement fragments during preview.
- [ ] Emit one replacement batch on pointer-up and clear preview state.

### Task 3: Persistent replacement and undo

**Files:**
- Modify: `frontend/src/components/PDFViewer.vue`
- Create: `frontend/src/components/pdfAnnotationHistory.ts`
- Create: `frontend/src/components/pdfAnnotationHistory.test.ts`

- [ ] Replace erase history entries with entries containing original strokes and persisted replacement strokes.
- [ ] Add helpers to create all replacement annotations, delete originals, and recover by refetching annotations after failure.
- [ ] Update local page strokes only after the replacement operation succeeds.
- [ ] Make erase undo delete replacement fragments and recreate original strokes.
- [ ] Remap older history references to restored annotation IDs so chained undo remains valid.
- [ ] Connect `PDFPageWrapper`'s `strokesReplaced` event to the replacement persistence handler.

### Task 4: Eraser mode labels and defaults

**Files:**
- Modify: `frontend/src/views/ReaderView.vue`

- [ ] Keep freehand erasure as the default.
- [ ] Rename the existing `stroke` mode in user-facing titles to local/freehand erasure.
- [ ] Keep rectangle mode and remove all wording that implies whole-stroke deletion.

### Task 5: Verification and review

**Files:**
- Review all files modified above.

- [ ] Run `npm.cmd run test:unit` in `frontend`.
- [ ] Run `npm.cmd run build` in `frontend`.
- [ ] Inspect `git diff --check` and the scoped diff for accidental unrelated edits.
- [ ] Request a focused code review of geometry, persistence failure handling, undo, and requirement coverage.
- [ ] Do not run Playwright or E2E tests.
