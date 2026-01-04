package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Calendar is a date picker component that displays a month view calendar
type Calendar struct {
	currentMonth  time.Time
	selectedDate  time.Time
	cursorDay     int // 1-31, which day in the month is selected
	dateInput     textinput.Model
	editingDate   bool // true when editing the date field at bottom
	width         int
	height        int
}

// CalendarResult is returned when a date is selected
type CalendarResult struct {
	Date     time.Time
	Canceled bool
}

// NewCalendar creates a new calendar picker
func NewCalendar(initialDate time.Time) Calendar {
	if initialDate.IsZero() {
		initialDate = time.Now()
	}

	// Create text input for date editing
	ti := textinput.New()
	ti.Placeholder = "YYYY-MM-DD"
	ti.CharLimit = 10
	ti.Width = 20
	ti.SetValue(initialDate.Format("2006-01-02"))

	// Start with current day selected
	return Calendar{
		currentMonth: time.Date(initialDate.Year(), initialDate.Month(), 1, 0, 0, 0, 0, time.Local),
		selectedDate: initialDate,
		cursorDay:    initialDate.Day(),
		dateInput:    ti,
		editingDate:  false,
		width:        28, // 7 days × 4 chars per day
		height:       10,
	}
}

// Init implements tea.Model
func (c Calendar) Init() tea.Cmd {
	return nil
}

// Update handles keyboard input
func (c Calendar) Update(msg tea.Msg) (Calendar, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If editing the date field, delegate to textinput
		if c.editingDate {
			switch msg.String() {
			case "enter":
				// Try to parse the entered date
				if date, err := time.Parse("2006-01-02", c.dateInput.Value()); err == nil {
					c.selectedDate = date
					c.currentMonth = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
					c.cursorDay = date.Day()
					c.editingDate = false
					c.dateInput.Blur()
					return c, nil
				}
				// Invalid date, stay in editing mode
				return c, nil
			case "esc":
				// Cancel editing
				c.editingDate = false
				c.dateInput.Blur()
				c.dateInput.SetValue(c.selectedDate.Format("2006-01-02"))
				return c, nil
			default:
				c.dateInput, cmd = c.dateInput.Update(msg)
				return c, cmd
			}
		}

		// Handle calendar navigation
		switch msg.String() {
		case "b", "B":
			// Previous month
			c.currentMonth = c.currentMonth.AddDate(0, -1, 0)
			c.adjustCursorDay()
			return c, nil

		case "n", "N":
			// Next month
			c.currentMonth = c.currentMonth.AddDate(0, 1, 0)
			c.adjustCursorDay()
			return c, nil

		case "left", "h":
			// Previous day
			c.cursorDay--
			if c.cursorDay < 1 {
				c.currentMonth = c.currentMonth.AddDate(0, -1, 0)
				c.cursorDay = daysInMonth(c.currentMonth)
			}
			c.updateSelectedDate()
			return c, nil

		case "right", "l":
			// Next day
			maxDays := daysInMonth(c.currentMonth)
			c.cursorDay++
			if c.cursorDay > maxDays {
				c.currentMonth = c.currentMonth.AddDate(0, 1, 0)
				c.cursorDay = 1
			}
			c.updateSelectedDate()
			return c, nil

		case "up", "k":
			// Previous week
			c.cursorDay -= 7
			if c.cursorDay < 1 {
				c.currentMonth = c.currentMonth.AddDate(0, -1, 0)
				maxDays := daysInMonth(c.currentMonth)
				c.cursorDay = maxDays + c.cursorDay
			}
			c.updateSelectedDate()
			return c, nil

		case "down", "j":
			// Next week
			maxDays := daysInMonth(c.currentMonth)
			c.cursorDay += 7
			if c.cursorDay > maxDays {
				c.currentMonth = c.currentMonth.AddDate(0, 1, 0)
				c.cursorDay = c.cursorDay - maxDays
			}
			c.updateSelectedDate()
			return c, nil

		case "e":
			// Edit date field
			c.editingDate = true
			c.dateInput.Focus()
			return c, textinput.Blink

		case "t":
			// Today
			today := time.Now()
			c.currentMonth = time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local)
			c.cursorDay = today.Day()
			c.updateSelectedDate()
			return c, nil
		}
	}

	return c, nil
}

// View renders the calendar
func (c Calendar) View() string {
	var b strings.Builder

	// Title: Month Year
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Align(lipgloss.Center).
		Width(c.width)

	title := fmt.Sprintf("%s %d", c.currentMonth.Format("January"), c.currentMonth.Year())
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	// Day headers (Su Mo Tu We Th Fr Sa)
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Bold(true)

	headers := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	for _, h := range headers {
		b.WriteString(headerStyle.Render(fmt.Sprintf(" %2s ", h)))
	}
	b.WriteString("\n")

	// Calculate first day of month and total days
	firstDay := c.currentMonth.Weekday() // 0 = Sunday
	maxDays := daysInMonth(c.currentMonth)

	// Day cell styles
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("205")).
		Bold(true)
	todayStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Underline(true)

	today := time.Now()
	isCurrentMonth := today.Year() == c.currentMonth.Year() && today.Month() == c.currentMonth.Month()

	// Render calendar grid
	day := 1
	for week := 0; week < 6; week++ {
		if day > maxDays {
			break
		}

		for weekday := 0; weekday < 7; weekday++ {
			if week == 0 && weekday < int(firstDay) {
				// Empty cells before first day
				b.WriteString("    ")
			} else if day > maxDays {
				// Empty cells after last day
				b.WriteString("    ")
			} else {
				// Render day
				cellText := fmt.Sprintf(" %2d ", day)

				if day == c.cursorDay {
					b.WriteString(cursorStyle.Render(cellText))
				} else if isCurrentMonth && day == today.Day() {
					b.WriteString(todayStyle.Render(cellText))
				} else {
					b.WriteString(normalStyle.Render(cellText))
				}

				day++
			}
		}
		b.WriteString("\n")
	}

	// Navigation hints (more compact)
	b.WriteString("\n")
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Width(c.width)
	b.WriteString(hintStyle.Render("B/N:month T:today E:edit\n"))
	b.WriteString(hintStyle.Render("↑↓←→:nav ⏎:ok ⎋:cancel\n"))

	// Date input field at bottom
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	b.WriteString(labelStyle.Render("Date: "))

	if c.editingDate {
		b.WriteString(c.dateInput.View())
	} else {
		dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		b.WriteString(dateStyle.Render(c.selectedDate.Format("2006-01-02")))
	}

	// Wrap in border with minimal padding
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(0, 1)

	return boxStyle.Render(b.String())
}

// GetSelectedDate returns the currently selected date
func (c Calendar) GetSelectedDate() time.Time {
	return c.selectedDate
}

// Helper functions

func daysInMonth(t time.Time) int {
	// Get the first day of next month, then subtract one day
	nextMonth := t.AddDate(0, 1, 0)
	firstOfNextMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	lastOfThisMonth := firstOfNextMonth.AddDate(0, 0, -1)
	return lastOfThisMonth.Day()
}

func (c *Calendar) adjustCursorDay() {
	maxDays := daysInMonth(c.currentMonth)
	if c.cursorDay > maxDays {
		c.cursorDay = maxDays
	}
	c.updateSelectedDate()
}

func (c *Calendar) updateSelectedDate() {
	c.selectedDate = time.Date(
		c.currentMonth.Year(),
		c.currentMonth.Month(),
		c.cursorDay,
		0, 0, 0, 0,
		time.Local,
	)
	c.dateInput.SetValue(c.selectedDate.Format("2006-01-02"))
}
