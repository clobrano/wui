package taskwarrior

import (
	"fmt"
	"time"

	"github.com/clobrano/wui/internal/core"
)

// MapToCore converts a TaskwarriorTask to a core.Task
func MapToCore(t TaskwarriorTask) core.Task {
	coreTask := core.Task{
		ID:          t.ID,
		UUID:        t.UUID,
		Description: t.Description,
		Project:     t.Project,
		Tags:        t.Tags,
		Priority:    t.Priority,
		Status:      t.Status,
		Urgency:     t.Urgency,
	}

	// Parse dates
	coreTask.Due, _ = parseTaskwarriorDate(t.Due)
	coreTask.Scheduled, _ = parseTaskwarriorDate(t.Scheduled)
	coreTask.Wait, _ = parseTaskwarriorDate(t.Wait)
	coreTask.Start, _ = parseTaskwarriorDate(t.Start)
	coreTask.Modified, _ = parseTaskwarriorDate(t.Modified)
	coreTask.End, _ = parseTaskwarriorDate(t.End)

	// Entry is required, parse with default to zero time on error
	entryTime, err := parseTaskwarriorDate(t.Entry)
	if err == nil && entryTime != nil {
		coreTask.Entry = *entryTime
	} else {
		coreTask.Entry = time.Time{}
	}

	// Parse dependencies (array of UUIDs)
	if t.Depends != nil && len(t.Depends) > 0 {
		coreTask.Depends = t.Depends
	} else {
		coreTask.Depends = []string{}
	}

	// Map annotations
	coreTask.Annotations = make([]core.Annotation, len(t.Annotations))
	for i, a := range t.Annotations {
		annotationTime, err := parseTaskwarriorDate(a.Entry)
		if err == nil && annotationTime != nil {
			coreTask.Annotations[i] = core.Annotation{
				Entry:       *annotationTime,
				Description: a.Description,
			}
		} else {
			// If date parsing fails, use zero time
			coreTask.Annotations[i] = core.Annotation{
				Entry:       time.Time{},
				Description: a.Description,
			}
		}
	}

	// Map UDAs (User Defined Attributes)
	coreTask.UDAs = make(map[string]string)
	for key, value := range t.UDA {
		// Convert interface{} to string
		if value != nil {
			coreTask.UDAs[key] = fmt.Sprintf("%v", value)
		}
	}

	return coreTask
}

// parseTaskwarriorDate parses a Taskwarrior date string
// Taskwarrior uses the format: 20251016T120000Z (ISO 8601 with Z suffix)
// Returns nil if the date string is empty, or error if parsing fails
func parseTaskwarriorDate(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	// Taskwarrior date format: YYYYMMDDTHHmmssZ
	tm, err := time.Parse("20060102T150405Z", dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %s: %w", dateStr, err)
	}

	return &tm, nil
}
