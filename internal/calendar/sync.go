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

// Sync performs the synchronization from Taskwarrior to Google Calendar
func (s *SyncClient) Sync(ctx context.Context) error {
	slog.Info("Starting sync", "calendar", s.calendarName, "filter", s.taskFilter)

	// Get the calendar ID by name
	calendarID, err := s.findCalendarByName(ctx, s.calendarName)
	if err != nil {
		return fmt.Errorf("failed to find calendar: %w", err)
	}

	slog.Info("Found calendar", "id", calendarID, "name", s.calendarName)

	// Get tasks from Taskwarrior
	tasks, err := s.taskClient.Export(s.taskFilter)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	slog.Info("Retrieved tasks", "count", len(tasks))

	// Get existing events from calendar
	existingEvents, err := s.getCalendarEvents(ctx, calendarID)
	if err != nil {
		return fmt.Errorf("failed to get calendar events: %w", err)
	}

	slog.Info("Retrieved existing calendar events", "count", len(existingEvents))

	// Separate events into task-originated and manual events
	eventMap := make(map[string]*calendar.Event)
	var manualEvents []*calendar.Event
	for _, event := range existingEvents {
		if uuid := extractUUIDFromEvent(event); uuid != "" {
			eventMap[uuid] = event
		} else {
			manualEvents = append(manualEvents, event)
		}
	}

	slog.Info("Event breakdown", "task-originated", len(eventMap), "manual", len(manualEvents))

	// First, create tasks from manual calendar events
	tasksCreatedFromEvents := 0
	for _, event := range manualEvents {
		uuid, err := s.createTaskFromEvent(ctx, calendarID, event)
		if err != nil {
			slog.Error("Failed to create task from event", "eventId", event.Id, "summary", event.Summary, "error", err)
			continue
		}
		if uuid != "" {
			// Add the updated event to eventMap to prevent duplicate creation
			eventMap[uuid] = event
			tasksCreatedFromEvents++
		}
	}

	// Now sync tasks to calendar
	created := 0
	updated := 0
	skipped := 0
	for _, task := range tasks {
		// Skip tasks without due date or scheduled date
		if (task.Due == nil || task.Due.IsZero()) && (task.Scheduled == nil || task.Scheduled.IsZero()) {
			slog.Debug("Skipping task without due date", "uuid", task.UUID, "description", task.Description)
			skipped++
			continue
		}

		if existingEvent, exists := eventMap[task.UUID]; exists {
			// Update existing event
			if s.shouldUpdateEvent(task, existingEvent) {
				if err := s.updateEvent(ctx, calendarID, task, existingEvent); err != nil {
					slog.Error("Failed to update event", "uuid", task.UUID, "error", err)
					continue
				}
				updated++
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

	slog.Info("Sync completed", "total", len(tasks), "created", created, "updated", updated, "skipped", skipped, "tasksFromEvents", tasksCreatedFromEvents)
	fmt.Printf("Sync completed: %d tasks, %d events created, %d events updated, %d skipped (no due date), %d tasks created from calendar\n", len(tasks), created, updated, skipped, tasksCreatedFromEvents)

	return nil
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

// getCalendarEvents retrieves all events from the calendar
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
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	return events.Items, nil
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

	// Check if the time has specific hour/minute (not just midnight)
	// If it's exactly midnight (00:00:00), treat it as an all-day event
	if eventTime.Hour() == 0 && eventTime.Minute() == 0 && eventTime.Second() == 0 {
		// Create all-day event
		event.Start = &calendar.EventDateTime{
			Date: eventTime.Format("2006-01-02"),
		}
		event.End = &calendar.EventDateTime{
			Date: eventTime.Format("2006-01-02"),
		}
	} else {
		// Create timed event with 1-hour duration
		// (Google Calendar requires end time != start time for timed events)
		endTime := eventTime.Add(time.Hour)
		event.Start = &calendar.EventDateTime{
			DateTime: eventTime.Format(time.RFC3339),
		}
		event.End = &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
		}
	}

	// Add color based on priority
	if task.Priority == "H" {
		event.ColorId = "11" // Red for high priority
	} else if task.Priority == "M" {
		event.ColorId = "5" // Yellow for medium priority
	}

	return event
}

// shouldUpdateEvent checks if an event needs to be updated
func (s *SyncClient) shouldUpdateEvent(task core.Task, event *calendar.Event) bool {
	// Build expected summary (with checkmark if completed)
	expectedSummary := task.Description
	if task.Status == "completed" {
		expectedSummary = "✓ " + expectedSummary
	}

	// Check if summary (description) changed
	if event.Summary != expectedSummary {
		return true
	}

	// Check if the date changed
	// Note: Tasks without dates are filtered before reaching this function
	var taskDate string
	if task.Due != nil && !task.Due.IsZero() {
		taskDate = task.Due.Format("2006-01-02")
	} else if task.Scheduled != nil && !task.Scheduled.IsZero() {
		taskDate = task.Scheduled.Format("2006-01-02")
	}

	if event.Start != nil && taskDate != "" && event.Start.Date != taskDate {
		return true
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

// parseEventMetadata extracts project and tags from an event description
// Expected format:
//   Project:<project name>
//   Tags: <tag1>, <tag2>
func parseEventMetadata(description string) (project string, tags []string) {
	if description == "" {
		return "", nil
	}

	lines := strings.Split(description, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse project
		if strings.HasPrefix(strings.ToLower(line), "project:") {
			project = strings.TrimSpace(line[8:]) // len("Project:") = 8
			continue
		}

		// Parse tags
		if strings.HasPrefix(strings.ToLower(line), "tags:") {
			tagStr := strings.TrimSpace(line[5:]) // len("Tags:") = 5
			if tagStr != "" {
				// Split by comma and trim each tag
				rawTags := strings.Split(tagStr, ",")
				for _, tag := range rawTags {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						tags = append(tags, tag)
					}
				}
			}
		}
	}

	return project, tags
}

// createTaskFromEvent creates a Taskwarrior task from a calendar event
// Returns the UUID of the created task, or empty string if no task was created
func (s *SyncClient) createTaskFromEvent(ctx context.Context, calendarID string, event *calendar.Event) (string, error) {
	if event.Summary == "" {
		slog.Debug("Skipping event without summary", "id", event.Id)
		return "", nil
	}

	// Parse metadata from event description
	project, tags := parseEventMetadata(event.Description)

	// Build task description with project and tags
	taskDesc := event.Summary

	// Add project if specified
	if project != "" {
		taskDesc = fmt.Sprintf("%s project:%s", taskDesc, project)
	}

	// Add tags if specified
	for _, tag := range tags {
		taskDesc = fmt.Sprintf("%s +%s", taskDesc, tag)
	}

	// Extract date/time from event
	var dueDate string
	if event.Start != nil {
		if event.Start.Date != "" {
			// All-day event - use date only
			dueDate = event.Start.Date
		} else if event.Start.DateTime != "" {
			// Timed event - preserve time information
			t, err := time.Parse(time.RFC3339, event.Start.DateTime)
			if err == nil {
				// Format as ISO 8601 datetime for Taskwarrior
				dueDate = t.Format("2006-01-02T15:04:05")
			}
		}
	}

	// Add due date/time if available
	if dueDate != "" {
		taskDesc = fmt.Sprintf("%s due:%s", taskDesc, dueDate)
	}

	// Create the task
	uuid, err := s.taskClient.Add(taskDesc)
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	slog.Info("Created task from calendar event", "uuid", uuid, "description", event.Summary)

	// Update the event description to include the UUID
	updatedDescription := fmt.Sprintf("Taskwarrior UUID: %s\n\n", uuid)
	if event.Description != "" {
		updatedDescription += event.Description
	} else {
		// Add project and tags info if they were parsed
		if project != "" {
			updatedDescription += fmt.Sprintf("Project: %s\n", project)
		}
		if len(tags) > 0 {
			updatedDescription += fmt.Sprintf("Tags: %s\n", strings.Join(tags, ", "))
		}
		updatedDescription += "Status: pending"
	}

	event.Description = updatedDescription

	// Update the event in the calendar
	_, err = s.calendarService.Events.Update(calendarID, event.Id, event).Context(ctx).Do()
	if err != nil {
		slog.Warn("Failed to update event with UUID", "error", err, "eventId", event.Id)
		// Don't return error - task was created successfully
	}

	return uuid, nil
}
