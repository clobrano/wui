package components

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clobrano/wui/internal/core"
)

// defaultSidebarStyles returns default styles for testing
func defaultSidebarStyles() SidebarStyles {
	return SidebarStyles{
		Border: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1),
		Title:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")),
		Label:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")),
		Value:          lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
		Dim:            lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		PriorityHigh:   lipgloss.Color("9"),
		PriorityMedium: lipgloss.Color("11"),
		PriorityLow:    lipgloss.Color("12"),
		DueOverdue:     lipgloss.Color("9"),
		StatusPending:  lipgloss.Color("11"),
		StatusDone:     lipgloss.Color("10"),
		StatusWaiting:  lipgloss.Color("12"),
		Tag:            lipgloss.Color("13"),
	}
}

func TestNewSidebar(t *testing.T) {
	sb := NewSidebar(40, 24, defaultSidebarStyles())

	if sb.width != 40 {
		t.Errorf("Expected width 40, got %d", sb.width)
	}
	if sb.height != 24 {
		t.Errorf("Expected height 24, got %d", sb.height)
	}
	if sb.task != nil {
		t.Error("Expected nil task")
	}
	if sb.offset != 0 {
		t.Errorf("Expected offset 0, got %d", sb.offset)
	}
}

func TestSetTask(t *testing.T) {
	sb := NewSidebar(40, 24, defaultSidebarStyles())
	task := &core.Task{
		ID:          1,
		UUID:        "test-uuid",
		Description: "Test task",
	}

	sb.SetTask(task)

	if sb.task == nil {
		t.Fatal("Expected task to be set")
	}
	if sb.task.ID != 1 {
		t.Errorf("Expected task ID 1, got %d", sb.task.ID)
	}
	if sb.offset != 0 {
		t.Errorf("Expected offset reset to 0, got %d", sb.offset)
	}
}

func TestSetSize(t *testing.T) {
	sb := NewSidebar(40, 24, defaultSidebarStyles())
	sb.SetSize(50, 30)

	if sb.width != 50 {
		t.Errorf("Expected width 50, got %d", sb.width)
	}
	if sb.height != 30 {
		t.Errorf("Expected height 30, got %d", sb.height)
	}
}

func TestSidebarViewEmpty(t *testing.T) {
	sb := NewSidebar(40, 24, defaultSidebarStyles())

	view := sb.View()

	if !strings.Contains(view, "No task selected") {
		t.Error("Expected 'No task selected' message")
	}
}

func TestViewWithTask(t *testing.T) {
	sb := NewSidebar(40, 24, defaultSidebarStyles())
	task := &core.Task{
		ID:          42,
		UUID:        "test-uuid-1234",
		Description: "Test task description",
		Project:     "TestProject",
		Status:      "pending",
		Priority:    "H",
	}

	sb.SetTask(task)
	view := sb.View()

	if !strings.Contains(view, "Task #42") {
		t.Error("Expected task number in view")
	}
	if !strings.Contains(view, "Test task description") {
		t.Error("Expected task description in view")
	}
	if !strings.Contains(view, "TestProject") {
		t.Error("Expected project name in view")
	}
}

func TestViewWithUUID(t *testing.T) {
	sb := NewSidebar(40, 24, defaultSidebarStyles())
	task := &core.Task{
		UUID:        "abc-123-def-456",
		Description: "Test task",
	}

	sb.SetTask(task)
	view := sb.View()

	if !strings.Contains(view, "abc-123-def-456") {
		t.Error("Expected UUID in view")
	}
}

func TestViewWithTags(t *testing.T) {
	sb := NewSidebar(40, 24, defaultSidebarStyles())
	task := &core.Task{
		ID:          1,
		Description: "Test task",
		Tags:        []string{"work", "urgent", "review"},
	}

	sb.SetTask(task)
	view := sb.View()

	if !strings.Contains(view, "work") {
		t.Error("Expected tag 'work' in view")
	}
	if !strings.Contains(view, "urgent") {
		t.Error("Expected tag 'urgent' in view")
	}
	if !strings.Contains(view, "review") {
		t.Error("Expected tag 'review' in view")
	}
}

func TestViewWithDates(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	sb := NewSidebar(40, 24, defaultSidebarStyles())
	task := &core.Task{
		ID:          1,
		Description: "Test task",
		Entry:       now,
		Due:         &tomorrow,
		Modified:    &yesterday,
	}

	sb.SetTask(task)
	view := sb.View()

	if !strings.Contains(view, "Dates") {
		t.Error("Expected 'Dates' section in view")
	}
	if !strings.Contains(view, "Due") {
		t.Error("Expected 'Due' date in view")
	}
	if !strings.Contains(view, "Created") {
		t.Error("Expected 'Created' date in view")
	}
}

func TestViewWithAnnotations(t *testing.T) {
	now := time.Now()
	sb := NewSidebar(40, 100, defaultSidebarStyles()) // Taller height so all content is visible
	task := &core.Task{
		ID:          1,
		Description: "Test task",
		Annotations: []core.Annotation{
			{Entry: now, Description: "First annotation"},
			{Entry: now.Add(-time.Hour), Description: "Second annotation"},
		},
	}

	sb.SetTask(task)
	view := sb.View()

	if !strings.Contains(view, "Annotations") {
		t.Errorf("Expected 'Annotations' section in view\nView:\n%s", view)
	}
	if !strings.Contains(view, "First annotation") {
		t.Errorf("Expected first annotation in view\nView:\n%s", view)
	}
	if !strings.Contains(view, "Second annotation") {
		t.Errorf("Expected second annotation in view\nView:\n%s", view)
	}
}

func TestViewWithDependencies(t *testing.T) {
	sb := NewSidebar(40, 24, defaultSidebarStyles())
	task := &core.Task{
		ID:          1,
		Description: "Test task",
		Depends:     []string{"uuid-1", "uuid-2"},
	}

	sb.SetTask(task)
	view := sb.View()

	if !strings.Contains(view, "Dependencies") {
		t.Error("Expected 'Dependencies' section in view")
	}
	if !strings.Contains(view, "Blocked by") {
		t.Error("Expected 'Blocked by' text in view")
	}
}

func TestViewWithUDAs(t *testing.T) {
	sb := NewSidebar(40, 24, defaultSidebarStyles())
	task := &core.Task{
		ID:          1,
		Description: "Test task",
		UDAs: map[string]string{
			"estimate": "2h",
			"sprint":   "sprint-1",
		},
	}

	sb.SetTask(task)
	view := sb.View()

	if !strings.Contains(view, "Custom Fields") {
		t.Error("Expected 'Custom Fields' section in view")
	}
}

func TestViewWithStatus(t *testing.T) {
	tests := []struct {
		status string
	}{
		{"pending"},
		{"completed"},
		{"deleted"},
		{"waiting"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			sb := NewSidebar(40, 24, defaultSidebarStyles())
			task := &core.Task{
				ID:          1,
				Description: "Test task",
				Status:      tt.status,
			}

			sb.SetTask(task)
			view := sb.View()

			if !strings.Contains(view, tt.status) {
				t.Errorf("Expected status '%s' in view", tt.status)
			}
		})
	}
}

func TestViewWithPriority(t *testing.T) {
	tests := []struct {
		priority string
	}{
		{"H"},
		{"M"},
		{"L"},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			sb := NewSidebar(40, 24, defaultSidebarStyles())
			task := &core.Task{
				ID:          1,
				Description: "Test task",
				Priority:    tt.priority,
			}

			sb.SetTask(task)
			view := sb.View()

			if !strings.Contains(view, tt.priority) {
				t.Errorf("Expected priority '%s' in view", tt.priority)
			}
		})
	}
}

func TestSidebarScrolling(t *testing.T) {
	sb := NewSidebar(40, 10, defaultSidebarStyles()) // Small height for testing scroll

	// Create task with lots of content
	annotations := make([]core.Annotation, 20)
	for i := 0; i < 20; i++ {
		annotations[i] = core.Annotation{
			Entry:       time.Now(),
			Description: "Annotation text",
		}
	}

	task := &core.Task{
		ID:          1,
		Description: "Test task",
		Annotations: annotations,
	}

	sb.SetTask(task)

	// Test scroll down
	initialOffset := sb.offset
	sb.scrollDown(5)

	if sb.offset <= initialOffset {
		t.Error("Expected offset to increase after scrolling down")
	}

	// Test scroll up
	currentOffset := sb.offset
	sb.scrollUp(3)

	if sb.offset >= currentOffset {
		t.Error("Expected offset to decrease after scrolling up")
	}

	// Test scroll doesn't go negative
	sb.scrollUp(1000)
	if sb.offset < 0 {
		t.Error("Expected offset to not go negative")
	}
}

func TestScrollingKeyboard(t *testing.T) {
	sb := NewSidebar(40, 10, defaultSidebarStyles())

	annotations := make([]core.Annotation, 20)
	for i := 0; i < 20; i++ {
		annotations[i] = core.Annotation{
			Entry:       time.Now(),
			Description: "Annotation",
		}
	}

	task := &core.Task{
		ID:          1,
		Description: "Test task",
		Annotations: annotations,
	}
	sb.SetTask(task)

	// Test ctrl+d (scroll down)
	initialOffset := sb.offset
	sb, _ = sb.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	if sb.offset <= initialOffset {
		t.Error("Expected offset to increase with ctrl+d")
	}

	// Test ctrl+u (scroll up)
	currentOffset := sb.offset
	sb, _ = sb.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	if sb.offset >= currentOffset {
		t.Error("Expected offset to decrease with ctrl+u")
	}
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		time            time.Time
		expected        string
		alternativeOk   bool
		alternativeText string
	}{
		{"just now", now, "just now", false, ""},
		{"1 minute ago", now.Add(-time.Minute), "1 minute ago", false, ""},
		{"5 minutes ago", now.Add(-5 * time.Minute), "5 minutes ago", false, ""},
		{"1 hour ago", now.Add(-time.Hour), "1 hour ago", false, ""},
		{"2 hours ago", now.Add(-2 * time.Hour), "2 hours ago", false, ""},
		{"yesterday", now.Add(-24 * time.Hour), "yesterday", true, "23 hours ago"},
		{"2 days ago", now.Add(-48 * time.Hour), "2 days ago", true, "yesterday"},
		{"tomorrow", now.Add(30 * time.Hour), "tomorrow", false, ""},
		{"in 2 days", now.Add(50 * time.Hour), "in 2 days", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRelativeTime(tt.time)
			if result != tt.expected {
				if tt.alternativeOk && strings.Contains(result, "hour") {
					// Allow hour-based responses near day boundaries
					return
				}
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected int // expected number of lines
	}{
		{
			name:     "short text",
			text:     "Hello world",
			width:    20,
			expected: 1,
		},
		{
			name:     "long text",
			text:     "This is a very long text that should be wrapped into multiple lines",
			width:    20,
			expected: 4,
		},
		{
			name:     "exact width",
			text:     "Exactly twenty chars",
			width:    20,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.text, tt.width)
			lines := strings.Split(result, "\n")
			if len(lines) != tt.expected {
				t.Errorf("Expected %d lines, got %d", tt.expected, len(lines))
			}
		})
	}
}

func TestViewUrgency(t *testing.T) {
	sb := NewSidebar(40, 24, defaultSidebarStyles())
	task := &core.Task{
		ID:          1,
		Description: "Test task",
		Urgency:     12.5,
	}

	sb.SetTask(task)
	view := sb.View()

	if !strings.Contains(view, "Urgency") {
		t.Error("Expected 'Urgency' field in view")
	}
	if !strings.Contains(view, "12.5") {
		t.Error("Expected urgency value in view")
	}
}

func TestOverdueHighlight(t *testing.T) {
	yesterday := time.Now().Add(-24 * time.Hour)
	sb := NewSidebar(40, 24, defaultSidebarStyles())
	task := &core.Task{
		ID:          1,
		Description: "Overdue task",
		Due:         &yesterday,
		Status:      "pending",
	}

	sb.SetTask(task)
	view := sb.View()

	// The view should contain the due date
	// Color coding is embedded in ANSI codes
	if !strings.Contains(view, "Due") {
		t.Error("Expected 'Due' in view for overdue task")
	}
}

func TestLongDescription(t *testing.T) {
	sb := NewSidebar(40, 24, defaultSidebarStyles())
	longDesc := strings.Repeat("This is a very long description that should be wrapped properly. ", 5)
	task := &core.Task{
		ID:          1,
		Description: longDesc,
	}

	sb.SetTask(task)
	view := sb.View()

	// Should not panic and should wrap text
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestDependencyLookup(t *testing.T) {
	sb := NewSidebar(40, 100, defaultSidebarStyles()) // Taller height so all content is visible

	// Create all tasks including the dependency
	allTasks := []core.Task{
		{
			ID:          1,
			UUID:        "dep-uuid-1234",
			Description: "Dependency task to complete first",
			Status:      "pending",
		},
		{
			ID:          2,
			UUID:        "main-uuid-5678",
			Description: "Main task that depends on task 1",
			Status:      "pending",
			Depends:     []string{"dep-uuid-1234"},
		},
	}

	// Set all tasks for dependency lookup
	sb.SetAllTasks(allTasks)

	// Set the main task as the current task
	sb.SetTask(&allTasks[1])

	// Render the view
	view := sb.View()

	// Verify the dependency section includes the task ID
	if !strings.Contains(view, "#1") {
		t.Error("Expected dependency task ID #1 in view")
	}

	// Verify the dependency description is shown (may be wrapped)
	if !strings.Contains(view, "Dependency task") {
		t.Errorf("Expected dependency description in view")
	}

	// Verify the "Blocked by" label is present
	if !strings.Contains(view, "Blocked by:") {
		t.Error("Expected 'Blocked by:' label in view")
	}
}

func TestDependencyNotFound(t *testing.T) {
	sb := NewSidebar(40, 100, defaultSidebarStyles()) // Taller height so all content is visible

	// Create task with dependency that doesn't exist in allTasks
	task := &core.Task{
		ID:          1,
		UUID:        "main-uuid",
		Description: "Task with missing dependency",
		Status:      "pending",
		Depends:     []string{"nonexistent-uuid"},
	}

	// Set empty allTasks
	sb.SetAllTasks([]core.Task{})
	sb.SetTask(task)

	// Render the view
	view := sb.View()

	// Should still show the UUID even if task not found
	if !strings.Contains(view, "nonexist") {
		t.Error("Expected dependency UUID in view even when task not found")
	}
}

func TestCompletedDependencyDisplay(t *testing.T) {
	sb := NewSidebar(40, 100, defaultSidebarStyles()) // Taller height so all content is visible

	// Create all tasks including completed and pending dependencies
	allTasks := []core.Task{
		{
			ID:          1,
			UUID:        "completed-dep-uuid",
			Description: "Completed dependency task",
			Status:      "completed",
		},
		{
			ID:          2,
			UUID:        "pending-dep-uuid",
			Description: "Pending dependency task",
			Status:      "pending",
		},
		{
			ID:          3,
			UUID:        "main-uuid",
			Description: "Main task with dependencies",
			Status:      "pending",
			Depends:     []string{"completed-dep-uuid", "pending-dep-uuid"},
		},
	}

	// Set all tasks for dependency lookup
	sb.SetAllTasks(allTasks)

	// Set the main task as the current task
	sb.SetTask(&allTasks[2])

	// Render the view
	view := sb.View()

	// Verify completed dependency shows with "x" prefix (no ID)
	if !strings.Contains(view, "x Completed dependency task") {
		t.Error("Expected completed dependency to show with 'x' prefix")
	}

	// Verify pending dependency shows with ID
	if !strings.Contains(view, "#2: Pending dependency task") {
		t.Error("Expected pending dependency to show with ID")
	}

	// Verify the "Blocked by" label is present
	if !strings.Contains(view, "Blocked by:") {
		t.Error("Expected 'Blocked by:' label in view")
	}
}
