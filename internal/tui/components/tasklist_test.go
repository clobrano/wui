package components

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/core"
)

// defaultTaskListStyles returns default styles for testing
func defaultTaskListStyles() TaskListStyles {
	return TaskListStyles{
		Header:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
		Separator:       lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		Selection:       lipgloss.NewStyle().Background(lipgloss.Color("12")).Foreground(lipgloss.Color("0")),
		PriorityHigh:    lipgloss.Color("9"),
		PriorityMedium:  lipgloss.Color("11"),
		PriorityLow:     lipgloss.Color("12"),
		DueOverdue:      lipgloss.Color("9"),
		TagColor:        lipgloss.Color("14"),
		StatusCompleted: lipgloss.Color("8"),
		StatusWaiting:   lipgloss.Color("8"),
		StatusActive:    lipgloss.Color("15"),
	}
}

// testColumns converts a simple string array to config.Columns for testing
func testColumns(names ...string) config.Columns {
	defaultLabels := map[string]string{
		"id":          "ID",
		"project":     "PROJECT",
		"priority":    "P",
		"due":         "DUE",
		"tags":        "TAGS",
		"annotation":  "A",
		"dependency":  "D",
		"description": "DESCRIPTION",
	}

	cols := make(config.Columns, len(names))
	for i, name := range names {
		label := defaultLabels[name]
		if label == "" {
			label = name
		}
		cols[i] = config.Column{Name: name, Label: label}
	}
	return cols
}

func TestNewTaskList(t *testing.T) {
	tl := NewTaskList(80, 24, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())

	if tl.width != 80 {
		t.Errorf("Expected width 80, got %d", tl.width)
	}
	if tl.height != 24 {
		t.Errorf("Expected height 24, got %d", tl.height)
	}
	if tl.cursor != 0 {
		t.Errorf("Expected cursor 0, got %d", tl.cursor)
	}
	if len(tl.tasks) != 0 {
		t.Errorf("Expected empty tasks, got %d tasks", len(tl.tasks))
	}
}

func TestSetTasks(t *testing.T) {
	tl := NewTaskList(80, 24, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tasks := []core.Task{
		{ID: 1, UUID: "uuid-1", Description: "Task 1"},
		{ID: 2, UUID: "uuid-2", Description: "Task 2"},
		{ID: 3, UUID: "uuid-3", Description: "Task 3"},
	}

	tl.SetTasks(tasks)

	if len(tl.tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tl.tasks))
	}
	if tl.cursor != 0 {
		t.Errorf("Expected cursor reset to 0, got %d", tl.cursor)
	}
}

func TestSetTasksResetsCursor(t *testing.T) {
	tl := NewTaskList(80, 24, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tasks := []core.Task{
		{ID: 1, UUID: "uuid-1", Description: "Task 1"},
		{ID: 2, UUID: "uuid-2", Description: "Task 2"},
	}

	tl.SetTasks(tasks)
	tl.cursor = 5 // Set cursor beyond new task list

	tl.SetTasks([]core.Task{{ID: 1, UUID: "uuid-1", Description: "Task 1"}})

	if tl.cursor != 0 {
		t.Errorf("Expected cursor reset to 0, got %d", tl.cursor)
	}
}

func TestNavigationDown(t *testing.T) {
	tl := NewTaskList(80, 24, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tl.SetTasks([]core.Task{
		{ID: 1, Description: "Task 1"},
		{ID: 2, Description: "Task 2"},
		{ID: 3, Description: "Task 3"},
	})

	// Move down
	tl.moveDown()
	if tl.cursor != 1 {
		t.Errorf("Expected cursor 1, got %d", tl.cursor)
	}

	tl.moveDown()
	if tl.cursor != 2 {
		t.Errorf("Expected cursor 2, got %d", tl.cursor)
	}

	// Try to move past end
	tl.moveDown()
	if tl.cursor != 2 {
		t.Errorf("Expected cursor to stay at 2, got %d", tl.cursor)
	}
}

func TestNavigationUp(t *testing.T) {
	tl := NewTaskList(80, 24, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tl.SetTasks([]core.Task{
		{ID: 1, Description: "Task 1"},
		{ID: 2, Description: "Task 2"},
		{ID: 3, Description: "Task 3"},
	})

	tl.cursor = 2

	// Move up
	tl.moveUp()
	if tl.cursor != 1 {
		t.Errorf("Expected cursor 1, got %d", tl.cursor)
	}

	tl.moveUp()
	if tl.cursor != 0 {
		t.Errorf("Expected cursor 0, got %d", tl.cursor)
	}

	// Try to move before start
	tl.moveUp()
	if tl.cursor != 0 {
		t.Errorf("Expected cursor to stay at 0, got %d", tl.cursor)
	}
}

func TestNavigationJumpToStart(t *testing.T) {
	tl := NewTaskList(80, 24, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tl.SetTasks([]core.Task{
		{ID: 1, Description: "Task 1"},
		{ID: 2, Description: "Task 2"},
		{ID: 3, Description: "Task 3"},
	})

	tl.cursor = 2
	tl.moveToStart()

	if tl.cursor != 0 {
		t.Errorf("Expected cursor 0, got %d", tl.cursor)
	}
}

func TestNavigationJumpToEnd(t *testing.T) {
	tl := NewTaskList(80, 24, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tl.SetTasks([]core.Task{
		{ID: 1, Description: "Task 1"},
		{ID: 2, Description: "Task 2"},
		{ID: 3, Description: "Task 3"},
	})

	tl.moveToEnd()

	if tl.cursor != 2 {
		t.Errorf("Expected cursor 2, got %d", tl.cursor)
	}
}

func TestKeyboardNavigation(t *testing.T) {
	tl := NewTaskList(80, 24, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tl.SetTasks([]core.Task{
		{ID: 1, Description: "Task 1"},
		{ID: 2, Description: "Task 2"},
		{ID: 3, Description: "Task 3"},
	})

	// Test 'j' (down)
	tl, _ = tl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if tl.cursor != 1 {
		t.Errorf("Expected cursor 1 after 'j', got %d", tl.cursor)
	}

	// Test 'k' (up)
	tl, _ = tl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if tl.cursor != 0 {
		t.Errorf("Expected cursor 0 after 'k', got %d", tl.cursor)
	}

	// Test 'G' (end)
	tl, _ = tl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if tl.cursor != 2 {
		t.Errorf("Expected cursor 2 after 'G', got %d", tl.cursor)
	}

	// Test 'g' (start)
	tl, _ = tl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if tl.cursor != 0 {
		t.Errorf("Expected cursor 0 after 'g', got %d", tl.cursor)
	}
}

func TestQuickJump(t *testing.T) {
	tl := NewTaskList(80, 10, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tasks := make([]core.Task, 20)
	for i := 0; i < 20; i++ {
		tasks[i] = core.Task{ID: i + 1, Description: "Task"}
	}
	tl.SetTasks(tasks)

	// Jump to position 3 (task index 2)
	tl.quickJump("3")
	if tl.cursor != 2 {
		t.Errorf("Expected cursor 2, got %d", tl.cursor)
	}

	// Jump to position 5 (task index 4)
	tl.quickJump("5")
	if tl.cursor != 4 {
		t.Errorf("Expected cursor 4, got %d", tl.cursor)
	}
}

func TestScrolling(t *testing.T) {
	tl := NewTaskList(80, 5, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles()) // Small height to test scrolling
	tasks := make([]core.Task, 20)
	for i := 0; i < 20; i++ {
		tasks[i] = core.Task{ID: i + 1, Description: "Task"}
	}
	tl.SetTasks(tasks)

	// Initially offset should be 0
	if tl.offset != 0 {
		t.Errorf("Expected initial offset 0, got %d", tl.offset)
	}

	// Move down several times
	for i := 0; i < 10; i++ {
		tl.moveDown()
	}

	// Offset should have adjusted to keep cursor visible
	if tl.offset == 0 {
		t.Error("Expected offset to have scrolled")
	}
	if tl.cursor < tl.offset || tl.cursor >= tl.offset+tl.height-1 {
		t.Errorf("Cursor %d is not within visible range [%d, %d)", tl.cursor, tl.offset, tl.offset+tl.height-1)
	}
}

func TestSelectedTask(t *testing.T) {
	tl := NewTaskList(80, 24, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tl.SetTasks([]core.Task{
		{ID: 1, UUID: "uuid-1", Description: "Task 1"},
		{ID: 2, UUID: "uuid-2", Description: "Task 2"},
	})

	task := tl.SelectedTask()
	if task == nil {
		t.Fatal("Expected selected task, got nil")
	}
	if task.ID != 1 {
		t.Errorf("Expected task ID 1, got %d", task.ID)
	}

	tl.moveDown()
	task = tl.SelectedTask()
	if task.ID != 2 {
		t.Errorf("Expected task ID 2, got %d", task.ID)
	}
}

func TestSelectedTaskEmpty(t *testing.T) {
	tl := NewTaskList(80, 24, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())

	task := tl.SelectedTask()
	if task != nil {
		t.Error("Expected nil for empty task list")
	}
}

func TestView(t *testing.T) {
	tl := NewTaskList(80, 10, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tl.SetTasks([]core.Task{
		{ID: 1, Description: "Test task 1", Project: "Project1"},
		{ID: 2, Description: "Test task 2", Project: "Project2"},
	})

	view := tl.View()

	if !strings.Contains(view, "Test task 1") {
		t.Error("Expected view to contain task description")
	}
	if !strings.Contains(view, "ID") {
		t.Error("Expected view to contain column header")
	}
}

func TestViewEmpty(t *testing.T) {
	tl := NewTaskList(80, 10, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())

	view := tl.View()

	if !strings.Contains(view, "No tasks") {
		t.Error("Expected view to show 'No tasks' message")
	}
}

func TestPriorityColorCoding(t *testing.T) {
	tl := NewTaskList(80, 10, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tl.SetTasks([]core.Task{
		{ID: 1, Description: "High priority", Priority: "H"},
		{ID: 2, Description: "Medium priority", Priority: "M"},
		{ID: 3, Description: "Low priority", Priority: "L"},
	})

	view := tl.View()

	// Just verify it renders without error
	// Color codes are embedded in the output
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestDueDateColorCoding(t *testing.T) {
	yesterday := time.Now().Add(-24 * time.Hour)
	tomorrow := time.Now().Add(24 * time.Hour)

	tl := NewTaskList(80, 10, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tl.SetTasks([]core.Task{
		{ID: 1, Description: "Overdue task", Due: &yesterday, Status: "pending"},
		{ID: 2, Description: "Future task", Due: &tomorrow, Status: "pending"},
	})

	view := tl.View()

	// Just verify it renders without error
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestTaskCount(t *testing.T) {
	tl := NewTaskList(80, 24, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tl.SetTasks([]core.Task{
		{ID: 1, Description: "Task 1"},
		{ID: 2, Description: "Task 2"},
		{ID: 3, Description: "Task 3"},
	})

	if tl.TaskCount() != 3 {
		t.Errorf("Expected task count 3, got %d", tl.TaskCount())
	}
}

func TestSelectedIndex(t *testing.T) {
	tl := NewTaskList(80, 24, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())
	tl.SetTasks([]core.Task{
		{ID: 1, Description: "Task 1"},
		{ID: 2, Description: "Task 2"},
	})

	if tl.SelectedIndex() != 0 {
		t.Errorf("Expected selected index 0, got %d", tl.SelectedIndex())
	}

	tl.moveDown()
	if tl.SelectedIndex() != 1 {
		t.Errorf("Expected selected index 1, got %d", tl.SelectedIndex())
	}
}

func TestCompletedTasksSortedToBottom(t *testing.T) {
	tl := NewTaskList(80, 24, testColumns("id", "project", "description", "due", "priority"), defaultTaskListStyles())

	// Create a mixed list of tasks with completed tasks in the middle
	tasks := []core.Task{
		{ID: 1, UUID: "uuid-1", Description: "Pending Task 1", Status: "pending"},
		{ID: 2, UUID: "uuid-2", Description: "Completed Task 1", Status: "completed"},
		{ID: 3, UUID: "uuid-3", Description: "Pending Task 2", Status: "pending"},
		{ID: 4, UUID: "uuid-4", Description: "Completed Task 2", Status: "completed"},
		{ID: 5, UUID: "uuid-5", Description: "Pending Task 3", Status: "pending"},
	}

	tl.SetTasks(tasks)

	// Verify that pending tasks come first
	if tl.tasks[0].Status == "completed" {
		t.Errorf("Expected first task to be pending, got %s", tl.tasks[0].Status)
	}
	if tl.tasks[1].Status == "completed" {
		t.Errorf("Expected second task to be pending, got %s", tl.tasks[1].Status)
	}
	if tl.tasks[2].Status == "completed" {
		t.Errorf("Expected third task to be pending, got %s", tl.tasks[2].Status)
	}

	// Verify that completed tasks come last
	if tl.tasks[3].Status != "completed" {
		t.Errorf("Expected fourth task to be completed, got %s", tl.tasks[3].Status)
	}
	if tl.tasks[4].Status != "completed" {
		t.Errorf("Expected fifth task to be completed, got %s", tl.tasks[4].Status)
	}

	// Verify total count
	if len(tl.tasks) != 5 {
		t.Errorf("Expected 5 tasks, got %d", len(tl.tasks))
	}
}
