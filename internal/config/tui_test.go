package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTUIConfigStruct(t *testing.T) {
	tui := &TUIConfig{
		SidebarWidth: 40,
		Tabs: []Tab{
			{Name: "Work", Filter: "+work"},
			{Name: "Home", Filter: "+home"},
		},
		Columns: Columns{
			{Name: "id", Label: "ID"},
			{Name: "project", Label: "PROJECT"},
			{Name: "description", Label: "DESCRIPTION"},
			{Name: "due", Label: "DUE"},
		},
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
	if len(tui.Tabs) != 2 {
		t.Errorf("Expected 2 tabs, got %d", len(tui.Tabs))
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

func TestTabStruct(t *testing.T) {
	tab := Tab{
		Name:   "Important Tasks",
		Filter: "priority:H status:pending",
	}

	if tab.Name != "Important Tasks" {
		t.Errorf("Expected name 'Important Tasks', got %s", tab.Name)
	}
	if tab.Filter != "priority:H status:pending" {
		t.Errorf("Expected filter 'priority:H status:pending', got %s", tab.Filter)
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
		Name:           "custom",
		PriorityHigh:   "red",
		PriorityMedium: "yellow",
		PriorityLow:    "blue",
		DueOverdue:     "red",
		DueToday:       "orange",
		DueSoon:        "yellow",
		SelectionBg:    "reverse",
		SidebarBorder:  "gray",
	}

	if theme.PriorityHigh != "red" {
		t.Errorf("Expected PriorityHigh 'red', got %s", theme.PriorityHigh)
	}
	if theme.DueOverdue != "red" {
		t.Errorf("Expected DueOverdue 'red', got %s", theme.DueOverdue)
	}
	if theme.SelectionBg != "reverse" {
		t.Errorf("Expected SelectionBg 'reverse', got %s", theme.SelectionBg)
	}
}

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()

	if theme == nil {
		t.Fatal("Expected theme, got nil")
	}

	// Check theme name
	if theme.Name != "dark" {
		t.Errorf("Expected Name 'dark', got %s", theme.Name)
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
	if theme.DueOverdue == "" {
		t.Error("Expected DueOverdue to be set")
	}
	if theme.DueToday == "" {
		t.Error("Expected DueToday to be set")
	}
	if theme.DueSoon == "" {
		t.Error("Expected DueSoon to be set")
	}
	if theme.SelectionBg == "" {
		t.Error("Expected SelectionBg to be set")
	}
	if theme.SidebarBorder == "" {
		t.Error("Expected SidebarBorder to be set")
	}
	if theme.HeaderFg == "" {
		t.Error("Expected HeaderFg to be set")
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
			if c.Name == col {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected column '%s' in default columns", col)
		}
	}

	// Check that all columns have labels
	for _, c := range cols {
		if c.Label == "" {
			t.Errorf("Expected column '%s' to have a label", c.Name)
		}
	}
}

func TestColumnsUnmarshal_OldFormat(t *testing.T) {
	// Test backward compatibility with old string array format
	yamlData := `
columns:
  - id
  - project
  - priority
  - due
  - description
`
	var cfg struct {
		Columns Columns `yaml:"columns"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	if err != nil {
		t.Fatalf("Failed to unmarshal old format: %v", err)
	}

	if len(cfg.Columns) != 5 {
		t.Errorf("Expected 5 columns, got %d", len(cfg.Columns))
	}

	// Check that columns were converted with default labels
	expectedColumns := map[string]string{
		"id":          "ID",
		"project":     "PROJECT",
		"priority":    "P",
		"due":         "DUE",
		"description": "DESCRIPTION",
	}

	for _, col := range cfg.Columns {
		expectedLabel, exists := expectedColumns[col.Name]
		if !exists {
			t.Errorf("Unexpected column name: %s", col.Name)
			continue
		}
		if col.Label != expectedLabel {
			t.Errorf("Expected label '%s' for column '%s', got '%s'", expectedLabel, col.Name, col.Label)
		}
	}
}

func TestColumnsUnmarshal_NewFormat(t *testing.T) {
	// Test new object format with custom labels
	yamlData := `
columns:
  - name: id
    label: "#"
  - name: project
    label: "Proj"
  - name: priority
    label: "!"
  - name: due
    label: "Due Date"
  - name: description
    label: "Task"
`
	var cfg struct {
		Columns Columns `yaml:"columns"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	if err != nil {
		t.Fatalf("Failed to unmarshal new format: %v", err)
	}

	if len(cfg.Columns) != 5 {
		t.Errorf("Expected 5 columns, got %d", len(cfg.Columns))
	}

	// Check that columns have custom labels
	expectedColumns := map[string]string{
		"id":          "#",
		"project":     "Proj",
		"priority":    "!",
		"due":         "Due Date",
		"description": "Task",
	}

	for _, col := range cfg.Columns {
		expectedLabel, exists := expectedColumns[col.Name]
		if !exists {
			t.Errorf("Unexpected column name: %s", col.Name)
			continue
		}
		if col.Label != expectedLabel {
			t.Errorf("Expected label '%s' for column '%s', got '%s'", expectedLabel, col.Name, col.Label)
		}
	}
}

func TestColumnsUnmarshal_WithLength(t *testing.T) {
	// Test new format with custom lengths
	yamlData := `
columns:
  - name: id
    label: "#"
    length: 5
  - name: project
    label: "Project"
    length: 20
  - name: description
    label: "Task"
    length: 40
  - name: due
    label: "Due"
`
	var cfg struct {
		Columns Columns `yaml:"columns"`
	}

	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	if err != nil {
		t.Fatalf("Failed to unmarshal columns with length: %v", err)
	}

	if len(cfg.Columns) != 4 {
		t.Errorf("Expected 4 columns, got %d", len(cfg.Columns))
	}

	// Check custom lengths
	expectedLengths := map[string]int{
		"id":          5,
		"project":     20,
		"description": 40,
		"due":         0, // No length specified
	}

	for _, col := range cfg.Columns {
		expectedLength, exists := expectedLengths[col.Name]
		if !exists {
			t.Errorf("Unexpected column name: %s", col.Name)
			continue
		}
		if col.Length != expectedLength {
			t.Errorf("Expected length %d for column '%s', got %d", expectedLength, col.Name, col.Length)
		}
	}
}
