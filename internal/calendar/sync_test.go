package calendar

import (
	"testing"
	"time"

	"github.com/clobrano/wui/internal/core"
	"google.golang.org/api/calendar/v3"
)

func timedTask(dur string) core.Task {
	due := time.Date(2026, 7, 8, 14, 0, 0, 0, time.Local)
	task := core.Task{
		UUID:        "test-uuid",
		Description: "Timed task",
		Status:      "pending",
		Due:         &due,
	}
	if dur != "" {
		task.UDAs = map[string]string{"dur": dur}
	}
	return task
}

func TestEventDuration(t *testing.T) {
	tests := []struct {
		name string
		dur  string
		want time.Duration
	}{
		{"no uda falls back to default", "", defaultEventDuration},
		{"iso duration", "PT30M", 30 * time.Minute},
		{"shorthand duration", "2h", 2 * time.Hour},
		{"invalid falls back to default", "not-a-duration", defaultEventDuration},
		{"zero falls back to default", "PT0S", defaultEventDuration},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := eventDuration(timedTask(tt.dur))
			if got != tt.want {
				t.Errorf("eventDuration(dur=%q) = %v, want %v", tt.dur, got, tt.want)
			}
		})
	}
}

func TestTaskToEventUsesDurUDA(t *testing.T) {
	s := &SyncClient{}

	// With a dur UDA, the event end should be start + dur.
	task := timedTask("PT45M")
	event := s.taskToEvent(task)

	start, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		t.Fatalf("failed to parse start: %v", err)
	}
	end, err := time.Parse(time.RFC3339, event.End.DateTime)
	if err != nil {
		t.Fatalf("failed to parse end: %v", err)
	}
	if got := end.Sub(start); got != 45*time.Minute {
		t.Errorf("event duration = %v, want %v", got, 45*time.Minute)
	}
}

func TestTaskToEventDefaultDuration(t *testing.T) {
	s := &SyncClient{}

	// Without a dur UDA, the default duration applies.
	event := s.taskToEvent(timedTask(""))

	start, _ := time.Parse(time.RFC3339, event.Start.DateTime)
	end, _ := time.Parse(time.RFC3339, event.End.DateTime)
	if got := end.Sub(start); got != defaultEventDuration {
		t.Errorf("event duration = %v, want %v", got, defaultEventDuration)
	}
}

func TestShouldUpdateEventOnDurationChange(t *testing.T) {
	s := &SyncClient{}

	// Build an event as it would exist with a 15-minute default duration.
	baseTask := timedTask("")
	existing := s.taskToEvent(baseTask)

	// The same task should not need an update.
	if s.shouldUpdateEvent(baseTask, existing) {
		t.Errorf("expected no update for unchanged task")
	}

	// Adding a dur UDA should now require an update to widen the event.
	changed := timedTask("PT1H")
	if !s.shouldUpdateEvent(changed, existing) {
		t.Errorf("expected update when 'dur' UDA changes the event duration")
	}

	// An event already matching the dur UDA should not need an update.
	matching := s.taskToEvent(changed)
	if s.shouldUpdateEvent(changed, matching) {
		t.Errorf("expected no update when event already matches 'dur' UDA")
	}
}

func TestTaskToEventMidnightAsTimedEvent(t *testing.T) {
	s := &SyncClient{}

	// Create a task with a due date at exactly midnight (no allDay UDA)
	midnight := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)
	task := core.Task{
		UUID:        "test-uuid",
		Description: "Midnight task",
		Status:      "pending",
		Due:         &midnight,
	}

	event := s.taskToEvent(task)

	// Should create a timed event (DateTime), not an all-day event (Date)
	if event.Start.DateTime == "" {
		t.Errorf("expected timed event (DateTime) for midnight task, got Date field instead")
	}
	if event.Start.Date != "" {
		t.Errorf("unexpected Date field for timed event")
	}

	// Parse and verify the exact time
	eventStartTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		t.Fatalf("failed to parse start time: %v", err)
	}
	if !eventStartTime.Equal(midnight) {
		t.Errorf("start time = %v, want %v", eventStartTime, midnight)
	}
}

func TestTaskToEventWithAllDayUDA(t *testing.T) {
	s := &SyncClient{}

	// Create a task with allDay:true UDA
	due := time.Date(2026, 3, 15, 14, 30, 0, 0, time.Local)
	task := core.Task{
		UUID:        "test-uuid",
		Description: "All-day meeting",
		Status:      "pending",
		Due:         &due,
		UDAs:        map[string]string{"allDay": "true"},
	}

	event := s.taskToEvent(task)

	// Should create an all-day event (Date), not a timed event
	if event.Start.Date == "" {
		t.Errorf("expected all-day event (Date) with allDay:true, got DateTime instead")
	}
	if event.Start.DateTime != "" {
		t.Errorf("unexpected DateTime field for all-day event")
	}

	// Verify the date is correct
	if event.Start.Date != "2026-03-15" {
		t.Errorf("start date = %q, want %q", event.Start.Date, "2026-03-15")
	}
}

func TestTaskToEventWithAllDayUDAVariations(t *testing.T) {
	s := &SyncClient{}
	due := time.Date(2026, 3, 15, 14, 30, 0, 0, time.Local)

	tests := []struct {
		name        string
		allDayValue string
		wantAllDay  bool
	}{
		{"allDay=true", "true", true},
		{"allDay=True", "True", true},
		{"allDay=TRUE", "TRUE", true},
		{"allDay=1", "1", true},
		{"allDay=yes", "yes", true},
		{"allDay=Yes", "Yes", true},
		{"allDay=false", "false", false},
		{"allDay=False", "False", false},
		{"allDay=0", "0", false},
		{"allDay=no", "no", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := core.Task{
				UUID:        "test-uuid",
				Description: "Test task",
				Status:      "pending",
				Due:         &due,
				UDAs:        map[string]string{"allDay": tt.allDayValue},
			}

			event := s.taskToEvent(task)

			hasDate := event.Start.Date != ""
			hasDateTime := event.Start.DateTime != ""

			if tt.wantAllDay && !hasDate {
				t.Errorf("expected all-day event (Date) for allDay=%q", tt.allDayValue)
			}
			if !tt.wantAllDay && !hasDateTime {
				t.Errorf("expected timed event (DateTime) for allDay=%q", tt.allDayValue)
			}
			if hasDate && hasDateTime {
				t.Errorf("event has both Date and DateTime")
			}
		})
	}
}

func TestShouldUpdateEventWhenAllDayChanges(t *testing.T) {
	s := &SyncClient{}

	// Create a timed task (no allDay UDA)
	due := time.Date(2026, 3, 15, 14, 0, 0, 0, time.Local)
	timedTask := core.Task{
		UUID:        "test-uuid",
		Description: "Meeting",
		Status:      "pending",
		Due:         &due,
	}

	// Create its corresponding event
	timedEvent := s.taskToEvent(timedTask)

	// Now create an all-day version of the same task
	allDayTask := core.Task{
		UUID:        "test-uuid",
		Description: "Meeting",
		Status:      "pending",
		Due:         &due,
		UDAs:        map[string]string{"allDay": "true"},
	}

	// The all-day task should detect the need for an update when compared to the timed event
	if !s.shouldUpdateEvent(allDayTask, timedEvent) {
		t.Errorf("expected update needed when switching from timed to all-day event")
	}

	// Conversely, the timed task should detect the need for an update when compared to an all-day event
	allDayEvent := s.taskToEvent(allDayTask)
	if !s.shouldUpdateEvent(timedTask, allDayEvent) {
		t.Errorf("expected update needed when switching from all-day to timed event")
	}

	// But if both are all-day, no update should be needed
	if s.shouldUpdateEvent(allDayTask, allDayEvent) {
		t.Errorf("expected no update when all-day task matches all-day event")
	}
}
