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
	// Footer has padding(1,1) which adds 2 extra lines (top and bottom)
	footerHeight := strings.Count(footer, "\n") + 1 + 2 // +2 for vertical padding

	// Ensure content doesn't exceed available space
	// Reserve space for the bottom border line (1 line)
	maxContentHeight := m.height - sectionsHeight - footerHeight
	if m.state == StateFilterInput || m.state == StateModifyInput ||
		m.state == StateAnnotateInput || m.state == StateNewTaskInput {
		maxContentHeight -= 2 // Input prompt takes 2 lines
	}

	// Add safety margin to ensure footer is always visible
	if maxContentHeight < 2 {
		maxContentHeight = 2 // At least 2 lines: 1 for content, 1 for bottom border
	}

	// Trim content if necessary, but keep the last line (bottom border)
	if len(contentLines) > maxContentHeight {
		// Keep the bottom border (last line) and trim the rest
		bottomBorderLine := contentLines[len(contentLines)-1]
		contentLines = contentLines[:maxContentHeight-1]
		contentLines = append(contentLines, bottomBorderLine)
		content = strings.Join(contentLines, "\n")
	}

	// Build sections slice for vertical join
	var sections []string
	sections = append(sections, sectionsBar)
	sections = append(sections, content)

	// Check input mode configuration
	inputMode := m.config.TUI.InputMode
	if inputMode == "" {
		inputMode = "floating" // Default to floating
	}

	// Input prompt area (if in input mode)
	if m.state == StateFilterInput || m.state == StateModifyInput ||
		m.state == StateAnnotateInput || m.state == StateNewTaskInput {

		if inputMode == "floating" {
			// Floating window will be overlaid after the base view is built
			// Don't add to sections - it will be rendered on top
		} else {
			// Bottom mode - add input prompt to sections
			sections = append(sections, m.renderInputPrompt())
		}
	}

	sections = append(sections, footer)

	baseView := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Overlay floating window if in floating mode and input state
	if inputMode == "floating" &&
		(m.state == StateFilterInput || m.state == StateModifyInput ||
			m.state == StateAnnotateInput || m.state == StateNewTaskInput) {
		baseView = m.renderFloatingInput(baseView)
	}

	// If calendar is active, overlay it on top of everything
	if m.calendarActive {
		calendarView := m.calendar.View()

		// Place calendar in the center of the screen as an overlay
		baseView = lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			calendarView,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
		)
	}

	return baseView
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
	var content string

	if m.viewMode == ViewModeListWithSidebar {
		// Render task list and sidebar side by side
		taskListView := m.taskList.View()
		sidebarView := m.sidebar.View()
		content = lipgloss.JoinHorizontal(lipgloss.Top, taskListView, sidebarView)
	} else if m.viewMode == ViewModeSmallTaskDetail {
		// Render full-screen task detail view (for small screens)
		content = m.sidebar.View()
	} else {
		// Render just the task list (ViewModeList or ViewModeSmall)
		content = m.taskList.View()
	}

	// Add bottom border line
	bottomBorder := m.styles.Separator.Width(m.width).Render(strings.Repeat("─", m.width))
	return content + "\n" + bottomBorder
}

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	return m.help.View()
}

// renderInputPrompt renders the input prompt area at the bottom
func (m Model) renderInputPrompt() string {
	var prompt, hint, inputView string

	switch m.state {
	case StateFilterInput:
		prompt = "Filter: "
		hint = "(Enter to apply, Esc to cancel)"
		inputView = m.filter.View()
	case StateModifyInput:
		prompt = "Modify: "
		hint = "(Enter to apply, Esc to cancel)"
		inputView = m.modifyInput.View()
	case StateAnnotateInput:
		prompt = "Annotate: "
		hint = "(Enter to apply, Esc to cancel)"
		inputView = m.annotateInput.View()
	case StateNewTaskInput:
		prompt = "New Task: "
		hint = "(Enter to create, Esc to cancel)"
		inputView = m.newTaskInput.View()
	default:
		return ""
	}

	// Calculate widths to prevent overlap
	// Available width = terminal width - padding (2 for left/right)
	availableWidth := m.width - 2
	promptWidth := lipgloss.Width(prompt)
	hintWidth := lipgloss.Width(hint)

	// Input should take remaining space, leaving room for prompt and hint
	// Add some spacing between input and hint
	spacing := 2
	inputWidth := availableWidth - promptWidth - hintWidth - spacing
	if inputWidth < 20 {
		// Minimum width for input
		inputWidth = 20
	}

	// Render components
	promptRendered := m.styles.InputPrompt.Render(prompt)
	hintRendered := m.styles.InputHint.Render(hint)

	// Build the content line with proper spacing
	// The hint should appear at the right edge
	remainingSpace := availableWidth - promptWidth - lipgloss.Width(inputView) - hintWidth
	if remainingSpace < 1 {
		remainingSpace = 1
	}
	spacer := strings.Repeat(" ", remainingSpace)

	content := promptRendered + inputView + spacer + hintRendered

	// Add a separator line above the input
	separator := m.styles.Separator.Width(m.width).Render(strings.Repeat("─", m.width))

	return separator + "\n" + lipgloss.NewStyle().
		Padding(0, 1).
		Render(content)
}

// renderFloatingInput renders the input prompt as a floating window overlaid on the base view
func (m Model) renderFloatingInput(baseView string) string {
	var title, hint, inputView string

	switch m.state {
	case StateFilterInput:
		title = "Filter Tasks"
		hint = "Enter: Apply  •  Esc: Cancel  •  ↑↓: History"
		inputView = m.filter.View()
	case StateModifyInput:
		title = "Modify Tasks"
		hint = "Enter: Apply  •  Esc: Cancel  •  ↑↓: History"
		inputView = m.modifyInput.View()
	case StateAnnotateInput:
		title = "Add Annotation"
		hint = "Enter: Apply  •  Esc: Cancel"
		inputView = m.annotateInput.View()
	case StateNewTaskInput:
		title = "New Task"
		hint = "Enter: Create  •  Esc: Cancel"
		inputView = m.newTaskInput.View()
	default:
		return baseView
	}

	// Get the dimensions
	baseHeight := lipgloss.Height(baseView)
	baseWidth := m.width

	// Build window content with input
	windowContent := inputView

	// Create the bordered window
	windowBox := m.styles.FloatingWindowBox.Render(
		m.styles.FloatingWindowTitle.Render(title) + "\n\n" +
			windowContent + "\n\n" +
			m.styles.InputHint.Render(hint),
	)

	// Calculate position (center horizontally, slightly above center vertically)
	windowHeight := lipgloss.Height(windowBox)
	windowWidth := lipgloss.Width(windowBox)

	// Position slightly above center for better visual balance
	verticalPos := max(0, (baseHeight-windowHeight)/2-2)

	// Split base view into lines
	baseLines := strings.Split(baseView, "\n")

	// Overlay the floating window
	// We'll replace lines at the vertical position with the window content
	windowLines := strings.Split(windowBox, "\n")

	for i, windowLine := range windowLines {
		lineIndex := verticalPos + i
		if lineIndex < len(baseLines) {
			// Center the window line horizontally
			padding := max(0, (baseWidth-windowWidth)/2)
			baseLines[lineIndex] = strings.Repeat(" ", padding) + windowLine
		}
	}

	return strings.Join(baseLines, "\n")
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
	if m.calendarActive {
		keybindings = "B/N: prev/next month | T: today | E: edit date | arrows/hjkl: navigate | enter: select | esc: cancel"
	} else {
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
			keybindings = "enter: apply | esc: cancel | tab: date picker (after due:/scheduled:/sched:)"
		case StateAnnotateInput:
			keybindings = "enter: apply | esc: cancel"
		case StateNewTaskInput:
			keybindings = "enter: create | esc: cancel | tab: date picker (after due:/scheduled:/sched:)"
		}
	}

	if keybindings != "" {
		parts = append(parts, keybindings)
	}

	footer := strings.Join(parts, " | ")

	return m.styles.Footer.
		Width(m.width).
		MaxWidth(m.width).
		Render(footer)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
