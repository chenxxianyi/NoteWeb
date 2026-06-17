# Block Format Menu Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand the notes editor right-click paragraph submenu into a block-format menu with first-stage working commands.

**Architecture:** Keep this phase local to `frontend/src/views/NotesView.vue`, extending the existing context menu data model and command dispatcher. Add Playwright coverage in `frontend/e2e/specs/note.spec.ts`.

**Tech Stack:** Vue 3, Tiptap StarterKit, TaskList/TaskItem, Playwright.

---

### Task 1: E2E Coverage

**Files:**
- Modify: `frontend/e2e/specs/note.spec.ts`

- [ ] Add a test that opens the right-click menu, hovers the paragraph row, and asserts first-stage block menu entries are visible.
- [ ] Add a test that clicks `三级标题` and verifies the editor contains an `h3`.

### Task 2: Block Menu Data

**Files:**
- Modify: `frontend/src/views/NotesView.vue`

- [ ] Replace the simple paragraph submenu with grouped block menu items.
- [ ] Include headings 1-6, paragraph, promote/demote heading, table/formula/code tools/callout disabled placeholders, quote, lists, list indent, and insert paragraph above/below.

### Task 3: Command Support

**Files:**
- Modify: `frontend/src/views/NotesView.vue`

- [ ] Add commands for headings 3-6.
- [ ] Add commands for promote/demote heading.
- [ ] Add commands for list indent/outdent and insert paragraph above/below.
- [ ] Keep unsupported commands disabled with tooltips.

### Task 4: Verification

**Files:**
- Verify: `frontend`

- [ ] Run targeted Playwright tests.
- [ ] Run `npm.cmd run build`.
