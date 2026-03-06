package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDirBrowser_NewDirBrowser(t *testing.T) {
	homeDir := os.TempDir()
	db := NewDirBrowser(homeDir)

	if db.homeDir != homeDir {
		t.Errorf("expected homeDir %s, got %s", homeDir, db.homeDir)
	}
	if db.currentPath != "" {
		t.Errorf("expected empty currentPath, got %s", db.currentPath)
	}
}

func TestDirBrowser_GetMatches(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	subDir1 := filepath.Join(tmpDir, "alpha")
	subDir2 := filepath.Join(tmpDir, "beta")
	subDir3 := filepath.Join(tmpDir, "gamma")

	os.Mkdir(subDir1, 0755)
	os.Mkdir(subDir2, 0755)
	os.Mkdir(subDir3, 0755)

	db := NewDirBrowser(tmpDir)
	matches := db.GetMatches("")

	if len(matches) != 3 {
		t.Errorf("expected 3 matches, got %d: %v", len(matches), matches)
	}

	// Check that matches are sorted
	if matches[0] != "alpha" || matches[1] != "beta" || matches[2] != "gamma" {
		t.Errorf("expected [alpha beta gamma], got %v", matches)
	}
}

func TestDirBrowser_GetMatches_PartialPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	subDir1 := filepath.Join(tmpDir, "app-backend")
	subDir2 := filepath.Join(tmpDir, "app-frontend")
	subDir3 := filepath.Join(tmpDir, "config")

	os.Mkdir(subDir1, 0755)
	os.Mkdir(subDir2, 0755)
	os.Mkdir(subDir3, 0755)

	db := NewDirBrowser(tmpDir)
	matches := db.GetMatches("app")

	if len(matches) != 2 {
		t.Errorf("expected 2 matches for 'app' prefix, got %d: %v", len(matches), matches)
	}
}

func TestDirBrowser_MoveSelection(t *testing.T) {
	tmpDir := t.TempDir()
	os.Mkdir(filepath.Join(tmpDir, "dir1"), 0755)
	os.Mkdir(filepath.Join(tmpDir, "dir2"), 0755)
	os.Mkdir(filepath.Join(tmpDir, "dir3"), 0755)

	db := NewDirBrowser(tmpDir)
	db.UpdateMatches()

	// Check initial selection
	if db.SelectedIndex() != 0 {
		t.Errorf("expected initial index 0, got %d", db.SelectedIndex())
	}

	// Move down
	db.MoveSelectionDown()
	if db.SelectedIndex() != 1 {
		t.Errorf("expected index 1, got %d", db.SelectedIndex())
	}

	// Move down again
	db.MoveSelectionDown()
	if db.SelectedIndex() != 2 {
		t.Errorf("expected index 2, got %d", db.SelectedIndex())
	}

	// Move down at end (should stay at 2)
	db.MoveSelectionDown()
	if db.SelectedIndex() != 2 {
		t.Errorf("expected index 2 (at end), got %d", db.SelectedIndex())
	}

	// Move up
	db.MoveSelectionUp()
	if db.SelectedIndex() != 1 {
		t.Errorf("expected index 1, got %d", db.SelectedIndex())
	}
}

func TestDirBrowser_CurrentPath(t *testing.T) {
	db := NewDirBrowser(os.TempDir())

	if db.CurrentPath() != "/" {
		t.Errorf("expected empty path to return '/', got %s", db.CurrentPath())
	}

	db.currentPath = "code"
	if db.CurrentPath() != "/code/" {
		t.Errorf("expected '/code/', got %s", db.CurrentPath())
	}

	db.currentPath = "code/projects"
	if db.CurrentPath() != "/code/projects/" {
		t.Errorf("expected '/code/projects/', got %s", db.CurrentPath())
	}
}

func TestDirBrowser_Backspace(t *testing.T) {
	db := NewDirBrowser(os.TempDir())

	db.currentPath = "code/projects/myapp"
	db.Backspace()

	if db.currentPath != "code/projects" {
		t.Errorf("expected 'code/projects', got %s", db.currentPath)
	}

	db.Backspace()
	if db.currentPath != "code" {
		t.Errorf("expected 'code', got %s", db.currentPath)
	}

	db.Backspace()
	if db.currentPath != "" {
		t.Errorf("expected empty path, got %s", db.currentPath)
	}
}

func TestDirBrowser_Reset(t *testing.T) {
	db := NewDirBrowser(os.TempDir())
	db.currentPath = "code"
	db.selectedIndex = 5
	db.showMatches = true
	db.scrollOffset = 10

	db.Reset()

	if db.currentPath != "" {
		t.Errorf("expected empty currentPath, got %s", db.currentPath)
	}
	if db.selectedIndex != 0 {
		t.Errorf("expected selectedIndex 0, got %d", db.selectedIndex)
	}
	if db.showMatches {
		t.Errorf("expected showMatches false")
	}
	if db.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0, got %d", db.scrollOffset)
	}
}

func TestDirBrowser_Scrolling(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 25 directories
	for i := 0; i < 25; i++ {
		name := fmt.Sprintf("dir%02d", i)
		os.Mkdir(filepath.Join(tmpDir, name), 0755)
	}

	db := NewDirBrowser(tmpDir)
	db.SetMaxVisibleRows(10)
	db.UpdateMatches()

	if len(db.matches) != 25 {
		t.Fatalf("expected 25 matches, got %d", len(db.matches))
	}

	// Check initial visible matches
	visible := db.GetVisibleMatches()
	if len(visible) != 10 {
		t.Errorf("expected 10 visible matches, got %d", len(visible))
	}

	// Check scroll info
	info := db.GetScrollInfo()
	if info != "1-10/25" {
		t.Errorf("expected scroll info '1-10/25', got %s", info)
	}

	// Move down several times
	for i := 0; i < 15; i++ {
		db.MoveSelectionDown()
	}

	// Should have scrolled
	if db.scrollOffset == 0 {
		t.Errorf("expected scroll offset > 0 after moving down 15 times, got 0")
	}

	visible = db.GetVisibleMatches()
	if len(visible) > 10 {
		t.Errorf("expected max 10 visible, got %d", len(visible))
	}
}

func TestDirBrowser_ScrollingVisibility(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 10 directories
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("dir%02d", i)
		os.Mkdir(filepath.Join(tmpDir, name), 0755)
	}

	db := NewDirBrowser(tmpDir)
	db.SetMaxVisibleRows(5)
	db.UpdateMatches()

	// Move to last item
	db.selectedIndex = 9
	db.ensureSelectionVisible()

	// Check that selected item is visible
	visible := db.GetVisibleMatches()
	if len(visible) != 5 {
		t.Errorf("expected 5 visible items, got %d", len(visible))
	}

	// Selected index relative to scroll offset should be in visible range
	relativeIndex := db.selectedIndex - db.scrollOffset
	if relativeIndex < 0 || relativeIndex >= len(visible) {
		t.Errorf("selected item at index %d not visible in range [%d,%d]",
			db.selectedIndex, db.scrollOffset, db.scrollOffset+len(visible))
	}
}
