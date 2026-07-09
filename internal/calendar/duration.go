package calendar

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Nominal durations for calendar-oriented units. Months and years are
// approximations, which is acceptable for sizing a calendar event block.
const (
	durDay   = 24 * time.Hour
	durWeek  = 7 * durDay
	durMonth = 30 * durDay
	durYear  = 365 * durDay
)

// ParseTaskDuration parses a Taskwarrior duration value into a time.Duration.
//
// Taskwarrior exports duration-typed UDAs as ISO 8601 durations such as
// "PT30M", "PT1H30M" or "P1DT2H". For robustness this also accepts common
// shorthand ("30min", "1h30m", "2d") and a bare number (interpreted as
// seconds, matching Taskwarrior's default unit).
//
// It returns an error if the value cannot be parsed. The returned duration may
// be zero or negative; callers decide how to treat such values.
func ParseTaskDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if s[0] == 'P' || s[0] == 'p' {
		return parseISO8601Duration(s)
	}
	return parseShorthandDuration(s)
}

// parseISO8601Duration parses the subset of ISO 8601 durations that Taskwarrior
// emits: P[nY][nM][nD][T[nH][nM][nS]] and the week form P[nW].
func parseISO8601Duration(s string) (time.Duration, error) {
	rest := s[1:] // strip leading 'P'
	if rest == "" {
		return 0, fmt.Errorf("invalid ISO 8601 duration %q", s)
	}

	var total time.Duration
	inTime := false
	numStart := 0
	sawField := false

	for i := 0; i < len(rest); i++ {
		c := rest[i]
		if c == 'T' || c == 't' {
			inTime = true
			numStart = i + 1
			continue
		}
		if (c >= '0' && c <= '9') || c == '.' {
			continue
		}

		numStr := rest[numStart:i]
		if numStr == "" {
			return 0, fmt.Errorf("invalid ISO 8601 duration %q: missing value before %q", s, string(c))
		}
		val, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid ISO 8601 duration %q: %w", s, err)
		}

		var unit time.Duration
		switch c {
		case 'Y', 'y':
			unit = durYear
		case 'W', 'w':
			unit = durWeek
		case 'D', 'd':
			unit = durDay
		case 'H', 'h':
			unit = time.Hour
		case 'S', 's':
			unit = time.Second
		case 'M', 'm':
			// 'M' means minutes after the 'T' separator, months before it.
			if inTime {
				unit = time.Minute
			} else {
				unit = durMonth
			}
		default:
			return 0, fmt.Errorf("invalid ISO 8601 duration %q: unknown unit %q", s, string(c))
		}

		total += time.Duration(val * float64(unit))
		sawField = true
		numStart = i + 1
	}

	if !sawField {
		return 0, fmt.Errorf("invalid ISO 8601 duration %q: no fields", s)
	}
	return total, nil
}

// parseShorthandDuration parses human/Taskwarrior shorthand such as "30min",
// "1h30m", "2d" or a bare number (seconds). It is a lenient fallback for values
// that are not in ISO 8601 form.
func parseShorthandDuration(s string) (time.Duration, error) {
	s = strings.ToLower(s)

	var total time.Duration
	i := 0
	sawField := false

	for i < len(s) {
		if s[i] == ' ' {
			i++
			continue
		}

		start := i
		for i < len(s) && ((s[i] >= '0' && s[i] <= '9') || s[i] == '.') {
			i++
		}
		if start == i {
			return 0, fmt.Errorf("invalid duration %q", s)
		}
		numStr := s[start:i]

		unitStart := i
		for i < len(s) && s[i] >= 'a' && s[i] <= 'z' {
			i++
		}
		unitStr := s[unitStart:i]

		val, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", s, err)
		}
		unit, err := shorthandUnit(unitStr)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", s, err)
		}

		total += time.Duration(val * float64(unit))
		sawField = true
	}

	if !sawField {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	return total, nil
}

// shorthandUnit maps a shorthand unit suffix to its duration. An empty suffix
// defaults to seconds, matching Taskwarrior's bare-number convention.
func shorthandUnit(u string) (time.Duration, error) {
	switch u {
	case "", "s", "sec", "secs", "second", "seconds":
		return time.Second, nil
	case "m", "min", "mins", "minute", "minutes":
		return time.Minute, nil
	case "h", "hr", "hrs", "hour", "hours":
		return time.Hour, nil
	case "d", "day", "days":
		return durDay, nil
	case "w", "wk", "wks", "week", "weeks":
		return durWeek, nil
	case "mo", "mon", "month", "months":
		return durMonth, nil
	case "y", "yr", "yrs", "year", "years":
		return durYear, nil
	default:
		return 0, fmt.Errorf("unknown duration unit %q", u)
	}
}
