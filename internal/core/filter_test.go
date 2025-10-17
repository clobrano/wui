package core

import "testing"

func TestFilterStruct(t *testing.T) {
	filter := Filter{
		Text:   "+work -waiting",
		Active: true,
	}

	if filter.Text != "+work -waiting" {
		t.Errorf("Expected Text '+work -waiting', got %s", filter.Text)
	}
	if !filter.Active {
		t.Error("Expected Active to be true")
	}
}

func TestSectionStruct(t *testing.T) {
	section := Section{
		Name:        "Next",
		Filter:      "status:pending -WAITING",
		Description: "Next tasks to work on",
	}

	if section.Name != "Next" {
		t.Errorf("Expected Name 'Next', got %s", section.Name)
	}
	if section.Filter != "status:pending -WAITING" {
		t.Errorf("Expected Filter 'status:pending -WAITING', got %s", section.Filter)
	}
	if section.Description != "Next tasks to work on" {
		t.Errorf("Expected Description 'Next tasks to work on', got %s", section.Description)
	}
}

func TestDefaultSections(t *testing.T) {
	sections := DefaultSections()

	if len(sections) < 5 {
		t.Errorf("Expected at least 5 default sections, got %d", len(sections))
	}

	// Test that expected sections exist
	expectedSections := map[string]string{
		"Next":     "( status:pending or status:active ) -WAITING",
		"Waiting":  "status:waiting",
		"Projects": "status:pending or status:active",
		"Tags":     "status:pending or status:active",
		"All":      "status:pending or status:active",
	}

	sectionMap := make(map[string]Section)
	for _, s := range sections {
		sectionMap[s.Name] = s
	}

	for name, expectedFilter := range expectedSections {
		section, exists := sectionMap[name]
		if !exists {
			t.Errorf("Expected section '%s' not found", name)
			continue
		}
		if section.Filter != expectedFilter {
			t.Errorf("Section '%s' has filter '%s', expected '%s'", name, section.Filter, expectedFilter)
		}
		if section.Description == "" {
			t.Errorf("Section '%s' has empty description", name)
		}
	}
}

func TestBookmarkStruct(t *testing.T) {
	bookmark := Bookmark{
		Name:   "Work Tasks",
		Filter: "+work due:today",
	}

	if bookmark.Name != "Work Tasks" {
		t.Errorf("Expected Name 'Work Tasks', got %s", bookmark.Name)
	}
	if bookmark.Filter != "+work due:today" {
		t.Errorf("Expected Filter '+work due:today', got %s", bookmark.Filter)
	}
}

func TestSectionsWithBookmarks(t *testing.T) {
	bookmarks := []Bookmark{
		{Name: "Work", Filter: "+work"},
		{Name: "Home", Filter: "+home"},
	}

	sections := SectionsWithBookmarks(bookmarks)

	// Should have default sections + bookmarks
	defaultCount := len(DefaultSections())
	expectedCount := defaultCount + len(bookmarks)

	if len(sections) != expectedCount {
		t.Errorf("Expected %d sections, got %d", expectedCount, len(sections))
	}

	// Verify bookmarks are included
	foundWork := false
	foundHome := false
	for _, s := range sections {
		if s.Name == "Work" && s.Filter == "+work" {
			foundWork = true
		}
		if s.Name == "Home" && s.Filter == "+home" {
			foundHome = true
		}
	}

	if !foundWork {
		t.Error("Expected to find 'Work' bookmark section")
	}
	if !foundHome {
		t.Error("Expected to find 'Home' bookmark section")
	}
}

func TestSectionsWithEmptyBookmarks(t *testing.T) {
	sections := SectionsWithBookmarks([]Bookmark{})

	// Should just return default sections
	defaultSections := DefaultSections()
	if len(sections) != len(defaultSections) {
		t.Errorf("Expected %d sections, got %d", len(defaultSections), len(sections))
	}
}
