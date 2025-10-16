package config

import (
	"testing"
)

func TestTUIConfigStruct(t *testing.T) {
	tui := &TUIConfig{
		SidebarWidth: 40,
		Bookmarks: []Bookmark{
			{Name: "Work", Filter: "+work"},
			{Name: "Home", Filter: "+home"},
		},
		Columns: []string{"id", "project", "description", "due"},
		Keybindings: map[string]string{
			"quit":   "q",
			"help":   "?",
			"done":   "d",
			"delete": "x",
		},
		Theme: &Theme{
			PriorityHigh:   "red",
			PriorityMedium: "yellow",
			PriorityLow:    "blue",
		},
	}

	if tui.SidebarWidth != 40 {
		t.Errorf("Expected SidebarWidth 40, got %d", tui.SidebarWidth)
	}
	if len(tui.Bookmarks) != 2 {
		t.Errorf("Expected 2 bookmarks, got %d", len(tui.Bookmarks))
	}
	if len(tui.Columns) != 4 {
		t.Errorf("Expected 4 columns, got %d", len(tui.Columns))
	}
	if len(tui.Keybindings) != 4 {
		t.Errorf("Expected 4 keybindings, got %d", len(tui.Keybindings))
	}
	if tui.Theme == nil {
		t.Error("Expected theme to be set")
	}
}

func TestBookmarkStruct(t *testing.T) {
	bookmark := Bookmark{
		Name:   "Important Tasks",
		Filter: "priority:H status:pending",
	}

	if bookmark.Name != "Important Tasks" {
		t.Errorf("Expected name 'Important Tasks', got %s", bookmark.Name)
	}
	if bookmark.Filter != "priority:H status:pending" {
		t.Errorf("Expected filter 'priority:H status:pending', got %s", bookmark.Filter)
	}
}

func TestDefaultTUIConfig(t *testing.T) {
	tui := DefaultTUIConfig()

	if tui == nil {
		t.Fatal("Expected TUI config, got nil")
	}

	if tui.SidebarWidth == 0 {
		t.Error("Expected default SidebarWidth to be set")
	}

	if len(tui.Columns) == 0 {
		t.Error("Expected default columns to be set")
	}

	if tui.Keybindings == nil || len(tui.Keybindings) == 0 {
		t.Error("Expected default keybindings to be set")
	}

	if tui.Theme == nil {
		t.Error("Expected default theme to be set")
	}
}

func TestThemeStruct(t *testing.T) {
	theme := &Theme{
		PriorityHigh:   "red",
		PriorityMedium: "yellow",
		PriorityLow:    "blue",
		Overdue:        "red",
		DueToday:       "orange",
		DueSoon:        "yellow",
		Selected:       "reverse",
		Border:         "gray",
	}

	if theme.PriorityHigh != "red" {
		t.Errorf("Expected PriorityHigh 'red', got %s", theme.PriorityHigh)
	}
	if theme.Overdue != "red" {
		t.Errorf("Expected Overdue 'red', got %s", theme.Overdue)
	}
	if theme.Selected != "reverse" {
		t.Errorf("Expected Selected 'reverse', got %s", theme.Selected)
	}
}

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()

	if theme == nil {
		t.Fatal("Expected theme, got nil")
	}

	// Check all color fields are set
	if theme.PriorityHigh == "" {
		t.Error("Expected PriorityHigh to be set")
	}
	if theme.PriorityMedium == "" {
		t.Error("Expected PriorityMedium to be set")
	}
	if theme.PriorityLow == "" {
		t.Error("Expected PriorityLow to be set")
	}
	if theme.Overdue == "" {
		t.Error("Expected Overdue to be set")
	}
	if theme.DueToday == "" {
		t.Error("Expected DueToday to be set")
	}
	if theme.DueSoon == "" {
		t.Error("Expected DueSoon to be set")
	}
	if theme.Selected == "" {
		t.Error("Expected Selected to be set")
	}
	if theme.Border == "" {
		t.Error("Expected Border to be set")
	}
}

func TestDefaultKeybindings(t *testing.T) {
	kb := DefaultKeybindings()

	if kb == nil || len(kb) == 0 {
		t.Fatal("Expected keybindings map, got nil or empty")
	}

	// Check essential keybindings exist
	essentialKeys := []string{"quit", "help", "done", "delete", "edit", "modify", "annotate", "new", "undo", "refresh"}
	for _, key := range essentialKeys {
		if _, exists := kb[key]; !exists {
			t.Errorf("Expected keybinding for '%s' to exist", key)
		}
	}
}

func TestDefaultColumns(t *testing.T) {
	cols := DefaultColumns()

	if cols == nil || len(cols) == 0 {
		t.Fatal("Expected columns, got nil or empty")
	}

	// Check essential columns exist
	essentialCols := []string{"id", "description"}
	for _, col := range essentialCols {
		found := false
		for _, c := range cols {
			if c == col {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected column '%s' in default columns", col)
		}
	}
}
