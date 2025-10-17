package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clobrano/wui/internal/core"
)

// DisplayMode indicates what the task list is displaying
type DisplayMode int

const (
	DisplayModeTasks DisplayMode = iota
	DisplayModeGroups
)

// TaskListStyles holds the styles needed for rendering the task list
type TaskListStyles struct {
	Header           lipgloss.Style
	Separator        lipgloss.Style
	Selection        lipgloss.Style
	PriorityHigh     lipgloss.Color
	PriorityMedium   lipgloss.Color
	PriorityLow      lipgloss.Color
	DueOverdue       lipgloss.Color
	TagColor         lipgloss.Color
	StatusCompleted  lipgloss.Color
	StatusWaiting    lipgloss.Color
	StatusActive     lipgloss.Color
}

// TaskList is a component for displaying and navigating a list of tasks or groups
type TaskList struct {
	tasks          []core.Task
	groups         []core.TaskGroup // For displaying groups (Projects/Tags)
	displayMode    DisplayMode      // What to display: tasks or groups
	cursor         int              // Selected task/group index
	width          int
	height         int
	displayColumns []string // Column names to display
	offset         int      // Scroll offset for viewport
	styles         TaskListStyles
}

// NewTaskList creates a new task list component
func NewTaskList(width, height int, styles TaskListStyles) TaskList {
	return TaskList{
		tasks:          []core.Task{},
		groups:         []core.TaskGroup{},
		displayMode:    DisplayModeTasks,
		cursor:         0,
		width:          width,
		height:         height,
		displayColumns: []string{"ID", "PROJECT", "P", "DUE", "TAGS", "DESCRIPTION"},
		offset:         0,
		styles:         styles,
	}
}

// SetTasks updates the task list and switches to task display mode
func (t *TaskList) SetTasks(tasks []core.Task) {
	t.tasks = tasks
	t.displayMode = DisplayModeTasks
	// Reset cursor if out of bounds
	if t.cursor >= len(t.tasks) {
		t.cursor = 0
	}
	t.updateScroll()
}

// SetGroups updates the groups and switches to group display mode
func (t *TaskList) SetGroups(groups []core.TaskGroup) {
	t.groups = groups
	t.displayMode = DisplayModeGroups
	// Reset cursor if out of bounds
	if t.cursor >= len(t.groups) {
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
	itemCount := t.itemCount()
	if itemCount == 0 {
		return
	}
	if t.cursor < itemCount-1 {
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
	if t.itemCount() > 0 {
		t.cursor = 0
		t.updateScroll()
	}
}

// moveToEnd jumps to last task
func (t *TaskList) moveToEnd() {
	itemCount := t.itemCount()
	if itemCount > 0 {
		t.cursor = itemCount - 1
		t.updateScroll()
	}
}

// quickJump jumps to a visible task by number (1-9)
func (t *TaskList) quickJump(key string) {
	num := int(key[0] - '0') // Convert '1'-'9' to 1-9
	targetIndex := t.offset + num - 1
	itemCount := t.itemCount()
	if targetIndex >= 0 && targetIndex < itemCount && targetIndex < t.offset+t.height-1 {
		t.cursor = targetIndex
	}
}

// itemCount returns the count of current items (tasks or groups)
func (t TaskList) itemCount() int {
	if t.displayMode == DisplayModeGroups {
		return len(t.groups)
	}
	return len(t.tasks)
}

// updateScroll adjusts the scroll offset to keep cursor visible
func (t *TaskList) updateScroll() {
	itemCount := t.itemCount()
	if itemCount == 0 {
		t.offset = 0
		return
	}

	visibleHeight := t.height - 2 // Subtract 2 for header rows (title + separator)

	// Cursor is above viewport
	if t.cursor < t.offset {
		t.offset = t.cursor
	}

	// Cursor is below viewport
	if t.cursor >= t.offset+visibleHeight {
		t.offset = t.cursor - visibleHeight + 1
	}

	// Don't scroll past the end
	maxOffset := itemCount - visibleHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if t.offset > maxOffset {
		t.offset = maxOffset
	}
}

// View renders the task list or group list
func (t TaskList) View() string {
	if t.displayMode == DisplayModeGroups {
		return t.renderGroupList()
	}
	return t.renderTaskList()
}

// renderTaskList renders the task list
func (t TaskList) renderTaskList() string {
	if len(t.tasks) == 0 {
		return lipgloss.NewStyle().
			Padding(2, 4).
			Render("No tasks found.")
	}

	var lines []string

	// Render column headers (returns 2 lines: header + separator)
	headerLines := strings.Split(t.renderHeader(), "\n")
	lines = append(lines, headerLines...)

	// Calculate visible range
	visibleHeight := t.height - 2 // Subtract 2 for header rows
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
	for len(lines) < t.height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderGroupList renders the group list (for Projects/Tags sections)
func (t TaskList) renderGroupList() string {
	if len(t.groups) == 0 {
		return lipgloss.NewStyle().
			Padding(2, 4).
			Render("No groups found.")
	}

	var lines []string

	// Render group header
	headerLines := strings.Split(t.renderGroupHeader(), "\n")
	lines = append(lines, headerLines...)

	// Calculate visible range
	visibleHeight := t.height - 2 // Subtract 2 for header rows
	endIdx := t.offset + visibleHeight
	if endIdx > len(t.groups) {
		endIdx = len(t.groups)
	}

	// Render visible groups
	for i := t.offset; i < endIdx; i++ {
		group := t.groups[i]
		isSelected := i == t.cursor
		quickJump := ""
		if i-t.offset < 9 {
			quickJump = fmt.Sprintf("%d", i-t.offset+1)
		}
		line := t.renderGroupLine(group, isSelected, quickJump)
		lines = append(lines, line)
	}

	// Fill remaining space
	for len(lines) < t.height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderHeader renders the column header row
func (t TaskList) renderHeader() string {
	cols := t.calculateColumnWidths()

	header := fmt.Sprintf("  %-*s %-*s %s %-*s %-*s %-*s",
		cols.id, "ID",
		cols.project, "PROJECT",
		"P",
		cols.due, "DUE",
		cols.tags, "TAGS",
		cols.description, "DESCRIPTION",
	)

	// Render header with underline separator
	styledHeader := t.styles.Header.Width(t.width).Render(header)
	separator := strings.Repeat("─", t.width)

	return styledHeader + "\n" + t.styles.Separator.Width(t.width).Render(separator)
}

// columnWidths holds calculated column widths
type columnWidths struct {
	id          int
	project     int
	priority    int
	due         int
	tags        int
	description int
}

// calculateColumnWidths determines column widths based on available space
func (t TaskList) calculateColumnWidths() columnWidths {
	const (
		cursorWidth   = 2  // "■ " or "1 "
		idWidth       = 4  // Sequential ID
		priorityWidth = 1  // H/M/L
		dueWidth      = 10 // Date format
		tagsWidth     = 15 // Tags column
		minProject    = 10
		minDesc       = 20
		spacing       = 7
	)

	fixedWidth := cursorWidth + idWidth + priorityWidth + dueWidth + tagsWidth + spacing
	remainingWidth := t.width - fixedWidth

	if remainingWidth < minProject+minDesc {
		return columnWidths{
			id:          idWidth,
			project:     minProject,
			priority:    priorityWidth,
			due:         dueWidth,
			tags:        tagsWidth,
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
		tags:        tagsWidth,
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

	// Priority - conditionally style based on selection
	priority := "-"
	var priorityText string
	if task.Priority != "" {
		priority = string(task.Priority[0])
		if !isSelected {
			// Apply color only when not selected
			priorityStyle := lipgloss.NewStyle()
			switch task.Priority {
			case "H":
				priorityStyle = priorityStyle.Foreground(t.styles.PriorityHigh)
			case "M":
				priorityStyle = priorityStyle.Foreground(t.styles.PriorityMedium)
			case "L":
				priorityStyle = priorityStyle.Foreground(t.styles.PriorityLow)
			}
			priorityText = priorityStyle.Render(priority)
		} else {
			priorityText = priority
		}
	} else {
		priorityText = priority
	}

	// Due date - conditionally style based on selection
	due := "-"
	var dueText string
	if task.Due != nil {
		due = task.FormatDueDate()
		if len(due) > cols.due {
			due = due[:cols.due]
		}
		if !isSelected && task.IsOverdue() {
			// Apply overdue color only when not selected
			dueStyle := lipgloss.NewStyle().Foreground(t.styles.DueOverdue)
			dueText = dueStyle.Render(due)
		} else {
			dueText = due
		}
	} else {
		dueText = due
	}

	// Tags - format with + prefix like taskwarrior
	tags := "-"
	if len(task.Tags) > 0 {
		var tagList []string
		for _, tag := range task.Tags {
			tagList = append(tagList, "+"+tag)
		}
		tags = strings.Join(tagList, " ")
		if len(tags) > cols.tags {
			tags = tags[:cols.tags-3] + "..."
		}
	}

	// Status icon prefix for description
	statusIcon := ""
	if task.Start != nil {
		// Task is started (has Start field)
		statusIcon = "▶ "
	} else if task.Status == "waiting" {
		statusIcon = "⏸ "
	}

	// Description - pad to fill remaining width (accounting for status icon)
	description := statusIcon + task.Description
	if len(description) > cols.description {
		description = description[:cols.description-3] + "..."
	}

	// Build line with conditionally styled text
	line := fmt.Sprintf("%s %-*s %-*s %s %-*s %-*s %-*s",
		cursor,
		cols.id, id,
		cols.project, project,
		priorityText,
		cols.due, dueText,
		cols.tags, tags,
		cols.description, description,
	)

	// Apply status-based styling
	var lineStyle lipgloss.Style

	if isSelected {
		lineStyle = t.styles.Selection
	} else {
		// Apply status styling based on task status
		lineStyle = lipgloss.NewStyle()

		// Check if task is started (has Start field)
		if task.Start != nil {
			// Bold and colored for started tasks
			lineStyle = lineStyle.
				Foreground(t.styles.StatusActive).
				Bold(true)
		} else {
			// Apply status styling
			switch task.Status {
			case "completed":
				// Strikethrough and dim color for completed tasks
				lineStyle = lineStyle.
					Foreground(t.styles.StatusCompleted).
					Strikethrough(true)
			case "waiting":
				// Dim and italic for waiting tasks
				lineStyle = lineStyle.
					Foreground(t.styles.StatusWaiting).
					Italic(true)
			case "deleted":
				// Very dim for deleted tasks
				lineStyle = lineStyle.
					Foreground(t.styles.StatusCompleted).
					Strikethrough(true)
			}
		}
	}

	return lineStyle.Width(t.width).Render(line)
}

// renderGroupHeader renders the header for group list view
func (t TaskList) renderGroupHeader() string {
	header := "  GROUP NAME                                        TASK COUNT"

	styledHeader := t.styles.Header.Width(t.width).Render(header)
	separator := strings.Repeat("─", t.width)

	return styledHeader + "\n" + t.styles.Separator.Width(t.width).Render(separator)
}

// renderGroupLine renders a single group row
func (t TaskList) renderGroupLine(group core.TaskGroup, isSelected bool, quickJump string) string {
	// Cursor or quick jump number
	cursor := " "
	if isSelected {
		cursor = "■"
	} else if quickJump != "" {
		cursor = quickJump
	}

	// Group name - truncate if too long
	groupName := group.Name
	maxNameWidth := t.width - 20 // Leave space for cursor and count
	if len(groupName) > maxNameWidth {
		groupName = groupName[:maxNameWidth-3] + "..."
	}

	// Task count
	countStr := fmt.Sprintf("%d", group.Count)
	if group.Count == 1 {
		countStr += " task"
	} else {
		countStr += " tasks"
	}

	line := fmt.Sprintf("%s %-*s %s",
		cursor,
		maxNameWidth, groupName,
		countStr,
	)

	if isSelected {
		return t.styles.Selection.Width(t.width).Render(line)
	}

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

// SelectedGroup returns the currently selected group, or nil if not in group mode
func (t TaskList) SelectedGroup() *core.TaskGroup {
	if t.displayMode != DisplayModeGroups || len(t.groups) == 0 {
		return nil
	}
	if t.cursor < 0 || t.cursor >= len(t.groups) {
		return nil
	}
	return &t.groups[t.cursor]
}

// SelectedIndex returns the index of the currently selected task
func (t TaskList) SelectedIndex() int {
	return t.cursor
}

// TaskCount returns the total number of tasks
func (t TaskList) TaskCount() int {
	return len(t.tasks)
}

// Cursor returns the current cursor position (exported for Model access)
func (t TaskList) Cursor() int {
	return t.cursor
}
