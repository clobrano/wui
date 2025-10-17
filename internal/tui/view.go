package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

	// Input prompt area (if in input mode)
	if m.state == StateFilterInput || m.state == StateModifyInput ||
	   m.state == StateAnnotateInput || m.state == StateNewTaskInput {
		sections = append(sections, m.renderInputPrompt())
	}

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
	case StateConfirm:
		return m.renderConfirm()
	default:
		// Always show task list (even when in input modes)
		return m.renderTaskListWithComponents()
	}
}

// renderTaskListWithComponents renders the task list using components
func (m Model) renderTaskListWithComponents() string {
	if m.viewMode == ViewModeListWithSidebar {
		// Render task list and sidebar side by side
		taskListView := m.taskList.View()
		sidebarView := m.sidebar.View()
		return lipgloss.JoinHorizontal(lipgloss.Top, taskListView, sidebarView)
	}

	// Render just the task list
	return m.taskList.View()
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

// renderInputPrompt renders the input prompt area at the bottom
func (m Model) renderInputPrompt() string {
	var prompt, hint, inputView string

	switch m.state {
	case StateFilterInput:
		prompt = "Filter: "
		hint = " (Enter to apply, Esc to cancel)"
		inputView = m.filter.View()
	case StateModifyInput:
		prompt = "Modify: "
		hint = " (Enter to apply, Esc to cancel)"
		inputView = m.modifyInput.View()
	case StateAnnotateInput:
		prompt = "Annotate: "
		hint = " (Enter to apply, Esc to cancel)"
		inputView = m.annotateInput.View()
	case StateNewTaskInput:
		prompt = "New Task: "
		hint = " (Enter to create, Esc to cancel)"
		inputView = m.newTaskInput.View()
	default:
		return ""
	}

	promptStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12"))

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	content := promptStyle.Render(prompt) + inputView + hintStyle.Render(hint)

	// Add a separator line above the input
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Width(m.width).
		Render(strings.Repeat("─", m.width))

	return separator + "\n" + lipgloss.NewStyle().
		Padding(0, 1).
		Render(content)
}

// renderConfirm renders the confirmation prompt
func (m Model) renderConfirm() string {
	message := "Confirm? (y/N)"

	if m.confirmAction == "delete" {
		selectedTask := m.taskList.SelectedTask()
		if selectedTask != nil {
			message = fmt.Sprintf("Delete task '%s'? (y/N)", selectedTask.Description)
		}
	}

	return lipgloss.NewStyle().
		Padding(2, 4).
		Render(message)
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
		keybindings = "d: done | x: delete | e: edit | n: new | m: modify | a: annotate | u: undo | /: filter | r: refresh"
	case StateHelp:
		keybindings = "?: close help"
	case StateFilterInput:
		keybindings = "enter: apply | esc: cancel"
	case StateConfirm:
		keybindings = "y: confirm | n: cancel"
	case StateModifyInput:
		keybindings = "enter: apply | esc: cancel"
	case StateAnnotateInput:
		keybindings = "enter: apply | esc: cancel"
	case StateNewTaskInput:
		keybindings = "enter: create | esc: cancel"
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
