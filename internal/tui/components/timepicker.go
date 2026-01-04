package components

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TimePicker is a time selection component for HH:MM format
type TimePicker struct {
	hour   int  // 0-23
	minute int  // 0-59
	focus  bool // true if focused on hour, false for minute
	width  int
	height int
}

// NewTimePicker creates a new time picker with default to current hour, minute 00
func NewTimePicker() TimePicker {
	now := time.Now()
	return TimePicker{
		hour:   now.Hour(),
		minute: 0, // Always default to :00
		focus:  true,
		width:  28,
		height: 8,
	}
}

// Update handles keyboard input
func (t TimePicker) Update(msg tea.Msg) (TimePicker, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if t.focus {
				// Increment hour
				t.hour = (t.hour + 1) % 24
			} else {
				// Increment minute by 5
				t.minute = (t.minute + 5) % 60
			}
			return t, nil

		case "down", "j":
			if t.focus {
				// Decrement hour
				t.hour = (t.hour - 1 + 24) % 24
			} else {
				// Decrement minute by 5
				t.minute = (t.minute - 5 + 60) % 60
			}
			return t, nil

		case "right", "l", "tab":
			// Move focus to minute
			t.focus = false
			return t, nil

		case "left", "h", "shift+tab":
			// Move focus to hour
			t.focus = true
			return t, nil

		case "n":
			// Set to current time (now)
			now := time.Now()
			t.hour = now.Hour()
			t.minute = 0 // Still round to :00
			return t, nil
		}
	}

	return t, nil
}

// View renders the time picker
func (t TimePicker) View() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Align(lipgloss.Center).
		Width(t.width)

	b.WriteString(titleStyle.Render("Select Time"))
	b.WriteString("\n\n")

	// Time display with focus indicators
	hourStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	minuteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	focusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("205")).
		Bold(true)

	hourText := fmt.Sprintf(" %02d ", t.hour)
	minuteText := fmt.Sprintf(" %02d ", t.minute)

	if t.focus {
		hourText = focusStyle.Render(hourText)
		minuteText = minuteStyle.Render(minuteText)
	} else {
		hourText = hourStyle.Render(hourText)
		minuteText = focusStyle.Render(minuteText)
	}

	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(":")

	timeDisplay := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(t.width).
		Render(hourText + separator + minuteText)

	b.WriteString(timeDisplay)
	b.WriteString("\n\n")

	// Navigation hints
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	b.WriteString(hintStyle.Render("↑↓:change Tab/←→:switch\n"))
	b.WriteString(hintStyle.Render("N:now (hour:00)\n"))
	b.WriteString(hintStyle.Render("⏎:select ⎋:cancel\n"))

	// Current selection
	b.WriteString("\n")
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	b.WriteString(labelStyle.Render("Time: "))

	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	b.WriteString(timeStyle.Render(t.GetFormattedTime()))

	// Wrap in border with minimal padding
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(0, 1)

	return boxStyle.Render(b.String())
}

// GetFormattedTime returns the time in HH:MM format
func (t TimePicker) GetFormattedTime() string {
	return fmt.Sprintf("%02d:%02d", t.hour, t.minute)
}

// GetHour returns the selected hour
func (t TimePicker) GetHour() int {
	return t.hour
}

// GetMinute returns the selected minute
func (t TimePicker) GetMinute() int {
	return t.minute
}
