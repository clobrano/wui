package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Filter is a component for entering task filter syntax
type Filter struct {
	textInput    textinput.Model
	history      []string // History of previous filter commands
	historyIndex int      // Current position in history (-1 means not navigating history)
	currentInput string   // Temporary storage for current input when navigating history
}

// NewFilter creates a new filter input component
func NewFilter() Filter {
	ti := textinput.New()
	ti.Placeholder = "Enter filter (e.g., +work -waiting priority:H)"
	ti.CharLimit = 200
	ti.Width = 50

	return Filter{
		textInput:    ti,
		history:      []string{},
		historyIndex: -1,
	}
}

// Focus sets focus on the filter input
func (f *Filter) Focus() tea.Cmd {
	return f.textInput.Focus()
}

// Blur removes focus from the filter input
func (f *Filter) Blur() {
	f.textInput.Blur()
}

// Value returns the current filter text
func (f Filter) Value() string {
	return f.textInput.Value()
}

// SetValue sets the filter text
func (f *Filter) SetValue(value string) {
	f.textInput.SetValue(value)
}

// SetWidth sets the width of the filter input
func (f *Filter) SetWidth(width int) {
	f.textInput.Width = width
}

// Update handles messages for the filter input
func (f Filter) Update(msg tea.Msg) (Filter, tea.Cmd) {
	var cmd tea.Cmd
	f.textInput, cmd = f.textInput.Update(msg)
	return f, cmd
}

// View renders the filter input
func (f Filter) View() string {
	return f.textInput.View()
}

// AddToHistory adds a filter command to the history
func (f *Filter) AddToHistory(value string) {
	// Don't add empty values or duplicates of the most recent entry
	if value == "" {
		return
	}
	if len(f.history) > 0 && f.history[len(f.history)-1] == value {
		return
	}

	f.history = append(f.history, value)
	f.historyIndex = -1 // Reset history navigation
}

// NavigateHistoryUp moves to the previous command in history
func (f *Filter) NavigateHistoryUp() {
	if len(f.history) == 0 {
		return
	}

	// If we're starting history navigation, save the current input
	if f.historyIndex == -1 {
		f.currentInput = f.textInput.Value()
		f.historyIndex = len(f.history)
	}

	// Move up in history (towards older entries)
	if f.historyIndex > 0 {
		f.historyIndex--
		f.textInput.SetValue(f.history[f.historyIndex])
		// Move cursor to end of text
		f.textInput.SetCursor(len(f.history[f.historyIndex]))
	}
}

// NavigateHistoryDown moves to the next command in history (or back to current input)
func (f *Filter) NavigateHistoryDown() {
	if f.historyIndex == -1 {
		return // Not navigating history
	}

	// Move down in history (towards newer entries)
	f.historyIndex++

	if f.historyIndex >= len(f.history) {
		// Reached the end, restore current input
		f.textInput.SetValue(f.currentInput)
		f.textInput.SetCursor(len(f.currentInput))
		f.historyIndex = -1
	} else {
		f.textInput.SetValue(f.history[f.historyIndex])
		f.textInput.SetCursor(len(f.history[f.historyIndex]))
	}
}

// ResetHistoryNavigation resets the history navigation state
func (f *Filter) ResetHistoryNavigation() {
	f.historyIndex = -1
	f.currentInput = ""
}
