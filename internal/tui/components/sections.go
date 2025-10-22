package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clobrano/wui/internal/core"
)

// SectionsStyles holds the styles needed for rendering sections
type SectionsStyles struct {
	Active   lipgloss.Style
	Inactive lipgloss.Style
	Count    lipgloss.Style
}

// Sections represents the section navigation component
type Sections struct {
	Items       []core.Section
	ActiveIndex int
	TaskCount   int
	Width       int
	styles      SectionsStyles
}

// SectionChangedMsg is sent when the active section changes
type SectionChangedMsg struct {
	Section core.Section
}

// NewSections creates a new Sections component
func NewSections(sections []core.Section, width int, styles SectionsStyles) Sections {
	return NewSectionsWithIndex(sections, width, styles, 0)
}

// NewSectionsWithIndex creates a new Sections component with a specific active index
func NewSectionsWithIndex(sections []core.Section, width int, styles SectionsStyles, activeIndex int) Sections {
	// Validate activeIndex
	if activeIndex < 0 || activeIndex >= len(sections) {
		activeIndex = 0
	}

	return Sections{
		Items:       sections,
		ActiveIndex: activeIndex,
		TaskCount:   0,
		Width:       width,
		styles:      styles,
	}
}

// Update handles messages for the Sections component
func (s Sections) Update(msg tea.Msg) (Sections, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return s.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		s.Width = msg.Width
		return s, nil
	}

	return s, nil
}

// handleKeyPress handles keyboard input for section navigation
func (s Sections) handleKeyPress(msg tea.KeyMsg) (Sections, tea.Cmd) {
	if len(s.Items) == 0 {
		return s, nil
	}

	oldIndex := s.ActiveIndex

	switch msg.Type {
	case tea.KeyTab, tea.KeyRight:
		// Tab or right arrow -> next section
		s.ActiveIndex = (s.ActiveIndex + 1) % len(s.Items)

	case tea.KeyShiftTab, tea.KeyLeft:
		// Shift+Tab or left arrow -> previous section
		s.ActiveIndex--
		if s.ActiveIndex < 0 {
			s.ActiveIndex = len(s.Items) - 1
		}

	case tea.KeyRunes:
		// Handle number keys 1-9 for quick section selection
		// Also handle h/l for vim-like navigation
		if len(msg.Runes) == 1 {
			key := msg.Runes[0]
			if key >= '1' && key <= '9' {
				index := int(key - '1')
				if index < len(s.Items) {
					s.ActiveIndex = index
					// Always send command for explicit number key selection
					return s, s.sectionChangedCmd()
				} else {
					// Out of range, don't change section
					return s, nil
				}
			} else if key == 'l' {
				// 'l' -> next section (vim-like)
				s.ActiveIndex = (s.ActiveIndex + 1) % len(s.Items)
			} else if key == 'h' {
				// 'h' -> previous section (vim-like)
				s.ActiveIndex--
				if s.ActiveIndex < 0 {
					s.ActiveIndex = len(s.Items) - 1
				}
			} else {
				// Not a recognized key, ignore
				return s, nil
			}
		} else {
			return s, nil
		}
	default:
		return s, nil
	}

	// If section changed, send a command
	if oldIndex != s.ActiveIndex {
		return s, s.sectionChangedCmd()
	}

	return s, nil
}

// sectionChangedCmd creates a command that sends a SectionChangedMsg
func (s Sections) sectionChangedCmd() tea.Cmd {
	return func() tea.Msg {
		return SectionChangedMsg{
			Section: s.GetActiveSection(),
		}
	}
}

// View renders the section navigation tabs
func (s Sections) View() string {
	if len(s.Items) == 0 {
		return ""
	}

	var tabs []string

	for i, section := range s.Items {
		var style lipgloss.Style

		if i == s.ActiveIndex {
			// Active section style
			style = s.styles.Active
		} else {
			// Inactive section style
			style = s.styles.Inactive
		}

		// Display only search icon for Search section
		displayName := section.Name
		if section.Name == "Search" {
			displayName = "⌕"
		}

		tabs = append(tabs, style.Render(displayName))
	}

	tabsLine := strings.Join(tabs, " ")

	// Add task count for active section if set
	if s.TaskCount > 0 {
		taskCountStr := s.styles.Count.Render(fmt.Sprintf("(%d)", s.TaskCount))
		tabsLine += " " + taskCountStr
	}

	// Ensure the line spans the full width and add newline for proper vertical spacing
	return lipgloss.NewStyle().Width(s.Width).Render(tabsLine)
}

// GetActiveSection returns the currently active section
func (s Sections) GetActiveSection() core.Section {
	if len(s.Items) == 0 || s.ActiveIndex < 0 || s.ActiveIndex >= len(s.Items) {
		return core.Section{}
	}
	return s.Items[s.ActiveIndex]
}

// SetTaskCount sets the task count for display
func (s *Sections) SetTaskCount(count int) {
	s.TaskCount = count
}

// SetSize updates the width of the component
func (s *Sections) SetSize(width int) {
	s.Width = width
}

// IsProjectsView returns true if the active section is the Projects view
func (s Sections) IsProjectsView() bool {
	activeSection := s.GetActiveSection()
	return activeSection.Name == "Projects"
}

// IsTagsView returns true if the active section is the Tags view
func (s Sections) IsTagsView() bool {
	activeSection := s.GetActiveSection()
	return activeSection.Name == "Tags"
}
