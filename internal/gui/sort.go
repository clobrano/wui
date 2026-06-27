package gui

import (
	"sort"
	"strings"
	"time"

	"github.com/clobrano/wui/internal/core"
)

// SortTasks sorts a copy of tasks using the given method and direction.
// Mirrors the logic in internal/tui/components/tasklist.go so both UIs behave identically.
func SortTasks(tasks []core.Task, method string, reverse bool) []core.Task {
	sorted := make([]core.Task, len(tasks))
	copy(sorted, tasks)

	sort.SliceStable(sorted, func(i, j int) bool {
		ti, tj := sorted[i], sorted[j]

		// Completed tasks always sink to the bottom.
		ci, cj := ti.Status == "completed", tj.Status == "completed"
		if ci != cj {
			return !ci
		}

		if method != "" {
			result := compareTasks(ti, tj, method)
			if result != 0 {
				if reverse {
					return result > 0
				}
				return result < 0
			}
		}
		return false
	})

	return sorted
}

func compareTasks(a, b core.Task, method string) int {
	switch method {
	case "alphabetic", "alpha", "description":
		return strings.Compare(strings.ToLower(a.Description), strings.ToLower(b.Description))
	case "due":
		return compareDates(a.Due, b.Due)
	case "scheduled":
		return compareDates(a.Scheduled, b.Scheduled)
	case "created", "entry":
		return compareTime(a.Entry, b.Entry)
	case "modified":
		return compareDates(a.Modified, b.Modified)
	case "urgency":
		if a.Urgency > b.Urgency {
			return -1
		} else if a.Urgency < b.Urgency {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func compareDates(a, b *time.Time) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return 1
	}
	if b == nil {
		return -1
	}
	return compareTime(*a, *b)
}

func compareTime(a, b time.Time) int {
	if a.Before(b) {
		return -1
	}
	if a.After(b) {
		return 1
	}
	return 0
}
