# Notes Context Menu Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a screenshot-style right-click menu in the notes editor body.

**Architecture:** Keep state, commands, and template markup inside `frontend/src/views/NotesView.vue`, reusing the current Tiptap command helpers. Add Playwright coverage in `frontend/e2e/specs/note.spec.ts` to lock the interaction.

**Tech Stack:** Vue 3, Tiptap, TypeScript, Playwright, CSS scoped styles.

---

### Task 1: E2E Coverage

**Files:**
- Modify: `frontend/e2e/specs/note.spec.ts`

- [ ] Add a Playwright test named `4. editor context menu opens and closes`.
- [ ] Right-click `.ne-body .ProseMirror`.
- [ ] Assert `.note-context-menu`, `.ncm-row`, and formatting buttons are visible.
- [ ] Click `.note-list` and assert the menu is hidden.

### Task 2: Menu State And Commands

**Files:**
- Modify: `frontend/src/views/NotesView.vue`

- [ ] Add context menu reactive state for open, x, y, submenu.
- [ ] Add handlers for contextmenu, outside click, Escape, scroll close, and viewport clamping.
- [ ] Add context menu command dispatcher that reuses `applyEditorCommand` for editor formatting and uses browser/Tiptap commands for cut/copy/delete.

### Task 3: Template And Styling

**Files:**
- Modify: `frontend/src/views/NotesView.vue`

- [ ] Add floating menu template under the editor body.
- [ ] Add grouped top icon buttons, middle format grid, and bottom rows.
- [ ] Add scoped CSS matching the screenshot: white card, 8px radius, subtle border, shadow, separators, compact icon buttons, disabled states, submenu flyout.

### Task 4: Verification

**Files:**
- Verify: `frontend`

- [ ] Run the targeted Playwright test and inspect failures.
- [ ] Run `npm.cmd run build`.
