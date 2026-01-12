package components

import (
	"testing"
)

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		name     string
		filter   string
		target   string
		expected bool
	}{
		// Basic matching
		{"exact match", "test", "test", true},
		{"prefix match", "tes", "test", true},
		{"suffix match", "est", "test", true},
		{"substring match", "es", "test", true},

		// Fuzzy matching
		{"fuzzy with gaps", "tt", "test", true},
		{"fuzzy scattered", "tst", "test", true},
		{"fuzzy all chars", "hme", "home-project", true},
		{"fuzzy with separator", "hmp", "home-my-project", true},
		{"fuzzy camelCase", "mypr", "myProject", true},

		// Case insensitive
		{"case insensitive upper", "TEST", "test", true},
		{"case insensitive mixed", "TeSt", "test", true},
		{"case insensitive reverse", "test", "TEST", true},

		// Non-matches
		{"wrong order", "tse", "test", false},
		{"missing char", "testz", "test", false},
		{"completely different", "xyz", "test", false},
		{"wrong char order", "rpoj", "project-home", false},

		// Edge cases
		{"empty filter", "", "test", true},
		{"empty target", "test", "", false},
		{"both empty", "", "", true},
		{"single char match", "t", "test", true},
		{"single char no match", "x", "test", false},

		// Real-world examples
		{"project fuzzy 1", "wui", "work-ui-project", true},
		{"project fuzzy 2", "hp", "home-project", true},
		{"project fuzzy 3", "abc", "a-better-choice", true},
		{"tag fuzzy", "wrk", "work", true},
		{"tag no match", "xyz", "work", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fuzzyMatch(tt.filter, tt.target)
			if result != tt.expected {
				t.Errorf("fuzzyMatch(%q, %q) = %v, expected %v", tt.filter, tt.target, result, tt.expected)
			}
		})
	}
}

func TestNewListPicker(t *testing.T) {
	items := []string{"home", "work", "project"}
	filter := "ho"
	title := "Projects"

	lp := NewListPicker(title, items, filter)

	if lp.title != title {
		t.Errorf("expected title %q, got %q", title, lp.title)
	}

	if len(lp.allItems) != len(items) {
		t.Errorf("expected %d items, got %d", len(items), len(lp.allItems))
	}

	if lp.filter != filter {
		t.Errorf("expected filter %q, got %q", filter, lp.filter)
	}

	// Should have filtered items based on fuzzy match
	if len(lp.filteredItems) == 0 {
		t.Error("expected filtered items to be populated")
	}
}

func TestListPicker_UpdateFilteredItems_FuzzyMatch(t *testing.T) {
	tests := []struct {
		name         string
		allItems     []string
		filter       string
		expectedMatches []string
	}{
		{
			name:         "fuzzy match scattered chars",
			allItems:     []string{"home-project", "work-project", "personal", "test"},
			filter:       "hp",
			expectedMatches: []string{"home-project"},
		},
		{
			name:         "fuzzy match multiple results",
			allItems:     []string{"work", "workout", "network", "framework"},
			filter:       "wrk",
			expectedMatches: []string{"work", "workout", "network", "framework"},
		},
		{
			name:         "empty filter shows all",
			allItems:     []string{"home", "work", "project"},
			filter:       "",
			expectedMatches: []string{"home", "work", "project"},
		},
		{
			name:         "no matches",
			allItems:     []string{"home", "work", "project"},
			filter:       "xyz",
			expectedMatches: []string{},
		},
		{
			name:         "case insensitive fuzzy",
			allItems:     []string{"HomeProject", "workProject", "testProject"},
			filter:       "hp",
			expectedMatches: []string{"HomeProject"},
		},
		{
			name:         "single char fuzzy",
			allItems:     []string{"home", "work", "project"},
			filter:       "h",
			expectedMatches: []string{"home"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lp := NewListPicker("Test", tt.allItems, tt.filter)

			if len(lp.filteredItems) != len(tt.expectedMatches) {
				t.Errorf("expected %d filtered items, got %d", len(tt.expectedMatches), len(lp.filteredItems))
				t.Logf("Expected: %v", tt.expectedMatches)
				t.Logf("Got: %v", lp.filteredItems)
				return
			}

			for i, expected := range tt.expectedMatches {
				if lp.filteredItems[i] != expected {
					t.Errorf("filtered item %d: expected %q, got %q", i, expected, lp.filteredItems[i])
				}
			}
		})
	}
}

func TestListPicker_SelectedItem(t *testing.T) {
	items := []string{"home", "work", "project"}
	lp := NewListPicker("Test", items, "")

	// Initial selection should be first item
	selected := lp.SelectedItem()
	if selected != "home" {
		t.Errorf("expected first item %q, got %q", "home", selected)
	}

	// Empty list should return empty string
	emptyLp := NewListPicker("Test", []string{}, "")
	if emptyLp.SelectedItem() != "" {
		t.Errorf("expected empty string for empty list, got %q", emptyLp.SelectedItem())
	}
}

func TestListPicker_HasItems(t *testing.T) {
	// List with items
	lp := NewListPicker("Test", []string{"home", "work"}, "")
	if !lp.HasItems() {
		t.Error("expected HasItems() to return true for non-empty list")
	}

	// Empty list
	emptyLp := NewListPicker("Test", []string{}, "")
	if emptyLp.HasItems() {
		t.Error("expected HasItems() to return false for empty list")
	}

	// Filtered to no matches
	filteredLp := NewListPicker("Test", []string{"home", "work"}, "xyz")
	if filteredLp.HasItems() {
		t.Error("expected HasItems() to return false when filter matches nothing")
	}
}

func TestListPicker_View(t *testing.T) {
	lp := NewListPicker("Projects", []string{"home", "work", "project"}, "")
	view := lp.View()

	if view == "" {
		t.Error("expected non-empty view output")
	}

	// View should contain title
	// Note: The view contains styled text, so we can't do simple string matching
	if len(view) < 10 {
		t.Error("expected view to contain title, items, and help text")
	}
}
