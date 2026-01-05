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
	Header          lipgloss.Style
	Separator       lipgloss.Style
	Selection       lipgloss.Style
	PriorityHigh    lipgloss.Color
	PriorityMedium  lipgloss.Color
	PriorityLow     lipgloss.Color
	DueOverdue      lipgloss.Color
	TagColor        lipgloss.Color
	StatusCompleted lipgloss.Color
	StatusWaiting   lipgloss.Color
	StatusActive    lipgloss.Color
}

// TaskList is a component for displaying and navigating a list of tasks or groups
type TaskList struct {
	tasks          []core.Task
	groups         []core.TaskGroup // For displaying groups (Projects/Tags)
	displayMode    DisplayMode      // What to display: tasks or groups
	cursor         int              // Selected task/group index
	selectedUUIDs  map[string]bool  // Multi-select: UUIDs of selected tasks
	width          int
	height         int
	displayColumns []string // Column names to display
	offset         int      // Scroll offset for viewport
	scrollBuffer   int      // Number of tasks to keep visible above/below cursor
	styles         TaskListStyles
	emptyMessage   string // Custom message to show when list is empty
}

// NewTaskList creates a new task list component
func NewTaskList(width, height int, columns []string, styles TaskListStyles) TaskList {
	// Default columns if none provided
	if len(columns) == 0 {
		columns = []string{"id", "project", "priority", "due", "description"}
	}

	// Limit to maximum 8 columns
	if len(columns) > 8 {
		columns = columns[:8]
	}

	// Normalize column names to lowercase
	normalizedColumns := make([]string, len(columns))
	for i, col := range columns {
		normalizedColumns[i] = strings.ToLower(col)
	}

	return TaskList{
		tasks:          []core.Task{},
		groups:         []core.TaskGroup{},
		displayMode:    DisplayModeTasks,
		cursor:         0,
		selectedUUIDs:  make(map[string]bool),
		width:          width,
		height:         height,
		displayColumns: normalizedColumns,
		offset:         0,
		scrollBuffer:   1, // Default: keep 1 task visible above/below cursor
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

// SetEmptyMessage sets a custom message to display when the list is empty
func (t *TaskList) SetEmptyMessage(message string) {
	t.emptyMessage = message
}

// SetScrollBuffer sets the number of tasks to keep visible above/below the cursor.
// A buffer of 1 means the selected task will have at least 1 task visible above
// and below it (when not at list boundaries). Set to 0 to disable buffering.
func (t *TaskList) SetScrollBuffer(buffer int) {
	if buffer < 0 {
		buffer = 0
	}
	t.scrollBuffer = buffer
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

// updateScroll adjusts the scroll offset to keep cursor visible with a configurable buffer.
//
// The scroll buffer (scrollBuffer) determines how many tasks remain visible above and below
// the selected task when scrolling. For example, with scrollBuffer=1:
//   - When moving down, scrolling triggers when cursor is 1 position from viewport bottom
//   - When moving up, scrolling triggers when cursor is 1 position from viewport top
//   - This maintains context around the selected task during navigation
//
// Adaptive buffer at list boundaries:
//   - At list start: buffer below cursor is maintained, buffer above cursor is 0
//   - At list end: buffer above cursor is maintained, buffer below cursor adapts to remaining tasks
//   - Example: With scrollBuffer=3, when on task 18 of 20 tasks, only 1 task remains below,
//     so the effective buffer below becomes 1 instead of 3
//   - This prevents jarring jumps and ensures all tasks are reachable
//   - With scrollBuffer=0: cursor can touch viewport edges (original behavior)
//
// Small screen handling:
//   - Small screens (width < 80) skip headers and use 2 lines per task
//   - visibleTasks calculation accounts for these differences
func (t *TaskList) updateScroll() {
	itemCount := t.itemCount()
	if itemCount == 0 {
		t.offset = 0
		return
	}

	// Calculate visible tasks accounting for screen size
	isSmallScreen := t.width < 80
	headerHeight := 0
	if !isSmallScreen {
		headerHeight = 2 // header + separator
	}
	visibleHeight := t.height - headerHeight

	// Calculate how many tasks can fit in viewport
	visibleTasks := visibleHeight
	if isSmallScreen {
		visibleTasks = visibleHeight / 2 // Each task takes 2 lines
	}

	// Ensure we have at least 1 visible task
	if visibleTasks < 1 {
		visibleTasks = 1
	}

	// Calculate max offset - cannot scroll past the last task
	maxOffset := itemCount - visibleTasks
	if maxOffset < 0 {
		maxOffset = 0
	}

	// Cursor moving up: scroll when cursor gets within buffer distance from top
	// Unless we're already at the start of the list
	if t.cursor < t.offset+t.scrollBuffer {
		t.offset = t.cursor - t.scrollBuffer
		if t.offset < 0 {
			t.offset = 0
		}
	}

	// Cursor moving down: scroll when cursor gets within buffer distance from bottom
	// Use adaptive buffer that reduces when approaching the end of the list
	tasksBelow := itemCount - t.cursor - 1
	effectiveBufferBelow := t.scrollBuffer
	if tasksBelow < effectiveBufferBelow {
		effectiveBufferBelow = tasksBelow
		if effectiveBufferBelow < 0 {
			effectiveBufferBelow = 0
		}
	}

	if t.cursor >= t.offset+visibleTasks-effectiveBufferBelow-1 {
		t.offset = t.cursor - visibleTasks + effectiveBufferBelow + 1
	}

	// Ensure cursor is always visible (safety check)
	// This handles edge cases where buffer logic might fail
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+visibleTasks {
		t.offset = t.cursor - visibleTasks + 1
	}

	// Don't scroll past the end of the list
	if t.offset > maxOffset {
		t.offset = maxOffset
	}

	// Final safety: ensure offset is non-negative
	if t.offset < 0 {
		t.offset = 0
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
		message := "No tasks found."
		if t.emptyMessage != "" {
			message = t.emptyMessage
		}
		return lipgloss.NewStyle().
			Padding(2, 4).
			Render(message)
	}

	// Check if we're in small screen mode
	isSmallScreen := t.width < 80

	var lines []string

	// Render column headers (returns 2 lines: header + separator)
	// Skip headers in small screen mode to save space
	if !isSmallScreen {
		headerLines := strings.Split(t.renderHeader(), "\n")
		lines = append(lines, headerLines...)
	}

	// Calculate visible range
	headerHeight := 0
	if !isSmallScreen {
		headerHeight = 2 // header + separator
	}

	visibleHeight := t.height - headerHeight

	// In small screen mode, each task takes 2 lines
	var endIdx int
	if isSmallScreen {
		maxVisibleTasks := visibleHeight / 2
		endIdx = t.offset + maxVisibleTasks
		if endIdx > len(t.tasks) {
			endIdx = len(t.tasks)
		}
	} else {
		endIdx = t.offset + visibleHeight
		if endIdx > len(t.tasks) {
			endIdx = len(t.tasks)
		}
	}

	// Render visible tasks
	for i := t.offset; i < endIdx; i++ {
		task := t.tasks[i]
		isCursor := i == t.cursor
		isMultiSelected := t.IsSelected(task.UUID)

		if isSmallScreen {
			// Small screen: render 2 lines per task
			taskLines := t.renderSmallScreenTaskLines(task, isCursor, isMultiSelected)
			lines = append(lines, taskLines...)
		} else {
			// Normal screen: render 1 line per task
			quickJump := ""
			if i-t.offset < 9 {
				quickJump = fmt.Sprintf("%d", i-t.offset+1)
			}
			line := t.renderTaskLine(task, isCursor, isMultiSelected, quickJump)
			lines = append(lines, line)
		}
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

	// Column name mappings
	columnNames := map[string]string{
		"id":          "ID",
		"project":     "PROJECT",
		"priority":    "P",
		"due":         "DUE",
		"tags":        "TAGS",
		"annotation":  "A",
		"dependency":  "D",
		"description": "DESCRIPTION",
	}

	// Build header dynamically based on displayColumns
	parts := []string{"  "} // Start with cursor space

	for _, col := range t.displayColumns {
		width := cols.widths[col]
		name := columnNames[col]
		if name == "" {
			name = strings.ToUpper(col)
		}

		// Priority, annotation, and dependency columns are single character, others are padded
		if col == "priority" || col == "annotation" || col == "dependency" {
			parts = append(parts, name+" ")
		} else {
			parts = append(parts, fmt.Sprintf("%-*s ", width, truncate(name, width)))
		}
	}

	header := strings.Join(parts, "")

	// Truncate header to width if necessary
	if len(header) > t.width {
		header = header[:t.width]
	}

	// Render header with exact width
	styledHeader := t.styles.Header.Render(header)

	// Separator should match the actual width
	separatorWidth := t.width
	if len(header) < t.width {
		separatorWidth = len(header)
	}
	separator := strings.Repeat("─", separatorWidth)

	return styledHeader + "\n" + t.styles.Separator.Render(separator)
}

// truncate truncates a string to the given length
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length <= 3 {
		return s[:length]
	}
	return s[:length-3] + "..."
}

// columnWidths holds calculated column widths
type columnWidths struct {
	widths map[string]int // Map of column name to width
}

// hasColumn checks if a column is enabled
func (t TaskList) hasColumn(name string) bool {
	for _, col := range t.displayColumns {
		if col == name {
			return true
		}
	}
	return false
}

// calculateColumnWidths determines column widths based on available space
func (t TaskList) calculateColumnWidths() columnWidths {
	const (
		cursorWidth     = 2  // "■ " or "1 "
		idWidth         = 4  // Sequential ID
		priorityWidth   = 1  // H/M/L
		dueWidth        = 10 // Date format
		tagsWidth       = 15 // Tags column
		annotationWidth = 1  // * or nothing
		dependencyWidth = 1  // * or nothing
		minProject      = 10
		minDesc         = 20
		spacing         = 7
	)

	fixedWidth := cursorWidth + idWidth + priorityWidth + dueWidth + tagsWidth + spacing
	remainingWidth := t.width - fixedWidth

	widths := make(map[string]int)

	// Set fixed widths for columns that are enabled
	if t.hasColumn("id") {
		widths["id"] = idWidth
	}
	if t.hasColumn("priority") {
		widths["priority"] = priorityWidth
	}
	if t.hasColumn("due") {
		widths["due"] = dueWidth
	}
	if t.hasColumn("tags") {
		widths["tags"] = tagsWidth
	}
	if t.hasColumn("annotation") {
		widths["annotation"] = annotationWidth
	}
	if t.hasColumn("dependency") {
		widths["dependency"] = dependencyWidth
	}

	// Calculate remaining space for flexible columns (project and description)
	if remainingWidth < minProject+minDesc {
		if t.hasColumn("project") {
			widths["project"] = minProject
		}
		if t.hasColumn("description") {
			widths["description"] = minDesc
		}
	} else {
		// Allocate 25% to project, 75% to description
		if t.hasColumn("project") && t.hasColumn("description") {
			projectWidth := remainingWidth / 4
			if projectWidth < minProject {
				projectWidth = minProject
			}
			if projectWidth > 20 {
				projectWidth = 20
			}
			widths["project"] = projectWidth
			widths["description"] = remainingWidth - projectWidth
		} else if t.hasColumn("project") {
			widths["project"] = remainingWidth
		} else if t.hasColumn("description") {
			widths["description"] = remainingWidth
		}
	}

	return columnWidths{widths: widths}
}

// renderTaskLine renders a single task row
func (t TaskList) renderTaskLine(task core.Task, isCursor bool, isMultiSelected bool, quickJump string) string {
	cols := t.calculateColumnWidths()

	// First column: cursor and multi-select indicator
	// Cursor (current line): "■", Multi-selected: "✓", Both: "◆", Neither: " "
	cursor := " "
	if isCursor && isMultiSelected {
		cursor = "◆" // Both cursor and selected
	} else if isCursor {
		cursor = "■" // Just cursor
	} else if isMultiSelected {
		cursor = "✓" // Just selected
	}
	// Note: quickJump numbers (1-9) are for keyboard shortcuts only, not displayed

	// ID column: task ID
	id := fmt.Sprintf("%d", task.ID)
	if task.ID == 0 {
		id = task.UUID
		if len(id) > cols.widths["id"] {
			id = id[:cols.widths["id"]]
		}
	}

	// Project
	project := task.Project
	if project == "" {
		project = "-"
	}
	if len(project) > cols.widths["project"] && cols.widths["project"] > 3 {
		project = project[:cols.widths["project"]-3] + "..."
	} else if len(project) > cols.widths["project"] {
		project = project[:cols.widths["project"]]
	}

	// Priority - conditionally style based on selection
	priority := "-"
	var priorityText string
	if task.Priority != "" {
		priority = string(task.Priority[0])
		if !isMultiSelected {
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
		if len(due) > cols.widths["due"] {
			due = due[:cols.widths["due"]]
		}
		if !isMultiSelected && task.IsOverdue() {
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
		if len(tags) > cols.widths["tags"] && cols.widths["tags"] > 3 {
			tags = tags[:cols.widths["tags"]-3] + "..."
		} else if len(tags) > cols.widths["tags"] {
			tags = tags[:cols.widths["tags"]]
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
	if len(description) > cols.widths["description"] && cols.widths["description"] > 3 {
		description = description[:cols.widths["description"]-3] + "..."
	} else if len(description) > cols.widths["description"] {
		description = description[:cols.widths["description"]]
	}

	// Annotation indicator - show "*" if task has annotations
	annotation := "-"
	if len(task.Annotations) > 0 {
		annotation = "*"
	}

	// Dependency indicator - show "*" if task has dependencies
	dependency := "-"
	if len(task.Depends) > 0 {
		dependency = "*"
	}

	// Build line dynamically based on displayColumns
	columnValues := map[string]string{
		"id":          id,
		"project":     project,
		"priority":    priorityText,
		"due":         dueText,
		"tags":        tags,
		"annotation":  annotation,
		"dependency":  dependency,
		"description": description,
	}

	parts := []string{cursor + " "}
	for _, col := range t.displayColumns {
		value := columnValues[col]
		width := cols.widths[col]

		// Priority, annotation, and dependency don't need padding, just add with space
		if col == "priority" || col == "annotation" || col == "dependency" {
			parts = append(parts, value+" ")
		} else {
			parts = append(parts, fmt.Sprintf("%-*s ", width, value))
		}
	}

	line := strings.Join(parts, "")

	// Apply status-based styling
	var lineStyle lipgloss.Style

	if isCursor || isMultiSelected {
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

// renderSmallScreenTaskLines renders a task as 2 lines for small screens
// Line 1: ID + Description
// Line 2: Due date (indented)
func (t TaskList) renderSmallScreenTaskLines(task core.Task, isCursor bool, isMultiSelected bool) []string {
	// Cursor indicator
	cursor := " "
	if isCursor && isMultiSelected {
		cursor = "◆" // Both cursor and selected
	} else if isCursor {
		cursor = "■" // Just cursor
	} else if isMultiSelected {
		cursor = "✓" // Just selected
	}

	// ID
	id := fmt.Sprintf("%d", task.ID)
	if task.ID == 0 {
		id = task.UUID
		if len(id) > 4 {
			id = id[:4]
		}
	}

	// Status icon prefix for description
	statusIcon := ""
	if task.Start != nil {
		statusIcon = "▶ "
	} else if task.Status == "waiting" {
		statusIcon = "⏸ "
	}

	// Line 1: ID + Description
	// Format: "■ 1 Description text here..."
	availableWidth := t.width - 7 // cursor(2) + id(4) + space(1)
	description := statusIcon + task.Description
	if len(description) > availableWidth && availableWidth > 3 {
		description = description[:availableWidth-3] + "..."
	} else if len(description) > availableWidth {
		description = description[:availableWidth]
	}
	line1 := fmt.Sprintf("%s %-4s %s", cursor, id, description)

	// Line 2: Due date (indented)
	// Format: "    Due: 2025-01-05" or "    Due: -" if no due date
	dueText := "-"
	var dueLine string
	if task.Due != nil {
		dueText = task.FormatDueDate()
		// Add overdue indicator
		if task.IsOverdue() {
			dueText += " ⚠"
		}
	}
	dueLine = fmt.Sprintf("    Due: %s", dueText)

	// Apply status-based styling to both lines
	var lineStyle lipgloss.Style
	if isCursor || isMultiSelected {
		lineStyle = t.styles.Selection
	} else {
		lineStyle = lipgloss.NewStyle()
		if task.Start != nil {
			lineStyle = lineStyle.Foreground(t.styles.StatusActive).Bold(true)
		} else {
			switch task.Status {
			case "completed":
				lineStyle = lineStyle.Foreground(t.styles.StatusCompleted).Strikethrough(true)
			case "waiting":
				lineStyle = lineStyle.Foreground(t.styles.StatusWaiting).Italic(true)
			case "deleted":
				lineStyle = lineStyle.Foreground(t.styles.StatusCompleted).Strikethrough(true)
			}
		}
	}

	styledLine1 := lineStyle.Width(t.width).Render(line1)
	styledLine2 := lineStyle.Width(t.width).Render(dueLine)

	return []string{styledLine1, styledLine2}
}

// renderGroupHeader renders the header for group list view
func (t TaskList) renderGroupHeader() string {
	header := "  PROJECT                                           TASK COUNT"

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

	// Format percentage if available
	percentStr := ""
	if group.Percentage >= 0 {
		// Format as [XX%] with padding for consistent width (e.g., [ 75%])
		percentStr = fmt.Sprintf("[%3d%%] ", group.Percentage)
	}

	// Add indentation based on depth level
	// Depth 0: no indentation
	// Depth 1: "- "
	// Depth 2: "  - "
	// Depth 3: "    - "
	indent := ""
	if group.Depth > 0 {
		// Add 2 spaces per depth level (after depth 1)
		spaces := strings.Repeat("  ", group.Depth-1)
		indent = spaces + "- "
	}

	// Construct name with percentage and indentation (use full project name)
	nameWithPrefix := percentStr + indent + group.Name

	// Calculate available width for the name
	maxNameWidth := t.width - 20 // Leave space for cursor and count
	if maxNameWidth < 1 {
		maxNameWidth = 1
	}

	// Truncate if too long
	if len(nameWithPrefix) > maxNameWidth && maxNameWidth > 3 {
		nameWithPrefix = nameWithPrefix[:maxNameWidth-3] + "..."
	} else if len(nameWithPrefix) > maxNameWidth {
		nameWithPrefix = nameWithPrefix[:maxNameWidth]
	}

	// Task count (optional, can be removed if not needed)
	countStr := ""
	if group.Count > 0 {
		if group.Count == 1 {
			countStr = fmt.Sprintf("%d task", group.Count)
		} else {
			countStr = fmt.Sprintf("%d tasks", group.Count)
		}
	}

	line := fmt.Sprintf("%s %-*s %s",
		cursor,
		maxNameWidth, nameWithPrefix,
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

// ToggleSelection toggles the selection state of the current task
func (t *TaskList) ToggleSelection() {
	if t.displayMode != DisplayModeTasks {
		return // Only works in task mode
	}

	task := t.SelectedTask()
	if task != nil {
		if t.selectedUUIDs[task.UUID] {
			delete(t.selectedUUIDs, task.UUID)
		} else {
			t.selectedUUIDs[task.UUID] = true
		}
	}
}

// ClearSelection clears all selected tasks
func (t *TaskList) ClearSelection() {
	t.selectedUUIDs = make(map[string]bool)
}

// GetSelectedTasks returns all currently selected tasks, or the cursor task if none selected
func (t TaskList) GetSelectedTasks() []core.Task {
	if len(t.selectedUUIDs) == 0 {
		// No multi-selection, return current task if any
		task := t.SelectedTask()
		if task != nil {
			return []core.Task{*task}
		}
		return []core.Task{}
	}

	// Return all selected tasks
	var selected []core.Task
	for _, task := range t.tasks {
		if t.selectedUUIDs[task.UUID] {
			selected = append(selected, task)
		}
	}
	return selected
}

// HasSelections returns true if any tasks are selected
func (t TaskList) HasSelections() bool {
	return len(t.selectedUUIDs) > 0
}

// IsSelected returns true if the given task UUID is selected
func (t TaskList) IsSelected(uuid string) bool {
	return t.selectedUUIDs[uuid]
}
