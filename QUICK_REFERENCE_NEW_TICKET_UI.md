# Quick Reference: New Ticket UI Components

## File Locations

| Component | File | Lines |
|-----------|------|-------|
| Form rendering | `internal/ui/view.go` | 821-1097 |
| Form state/init | `internal/ui/model.go` | 106-131, 165-170 |
| Form handlers | `internal/ui/model.go` | 1100-1188 |
| Field navigation | `internal/ui/model.go` | 1469-1521 |
| Form submission | `internal/ui/model.go` | 1553-1617 |
| Mode constants | `internal/ui/model.go` | 27-44 |
| Overlay rendering | `internal/ui/view.go` | 1270-1279 |
| Sidebar (reference) | `internal/ui/view.go` | 40-51, 1730-1860 |
| Ticket struct | `internal/board/board.go` | 69-98 |

## Key State Variables

```go
m.mode                    // Current mode: ModeCreateTicket, ModeEditTicket, etc.
m.ticketFormField         // Which field is focused: 0-8
m.descInput               // textarea.Model for description
m.titleInput              // textinput.Model for title
m.branchInput             // textinput.Model for branch
m.labelsInput             // textinput.Model for labels
m.ticketPriority          // int: 1-5
m.ticketUseWorktree       // bool: use worktree or main repo
m.ticketAgent             // string: selected agent name
m.selectedProject         // *project.Project: selected project
m.selectedBlockers        // map[TicketID]bool: multi-selected blockers
m.formScrollOffset        // int: vertical scroll position in form
m.formFieldLines          // map[int]int: track line positions for each field
m.editingTicketID         // TicketID: which ticket being edited (empty if creating)
m.branchLocked            // bool: locked after worktree created
m.agentLocked             // bool: locked after agent spawned
```

## Form Field Constants

```go
formFieldTitle       = 0  // textinput
formFieldDescription = 1  // textarea
formFieldBranch      = 2  // textinput
formFieldLabels      = 3  // textinput
formFieldPriority    = 4  // selector
formFieldWorktree    = 5  // toggle
formFieldAgent       = 6  // selector
formFieldBlockedBy   = 7  // multi-select (edit only)
formFieldProject     = 8  // selector (creation only)
```

## Key Functions

### Entry Points
- `createNewTicket()` (model.go) - Open form for new ticket
- `editTicket()` (model.go) - Open form to edit existing ticket

### Rendering
- `renderTicketForm()` (view.go:821) - Main form rendering
- `renderWithOverlay(overlay string)` (view.go:1270) - Center overlay on screen
- `renderPrioritySelector()` (view.go:1099) - Render priority field
- `renderWorktreeSelector()` (view.go:1131) - Render worktree field
- `renderAgentSelector()` (view.go:1154) - Render agent field
- `renderBlockerSelector()` (view.go:1181) - Render blocker multi-select
- `renderProjectSelector()` (view.go:1552) - Render project field

### Handlers
- `handleCreateTicketMode(msg)` (model.go:1100) - Key handling for creation
- `handleEditTicketMode(msg)` (model.go:1104) - Key handling for editing
- `handleTicketForm(msg, isEdit)` (model.go:1108) - Shared form handler
- `handleTicketFormMouse(msg)` (model.go:1020) - Mouse handling

### Navigation
- `nextFormField(isEdit)` (model.go:1469) - Move to next field
- `prevFormField(isEdit)` (model.go:1496) - Move to previous field
- `focusCurrentField()` (model.go:1532) - Focus current field's input
- `blurAllFormFields()` (model.go:1523) - Blur all inputs

### Submission
- `saveTicketForm(isEdit)` (model.go:1553) - Validate, save, and exit form

## Keyboard Shortcuts (Form Mode)

| Key | Action |
|-----|--------|
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `Ctrl+S` | Save ticket |
| `Enter` | On title: save; on project: select |
| `Esc` | Cancel and exit |
| `↑/↓/←/→` | Field-specific (priority, agents, etc.) |
| `Space` | Toggle (worktree, blockers) |
| `1-5` | Priority quick-select |
| `Y/N` | Worktree quick-select |
| `Wheel Up/Down` | Scroll form |

## Mouse Support

- **Wheel scroll** - Scroll form vertically
- **Click on label/input** - Focus that field
- **Click in input area** - Position cursor in text

## Render Pattern: Sidebar + Board (Reference)

The existing sidebar + board layout shows how to build two-panel forms:

```go
// In view.go:40-51
sidebar := m.renderSidebar()
board := m.renderBoard()
if sidebar != "" {
    b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, sidebar, board))
}
```

Key techniques:
- `lipgloss.JoinHorizontal(align, ...panels)` - Side-by-side layout
- Fixed sidebar width (24 chars), flexible board width
- Tab key switches focus between panels
- Each panel renders independently

## Description Field Details

### Current Implementation
- Type: `textarea.Model` (Bubbletea component)
- Width: 40 characters (form field)
- Height: 4 lines (in form view)
- Storage: `Ticket.Description` field (string, no limit)
- No markdown support, plain text only

### Rendering (view.go:935-943)
```go
// Split by newlines and render each line
descLines := strings.Split(m.descInput.View(), "\n")
for _, dl := range descLines {
    lines = append(lines, "  "+dl)
}
```

## Form Overlay Rendering

The form uses `lipgloss.Place()` to center a dialog:

```go
lipgloss.Place(
    m.width,           // Full terminal width
    m.height,          // Full terminal height
    lipgloss.Center,   // Horizontally centered
    lipgloss.Center,   // Vertically centered
    overlay,           // The form content
    lipgloss.WithWhitespaceChars(" "),      // Dark background
    lipgloss.WithWhitespaceForeground(dark), // Color
)
```

## Form Dimensions

- **Width:** `min(60, m.width-4)` but >= 40 chars
- **Height:** `m.height - 10` (for scrollable viewport)
- **Total form lines:** Variable, depends on field count and descriptions
- **Scroll indicator:** "▲ N more above" / "▼ N more below"

## Color Usage

- Form title: `colors.success` (green)
- Field labels (unfocused): `colors.subtext`
- Field labels (focused): `colors.info` (blue)
- Field descriptions: `colors.muted` (gray, italic)
- Locked fields: `colors.muted` (gray, italic)
- Border: `colors.success` (green)
- Overlay background: Dark/base color

## Extension Points

To add a description preview panel:

1. **Rendering:**
   - Create `renderDescriptionPreview()` function
   - Use `lipgloss.JoinHorizontal()` to place preview next to form fields
   - Update `renderTicketForm()` to return two-panel layout

2. **State:**
   - May need to add `showDescriptionPreview bool` flag
   - Could use Alt+Tab or dedicated key to toggle preview focus

3. **Scrolling:**
   - Form scrolls independently (existing `formScrollOffset`)
   - Preview has its own scroll offset if needed
   - `formFieldLines` helps form auto-scroll, may need similar for preview

4. **Mouse Interaction:**
   - Update `handleTicketFormMouse()` to account for preview panel width
   - May need separate scroll/click zones for left and right panels
