# New Ticket UI - Code Flow Diagrams

## 1. Creating a New Ticket - State Flow

```
User presses 'n' in ModeNormal
         ↓
handleNormalMode() case "n"
         ↓
createNewTicket() called
         ↓
┌─────────────────────────────────────┐
│ 1. Clear form fields                │
│ 2. Initialize descInput, titleInput │
│ 3. Load agents, pick default        │
│ 4. Load projects, select first      │
│ 5. Load blocker candidates         │
│ 6. Reset flags:                     │
│    - branchLocked = false           │
│    - agentLocked = false            │
│ 7. Set m.mode = ModeCreateTicket    │
│ 8. Set ticketFormField = 0 (title)  │
│ 9. focusCurrentField() -> Focus on  │
│    titleInput                       │
│ 10. Return with Cmd: textinput.Blink
└─────────────────────────────────────┘
         ↓
Next Update() call:
   m.mode == ModeCreateTicket
         ↓
View() renders form:
   View() → renderWithOverlay(
              renderTicketForm()
            )
         ↓
Form appears on screen centered
```

## 2. Form Rendering - renderTicketForm() Flow

```
renderTicketForm() called
         ↓
Determine form mode (Create vs Edit)
         ↓
Calculate field indicators:
  - Which field is focused? → Set focus indicator "▸"
  - Which fields are locked? → Set lock indicator
         ↓
Build form content as []string:
┌─────────────────────────────────────────────────┐
│ Title section:                                   │
│  - fieldStartLines[0] = line 0                  │
│  - "▸ Title          0/100"                     │
│  - "  Brief summary..."                         │
│  - "  [titleInput.View()]"                      │
│  - ""                                           │
│  - fieldEndLines[0] = line 4                    │
│                                                 │
│ Description section:                            │
│  - fieldStartLines[1] = line 5                  │
│  - "  Description"                              │
│  - "  Details, context, or criteria"            │
│  - For each line in descInput.View():           │
│    - "  [line content]"                         │
│  - ""                                           │
│  - fieldEndLines[1] = line N                    │
│                                                 │
│ [Branch, Labels, Priority, Worktree, Agent,    │
│  BlockedBy, Project sections...]                │
└─────────────────────────────────────────────────┘
         ↓
Calculate scroll metrics:
  - totalLines = len(all lines)
  - viewportHeight = m.height - 10
  - needsScroll = totalLines > viewportHeight
         ↓
If needs scroll:
  - Find focused field's start/end lines
  - Auto-scroll to keep focused field visible
  - Set m.formScrollOffset accordingly
  - Add "▲ N more above" and "▼ N more below" indicators
         ↓
Extract visible lines:
  lines[formScrollOffset : formScrollOffset+availableHeight]
         ↓
Build footer:
  "[Tab] Next  [Ctrl+S] Create  [Esc] Cancel"
         ↓
Wrap in border:
  lipgloss.NewStyle()
    .Border(RoundedBorder)
    .BorderForeground(colors.success)
    .Padding(1, 2)
    .Width(formWidth)
    .Render(content)
         ↓
Return form string
```

## 3. Field Navigation - Tab/Shift+Tab Flow

```
User presses Tab while in form
         ↓
handleTicketForm(msg, isEdit)
         ↓
Case "tab":
  ├─ If showAddProjectForm and addProjectPath != "":
  │    └─ createProjectFromPath()
  │         └─ showAddProjectForm = false
  │         └─ Go to next field
  │
  └─ nextFormField(isEdit)
       └─ ┌─────────────────────────────────────┐
          │ 1. blurAllFormFields()               │
          │    (blur titleInput, descInput, ...) │
          │                                     │
          │ 2. m.ticketFormField++              │
          │                                     │
          │ 3. Calculate maxField:              │
          │    - If isEdit: maxField = 7        │
          │    - Else: maxField = 8             │
          │                                     │
          │ 4. Loop until valid field:          │
          │    - If > maxField: wrap to 0       │
          │    - If == formFieldBranch &&       │
          │      branchLocked: skip             │
          │    - If == formFieldAgent &&        │
          │      agentLocked: skip              │
          │    - Break on first valid           │
          │                                     │
          │ 5. focusCurrentField()              │
          │    - Focus appropriate input        │
          │    (titleInput, descInput, etc.)    │
          │                                     │
          │ 6. Reset form scroll to show field  │
          └─────────────────────────────────────┘
             ↓
             Return m, nil
             ↓
             Next Update() call renders form at new position
```

## 4. Field Input - Key Press Flow

```
User types "hello" in title field while focused on Title
         ↓
Update() receives tea.KeyMsg("h"), tea.KeyMsg("e"), etc.
         ↓
handleTicketForm(msg, isEdit)
         ↓
Match m.ticketFormField:
  case formFieldTitle (0):
    └─ m.titleInput.Update(msg)
       └─ titleInput handles key: appends "h" to value
       └─ Returns updated textinput.Model
       └─ Value is now "h"
       ↓
  case formFieldDescription (1):
    └─ m.descInput.Update(msg)
       └─ descInput handles key
       └─ Value updated with text
       ↓
  case formFieldBranch (2):
    └─ If not branchLocked:
         └─ m.branchInput.Update(msg)
       ↓
  [... similar for other text input fields ...]
       ↓
Return m, cmd (usually nil or textinput animation)
       ↓
Next View() call renders updated values:
  "  [hello|]"  (with cursor)
```

## 5. Description Rendering - From Input to View

```
m.descInput contains multi-line text
         ↓
In renderTicketForm(), when rendering description section:
         ↓
  descLines := strings.Split(m.descInput.View(), "\n")
  for _, dl := range descLines {
      lines = append(lines, "  "+dl)
  }
         ↓
If description is:
  "This is a task\nwith multiple\nlines"
         ↓
Result in form:
  "  Description"
  "  Details, context, or criteria"
  "  This is a task"
  "  with multiple"
  "  lines"
  ""
         ↓
If form is scrolled, these lines may be partially visible
based on formScrollOffset
         ↓
When focused, entire form auto-scrolls to show
description field fully
```

## 6. Saving - Submit Flow

```
User presses Ctrl+S
         ↓
handleTicketForm(msg, isEdit)
         ↓
Case "ctrl+s":
  └─ saveTicketForm(isEdit)
     └─ ┌──────────────────────────────────────────┐
        │ 1. Validate:                             │
        │    - title = titleInput.Value().Trim()   │
        │    - if title == "": notify error, exit  │
        │    - if selectedProject == nil: error    │
        │                                          │
        │ 2. Collect data:                         │
        │    - desc = descInput.Value().Trim()     │
        │    - branch = branchInput.Value().Trim() │
        │    - labels = parseLabels(...)           │
        │    - blockedBy = collectBlockers()       │
        │                                          │
        │ 3. Create or Update:                     │
        │                                          │
        │    IF isEdit:                            │
        │      ticket, _ = globalStore.Get(...)    │
        │      ticket.Title = title                │
        │      ticket.Description = desc           │
        │      ticket.Touch()                      │
        │                                          │
        │    ELSE:                                 │
        │      ticket = NewTicket(title, projID)   │
        │      ticket.Description = desc           │
        │      globalStore.Add(ticket)             │
        │                                          │
        │ 4. Persist:                              │
        │    globalStore.Save(ticket)              │
        │                                          │
        │ 5. Exit form:                            │
        │    - m.mode = ModeNormal                 │
        │    - blurAllFormFields()                 │
        │    - editingTicketID = ""                │
        │    - branchLocked = false                │
        │    - refreshColumnTickets()              │
        │    - notify("Created: " + title)         │
        │                                          │
        │ 6. Return m, nil                         │
        └──────────────────────────────────────────┘
             ↓
             View() now renders board instead of form
             ↓
             New/updated ticket visible in board
```

## 7. Cancellation - Exit Flow

```
User presses Esc
         ↓
handleTicketForm(msg, isEdit)
         ↓
Case "esc":
  ├─ If showAddProjectForm:
  │    └─ showAddProjectForm = false
  │    └─ addProjectPath.Blur()
  │    └─ return m, nil (stay in form, just close project picker)
  │
  └─ Else:
       └─ m.mode = ModeNormal
       └─ blurAllFormFields()
       └─ editingTicketID = ""
       └─ branchLocked = false
       └─ return m, nil
            ↓
            Next View() call renders board (no form)
            ↓
            Form closed, changes discarded
```

## 8. Two-Panel Layout Pattern (Sidebar Reference)

```
Main View() in view.go:
         ↓
var b strings.Builder
b.WriteString(m.renderHeader())
b.WriteString("\n")
         ↓
sidebar := m.renderSidebar()     // Returns string or ""
board := m.renderBoard()         // Returns string
         ↓
if sidebar != "" {
    b.WriteString(
        lipgloss.JoinHorizontal(
            lipgloss.Top,
            sidebar,  // Left panel (24 chars)
            board     // Right panel (remaining width)
        )
    )
} else {
    b.WriteString(board)
}
         ↓
Sidebar content:            Board content:
┌──────────────────┐        ┌──────────────────────────┐
│ All Projects     │ + Col1│ Col2│ Col3│             │
│ ● Project 1      │  │    │     │     │             │
│ ○ Project 2      │  │Ticket  │     │             │
│ ○ Project 3      │  │ Card   │     │             │
│ ○ + Add project  │  │        │     │             │
└──────────────────┘        └──────────────────────────┘
  sidebarWidth = 24        width - 24 - 1
```

## 9. Modal Overlay Rendering

```
Form is centered with:
         ↓
lipgloss.Place(
    m.width,          // 120 (full terminal width)
    m.height,         // 40 (full terminal height)
    lipgloss.Center,  // Horizontally centered
    lipgloss.Center,  // Vertically centered
    formString,       // The actual form
    lipgloss.WithWhitespaceChars(" "),        // Fill with spaces
    lipgloss.WithWhitespaceForeground(dark)   // Dark color
)
         ↓
Result:
┌──────────────────────────────────────────────┐
│                                              │
│       [Dark background with spaces]          │
│                                              │
│              ┌──────────────────┐            │
│              │  ◈ New Ticket    │            │
│              │  ▸ Title         │            │
│              │    [input]       │            │
│              │                  │            │
│              │  Description     │            │
│              │    [textarea]    │            │
│              │                  │            │
│              │  [Footer hints]  │            │
│              └──────────────────┘            │
│                                              │
│       [Dark background with spaces]          │
│                                              │
└──────────────────────────────────────────────┘

Form is centered both horizontally and vertically
Board underneath is completely obscured
```

## 10. Form Scroll Auto-Behavior

```
User navigates to Description field (formFieldDescription)
         ↓
renderTicketForm() calculates:
  startLine = fieldStartLines[1] = 5
  endLine = fieldEndLines[1] = 12
  fieldHeight = 12 - 5 + 1 = 8 lines
  viewportHeight = m.height - 10 = 30
         ↓
Is field visible?
  └─ if endLine >= m.formScrollOffset + viewportHeight:
       └─ m.formScrollOffset = endLine - viewportHeight + 1
  └─ if startLine < m.formScrollOffset:
       └─ m.formScrollOffset = startLine
         ↓
Example: User scrolls down too far
  prevScrollOffset = 20 (beyond description)
  startLine = 5
  5 < 20? YES
  m.formScrollOffset = 5  (auto-scroll back to show description)
         ↓
Next render includes lines 5-34 (visible portion)
Description field is now fully visible
```

## Key Interactions Summary

| User Action | Handler | Result |
|-------------|---------|--------|
| `n` | handleNormalMode() → createNewTicket() | Open form, focus title |
| `Tab` | handleTicketForm() → nextFormField() | Focus next field, auto-scroll |
| `Shift+Tab` | handleTicketForm() → prevFormField() | Focus prev field, auto-scroll |
| Type text | handleTicketForm() → m.fieldInput.Update() | Text appended to field |
| `Ctrl+S` | handleTicketForm() → saveTicketForm() | Validate, save, close form |
| `Esc` | handleTicketForm() | Close form, discard changes |
| Wheel scroll | handleTicketFormMouse() | Scroll form content |
| Click field | handleTicketFormMouse() | Focus that field |
