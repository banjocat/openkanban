package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// DirBrowser implements shell-like tab-completion for directory selection.
// It's confined to the homeDir and cannot escape with "../".
type DirBrowser struct {
	homeDir       string   // $HOME - cannot escape this boundary
	currentPath   string   // Accumulated path (relative to homeDir)
	matches       []string // Matching subdirectories at current level
	selectedIndex int      // Currently selected match (for j/k navigation)
	showMatches   bool     // Whether popup should be shown
	scrollOffset  int      // Top visible item in the list
	maxVisibleRows int     // Maximum rows to display (set by caller)
}

// New creates a new DirBrowser starting at homeDir.
func NewDirBrowser(homeDir string) *DirBrowser {
	return &DirBrowser{
		homeDir:        homeDir,
		currentPath:    "",
		matches:        []string{},
		selectedIndex:  0,
		showMatches:    false,
		scrollOffset:   0,
		maxVisibleRows: 10, // Default to 10 visible rows
	}
}

// GetMatches returns subdirectories matching the partial path.
// If partial path is empty, returns all subdirs in currentPath.
// Returns empty slice if path doesn't exist.
func (db *DirBrowser) GetMatches(partialPath string) []string {
	// Build the directory to list from
	currentFullPath := filepath.Join(db.homeDir, db.currentPath)

	// Security: ensure we don't escape homeDir
	absPath, err := filepath.Abs(currentFullPath)
	if err != nil {
		return []string{}
	}
	relPath, err := filepath.Rel(db.homeDir, absPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return []string{} // Can't escape homeDir
	}

	// List entries in current directory
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return []string{}
	}

	var matches []string
	partialPath = strings.TrimSpace(partialPath)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue // Only directories
		}
		name := entry.Name()
		if partialPath == "" || strings.HasPrefix(name, partialPath) {
			matches = append(matches, name)
		}
	}

	sort.Strings(matches)
	return matches
}

// UpdateMatches refreshes the matches list for the current state.
// Call this after path changes.
func (db *DirBrowser) UpdateMatches() {
	db.matches = db.GetMatches("")
	db.selectedIndex = 0
	db.scrollOffset = 0
	db.showMatches = len(db.matches) > 0
}

// SetMaxVisibleRows sets the maximum number of directory rows to display.
func (db *DirBrowser) SetMaxVisibleRows(rows int) {
	if rows > 0 {
		db.maxVisibleRows = rows
	}
}

// GetVisibleMatches returns the matches that fit in the current viewport.
func (db *DirBrowser) GetVisibleMatches() []string {
	if len(db.matches) == 0 {
		return []string{}
	}

	end := db.scrollOffset + db.maxVisibleRows
	if end > len(db.matches) {
		end = len(db.matches)
	}

	return db.matches[db.scrollOffset:end]
}

// GetScrollInfo returns the scroll position for display (e.g., "3/50").
func (db *DirBrowser) GetScrollInfo() string {
	if len(db.matches) == 0 {
		return ""
	}
	visibleEnd := db.scrollOffset + db.maxVisibleRows
	if visibleEnd > len(db.matches) {
		visibleEnd = len(db.matches)
	}
	return fmt.Sprintf("%d-%d/%d", db.scrollOffset+1, visibleEnd, len(db.matches))
}

// ensureSelectionVisible adjusts scroll offset to keep selected item in view.
func (db *DirBrowser) ensureSelectionVisible() {
	if db.selectedIndex < db.scrollOffset {
		db.scrollOffset = db.selectedIndex
	} else if db.selectedIndex >= db.scrollOffset+db.maxVisibleRows {
		db.scrollOffset = db.selectedIndex - db.maxVisibleRows + 1
	}
}

// TabComplete displays matches for the current input path.
// It parses the input, sets currentPath, and updates matches for the next level.
// Returns whether matches were found.
func (db *DirBrowser) TabComplete(inputText string) bool {
	inputText = strings.TrimSpace(inputText)
	if inputText == "" {
		// Empty input - show matches in home
		db.currentPath = ""
		db.UpdateMatches()
		return db.showMatches
	}

	// Normalize path: remove leading slash, trailing slash
	inputText = strings.TrimPrefix(inputText, "/")
	inputText = strings.TrimSuffix(inputText, "/")

	// Validate the input path exists
	fullPath := filepath.Join(db.homeDir, inputText)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return false
	}

	relPath, err := filepath.Rel(db.homeDir, absPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return false // Can't escape homeDir
	}

	// Check if path exists and is a directory
	info, err := os.Stat(absPath)
	if err != nil || !info.IsDir() {
		return false
	}

	// Set current path and update matches
	db.currentPath = inputText
	db.UpdateMatches()
	return db.showMatches
}

// SelectMatch returns the next path segment by completing the currently selected match.
// The input text is updated with the selected match, ready for further drilling.
func (db *DirBrowser) SelectMatch() string {
	if !db.showMatches || db.selectedIndex >= len(db.matches) {
		return db.CurrentPath()
	}

	selected := db.matches[db.selectedIndex]

	if db.currentPath != "" {
		db.currentPath = db.currentPath + "/" + selected
	} else {
		db.currentPath = selected
	}

	// Refresh matches for next level
	db.UpdateMatches()
	return "/" + db.currentPath + "/"
}

// MoveSelectionDown moves to the next match (j in vim).
func (db *DirBrowser) MoveSelectionDown() {
	if db.selectedIndex < len(db.matches)-1 {
		db.selectedIndex++
		db.ensureSelectionVisible()
	}
}

// MoveSelectionUp moves to the previous match (k in vim).
func (db *DirBrowser) MoveSelectionUp() {
	if db.selectedIndex > 0 {
		db.selectedIndex--
		db.ensureSelectionVisible()
	}
}

// Backspace removes the last path component and regenerates matches.
func (db *DirBrowser) Backspace() string {
	parts := strings.Split(db.currentPath, "/")
	if len(parts) > 1 {
		parts = parts[:len(parts)-1]
		db.currentPath = strings.Join(parts, "/")
	} else if len(parts) == 1 && parts[0] != "" {
		db.currentPath = ""
	}

	db.UpdateMatches()
	return db.CurrentPath()
}

// CurrentPath returns the full absolute path for display.
func (db *DirBrowser) CurrentPath() string {
	if db.currentPath == "" {
		return "/"
	}
	return "/" + db.currentPath + "/"
}

// FullAbsPath returns the absolute path on the filesystem.
func (db *DirBrowser) FullAbsPath() string {
	return filepath.Join(db.homeDir, db.currentPath)
}

// GetMatches returns the current list of matching subdirectories.
func (db *DirBrowser) GetCurrentMatches() []string {
	return db.matches
}

// GetSelectedMatch returns the currently selected match name.
func (db *DirBrowser) GetSelectedMatch() string {
	if db.selectedIndex < len(db.matches) {
		return db.matches[db.selectedIndex]
	}
	return ""
}

// SelectedIndex returns the current selection index.
func (db *DirBrowser) SelectedIndex() int {
	return db.selectedIndex
}

// ScrollOffset returns the current scroll offset.
func (db *DirBrowser) ScrollOffset() int {
	return db.scrollOffset
}

// ShowMatches returns whether the matches popup should be visible.
func (db *DirBrowser) ShowMatches() bool {
	return db.showMatches
}

// Validate checks if the current path is a valid git repository.
func (db *DirBrowser) Validate() error {
	fullPath := db.FullAbsPath()

	// Check directory exists
	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	// Check if it's a git repository using git command
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = fullPath
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("not a git repository")
	}

	return nil
}

// Reset clears the browser state.
func (db *DirBrowser) Reset() {
	db.currentPath = ""
	db.matches = []string{}
	db.selectedIndex = 0
	db.scrollOffset = 0
	db.showMatches = false
}
