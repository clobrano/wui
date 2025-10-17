package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clobrano/wui/internal/core"
)

// TaskList is a component for displaying and navigating a list of tasks
type TaskList struct {
	tasks          []core.Task
	cursor         int // Selected task index
	width          int
	height         int
	displayColumns []string // Column names to display
	offset         int      // Scroll offset for viewport
}

// NewTaskList creates a new task list component
func NewTaskList(width, height int) TaskList {
	return TaskList{
		tasks:          []core.Task{},
		cursor:         0,
		width:          width,
		height:         height,
		displayColumns: []string{"ID", "PROJECT", "P", "DUE", "DESCRIPTION"},
		offset:         0,
	}
}

// SetTasks updates the task list
func (t *TaskList) SetTasks(tasks []core.Task) {
	t.tasks = tasks
	// Reset cursor if out of bounds
	if t.cursor >= len(t.tasks) {
		t.cursor = 0
	}
	t.updateScroll()
}

// SetSize updates the component dimensions
func (t *TaskList) SetSize(width, height int) {
	t.width = width
	t.height = height
	t.updateScroll()
}

// Update handles messages for the task list
func (t TaskList) Update(msg tea.Msg) (TaskList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return t.handleKey(msg), nil
	}
	return t, nil
}

// handleKey processes keyboard input
func (t TaskList) handleKey(msg tea.KeyMsg) TaskList {
	switch msg.String() {
	case "j", "down":
		t.moveDown()
	case "k", "up":
		t.moveUp()
	case "g":
		t.moveToStart()
	case "G":
		t.moveToEnd()
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Quick jump to visible task by number
		t.quickJump(msg.String())
	}
	return t
}

// moveDown moves cursor down one position
func (t *TaskList) moveDown() {
	if len(t.tasks) == 0 {
		return
	}
	if t.cursor < len(t.tasks)-1 {
		t.cursor++
		t.updateScroll()
	}
}

// moveUp moves cursor up one position
func (t *TaskList) moveUp() {
	if t.cursor > 0 {
		t.cursor--
		t.updateScroll()
	}
}

// moveToStart jumps to first task
func (t *TaskList) moveToStart() {
	if len(t.tasks) > 0 {
		t.cursor = 0
		t.updateScroll()
	}
}

// moveToEnd jumps to last task
func (t *TaskList) moveToEnd() {
	if len(t.tasks) > 0 {
		t.cursor = len(t.tasks) - 1
		t.updateScroll()
	}
}

// quickJump jumps to a visible task by number (1-9)
func (t *TaskList) quickJump(key string) {
	num := int(key[0] - '0') // Convert '1'-'9' to 1-9
	targetIndex := t.offset + num - 1
	if targetIndex >= 0 && targetIndex < len(t.tasks) && targetIndex < t.offset+t.height-1 {
		t.cursor = targetIndex
	}
}

// updateScroll adjusts the scroll offset to keep cursor visible
func (t *TaskList) updateScroll() {
	if len(t.tasks) == 0 {
		t.offset = 0
		return
	}

	visibleHeight := t.height - 1 // Subtract 1 for header row

	// Cursor is above viewport
	if t.cursor < t.offset {
		t.offset = t.cursor
	}

	// Cursor is below viewport
	if t.cursor >= t.offset+visibleHeight {
		t.offset = t.cursor - visibleHeight + 1
	}

	// Don't scroll past the end
	maxOffset := len(t.tasks) - visibleHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if t.offset > maxOffset {
		t.offset = maxOffset
	}
}

// View renders the task list
func (t TaskList) View() string {
	if len(t.tasks) == 0 {
		return lipgloss.NewStyle().
			Padding(2, 4).
			Render("No tasks found.")
	}

	var lines []string

	// Render column headers
	lines = append(lines, t.renderHeader())

	// Calculate visible range
	visibleHeight := t.height - 1 // Subtract header
	endIdx := t.offset + visibleHeight
	if endIdx > len(t.tasks) {
		endIdx = len(t.tasks)
	}

	// Render visible tasks
	for i := t.offset; i < endIdx; i++ {
		task := t.tasks[i]
		isSelected := i == t.cursor
		quickJump := ""
		if i-t.offset < 9 {
			quickJump = fmt.Sprintf("%d", i-t.offset+1)
		}
		line := t.renderTaskLine(task, isSelected, quickJump)
		lines = append(lines, line)
	}

	// Fill remaining space
	for i := len(lines) - 1; i < t.height; i++ {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderHeader renders the column header row
func (t TaskList) renderHeader() string {
	cols := t.calculateColumnWidths()

	header := fmt.Sprintf("  %-*s %-*s %s %-*s %s",
		cols.id, "ID",
		cols.project, "PROJECT",
		"P",
		cols.due, "DUE",
		"DESCRIPTION",
	)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("8")).
		Width(t.width)

	return headerStyle.Render(header)
}

// columnWidths holds calculated column widths
type columnWidths struct {
	id          int
	project     int
	priority    int
	due         int
	description int
}

// calculateColumnWidths determines column widths based on available space
func (t TaskList) calculateColumnWidths() columnWidths {
	const (
		cursorWidth   = 2  // "■ " or "1 "
		idWidth       = 4  // Sequential ID
		priorityWidth = 1  // H/M/L
		dueWidth      = 10 // Date format
		minProject    = 10
		minDesc       = 20
		spacing       = 6
	)

	fixedWidth := cursorWidth + idWidth + priorityWidth + dueWidth + spacing
	remainingWidth := t.width - fixedWidth

	if remainingWidth < minProject+minDesc {
		return columnWidths{
			id:          idWidth,
			project:     minProject,
			priority:    priorityWidth,
			due:         dueWidth,
			description: minDesc,
		}
	}

	// Allocate 25% to project, 75% to description
	projectWidth := remainingWidth / 4
	if projectWidth < minProject {
		projectWidth = minProject
	}
	if projectWidth > 20 {
		projectWidth = 20
	}

	descWidth := remainingWidth - projectWidth

	return columnWidths{
		id:          idWidth,
		project:     projectWidth,
		priority:    priorityWidth,
		due:         dueWidth,
		description: descWidth,
	}
}

// renderTaskLine renders a single task row
func (t TaskList) renderTaskLine(task core.Task, isSelected bool, quickJump string) string {
	cols := t.calculateColumnWidths()

	// Cursor or quick jump number
	cursor := " "
	if isSelected {
		cursor = "■"
	} else if quickJump != "" {
		cursor = quickJump
	}

	// ID
	id := fmt.Sprintf("%d", task.ID)
	if task.ID == 0 {
		id = task.UUID
		if len(id) > cols.id {
			id = id[:cols.id]
		}
	}

	// Project
	project := task.Project
	if project == "" {
		project = "-"
	}
	if len(project) > cols.project {
		project = project[:cols.project-3] + "..."
	}

	// Priority with color coding
	priority := "-"
	priorityStyle := lipgloss.NewStyle()
	if task.Priority != "" {
		priority = string(task.Priority[0])
		switch task.Priority {
		case "H":
			priorityStyle = priorityStyle.Foreground(lipgloss.Color("9")) // Red
		case "M":
			priorityStyle = priorityStyle.Foreground(lipgloss.Color("11")) // Yellow
		case "L":
			priorityStyle = priorityStyle.Foreground(lipgloss.Color("12")) // Blue
		}
	}

	// Due date with color coding
	due := "-"
	dueStyle := lipgloss.NewStyle()
	if task.Due != nil {
		due = task.FormatDueDate()
		if len(due) > cols.due {
			due = due[:cols.due]
		}
		if task.IsOverdue() {
			dueStyle = dueStyle.Foreground(lipgloss.Color("9")) // Red - overdue
		}
		// TODO: Add "today" and "soon" color coding
	}

	// Description
	description := task.Description
	if len(description) > cols.description {
		description = description[:cols.description-3] + "..."
	}

	line := fmt.Sprintf("%s %-*s %-*s %s %-*s %s",
		cursor,
		cols.id, id,
		cols.project, project,
		priorityStyle.Render(priority),
		cols.due, dueStyle.Render(due),
		description,
	)

	if isSelected {
		style := lipgloss.NewStyle().
			Background(lipgloss.Color("12")).
			Foreground(lipgloss.Color("0")).
			Width(t.width)
		return style.Render(line)
	}

	// Ensure non-selected lines also use full width
	normalStyle := lipgloss.NewStyle().Width(t.width)
	return normalStyle.Render(line)
}

// SelectedTask returns the currently selected task, or nil if no tasks
func (t TaskList) SelectedTask() *core.Task {
	if len(t.tasks) == 0 || t.cursor < 0 || t.cursor >= len(t.tasks) {
		return nil
	}
	return &t.tasks[t.cursor]
}

// SelectedIndex returns the index of the currently selected task
func (t TaskList) SelectedIndex() int {
	return t.cursor
}

// TaskCount returns the total number of tasks
func (t TaskList) TaskCount() int {
	return len(t.tasks)
}
