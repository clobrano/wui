package calendar

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/clobrano/wui/internal/core"
	"github.com/clobrano/wui/internal/taskwarrior"
	"google.golang.org/api/calendar/v3"
)

// SyncClient handles synchronization between Taskwarrior and Google Calendar
type SyncClient struct {
	calendarService *calendar.Service
	taskClient      *taskwarrior.Client
	calendarName    string
	taskFilter      string
}

// NewSyncClient creates a new sync client
func NewSyncClient(ctx context.Context, taskClient *taskwarrior.Client, credentialsPath, tokenPath, calendarName, taskFilter string) (*SyncClient, error) {
	// Get authenticated calendar service
	calService, err := GetOAuth2Client(ctx, credentialsPath, tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar client: %w", err)
	}

	return &SyncClient{
		calendarService: calService,
		taskClient:      taskClient,
		calendarName:    calendarName,
		taskFilter:      taskFilter,
	}, nil
}

// SyncResult contains the results of a sync operation
type SyncResult struct {
	Total    int
	Created  int
	Updated  int
	Deleted  int
	Skipped  int
	Warnings []string
}

// Sync performs the synchronization from Taskwarrior to Google Calendar
func (s *SyncClient) Sync(ctx context.Context) (*SyncResult, error) {
	slog.Info("Starting sync", "calendar", s.calendarName, "filter", s.taskFilter)

	result := &SyncResult{
		Warnings: make([]string, 0),
	}

	// Get the calendar ID by name
	calendarID, err := s.findCalendarByName(ctx, s.calendarName)
	if err != nil {
		return nil, fmt.Errorf("failed to find calendar: %w", err)
	}

	slog.Info("Found calendar", "id", calendarID, "name", s.calendarName)

	// Get tasks from Taskwarrior
	tasks, err := s.taskClient.Export(s.taskFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	slog.Info("Retrieved tasks", "count", len(tasks))

	// Get existing events from calendar
	existingEvents, err := s.getCalendarEvents(ctx, calendarID)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar events: %w", err)
	}

	slog.Info("Retrieved existing calendar events", "count", len(existingEvents))

	// Build map of existing events by task UUID
	eventMap := make(map[string]*calendar.Event)
	for _, event := range existingEvents {
		if uuid := extractUUIDFromEvent(event); uuid != "" {
			eventMap[uuid] = event
		}
	}

	// Build map of tasks by UUID for deletion tracking
	taskMap := make(map[string]core.Task)
	for _, task := range tasks {
		taskMap[task.UUID] = task
	}

	// Sync each task
	created := 0
	updated := 0
	deleted := 0
	skipped := 0
	warnings := 0
	for _, task := range tasks {
		// Check if task has no due date and no scheduled date
		hasNoDates := (task.Due == nil || task.Due.IsZero()) && (task.Scheduled == nil || task.Scheduled.IsZero())

		if hasNoDates {
			// If task has an existing event, delete it
			if existingEvent, exists := eventMap[task.UUID]; exists {
				slog.Info("Deleting event for task without dates", "uuid", task.UUID, "description", task.Description)
				if err := s.deleteEvent(ctx, calendarID, existingEvent.Id); err != nil {
					slog.Error("Failed to delete event", "uuid", task.UUID, "error", err)
					continue
				}
				deleted++
				// Remove from eventMap to avoid processing again
				delete(eventMap, task.UUID)
			} else {
				slog.Debug("Skipping task without due date or scheduled date", "uuid", task.UUID, "description", task.Description)
				skipped++
			}
			continue
		}

		// Check if scheduled is after due and log warning
		if task.Scheduled != nil && !task.Scheduled.IsZero() && task.Due != nil && !task.Due.IsZero() {
			if task.Scheduled.After(*task.Due) || task.Scheduled.Equal(*task.Due) {
				warningMsg := fmt.Sprintf("Task '%s' has scheduled time (%s) after due time (%s)",
					task.Description,
					task.Scheduled.Format("2006-01-02 15:04"),
					task.Due.Format("2006-01-02 15:04"))
				slog.Warn("Task has scheduled time after or equal to due time",
					"uuid", task.UUID,
					"description", task.Description,
					"scheduled", task.Scheduled.Format("2006-01-02 15:04:05"),
					"due", task.Due.Format("2006-01-02 15:04:05"))
				fmt.Printf("⚠️  WARNING: %s\n", warningMsg)
				result.Warnings = append(result.Warnings, warningMsg)
				warnings++
			}
		}

		if existingEvent, exists := eventMap[task.UUID]; exists {
			// Update existing event
			slog.Info("Found existing event for task",
				"uuid", task.UUID,
				"description", task.Description,
				"task_due", task.Due,
				"task_scheduled", task.Scheduled)

			// Log event details
			slog.Info("Existing event details",
				"uuid", task.UUID,
				"event_summary", existingEvent.Summary,
				"event_start", existingEvent.Start,
				"event_has_reminders", existingEvent.Reminders != nil,
				"event_reminders_use_default", existingEvent.Reminders != nil && existingEvent.Reminders.UseDefault)

			if existingEvent.Reminders != nil && len(existingEvent.Reminders.Overrides) > 0 {
				slog.Info("Event reminder details",
					"uuid", task.UUID,
					"reminder_minutes", existingEvent.Reminders.Overrides[0].Minutes,
					"reminder_method", existingEvent.Reminders.Overrides[0].Method)
			}

			needsUpdate := s.shouldUpdateEvent(task, existingEvent)
			slog.Info("Update check result",
				"uuid", task.UUID,
				"description", task.Description,
				"needs_update", needsUpdate)
			if needsUpdate {
				if err := s.updateEvent(ctx, calendarID, task, existingEvent); err != nil {
					slog.Error("Failed to update event", "uuid", task.UUID, "error", err)
					continue
				}
				slog.Info("Updated event", "uuid", task.UUID, "description", task.Description)
				updated++
			} else {
				slog.Debug("Event already up to date", "uuid", task.UUID)
			}
		} else {
			// Create new event
			if err := s.createEvent(ctx, calendarID, task); err != nil {
				slog.Error("Failed to create event", "uuid", task.UUID, "error", err)
				continue
			}
			created++
		}
	}

	// Populate result
	result.Total = len(tasks)
	result.Created = created
	result.Updated = updated
	result.Deleted = deleted
	result.Skipped = skipped

	slog.Info("Sync completed", "total", result.Total, "created", result.Created, "updated", result.Updated, "deleted", result.Deleted, "skipped", result.Skipped, "warnings", len(result.Warnings))

	// Print summary (for CLI mode and visibility)
	fmt.Printf("\nSync completed: %d tasks, %d created, %d updated, %d deleted, %d skipped (no dates)\n",
		result.Total, result.Created, result.Updated, result.Deleted, result.Skipped)
	if len(result.Warnings) > 0 {
		fmt.Printf("⚠️  %d WARNINGS: Tasks with scheduled > due\n", len(result.Warnings))
	}

	return result, nil
}

// findCalendarByName finds a calendar ID by its name
func (s *SyncClient) findCalendarByName(ctx context.Context, name string) (string, error) {
	calendarList, err := s.calendarService.CalendarList.List().Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to list calendars: %w", err)
	}

	for _, cal := range calendarList.Items {
		if cal.Summary == name {
			return cal.Id, nil
		}
	}

	return "", fmt.Errorf("calendar '%s' not found", name)
}

// getCalendarEvents retrieves events from the calendar that were created by this tool
func (s *SyncClient) getCalendarEvents(ctx context.Context, calendarID string) ([]*calendar.Event, error) {
	// Get events from the past 30 days to the next 365 days
	timeMin := time.Now().AddDate(0, 0, -30).Format(time.RFC3339)
	timeMax := time.Now().AddDate(1, 0, 0).Format(time.RFC3339)

	events, err := s.calendarService.Events.List(calendarID).
		Context(ctx).
		TimeMin(timeMin).
		TimeMax(timeMax).
		SingleEvents(true).
		OrderBy("startTime").
		Fields("items(id,summary,description,start,end,colorId,reminders)").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	// Filter events created by wui (check for UUID in description)
	var wuiEvents []*calendar.Event
	for _, event := range events.Items {
		if extractUUIDFromEvent(event) != "" {
			wuiEvents = append(wuiEvents, event)
		}
	}

	return wuiEvents, nil
}

// createEvent creates a new calendar event from a task
func (s *SyncClient) createEvent(ctx context.Context, calendarID string, task core.Task) error {
	event := s.taskToEvent(task)

	_, err := s.calendarService.Events.Insert(calendarID, event).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	slog.Debug("Created event", "uuid", task.UUID, "description", task.Description)
	return nil
}

// updateEvent updates an existing calendar event from a task
func (s *SyncClient) updateEvent(ctx context.Context, calendarID string, task core.Task, existingEvent *calendar.Event) error {
	event := s.taskToEvent(task)
	event.Id = existingEvent.Id

	_, err := s.calendarService.Events.Update(calendarID, existingEvent.Id, event).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	slog.Debug("Updated event", "uuid", task.UUID, "description", task.Description)
	return nil
}

// deleteEvent deletes a calendar event
func (s *SyncClient) deleteEvent(ctx context.Context, calendarID string, eventID string) error {
	err := s.calendarService.Events.Delete(calendarID, eventID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	slog.Debug("Deleted event", "event_id", eventID)
	return nil
}

// taskToEvent converts a Taskwarrior task to a Google Calendar event
func (s *SyncClient) taskToEvent(task core.Task) *calendar.Event {
	// Add checkmark for completed tasks
	summary := task.Description
	if task.Status == "completed" {
		summary = "✓ " + summary
	}

	event := &calendar.Event{
		Summary: summary,
		Description: fmt.Sprintf("Taskwarrior UUID: %s\n\nProject: %s\nTags: %s\nStatus: %s",
			task.UUID,
			task.Project,
			strings.Join(task.Tags, ", "),
			task.Status,
		),
	}

	// Set event time based on due date or scheduled date
	// Note: Tasks without dates are filtered before reaching this function
	var eventTime time.Time
	if task.Due != nil && !task.Due.IsZero() {
		eventTime = *task.Due
	} else if task.Scheduled != nil && !task.Scheduled.IsZero() {
		eventTime = *task.Scheduled
	}

	// Check if the task has a specific time component (not midnight)
	hasTime := eventTime.Hour() != 0 || eventTime.Minute() != 0 || eventTime.Second() != 0

	if hasTime {
		// Create timed event with specific time
		event.Start = &calendar.EventDateTime{
			DateTime: eventTime.Format(time.RFC3339),
		}
		// Default to 15 minutes duration for timed events
		endTime := eventTime.Add(15 * time.Minute)
		event.End = &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
		}
	} else {
		// Create all-day event for tasks at midnight
		event.Start = &calendar.EventDateTime{
			Date: eventTime.Format("2006-01-02"),
		}
		event.End = &calendar.EventDateTime{
			Date: eventTime.Format("2006-01-02"),
		}
	}

	// Add color based on priority
	if task.Priority == "H" {
		event.ColorId = "11" // Red for high priority
	} else if task.Priority == "M" {
		event.ColorId = "5" // Yellow for medium priority
	}

	// Set notification based on scheduled field
	if task.Scheduled != nil && !task.Scheduled.IsZero() {
		var reminderMinutes int64
		var hasWarning bool

		if task.Due != nil && !task.Due.IsZero() {
			// Both scheduled and due exist
			timeDiff := task.Due.Sub(*task.Scheduled)

			if timeDiff > 0 {
				// Scheduled is before due - normal case
				// Set notification at scheduled time (time difference before the event)
				reminderMinutes = int64(timeDiff.Minutes())
			} else {
				// Scheduled is after or equal to due - warning case
				hasWarning = true
				// Still set a notification at a default time
				reminderMinutes = 15
			}
		} else {
			// Only scheduled exists (event time is based on scheduled)
			// Set a default notification
			reminderMinutes = 15
		}

		// Add warning to description if scheduled is after due
		if hasWarning {
			event.Description = event.Description + "\n\n⚠️ WARNING: Scheduled time is after due time!"
		}

		// Set the reminder
		// ForceSendFields ensures UseDefault:false is explicitly sent to override calendar defaults
		event.Reminders = &calendar.EventReminders{
			UseDefault:      false,
			ForceSendFields: []string{"UseDefault"},
			Overrides: []*calendar.EventReminder{
				{
					Method:  "popup",
					Minutes: reminderMinutes,
				},
			},
		}
	} else {
		// No scheduled field, use default calendar reminders
		event.Reminders = &calendar.EventReminders{
			UseDefault: true,
		}
	}

	return event
}

// shouldUpdateEvent checks if an event needs to be updated
func (s *SyncClient) shouldUpdateEvent(task core.Task, event *calendar.Event) bool {
	slog.Debug("Comparing event with task",
		"uuid", task.UUID,
		"task_desc", task.Description,
		"event_summary", event.Summary,
		"event_has_reminders", event.Reminders != nil,
		"event_reminders_use_default", event.Reminders != nil && event.Reminders.UseDefault)

	// Build expected summary (with checkmark if completed)
	expectedSummary := task.Description
	if task.Status == "completed" {
		expectedSummary = "✓ " + expectedSummary
	}

	// Check if summary (description) changed
	if event.Summary != expectedSummary {
		slog.Debug("Summary changed", "uuid", task.UUID, "expected", expectedSummary, "actual", event.Summary)
		return true
	}

	// Build expected description to check if it changed (including warning)
	expectedDescription := fmt.Sprintf("Taskwarrior UUID: %s\n\nProject: %s\nTags: %s\nStatus: %s",
		task.UUID,
		task.Project,
		strings.Join(task.Tags, ", "),
		task.Status,
	)

	// Add warning to expected description if scheduled is after due
	if task.Scheduled != nil && !task.Scheduled.IsZero() && task.Due != nil && !task.Due.IsZero() {
		if !task.Scheduled.Before(*task.Due) { // scheduled >= due
			expectedDescription = expectedDescription + "\n\n⚠️ WARNING: Scheduled time is after due time!"
		}
	}

	// Check if description changed
	if event.Description != expectedDescription {
		slog.Debug("Description changed",
			"uuid", task.UUID,
			"event_desc_len", len(event.Description),
			"expected_desc_len", len(expectedDescription))
		return true
	}

	// Check if the date or time changed
	// Note: Tasks without dates are filtered before reaching this function
	var taskTime time.Time
	if task.Due != nil && !task.Due.IsZero() {
		taskTime = *task.Due
	} else if task.Scheduled != nil && !task.Scheduled.IsZero() {
		taskTime = *task.Scheduled
	}

	slog.Debug("Checking date/time",
		"uuid", task.UUID,
		"task_time", taskTime,
		"task_due", task.Due,
		"task_scheduled", task.Scheduled)

	// Check if the task has a specific time component
	hasTime := taskTime.Hour() != 0 || taskTime.Minute() != 0 || taskTime.Second() != 0

	if event.Start != nil {
		if hasTime {
			// Compare DateTime for timed events
			if event.Start.DateTime != "" {
				eventStartTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
				if err == nil && !eventStartTime.Equal(taskTime) {
					return true
				}
			} else if event.Start.Date != "" {
				// Event is all-day but task has time, needs update
				return true
			}
		} else {
			// Compare Date for all-day events
			taskDate := taskTime.Format("2006-01-02")
			if event.Start.Date != "" {
				if event.Start.Date != taskDate {
					return true
				}
			} else if event.Start.DateTime != "" {
				// Event is timed but task is all-day, needs update
				return true
			}
		}
	}

	// Check if status changed by examining the event description
	if event.Description != "" {
		// Extract status from description
		statusPrefix := "\nStatus: "
		if idx := strings.Index(event.Description, statusPrefix); idx >= 0 {
			start := idx + len(statusPrefix)
			end := len(event.Description)
			// Find the end of the status line
			if newlineIdx := strings.Index(event.Description[start:], "\n"); newlineIdx >= 0 {
				end = start + newlineIdx
			}
			eventStatus := strings.TrimSpace(event.Description[start:end])
			if eventStatus != task.Status {
				return true
			}
		}
	}

	// Check if priority changed (affects color coding)
	expectedColorId := ""
	if task.Priority == "H" {
		expectedColorId = "11" // Red for high priority
	} else if task.Priority == "M" {
		expectedColorId = "5" // Yellow for medium priority
	}

	if expectedColorId != "" && event.ColorId != expectedColorId {
		return true
	}

	// Check if reminder/notification changed based on scheduled field
	taskHasScheduled := task.Scheduled != nil && !task.Scheduled.IsZero()
	eventHasCustomReminders := event.Reminders != nil && !event.Reminders.UseDefault && len(event.Reminders.Overrides) > 0

	slog.Debug("Checking reminders",
		"uuid", task.UUID,
		"task_has_scheduled", taskHasScheduled,
		"event_has_custom_reminders", eventHasCustomReminders,
		"event_reminders_nil", event.Reminders == nil,
		"event_reminders_use_default", event.Reminders != nil && event.Reminders.UseDefault)

	if taskHasScheduled {
		// Task has scheduled, so event should have custom reminders
		var expectedReminderMinutes int64

		if task.Due != nil && !task.Due.IsZero() {
			timeDiff := task.Due.Sub(*task.Scheduled)
			if timeDiff > 0 {
				expectedReminderMinutes = int64(timeDiff.Minutes())
			} else {
				expectedReminderMinutes = 15
			}
		} else {
			expectedReminderMinutes = 15
		}

		slog.Debug("Task has scheduled, checking reminders",
			"uuid", task.UUID,
			"expected_reminder_minutes", expectedReminderMinutes,
			"scheduled", task.Scheduled,
			"due", task.Due)

		// Check if event has custom reminders
		if !eventHasCustomReminders {
			slog.Debug("Event missing custom reminders", "uuid", task.UUID, "expected_minutes", expectedReminderMinutes)
			return true
		}

		// Check if the reminder minutes match
		actualReminderMinutes := event.Reminders.Overrides[0].Minutes
		slog.Debug("Comparing reminder minutes",
			"uuid", task.UUID,
			"expected", expectedReminderMinutes,
			"actual", actualReminderMinutes)
		if actualReminderMinutes != expectedReminderMinutes {
			slog.Debug("Reminder minutes mismatch", "uuid", task.UUID, "expected", expectedReminderMinutes, "actual", actualReminderMinutes)
			return true
		}
	} else {
		// Task has no scheduled time, event shouldn't have custom reminders
		if eventHasCustomReminders {
			slog.Debug("Event has unwanted custom reminders", "uuid", task.UUID)
			return true
		}
	}

	return false
}

// extractUUIDFromEvent extracts the Taskwarrior UUID from an event description
func extractUUIDFromEvent(event *calendar.Event) string {
	if event.Description == "" {
		return ""
	}

	// Look for "Taskwarrior UUID: " prefix
	prefix := "Taskwarrior UUID: "
	if idx := strings.Index(event.Description, prefix); idx >= 0 {
		start := idx + len(prefix)
		end := strings.Index(event.Description[start:], "\n")
		if end < 0 {
			return strings.TrimSpace(event.Description[start:])
		}
		return strings.TrimSpace(event.Description[start : start+end])
	}

	return ""
}
