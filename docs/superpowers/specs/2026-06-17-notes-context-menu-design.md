# Notes Context Menu Design

## Goal

Add a screenshot-style right-click menu inside the notes editor body. The menu prioritizes visual fidelity and discoverability; clipboard actions that browsers restrict can appear disabled or show a short status message.

## Scope

- Trigger only from the Tiptap note body, not the note list, title input, toolbar, or app shell.
- Use a floating menu with grouped icon buttons and text rows matching the provided reference.
- Reuse existing editor commands for formatting.
- Keep implementation local to `frontend/src/views/NotesView.vue` for this iteration.

## Behavior

- Right-clicking the editor body prevents the native context menu and opens the custom menu at the pointer position.
- The menu closes when clicking outside it, pressing Escape, scrolling the editor, or executing a command.
- Direct actions: cut, copy, delete selection, bold, italic, inline code, link, quote, ordered list, unordered list, task list.
- Submenu rows: `复制 / 粘贴为...`, `段落`, and `插入`.
- Paste-related restricted options are shown as disabled helper choices in the submenu.
- Position is clamped so the menu stays inside the viewport.

## Testing

- Add Playwright coverage to open the menu with right click and assert major controls are visible.
- Assert clicking outside closes the menu.
- Keep existing build verification.
