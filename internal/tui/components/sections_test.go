package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clobrano/wui/internal/core"
)

// defaultSectionsStyles returns default styles for testing
func defaultSectionsStyles() SectionsStyles {
	return SectionsStyles{
		Active:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("63")).Padding(0, 1),
		Inactive: lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Padding(0, 1),
		Count:    lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Padding(0, 1),
	}
}

// Test constructor (10.1-10.2)
func TestNewSections(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())

	if s.ActiveIndex != 0 {
		t.Errorf("Expected ActiveIndex to be 0, got %d", s.ActiveIndex)
	}
	if len(s.Items) != len(sections) {
		t.Errorf("Expected %d sections, got %d", len(sections), len(s.Items))
	}
	if s.Width != 100 {
		t.Errorf("Expected width to be 100, got %d", s.Width)
	}
}

func TestNewSectionsEmpty(t *testing.T) {
	s := NewSections([]core.Section{}, 80, defaultSectionsStyles())

	if s.ActiveIndex != 0 {
		t.Errorf("Expected ActiveIndex to be 0, got %d", s.ActiveIndex)
	}
	if len(s.Items) != 0 {
		t.Errorf("Expected 0 sections, got %d", len(s.Items))
	}
}

// Test section switching with Tab (10.3)
func TestSectionsUpdateTab(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())

	// Press Tab to move to next section
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updated, cmd := s.Update(msg)

	if updated.ActiveIndex != 1 {
		t.Errorf("Expected ActiveIndex to be 1 after Tab, got %d", updated.ActiveIndex)
	}
	if cmd == nil {
		t.Error("Expected command to be returned for section change")
	}

	// Verify it returns a SectionChangedMsg
	if cmd != nil {
		result := cmd()
		if msg, ok := result.(SectionChangedMsg); !ok {
			t.Errorf("Expected SectionChangedMsg, got %T", result)
		} else {
			if msg.Section.Name != "Waiting" {
				t.Errorf("Expected section 'Waiting', got %s", msg.Section.Name)
			}
		}
	}
}

func TestSectionsUpdateTabWrapAround(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())
	s.ActiveIndex = len(sections) - 1 // Last section

	// Press Tab should wrap to first section
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ := s.Update(msg)

	if updated.ActiveIndex != 0 {
		t.Errorf("Expected ActiveIndex to wrap to 0, got %d", updated.ActiveIndex)
	}
}

// Test section switching with Shift+Tab (10.3)
func TestSectionsUpdateShiftTab(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())
	s.ActiveIndex = 2 // Start at third section

	// Press Shift+Tab to move to previous section
	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	updated, cmd := s.Update(msg)

	if updated.ActiveIndex != 1 {
		t.Errorf("Expected ActiveIndex to be 1 after Shift+Tab, got %d", updated.ActiveIndex)
	}
	if cmd == nil {
		t.Error("Expected command to be returned for section change")
	}
}

func TestSectionsUpdateShiftTabWrapAround(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())
	s.ActiveIndex = 0 // First section

	// Press Shift+Tab should wrap to last section
	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	updated, _ := s.Update(msg)

	if updated.ActiveIndex != len(sections)-1 {
		t.Errorf("Expected ActiveIndex to wrap to %d, got %d", len(sections)-1, updated.ActiveIndex)
	}
}

// Test number key navigation (1-9)
func TestSectionsUpdateNumberKey(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())

	tests := []struct {
		key           rune
		expectedIndex int
	}{
		{'1', 0},
		{'2', 1},
		{'3', 2},
		{'4', 3},
		{'5', 4},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}}
			updated, cmd := s.Update(msg)

			if updated.ActiveIndex != tt.expectedIndex {
				t.Errorf("Expected ActiveIndex to be %d for key %c, got %d", tt.expectedIndex, tt.key, updated.ActiveIndex)
			}
			if cmd == nil {
				t.Error("Expected command to be returned for section change")
			}
		})
	}
}

func TestSectionsUpdateNumberKeyOutOfRange(t *testing.T) {
	sections := core.DefaultSections() // 5 sections
	s := NewSections(sections, 100, defaultSectionsStyles())

	// Press '9' when only 5 sections exist
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'9'}}
	updated, cmd := s.Update(msg)

	// Should not change section
	if updated.ActiveIndex != 0 {
		t.Errorf("Expected ActiveIndex to remain 0, got %d", updated.ActiveIndex)
	}
	if cmd != nil {
		t.Error("Expected no command for out of range number")
	}
}

// Test view rendering (10.4, 10.6)
func TestSectionsView(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())

	view := s.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Should contain all section names
	for _, section := range sections {
		if !strings.Contains(view, section.Name) {
			t.Errorf("Expected view to contain section name %s", section.Name)
		}
	}
}

func TestSectionsViewHighlightActive(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())
	s.ActiveIndex = 1 // Waiting section

	view := s.View()

	// The active section should be visually distinct
	// We check that it contains the section name
	if !strings.Contains(view, "Waiting") {
		t.Error("Expected view to contain 'Waiting'")
	}
}

func TestSectionsViewEmpty(t *testing.T) {
	s := NewSections([]core.Section{}, 100, defaultSectionsStyles())

	view := s.View()

	// Empty sections should still return a string (maybe just empty or placeholder)
	if view == "" {
		// This is acceptable
	}
}

// Test custom sections (10.9)
func TestSectionsWithBookmarks(t *testing.T) {
	bookmarks := []core.Bookmark{
		{Name: "Work", Filter: "+work status:pending"},
		{Name: "Home", Filter: "+home status:pending"},
	}
	sections := core.SectionsWithBookmarks(bookmarks)
	s := NewSections(sections, 100, defaultSectionsStyles())

	// Should have default sections + bookmarks
	expectedCount := len(core.DefaultSections()) + len(bookmarks)
	if len(s.Items) != expectedCount {
		t.Errorf("Expected %d sections (default + bookmarks), got %d", expectedCount, len(s.Items))
	}

	// Check that bookmark sections are present
	view := s.View()
	if !strings.Contains(view, "Work") {
		t.Error("Expected view to contain bookmark 'Work'")
	}
	if !strings.Contains(view, "Home") {
		t.Error("Expected view to contain bookmark 'Home'")
	}
}

// Test active section getter
func TestGetActiveSection(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())
	s.ActiveIndex = 2

	activeSection := s.GetActiveSection()

	if activeSection.Name != sections[2].Name {
		t.Errorf("Expected active section to be %s, got %s", sections[2].Name, activeSection.Name)
	}
	if activeSection.Filter != sections[2].Filter {
		t.Errorf("Expected filter %s, got %s", sections[2].Filter, activeSection.Filter)
	}
}

func TestGetActiveSectionEmpty(t *testing.T) {
	s := NewSections([]core.Section{}, 100, defaultSectionsStyles())

	activeSection := s.GetActiveSection()

	// Should return empty section
	if activeSection.Name != "" {
		t.Errorf("Expected empty section name, got %s", activeSection.Name)
	}
}

// Test task count integration (10.10)
func TestSetTaskCount(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())

	s.SetTaskCount(42)

	if s.TaskCount != 42 {
		t.Errorf("Expected task count to be 42, got %d", s.TaskCount)
	}

	view := s.View()
	// View should contain the task count when set
	if !strings.Contains(view, "42") {
		t.Error("Expected view to contain task count '42'")
	}
}

func TestSetTaskCountZero(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())

	s.SetTaskCount(0)

	if s.TaskCount != 0 {
		t.Errorf("Expected task count to be 0, got %d", s.TaskCount)
	}
}

// Test window size updates
func TestSectionsUpdateWindowSize(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, cmd := s.Update(msg)

	if updated.Width != 120 {
		t.Errorf("Expected width to be updated to 120, got %d", updated.Width)
	}
	if cmd != nil {
		t.Error("Expected no command for window size update")
	}
}

// Test SetSize method
func TestSectionsSetSize(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())

	s.SetSize(150)

	if s.Width != 150 {
		t.Errorf("Expected width to be 150, got %d", s.Width)
	}
}

// Test SectionChangedMsg type
func TestSectionChangedMsg(t *testing.T) {
	section := core.Section{
		Name:        "Test",
		Filter:      "status:pending",
		Description: "Test section",
	}

	msg := SectionChangedMsg{Section: section}

	if msg.Section.Name != "Test" {
		t.Errorf("Expected section name 'Test', got %s", msg.Section.Name)
	}
}

// Test that unchanged keys return model unchanged
func TestSectionsUpdateOtherKeys(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())
	initialIndex := s.ActiveIndex

	// Press an unrelated key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updated, cmd := s.Update(msg)

	if updated.ActiveIndex != initialIndex {
		t.Errorf("Expected ActiveIndex to remain %d, got %d", initialIndex, updated.ActiveIndex)
	}
	if cmd != nil {
		t.Error("Expected no command for unrelated key")
	}
}

// Test Projects section flag (10.7)
func TestIsProjectsSection(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())

	// Find the Projects section index
	projectsIndex := -1
	for i, section := range sections {
		if section.Name == "Projects" {
			projectsIndex = i
			break
		}
	}

	if projectsIndex == -1 {
		t.Fatal("Projects section not found in default sections")
	}

	s.ActiveIndex = projectsIndex

	if !s.IsProjectsView() {
		t.Error("Expected IsProjectsView to return true for Projects section")
	}

	// Test other sections
	s.ActiveIndex = 0
	if s.IsProjectsView() {
		t.Error("Expected IsProjectsView to return false for non-Projects section")
	}
}

// Test Tags section flag (10.8)
func TestIsTagsSection(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())

	// Find the Tags section index
	tagsIndex := -1
	for i, section := range sections {
		if section.Name == "Tags" {
			tagsIndex = i
			break
		}
	}

	if tagsIndex == -1 {
		t.Fatal("Tags section not found in default sections")
	}

	s.ActiveIndex = tagsIndex

	if !s.IsTagsView() {
		t.Error("Expected IsTagsView to return true for Tags section")
	}

	// Test other sections
	s.ActiveIndex = 0
	if s.IsTagsView() {
		t.Error("Expected IsTagsView to return false for non-Tags section")
	}
}

// Test vim-like h/l navigation
func TestSectionsUpdateVimKeysL(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())

	// Press 'l' to move to next section (vim-like)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	updated, cmd := s.Update(msg)

	if updated.ActiveIndex != 1 {
		t.Errorf("Expected ActiveIndex to be 1 after 'l', got %d", updated.ActiveIndex)
	}
	if cmd == nil {
		t.Error("Expected command to be returned for section change")
	}
}

func TestSectionsUpdateVimKeysH(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())
	s.ActiveIndex = 2 // Start at third section

	// Press 'h' to move to previous section (vim-like)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	updated, cmd := s.Update(msg)

	if updated.ActiveIndex != 1 {
		t.Errorf("Expected ActiveIndex to be 1 after 'h', got %d", updated.ActiveIndex)
	}
	if cmd == nil {
		t.Error("Expected command to be returned for section change")
	}
}

func TestSectionsUpdateVimKeysWrapAround(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())

	// Test 'h' wrap around from first to last
	s.ActiveIndex = 0
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	updated, _ := s.Update(msg)

	if updated.ActiveIndex != len(sections)-1 {
		t.Errorf("Expected ActiveIndex to wrap to %d, got %d", len(sections)-1, updated.ActiveIndex)
	}

	// Test 'l' wrap around from last to first
	s.ActiveIndex = len(sections) - 1
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	updated, _ = s.Update(msg)

	if updated.ActiveIndex != 0 {
		t.Errorf("Expected ActiveIndex to wrap to 0, got %d", updated.ActiveIndex)
	}
}

// Test arrow key navigation
func TestSectionsUpdateArrowKeys(t *testing.T) {
	sections := core.DefaultSections()
	s := NewSections(sections, 100, defaultSectionsStyles())

	// Test right arrow (next section)
	msg := tea.KeyMsg{Type: tea.KeyRight}
	updated, cmd := s.Update(msg)

	if updated.ActiveIndex != 1 {
		t.Errorf("Expected ActiveIndex to be 1 after right arrow, got %d", updated.ActiveIndex)
	}
	if cmd == nil {
		t.Error("Expected command to be returned for section change")
	}

	// Test left arrow (previous section)
	msg = tea.KeyMsg{Type: tea.KeyLeft}
	updated, cmd = updated.Update(msg)

	if updated.ActiveIndex != 0 {
		t.Errorf("Expected ActiveIndex to be 0 after left arrow, got %d", updated.ActiveIndex)
	}
	if cmd == nil {
		t.Error("Expected command to be returned for section change")
	}
}
