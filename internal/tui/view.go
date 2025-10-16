package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/clobrano/wui/internal/core"
)

// View renders the TUI to a string
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var sections []string

	// Header
	sections = append(sections, m.renderHeader())

	// Main content area
	sections = append(sections, m.renderContent())

	// Footer with keybindings
	sections = append(sections, m.renderFooter())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderHeader renders the header section
func (m Model) renderHeader() string {
	title := "wui - Warrior UI"

	// Add current section if available
	if m.currentSection != nil {
		title += " | " + m.currentSection.Name
	}

	// Add active filter if not default
	if m.activeFilter != "" && m.currentSection != nil && m.activeFilter != m.currentSection.Filter {
		title += fmt.Sprintf(" | Filter: %s", m.activeFilter)
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Padding(0, 1)

	return headerStyle.Render(title)
}

// renderContent renders the main content based on current state
func (m Model) renderContent() string {
	switch m.state {
	case StateHelp:
		return m.renderHelp()
	case StateFilterInput:
		return m.renderFilterInput()
	case StateConfirm:
		return m.renderConfirm()
	default:
		return m.renderTaskList()
	}
}

// renderTaskList renders the task list (and sidebar if enabled)
func (m Model) renderTaskList() string {
	if len(m.tasks) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Padding(2, 4).
			Render("No tasks found. Press 'n' to create a new task.")
		// Fill remaining space
		remaining := m.height - 4 // subtract header and footer
		if remaining > 3 {
			emptyMsg += strings.Repeat("\n", remaining-3)
		}
		return emptyMsg
	}

	var lines []string

	// Calculate available width for task list
	taskListWidth := m.width
	sidebarWidth := 0
	if m.viewMode == ViewModeListWithSidebar {
		sidebarWidth = m.width / 3
		if sidebarWidth < 30 {
			sidebarWidth = 30
		}
		taskListWidth = m.width - sidebarWidth
	}

	// Calculate available height for task list (subtract header, column header, and footer)
	availableHeight := m.height - 5 // header(1) + column_header(1) + footer(2) + padding(1)
	if availableHeight < 1 {
		availableHeight = 10
	}

	// Add column headers
	lines = append(lines, m.renderColumnHeaders(taskListWidth))

	// Render tasks
	startIdx := 0
	endIdx := len(m.tasks)

	// Simple scrolling: show window around selected task
	if len(m.tasks) > availableHeight {
		// Center selected task in viewport
		halfHeight := availableHeight / 2
		startIdx = m.selectedIndex - halfHeight
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + availableHeight
		if endIdx > len(m.tasks) {
			endIdx = len(m.tasks)
			startIdx = endIdx - availableHeight
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	for i := startIdx; i < endIdx; i++ {
		task := m.tasks[i]
		line := m.renderTaskLine(task, i == m.selectedIndex, taskListWidth)
		lines = append(lines, line)
	}

	// Fill remaining vertical space
	renderedLines := len(lines)
	if renderedLines < availableHeight+1 { // +1 for header
		for i := 0; i < availableHeight+1-renderedLines; i++ {
			lines = append(lines, "")
		}
	}

	taskListView := strings.Join(lines, "\n")

	// Show sidebar if enabled
	if m.viewMode == ViewModeListWithSidebar && m.selectedIndex < len(m.tasks) {
		selectedTask := m.tasks[m.selectedIndex]
		sidebar := m.renderSidebar(selectedTask, availableHeight+1)

		taskListStyle := lipgloss.NewStyle().Width(taskListWidth)
		taskListView = taskListStyle.Render(taskListView)

		return lipgloss.JoinHorizontal(lipgloss.Top, taskListView, sidebar)
	}

	return taskListView
}

// renderColumnHeaders renders the column headers for the task list
func (m Model) renderColumnHeaders(width int) string {
	// Calculate column widths dynamically
	cols := m.calculateColumnWidths(width)

	cursor := " "
	id := "ID"
	project := "PROJECT"
	priority := "P"
	due := "DUE"
	description := "DESCRIPTION"

	header := fmt.Sprintf("%s %-*s %-*s %s %-*s %s",
		cursor,
		cols.id, id,
		cols.project, project,
		priority,
		cols.due, due,
		description,
	)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("8")).
		Width(width)

	return headerStyle.Render(header)
}

// columnWidths holds the calculated widths for each column
type columnWidths struct {
	id          int
	project     int
	priority    int
	due         int
	description int
}

// calculateColumnWidths determines column widths based on available space
func (m Model) calculateColumnWidths(width int) columnWidths {
	// Fixed widths for some columns
	const (
		cursorWidth   = 2  // " " or "■ "
		idWidth       = 4  // Sequential ID (1-9999 should fit)
		priorityWidth = 1  // H/M/L
		dueWidth      = 10 // YYYY-MM-DD or "Today"
		minProject    = 10
		minDesc       = 20
		spacing       = 6 // spaces between columns
	)

	// Calculate remaining space for flexible columns
	fixedWidth := cursorWidth + idWidth + priorityWidth + dueWidth + spacing
	remainingWidth := width - fixedWidth

	if remainingWidth < minProject+minDesc {
		// Minimal widths
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

// renderTaskLine renders a single task line
func (m Model) renderTaskLine(task core.Task, isSelected bool, width int) string {
	cols := m.calculateColumnWidths(width)

	// Cursor
	cursor := " "
	if isSelected {
		cursor = "■"
	}

	// ID (taskwarrior's sequential ID)
	id := fmt.Sprintf("%d", task.ID)
	if task.ID == 0 {
		// Fall back to UUID prefix if ID is not set (e.g., for completed tasks)
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

	// Priority
	priority := "-"
	if task.Priority != "" {
		priority = string(task.Priority[0]) // H, M, L
	}

	// Due date
	due := "-"
	if task.Due != nil {
		due = task.FormatDueDate()
		if len(due) > cols.due {
			due = due[:cols.due]
		}
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
		priority,
		cols.due, due,
		description,
	)

	if isSelected {
		style := lipgloss.NewStyle().
			Background(lipgloss.Color("12")).
			Foreground(lipgloss.Color("0")).
			Width(width)
		return style.Render(line)
	}

	return line
}

// renderSidebar renders the task detail sidebar
func (m Model) renderSidebar(task core.Task, height int) string {
	lines := []string{
		"Task Details",
		"",
		fmt.Sprintf("UUID: %s", task.UUID),
		fmt.Sprintf("Description: %s", task.Description),
		fmt.Sprintf("Project: %s", task.Project),
		fmt.Sprintf("Status: %s", task.Status),
		fmt.Sprintf("Priority: %s", task.Priority),
	}

	if task.Due != nil {
		lines = append(lines, fmt.Sprintf("Due: %s", task.FormatDueDate()))
	}

	if len(task.Tags) > 0 {
		lines = append(lines, fmt.Sprintf("Tags: %v", task.Tags))
	}

	content := strings.Join(lines, "\n")

	sidebarWidth := m.width / 3
	if sidebarWidth < 30 {
		sidebarWidth = 30
	}

	sidebarStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		Padding(1, 2).
		Width(sidebarWidth).
		Height(height)

	return sidebarStyle.Render(content)
}

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	help := []string{
		"Help - Keyboard Shortcuts",
		"",
		"Navigation:",
		"  j/↓       - Move down",
		"  k/↑       - Move up",
		"  g         - Jump to first",
		"  G         - Jump to last",
		"  Tab       - Toggle sidebar",
		"",
		"Actions:",
		"  r         - Refresh tasks",
		"  q         - Quit",
		"",
		"Press ? or Esc to close help",
	}

	return lipgloss.NewStyle().
		Padding(2, 4).
		Render(strings.Join(help, "\n"))
}

// renderFilterInput renders the filter input prompt
func (m Model) renderFilterInput() string {
	return lipgloss.NewStyle().
		Padding(2, 4).
		Render("Filter: (Enter to apply, Esc to cancel)")
}

// renderConfirm renders the confirmation prompt
func (m Model) renderConfirm() string {
	return lipgloss.NewStyle().
		Padding(2, 4).
		Render("Confirm? (y/N)")
}

// renderFooter renders the footer with keybindings
func (m Model) renderFooter() string {
	var parts []string

	// Show error message if present
	if m.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)
		parts = append(parts, errorStyle.Render("Error: "+m.errorMessage))
	} else if m.statusMessage != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))
		parts = append(parts, statusStyle.Render(m.statusMessage))
	}

	// Show keybindings based on state
	keybindings := ""
	switch m.state {
	case StateNormal:
		keybindings = "q: quit | ?: help | r: refresh | tab: sidebar"
	case StateHelp:
		keybindings = "?: close help"
	case StateFilterInput:
		keybindings = "enter: apply | esc: cancel"
	case StateConfirm:
		keybindings = "y: confirm | n: cancel"
	}

	if keybindings != "" {
		parts = append(parts, keybindings)
	}

	footer := strings.Join(parts, " | ")

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(1, 1)

	return footerStyle.Render(footer)
}
