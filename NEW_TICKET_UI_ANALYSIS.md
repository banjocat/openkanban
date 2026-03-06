# OpenKanban "New Ticket" UI Implementation Analysis

## Overview

The current "New Ticket" UI in OpenKanban implements a **single-panel modal form** that overlays the kanban board. This document analyzes the existing implementation to inform future improvements like adding a description preview panel.

---

## 1. The Modal/Form Entry Point

### Location: `internal/ui/view.go` - `renderTicketForm()`

**Lines: 821-1097**

The form is triggered when the app enters either:
- `ModeCreateTicket` - Creating a new ticket
- `ModeEditTicket` - Editing an existing ticket

**Key Code Flow:**
```go
// In View() method (view.go:59-60)
if m.mode == ModeCreateTicket || m.mode == ModeEditTicket {
    return m.renderWithOverlay(m.renderTicketForm())
}
```

The `renderWithOverlay()` function (view.go:1270-1279) centers the form on the screen using `lipgloss.Place()` with a dark overlay background.

### Form Triggering

**How to open the form:**

In `model.go` line 588-589 (handleNormalMode):
```go
case "n":
    return m.createNewTicket()
```

Or edit with "e" key (line 590-591).

---

## 2. Description Field - Current Implementation

### Current Approach: **Single Text Area**

**Field Definition:**
- Type: `textarea.Model` (Bubbletea textarea component)
- Variable: `m.descInput` (model.go:107)
- Form field ID: `formFieldDescription = 1` (model.go:54)

**Initialization (model.go:165-170):**
```go
di := textarea.New()
di.Placeholder = "Optional description..."
di.CharLimit = 0          // No character limit
di.SetWidth(40)
di.SetHeight(4)           // 4 lines tall
di.ShowLineNumbers = false
```

**Rendering (view.go:935-943):**
```go
fieldStartLines[formFieldDescription] = currentLine
lines = append(lines, descFocus+descLabel.Render("Description"))
lines = append(lines, "  "+descriptionStyle.Render("Details, context, or acceptance criteria"))
descLines := strings.Split(m.descInput.View(), "\n")
for _, dl := range descLines {
    lines = append(lines, "  "+dl)
}
lines = append(lines, "")
fieldEndLines[formFieldDescription] = len(lines) - 1
```

**Storage (board.go:73):**
```go
type Ticket struct {
    Description string `json:"description,omitempty"`
    // ...
}
```

**Saving (model.go:1565):**
```go
desc := strings.TrimSpace(m.descInput.Value())
// Later assigned to ticket:
ticket.Description = desc
```

---

## 3. Existing Two-Panel Layout Reference

### The Sidebar + Board Layout Pattern

OpenKanban **already implements a horizontal two-panel layout** with the sidebar and board!

**Location: `view.go` lines 40-51**

```go
var b strings.Builder

b.WriteString(m.renderHeader())
b.WriteString("\n")

sidebar := m.renderSidebar()
board := m.renderBoard()
if sidebar != "" {
    b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, sidebar, board))
} else {
    b.WriteString(board)
}
```

**Key Points:**
- Uses `lipgloss.JoinHorizontal()` to join panels side-by-side
- Sidebar is **optional** (can be toggled with `[` key)
- Sidebar width is **fixed at 24 characters** (model.go:255: `sidebarWidth: 24`)
- Board is **responsive** to remaining width

### Sidebar Structure (view.go:1730-1860)

The sidebar renders:
1. Project list with selection indicators
2. "All Projects" toggle
3. "+ Add project" option

Navigation:
- `j/k` or `Up/Down` to navigate
- `Enter` to select
- `[` or `l` to focus/unfocus sidebar
- `tab` to switch between sidebar and board

---

## 4. UI Mode Handling for Ticket Creation

### Mode Constants (model.go:27-44)

```go
type Mode string

const (
    ModeNormal        Mode = "NORMAL"
    ModeInsert        Mode = "INSERT"
    ModeCommand       Mode = "COMMAND"
    ModeHelp          Mode = "HELP"
    ModeConfirm       Mode = "CONFIRM"
    ModeCreateTicket  Mode = "CREATE"      // New ticket
    ModeEditTicket    Mode = "EDIT"        // Edit ticket
    ModeAgentView     Mode = "AGENT"
    ModeSettings      Mode = "SETTINGS"
    ModeShuttingDown  Mode = "SHUTTING_DOWN"
    ModeSpawning      Mode = "SPAWNING"
    ModeFilter        Mode = "FILTER"
    ModeCreateProject Mode = "NEW_PROJECT"
    ModeFileBrowser   Mode = "FILE_BROWSER"
)
```

### Mode Entry

**Create New Ticket:**
The `createNewTicket()` function initializes form state and sets `m.mode = ModeCreateTicket`

**Edit Existing Ticket:**
The `editTicket()` function loads ticket data into form fields and sets `m.mode = ModeEditTicket`

### Mode Handler: `handleCreateTicketMode()` and `handleEditTicketMode()`

**Location: model.go:1100-1188**

Both delegate to `handleTicketForm(msg tea.KeyMsg, isEdit bool)`

**Key Handlers:**
- `Tab` / `Shift+Tab` - Navigate between fields
- `Ctrl+S` - Save ticket
- `Enter` - On title field: save; on project field: select
- `Esc` - Cancel and return to ModeNormal
- `Up/Down/Left/Right` - Field-specific (priority, agent selection, etc.)

### Form Field Navigation

**Current Form Fields (8 total for creation, 9 for edit):**

```go
const (
    formFieldTitle       = 0   // textinput
    formFieldDescription = 1   // textarea
    formFieldBranch      = 2   // textinput (auto-generated)
    formFieldLabels      = 3   // textinput (comma-separated)
    formFieldPriority    = 4   // selector (1-5, arrows to change)
    formFieldWorktree    = 5   // toggle (Space or Y/N)
    formFieldAgent       = 6   // selector (arrows to navigate)
    formFieldBlockedBy   = 7   // multi-select with filter (edit only)
    formFieldProject     = 8   // selector (creation only)
)
```

**Navigation Flow (model.go:1469-1521):**
- `nextFormField()` - Cycles forward, skips locked fields
- `prevFormField()` - Cycles backward
- Skip `formFieldBranch` if `branchLocked` (after worktree created)
- Skip `formFieldAgent` if `agentLocked` (after agent spawned)

**Focus Handling:**
- `focusCurrentField()` calls `.Focus()` on the appropriate input
- `blurAllFormFields()` blurs all inputs when changing fields

---

## 5. Full Form Rendering Structure

### Form Content Layout (view.go:821-1097)

The form renders as a **vertical scroll-able list** with sections for each field:

```
┌─────────────────────────────────┐
│ ◈ New Ticket                    │
│                                 │
│ ▸ Title                    0/100│
│   Brief summary of the task     │
│   [text input showing...]       │
│                                 │
│   Description                   │
│   Details, context, or criteria │
│   [textarea line 1]             │
│   [textarea line 2]             │
│   [textarea line 3]             │
│   [textarea line 4]             │
│                                 │
│   Branch                        │
│   Auto-generated from title ... │
│   [text input]                  │
│                                 │
│   Labels                        │
│   Comma-separated tags (e.g...) │
│   [text input]                  │
│                                 │
│   ... [Priority, Worktree, Agent, etc.] ...
│                                 │
│ [Tab] Next  [Ctrl+S] Create  [Esc] Cancel
│                                 │
└─────────────────────────────────┘
```

**Key Features:**

1. **Field Indicators:**
   - `▸` prefix = Currently focused field (highlight)
   - `  ` prefix = Unfocused field
   
2. **Field Organization:**
   - Each field has: label + description + input(s) + blank line
   - Currently focused field is highlighted in blue/info color
   
3. **Scrolling:**
   - Form height: `formViewportHeight()` = `height - 10`
   - If total form > viewport height: adds scroll indicators `▲ N more above` / `▼ N more below`
   - Auto-scrolls to keep focused field visible
   
4. **Width Constraints:**
   - Form width: `min(60, width-4)` but not less than 40 chars
   - Responsive to terminal size

### Rendering Style Layers (view.go:830-837)

```go
titleStyle := lipgloss.NewStyle().Foreground(colors.success).Bold(true)
labelStyle := lipgloss.NewStyle().Foreground(colors.subtext)
activeLabelStyle := lipgloss.NewStyle().Foreground(colors.info).Bold(true)
lockedStyle := lipgloss.NewStyle().Foreground(colors.muted).Italic(true)
descriptionStyle := lipgloss.NewStyle().Foreground(colors.muted).Italic(true)
```

---

## 6. Form State Management

### Form State Variables (model.go:106-131)

```go
type Model struct {
    // Text inputs
    titleInput         textinput.Model
    descInput          textarea.Model        // Description
    branchInput        textinput.Model
    labelsInput        textinput.Model
    projectInput       textinput.Model
    blockerFilterInput textinput.Model
    
    // Field selection/values
    ticketPriority     int                   // 1-5
    ticketUseWorktree  bool
    ticketAgent        string
    agentListIndex     int                   // For agent selector
    projectListIndex   int                   // For project selector
    blockerListIndex   int                   // For blocker selector
    
    // Form state
    ticketFormField    int                   // Current active field (0-8)
    editingTicketID    board.TicketID        // If editing, which ticket
    branchLocked       bool                  // Locked after worktree created
    agentLocked        bool                  // Locked after agent spawned
    selectedBlockers   map[board.TicketID]bool
    blockerCandidates  []*board.Ticket
    
    // Scroll state
    formScrollOffset int
    formFieldLines   map[int]int             // Track line positions for each field
    
    // Project selection
    selectedProject    *project.Project
    showAddProjectForm bool                  // Show "add new project" form
}
```

### Initialization

When creating a new ticket:
1. Clear all form field inputs
2. Load agent names and pick default
3. Load project list and pre-select first project
4. Initialize blocker candidates from all non-archived tickets
5. Reset all state flags (`branchLocked = false`, `agentLocked = false`, etc.)
6. Set `mode = ModeCreateTicket`
7. Set focus to `formFieldTitle`

---

## 7. Key Form Interaction Patterns

### Text Input Fields (textinput)
- **Title, Branch, Labels, Project Input:**
- Standard text editing: typing appends, backspace deletes
- Updates propagate to `m.Update()` via `tea.KeyMsg`
- Character limits enforced (title: 100 chars)

### Text Area (textarea)
- **Description:**
- Multi-line editing with line wrapping
- No character limit
- Returns from `m.descInput.View()` as `string` with `\n` separators

### Selector Fields (custom rendering)
- **Priority (1-5):**
  - Rendered with radio buttons: `● Selected` vs `○ 1 2 3 4 5`
  - Navigate with `← →` or type `1-5` directly
  
- **Worktree (toggle):**
  - Rendered as: `● Worktree   ○ Main Repo`
  - Toggle with `Space`, `Enter`, or `Y/N`
  
- **Agent (list):**
  - Rendered as: `● agent1   ○ agent2   ○ agent3`
  - Navigate with `← →`
  - Show hint if locked after spawn
  
- **Blocked By (multi-select with filter):**
  - Filter input for searching
  - List of checkboxes: `[✓] ticket-name   [project]`
  - Navigate with `↑ ↓`, toggle with `Space/Enter`
  
- **Project (single-select list):**
  - Radio-button list: `● project1   ○ project2`
  - "+ Add project..." option at bottom
  - Navigate with `↑ ↓`

---

## 8. Overlay Rendering System

### `renderWithOverlay()` Pattern (view.go:1270-1279)

```go
func (m *Model) renderWithOverlay(overlay string) string {
    return lipgloss.Place(
        m.width,
        m.height,
        lipgloss.Center,
        lipgloss.Center,
        overlay,
        lipgloss.WithWhitespaceChars(" "),
        lipgloss.WithWhitespaceForeground(m.colors.base),
    )
}
```

**How it works:**
1. Takes the entire terminal canvas
2. Centers the overlay (form/dialog) both horizontally and vertically
3. Fills background with space characters in a dark color
4. Completely obscures the board underneath

**Similar Overlays:**
- Settings panel (view.go:1282-1354)
- Help dialog (view.go:730-800)
- Confirm dialog (view.go:690-809)
- Create Project form (view.go:1605-1674)

---

## 9. Mouse Support

### Form Mouse Handling (model.go:1020-1098)

The form supports:
1. **Wheel scrolling:** `Wheel Up/Down` to scroll form
2. **Field clicking:** Click on a field label/input to focus it
3. **Text input interaction:** Bubbletea's textinput handles mouse clicks

**Mouse Hit Testing Logic (model.go:1041-1080):**
- Calculate relative Y from form top
- Map Y ranges to form fields
- Set `m.ticketFormField` and call `focusCurrentField()`
- Special handling for project selector dropdown

---

## 10. Saving & Exit Flow

### Form Submission (model.go:1553-1617)

**Function: `saveTicketForm(isEdit bool)`**

1. **Validation:**
   - Title cannot be empty
   - Project must be selected
   
2. **Data Collection:**
   ```go
   title := strings.TrimSpace(m.titleInput.Value())
   desc := strings.TrimSpace(m.descInput.Value())
   branchName := strings.TrimSpace(m.branchInput.Value())
   labels := m.parseLabels(m.labelsInput.Value())
   blockedBy := m.collectSelectedBlockers()
   ```

3. **Create or Update:**
   - **Create mode:** `board.NewTicket()` with collected values
   - **Edit mode:** Update existing ticket fields
   - Save to `m.globalStore.Save(ticket)`

4. **Exit:**
   - Set `m.mode = ModeNormal`
   - Clear all form inputs
   - Refresh column display
   - Show success notification

### Cancellation

**Key: Esc**
- Sets `m.mode = ModeNormal`
- Calls `m.blurAllFormFields()`
- Clears `m.editingTicketID`
- Does NOT save

---

## 11. Current Limitations (for Design Reference)

1. **Single Panel:** Description is inline with form, no separate view
2. **No Markdown Support:** Plain text only, no preview
3. **No Rich Editing:** No syntax highlighting or formatting
4. **Height Constraint:** Description textarea limited to 4 visible lines in form
5. **Scroll Coordination:** When description is long, entire form scrolls, not just description
6. **No References:** Can't reference other tickets inline in description

---

## 12. Integration Points for Enhancement

If adding a two-panel layout (description + form), consider:

### Pattern to Follow: Sidebar + Board

The sidebar + board layout already provides:
- `lipgloss.JoinHorizontal()` for side-by-side rendering
- Responsive width distribution
- Tab-based focus switching (sidebar ↔ board)
- Clean separation of concerns

### Potential Two-Panel Form Structure

```
┌────────────────────────────────────────────────┐
│              ◈ New Ticket                      │
├─────────────────────┬──────────────────────────┤
│  Form Fields:       │                          │
│  ▸ Title           │   Description Preview    │
│    [input]         │                          │
│                    │   (live updated as      │
│  Description       │    user types in        │
│  [hidden]          │    form)                │
│                    │                          │
│  Branch            │                          │
│  [input]           │                          │
│                    │                          │
│  Priority          │                          │
│  ● Critical ○ High │                          │
│                    │                          │
│  ... more fields   │                          │
│                    │                          │
├─────────────────────┴──────────────────────────┤
│ [Tab] Cycle Fields  [Ctrl+S] Save  [Esc] Cancel
└────────────────────────────────────────────────┘
```

### Implementation Considerations

1. **Focus Management:**
   - Tab cycles through left panel fields
   - Preview updates in sync with `m.descInput`
   - Consider Tab+Shift focus to preview if markdown support added

2. **Width Distribution:**
   - Similar to sidebar pattern: fixed left + flexible right
   - Maybe 40 chars form + remaining for preview

3. **Scroll Synchronization:**
   - Form scrolls vertically (existing pattern)
   - Preview scrolls independently if description is long

4. **State Tracking:**
   - Keep existing `formFieldLines` map for form scrolling
   - Add similar tracking for preview content

---

## Summary

The current "New Ticket" UI is a **single-panel scrollable modal form** that:
- Uses `lipgloss.Place()` to center and overlay on the board
- Manages 8-9 form fields through a `ticketFormField` state variable
- Stores description as a `textarea.Model` with up to 4 visible lines
- Supports keyboard navigation (Tab/Shift+Tab) and mouse interaction
- Provides inline field selectors for priorities, agents, projects, etc.

The codebase already demonstrates a **two-panel layout pattern** (sidebar + board) that could serve as a reference for adding a description preview panel to the form.
