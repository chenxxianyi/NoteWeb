# Delete Note Button Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a visible, confirmed delete action for the currently selected note.

**Architecture:** Reuse the existing `noteStore.remove(id)` API path. Keep the UI local to `frontend/src/views/NotesView.vue`, placing a trash button beside the existing save action and selecting the next available note after deletion.

**Tech Stack:** Vue 3, Pinia, lucide-vue-next, Playwright.

---

### Task 1: Delete Current Note

**Files:**
- Modify: `frontend/e2e/specs/note.spec.ts`
- Modify: `frontend/src/views/NotesView.vue`

- [ ] **Step 1: Write the failing test**

Add a Playwright test that creates two notes, deletes the active one through the footer trash button, confirms the dialog, and verifies the deleted note leaves the list while another note remains selected.

- [ ] **Step 2: Run test to verify it fails**

Run: `npm.cmd exec playwright test --config e2e/playwright.config.ts e2e/specs/note.spec.ts -g "deletes the current note"`

Expected: FAIL because the delete button is not present.

- [ ] **Step 3: Write minimal implementation**

Add `deleteCurrentNote()`, call `noteStore.remove(id)`, pick the next or previous filtered note, and show a status toast. Add a footer trash icon button with accessible label.

- [ ] **Step 4: Run test and build**

Run: `npm.cmd exec playwright test --config e2e/playwright.config.ts e2e/specs/note.spec.ts -g "deletes the current note"`

Run: `npm.cmd run build`

Expected: both commands exit 0.
