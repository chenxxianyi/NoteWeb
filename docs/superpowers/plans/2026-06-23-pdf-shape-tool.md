# PDF Shape Tool Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add editable line, arrow, rectangle, and ellipse annotations to the PDF reader.

**Architecture:** A pure TypeScript geometry module defines shape contours, hit testing, handles, movement, and resizing. The page wrapper renders persisted drawings on the base canvas and shape previews/selection controls on an overlay canvas; PDFViewer persists all drawing variants through the existing annotation APIs.

**Tech Stack:** Vue 3, TypeScript, Canvas 2D, Node test runner, Pinia.

---

### Task 1: Shape geometry

**Files:**
- Create: `frontend/src/components/pdfShapeGeometry.ts`
- Create: `frontend/src/components/pdfShapeGeometry.test.ts`
- Modify: `frontend/package.json`

- [ ] Add failing tests for line, arrow, rectangle, and ellipse contours.
- [ ] Add failing tests for bounds, hit testing, handles, move, and resize.
- [ ] Run unit tests and confirm the shape helpers are missing.
- [ ] Implement the pure shape helpers.
- [ ] Run unit tests and confirm they pass.

### Task 2: Shared drawing types and history

**Files:**
- Create: `frontend/src/components/pdfDrawingTypes.ts`
- Modify: `frontend/src/components/pdfAnnotationHistory.ts`
- Modify: `frontend/src/components/pdfAnnotationHistory.test.ts`

- [ ] Define freehand, shape, and drawing union types.
- [ ] Extend history remapping to edit replacements.
- [ ] Add a failing history test, then implement and verify it.

### Task 3: Canvas creation, selection, move, resize, and erasure conversion

**Files:**
- Modify: `frontend/src/components/PDFPageWrapper.vue`

- [ ] Accept the drawing union and shape type props.
- [ ] Add an overlay canvas for shape previews and selection controls.
- [ ] Draw all four shape types.
- [ ] Implement shape creation.
- [ ] Implement topmost-shape selection and empty-space deselection.
- [ ] Implement body drag movement and handle resizing.
- [ ] Convert shapes to pen contours before partial erasure.
- [ ] Emit creation, edit, and erasure replacement events.

### Task 4: Persistence and undo

**Files:**
- Modify: `frontend/src/components/PDFViewer.vue`

- [ ] Load and serialize both freehand and shape annotations.
- [ ] Persist shape creation.
- [ ] Persist move/resize through atomic replacement.
- [ ] Add edit history and undo.
- [ ] Preserve shape-to-stroke erasure undo.

### Task 5: Toolbar

**Files:**
- Modify: `frontend/src/views/ReaderView.vue`

- [ ] Extend the active tool union with select and shape.
- [ ] Add a Select button.
- [ ] Add a Shape button and popup with the four shape choices.
- [ ] Pass the selected shape type to PDFViewer.
- [ ] Show color and width only for tools that use them.

### Task 6: Verification

**Files:**
- Review all files above.

- [ ] Run `npm.cmd run test:unit`.
- [ ] Run `npm.cmd run build`.
- [ ] Run `git diff --check`.
- [ ] Verify no unrelated files changed.
- [ ] Do not run Playwright or E2E tests.
