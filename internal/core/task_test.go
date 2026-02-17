package core

import (
	"strings"
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
			name:     "has due date with time",
			due:      timePtr(time.Date(2025, 10, 20, 14, 30, 0, 0, time.Local)),
			expected: "2025-10-20 14:30",
		},
		{
			name:     "has due date at midnight",
			due:      timePtr(time.Date(2025, 10, 20, 0, 0, 0, 0, time.Local)),
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

func TestToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		task     Task
		expected string
	}{
		{
			name: "pending task",
			task: Task{
				Description: "Test task",
				UUID:        "12345678-abcd-efgh",
				Status:      "pending",
			},
			expected: "* [ ] Test task (12345678)",
		},
		{
			name: "completed task",
			task: Task{
				Description: "Completed task",
				UUID:        "abcd1234",
				Status:      "completed",
			},
			expected: "* [x] Completed task (abcd1234)",
		},
		{
			name: "started task",
			task: Task{
				Description: "Active task",
				UUID:        "xyz123456789",
				Status:      "pending",
				Start:       timePtr(time.Now()),
			},
			expected: "* [S] Active task (xyz12345)",
		},
		{
			name: "deleted task",
			task: Task{
				Description: "Deleted task",
				UUID:        "delete-me-uuid",
				Status:      "deleted",
			},
			expected: "* [d] Deleted task (delete-m)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.task.ToMarkdown()
			if result != tt.expected {
				t.Errorf("ToMarkdown() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestFormatRelativeDateFrom(t *testing.T) {
	now := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		date     *time.Time
		expected string
	}{
		{"nil date", nil, ""},
		// Past dates (with "ago")
		{"just now", timePtr(now.Add(-10 * time.Second)), "now"},
		{"1 min ago", timePtr(now.Add(-1 * time.Minute)), "1 min ago"},
		{"5 min ago", timePtr(now.Add(-5 * time.Minute)), "5 min ago"},
		{"1 hour ago", timePtr(now.Add(-1 * time.Hour)), "1 hour ago"},
		{"3 hours ago", timePtr(now.Add(-3 * time.Hour)), "3 hours ago"},
		{"yesterday", timePtr(now.Add(-30 * time.Hour)), "yesterday"},
		{"3 days ago", timePtr(now.Add(-3 * 24 * time.Hour)), "3 days ago"},
		{"1 week ago", timePtr(now.Add(-8 * 24 * time.Hour)), "1 week ago"},
		{"3 weeks ago", timePtr(now.Add(-22 * 24 * time.Hour)), "3 weeks ago"},
		{"1 month ago", timePtr(now.Add(-45 * 24 * time.Hour)), "1 month ago"},
		{"6 months ago", timePtr(now.Add(-180 * 24 * time.Hour)), "6 months ago"},
		{"1 year ago", timePtr(now.Add(-400 * 24 * time.Hour)), "1 year ago"},
		{"3 years ago", timePtr(now.Add(-3 * 365 * 24 * time.Hour)), "3 years ago"},
		// Future dates (with "+" prefix)
		{"+moments", timePtr(now.Add(10 * time.Second)), "now"},
		{"+1 min", timePtr(now.Add(1 * time.Minute)), "+1 min"},
		{"+5 min", timePtr(now.Add(5 * time.Minute)), "+5 min"},
		{"+1 hour", timePtr(now.Add(1 * time.Hour)), "+1 hour"},
		{"+3 hours", timePtr(now.Add(3 * time.Hour)), "+3 hours"},
		{"tomorrow", timePtr(now.Add(30 * time.Hour)), "tomorrow"},
		{"+3 days", timePtr(now.Add(3 * 24 * time.Hour)), "+3 days"},
		{"+1 week", timePtr(now.Add(8 * 24 * time.Hour)), "+1 week"},
		{"+3 weeks", timePtr(now.Add(22 * 24 * time.Hour)), "+3 weeks"},
		{"+1 month", timePtr(now.Add(45 * 24 * time.Hour)), "+1 month"},
		{"+6 months", timePtr(now.Add(180 * 24 * time.Hour)), "+6 months"},
		{"+1 year", timePtr(now.Add(400 * 24 * time.Hour)), "+1 year"},
		{"+3 years", timePtr(now.Add(3 * 365 * 24 * time.Hour)), "+3 years"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRelativeDateFrom(tt.date, now)
			if result != tt.expected {
				t.Errorf("formatRelativeDateFrom() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGetPropertyReturnsAbsoluteDates(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	task := Task{
		Due:    &yesterday,
		Status: "pending",
		Entry:  now,
	}

	// GetProperty should return absolute date format (YYYY-MM-DD)
	dueVal, ok := task.GetProperty("due")
	if !ok {
		t.Fatal("Expected GetProperty('due') to return ok=true")
	}
	// Should be an absolute date, not relative
	if len(dueVal) < 10 || dueVal[4] != '-' || dueVal[7] != '-' {
		t.Errorf("Expected absolute date format YYYY-MM-DD, got %q", dueVal)
	}

	// Entry (always set) should return absolute date
	entryVal, ok := task.GetProperty("entry")
	if !ok {
		t.Fatal("Expected GetProperty('entry') to return ok=true")
	}
	if len(entryVal) < 10 || entryVal[4] != '-' || entryVal[7] != '-' {
		t.Errorf("Expected absolute date format YYYY-MM-DD, got %q", entryVal)
	}
}

func TestGetDateValue(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	task := Task{
		Due:   &yesterday,
		Entry: now,
	}

	// Due should return the pointer
	dueVal := task.GetDateValue("due")
	if dueVal == nil {
		t.Fatal("Expected GetDateValue('due') to return non-nil")
	}
	if !dueVal.Equal(yesterday) {
		t.Errorf("Expected due date to match, got %v", dueVal)
	}

	// Entry should return a pointer to Entry
	entryVal := task.GetDateValue("entry")
	if entryVal == nil {
		t.Fatal("Expected GetDateValue('entry') to return non-nil")
	}

	// Nil date field
	task.Scheduled = nil
	schedVal := task.GetDateValue("scheduled")
	if schedVal != nil {
		t.Errorf("Expected nil for unset scheduled, got %v", schedVal)
	}

	// Non-date field
	unknownVal := task.GetDateValue("description")
	if unknownVal != nil {
		t.Errorf("Expected nil for non-date field, got %v", unknownVal)
	}

	// FormatRelativeDate on past date should contain "ago" or "yesterday"
	relVal := FormatRelativeDate(dueVal)
	if relVal != "yesterday" && !strings.Contains(relVal, "ago") {
		t.Errorf("Expected relative date with 'ago' or 'yesterday', got %q", relVal)
	}
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
