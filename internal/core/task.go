package core

import (
	"fmt"
	"strings"
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

// GetProperty returns the value of any task property as a formatted string
// Returns empty string and false if the property doesn't exist
// Supports all standard taskwarrior properties and UDAs
func (t *Task) GetProperty(name string) (string, bool) {
	switch name {
	case "id":
		if t.ID == 0 {
			return t.UUID, true
		}
		return fmt.Sprintf("%d", t.ID), true
	case "uuid":
		return t.UUID, true
	case "short_uuid":
		// Return first 8 characters of UUID
		if len(t.UUID) > 8 {
			return t.UUID[:8], true
		}
		return t.UUID, true
	case "description":
		return t.Description, true
	case "project":
		if t.Project == "" {
			return "-", true
		}
		return t.Project, true
	case "priority":
		if t.Priority == "" {
			return "-", true
		}
		return string(t.Priority[0]), true
	case "status":
		return t.Status, true
	case "tags":
		if len(t.Tags) == 0 {
			return "-", true
		}
		var tagList []string
		for _, tag := range t.Tags {
			tagList = append(tagList, "+"+tag)
		}
		return strings.Join(tagList, ", "), true
	case "due":
		if t.Due == nil {
			return "-", true
		}
		return FormatRelativeDate(t.Due), true
	case "scheduled":
		if t.Scheduled == nil {
			return "-", true
		}
		return FormatRelativeDate(t.Scheduled), true
	case "wait":
		if t.Wait == nil {
			return "-", true
		}
		return FormatRelativeDate(t.Wait), true
	case "start":
		if t.Start == nil {
			return "-", true
		}
		return FormatRelativeDate(t.Start), true
	case "entry":
		return FormatRelativeDate(&t.Entry), true
	case "modified":
		if t.Modified == nil {
			return "-", true
		}
		return FormatRelativeDate(t.Modified), true
	case "end":
		if t.End == nil {
			return "-", true
		}
		return FormatRelativeDate(t.End), true
	case "urgency":
		return fmt.Sprintf("%.1f", t.Urgency), true
	case "annotation":
		if len(t.Annotations) > 0 {
			return "*", true
		}
		return "-", true
	case "dependency":
		if len(t.Depends) > 0 {
			return "*", true
		}
		return "-", true
	default:
		// Check if it's a UDA
		if val := t.GetUDA(name); val != "" {
			return val, true
		}
		return "", false
	}
}

// formatDate formats a time.Time pointer as YYYY-MM-DD or YYYY-MM-DD HH:MM
// Shows time only if it's not midnight. Converts to local timezone for display.
func (t *Task) formatDate(date *time.Time) string {
	if date == nil {
		return ""
	}
	// Convert to local timezone
	localTime := date.Local()

	// Show time if it's not midnight (00:00:00) in local time
	if localTime.Hour() == 0 && localTime.Minute() == 0 && localTime.Second() == 0 {
		return localTime.Format("2006-01-02")
	}
	return localTime.Format("2006-01-02 15:04")
}

// FormatRelativeDate returns a compact human-readable relative date string.
// Past dates include "ago" suffix (e.g., "2 weeks ago").
// Future dates include "in" prefix (e.g., "in 3 days").
func FormatRelativeDate(date *time.Time) string {
	return formatRelativeDateFrom(date, time.Now())
}

// formatRelativeDateFrom returns a relative date string computed against a given reference time.
// Extracted for testability.
func formatRelativeDateFrom(date *time.Time, now time.Time) string {
	if date == nil {
		return ""
	}
	diff := now.Sub(*date)

	if diff < 0 {
		// Future dates
		diff = -diff
		switch {
		case diff < time.Minute:
			return "now"
		case diff < time.Hour:
			mins := int(diff.Minutes())
			if mins == 1 {
				return "in 1 min"
			}
			return fmt.Sprintf("in %d min", mins)
		case diff < 24*time.Hour:
			hours := int(diff.Hours())
			if hours == 1 {
				return "in 1 hour"
			}
			return fmt.Sprintf("in %d hours", hours)
		case diff < 48*time.Hour:
			return "tomorrow"
		case diff < 7*24*time.Hour:
			days := int(diff.Hours() / 24)
			return fmt.Sprintf("in %d days", days)
		case diff < 30*24*time.Hour:
			weeks := int(diff.Hours() / 24 / 7)
			if weeks == 1 {
				return "in 1 week"
			}
			return fmt.Sprintf("in %d weeks", weeks)
		case diff < 365*24*time.Hour:
			months := int(diff.Hours() / 24 / 30)
			if months == 1 {
				return "in 1 month"
			}
			return fmt.Sprintf("in %d months", months)
		default:
			years := int(diff.Hours() / 24 / 365)
			if years == 1 {
				return "in 1 year"
			}
			return fmt.Sprintf("in %d years", years)
		}
	}

	// Past dates
	switch {
	case diff < time.Minute:
		return "now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d min ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 48*time.Hour:
		return "yesterday"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(diff.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
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
