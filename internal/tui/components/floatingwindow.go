package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FloatingWindowStyles contains styling for the floating window
type FloatingWindowStyles struct {
	Border lipgloss.Style
	Title  lipgloss.Style
	Hint   lipgloss.Style
}

// FloatingWindow represents a centered overlay window
type FloatingWindow struct {
	title  string
	hint   string
	styles FloatingWindowStyles
}

// NewFloatingWindow creates a new floating window component
func NewFloatingWindow(title, hint string, styles FloatingWindowStyles) FloatingWindow {
	return FloatingWindow{
		title:  title,
		hint:   hint,
		styles: styles,
	}
}

// SetTitle updates the window title
func (fw *FloatingWindow) SetTitle(title string) {
	fw.title = title
}

// SetHint updates the hint text
func (fw *FloatingWindow) SetHint(hint string) {
	fw.hint = hint
}

// Render renders the floating window with the given content at the center of the screen
// content is typically the input field view
// width and height are the dimensions of the terminal
func (fw FloatingWindow) Render(content string, width, height int) string {
	// Calculate window dimensions
	// Width: 60% of terminal width, min 40, max 80
	windowWidth := max(40, min(80, width*60/100))

	// Build the window content
	var windowContent strings.Builder

	// Add title
	if fw.title != "" {
		title := fw.styles.Title.Render(fw.title)
		windowContent.WriteString(title)
		windowContent.WriteString("\n\n")
	}

	// Add input content
	windowContent.WriteString(content)
	windowContent.WriteString("\n")

	// Add hint
	if fw.hint != "" {
		hint := fw.styles.Hint.Render(fw.hint)
		windowContent.WriteString("\n")
		windowContent.WriteString(hint)
	}

	// Apply border and padding
	bordered := fw.styles.Border.
		Width(windowWidth - 4). // Account for border and padding
		Render(windowContent.String())

	// Center the window on the screen
	// Place it slightly above center for better visual balance
	verticalOffset := max(0, (height-lipgloss.Height(bordered))/2-2)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Top,
		bordered,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	) + strings.Repeat("\n", verticalOffset)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
