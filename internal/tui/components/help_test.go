package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewHelp(t *testing.T) {
	styles := DefaultHelpStyles()
	help := NewHelp(80, 24, styles)

	if help.width != 80 {
		t.Errorf("Expected width 80, got %d", help.width)
	}
	if help.height != 24 {
		t.Errorf("Expected height 24, got %d", help.height)
	}
	if len(help.groups) == 0 {
		t.Error("Expected default keybinding groups to be set")
	}
}

func TestDefaultKeybindingGroups(t *testing.T) {
	groups := defaultKeybindingGroups()

	expectedGroups := []string{
		"Task Navigation",
		"Section Navigation",
		"Task Actions",
		"View Controls",
		"Sidebar Scrolling (when sidebar is open)",
		"Other",
	}

	if len(groups) != len(expectedGroups) {
		t.Errorf("Expected %d groups, got %d", len(expectedGroups), len(groups))
	}

	for i, expected := range expectedGroups {
		if i >= len(groups) {
			break
		}
		if groups[i].Title != expected {
			t.Errorf("Expected group %d to be '%s', got '%s'", i, expected, groups[i].Title)
		}
	}
}

func TestHelp_SetKeybindings(t *testing.T) {
	styles := DefaultHelpStyles()
	help := NewHelp(80, 24, styles)

	customGroups := []KeybindingGroup{
		{
			Title: "Custom Group",
			Bindings: []Keybinding{
				{Keys: []string{"x"}, Description: "Custom action"},
			},
		},
	}

	help.SetKeybindings(customGroups)

	if len(help.groups) != 1 {
		t.Errorf("Expected 1 group after SetKeybindings, got %d", len(help.groups))
	}
	if help.groups[0].Title != "Custom Group" {
		t.Errorf("Expected group title 'Custom Group', got '%s'", help.groups[0].Title)
	}
}

func TestHelp_SetSize(t *testing.T) {
	styles := DefaultHelpStyles()
	help := NewHelp(80, 24, styles)

	help.SetSize(100, 30)

	if help.width != 100 {
		t.Errorf("Expected width 100, got %d", help.width)
	}
	if help.height != 30 {
		t.Errorf("Expected height 30, got %d", help.height)
	}
	if help.viewport.Width != 96 { // 100 - 4 for padding
		t.Errorf("Expected viewport width 96, got %d", help.viewport.Width)
	}
	if help.viewport.Height != 26 { // 30 - 4 for padding
		t.Errorf("Expected viewport height 26, got %d", help.viewport.Height)
	}
}

func TestHelp_Update_KeyNavigation(t *testing.T) {
	styles := DefaultHelpStyles()
	help := NewHelp(80, 24, styles)

	tests := []struct {
		name string
		key  string
	}{
		{"Down", "j"},
		{"Up", "k"},
		{"HalfPageDown", "ctrl+d"},
		{"HalfPageUp", "ctrl+u"},
		{"PageDown", "ctrl+f"},
		{"PageUp", "ctrl+b"},
		{"Top", "g"},
		{"Bottom", "G"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			_, cmd := help.Update(msg)
			// Just verify it doesn't panic
			_ = cmd
		})
	}
}

func TestHelp_Update_WindowSize(t *testing.T) {
	styles := DefaultHelpStyles()
	help := NewHelp(80, 24, styles)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedHelp, _ := help.Update(msg)

	if updatedHelp.width != 120 {
		t.Errorf("Expected width 120, got %d", updatedHelp.width)
	}
	if updatedHelp.height != 40 {
		t.Errorf("Expected height 40, got %d", updatedHelp.height)
	}
}

func TestHelp_View(t *testing.T) {
	styles := DefaultHelpStyles()
	help := NewHelp(80, 24, styles)

	view := help.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Check for some expected content
	if !strings.Contains(view, "Keyboard Shortcuts") {
		t.Error("Expected view to contain 'Keyboard Shortcuts' title")
	}
}

func TestHelp_View_ZeroSize(t *testing.T) {
	styles := DefaultHelpStyles()
	help := NewHelp(0, 0, styles)

	view := help.View()

	if view != "" {
		t.Error("Expected empty view for zero size")
	}
}

func TestRenderHelpContent(t *testing.T) {
	styles := DefaultHelpStyles()
	groups := []KeybindingGroup{
		{
			Title: "Test Group",
			Bindings: []Keybinding{
				{Keys: []string{"a", "b"}, Description: "Test action"},
			},
		},
	}

	content := renderHelpContent(groups, styles)

	if !strings.Contains(content, "Test Group") {
		t.Error("Expected content to contain group title")
	}
	if !strings.Contains(content, "a/b") {
		t.Error("Expected content to contain joined keys")
	}
	if !strings.Contains(content, "Test action") {
		t.Error("Expected content to contain binding description")
	}
}

func TestKeybinding_MultipleKeys(t *testing.T) {
	styles := DefaultHelpStyles()
	groups := []KeybindingGroup{
		{
			Title: "Navigation",
			Bindings: []Keybinding{
				{Keys: []string{"j", "↓", "down"}, Description: "Move down"},
			},
		},
	}

	content := renderHelpContent(groups, styles)

	if !strings.Contains(content, "j/↓/down") {
		t.Error("Expected content to contain all alternative keys joined with /")
	}
}

func TestDefaultHelpStyles(t *testing.T) {
	styles := DefaultHelpStyles()

	// Just verify all styles are initialized (non-nil check via rendering)
	if styles.Title.Render("test") == "" {
		t.Error("Expected Title style to be initialized")
	}
	if styles.GroupTitle.Render("test") == "" {
		t.Error("Expected GroupTitle style to be initialized")
	}
	if styles.Key.Render("test") == "" {
		t.Error("Expected Key style to be initialized")
	}
	if styles.Description.Render("test") == "" {
		t.Error("Expected Description style to be initialized")
	}
}
