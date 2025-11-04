package core

import (
	"sort"
	"testing"
)

// Test extracting main project from nested project names
func TestExtractMainProject(t *testing.T) {
	tests := []struct {
		name     string
		project  string
		expected string
	}{
		{"Simple project", "Work", "Work"},
		{"Nested project", "Work.company1", "Work"},
		{"Deeply nested", "Work.company1.backend", "Work"},
		{"Empty project", "", ""},
		{"Just dot", ".", ""},
		{"Dot at start", ".Work", ""},
		{"Dot at end", "Work.", "Work"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractMainProject(tt.project)
			if result != tt.expected {
				t.Errorf("ExtractMainProject(%q) = %q, want %q", tt.project, result, tt.expected)
			}
		})
	}
}

// Test grouping tasks by project
func TestGroupByProject(t *testing.T) {
	tasks := []Task{
		{UUID: "1", Description: "Task 1", Project: "Work"},
		{UUID: "2", Description: "Task 2", Project: "Work.company1"},
		{UUID: "3", Description: "Task 3", Project: "Home"},
		{UUID: "4", Description: "Task 4", Project: "Work.company2"},
		{UUID: "5", Description: "Task 5", Project: ""},
		{UUID: "6", Description: "Task 6", Project: "Shopping"},
	}

	groups := GroupByProject(tasks)

	// Check that we have the expected groups
	expectedGroups := map[string]int{
		"Work":     3, // Task 1, 2, 4
		"Home":     1, // Task 3
		"(none)":   1, // Task 5
		"Shopping": 1, // Task 6
	}

	if len(groups) != len(expectedGroups) {
		t.Errorf("Expected %d groups, got %d", len(expectedGroups), len(groups))
	}

	for _, group := range groups {
		expectedCount, exists := expectedGroups[group.Name]
		if !exists {
			t.Errorf("Unexpected group: %s", group.Name)
			continue
		}
		if group.Count != expectedCount {
			t.Errorf("Group %s: expected %d tasks, got %d", group.Name, expectedCount, group.Count)
		}
		if len(group.Tasks) != expectedCount {
			t.Errorf("Group %s: expected %d tasks in slice, got %d", group.Name, expectedCount, len(group.Tasks))
		}
	}
}

// Test grouping tasks by tag
func TestGroupByTag(t *testing.T) {
	tasks := []Task{
		{UUID: "1", Description: "Task 1", Tags: []string{"work", "urgent"}},
		{UUID: "2", Description: "Task 2", Tags: []string{"work", "backend"}},
		{UUID: "3", Description: "Task 3", Tags: []string{"home"}},
		{UUID: "4", Description: "Task 4", Tags: []string{"work"}},
		{UUID: "5", Description: "Task 5", Tags: []string{}},
	}

	groups := GroupByTag(tasks)

	// Check that we have the expected groups
	// Tasks can appear in multiple tag groups
	expectedGroups := map[string]int{
		"work":    3, // Task 1, 2, 4
		"urgent":  1, // Task 1
		"backend": 1, // Task 2
		"home":    1, // Task 3
		"(none)":  1, // Task 5
	}

	if len(groups) != len(expectedGroups) {
		t.Errorf("Expected %d groups, got %d", len(expectedGroups), len(groups))
	}

	for _, group := range groups {
		expectedCount, exists := expectedGroups[group.Name]
		if !exists {
			t.Errorf("Unexpected group: %s", group.Name)
			continue
		}
		if group.Count != expectedCount {
			t.Errorf("Group %s: expected %d tasks, got %d", group.Name, expectedCount, group.Count)
		}
	}
}

// Test that groups are sorted by name
func TestGroupsSorted(t *testing.T) {
	tasks := []Task{
		{UUID: "1", Project: "Zebra"},
		{UUID: "2", Project: "Apple"},
		{UUID: "3", Project: "Middle"},
	}

	groups := GroupByProject(tasks)

	// Check that groups are sorted (excluding "(none)" which should be last)
	var names []string
	for _, group := range groups {
		names = append(names, group.Name)
	}

	// Verify it's sorted
	if !sort.StringsAreSorted(names[:len(names)-1]) {
		// Check if "(none)" is at the end
		if names[len(names)-1] != "(none)" {
			t.Errorf("Groups not sorted correctly: %v", names)
		}
		// Check the rest are sorted
		if !sort.StringsAreSorted(names[:len(names)-1]) {
			t.Errorf("Groups (excluding (none)) not sorted: %v", names[:len(names)-1])
		}
	}
}

// Test filtering tasks by group
func TestFilterTasksByGroup(t *testing.T) {
	tasks := []Task{
		{UUID: "1", Description: "Task 1", Project: "Work"},
		{UUID: "2", Description: "Task 2", Project: "Work.company1"},
		{UUID: "3", Description: "Task 3", Project: "Home"},
	}

	groups := GroupByProject(tasks)

	// Find the Work group
	var workGroup *TaskGroup
	for i := range groups {
		if groups[i].Name == "Work" {
			workGroup = &groups[i]
			break
		}
	}

	if workGroup == nil {
		t.Fatal("Work group not found")
	}

	// Check that the Work group contains the right tasks
	if len(workGroup.Tasks) != 2 {
		t.Errorf("Expected 2 tasks in Work group, got %d", len(workGroup.Tasks))
	}

	// Verify the tasks are correct
	foundTask1 := false
	foundTask2 := false
	for _, task := range workGroup.Tasks {
		if task.UUID == "1" {
			foundTask1 = true
		}
		if task.UUID == "2" {
			foundTask2 = true
		}
	}

	if !foundTask1 || !foundTask2 {
		t.Error("Work group doesn't contain expected tasks")
	}
}

// Test empty task list
func TestGroupByProjectEmpty(t *testing.T) {
	tasks := []Task{}
	groups := GroupByProject(tasks)

	if len(groups) != 0 {
		t.Errorf("Expected 0 groups for empty task list, got %d", len(groups))
	}
}

func TestGroupByTagEmpty(t *testing.T) {
	tasks := []Task{}
	groups := GroupByTag(tasks)

	if len(groups) != 0 {
		t.Errorf("Expected 0 groups for empty task list, got %d", len(groups))
	}
}

// Test all tasks without project
func TestGroupByProjectAllNone(t *testing.T) {
	tasks := []Task{
		{UUID: "1", Description: "Task 1", Project: ""},
		{UUID: "2", Description: "Task 2", Project: ""},
	}

	groups := GroupByProject(tasks)

	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}

	if groups[0].Name != "(none)" {
		t.Errorf("Expected group name '(none)', got %q", groups[0].Name)
	}

	if groups[0].Count != 2 {
		t.Errorf("Expected 2 tasks in (none) group, got %d", groups[0].Count)
	}
}

// Test TaskGroup type
func TestTaskGroupType(t *testing.T) {
	group := TaskGroup{
		Name:  "Test",
		Count: 5,
		Tasks: []Task{{UUID: "1"}},
	}

	if group.Name != "Test" {
		t.Errorf("Expected name 'Test', got %q", group.Name)
	}
	if group.Count != 5 {
		t.Errorf("Expected count 5, got %d", group.Count)
	}
	if len(group.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(group.Tasks))
	}
}

func TestGroupProjectsByHierarchy(t *testing.T) {
	// Test data with various hierarchy levels
	summaries := []ProjectSummary{
		{Name: "M8s", Percentage: 51},
		{Name: "M8s.helm", Percentage: 0},
		{Name: "M8s.SNR", Percentage: 47},
		{Name: "M8s.SNR.RHWA12", Percentage: 0}, // Should now be included with depth 2
		{Name: "WUI", Percentage: 0},
		{Name: "dty", Percentage: 87},
		{Name: "dty.condominio", Percentage: 87},
	}

	groups := GroupProjectsByHierarchy(summaries, []Task{})

	// Should include ALL projects regardless of depth
	expected := map[string]struct {
		percentage int
		isSubitem  bool
		depth      int
	}{
		"M8s":            {51, false, 0},
		"M8s.helm":       {0, true, 1},
		"M8s.SNR":        {47, true, 1},
		"M8s.SNR.RHWA12": {0, true, 2}, // Now included with depth 2
		"WUI":            {0, false, 0},
		"dty":            {87, false, 0},
		"dty.condominio": {87, true, 1},
	}

	if len(groups) != len(expected) {
		t.Errorf("Expected %d groups, got %d", len(expected), len(groups))
	}

	// Verify each group
	for _, group := range groups {
		exp, found := expected[group.Name]
		if !found {
			t.Errorf("Unexpected group: %s", group.Name)
			continue
		}
		if group.Percentage != exp.percentage {
			t.Errorf("Group %s: expected %d%%, got %d%%", group.Name, exp.percentage, group.Percentage)
		}
		if group.IsSubitem != exp.isSubitem {
			t.Errorf("Group %s: expected IsSubitem=%v, got %v", group.Name, exp.isSubitem, group.IsSubitem)
		}
		if group.Depth != exp.depth {
			t.Errorf("Group %s: expected Depth=%d, got %d", group.Name, exp.depth, group.Depth)
		}
	}
}
