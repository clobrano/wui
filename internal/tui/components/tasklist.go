package components

import (
	"fmt"
	"log"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clobrano/wui/internal/config"
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
	tasks               []core.Task
	groups              []core.TaskGroup // For displaying groups (Projects/Tags)
	displayMode         DisplayMode      // What to display: tasks or groups
	cursor              int              // Selected task/group index
	selectedUUIDs       map[string]bool  // Multi-select: UUIDs of selected tasks
	width               int
	height              int
	displayColumns      []string          // Column names to display
	columnLabels        map[string]string // Map of column name to display label
	columnLengths       map[string]int    // Map of column name to custom max length (0 = use default)
	narrowViewFields    []string          // Field names to display in narrow view (below description)
	narrowViewLabels    map[string]string // Map of field name to display label for narrow view
	narrowViewLengths   map[string]int    // Map of field name to custom max length for narrow view
	offset              int               // Scroll offset for viewport
	scrollBuffer        int               // Number of tasks to keep visible above/below cursor
	styles              TaskListStyles
	emptyMessage        string // Custom message to show when list is empty
}

// NewTaskList creates a new task list component
func NewTaskList(width, height int, columns config.Columns, narrowViewFields config.Columns, styles TaskListStyles) TaskList {
	// Default columns if none provided
	if len(columns) == 0 {
		columns = config.DefaultColumns()
	}

	// Limit to maximum 8 columns
	if len(columns) > 8 {
		columns = columns[:8]
	}

	// Build column names, labels, and lengths maps
	normalizedColumns := make([]string, len(columns))
	columnLabels := make(map[string]string)
	columnLengths := make(map[string]int)
	for i, col := range columns {
		normalizedName := strings.ToLower(col.Name)
		normalizedColumns[i] = normalizedName
		columnLabels[normalizedName] = col.Label
		columnLengths[normalizedName] = col.Length
	}

	// Default narrow view fields if none provided
	if len(narrowViewFields) == 0 {
		narrowViewFields = config.DefaultNarrowViewFields()
	}

	// Build narrow view field names, labels, and lengths maps
	normalizedNarrowViewFields := make([]string, len(narrowViewFields))
	narrowViewLabels := make(map[string]string)
	narrowViewLengths := make(map[string]int)
	for i, field := range narrowViewFields {
		normalizedName := strings.ToLower(field.Name)
		normalizedNarrowViewFields[i] = normalizedName
		narrowViewLabels[normalizedName] = field.Label
		narrowViewLengths[normalizedName] = field.Length
	}

	return TaskList{
		tasks:             []core.Task{},
		groups:            []core.TaskGroup{},
		displayMode:       DisplayModeTasks,
		cursor:            0,
		selectedUUIDs:     make(map[string]bool),
		width:             width,
		height:            height,
		displayColumns:    normalizedColumns,
		columnLabels:      columnLabels,
		columnLengths:     columnLengths,
		narrowViewFields:  normalizedNarrowViewFields,
		narrowViewLabels:  narrowViewLabels,
		narrowViewLengths: narrowViewLengths,
		offset:            0,
		scrollBuffer:      1, // Default: keep 1 task visible above/below cursor
		styles:            styles,
	}
}

// SetTasks updates the task list and switches to task display mode
func (t *TaskList) SetTasks(tasks []core.Task) {
	t.SetTasksWithSort(tasks, "", false)
}

// SetTasksWithSort updates the task list with custom sorting
func (t *TaskList) SetTasksWithSort(tasks []core.Task, sortMethod string, reverse bool) {
	// Sort tasks: non-completed tasks first, completed tasks last
	// Use stable sort to maintain original order within each group
	sortedTasks := make([]core.Task, len(tasks))
	copy(sortedTasks, tasks)

	sort.SliceStable(sortedTasks, func(i, j int) bool {
		taskI := sortedTasks[i]
		taskJ := sortedTasks[j]

		// First priority: Completed tasks should come after non-completed tasks
		isCompletedI := taskI.Status == "completed"
		isCompletedJ := taskJ.Status == "completed"

		if isCompletedI != isCompletedJ {
			return !isCompletedI // true if i is not completed (i comes first)
		}

		// Second priority: Apply custom sorting if specified
		if sortMethod != "" {
			result := compareTasks(taskI, taskJ, sortMethod)
			if result != 0 {
				if reverse {
					return result > 0
				}
				return result < 0
			}
		}

		// Otherwise maintain original order (stable sort)
		return false
	})

	t.tasks = sortedTasks
	t.displayMode = DisplayModeTasks
	// Reset cursor if out of bounds
	if t.cursor >= len(t.tasks) {
		t.cursor = 0
	}
	t.updateScroll()
}

// compareTasks compares two tasks based on the specified sort method
// Returns: -1 if taskI < taskJ, 0 if equal, 1 if taskI > taskJ
func compareTasks(taskI, taskJ core.Task, sortMethod string) int {
	switch sortMethod {
	case "alphabetic", "alpha", "description":
		// Sort by description alphabetically
		if taskI.Description < taskJ.Description {
			return -1
		} else if taskI.Description > taskJ.Description {
			return 1
		}
		return 0

	case "due":
		// Sort by due date (tasks without due date go last)
		if taskI.Due == nil && taskJ.Due == nil {
			return 0
		}
		if taskI.Due == nil {
			return 1 // taskI goes after taskJ
		}
		if taskJ.Due == nil {
			return -1 // taskI goes before taskJ
		}
		if taskI.Due.Before(*taskJ.Due) {
			return -1
		} else if taskI.Due.After(*taskJ.Due) {
			return 1
		}
		return 0

	case "scheduled":
		// Sort by scheduled date (tasks without scheduled date go last)
		if taskI.Scheduled == nil && taskJ.Scheduled == nil {
			return 0
		}
		if taskI.Scheduled == nil {
			return 1
		}
		if taskJ.Scheduled == nil {
			return -1
		}
		if taskI.Scheduled.Before(*taskJ.Scheduled) {
			return -1
		} else if taskI.Scheduled.After(*taskJ.Scheduled) {
			return 1
		}
		return 0

	case "created", "entry":
		// Sort by creation date (Entry is always set, non-pointer)
		if taskI.Entry.Before(taskJ.Entry) {
			return -1
		} else if taskI.Entry.After(taskJ.Entry) {
			return 1
		}
		return 0

	case "modified":
		// Sort by modified date (tasks without modified date go last)
		if taskI.Modified == nil && taskJ.Modified == nil {
			return 0
		}
		if taskI.Modified == nil {
			return 1
		}
		if taskJ.Modified == nil {
			return -1
		}
		if taskI.Modified.Before(*taskJ.Modified) {
			return -1
		} else if taskI.Modified.After(*taskJ.Modified) {
			return 1
		}
		return 0

	default:
		// Unknown sort method, maintain original order
		return 0
	}
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

	// Build header dynamically based on displayColumns
	parts := []string{"  "} // Start with cursor space

	for _, col := range t.displayColumns {
		width := cols.widths[col]
		// Use custom label from configuration, or uppercase column name as fallback
		name := t.columnLabels[col]
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

// getColumnWidth returns the default width for a column type
func getColumnWidth(columnName string) (width int, isFixed bool) {
	switch columnName {
	// Single-character columns
	case "priority", "annotation", "dependency":
		return 1, true
	// ID column
	case "id":
		return 4, true
	// UUID column (short form)
	case "uuid":
		return 8, true
	// Date columns (YYYY-MM-DD HH:MM format - 16 chars max)
	case "due", "scheduled", "wait", "start", "entry", "modified", "end":
		return 16, true
	// Tags column
	case "tags":
		return 15, true
	// Urgency column
	case "urgency":
		return 5, true
	// Status column
	case "status":
		return 10, true
	// Variable-width columns
	case "project":
		return 10, false // minimum width, can grow
	case "description":
		return 20, false // minimum width, can grow
	default:
		// Unknown columns get a default width
		return 15, true
	}
}

// calculateColumnWidths determines column widths based on available space
func (t TaskList) calculateColumnWidths() columnWidths {
	const (
		cursorWidth = 2 // "■ " or "  "
		spacing     = 1 // Space between columns
	)

	widths := make(map[string]int)

	// Calculate fixed width and identify flexible columns
	fixedWidth := cursorWidth
	var flexibleColumns []string

	for _, col := range t.displayColumns {
		// Check if user specified a custom length for this column
		if customLength, hasCustom := t.columnLengths[col]; hasCustom && customLength > 0 {
			// Use custom length specified by user
			widths[col] = customLength
			fixedWidth += customLength + spacing
		} else {
			// Use default width calculation
			width, isFixed := getColumnWidth(col)
			if isFixed {
				widths[col] = width
				fixedWidth += width + spacing
			} else {
				flexibleColumns = append(flexibleColumns, col)
			}
		}
	}

	// Calculate remaining width for flexible columns
	remainingWidth := t.width - fixedWidth
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	// Distribute remaining width among flexible columns
	if len(flexibleColumns) > 0 {
		// Special handling for project and description combination
		hasProject := false
		hasDescription := false
		for _, col := range flexibleColumns {
			if col == "project" {
				hasProject = true
			} else if col == "description" {
				hasDescription = true
			}
		}

		if hasProject && hasDescription {
			// Allocate 25% to project, 75% to description
			projectWidth := remainingWidth / 4
			minProject, _ := getColumnWidth("project")
			if projectWidth < minProject {
				projectWidth = minProject
			}
			if projectWidth > 20 {
				projectWidth = 20
			}
			widths["project"] = projectWidth
			widths["description"] = remainingWidth - projectWidth

			// Remove project and description from flexibleColumns
			newFlexible := []string{}
			for _, col := range flexibleColumns {
				if col != "project" && col != "description" {
					newFlexible = append(newFlexible, col)
				}
			}
			flexibleColumns = newFlexible
		}

		// Distribute remaining space equally among other flexible columns
		if len(flexibleColumns) > 0 {
			widthPerColumn := remainingWidth / len(flexibleColumns)
			for _, col := range flexibleColumns {
				minWidth, _ := getColumnWidth(col)
				if widthPerColumn < minWidth {
					widthPerColumn = minWidth
				}
				widths[col] = widthPerColumn
			}
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

	// Build line dynamically based on displayColumns
	parts := []string{cursor + " "}
	for _, col := range t.displayColumns {
		value, exists := task.GetProperty(col)
		width := cols.widths[col]

		// If property doesn't exist, show empty value and log warning
		if !exists {
			log.Printf("Warning: unknown column '%s' - property not found in task", col)
			value = "-"
		}

		// Handle description separately (add status icons)
		if col == "description" {
			// Add status icon prefix for description
			statusIcon := ""
			if task.Start != nil {
				statusIcon = "▶ "
			} else if task.Status == "waiting" {
				statusIcon = "⏸ "
			}
			value = statusIcon + task.Description
		}

		// Truncate value if it exceeds column width
		// Date columns should be truncated without ellipses
		isDateColumn := col == "due" || col == "scheduled" || col == "wait" ||
			col == "start" || col == "entry" || col == "modified" || col == "end"

		if len(value) > width {
			if isDateColumn {
				// Hard truncate date columns without ellipses
				value = value[:width]
			} else if width > 3 {
				// Add ellipses for other columns
				value = value[:width-3] + "..."
			} else {
				value = value[:width]
			}
		}

		// Pad value to column width BEFORE applying styling
		// Single-character columns don't need padding
		if col == "priority" || col == "annotation" || col == "dependency" {
			// These will be handled with styling below, just add space
		} else {
			// Pad to width for consistent column alignment
			value = fmt.Sprintf("%-*s", width, value)
		}

		// Apply styling for specific columns AFTER padding
		switch col {
		case "priority":
			// Apply color coding for priority (only when not highlighted)
			if !isCursor && !isMultiSelected && task.Priority != "" {
				priorityStyle := lipgloss.NewStyle()
				switch task.Priority {
				case "H":
					priorityStyle = priorityStyle.Foreground(t.styles.PriorityHigh)
				case "M":
					priorityStyle = priorityStyle.Foreground(t.styles.PriorityMedium)
				case "L":
					priorityStyle = priorityStyle.Foreground(t.styles.PriorityLow)
				}
				value = priorityStyle.Render(value)
			}
			parts = append(parts, value+" ")
			continue

		case "due":
			// Apply color coding for overdue tasks (only when not highlighted)
			if !isCursor && !isMultiSelected && task.IsOverdue() {
				dueStyle := lipgloss.NewStyle().Foreground(t.styles.DueOverdue)
				value = dueStyle.Render(value)
			}
		}

		// Add to parts with trailing space
		parts = append(parts, value+" ")
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

// renderSmallScreenTaskLines renders a task as multiple lines for small screens
// Line 1: Cursor (1) + Space (1) + ID field (2, left-aligned) + Indent (2) + Description
// Line 2+: Configured fields (6 space indent to align with description)
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
		if len(id) > 2 {
			id = id[:2]
		}
	}

	// Status icon prefix for description
	statusIcon := ""
	if task.Start != nil {
		statusIcon = "▶ "
	} else if task.Status == "waiting" {
		statusIcon = "⏸ "
	}

	// Line 1: ID + Description with 2 space indent
	// Format: "■ 1  Description text here..."
	availableWidth := t.width - 6 // cursor(2) + id(2) + indent(2)
	description := statusIcon + task.Description
	if len(description) > availableWidth && availableWidth > 3 {
		description = description[:availableWidth-3] + "..."
	} else if len(description) > availableWidth {
		description = description[:availableWidth]
	}
	line1 := fmt.Sprintf("%s %-2s  %s", cursor, id, description)

	// Apply status-based styling to all lines
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

	// Start building the result with line 1
	lines := []string{lineStyle.Width(t.width).Render(line1)}

	// Add configured narrow view fields
	for _, fieldName := range t.narrowViewFields {
		label := t.narrowViewLabels[fieldName]
		if label == "" {
			label = strings.Title(fieldName)
		}

		// Get the field value
		value, exists := task.GetProperty(fieldName)
		if !exists {
			value = "-"
		}

		// Special handling for due dates: add overdue indicator
		if fieldName == "due" && task.Due != nil && task.IsOverdue() {
			value += " ⚠"
		}

		// Format: "      Label: value" (6 spaces to align with description)
		fieldLine := fmt.Sprintf("      %s: %s", label, value)

		// Apply length limit if configured
		maxLength := t.narrowViewLengths[fieldName]
		if maxLength > 0 && len(fieldLine) > maxLength {
			if maxLength > 3 {
				fieldLine = fieldLine[:maxLength-3] + "..."
			} else {
				fieldLine = fieldLine[:maxLength]
			}
		}

		lines = append(lines, lineStyle.Width(t.width).Render(fieldLine))
	}

	return lines
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
