# OpenKanban New Ticket UI - Documentation Index

This directory contains comprehensive documentation about the "New Ticket" UI implementation in OpenKanban, including the modal form, description field handling, two-panel layout patterns, and UI mode management.

## Documents

### 1. [NEW_TICKET_UI_ANALYSIS.md](NEW_TICKET_UI_ANALYSIS.md)
**Complete architectural deep-dive** - Start here for understanding the full system.

**Covers:**
- Modal/form entry point and triggering mechanism
- Description field implementation (textarea.Model)
- Existing two-panel layout patterns (sidebar + board)
- UI mode system (ModeCreateTicket, ModeEditTicket)
- Full form rendering structure and layout
- Form state management and variables
- Key form interaction patterns
- Overlay rendering system
- Mouse support and interaction
- Form submission and exit flows
- Current limitations and design constraints
- Extension points for enhancement

**Best for:** Understanding the overall architecture, how pieces fit together, and design principles

---

### 2. [QUICK_REFERENCE_NEW_TICKET_UI.md](QUICK_REFERENCE_NEW_TICKET_UI.md)
**Fast lookup guide** - Use this when you need to find something quickly.

**Includes:**
- File locations and line numbers
- Key state variables with descriptions
- Form field constants and mappings
- Function signatures and locations
- Keyboard shortcuts reference
- Mouse interaction support
- Color usage and styling
- Form dimensions and constraints
- Extension points for adding new features

**Best for:** Finding specific code locations, variable names, function calls, and quick facts

---

### 3. [NEW_TICKET_UI_FLOW_DIAGRAMS.md](NEW_TICKET_UI_FLOW_DIAGRAMS.md)
**Visual code flows** - See how things work step-by-step.

**Includes:**
- Creating a new ticket state flow
- Form rendering pipeline
- Field navigation (Tab/Shift+Tab)
- Text input handling
- Description field rendering
- Form submission flow
- Cancellation flow
- Two-panel layout pattern
- Modal overlay rendering
- Form auto-scroll behavior
- Interaction summary table

**Best for:** Understanding the sequence of events, how state changes flow through the code, and debugging issues

---

## Quick Navigation

### I want to...

**Understand how the form works**
→ Start with [NEW_TICKET_UI_ANALYSIS.md](NEW_TICKET_UI_ANALYSIS.md) sections 1-5

**Find where something is in the code**
→ Use [QUICK_REFERENCE_NEW_TICKET_UI.md](QUICK_REFERENCE_NEW_TICKET_UI.md) file locations table

**Trace how user actions flow through the code**
→ Read [NEW_TICKET_UI_FLOW_DIAGRAMS.md](NEW_TICKET_UI_FLOW_DIAGRAMS.md)

**Add a two-panel description preview**
→ Read sections 3, 8, and 12 of [NEW_TICKET_UI_ANALYSIS.md](NEW_TICKET_UI_ANALYSIS.md) plus the sidebar pattern in [QUICK_REFERENCE_NEW_TICKET_UI.md](QUICK_REFERENCE_NEW_TICKET_UI.md)

**Understand form field management**
→ See section 4 and 7 of [NEW_TICKET_UI_ANALYSIS.md](NEW_TICKET_UI_ANALYSIS.md) and the field constants in [QUICK_REFERENCE_NEW_TICKET_UI.md](QUICK_REFERENCE_NEW_TICKET_UI.md)

**Debug form rendering**
→ Use [NEW_TICKET_UI_FLOW_DIAGRAMS.md](NEW_TICKET_UI_FLOW_DIAGRAMS.md) section "Form Rendering Pipeline" and [NEW_TICKET_UI_ANALYSIS.md](NEW_TICKET_UI_ANALYSIS.md) section 5

**Understand state management**
→ See [NEW_TICKET_UI_ANALYSIS.md](NEW_TICKET_UI_ANALYSIS.md) section 6 and [QUICK_REFERENCE_NEW_TICKET_UI.md](QUICK_REFERENCE_NEW_TICKET_UI.md) key state variables

---

## Key Files in Codebase

| File | Lines | Purpose |
|------|-------|---------|
| `internal/ui/model.go` | 3080 | Model state and event handlers |
| `internal/ui/view.go` | 2012 | Form rendering and display |
| `internal/board/board.go` | 160 | Ticket data structure |

---

## Key Concepts

### Form Modes
- **ModeCreateTicket** - Creating a new ticket
- **ModeEditTicket** - Editing an existing ticket

### Form Fields (in order)
1. Title (textinput)
2. Description (textarea)
3. Branch (textinput)
4. Labels (textinput)
5. Priority (selector 1-5)
6. Worktree (toggle)
7. Agent (selector)
8. Blocked By (multi-select, edit only)
9. Project (selector, creation only)

### Rendering Approach
- Single-panel modal form overlay
- Uses `lipgloss.Place()` to center on screen
- Scrollable if content exceeds viewport
- Auto-scrolls to keep focused field visible

### Description Field
- Type: `textarea.Model` (from Bubbletea)
- No character limit
- Plain text only
- Rendered inline with form fields

### Two-Panel Pattern Reference
The codebase demonstrates a working two-panel layout with:
- Sidebar (left, fixed 24 chars)
- Board (right, flexible)
- Uses `lipgloss.JoinHorizontal()` to join panels
- Tab key switches focus between panels

---

## Common Tasks

### Adding a keyboard shortcut in the form
1. Find `handleTicketForm()` in model.go (line 1108)
2. Add case in the switch statement
3. Implement the action

### Modifying form layout
1. Edit `renderTicketForm()` in view.go (line 821)
2. Modify the form content building logic (lines 924-1007)
3. Adjust scroll metrics if needed (lines 1011-1045)

### Adding a new form field
1. Add new constant in model.go (after line 61)
2. Add new input model in Model struct (around line 106)
3. Initialize in NewModel() (around line 159)
4. Add rendering in renderTicketForm() (view.go:821)
5. Add handler in handleTicketForm() (model.go:1108)
6. Update nextFormField/prevFormField navigation
7. Update saveTicketForm() to collect value

### Implementing a two-panel form
1. Review sidebar pattern in view.go:40-51
2. Create left panel render function
3. Create right panel render function
4. Use `lipgloss.JoinHorizontal()` to combine
5. Implement Tab to switch focus
6. Add state variable to track focused panel

---

## Structure Overview

```
New Ticket UI System
├── Entry Point: Press 'n' in board view
│   └── handleNormalMode() → createNewTicket()
│
├── State Management (model.go)
│   ├── Mode: ModeCreateTicket, ModeEditTicket
│   ├── Form fields: titleInput, descInput, branchInput, etc.
│   ├── Navigation: ticketFormField (0-8)
│   └── Scroll: formScrollOffset, formFieldLines
│
├── Event Handlers (model.go)
│   ├── handleCreateTicketMode()
│   ├── handleEditTicketMode()
│   ├── handleTicketForm() - Shared handler
│   ├── handleTicketFormMouse() - Mouse support
│   ├── nextFormField() - Tab navigation
│   ├── prevFormField() - Shift+Tab navigation
│   └── saveTicketForm() - Submit handler
│
├── Rendering (view.go)
│   ├── renderTicketForm() - Main form (lines 821-1097)
│   ├── renderWithOverlay() - Center overlay (lines 1270-1279)
│   ├── renderPrioritySelector() - Priority field (lines 1099-1129)
│   ├── renderWorktreeSelector() - Worktree field (lines 1131-1152)
│   ├── renderAgentSelector() - Agent field (lines 1154-1179)
│   ├── renderBlockerSelector() - Blocker field (lines 1181-1268)
│   └── renderProjectSelector() - Project field (lines 1552-1599)
│
└── Data Storage (board.go)
    └── Ticket struct with Description field
```

---

## Related Documentation

- [AGENTS.md](../AGENTS.md) - Architecture overview and code map
- [internal/ui/AGENTS.md](../internal/ui/AGENTS.md) - UI package patterns
- [internal/CLAUDE.md](../internal/CLAUDE.md) - Internal package conventions

---

## Tips for Navigation

1. **Start with the overview** - Read [NEW_TICKET_UI_ANALYSIS.md](NEW_TICKET_UI_ANALYSIS.md) first to get context
2. **Find code locations** - Use [QUICK_REFERENCE_NEW_TICKET_UI.md](QUICK_REFERENCE_NEW_TICKET_UI.md) file locations table
3. **Trace execution** - Use [NEW_TICKET_UI_FLOW_DIAGRAMS.md](NEW_TICKET_UI_FLOW_DIAGRAMS.md) to follow code paths
4. **Understand state** - The form state variables in [QUICK_REFERENCE_NEW_TICKET_UI.md](QUICK_REFERENCE_NEW_TICKET_UI.md) explain what each variable does
5. **See patterns** - The sidebar + board layout shows how to build two-panel UIs

---

Last Updated: 2026-03-06  
Documentation Version: 1.0
