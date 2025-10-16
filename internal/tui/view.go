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
		return lipgloss.NewStyle().
			Padding(2, 4).
			Render("No tasks found. Press 'n' to create a new task.")
	}

	var lines []string

	// Calculate available height for task list (subtract header and footer)
	availableHeight := m.height - 4
	if availableHeight < 1 {
		availableHeight = 10
	}

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
		line := m.renderTaskLine(task, i == m.selectedIndex)
		lines = append(lines, line)
	}

	taskListView := strings.Join(lines, "\n")

	// Show sidebar if enabled
	if m.viewMode == ViewModeListWithSidebar && m.selectedIndex < len(m.tasks) {
		selectedTask := m.tasks[m.selectedIndex]
		sidebar := m.renderSidebar(selectedTask)

		// Split view horizontally
		taskListWidth := m.width * 2 / 3
		if taskListWidth < 40 {
			taskListWidth = 40
		}

		taskListStyle := lipgloss.NewStyle().Width(taskListWidth)
		taskListView = taskListStyle.Render(taskListView)

		return lipgloss.JoinHorizontal(lipgloss.Top, taskListView, sidebar)
	}

	return taskListView
}

// renderTaskLine renders a single task line
func (m Model) renderTaskLine(task core.Task, isSelected bool) string {
	// Simple format: [>] ID Project Description
	cursor := " "
	if isSelected {
		cursor = "■"
	}

	// Format ID (first 8 chars of UUID)
	id := task.UUID
	if len(id) > 8 {
		id = id[:8]
	}

	// Format project
	project := task.Project
	if project == "" {
		project = "-"
	}
	if len(project) > 15 {
		project = project[:12] + "..."
	}

	// Format description
	description := task.Description
	maxDescLen := 50
	if len(description) > maxDescLen {
		description = description[:maxDescLen-3] + "..."
	}

	line := fmt.Sprintf("%s %s %-15s %s", cursor, id, project, description)

	if isSelected {
		style := lipgloss.NewStyle().
			Background(lipgloss.Color("12")).
			Foreground(lipgloss.Color("0"))
		return style.Render(line)
	}

	return line
}

// renderSidebar renders the task detail sidebar
func (m Model) renderSidebar(task core.Task) string {
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

	sidebarStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		Padding(1, 2).
		Width(m.width / 3)

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
