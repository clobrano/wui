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
	tasks, err := s.taskClient.GetTasks(s.taskFilter)
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

	// Build map of existing events by task UUID
	eventMap := make(map[string]*calendar.Event)
	for _, event := range existingEvents {
		if uuid := extractUUIDFromEvent(event); uuid != "" {
			eventMap[uuid] = event
		}
	}

	// Sync each task
	created := 0
	updated := 0
	for _, task := range tasks {
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

	slog.Info("Sync completed", "created", created, "updated", updated)
	fmt.Printf("Sync completed: %d created, %d updated\n", created, updated)

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

// taskToEvent converts a Taskwarrior task to a Google Calendar event
func (s *SyncClient) taskToEvent(task core.Task) *calendar.Event {
	event := &calendar.Event{
		Summary: task.Description,
		Description: fmt.Sprintf("Taskwarrior UUID: %s\n\nProject: %s\nTags: %s\nStatus: %s",
			task.UUID,
			task.Project,
			strings.Join(task.Tags, ", "),
			task.Status,
		),
	}

	// Set event time based on due date or scheduled date
	var eventTime time.Time
	if !task.Due.IsZero() {
		eventTime = task.Due
	} else if !task.Scheduled.IsZero() {
		eventTime = task.Scheduled
	} else {
		// If no date is set, use today
		eventTime = time.Now()
	}

	// Create all-day event
	event.Start = &calendar.EventDateTime{
		Date: eventTime.Format("2006-01-02"),
	}
	event.End = &calendar.EventDateTime{
		Date: eventTime.Format("2006-01-02"),
	}

	// Add color based on priority or status
	if task.Status == "completed" {
		event.ColorId = "8" // Gray for completed
	} else if task.Priority == "H" {
		event.ColorId = "11" // Red for high priority
	} else if task.Priority == "M" {
		event.ColorId = "5" // Yellow for medium priority
	}

	return event
}

// shouldUpdateEvent checks if an event needs to be updated
func (s *SyncClient) shouldUpdateEvent(task core.Task, event *calendar.Event) bool {
	// Check if summary (description) changed
	if event.Summary != task.Description {
		return true
	}

	// Check if the date changed
	var taskDate string
	if !task.Due.IsZero() {
		taskDate = task.Due.Format("2006-01-02")
	} else if !task.Scheduled.IsZero() {
		taskDate = task.Scheduled.Format("2006-01-02")
	} else {
		taskDate = time.Now().Format("2006-01-02")
	}

	if event.Start != nil && event.Start.Date != taskDate {
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
