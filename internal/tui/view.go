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

	// Build the view components
	sectionsBar := m.renderSections()
	content := m.renderContent()
	footer := m.renderFooter()

	// Calculate actual heights
	sectionsHeight := strings.Count(sectionsBar, "\n") + 1
	contentLines := strings.Split(content, "\n")
	footerHeight := strings.Count(footer, "\n") + 1

	// Ensure content doesn't exceed available space
	maxContentHeight := m.height - sectionsHeight - footerHeight
	if m.state == StateFilterInput || m.state == StateModifyInput ||
		m.state == StateAnnotateInput || m.state == StateNewTaskInput {
		maxContentHeight -= 2 // Input prompt takes 2 lines
	}

	// Trim content if necessary
	if len(contentLines) > maxContentHeight {
		contentLines = contentLines[:maxContentHeight]
		content = strings.Join(contentLines, "\n")
	}

	// Build sections slice for vertical join
	var sections []string
	sections = append(sections, sectionsBar)
	sections = append(sections, content)

	// Input prompt area (if in input mode)
	if m.state == StateFilterInput || m.state == StateModifyInput ||
		m.state == StateAnnotateInput || m.state == StateNewTaskInput {
		sections = append(sections, m.renderInputPrompt())
	}

	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderHeader renders the header section
func (m Model) renderHeader() string {
	title := "wui - Warrior UI"

	// Add active filter if not default
	if m.activeFilter != "" && m.currentSection != nil && m.activeFilter != m.currentSection.Filter {
		title += fmt.Sprintf(" | Filter: %s", m.activeFilter)
	}

	return m.styles.Header.Width(m.width).Render(title)
}

// renderSections renders the sections navigation bar
func (m Model) renderSections() string {
	return m.sections.View()
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
		"Task Navigation:",
		"  j/↓       - Move down",
		"  k/↑       - Move up",
		"  g         - Jump to first",
		"  G         - Jump to last",
		"  1-9       - Quick jump to task",
		"",
		"Section Navigation:",
		"  Tab/l/→   - Next section",
		"  Shift+Tab/h/← - Previous section",
		"  1-5       - Jump to section",
		"",
		"Task Actions:",
		"  d         - Mark done",
		"  s         - Start/Stop toggle",
		"  x         - Delete",
		"  e         - Edit",
		"  n         - New task",
		"  m         - Modify (e.g. wait:tomorrow)",
		"  a         - Annotate",
		"  u         - Undo",
		"",
		"Other:",
		"  Enter     - Toggle sidebar",
		"  /         - Filter tasks",
		"  r         - Refresh tasks",
		"  ?         - Toggle help",
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

	content := m.styles.InputPrompt.Render(prompt) + inputView + m.styles.InputHint.Render(hint)

	// Add a separator line above the input
	separator := m.styles.Separator.Width(m.width).Render(strings.Repeat("─", m.width))

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

	// Show loading indicator if loading
	if m.isLoading {
		parts = append(parts, m.styles.LoadingIndicator.Render("⣾ Loading..."))
	} else if m.errorMessage != "" {
		// Show error message if present
		parts = append(parts, m.styles.Error.Render("✗ "+m.errorMessage))
	} else if m.statusMessage != "" {
		parts = append(parts, m.styles.Success.Render("✓ "+m.statusMessage))
	}

	// Show keybindings based on state
	keybindings := ""
	switch m.state {
	case StateNormal:
		keybindings = "d: done | s: start/stop | x: delete | e: edit | n: new | m: modify | a: annotate | u: undo"
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

	return m.styles.Footer.Render(footer)
}
