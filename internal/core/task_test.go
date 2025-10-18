package core

import (
	"testing"
	"time"
)

func TestTaskBasicFields(t *testing.T) {
	task := Task{
		UUID:        "abc-123",
		Description: "Test task",
		Project:     "TestProject",
		Tags:        []string{"work", "urgent"},
		Priority:    "H",
		Status:      "pending",
	}

	if task.UUID != "abc-123" {
		t.Errorf("Expected UUID abc-123, got %s", task.UUID)
	}
	if task.Description != "Test task" {
		t.Errorf("Expected Description 'Test task', got %s", task.Description)
	}
	if task.Project != "TestProject" {
		t.Errorf("Expected Project TestProject, got %s", task.Project)
	}
	if len(task.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(task.Tags))
	}
	if task.Priority != "H" {
		t.Errorf("Expected Priority H, got %s", task.Priority)
	}
	if task.Status != "pending" {
		t.Errorf("Expected Status pending, got %s", task.Status)
	}
}

func TestTaskDates(t *testing.T) {
	now := time.Now()
	task := Task{
		Due:       &now,
		Scheduled: &now,
		Wait:      &now,
		Entry:     now,
		Modified:  &now,
	}

	if task.Due == nil {
		t.Error("Expected Due to be set")
	}
	if task.Scheduled == nil {
		t.Error("Expected Scheduled to be set")
	}
	if task.Wait == nil {
		t.Error("Expected Wait to be set")
	}
	if task.Entry.IsZero() {
		t.Error("Expected Entry to be set")
	}
	if task.Modified == nil {
		t.Error("Expected Modified to be set")
	}
}

func TestTaskDependencies(t *testing.T) {
	task := Task{
		Depends: []string{"task1", "task2"},
	}

	if len(task.Depends) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(task.Depends))
	}
}

func TestTaskAnnotations(t *testing.T) {
	now := time.Now()
	task := Task{
		Annotations: []Annotation{
			{Entry: now, Description: "First note"},
			{Entry: now, Description: "Second note"},
		},
	}

	if len(task.Annotations) != 2 {
		t.Errorf("Expected 2 annotations, got %d", len(task.Annotations))
	}
	if task.Annotations[0].Description != "First note" {
		t.Errorf("Expected 'First note', got %s", task.Annotations[0].Description)
	}
}

func TestTaskUDAs(t *testing.T) {
	task := Task{
		UDAs: map[string]string{
			"estimate": "2h",
			"category": "development",
		},
	}

	if len(task.UDAs) != 2 {
		t.Errorf("Expected 2 UDAs, got %d", len(task.UDAs))
	}
	if task.UDAs["estimate"] != "2h" {
		t.Errorf("Expected estimate '2h', got %s", task.UDAs["estimate"])
	}
}

func TestGetUDA(t *testing.T) {
	task := Task{
		UDAs: map[string]string{
			"estimate": "2h",
			"category": "development",
		},
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"existing key", "estimate", "2h"},
		{"existing key 2", "category", "development"},
		{"missing key", "nonexistent", ""},
		{"nil UDAs", "any", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "nil UDAs" {
				task.UDAs = nil
			}
			result := task.GetUDA(tt.key)
			if result != tt.expected {
				t.Errorf("GetUDA(%s) = %s, expected %s", tt.key, result, tt.expected)
			}
		})
	}
}

func TestFormatDueDate(t *testing.T) {
	tests := []struct {
		name     string
		due      *time.Time
		expected string
	}{
		{
			name:     "no due date",
			due:      nil,
			expected: "",
		},
		{
			name:     "has due date",
			due:      timePtr(time.Date(2025, 10, 20, 14, 30, 0, 0, time.UTC)),
			expected: "2025-10-20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := Task{Due: tt.due}
			result := task.FormatDueDate()
			if result != tt.expected {
				t.Errorf("FormatDueDate() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

func TestIsOverdue(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	tests := []struct {
		name     string
		due      *time.Time
		status   string
		expected bool
	}{
		{"no due date", nil, "pending", false},
		{"completed task", &yesterday, "completed", false},
		{"deleted task", &yesterday, "deleted", false},
		{"due yesterday", &yesterday, "pending", true},
		{"due tomorrow", &tomorrow, "pending", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := Task{
				Due:    tt.due,
				Status: tt.status,
			}
			result := task.IsOverdue()
			if result != tt.expected {
				t.Errorf("IsOverdue() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsDueToday(t *testing.T) {
	now := time.Now()
	year, month, day := now.Date()
	today := time.Date(year, month, day, 12, 0, 0, 0, now.Location())
	yesterday := today.Add(-24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	tests := []struct {
		name     string
		due      *time.Time
		status   string
		expected bool
	}{
		{"no due date", nil, "pending", false},
		{"completed task due today", &today, "completed", false},
		{"deleted task due today", &today, "deleted", false},
		{"due today", &today, "pending", true},
		{"due yesterday", &yesterday, "pending", false},
		{"due tomorrow", &tomorrow, "pending", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := Task{
				Due:    tt.due,
				Status: tt.status,
			}
			result := task.IsDueToday()
			if result != tt.expected {
				t.Errorf("IsDueToday() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsDueSoon(t *testing.T) {
	now := time.Now()
	in3Days := now.Add(3 * 24 * time.Hour)
	in6Days := now.Add(6 * 24 * time.Hour)
	in8Days := now.Add(8 * 24 * time.Hour)
	yesterday := now.Add(-24 * time.Hour)

	tests := []struct {
		name     string
		due      *time.Time
		status   string
		expected bool
	}{
		{"no due date", nil, "pending", false},
		{"completed task due in 3 days", &in3Days, "completed", false},
		{"deleted task due in 3 days", &in3Days, "deleted", false},
		{"due in 3 days", &in3Days, "pending", true},
		{"due in 6 days", &in6Days, "pending", true},
		{"due in 8 days", &in8Days, "pending", false},
		{"due yesterday", &yesterday, "pending", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := Task{
				Due:    tt.due,
				Status: tt.status,
			}
			result := task.IsDueSoon()
			if result != tt.expected {
				t.Errorf("IsDueSoon() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
