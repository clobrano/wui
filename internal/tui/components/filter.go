package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Filter is a component for entering task filter syntax
type Filter struct {
	textInput textinput.Model
}

// NewFilter creates a new filter input component
func NewFilter() Filter {
	ti := textinput.New()
	ti.Placeholder = "Enter filter (e.g., +work -waiting priority:H)"
	ti.CharLimit = 200
	ti.Width = 50

	return Filter{
		textInput: ti,
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
