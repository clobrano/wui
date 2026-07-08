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
