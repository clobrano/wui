package gui

import (
	"fmt"
	"time"
)

// ShortRelDate formats a date as a compact relative string for task list rows.
// Examples: "today", "tmrw", "yest", "3d", "2w", "1m", "1y",
// and past equivalents "3d ago", "2w ago", "1m ago", "1y ago".
func ShortRelDate(t *time.Time) string {
	return shortRelDateFrom(t, time.Now())
}

// ShortRelDateFrom is the testable variant of ShortRelDate with an explicit reference time.
func ShortRelDateFrom(t *time.Time, now time.Time) string {
	return shortRelDateFrom(t, now)
}

func shortRelDateFrom(t *time.Time, now time.Time) string {
	if t == nil {
		return ""
	}

	// Normalise to date boundaries (midnight local time).
	localNow := now.Local()
	nowDay := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, localNow.Location())
	localT := t.Local()
	tDay := time.Date(localT.Year(), localT.Month(), localT.Day(), 0, 0, 0, 0, localT.Location())

	days := int(tDay.Sub(nowDay).Hours() / 24)

	switch {
	case days == 0:
		return "today"
	case days == 1:
		return "tmrw"
	case days == -1:
		return "yest"
	case days > 0:
		return compactFuture(days)
	default:
		return compactPast(-days)
	}
}

func compactFuture(days int) string {
	switch {
	case days < 7:
		return fmt.Sprintf("%dd", days)
	case days < 30:
		return fmt.Sprintf("%dw", days/7)
	case days < 365:
		return fmt.Sprintf("%dm", days/30)
	default:
		return fmt.Sprintf("%dy", days/365)
	}
}

func compactPast(days int) string {
	switch {
	case days < 7:
		return fmt.Sprintf("%dd ago", days)
	case days < 30:
		return fmt.Sprintf("%dw ago", days/7)
	case days < 365:
		return fmt.Sprintf("%dm ago", days/30)
	default:
		return fmt.Sprintf("%dy ago", days/365)
	}
}

// LongRelDate formats a date as a verbose relative string for the detail view.
// Examples: "today", "tomorrow", "yesterday", "in 3 days", "2 weeks ago".
// Returns the relative part; callers combine it with the absolute date.
func LongRelDate(t *time.Time) string {
	return longRelDateFrom(t, time.Now())
}

// LongRelDateFrom is the testable variant of LongRelDate with an explicit reference time.
func LongRelDateFrom(t *time.Time, now time.Time) string {
	return longRelDateFrom(t, now)
}

func longRelDateFrom(t *time.Time, now time.Time) string {
	if t == nil {
		return ""
	}

	localNow := now.Local()
	nowDay := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, localNow.Location())
	localT := t.Local()
	tDay := time.Date(localT.Year(), localT.Month(), localT.Day(), 0, 0, 0, 0, localT.Location())

	days := int(tDay.Sub(nowDay).Hours() / 24)

	switch {
	case days == 0:
		return "today"
	case days == 1:
		return "tomorrow"
	case days == -1:
		return "yesterday"
	case days > 0:
		return verboseFuture(days)
	default:
		return verbosePast(-days)
	}
}

func verboseFuture(days int) string {
	switch {
	case days < 7:
		return fmt.Sprintf("in %d day%s", days, plural(days))
	case days < 30:
		w := days / 7
		return fmt.Sprintf("in %d week%s", w, plural(w))
	case days < 365:
		m := days / 30
		return fmt.Sprintf("in %d month%s", m, plural(m))
	default:
		y := days / 365
		return fmt.Sprintf("in %d year%s", y, plural(y))
	}
}

func verbosePast(days int) string {
	switch {
	case days < 7:
		return fmt.Sprintf("%d day%s ago", days, plural(days))
	case days < 30:
		w := days / 7
		return fmt.Sprintf("%d week%s ago", w, plural(w))
	case days < 365:
		m := days / 30
		return fmt.Sprintf("%d month%s ago", m, plural(m))
	default:
		y := days / 365
		return fmt.Sprintf("%d year%s ago", y, plural(y))
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// FormatAbsDate formats a time.Time as "Jan 2, 2006" for the detail view.
func FormatAbsDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Local().Format("Jan 2, 2006")
}

// DueCSSClass returns a CSS class name for due-date row tinting.
// Classes: "due-overdue", "due-today", "due-soon" (< 7 days), or "".
// Comparison is done at day granularity (midnight local time).
func DueCSSClass(t *time.Time) string {
	return dueCSSClassFrom(t, time.Now())
}

// DueCSSClassFrom is the testable variant with an explicit reference time.
func DueCSSClassFrom(t *time.Time, now time.Time) string {
	return dueCSSClassFrom(t, now)
}

func dueCSSClassFrom(t *time.Time, now time.Time) string {
	if t == nil {
		return ""
	}
	localNow := now.Local()
	nowDay := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, localNow.Location())
	localT := t.Local()
	tDay := time.Date(localT.Year(), localT.Month(), localT.Day(), 0, 0, 0, 0, localT.Location())
	days := int(tDay.Sub(nowDay).Hours() / 24)
	switch {
	case days < 0:
		return "due-overdue"
	case days == 0:
		return "due-today"
	case days < 7:
		return "due-soon"
	default:
		return ""
	}
}
