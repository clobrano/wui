package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MessageType represents the type of status message
type MessageType int

const (
	// MessageInfo is an informational message
	MessageInfo MessageType = iota
	// MessageSuccess is a success message
	MessageSuccess
	// MessageError is an error message
	MessageError
	// MessageWarning is a warning message
	MessageWarning
)

// StatusMessage represents a transient message to display
type StatusMessage struct {
	Content  string
	Type     MessageType
	Duration time.Duration
}

// ClearStatusMsg is sent when the status message should be cleared
type ClearStatusMsg struct{}

// StatusBar component displays transient status messages
type StatusBar struct {
	width          int
	currentMessage *StatusMessage
	showKeybindings bool
	keybindings    string
	styles         StatusBarStyles
}

// StatusBarStyles contains styling for the status bar
type StatusBarStyles struct {
	Normal  lipgloss.Style
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Keys    lipgloss.Style
}

// DefaultStatusBarStyles returns the default status bar styles
func DefaultStatusBarStyles() StatusBarStyles {
	return StatusBarStyles{
		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true),
		Keys: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}

// NewStatusBar creates a new status bar
func NewStatusBar(width int, showKeybindings bool, styles StatusBarStyles) StatusBar {
	return StatusBar{
		width:           width,
		currentMessage:  nil,
		showKeybindings: showKeybindings,
		keybindings:     "",
		styles:          styles,
	}
}

// SetMessage sets a new status message with a duration
// Returns a command that will clear the message after the duration
func (s *StatusBar) SetMessage(content string, msgType MessageType, duration time.Duration) tea.Cmd {
	s.currentMessage = &StatusMessage{
		Content:  content,
		Type:     msgType,
		Duration: duration,
	}

	// Return command to clear message after duration
	if duration > 0 {
		return tea.Tick(duration, func(time.Time) tea.Msg {
			return ClearStatusMsg{}
		})
	}
	return nil
}

// SetKeybindings sets the keybindings text to display
func (s *StatusBar) SetKeybindings(keys string) {
	s.keybindings = keys
}

// SetWidth updates the width of the status bar
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// ClearMessage clears the current status message
func (s *StatusBar) ClearMessage() {
	s.currentMessage = nil
}

// Update handles messages for the status bar
func (s StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	switch msg := msg.(type) {
	case ClearStatusMsg:
		s.currentMessage = nil
		return s, nil
	case tea.WindowSizeMsg:
		s.width = msg.Width
		return s, nil
	}
	return s, nil
}

// View renders the status bar
func (s StatusBar) View() string {
	if s.width == 0 {
		return ""
	}

	var parts []string

	// Show status message if present (takes priority)
	if s.currentMessage != nil {
		var style lipgloss.Style
		prefix := ""

		switch s.currentMessage.Type {
		case MessageSuccess:
			style = s.styles.Success
			prefix = "✓ "
		case MessageError:
			style = s.styles.Error
			prefix = "✗ "
		case MessageWarning:
			style = s.styles.Warning
			prefix = "⚠ "
		case MessageInfo:
			style = s.styles.Normal
			prefix = "ℹ "
		}

		parts = append(parts, style.Render(prefix+s.currentMessage.Content))
	}

	// Show keybindings if enabled and no message (or append if there's space)
	if s.showKeybindings && s.keybindings != "" {
		if s.currentMessage == nil {
			parts = append(parts, s.styles.Keys.Render(s.keybindings))
		}
	}

	// Join parts with separator
	content := ""
	if len(parts) > 0 {
		content = lipgloss.JoinHorizontal(lipgloss.Left, parts...)
	}

	// Render with padding and width constraint
	return lipgloss.NewStyle().
		Width(s.width).
		MaxWidth(s.width).
		Padding(0, 1).
		Render(content)
}
