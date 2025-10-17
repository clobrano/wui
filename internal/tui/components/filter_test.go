package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewFilter(t *testing.T) {
	filter := NewFilter()

	if filter.textInput.Value() != "" {
		t.Errorf("expected empty initial value, got %q", filter.textInput.Value())
	}

	if filter.textInput.Placeholder == "" {
		t.Error("expected non-empty placeholder")
	}
}

func TestFilter_Focus(t *testing.T) {
	filter := NewFilter()

	// Initially should not be focused
	if filter.textInput.Focused() {
		t.Error("expected filter to not be focused initially")
	}

	// Focus should work
	filter.Focus()
	if !filter.textInput.Focused() {
		t.Error("expected filter to be focused after Focus()")
	}
}

func TestFilter_Blur(t *testing.T) {
	filter := NewFilter()
	filter.Focus()

	filter.Blur()
	if filter.textInput.Focused() {
		t.Error("expected filter to not be focused after Blur()")
	}
}

func TestFilter_Value(t *testing.T) {
	filter := NewFilter()

	testValue := "+work -waiting"
	filter.SetValue(testValue)

	if filter.Value() != testValue {
		t.Errorf("expected value %q, got %q", testValue, filter.Value())
	}
}

func TestFilter_Update_TextInput(t *testing.T) {
	filter := NewFilter()
	filter.Focus()

	// Simulate typing
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t', 'e', 's', 't'}}
	updatedFilter, _ := filter.Update(msg)

	// The value should be updated by the textinput
	if updatedFilter.Value() == "" {
		t.Error("expected non-empty value after typing")
	}
}

func TestFilter_View(t *testing.T) {
	filter := NewFilter()
	view := filter.View()

	if view == "" {
		t.Error("expected non-empty view output")
	}

	// View should contain the prompt
	if len(view) < 5 {
		t.Error("expected view to contain prompt and input field")
	}
}

func TestFilter_SetValue(t *testing.T) {
	filter := NewFilter()

	tests := []struct {
		name  string
		value string
	}{
		{"simple filter", "+work"},
		{"complex filter", "+work -waiting priority:H"},
		{"empty filter", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter.SetValue(tt.value)
			if filter.Value() != tt.value {
				t.Errorf("expected %q, got %q", tt.value, filter.Value())
			}
		})
	}
}
