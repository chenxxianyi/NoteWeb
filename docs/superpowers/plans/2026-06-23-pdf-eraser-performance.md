# PDF Eraser Performance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove freehand eraser drag jank and reduce pointer-up persistence to one atomic request.

**Architecture:** The canvas provides incremental pixel preview during drag, while exact vector splitting runs once on pointer-up after path simplification and bounding-box filtering. The Go backend atomically creates fragments and soft-deletes originals in one transaction.

**Tech Stack:** Vue 3, TypeScript, Canvas 2D, Node test runner, Go, Gin, GORM, MySQL.

---

### Task 1: Geometry acceleration helpers

**Files:**
- Modify: `frontend/src/components/pdfAnnotationGeometry.ts`
- Modify: `frontend/src/components/pdfAnnotationGeometry.test.ts`

- [ ] Add failing tests for width-aware stroke bounds, gesture bounds, intersection filtering, and distance-based path simplification.
- [ ] Run `npm.cmd run test:unit` and confirm the new tests fail because helpers are absent.
- [ ] Implement `getStrokeBounds`, `getPathBounds`, `boundsIntersect`, and `simplifyPathByDistance`.
- [ ] Run unit tests and confirm they pass.

### Task 2: Frame-bounded drag preview

**Files:**
- Modify: `frontend/src/components/PDFPageWrapper.vue`

- [ ] Cache `getBoundingClientRect()` from pointer-down to pointer-up.
- [ ] Queue coalesced pointer points and flush freehand `destination-out` strokes with `requestAnimationFrame`.
- [ ] Replace canvas cursor and rectangle drawing with DOM overlays.
- [ ] Remove geometry splitting and full redraw from pointer-move.
- [ ] On pointer-up, simplify the path, filter candidates by cached bounds, and split only candidates.
- [ ] Cache `Path2D` per stroke and change the stroke watcher to shallow array updates.
- [ ] Restore the stored canvas on cancellation or persistence failure.

### Task 3: Atomic Go replacement endpoint

**Files:**
- Modify: `backend-go/internal/repository/annotation_repo.go`
- Modify: `backend-go/internal/service/annotation_service.go`
- Modify: `backend-go/internal/handlers/annotation.go`
- Modify: `backend-go/cmd/server/main.go`
- Create: `backend-go/internal/service/annotation_service_test.go`

- [ ] Introduce a repository interface so service conversion can be tested with a fake.
- [ ] Add failing service tests for replacement annotation conversion and repository error propagation.
- [ ] Implement request/response conversion in the service.
- [ ] Implement ownership validation, create, and soft-delete in one GORM transaction.
- [ ] Add the authenticated `POST /annotations/replace` route.
- [ ] Run `go test ./...`.

### Task 4: Frontend batch persistence

**Files:**
- Modify: `frontend/src/api/annotation.ts`
- Modify: `frontend/src/stores/annotationStore.ts`
- Modify: `frontend/src/components/PDFViewer.vue`

- [ ] Add the typed batch replacement API and store method.
- [ ] Replace sequential fragment POST/delete loops with one batch request.
- [ ] Use the same batch request for erase undo.
- [ ] Remove best-effort multi-request rollback.
- [ ] Tell the page wrapper to restore its preview if the transaction fails.

### Task 5: Verification

**Files:**
- Review all files above.

- [ ] Run `npm.cmd run test:unit`.
- [ ] Run `npm.cmd run build`.
- [ ] Run `go test ./...`.
- [ ] Run `go build ./cmd/server`.
- [ ] Run the synthetic eraser benchmark and compare with the recorded baseline.
- [ ] Run `git diff --check`.
- [ ] Do not run Playwright or E2E tests.
