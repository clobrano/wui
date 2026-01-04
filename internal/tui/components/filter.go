package components

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// Filter is a component for entering task filter syntax
type Filter struct {
	textArea     textarea.Model
	history      []string // History of previous filter commands
	historyIndex int      // Current position in history (-1 means not navigating history)
	currentInput string   // Temporary storage for current input when navigating history
}

// NewFilter creates a new filter input component
func NewFilter() Filter {
	ta := textarea.New()
	ta.Placeholder = "Type your text here..."
	ta.CharLimit = 1000
	ta.SetWidth(60)
	ta.SetHeight(4)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = ta.FocusedStyle.Base

	return Filter{
		textArea:     ta,
		history:      []string{},
		historyIndex: -1,
	}
}

// Focus sets focus on the filter input
func (f *Filter) Focus() tea.Cmd {
	f.textArea.Focus()
	return nil
}

// Blur removes focus from the filter input
func (f *Filter) Blur() {
	f.textArea.Blur()
}

// Value returns the current filter text
func (f Filter) Value() string {
	return f.textArea.Value()
}

// SetValue sets the filter text
func (f *Filter) SetValue(value string) {
	f.textArea.SetValue(value)
}

// SetWidth sets the width of the filter input
func (f *Filter) SetWidth(width int) {
	f.textArea.SetWidth(width)
}

// SetHeight sets the height of the filter input
func (f *Filter) SetHeight(height int) {
	f.textArea.SetHeight(height)
}

// Update handles messages for the filter input
func (f Filter) Update(msg tea.Msg) (Filter, tea.Cmd) {
	var cmd tea.Cmd
	f.textArea, cmd = f.textArea.Update(msg)
	return f, cmd
}

// View renders the filter input
func (f Filter) View() string {
	return f.textArea.View()
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
		f.currentInput = f.textArea.Value()
		f.historyIndex = len(f.history)
	}

	// Move up in history (towards older entries)
	if f.historyIndex > 0 {
		f.historyIndex--

		// If this entry matches the current input, skip it and go to the previous one
		if f.historyIndex == len(f.history)-1 && f.history[f.historyIndex] == f.currentInput {
			if f.historyIndex > 0 {
				f.historyIndex--
			}
		}

		f.textArea.SetValue(f.history[f.historyIndex])
		// Move cursor to end of text
		f.textArea.CursorEnd()
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
		f.textArea.SetValue(f.currentInput)
		f.textArea.CursorEnd()
		f.historyIndex = -1
	} else {
		f.textArea.SetValue(f.history[f.historyIndex])
		f.textArea.CursorEnd()
	}
}

// ResetHistoryNavigation resets the history navigation state
func (f *Filter) ResetHistoryNavigation() {
	f.historyIndex = -1
	f.currentInput = ""
}
