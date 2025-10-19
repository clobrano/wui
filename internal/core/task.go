package core

import (
	"fmt"
	"time"
)

// Task represents a task in the domain model (UI-agnostic)
type Task struct {
	// Core fields
	ID          int    // Taskwarrior's sequential ID (1, 2, 3, ...)
	UUID        string // Unique identifier
	Description string
	Project     string
	Tags        []string
	Priority    string // H, M, L, or empty
	Status      string // pending, completed, deleted, waiting, recurring

	// Date fields
	Due       *time.Time // Due date
	Scheduled *time.Time // Scheduled date
	Wait      *time.Time // Wait until date
	Start     *time.Time // Start date (when task was started)
	Entry     time.Time  // Created date
	Modified  *time.Time // Last modified date
	End       *time.Time // Completion date

	// Dependencies
	Depends []string // UUIDs of tasks this depends on

	// Annotations
	Annotations []Annotation

	// User Defined Attributes
	UDAs map[string]string

	// Urgency calculation (from Taskwarrior)
	Urgency float64
}

// Annotation represents a note added to a task
type Annotation struct {
	Entry       time.Time
	Description string
}

// GetUDA returns the value of a User Defined Attribute by key
// Returns empty string if the key doesn't exist or UDAs is nil
func (t *Task) GetUDA(key string) string {
	if t.UDAs == nil {
		return ""
	}
	return t.UDAs[key]
}

// FormatDueDate returns the due date formatted as YYYY-MM-DD or YYYY-MM-DD HH:MM
// Shows time only if it's not midnight. Returns empty string if due date is not set.
// Converts to local timezone for display.
func (t *Task) FormatDueDate() string {
	if t.Due == nil {
		return ""
	}
	// Convert to local timezone
	localTime := t.Due.Local()

	// Show time if it's not midnight (00:00:00) in local time
	if localTime.Hour() == 0 && localTime.Minute() == 0 && localTime.Second() == 0 {
		return localTime.Format("2006-01-02")
	}
	return localTime.Format("2006-01-02 15:04")
}

// IsOverdue returns true if the task is overdue
// A task is overdue if it has a due date in the past and is not completed or deleted
func (t *Task) IsOverdue() bool {
	if t.Due == nil {
		return false
	}
	if t.Status == "completed" || t.Status == "deleted" {
		return false
	}
	return t.Due.Before(time.Now())
}

// IsDueToday returns true if the task is due today
func (t *Task) IsDueToday() bool {
	if t.Due == nil {
		return false
	}
	if t.Status == "completed" || t.Status == "deleted" {
		return false
	}
	now := time.Now()
	year, month, day := now.Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)

	return !t.Due.Before(today) && t.Due.Before(tomorrow)
}

// IsDueSoon returns true if the task is due within the next 7 days
func (t *Task) IsDueSoon() bool {
	if t.Due == nil {
		return false
	}
	if t.Status == "completed" || t.Status == "deleted" {
		return false
	}
	now := time.Now()
	sevenDaysFromNow := now.AddDate(0, 0, 7)

	return t.Due.After(now) && t.Due.Before(sevenDaysFromNow)
}

// ToMarkdown formats the task as a markdown checklist item
// Format: * [ ] Description (short-uuid)
// Status markers: [ ] pending, [x] completed, [S] started, [d] deleted
func (t *Task) ToMarkdown() string {
	// Determine status marker
	var statusMarker string
	switch t.Status {
	case "completed":
		statusMarker = "x"
	case "deleted":
		statusMarker = "d"
	default:
		// pending or waiting - check if started
		if t.Start != nil {
			statusMarker = "S"
		} else {
			statusMarker = " "
		}
	}

	// Get short UUID (first 8 characters)
	shortUUID := t.UUID
	if len(shortUUID) > 8 {
		shortUUID = shortUUID[:8]
	}

	return fmt.Sprintf("* [%s] %s (%s)", statusMarker, t.Description, shortUUID)
}
